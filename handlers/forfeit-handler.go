package handlers

import (
	"errors"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
)

// Forfeit handles requests to forfeit a battle. If the battle is in a waiting
// state, then the trainer stops waiting. If the battle has started, a forfeit
// will result in the trainer losing the match.
type Forfeit struct {
}

func (h *Forfeit) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	battleData := ctx.Value("battle data").(*battleData)

	// Assert that the trainer is currently in either battling mode or battle
	// waiting mode
	if requester.trainer.GetTrainer().Mode != pkmn.BattlingTrainerMode {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     challengingWhenInWrongMode,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not send challenging when in wrong mode template", err: err}
		}
		return nil // No more work to do
	}

	// Assert that the required data is inside of the battle data object
	if !battleData.hasBattle() {
		return handlerError{user: "could not load battle data", err: errors.New("no battle object in battle data")}
	}

	// Find the UUID of the opponent
	opponentUUID := battleData.battle.GetBattle().P1
	if opponentUUID == requester.trainer.GetTrainer().UUID {
		opponentUUID = battleData.battle.GetBattle().P2
	}

	// Take the current trainer out of battle mode
	requester.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode

	if battleData.battle.GetBattle().Mode == pkmn.WaitingBattleMode {
		// The battle hadn't started yet, so nobody loses

		// Construct the template letting everyone know that the trainer
		// forfeitted
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     waitingForfeitTemplate,
			TemplInfo: battleData.battle.GetBattle(),
			Public:    true})
		if err != nil {
			return handlerError{user: "could not populate waiting forfeit template", err: err}
		}
	} else if battleData.battle.GetBattle().Mode == pkmn.StartedBattleMode {
		// The battle has started, so the forfeitter will lose

		// Load the opponent
		opponent, err := loadBasicTrainerData(ctx, s.DB, opponentUUID)
		if err != nil {
			return handlerError{user: "could not load opponent information", err: err}
		}

		// Take the opponent out of battle mode
		opponent.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode

		// Count this as a loss for the forfeitting trainer
		requester.trainer.GetTrainer().Losses++
		opponent.trainer.GetTrainer().Wins++

		// Construct a template letting everyone know that the requesting
		// trainer forfeitted
		battlingForfeitTemplInfo := struct {
			Forfeitter string
			Opponent   string
		}{
			Forfeitter: requester.trainer.GetTrainer().Name,
			Opponent:   opponent.trainer.GetTrainer().Name}

		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Type:      messaging.Good,
			Templ:     battlingForfeitTemplate,
			TemplInfo: battlingForfeitTemplInfo,
			Public:    true})
		if err != nil {
			return handlerError{user: "could not populate battling forfeit template", err: err}
		}

		// Save the opponent
		err = saveBasicTrainerData(ctx, s.DB, opponent)
		if err != nil {
			return handlerError{user: "could not save opponent", err: err}
		}
	}

	// Save changes to the current trainer
	err := s.DB.SaveTrainer(ctx, requester.trainer)
	if err != nil {
		return handlerError{user: "could not save the requesting trainer", err: err}
	}

	// The battle is over, so it should be deleted
	s.Log.Infof(ctx, "deleting a battle %s/%s", battleData.battle.GetBattle().P1, battleData.battle.GetBattle().P2)
	err = s.DB.PurgeBattle(ctx, battleData.battle.GetBattle().P1, battleData.battle.GetBattle().P2)
	if err != nil {
		return handlerError{user: "could not delete the battle now that it's over", err: err}
	}

	return nil
}
