package pkmn

import "github.com/pkg/errors"

type Pokemon struct {
	UUID string

	SpriteURL string

	ID     int
	Name   string
	Height int
	Weight int
	Type1  string
	Type2  string

	Level int

	HP        Stat
	Attack    Stat
	Defense   Stat
	SpAttack  Stat
	SpDefense Stat
	Speed     Stat

	Move1 int
	Move2 int
	Move3 int
	Move4 int

	CatchRate  int
	GrowthRate GrowthRate

	Slot int

	BaseExperience int
	Experience     int
}

type GrowthRate int

const (
	_ GrowthRate = iota
	ErraticGrowthRate
	FastGrowthRate
	MediumFastGrowthRate
	MediumSlowGrowthRate
	SlowGrowthRate
	FluctuatingGrowthRate
)

// MoveCount returns the number of moves the Pokemon has.
func (pkmn *Pokemon) MoveCount() int {
	cnt := 0

	if pkmn.Move1 != 0 {
		cnt++
	}
	if pkmn.Move2 != 0 {
		cnt++
	}
	if pkmn.Move3 != 0 {
		cnt++
	}
	if pkmn.Move4 != 0 {
		cnt++
	}

	return cnt
}

// MoveIDsAsSlice returns the Pokemon's move IDs as a slice.
func (pkmn *Pokemon) MoveIDsAsSlice() []int {
	moves := make([]int, pkmn.MoveCount())

	if pkmn.Move1 != 0 {
		moves[0] = pkmn.Move1
	}
	if pkmn.Move2 != 0 {
		moves[1] = pkmn.Move2
	}
	if pkmn.Move3 != 0 {
		moves[2] = pkmn.Move3
	}
	if pkmn.Move4 != 0 {
		moves[3] = pkmn.Move4
	}

	return moves
}

func (pkmn *Pokemon) LearnMove(moveID int) error {
	if pkmn.Move1 == 0 {
		pkmn.Move1 = moveID
	} else if pkmn.Move2 == 0 {
		pkmn.Move2 = moveID
	} else if pkmn.Move3 == 0 {
		pkmn.Move3 = moveID
	} else if pkmn.Move4 == 0 {
		pkmn.Move4 = moveID
	} else {
		return errors.New("attempt to give a Pokemon a new move when all move slots are full")
	}

	return nil
}
