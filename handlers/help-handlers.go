package handlers

import (
	"bytes"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"

	"golang.org/x/net/context"
)

func waitingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := waitingHelpTemplate.Execute(templData, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting help template")
		log.Errorf(ctx, "while populating waiting help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
}

func battleWaitingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := battleWaitingHelpTemplate.Execute(templData, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battle waiting help template")
		log.Errorf(ctx, "while populating battle waiting help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
}

func battlingHelpHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := battlingHelpTemplate.Execute(templData, "")
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battling help template")
		log.Errorf(ctx, "while populating battling help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
}
