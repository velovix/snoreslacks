package pkmn

import (
	"math/rand"

	"github.com/pkg/errors"
)

const MaxPartySize = 6

type TrainerMode int

const (
	_ TrainerMode = iota
	StarterTrainerMode
	WaitingTrainerMode
	BattlingTrainerMode
	ForgetMoveTrainerMode
)

// TrainerType represents different classes of trainers.
type TrainerType int

const (
	_ TrainerType = iota
	// HumanTrainerType is a human user.
	HumanTrainerType
	// WildTrainerType is a stand-in trainer directing a single wild Pokemon.
	WildTrainerType
	// GymLeaderTrainerType is a gym leader.
	GymLeaderTrainerType
)

// Trainer is a named person (human or otherwise) that owns Pokemon and can
// engage in battles and other activities.
type Trainer struct {
	UUID string
	Name string
	Mode TrainerMode
	Type TrainerType

	KantoBadges         int
	KantoEncounterLevel int

	JohtoBadges         int
	JohtoEncounterLevel int

	HoennBadges         int
	HoennEncounterLevel int

	SinnohBadges         int
	SinnohEncounterLevel int

	UnovaBadges         int
	UnovaEncounterLevel int

	KalosBadges         int
	KalosEncounterLevel int

	Wins   int
	Losses int
}

func (t *Trainer) PickMove(moves []int, moveCnt int) (int, error) {
	switch t.Type {
	case HumanTrainerType:
		// Human trainers can't have their moves picked automatically
		return -1, errors.New("attempt to have a human trainer's move picked automatically")
	case WildTrainerType:
		// Wild Pokemon get their moves picked randomly
		return moves[rand.Intn(moveCnt)], nil
	case GymLeaderTrainerType:
		// Gym leaders pick their Pokemon's moves randomly. This will probably
		// get more inspiring in the future
		return moves[rand.Intn(moveCnt)], nil
	default:
		panic("unimplemented trainer type")
	}
}
