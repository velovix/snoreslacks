package handlers

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/satori/go.uuid"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
)

// WildEncounter handles requests to battle a wild Pokemon.
type WildEncounter struct {
	Services
}

func (h *WildEncounter) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Assert that the trainer is currently in waiting mode
	if requester.trainer.GetTrainer().Mode != pkmn.WaitingTrainerMode {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     wildEncounterWhenInWrongModeTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not populate wild encounter when in wrong mode template", err: err}
		}
	}

	// Put the trainer in battle mode
	requester.trainer.GetTrainer().Mode = pkmn.BattlingTrainerMode

	// Randomly decide what wild Pokemon the trainer will encounter
	wildEntry := pkmn.RandomWildPokemon(pkmn.AvailableWildPokemon(*requester.trainer.GetTrainer()))
	// Get PokeAPI Data on the Pokemon
	apiPkmn, err := s.Fetcher.FetchPokemon(ctx, client, wildEntry.ID)
	if err != nil {
		return handlerError{user: "unable to fetch Pokemon information", err: err}
	}
	// Create the pkmn.Pokemon from the PokeAPI data
	pokeWild, err := pokeapi.NewPokemon(ctx, client, s.Fetcher, apiPkmn, wildEntry.MedianLevel)
	if err != nil {
		return handlerError{user: "unable to fetch Pokemon information", err: err}
	}
	// Wrap the new Pokemon in a database object
	wild := s.DB.NewPokemon(pokeWild)

	// Create an ephemeral trainer to own the wild Pokemon
	wildTrainer := s.DB.NewTrainer(pkmn.Trainer{
		UUID: uuid.NewV4().String(),
		Name: wild.GetPokemon().Name,
		Mode: pkmn.BattlingTrainerMode,
		Type: pkmn.WildTrainerType})

	// Create a battle between the trainer and wild Pokemon
	b := s.DB.NewBattle(pkmn.Battle{
		P1:   requester.trainer.GetTrainer().UUID,
		P2:   wildTrainer.GetTrainer().UUID,
		Mode: pkmn.StartedBattleMode})
	// Create trainer battle info
	trainerBattleInfo := s.DB.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{TrainerUUID: requester.trainer.GetTrainer().UUID})
	wildBattleInfo := s.DB.NewTrainerBattleInfo(pkmn.TrainerBattleInfo{TrainerUUID: wildTrainer.GetTrainer().UUID})

	s.Log.Infof(ctx, "creating a new battle: %+v", b)

	// Create Pokemon battle info for all the trainer's Pokemon and the wild Pokemon
	pkmnBIs := make([]database.PokemonBattleInfo, 0, 7)
	pkmnBIs = append(pkmnBIs, s.DB.NewPokemonBattleInfo( // Add the wild Pokemon battle info
		pkmn.PokemonBattleInfo{
			PkmnUUID: wild.GetPokemon().UUID,
			CurrHP:   pkmn.CalcOOBHP(wild.GetPokemon().HP, *wild.GetPokemon())}))
	// Add the trainer's Pokemon battle info
	for _, p := range requester.pkmn {
		pkmnBIs = append(pkmnBIs, s.DB.NewPokemonBattleInfo(pkmn.PokemonBattleInfo{
			PkmnUUID: p.GetPokemon().UUID,
			CurrHP:   pkmn.CalcOOBHP(p.GetPokemon().HP, *p.GetPokemon())}))
	}

	// Send message telling the trainer that an encounter happened
	templInfo := struct {
		WildPokemonName string
	}{
		WildPokemonName: wild.GetPokemon().Name}
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     wildBattleStartedTemplate,
		TemplInfo: templInfo,
		Image:     wild.GetPokemon().SpriteURL})
	if err != nil {
		return handlerError{user: "could not populate wild battle started template", err: err}
	}

	// Send the trainer their action options
	err = makeActionOptions(ctx, s, requester, trainerBattleInfo, b)
	if err != nil {
		return handlerError{user: "could not send action options", err: err}
	}

	// Save database objects
	err = s.DB.SaveTrainerBattleInfo(ctx, b, trainerBattleInfo)
	if err != nil {
		return handlerError{user: "could not save trainer battle info", err: err}
	}
	err = s.DB.SaveTrainerBattleInfo(ctx, b, wildBattleInfo)
	if err != nil {
		return handlerError{user: "could not save wild trainer battle info", err: err}
	}
	for _, bi := range pkmnBIs {
		err = s.DB.SavePokemonBattleInfo(ctx, b, bi)
		if err != nil {
			return handlerError{user: "could not save Pokemon battle info", err: err}
		}
	}
	err = s.DB.SavePokemon(ctx, wildTrainer, wild)
	if err != nil {
		return handlerError{user: "could not save Pokemon", err: err}
	}
	err = s.DB.SaveBattle(ctx, b)
	if err != nil {
		return handlerError{user: "could not save battle", err: err}
	}
	err = s.DB.SaveTrainer(ctx, wildTrainer)
	if err != nil {
		return handlerError{user: "could not save wild trainer", err: err}
	}
	err = s.DB.SaveTrainer(ctx, requester.trainer)
	if err != nil {
		return handlerError{user: "could not save trainer", err: err}
	}

	return nil
}

// CatchPokemon handles requests to catch a Pokemon.
type CatchPokemon struct {
	Services
}

func (h *CatchPokemon) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	battleData := ctx.Value("battle data").(*battleData)

	// Assert that the trainer is in battle mode
	if requester.trainer.GetTrainer().Mode != pkmn.BattlingTrainerMode {
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     catchWhenInWrongModeTemplate,
			TemplInfo: nil})
		if err != nil {
			return handlerError{user: "could not populate catch when in wrong mode template", err: err}
		}
	}

	// Assert that all necessary data is in the battle data object
	if !battleData.isComplete() {
		return handlerError{user: "could not load battle data", err: errors.New("incomplete battle data object")}
	}

	if battleData.opponent.trainer.GetTrainer().Type != pkmn.WildTrainerType {
		// You can only catch wild Pokemon! Let the user know that fact.
		err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     cannotCatchTrainerPokemonTemplate,
			TemplInfo: nil,
			Type:      messaging.Important})
		if err != nil {
			return handlerError{user: "could not populate cannot catch trainer Pokemon template", err: err}
		}
		return nil // There's nothing else to do
	}

	// Set up the next action as a catch action
	battleData.requester.battleInfo.GetTrainerBattleInfo().FinishedTurn = true
	battleData.requester.battleInfo.GetTrainerBattleInfo().NextBattleAction = pkmn.BattleAction{
		Type: pkmn.CatchBattleActionType,
		Val:  0} // The val parameter has no meaning for catches

	// Get a turn processor ready to do any required processing
	tp := turnProcessor{Services: s}

	var battleOver bool
	// Do any work required to get the opponent ready for the turn to be
	// processed
	ready, err := preprocessTurn(ctx, s, battleData)
	if err != nil {
		return handlerError{user: "could not do preprocessing on the current turn", err: err}
	}
	if ready {
		// The opponent is ready and the turn may be processed
		battleOver, err = tp.process(ctx, battleData)
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
