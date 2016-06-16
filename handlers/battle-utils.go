package handlers

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"

	"golang.org/x/net/context"
)

// makeTextHPBar creates an HP bar for the given Pokemon in text.
func makeTextHPBar(p *pkmn.Pokemon, pBI *pkmn.PokemonBattleInfo) string {
	percentHealth := (float64(pBI.CurrHP) / float64(pkmn.CalcOOBHP(p.HP, *p))) * 100.0
	bar := "["

	for i := 0; i < 15; i++ {
		if percentHealth > (100.0 * (float64(i) / 15.0)) {
			bar += "#"
		} else {
			bar += " "
		}
	}

	return bar + "]"
}

// makeActionOptions makes and sends each player their move and party switching
// options.
func makeActionOptions(ctx context.Context, s Services, trainerData basicTrainerData, trainerDataBI database.TrainerBattleInfo, b database.Battle) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Get the current Pokemon
	currPkmn := trainerData.pkmn[trainerDataBI.GetTrainerBattleInfo().CurrPkmnSlot]

	// Construct a move lookup table
	var mlt pkmn.MoveLookupTable
	mlt.TrainerUUID = trainerData.trainer.GetTrainer().UUID
	mlt.Moves = make([]pkmn.MoveLookupElement, currPkmn.GetPokemon().MoveCount())

	// Construct the move lookup elements
	moves := currPkmn.GetPokemon().MoveIDsAsSlice()
	for i, moveID := range moves {
		// Fetch move info from PokeAPI
		apiMove, err := s.Fetcher.FetchMove(ctx, client, moveID)
		if err != nil {
			return err
		}

		// Create a pkmn.Move from the PokeAPI data
		move, err := pokeapi.NewMove(apiMove)
		if err != nil {
			return err
		}

		mlt.Moves[i] = pkmn.MoveLookupElement{
			ID:       i,
			MoveID:   moveID,
			MoveName: move.Name}
	}

	// Fetch the trainer's party
	party, err := s.DB.LoadParty(ctx, trainerData.trainer)
	if err != nil {
		return err
	}

	// Construct a party member lookup table
	var pmlt pkmn.PartyMemberLookupTable
	pmlt.TrainerUUID = trainerData.trainer.GetTrainer().UUID
	pmlt.Members = make([]pkmn.PartyMemberLookupElement, len(party))

	// Construct the party member lookup elements
	for i, val := range party {
		element := pkmn.PartyMemberLookupElement{
			ID:       i,
			SlotID:   i,
			PkmnName: val.GetPokemon().Name}
		pmlt.Members[i] = element
	}

	// Send action options to the player
	templInfo := struct {
		CurrPokemonName string
		MoveTable       pkmn.MoveLookupTable
		PartyTable      pkmn.PartyMemberLookupTable
	}{
		CurrPokemonName: currPkmn.GetPokemon().Name,
		MoveTable:       mlt,
		PartyTable:      pmlt}
	err = messaging.SendTempl(client, trainerData.lastContactURL, messaging.TemplMessage{
		Templ:     actionOptionsTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return err
	}

	// Save data if all went well
	err = s.DB.SaveMoveLookupTable(ctx, s.DB.NewMoveLookupTable(mlt), b)
	if err != nil {
		return err
	}
	err = s.DB.SavePartyMemberLookupTable(ctx, s.DB.NewPartyMemberLookupTable(pmlt), b)
	if err != nil {
		return err
	}

	return nil
}
