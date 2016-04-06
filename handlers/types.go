package handlers

import "github.com/velovix/snoreslacks/database"

// trainerData contains trainer info retrieved from the database.
type trainerData struct {
	database.Trainer

	lastContactURL string
	pkmn           []database.Pokemon
}

func givePokemon(party []database.Pokemon, pkmn database.Pokemon) ([]database.Pokemon, bool) {
	if len(party) >= 6 {
		return party, false
	}

	return append(party, pkmn), true
}
