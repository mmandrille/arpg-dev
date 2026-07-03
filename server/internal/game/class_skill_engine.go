package game

func (s *Sim) trySkillReflectOnPlayerBlock(player *entity, attacker *entity, damageRange DamageRange, corr string, res *TickResult) {
	if player == nil || attacker == nil || attacker.kind != monsterEntity || attacker.hp <= 0 {
		return
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		state := s.skillEffects[stateKey]
		if !state.ReflectOnBlock || state.TargetID != player.id || s.tick >= state.EndsTick || state.Percent <= 0 {
			continue
		}
		raw := s.rollRange(damageRange)
		reflect := percentOf(raw, state.Percent)
		if reflect < 1 {
			reflect = 1
		}
		s.applyUniqueDirectDamage(player.id, attacker, state.SkillID, reflect, damageTypeForce, corr, res)

		return
	}
}

func skillBleedValues(r *Rules, bleed SkillBleedDef, rank int) (damagePercentMaxHP, durationTicks, intervalTicks int) {
	if rank < 1 {
		rank = 1
	}
	if r == nil {
		return 0, 0, 10
	}
	damagePercentMaxHP = r.rankScaledInt(bleed.DamagePercentMaxHP, bleed.DamagePercentMaxHPPerRank, rank)
	durationTicks = r.rankScaledInt(bleed.DurationTicks, bleed.DurationTicksPerRank, rank)
	intervalTicks = bleed.IntervalTicks
	if intervalTicks <= 0 {
		intervalTicks = 10
	}
	if durationTicks < 1 {
		durationTicks = 1
	}
	if damagePercentMaxHP < 1 {
		damagePercentMaxHP = 1
	}

	return damagePercentMaxHP, durationTicks, intervalTicks
}

func skillMarkValues(r *Rules, mark SkillMarkDef, rank int) (damageBonusPercent, durationTicks int) {
	if rank < 1 {
		rank = 1
	}
	if r == nil {
		return 0, 0
	}
	damageBonusPercent = r.rankScaledInt(mark.DamageBonusPercent, mark.DamageBonusPercentPerRank, rank)
	durationTicks = mark.DurationTicks
	if damageBonusPercent < 1 {
		damageBonusPercent = 1
	}
	if durationTicks < 1 {
		durationTicks = 1
	}

	return damageBonusPercent, durationTicks
}

func (s *Sim) startSkillBleed(player *entity, target *entity, skillID string, bleed SkillBleedDef, rank int, correlationID string, res *TickResult) {
	if player == nil || target == nil || target.kind != monsterEntity || target.hp <= 0 || bleed.DurationTicks <= 0 {
		return
	}
	damagePercentMaxHP, durationTicks, intervalTicks := skillBleedValues(s.rules, bleed, rank)
	effectID := bleed.EffectID
	if effectID == "" {
		effectID = "bleed"
	}
	if s.bleedDots == nil {
		s.bleedDots = make(map[uint64]bleedDotState)
	}
	dot := bleedDotState{
		SourcePlayerID:     player.id,
		TargetID:           target.id,
		SkillID:            skillID,
		EffectID:           effectID,
		DamagePercentMaxHP: damagePercentMaxHP,
		IntervalTicks:      intervalTicks,
		NextTick:           s.tick + uint64(intervalTicks),
		RemainingTicks:     durationTicks,
		TotalTicks:         durationTicks,
		CorrelationID:      correlationID,
	}
	s.bleedDots[target.id] = dot
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(player.id),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Amount:         intPtr(damagePercentMaxHP),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}

func (s *Sim) startSkillMark(player *entity, target *entity, skillID string, mark SkillMarkDef, rank int, correlationID string, res *TickResult) {
	if player == nil || target == nil || target.kind != monsterEntity || target.hp <= 0 || mark.DurationTicks <= 0 {
		return
	}
	damageBonusPercent, durationTicks := skillMarkValues(s.rules, mark, rank)
	if s.rogueMarks == nil {
		s.rogueMarks = make(map[uint64]rogueMarkState)
	}
	effectID := mark.EffectID
	if effectID == "" {
		effectID = "rogue_mark"
	}
	markState := rogueMarkState{
		SourcePlayerID:     player.id,
		TargetID:           target.id,
		SkillID:            skillID,
		Rank:               rank,
		DamageBonusPercent: damageBonusPercent,
		EndsTick:           s.tick + uint64(durationTicks),
		TotalTicks:         durationTicks,
		EffectID:           effectID,
		CorrelationID:      correlationID,
	}
	s.rogueMarks[target.id] = markState
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(player.id),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Rank:           intPtr(rank),
		Amount:         intPtr(damageBonusPercent),
		RemainingTicks: intPtr(durationTicks),
		TotalTicks:     intPtr(durationTicks),
	})
}
