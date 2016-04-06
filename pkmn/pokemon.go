package pkmn

type Pokemon struct {
	UUID string

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

	Slot int
}
