package app

import (
	"errors"

	"github.com/velovix/snoreslacks/pokeapi"
)

type damageClass int

const (
	_ damageClass = iota
	statusDamageClass
	physicalDamageClass
	specialDamageClass
)

type ailment int

const (
	noAilment ailment = iota
	paralysisAilment
	poisonAilment
	freezeAilment
	burnAilment
	sleepAilment
	confusionAilment
)

type statType int

const (
	_ statType = iota
	attackStatType
	defenseStatType
	specialAttackStatType
	specialDefenseStatType
	speedStatType
	evasionStatType
	accuracyStatType
)

type move struct {
	ID              int
	Name            string
	Accuracy        int
	EffectChance    int
	PP              int
	Priority        int
	Power           int
	DamageClass     damageClass
	EffectEntry     string
	Ailment         ailment
	HasMultipleHits bool
	MinHits         int
	MaxHits         int
	Drain           int
	Healing         int
	CritRate        int
	AilmentChance   int
	FlinchChance    int
	StatChance      int
	Type            string
	StatChanges     []struct {
		Change int
		Stat   statType
	}
}

// newMove creates a new move from the PokeAPI move data.
func newMove(apiMove pokeapi.Move) (move, error) {
	var m move

	m.ID = apiMove.ID
	m.Name = apiMove.Name
	m.Accuracy = apiMove.Accuracy
	m.EffectChance = apiMove.EffectChance
	m.PP = apiMove.PP
	m.Priority = apiMove.Priority
	m.Power = apiMove.Power
	// Assign the damange class
	switch apiMove.DamageClass.Name {
	case "physical":
		m.DamageClass = physicalDamageClass
	case "status":
		m.DamageClass = statusDamageClass
	case "special":
		m.DamageClass = specialDamageClass
	default:
		return move{}, errors.New("unsupported damage class '" + apiMove.DamageClass.Name + "'")
	}
	// Assign the english effect entry
	for _, entr := range apiMove.EffectEntries {
		if entr.Language.Name == "en" {
			m.EffectEntry = entr.Effect
			break
		}
	}
	// Assign the ailment
	switch apiMove.Meta.Ailment.Name {
	case "none":
		m.Ailment = noAilment
	case "paralysis":
		m.Ailment = paralysisAilment
	case "poison":
		m.Ailment = poisonAilment
	case "freeze":
		m.Ailment = freezeAilment
	case "burn":
		m.Ailment = burnAilment
	case "sleep":
		m.Ailment = sleepAilment
	case "confusion":
		m.Ailment = confusionAilment
	default:
		return move{}, errors.New("unsupported ailment '" + apiMove.Meta.Ailment.Name + "'")
	}
	// Check if the move hits multiple times
	m.HasMultipleHits = apiMove.Meta.MinHits != nil
	if m.HasMultipleHits {
		m.MinHits = *apiMove.Meta.MinHits
		m.MaxHits = *apiMove.Meta.MaxHits
	}
	m.Drain = apiMove.Meta.Drain
	m.Healing = apiMove.Meta.Healing
	m.CritRate = apiMove.Meta.CritRate
	m.AilmentChance = apiMove.Meta.AilmentChance
	m.FlinchChance = apiMove.Meta.FlinchChance
	m.StatChance = apiMove.Meta.StatChance
	m.Type = apiMove.Type.Name
	// Assign the potential stat changes
	m.StatChanges = make([]struct {
		Change int
		Stat   statType
	}, 0)
	for _, val := range apiMove.StatChanges {
		change := val.Change
		var stat statType
		switch val.Stat.Name {
		case "attack":
			stat = attackStatType
		case "defense":
			stat = defenseStatType
		case "special-attack":
			stat = specialAttackStatType
		case "special-defense":
			stat = specialDefenseStatType
		case "speed":
			stat = speedStatType
		case "accuracy":
			stat = accuracyStatType
		case "evasion":
			stat = evasionStatType
		default:
			return move{}, errors.New("unsupported stat type '" + val.Stat.Name + "'")
		}

		m.StatChanges = append(m.StatChanges, struct {
			Change int
			Stat   statType
		}{Change: change, Stat: stat})
	}

	return m, nil
}
