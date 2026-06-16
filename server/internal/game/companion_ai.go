package game

import "math"

const (
	companionAssistRadius     = 5.0
	companionFollowDistance   = 1.1
	companionFollowStopRadius = 1.6
)

func (s *Sim) newPresetMonsterOrCompanion(level *LevelState, preset WorldEntity, ownerID uint64) *entity {
	def := s.rules.Monsters[preset.MonsterDefID]
	monster := &entity{
		kind:         preset.Type,
		pos:          preset.Position,
		spawnPos:     preset.Position,
		hp:           def.MaxHP,
		maxHP:        def.MaxHP,
		monsterDefID: preset.MonsterDefID,
		lootTable:    def.LootTable,
		aiMode:       monsterAIModeIdle,
	}
	if preset.Type == companionEntity {
		monster.ownerID = ownerID
		monster.monsterAttackDamage = def.AttackDamage
		monster.monsterAttackCooldown = def.AttackCooldown
		return monster
	}
	s.applyPartyHPScale(level, monster)
	return monster
}

func (s *Sim) summonCompanion(owner *entity, skillID string, def SkillDef, rank int, res *TickResult) *entity {
	if owner == nil {
		return nil
	}
	level := s.activeLevel()
	s.pruneCompanionsForNewSpawn(owner.id, skillID, companionLimitAtRank(def.Companion.Limit, rank), res)
	monsterDef := s.rules.Monsters[def.Companion.MonsterDefID]
	statPercent := companionHeroStatPercent(def, rank)
	hp := monsterDef.MaxHP
	damage := monsterDef.AttackDamage
	visualScale := def.Companion.VisualScale
	if statPercent > 0 {
		hp = scalePositiveInt(owner.maxHP, statPercent)
		derived := s.DerivedStatsView()
		damage = &DamageRange{
			Min: scalePositiveInt(int(math.Round(derived.DamageMin)), statPercent),
			Max: scalePositiveInt(int(math.Round(derived.DamageMax)), statPercent),
		}
		visualScale = float64(statPercent) / 100.0
	}
	companion := &entity{
		kind:                  companionEntity,
		pos:                   companionSpawnPosition(owner),
		spawnPos:              owner.pos,
		hp:                    hp,
		maxHP:                 hp,
		ownerID:               owner.id,
		monsterDefID:          def.Companion.MonsterDefID,
		lootTable:             monsterDef.LootTable,
		speed:                 monsterDef.MoveSpeed,
		monsterAttackDamage:   damage,
		monsterAttackCooldown: monsterDef.AttackCooldown,
		aiMode:                monsterAIModeIdle,
		sourceSkillID:         skillID,
		visualModel:           def.Companion.VisualModel,
		visualTint:            def.Companion.VisualTint,
		visualScale:           visualScale,
	}
	companion.id = s.alloc()
	level.entities[companion.id] = companion
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(companion))})
	return companion
}

func (s *Sim) handleSummonCompanionSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	companion := s.summonCompanion(player, skillID, def, rank, res)
	targetID := uint64(0)
	if companion != nil {
		targetID = companion.id
	}
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, "")
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) handleReviveCompanionSkillCast(in Input, res *TickResult, player *entity, skillID string, def SkillDef, rank int, manaCost int) {
	target, rejectReason := s.reviveCompanionTarget(in.CastSkill)
	if rejectReason != "" {
		res.reject(in.MessageID, rejectReason)
		return
	}
	s.activeLevel().move = nil
	s.clearAutoNav()
	cooldownTicks := s.commitSkillSpend(player, skillID, def, manaCost)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
	companion := s.reviveMonsterCompanion(player, target, skillID, def, rank, res)
	targetID := uint64(0)
	if companion != nil {
		targetID = companion.id
	}
	s.appendSkillCastEvent(res, player, skillID, rank, manaCost, in.CorrelationID, targetID, "")
	s.appendSkillCooldownUpdate(res)
	s.appendSkillCooldownStartedEvent(res, player, skillID, in.CorrelationID, cooldownTicks)
	res.ack(in.MessageID)
}

func (s *Sim) reviveCompanionTarget(cast *CastSkillIntent) (*entity, string) {
	if cast == nil || cast.TargetID == "" {
		return nil, "invalid_target"
	}
	target := s.findEntity(cast.TargetID)
	if target == nil || target.kind != monsterEntity {
		return nil, "invalid_target"
	}
	if target.isBoss {
		return nil, "cannot_revive_boss"
	}
	if target.hp > 0 {
		return nil, "target_not_dead"
	}
	return target, ""
}

