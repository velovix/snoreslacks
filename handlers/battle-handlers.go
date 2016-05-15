package handlers

import (
	"errors"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
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

// processTurn runs all required steps for a single battle turn, and returns
// the modified Pokemon battle info. This function does not propigate errors
// It deals with them like a handler does, so the caller should not worry about
// notifying the user if this function fails. If the third return value is
// true, then the battle is over and should be stopped. If the fourth return
// value is false, that should be taken as a message that the current operation
// should be aborted.
//
// This function should not save information to the database. Instead, it
// should return data that needs to be saved for the caller to deal with.
func processTurn(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	p1, p2 trainerData, p1BI, p2BI database.TrainerBattleInfo,
	b database.Battle) (database.PokemonBattleInfo, database.PokemonBattleInfo, bool, bool) {

	var err error

	// Get the trainers' current Pokemon
	p1Pkmn := p1.pkmn[p1BI.GetTrainerBattleInfo().CurrPkmnSlot]
	p2Pkmn := p2.pkmn[p2BI.GetTrainerBattleInfo().CurrPkmnSlot]

	// Loads the battle info of a given Pokemon. Returns the battle info or
	// false if something went wrong. This function does the right thing in
	// case of errors so the caller can operate under the assumption that the
	// user has been notified and that the current action should be aborted.
	loadPokemonBI := func(pkmn database.Pokemon) (database.PokemonBattleInfo, bool) {
		pkmnBI, _, err := loadPokemonBattleInfo(ctx, db, log, client, p1.lastContactURL, true,
			"could not get Pokemon battle info",
			b, pkmn.GetPokemon().UUID)
		if err != nil {
			return nil, false // Abort operation
		}
		return pkmnBI, true
	}

	// Create a helper function for loading moves
	loadMove := func(id int) (pkmn.Move, error) {
		// Load the move from PokeAPI
		apiMove, err := fetcher.FetchMove(ctx, client, id)
		if err != nil {
			return pkmn.Move{}, err
		}
		// Use the PokeAPI data to create a pkmn.Move
		move, err := pokeapi.NewMove(apiMove)
		if err != nil {
			return pkmn.Move{}, err
		}
		return move, nil
	}

	// Load player 1's move, if necessary
	var p1PkmnMove, p2PkmnMove pkmn.Move
	if p1BI.GetTrainerBattleInfo().NextBattleAction.Type == pkmn.MoveBattleActionType {
		// Player 1 is using a move, so we need to load it
		p1PkmnMove, err = loadMove(p1BI.GetTrainerBattleInfo().NextBattleAction.Val)
		if err != nil {
			regularSlackRequest(client, p1.lastContactURL, "could not load move information")
			log.Errorf(ctx, "%s", err)
			return nil, nil, false, false
		}
	}
	if p2BI.GetTrainerBattleInfo().NextBattleAction.Type == pkmn.MoveBattleActionType {
		// Player 2 is using a move, so we need to load it
		p2PkmnMove, err = loadMove(p2BI.GetTrainerBattleInfo().NextBattleAction.Val)
		if err != nil {
			regularSlackRequest(client, p1.lastContactURL, "could not load move information")
			log.Errorf(ctx, "%s", err)
			return nil, nil, false, false
		}
	}

	// Runs a move turn for a single player. The first return value is false if
	// the next Pokemon should not be allowed to use their move (generally
	// meaning that it fainted. The second returns false if something didn't
	// succeed. This function does the right thing in case of errors, so the
	// only thing callers have to do is operate under the assumption that this
	// operation should be aborted.
	runPlayerMove := func(t trainerData, tBI database.TrainerBattleInfo,
		opponent trainerData,
		userPkmn, targetPkmn database.Pokemon,
		userPkmnBI, targetPkmnBI database.PokemonBattleInfo,
		move pkmn.Move) (bool, bool) {

		var mr pkmn.MoveReport

		if userPkmnBI.GetPokemonBattleInfo().CurrHP <= 0 {
			// The trainer tried to have a fainted Pokemon use a move
			err = regularSlackTemplRequest(client, t.lastContactURL, faintedPokemonUsingMoveTemplate, nil)
			if err != nil {
				regularSlackRequest(client, t.lastContactURL, "failed to populate failed Pokemon using move template")
				log.Errorf(ctx, "while sending a fainted Pokemon using move template: %s", err)
				return false, false
			}
			return false, false
		}

		// Use the move
		mr, err = pkmn.RunMove(userPkmn.GetPokemon(), targetPkmn.GetPokemon(),
			userPkmnBI.GetPokemonBattleInfo(), targetPkmnBI.GetPokemonBattleInfo(), move)
		if err != nil {
			regularSlackRequest(client, t.lastContactURL, "could not run move")
			log.Errorf(ctx, "while running a move: %s", err)
			return false, false
		}

		// Send the move report
		templInfo := struct {
			pkmn.MoveReport
			UserTrainerName   string
			UserPokemonName   string
			UserHPBar         string
			TargetTrainerName string
			TargetPokemonName string
			MoveName          string
		}{
			MoveReport:        mr,
			UserTrainerName:   t.GetTrainer().Name,
			UserPokemonName:   userPkmn.GetPokemon().Name,
			UserHPBar:         makeTextHPBar(targetPkmn.GetPokemon(), targetPkmnBI.GetPokemonBattleInfo()),
			TargetTrainerName: opponent.GetTrainer().Name,
			TargetPokemonName: targetPkmn.GetPokemon().Name,
			MoveName:          move.Name}
		err = publicSlackTemplRequest(client, t.lastContactURL, moveReportTemplate, templInfo)
		if err != nil {
			regularSlackRequest(client, t.lastContactURL, "could not populate move report template")
			log.Errorf(ctx, "while sending a move report: %s", err)
			return false, false
		}

		if mr.TargetFainted {
			// The target fainted, so we need to tell the caller that this move
			// should mark the end of the turn
			return false, true
		}

		return true, true
	}

	// Runs a switch turn for a single player. Returns false if something didn't
	// succeed. This function does the right thing in case of errors, so the
	// only thing callers have to do is operate under the assumption that this
	// operation should be aborted.
	runPlayerSwitch := func(t trainerData, tBI database.TrainerBattleInfo) bool {
		// Switch Pokemon
		prevPkmn := tBI.GetTrainerBattleInfo().CurrPkmnSlot
		newPkmn := tBI.GetTrainerBattleInfo().NextBattleAction.Val
		tBI.GetTrainerBattleInfo().CurrPkmnSlot = newPkmn

		// Send the switch message
		templInfo := struct {
			Switcher         string
			WithdrawnPokemon string
			SelectedPokemon  string
		}{
			Switcher:         t.GetTrainer().Name,
			WithdrawnPokemon: t.pkmn[prevPkmn].GetPokemon().Name,
			SelectedPokemon:  t.pkmn[newPkmn].GetPokemon().Name}
		err = regularSlackTemplRequest(client, t.lastContactURL, switchPokemonTemplate, templInfo)
		if err != nil {
			regularSlackRequest(client, t.lastContactURL, "could not populate switch Pokemon template")
			log.Errorf(ctx, "while sending a switch Pokemon template: %s", err)
			return false
		}

		return true
	}

	var p1PkmnBI, p2PkmnBI database.PokemonBattleInfo
	var success bool

	// Check what battle action player 1 is doing
	switch p1BI.GetTrainerBattleInfo().NextBattleAction.Type {
	case pkmn.SwitchBattleActionType:
		// Check what battle action player 2 is doing
		switch p2BI.GetTrainerBattleInfo().NextBattleAction.Type {
		case pkmn.SwitchBattleActionType:
			// Both players are switching, meaning that order doesn't matter
			// and player 1 will go first by default

			// Switch Pokemon
			if !runPlayerSwitch(p1, p1BI) || !runPlayerSwitch(p2, p2BI) {
				return nil, nil, false, false
			}
		case pkmn.MoveBattleActionType:
			// Player 1 is switching and player 2 is using a move. Player 1
			// will go first because switching always runs first.

			// Switch Pokemon
			if !runPlayerSwitch(p1, p1BI) {
				return nil, nil, false, false
			}

			// Load the Pokemon battle info now that the switch is complete
			p1PkmnBI, success = loadPokemonBI(p1Pkmn)
			if !success {
				return nil, nil, false, false
			}
			p2PkmnBI, success = loadPokemonBI(p2Pkmn)
			if !success {
				return nil, nil, false, false
			}

			// Run the move
			_, success = runPlayerMove(p2, p2BI, p1, p2Pkmn, p1Pkmn, p2PkmnBI, p1PkmnBI, p2PkmnMove)
			if !success {
				return nil, nil, false, false
			}
		}
	case pkmn.MoveBattleActionType:
		// Check what battle action player 2 is doing
		switch p2BI.GetTrainerBattleInfo().NextBattleAction.Type {
		case pkmn.SwitchBattleActionType:
			// Player 1 is using a move and player 2 is switching. Player 2
			// will go first because switching always runs first.

			// Switch Pokemon
			if !runPlayerSwitch(p2, p2BI) {
				return nil, nil, false, false
			}

			// Load the Pokemon battle info now that the switch is complete
			p1PkmnBI, success = loadPokemonBI(p1Pkmn)
			if !success {
				return nil, nil, false, false
			}
			p2PkmnBI, success = loadPokemonBI(p2Pkmn)
			if !success {
				return nil, nil, false, false
			}

			// Run the move
			_, success = runPlayerMove(p1, p1BI, p2, p1Pkmn, p2Pkmn, p1PkmnBI, p2PkmnBI, p1PkmnMove)
			if !success {
				return nil, nil, false, false
			}
		case pkmn.MoveBattleActionType:
			// Player 1 and player 2 are both using moves. We have to consult
			// the speed of the Pokemon and the priority of the moves to decide
			// who goes first.

			// Load the Pokemon battle info
			p1PkmnBI, success = loadPokemonBI(p1Pkmn)
			if !success {
				return nil, nil, false, false
			}
			p2PkmnBI, success = loadPokemonBI(p2Pkmn)
			if !success {
				return nil, nil, false, false
			}

			// Figure out who goes first
			first := pkmn.CalcMoveOrder(*p1Pkmn.GetPokemon(), *p2Pkmn.GetPokemon(),
				*p1PkmnBI.GetPokemonBattleInfo(), *p2PkmnBI.GetPokemonBattleInfo(), p1PkmnMove, p2PkmnMove)
			// Run both moves in the right order, or just the first if the
			// second Pokemon feints from the attack
			switch first {
			case 1:
				// Player 1 will move first
				runNextMove, success := runPlayerMove(p1, p1BI, p2, p1Pkmn, p2Pkmn, p1PkmnBI, p2PkmnBI, p1PkmnMove)
				if !success {
					return nil, nil, false, false
				}
				if runNextMove {
					_, success = runPlayerMove(p2, p2BI, p1, p2Pkmn, p1Pkmn, p2PkmnBI, p1PkmnBI, p2PkmnMove)
					if !success {
						return nil, nil, false, false
					}
				}
			case 2:
				// Player 2 will move first
				runNextMove, success := runPlayerMove(p2, p2BI, p1, p2Pkmn, p1Pkmn, p2PkmnBI, p1PkmnBI, p2PkmnMove)
				if !success {
					return nil, nil, false, false
				}
				if runNextMove {
					_, success = runPlayerMove(p1, p1BI, p2, p1Pkmn, p2Pkmn, p1PkmnBI, p2PkmnBI, p1PkmnMove)
					if !success {
						return nil, nil, false, false
					}
				}
			}
		}
	}

	// Checks if the given trainer has no more Pokemon able to fight. The first
	// return value is true if the player has lost. The second return value is
	// false if this check failed in some way. This function handles errors
	// properly on its own, so the caller should treat this value as an
	// indicator to abort the current operation.
	checkIfPlayerLost := func(t trainerData, opponent trainerData) (bool, bool) {
		// Get the trainer's party
		party, _, err := loadParty(ctx, db, log, client, t.lastContactURL, true,
			"failed to load party members",
			t.Trainer)

		for _, member := range party {
			// Load the party member's battle info
			pkmnBI, _, err := loadPokemonBattleInfo(ctx, db, log, client, t.lastContactURL, true,
				"failed to load Pokemon battle info",
				b, member.GetPokemon().UUID)
			if err != nil {
				return false, false
			}

			// Check if the Pokemon is still able to fight
			if pkmnBI.GetPokemonBattleInfo().CurrHP > 0 {
				// There's one Pokemon in the party still able to fight, so the
				// trainer has not lost yet
				return false, true
			}
		}

		// The trainer has lost. Let the world know.
		templInfo := struct {
			LostTrainer string
			WonTrainer  string
		}{
			LostTrainer: t.GetTrainer().Name,
			WonTrainer:  opponent.GetTrainer().Name}
		err = regularSlackTemplRequest(client, t.lastContactURL, trainerLostTemplate, templInfo)
		if err != nil {
			regularSlackRequest(client, t.lastContactURL, "could not populate trainer lost template")
			log.Infof(ctx, "while sending trainer lost template: %s", err)
			return false, false
		}

		return true, true
	}

	// Check if a player has lost the battle.
	// TODO(velovix): Add an intelligent response to when two players lose at
	// the same time.
	battleOver := false
	var lostTrainerName, wonTrainerName string
	if p1PkmnBI.GetPokemonBattleInfo().CurrHP <= 0 {
		_, success := checkIfPlayerLost(p1, p2)
		if !success {
			return nil, nil, false, false
		}
		lostTrainerName = p1.GetTrainer().Name
		wonTrainerName = p2.GetTrainer().Name
		battleOver = true
	} else if p2PkmnBI.GetPokemonBattleInfo().CurrHP <= 0 {
		_, success := checkIfPlayerLost(p2, p1)
		if !success {
			return nil, nil, false, false
		}
		battleOver = true
		lostTrainerName = p2.GetTrainer().Name
		wonTrainerName = p1.GetTrainer().Name
	}

	if battleOver {
		// Put the trainers back in waiting mode
		p1.GetTrainer().Mode = pkmn.WaitingTrainerMode
		p2.GetTrainer().Mode = pkmn.WaitingTrainerMode

		// Send a notification that the battle is over
		trainerLostTemplInfo := struct {
			LostTrainer string
			WonTrainer  string
		}{
			LostTrainer: lostTrainerName,
			WonTrainer:  wonTrainerName}
		err = publicSlackTemplRequest(client, p1.lastContactURL, trainerLostTemplate, trainerLostTemplInfo)
		if err != nil {
			regularSlackRequest(client, p1.lastContactURL, "failed to parse trainer lost template")
			log.Errorf(ctx, "while sending trainer lost template: %s", err)
			return nil, nil, false, false
		}
	}

	p1BI.GetTrainerBattleInfo().FinishedTurn = false
	p2BI.GetTrainerBattleInfo().FinishedTurn = false

	return p1PkmnBI, p2PkmnBI, battleOver, true
}

// fisherYates creates a slice of [from, to] and shuffles it using the
// Fisher-Yates algorithm.
func fisherYates(from, to int) []int {
	values := make([]int, to-(from-1))
	for i := range values {
		values[i] = from + i
	}
	for i := len(values) - 1; i >= 1; i-- {
		n := rand.Intn(i + 1)
		temp := values[i]
		values[i] = values[n]
		values[n] = temp
	}

	return values
}

// sendActionOptions sends each player their move and party switching options.
func sendActionOptions(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData, currTrainerBI database.TrainerBattleInfo, b database.Battle) error {

	// Get the current Pokemon
	currPkmn := currTrainer.pkmn[currTrainerBI.GetTrainerBattleInfo().CurrPkmnSlot]

	// Construct a move lookup table
	var mlt pkmn.MoveLookupTable
	mlt.TrainerName = currTrainer.GetTrainer().Name
	mlt.Moves = make([]pkmn.MoveLookupElement, currPkmn.GetPokemon().MoveCount())

	// Construct the move lookup elements
	moveOrder := fisherYates(1, currPkmn.GetPokemon().MoveCount())
	moves := currPkmn.GetPokemon().MoveIDsAsSlice()
	for i, moveID := range moves {
		// Fetch move info from PokeAPI
		apiMove, err := fetcher.FetchMove(ctx, client, moveID)
		if err != nil {
			return err
		}

		// Create a pkmn.Move from the PokeAPI data
		move, err := pokeapi.NewMove(apiMove)
		if err != nil {
			return err
		}

		mlt.Moves[i] = pkmn.MoveLookupElement{
			ID:       moveOrder[i],
			MoveID:   moveID,
			MoveName: move.Name}
	}

	// Fetch the trainer's party
	party, _, err := db.LoadParty(ctx, currTrainer.Trainer)
	if err != nil {
		return err
	}

	// Construct a party member lookup table
	var pmlt pkmn.PartyMemberLookupTable
	pmlt.TrainerName = currTrainer.GetTrainer().Name
	pmlt.Members = make([]pkmn.PartyMemberLookupElement, len(party))

	// Construct the party member lookup elements
	partyOrder := fisherYates(1, len(party))
	for i, val := range party {
		element := pkmn.PartyMemberLookupElement{
			ID:       partyOrder[i],
			SlotID:   i,
			PkmnName: val.GetPokemon().Name}
		pmlt.Members[i] = element
	}

	templInfo := struct {
		CurrPokemonName string
		MoveTable       pkmn.MoveLookupTable
		PartyTable      pkmn.PartyMemberLookupTable
	}{
		CurrPokemonName: currPkmn.GetPokemon().Name,
		MoveTable:       mlt,
		PartyTable:      pmlt}
	err = regularSlackTemplRequest(client, currTrainer.lastContactURL, actionOptionsTemplate, templInfo)
	if err != nil {
		return err
	}

	// Save data if all went well
	err = db.SaveMoveLookupTable(ctx, db.NewMoveLookupTable(mlt), b)
	if err != nil {
		return err
	}
	err = db.SavePartyMemberLookupTable(ctx, db.NewPartyMemberLookupTable(pmlt), b)
	if err != nil {
		return err
	}

	return nil
}

// challengeHandler handles requests to start a battle with another trainer.
func challengeHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData) {

	// Check if the command was used correctly
	if len(r.commandParams) != 1 {
		regularSlackRequest(client, r.responseURL, "invalid number of parameters in command")
		return
	}

	opponentName := r.commandParams[0]

	// Check if the opponent exists
	opponent, found, err := buildTrainerData(ctx, db, opponentName)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not read oppponent trainer info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		// Construct the template notifying the trainer that the opponent
		// doesn't exist
		err := regularSlackTemplRequest(client, currTrainer.lastContactURL, noSuchTrainerExistsTemplate, opponentName)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate no such trainer exists template")
			log.Errorf(ctx, "while sending a no such trainer exists template: %s", err)
			return
		}
		return
	}

	// Check if the opponent is already waiting on trainer to join a battle
	b, found, err := loadBattle(ctx, db, log, client, currTrainer.lastContactURL, false,
		"could not check for a waiting battle",
		opponentName, currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}
	var p1BattleInfo database.TrainerBattleInfo
	var p2BattleInfo database.TrainerBattleInfo

	currTrainer.GetTrainer().Mode = pkmn.BattlingTrainerMode

	// Battle info for every trainer's Pokemon. This will only get filled if
	// a battle is starting
	var pkmnBattleInfos []pkmn.PokemonBattleInfo

	if found {
		// We will join an existing battle

		log.Infof(ctx, "joining an existing battle: %+v", b)

		// Load the player battle info
		p1BattleInfo, _, err = loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
			"could not load player 1 battle info",
			b, b.GetBattle().P1)
		if err != nil {
			return // Abort operation
		}
		p2BattleInfo, _, err = loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
			"could not load player 2 battle info",
			b, b.GetBattle().P2)
		if err != nil {
			return // Abort operation
		}

		log.Infof(ctx, "current trainer has %s Pokemon", len(currTrainer.pkmn))
		log.Infof(ctx, "opponent has %s Pokemon", len(opponent.pkmn))

		// Make the battle info for each Pokemon
		for _, p := range currTrainer.pkmn {
			pkmnBattleInfos = append(pkmnBattleInfos, pkmn.PokemonBattleInfo{
				PkmnUUID: p.GetPokemon().UUID,
				CurrHP:   pkmn.CalcOOBHP(p.GetPokemon().HP, *p.GetPokemon())})
		}
		for _, p := range opponent.pkmn {
			pkmnBattleInfos = append(pkmnBattleInfos, pkmn.PokemonBattleInfo{
				PkmnUUID: p.GetPokemon().UUID,
				CurrHP:   pkmn.CalcOOBHP(p.GetPokemon().HP, *p.GetPokemon())})
		}

		b.GetBattle().Mode = pkmn.StartedBattleMode // Start the battle

		// Notify everyone that a battle has started
		err = publicSlackTemplRequest(client, currTrainer.lastContactURL, battleStartedTemplate, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battle started template")
			log.Errorf(ctx, "while sending a battle started template: %s", err)
			return
		}

		// Get the battle info of the current trainer
		currTrainerBI := p1BattleInfo
		if currTrainer.GetTrainer().Name == b.GetBattle().P2 {
			currTrainerBI = p2BattleInfo
		}
		// Get the battle info of the opponent
		opponentBI := p1BattleInfo
		if opponentName == b.GetBattle().P2 {
			opponentBI = p2BattleInfo
		}

		// Send action options to the current trainer
		err = sendActionOptions(ctx, db, log, client, r, fetcher, currTrainer, currTrainerBI, b)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
		// Send action options to the opponent
		err = sendActionOptions(ctx, db, log, client, r, fetcher, opponent, opponentBI, b)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	} else {
		// We will create a new battle and wait for the opponent

		b = db.NewBattle(pkmn.Battle{
			P1:   currTrainer.GetTrainer().Name,
			P2:   opponentName,
			Mode: pkmn.WaitingBattleMode})
		p1BattleInfo = db.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{Name: currTrainer.GetTrainer().Name})
		p2BattleInfo = db.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{Name: opponentName})

		log.Infof(ctx, "creating a new battle: %+v", b)

		// Notify everyone that the trainer is waiting for a battle
		err := publicSlackTemplRequest(client, currTrainer.lastContactURL, waitingForBattleTemplate, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting for battle template")
			log.Errorf(ctx, "%s", err)
			return
		}
	}

	// Save data if the request was received
	err = db.SaveBattle(ctx, b)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	log.Infof(ctx, "Type of battle info: %s", reflect.TypeOf(p1BattleInfo))
	err = db.SaveTrainerBattleInfo(ctx, b, p1BattleInfo)
	if err != nil {
		log.Errorf(ctx, "%s", err)
	}
	err = db.SaveTrainerBattleInfo(ctx, b, p2BattleInfo)
	if err != nil {
		log.Errorf(ctx, "%s", err)
	}
	err = db.SaveTrainer(ctx, currTrainer.Trainer)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	for _, pbi := range pkmnBattleInfos {
		err = db.SavePokemonBattleInfo(ctx, b, db.NewPokemonBattleInfo(pbi))
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
}

