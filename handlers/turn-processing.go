package handlers

import (
	"errors"
	"math/rand"
	"text/template"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
)

// preprocessTurn checks if both members of the battle are ready to have the
// current turn processed, and returns true if they are.
//
// If the user is facing a human, this function will only let the turn be
// processed if the opponent has picked their action. If the user is facing
// a bot of some kind, this method will choose their action for them.
func preprocessTurn(ctx context.Context, s Services, bd *battleData) (bool, error) {
	switch bd.opponent.trainer.GetTrainer().Type {
	case pkmn.HumanTrainerType:
		// The opponent is a fellow human, so both trainers have to be given a
		// chance to pick their move. Check if both trainers have chosen their
		// move.
		if bd.requester.battleInfo.GetTrainerBattleInfo().FinishedTurn &&
			bd.opponent.battleInfo.GetTrainerBattleInfo().FinishedTurn {
			return true, nil
		}

		// The opponent has not picked their action yet
		return false, nil
	case pkmn.WildTrainerType:
		// The opponent is a wild Pokemon, so their move will be chosen
		// algorithmically
		opponentPkmn := bd.opponent.activePkmn()
		moveCnt := opponentPkmn.GetPokemon().MoveCount()
		moves := opponentPkmn.GetPokemon().MoveIDsAsSlice()

		// Find move
		moveID, err := bd.opponent.trainer.GetTrainer().PickMove(moves, moveCnt)
		if err != nil {
			return false, err
		}

		s.Log.Infof(ctx, "wild opponent will be using a move: %v", moveID)

		// Set up the next action to be a move action
		bd.opponent.battleInfo.GetTrainerBattleInfo().FinishedTurn = true
		bd.opponent.battleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
			Type: pkmn.MoveBattleActionType,
			Val:  moveID}

		// Now that the opponent has their choice of action ready, the turn is
		// ready to be processed
		return true, nil
	default:
		panic("unsupported trainer type")
	}
}

type turnProcessor struct {
	Services
}

// checkIfPlayerLost returns true if the given trainer has no more Pokemon that
// are able to fight.
func (tp *turnProcessor) checkIfPlayerLost(ctx context.Context, bd *battleData, checkee, opponent *battleTrainerData) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	for _, memberBI := range checkee.pkmnBattleInfo {
		// Check if the Pokemon is still able to fight
		if memberBI.GetPokemonBattleInfo().CurrHP > 0 {
			// There's one Pokemon in the party still able to fight, so the
			// trainer has not lost yet
			return false, nil
		}
	}

	// The trainer has lost. Let the world know.
	templInfo := struct {
		LostTrainer string
		WonTrainer  string
	}{
		LostTrainer: checkee.trainer.GetTrainer().Name,
		WonTrainer:  opponent.trainer.GetTrainer().Name}
	err := messaging.SendTempl(client, checkee.lastContactURL, messaging.TemplMessage{
		Templ:     trainerLostTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return false, err
	}

	return true, nil
}

