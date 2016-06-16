package handlers

import "golang.org/x/net/context"

// SwitchPokemon handles requests to switch Pokemon. This function will queue
// up a switch to be run once both trainers finish selecting the action they
// will take.
type SwitchPokemon struct {
	Services
}

func (h *SwitchPokemon) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	/*slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(basicTrainerData)

	// Load the battle data
	battleData, err := loadBattleData(ctx, s.DB, requester)
	if err != nil {
		return handlerError{user: "could not load battle data", err: err}
	}

	// Set up the next battle action to be a switch Pokemon action
	battleData.curr.battleInfo.GetTrainerBattleInfo().FinishedTurn = true
	battleData.curr.battleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.SwitchBattleActionType,
		Val:  5}

	// Send confirmation that the switch was received
	err = messaging.SendTempl(client, battleData.curr.trainer.lastContactURL, messaging.TemplMessage{
		Templ:     switchConfirmationTemplate,
		TemplInfo: "some crap",
		Public:    true})
	if err != nil {
		return handlerError{user: "could not populate switch confirmation template", err: err}
	}

	// Check if both trainers have chosen their move and run the turns if so
	var p1PkmnBI, p2PkmnBI database.PokemonBattleInfo
	var success, battleOver bool
	if trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn &&
		opponentBattleInfo.GetTrainerBattleInfo().FinishedTurn {

		if requester.trainer.GetTrainer().UUID == b.GetBattle().P1 {
			// If the current trainer is player one
			p1PkmnBI, p2PkmnBI, battleOver, success = tp.process(ctx, client, req, requester, opponent, trainerBattleInfo, opponentBattleInfo, b)
			if !success {
				return // Abort operation
			}
		} else {
			// If the current trainer is player two
			p1PkmnBI, p2PkmnBI, battleOver, success = tp.process(ctx, client, req, opponent, requester, opponentBattleInfo, trainerBattleInfo, b)
			if !success {
				return // Abort operation
			}
		}
	}

	if battleOver {
		// Get the trainers out of the battle if it is over
		requester.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
		opponent.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
	}

	// Save data if all has gone well
	err = s.DB.SaveTrainer(ctx, requester.trainer)
	if err != nil {
		s.Log.Errorf(ctx, "%s", err)
		return
	}
	err = s.DB.SaveTrainer(ctx, opponent.Trainer)
	if err != nil {
		s.Log.Errorf(ctx, "%s", err)
		return
	}
	err = s.DB.SaveBattle(ctx, b)
	if err != nil {
		s.Log.Errorf(ctx, "%s", err)
		return
	}
	err = s.DB.SaveTrainerBattleInfo(ctx, b, trainerBattleInfo)
	if err != nil {
		s.Log.Errorf(ctx, "%s", err)
		return
	}
	if p1PkmnBI != nil {
		err = s.DB.SavePokemonBattleInfo(ctx, b, p1PkmnBI)
		if err != nil {
			s.Log.Errorf(ctx, "%s", err)
			return
		}
	}
	if p2PkmnBI != nil {
		err = s.DB.SavePokemonBattleInfo(ctx, b, p2PkmnBI)
		if err != nil {
			s.Log.Errorf(ctx, "%s", err)
			return
		}
	}
	if battleOver {
		err = s.DB.PurgeBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
		if err != nil {
			s.Log.Errorf(ctx, "%s", err)
			return
		}
	}*/

	return nil
}
