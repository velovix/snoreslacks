package handlers

import (
	"fmt"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"golang.org/x/net/context"
)

// loadTrainer is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadTrainer(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp, name string) (trainerData, bool, error) {

	errLog := "while loading a trainer"

	t, found, err := buildTrainerData(ctx, db, name)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return trainerData{}, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a trainer with the name '%s' but none exists", name)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return trainerData{}, false, err
		} else {
			return trainerData{}, false, nil
		}
	}

	return t, true, nil
}

// loadPokemon is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadPokemon(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp, uuid string) (database.Pokemon, bool, error) {

	errLog := "while loading a Pokemon"

	b, found, err := db.LoadPokemon(ctx, uuid)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a Pokemon with the UUID '%s' but none exists", uuid)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return b, true, nil
}

// loadParty is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadParty(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp string, t database.Trainer) ([]database.Pokemon, bool, error) {

	errLog := "while loading a party"

	party, found, err := db.LoadParty(ctx, t)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a party under the trainer '%s' but none exists", t.GetTrainer().Name)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return party, true, nil
}

// loadMoveLookupTables is a helper function that does database requests and
// intelligently reacts to the results. If required is true, missing objects
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadMoveLookupTables(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp string, b database.Battle) ([]database.MoveLookupTable, bool, error) {

	errLog := "while loading move lookup tables"

	mlts, found, err := db.LoadMoveLookupTables(ctx, b)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected move lookup tables but none exist")
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return mlts, true, nil
}

// loadBattle is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadBattle(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp, p1Name, p2Name string) (database.Battle, bool, error) {

	errLog := "while loading a battle"

	b, found, err := db.LoadBattle(ctx, p1Name, p2Name)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a battle with the players '%s', '%s' but none exists", p1Name, p2Name)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return b, true, nil
}

// loadBattleTrainerIsIn is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadBattleTrainerIsIn(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp, tName string) (database.Battle, bool, error) {

	errLog := "while loading a battle a trainer is in"

	b, found, err := db.LoadBattleTrainerIsIn(ctx, tName)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a battle that trainer '%s' is in but none exists", tName)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return b, true, nil
}

// loadPartyMemberLookupTables is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadPartyMemberLookupTables(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp string, b database.Battle) ([]database.PartyMemberLookupTable, bool, error) {

	errLog := "while loading party member lookup tables"

	pmlt, found, err := db.LoadPartyMemberLookupTables(ctx, b)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected party member lookup tables but none exists")
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return pmlt, true, nil
}

// loadTrainerBattleInfo is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadTrainerBattleInfo(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp string, b database.Battle, tName string) (database.TrainerBattleInfo, bool, error) {

	log.Infof(ctx, "oh boy! We're loading a trainer battle info!")

	errLog := "while loading trainer battle info"

	tbi, found, err := db.LoadTrainerBattleInfo(ctx, b, tName)
	if err != nil {
		log.Infof(ctx, "Some kind of error happened!")
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		log.Infof(ctx, "The thing wasn't found")
		if required {
			log.Infof(ctx, "The thing wasn't found but is required")
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a trainer battle info for a trainer named '%s' but none exists", tName)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			log.Infof(ctx, "That's okay though, because it isn't required")
			return nil, false, nil
		}
	}

	log.Infof(ctx, "Nothing went wrong. Here is the thing: %v", tbi)

	return tbi, true, nil
}

// loadPokemonBattleInfo is a helper function that does database requests and
// intelligently reacts to the results. If required is true, a missing object
// will result in an error and the second return value can be safely ignored.
// This function will make a regular slack response to the user if an error
// occurs with the given data and log the error with the given information.
func loadPokemonBattleInfo(ctx context.Context, db database.Database, log logging.Logger,
	client *http.Client, url string, required bool,
	errResp string, b database.Battle, uuid string) (database.PokemonBattleInfo, bool, error) {

	errLog := "while loading Pokemon battle info"

	pbi, found, err := db.LoadPokemonBattleInfo(ctx, b, uuid)
	if err != nil {
		regularSlackRequest(client, url, errResp)
		log.Errorf(ctx, "%s: %s", errLog, err)
		return nil, false, err
	}
	if !found {
		if required {
			regularSlackRequest(client, url, errResp)
			err := fmt.Errorf("expected a Pokemon battle info for a Pokemon with the UUID '%s' but none exists", uuid)
			log.Errorf(ctx, "%s: %s", errLog, err)
			return nil, false, err
		} else {
			return nil, false, nil
		}
	}

	return pbi, true, nil
}
