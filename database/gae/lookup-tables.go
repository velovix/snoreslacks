package gaedatabase

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type GAEMoveLookupTable struct {
	pkmn.MoveLookupTable
}

// NewMoveLookupTable creates a database move lookup table that is ready
// to be saved from the given pkmn.MoveLookupTable.
func (db GAEDatabase) NewMoveLookupTable(mlt pkmn.MoveLookupTable) database.MoveLookupTable {
	return &GAEMoveLookupTable{MoveLookupTable: mlt}
}

func (mlt *GAEMoveLookupTable) GetMoveLookupTable() *pkmn.MoveLookupTable {
	return &mlt.MoveLookupTable
}

type GAEPartyMemberLookupTable struct {
	pkmn.PartyMemberLookupTable
}

// NewPartyMemberLookupTable creates a database party member lookup table
// that is ready to be saved from the given pkmn.PartyMemberLookupTable.
func (db GAEDatabase) NewPartyMemberLookupTable(pmlt pkmn.PartyMemberLookupTable) database.PartyMemberLookupTable {
	return &GAEPartyMemberLookupTable{PartyMemberLookupTable: pmlt}
}

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

func (db GAEDatabase) LoadMoveLookupTables(ctx context.Context, dbb database.Battle) ([]database.MoveLookupTable, bool, error) {
	b, ok := dbb.(*GAEBattle)
	if !ok {
		panic("The given battle is not of the right type for this implementation. Are you using two implementations by mistake?")
	}

	battleKey := datastore.NewKey(ctx, "battle", battleName(b), 0, nil)

	var tables []database.MoveLookupTable

	_, err := datastore.NewQuery("move lookup table").
		Ancestor(battleKey).
		GetAll(ctx, &tables)
	if err != nil {
		return make([]database.MoveLookupTable, 0), false, err
	}
	if len(tables) == 0 {
		return tables, false, nil
	}

	return tables, true, nil
}

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
