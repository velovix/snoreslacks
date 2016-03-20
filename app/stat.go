package app

type stat struct {
	Base int
	IV   int
	EV   int
}

func calcHP(s stat, pkmn pokemon) int {
	return (((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + pkmn.Level + 10
}

func calcStat(s stat, pkmn pokemon) int {
	return ((((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + 5) // * Nature
}
