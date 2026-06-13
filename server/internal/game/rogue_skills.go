package game

import "math"

type poisonDotState struct {
	SourcePlayerID uint64
	TargetID       uint64
	SkillID        string
	Rank           int
	DamagePerTick  int
	NextTick       uint64
	RemainingTicks int
	CorrelationID  string
}

func clonePoisonDots(in map[uint64]poisonDotState) map[uint64]poisonDotState {
	if len(in) == 0 {
		return make(map[uint64]poisonDotState)
	}
	out := make(map[uint64]poisonDotState, len(in))
	for targetID, dot := range in { //nolint:determinism — pure map clone, output is a map
		out[targetID] = dot
	}
	return out
}

func (s *Sim) dashRange(def SkillDef, rank int) float64 {
	if rank < 1 {
		rank = 1
	}
	r := def.Dash.RangeBase + def.Dash.RangePerRank*float64(rank-1)
	if r <= 0 {
		return def.Cone.Range
	}
	return r
}

func (s *Sim) dashDamagePercent(def SkillDef, rank int) int {
	percent := def.Dash.DamagePercentBase
	if def.Dash.DamagePercentPerMagic > 0 {
		bonus := s.progression.BaseStats.Magic * def.Dash.DamagePercentPerMagic
		if bonus > def.Dash.MaxDamageBonusPercent {
			bonus = def.Dash.MaxDamageBonusPercent
		}
		percent += bonus
	}
	if percent < 1 {
		return 1
	}
	return percent
}

func (s *Sim) handleDashSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	dashRange := s.dashRange(def, rank)
	dir, targetID, rejectReason := s.skillCastDirectionWithRange(def, in.CastSkill, player, dashRange)
	if rejectReason != "" {
		if rejectReason == "target_out_of_range" && in.CastSkill != nil && in.CastSkill.TargetID != "" {
			s.beginSkillAutoNav(in, res, dashRange, false)
			return
		}
		res.reject(in.MessageID, rejectReason)
		return
	}
	targets := s.dashSkillTargets(player, dir, dashRange)
	if len(targets) == 0 {
		res.reject(in.MessageID, "no_valid_targets")
		return
	}

	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	start := player.pos
	player.pos = s.resolveDashEndpoint(player.pos, dir, s.dashTravelDistance(player, dir, dashRange, targets))
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	s.appendConeSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, dir, SkillConeDef{
		Range:        dashRange,
		AngleDegrees: def.Cone.AngleDegrees,
		DamageSource: def.Cone.DamageSource,
	})
	res.Events[len(res.Events)-1].Position = cloneVec2Ptr(&start)
	s.applyDashSkill(player, skillID, def, rank, targets, in.CorrelationID, res)
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) dashSkillTargets(player *entity, dir Vec2, dashRange float64) []*entity {
	targets := []*entity{}
	if player == nil {
		return targets
	}
	dir = normalize(dir)
	if dir.X == 0 && dir.Y == 0 {
		return targets
	}
	end := Vec2{X: player.pos.X + dir.X*dashRange, Y: player.pos.Y + dir.Y*dashRange}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		target := s.activeLevel().entities[id]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 {
			continue
		}
		if _, ok := segmentIntersectsCircle(player.pos, end, target.pos, monsterRadius+playerRadius); !ok {
			continue
		}
		targets = append(targets, target)
	}
	return targets
}

func (s *Sim) dashTravelDistance(player *entity, dir Vec2, dashRange float64, targets []*entity) float64 {
	if player == nil || len(targets) == 0 {
		return dashRange
	}
	dir = normalize(dir)
	farthest := 0.0
	for _, target := range targets {
		if target == nil {
			continue
		}
		toTarget := Vec2{X: target.pos.X - player.pos.X, Y: target.pos.Y - player.pos.Y}
		projection := toTarget.X*dir.X + toTarget.Y*dir.Y
		if projection > farthest {
			farthest = projection
		}
	}
	travel := farthest + playerRadius + monsterRadius + 0.25
	if travel <= 0 || travel > dashRange {
		return dashRange
	}
	return travel
}