// forfeitHandler handles requests to forfeit a battle. If the battle is in a
// waiting state, then the trainer stops waiting. If the battle has started, a
// forfeit will result in the trainer losing the match.
func forfeitHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData) {

	// Load the battle the trainer is in
	b, _, err := loadBattleTrainerIsIn(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not load battle info",
		currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}

	// Figure out who the opponent is
	var opponentName string
	if currTrainer.GetTrainer().Name == b.GetBattle().P1 {
		opponentName = b.GetBattle().P2
	} else {
		opponentName = b.GetBattle().P1
	}

	// Take the current trainer out of battle mode
	currTrainer.GetTrainer().Mode = pkmn.WaitingTrainerMode

	if b.GetBattle().Mode == pkmn.WaitingBattleMode {
		// The battle hadn't started yet, so nobody loses

		// Construct the template letting everyone know that the trainer
		// forfeitted
		err := publicSlackTemplRequest(client, currTrainer.lastContactURL, waitingForfeitTemplate, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting forfeit template")
			log.Errorf(ctx, "while sending waiting forfeit template: %s", err)
			return
		} else {
			// Save trainers and battle if slack received the request
			err = db.SaveTrainer(ctx, currTrainer.Trainer)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			// The battle is over, so it should be deleted
			log.Infof(ctx, "deleting a battle %s/%s", b.GetBattle().P1, b.GetBattle().P2)
			err = db.PurgeBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
		}
	} else if b.GetBattle().Mode == pkmn.StartedBattleMode {
		// The battle has started, so the forfeitter will lose

		// Load the opponent
		opponent, _, err := loadTrainer(ctx, db, log, client, currTrainer.lastContactURL, true,
			"could not read opponent info",
			opponentName)
		if err != nil {
			return // Abort operation
		}

		// Take the opponent out of battle mode
		opponent.GetTrainer().Mode = pkmn.WaitingTrainerMode

		// Count this as a loss for the trainer
		currTrainer.GetTrainer().Losses++
		opponent.GetTrainer().Wins++

		battlingForfeitTemplInfo := struct {
			Forfeitter string
			Opponent   string
		}{
			Forfeitter: currTrainer.GetTrainer().Name,
			Opponent:   opponent.GetTrainer().Name}

		err = publicSlackTemplRequest(client, currTrainer.lastContactURL, battlingForfeitTemplate, battlingForfeitTemplInfo)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battling forfeit template")
			log.Errorf(ctx, "while sending battling forfeit template: %s", err)
			return
		} else {
			// Save changes if Slack received the request
			err = db.SaveTrainer(ctx, currTrainer.Trainer)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			err = db.SaveTrainer(ctx, opponent.Trainer)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			// The battle is over, so it should be deleted
			log.Infof(ctx, "deleting a battle %s/%s", b.GetBattle().P1, b.GetBattle().P2)
			err = db.PurgeBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
		}
	}
}

