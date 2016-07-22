package handlers

import (
	"errors"
	"strconv"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
)

// SwitchPokemon handles requests to switch Pokemon. This function will queue
// up a switch to be run once both trainers finish selecting the action they
// will take.
type SwitchPokemon struct {
	Services
}

func (h *SwitchPokemon) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	battleData := ctx.Value("battle data").(*battleData)

	// Assert that the trainer is in battle mode
	if requester.trainer.GetTrainer().Mode != pkmn.BattlingTrainerMode {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     switchingWhenInWrongModeTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "failed to populate switching when in wrong mode template", err: err}
		}
		return nil // No more work to do
	}

	// Assert that all necessary data is in the battle data object
	if !battleData.isComplete() {
		return handlerError{user: "could not load battle data", err: errors.New("incomplete battle data object")}
	}

	// Check if the command looks correct
	if len(slackReq.CommandParams) != 1 {
		return sendInvalidCommand(client, requester.lastContactURL)
	}

	// Extract the party slot ID from the command
	partySlotID, err := strconv.Atoi(slackReq.CommandParams[0])
	if err != nil {
		return sendInvalidCommand(client, requester.lastContactURL)
	}

	// Check if the party slot ID is valid
	if partySlotID < 1 || partySlotID > len(battleData.requester.pkmn) {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     invalidPartySlotTemplate,
			TemplInfo: partySlotID})
		if err != nil {
			return handlerError{user: "could not populate invalid party slot template", err: err}
		}
		return nil // There is nothing else to do
	}

	// Check that the Pokemon to be switched to is a different Pokemon from the
	// one currently out
	if partySlotID-1 == battleData.requester.battleInfo.GetTrainerBattleInfo().CurrPkmnSlot {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     switchToCurrentPokemonTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not populate switch to current Pokemon template", err: err}
		}
		return nil // There is nothing else to do
	}

	// Check that the Pokemon to be switched to can fight
	if battleData.requester.pkmnBattleInfo[partySlotID-1].GetPokemonBattleInfo().CurrHP <= 0 {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     switchToFaintedPokemonTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not populate switch to fainted Pokemon template", err: err}
		}
		return nil // There is nothing else to do
	}

	// Set up the next battle action to be a switch Pokemon action
	battleData.requester.battleInfo.GetTrainerBattleInfo().FinishedTurn = true
	battleData.requester.battleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.SwitchBattleActionType,
		Val:  partySlotID - 1}

	// Send confirmation that the switch was received
	err = messaging.SendTempl(client, battleData.requester.lastContactURL, messaging.TemplMessage{
		Templ:     switchConfirmationTemplate,
		TemplInfo: requester.pkmn[partySlotID-1].GetPokemon().Name,
		Public:    false})
	if err != nil {
		return handlerError{user: "could not populate switch confirmation template", err: err}
	}

	// Get a turn processor ready to do any required processing
	tp := turnProcessor{Services: s}

	var battleOver bool
	// Do any work required to get the opponent ready for the turn to be
	// processed
	ready, err := preprocessTurn(ctx, s, battleData)
	if err != nil {
		return handlerError{user: "could not do preprocessing on the current turn", err: err}
	}
	if ready {
		// The opponent is ready and the turn may be processed
		battleOver, err = tp.process(ctx, battleData)
		if err != nil {
			return handlerError{user: "could not process the current turn", err: err}
		}
	}

	// Save data if all has gone well
	err = saveBattleData(ctx, s.DB, battleData)
	if err != nil {
		return handlerError{user: "could not save battle session", err: err}
	}
	if battleOver && battleData.opponent.trainer.GetTrainer().Type == pkmn.WildTrainerType {
		// The battle is over and the trainer is one-time-use. It's time to
		// destroy him.
		err = s.DB.PurgeTrainer(ctx, battleData.opponent.trainer.GetTrainer().UUID)
		if err != nil {
			return handlerError{user: "could not purge the wild trainer", err: err}
		}
	}
	if battleOver {
		// Delete the battle if it has ended
		err = s.DB.PurgeBattle(ctx, battleData.battle.GetBattle().P1, battleData.battle.GetBattle().P2)
		if err != nil {
			return handlerError{user: "could not delete a battle", err: err}
		}
	}

	return nil
}
