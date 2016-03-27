package app

// moveLookupElement represents a single match between a scrambled ID and the
// corresponding move ID.
type moveLookupElement struct {
	ID     int
	MoveID int
}

// moveLookupTable contains a collection of matches between scrambled IDs and
// move IDs.
type moveLookupTable struct {
	TrainerName string
	Moves       []moveLookupElement
}

// Lookup finds the move ID that corresponds to the given scambled ID.
func (table moveLookupTable) lookup(id int) int {
	for _, val := range table.Moves {
		if val.ID == id {
			return val.MoveID
		}
	}

	return -1
}

// partyMemberLookupElement represents a single match between a scrambled ID
// and the corresponding party slot.
type partyMemberLookupElement struct {
	ID     int
	SlotID int
}

// partyMemberLookupTable contains a collection of matches between scrambled
// IDs and party slots.
type partyMemberLookupTable struct {
	TrainerName string
	Members     []partyMemberLookupElement
}

// Lookup finds the party slot that corresponds to the given scrambled ID.
func (table partyMemberLookupTable) lookup(id int) int {
	for _, val := range table.Members {
		if val.ID == id {
			return val.SlotID
		}
	}

	return -1
}

// battleActionType represents the type of an action that can be made in a
// turn.
type battleActionType int

const (
	_ battleActionType = iota
	moveBattleActionType
	switchBattleActionType
)

// battleAction represents a single action made by a trianer. The val parameter
// is a generic value that means different things depending on the battle
// action type.
type battleAction struct {
	Type battleActionType
	Val  int
}

// trainerBattleInfo contains information on the battling status of a single
// trainer.
type trainerBattleInfo struct {
	Name             string
	FinishedTurn     bool
	NextBattleAction battleAction
	CurrPokemonSlot  int
}

// battleMode represents what point the battle is in.
type battleMode int

const (
	_ battleMode = iota
	waitingBattleMode
	startedBattleMode
)

// battle represents a battle between two players. It contains data about
// battle progress and trainer info.
type battle struct {
	P1 trainerBattleInfo
	P2 trainerBattleInfo

	Mode battleMode
}
