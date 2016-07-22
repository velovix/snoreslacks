package handlers

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
	"golang.org/x/net/context"
)

// NoForgetMove gives up on learning a new move and keeps the old ones.
type NoForgetMove struct {
}

func (h *NoForgetMove) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Get the Pokemon in question
	var pk database.Pokemon
	for _, partyMember := range requester.pkmn {
		// Find the first Pokemon that is waiting to level up, which is always
		// the one we're currently talking about
		if partyMember.GetPokemon().ReadyToLevelUp() {
			pk = partyMember
			break
		}
	}

	// Load information on the move that won't be learned
	learnableMoves, err := pokeapi.FetchLevelLearnableMoveIDs(ctx, client, s.Fetcher, pk.GetPokemon().ID)
	if err != nil {
		return handlerError{user: "failed to fetch learnable moves", err: err}
	}
	move, err := loadMove(ctx, client, s.Fetcher, learnableMoves[pk.GetPokemon().Level+1][0])
	if err != nil {
		return handlerError{user: "failed to load move info", err: err}
	}

	// Let the user know the Pokemon gave up on learning the move
	templInfo := struct {
		PokemonName string
		MoveName    string
	}{
		PokemonName: pk.GetPokemon().Name,
		MoveName:    move.Name}
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     giveUpLearningMoveTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return handlerError{user: "could not populate give up learning move template", err: err}
	}

	// Level up the Pokemon manually to show that all conflicts have been
	// resolved
	pk.GetPokemon().Level += 1
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     levelUpTemplate,
		TemplInfo: pk.GetPokemon()})
	if err != nil {
		return handlerError{user: "could not populate level up template", err: err}
	}

	// Continue the process of leveling up the party
	problemSlot, err := levelUpPartyIfPossible(ctx, s, requester)
	if problemSlot == -1 {
		// The process is complete
		requester.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
	}

	// Save data if all went well
	err = saveBasicTrainerData(ctx, s.DB, requester)
	if err != nil {
		return handlerError{user: "could not save basic trainer data", err: err}
	}

	return nil
}

// ForgetMove allows the user to pick an existing move to be deleted in favor
// of a new one.
type ForgetMove struct {
}

func (h *ForgetMove) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)

	// Check if the command looks correct
	if len(slackReq.CommandParams) != 1 {
		return sendInvalidCommand(client, requester.lastContactURL)
	}

	// Get the slot number of the move to be replaced
	replacedMoveSlotID, err := strconv.Atoi(slackReq.CommandParams[0])
	if err != nil {
		return sendInvalidCommand(client, requester.lastContactURL)
	}

	// Get the Pokemon in question
	var pk database.Pokemon
	for _, partyMember := range requester.pkmn {
		// Find the first Pokemon that is waiting to level up, which is always
		// the one we're currently talking about
		if partyMember.GetPokemon().ReadyToLevelUp() {
			pk = partyMember
			break
		}
	}

	// Check if the given slot ID is valid
	if replacedMoveSlotID < 1 || replacedMoveSlotID > pk.GetPokemon().MoveCount() {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     invalidMoveSlotTemplate,
			TemplInfo: replacedMoveSlotID})
		if err != nil {
			return handlerError{user: "could not populate invalid move slot tempalate", err: err}
		}
		return nil // There is nothing else to do
	}

	// Load information on the move that will be replacing the selected move
	learnableMoves, err := pokeapi.FetchLevelLearnableMoveIDs(ctx, client, s.Fetcher, pk.GetPokemon().ID)
	if err != nil {
		return handlerError{user: "failed to fetch learnable moves", err: err}
	}
	newMove, err := loadMove(ctx, client, s.Fetcher, learnableMoves[pk.GetPokemon().Level+1][0])
	if err != nil {
		return handlerError{user: "failed to load move info", err: err}
	}
	// Load information on the move that will be replaced
	oldMove, err := loadMove(ctx, client, s.Fetcher, pk.GetPokemon().MoveIDsAsSlice()[replacedMoveSlotID-1])
	if err != nil {
		return handlerError{user: "failed to load new move info", err: err}
	}

	// Replace the move
	err = pk.GetPokemon().ReplaceMove(replacedMoveSlotID, newMove.ID)
	if err != nil {
		return handlerError{user: "unable to replace move", err: errors.Wrap(err, "while replacing move")}
	}

	// Send confirmation of the replace
	templInfo := struct {
		PokemonName string
		OldMoveName string
		NewMoveName string
	}{
		PokemonName: pk.GetPokemon().Name,
		OldMoveName: oldMove.Name,
		NewMoveName: newMove.Name}
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     replacedMoveTemplate,
		TemplInfo: templInfo})
	if err != nil {
		return handlerError{user: "failed to populate replaced move template", err: err}
	}

	// Level up the Pokemon manually to show that all conflicts have been
	// resolved
	pk.GetPokemon().Level += 1
	err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     levelUpTemplate,
		TemplInfo: pk.GetPokemon()})
	if err != nil {
		return handlerError{user: "could not populate level up template", err: err}
	}

	// Continue the process of leveling up the party
	problemSlot, err := levelUpPartyIfPossible(ctx, s, requester)
	if problemSlot == -1 {
		// The process is complete
		requester.trainer.GetTrainer().Mode = pkmn.WaitingTrainerMode
	}

	// Save data if all went well
	err = saveBasicTrainerData(ctx, s.DB, requester)
	if err != nil {
		return handlerError{user: "could not save basic trainer data", err: err}
	}

	return nil
}