// useMoveHandler handles requests to use a Pokemon move. This function will
// queue up a move to be run once both trainers finish selecting the action
// they will take.
func useMoveHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData) {

	// Check if the command looks correct
	if len(r.commandParams) != 1 {
		err := regularSlackTemplRequest(client, currTrainer.lastContactURL, invalidCommandTemplate, nil)
		if err != nil {
			log.Errorf(ctx, "while sending an invalid command template: %s", err)
			return
		}
		return
	}

	// Find the battle the player is in
	b, _, err := loadBattleTrainerIsIn(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not read battle info",
		currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}

	// Figure out the opponent name
	opponentName := b.GetBattle().P1
	if currTrainer.GetTrainer().Name == b.GetBattle().P1 {
		opponentName = b.GetBattle().P2
	}

	// Load battle info for the trainer
	trainerBattleInfo, _, err := loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not get trainer battle info",
		b, currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}

	// Load trainer info for the opponent
	opponent, found, err := buildTrainerData(ctx, db, opponentName)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not build info for opponent")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not build info for opponent")
		log.Errorf(ctx, "%s", errors.New("trainer is in a battle but their opponent doesn't have trainer info"))
		return
	}
	// Load battle info for the opponent
	opponentBattleInfo, _, err := loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not get trainer battle info",
		b, opponentName)
	if err != nil {
		return // Abort operation
	}

	// Extract the move ID from the command
	scrambledID, err := strconv.Atoi(r.commandParams[0])
	if err != nil {
		err = regularSlackTemplRequest(client, currTrainer.lastContactURL, invalidCommandTemplate, nil)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate invalid command template")
			log.Errorf(ctx, "while sending invalid command template: %s", err)
			return
		}
		return
	}

	// Get all the move lookup tables
	mlts, _, err := loadMoveLookupTables(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not fetch move lookup tables",
		b)
	if err != nil {
		return // Abort operation
	}
	// Find the trainer's move lookup table
	var mlt *pkmn.MoveLookupTable
	for _, val := range mlts {
		if val.GetMoveLookupTable().TrainerName == currTrainer.GetTrainer().Name {
			mlt = val.GetMoveLookupTable()
			break
		}
	}
	if mlt == nil {
		// The trainer doesn't have a move lookup table
		regularSlackRequest(client, currTrainer.lastContactURL, "could not fetch move lookup table")
		log.Errorf(ctx, "%s", errors.New("trainer is in a battle but has no move lookup table"))
		return
	}
	moveID := mlt.Lookup(scrambledID)

	// Set up the next action to be a move action
	trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn = true
	trainerBattleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.MoveBattleActionType,
		Val:  moveID}

	// Get the move information
	apiMove, err := fetcher.FetchMove(ctx, client, trainerBattleInfo.GetTrainerBattleInfo().NextBattleAction.Val)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not fetch move information")
		log.Errorf(ctx, "%s", err)
		return
	}
	move, err := pokeapi.NewMove(apiMove)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not fetch move information")
		log.Errorf(ctx, "%s", err)
		return
	}

	// Send confirmation that the move was received
	err = regularSlackTemplRequest(client, currTrainer.lastContactURL, moveConfirmationTemplate, move.Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate move confirmation template")
		log.Errorf(ctx, "while sending move confirmation template: %s", err)
		return
	}

	// Check if both trainers have chosen their move and run the turns if so
	var p1PkmnBI, p2PkmnBI database.PokemonBattleInfo
	var success, battleOver bool
	if trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn &&
		opponentBattleInfo.GetTrainerBattleInfo().FinishedTurn {

		if currTrainer.GetTrainer().Name == b.GetBattle().P1 {
			// If the current trainer is player one
			p1PkmnBI, p2PkmnBI, battleOver, success = processTurn(ctx, db, log, client, r, fetcher, currTrainer, opponent, trainerBattleInfo, opponentBattleInfo, b)
			if !success {
				return // Abort operation
			}
		} else {
			// If the current trainer is player two
			p1PkmnBI, p2PkmnBI, battleOver, success = processTurn(ctx, db, log, client, r, fetcher, opponent, currTrainer, opponentBattleInfo, trainerBattleInfo, b)
			if !success {
				return // Abort operation
			}
		}
	}

	if battleOver {
		// Get the trainers out of the battle if it is over
		currTrainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
		opponent.GetTrainer().Mode = pkmn.WaitingTrainerMode
	}

	// Save data if all has gone well
	err = db.SaveBattle(ctx, b)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainerBattleInfo(ctx, b, trainerBattleInfo)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainerBattleInfo(ctx, b, opponentBattleInfo)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainer(ctx, currTrainer.Trainer)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainer(ctx, opponent.Trainer)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	if p1PkmnBI != nil {
		err = db.SavePokemonBattleInfo(ctx, b, p1PkmnBI)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
	if p2PkmnBI != nil {
		err = db.SavePokemonBattleInfo(ctx, b, p2PkmnBI)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
	if battleOver {
		// Delete the battle if it has ended
		err = db.PurgeBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
}

// switchPokemonHandler handles requests to switch Pokemon. This function will
// queue up a switch to be run once both trainers finish selecting the action
// they will take.
func switchPokemonHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData) {

	// Find the battle the player is in
	b, _, err := loadBattleTrainerIsIn(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not read battle info",
		currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}

	// Figure out the opponent's name
	opponentName := b.GetBattle().P1
	if b.GetBattle().P1 == currTrainer.GetTrainer().Name {
		opponentName = b.GetBattle().P2
	}

	// Load the opponent
	opponent, found, err := buildTrainerData(ctx, db, opponentName)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get opponent trainer info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get opponent data")
		log.Errorf(ctx, "%s", errors.New("trainer is in a battle with an opponent that doesn't exist"))
		return
	}

	// Load battle info for the trainer
	trainerBattleInfo, _, err := loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not get trainer battle info",
		b, currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}
	// Load battle info for the opponent
	opponentBattleInfo, _, err := loadTrainerBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, true,
		"could not get opponent battle info",
		b, opponentName)
	if err != nil {
		return // Abort operation
	}

	// Set up the next battle action to be a switch Pokemon action
	trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn = true
	trainerBattleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.SwitchBattleActionType,
		Val:  5}

	// Send confirmation that the switch was received
	err = publicSlackTemplRequest(client, currTrainer.lastContactURL, switchConfirmationTemplate, "some crap")
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate switch confirmation template")
		log.Errorf(ctx, "while sending switch confirmation template: %s", err)
		return
	}

	// Check if both trainers have chosen their move and run the turns if so
	var p1PkmnBI, p2PkmnBI database.PokemonBattleInfo
	var success, battleOver bool
	if trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn &&
		opponentBattleInfo.GetTrainerBattleInfo().FinishedTurn {

		if currTrainer.GetTrainer().Name == b.GetBattle().P1 {
			// If the current trainer is player one
			p1PkmnBI, p2PkmnBI, battleOver, success = processTurn(ctx, db, log, client, r, fetcher, currTrainer, opponent, trainerBattleInfo, opponentBattleInfo, b)
			if !success {
				return // Abort operation
			}
		} else {
			// If the current trainer is player two
			p1PkmnBI, p2PkmnBI, battleOver, success = processTurn(ctx, db, log, client, r, fetcher, opponent, currTrainer, opponentBattleInfo, trainerBattleInfo, b)
			if !success {
				return // Abort operation
			}
		}
	}

	if battleOver {
		// Get the trainers out of the battle if it is over
		currTrainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
		opponent.GetTrainer().Mode = pkmn.WaitingTrainerMode
	}

	// Save data if all has gone well
	err = db.SaveTrainer(ctx, currTrainer.Trainer)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainer(ctx, opponent.Trainer)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveBattle(ctx, b)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	err = db.SaveTrainerBattleInfo(ctx, b, trainerBattleInfo)
	if err != nil {
		log.Errorf(ctx, "%s", err)
		return
	}
	if p1PkmnBI != nil {
		err = db.SavePokemonBattleInfo(ctx, b, p1PkmnBI)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
	if p2PkmnBI != nil {
		err = db.SavePokemonBattleInfo(ctx, b, p2PkmnBI)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
	if battleOver {
		err = db.PurgeBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
	}
}
