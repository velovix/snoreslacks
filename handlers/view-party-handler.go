package handlers

import (
	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pkmn"
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

// ViewParty manages requests to print the trainer's full party.
type ViewParty struct {
	Services
}

// makeSinglePartyEntry returns template info on a single Pokemon in the party
// given the information that can't necessarily be descerned from the Pokemon
// object.
func (h *ViewParty) makeSinglePartyEntry(p *pkmn.Pokemon, currHP int, statusCondition string) viewSinglePokemonTemplateInfo {
	return viewSinglePokemonTemplateInfo{
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
		StatusCondition: statusCondition}
}

func (h *ViewParty) runTask(ctx context.Context, s Services) error {
	// Load request-specific objects
	client := ctx.Value("client").(messaging.Client)
	requester := ctx.Value("requesting trainer").(*basicTrainerData)
	battleData := ctx.Value("battle data").(*battleData)

	// The trainer is in a battle if the battle data is fully constructed
	inBattle := battleData.isComplete()

	var viewPartyTemplateInfo []viewSinglePokemonTemplateInfo

	if inBattle {
		s.Log.Infof(ctx, "the trainer is in a battle so we will show additional information")
	}

	// Create a party entry for each Pokemon
	for _, val := range requester.pkmn {
		p := val.GetPokemon()

		var statusCondition string
		var currHP int
		if inBattle {
			// Fill in special Pokemon in-battle data if need be

			inBattleStats, err := s.DB.LoadPokemonBattleInfo(ctx, battleData.battle, p.UUID)
			if err != nil {
				return handlerError{user: "could not fetch Pokemon battle info", err: err}
			}

			statusCondition = partyInfoAilmentText(inBattleStats.GetPokemonBattleInfo().Ailment)
			currHP = inBattleStats.GetPokemonBattleInfo().CurrHP
		}

		// Add a new entry to the party list
		viewPartyTemplateInfo = append(viewPartyTemplateInfo, h.makeSinglePartyEntry(p, currHP, statusCondition))
	}

	// Send the template
	var err error
	if inBattle {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     viewPartyInBattleTemplate,
			TemplInfo: viewPartyTemplateInfo})
	} else {
		err = messaging.SendTempl(client, requester.lastContactURL, messaging.TemplMessage{
			Templ:     viewPartyTemplate,
			TemplInfo: viewPartyTemplateInfo})
	}
	if err != nil {
		return handlerError{user: "could not populate view party template", err: err}
	}

	return nil
}
