package pokeapi

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	apiURL = "http://pokeapi.co/api/v2/"

	pokemonEP        = "pokemon/"
	moveEP           = "move/"
	pokemonSpeciesEP = "pokemon-species/"
)

func idFromURL(url string) (int, error) {
	splitData := strings.Split(url, "/")

	id, err := strconv.Atoi(splitData[len(splitData)-2])
	if err != nil {
		return -1, errors.Wrap(err, "parsing ID from URL")
	}

	return id, nil
}