func (s *Sim) reviveMonsterCompanion(owner *entity, target *entity, skillID string, def SkillDef, rank int, res *TickResult) *entity {
	if owner == nil || target == nil {
		return nil
	}
	level := s.activeLevel()
	s.pruneCompanionsForNewSpawn(owner.id, skillID, companionLimitAtRank(def.Revive.Limit, rank), res)
	delete(level.entities, target.id)
	res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(target.id)})

	monsterDef := s.rules.Monsters[target.monsterDefID]
	powerPercent := revivePowerPercent(def, rank)
	durationTicks := reviveDurationTicks(def, rank)
	maxHP := scalePositiveInt(monsterDef.MaxHP, powerPercent)
	companion := &entity{
		kind:                  companionEntity,
		pos:                   target.pos,
		spawnPos:              target.pos,
		hp:                    maxHP,
		maxHP:                 maxHP,
		ownerID:               owner.id,
		monsterDefID:          target.monsterDefID,
		lootTable:             "no_drop",
		speed:                 monsterDef.MoveSpeed,
		monsterAttackDamage:   scaleRevivedDamageRange(monsterDef.AttackDamage, powerPercent),
		monsterAttackCooldown: monsterDef.AttackCooldown,
		monsterHitChance:      target.monsterHitChance,
		monsterCritChance:     target.monsterCritChance,
		monsterBlockPercent:   target.monsterBlockPercent,
		monsterArmor:          target.monsterArmor,
		aiMode:                monsterAIModeIdle,
		sourceSkillID:         skillID,
		expiresTick:           s.tick + uint64(durationTicks) + 1,
		totalDurationTicks:    durationTicks,
		visualModel:           target.visualModel,
		visualTint:            target.visualTint,
		visualScale:           target.visualScale,
	}
	companion.id = s.alloc()
	level.entities[companion.id] = companion
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(companion))})
	return companion
}

func (s *Sim) pruneCompanionsForNewSpawn(ownerID uint64, skillID string, limit int, res *TickResult) {
	if limit < 1 {
		limit = 1
	}
	type companionRef struct {
		level *LevelState
		id    uint64
	}
	existingRefs := make([]companionRef, 0)
	for _, levelNum := range s.sortedLevelNums() {
		level := s.levels[levelNum]
		if level == nil {
			continue
		}
		for _, id := range sortedEntityIDs(level.entities) {
			existing := level.entities[id]
			if existing == nil || existing.kind != companionEntity || existing.ownerID != ownerID || existing.sourceSkillID != skillID {
				continue
			}
			existingRefs = append(existingRefs, companionRef{level: level, id: id})
		}
	}
	removeCount := len(existingRefs) - limit + 1
	for i := 0; i < removeCount; i++ {
		id := existingRefs[i].id
		delete(existingRefs[i].level.entities, id)
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(id)})
	}
}

func (s *Sim) transferOwnedCompanionsToLevel(ownerID uint64, current, dest *LevelState, ownerPos Vec2, res *TickResult) {
	if current == nil || dest == nil || current == dest {
		return
	}
	offsetIndex := 0
	for _, id := range sortedEntityIDs(current.entities) {
		companion := current.entities[id]
		if companion == nil || companion.kind != companionEntity || companion.ownerID != ownerID {
			continue
		}
		delete(current.entities, id)
		companion.pos = companionTravelPosition(ownerPos, offsetIndex)
		companion.spawnPos = ownerPos
		companion.targetID = 0
		dest.entities[id] = companion
		res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(id)})
		offsetIndex++
	}
}

func scalePositiveInt(value int, percent int) int {
	if value <= 0 || percent <= 0 {
		return 1
	}
	scaled := int(math.Ceil(float64(value) * float64(percent) / 100.0))
	if scaled < 1 {
		return 1
	}
	return scaled
}

func scaleRevivedDamageRange(damage *DamageRange, percent int) *DamageRange {
	if damage == nil {
		return nil
	}
	return &DamageRange{
		Min: scalePositiveInt(damage.Min, percent),
		Max: scalePositiveInt(damage.Max, percent),
	}
}

func companionSpawnPosition(owner *entity) Vec2 {
	return Vec2{X: owner.pos.X + companionFollowDistance, Y: owner.pos.Y}
}

func companionTravelPosition(ownerPos Vec2, index int) Vec2 {
	offsets := []Vec2{
		{X: companionFollowDistance, Y: 0},
		{X: -companionFollowDistance, Y: 0},
		{X: 0, Y: companionFollowDistance},
		{X: 0, Y: -companionFollowDistance},
	}
	offset := offsets[index%len(offsets)]
	ring := 1 + index/len(offsets)
	return Vec2{
		X: ownerPos.X + offset.X*float64(ring),
		Y: ownerPos.Y + offset.Y*float64(ring),
	}
}

