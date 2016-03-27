package app

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type dao interface {
	// saveTrainer saves the given trainer.
	saveTrainer(ctx context.Context, t trainer) error
	// loadTrainer returns a trainer with the given name. The second return
	// value is true if the trainer exists and was retrieved, false otherwise.
	loadTrainer(ctx context.Context, name string) (trainer, bool, error)

	// saveBattle saves the given battle.
	saveBattle(ctx context.Context, b battle) error
	// loadBattle returns a battle that the given trainer name is involved in.
	// The second return value is true if the battle exists and was retrieved,
	// false otherwise.
	loadBattle(ctx context.Context, p1Name, p2Name string) (battle, bool, error)
	// loadBattleTrainerIsIn returns a battle the given trainer name is involved
	// in.
	loadBattleTrainerIsIn(ctx context.Context, pName string) (battle, bool, error)
	// deleteTrainer deletes the battle that the two trainers are involved in.
	deleteBattle(ctx context.Context, p1Name, p2Name string) error

	// saveMoveLookupTable saves a move lookup table to the database given the
	// battle object it belongs to.
	saveMoveLookupTable(ctx context.Context, table moveLookupTable, b battle) error
	// loadMoveLookupTables loads all the move lookup tables that the given
	// battle object owns.
	loadMoveLookupTables(ctx context.Context, b battle) ([]moveLookupTable, bool, error)
	// savePartyMemberLookupTable saves a party member lookup table to the
	// database given the battle object it belongs to.
	savePartyMemberLookupTable(ctx context.Context, table partyMemberLookupTable, b battle) error
	// loadPartyMemberLookupTables loads all the party member lookup tables
	// taht the given battle object owns.
	loadPartyMemberLookupTables(ctx context.Context, b battle) ([]partyMemberLookupTable, bool, error)
}

type appengineDatastoreDAO struct{}

// saveTrainer saves the trainer to datastore.
func (ds appengineDatastoreDAO) saveTrainer(ctx context.Context, t trainer) error {
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
			// The Pokemon is stored anonymously, but it contains a UID that is used to
			// make sure that this save overrides the old version of the Pokemon
			var pkmnKey *datastore.Key
			if val.UID == "" {
				// This must be the Pokemon's first time being saved

				pkmnKey = datastore.NewIncompleteKey(ctx, "pokemon", trainerKey)
			} else {
				// Save over the existing data
				pkmnKey, err = datastore.DecodeKey(val.UID)
				if err != nil {
					return err
				}
			}

			_, err = datastore.Put(ctx, pkmnKey, &val)
			if err != nil {
				return err
			}
		}

		return nil
	}, nil)
}

// loadTrainer loads a trainer from datastore.
func (ds appengineDatastoreDAO) loadTrainer(ctx context.Context, name string) (trainer, bool, error) {
	trainerKey := datastore.NewKey(ctx, "trainer", name, 0, nil)
	var t trainer

	// Load the trainer
	err := datastore.Get(ctx, trainerKey, &t)
	if err == datastore.ErrNoSuchEntity {
		return trainer{}, false, nil
	} else if err != nil {
		return trainer{}, false, err
	}

	// Load the trainer's Pokemon
	keys, err := datastore.NewQuery("pokemon").
		Ancestor(trainerKey).
		GetAll(ctx, &t.pkmn)
	if err != nil {
		return trainer{}, false, err
	}

	// Associate the Pokemon's keys with the Pokemon, so that changes can be saved
	// over this instance.
	for i := range t.pkmn {
		t.pkmn[i].UID = keys[i].Encode()
	}

	log.Infof(ctx, "loaded trainer: %v", t)

	return t, true, nil
}

// saveBattle saves a battle to the datastore.
func (ds appengineDatastoreDAO) saveBattle(ctx context.Context, b battle) error {
	// Save the battle anonymously. It will by identified by who is involved
	battleKey := datastore.NewKey(ctx, "battle", b.P1.Name+"+"+b.P2.Name, 0, nil)
	_, err := datastore.Put(ctx, battleKey, &b)
	if err != nil {
		return err
	}
	return nil
}

