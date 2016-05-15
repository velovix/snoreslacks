package handlers

import (
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"

	"golang.org/x/net/context"
)

// waitingHelpHandler sends help information to the user when they are not
// doing anything in particular.
func waitingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Send the templated info
	err := regularSlackTemplRequest(client, currTrainer.lastContactURL, waitingHelpTemplate, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting help template")
		log.Errorf(ctx, "while populating waiting help template: %s", err)
		return
	}
}

// battleWaitingHelpHandler sends help information to the user when they are
// waiting for a battle to start
func battleWaitingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Send the templated info
	err := regularSlackTemplRequest(client, currTrainer.lastContactURL, battleWaitingHelpTemplate, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battle waiting help template")
		log.Errorf(ctx, "while populating battle waiting help template: %s", err)
		return
	}
}

// battlingHelpHandler sends help information to the user when they are in a
// battle.
func battlingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Send the templated info
	err := regularSlackTemplRequest(client, currTrainer.lastContactURL, battlingHelpTemplate, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battling help template")
		log.Errorf(ctx, "while populating battling help template: %s", err)
	}
}
