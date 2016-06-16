package gaepokeapi

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"strings"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pokeapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// idFromURL attempts to extract an ID from a PokeAPI resource URL.
// PokeAPI URLs should be in the form "http://pokeapi.co/api/v2/ability/34/",
// so the ID should be the second to last element if the string is split
// by slashes.
func idFromURL(url string) (int, error) {
	betweenSlashes := strings.Split(url, "/")
	idStr := betweenSlashes[len(betweenSlashes)-2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GAEFetcher is the App Engine implementation of a PokeAPI fetcher.
type GAEFetcher struct{}

// FetchMove wraps around the FetchMove method provided by the PokeAPI
// package and provides in-memory caching. If the move is not cached, it
// will ask PokeAPI to request the data. If either that request or the cache
// operation fails, an error is returned.
func (f GAEFetcher) FetchMove(ctx context.Context, client messaging.Client, id int) (pokeapi.Move, error) {
	cacheKey := "pokeapi.move." + strconv.Itoa(id)

	// Try the cache for the move
	item, err := memcache.Get(ctx, "pokeapi.move."+strconv.Itoa(id))

	if err == memcache.ErrCacheMiss {
		// The move is not in the cache, so we have to ask PokeAPI

		move, err := pokeapi.FetchMove(id, client)
		if err != nil {
			return pokeapi.Move{}, err
		}
		// Encode the move structure as a gob
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		err = enc.Encode(move)
		if err != nil {
			return pokeapi.Move{}, err
		}

		// Add the data to the cache
		cacheItem := &memcache.Item{
			Key:   cacheKey,
			Value: buf.Bytes()}
		if err := memcache.Add(ctx, cacheItem); err == memcache.ErrNotStored {
			// Another request may have beaten us to the punch on caching this item. Not a big deal
			log.Infof(ctx, "attempted to cache %s when it already exists", cacheKey)
		} else if err != nil {
			return pokeapi.Move{}, err
		}

		return move, nil
	} else if err != nil {
		// Some miscellaneous cache error occurred
		return pokeapi.Move{}, err
	}

	// The move is in the cache. Decode it from a gob
	var move pokeapi.Move
	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&move)
	if err != nil {
		return pokeapi.Move{}, err
	}
	return move, nil

}

// FetchPokemon wraps around the FetchPokemon method provided by the PokeAPI
// package and provides in-memory caching. If the Pokemon is not cached, it
// will ask PokeAPI to request the data. If either that request or the cache
// operation fails, an error is returned.
func (f GAEFetcher) FetchPokemon(ctx context.Context, client messaging.Client, id int) (pokeapi.Pokemon, error) {
	cacheKey := "pokeapi.pokemon." + strconv.Itoa(id)

	// Try the cache for the Pokemon
	item, err := memcache.Get(ctx, "pokeapi.pokemon."+strconv.Itoa(id))

	if err == memcache.ErrCacheMiss {
		// The Pokemon is not in the cache, so we have to ask PokeAPI

		pkmn, err := pokeapi.FetchPokemon(id, client)
		if err != nil {
			return pokeapi.Pokemon{}, err
		}
		// Encode the Pokemon structure as a gob
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		err = enc.Encode(pkmn)
		if err != nil {
			return pokeapi.Pokemon{}, err
		}

		// Add the data to the cache
		cacheItem := &memcache.Item{
			Key:   cacheKey,
			Value: buf.Bytes()}
		if err := memcache.Add(ctx, cacheItem); err == memcache.ErrNotStored {
			// Another request may have beaten us to the punch on caching this item. Not a big deal
			log.Infof(ctx, "attempted to cache %s when it already exists", cacheKey)
		} else if err != nil {
			return pokeapi.Pokemon{}, err
		}

		return pkmn, nil
	} else if err != nil {
		// Some miscellaneous cache error occurred
		return pokeapi.Pokemon{}, err
	}

	// The Pokemon is in the cache. Decode it from a gob
	var pkmn pokeapi.Pokemon
	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&pkmn)
	if err != nil {
		return pokeapi.Pokemon{}, err
	}
	return pkmn, nil
}

// FetchPokemonSpecies wraps around the FetchPokemonSpecies method provided by
// the PokeAPI package and provides in-memory caching. If the pokemon species
// is not cached, it will ask PokeAPI to request the data. If either that
// request or the cache operation fails, an error is returned.
func (f GAEFetcher) FetchPokemonSpecies(ctx context.Context, client messaging.Client, id int) (pokeapi.PokemonSpecies, error) {
	cacheKey := "pokeapi.pokemonSpecies." + strconv.Itoa(id)

	// Try the cache for the pokemon species
	item, err := memcache.Get(ctx, "pokeapi.pokemonSpecies."+strconv.Itoa(id))

	if err == memcache.ErrCacheMiss {
		// The pokemon species is not in the cache, so we have to ask PokeAPI

		pokemonSpecies, err := pokeapi.FetchPokemonSpecies(id, client)
		if err != nil {
			return pokeapi.PokemonSpecies{}, err
		}
		// Encode the pokemon species structure as a gob
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		err = enc.Encode(pokemonSpecies)
		if err != nil {
			return pokeapi.PokemonSpecies{}, err
		}

		// Add the data to the cache
		cacheItem := &memcache.Item{
			Key:   cacheKey,
			Value: buf.Bytes()}
		if err := memcache.Add(ctx, cacheItem); err == memcache.ErrNotStored {
			// Another request may have beaten us to the punch on caching this item. Not a big deal
			log.Infof(ctx, "attempted to cache %s when it already exists", cacheKey)
		} else if err != nil {
			return pokeapi.PokemonSpecies{}, err
		}

		return pokemonSpecies, nil
	} else if err != nil {
		// Some miscellaneous cache error occurred
		return pokeapi.PokemonSpecies{}, err
	}

	// The pokemonSpecies is in the cache. Decode it from a gob
	var pokemonSpecies pokeapi.PokemonSpecies
	buf := bytes.NewBuffer(item.Value)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&pokemonSpecies)
	if err != nil {
		return pokeapi.PokemonSpecies{}, err
	}
	return pokemonSpecies, nil
}

func init() {
	pokeapi.Register("gae", GAEFetcher{})
}
