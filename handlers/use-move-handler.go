package handlers

import (
	"strconv"

	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
)

// UseMove handles requests to use a Pokemon move. This function will
// queue up a move to be run once both trainers finish selecting the action
// they will take.
type UseMove struct {
	Services
}

func (h *UseMove) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(basicTrainerData)

	// Check if the command looks correct
	if len(slackReq.CommandParams) != 1 {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     invalidCommandTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "failed to populate invalid command template", err: err}
		}
	}

	// Load the battle data
	battleData, err := loadBattleData(ctx, s.DB, requester)
	if err != nil {
		return handlerError{user: "could not load battle data", err: err}
	}

	// Extract the move ID from the command
	scrambledID, err := strconv.Atoi(slackReq.CommandParams[0])
	if err != nil {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     invalidCommandTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not populate invalid command template", err: err}
		}
		return nil // There is nothing else to process
	}

	// Get all the move lookup tables
	mlts, err := s.DB.LoadMoveLookupTables(ctx, battleData.battle)
	if err != nil {
		return handlerError{user: "could not fetch move lookup tables", err: err}
	}
	// Find the trainer's move lookup table
	var mlt *pkmn.MoveLookupTable
	for _, val := range mlts {
		if val.GetMoveLookupTable().TrainerUUID == requester.trainer.GetTrainer().UUID {
			mlt = val.GetMoveLookupTable()
			break
		}
	}
	if mlt == nil {
		// The trainer doesn't have a move lookup table
		return handlerError{user: "could not fetch move lookup table", err: err}
	}
	moveID := mlt.Lookup(scrambledID)

	// Set up the next action to be a move action
	battleData.requester.battleInfo.GetTrainerBattleInfo().FinishedTurn = true
	battleData.requester.battleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.MoveBattleActionType,
		Val:  moveID}

	// Get the move information
	apiMove, err := s.Fetcher.FetchMove(ctx, client, battleData.requester.battleInfo.GetTrainerBattleInfo().NextBattleAction.Val)
	if err != nil {
		return handlerError{user: "could not fetch move information", err: err}
	}
	move, err := pokeapi.NewMove(apiMove)
	if err != nil {
		return handlerError{user: "could not fetch move information", err: err}
	}

	// Send confirmation that the move was received
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     moveConfirmationTemplate,
		TemplInfo: move.Name})
	if err != nil {
		return handlerError{user: "could not populate move confirmation template", err: err}
	}

	// Get a turn processor ready to do any required processing
	tp := turnProcessor{Services: s}

	var battleOver bool
	// Do any work required to get the opponent ready for the turn to be
	// processed
	ready, err := preprocessTurn(ctx, s, &battleData)
	if err != nil {
		return handlerError{user: "could not do preprocessing on the current turn", err: err}
	}
	if ready {
		// The opponent is ready and the turn may be processed
		battleOver, err = tp.process(ctx, &battleData)
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
