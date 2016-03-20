package app

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"strconv"
	"strings"

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

// fetchMove wraps around the FetchMove method provided by the PokeAPI
// package and provides in-memory caching. If the move is not cached, it
// will ask PokeAPI to request the data. If either that request or the cache
// operation fails, an error is returned.
func fetchMove(id int, client *http.Client, ctx context.Context) (pokeapi.Move, error) {
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
		enc.Encode(move)

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
	dec.Decode(&move)
	return move, nil

}

// fetchPokemon wraps around the FetchPokemon method provided by the PokeAPI
// package and provides in-memory caching. If the Pokemon is not cached, it
// will ask PokeAPI to request the data. If either that request or the cache
// operation fails, an error is returned.
func fetchPokemon(id int, client *http.Client, ctx context.Context) (pokeapi.Pokemon, error) {
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
		enc.Encode(pkmn)

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
	dec.Decode(&pkmn)
	return pkmn, nil
}
