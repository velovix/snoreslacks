package pkmn

// BattleActionType represents the type of an action that can be made in a
// turn.
type BattleActionType int

const (
	_ BattleActionType = iota
	MoveBattleActionType
	SwitchBattleActionType
)

// BattleAction represents a single action made by a trainer. The val parameter
// is a generic value that means different things depending on the battle
// action type.
type BattleAction struct {
	Type BattleActionType
	Val  int
}

// PokemonBattleInfo contains information on the battling status of a single
// Pokemon. This structure acts as a place for Pokemon data to be held when it
// pertains only to the battle at hand and isn't otherwise permanent.
type PokemonBattleInfo struct {
	PkmnUUID string

	CurrHP     int
	AttStage   int
	DefStage   int
	SpAttStage int
	SpDefStage int
	SpeedStage int

	Burned    bool
	Frozen    bool
	Paralyzed bool
	Poisoned  bool
	Asleep    bool
	Confused  bool
}

// TrainerBattleInfo contains information on the battling status of a single
// trainer. This structure acts as a place for Trainer data to be held when it
// pertains only to the battle at hand and isn't otherwise permanent.
type TrainerBattleInfo struct {
	Name             string
	FinishedTurn     bool
	NextBattleAction BattleAction
	CurrPkmnSlot     int
}

// BattleMode represents what point the battle is in.
type BattleMode int

const (
	_ BattleMode = iota
	WaitingBattleMode
	StartedBattleMode
)

// Battle represents a battle between two players. It contains data about
// battle progress and trainer info.
type Battle struct {
	P1 string
	P2 string

	Mode BattleMode
}
