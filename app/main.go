package app

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// mainHandler responds to Slack slash requests. Generally, it will delegate
// work to more specific handlers.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	// Get an App Engine context for this request
	ctx := appengine.NewContext(r)

	// Parse the POST form data
	r.ParseForm()

	// Check if the request is coming from slack
	_, ok := r.Form["token"]
	if !ok {
		http.Error(w, "no token included in request body", 400)
		return
	}
	token := r.Form["token"][0]
	if token != config.Token { // Compare the tokens
		http.Error(w, "invalid token", 400)
		return
	}

	// Extract the text for later use
	_, ok = r.Form["text"]
	if !ok {
		http.Error(w, "no text included in request body", 400)
		return
	}
	cmd := r.Form["text"][0]
	cmd = strings.ToUpper(strings.Trim(cmd, " "))

	// Extract the username
	_, ok = r.Form["user_name"]
	if !ok {
		http.Error(w, "no username included in request body", 400)
		return
	}
	username := r.Form["user_name"][0]

	log.Infof(ctx, "got text '%s' from '%s'", cmd, username)

	// Read in the trainer data
	currTrainer, err := loadTrainer(ctx, username)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			// If the trainer doesn't exist, send the request off to the new trainer handler

			log.Infof(ctx, "'%s' is a new trainer", username)
			newTrainerHandler(w, r, ctx, currTrainer)
			return
		} else {
			// Some error has occurred reading the trainer data. This should not happen
			http.Error(w, "could not pull in information for trainer '"+username+"'", 500)
			log.Errorf(ctx, "while pulling trainer information: %s", err)
			return
		}
	}

	switch currTrainer.Mode {
	case starterTrainerMode:
		// The trainer is choosing their starter

		log.Infof(ctx, "'%s' is picking their starter", username)
		choosingStarterHandler(w, r, ctx, currTrainer)
	case waitingTrainerMode:
		// The trainer is in no particular state and is requesting that to change

		switch cmd {
		case "":
			// The user doesn't know what to do

			waitingHelpHandler(w, r, ctx, currTrainer)
		case "PARTY":
			// The user wants to see their party

			log.Infof(ctx, "'%s' wants to see their party", username)
			viewPartyHandler(w, r, ctx, currTrainer)
		}
	}
}

func init() {
	http.HandleFunc("/", mainHandler)
}
