package handlers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// challengeHandler handles requests to start a battle with another ttrainer.
func challengeHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	// Check if the command was used correctly
	if len(r.commandParams) != 1 {
		regularSlackRequest(client, r.responseURL, "invalid number of parameters in command")
		return
	}

	opponentName := r.commandParams[0]

	// Check if the opponent exists
	_, found, err := db.LoadTrainer(ctx, opponentName)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not read oppponent trainer info")
		log.Errorf(ctx, "%s", err)
		return
	}
	if !found {
		// Construct the template notifying the trainer that the opponent
		// doesn't exist
		templData := &bytes.Buffer{}
		err := noSuchTrainerExistsTemplate.Execute(templData, "")
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
		// Send action options to the current trainer
		/*err = sendActionOptions(ctx, db, log, client, r, currTrainer, b)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}
		// Send action options to the opponent
		err = sendActionOptions(ctx, db, log, client, r, opponent, b)
		if err != nil {
			log.Errorf(ctx, "%s", err)
			return
		}*/
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
func forfeitHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
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

		err = battlingForfeitTemplate.Execute(templData, b)
		if err != nil {
			regularSlackRequest(client, currTrainer.lastContactURL, "could not populate battling forfeit template")
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
			err = db.SaveTrainer(ctx, opponent)
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
