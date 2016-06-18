package handlers

import (
	"github.com/pkg/errors"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"

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
func makeActionOptions(ctx context.Context, s Services, trainerData *basicTrainerData, trainerDataBI database.TrainerBattleInfo, b database.Battle) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Get the current Pokemon
	currPkmn := trainerData.pkmn[trainerDataBI.GetTrainerBattleInfo().CurrPkmnSlot]

	// Create the move selector
	var moveSlots []string
	for _, moveID := range currPkmn.GetPokemon().MoveIDsAsSlice() {
		// Load the move info to get the name
		move, err := loadMove(ctx, client, s.Fetcher, moveID)
		if err != nil {
			return errors.Wrap(err, "making action options")
		}
		moveSlots = append(moveSlots, move.Name)
	}

	// Create the party selector
	var partySlots []string
	for _, pkmn := range trainerData.pkmn {
		partySlots = append(partySlots, pkmn.GetPokemon().Name)
	}

	// Send action options to the player
	templInfo := struct {
		CurrPokemonName string
		MoveSlots       []string
		PartySlots      []string
	}{
		CurrPokemonName: currPkmn.GetPokemon().Name,
		MoveSlots:       moveSlots,
		PartySlots:      partySlots}
	err := messaging.SendTempl(client, trainerData.lastContactURL, messaging.TemplMessage{
		Templ:     actionOptionsTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return err
	}

	return nil
}
