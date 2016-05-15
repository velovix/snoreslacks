package gaedatabase

import (
	"errors"
	"fmt"

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

// battleNameFromTrainerNames generates the name of a battle that contains
// the two players.
func battleNameFromTrainerNames(p1, p2 string) string {
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

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	_, err := datastore.Put(ctx, battleKey, b)
	if err != nil {
		return err
	}
	return nil
}

// LoadBattle loads a battle from the datastore.
func (db GAEDatabase) LoadBattle(ctx context.Context, p1Name, p2Name string) (database.Battle, bool, error) {
	var battle GAEBattle

	battleKey := datastore.NewKey(ctx, "battle", battleNameFromTrainerNames(p1Name, p2Name), 0, nil)

	err := datastore.Get(ctx, battleKey, &battle)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return &GAEBattle{}, false, nil
		}
		return &GAEBattle{}, false, err
	}

	return &battle, true, nil
}

// LoadBattleTrainerIsIn loads a battle that the trainer is participating in,
// or false as the second value if the trainer is not in any battles.
func (db GAEDatabase) LoadBattleTrainerIsIn(ctx context.Context, pName string) (database.Battle, bool, error) {
	var battles []GAEBattle

	// See if there's a Battle where the player is P1
	_, err := datastore.NewQuery("battle").
		Filter("P1 =", pName).
		GetAll(ctx, &battles)
	if err != nil {
		return &GAEBattle{}, false, err
	}
	if len(battles) == 1 {
		// The battle is found
		return &battles[0], true, nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return &GAEBattle{}, false, errors.New(pName + " appears to be in more than one battle at once")
	}

	// See if there's a Battle where the player is P2
	_, err = datastore.NewQuery("battle").
		Filter("P2 =", pName).
		GetAll(ctx, &battles)
	if err != nil {
		return &GAEBattle{}, false, err
	}
	if len(battles) == 1 {
		// The battle is found
		return &battles[0], true, nil
	} else if len(battles) > 1 {
		// The player is in more than one battle at once. This should not happen
		return &GAEBattle{}, false, errors.New(pName + " appears to be in more than one battle at once")
	}

	// No battle of this type exists
	return &GAEBattle{}, false, nil
}

// DeleteBattle deletes the battle from the datastore
func (db GAEDatabase) DeleteBattle(ctx context.Context, p1Name, p2Name string) error {
	battleKey := datastore.NewKey(ctx, "battle", battleNameFromTrainerNames(p1Name, p2Name), 0, nil)
	return datastore.Delete(ctx, battleKey)
}

// PurgeBattle deletes the battle from the Datastore and any relating data.
func (db GAEDatabase) PurgeBattle(ctx context.Context, p1Name, p2Name string) error {
	b, found, err := db.LoadBattle(ctx, p1Name, p2Name)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("no battle found with player 1: %s player 2: %s", p1Name, p2Name)
	}

	err = db.DeleteTrainerBattleInfos(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeletePokemonBattleInfos(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeleteMoveLookupTables(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeletePartyMemberLookupTables(ctx, b)
	if err != nil {
		return err
	}
	err = db.DeleteBattle(ctx, p1Name, p2Name)
	if err != nil {
		return err
	}

	return nil
}
