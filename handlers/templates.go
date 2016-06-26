package handlers

import "text/template"

var invalidCommandTemplateText = `
Invalid command format.
`
var invalidCommandTemplate *template.Template

var initialResponseTemplateText = `
Hold tight trainer, your request is being processed...
`
var initialResponseTemplate *template.Template

// Starter message template. This template will be shown when the user is
// first becoming a trainer.
var starterMessageTemplateText = `
*Professor Oak*
Hello there! Welcome to the world of Pokémon! My name is Oak! People call me the Pokémon Professor!
This world is inhabited by creatures called Pokémon! For some people, Pokémon are pets. Others use them for fights. Myself... I study Pokémon as a profession.
{{.Username}}! Your very own Pokémon legend is about to unfold! A world of dreams and adventures with Pokémon awaits! Let's go!
Let's get you a starter Pokémon!

To pick your starter, respond with {{ .SlashCommand }}, followed by the name of the starter you want!

{{ range .Starters }}
	*{{ .Name }}* (No. {{ .ID }})
{{ printf "\u0060\u0060\u0060" }}Type 1: {{ .Type1 }}
Type 2: {{ .Type2 }}
Height: {{ .Height }}
Weight: {{ .Weight }}
{{ printf "\u0060\u0060\u0060" }}

{{end}}
`
var starterMessageTemplate *template.Template

// Starter picked template. This template will be shown when the trainer picks
// a valid starter.
var starterPickedTemplateText = `
*Professor Oak*
So, you want {{ .PkmnName }}! This Pokémon is really energetic!
{{ .TrainerName }} received a {{ .PkmnName }}!

(You can check out your party with the "{{ .CommandName }} party" command)
`
var starterPickedTemplate *template.Template

// Starter instructions template. This template will be shown when a trainer
// should be choosing a starter and it doesn't seem like they know what they're
// doing.
var starterInstructionsTemplateText = `
You need to choose your starter before you can start playing!
`
var starterInstructionsTemplate *template.Template

// Invalid starter template. This template will be shown when a trainer
// attempts to pick a starter that doesn't exist.
var invalidStarterTemplateText = `
*Professor Oak*
{{ . }}?? I don't have any Pokémon like that!
`
var invalidStarterTemplate *template.Template

// View party in battle template. This template will be shown when a trainer
// wants to view their party.
var viewPartyInBattleTemplateText = `
{{ range . }}
	*{{ .Name }}* (No. {{ .ID }})
{{ printf "\u0060\u0060\u0060" -}}
IN-BATTLE STATS
  HP     : {{ .CurrHP }} / {{ .HP }}
  Status : {{ .StatusCondition }}

BASE STATS
  Level : {{ printf "%3d" .Level     }}    Types : {{ .Type1 }} {{ .Type2 }}
  HP    : {{ printf "%3d" .HP        }}    Att   : {{ printf "%3d" .Attack }}
  Def   : {{ printf "%3d" .Defense   }}    SpAtt : {{ printf "%3d" .SpAttack }}
  SpDef : {{ printf "%3d" .SpDefense }}    Speed : {{ printf "%3d" .Speed }}
{{ printf "\u0060\u0060\u0060" }}
{{ end }}
`
var viewPartyInBattleTemplate *template.Template

// View party template. This template will be shown when a trainer wants
// to view their party.
var viewPartyTemplateText = `
{{ range . }}
	*{{ .Name }}* (No. {{ .ID }})
{{ printf "\u0060\u0060\u0060" -}}
BASE STATS
  Level : {{ printf "%3d" .Level     }}    Types : {{ .Type1 }} {{ .Type2 }}
  HP    : {{ printf "%3d" .HP        }}    Att   : {{ printf "%3d" .Attack }}
  Def   : {{ printf "%3d" .Defense   }}    SpAtt : {{ printf "%3d" .SpAttack }}
  SpDef : {{ printf "%3d" .SpDefense }}    Speed : {{ printf "%3d" .Speed }}
{{ printf "\u0060\u0060\u0060" }}
{{ end }}
`
var viewPartyTemplate *template.Template

