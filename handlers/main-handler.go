package handlers

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
)

// MainHandler responds to Slack slash requests.
func MainHandler(ctx context.Context, w http.ResponseWriter, r *http.Request,
	db database.Database, log logging.Logger, client *http.Client,
	fetcher pokeapi.Fetcher, token string) {

	var err error

	// Create the Slack request
	slackReq, err := newSlackRequest(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if slackReq.token != token { // Compare the tokens
		http.Error(w, "invalid token", 400)
		return
	}

	log.Infof(ctx, "got text '%s' from '%s'", slackReq.text, slackReq.username)

	startTime := time.Now()

	done := make(chan struct{}, 1)    // Signals when we're done processing this request
	timeout := make(chan struct{}, 1) // Signals when we've been taking a long time processing this request

	// Sends a signal after five seconds indicating that it has taken a while
	// for the server to process this request and we need to assure the trainer
	// that the request was received.
	go func() {
		time.Sleep(1 * time.Second)
		select {
		case <-done:
			return // Don't send the timeout if the request has finished
		default:
			timeout <- struct{}{}
		}
	}()

	// Sends a message assuring the trainer that the request is being processed
	// if the server is taking a long time to complete a task.
	go func(r slackRequest) {
		select {
		case <-done:
			return // The request has finished processing
		case <-timeout:
			// Time to send the message
		}

		// Send the templated message
		err = sendTemplMessage(client, r.responseURL, templMessage{
			templ:     initialResponseTemplate,
			templInfo: nil})
		if err != nil {
			sendMessage(client, r.responseURL, message{
				text: "could not populate initial response template",
				t:    errorMsgType})
			log.Errorf(ctx, "while sending the initial response template: %s", err)
			return
		}
	}(slackReq)

	// Notify the slow response message after we finish processing
	// the request
	defer func() {
		log.Infof(ctx, "finished processing request for '%s' (processing took %s)", slackReq.username, time.Since(startTime))
		done <- struct{}{}
	}()

	// Get information on the current trainer
	currTrainer, found, err := buildTrainerData(ctx, db, slackReq.username)
	if err != nil {
		// Some error happened while building a trainerData. This should not happen
		log.Errorf(ctx, "while building trainer data: %s", err)
		sendMessage(client, slackReq.responseURL, message{
			text: "could not build trainer data",
			t:    errorMsgType})
	}

	// Set the last known contact URL to the one from this request
	currTrainer.lastContactURL = slackReq.responseURL

	if !found {
		// If the trainer doesn't exist, send the request off to the new trainer handler

		log.Infof(ctx, "'%s' is a new trainer", slackReq.username)
		newTrainerHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
		return

		// A careful mind might notice that the last contact URL doesn't get
		// saved to the database if the trainer is a new trainer. This is
		// because we don't have a trainer to associate the URL with yet and
		// we never have to send more than one request to a brand new trainer
		// so not having the data isn't an issue.
	}

	// Save the last contact URL for future use
	err = db.SaveLastContactURL(ctx, currTrainer.Trainer, currTrainer.lastContactURL)
	if err != nil {
		// Some error has occurred saving the last contact URL. This should not happen
		sendMessage(client, slackReq.responseURL, message{
			text: "could not save the last contact URL for trainer '" + slackReq.username + "'",
			t:    errorMsgType})
		log.Errorf(ctx, "%s", err)
		return
	}

	switch currTrainer.GetTrainer().Mode {
	case pkmn.StarterTrainerMode:
		// The trainer is choosing their starter

		log.Infof(ctx, "'%s' is picking their starter", slackReq.username)
		choosingStarterHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
	case pkmn.WaitingTrainerMode:
		// The trainer is in no particular state and is requesting that to change

		switch slackReq.commandName {
		case "":
			// The user doesn't know what to do

			log.Infof(ctx, "'%s' is looking for a list of commands", slackReq.username)
			waitingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
		case "PARTY":
			// The user wants to see their party

			log.Infof(ctx, "'%s' wants to see their party", slackReq.username)
			viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
		case "BATTLE":
			// The user wants to battle

			log.Infof(ctx, "'%s' is looking for a battle", slackReq.username)
			challengeHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
		}
	case pkmn.BattlingTrainerMode:
		// The trainer is battling or waiting to battle

		// Get the battle the trainer is in
		b, exists, err := db.LoadBattleTrainerIsIn(ctx, currTrainer.GetTrainer().Name)
		if err != nil {
			sendMessage(client, slackReq.responseURL, message{
				text: "could not load the battle the trainer is in",
				t:    errorMsgType})
			log.Errorf(ctx, "while trying to find what battle the trainer is in: %s", err)
			return
		}
		if !exists {
			sendMessage(client, slackReq.responseURL, message{
				text: "trainer is in battling mode, but is not in a battle",
				t:    errorMsgType})
			log.Errorf(ctx, "trainer is in battling mode, but is not in a battle")
			return
		}

		switch b.GetBattle().Mode {
		case pkmn.WaitingBattleMode:
			// The trainer is waiting to battle
			switch slackReq.commandName {
			case "":
				// The user doesn't know what to do

				log.Infof(ctx, "'%s' is looking for a list of commands while in waiting battle mode", slackReq.username)
				battleWaitingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
			case "PARTY":
				// The user wants to see their party

				log.Infof(ctx, "'%s' wants to see their party", slackReq.username)
				viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
			case "FORFEIT":
				// The user wants to stop waiting to battle

				log.Infof(ctx, "'%s' wants to forfeit waiting", slackReq.username)
				forfeitHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
			}
		case pkmn.StartedBattleMode:
			// The trainer is currently battling
			switch slackReq.commandName {
			case "":
				// The user doesn't know what to do

				log.Infof(ctx, "'%s' is looking for a list of commands while in started battle mode", slackReq.username)
				battlingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
			case "PARTY":
				// The user wants to see their party

				log.Infof(ctx, "'%s' wants to see their party", slackReq.username)
				viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
			case "FORFEIT":
				// The user wants to voluntarily lose the match

				log.Infof(ctx, "'%s' wants to forfeit the match", slackReq.username)
				forfeitHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
			case "USE":
				// The user wants to use a Pokemon move

				log.Infof(ctx, "'%s' wants to use a move", slackReq.username)
				useMoveHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
			case "SWITCH":
				// The user wants to switch Pokemon

				log.Infof(ctx, "'%s' wants to switch Pokemon", slackReq.username)
				switchPokemonHandler(ctx, db, log, client, slackReq, fetcher, currTrainer)
			}
		}
	}
}
