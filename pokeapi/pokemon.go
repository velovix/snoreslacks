package pokeapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
// executes FetchPokemonAsBytes and MakePokemonFromBytes as one operation.
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
	resp, err := client.Get(apiUrl + pokemonEP + strconv.Itoa(id) + "/")
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
