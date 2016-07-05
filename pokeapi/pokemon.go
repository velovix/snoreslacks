package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"

	"github.com/satori/go.uuid"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// Pokemon contains the response data from PokeAPI regarding a Pokemon.
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	BaseExperience int    `json:"base_experience"`
	Types          []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Sprites struct {
		FrontDefault string `json:"front_default"`
	} `json:"sprites"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
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
// executes FetchPokemonAsBytes and MakePokemonFromBytes as one operation. This
// function should be avoided in favor of using a Fetcher.
func FetchPokemon(id int, client messaging.Client) (Pokemon, error) {
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
func FetchPokemonAsBytes(id int, client messaging.Client) ([]byte, error) {
	// Query the API
	resp, err := client.Get(apiURL + pokemonEP + strconv.Itoa(id) + "/")
	if err != nil {
		return nil, errors.Wrap(err, "fetching a Pokemon")
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading Pokemon data")
	}

	return data, nil
}

// MakePokemonFromBytes constructs a Pokemon object from the data, or an error
// if the data is not a valid PokeAPI response.
func MakePokemonFromBytes(data []byte) (Pokemon, error) {
	var p Pokemon
	err := json.Unmarshal(data, &p)
	if err != nil {
		return Pokemon{}, errors.Wrap(err, "parsing Pokemon data")
	}

	return p, nil
}

// NewPokemon creates a new Pokemon from the given PokeAPI Pokemon data.
func NewPokemon(ctx context.Context, client messaging.Client, fetcher Fetcher, apiPkmn Pokemon, level int) (pkmn.Pokemon, error) {
	var p pkmn.Pokemon

	uuid := uuid.NewV4()
	p.UUID = uuid.String()

	p.ID = apiPkmn.ID
	p.Name = apiPkmn.Name
	p.Height = apiPkmn.Height
	p.Weight = apiPkmn.Weight
	p.BaseExperience = apiPkmn.BaseExperience
	// Fill in the type values
	for _, val := range apiPkmn.Types {
		if val.Slot != 1 && val.Slot != 2 {
			return pkmn.Pokemon{}, errors.New("unsupported type slot " + strconv.Itoa(val.Slot))
		}
	}

	// Fill in the sprite info
	p.SpriteURL = apiPkmn.Sprites.FrontDefault

	// Fill in the stat values
	for _, val := range apiPkmn.Stats {
		switch val.Stat.Name {
		case "hp":
			p.HP = pkmn.Stat{Base: val.BaseStat}
		case "attack":
			p.Attack = pkmn.Stat{Base: val.BaseStat}
		case "defense":
			p.Defense = pkmn.Stat{Base: val.BaseStat}
		case "special-attack":
			p.SpAttack = pkmn.Stat{Base: val.BaseStat}
		case "special-defense":
			p.SpDefense = pkmn.Stat{Base: val.BaseStat}
		case "speed":
			p.Speed = pkmn.Stat{Base: val.BaseStat}
		default:
			return pkmn.Pokemon{}, errors.New("unsupported stat '" + val.Stat.Name + "'")
		}
	}

	// Assign types
	for _, t := range apiPkmn.Types {
		if t.Slot == 1 {
			p.Type1 = t.Type.Name
		} else if t.Slot == 2 {
			p.Type2 = t.Type.Name
		} else {
			return pkmn.Pokemon{}, errors.New("unsupported type slot '" + strconv.Itoa(t.Slot) + "'")
		}
	}

	// Get species information for the catch rate and growth rate
	speciesID, err := idFromURL(apiPkmn.Species.URL)
	if err != nil {
		return pkmn.Pokemon{}, err
	}
	species, err := fetcher.FetchPokemonSpecies(ctx, client, speciesID)
	if err != nil {
		return pkmn.Pokemon{}, err
	}
	p.CatchRate = species.CaptureRate

	// Fill up the growth rate value
	switch species.GrowthRate.Name {
	case "erratic":
		p.GrowthRate = pkmn.ErraticGrowthRate
	case "fast":
		p.GrowthRate = pkmn.FastGrowthRate
	case "medium":
		p.GrowthRate = pkmn.MediumFastGrowthRate
	case "medium-slow":
		p.GrowthRate = pkmn.MediumSlowGrowthRate
	case "slow":
		p.GrowthRate = pkmn.SlowGrowthRate
	case "fluctuating":
		p.GrowthRate = pkmn.FluctuatingGrowthRate
	default:
		return pkmn.Pokemon{}, errors.New("unsupported growth rate '" + species.GrowthRate.Name + "'")
	}
	p.Level = level
	p.Experience = pkmn.RequiredExperience(p.GrowthRate, p.Level)

	// Learn any necessary moves
	moveIDs := make([]int, 4)
	currMoveSlot := 0
	for _, val := range apiPkmn.Moves {
		for _, vgd := range val.VersionGroupDetails {
			if vgd.MoveLearnMethod.Name == "level-up" && vgd.LevelLearnedAt <= p.Level {
				// Get the move ID from the move information
				moveID, err := idFromURL(val.Move.URL)
				if err != nil {
					return pkmn.Pokemon{}, errors.Wrap(err, "parsing an ID from a URL")
				}
				// Learn the move
				moveIDs[currMoveSlot] = moveID
				currMoveSlot++
				currMoveSlot = currMoveSlot % 4
				break
			}
		}
	}

	// Record the IDs of the moves to be stored in the database
	p.Move1 = moveIDs[0]
	p.Move2 = moveIDs[1]
	p.Move3 = moveIDs[2]
	p.Move4 = moveIDs[3]

	return p, nil
}

// FetchLevelLearnableMoveIDs returns a map containing information on all moves
// that can be learned by the Pokemon with the given ID by leveling up alone.
// The returned map has levels as the key and a slice of moves that can be
// learned at that level as its value.
func FetchLevelLearnableMoveIDs(ctx context.Context, client messaging.Client, fetcher Fetcher, pkmnID int) (map[int][]int, error) {
	// Fetch Pokemon info from PokeAPI to extract move information from
	apiPkmn, err := fetcher.FetchPokemon(ctx, client, pkmnID)
	if err != nil {
		return make(map[int][]int), err
	}

	moveIDs := make(map[int][]int)

	// Loop through all move info
	for _, moveInfo := range apiPkmn.Moves {
		// Some moves can be learned in a multitude of ways, so every way has
		// to be checked
		for _, vgd := range moveInfo.VersionGroupDetails {
			// Only include moves that can be learned by level up
			if vgd.MoveLearnMethod.Name == "level-up" {
				// Extract the move ID
				moveID, err := idFromURL(moveInfo.Move.URL)
				if err != nil {
					return make(map[int][]int), err
				}
				// Add this move ID to the list of learnable moves at the
				// level this move can be learned at
				moveIDs[vgd.LevelLearnedAt] = append(moveIDs[vgd.LevelLearnedAt], moveID)
			}
		}
	}

	return moveIDs, nil
}
