package pkmn

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

	CatchRate int

	Slot int
}

// MoveCount returns the number of moves the Pokemon has.
func (pkmn Pokemon) MoveCount() int {
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
func (pkmn Pokemon) MoveIDsAsSlice() []int {
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
