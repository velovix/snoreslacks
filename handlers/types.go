package handlers

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
)

// trainerData contains trainer info retrieved from the database.
type trainerData struct {
	database.Trainer

	lastContactURL string
	pkmn           []database.Pokemon
}

// buildTrainerData loads information the given trainer name and fills in a
// trainerData object.
func buildTrainerData(ctx context.Context, db database.Database, name string) (trainerData, bool, error) {
	var td trainerData
	var found bool
	var err error

	// Read in the trainer data
	td.Trainer, found, err = db.LoadTrainer(ctx, name)
	if err != nil {
		// Some error has occurred reading the trainer data. This should not happen
		return trainerData{}, false, err
	}
	if !found {
		// The trainer does not exist
		td.Trainer = db.NewTrainer(pkmn.Trainer{})
		return td, false, nil
	}

	// Load the trainer's party
	td.pkmn, _, err = db.LoadParty(ctx, td.Trainer)
	if err != nil {
		// Some error has occurred loading the trainer's party. This should not happen
		return trainerData{}, false, err
	}

	// Load the last contact URL if it exists
	td.lastContactURL, found, err = db.LoadLastContactURL(ctx, td.Trainer)
	if err != nil {
		// Some error has occurred loading the trainer's last contact URL. This should not happen
		return trainerData{}, false, err
	}

	return td, true, nil
}

func givePokemon(party []database.Pokemon, pkmn database.Pokemon) ([]database.Pokemon, bool) {
	if len(party) >= 6 {
		return party, false
	}

	return append(party, pkmn), true
}
