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
	viewPartyTemplateInfo := make([]viewSinglePokemonTemplateInfo, 0)

	// Check if the trainer is in a battle
	b, found, err := db.LoadBattleTrainerIsIn(ctx, currTrainer.GetTrainer().Name)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not fetch battle trainer may be in")
		log.Errorf(ctx, "while checking if the trainer is in a battle: %s", err)
		return
	}
	inBattle := found

	for _, val := range currTrainer.pkmn {
		p := val.GetPokemon()

		var statusCondition string
		var currHP int
		if inBattle {
			// Fill in special Pokemon in-battle data if need be

			var statusCondition string
			inBattleStats, found, err := db.LoadPokemonBattleInfo(ctx, b, p.UUID)
			if err != nil {
				regularSlackRequest(client, currTrainer.lastContactURL, "could not fetch Pokemon battle info")
				log.Errorf(ctx, "while fetching Pokemon battle info: %s", err)
				return
			}

			if found {
				// The Pokemon has been in this battle, so there is battle info available

				switch inBattleStats.GetPokemonBattleInfo().Ailment {
				case pkmn.PoisonAilment:
					statusCondition += "poisoned "
				case pkmn.FreezeAilment:
					statusCondition += "frozen "
				case pkmn.ParalysisAilment:
					statusCondition += "paralyzed "
				case pkmn.BurnAilment:
					statusCondition += "burned "
				case pkmn.SleepAilment:
					statusCondition += "asleep "
				}

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

	// Populate the template
	templData := &bytes.Buffer{}
	err = viewPartyTemplate.Execute(templData, viewPartyTemplateInfo)
	if err != nil {
		regularSlackRequest(client, currTrainer.lastContactURL, "could not populate view party template")
		log.Errorf(ctx, "while populating view party template: %s", err)
		return
	}

	regularSlackRequest(client, currTrainer.lastContactURL, string(templData.Bytes()))
}
