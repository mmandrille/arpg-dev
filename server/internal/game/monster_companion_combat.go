package game

import "math"

func (s *Sim) monsterAttackTarget(monster *entity, player *entity, def MonsterDef) *entity {
	if monster == nil || player == nil || player.hp <= 0 {
		return nil
	}
	if companion := s.monsterEngagedCompanionTarget(monster, player, def); companion != nil {
		return companion
	}
	if s.monsterInAttackRange(monster, player, def) {
		return player
	}
	return nil
}

func (s *Sim) monsterEngagedCompanionTarget(monster *entity, player *entity, def MonsterDef) *entity {
	var best *entity
	bestDist := math.MaxFloat64
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		companion := s.activeLevel().entities[id]
		if companion == nil || companion.kind != companionEntity || companion.ownerID != player.id || companion.hp <= 0 {
			continue
		}
		if companion.targetID != monster.id {
			continue
		}
		if !s.monsterInAttackRangeOfEntity(monster, companion, def, monsterRadius) {
			continue
		}
		dist := distance(monster.pos, companion.pos)
		if best == nil || dist < bestDist-1e-9 || (math.Abs(dist-bestDist) <= 1e-9 && companion.id < best.id) {
			best = companion
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) monsterInAttackRangeOfEntity(monster *entity, target *entity, def MonsterDef, targetRadius float64) bool {
	return meleeInRange(distance(target.pos, monster.pos), s.monsterAttackReach(def), targetRadius)
}

func (s *Sim) damageCompanionByMonster(monster *entity, companion *entity, damageRange DamageRange, corr string, res *TickResult) combatResolution {
	attackerStats := s.monsterEffectiveCombatStats(monster, damageRange)
	defenderStats := s.monsterEffectiveCombatStats(companion, DamageRange{})
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	if outcome.Hit && !outcome.Blocked {
		companion.hp -= outcome.Damage
		if companion.hp < 0 {
			companion.hp = 0
		}
	}
	if outcome.Hit && !outcome.Blocked && companion.hp > 0 {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(companion))})
	}
	eventType := s.companionCombatEventType(outcome, companion)
	res.Events = append(res.Events, combatEvent(eventType, monster.id, companion.id, corr, outcome))
	if companion.hp == 0 {
		s.finishCompanionDeath(companion, monster, res)
	}
	return outcome
}

func (s *Sim) companionCombatEventType(outcome combatResolution, companion *entity) string {
	if outcome.Outcome == "miss" {
		return "attack_missed"
	}
	if companion != nil && companion.hp == 0 {
		return "companion_killed"
	}
	return "companion_damaged"
}

func (s *Sim) finishCompanionDeath(companion *entity, killer *entity, res *TickResult) {
	if companion == nil || companion.kind != companionEntity {
		return
	}
	delete(s.activeLevel().entities, companion.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(companion.id)})
	if companion.sourceSkillID != mercenaryHireSourceID || companion.monsterDefID != mercenaryGuardMonsterDefID {
		return
	}
	res.Events = append(res.Events, Event{
		EventType:      "mercenary_lost",
		EntityID:       idStr(companion.id),
		SourceEntityID: idStr(entityID(killer)),
		TargetEntityID: idStr(companion.id),
		Service:        mercenaryService,
		OfferID:        mercenaryGuardOfferID,
		MonsterDefID:   mercenaryGuardMonsterDefID,
	})
}

func (s *Sim) resolveMonsterProjectileCompanionHit(p *entity, hit projectileHit, res *TickResult) {
	owner := s.activeLevel().entities[p.ownerID]
	target := s.activeLevel().entities[hit.entityID]
	if owner == nil || owner.kind != monsterEntity || owner.hp <= 0 || target == nil || target.kind != companionEntity || target.hp <= 0 {
		res.Events = append(res.Events, Event{EventType: "projectile_expired", CorrelationID: p.sourceCorrID})
		return
	}
	s.damageCompanionByMonster(owner, target, p.damageRange, p.sourceCorrID, res)
}
