package gaedatabase

import (
	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// battleName generates the name of a battle in the database in the correct
// format.
func battleName(b *GAEBattle) string {
	return b.P1 + "/" + b.P2
}

// battleNameFromTrainerUUIDs generates the name of a battle that contains
// the two players.
func battleNameFromTrainerUUIDs(p1, p2 string) string {
	return p1 + "/" + p2
}

// GAEBattle is a battle database wrapper object for datastore.
type GAEBattle struct {
	pkmn.Battle
}

// NewBattle creates a database battle that is ready to be saved from the
// given pkmn.Battle.
func (db GAEDatabase) NewBattle(b pkmn.Battle) database.Battle {
	return &GAEBattle{Battle: b}
}

// GetBattle returns the underlying battle from the database object.
func (b *GAEBattle) GetBattle() *pkmn.Battle {
	return &b.Battle
}

// SaveBattle saves a battle to the datastore.
func (db GAEDatabase) SaveBattle(ctx context.Context, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, battleKindName, battleName(b), 0, nil)

	_, err := datastore.Put(ctx, battleKey, b)
	if err != nil {
		return errors.Wrap(err, "saving battle")
	}
	return nil
}

// LoadBattle loads a battle from the datastore.
func (db GAEDatabase) LoadBattle(ctx context.Context, p1uuid, p2uuid string) (database.Battle, error) {
	var battle GAEBattle

	battleKey := datastore.NewKey(ctx, battleKindName, battleNameFromTrainerUUIDs(p1uuid, p2uuid), 0, nil)

	err := datastore.Get(ctx, battleKey, &battle)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return &GAEBattle{}, errors.Wrap(database.ErrNoResults, "loading battle")
		}
		return &GAEBattle{}, errors.Wrap(err, "loading battle")
	}

	return &battle, nil
}

// LoadBattleTrainerIsIn loads a battle that the trainer is participating in,
// or false as the second value if the trainer is not in any battles.
func (db GAEDatabase) LoadBattleTrainerIsIn(ctx context.Context, tuuid string) (database.Battle, error) {
	var battles []GAEBattle

	// See if there's a Battle where the player is P1
	_, err := datastore.NewQuery(battleKindName).
		Filter("P1 =", tuuid).
		GetAll(ctx, &battles)
	if err != nil {
		return &GAEBattle{}, errors.Wrap(err, "loading battle trainer is in")
	}
	if len(battles) == 1 {
		// The battle is found
		return &battles[0], nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return &GAEBattle{}, errors.New(tuuid + " appears to be in more than one battle at once")
	}

	// See if there's a Battle where the player is P2
	_, err = datastore.NewQuery(battleKindName).
		Filter("P2 =", tuuid).
		GetAll(ctx, &battles)
	if err != nil {
		return &GAEBattle{}, errors.Wrap(err, "loading battle trainer is in")
	}
	if len(battles) == 1 {
		// The battle is found
		return &battles[0], nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return &GAEBattle{}, errors.New(tuuid + " appears to be in more than one battle at once")
	}

	// No battle of this type exists
	return &GAEBattle{}, errors.Wrap(database.ErrNoResults, "loading battle trainer is in")
}

// DeleteBattle deletes the battle from the datastore
func (db GAEDatabase) DeleteBattle(ctx context.Context, p1uuid, p2uuid string) error {
	battleKey := datastore.NewKey(ctx, battleKindName, battleNameFromTrainerUUIDs(p1uuid, p2uuid), 0, nil)
	return datastore.Delete(ctx, battleKey)
}

// PurgeBattle deletes the battle from the Datastore and any relating data.
func (db GAEDatabase) PurgeBattle(ctx context.Context, p1uuid, p2uuid string) error {
	b, err := db.LoadBattle(ctx, p1uuid, p2uuid)
	if err != nil {
		if err == database.ErrNoResults {
			return errors.Errorf("no battle found with player 1: %s player 2: %s", p1uuid, p2uuid)
		}
		return err
	}

	err = db.DeleteTrainerBattleInfos(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeletePokemonBattleInfos(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeleteBattle(ctx, p1uuid, p2uuid)
	if err != nil {
		return err
	}

	return nil
}
