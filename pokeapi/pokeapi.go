package pokeapi

import (
	"strconv"
	"strings"
)

const (
	apiUrl = "http://pokeapi.co/api/v2/"

	pokemonEP = "pokemon/"
	moveEP    = "move/"
)

func idFromURL(url string) (int, error) {
	splitData := strings.Split(url, "/")

	return strconv.Atoi(splitData[len(splitData)-2])
}
