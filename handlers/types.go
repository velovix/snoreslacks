package handlers

import "github.com/velovix/snoreslacks/database"

// givePokemon adds a Pokemon to the given party, or returns false if the
// player already has the maximum amount of Pokemon.
func givePokemon(party []database.Pokemon, pkmn database.Pokemon) ([]database.Pokemon, bool) {
	if len(party) >= 6 {
		return party, false
	}

	return append(party, pkmn), true
}
