package gaedatabase

import (
	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// GAEPokemon is the database object wrapper of a Pokemon for datastore.
type GAEPokemon struct {
	pkmn.Pokemon
}

// NewPokemon creates a database Pokemon that is ready to be saved from the
// given pkmn.Pokemon.
func (db GAEDatabase) NewPokemon(p pkmn.Pokemon) database.Pokemon {
	return &GAEPokemon{Pokemon: p}
}

// GetPokemon returns the underlying Pokemon from the database object.
func (pkmn *GAEPokemon) GetPokemon() *pkmn.Pokemon {
	return &pkmn.Pokemon
}

// SavePokemon saves the given Pokemon as owned by the given trainer.
func (db GAEDatabase) SavePokemon(ctx context.Context, dbt database.Trainer, dbpkmn database.Pokemon) error {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	pkmn, ok := dbpkmn.(*GAEPokemon)
	if !ok {
		panic("The given Pokemon is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	trainerKey := datastore.NewKey(ctx, trainerKindName, t.Name, 0, nil)
	pkmnKey := datastore.NewKey(ctx, pokemonKindName, pkmn.UUID, 0, trainerKey)

	_, err := datastore.Put(ctx, pkmnKey, pkmn)
	if err != nil {
		return errors.Wrap(err, "saving Pokemon")
	}

	return nil
}

// LoadPokemon loads a Pokemon with the given UUID. The second return value
// is true if the Pokemon exists, false otherwise.
func (db GAEDatabase) LoadPokemon(ctx context.Context, uuid string) (database.Pokemon, error) {
	var pkmns []*GAEPokemon

	_, err := datastore.NewQuery(pokemonKindName).
		Filter("UUID =", uuid).
		GetAll(ctx, &pkmns)
	if err != nil {
		return &GAEPokemon{}, errors.Wrap(err, "loading Pokemon")
	}

	if len(pkmns) == 0 {
		return &GAEPokemon{}, errors.Wrap(database.ErrNoResults, "loading Pokemon")
	}

	return pkmns[0], nil
}

// DeletePokemon deletes a Pokemon with the given UUID.
func (db GAEDatabase) DeletePokemon(ctx context.Context, uuid string) error {
	var pkmns []*GAEPokemon

	keys, err := datastore.NewQuery(pokemonKindName).
		Filter("UUID =", uuid).
		GetAll(ctx, &pkmns)
	if err != nil {
		return errors.Wrap(err, "deleting Pokemon")
	}

	if len(pkmns) == 0 {
		return errors.New("no Pokemon with the UUID " + uuid + " found to delete")
	}
	if len(pkmns) > 1 {
		return errors.New("multiple Pokemon with the UUID " + uuid + " found while looking for a Pokemon to delete")
	}

	return datastore.Delete(ctx, keys[0])
}

// SaveParty saves a batch of Pokemon as owend by the given trainer.
func (db GAEDatabase) SaveParty(ctx context.Context, dbt database.Trainer, party []database.Pokemon) error {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	for _, dbpkmn := range party {
		pkmn, ok := dbpkmn.(*GAEPokemon)
		if !ok {
			panic("One of the given Pokemon is not of the right type for this implementation. Are you using two implementations by mistake?")
		}

		err := db.SavePokemon(ctx, t, pkmn)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadParty returns all the Pokemon in the given trainer's party. The
// second return value is true if any Pokemon were found, false otherwise.
func (db GAEDatabase) LoadParty(ctx context.Context, dbt database.Trainer) ([]database.Pokemon, error) {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	trainerKey := datastore.NewKey(ctx, trainerKindName, t.Name, 0, nil)

	var gaeParty []*GAEPokemon
	_, err := datastore.NewQuery(pokemonKindName).
		Ancestor(trainerKey).
		GetAll(ctx, &gaeParty)
	if err != nil {
		return make([]database.Pokemon, 0), errors.Wrap(err, "querying for Pokemon")
	}

	if len(gaeParty) == 0 {
		return make([]database.Pokemon, 0), errors.Wrap(database.ErrNoResults, "loading party")
	}

	// Create the interface representation of the party
	party := make([]database.Pokemon, len(gaeParty))
	for i, val := range gaeParty {
		party[i] = val
	}

	return party, nil
}
