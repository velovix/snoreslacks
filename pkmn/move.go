package pkmn

import (
	"log"
	"math/rand"

	"github.com/pkg/errors"
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

type MoveReport struct {
	Missed           bool
	TargetsUser      bool
	UserHealing      int
	UserFainted      bool
	TargetDamage     int
	TargetDrain      int
	TargetFainted    bool
	AttStageChange   int
	DefStageChange   int
	SpAttStageChange int
	SpDefStageChange int
	SpeedStageChange int
	CriticalHit      bool
	Effectiveness    int
	Poisoned         bool
	Paralyzed        bool
	Asleep           bool
	Frozen           bool
	Burned           bool
}

// CalcMoveOrder calculates which move should go first based on the move
// itself and the user of the move. The function returns 1 if Pokemon 1 goes
// first, or 2 if Pokemon 2 goes first.
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

// critChance returns the critical hit chance in percentage based on the given
// crit rate.
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

// calcDamage calculates the damage the target will take if the user uses the
// given move. It returns the damage given, the type effectiveness (positive if
// super effective, negative if not very effective, zero if regular), true if it
// was a critical hit, and potentially an error.
func calcDamage(user, target *Pokemon, userBI, targetBI *PokemonBattleInfo, move Move) (int, int, bool, error) {
	// Convert target type names to type objects
	targetType1, ok := NameToType(target.Type1)
	if !ok {
		return 0, 0, false, errors.New("no type found for type name '" + target.Type1 + "'")
	}
	var targetType2 Type
	if target.Type2 != "" {
		targetType2, ok = NameToType(target.Type2)
		if !ok {
			return 0, 0, false, errors.New("no type found for type name '" + target.Type2 + "'")
		}
	}

	log.Printf("Target types: %+v %+v", targetType1, targetType2)

	// Calculate same type attack bonus
	stab := 1.0
	if user.Type1 == move.Type || user.Type2 == move.Type {
		stab = 1.5
	}
	log.Printf("Stab: %v", stab)
	// Calculate type effectiveness
	typeEff := 1.0
	typeEff *= targetType1.Mod(move.Type)
	log.Printf("Type 1 mod: %v", typeEff)
	if target.Type2 != "" {
		typeEff *= targetType2.Mod(move.Type)
		log.Printf("Type 2 mod: %v", targetType2.Mod(move.Type))
	}
	// Calculate critical hit effectiveness
	crit := 1.0
	if float64(rand.Intn(100)+1) <= critChance(move.CritRate) {
		crit = 1.5
	}
	log.Printf("Critical hit: %v", crit)
	// Calculate the random number
	random := float64(rand.Intn(15)+85) / 100.0
	log.Printf("Random: %v", crit)

	// Calculate the modifier
	modifier := stab * typeEff * crit * random
	log.Printf("Modifier: %v", modifier)

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
	log.Printf("Att: %v", att)
	log.Printf("Def: %v", def)

	// Calculate the damage
	return int(((((2.0*float64(user.Level)+10)/250.0)*(att/def))*float64(move.Power) + 2.0) * modifier), int(typeEff), (crit > 1.0), nil
}

// RunMove uses the move on the target.
func RunMove(user, target *Pokemon, userBI, targetBI *PokemonBattleInfo, move Move) (MoveReport, error) {
	var mr MoveReport

	if rand.Intn(100)+1 > move.Accuracy*(CalcIBAccuracy(*user, *userBI)/CalcIBEvasion(*target, *targetBI)) {
		// The move missed, so we have nothing to do
		mr.Missed = true
		return mr, nil
	}

	// Check if the move targets the user
	if move.Target == SelfMoveTarget {
		// The target is directed at the user, so we'll change the target to
		// the user
		target = user
		targetBI = userBI
		mr.TargetsUser = true
	}

	// Check if the move heals
	if move.Healing > 0 {
		// The move heals the user by the given percent
		maxHP := CalcIBHP(*user, *userBI) * move.Healing
		userBI.CurrHP += maxHP * move.Healing
		if userBI.CurrHP > maxHP {
			userBI.CurrHP = maxHP
		}
		mr.UserHealing = maxHP * move.Healing
	}

	// Check what kind of damage class the move is in
	if move.DamageClass == StatusDamageClass {
		// Apply all status effects onto the target
		for _, statChange := range move.StatChanges {
			switch statChange.Stat {
			case AttackStatType:
				targetBI.AttStage += statChange.Change
				mr.AttStageChange = statChange.Change
			case DefenseStatType:
				targetBI.DefStage += statChange.Change
				mr.DefStageChange = statChange.Change
			case SpecialAttackStatType:
				targetBI.SpAttStage += statChange.Change
				mr.SpAttStageChange = statChange.Change
			case SpecialDefenseStatType:
				targetBI.SpDefStage += statChange.Change
				mr.SpDefStageChange = statChange.Change
			case SpeedStatType:
				targetBI.SpeedStage += statChange.Change
				mr.SpeedStageChange = statChange.Change
			}
		}

		// We don't care about type effectiveness for status moves
		mr.Effectiveness = 1.0
	} else {
		// Calculate the damage done by the move
		damage, effectiveness, crit, err := calcDamage(user, target, userBI, targetBI, move)
		if err != nil {
			return MoveReport{}, err
		}

		// Deal the damage
		targetBI.CurrHP -= damage
		if targetBI.CurrHP < 0 {
			targetBI.CurrHP = 0
			mr.TargetFainted = true // The move made the opponent faint
		}
		mr.TargetDamage = damage
		mr.Effectiveness = effectiveness
		mr.CriticalHit = crit

		// Check if the move has HP drain or knockback
		if move.Drain > 0 {
			// The move heals or hurts the user by a percent of the damage done
			userBI.CurrHP += damage * move.Drain
			maxHP := CalcIBHP(*user, *userBI) * move.Healing
			if userBI.CurrHP > maxHP {
				userBI.CurrHP = maxHP
			} else if userBI.CurrHP < 0 {
				userBI.CurrHP = 0
				mr.UserFainted = true // Knockback made the user faint
			}
			mr.TargetDrain = damage * move.Drain
		}
	}

	// Check if the move has an ailment effect
	if move.Ailment != NoAilment {
		// Attempt to inflict an ailment on the target

		// Check if the ailment "hit"
		if rand.Intn(100)+1 <= move.EffectChance {
			// Inflict the ailment so long as the target isn't already
			// suffering from an ailment
			if targetBI.Ailment == NoAilment {
				targetBI.Ailment = move.Ailment
				switch targetBI.Ailment {
				case PoisonAilment:
					mr.Poisoned = true
				case ParalysisAilment:
					mr.Paralyzed = true
				case SleepAilment:
					mr.Asleep = true
				case FreezeAilment:
					mr.Frozen = true
				case BurnAilment:
					mr.Burned = true
				}
			}
		}
	}

	return mr, nil
}