func (s *Sim) resolveDashEndpoint(start Vec2, dir Vec2, dashRange float64) Vec2 {
	dir = normalize(dir)
	end := start
	steps := int(math.Ceil(dashRange / 0.25))
	if steps < 1 {
		steps = 1
	}
	step := dashRange / float64(steps)
	for i := 0; i < steps; i++ {
		candidate := Vec2{X: end.X + dir.X*step, Y: end.Y + dir.Y*step}
		if s.playerDashPositionBlocked(candidate) {
			return end
		}
		end = candidate
	}
	return end
}

func (s *Sim) playerDashPositionBlocked(pos Vec2) bool {
	for _, wall := range s.activeWalls() {
		if circleIntersectsAABB(pos, playerRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		e := s.activeLevel().entities[id]
		if e == nil || e.kind != interactableEntity || e.state != interactableClosed {
			continue
		}
		if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
			if circleIntersectsAABB(pos, playerRadius, e.pos, def.BarrierWhenClosed.Size) {
				return true
			}
		}
	}
	return false
}

func (s *Sim) applyDashSkill(player *entity, skillID string, def SkillDef, rank int, targets []*entity, correlationID string, res *TickResult) {
	damageRange := percentDamageRange(s.resolvePlayerAttackDamage(), s.dashDamagePercent(def, rank))
	for _, target := range targets {
		if target == nil || target.hp <= 0 {
			continue
		}
		beforeEvents := len(res.Events)
		s.damageMonsterByPlayerSkillTyped(target, player.id, correlationID, res, damageRange, s.skillDamageType(def))
		for i := beforeEvents; i < len(res.Events); i++ {
			if res.Events[i].EventType == "monster_damaged" && res.Events[i].TargetEntityID == idStr(target.id) {
				res.Events[i].SkillID = skillID
			}
		}
	}
}

