package pokeapi

import (
	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/messaging"

	"golang.org/x/net/context"
)

// Fetcher describes an object that fetches PokeAPI data in some way.
// The object should implement some kind of caching. Users should expect that
// any method call may take any amount of time and may fail.
type Fetcher interface {
	FetchPokemon(ctx context.Context, client messaging.Client, id int) (Pokemon, error)
	FetchMove(ctx context.Context, client messaging.Client, id int) (Move, error)
	FetchPokemonSpecies(ctx context.Context, client messaging.Client, id int) (PokemonSpecies, error)
}

var registered map[string]Fetcher

func init() {
	registered = make(map[string]Fetcher)
}

// Register registers an implementation of the Fetcher interface.
func Register(name string, fetcher Fetcher) {
	registered[name] = fetcher
}

// Get returns an implementation of the Fetcher with the given name of an error
// if no such implementation exists.
func Get(name string) (Fetcher, error) {
	if f, ok := registered[name]; ok {
		return f, nil
	}

	return nil, errors.New("no Fetcher implementation called '" + name + "' found")
}
