package pkmn

import (
	"errors"
	"math/rand"
)

type DamageClass int

const (
	_ DamageClass = iota
	StatusDamageClass
	PhysicalDamageClass
	SpecialDamageClass
)

type Ailment int

const (
	NoAilment Ailment = iota
	ParalysisAilment
	PoisonAilment
	FreezeAilment
	BurnAilment
	SleepAilment
	ConfusionAilment
)

type StatType int

const (
	_ StatType = iota
	AttackStatType
	DefenseStatType
	SpecialAttackStatType
	SpecialDefenseStatType
	SpeedStatType
	EvasionStatType
	AccuracyStatType
)

type MoveTarget int

const (
	_ MoveTarget = iota
	SelfMoveTarget
	EnemyMoveTarget
)

// Move represents a Pokemon move.
type Move struct {
	ID              int
	Name            string
	Accuracy        int
	EffectChance    int
	PP              int
	Priority        int
	Power           int
	DamageClass     DamageClass
	EffectEntry     string
	Ailment         Ailment
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
	Target          MoveTarget
	StatChanges     []struct {
		Change int
		Stat   StatType
	}
}

func CalcMoveOrder(pkmn1, pkmn2 Pokemon, pkmnBI1, pkmnBI2 PokemonBattleInfo, move1, move2 Move) int {
	// Check if the moves have different priority and find move order based
	// on that if possible. Moves with a higher priority go before moves
	// with a lower priority
	if move1.Priority > move2.Priority {
		return 1
	} else if move2.Priority > move1.Priority {
		return 2
	}

	// Check if the Pokemon have different speeds and find the move order
	// based off of that if possible. Pokemon with higher speeds go first.
	if CalcIBSpeed(pkmn1, pkmnBI1) > CalcIBSpeed(pkmn2, pkmnBI2) {
		return 1
	} else if CalcIBSpeed(pkmn1, pkmnBI1) < CalcIBSpeed(pkmn2, pkmnBI2) {
		return 2
	}

	// The Pokemon are an equal match in priority and speed, so chaos will be
	// our guide. Randomly choose move order.
	return rand.Intn(1) + 1
}

func critChance(critRate int) float64 {
	if critRate == 0 {
		return 6.25
	} else if critRate == 1 {
		return 12.5
	} else if critRate == 2 {
		return 50
	} else {
		return 100
	}
}

func calcDamage(user, target *Pokemon, userBI, targetBI *PokemonBattleInfo, move Move) (int, error) {
	// Convert target type names to type objects
	targetType1, ok := NameToType(target.Type1)
	if !ok {
		return 0, errors.New("no type found for type name '" + target.Type1 + "'")
	}
	targetType2, ok := NameToType(target.Type2)
	if !ok {
		return 0, errors.New("no type found for type name '" + target.Type2 + "'")
	}

	// Calculate same type attack bonus
	stab := 1.0
	if user.Type1 == move.Type || user.Type2 == move.Type {
		stab = 1.5
	}
	// Calculate type effectiveness
	typeEff := targetType1.Mod(move.Type) * targetType2.Mod(move.Type)
	// Calculate critical hit effectiveness
	crit := 1.0
	if float64(rand.Intn(100)+1) <= critChance(move.CritRate) {
		crit = 1.5
	}
	// Calculate the random number
	random := float64(rand.Intn(15)+85) / 100.0

	// Calculate the modifier
	modifier := stab * typeEff * crit * random

	// Calculate the user's special or physical attack and the target's
	// physical or special defense
	var att, def float64
	if move.DamageClass == PhysicalDamageClass {
		att = float64(CalcIBAttack(*user, *userBI))
		def = float64(CalcIBDefense(*target, *targetBI))
	} else if move.DamageClass == SpecialDamageClass {
		att = float64(CalcIBSpAtt(*user, *userBI))
		def = float64(CalcIBDefense(*target, *targetBI))
	}

	// Calculate the damage
	return int(((((2.0*float64(user.Level)+10)/250.0)*(att/def))*float64(move.Power) + 2.0) * modifier), nil
}

func RunMove(user, target *Pokemon, userBI, targetBI *PokemonBattleInfo, move Move) error {
	if rand.Intn(100)+1 > move.Accuracy {
		// The move missed, so we have nothing to do
		return nil
	}

	if move.Target == SelfMoveTarget {
		// The target is directed at the user, so we'll change the target to
		// the user
		target = user
		targetBI = userBI
	}

	if move.DamageClass == StatusDamageClass {
		// Apply all status effects onto the target
		for _, statChange := range move.StatChanges {
			switch statChange.Stat {
			case AttackStatType:
				targetBI.AttStage += statChange.Change
			case DefenseStatType:
				targetBI.DefStage += statChange.Change
			case SpecialAttackStatType:
				targetBI.SpAttStage += statChange.Change
			case SpecialDefenseStatType:
				targetBI.SpDefStage += statChange.Change
			case SpeedStatType:
				targetBI.SpeedStage += statChange.Change
			}
		}
	} else {
		// Calculate the damage done by the move
		damage, err := calcDamage(user, target, userBI, targetBI, move)
		if err != nil {
			return err
		}

		// Deal the damage
		targetBI.CurrHP -= damage
		if targetBI.CurrHP < 0 {
			targetBI.CurrHP = 0
		}
	}

	if move.Ailment != NoAilment {
		// Attempt to inflict an ailment on the target

		// Check if the ailment "hit"
		if rand.Intn(100)+1 <= move.EffectChance {
			// Inflict the ailment so long as the target isn't already
			// suffering from an ailment
			if targetBI.Ailment == NoAilment {
				targetBI.Ailment = move.Ailment
			}
		}
	}

	return nil
}
