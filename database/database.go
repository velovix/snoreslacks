// Package database contains a specification for a database API for
// Snoreslacks. This includes database object wrappers for each database type
// and a single monoloithic interface for database actions.
//
// Applications should interact with this package directly, not its
// implementations. Implementations should register themselves with this
// package.
package database

import (
	goerrors "errors"

	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// ErrNoResults is returned when a database query that expects an object can't
// find one, like when a call to LoadTrainer cannot find a trainer with the
// given UUID.
var ErrNoResults = goerrors.New("database: no results found")

// IsNoResults returns true if the given error is as a result of there being
// no results from a query.
//
// This can be useful information because, in some circumstances, a missing
// entry might be a recoverable issue.
func IsNoResults(err error) bool {
	return errors.Cause(err) == ErrNoResults
}

// Pokemon describes a database representation of a Pokemon that can emit a
// pkmn.Pokemon.
type Pokemon interface {
	GetPokemon() *pkmn.Pokemon
}

// Trainer describes a database representation of a trainer that can emit a
// pkmn.Trainer.
type Trainer interface {
	GetTrainer() *pkmn.Trainer
}

// Battle describes a database representation of a battle that can emit a
// pkmn.Battle.
type Battle interface {
	GetBattle() *pkmn.Battle
}

// TrainerBattleInfo describes a database representation of a trainer battle
// info that can emit a pkmn.TrainerBattleInfo.
type TrainerBattleInfo interface {
	GetTrainerBattleInfo() *pkmn.TrainerBattleInfo
}

// PokemonBattleInfo describes a database representation of a Pokemon battle
// info that can emit a pkmn.PokemonBattleInfo.
type PokemonBattleInfo interface {
	GetPokemonBattleInfo() *pkmn.PokemonBattleInfo
}

// Database describes an object that is able to save and load Pokemon
// constructs.
type Database interface {
	// NewTrainer creates a database trainer that is ready to be saved from the
	// given pkmn.Trainer.
	NewTrainer(t pkmn.Trainer) Trainer
	// SaveTrainer saves the given Trainer.
	SaveTrainer(ctx context.Context, t Trainer) error
	// LoadTrainer returns a Trainer with the given UUID. The second return
	// value is true if the Trainer exists and was retrieved, false otherwise.
	LoadTrainer(ctx context.Context, uuid string) (Trainer, error)
	// DeleteTrainer deletes the trainer with the given UUID from the database.
	DeleteTrainer(ctx context.Context, uuid string) error
	// PurgeTrainer deletes the trainer with the given UUID and all of their
	// Pokemon from the database.
	PurgeTrainer(ctx context.Context, uuid string) error
	// SaveLastContactURL saves the given last contact URL as associated with
	// the given trainer.
	SaveLastContactURL(ctx context.Context, t Trainer, url string) error
	// LoadLastContactURL loads the last contact URL associated with the given
	// trainer. The second return value is true if there is a last contact
	// URL associated with this trainer, false otherwise.
	LoadLastContactURL(ctx context.Context, t Trainer) (string, error)
	// LoadUUIDFromHumanTrainerName finds the corresponding UUID for the given
	// name of a human (non-NPC) trainer. The second return value is true if a
	// trainer with the given name was found.
	LoadUUIDFromHumanTrainerName(ctx context.Context, name string) (string, error)

	// NewPokemon creates a database Pokemon that is ready to be saved from the
	// given pkmn.Pokemon.
	NewPokemon(p pkmn.Pokemon) Pokemon
	// SavePokemon saves the given Pokemon as owned by the given trainer.
	SavePokemon(ctx context.Context, t Trainer, pkmn Pokemon) error
	// LoadPokemon loads a Pokemon with the given UUID. The second return value
	// is true if the Pokemon exists, false otherwise.
	LoadPokemon(ctx context.Context, uuid string) (Pokemon, error)
	// DeletePokemon deletes a Pokemon with the given UUID.
	DeletePokemon(ctx context.Context, uuid string) error
	// SaveParty saves a batch of Pokemon as owend by the given trainer.
	SaveParty(ctx context.Context, t Trainer, party []Pokemon) error
	// LoadParty returns all the Pokemon in the given trainer's party. The
	// second return value is true if any Pokemon were found, false otherwise.
	LoadParty(ctx context.Context, t Trainer) ([]Pokemon, error)

	// NewBattle creates a database battle that is ready to be saved from the
	// given pkmn.Battle.
	NewBattle(b pkmn.Battle) Battle
	// SaveBattle saves the given battle.
	SaveBattle(ctx context.Context, b Battle) error
	// LoadBattle returns a battle that the given trainer UUID is involved in.
	// The second return value is true if the battle exists and was retrieved,
	// false otherwise.
	LoadBattle(ctx context.Context, p1uuid, p2uuid string) (Battle, error)
	// LoadBattleTrainerIsIn returns a battle the given trainer UUID is involved
	// in.
	LoadBattleTrainerIsIn(ctx context.Context, tuuid string) (Battle, error)
	// DeleteTrainer deletes the battle that the two trainers are involved in.
	DeleteBattle(ctx context.Context, p1uuid, p2uuid string) error
	// PurgeBattle deletes the battle that the two trainers are involved in and
	// purges any relating data having to do with this battle.
	PurgeBattle(ctx context.Context, p1uuid, p2uuid string) error

	// NewTrainerBattleInfo creates a new trainer battle info that is ready to
	// be saved from the given pkmn.TrainerBattleInfo.
	NewTrainerBattleInfo(tbi pkmn.TrainerBattleInfo) TrainerBattleInfo
	// SaveTrainerBattleInfo saves the given trainer battle info given the
	// battle that it pertains to.
	SaveTrainerBattleInfo(ctx context.Context, b Battle, tbi TrainerBattleInfo) error
	// LoadTrainerBattleInfo returns a trainer battle info for the given
	// trainer UUID and battle. The second return value is true if the battle
	// info exists and was retrieved, false otherwise.
	LoadTrainerBattleInfo(ctx context.Context, b Battle, tuuid string) (TrainerBattleInfo, error)
	// DeleteTrainerBattleInfos deletes all trainer battle infos under the
	// given battle.
	DeleteTrainerBattleInfos(ctx context.Context, b Battle) error

	// NewPokemonBattleInfo creates a new Pokemon battle info that is ready to
	// be saved from the given pkmn.PokemonBattleInfo.
	NewPokemonBattleInfo(pbi pkmn.PokemonBattleInfo) PokemonBattleInfo
	// SavePokemonBattleInfo saves the given Pokemon battle info given the
	// battle it pertains to.
	SavePokemonBattleInfo(ctx context.Context, b Battle, pbi PokemonBattleInfo) error
	// LoadPokemonBattleInfo returns a Pokemon battle info for the given
	// Pokemon UUID and battle. The second return value is true if the battle
	// info exists and was retrieved, false otherwise.
	LoadPokemonBattleInfo(ctx context.Context, b Battle, uuid string) (PokemonBattleInfo, error)
	// DeletePokemonBattleInfos deletes all Pokemon battle infos under the
	// given battle.
	DeletePokemonBattleInfos(ctx context.Context, b Battle) error
}

var implementations map[string]Database

func init() {
	implementations = make(map[string]Database)
}

// Register registers an implementation of the Database interface under the
// given name.
func Register(name string, db Database) {
	implementations[name] = db
}

// Get returns the implementation of Database with the given name, or an error
// if no such implementation exists.
func Get(name string) (Database, error) {
	if db, ok := implementations[name]; ok {
		return db, nil
	}

	return nil, errors.New("no Database implementation called '" + name + "' found")
}
