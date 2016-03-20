package app

import (
	"bytes"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// viewPartyHandler manages requests to print the trainer's full party.
func viewPartyHandler(w http.ResponseWriter, r *http.Request, ctx context.Context, currTrainer trainer) {
	viewPartyTemplateInfo := make([]viewSinglePokemonTemplateInfo, len(currTrainer.pkmn))

	for _, val := range currTrainer.pkmn {
		viewPartyTemplateInfo = append(viewPartyTemplateInfo,
			viewSinglePokemonTemplateInfo{
				Name:      val.Name,
				ID:        val.ID,
				Level:     val.Level,
				Type1:     val.Type1,
				Type2:     val.Type2,
				HP:        calcHP(val.HP, val),
				Attack:    calcStat(val.Attack, val),
				Defense:   calcStat(val.Defense, val),
				SpAttack:  calcStat(val.SpAttack, val),
				SpDefense: calcStat(val.SpDefense, val),
				Speed:     calcStat(val.Speed, val)})
	}

	// Populate the template
	templData := &bytes.Buffer{}
	err := viewPartyTemplate.Execute(templData, viewPartyTemplateInfo)
	if err != nil {
		http.Error(w, "could not populate view party template", 500)
		log.Errorf(ctx, "while populating view party template: %s", err)
		return
	}

	regularSlackResponse(w, r, string(templData.Bytes()))
}
