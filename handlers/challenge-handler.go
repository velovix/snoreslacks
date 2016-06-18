package handlers

import (
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"golang.org/x/net/context"
)

// Challenge handles requests to start a battle with another trainer.
type Challenge struct {
}

// sendInitialPkmnMessage sends a public message alerting everyone of the
// Pokemon the trainer is starting the battle with, along with a sprite of that
// Pokemon. It is assumed that the trainer starts the battle with the first
// Pokemon in the party, like in the main series games.
func (h *Challenge) sendInitialPkmnMessage(ctx context.Context, t *basicTrainerData) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Construct and send message
	initialPokemonTemplInfo := struct {
		TrainerName string
		PokemonName string
	}{
		TrainerName: t.trainer.GetTrainer().Name,
		PokemonName: t.pkmn[0].GetPokemon().Name}
	return messaging.SendTempl(client, t.lastContactURL, messaging.TemplMessage{
		Public:    true,
		Image:     t.pkmn[0].GetPokemon().SpriteURL,
		Templ:     initialPokemonSendOutTemplate,
		TemplInfo: initialPokemonTemplInfo})
}

func (h *Challenge) preprocess(ctx context.Context, s Services) (context.Context, bool, error) {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Check if the command was used correctly
	if len(slackReq.CommandParams) != 1 {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     invalidCommandTemplate,
			TemplInfo: nil})
		if err != nil {
			return ctx, false, handlerError{user: "could not send invalid command template", err: err}
		}
		return ctx, false, nil // No more work to do
	}

	// Trainers are identified by their UUID, not their display name, so we
	// have to do a reverse lookup for the UUID. Slack takes care of making
	// sure usernames are original, so this shouldn't be a problem.
	opponentName := slackReq.CommandParams[0]
	opponentUUID, err := s.DB.LoadUUIDFromHumanTrainerName(ctx, opponentName)
	if database.IsNoResults(err) {
		// Construct the template notifying the trainer that the opponent
		// doesn't exist
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     noSuchTrainerExistsTemplate,
			TemplInfo: opponentName})
		if err != nil {
			return ctx, false, handlerError{user: "could not populate no such trainer exists template", err: err}
		}
		return ctx, false, nil // This request has been finished
	} else if err != nil {
		// Some generic error occurred
		s.Log.Infof(ctx, "an error occurred!")
		return ctx, false, handlerError{user: "could not convert the opponent's username to a UUID", err: err}
	}

	ctx = context.WithValue(ctx, "opponent UUID", opponentUUID)

	return ctx, true, nil
}

