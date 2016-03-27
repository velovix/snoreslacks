package app

import (
	"bytes"
	"net/http"

	"golang.org/x/net/context"
)

func waitingHelpHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := waitingHelpTemplate.Execute(templData, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate waiting help template")
		log.errorf(ctx, "while populating waiting help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
}

func battleWaitingHelpHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := battleWaitingHelpTemplate.Execute(templData, nil)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate battle waiting help template")
		log.errorf(ctx, "while populating battle waiting help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
}

func battlingHelpHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := battlingHelpTemplate.Execute(templData, "")
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate battling help template")
		log.errorf(ctx, "while populating battling help template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
}
