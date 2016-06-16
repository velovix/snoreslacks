package gaedatabase

import (
	"github.com/pkg/errors"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// GAETrainer is a database object wrapper of a trainer for datastore.
type GAETrainer struct {
	pkmn.Trainer
}

// NewTrainer creates a database trainer that is ready to be saved from the
// given pkmn.Trainer.
func (db GAEDatabase) NewTrainer(t pkmn.Trainer) database.Trainer {
	return &GAETrainer{Trainer: t}
}

// GetTrainer returns the underlying trainer from the database object.
func (t *GAETrainer) GetTrainer() *pkmn.Trainer {
	return &t.Trainer
}

// SaveTrainer saves the trainer to datastore.
func (db GAEDatabase) SaveTrainer(ctx context.Context, dbt database.Trainer) error {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	// Save the trainer data
	trainerKey := datastore.NewKey(ctx, trainerKindName, t.UUID, 0, nil)
	_, err := datastore.Put(ctx, trainerKey, t)
	if err != nil {
		return errors.Wrap(err, "saving trainer")
	}

	return nil
}

// LoadTrainer Loads a trainer from datastore.
func (db GAEDatabase) LoadTrainer(ctx context.Context, uuid string) (database.Trainer, error) {
	trainerKey := datastore.NewKey(ctx, trainerKindName, uuid, 0, nil)
	var t GAETrainer

	// Load the Trainer
	err := datastore.Get(ctx, trainerKey, &t)
	if err == datastore.ErrNoSuchEntity {
		return &GAETrainer{}, errors.Wrap(database.ErrNoResults, "loading trainer")
	} else if err != nil {
		return &GAETrainer{}, errors.Wrap(err, "loading trainer")
	}

	return &t, nil
}

// DeleteTrainer deletes the trainer from the database with the given UUID.
func (db GAEDatabase) DeleteTrainer(ctx context.Context, uuid string) error {
	trainerKey := datastore.NewKey(ctx, trainerKindName, uuid, 0, nil)

	err := datastore.Delete(ctx, trainerKey)
	if err != nil {
		return errors.Wrap(err, "deleting trainer")
	}

	return nil
}

// PurgeTrainer deletes the trainer with the given UUID and all of their
// Pokemon from the database.
func (db GAEDatabase) PurgeTrainer(ctx context.Context, uuid string) error {
	trainerKey := datastore.NewKey(ctx, trainerKindName, uuid, 0, nil)

	// Find and delete all the trainer's Pokemon
	query := datastore.NewQuery(pokemonKindName).
		Ancestor(trainerKey)
	for t := query.Run(ctx); ; {
		var pkmn database.Pokemon
		key, err := t.Next(&pkmn)
		if err == datastore.Done {
			// All the trainer's Pokemon have been deleted
			break
		}
		if err != nil {
			return errors.Wrap(err, "deleting party")
		}
		datastore.Delete(ctx, key)
	}

	// Delete the trainer
	err := datastore.Delete(ctx, trainerKey)
	if err != nil {
		return errors.Wrap(err, "deleting trainer")
	}

	return nil
}

// SaveLastContactURL saves the given last contact URL as associated with
// the given trainer.
func (db GAEDatabase) SaveLastContactURL(ctx context.Context, dbt database.Trainer, url string) error {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	urlKey := datastore.NewKey(ctx, lastContactURLKindName, t.UUID, 0, nil)

	urlContainer := struct {
		URL string
	}{URL: url}

	_, err := datastore.Put(ctx, urlKey, &urlContainer)
	if err != nil {
		return errors.Wrap(err, "saving last contact URL")
	}

	return nil
}

// LoadLastContactURL loads the last contact URL associated with the given
// trainer. The second return value is true if there is a last contact
// URL associated with this trainer, false otherwise.
func (db GAEDatabase) LoadLastContactURL(ctx context.Context, dbt database.Trainer) (string, error) {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	urlKey := datastore.NewKey(ctx, lastContactURLKindName, t.UUID, 0, nil)

	var url struct {
		URL string
	}
	err := datastore.Get(ctx, urlKey, &url)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return "", errors.Wrap(database.ErrNoResults, "loading last contact URL")
		}
		return "", errors.Wrap(err, "loading last contact URL")
	}

	return url.URL, nil
}

// LoadUUIDFromHumanTrainerName finds the corresponding UUID for the given
// name of a human (non-NPC) trainer.
func (db GAEDatabase) LoadUUIDFromHumanTrainerName(ctx context.Context, name string) (string, error) {
	var trainers []database.Trainer

	_, err := datastore.NewQuery(trainerKindName).
		Filter("Name =", name).
		GetAll(ctx, &trainers)
	if err != nil {
		return "", err
	}
	if len(trainers) == 0 {
		return "", errors.Wrap(database.ErrNoResults, "loading UUID from human trainer name")
	}
	if len(trainers) > 0 {
		return "", errors.Errorf("multiple human trainers share the same name '%s'", name)
	}

	return trainers[0].GetTrainer().Name, nil
}
