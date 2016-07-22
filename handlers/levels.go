package handlers

import (
	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
	"github.com/velovix/snoreslacks/pokeapi"
	"golang.org/x/net/context"
)

// levelUpPartyIfPossible checks if any of the given trainer's Pokemon are
// ready to level up. The Pokemon may be leveled up if all new moves can be
// learned automatically, or the trainer will be prompted to forget a move if a
// Pokemon can learn a move but there are no move slots left. This function
// returns the party slot of the Pokemon that created the conflict, or -1 if no
// conflict occurred.
func levelUpPartyIfPossible(ctx context.Context, s Services, t *basicTrainerData) (int, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)

	for slot, p := range t.pkmn {
		for p.GetPokemon().ReadyToLevelUp() {
			unlearnedMoveID, err := learnNewMovesIfPossible(ctx, s, t, p)
			if err != nil {
				return -1, err
			}
			if unlearnedMoveID == -1 {
				// No moves were left unlearned.  We don't have to prompt the
				// trainer to forget any moves, so the Pokemon can be leveled
				// up immediately
				p.GetPokemon().Level += 1
				// Let the trainer know
				err := messaging.SendTempl(client, t.lastContactURL, messaging.TemplMessage{
					Templ:     levelUpTemplate,
					TemplInfo: p.GetPokemon()})
				if err != nil {
					return -1, err
				}
			} else {
				// The user must be prompted to either forget a move or give up
				// on learning this move.
				t.trainer.GetTrainer().Mode = pkmn.ForgetMoveTrainerMode
				// Fetch info on the unlearned move
				unlearnedMove, err := loadMove(ctx, client, s.Fetcher, unlearnedMoveID)
				if err != nil {
					return -1, err
				}
				// Prompt the trainer
				templInfo := struct {
					MoveName     string
					PokemonName  string
					SlashCommand string
					MoveSlots    []string
				}{
					MoveName:     unlearnedMove.Name,
					PokemonName:  p.GetPokemon().Name,
					SlashCommand: slackReq.SlashCommand}
				// Populate the existing move slots of the Pokemon
				for _, moveID := range p.GetPokemon().MoveIDsAsSlice() {
					move, err := loadMove(ctx, client, s.Fetcher, moveID)
					if err != nil {
						return -1, err
					}
					templInfo.MoveSlots = append(templInfo.MoveSlots, move.Name)
				}
				err = messaging.SendTempl(client, t.lastContactURL, messaging.TemplMessage{
					TemplInfo: templInfo,
					Templ:     forgetMoveTemplate})
				if err != nil {
					return -1, err
				}

				// A conflict occurred
				return slot, nil
			}
		}
	}

	// The level up process has completed without conflicts
	return -1, nil
}

// checkForLearnableMoves checks if the given Pokemon will learn any moves when
// it levels up. This function will teach the Pokemon all new moves if
// possible. This function returns the ID of the move that caused a conflict or
// -1 if there were no conflicts.
func learnNewMovesIfPossible(ctx context.Context, s Services, t *basicTrainerData, p database.Pokemon) (int, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Assert that the Pokemon is ready to level up
	if !p.GetPokemon().ReadyToLevelUp() {
		return -1, errors.New("attempt to check for learnable moves when the Pokemon isn't ready to level up yet")
	}

	// Fetch all moves this Pokemon can learn via level up
	learnableMoves, err := pokeapi.FetchLevelLearnableMoveIDs(ctx, client, s.Fetcher, p.GetPokemon().ID)
	if err != nil {
		return -1, errors.Wrap(err, "while checking for learnable moves")
	}

	if len(learnableMoves[p.GetPokemon().Level+1]) > 0 {
		// There is a new move the Pokemon will learn upon level up
		newMoveID := learnableMoves[p.GetPokemon().Level+1][0]
		if p.GetPokemon().MoveCount() >= 4 {
			// The Pokemon has no empty move slots, so either an existing move
			// must be forgotten or this move will not be learned. We have to
			// ask the trainer before the move can be learned and before the
			// Pokemon can officially level up.
			return newMoveID, nil
		} else {
			// We can learn the move automatically
			p.GetPokemon().LearnMove(newMoveID)
			// Fetch info on the new move
			newMove, err := loadMove(ctx, client, s.Fetcher, newMoveID)
			if err != nil {
				return -1, errors.Wrap(err, "while teaching a Pokemon a move")
			}
			// Let the user know
			templInfo := struct {
				PokemonName string
				MoveName    string
			}{
				PokemonName: p.GetPokemon().Name,
				MoveName:    newMove.Name}
			err = messaging.SendTempl(client, t.lastContactURL, messaging.TemplMessage{
				TemplInfo: templInfo,
				Templ:     learnedMoveTemplate})
			if err != nil {
				return -1, errors.Wrap(err, "while teaching a Pokemon a move")
			}
		}
	}

	return -1, nil
}
