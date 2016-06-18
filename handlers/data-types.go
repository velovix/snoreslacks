package handlers

import "github.com/velovix/snoreslacks/database"

// basicTrainerData contains trainer info retrieved from the database.
type basicTrainerData struct {
	trainer        database.Trainer
	lastContactURL string
	pkmn           []database.Pokemon
}

// battleTrainerData contains the information of a single player in a battle
// session as well as all basic trainer data.
type battleTrainerData struct {
	*basicTrainerData
	battleInfo     database.TrainerBattleInfo
	pkmnBattleInfo []database.PokemonBattleInfo
}

// activePkmn returns the current active Pokemon of the trainer.
func (btd *battleTrainerData) activePkmn() database.Pokemon {
	return btd.pkmn[btd.battleInfo.GetTrainerBattleInfo().CurrPkmnSlot]
}

// activePkmnBattleInfo returns the current active Pokemon of the trainer.
func (btd *battleTrainerData) activePkmnBattleInfo() database.PokemonBattleInfo {
	return btd.pkmnBattleInfo[btd.battleInfo.GetTrainerBattleInfo().CurrPkmnSlot]
}

// battleData is a container of most all commonly used information when
// dealing with a pre-existing battle.
type battleData struct {
	battle              database.Battle
	requester, opponent *battleTrainerData
}

// HasBattle returns true if the battle data includes at least a battle.
func (bd battleData) hasBattle() bool {
	return bd.battle != nil
}

// IsComplete returns true if the battle data includes all the data it
// can.
func (bd battleData) isComplete() bool {
	return bd.battle != nil && bd.requester != nil && bd.opponent != nil
}