// Runs a move action for a single player.
//
// The first return value is true if the trainer running a turn after this one
// (if this is not the last turn) should continue their turn, and false
// otherwise. This might be false if this move made the other Pokemon faint.
func (tp *turnProcessor) runMove(ctx context.Context, user, target *battleTrainerData, move pkmn.Move) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	var mr pkmn.MoveReport
	var err error

	if user.activePkmnBattleInfo().GetPokemonBattleInfo().CurrHP <= 0 {
		// The trainer tried to have a fainted Pokemon use a move
		err = messaging.SendTempl(client, user.lastContactURL, messaging.TemplMessage{
			Templ:     faintedPokemonUsingMoveTemplate,
			TemplInfo: nil})
		if err != nil {
			return false, handlerError{user: "failed to populate fainted Pokemon using move template", err: err}
		}
		return false, nil
	}

	// Use the move
	mr, err = pkmn.RunMove(user.activePkmn().GetPokemon(), target.activePkmn().GetPokemon(),
		user.activePkmnBattleInfo().GetPokemonBattleInfo(), target.activePkmnBattleInfo().GetPokemonBattleInfo(), move)
	if err != nil {
		return false, handlerError{user: "could not run move", err: err}
	}

	// Each action in the template prefaces itself with the name of the Pokemon
	// in question. For instance, "the wild bulbsaur used tackle" or
	// "ash.ketchum's pikachu is poisoned!". We need to figuire out which of
	// these prefixes is appopriate for each Pokemon
	userActionPrefix := user.trainer.GetTrainer().Name + "'s"
	if user.trainer.GetTrainer().Type == pkmn.WildTrainerType {
		userActionPrefix = "The wild "
	}
	targetActionPrefix := target.trainer.GetTrainer().Name + "'s"
	if target.trainer.GetTrainer().Type == pkmn.WildTrainerType {
		targetActionPrefix = "The wild "
	}

	// Send the move report
	templInfo := struct {
		pkmn.MoveReport
		UserActionPrefix   string
		UserPokemonName    string
		UserHPBar          string
		TargetActionPrefix string
		TargetPokemonName  string
		MoveName           string
	}{
		MoveReport:         mr,
		UserActionPrefix:   userActionPrefix,
		UserPokemonName:    user.activePkmn().GetPokemon().Name,
		UserHPBar:          makeTextHPBar(target.activePkmn().GetPokemon(), target.activePkmnBattleInfo().GetPokemonBattleInfo()),
		TargetActionPrefix: targetActionPrefix,
		TargetPokemonName:  target.activePkmn().GetPokemon().Name,
		MoveName:           move.Name}
	err = messaging.SendTempl(client, user.lastContactURL, messaging.TemplMessage{
		Templ:     moveReportTemplate,
		TemplInfo: templInfo,
		Public:    true})
	if err != nil {
		return false, handlerError{user: "could not populate move report template", err: err}
	}

	if mr.TargetFainted {
		// The target fainted, so we need to tell the caller that this move
		// should mark the end of the turn
		return false, nil
	}

	return true, nil
}

// Runs a switch action for a single player.
func (tp *turnProcessor) runSwitch(ctx context.Context, user *battleTrainerData) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Switch Pokemon
	prevPkmn := user.battleInfo.GetTrainerBattleInfo().CurrPkmnSlot
	newPkmn := user.battleInfo.GetTrainerBattleInfo().NextBattleAction.Val
	user.battleInfo.GetTrainerBattleInfo().CurrPkmnSlot = newPkmn

	var err error

	// Send the switch message
	templInfo := struct {
		Switcher         string
		WithdrawnPokemon string
		SelectedPokemon  string
	}{
		Switcher:         user.trainer.GetTrainer().Name,
		WithdrawnPokemon: user.pkmn[prevPkmn].GetPokemon().Name,
		SelectedPokemon:  user.pkmn[newPkmn].GetPokemon().Name}
	err = messaging.SendTempl(client, user.lastContactURL, messaging.TemplMessage{
		Templ:     switchPokemonTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return handlerError{user: "could not populate switch Pokemon template", err: err}
	}

	return nil
}

// Runs a catch action for a single player
func (tp *turnProcessor) runCatch(ctx context.Context, user, target *battleTrainerData) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Find the catch rate of the trainer's current Pokemon
	catchRate := pkmn.CatchRate(*target.activePkmn().GetPokemon(),
		*target.activePkmnBattleInfo().GetPokemonBattleInfo())
	// Generate a random number indicating whether or not the user caught the
	// Pokemon
	catchScore := rand.Float64() * 256.0

	// Whether or not the user caught the Pokemon decides the template we will
	// send them
	var success bool
	var templ *template.Template
	if catchScore <= catchRate {
		// The Pokemon was caught
		success = true
		templ = pokemonCaughtTemplate
		// The caught Pokemon is essentially out of commission for this battle
		target.activePkmnBattleInfo().GetPokemonBattleInfo().CurrHP = 0
		// Give the Pokemon to the trainer
		tp.Log.Infof(ctx, "giving %v the %v", user.trainer.GetTrainer().Name, target.activePkmn().GetPokemon().Name)
		user.pkmn, success = givePokemon(user.pkmn, tp.DB.NewPokemon(*target.activePkmn().GetPokemon()))
		if !success {
			return false, handlerError{user: "trainer already has the maximum amount of Pokemon", err: errors.New("trainer already has the maximum amount of Pokemon")}
		}
	} else {
		// The Pokemon was not caught
		success = false
		templ = pokemonNotCaughtTemplate
	}

	// Send the message containing the results
	err := messaging.SendTempl(client, user.lastContactURL, messaging.TemplMessage{
		Templ:     templ,
		TemplInfo: target.activePkmn().GetPokemon().Name})
	if err != nil {
		return false, handlerError{user: "could not populate a Pokemon caught template", err: err}
	}

	// The turn should end prematurely if the catch was successful
	return !success, nil
}

