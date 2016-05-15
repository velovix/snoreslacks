package gaedatabase

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// GAEMoveLookupTable is the database object wrapper of a move lookup table for
// datastore.
type GAEMoveLookupTable struct {
	pkmn.MoveLookupTable
}

// NewMoveLookupTable creates a database move lookup table that is ready
// to be saved from the given pkmn.MoveLookupTable.
func (db GAEDatabase) NewMoveLookupTable(mlt pkmn.MoveLookupTable) database.MoveLookupTable {
	return &GAEMoveLookupTable{MoveLookupTable: mlt}
}

// GetMoveLookupTable returns the underlying move lookup table from the
// database object.
func (mlt *GAEMoveLookupTable) GetMoveLookupTable() *pkmn.MoveLookupTable {
	return &mlt.MoveLookupTable
}

// GAEPartyMemberLookupTable is the database wrapper object of a party member
// lookup table for datastore.
type GAEPartyMemberLookupTable struct {
	pkmn.PartyMemberLookupTable
}

// NewPartyMemberLookupTable creates a database party member lookup table
// that is ready to be saved from the given pkmn.PartyMemberLookupTable.
func (db GAEDatabase) NewPartyMemberLookupTable(pmlt pkmn.PartyMemberLookupTable) database.PartyMemberLookupTable {
	return &GAEPartyMemberLookupTable{PartyMemberLookupTable: pmlt}
}

// GetPartyMemberLookupTable returns the underlying party member lookup table
// from the database object.
func (pmlt *GAEPartyMemberLookupTable) GetPartyMemberLookupTable() *pkmn.PartyMemberLookupTable {
	return &pmlt.PartyMemberLookupTable
}

// SaveMoveLookupTable saves the lookup table to the Datastore.
func (db GAEDatabase) SaveMoveLookupTable(ctx context.Context, dbmlt database.MoveLookupTable, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	mlt, ok := dbmlt.(*GAEMoveLookupTable)
	if !ok {
		panic("The given move lookup table is not of the right type for this implementation. Are you using two implementations by misake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	mltKey := datastore.NewIncompleteKey(ctx, "move lookup table", battleKey)

	_, err := datastore.Put(ctx, mltKey, mlt)
	if err != nil {
		return err
	}

	return nil
}

// LoadMoveLookupTables loads all move lookup tables attached to the given
// battle. If none are found, an empty slice is returned and the second return
// value is false.
func (db GAEDatabase) LoadMoveLookupTables(ctx context.Context, dbb database.Battle) ([]database.MoveLookupTable, bool, error) {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	var gaeTables []*GAEMoveLookupTable

	_, err := datastore.NewQuery("move lookup table").
		Ancestor(battleKey).
		GetAll(ctx, &gaeTables)
	if err != nil {
		return make([]database.MoveLookupTable, 0), false, err
	}
	if len(gaeTables) == 0 {
		return make([]database.MoveLookupTable, 0), false, nil
	}

	tables := make([]database.MoveLookupTable, len(gaeTables))
	for i, val := range gaeTables {
		tables[i] = val
	}

	return tables, true, nil
}

// DeleteMoveLookupTables deletes all move lookup tables under the given battle.
func (db GAEDatabase) DeleteMoveLookupTables(ctx context.Context, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	// Find all move lookup tables under this battle
	keys, err := datastore.NewQuery("move lookup table").
		KeysOnly().
		Ancestor(battleKey).
		GetAll(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all the move lookup tables
	for _, key := range keys {
		err = datastore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

// SavePartyMemberLookupTable saves the given party member lookup table under
// the given battle.
func (db GAEDatabase) SavePartyMemberLookupTable(ctx context.Context, dbpmlt database.PartyMemberLookupTable, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}
	pmlt, ok := dbpmlt.(*GAEPartyMemberLookupTable)
	if !ok {
		panic("The given party member lookup table is not of the right type for this implementation. Are you using two implementations by misake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)
	pmltKey := datastore.NewIncompleteKey(ctx, "party member lookup table", battleKey)

	_, err := datastore.Put(ctx, pmltKey, pmlt)
	if err != nil {
		return err
	}

	return nil
}

// LoadPartyMemberLookupTables loads all party member lookup tables attached to
// the given battle. If none are found, an empty slice is returned and the
// second return value is false.
func (db GAEDatabase) LoadPartyMemberLookupTables(ctx context.Context, dbb database.Battle) ([]database.PartyMemberLookupTable, bool, error) {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	var tables []database.PartyMemberLookupTable

	_, err := datastore.NewQuery("party member lookup table").
		Ancestor(battleKey).
		GetAll(ctx, &tables)
	if err != nil {
		return make([]database.PartyMemberLookupTable, 0), false, err
	}
	if len(tables) == 0 {
		return tables, false, nil
	}

	return tables, true, nil
}

// DeletePartyMemberLookupTables deletes all party member lookup tables under
// the given battle.
func (db GAEDatabase) DeletePartyMemberLookupTables(ctx context.Context, dbb database.Battle) error {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	// Find all party member lookup tables under this battle
	keys, err := datastore.NewQuery("party member lookup table").
		KeysOnly().
		Ancestor(battleKey).
		GetAll(ctx, nil)
	if err != nil {
		return err
	}

	// Delete all the party member lookup tables
	for _, key := range keys {
		err = datastore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}
