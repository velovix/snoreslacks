package snoreslacks

import (
	"bytes"
	"log"
	"net/http"
	"text/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// newTrainerHandler manages requests made by new trainers. It will create the
// trainer data for this Slack user and respond with information about choosing
// their starter.
func newTrainerHandler(w http.ResponseWriter, r *http.Request, currTrainer trainer) {
	// Construct a new trainer
	currTrainer.name = r.Form["user_name"][0] // We can assume the request is a valid Slack slash request at this point
	currTrainer.pkmn = make([]pokemon, 0)
	currTrainer.mode = starterTrainerMode // The trainer needs to choose its starter

	// Create the data that will go in the starter message template
	starterMessageTemplateInfo := struct {
		Username string
		Starters []pokemon
	}{
		Username: currTrainer.name,
		Starters: []pokemon{pokemon{}, pokemon{}, pokemon{}, pokemon{}}}

	// Populate the template
	templData := &bytes.Buffer{}
	templ, err := template.New("starters").Parse(starterMessageTemplate)
	if err != nil {
		http.Error(w, "could not compile starter message template", 500)
		log.Println("while parsing starter message template:", err)
		return
	}
	err = templ.Execute(templData, starterMessageTemplateInfo)
	if err != nil {
		http.Error(w, "could not populate starter message template", 500)
		log.Println("while populating starter message template:", err)
		return
	}

	regularSlackResponse(w, r, string(templData.Bytes()))
}

// mainHandler responds to Slack slash requests. Generally, it will delegate
// work to more specific handlers.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	// Get an App Engine context for this request
	ctx := appengine.NewContext(r)

	// Parse the POST form data
	r.ParseForm()

	// Check if the request is coming from slack
	/*_, ok := r.Form["token"]
	if !ok {
		http.Error(w, "no token included in request body", 400)
		return
	}
	token := r.Form["token"][0]
	expectedToken, err := getTokenFromConfig(ctx) // Read the expected token from the database
	if err != nil {
		http.Error(w, "request authentication has not been set up properly", 500)
		log.Println("while pulling expected token from config:", err)
		return
	}
	if token != expectedToken { // Compare the tokens
		http.Error(w, "invalid token", 400)
		return
	}*/

	// Extract the username
	_, ok := r.Form["user_name"]
	if !ok {
		http.Error(w, "no username included in request body", 400)
		return
	}
	username := r.Form["user_name"][0]

	// Create the key identifiying the current trainer
	trainerKey := datastore.NewKey(ctx, "trainer", username, 0, nil)
	currTrainer := trainer{}

	// Read in the trainer data
	err := datastore.Get(ctx, trainerKey, &currTrainer)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			// If the trainer doesn't exist, send the request off to the new trainer handler
			newTrainerHandler(w, r, currTrainer)
			return
		} else {
			// Some error has occurred reading the trainer data. This should not happen
			http.Error(w, "could not pull in information for trainer '"+username+"'", 500)
			log.Println("while pulling trainer information:", err)
			return
		}
	}
}

func init() {
	log.Println("Starting Snoreslacks server...")

	http.HandleFunc("/", mainHandler)
}
