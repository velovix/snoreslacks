package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Move struct {
	ID           *int    `json:"id"`
	Name         *string `json:"name"`
	Accuracy     *int    `json:"accuracy"`
	EffectChance *int    `json:"effect_chance"`
	PP           *int    `json:"pp"`
	Priority     *int    `json:"priority"`
	Power        *int    `json:"power"`
	DamageClass  *struct {
		Name string `json:"name"`
	} `json:"damage_class"`
	EffectEntries []*struct {
		Effect   *string `json:"effect"`
		Language []*struct {
			Name string `json:"name"`
		} `json:"language"`
	} `json:"effect_entires"`
}

// FetchMove queries PokeAPI for the move with the given id.
func FetchMove(id int) (Move, error) {
	// Query the API
	resp, err := http.Get(apiUrl + moveEP + strconv.Itoa(id) + "/")
	if err != nil {
		return Move{}, err
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Move{}, err
	}

	// Unmarshal the response into a Move object
	var p Move
	err = json.Unmarshal(data, &p)
	if err != nil {
		return Move{}, err
	}

	return p, nil
}