// loadBattle loads a battle from the datastore.
func (ds appengineDatastoreDAO) loadBattle(ctx context.Context, p1Name, p2Name string) (battle, bool, error) {
	var battles []battle

	// Look for a battle involving the two players
	_, err := datastore.NewQuery("battle").
		Filter("P1Name =", p1Name).
		Filter("P2Name =", p2Name).
		GetAll(ctx, &battles)
	if err != nil {
		return battle{}, false, err
	}
	if len(battles) == 1 {
		// The battle is found
		return battles[0], true, nil
	} else if len(battles) > 1 {
		// The players are in more than one battle at once. This should not happen
		return battle{}, false, errors.New(p1Name + " and " + p2Name + " appear to be in more than one battle at once")
	} else if len(battles) == 0 {
		// No battle was found with the given criteria
		return battle{}, false, nil
	}

	return battles[0], true, nil
}

func (ds appengineDatastoreDAO) loadBattleTrainerIsIn(ctx context.Context, pName string) (battle, bool, error) {
	var battles []battle

	// See if there's a battle where the player is P1
	_, err := datastore.NewQuery("battle").
		Filter("P1Name =", pName).
		GetAll(ctx, &battles)
	if err != nil {
		return battle{}, false, err
	}
	if len(battles) == 1 {
		// The battle is found
		return battles[0], true, nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return battle{}, false, errors.New(pName + " appears to be in more than one battle at once")
	}

	// See if there's a battle where the player is P2
	_, err = datastore.NewQuery("battle").
		Filter("P2Name =", pName).
		GetAll(ctx, &battles)
	if err != nil {
		return battle{}, false, err
	}
	if len(battles) == 1 {
		// The battle is found
		return battles[0], true, nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return battle{}, false, errors.New(pName + " appears to be in more than one battle at once")
	}

	// No battle of this type exists
	return battle{}, false, nil
}

// deleteBattle deletes the battle from the Datastore
func (ds appengineDatastoreDAO) deleteBattle(ctx context.Context, p1Name, p2Name string) error {
	battleKey := datastore.NewKey(ctx, "battle", p1Name+"+"+p2Name, 0, nil)
	return datastore.Delete(ctx, battleKey)
}

// saveMoveLookupTable saves the lookup table to the Datastore.
func (ds appengineDatastoreDAO) saveMoveLookupTable(ctx context.Context, table moveLookupTable, b battle) error {
	battleKey := datastore.NewKey(ctx, "battle", b.P1.Name+b.P2.Name, 0, nil)
	tableKey := datastore.NewIncompleteKey(ctx, "moveLookupTable", battleKey)

	_, err := datastore.Put(ctx, tableKey, &b)
	if err != nil {
		return err
	}

	return nil
}

func (ds appengineDatastoreDAO) loadMoveLookupTables(ctx context.Context, b battle) ([]moveLookupTable, bool, error) {
	battleKey := datastore.NewKey(ctx, "battle", b.P1.Name+b.P2.Name, 0, nil)

	var tables []moveLookupTable

	_, err := datastore.NewQuery("moveLookupTable").
		Ancestor(battleKey).
		GetAll(ctx, &tables)
	if err != nil {
		return make([]moveLookupTable, 0), false, err
	}
	if len(tables) == 0 {
		return tables, false, nil
	}

	return tables, true, nil
}

func (ds appengineDatastoreDAO) savePartyMemberLookupTable(ctx context.Context, table partyMemberLookupTable, b battle) error {
	battleKey := datastore.NewKey(ctx, "battle", b.P1.Name+b.P2.Name, 0, nil)
	tableKey := datastore.NewIncompleteKey(ctx, "partyMemberLookupTable", battleKey)

	_, err := datastore.Put(ctx, tableKey, &b)
	if err != nil {
		return err
	}

	return nil
}

func (ds appengineDatastoreDAO) loadPartyMemberLookupTables(ctx context.Context, b battle) ([]partyMemberLookupTable, bool, error) {
	battleKey := datastore.NewKey(ctx, "battle", b.P1.Name+b.P2.Name, 0, nil)

	var tables []partyMemberLookupTable

	_, err := datastore.NewQuery("partyMemberLookupTable").
		Ancestor(battleKey).
		GetAll(ctx, &tables)
	if err != nil {
		return make([]partyMemberLookupTable, 0), false, err
	}
	if len(tables) == 0 {
		return tables, false, nil
	}

	return tables, true, nil
}
