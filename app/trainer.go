package app

import "errors"

const MaxPartySize = 6

type trainerMode int

const (
	_ trainerMode = iota
	starterTrainerMode
	waitingTrainerMode
	battlingTrainerMode
)

type trainer struct {
	Name string
	pkmn []pokemon
	Mode trainerMode

	Wins   int
	Losses int

	// The last URL that the trainer can be contacted with. This URL always
	// has the possiblity of being out of date, and should only be used right
	// after being updated.
	LastContactURL string
}

// givePokemon gives the trainer a new Pokemon, or returns an error if the
// trainer has a full party.
func (t *trainer) givePokemon(pkmn pokemon) error {
	destSlot := 1
	// Find an empty slot for the new Pokemon
	for _, val := range t.pkmn {
		if val.Slot == destSlot {
			destSlot++
			if destSlot > MaxPartySize {
				return errors.New("the trainer already has a full party so a new Pokemon cannot be added")
			}
		}
	}
	pkmn.Slot = destSlot

	t.pkmn = append(t.pkmn, pkmn)
	return nil
}
