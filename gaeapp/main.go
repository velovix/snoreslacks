package gaeapp

import (
	"net/http"

	"github.com/velovix/snoreslacks/ctxman"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/handlers"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pokeapi"
	"github.com/velovix/snoreslacks/tasking"

	// Get the GAE implementations
	_ "github.com/velovix/snoreslacks/ctxman/gae"
	_ "github.com/velovix/snoreslacks/database/gae"
	_ "github.com/velovix/snoreslacks/logging/gae"
	_ "github.com/velovix/snoreslacks/messaging/gae"
	_ "github.com/velovix/snoreslacks/pokeapi/gae"
	_ "github.com/velovix/snoreslacks/tasking/gae"
)

func init() {
	loadConfig()

	// Get all Google App Engine dependencies
	ctxCreator, err := ctxman.Get("gae")
	if err != nil {
		panic(err)
	}
	db, err := database.Get("gae")
	if err != nil {
		panic(err)
	}
	log, err := logging.Get("gae")
	if err != nil {
		panic(err)
	}
	clientCreator, err := messaging.GetClientCreator("gae")
	if err != nil {
		panic(err)
	}
	fetcher, err := pokeapi.Get("gae")
	if err != nil {
		panic(err)
	}
	queue, err := tasking.Get("gae")

	// Inject dependencies into the services collection to be used by handlers.
	services := handlers.Services{
		CtxCreator:    ctxCreator,
		DB:            db,
		Log:           log,
		ClientCreator: clientCreator,
		Fetcher:       fetcher,
		WorkQueue:     queue}

	http.Handle(handlers.WaitingHelpURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.WaitingHelp{}})

	http.Handle(handlers.BattleWaitingHelpURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.BattleWaitingHelp{}})

	http.Handle(handlers.BattlingHelpURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.BattlingHelp{}})

	http.Handle(handlers.ChallengeURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.Challenge{}})

	http.Handle(handlers.ForfeitURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.Forfeit{}})

	http.Handle(handlers.NewTrainerURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.NewTrainer{}})

	http.Handle(handlers.ChoosingStarterURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.ChoosingStarter{}})

	http.Handle(handlers.UseMoveURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.UseMove{}})

	http.Handle(handlers.SwitchPokemonURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.SwitchPokemon{}})

	http.Handle(handlers.CatchPokemonURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.CatchPokemon{}})

	http.Handle(handlers.ViewPartyURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.ViewParty{}})

	http.Handle(handlers.WildEncounterURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.WildEncounter{}})

	http.Handle(handlers.ForgetMoveHelpURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.ForgetMoveHelp{}})

	http.Handle(handlers.ForgetMoveURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.ForgetMove{}})

	http.Handle(handlers.NoForgetMoveURL, handlers.Runner{
		Servs: services,
		Task:  &handlers.NoForgetMove{}})

	// Set up the main handler to respond to Slack requests
	mainHandler := &handlers.Main{
		Services: services,
		Token:    config.Token}
	http.Handle(handlers.MainURL, mainHandler)
}
