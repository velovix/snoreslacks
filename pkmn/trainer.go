package pkmn

const MaxPartySize = 6

type TrainerMode int

const (
	_ TrainerMode = iota
	StarterTrainerMode
	WaitingTrainerMode
	BattlingTrainerMode
)

type Trainer struct {
	Name string
	Mode TrainerMode

	Wins   int
	Losses int
}
