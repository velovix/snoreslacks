package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Move struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Accuracy     int    `json:"accuracy"`
	EffectChance int    `json:"effect_chance"`
	PP           int    `json:"pp"`
	Priority     int    `json:"priority"`
	Power        int    `json:"power"`
	DamageClass  struct {
		Name string `json:"name"`
	} `json:"damage_class"`
	EffectEntries []struct {
		Effect   string `json:"effect"`
		Language struct {
			Name string `json:"name"`
		} `json:"language"`
	} `json:"effect_entires"`
	Meta struct {
		Ailment struct {
			Name string `json:"name"`
		} `json:"ailment"`
		Category struct {
			Name string `json:"name"`
		}
		MinHits       *int `json:"min_hits"`
		MaxHits       *int `json:"max_hits"`
		MinTurns      *int `json:"min_turns"`
		MaxTurns      *int `json:"max_turns"`
		Drain         int  `json:"drain"`
		Healing       int  `json:"healing"`
		CritRate      int  `json:"crit_rate"`
		AilmentChance int  `json:"ailment_chance"`
		FlinchChance  int  `json:"flinch_chance"`
		StatChance    int  `json:"stat_chance"`
	} `json:"meta"`
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
	StatChanges []struct {
		Change int `json:"change"`
		Stat   struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stat_changes"`
}

// FetchMove queries PokeAPI for the move with the given id.
func FetchMove(id int, client *http.Client) (Move, error) {
	// Query the API
	resp, err := client.Get(apiUrl + moveEP + strconv.Itoa(id) + "/")
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

// FetchMoveFromURL queries the given URl for a move and returns the result.
func FetchMoveFromURL(url string, client *http.Client) (Move, error) {
	// Query the API
	resp, err := client.Get(url)
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