func (e *entity) applyMonsterLikeViewFields(ev *EntityView) {
	ev.MonsterDefID = e.monsterDefID
	ev.MonsterPackID = e.monsterPackID
	ev.MonsterPackLeader = e.monsterPackLeader
	if e.kind == companionEntity {
		ev.OwnerID = idStr(e.ownerID)
		ev.CompanionStance = e.companionStanceOrDefault()
		if e.targetID != 0 {
			ev.TargetID = idStr(e.targetID)
		}
	}
	if e.monsterRarityID != "" {
		ev.Rarity = e.monsterRarityID
	}
	ev.IsBoss = e.isBoss
	ev.BossTemplateID = e.bossTemplateID
	ev.VisualModel = e.visualModel
	ev.VisualScale = e.visualScale
	ev.VisualTint = e.visualTint
	if e.isBoss && e.bossPhaseKind != "" {
		ev.BossPhase = e.bossPhaseView()
	}
	if e.isBoss {
		ev.Enraged = e.bossEnraged
		if e.bossEnrageThreshold > 0 {
			ev.EnrageHealthRatioThreshold = e.bossEnrageThreshold
		}
	}
}

func (s *Sim) advanceCompanions(res *TickResult) {
	level := s.activeLevel()
	if level == nil {
		return
	}
	for _, id := range sortedEntityIDs(level.entities) {
		companion := level.entities[id]
		if companion == nil || companion.kind != companionEntity || companion.hp <= 0 {
			continue
		}
		if companion.expiresTick > 0 && s.tick >= companion.expiresTick {
			delete(level.entities, id)
			res.Changes = append(res.Changes, Change{Op: OpEntityRemove, EntityID: idStr(id)})
			continue
		}
		owner := level.entities[companion.ownerID]
		if owner == nil || owner.kind != playerEntity || owner.hp <= 0 {
			continue
		}
		target := s.companionTarget(companion)
		if target != nil {
			companion.targetID = target.id
		} else {
			companion.targetID = 0
		}
		s.advanceCompanionMovement(companion, owner, target, res)
		if target != nil {
			s.advanceCompanionAttack(companion, target, res)
		}
	}
}

func (s *Sim) companionTarget(companion *entity) *entity {
	level := s.activeLevel()
	if companion.companionStanceOrDefault() == CompanionStancePassive {
		return nil
	}
	if companion.targetID != 0 {
		target := level.entities[companion.targetID]
		if s.validCompanionTarget(companion, target) {
			return target
		}
	}
	var best *entity
	bestDist := math.MaxFloat64
	for _, id := range sortedEntityIDs(level.entities) {
		target := level.entities[id]
		if !s.validCompanionTarget(companion, target) {
			continue
		}
		dist := distance(companion.pos, target.pos)
		if best == nil || dist < bestDist-1e-9 || (math.Abs(dist-bestDist) <= 1e-9 && target.id < best.id) {
			best = target
			bestDist = dist
		}
	}
	return best
}

func (s *Sim) validCompanionTarget(companion *entity, target *entity) bool {
	if companion == nil || target == nil || target.kind != monsterEntity || target.hp <= 0 {
		return false
	}
	switch companion.companionStanceOrDefault() {
	case CompanionStancePassive:
		return false
	case CompanionStanceDefend:
		owner := s.activeLevel().entities[companion.ownerID]
		return owner != nil && owner.kind == playerEntity && owner.hp > 0 && distance(owner.pos, target.pos) <= companionAssistRadius
	default:
		return distance(companion.pos, target.pos) <= companionAssistRadius
	}
}

func (s *Sim) advanceCompanionMovement(companion *entity, owner *entity, target *entity, res *TickResult) {
	goal, ok := s.companionMovementGoal(companion, owner, target)
	if !ok {
		return
	}
	nav := s.activeNav()
	speed := companion.speed
	if speed <= 0 {
		speed = nav.CellSize
	}
	var steps []Vec2
	if target == nil {
		blocked := s.buildCompanionBlockedFn(companion)
		var pathOK bool
		steps, pathOK = PlanPath(nav, companion.pos, goal, blocked)
		if !pathOK && distance(companion.pos, goal) > nav.CellSize+nav.StopDistance {
			return
		}
	}
	before := companion.pos
	companion.pos = s.resolveCompanionMovement(companion, s.monsterMoveDelta(companion.pos, goal, steps, speed))
	if companion.pos != before {
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(companion))})
	}
}

func (s *Sim) buildCompanionBlockedFn(companion *entity) func(gx, gy int) bool {
	return func(gx, gy int) bool {
		center := gridToWorld(s.activeNav(), gridCell{x: gx, y: gy})
		return s.companionPositionBlocked(companion, center)
	}
}

