package app

import (
	"bytes"
	"net/http"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
)

func waitingHelpHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, currTrainer trainer) {
	// Populate the template
	templData := &bytes.Buffer{}
	err := waitingHelpTemplate.Execute(templData, waitingHelpTemplate)
	if err != nil {
		http.Error(w, "could not populate waiting help template", 500)
		log.Errorf(ctx, "while populating waiting help template: %s", err)
		return
	}

	regularSlackResponse(w, r, string(templData.Bytes()))
}
