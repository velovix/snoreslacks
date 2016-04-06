package gaedatabase

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type GAEPokemon struct {
	pkmn.Pokemon
}

// NewPokemon creates a database Pokemon that is ready to be saved from the
// given pkmn.Pokemon.
func (db GAEDatabase) NewPokemon(p pkmn.Pokemon) database.Pokemon {
	return GAEPokemon{Pokemon: p}
}

func (pkmn GAEPokemon) GetPokemon() *pkmn.Pokemon {
	return &pkmn.Pokemon
}

// SavePokemon saves the given Pokemon as owned by the given trainer.
func (db GAEDatabase) SavePokemon(ctx context.Context, dbt database.Trainer, dbpkmn database.Pokemon) error {
	t, ok := dbt.(GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	pkmn, ok := dbpkmn.(GAEPokemon)
	if !ok {
		panic("The given Pokemon is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	trainerKey := datastore.NewKey(ctx, "trainer", t.Name, 0, nil)
	pkmnKey := datastore.NewKey(ctx, "pokemon", pkmn.UUID, 0, trainerKey)

	_, err := datastore.Put(ctx, pkmnKey, &pkmn)
	if err != nil {
		return err
	}

	return nil
}

// LoadPokemon loads a Pokemon with the given UUID. The second return value
// is true if the Pokemon exists, false otherwise.
func (db GAEDatabase) LoadPokemon(ctx context.Context, uuid string) (database.Pokemon, bool, error) {
	var pkmns []GAEPokemon

	_, err := datastore.NewQuery("pokemon").
		Filter("UUID =", uuid).
		GetAll(ctx, &pkmns)
	if err != nil {
		return GAEPokemon{}, false, err
	}

	if len(pkmns) == 0 {
		return GAEPokemon{}, false, nil
	}

	return pkmns[0], true, nil
}

// SaveParty saves a batch of Pokemon as owend by the given trainer.
func (db GAEDatabase) SaveParty(ctx context.Context, dbt database.Trainer, party []database.Pokemon) error {
	t, ok := dbt.(GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		for _, dbpkmn := range party {
			pkmn, ok := dbpkmn.(GAEPokemon)
			if !ok {
				panic("One of the given Pokemon is not of the right type for this implementation. Are you using two implementations by mistake?")
			}

			err := db.SavePokemon(ctx, t, pkmn)
			if err != nil {
				return err
			}
		}
		return nil
	}, nil)

}

// LoadParty returns all the Pokemon in the given trainer's party. The
// second return value is true if any Pokemon were found, false otherwise.
func (db GAEDatabase) LoadParty(ctx context.Context, dbt database.Trainer) ([]database.Pokemon, bool, error) {
	t, ok := dbt.(GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	trainerKey := datastore.NewKey(ctx, "trainer", t.Name, 0, nil)

	var gaeParty []GAEPokemon
	_, err := datastore.NewQuery("pokemon").
		Ancestor(trainerKey).
		GetAll(ctx, &gaeParty)
	if err != nil {
		return make([]database.Pokemon, 0), false, err
	}

	if len(gaeParty) == 0 {
		return make([]database.Pokemon, 0), false, nil
	}

	// Create the interface representation of the party
	party := make([]database.Pokemon, len(gaeParty))
	for i, val := range gaeParty {
		party[i] = val
	}

	return party, true, nil
}
