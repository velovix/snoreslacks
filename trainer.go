package snoreslacks

type trainerMode int

const (
	_ trainerMode = iota
	starterTrainerMode
	waitingTrainerMode
)

type trainer struct {
	name string
	pkmn []pokemon
	mode trainerMode
}
