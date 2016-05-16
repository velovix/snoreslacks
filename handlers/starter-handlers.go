package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"

	"golang.org/x/net/context"
)

// fetchStarters returns the list of Pokemon that have the special
// distinguishment of being starters.
func fetchStarters(ctx context.Context, client *http.Client, fetcher pokeapi.Fetcher) ([]pkmn.Pokemon, error) {
	// The National Pokedex IDs of all original starters
	starterIDs := []int{1, 4, 7}

	pkmn := make([]pkmn.Pokemon, len(starterIDs))
	for i, val := range starterIDs {
		// Fetch the PokeAPI data
		apiPkmn, err := fetcher.FetchPokemon(ctx, client, val)
		if err != nil {
			return nil, errors.New("while fetching data: " + err.Error())
		}
		// Create the Pokemon from that data
		pkmn[i], err = pokeapi.NewPokemon(ctx, client, fetcher, apiPkmn)
		if err != nil {
			return nil, errors.New("while creating the Pokemon: " + err.Error())
		}
	}

	return pkmn, nil
}

// newTrainerHandler manages requests made by new trainers. It will create the
// trainer data for this Slack user and respond with information about choosing
// their starter.
func newTrainerHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher, currTrainer trainerData) {

	// Construct a new trainer
	currTrainer.GetTrainer().Name = r.username
	currTrainer.GetTrainer().Mode = pkmn.StarterTrainerMode // The trainer needs to choose its starter

	log.Infof(ctx, "created a new trainer: %+v", *currTrainer.GetTrainer())

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client, fetcher)
	if err != nil {
		sendMessage(client, currTrainer.lastContactURL, message{
			text: "could not fetch information on starters",
			t:    errorMsgType})
		log.Errorf(ctx, "while fetching starters: %s", err)
		return
	}

	// Create the data that will go in the starter message template
	starterMessageTemplateInfo := struct {
		Username string
		Starters []pkmn.Pokemon
	}{
		Username: currTrainer.GetTrainer().Name,
		Starters: starters}

	err = sendTemplMessage(client, currTrainer.lastContactURL, templMessage{
		templ:     starterMessageTemplate,
		templInfo: starterMessageTemplateInfo})
	if err != nil {
		// The trainer did not get our response. Abort operation
		sendMessage(client, currTrainer.lastContactURL, message{
			text: "could not populate starter message template",
			t:    errorMsgType})
		log.Errorf(ctx, "while sending starter message template: %s", err)
		return
	} else {
		// Save the trainer to the database if the Slack request was successful
		err = db.SaveTrainer(ctx, currTrainer.Trainer)
		if err != nil {
			sendMessage(client, currTrainer.lastContactURL, message{
				text: "could not save trainer information",
				t:    errorMsgType})
			log.Errorf(ctx, "while saving a new trainer: %s", err)
			return
		}
	}
}

// choosingStarterHandler manages requests where the trainer says what starter
// they want. It will give the trainer that starter and allow them to play
// normally.
func choosingStarterHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher, currTrainer trainerData) {

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client, fetcher)
	if err != nil {
		sendMessage(client, currTrainer.lastContactURL, message{
			text: "could not fetch information on starters",
			t:    errorMsgType})
		log.Errorf(ctx, "while fetching starters: %s", err)
		return
	}

	// Check if the user chose a valid starter
	validStarter := false
	for _, val := range starters {
		if strings.ToUpper(val.Name) == strings.ToUpper(r.text) {
			// Give the trainer the starter
			var success bool
			currTrainer.pkmn, success = givePokemon(currTrainer.pkmn, db.NewPokemon(val))
			if !success {
				// This contingency should never happen and is a sign of
				// something seriously wrong
				sendMessage(client, currTrainer.lastContactURL, message{
					text: "starting trainer already has the maximum amount of Pokemon",
					t:    errorMsgType})
				log.Errorf(ctx, "%s", errors.New("a new trainer already has a full party of Pokemon"))
				return
			}
			currTrainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
			validStarter = true
			break
		}
	}

	if !validStarter {
		// The trainer chose a starter that doesn't exist or failed ot choose a
		// starter at all

		if r.text == "" {
			// The trainer sent an empty request
			err = sendTemplMessage(client, currTrainer.lastContactURL, templMessage{
				templ:     starterInstructionsTemplate,
				templInfo: nil})
			if err != nil {
				sendMessage(client, currTrainer.lastContactURL, message{
					text: "could not populate starter instructions template",
					t:    errorMsgType})
				log.Errorf(ctx, "while sending a starter instructions template: %s", err)
				return
			}
		} else {
			// The trainer requested a starter, but it doesn't exist
			invalidStarterTemplateInfo := strings.ToLower(r.text)
			err = sendTemplMessage(client, currTrainer.lastContactURL, templMessage{
				templ:     invalidStarterTemplate,
				templInfo: invalidStarterTemplateInfo})
			if err != nil {
				sendMessage(client, currTrainer.lastContactURL, message{
					text: "could not populate invalid starter template",
					t:    errorMsgType})
				log.Errorf(ctx, "while sending invalid starter template: %s", err)
				return
			}
		}
		return
	}

	// The player has chosen their starter

	starterPickedTemplateInfo :=
		struct {
			PkmnName    string
			TrainerName string
		}{
			PkmnName:    strings.ToLower(r.text),
			TrainerName: currTrainer.GetTrainer().Name}

	// Populate the template
	err = sendTemplMessage(client, currTrainer.lastContactURL, templMessage{
		templ:     starterPickedTemplate,
		templInfo: starterPickedTemplateInfo})
	if err != nil {
		sendMessage(client, currTrainer.lastContactURL, message{
			text: "could not populate starter picked template",
			t:    errorMsgType})
		log.Errorf(ctx, "while sending starter picked template: %s", err)
		return
	} else {
		// Save the trainer and party if Slack received the request
		err = db.SaveTrainer(ctx, currTrainer.Trainer)
		if err != nil {
			sendMessage(client, currTrainer.lastContactURL, message{
				text: "could not save trainer information",
				t:    errorMsgType})
			log.Errorf(ctx, "while saving trainer information: %s", err)
			return
		}
		err = db.SaveParty(ctx, currTrainer.Trainer, currTrainer.pkmn)
		if err != nil {
			sendMessage(client, currTrainer.lastContactURL, message{
				text: "could not save party information",
				t:    errorMsgType})
			log.Errorf(ctx, "%s", err)
		}
	}

}