type viewSinglePokemonTemplateInfo struct {
	Name            string
	ID              int
	Level           int
	Type1           string
	Type2           string
	HP              int
	CurrHP          int
	Attack          int
	Defense         int
	SpAttack        int
	SpDefense       int
	Speed           int
	StatusCondition string
}

// Waiting help template. This template will be shown when the player is
// looking for a list of commands and isn't in a special state.
var waitingHelpTemplateText = `
You are currently doing nothing in particular.

{{ . }} *party*
View the list of Pokémon you have in your party, including their stats and other useful information.

{{ . }} *battle* _username_
Request a battle with a trainer that has the given username. The user has to be a trainer. The user can accept by using this command with your username.

{{ . }} *wild*
Jump into a wild Pokémon encounter! More wild Pokemon become available to you as you progress through the game.
`
var waitingHelpTemplate *template.Template

// Battle waiting help template. Shows when the user is looking for a list of
// commands and is waiting to start a battle.
var battleWaitingHelpTemplateText = `
You are currently waiting for your opponent to accept your challenge.

{{ . }} *party*
View the list of Pokémon you have in your party, including their stats and other useful information.

{{ . }} *forfeit*
Stops waiting for your opponent to accept your challenge. It does not count as a loss so long as you're still waiting.
`
var battleWaitingHelpTemplate *template.Template

var battlingHelpTemplateText = `
You are currently in a battle.

{{ . }} *party*
View the list of Pokémon you have in your party, including their stats and other useful information.

{{ . }} *use* _id_
Uses the move with the given ID

{{ . }} *switch* _slot_
Switch to the Pokémon with the given ID.

{{ . }} *catch*
Throws a Pokéball at the Pokemon. Just don't do this to Pokémon that already have an owner!

{{ . }} *forfeit*
Leave the battle. This counts as a loss for you.
`
var battlingHelpTemplate *template.Template

// No such trainer exists template. This template will be shown when the
// trainer wants to interact with another trainer that isn't registred.
var noSuchTrainerExistsTemplateText = `
There's no trainer with the name '{{ . }}'. Is the Slack user registered as a trainer?
`
var noSuchTrainerExistsTemplate *template.Template

// Waiting for battle template. This template will be shown when a trainer
// is waiting to battle another trainer.
var waitingForBattleTemplateText = `
Trainer {{ .Challenger }} wants to battle {{ .Opponent }}!

{{ .Opponent }}, if you accept this challenge, type "{{ .SlashCommand }} battle {{ .Challenger }}" to start the battle!
`
var waitingForBattleTemplate *template.Template

// Battle started template. This template will be shown when a battle between
// trainers begins.
var battleStartedTemplateText = `
A battle has started between {{ .Challenger }} and {{ .Opponent }}!
`
var battleStartedTemplate *template.Template

var waitingForfeitTemplateText = `
{{ .Forfeitter }} has stopped challenging {{ .Opponent }}.
`
var waitingForfeitTemplate *template.Template

var battlingForfeitTemplateText = `
{{ .Forfeitter }} has forfeitted the match, making {{ .Opponent }} the winner by default!
`
var battlingForfeitTemplate *template.Template

// Not trainer's turn template. Shown when the trainer attempts to make a move
// when it isn't their turn any more.
var notTrainersTurnTemplateText = `
You have already made your move this turn.
`
var notTrainersTurnTemplate *template.Template

// Move confirmation template. Shows when the server successfully processes a
// request to use a move.
var moveConfirmationTemplateText = `
You will be using {{ . }} next turn.
`
var moveConfirmationTemplate *template.Template

// Switch confirmation template. Shows when the server successfully processes a
// request to switch Pokemon.
var switchConfirmationTemplateText = `
You will be switching to {{ . }} next turn.
`
var switchConfirmationTemplate *template.Template

