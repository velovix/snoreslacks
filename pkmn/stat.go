package pkmn

// Stat is the information on a single stat.
type Stat struct {
	Base int
	IV   int
	EV   int
}

func statMod(mod int) float64 {
	if mod >= 0 {
		return float64(mod+2) / 2.0
	} else {
		return 2.0 / float64((-mod)+2)
	}
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
	return CalcOOBStat(pkmn.HP, pkmn)
}

// CalcIBAttack calculates the in-battle attack stat of the Pokemon.
func CalcIBAttack(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return int(float64(((((2*pkmn.Attack.Base + pkmn.Attack.IV + (pkmn.Attack.EV / 4)) * pkmn.Level) / 100) + 5)) * statMod(pkmnBI.AttStage))
}

// CalcIBDefense calculates the in-battle defense stat of the Pokemon.
func CalcIBDefense(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return int(float64(((((2*pkmn.Defense.Base + pkmn.Defense.IV + (pkmn.Defense.EV / 4)) * pkmn.Level) / 100) + 5)) * statMod(pkmnBI.DefStage))
}

// CalcIBSpAtt calculates the in-battle special attack stat of the Pokemon.
func CalcIBSpAtt(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return int(float64(((((2*pkmn.SpAttack.Base + pkmn.SpAttack.IV + (pkmn.SpAttack.EV / 4)) * pkmn.Level) / 100) + 5)) * statMod(pkmnBI.SpAttStage))
}

// CalcIBSpDef calculates the in-battle special defense stat of the Pokemon.
func CalcIBSpDef(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return int(float64(((((2*pkmn.SpDefense.Base + pkmn.SpDefense.IV + (pkmn.SpDefense.EV / 4)) * pkmn.Level) / 100) + 5)) * statMod(pkmnBI.SpDefStage))
}

// CalcIBSpeed calculates the in-battle speed stat of the Pokemon.
func CalcIBSpeed(pkmn Pokemon, pkmnBI PokemonBattleInfo) int {
	return int(float64(((((2*pkmn.Speed.Base + pkmn.Speed.IV + (pkmn.Speed.EV / 4)) * pkmn.Level) / 100) + 5)) * statMod(pkmnBI.SpeedStage))
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
