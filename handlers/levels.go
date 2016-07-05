package handlers

import (
	"github.com/pkg/errors"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pokeapi"
	"golang.org/x/net/context"
)

// levelUpPartyIfPossible checks if any of the given trainer's Pokemon are
// ready to level up. The Pokemon may be leveled up if all new moves can be
// learned automatically, or the trainer will be prompted to forget a move if a
// Pokemon can learn a move but there are no move slots left. This function
// returns true if the level up process has completed without conflicts.
func levelUpPartyIfPossible(ctx context.Context, s Services, t *basicTrainerData) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	for _, p := range t.pkmn {
		for p.GetPokemon().ReadyToLevelUp() {
			allMovesLearned, err := learnNewMovesIfPossible(ctx, s, t, p)
			if err != nil {
				return false, err
			}
			if allMovesLearned {
				// We don't have to prompt the trainer to forget any moves, so
				// the Pokemon can be leveled up immediately
				p.GetPokemon().Level += 1
				// Let the trainer know
				err := messaging.SendTempl(client, t.lastContactURL, messaging.TemplMessage{
					Templ:     levelUpTemplate,
					TemplInfo: p.GetPokemon()})
				if err != nil {
					return false, err
				}
			} else {
				// The user must be prompted to either forget a move or give up
				// on learning this move.
			}
		}
	}

	// The level up process has completed without conflicts
	return true, nil
}

// checkForLearnableMoves checks if the given Pokemon will learn any moves when
// it levels up. This function will teach the Pokemon all new moves if
// possible. This function returns true if all new moves were learned without
// conflicts.
func learnNewMovesIfPossible(ctx context.Context, s Services, t *basicTrainerData, p database.Pokemon) (bool, error) {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)

	// Assert that the Pokemon is ready to level up
	if !p.GetPokemon().ReadyToLevelUp() {
		return false, errors.New("attempt to check for learnable moves when the Pokemon isn't ready to level up yet")
	}

	// Fetch all moves this Pokemon can learn via level up
	learnableMoves, err := pokeapi.FetchLevelLearnableMoveIDs(ctx, client, s.Fetcher, p.GetPokemon().ID)
	if err != nil {
		return false, errors.Wrap(err, "while checking for learnable moves")
	}

	if len(learnableMoves[p.GetPokemon().Level+1]) > 0 {
		// There is a new move the Pokemon will learn upon level up
		if p.GetPokemon().MoveCount() >= 4 {
			// The Pokemon has no empty move slots, so either an existing move
			// must be forgotten or this move will not be learned. We have to
			// ask the trainer before the move can be learned and before the
			// Pokemon can officially level up.
			return false, nil
		} else {
			newMoveID := learnableMoves[p.GetPokemon().Level+1][0]
			// We can learn the move automatically
			p.GetPokemon().LearnMove(newMoveID)
			// Fetch info on the new move
			newAPIMove, err := s.Fetcher.FetchMove(ctx, client, newMoveID)
			if err != nil {
				return false, errors.Wrap(err, "while teaching a Pokemon a move")
			}
			newMove, err := pokeapi.NewMove(newAPIMove)
			if err != nil {
				return false, errors.Wrap(err, "while teaching a Pokemon a move")
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
				return false, errors.Wrap(err, "while teaching a Pokemon a move")
			}
		}
	}

	return true, nil
}
