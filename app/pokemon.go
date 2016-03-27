package app

import (
	"errors"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/pokeapi"
)

type pokemon struct {
	ID     int
	Name   string
	Height int
	Weight int
	Type1  string
	Type2  string

	Level int

	HP        stat
	Attack    stat
	Defense   stat
	SpAttack  stat
	SpDefense stat
	Speed     stat

	moves []move `datastore:"-"`

	Move1 int
	Move2 int
	Move3 int
	Move4 int

	Slot int

	// Pokemon will probably be stored anonymously, whatever that may meen for the
	// database implementation. A unique identifier should be stored here when the
	// Pokemon is loaded so changes to the Pokemon can be saved over the old
	// instance.
	UID string `datastore:"-"`
}

// newPokemon creates a new Pokemon from the given PokeAPI Pokemon data.
func newPokemon(apiPkmn pokeapi.Pokemon, client *http.Client, ctx context.Context) (pokemon, error) {
	var pkmn pokemon

	pkmn.ID = apiPkmn.ID
	pkmn.Name = apiPkmn.Name
	pkmn.Height = apiPkmn.Height
	pkmn.Weight = apiPkmn.Weight
	// Fill in the type values
	for _, val := range apiPkmn.Types {
		if val.Slot != 1 && val.Slot != 2 {
			return pokemon{}, errors.New("PokeAPI set a type as slot " + strconv.Itoa(val.Slot) + ", which is invalid")
		}
		if _, ok := types[val.Type.Name]; ok {
			if val.Slot == 1 {
				pkmn.Type1 = val.Type.Name
			} else if val.Slot == 2 {
				pkmn.Type2 = val.Type.Name
			}
		} else {
			return pokemon{}, errors.New("PokeAPI returned an unknown type '" + val.Type.Name + "'")
		}
	}

	// Fill in the stat values
	for _, val := range apiPkmn.Stats {
		switch val.Stat.Name {
		case "hp":
			pkmn.HP = stat{Base: val.BaseStat}
		case "attack":
			pkmn.Attack = stat{Base: val.BaseStat}
		case "defense":
			pkmn.Defense = stat{Base: val.BaseStat}
		case "special-attack":
			pkmn.SpAttack = stat{Base: val.BaseStat}
		case "special-defense":
			pkmn.SpDefense = stat{Base: val.BaseStat}
		case "speed":
			pkmn.Speed = stat{Base: val.BaseStat}
		default:
			return pokemon{}, errors.New("PokeAPI returned an unknown stat '" + val.Stat.Name + "'")
		}
	}

	pkmn.Level = 5

	// Learn any necessary moves
	pkmn.moves = make([]move, 4)
	moveIDs := make([]int, 4)
	currMoveSlot := 0
	for _, val := range apiPkmn.Moves {
		for _, vgd := range val.VersionGroupDetails {
			if vgd.MoveLearnMethod.Name == "level-up" && vgd.LevelLearnedAt <= pkmn.Level {
				// Get the move ID from the mvoe information
				moveID, err := idFromURL(val.Move.URL)
				if err != nil {
					return pokemon{}, err
				}
				// Fetch the move from PokeAPI
				apiMove, err := fetchMove(moveID, client, ctx)
				if err != nil {
					return pokemon{}, err
				}
				// Create a move from the PokeAPI data
				m, err := newMove(apiMove)
				if err != nil {
					return pokemon{}, err
				}
				// Learn the move
				pkmn.moves[currMoveSlot] = m
				moveIDs[currMoveSlot] = moveID
				currMoveSlot++
				currMoveSlot = currMoveSlot % 4
				break
			}
		}
	}

	// Record the IDs of the moves to be stored in the database
	pkmn.Move1 = moveIDs[0]
	pkmn.Move2 = moveIDs[1]
	pkmn.Move3 = moveIDs[2]
	pkmn.Move4 = moveIDs[3]

	return pkmn, nil
}
