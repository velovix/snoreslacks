// Package gaedatabase provides an implementation of the database interface for
// the Datastore platform. This package should not be used directly. Instead,
// it should be imported once in your project for its side-effects.
//
// 	import _ "github.com/velovix/snoreslacks/database/gae"
// 	...
//	db, err := database.Get("gae")
package gaedatabase

import "github.com/velovix/snoreslacks/database"

const (
	pokemonKindName           = "Pokemon"
	trainerKindName           = "Trainer"
	battleKindName            = "Battle"
	trainerBattleInfoKindName = "TrainerBattleInfo"
	pokemonBattleInfoKindName = "PokemonBattleInfo"
	lastContactURLKindName    = "LastContactURL"
)

// GAEDatabase is the datastore implementation of the database interface.
type GAEDatabase struct{}

func init() {
	database.Register("gae", GAEDatabase{})
}
