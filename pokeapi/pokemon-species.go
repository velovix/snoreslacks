package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
	"github.com/velovix/snoreslacks/messaging"
)

type PokemonSpecies struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CaptureRate int    `json:"capture_rate"`
	GrowthRate  struct {
		Name string `json:"name"`
	} `json:"growth_rate"`
}

// FetchPokemonSpecies queries PokeAPI directly for a Pokemon species with
// the given id. This function should be avoided in favor of using a Fetcher.
func FetchPokemonSpecies(id int, client messaging.Client) (PokemonSpecies, error) {
	// Query the API
	resp, err := client.Get(apiURL + pokemonSpeciesEP + strconv.Itoa(id) + "/")
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "fetching a Pokemon species")
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "reading Pokemon species data")
	}

	// Unmarshal the response into a Pokemon species object
	var p PokemonSpecies
	err = json.Unmarshal(data, &p)
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "parsing Pokemon species data")
	}

	return p, nil
}

// FetchPokemonSpecies fetches species information from the given URL using the
// given client. This function should be avoided in favor of using a Fetcher.
func FetchPokemonSpeciesFromURL(url string, client messaging.Client) (PokemonSpecies, error) {
	// Query the API
	resp, err := client.Get(url)
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "fetching a Pokemon species")
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "reading Pokemon species data")
	}

	// Unmarshal the response into a Pokemon species object
	var p PokemonSpecies
	err = json.Unmarshal(data, &p)
	if err != nil {
		return PokemonSpecies{}, errors.Wrap(err, "parsing Pokemon species data")
	}

	return p, nil
}
