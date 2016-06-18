// Package gaedatabase provides an implementation of the database interface for
// the Datastore platform. This package should not be used directly. Instead,
// it should be imported once in your project for its side-effects.
//
// 	import _ "github.com/velovix/snoreslacks/database/gae"
// 	...
//	db, err := database.Get("gae")
package gaedatabase

import (
	"github.com/velovix/snoreslacks/database"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

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

// Transaction runs the given function in a transaction, meaning that the
// modified fields are locked down and can't be changed by other
// goroutines. It may also be able to roll back changes if an error occurs.
func (db GAEDatabase) Transaction(ctx context.Context, f func(context.Context) error) error {
	return datastore.RunInTransaction(ctx, f, &datastore.TransactionOptions{
		XG:       true,
		Attempts: 0})
}
