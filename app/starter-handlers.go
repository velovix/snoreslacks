package app

import (
	"bytes"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// fetchStarters returns the list of Pokemon that have the special
// distinguishment of being starters.
func fetchStarters(ctx context.Context, client *http.Client) ([]pokemon, error) {
	// The National Pokedex IDs of all original starters
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
func newTrainerHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Construct a new trainer
	currTrainer.Name = r.username
	currTrainer.pkmn = make([]pokemon, 0)
	currTrainer.Mode = starterTrainerMode // The trainer needs to choose its starter

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not fetch information on starters")
		log.warningf(ctx, "while fetching starters: %s", err)
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
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate starter message template")
		log.errorf(ctx, "while populating starter message template: %s", err)
		return
	}

	err = regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
	if err != nil {
		log.errorf(ctx, "while sending a Slack request: %s", err)
	} else {
		// Save the trainer to the database if the Slack request was successful
		err = db.saveTrainer(ctx, currTrainer)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not save trainer information")
			log.errorf(ctx, "while saving a new trainer: %s", err)
			return
		}

	}
}

// choosingStarterHandler manages requests where the trainer says what starter
// they want. It will give the trainer that starter and allow them to play
// normally.
func choosingStarterHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not fetch information on starters")
		log.warningf(ctx, "while fetching starters: %s", err)
		return
	}

	validStarter := false
	for _, val := range starters {
		if strings.ToUpper(val.Name) == strings.ToUpper(r.text) {
			// Give the trainer the starter
			currTrainer.givePokemon(val)
			currTrainer.Mode = waitingTrainerMode
			validStarter = true
			break
		}
	}

	if !validStarter {
		// The trainer chose a starter that doesn't exist or failed ot choose a
		// starter at all

		templData := &bytes.Buffer{}
		if r.text == "" {
			// The trainer sent an empty request
			err = starterInstructionsTemplate.Execute(templData, nil)
			if err != nil {
				regularSlackRequest(client, currTrainer.LastContactURL, "could not populate starter instructions template")
				log.errorf(ctx, "while populating a starter instructions template: %s", err)
				return
			}
		} else {
			// The trainer requested a starter, but it doesn't exist
			invalidStarterTemplateInfo := strings.ToLower(r.text)
			err = invalidStarterTemplate.Execute(templData, invalidStarterTemplateInfo)
			if err != nil {
				regularSlackRequest(client, currTrainer.LastContactURL, "could not populate invalid starter template")
				log.errorf(ctx, "while populating invalid starter template: %s", err)
				return
			}
		}

		regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
		return
	}

	// The player has chosen their starter

	starterPickedTemplateInfo :=
		struct {
			PkmnName    string
			TrainerName string
		}{
			PkmnName:    strings.ToLower(r.text),
			TrainerName: currTrainer.Name}

	// Populate the template
	templData := &bytes.Buffer{}
	err = starterPickedTemplate.Execute(templData, starterPickedTemplateInfo)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate starter picked template")
		log.errorf(ctx, "while populating starter picked template: %s", err)
		return
	}

	// Save the trainer
	err = db.saveTrainer(ctx, currTrainer)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not save trainer information")
		log.errorf(ctx, "while saving trainer information: %s", err)
		return
	}

	err = regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
	if err != nil {
		log.errorf(ctx, "while sending a Slack request: %s", err)
	} else {
		// Save the trainer if Slack received the request
		err = db.saveTrainer(ctx, currTrainer)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not save trainer information")
			log.errorf(ctx, "while saving trainer information: %s", err)
			return
		}
	}

}
