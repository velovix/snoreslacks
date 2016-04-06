package handlers

import "html/template"

var initialResponseTemplateText = `
Hey trainer, your request is being processed...
`
var initialResponseTemplate *template.Template

// Starter message template. This template will be shown when the user is
// first becoming a trainer.
var starterMessageTemplateText = `
*Professor Oak*
Hello there! Welcome to the world of Pokémon! My name is Oak! People call me the Pokémon Professor!
This world is inhabited by creatures called Pokémon! For some people, Pokémon are pets. Others use them for fights. Myself... I study Pokémon as a profession.
{{.Username}}! Your very own Pokémon legend is about to unfold! A world of dreams and adventures with Pokémon awaits! Let's go!
Let's get you a starter Pokémon! Respond with the name of the Pokémon you would like to have to get you started.

{{ range .Starters }}
	*{{ .Name }}* (No. {{ .ID }})
{{ printf "\u0060\u0060\u0060" }} {{/* Three grav accents to denote a code block */}}
Type 1: {{ .Type1 }}
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

(You can check out your party with the "party" keyword)
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

// View party template. This template will be shown when a trainer wants to
// view their party.
var viewPartyTemplateText = `
*Your Party*

{{ range . }}
	*{{ .Name }}* (No. {{ .ID }})
{{ printf "\u0060\u0060\u0060" }} {{/* Three grav accents to denote a code block */}}
Level : {{ printf "%03d" .Level     }}    Types : {{ .Type1 }} {{ .Type2 }}
HP    : {{ printf "%03d" .HP        }}    Att   : {{ printf "%03d" .Attack }}
Def   : {{ printf "%03d" .Defense   }}    SpAtt : {{ printf "%03d" .SpAttack }}
SpDef : {{ printf "%03d" .SpDefense }}    Speed : {{ printf "%03d" .Speed }}
{{ printf "\u0060\u0060\u0060" }}
{{ end }}
`
var viewPartyTemplate *template.Template

type viewSinglePokemonTemplateInfo struct {
	Name      string
	ID        int
	Level     int
	Type1     string
	Type2     string
	HP        int
	Attack    int
	Defense   int
	SpAttack  int
	SpDefense int
	Speed     int
}

// Waiting help template. This template will be shown when the player is
// looking for a list of commands and isn't in a special state.
var waitingHelpTemplateText = `
You are currently doing nothing in particular.

*party*
View the list of Pokémon you have in your party, including their stats and other useful information.

*battle* _username_
Request a battle with a trainer that has the given username. The user has to be a trainer. The user can accept by using this command with your username.
`
var waitingHelpTemplate *template.Template

// Battle waiting help template. Shows when the user is looking for a list of
// commands and is waiting to start a battle.
var battleWaitingHelpTemplateText = `
You are currently waiting for your opponent to accept your challenge.

*party*
View the list of Pokémon you have in your party, including their stats and other useful information.

*forfeit*
Stops waiting for your opponent to accept your challenge. It does not count as a loss so long as you're still waiting.
`
var battleWaitingHelpTemplate *template.Template

var battlingHelpTemplateText = `
You are currently in a battle.

*party*
View the list of Pokémon you have in your party, including their stats and other useful information.

*use* _id_
Uses the move with the given ID. These IDs are scrambled every turn so that the opponent doesn't know what move you've chosen.

*switch* _slot_
Switch to the Pokémon in the given slot, starting at 1.

*forfeit*
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
Trainer {{ .P1.Name }} wants to battle {{ .P2.Name }}!
`
var waitingForBattleTemplate *template.Template

// Battle started template. This template will be shown when a battle between
// trainers begins.
var battleStartedTemplateText = `
A battle has started between {{ .P1.Name }} and {{ .P2.Name }}!
`
var battleStartedTemplate *template.Template

var waitingForfeitTemplateText = `
{{ .P1.Name }} has stopped challenging {{ .P2.Name }}.
`
var waitingForfeitTemplate *template.Template

var battlingForfeitTemplateText = `
{{ .P1.Name }} has forfeitted the match, making {{ .P2.Name }} the winner by default!
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

var actionOptionsTemplateText = `
To select an action, use the "move" or "switch" command along with the ID of your choice. The IDs are scrambled so your opponent can't know what move you'll use.
*Current Pokémon*: {{ .CurrPokemonName }}
	*Moves*
{{ printf "\u0060\u0060\u0060" }}
{{ range .MoveTable.Moves }}
{{ .ID }}: {{ printf "%12d" .MoveName }}
{{ end }}
{{ printf "\u0060\u0060\u0060" }}
	*Party*
{{ printf "\u0060\u0060\u0060" }}
{{ range .PartyTable.Members }}
{{ .ID }}: {{ printf "%12d" .PkmnName }}
{{ end }}
{{ printf "\u0060\u0060\u0060" }}
`
var actionOptionsTemplate *template.Template

// Parse all templates.
func init() {
	initialResponseTemplate = template.Must(template.New("").Parse(initialResponseTemplateText))
	starterMessageTemplate = template.Must(template.New("").Parse(starterMessageTemplateText))
	starterInstructionsTemplate = template.Must(template.New("").Parse(starterInstructionsTemplateText))
	invalidStarterTemplate = template.Must(template.New("").Parse(invalidStarterTemplateText))
	starterPickedTemplate = template.Must(template.New("").Parse(starterPickedTemplateText))
	viewPartyTemplate = template.Must(template.New("").Parse(viewPartyTemplateText))
	waitingHelpTemplate = template.Must(template.New("").Parse(waitingHelpTemplateText))
	noSuchTrainerExistsTemplate = template.Must(template.New("").Parse(noSuchTrainerExistsTemplateText))
	battleStartedTemplate = template.Must(template.New("").Parse(battleStartedTemplateText))
	waitingForBattleTemplate = template.Must(template.New("").Parse(waitingForBattleTemplateText))
	battleWaitingHelpTemplate = template.Must(template.New("").Parse(battleWaitingHelpTemplateText))
	waitingForfeitTemplate = template.Must(template.New("").Parse(waitingForfeitTemplateText))
	battlingForfeitTemplate = template.Must(template.New("").Parse(battlingForfeitTemplateText))
	battlingHelpTemplate = template.Must(template.New("").Parse(battlingHelpTemplateText))
	moveConfirmationTemplate = template.Must(template.New("").Parse(moveConfirmationTemplateText))
	switchConfirmationTemplate = template.Must(template.New("").Parse(moveConfirmationTemplateText))
	actionOptionsTemplate = template.Must(template.New("").Parse(actionOptionsTemplateText))
}