// runTurn runs the turn of the user on the target. It returns true if the turn
// should continue.
//
// The given move object is only used as appropriate, so it's acceptable to
// pass an empty move object if that trainer is not using a move.
func (tp *turnProcessor) runTurn(ctx context.Context, user, target *battleTrainerData, move pkmn.Move) (bool, error) {
	switch user.battleInfo.GetTrainerBattleInfo().NextBattleAction.Type {
	case pkmn.MoveBattleActionType:
		return tp.runMove(ctx, user, target, move)
	case pkmn.SwitchBattleActionType:
		return true, tp.runSwitch(ctx, user)
	case pkmn.CatchBattleActionType:
		return tp.runCatch(ctx, user, target)
	default:
		panic("unsupported battle action type")
	}
}

// doesCurrTrainerGoFirst returns true if the current trainer should have their
// battle action performed first.
//
// The given move objects are only used if necessary. Passing an empty object
// is appropriate if that trainer is not using a move this turn.
func (tp *turnProcessor) doesCurrTrainerGoFirst(ctx context.Context, bd *battleData, currMove, opponentMove pkmn.Move) bool {
	// Make this info quicker to access
	currAction := bd.requester.battleInfo.GetTrainerBattleInfo().NextBattleAction.Type
	opponentAction := bd.opponent.battleInfo.GetTrainerBattleInfo().NextBattleAction.Type

	// Calculate the base priority of the trainers' actions.
	currPriority := pkmn.BattleActionTypePriority(currAction)
	opponentPriority := pkmn.BattleActionTypePriority(opponentAction)

	if currPriority != opponentPriority {
		// One battle action should go first by virtue of its type (i.e.
		// switching goes before moves), so we don't have to do anything else.
		return currPriority > opponentPriority
	}

	// The two battle actions have the same base priority, so we have to do
	// some further processing. Each combination of battle actions have their
	// own rules.

	// Since only human trainers should have the ability to catch Pokemon
	// and only non-human trainers can have their Pokemon caught, there's
	// no correct way that this should happen.
	if currAction == pkmn.CatchBattleActionType && opponentAction == pkmn.CatchBattleActionType {
		tp.Log.Warningf(ctx, "two trainers were able to attempt a valid catch action against each other, which should be impossible: %s and %s",
			bd.requester.trainer.GetTrainer().Name, bd.opponent.trainer.GetTrainer().Name)

		return true
	}

	// When two trainers are switching Pokemon on the same turn, the order that
	// it happens is inconsequential, so we default to having the current
	// trainer switch first.
	if currAction == pkmn.SwitchBattleActionType && opponentAction == pkmn.SwitchBattleActionType {
		return true
	}

	// When two trainers are using a move, we must consult the nature of the
	// moves and the Pokemon themselves to see who goes first.
	if currAction == pkmn.MoveBattleActionType && opponentAction == pkmn.MoveBattleActionType {
		goesFirst := pkmn.CalcMoveOrder(*bd.requester.activePkmn().GetPokemon(),
			*bd.opponent.activePkmn().GetPokemon(),
			*bd.requester.activePkmnBattleInfo().GetPokemonBattleInfo(),
			*bd.opponent.activePkmnBattleInfo().GetPokemonBattleInfo(),
			currMove, opponentMove)

		// Return true if goesFirst is 1, meaning that the first trainer we
		// gave CalcMoveOrder goes first, which is the current trainer.
		return goesFirst == 1
	}

	panic("unsupported combination of battle actions")
}

