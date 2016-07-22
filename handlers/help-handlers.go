package handlers

import (
	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/messaging"
)

// WaitingHelp sends help information to the user when they are not
// doing anything in particular.
type WaitingHelp struct {
}

func (h *WaitingHelp) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Send the templated info
	err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     waitingHelpTemplate,
		TemplInfo: slackReq.SlashCommand})
	if err != nil {
		return handlerError{user: "could not populate waiting help template", err: err}
	}

	return nil
}

// BattleWaitingHelp sends help information to the user when they are
// waiting for a battle to start.
type BattleWaitingHelp struct {
}

func (h *BattleWaitingHelp) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Send the templated info
	err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     battleWaitingHelpTemplate,
		TemplInfo: slackReq.SlashCommand})
	if err != nil {
		return handlerError{user: "could not populate battle waiting help template", err: err}
	}

	return nil
}

// BattlingHelp sends help information to the user when they are in a
// battle.
type BattlingHelp struct {
}

func (h *BattlingHelp) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Send the templated info
	err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     battlingHelpTemplate,
		TemplInfo: slackReq.SlashCommand})
	if err != nil {
		return handlerError{user: "could not populate battling help template", err: err}
	}

	return nil
}

// ForgetMoveHelp sends help information to the user when they are deciding
// whether or not to forget a move.
type ForgetMoveHelp struct {
}

func (h *ForgetMoveHelp) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	slackReq := ctx.Value("slack request").(messaging.SlackRequest)
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)

	// Send the templated info
	err := messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
		Templ:     forgetMoveHelpTemplate,
		TemplInfo: slackReq.SlashCommand})
	if err != nil {
		return handlerError{user: "could not populate battling help template", err: err}
	}

	return nil
}
