package pkmn

import (
	"math/rand"
	"time"
)

// WildEntry describtes the encounter details of the Pokemon as indicated by
// its national Pokedex ID. The encounter rate should be some value between
// 1 and 100.
type WildEntry struct {
	ID          int
	Probability int
	MedianLevel int
}

var kantoWilds = [][]WildEntry{
	{{16, 100, 3}, {19, 100, 3}, {21, 100, 3}, {29, 40, 3}, {32, 40, 3}, {56, 30, 3}, {10, 100, 3}, {13, 90, 3}, {11, 70, 3}, {14, 70, 3}, {25, 10, 3}}}

var johtoWilds = [][]WildEntry{
	{{161, 100, 5}, {163, 90, 5}, {187, 30, 5}, {60, 30, 5}, {165, 70, 5}, {167, 70, 5}, {69, 50, 5}, {92, 5, 5}, {206, 20, 5}, {216, 20, 5}}}

var hoennWilds = [][]WildEntry{
	{{261, 100, 6}, {263, 100, 6}, {265, 100, 6}, {270, 70, 6}, {273, 20, 6}, {280, 10, 6}, {283, 10, 6}, {278, 70, 6}}}

var sinnohWilds = [][]WildEntry{
	{{396, 100, 7}, {399, 100, 7}, {401, 50, 7}, {403, 70, 7}, {406, 70, 7}, {63, 10, 7}, {417, 40, 7}}}

var unovaWilds = [][]WildEntry{
	{{504, 100, 8}, {506, 100, 8}, {509, 80, 8}, {531, 20, 8}, {550, 40, 8}, {517, 10, 8}, {519, 100, 8}}}

var kalosWilds = [][]WildEntry{
	{{661, 100, 9}, {659, 100, 9}, {664, 100, 9}, {511, 10, 9}, {513, 10, 9}, {515, 10, 9}, {25, 20, 9}, {298, 30, 9}, {412, 10, 9}}}

// AvailableWildPokemon returns a list of all Pokemon that the given trainer
// may encounter split up by region.
func AvailableWildPokemon(t Trainer) [][]WildEntry {
	var wilds [][]WildEntry

	for i := 0; i < t.KantoEncounterLevel; i++ {
		wilds = append(wilds, kantoWilds[i])
	}
	for i := 0; i < t.JohtoEncounterLevel; i++ {
		wilds = append(wilds, johtoWilds[i])
	}
	for i := 0; i < t.HoennEncounterLevel; i++ {
		wilds = append(wilds, hoennWilds[i])
	}
	for i := 0; i < t.SinnohEncounterLevel; i++ {
		wilds = append(wilds, sinnohWilds[i])
	}
	for i := 0; i < t.UnovaEncounterLevel; i++ {
		wilds = append(wilds, unovaWilds[i])
	}
	for i := 0; i < t.KalosEncounterLevel; i++ {
		wilds = append(wilds, kalosWilds[i])
	}

	return wilds
}

// RandomWildPokemon picks a random wild Pokemon from the list of
// possibilities by first picking a region and random, then picking a Pokemon
// from that region based on the probablities of each.
func RandomWildPokemon(wilds [][]WildEntry) WildEntry {
	// Seed the random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Randomly decide which region the Pokemon will be from
	region := rng.Intn(len(wilds))
	regionWilds := wilds[region]

	// Sum up the full probability of all Pokemon in the region
	var probabilitySum int
	for _, wild := range regionWilds {
		probabilitySum += wild.Probability
	}

	// Randomly decide the wild Pokemon
	num := rng.Intn(probabilitySum)
	var currProbability int
	for _, wild := range regionWilds {
		currProbability += wild.Probability
		if currProbability >= num {
			return wild
		}
	}

	panic("A random wild Pokemon was not chosen!")
}

// CatchRate returns the catch rate of the Pokemon.
func CatchRate(p Pokemon, pBI PokemonBattleInfo) float64 {
	hp := float64(CalcOOBHP(p.HP, p))
	var bonus float64
	switch pBI.Ailment {
	default:
		bonus = 1.0
	case SleepAilment, FreezeAilment:
		bonus = 2.0
	case ParalysisAilment, PoisonAilment, BurnAilment:
		bonus = 1.5
	}
	return (((3.0*hp - (2.0 * float64(pBI.CurrHP))) * float64(p.CatchRate)) / (3.0 * hp)) * bonus
}