// process runs all required steps for a single battle turn.
//
// If the first return value is true, then the battle has ended as a result of
// this turn.
func (tp *turnProcessor) process(ctx context.Context, bd *battleData) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	var err error

	tp.Log.Infof(ctx, "the turn will be processed now")

	// Make these values easier to access
	curr := bd.requester
	opponent := bd.opponent

	// Load trainer moves if necessary
	var currPkmnMove, opponentPkmnMove pkmn.Move
	if curr.battleInfo.GetTrainerBattleInfo().NextBattleAction.Type == pkmn.MoveBattleActionType {
		// Current trainer is using a move, so we need to load it
		currPkmnMove, err = loadMove(ctx, client, tp.Fetcher, curr.battleInfo.GetTrainerBattleInfo().NextBattleAction.Val)
		if err != nil {
			return false, err
		}
	}
	if opponent.battleInfo.GetTrainerBattleInfo().NextBattleAction.Type == pkmn.MoveBattleActionType {
		// Opponent is using a move, so we need to load it
		opponentPkmnMove, err = loadMove(ctx, client, tp.Fetcher, opponent.battleInfo.GetTrainerBattleInfo().NextBattleAction.Val)
		if err != nil {
			return false, err
		}
	}

	if opponent.trainer.GetTrainer().Type != pkmn.HumanTrainerType {
		// Opponent is a bot of some kind, so all messages should go to the
		// current trainer. We will reroute opponent messages to their URL.
		opponent.lastContactURL = curr.lastContactURL
	}

	// Run the trainers' turns in an order depending on who goes first.
	currGoesFirst := tp.doesCurrTrainerGoFirst(ctx, bd, currPkmnMove, opponentPkmnMove)
	if currGoesFirst {
		runNextAction, err := tp.runTurn(ctx, curr, opponent, currPkmnMove)
		if err != nil {
			return false, err
		}
		if runNextAction {
			_, err = tp.runTurn(ctx, opponent, curr, opponentPkmnMove)
		}
	} else {
		runNextAction, err := tp.runTurn(ctx, opponent, curr, opponentPkmnMove)
		if err != nil {
			return false, err
		}
		if runNextAction {
			_, err = tp.runTurn(ctx, curr, opponent, currPkmnMove)
			if err != nil {
				return false, err
			}
		}
	}

	// Check if the requester has lost
	battleOver := false
	var lostTrainerName, wonTrainerName string
	playerLost, err := tp.checkIfPlayerLost(ctx, bd, curr, opponent)
	if err != nil {
		return false, err
	}
	if playerLost {
		lostTrainerName = curr.trainer.GetTrainer().Name
		wonTrainerName = opponent.trainer.GetTrainer().Name
		battleOver = true
	}
	// Check if the opponent has lost
	playerLost, err = tp.checkIfPlayerLost(ctx, bd, opponent, curr)
	if err != nil {
		return false, err
	}
	if playerLost {
		lostTrainerName = opponent.trainer.GetTrainer().Name
		wonTrainerName = curr.trainer.GetTrainer().Name
		battleOver = true
	}

	if battleOver {
		// Put the trainers back in waiting mode
		curr.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
		opponent.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode

		// Send a notification that the battle is over
		trainerLostTemplInfo := struct {
			LostTrainer string
			WonTrainer  string
		}{
			LostTrainer: lostTrainerName,
			WonTrainer:  wonTrainerName}
		err = messaging.SendTempl(client, curr.lastContactURL, messaging.TemplMessage{
			Type:      messaging.Good,
			Templ:     trainerLostTemplate,
			TemplInfo: trainerLostTemplInfo,
			Public:    true})
		if err != nil {
			return false, err
		}
	}

	curr.battleInfo.GetTrainerBattleInfo().FinishedTurn = false
	opponent.battleInfo.GetTrainerBattleInfo().FinishedTurn = false

	return battleOver, nil
}