func percentDamageRange(base DamageRange, percent int) DamageRange {
	if percent < 1 {
		percent = 1
	}
	minDamage := int(math.Round(float64(base.Min) * float64(percent) / 100.0))
	maxDamage := int(math.Round(float64(base.Max) * float64(percent) / 100.0))
	if minDamage < 1 {
		minDamage = 1
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func (s *Sim) startPoisonDot(player *entity, target *entity, skillID string, def SkillDef, sourceDamage int, correlationID string, res *TickResult) {
	if player == nil || target == nil || sourceDamage < 0 {
		return
	}
	rank := s.effectiveSkillRank(skillID)
	if rank < 1 {
		rank = 1
	}
	percent := def.Poison.DamagePercentBase + def.Poison.DamagePercentPerRank*(rank-1)
	damage := int(math.Round(float64(sourceDamage) * float64(percent) / 100.0))
	if sourceDamage > 0 && damage < 1 {
		damage = 1
	}
	duration := def.Poison.DurationTicks + s.progression.BaseStats.Magic*def.Poison.MagicDurationTicksPerPoint
	if duration < 1 {
		duration = 1
	}
	s.poisonDots[target.id] = poisonDotState{
		SourcePlayerID: player.id,
		TargetID:       target.id,
		SkillID:        skillID,
		Rank:           rank,
		DamagePerTick:  damage,
		NextTick:       s.tick + 10,
		RemainingTicks: duration,
		CorrelationID:  correlationID,
	}
	res.Events = append(res.Events, Event{
		EventType:      "skill_effect_started",
		EntityID:       idStr(target.id),
		SourceEntityID: idStr(player.id),
		TargetEntityID: idStr(target.id),
		CorrelationID:  correlationID,
		SkillID:        skillID,
		Rank:           intPtr(rank),
		Amount:         intPtr(damage),
		RemainingTicks: intPtr(duration),
		TotalTicks:     intPtr(duration),
	})
}

func (s *Sim) advancePoisonDots(res *TickResult) {
	if len(s.poisonDots) == 0 {
		return
	}
	for _, targetID := range sortedUint64Keys(s.poisonDots) {
		dot := s.poisonDots[targetID]
		target := s.activeLevel().entities[targetID]
		if target == nil || target.kind != monsterEntity || target.hp <= 0 || dot.RemainingTicks <= 0 {
			delete(s.poisonDots, targetID)
			continue
		}
		if s.tick < dot.NextTick {
			continue
		}
		rawDamage := dot.DamagePerTick
		damage := s.applyResistanceToDamage(rawDamage, s.monsterResistance(target, damageTypePoison))
		if damage > target.hp {
			damage = target.hp
		}
		target.hp -= damage
		dot.RemainingTicks -= 10
		dot.NextTick += 10
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
		res.Events = append(res.Events, Event{
			EventType:       "monster_damaged",
			EntityID:        idStr(target.id),
			SourceEntityID:  idStr(dot.SourcePlayerID),
			TargetEntityID:  idStr(target.id),
			CorrelationID:   dot.CorrelationID,
			SkillID:         dot.SkillID,
			Rank:            intPtr(dot.Rank),
			Damage:          intPtr(damage),
			DamageType:      damageTypePoison,
			Outcome:         "hit",
			RawDamage:       intPtr(rawDamage),
			MitigatedDamage: intPtr(rawDamage),
		})
		if target.hp == 0 {
			s.finishMonsterKill(target, dot.SourcePlayerID, dot.CorrelationID, res)
			delete(s.poisonDots, targetID)
			continue
		}
		if dot.RemainingTicks <= 0 {
			delete(s.poisonDots, targetID)
			res.Events = append(res.Events, Event{
				EventType:      "skill_effect_ended",
				EntityID:       idStr(target.id),
				SourceEntityID: idStr(dot.SourcePlayerID),
				TargetEntityID: idStr(target.id),
				SkillID:        dot.SkillID,
			})
			continue
		}
		s.poisonDots[targetID] = dot
	}
}

func (s *Sim) consumeBasicAttack(in Input, res *TickResult) (string, bool) {
	if s.tick >= s.nextBasicAttackTick {
		s.nextBasicAttackTick = s.tick + uint64(s.DerivedStatsView().AttackIntervalTicks)
		return mainHandSlot, true
	}
	if s.rogueOffHandReady() {
		s.nextOffHandAttackTick = s.tick + uint64(s.offHandAttackIntervalTicks())
		return offHandSlot, true
	}
	res.reject(in.MessageID, "basic_attack_on_cooldown")
	return "", false
}

func (s *Sim) resolvePlayerAttackDamageForSlot(slot string) DamageRange {
	if slot != offHandSlot {
		return s.resolvePlayerAttackDamage()
	}
	item := s.findItemByID(s.equipped[offHandSlot])
	if item == nil {
		return s.resolvePlayerAttackDamage()
	}
	character := s.characterDerivedStatsView()
	baseMin, baseMax, minRoll, maxRoll, _, _, ok := s.weaponDamageContributions(item)
	if !ok {
		return s.resolvePlayerAttackDamage()
	}
	minDamage := int(math.Floor(character.DamageMin + baseMin + minRoll))
	maxDamage := int(math.Floor(character.DamageMax + baseMax + maxRoll))
	if minDamage < 0 {
		minDamage = 0
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return DamageRange{Min: minDamage, Max: maxDamage}
}

func (s *Sim) rogueOffHandReady() bool {
	if s.progression.CharacterClass != "rogue" || s.tick < s.nextOffHandAttackTick {
		return false
	}
	item := s.findItemByID(s.equipped[offHandSlot])
	return s.canOffhandWeapon(item)
}

func (s *Sim) offHandAttackIntervalTicks() int {
	interval := int(math.Ceil(float64(s.DerivedStatsView().AttackIntervalTicks) / 1.5))
	if interval < 1 {
		return 1
	}
	return interval
}
