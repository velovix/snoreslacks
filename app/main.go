package app

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

// mainHandler responds to Slack slash requests. Generally, it will delegate
// work to more specific handlers.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	// Create the context and "inject" database and logging dependencies
	ctx := appengine.NewContext(r)
	db := appengineDatastoreDAO{}
	log := appengineLogger{}
	client := urlfetch.Client(ctx)

	// Create the Slack request
	slackReq, err := newSlackRequest(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if slackReq.token != config.Token { // Compare the tokens
		http.Error(w, "invalid token", 400)
		return
	}

	log.infof(ctx, "got text '%s' from '%s'", slackReq.text, slackReq.username)

	// Read in the trainer data
	currTrainer, found, err := db.loadTrainer(ctx, slackReq.username)
	if err != nil {
		// Some error has occurred reading the trainer data. This should not happen
		http.Error(w, "could not pull in information for trainer '"+slackReq.username+"'", 500)
		log.errorf(ctx, "while pulling trainer information: %s", err)
		return
	}
	if !found {
		// If the trainer doesn't exist, send the request off to the new trainer handler

		log.infof(ctx, "'%s' is a new trainer", slackReq.username)
		newTrainerHandler(ctx, db, log, client, slackReq, currTrainer)
		return
	}

	// Set the last known contact URL to the one from this request
	currTrainer.LastContactURL = slackReq.responseURL

	switch currTrainer.Mode {
	case starterTrainerMode:
		// The trainer is choosing their starter

		log.infof(ctx, "'%s' is picking their starter", slackReq.username)
		choosingStarterHandler(ctx, db, log, client, slackReq, currTrainer)
	case waitingTrainerMode:
		// The trainer is in no particular state and is requesting that to change

		switch slackReq.commandName {
		case "":
			// The user doesn't know what to do

			waitingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
		case "PARTY":
			// The user wants to see their party

			log.infof(ctx, "'%s' wants to see their party", slackReq.username)
			viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
		case "BATTLE":
			// The user wants to battle

			log.infof(ctx, "'%s' is looking for a battle", slackReq.username)
			challengeHandler(ctx, db, log, client, slackReq, currTrainer)
		}
	case battlingTrainerMode:
		// The trainer is battling or waiting to battle

		// Get the battle the trainer is in
		b, exists, err := db.loadBattleTrainerIsIn(ctx, currTrainer.Name)
		if err != nil {
			http.Error(w, "could not load the battle the trainer is in", 500)
			log.errorf(ctx, "while trying to find what battle the trainer is in: %s", err)
			return
		}
		if !exists {
			http.Error(w, "trainer is in battling mode, but is not in a battle", 500)
			log.errorf(ctx, "while trying to find what battle the trainer is in: %s", err)
			return
		}

		switch b.Mode {
		case waitingBattleMode:
			// The trainer is waiting to battle
			switch slackReq.commandName {
			case "":
				// The user doesn't know what to do

				battleWaitingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
			case "PARTY":
				// The user wants to see their party

				log.infof(ctx, "'%s' wants to see their party", slackReq.username)
				viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
			case "FORFEIT":
				// The user wants to stop waiting to battle

				log.infof(ctx, "'%s' wants to forfeit waiting", slackReq.username)
				forfeitHandler(ctx, db, log, client, slackReq, currTrainer)
			}
		case startedBattleMode:
			// The trainer is currently battling
			switch slackReq.commandName {
			case "":
				// The user doesn't know what to do

				battlingHelpHandler(ctx, db, log, client, slackReq, currTrainer)
			case "PARTY":
				// The user wants to see their party

				log.infof(ctx, "'%s' wants to see their party", slackReq.username)
				viewPartyHandler(ctx, db, log, client, slackReq, currTrainer)
			case "FORFEIT":
				// The user wants to voluntarily lose the match

				log.infof(ctx, "'%s' wants to forfeit the match", slackReq.username)
				forfeitHandler(ctx, db, log, client, slackReq, currTrainer)
			case "USE":
				// The user wants to use a Pokemon move

				log.infof(ctx, "'%s' wants to use a move", slackReq.username)
				useMoveHandler(ctx, db, log, client, slackReq, currTrainer)
			case "SWITCH":
				// The user wants to switch Pokemon

				log.infof(ctx, "'%s' wants to switch Pokemon", slackReq.username)
				switchPokemonHandler(ctx, db, log, client, slackReq, currTrainer)
			}
		}
	}
}

func init() {
	http.HandleFunc("/", mainHandler)
}
