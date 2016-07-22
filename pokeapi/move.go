package pokeapi

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
)

// Move describes a PokeAPI move.
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
	Target struct {
		Name string `json:"name"`
	} `json:"target"`
	StatChanges []struct {
		Change int `json:"change"`
		Stat   struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stat_changes"`
}

// FetchMove queries PokeAPI directly for the move with the given id. This
// function should be avoided in favor of using a Fetcher.
func FetchMove(id int, client messaging.Client) (Move, error) {
	// Query the API
	resp, err := client.Get(apiURL + moveEP + strconv.Itoa(id) + "/")
	if err != nil {
		return Move{}, errors.Wrap(err, "fetching a move")
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Move{}, errors.Wrap(err, "reading move data")
	}

	// Unmarshal the response into a Move object
	var p Move
	err = json.Unmarshal(data, &p)
	if err != nil {
		return Move{}, errors.Wrap(err, "parsing move data")
	}

	return p, nil
}

// FetchMoveFromURL queries the given URL for a move and returns the result.
// This function should be avoided in favor of using a Fetcher.
func FetchMoveFromURL(url string, client messaging.Client) (Move, error) {
	// Query the API
	resp, err := client.Get(url)
	if err != nil {
		return Move{}, errors.Wrap(err, "fetching a move")
	}
	defer resp.Body.Close()

	// Read the response data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Move{}, errors.Wrap(err, "reading move data")
	}

	// Unmarshal the response into a Move object
	var p Move
	err = json.Unmarshal(data, &p)
	if err != nil {
		return Move{}, errors.Wrap(err, "parsing move data")
	}

	return p, nil
}

// NewMove creates a new move from the PokeAPI move data.
func NewMove(apiMove Move) (pkmn.Move, error) {
	var m pkmn.Move

	m.ID = apiMove.ID
	m.Name = apiMove.Name
	m.Accuracy = apiMove.Accuracy
	m.EffectChance = apiMove.EffectChance
	m.PP = apiMove.PP
	m.Priority = apiMove.Priority
	m.Power = apiMove.Power
	// Assign the damange class
	switch apiMove.DamageClass.Name {
	case "physical":
		m.DamageClass = pkmn.PhysicalDamageClass
	case "status":
		m.DamageClass = pkmn.StatusDamageClass
	case "special":
		m.DamageClass = pkmn.SpecialDamageClass
	default:
		return pkmn.Move{}, errors.New("unsupported damage class '" + apiMove.DamageClass.Name + "'")
	}
	// Assign the english effect entry
	for _, entr := range apiMove.EffectEntries {
		if entr.Language.Name == "en" {
			m.EffectEntry = entr.Effect
			break
		}
	}
	// Assign the ailment
	switch apiMove.Meta.Ailment.Name {
	case "none":
		m.Ailment = pkmn.NoAilment
	case "paralysis":
		m.Ailment = pkmn.ParalysisAilment
	case "poison":
		m.Ailment = pkmn.PoisonAilment
	case "freeze":
		m.Ailment = pkmn.FreezeAilment
	case "burn":
		m.Ailment = pkmn.BurnAilment
	case "sleep":
		m.Ailment = pkmn.SleepAilment
	case "confusion":
		m.Ailment = pkmn.ConfusionAilment
	default:
		// Not all ailments are supported at this time
		m.Ailment = pkmn.NoAilment
	}
	// Check if the move hits multiple times
	m.HasMultipleHits = apiMove.Meta.MinHits != nil
	if m.HasMultipleHits {
		m.MinHits = *apiMove.Meta.MinHits
		m.MaxHits = *apiMove.Meta.MaxHits
	}
	m.Drain = apiMove.Meta.Drain
	m.Healing = apiMove.Meta.Healing
	m.CritRate = apiMove.Meta.CritRate
	m.AilmentChance = apiMove.Meta.AilmentChance
	m.FlinchChance = apiMove.Meta.FlinchChance
	m.StatChance = apiMove.Meta.StatChance
	m.Type = apiMove.Type.Name
	// Assign the target
	if apiMove.Target.Name == "user" {
		m.Target = pkmn.SelfMoveTarget
	} else {
		m.Target = pkmn.EnemyMoveTarget
	}
	// Assign the potential stat changes
	m.StatChanges = make([]struct {
		Change int
		Stat   pkmn.StatType
	}, 0)
	for _, val := range apiMove.StatChanges {
		change := val.Change
		var stat pkmn.StatType
		switch val.Stat.Name {
		case "attack":
			stat = pkmn.AttackStatType
		case "defense":
			stat = pkmn.DefenseStatType
		case "special-attack":
			stat = pkmn.SpecialAttackStatType
		case "special-defense":
			stat = pkmn.SpecialDefenseStatType
		case "speed":
			stat = pkmn.SpeedStatType
		case "accuracy":
			stat = pkmn.AccuracyStatType
		case "evasion":
			stat = pkmn.EvasionStatType
		default:
			return pkmn.Move{}, errors.New("unsupported stat type '" + val.Stat.Name + "'")
		}

		m.StatChanges = append(m.StatChanges, struct {
			Change int
			Stat   pkmn.StatType
		}{Change: change, Stat: stat})
	}

	return m, nil
}
