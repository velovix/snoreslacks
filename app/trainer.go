package app

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type trainerMode int

const (
	_ trainerMode = iota
	starterTrainerMode
	waitingTrainerMode
	challengingTrainerMode
)

// saveTrainer saves the trainer to the database.
func saveTrainer(ctx context.Context, t trainer) error {
	// Save both the trainer and the party in a single transaction
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// Save the trainer data
		trainerKey := datastore.NewKey(ctx, "trainer", t.Name, 0, nil)
		_, err := datastore.Put(ctx, trainerKey, &t)
		if err != nil {
			return err
		}

		// Save each Pokemon and make them children of the trainer
		for _, val := range t.pkmn {
			pkmnKey := datastore.NewKey(ctx, "pokemon", val.Name, 0, trainerKey)
			_, err := datastore.Put(ctx, pkmnKey, &val)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil)
}

// loadTrainer loads a trainer for the database.
func loadTrainer(ctx context.Context, name string) (trainer, error) {
	trainerKey := datastore.NewKey(ctx, "trainer", name, 0, nil)
	t := trainer{}

	err := datastore.Get(ctx, trainerKey, &t)
	if err != nil {
		return trainer{}, err
	}

	_, err = datastore.NewQuery("pokemon").Ancestor(trainerKey).GetAll(ctx, &t.pkmn)
	if err != nil {
		return trainer{}, err
	}

	log.Infof(ctx, "loaded trainer: %v", t)

	return t, nil
}

type trainer struct {
	Name string
	pkmn []pokemon
	Mode trainerMode
}
