package pkmn

type Stat struct {
	Base int
	IV   int
	EV   int
}

func CalcStat(s Stat, pkmn Pokemon) int {
	return ((((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + 5)
}

func CalcHP(s Stat, pkmn Pokemon) int {
	return ((((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + pkmn.Level + 10)
}
