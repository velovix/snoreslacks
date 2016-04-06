package handlers

import (
	"bytes"
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// viewPartyHandler manages requests to print the trainer's full party.
func viewPartyHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	viewPartyTemplateInfo := make([]viewSinglePokemonTemplateInfo, len(currTrainer.pkmn))

	for _, val := range currTrainer.pkmn {
		p := val.GetPokemon()
		viewPartyTemplateInfo = append(viewPartyTemplateInfo,
			viewSinglePokemonTemplateInfo{
				Name:      p.Name,
				ID:        p.ID,
				Level:     p.Level,
				Type1:     p.Type1,
				Type2:     p.Type2,
				HP:        pkmn.CalcHP(p.HP, *p),
				Attack:    pkmn.CalcStat(p.Attack, *p),
				Defense:   pkmn.CalcStat(p.Defense, *p),
				SpAttack:  pkmn.CalcStat(p.SpAttack, *p),
				SpDefense: pkmn.CalcStat(p.SpDefense, *p),
				Speed:     pkmn.CalcStat(p.Speed, *p)})
	}

	// Populate the template
	templData := &bytes.Buffer{}
	err := viewPartyTemplate.Execute(templData, viewPartyTemplateInfo)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate view party template")
		log.Errorf(ctx, "while populating view party template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
}