// Action options template. Shows the battle options a trainer has.
var actionOptionsTemplateText = `
To select an action, use the "move" or "switch" command along with the ID of your choice.
*Current Pokémon*: {{ .CurrPokemonName }}
{{ printf "\u0060\u0060\u0060" -}}
MOVES
{{ range $id, $moveName := .MoveSlots }}  {{ toBaseOne $id }}: {{ $moveName }}
{{ end -}}
PARTY
{{ range $id, $pkmnName := .PartySlots }}  {{ toBaseOne $id }}: {{ $pkmnName }}
{{ end -}}
{{ printf "\u0060\u0060\u0060" }}
`
var actionOptionsTemplate *template.Template

// Move report template. Contains a full textual representation of a move
// report, telling trainers what happened when a move was used.
var moveReportTemplateText = `
{{ .UserActionPrefix }} {{ .UserPokemonName }} used {{ .MoveName }}!
{{ if .Missed -}}
But the attack missed!
{{ else if .CriticalHit -}}
Critical hit!
{{ else -}}
{{- end -}}
{{ if gt .Effectiveness 1 -}}
It's super effective!
{{ else if lt .Effectiveness 1 -}}
It's not very effective...
{{ else if eq .Effectiveness 0 -}}
The move had no effect...
{{ else -}}
{{- end -}}
{{ if .TargetDamage -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} took {{ .TargetDamage }} damage!
{{ else -}}
{{- end -}}
{{ if .TargetDrain -}}
{{ .UserActionPrefix }} {{ .UserPokemonName }} drained {{ .TargetDrain }} HP from {{ .TargetActionPrefix }} {{ .TargetPokemonName }}!
{{ else -}}
{{- end -}}
{{ if gt .UserHealing 0 -}}
{{ .UserActionPrefix }} {{ .UserPokemonName }} healed {{ .UserHealing }} HP!
{{ else if lt .UserHealing 0 -}}
{{ .UserActionPrefix }} {{ .UserPokemonName }} suffered knockback damage...
{{ else -}}
{{- end -}}
{{ if .UserFainted -}}
{{ .UserActionPrefix }} {{ .UserPokemonName }} has fainted!
{{ else -}}
{{- end -}}
{{ if .TargetFainted -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has fainted!
{{ else -}}
{{- end -}}
{{ if gt .AttStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s attack has increased!
{{ else if lt .AttStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s attack has decreased!
{{ else -}}
{{- end -}}
{{ if gt .DefStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s defense has increased!
{{ else if lt .DefStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s defense has decreased!
{{ else -}}
{{- end -}}
{{ if gt .SpAttStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s special attack has increased!
{{ else if lt .SpAttStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s special attack has decreased!
{{ else -}}
{{- end -}}
{{ if gt .SpDefStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s special defense has increased!
{{ else if lt .SpDefStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s special defense has decreased!
{{ else -}}
{{- end -}}
{{ if gt .SpeedStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s speed has increased!
{{ else if lt .SpeedStageChange 0 -}}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }}'s speed has decreased!
{{ else -}}
{{- end -}}
{{ if .Poisoned }}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has been poisoned!
{{ else -}}
{{- end -}}
{{ if .Paralyzed }}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has been paralyzed!
{{ else -}}
{{- end -}}
{{ if .Asleep }}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has fallen asleep!
{{ else -}}
{{- end -}}
{{ if .Frozen }}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has been frozen!
{{ else -}}
{{- end -}}
{{ if .Burned }}
{{ .TargetActionPrefix }} {{ .TargetPokemonName }} has been burned!
{{ else -}}
{{- end -}}
{{ printf "\u0060" }}{{ printf "%-15s" .TargetPokemonName }}: {{ .UserHPBar }}{{ printf "\u0060" }}
`
var moveReportTemplate *template.Template

var switchPokemonTemplateText = `
{{ .Switcher }} has withdrawn {{ .WithdrawnPokemon }}.
{{ .Switcher }} sent out {{ .SelectedPokemon }}! (Lv. {{ .SelectedLevel }})
`
var switchPokemonTemplate *template.Template

var initialPokemonSendOutTemplateText = `
{{ .TrainerName }} sent out {{ .PokemonName }}! (Lv. {{ .Level }})
`
var initialPokemonSendOutTemplate *template.Template

