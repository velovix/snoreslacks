package gaedatabase

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// GAETrainerBattleInfo is a database wrapper around a trainer battle info for
// Datastore.
type GAETrainerBattleInfo struct {
	pkmn.TrainerBattleInfo
}

// NewTrainerBattleInfo creates a new trainer battle info that is ready to
// be saved from the given pkmn.TrainerBattleInfo.
func (db GAEDatabase) NewTrainerBattleInfo(tbi pkmn.TrainerBattleInfo) database.TrainerBattleInfo {
	return &GAETrainerBattleInfo{TrainerBattleInfo: tbi}
}

// GetTrainerBattleInfo returns the underlying trainer battle info from the
// database object, which may by modified and saved.
func (tbi *GAETrainerBattleInfo) GetTrainerBattleInfo() *pkmn.TrainerBattleInfo {
	return &tbi.TrainerBattleInfo
}

// GAEPokemonBattleInfo is a database wrapper around a Pokemon battle info for
// Datastore.
type GAEPokemonBattleInfo struct {
	pkmn.PokemonBattleInfo
}

// NewPokemonBattleInfo creates a new Pokemon battle info that is ready to
// be saved from the given Pokemon battle info.
func (db GAEDatabase) NewPokemonBattleInfo(pbi pkmn.PokemonBattleInfo) database.PokemonBattleInfo {
	return &GAEPokemonBattleInfo{PokemonBattleInfo: pbi}
}

// GetPokemonBattleInfo returns the underlying Pokemon battle info from the
// database object, which may be modified and saved.
func (pbi *GAEPokemonBattleInfo) GetPokemonBattleInfo() *pkmn.PokemonBattleInfo {
	return &pbi.PokemonBattleInfo
}

// SaveTrainerBattleInfo saves the given trainer battle info.
func (db GAEDatabase) SaveTrainerBattleInfo(ctx context.Context, dbb database.Battle, dbtbi database.TrainerBattleInfo) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	tbi, ok := dbtbi.(*GAETrainerBattleInfo)
	if !ok {
		panic("The given trainer battle info is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	tbiKey := datastore.NewKey(ctx, "trainer battle info", tbi.Name, 0, battleKey)

	_, err := datastore.Put(ctx, tbiKey, tbi)
	if err != nil {
		return err
	}

	return nil
}

// LoadTrainerBattleInfo returns a trainer battle info for the given
// trainer name. The second return value is true if the battle info exists
// and was retrieved, false otherwise.
func (db GAEDatabase) LoadTrainerBattleInfo(ctx context.Context, dbb database.Battle, tName string) (database.TrainerBattleInfo, bool, error) {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	tbiKey := datastore.NewKey(ctx, "trainer battle info", tName, 0, battleKey)

	var tbi GAETrainerBattleInfo
	err := datastore.Get(ctx, tbiKey, &tbi)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return &GAETrainerBattleInfo{}, false, nil
		}

		return &GAETrainerBattleInfo{}, false, err
	}

	return &tbi, true, nil
}

// DeleteTrainerBattleInfos deletes all trainer battle infos under the given
// battle.
func (db GAEDatabase) DeleteTrainerBattleInfos(ctx context.Context, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	// Find all the trainer battle infos under this battle
	keys, err := datastore.NewQuery("trainer battle info").
		KeysOnly().
		Ancestor(battleKey).
		GetAll(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all the trainer battle infos
	for _, key := range keys {
		err = datastore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

// SavePokemonBattleInfo saves the given Pokemon battle info.
func (db GAEDatabase) SavePokemonBattleInfo(ctx context.Context, dbb database.Battle, dbpbi database.PokemonBattleInfo) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	pbi, ok := dbpbi.(*GAEPokemonBattleInfo)
	if !ok {
		panic("The given Pokemon battle info is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	pbiKey := datastore.NewKey(ctx, "pokemon battle info", pbi.PkmnUUID, 0, battleKey)

	_, err := datastore.Put(ctx, pbiKey, pbi)
	if err != nil {
		return err
	}

	return nil
}

// LoadPokemonBattleInfo returns a Pokemon battle info for the given
// Pokemon UUID. The second return value is true if the battle info exists
// and was retrieved, false otherwise.
func (db GAEDatabase) LoadPokemonBattleInfo(ctx context.Context, dbb database.Battle, uuid string) (database.PokemonBattleInfo, bool, error) {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	pbiKey := datastore.NewKey(ctx, "pokemon battle info", uuid, 0, battleKey)

	var pbi GAEPokemonBattleInfo
	err := datastore.Get(ctx, pbiKey, &pbi)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return &GAEPokemonBattleInfo{}, false, nil
		}
		return &GAEPokemonBattleInfo{}, false, err
	}

	return &pbi, true, nil
}

// DeletePokemonBattleInfos deletes all Pokemon battle infos under the given
// battle.
func (db GAEDatabase) DeletePokemonBattleInfos(ctx context.Context, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	// Find all the Pokemon battle infos under this battle
	keys, err := datastore.NewQuery("pokemon battle info").
		KeysOnly().
		Ancestor(battleKey).
		GetAll(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all the pokemon battle infos
	for _, key := range keys {
		err = datastore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}
