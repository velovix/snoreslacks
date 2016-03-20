package app

import (
	"bytes"
	"net/http"
	"strings"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"golang.org/x/net/context"
)

func fetchStarters(ctx context.Context, client *http.Client) ([]pokemon, error) {
	// The National Pokedex IDs of all starters up to Generation 6
	starterIDs := []int{1, 4, 7}

	pkmn := make([]pokemon, len(starterIDs))
	for i, val := range starterIDs {
		// Fetch the PokeAPI data
		apiPkmn, err := fetchPokemon(val, client, ctx)
		if err != nil {
			return nil, err
		}
		// Create the Pokemon from that data
		pkmn[i], err = newPokemon(apiPkmn, client, ctx)
		if err != nil {
			return nil, err
		}
	}

	return pkmn, nil
}

// newTrainerHandler manages requests made by new trainers. It will create the
// trainer data for this Slack user and respond with information about choosing
// their starter.
func newTrainerHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, currTrainer trainer) {
	// Construct a new trainer
	currTrainer.Name = r.Form["user_name"][0] // We can assume the request is a valid Slack slash request at this point
	currTrainer.pkmn = make([]pokemon, 0)
	currTrainer.Mode = starterTrainerMode // The trainer needs to choose its starter

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, urlfetch.Client(ctx))
	if err != nil {
		http.Error(w, "could not fetch information on starters", 500)
		log.Warningf(ctx, "while fetching starters: %s", err)
		return
	}

	// Create the data that will go in the starter message template
	starterMessageTemplateInfo := struct {
		Username string
		Starters []pokemon
	}{
		Username: currTrainer.Name,
		Starters: starters}

	// Populate the template
	templData := &bytes.Buffer{}
	err = starterMessageTemplate.Execute(templData, starterMessageTemplateInfo)
	if err != nil {
		http.Error(w, "could not populate starter message template", 500)
		log.Errorf(ctx, "while populating starter message template: %s", err)
		return
	}

	// Save the trainer to the database
	err = saveTrainer(ctx, currTrainer)
	if err != nil {
		http.Error(w, "could not save trainer information", 500)
		log.Errorf(ctx, "while saving a new trainer: %s", err)
		return
	}

	regularSlackResponse(w, r, string(templData.Bytes()))
}

func choosingStarterHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, currTrainer trainer) {
	// Fetch information on the starters
	starters, err := fetchStarters(ctx, urlfetch.Client(ctx))
	if err != nil {
		http.Error(w, "could not fetch information on starters", 500)
		log.Warningf(ctx, "while fetching starters: %s", err)
		return
	}

	cmd := strings.Trim(r.PostFormValue("text"), " ")

	validStarter := false
	for _, val := range starters {
		if strings.ToUpper(val.Name) == strings.ToUpper(cmd) {
			// Give the trainer the starter
			currTrainer.pkmn = append(currTrainer.pkmn, val)
			currTrainer.Mode = waitingTrainerMode
			validStarter = true
			break
		}
	}

	if !validStarter {
		// The trainer chose a starter that doesn't exist or failed ot choose a
		// starter at all

		templData := &bytes.Buffer{}
		if cmd == "" {
			// The trainer sent an empty request
			err = starterInstructionsTemplate.Execute(templData, nil)
			if err != nil {
				http.Error(w, "could not populate starter instructions template", 500)
				log.Errorf(ctx, "while populating a starter instructions template: %s", err)
				return
			}
		} else {
			// The trainer requested a starter, but it doesn't exist
			invalidStarterTemplateInfo := strings.ToLower(cmd)
			err = invalidStarterTemplate.Execute(templData, invalidStarterTemplateInfo)
			if err != nil {
				http.Error(w, "could not populate invalid starter template", 500)
				log.Errorf(ctx, "while populating invalid starter template: %s", err)
				return
			}
		}

		regularSlackResponse(w, r, string(templData.Bytes()))
		return
	}

	// The player has chosen their starter

	starterPickedTemplateInfo :=
		struct {
			PkmnName    string
			TrainerName string
		}{
			PkmnName:    strings.ToLower(cmd),
			TrainerName: currTrainer.Name}

	// Populate the template
	templData := &bytes.Buffer{}
	err = starterPickedTemplate.Execute(templData, starterPickedTemplateInfo)
	if err != nil {
		http.Error(w, "could not populate starter picked template", 500)
		log.Errorf(ctx, "while populating starter picked template: %s", err)
		return
	}

	// Save the trainer
	err = saveTrainer(ctx, currTrainer)
	if err != nil {
		http.Error(w, "could not save trainer information", 500)
		log.Errorf(ctx, "while saving trainer information: %s", err)
		return
	}

	regularSlackResponse(w, r, string(templData.Bytes()))
}
