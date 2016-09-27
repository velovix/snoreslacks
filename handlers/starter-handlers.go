package handlers

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"

	"golang.org/x/net/context"
)

// starterLevel is the level that starter Pokemon should be at when chosen.
const starterLevel = 5

// fetchStarters returns the list of Pokemon that have the special
// distinguishment of being starters.
func fetchStarters(ctx context.Context, client messaging.Client, fetcher pokeapi.Fetcher) ([]pkmn.Pokemon, error) {
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
		pkmn[i], err = pokeapi.NewPokemon(ctx, client, fetcher, apiPkmn, starterLevel)
		if err != nil {
			return nil, errors.New("while creating the Pokemon: " + err.Error())
		}
	}

	return pkmn, nil
}

// NewTrainer manages requests made by new trainers. It will create the trainer
// data for this Slack user and respond with information about choosing their
// starter.
type NewTrainer struct {
}

func (h *NewTrainer) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)

	// Assert that no such trainer exists yet
	_, err := s.DB.LoadTrainer(ctx, slackReq.UserID)
	if err == nil {
		// The trainer already exists, but we don't want to error out at the
		// user because users don't explitly ask to become a trainer.
		s.Log.Warningf(ctx, "attempt to create a new trainer when the trainer already exists")
		return nil
	} else if err != nil && !database.IsNoResults(err) {
		// A miscellaneous error occurred
		return handlerError{user: "could not check if trainer exists",
			err: errors.Wrap(err, "while checking if trainer exists")}
	}

	// Construct a new trainer
	requester := s.DB.NewTrainer(pkmn.Trainer{
		UUID:                 slackReq.UserID, // Use Slack's user IDs as a unique identifier
		Name:                 slackReq.Username,
		Mode:                 pkmn.StarterTrainerMode, // The trainer needs to choose its starter
		Type:                 pkmn.HumanTrainerType,   // The trainer is a human being, not an NPC
		KantoEncounterLevel:  1,                       // Grant the user access to level 1 encounters
		JohtoEncounterLevel:  1,
		HoennEncounterLevel:  1,
		SinnohEncounterLevel: 1,
		UnovaEncounterLevel:  1,
		KalosEncounterLevel:  1,
	})

	s.Log.Infof(ctx, "created a new trainer: %+v", *requester.GetTrainer())

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client, s.Fetcher)
	if err != nil {
		return handlerError{user: "could not fetch information on starters", err: err}
	}

	// Create the data that will go in the starter message template
	starterMessageTemplateInfo := struct {
		Username     string
		Starters     []pkmn.Pokemon
		SlashCommand string
	}{
		Username:     requester.GetTrainer().Name,
		Starters:     starters,
		SlashCommand: slackReq.SlashCommand}

	err = messaging.SendTempl(client, slackReq.ResponseURL, messaging.TemplMessage{
		Templ:     starterMessageTemplate,
		TemplInfo: starterMessageTemplateInfo})
	if err != nil {
		return handlerError{user: "could not populate starter message template", err: err}
	}

	// Save the trainer to the database if all else was successful
	err = s.DB.SaveTrainer(ctx, requester)
	if err != nil {
		return handlerError{user: "could not save trainer information", err: err}
	}

	return nil
}

// ChoosingStarter manages requests where the trainer says what starter they
// want. It will give the trainer that starter and allow them to play normally.
type ChoosingStarter struct {
}

func (h *ChoosingStarter) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Assert that the trainer does not have any Pokemon yet
	if len(requester.pkmn) != 0 {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     choosingStarterWhenInWrongModeTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not sned choosing starter when in wrong mode template", err: err}
		}
		return nil // No more work to do
	}

	// Fetch information on the starters
	starters, err := fetchStarters(ctx, client, s.Fetcher)
	if err != nil {
		return handlerError{user: "could not fetch information on starters", err: err}
	}

	// Check if the user chose a valid starter
	validStarter := false
	for _, val := range starters {
		if strings.ToUpper(val.Name) == strings.ToUpper(slackReq.Text) {
			// Give the trainer the starter
			var success bool
			requester.pkmn, success = givePokemon(requester.pkmn, s.DB.NewPokemon(val))
			if !success {
				// This contingency should never happen and is a sign of
				// something seriously wrong
				return handlerError{user: "starting trainer already has the maximum amount of Pokemon", err: err}
			}
			requester.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
			validStarter = true
			break
		}
	}

	if !validStarter {
		// The trainer chose a starter that doesn't exist or failed to choose a
		// starter at all

		if slackReq.Text == "" {
			// The trainer sent an empty request
			err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
				Templ:     starterInstructionsTemplate,
				TemplInfo: nil})
			if err != nil {
				return handlerError{user: "could not populate starter instructions template", err: err}
			}
		} else {
			// The trainer requested a starter, but it doesn't exist
			invalidStarterTemplateInfo := strings.ToLower(slackReq.Text)
			err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
				Templ:     invalidStarterTemplate,
				TemplInfo: invalidStarterTemplateInfo})
			if err != nil {
				return handlerError{user: "could not populate invalid starter template", err: err}
			}
		}
		return nil // There is no more work to be done
	}

	// The player has chosen their starter

	starterPickedTemplateInfo :=
		struct {
			PkmnName    string
			TrainerName string
			CommandName string
		}{
			PkmnName:    strings.ToLower(slackReq.Text),
			TrainerName: requester.trainer.GetTrainer().Name,
			CommandName: slackReq.SlashCommand}

	// Populate the template
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     starterPickedTemplate,
		TemplInfo: starterPickedTemplateInfo})
	if err != nil {
		return handlerError{user: "could not populate starter picked template", err: err}
	}

	// Save the trainer and party if Slack received the request
	err = s.DB.SaveTrainer(ctx, requester.trainer)
	if err != nil {
		return handlerError{user: "could not save trainer information", err: err}
	}
	err = s.DB.SaveParty(ctx, requester.trainer, requester.pkmn)
	if err != nil {
		return handlerError{user: "could not save party information", err: err}
	}

	return nil
}
