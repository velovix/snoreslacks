package pkmn

type Stat struct {
	Base int
	IV   int
	EV   int
}

// CalcOOBStat calculates the actual stat value of any stat besides HP, not
// including in-battle effects.
func CalcOOBStat(s Stat, pkmn Pokemon) int {
	return ((((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + 5)
}

// CalcOOBHP calculates the actual HP value of a Pokemon, not including in-battle
// effects.
func CalcOOBHP(s Stat, pkmn Pokemon) int {
	return ((((2*s.Base + s.IV + (s.EV / 4)) * pkmn.Level) / 100) + pkmn.Level + 10)
}

// CalcIBHP calculates the in-battle attack stat of the Pokemon.
func CalcIBHP(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBAttack calculates the in-battle attack stat of the Pokemon.
func CalcIBAttack(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBDefense calculates the in-battle defense stat of the Pokemon.
func CalcIBDefense(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBSpAtt calculates the in-battle special attack stat of the Pokemon.
func CalcIBSpAtt(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBSpDef calculates the in-battle special defense stat of the Pokemon.
func CalcIBSpDef(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBSpeed calculates the in-battle speed stat of the Pokemon.
func CalcIBSpeed(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return 0
}

// CalcIBAccuracy calculates the in-battle accuracy of the Pokemon in percent.
// This value alone isn't very meaningful, but it works when plugged into the
// move accuracy function.
func CalcIBAccuracy(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	if pkmnBI.AccuracyStage > 0 {
		return int(100.0 * ((3.0 + float64(pkmnBI.AccuracyStage)) / 3.0))
	} else if pkmnBI.AccuracyStage < 0 {
		return int(100.0 * (3.0 / (3.0 + float64(pkmnBI.AccuracyStage))))
	} else {
		return 100
	}
}

// CalcIBEvasion calculates the in-battle evasion of the Pokemon in percent.
// This value alone isn't very meaningful, but it works when plugged into the
// move accuracy function.
func CalcIBEvasion(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	if pkmnBI.EvasionStage > 0 {
		return int(100.0 * (3.0 / (3.0 + float64(pkmnBI.EvasionStage))))
	} else if pkmnBI.EvasionStage < 0 {
		return int(100.0 * ((3.0 + float64(pkmnBI.EvasionStage)) / 3.0))
	} else {
		return 100
	}
}
