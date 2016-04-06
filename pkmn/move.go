package pkmn

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
	StatChanges     []struct {
		Change int
		Stat   StatType
	}
}