func (h *Challenge) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	opponentUUID := ctx.Value("opponent UUID").(string)

	// No checking has to be done on the command itself because, in this case,
	// the preprocessor takes care of that

	// Assert that the trainer is currently in waiting mode
	if requester.trainer.GetTrainer().Mode != pkmn.WaitingTrainerMode {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     challengingWhenInWrongMode,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not send challenging when in wrong mode template", err: err}
		}
		return nil // No more work to do
	}

	// Load some info on the opponent
	opponent, err := loadBasicTrainerData(ctx, s.DB, opponentUUID)
	if err != nil {
		// We know the opponent exists from the preprocessing, so this should
		// not happen unless something is wrong
		return handlerError{user: "could not load opponent information", err: err}
	}

	found := true
	// Check if the opponent is already waiting on trainer to join a battle
	b, err := s.DB.LoadBattle(ctx, opponentUUID, requester.trainer.GetTrainer().UUID)
	if database.IsNoResults(err) {
		// The battle was not found
		found = false
	} else if err != nil {
		// A generic error occurred
		return handlerError{user: "could not check for a waiting battle", err: err}
	}
	var p1BattleInfo database.TrainerBattleInfo
	var p2BattleInfo database.TrainerBattleInfo

	requester.trainer.GetTrainer().Mode = pkmn.BattlingTrainerMode

	// Battle info for every trainer's Pokemon. This will only get filled if
	// a battle is starting
	var pkmnBattleInfos []pkmn.PokemonBattleInfo

	if found {
		// We will join an existing battle

		s.Log.Infof(ctx, "joining an existing battle: %+v", b)

		// Load the player battle info
		p1BattleInfo, err = s.DB.LoadTrainerBattleInfo(ctx, b, b.GetBattle().P1)
		if err != nil {
			return handlerError{user: "could not load player 1 battle info", err: err}
		}
		p2BattleInfo, err = s.DB.LoadTrainerBattleInfo(ctx, b, b.GetBattle().P2)
		if err != nil {
			return handlerError{user: "could not load player 2 battle info", err: err}
		}

		// Make the battle info for each Pokemon
		for _, p := range requester.pkmn {
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
		templInfo := struct {
			Challenger string
			Opponent   string
		}{
			Challenger: requester.trainer.GetTrainer().Name,
			Opponent:   opponent.trainer.GetTrainer().Name}
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     battleStartedTemplate,
			TemplInfo: templInfo,
			Public:    true})
		if err != nil {
			return handlerError{user: "could not populate battle started template", err: err}
		}

		// Send message about the current trainer's first Pokemon
		err = h.sendInitialPkmnMessage(ctx, requester)
		if err != nil {
			// This request is non-critical, so it's okay if it fails
			s.Log.Errorf(ctx, "while sending out information on the first starter: %s", err)
		}
		// Send message about the opponent's first Pokemon
		err = h.sendInitialPkmnMessage(ctx, opponent)
		if err != nil {
			// This request is non-critical, so it's okay if it fails
			s.Log.Errorf(ctx, "while sending out information on the first starter: %s", err)
		}

		// Get the battle info of the current trainer
		requesterBI := p1BattleInfo
		if requester.trainer.GetTrainer().UUID == b.GetBattle().P2 {
			requesterBI = p2BattleInfo
		}
		// Get the battle info of the opponent
		opponentBI := p1BattleInfo
		if opponentUUID == b.GetBattle().P2 {
			opponentBI = p2BattleInfo
		}

		// Make action options for the current trainer
		err = makeActionOptions(ctx, s, requester, requesterBI, b)
		if err != nil {
			return handlerError{user: "could not send action options", err: err}
		}
		// Make action options for the opponent
		err = makeActionOptions(ctx, s, opponent, opponentBI, b)
		if err != nil {
			return handlerError{user: "could not send action options", err: err}
		}
	} else {
		// We will create a new battle and wait for the opponent

		b = s.DB.NewBattle(pkmn.Battle{
			P1:   requester.trainer.GetTrainer().UUID,
			P2:   opponentUUID,
			Mode: pkmn.WaitingBattleMode})
		p1BattleInfo = s.DB.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{TrainerUUID: requester.trainer.GetTrainer().UUID})
		p2BattleInfo = s.DB.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{TrainerUUID: opponentUUID})

		s.Log.Infof(ctx, "creating a new battle: %+v", b)

		// Notify everyone that the trainer is waiting for a battle
		templInfo := struct {
			Challenger string
			Opponent   string
		}{
			Challenger: requester.trainer.GetTrainer().Name,
			Opponent:   opponent.trainer.GetTrainer().Name}
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Type:      messaging.Important,
			Templ:     waitingForBattleTemplate,
			TemplInfo: templInfo,
			Public:    true})
		if err != nil {
			return handlerError{user: "could not populatle waiting for battle template", err: err}
		}
	}

	// Save data if the request was received
	err = s.DB.SaveBattle(ctx, b)
	if err != nil {
		return handlerError{user: "could not save battle", err: err}
	}
	err = s.DB.SaveTrainerBattleInfo(ctx, b, p1BattleInfo)
	if err != nil {
		return handlerError{user: "could not save trainer battle info", err: err}
	}
	err = s.DB.SaveTrainerBattleInfo(ctx, b, p2BattleInfo)
	if err != nil {
		return handlerError{user: "could not save trainer battle info", err: err}
	}
	err = s.DB.SaveTrainer(ctx, requester.trainer)
	if err != nil {
		return handlerError{user: "could not save trainer", err: err}
	}
	for _, pbi := range pkmnBattleInfos {
		err = s.DB.SavePokemonBattleInfo(ctx, b, s.DB.NewPokemonBattleInfo(pbi))
		if err != nil {
			return handlerError{user: "could not save Pokemon battle info", err: err}
		}
	}

	return nil
}
