package handlers

import (
	"bytes"
	"errors"
	"math/rand"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"

	"golang.org/x/net/context"
)

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
	templData := &bytes.Buffer{}

	err = actionOptionsTemplate.Execute(templData, templInfo)
	if err != nil {
		return err
	}

	err = regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
	if err == nil {
		err = db.SaveMoveLookupTable(ctx, db.NewMoveLookupTable(mlt), b)
		if err != nil {
			return err
		}
		err = db.SavePartyMemberLookupTable(ctx, db.NewPartyMemberLookupTable(pmlt), b)
		if err != nil {
			return err
		}
	}

	return nil
}

func processMoves(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	p1, p2 trainerData, p1Info, p2Info database.TrainerBattleInfo,
	b database.Battle) {

	p1ActionType := p1Info.GetTrainerBattleInfo().NextBattleAction.Type
	p2ActionType := p2Info.GetTrainerBattleInfo().NextBattleAction.Type

	// Check if the turn is in fact over
	if p1ActionType == 0 || p2ActionType == 0 {
		regularSlackRequest(client, p1.lastContactURL, "the server attempted to end the turn too early")
		log.Errorf(ctx, "the server attempted to end a turn too early")
		return
	}

	if p1ActionType == pkmn.MoveBattleActionType && p2ActionType == pkmn.MoveBattleActionType {
		// Both players are using a regular Pokemon move, so priority needs to
		// be calculated
	} else if p1ActionType == pkmn.SwitchBattleActionType && p2ActionType == pkmn.SwitchBattleActionType {
		// Both players are switching Pokemon, so the order doesn't matter, but
		// player 1 will go first by default
	} else if p1ActionType == pkmn.SwitchBattleActionType && p2ActionType == pkmn.MoveBattleActionType {
		// Player 1 is switching out, so they will always go first
	} else if p2ActionType == pkmn.SwitchBattleActionType && p1ActionType == pkmn.MoveBattleActionType {
		// Player 2 is switching out, so they will always go first
	}
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
		templData := &bytes.Buffer{}
		err := noSuchTrainerExistsTemplate.Execute(templData, opponentName)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate no such trainer exists template")
			log.Errorf(ctx, "%s", err)
			return
		}
		regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
		return
	}

	// Check if the opponent is already waiting on trainer to join a battle
	b, found, err := db.LoadBattle(ctx, opponentName, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not check for a waiting battle")
		log.Errorf(ctx, "%s", err)
		return
	}
	var p1BattleInfo database.TrainerBattleInfo
	var p2BattleInfo database.TrainerBattleInfo

	currTrainer.GetTrainer().Mode = pkmn.BattlingTrainerMode

	templData := &bytes.Buffer{}

	if found {
		// We will join an existing battle

		log.Infof(ctx, "joining an existing battle: %+v", b)

		// Load the player battle info
		p1BattleInfo, found, err = db.LoadTrainerBattleInfo(ctx, b, b.GetBattle().P1)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not load player 1 battle info")
			log.Errorf(ctx, "%s", err)
			return
		}
		if !found {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not load player 1 battle info")
			log.Errorf(ctx, "a trainer is in a battle but has no battle info")
			return
		}
		p2BattleInfo, found, err = db.LoadTrainerBattleInfo(ctx, b, b.GetBattle().P2)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not load player 2 battle info")
			log.Errorf(ctx, "%s", err)
			return
		}
		if !found {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not load player 2 battle info")
			log.Errorf(ctx, "a trainer is in a battle but has no battle info")
			return
		}

		b.GetBattle().Mode = pkmn.StartedBattleMode // Start the battle

		// Notify everyone that a battle has started
		err := battleStartedTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battle started template")
			log.Errorf(ctx, "%s", err)
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
		err := waitingForBattleTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting for battle template")
			log.Errorf(ctx, "%s", err)
			return
		}
	}

	err = publicSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if the request was received
		err = db.SaveBattle(ctx, b)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
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
	} else {
		log.Errorf(ctx, "%s", err)
	}
}

