package gaedatabase

import (
	"log"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type GAETrainer struct {
	pkmn.Trainer
}

// NewTrainer creates a database trainer that is ready to be saved from the
// given pkmn.Trainer.
func (db GAEDatabase) NewTrainer(t pkmn.Trainer) database.Trainer {
	return &GAETrainer{Trainer: t}
}

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
	trainerKey := datastore.NewKey(ctx, "trainer", t.Name, 0, nil)
	_, err := datastore.Put(ctx, trainerKey, t)
	if err != nil {
		return err
	}

	return nil
}

// LoadTrainer Loads a trainer from datastore.
func (db GAEDatabase) LoadTrainer(ctx context.Context, name string) (database.Trainer, bool, error) {
	trainerKey := datastore.NewKey(ctx, "trainer", name, 0, nil)
	var t GAETrainer

	// Load the Trainer
	err := datastore.Get(ctx, trainerKey, &t)
	if err == datastore.ErrNoSuchEntity {
		return &GAETrainer{}, false, nil
	} else if err != nil {
		return &GAETrainer{}, false, err
	}

	return &t, true, nil
}

// SaveLastContactURL saves the given last contact URL as associated with
// the given trainer.
func (db GAEDatabase) SaveLastContactURL(ctx context.Context, dbt database.Trainer, url string) error {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	urlKey := datastore.NewKey(ctx, "last contact url", t.Name, 0, nil)

	urlContainer := struct {
		URL string
	}{URL: url}

	_, err := datastore.Put(ctx, urlKey, &urlContainer)
	if err != nil {
		log.Printf("Oh no! Error! %s", err)
		return err
	}

	return nil
}

// LoadLastContactURL loads the last contact URL associated with the given
// trainer. The second return value is true if there is a last contact
// URL associated with this trainer, false otherwise.
func (db GAEDatabase) LoadLastContactURL(ctx context.Context, dbt database.Trainer) (string, bool, error) {
	t, ok := dbt.(*GAETrainer)
	if !ok {
		panic("The given trainer is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	urlKey := datastore.NewKey(ctx, "last contact url", t.Name, 0, nil)

	var url string
	err := datastore.Get(ctx, urlKey, &url)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return "", false, nil
		} else {
			return "", false, err
		}
	}

	return url, true, nil
}
