package game

// SkillDamageBurstHit is one authoritative hit inside a skill_damage_burst event.
type SkillDamageBurstHit struct {
	TargetEntityID  string `json:"target_entity_id"`
	MonsterDefID    string `json:"monster_def_id,omitempty"`
	Damage          *int   `json:"damage,omitempty"`
	DamageType      string `json:"damage_type,omitempty"`
	Outcome         string `json:"outcome,omitempty"`
	Killed          bool   `json:"killed,omitempty"`
	Critical        *bool  `json:"critical,omitempty"`
	Blocked         *bool  `json:"blocked,omitempty"`
	RawDamage       *int   `json:"raw_damage,omitempty"`
	MitigatedDamage *int   `json:"mitigated_damage,omitempty"`
}

func skillDamageBurstHitFromEvent(ev Event) SkillDamageBurstHit {
	hit := SkillDamageBurstHit{
		TargetEntityID: ev.TargetEntityID,
		MonsterDefID:   ev.MonsterDefID,
		Damage:         ev.Damage,
		DamageType:     ev.DamageType,
		Outcome:        ev.Outcome,
		Critical:       ev.Critical,
		Blocked:        ev.Blocked,
		RawDamage:      ev.RawDamage,
		MitigatedDamage: ev.MitigatedDamage,
	}
	if ev.EventType == "monster_killed" {
		hit.Killed = true
	}
	return hit
}

func (s *Sim) collapseSkillDamageBurst(res *TickResult, playerID uint64, skillID, correlationID string, startEventLen int) {
	if startEventLen >= len(res.Events) {
		return
	}
	hits := make([]SkillDamageBurstHit, 0)
	kept := append([]Event(nil), res.Events[:startEventLen]...)
	for _, ev := range res.Events[startEventLen:] {
		if ev.EventType == "monster_damaged" && ev.SkillID == skillID && ev.CorrelationID == correlationID {
			hits = append(hits, skillDamageBurstHitFromEvent(ev))
			continue
		}
		kept = append(kept, ev)
	}
	if len(hits) == 0 {
		return
	}
	res.Events = kept
	res.Events = append(res.Events, Event{
		EventType:     "skill_damage_burst",
		EntityID:      idStr(playerID),
		SourceEntityID: idStr(playerID),
		SkillID:       skillID,
		CorrelationID: correlationID,
		Hits:          hits,
	})
}

func (s *Sim) handleInstantProjectileSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	castRange := s.effectiveProjectileRangeForSkill(skillID, def.Projectile.Range)
	dir, targetID, rejectReason := s.skillCastDirectionWithRange(def, in.CastSkill, player, castRange)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, castRange, true)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}
	if s.maybeBeginDirectionalSkillAutoNav(in, res, player, def, dir, castRange, true) {
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendProjectileSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, dir, def.Projectile)
	targets := s.rangerLineTargets(player, dir, def.Projectile.Range)
	if len(targets) > 0 {
		startEvents := len(res.Events)
		damageRange := s.scaleSkillDamageForMagic(def, rank, s.skillDamageRangeForSkill(skillID, def, rank))
		target := targets[0].Target
		s.damageMonsterByPlayerSkillTypedWithID(target, player.id, skillID, in.CorrelationID, res, damageRange, s.skillDamageType(def))
		for i := startEvents; i < len(res.Events); i++ {
			if res.Events[i].EventType == "monster_damaged" && res.Events[i].TargetEntityID == idStr(target.id) {
				res.Events[i].SkillID = skillID
			}
		}
	}
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func skillUsesInstantResolution(def SkillDef) bool {
	switch def.Resolution {
	case "instant_ray", "instant_aoe":
		return true
	default:
		return def.Volley.ArrowCount > 0 || def.Pierce.MaxHits > 0 || def.Root.DurationTicks > 0
	}
}
