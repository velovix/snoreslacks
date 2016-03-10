package snoreslacks

var starterMessageTemplate = `
*Professor Oak*
Hello there! Welcome to the world of Pokèmon! My name is Oak! People call me the Pokèmon Professor!

This world is inhabited by creatures called Pokèmon! For some people, Pokèmon are pets. Others use them for fights. Myself... I study Pokèmon as a profession.

{{.Username}}! Your very own Pokèmon legend is about to unfold! A world of drams and adventures with Pokèmon awaits! Let's go!

Let's get you a starter Pokèmon! Respond with the name of the Pokèmon you would like to have to get you started.

{{ range .Starters }}
	*{{ .Name }}* (Pokedex No. {{ .ID }})
	{{ range $key, $value := .Types }}
		Type {{ $key }}: {{ $value }}
	{{ end }}
	Height: {{ .Height }}
	Weight: {{ .Weight }}


{{end}}
`
