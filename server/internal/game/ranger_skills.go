package game

import (
	"math"
	"sort"
)

type rangerLineTarget struct {
	Target *entity
	T      float64
}

func (s *Sim) handleRangerProjectileSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	dir, targetID, rejectReason := s.skillCastDirection(def, in.CastSkill, player)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, def.Projectile.Range, true)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}
	targets := s.rangerLineTargets(player, dir, def.Projectile.Range)
	if len(targets) == 0 {
		res.reject(in.MessageID, "no_valid_targets")
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendRangerShotCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, dir, def)
	s.applyRangerShot(player, skillID, def, rank, targets, in.CorrelationID, res)
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) appendRangerShotCastEvent(res *TickResult, player *entity, skillID string, rank int, manaCost int, correlationID string, targetID uint64, dir Vec2, def SkillDef) {
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, correlationID, targetID, def.Projectile.Visual)
	if len(res.Events) == 0 {
		return
	}
	event := &res.Events[len(res.Events)-1]
	event.Position = cloneVec2Ptr(&player.pos)
	event.Direction = cloneVec2Ptr(&dir)
	event.Range = floatPtr(def.Projectile.Range)
}

func (s *Sim) rangerLineTargets(player *entity, dir Vec2, shotRange float64) []rangerLineTarget {
	if player == nil || shotRange <= 0 {
		return []rangerLineTarget{}
	}
	dir = normalize(dir)
	if dir.X == 0 && dir.Y == 0 {
		return []rangerLineTarget{}
	}
	start := player.pos
	end := Vec2{X: start.X + dir.X*shotRange, Y: start.Y + dir.Y*shotRange}
	targets := []rangerLineTarget{}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		t, ok := segmentIntersectsCircle(start, end, target.pos, monsterRadius+projectileRadius)
		if !ok {
			continue
		}
		targets = append(targets, rangerLineTarget{Target: target, T: t})
	}
	sort.SliceStable(targets, func(i, j int) bool {
		if math.Abs(targets[i].T-targets[j].T) > 0.000000001 {
			return targets[i].T < targets[j].T
		}
		return targets[i].Target.id < targets[j].Target.id
	})
	return targets
}

func (s *Sim) applyRangerShot(player *entity, skillID string, def SkillDef, rank int, targets []rangerLineTarget, correlationID string, res *TickResult) {
	maxHits := 1
	if def.Pierce.MaxHits > 0 {
		maxHits = def.Pierce.MaxHits
	}
	damageRange := s.scaleSkillDamageForMagic(def, rank, skillDamageRange(def, rank))
	for hitIndex, row := range targets {
		if hitIndex >= maxHits {
			break
		}
		target := row.Target
		if target == nil || target.hp <= 0 {
			continue
		}
		hitDamage := damageRange
		if hitIndex > 0 && def.Pierce.DamagePercentPerExtraHit > 0 {
			hitDamage = scaleDamageRangePercent(damageRange, def.Pierce.DamagePercentPerExtraHit)
		}
		beforeEvents := len(res.Events)
		outcome := s.damageMonsterByPlayerSkillTyped(target, player.id, correlationID, res, hitDamage, s.skillDamageType(def))
		for i := beforeEvents; i < len(res.Events); i++ {
			if res.Events[i].EventType == "monster_damaged" && res.Events[i].TargetEntityID == idStr(target.id) {
				res.Events[i].SkillID = skillID
			}
		}
		if def.Root.DurationTicks > 0 && outcome.Damage > 0 && target.hp > 0 {
			s.applyMonsterRoot(target, player.id, skillID, def.Root, correlationID, res)
			break
		}
	}
}

func scaleDamageRangePercent(in DamageRange, percent int) DamageRange {
	if percent >= 100 {
		return in
	}
	if percent <= 0 {
		return DamageRange{}
	}
	return DamageRange{
		Min: scaleDamageValuePercent(in.Min, percent),
		Max: scaleDamageValuePercent(in.Max, percent),
	}
}

func scaleDamageValuePercent(value int, percent int) int {
	if value <= 0 {
		return 0
	}
	out := int(math.Floor(float64(value) * float64(percent) / 100.0))
	if out < 1 {
		return 1
	}
	return out
}

func (s *Sim) applyMonsterRoot(target *entity, sourceID uint64, skillID string, root SkillRootDef, correlationID string, res *TickResult) {
	if target == nil || target.kind != monsterEntity || target.hp <= 0 || root.DurationTicks <= 0 {
		return
	}
	effectID := root.EffectID
	if effectID == "" {
		effectID = skillID
	}
	stateKey := "root:" + skillID + ":" + idStr(target.id)
	s.skillEffects[stateKey] = skillEffectState{
		SkillID:    skillID,
		TargetID:   target.id,
		Stats:      []string{"root"},
		Percent:    100,
		EffectID:   effectID,
		EndsTick:   s.tick + uint64(root.DurationTicks),
		TotalTicks: root.DurationTicks,
	}
	target.effectIDs = sortedUniqueStrings(append(target.effectIDs, effectID))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(sourceID),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Amount:         intPtr(100),
		RemainingTicks: intPtr(root.DurationTicks),
		TotalTicks:     intPtr(root.DurationTicks),
	})
}

func (s *Sim) monsterRooted(monster *entity) bool {
	if monster == nil {
		return false
	}
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.TargetID == monster.id && effect.EndsTick > s.tick && containsStringValue(effect.Stats, "root") {
			return true
		}
	}
	return false
}
