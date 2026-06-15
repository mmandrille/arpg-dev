package game

import "fmt"

func (s *Sim) applySkillBuff(player *entity, skillID string, def SkillDef, rank int, correlationID string, res *TickResult) {
	if player == nil {
		return
	}
	for _, effect := range def.Effects {
		if effect.Type != "stat_percent_buff" {
			continue
		}
		percent := skillEffectPercent(effect, rank)
		scale := 1.0
		if effect.VisualScale {
			scale += float64(percent) / 100.0
		}
		totalTicks := effect.DurationTicks
		s.skillEffects[skillID] = skillEffectState{
			SkillID:     skillID,
			TargetID:    player.id,
			Stats:       cloneStringSlice(effect.Stats),
			Percent:     percent,
			VisualScale: scale,
			EffectID:    skillID,
			EndsTick:    s.tick + uint64(totalTicks),
			TotalTicks:  totalTicks,
		}
		s.syncActivePlayerVisualScale()
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(player.id),
			CorrelationID:  correlationID,
			SkillID:        skillID,
			Rank:           intPtr(rank),
			Amount:         intPtr(percent),
			RemainingTicks: intPtr(totalTicks),
			TotalTicks:     intPtr(totalTicks),
		})
	}
	s.appendCharacterProgressionUpdate(res)
}

func (s *Sim) areaStatBuffApplications(player *entity, def SkillDef, rank int, cast *CastSkillIntent) ([]skillBuffApplication, string) {
	if player == nil {
		return nil, "player_dead"
	}
	applications := []skillBuffApplication{}
	for _, effect := range def.Effects {
		if effect.Type != "area_stat_percent_buff" && effect.Type != "area_immunity_buff" {
			continue
		}
		percent := 0
		if effect.Type == "area_stat_percent_buff" {
			percent = s.scaleSkillPercentForMagic(def, rank, effect, skillEffectPercent(effect, rank))
		}
		targets := s.healSkillTargets(player.pos, effect, player.id, s.scaleSkillRadiusForMagic(def, rank, effect))
		for _, target := range targets {
			applications = append(applications, skillBuffApplication{
				Target:      target,
				Effect:      effect,
				Percent:     percent,
				VisualScale: 1.0,
			})
		}
	}
	return applications, ""
}

func (s *Sim) applyAreaStatBuff(player *entity, skillID string, rank int, applications []skillBuffApplication, correlationID string, res *TickResult) {
	activePlayerProgressionChanged := false
	for _, app := range applications {
		target := app.Target
		if target == nil || target.kind != playerEntity || target.hp <= 0 {
			continue
		}
		effectID := app.Effect.EffectID
		if effectID == "" {
			effectID = skillID
		}
		stateKey := skillID
		if target.id != s.playerID {
			stateKey = fmt.Sprintf("%s:%d", skillID, target.id)
		}
		totalTicks := app.Effect.DurationTicks
		s.skillEffects[stateKey] = skillEffectState{
			SkillID:     skillID,
			TargetID:    target.id,
			Stats:       cloneStringSlice(app.Effect.Stats),
			Percent:     app.Percent,
			VisualScale: app.VisualScale,
			EffectID:    effectID,
			EndsTick:    s.tick + uint64(totalTicks),
			TotalTicks:  totalTicks,
		}
		target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
		if target.id == s.playerID {
			activePlayerProgressionChanged = true
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:      "skill_effect_started",
			EntityID:       idStr(target.id),
			SourceEntityID: idStr(player.id),
			TargetEntityID: idStr(target.id),
			CorrelationID:  correlationID,
			SkillID:        skillID,
			Rank:           intPtr(rank),
			Amount:         intPtr(app.Percent),
			RemainingTicks: intPtr(totalTicks),
			TotalTicks:     intPtr(totalTicks),
		})
	}
	if activePlayerProgressionChanged {
		s.appendCharacterProgressionUpdate(res)
	}
}
