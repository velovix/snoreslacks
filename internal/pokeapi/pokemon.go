package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Pokemon contains the response data from PokeAPI regarding a Pokemon.
type Pokemon struct {
	ID     *int    `json:"id"`
	Name   *string `json:"name"`
	Height *int    `json:"height"`
	Weight *int    `json:"weight"`
	Types  []*struct {
		Slot *int `json:"slot"`
		Type struct {
			Name *string `json:"name"`
			URL  *string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Stats []*struct {
		BaseStat *int `json:"base_stat"`
		Effort   *int `json:"effort"`
		Stat     *struct {
			Name *string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Moves []*struct {
		Move *struct {
			Name *string `json:"name"`
			URL  *string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []*struct {
			LevelLearnedAt  *int `json:"level_learned_at"`
			MoveLearnMethod *struct {
				Name *string `json:"name"`
			} `json:"move_learn_method"`
		} `json:"version_group_details"`
	} `json:"moves"`
}

// FetchPokemon queries PokeAPI for the Pokemon with the given id.
func FetchPokemon(id int) (Pokemon, error) {
	// Query the API
	resp, err := http.Get(apiUrl + pokemonEP + strconv.Itoa(id) + "/")
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Pokemon{}, err
	}

	// Unmarshal the response into a Pokemon object
	var p Pokemon
	err = json.Unmarshal(data, &p)
	if err != nil {
		return Pokemon{}, err
	}

	return p, nil
}