var faintedPokemonUsingMoveTemplateText = `
A fainted Pokémon cannot use a move. You must switch to a battle-ready Pokémon first.
`
var faintedPokemonUsingMoveTemplate *template.Template

var trainerLostTemplateText = `
{{ .LostTrainer }} is out of usable Pokémon. {{ .WonTrainer }} has won the battle!
`
var trainerLostTemplate *template.Template

var wildBattleStartedTemplateText = `
A wild {{ .WildPokemonName }} appeared! (Lv. {{ .Level }})
`
var wildBattleStartedTemplate *template.Template

var cannotCatchTrainerPokemonTemplateText = `
The trainer blocked the ball! Don't be a thief!
`
var cannotCatchTrainerPokemonTemplate *template.Template

var pokemonCaughtTemplateText = `
You threw a Pokeball at the wild {{ . }}...
The wild {{ . }} was caught!
`
var pokemonCaughtTemplate *template.Template

var pokemonNotCaughtTemplateText = `
You threw a Pokeball at the wild {{ . }}...
But it broke out!
`
var pokemonNotCaughtTemplate *template.Template

var invalidMoveSlotTemplateText = `
Your Pokémon does not have a move in slot {{ . }}!
`
var invalidMoveSlotTemplate *template.Template

var challengingWhenInWrongModeTemplateText = `
You cannot challenge a trainer right now!
`
var challengingWhenInWrongModeTemplate *template.Template

var forfeittingWhenInWrongModeTemplateText = `
You cannot forfeit a battle right now!
`
var forfeittingWhenInWrongModeTemplate *template.Template

var choosingStarterWhenInWrongModeTemplateText = `
You cannot choose a starter more than once!
`
var choosingStarterWhenInWrongModeTemplate *template.Template

var usingMoveWhenInWrongModeTemplateText = `
You cannot choose a move right now!
`
var usingMoveWhenInWrongModeTemplate *template.Template

var switchingWhenInWrongModeTemplateText = `
You cannot switch Pokémon right now!
`
var switchingWhenInWrongModeTemplate *template.Template

var wildEncounterWhenInWrongModeTemplateText = `
You cannot engage in a wild encounter right now!
`
var wildEncounterWhenInWrongModeTemplate *template.Template

var catchWhenInWrongModeTemplateText = `
You cannot catch a Pokémon right now!
`
var catchWhenInWrongModeTemplate *template.Template

var noChallengingSelfTemplateText = `
You cannot battle yourself!
`
var noChallengingSelfTemplate *template.Template

var invalidPartySlotTemplateText = `
You don't have a Pokémon in slot {{ . }}!
`
var invalidPartySlotTemplate *template.Template

var switchToFaintedPokemonTemplateText = `
There is no will to fight!
`
var switchToFaintedPokemonTemplate *template.Template

var switchToCurrentPokemonTemplateText = `
The Pokémon is already in battle!
`
var switchToCurrentPokemonTemplate *template.Template

// toBaseOne converts the given number from base-zero to base-one by adding one
// to it. This is intended to be used in templates.
func toBaseOne(i int) int {
	return i + 1
}

