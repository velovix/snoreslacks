package pkmn

import "math"

func (pkmn *Pokemon) LevelUp(learnableMoves map[int][]int) {

}

// Experience returns the experience gained by defeating the given Pokemon.
func Experience(faintedPkmn Pokemon, wild bool) int {
	var wildMod float64
	if wild {
		wildMod = 1.0
	} else {
		wildMod = 1.5
	}

	baseExp := float64(faintedPkmn.BaseExperience)
	level := float64(faintedPkmn.Level)

	return int((baseExp * wildMod * level) / (7.0))
}

// ReadyToLevelUp returns true if the Pokemon is ready to level up.
func (pkmn *Pokemon) ReadyToLevelUp() bool {
	return pkmn.Experience >= RequiredExperience(pkmn.GrowthRate, pkmn.Level+1)
}

// RequiredExperience returns the necessary experience to reach the given level
// n for the given growth pattern.
func RequiredExperience(growthRate GrowthRate, n int) int {
	switch growthRate {
	case ErraticGrowthRate:
		return ErraticExp(n + 1)
	case FastGrowthRate:
		return FastExp(n + 1)
	case MediumFastGrowthRate:
		return MediumFastExp(n + 1)
	case MediumSlowGrowthRate:
		return MediumSlowExp(n + 1)
	case SlowGrowthRate:
		return SlowExp(n + 1)
	case FluctuatingGrowthRate:
		return FluctuatingExp(n + 1)
	default:
		panic("invalid growth rate")
	}
}

// ErraticExp returns the necessary experience to reach the given level n for
// the "erratic" growth pattern.
func ErraticExp(n int) int {
	nf := float64(n)

	if n <= 50 {
		return int((math.Pow(nf, 3.0) * (100.0 - nf)) / 50.0)
	} else if n <= 68 {
		return int((math.Pow(nf, 3.0) * (150.0 - nf)) / 100.0)
	} else if n <= 98 {
		return int((math.Pow(nf, 3.0) * ((1911.0 - 10.0*nf) / 3.0)) / 500.0)
	} else if n <= 100 {
		return int((math.Pow(nf, 3.0) * (160.0 - nf)) / 100.0)
	} else {
		panic("invalid level")
	}
}

// FastExp returns the necessary experience to reach the given level n for the
// "fast" growth pattern.
func FastExp(n int) int {
	nf := float64(n)
	return int((4.0 * math.Pow(nf, 3.0)) / 5.0)
}

// MediumFastExp returns the necessary experience to reach the given level n
// for the "medium fast" growth pattern.
func MediumFastExp(n int) int {
	nf := float64(n)
	return int(math.Pow(nf, 3.0))
}

// MediumSlowExp returns the necessary experience to reach the given level n
// for the "medium slow" growth pattern.
func MediumSlowExp(n int) int {
	nf := float64(n)
	return int((6.0/5.0)*math.Pow(nf, 3.0) - 15.0*math.Pow(nf, 2.0) + 100.0*nf - 140)
}

// SlowExp returns the necessary experience to reach the given level n for the
// "slow" growth pattern.
func SlowExp(n int) int {
	nf := float64(n)
	return int((5.0 * math.Pow(nf, 3.0)) / 4.0)
}

// FluctuatingExp returns the necessary experience to reach the given level n
// for the "fluctuating" growth pattern.
func FluctuatingExp(n int) int {
	nf := float64(n)
	if n <= 15 {
		return int(math.Pow(nf, 3.0) * ((((nf + 1.0) / 3.0) + 24.0) / 50.0))
	} else if n <= 36 {
		return int(math.Pow(nf, 3.0) * ((nf + 14.0) / 50.0))
	} else if n <= 100 {
		return int(math.Pow(nf, 3.0) * (((nf / 2.0) + 32.0) / 50.0))
	} else {
		panic("invalid level")
	}
}
