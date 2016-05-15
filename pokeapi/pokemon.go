package pokeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/satori/go.uuid"
	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// Pokemon contains the response data from PokeAPI regarding a Pokemon.
type Pokemon struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Types  []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Moves []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
			} `json:"move_learn_method"`
		} `json:"version_group_details"`
	} `json:"moves"`
}

// FetchPokemon queries the API for the Pokemon data that corresponds to the
// given ID, then interprets the data as a Pokemon structure. It essentially
// executes FetchPokemonAsBytes and MakePokemonFromBytes as one operation. This
// function should be avoided in favor of using a Fetcher.
func FetchPokemon(id int, client *http.Client) (Pokemon, error) {
	// Get the response data
	data, err := FetchPokemonAsBytes(id, client)
	if err != nil {
		return Pokemon{}, err
	}

	// Interpret the response data
	return MakePokemonFromBytes(data)
}

// FetchPokemonAsBytes queries the API for the Pokemon data that corresponds to
// the given ID. It returns the raw bytes of the response, or an error if the
// request failed.
func FetchPokemonAsBytes(id int, client *http.Client) ([]byte, error) {
	// Query the API
	resp, err := client.Get(apiURL + pokemonEP + strconv.Itoa(id) + "/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("For", id, ":", string(data))

	return data, nil
}

// MakePokemonFromBytes constructs a Pokemon object from the data, or an error
// if the data is not a valid PokeAPI response.
func MakePokemonFromBytes(data []byte) (Pokemon, error) {
	var p Pokemon
	err := json.Unmarshal(data, &p)
	if err != nil {
		return Pokemon{}, err
	}

	return p, nil
}

// NewPokemon creates a new Pokemon from the given PokeAPI Pokemon data.
func NewPokemon(ctx context.Context, client *http.Client, fetcher Fetcher, apiPkmn Pokemon) (pkmn.Pokemon, error) {
	var p pkmn.Pokemon

	uuid := uuid.NewV4()
	p.UUID = uuid.String()

	p.ID = apiPkmn.ID
	p.Name = apiPkmn.Name
	p.Height = apiPkmn.Height
	p.Weight = apiPkmn.Weight
	// Fill in the type values
	for _, val := range apiPkmn.Types {
		if val.Slot != 1 && val.Slot != 2 {
			return pkmn.Pokemon{}, errors.New("PokeAPI set a type as slot " + strconv.Itoa(val.Slot) + ", which is invalid")
		}
	}

	// Fill in the stat values
	for _, val := range apiPkmn.Stats {
		switch val.Stat.Name {
		case "hp":
			p.HP = pkmn.Stat{Base: val.BaseStat}
		case "attack":
			p.Attack = pkmn.Stat{Base: val.BaseStat}
		case "defense":
			p.Defense = pkmn.Stat{Base: val.BaseStat}
		case "special-attack":
			p.SpAttack = pkmn.Stat{Base: val.BaseStat}
		case "special-defense":
			p.SpDefense = pkmn.Stat{Base: val.BaseStat}
		case "speed":
			p.Speed = pkmn.Stat{Base: val.BaseStat}
		default:
			return pkmn.Pokemon{}, errors.New("PokeAPI returned an unknown stat '" + val.Stat.Name + "'")
		}
	}

	// Assign types
	for _, t := range apiPkmn.Types {
		if t.Slot == 1 {
			p.Type1 = t.Type.Name
		} else if t.Slot == 2 {
			p.Type2 = t.Type.Name
		} else {
			return pkmn.Pokemon{}, errors.New("PokeAPI returned an invalid type slot '" + strconv.Itoa(t.Slot) + "'")
		}
	}

	p.Level = 5

	// Learn any necessary moves
	moveIDs := make([]int, 4)
	currMoveSlot := 0
	for _, val := range apiPkmn.Moves {
		for _, vgd := range val.VersionGroupDetails {
			if vgd.MoveLearnMethod.Name == "level-up" && vgd.LevelLearnedAt <= p.Level {
				// Get the move ID from the mvoe information
				moveID, err := idFromURL(val.Move.URL)
				if err != nil {
					return pkmn.Pokemon{}, errors.New("while parsing an ID from a URL: " + err.Error())
				}
				// Learn the move
				moveIDs[currMoveSlot] = moveID
				currMoveSlot++
				currMoveSlot = currMoveSlot % 4
				break
			}
		}
	}

	// Record the IDs of the moves to be stored in the database
	p.Move1 = moveIDs[0]
	p.Move2 = moveIDs[1]
	p.Move3 = moveIDs[2]
	p.Move4 = moveIDs[3]

	return p, nil
}