// Parse all templates.
func init() {
	// Make our standard library of template functions available
	funcMap := template.FuncMap{
		"toBaseOne": toBaseOne}

	invalidCommandTemplate = template.Must(template.New("").Funcs(funcMap).Parse(invalidCommandTemplateText))
	initialResponseTemplate = template.Must(template.New("").Funcs(funcMap).Parse(initialResponseTemplateText))
	starterMessageTemplate = template.Must(template.New("").Funcs(funcMap).Parse(starterMessageTemplateText))
	starterInstructionsTemplate = template.Must(template.New("").Funcs(funcMap).Parse(starterInstructionsTemplateText))
	invalidStarterTemplate = template.Must(template.New("").Funcs(funcMap).Parse(invalidStarterTemplateText))
	starterPickedTemplate = template.Must(template.New("").Funcs(funcMap).Parse(starterPickedTemplateText))
	viewPartyTemplate = template.Must(template.New("").Funcs(funcMap).Parse(viewPartyTemplateText))
	viewPartyInBattleTemplate = template.Must(template.New("").Funcs(funcMap).Parse(viewPartyInBattleTemplateText))
	waitingHelpTemplate = template.Must(template.New("").Funcs(funcMap).Parse(waitingHelpTemplateText))
	noSuchTrainerExistsTemplate = template.Must(template.New("").Funcs(funcMap).Parse(noSuchTrainerExistsTemplateText))
	battleStartedTemplate = template.Must(template.New("").Funcs(funcMap).Parse(battleStartedTemplateText))
	waitingForBattleTemplate = template.Must(template.New("").Funcs(funcMap).Parse(waitingForBattleTemplateText))
	battleWaitingHelpTemplate = template.Must(template.New("").Funcs(funcMap).Parse(battleWaitingHelpTemplateText))
	waitingForfeitTemplate = template.Must(template.New("").Funcs(funcMap).Parse(waitingForfeitTemplateText))
	battlingForfeitTemplate = template.Must(template.New("").Funcs(funcMap).Parse(battlingForfeitTemplateText))
	battlingHelpTemplate = template.Must(template.New("").Funcs(funcMap).Parse(battlingHelpTemplateText))
	moveConfirmationTemplate = template.Must(template.New("").Funcs(funcMap).Parse(moveConfirmationTemplateText))
	switchConfirmationTemplate = template.Must(template.New("").Funcs(funcMap).Parse(switchConfirmationTemplateText))
	actionOptionsTemplate = template.Must(template.New("").Funcs(funcMap).Parse(actionOptionsTemplateText))
	moveReportTemplate = template.Must(template.New("").Funcs(funcMap).Parse(moveReportTemplateText))
	switchPokemonTemplate = template.Must(template.New("").Funcs(funcMap).Parse(switchPokemonTemplateText))
	initialPokemonSendOutTemplate = template.Must(template.New("").Funcs(funcMap).Parse(initialPokemonSendOutTemplateText))
	faintedPokemonUsingMoveTemplate = template.Must(template.New("").Funcs(funcMap).Parse(faintedPokemonUsingMoveTemplateText))
	trainerLostTemplate = template.Must(template.New("").Funcs(funcMap).Parse(trainerLostTemplateText))
	wildBattleStartedTemplate = template.Must(template.New("").Funcs(funcMap).Parse(wildBattleStartedTemplateText))
	cannotCatchTrainerPokemonTemplate = template.Must(template.New("").Funcs(funcMap).Parse(cannotCatchTrainerPokemonTemplateText))
	pokemonCaughtTemplate = template.Must(template.New("").Funcs(funcMap).Parse(pokemonCaughtTemplateText))
	pokemonNotCaughtTemplate = template.Must(template.New("").Funcs(funcMap).Parse(pokemonNotCaughtTemplateText))
	invalidMoveSlotTemplate = template.Must(template.New("").Funcs(funcMap).Parse(invalidMoveSlotTemplateText))
	challengingWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(challengingWhenInWrongModeTemplateText))
	forfeittingWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(forfeittingWhenInWrongModeTemplateText))
	choosingStarterWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(choosingStarterWhenInWrongModeTemplateText))
	usingMoveWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(usingMoveWhenInWrongModeTemplateText))
	wildEncounterWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(wildEncounterWhenInWrongModeTemplateText))
	catchWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(catchWhenInWrongModeTemplateText))
	noChallengingSelfTemplate = template.Must(template.New("").Funcs(funcMap).Parse(noChallengingSelfTemplateText))
	switchingWhenInWrongModeTemplate = template.Must(template.New("").Funcs(funcMap).Parse(switchingWhenInWrongModeTemplateText))
	invalidPartySlotTemplate = template.Must(template.New("").Funcs(funcMap).Parse(invalidPartySlotTemplateText))
	switchToFaintedPokemonTemplate = template.Must(template.New("").Funcs(funcMap).Parse(switchToFaintedPokemonTemplateText))
	switchToCurrentPokemonTemplate = template.Must(template.New("").Funcs(funcMap).Parse(switchToCurrentPokemonTemplateText))
}
