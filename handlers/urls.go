package handlers

const MainURL = "/"

// Workers
const (
	workerPrefix         = "/work"
	WaitingHelpURL       = workerPrefix + "/waiting-help"
	BattleWaitingHelpURL = workerPrefix + "/battle-waiting-help"
	BattlingHelpURL      = workerPrefix + "/battling-help"
	ChallengeURL         = workerPrefix + "/challenge"
	ForfeitURL           = workerPrefix + "/forfeit"
	NewTrainerURL        = workerPrefix + "/new-trainer"
	ChoosingStarterURL   = workerPrefix + "/choosing-starter"
	UseMoveURL           = workerPrefix + "/use-move"
	SwitchPokemonURL     = workerPrefix + "/switch-pokemon"
	CatchPokemonURL      = workerPrefix + "/catch-pokemon"
	ViewPartyURL         = workerPrefix + "/view-party"
	WildEncounterURL     = workerPrefix + "/wild"
	ForgetMoveHelpURL    = workerPrefix + "/forget-move-help"
	NoForgetMoveURL      = workerPrefix + "/no-forget-move"
	ForgetMoveURL        = workerPrefix + "/forget-move"
)
