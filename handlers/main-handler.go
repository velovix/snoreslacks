package handlers

import (
	"bytes"
	"encoding/gob"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
)

// Main responds to Slack slash requests.
type Main struct {
	Services

	Token string
}

func (h *Main) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	// Create the request context
	ctx, err := h.CtxCreator.Create(r)
	if err != nil {
		http.Error(w, "error processing context", 500)
		return
	}
	// Create a client for making HTTP requests
	client, err := h.ClientCreator.Create(ctx)
	if err != nil {
		http.Error(w, "error creating HTTP client", 500)
		return
	}
	// Create the Slack request
	slackReq, err := messaging.NewSlackRequest(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if slackReq.Token != h.Token { // Compare the tokens
		http.Error(w, "invalid token", 400)
		return
	}

	// Encode the Slack request into a binary blob to be sent off to workers
	slackReqBlob := &bytes.Buffer{}
	err = gob.NewEncoder(slackReqBlob).Encode(slackReq)
	if err != nil {
		// An error happened while encoding the Slack request
		h.Log.Errorf(ctx, "while encoding the Slack request: %s", err)
		messaging.Send(client, slackReq.ResponseURL, messaging.Message{
			Text: "could not encode Slack request",
			Type: messaging.Error})
		return
	}

	h.Log.Infof(ctx, "got text '%s' from '%s'", slackReq.Text, slackReq.Username)

	found := true
	// Get information on the current trainer
	requester, err := loadBasicTrainerData(ctx, h.DB, slackReq.UserID)
	if database.IsNoResults(err) {
		// The trainer could not be found
		found = false
	} else if err != nil {
		// Some error happened while building a trainerData. This should not happen
		h.Log.Errorf(ctx, "while building trainer data: %s", err)
		messaging.Send(client, slackReq.ResponseURL, messaging.Message{
			Text: "could not build trainer data",
			Type: messaging.Error})
		return
	}

	// Set the last known contact URL to the one from this request
	requester.lastContactURL = slackReq.ResponseURL

	if !found {
		// If the trainer doesn't exist, send the request off to the new trainer handler

		h.Log.Infof(ctx, "'%s' is a new trainer", slackReq.Username)
		h.WorkQueue.Add(ctx, NewTrainerURL, slackReqBlob.Bytes())
		return

		// A careful mind might notice that the last contact URL doesn't get
		// saved to the database if the trainer is a new trainer. This is
		// because we don't have a trainer to associate the URL with yet and
		// we never have to send more than one request to a brand new trainer
		// so not having the data isn't an issue.
	}

	// Save the last contact URL for future use
	err = h.DB.SaveLastContactURL(ctx, requester.trainer, requester.lastContactURL)
	if err != nil {
		// Some error has occurred saving the last contact URL. This should not happen
		messaging.Send(client, slackReq.ResponseURL, messaging.Message{
			Text: "could not save the last contact URL for trainer '" + slackReq.Username + "'",
			Type: messaging.Error})
		h.Log.Errorf(ctx, "%s", err)
		return
	}

	switch requester.trainer.GetTrainer().Mode {
	case pkmn.StarterTrainerMode:
		// The trainer is choosing their starter

		h.Log.Infof(ctx, "'%s' is picking their starter", slackReq.Username)
		h.WorkQueue.Add(ctx, ChoosingStarterURL, slackReqBlob.Bytes())
	case pkmn.WaitingTrainerMode:
		// The trainer is in no particular state and is requesting that to change

		switch slackReq.CommandName {
		default:
			// The user doesn't know what to do

			h.Log.Infof(ctx, "'%s' is looking for a list of commands", slackReq.Username)
			h.WorkQueue.Add(ctx, WaitingHelpURL, slackReqBlob.Bytes())
		case "PARTY":
			// The user wants to see their party

			h.Log.Infof(ctx, "'%s' wants to see their party", slackReq.Username)
			h.WorkQueue.Add(ctx, ViewPartyURL, slackReqBlob.Bytes())
		case "BATTLE":
			// The user wants to battle

			h.Log.Infof(ctx, "'%s' is looking for a battle", slackReq.Username)
			h.WorkQueue.Add(ctx, ChallengeURL, slackReqBlob.Bytes())
		case "WILD":
			// The user wants to encounter a wild Pokemon

			h.Log.Infof(ctx, "'%s' is looking to encounter a wild Pokemon", slackReq.Username)
			h.WorkQueue.Add(ctx, WildEncounterURL, slackReqBlob.Bytes())
		}
	case pkmn.BattlingTrainerMode:
		// The trainer is battling or waiting to battle

		// Get the battle the trainer is in
		b, err := h.DB.LoadBattleTrainerIsIn(ctx, requester.trainer.GetTrainer().UUID)
		if err != nil {
			messaging.Send(client, slackReq.ResponseURL, messaging.Message{
				Text: "could not load the battle the trainer is in",
				Type: messaging.Error})
			h.Log.Errorf(ctx, "while trying to find what battle the trainer is in: %s", err)
			return
		}

		switch b.GetBattle().Mode {
		case pkmn.WaitingBattleMode:
			// The trainer is waiting to battle
			switch slackReq.CommandName {
			default:
				// The user doesn't know what to do

				h.Log.Infof(ctx, "'%s' is looking for a list of commands while in waiting battle mode", slackReq.Username)
				h.WorkQueue.Add(ctx, BattleWaitingHelpURL, slackReqBlob.Bytes())
			case "PARTY":
				// The user wants to see their party

				h.Log.Infof(ctx, "'%s' wants to see their party", slackReq.Username)
				h.WorkQueue.Add(ctx, ViewPartyURL, slackReqBlob.Bytes())
			case "FORFEIT":
				// The user wants to stop waiting to battle

				h.Log.Infof(ctx, "'%s' wants to forfeit waiting", slackReq.Username)
				h.WorkQueue.Add(ctx, ForfeitURL, slackReqBlob.Bytes())
			}
		case pkmn.StartedBattleMode:
			// The trainer is currently battling
			switch slackReq.CommandName {
			default:
				// The user doesn't know what to do

				h.Log.Infof(ctx, "'%s' is looking for a list of commands while in started battle mode", slackReq.Username)
				h.WorkQueue.Add(ctx, BattlingHelpURL, slackReqBlob.Bytes())
			case "PARTY":
				// The user wants to see their party

				h.Log.Infof(ctx, "'%s' wants to see their party", slackReq.Username)
				h.WorkQueue.Add(ctx, ViewPartyURL, slackReqBlob.Bytes())
			case "FORFEIT":
				// The user wants to voluntarily lose the match

				h.Log.Infof(ctx, "'%s' wants to forfeit the match", slackReq.Username)
				h.WorkQueue.Add(ctx, ForfeitURL, slackReqBlob.Bytes())
			case "USE":
				// The user wants to use a Pokemon move

				h.Log.Infof(ctx, "'%s' wants to use a move", slackReq.Username)
				h.WorkQueue.Add(ctx, UseMoveURL, slackReqBlob.Bytes())
			case "SWITCH":
				// The user wants to switch Pokemon

				h.Log.Infof(ctx, "'%s' wants to switch Pokemon", slackReq.Username)
				h.WorkQueue.Add(ctx, SwitchPokemonURL, slackReqBlob.Bytes())
			case "CATCH":
				// The user wants to catch the Pokemon

				h.Log.Infof(ctx, "'%s' wants to catch the Pokemon", slackReq.Username)
				h.WorkQueue.Add(ctx, CatchPokemonURL, slackReqBlob.Bytes())
			}
		}
	}
}
