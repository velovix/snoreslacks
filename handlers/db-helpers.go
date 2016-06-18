package handlers

import (
	"github.com/pkg/errors"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
	"golang.org/x/net/context"
)

// loadBasicTrainerData loads some basic information about the trainer. It
// assumes that all information will be present and errors out if it is not.
func loadBasicTrainerData(ctx context.Context, db database.Database, uuid string) (*basicTrainerData, error) {
	var td basicTrainerData
	var err error

	// Read in the trainer data
	td.trainer, err = db.LoadTrainer(ctx, uuid)
	if err != nil {
		return &basicTrainerData{}, err
	}

	// Load the trainer's party
	td.pkmn, err = db.LoadParty(ctx, td.trainer)
	// It's okay if the trainer doesn't have any Pokemon yet
	if err != nil && !database.IsNoResults(err) {
		return &basicTrainerData{}, err
	}

	// Load the last contact URL if it exists
	td.lastContactURL, err = db.LoadLastContactURL(ctx, td.trainer)
	// it's okay if the trainer doesn't have a last contact URL yet
	if err != nil && !database.IsNoResults(err) {
		return &basicTrainerData{}, err
	}

	return &td, nil
}

// saveBasicTrainerData saves all relevant information contained in the data
// structure to the database.
func saveBasicTrainerData(ctx context.Context, db database.Database, btd *basicTrainerData) error {
	// The last contact URL is not saved here because it is saved by the main
	// handler

	// Save the party
	err := db.SaveParty(ctx, btd.trainer, btd.pkmn)
	if err != nil {
		return err
	}

	// Save the trainer
	err = db.SaveTrainer(ctx, btd.trainer)
	if err != nil {
		return err
	}

	return nil
}

// loadBattleTrainerData loads battle information on the trainer as well as
// some basic trainer information, given the battle that this data is in
// reference to. It assumes all relevant information is available and errors
// out if this is not the case.
func loadBattleTrainerData(ctx context.Context, db database.Database, b database.Battle, uuid string) (*battleTrainerData, error) {
	var err error
	var btd battleTrainerData

	// Load the basic trainer data
	btd.basicTrainerData, err = loadBasicTrainerData(ctx, db, uuid)
	if err != nil {
		return &battleTrainerData{}, errors.Wrap(err, "loading battle trainer data")
	}

	// Load the trainer's battle info
	btd.battleInfo, err = db.LoadTrainerBattleInfo(ctx, b, uuid)
	if err != nil {
		return &battleTrainerData{}, errors.Wrap(err, "loading battle trainer data")
	}

	// Load each Pokemon's battle info
	for _, pkmn := range btd.pkmn {
		pbi, err := db.LoadPokemonBattleInfo(ctx, b, pkmn.GetPokemon().UUID)
		if err != nil {
			return &battleTrainerData{}, errors.Wrap(err, "loading battle trainer data")
		}
		btd.pkmnBattleInfo = append(btd.pkmnBattleInfo, pbi)
	}

	return &btd, nil

}

// loadBattleData loads the current battle the given trainer is in along with a
// bunch of other useful information about the battle. It will return an
// incomplete object if one of its components is missing or if a database error
// occurs, but will still return those errors. It's safe to use this object if
// it returns a NoResults error, but not otherwise.
func loadBattleData(ctx context.Context, db database.Database, requester *basicTrainerData) (*battleData, error) {
	var err error
	var bd battleData

	// Load the battle the trainer is in
	bd.battle, err = db.LoadBattleTrainerIsIn(ctx, requester.trainer.GetTrainer().UUID)
	if err != nil {
		return &bd, errors.Wrap(err, "loading battle data")
	}

	// Load the requester information
	bd.requester, err = loadBattleTrainerData(ctx, db, bd.battle, requester.trainer.GetTrainer().UUID)
	if err != nil {
		return &bd, errors.Wrap(err, "loading battle data")
	}

	// Figure out the UUID of the opponent
	opponentUUID := bd.battle.GetBattle().P1
	if opponentUUID == requester.trainer.GetTrainer().UUID {
		opponentUUID = bd.battle.GetBattle().P2
	}
	// Load the opponent information
	bd.opponent, err = loadBattleTrainerData(ctx, db, bd.battle, opponentUUID)
	if err != nil {
		return &bd, errors.Wrap(err, "loading battle data")
	}

	return &bd, nil
}

// saveBattleData saves all objects loaded in the battle data.
func saveBattleData(ctx context.Context, db database.Database, bd *battleData) error {

	// Save the battle
	err := db.SaveBattle(ctx, bd.battle)
	if err != nil {
		return errors.Wrap(err, "saving battle data")
	}

	// Save the trainer battle infos
	err = db.SaveTrainerBattleInfo(ctx, bd.battle, bd.requester.battleInfo)
	if err != nil {
		return errors.Wrap(err, "saving battle data")
	}
	err = db.SaveTrainerBattleInfo(ctx, bd.battle, bd.opponent.battleInfo)
	if err != nil {
		return errors.Wrap(err, "saving battle data")
	}

	// Save the Pokemon battle infos
	for _, pkmnBI := range bd.requester.pkmnBattleInfo {
		err = db.SavePokemonBattleInfo(ctx, bd.battle, pkmnBI)
		if err != nil {
			return errors.Wrap(err, "saving battle data")
		}
	}
	for _, pkmnBI := range bd.opponent.pkmnBattleInfo {
		err = db.SavePokemonBattleInfo(ctx, bd.battle, pkmnBI)
		if err != nil {
			return errors.Wrap(err, "saving battle data")
		}
	}

	// Save the Pokemon
	for _, pkmn := range bd.requester.pkmn {
		err = db.SavePokemon(ctx, bd.requester.trainer, pkmn)
		if err != nil {
			return errors.Wrap(err, "saving battle data")
		}
	}
	for _, pkmn := range bd.opponent.pkmn {
		err = db.SavePokemon(ctx, bd.opponent.trainer, pkmn)
		if err != nil {
			return errors.Wrap(err, "saving battle data")
		}
	}

	// Save the trainers
	err = db.SaveTrainer(ctx, bd.requester.trainer)
	if err != nil {
		return errors.Wrap(err, "saving battle data")
	}
	err = db.SaveTrainer(ctx, bd.opponent.trainer)
	if err != nil {
		return errors.Wrap(err, "saving battle data")
	}
	return nil
}

// loadMove fetches the move info from PokeAPI and returns a move object.
func loadMove(ctx context.Context, client messaging.Client, fetcher pokeapi.Fetcher, id int) (pkmn.Move, error) {
	// Load the move from PokeAPI
	apiMove, err := fetcher.FetchMove(ctx, client, id)
	if err != nil {
		return pkmn.Move{}, errors.Wrap(err, "loading a move")
	}
	// Use the PokeAPI data to create a pkmn.Move
	move, err := pokeapi.NewMove(apiMove)
	if err != nil {
		return pkmn.Move{}, errors.Wrap(err, "loading a move")
	}
	return move, nil
}
