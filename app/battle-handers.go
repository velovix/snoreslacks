package app

import (
	"bytes"
	"errors"
	"net/http"

	"golang.org/x/net/context"
)

func challengeHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Check if the command was used correctly
	if len(r.commandParams) != 1 {
		regularSlackRequest(client, r.responseURL, "invalid number of parameters in command")
		return
	}

	opponentName := r.commandParams[0]

	// Check if the opponent exists
	_, found, err := db.loadTrainer(ctx, opponentName)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not read oppponent trainer info")
		log.errorf(ctx, "%s", err)
		return
	}
	if !found {
		// Construct the template notifying the trainer that the opponent
		// doesn't exist
		templData := &bytes.Buffer{}
		err := noSuchTrainerExistsTemplate.Execute(templData, "")
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not populate no such trainer exists template")
			log.errorf(ctx, "%s", err)
			return
		}
		regularSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
		return
	}

	// Check if the opponent is already waiting on trainer to join a battle
	b, found, err := db.loadBattle(ctx, opponentName, currTrainer.Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not check for a waiting battle")
		log.errorf(ctx, "%s", err)
		return
	}

	currTrainer.Mode = battlingTrainerMode

	templData := &bytes.Buffer{}

	if found {
		// We will join an existing battle

		b.Mode = startedBattleMode // Start the battle

		// Notify everyone that a battle has started
		err := battleStartedTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not populate battle started template")
			log.errorf(ctx, "%s", err)
			return
		}
	} else {
		// We will create a new battle and wait for the opponent

		b = battle{
			P1:   trainerBattleInfo{Name: currTrainer.Name},
			P2:   trainerBattleInfo{Name: opponentName},
			Mode: waitingBattleMode}

		// Notify everyone that the trainer is waiting for a battle
		err := waitingForBattleTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not populate waiting for battle template")
			log.errorf(ctx, "%s", err)
			return
		}
	}

	err = publicSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if the request was received
		err = db.saveBattle(ctx, b)
		if err != nil {
			log.errorf(ctx, "%s", err)
			return
		}
		err = db.saveTrainer(ctx, currTrainer)
		if err != nil {
			log.errorf(ctx, "%s", err)
			return
		}
	} else {
		log.errorf(ctx, "%s", err)
	}
}

func forfeitHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Load the battle the trainer is in
	b, found, err := db.loadBattleTrainerIsIn(ctx, currTrainer.Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not load battle info")
		log.errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.LastContactURL, "trainer is in battle mode but isn't in a battle")
		log.errorf(ctx, "trainer is in battle mode but isn't in a battle")
		return
	}

	var opponentName string
	if currTrainer.Name == b.P1.Name {
		opponentName = b.P2.Name
	} else {
		opponentName = b.P1.Name
	}

	// Get the trainers out of battle mode
	currTrainer.Mode = waitingTrainerMode

	templData := &bytes.Buffer{}

	if b.Mode == waitingBattleMode {
		// The battle hadn't started yet, so nobody loses

		// Construct the template letting everyone know that the trainer
		// forfeitted
		err := waitingForfeitTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not populate waiting forfeit template")
			log.errorf(ctx, "%s", err)
			return
		}

		err = publicSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
		if err == nil {
			// Save trainers and battle if slack received the request
			err = db.saveTrainer(ctx, currTrainer)
			if err != nil {
				log.errorf(ctx, "%s", err)
				return
			}
			err = db.saveBattle(ctx, b)
			if err != nil {
				log.errorf(ctx, "%s", err)
				return
			}
		}
	} else if b.Mode == startedBattleMode {
		// Load the opponent
		opponent, found, err := db.loadTrainer(ctx, opponentName)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not read opponent info")
			log.errorf(ctx, "%s", err)
			return
		}
		if !found {
			regularSlackRequest(client, currTrainer.LastContactURL, "opponent does not exist")
			log.errorf(ctx, "opponent does not exist")
			return
		}

		// Take the opponent out of battle mode
		opponent.Mode = waitingTrainerMode

		// Count this as a loss for the trainer
		currTrainer.Losses++
		opponent.Wins++

		err = battlingForfeitTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.LastContactURL, "could not populate battling forfeit template")
			log.errorf(ctx, "%s", err)
			return
		}

		err = publicSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
		if err == nil {
			// Save trainers and battle if slack received the request
			err = db.saveTrainer(ctx, currTrainer)
			if err != nil {
				log.errorf(ctx, "%s", err)
				return
			}
			err = db.saveTrainer(ctx, opponent)
			if err != nil {
				log.errorf(ctx, "%s", err)
				return
			}
			err = db.saveBattle(ctx, b)
			if err != nil {
				log.errorf(ctx, "%s", err)
				return
			}
		}
	}
}

func useMoveHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Find the battle the player is in
	b, found, err := db.loadBattleTrainerIsIn(ctx, currTrainer.Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not read battle info")
		log.errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.LastContactURL, "trainer is in battle mode but isn't in a battle")
		log.errorf(ctx, "%s", errors.New("trainer is in battle mode but isn't in a battle"))
		return
	}

	// Find which player the trainer is
	var trainerInfo *trainerBattleInfo
	if currTrainer.Name == b.P1.Name {
		trainerInfo = &b.P1
	} else {
		trainerInfo = &b.P2
	}

	trainerInfo.FinishedTurn = true
	trainerInfo.NextBattleAction = battleAction{
		Type: moveBattleActionType,
		Val:  5}

	templData := &bytes.Buffer{}
	err = moveConfirmationTemplate.Execute(templData, "some crap")
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate move confirmation template")
		log.errorf(ctx, "%s", err)
		return
	}

	err = publicSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if Slack got the request
		err = db.saveBattle(ctx, b)
		if err != nil {
			log.errorf(ctx, "%s", err)
			return
		}
	} else {
		log.errorf(ctx, "%s", err)
	}
}

func switchPokemonHandler(ctx context.Context, db dao, log logger, client *http.Client, r slackRequest, currTrainer trainer) {
	// Find the battle the player is in
	b, found, err := db.loadBattleTrainerIsIn(ctx, currTrainer.Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not read battle info")
		log.errorf(ctx, "%s", err)
		return
	}
	if !found {
		regularSlackRequest(client, currTrainer.LastContactURL, "trainer is in battle mode but isn't in a battle")
		log.errorf(ctx, "%s", errors.New("trainer is in battle mode but isn't in a battle"))
		return
	}

	// Find which player the trainer is
	var trainerInfo *trainerBattleInfo
	if currTrainer.Name == b.P1.Name {
		trainerInfo = &b.P1
	} else {
		trainerInfo = &b.P2
	}

	trainerInfo.FinishedTurn = true
	trainerInfo.NextBattleAction = battleAction{
		Type: switchBattleActionType,
		Val:  5}

	templData := &bytes.Buffer{}
	err = switchConfirmationTemplate.Execute(templData, "some crap")
	if err != nil {
		regularSlackRequest(client, currTrainer.LastContactURL, "could not populate switch confirmation template")
		log.errorf(ctx, "%s", err)
		return
	}

	err = publicSlackRequest(client, currTrainer.LastContactURL, string(templData.Bytes()))
	if err == nil {
		// Save data if Slack got the request
		err = db.saveBattle(ctx, b)
		if err != nil {
			log.errorf(ctx, "%s", err)
			return
		}
	} else {
		log.errorf(ctx, "%s", err)
	}

}
