package pkmn

import "fmt"

// BattleActionType represents the type of an action that can be made in a
// turn.
type BattleActionType int

const (
	_ BattleActionType = iota
	MoveBattleActionType
	SwitchBattleActionType
	CatchBattleActionType
)

// BattleActionTypePriority returns a number representing the relative priority
// of a battle action. If the priority of one battle action is higher than
// another, then that action should be performed first by default. If two
// battle actions tie in priority, some other means has to be used to decide
// who goes first.
func BattleActionTypePriority(t BattleActionType) int {
	switch t {
	case MoveBattleActionType:
		return 0
	case SwitchBattleActionType:
		return 1
	case CatchBattleActionType:
		return 2
	default:
		panic(fmt.Sprintf("invalid battle action type %v", t))
	}
}

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

	CurrHP        int
	AttStage      int
	DefStage      int
	SpAttStage    int
	SpDefStage    int
	SpeedStage    int
	AccuracyStage int
	EvasionStage  int

	Ailment  Ailment
	Confused bool
}

// TrainerBattleInfo contains information on the battling status of a single
// trainer. This structure acts as a place for Trainer data to be held when it
// pertains only to the battle at hand and isn't otherwise permanent.
type TrainerBattleInfo struct {
	TrainerUUID      string
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