func (s *Sim) companionPositionBlocked(companion *entity, pos Vec2) bool {
	for _, wall := range s.activeWalls() {
		if circleIntersectsAABB(pos, monsterRadius, wall.pos, wall.size) {
			return true
		}
	}
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		if id == companion.id || id == companion.ownerID {
			continue
		}
		e := s.activeLevel().entities[id]
		if e == nil {
			continue
		}
		if (e.kind == monsterEntity || e.kind == companionEntity) && e.hp > 0 {
			if circlesOverlap(pos, monsterRadius, e.pos, monsterRadius) {
				return true
			}
			continue
		}
		if e.kind == interactableEntity && e.state == interactableClosed {
			if def, ok := s.rules.Interactables[e.interactableDefID]; ok && def.BarrierWhenClosed != nil {
				if circleIntersectsAABB(pos, monsterRadius, e.pos, def.BarrierWhenClosed.Size) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Sim) resolveCompanionMovement(companion *entity, delta Vec2) Vec2 {
	candidate := Vec2{X: companion.pos.X + delta.X, Y: companion.pos.Y + delta.Y}
	if !s.companionPositionBlocked(companion, candidate) {
		return candidate
	}
	xOnly := Vec2{X: companion.pos.X + delta.X, Y: companion.pos.Y}
	if delta.X != 0 && !s.companionPositionBlocked(companion, xOnly) {
		return xOnly
	}
	yOnly := Vec2{X: companion.pos.X, Y: companion.pos.Y + delta.Y}
	if delta.Y != 0 && !s.companionPositionBlocked(companion, yOnly) {
		return yOnly
	}
	return companion.pos
}

func (s *Sim) companionMovementGoal(companion *entity, owner *entity, target *entity) (Vec2, bool) {
	if target != nil {
		def := s.rules.Monsters[companion.monsterDefID]
		if s.companionInAttackRange(companion, target, def) {
			return Vec2{}, false
		}
		return s.companionAttackGoal(companion, target, def)
	}
	if distance(companion.pos, owner.pos) <= companionFollowStopRadius {
		return Vec2{}, false
	}
	return Vec2{X: owner.pos.X - companionFollowDistance, Y: owner.pos.Y}, true
}

func (s *Sim) companionAttackGoal(companion *entity, target *entity, def MonsterDef) (Vec2, bool) {
	reach := s.monsterAttackReach(def) + monsterRadius - 0.05
	minSeparation := monsterRadius + monsterRadius + 0.05
	if reach < minSeparation {
		reach = minSeparation
	}
	dir := normalize(Vec2{X: companion.pos.X - target.pos.X, Y: companion.pos.Y - target.pos.Y})
	if dir.X == 0 && dir.Y == 0 {
		dir = Vec2{X: -1}
	}
	goal := Vec2{X: target.pos.X + dir.X*reach, Y: target.pos.Y + dir.Y*reach}
	if !s.positionInNavigationBounds(s.activeNav(), goal) || s.monsterPositionBlocked(goal, companion.id) {
		return target.pos, true
	}
	return goal, true
}

func (s *Sim) advanceCompanionAttack(companion *entity, target *entity, res *TickResult) {
	def := s.rules.Monsters[companion.monsterDefID]
	if def.AttackDamage == nil && companion.monsterAttackDamage == nil {
		return
	}
	if !s.companionInAttackRange(companion, target, def) {
		return
	}
	cooldown := def.AttackCooldown
	if companion.monsterAttackCooldown > 0 {
		cooldown = companion.monsterAttackCooldown
	}
	if cooldown <= 0 {
		return
	}
	if companion.hasAttacked && s.tick-companion.lastAttackTick < uint64(cooldown) {
		return
	}
	damage := def.AttackDamage
	if companion.monsterAttackDamage != nil {
		damage = companion.monsterAttackDamage
	}
	companion.lastAttackTick = s.tick
	companion.hasAttacked = true
	s.damageMonsterByCompanion(target, companion, *damage, res)
}

func (s *Sim) companionInAttackRange(companion *entity, target *entity, def MonsterDef) bool {
	return meleeInRange(distance(target.pos, companion.pos), s.monsterAttackReach(def), monsterRadius+0.05)
}

func (s *Sim) damageMonsterByCompanion(target *entity, companion *entity, damageRange DamageRange, res *TickResult) combatResolution {
	defenderStats := s.monsterEffectiveCombatStats(target, DamageRange{})
	attackerStats := s.monsterEffectiveCombatStats(companion, damageRange)
	outcome := s.resolveCombat(attackerStats, defenderStats, damageRange)
	s.applyMonsterResistanceToOutcome(target, damageTypeForce, &outcome)
	if outcome.Hit && !outcome.Blocked {
		target.hp -= outcome.Damage
		if target.hp < 0 {
			target.hp = 0
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(target))})
	}
	res.Events = append(res.Events, combatEvent(s.combatEventType(monsterEntity, outcome), companion.id, target.id, "", outcome))
	if target.hp == 0 {
		s.finishMonsterKill(target, companion.id, "", res)
	}
	return outcome
}
