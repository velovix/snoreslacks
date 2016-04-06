package database

import (
	"errors"

	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

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

// MoveLookupTable describes a database representation of a move lookup table
// that can emit a pkmn.MoveLookupTable
type MoveLookupTable interface {
	GetMoveLookupTable() *pkmn.MoveLookupTable
}

// PartyMemberLookupTable describes a database representation of a party
// member lookup table that can emit a pkmn.PartyMemberLookupTable.
type PartyMemberLookupTable interface {
	GetPartyMemberLookupTable() *pkmn.PartyMemberLookupTable
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
	// LoadTrainer returns a Trainer with the given name. The second return
	// value is true if the Trainer exists and was retrieved, false otherwise.
	LoadTrainer(ctx context.Context, name string) (Trainer, bool, error)
	// SaveLastContactURL saves the given last contact URL as associated with
	// the given trainer.
	SaveLastContactURL(ctx context.Context, t Trainer, url string) error
	// LoadLastContactURL loads the last contact URL associated with the given
	// trainer. The second return value is true if there is a last contact
	// URL associated with this trainer, false otherwise.
	LoadLastContactURL(ctx context.Context, t Trainer) (string, bool, error)

	// NewPokemon creates a database Pokemon that is ready to be saved from the
	// given pkmn.Pokemon.
	NewPokemon(p pkmn.Pokemon) Pokemon
	// SavePokemon saves the given Pokemon as owned by the given trainer.
	SavePokemon(ctx context.Context, t Trainer, pkmn Pokemon) error
	// LoadPokemon loads a Pokemon with the given UUID. The second return value
	// is true if the Pokemon exists, false otherwise.
	LoadPokemon(ctx context.Context, uuid string) (Pokemon, bool, error)
	// SaveParty saves a batch of Pokemon as owend by the given trainer.
	SaveParty(ctx context.Context, t Trainer, party []Pokemon) error
	// LoadParty returns all the Pokemon in the given trainer's party. The
	// second return value is true if any Pokemon were found, false otherwise.
	LoadParty(ctx context.Context, t Trainer) ([]Pokemon, bool, error)

	// NewBattle creates a database battle that is ready to be saved from the
	// given pkmn.Battle.
	NewBattle(b pkmn.Battle) Battle
	// SaveBattle saves the given battle.
	SaveBattle(ctx context.Context, b Battle) error
	// LoadBattle returns a battle that the given Trainer name is involved in.
	// The second return value is true if the battle exists and was retrieved,
	// false otherwise.
	LoadBattle(ctx context.Context, p1Name, p2Name string) (Battle, bool, error)
	// LoadBattleTrainerIsIn returns a battle the given Trainer name is involved
	// in.
	LoadBattleTrainerIsIn(ctx context.Context, pName string) (Battle, bool, error)
	// DeleteTrainer deletes the battle that the two Trainers are involved in.
	DeleteBattle(ctx context.Context, p1Name, p2Name string) error

	// NewMoveLookupTable creates a database move lookup table that is ready
	// to be saved from the given pkmn.MoveLookupTable.
	NewMoveLookupTable(mlt pkmn.MoveLookupTable) MoveLookupTable
	// SaveMoveLookupTable Saves a move lookup table to the database given the
	// battle object it belongs to.
	SaveMoveLookupTable(ctx context.Context, table MoveLookupTable, b Battle) error
	// LoadMoveLookupTables Loads all the move lookup tables that the given
	// battle object owns.
	LoadMoveLookupTables(ctx context.Context, b Battle) ([]MoveLookupTable, bool, error)

	// NewPartyMemberLookupTable creates a database party member lookup table
	// that is ready to be saved from the given pkmn.PartyMemberLookupTable.
	NewPartyMemberLookupTable(pmlt pkmn.PartyMemberLookupTable) PartyMemberLookupTable
	// SavePartyMemberLookupTable Saves a party member lookup table to the
	// database given the Battle object it belongs to.
	SavePartyMemberLookupTable(ctx context.Context, table PartyMemberLookupTable, b Battle) error
	// LoadPartyMemberLookupTables Loads all the party member lookup tables
	// that the given Battle object owns.
	LoadPartyMemberLookupTables(ctx context.Context, b Battle) ([]PartyMemberLookupTable, bool, error)

	// NewTrainerBattleInfo creates a new trainer battle info that is ready to
	// be saved from the given pkmn.TrainerBattleInfo.
	NewTrainerBattleInfo(tbi pkmn.TrainerBattleInfo) TrainerBattleInfo
	// SaveTrainerBattleInfo saves the given trainer battle info given the
	// battle that it pertains to.
	SaveTrainerBattleInfo(ctx context.Context, b Battle, tbi TrainerBattleInfo) error
	// LoadTrainerBattleInfo returns a trainer battle info for the given
	// trainer name and battle. The second return value is true if the battle
	// info exists and was retrieved, false otherwise.
	LoadTrainerBattleInfo(ctx context.Context, b Battle, tName string) (TrainerBattleInfo, bool, error)

	// NewPokemonBattleInfo creates a new Pokemon battle info that is ready to
	// be saved from the given pkmn.PokemonBattleInfo.
	NewPokemonBattleInfo(pbi pkmn.PokemonBattleInfo) PokemonBattleInfo
	// SavePokemonBattleInfo saves the given Pokemon battle info given the
	// battle it pertains to.
	SavePokemonBattleInfo(ctx context.Context, b Battle, pbi PokemonBattleInfo) error
	// LoadPokemonBattleInfo returns a Pokemon battle info for the given
	// Pokemon UUID and battle. The second return value is true if the battle
	// info exists and was retrieved, false otherwise.
	LoadPokemonBattleInfo(ctx context.Context, b Battle, uuid string) (PokemonBattleInfo, bool, error)
}

var implementations map[string]Database

func init() {
	implementations = make(map[string]Database)
}

func Register(name string, db Database) {
	implementations[name] = db
}

func Get(name string) (Database, error) {
	if db, ok := implementations[name]; ok {
		return db, nil
	}

	return nil, errors.New("no Database implementation called '" + name + "' found")
}
