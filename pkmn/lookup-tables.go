package pkmn

// MoveLookupElement represents a single match between a scrambled ID and the
// corresponding move ID.
type MoveLookupElement struct {
	ID       int
	MoveID   int
	MoveName string
}

// MoveLookupTable contains a collection of matches between scrambled IDs and
// move IDs.
type MoveLookupTable struct {
	TrainerName string
	Moves       []MoveLookupElement
}

// Lookup finds the move ID that corresponds to the given scambled ID.
func (table MoveLookupTable) Lookup(id int) int {
	for _, val := range table.Moves {
		if val.ID == id {
			return val.MoveID
		}
	}

	return -1
}

// PartyMemberLookupElement represents a single match between a scrambled ID
// and the corresponding party slot.
type PartyMemberLookupElement struct {
	ID       int
	SlotID   int
	PkmnName string
}

// PartyMemberLookupTable contains a collection of matches between scrambled
// IDs and party slots.
type PartyMemberLookupTable struct {
	TrainerName string
	Members     []PartyMemberLookupElement
}

// Lookup finds the party slot that corresponds to the given scrambled ID.
func (table PartyMemberLookupTable) Lookup(id int) int {
	for _, val := range table.Members {
		if val.ID == id {
			return val.SlotID
		}
	}

	return -1
}
