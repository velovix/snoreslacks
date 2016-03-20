package app

import "html/template"

// Starter message template. This template will be shown when the user is
// first becoming a trainer.
var starterMessageTemplateText = `
*Professor Oak*
Hello there! Welcome to the world of Pokèmon! My name is Oak! People call me the Pokèmon Professor!

This world is inhabited by creatures called Pokèmon! For some people, Pokèmon are pets. Others use them for fights. Myself... I study Pokèmon as a profession.

{{.Username}}! Your very own Pokèmon legend is about to unfold! A world of dreams and adventures with Pokèmon awaits! Let's go!

Let's get you a starter Pokèmon! Respond with the name of the Pokèmon you would like to have to get you started.

{{ range .Starters }}
	*{{ .Name }}* (No. {{ .ID }})
	Type 1: {{ .Type1 }}
	Type 2: {{ .Type2 }}
	Height: {{ .Height }}
	Weight: {{ .Weight }}


{{end}}
`
var starterMessageTemplate *template.Template

// Starter picked template. This template will be shown when the trainer picks
// a valid starter.
var starterPickedTemplateText = `
*Professor Oak*
So, you want {{ .PkmnName }}! This Pokèmon is really energetic!

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
{{ . }}?? I don't have any Pokèmon like that!
`
var invalidStarterTemplate *template.Template

// View party template. This template will be shown when a trainer wants to
// view their party.
var viewPartyTemplateText = `
*Your Party*
{{ range . }}
	*{{ .Name }}* (No. {{ .ID }})
	Level: {{ .Level }}
	Types: {{ .Type1 }} {{ .Type2 }}
	HP: {{ .HP }}
	Attack: {{ .Attack }}
	Defense: {{ .Defense }}
	SpAttack: {{ .SpAttack }}
	SpDefense: {{ .SpDefense }}
	Speed: {{ .Speed }}
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
View the list of Pokemon you have in your party, including their stats and
other useful information.

*battle* _username_
Request a battle with a trainer that has the given username. The user has to be
a trainer. The user can accept by using this command with your username.
`
var waitingHelpTemplate *template.Template

// Parse all templates.
func init() {
	starterMessageTemplate = template.Must(template.New("").Parse(starterMessageTemplateText))
	starterInstructionsTemplate = template.Must(template.New("").Parse(starterInstructionsTemplateText))
	invalidStarterTemplate = template.Must(template.New("").Parse(invalidStarterTemplateText))
	starterPickedTemplate = template.Must(template.New("").Parse(starterPickedTemplateText))
	viewPartyTemplate = template.Must(template.New("").Parse(viewPartyTemplateText))
	waitingHelpTemplate = template.Must(template.New("").Parse(waitingHelpTemplateText))
}
