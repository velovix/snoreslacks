package handlers

import (
	"net/http"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pkmn"

	"golang.org/x/net/context"
)

// partyInfoAilmentText returns text describing the given ailment in a format
// for the party viewer.
func partyInfoAilmentText(ailment pkmn.Ailment) string {
	switch ailment {
	case pkmn.NoAilment:
		return "fine"
	case pkmn.PoisonAilment:
		return "poisoned"
	case pkmn.FreezeAilment:
		return "frozen"
	case pkmn.ParalysisAilment:
		return "paralyzed"
	case pkmn.BurnAilment:
		return "burned"
	case pkmn.SleepAilment:
		return "asleep"
	default:
		panic("unsupported ailment type")
	}
}

// viewPartyHandler manages requests to print the trainer's full party.
func viewPartyHandler(ctx context.Context, db database.Database, log logging.Logger, client *http.Client, r slackRequest, currTrainer trainerData) {
	var viewPartyTemplateInfo []viewSinglePokemonTemplateInfo

	// Check if the trainer is in a battle
	b, inBattle, err := loadBattleTrainerIsIn(ctx, db, log, client, currTrainer.lastContactURL, false,
		"while checking if the trainer is in a battle",
		currTrainer.GetTrainer().Name)
	if err != nil {
		return // Abort operation
	}

	if inBattle {
		log.Infof(ctx, "the trainer is in a battle so we will show additional information")
	}

	for _, val := range currTrainer.pkmn {
		p := val.GetPokemon()

		var statusCondition string
		var currHP int
		if inBattle {
			// Fill in special Pokemon in-battle data if need be

			inBattleStats, found, err := loadPokemonBattleInfo(ctx, db, log, client, currTrainer.lastContactURL, false,
				"while fetching Pokemon battle info",
				b, p.UUID)
			if err != nil {
				return // Abort operation
			}

			if found {
				// The Pokemon has been in this battle, so there is battle info available
				statusCondition = partyInfoAilmentText(inBattleStats.GetPokemonBattleInfo().Ailment)
				currHP = inBattleStats.GetPokemonBattleInfo().CurrHP
			} else {
				// The trainer is in a battle, but this Pokemon has not been in
				// a battle yet, so we can assume the Pokemon is unscathed.
				currHP = pkmn.CalcOOBHP(p.HP, *p)
				statusCondition = "none"
			}
		}

		viewPartyTemplateInfo = append(viewPartyTemplateInfo,
			viewSinglePokemonTemplateInfo{
				Name:  p.Name,
				ID:    p.ID,
				Level: p.Level,
				Type1: p.Type1,
				Type2: p.Type2,

				HP:        pkmn.CalcOOBHP(p.HP, *p),
				Attack:    pkmn.CalcOOBStat(p.Attack, *p),
				Defense:   pkmn.CalcOOBStat(p.Defense, *p),
				SpAttack:  pkmn.CalcOOBStat(p.SpAttack, *p),
				SpDefense: pkmn.CalcOOBStat(p.SpDefense, *p),
				Speed:     pkmn.CalcOOBStat(p.Speed, *p),

				CurrHP:          currHP,
				StatusCondition: statusCondition})
	}

	// Send the template
	if inBattle {
		err = regularSlackTemplRequest(client, currTrainer.lastContactURL, viewPartyInBattleTemplate, viewPartyTemplateInfo)
	} else {
		err = regularSlackTemplRequest(client, currTrainer.lastContactURL, viewPartyTemplate, viewPartyTemplateInfo)
	}
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate view party template")
		log.Errorf(ctx, "while populating view party template: %s", err)
		return
	}
}