// forfeitHandler handles requests to forfeit a battle. If the battle is in a
// waiting state, then the trainer stops waiting. If the battle has started, a
// forfeit will result in the trainer losing the match.
func forfeitHandler(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, r slackRequest, fetcher pokeapi.Fetcher,
	currTrainer trainerData) {

	// Load the battle the trainer is in
	b, found, err := db.LoadBattleTrainerIsIn(ctx, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not load battle info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "trainer is in battle mode but isn't in a battle")
		log.Errorf(ctx, "trainer is in battle mode but isn't in a battle")
		return
	}

	var opponentName string
	if currTrainer.GetTrainer().Name == b.GetBattle().P1 {
		opponentName = b.GetBattle().P2
	} else {
		opponentName = b.GetBattle().P1
	}

	// Get the trainers out of battle mode
	currTrainer.GetTrainer().Mode = pkmn.WaitingTrainerMode

	templData := &bytes.Buffer{}

	if b.GetBattle().Mode == pkmn.WaitingBattleMode {
		// The battle hadn't started yet, so nobody loses

		// Construct the template letting everyone know that the trainer
		// forfeitted
		err := waitingForfeitTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate waiting forfeit template")
			log.Errorf(ctx, "%s", err)
			return
		}

		err = publicSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
		if err == nil {
			// Save trainers and battle if slack received the request
			err = db.SaveTrainer(ctx, currTrainer)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			err = db.SaveBattle(ctx, b)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
		}
	} else if b.GetBattle().Mode == pkmn.StartedBattleMode {
		// The battle has started, so the forfeitter will lose

		// Load the opponent
		opponent, found, err := db.LoadTrainer(ctx, opponentName)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not read opponent info")
			log.Errorf(ctx, "%s", err)
			return
		}
		if !found {
			regularSlackRequest(client, currTrainer.lastContactURL, "opponent does not exist")
			log.Errorf(ctx, "opponent does not exist")
			return
		}

		// Take the opponent out of battle mode
		opponent.GetTrainer().Mode = pkmn.WaitingTrainerMode

		// Count this as a loss for the trainer
		currTrainer.GetTrainer().Losses++
		opponent.GetTrainer().Wins++

		templInfo := struct {
			Forfeitter string
			Opponent   string
		}{
			Forfeitter: currTrainer.GetTrainer().Name,
			Opponent:   opponent.GetTrainer().Name}

		err = battlingForfeitTemplate.Execute(templData, templInfo)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battling forfeit template")
			log.Errorf(ctx, "%s", err)
			return
		}

		err = publicSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
		if err == nil {
			// Save trainers and battle if slack received the request
			err = db.SaveTrainer(ctx, currTrainer.Trainer)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			err = db.SaveTrainer(ctx, opponent)
			if err != nil {
				log.Errorf(ctx, "%s", err)
				return
			}
			err = db.DeleteBattle(ctx, b.GetBattle().P1, b.GetBattle().P2)
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
func useMoveHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Find the battle the player is in
	b, found, err := db.LoadBattleTrainerIsIn(ctx, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not read battle info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "trainer is in battle mode but isn't in a battle")
		log.Errorf(ctx, "%s", errors.New("trainer is in battle mode but isn't in a battle"))
		return
	}

	trainerBattleInfo, found, err := db.LoadTrainerBattleInfo(ctx, b, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get trainer battle info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get trainer battle info")
		log.Errorf(ctx, "%s", errors.New("trainer is in a battle but they don't have battle info"))
		return
	}

	// Set up the next action to be a move action
	trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn = true
	trainerBattleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.MoveBattleActionType,
		Val:  5}

	// Send confirmation that the move was received
	templData := &bytes.Buffer{}
	err = moveConfirmationTemplate.Execute(templData, "some crap")
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate move confirmation template")
		log.Errorf(ctx, "%s", err)
		return
	}

	err = publicSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if Slack got the request
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
	} else {
		log.Errorf(ctx, "%s", err)
	}
}

// switchPokemonHandler handles requests to switch Pokemon. This function will
// queue up a switch too be run once both trainers finish selecting the action
// they will take.
func switchPokemonHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Find the battle the player is in
	b, found, err := db.LoadBattleTrainerIsIn(ctx, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not read battle info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "trainer is in battle mode but isn't in a battle")
		log.Errorf(ctx, "%s", errors.New("trainer is in battle mode but isn't in a battle"))
		return
	}

	// Load battle info for the trainer
	trainerBattleInfo, found, err := db.LoadTrainerBattleInfo(ctx, b, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get trainer battle info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not get trainer battle info")
		log.Errorf(ctx, "%s", errors.New("trainer is in a battle but they don't have battle info"))
		return
	}

	// Set up the next battle action to be a switch Pokemon action
	trainerBattleInfo.GetTrainerBattleInfo().FinishedTurn = true
	trainerBattleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.SwitchBattleActionType,
		Val:  5}

	// Send confirmation that the switch was received
	templData := &bytes.Buffer{}
	err = switchConfirmationTemplate.Execute(templData, "some crap")
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate switch confirmation template")
		log.Errorf(ctx, "%s", err)
		return
	}

	err = publicSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if Slack got the request
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
	} else {
		log.Errorf(ctx, "%s", err)
	}
}
