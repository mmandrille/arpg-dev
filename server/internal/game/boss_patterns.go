package game

import "math"

func (s *Sim) advanceBossPhases(res *TickResult) {
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		boss := s.activeLevel().entities[id]
		if boss == nil || boss.kind != monsterEntity || !boss.isBoss || boss.hp <= 0 {
			continue
		}
		runtime, ok := s.ensureBossPhase(boss, res)
		if !ok {
			continue
		}
		if boss.bossPhaseKind == "active" {
			s.applyBossActivePhase(boss, runtime.phase, res)
		}
		if s.tick+1 >= boss.bossPhaseEnds {
			s.endBossPhase(boss, runtime, res)
		}
	}
}

func (s *Sim) ensureBossPhase(boss *entity, res *TickResult) (bossPhaseRuntime, bool) {
	if boss.bossPhaseKind != "" && s.tick < boss.bossPhaseEnds {
		return s.currentBossPhase(boss)
	}
	if boss.bossCooldownEnds > s.tick {
		return bossPhaseRuntime{}, false
	}
	next, ok := s.nextBossPhase(boss)
	if !ok {
		return bossPhaseRuntime{}, false
	}
	boss.bossPatternID = next.patternID
	boss.bossPhaseIndex = next.index
	boss.bossPhaseKind = next.phase.Kind
	boss.bossPhaseStarted = s.tick
	boss.bossPhaseEnds = s.tick + uint64(next.phase.DurationTicks)
	boss.bossActiveHit = map[uint64]bool{}
	boss.bossPhaseExecuted = false
	s.captureBossPhaseAim(boss, next.phase)
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(boss))})
	res.Events = append(res.Events, bossPhaseEvent("boss_phase_started", boss, next))
	return next, true
}

func (s *Sim) currentBossPhase(boss *entity) (bossPhaseRuntime, bool) {
	pattern, ok := s.rules.BossPatterns[boss.bossPatternID]
	if !ok || boss.bossPhaseIndex < 0 || boss.bossPhaseIndex >= len(pattern.Phases) {
		return bossPhaseRuntime{}, false
	}
	return bossPhaseRuntime{
		patternID: boss.bossPatternID,
		index:     boss.bossPhaseIndex,
		phase:     pattern.Phases[boss.bossPhaseIndex],
	}, true
}

func (s *Sim) nextBossPhase(boss *entity) (bossPhaseRuntime, bool) {
	template, ok := s.rules.BossTemplates[boss.bossTemplateID]
	if !ok || len(template.PatternDeck) == 0 {
		return bossPhaseRuntime{}, false
	}
	patternID := boss.bossPatternID
	if patternID == "" {
		boss.bossPatternDeckIndex = 0
		patternID = template.PatternDeck[0]
		boss.bossPatternID = patternID
	}
	pattern, ok := s.rules.BossPatterns[patternID]
	if !ok || len(pattern.Phases) == 0 {
		return bossPhaseRuntime{}, false
	}
	nextIndex := boss.bossPhaseIndex + 1
	if boss.bossPhaseKind == "" {
		nextIndex = 0
	}
	if nextIndex >= len(pattern.Phases) {
		nextIndex = 0
	}
	return bossPhaseRuntime{patternID: patternID, index: nextIndex, phase: pattern.Phases[nextIndex]}, true
}

func (s *Sim) endBossPhase(boss *entity, runtime bossPhaseRuntime, res *TickResult) {
	res.Events = append(res.Events, bossPhaseEvent("boss_phase_ended", boss, runtime))
	pattern := s.rules.BossPatterns[runtime.patternID]
	if runtime.index >= len(pattern.Phases)-1 {
		boss.bossCooldownEnds = s.tick + 1 + uint64(pattern.CooldownTicks)
		boss.bossPhaseKind = ""
		boss.bossPhaseIndex = -1
		boss.bossPhaseStarted = 0
		boss.bossPhaseEnds = 0
		boss.bossActiveHit = nil
		boss.bossPhaseExecuted = false
		boss.bossPhaseHasAim = false
		s.advanceBossPatternDeck(boss)
	} else {
		next := bossPhaseRuntime{patternID: runtime.patternID, index: runtime.index + 1, phase: pattern.Phases[runtime.index+1]}
		boss.bossPhaseIndex = next.index
		boss.bossPhaseKind = next.phase.Kind
		boss.bossPhaseStarted = s.tick + 1
		boss.bossPhaseEnds = s.tick + 1 + uint64(next.phase.DurationTicks)
		boss.bossActiveHit = map[uint64]bool{}
		boss.bossPhaseExecuted = false
		if next.phase.Kind == "telegraph" {
			s.captureBossPhaseAim(boss, next.phase)
		}
		res.Events = append(res.Events, bossPhaseEvent("boss_phase_started", boss, next))
	}
	res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(boss))})
}

func (s *Sim) advanceBossPatternDeck(boss *entity) {
	template, ok := s.rules.BossTemplates[boss.bossTemplateID]
	if !ok || len(template.PatternDeck) == 0 {
		return
	}
	nextIndex := boss.bossPatternDeckIndex + 1
	if nextIndex < 0 || nextIndex >= len(template.PatternDeck) {
		nextIndex = 0
	}
	boss.bossPatternDeckIndex = nextIndex
	boss.bossPatternID = template.PatternDeck[nextIndex]
}

func (s *Sim) applyBossActivePhase(boss *entity, phase BossPatternPhase, res *TickResult) {
	if phase.SummonMonsterDefID != "" {
		s.applyBossSummonPhase(boss, phase, res)
	}
	if phase.Damage == nil {
		return
	}
	for _, playerID := range sortedPlayerIDs(s.players) {
		ps := s.players[playerID]
		if ps == nil || !ps.Connected || ps.CurrentLevel != s.currentLevel || boss.bossActiveHit[playerID] {
			continue
		}
		player := s.activeLevel().entities[playerID]
		if player == nil || player.hp <= 0 || !bossPhaseHitsPlayer(boss, player, phase) {
			continue
		}
		s.usePlayer(ps)
		scaledDamage := s.scaleMonsterDamageForParty(s.currentLevel, *phase.Damage)
		if outcome, immune := s.playerDamageImmunityOutcome(player); immune {
			boss.bossActiveHit[playerID] = true
			res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), boss.id, player.id, "", outcome))
			continue
		}
		attackerStats := s.monsterEffectiveCombatStats(boss, scaledDamage)
		defenderStats, _ := s.playerEffectiveCombatStats()
		outcome := s.resolveCombat(attackerStats, defenderStats, scaledDamage)
		boss.bossActiveHit[playerID] = true
		if !outcome.Hit || outcome.Blocked {
			s.triggerUniqueEffectsAfterPlayerAvoidedHit(player, boss, "", res)
			res.Events = append(res.Events, combatEvent(s.combatEventType(playerEntity, outcome), boss.id, player.id, "", outcome))
			continue
		}
		outcome = s.applyUniqueEffectsBeforePlayerDamage(player, boss, "", res, outcome, uniqueIncomingDamageSource{})
		player.hp -= outcome.Damage
		if player.hp < 0 {
			player.hp = 0
		}
		res.Changes = append(res.Changes, Change{Op: OpEntityUpdate, Entity: ptrEntityView(s.entityView(player))})
		eventType := "player_damaged"
		if player.hp == 0 {
			eventType = "player_killed"
		}
		res.Events = append(res.Events, combatEvent(eventType, boss.id, player.id, "", outcome))
		s.triggerUniqueEffectsAfterPlayerDamage(player, boss, "", res, outcome)
	}
}

func (s *Sim) applyBossSummonPhase(boss *entity, phase BossPatternPhase, res *TickResult) {
	if boss.bossPhaseExecuted || phase.SummonCount <= 0 || phase.SummonMonsterDefID == "" {
		return
	}
	boss.bossPhaseExecuted = true
	spawned := 0
	for _, pos := range s.bossSummonPositions(boss, phase) {
		add := s.newBossSummonedAdd(phase.SummonMonsterDefID, pos)
		if add == nil {
			continue
		}
		s.activeLevel().entities[add.id] = add
		res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(add))})
		spawned++
		if spawned >= phase.SummonCount {
			break
		}
	}
	if spawned == 0 {
		return
	}
	res.Events = append(res.Events, Event{
		EventType:     "boss_summoned_adds",
		EntityID:      idStr(boss.id),
		MonsterDefID:  phase.SummonMonsterDefID,
		PatternID:     boss.bossPatternID,
		PhaseIndex:    intPtr(boss.bossPhaseIndex),
		PhaseKind:     boss.bossPhaseKind,
		Amount:        intPtr(spawned),
		Position:      &boss.pos,
		DurationTicks: intPtr(phase.DurationTicks),
	})
}

func (s *Sim) bossSummonPositions(boss *entity, phase BossPatternPhase) []Vec2 {
	radius := phase.SummonRadius
	if radius <= 0 {
		return nil
	}
	directions := []Vec2{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
		{X: 0.70710678118, Y: 0.70710678118},
		{X: -0.70710678118, Y: 0.70710678118},
		{X: 0.70710678118, Y: -0.70710678118},
		{X: -0.70710678118, Y: -0.70710678118},
	}
	radii := []float64{radius, radius * 0.7, radius * 0.45}
	out := make([]Vec2, 0, phase.SummonCount)
	for _, candidateRadius := range radii {
		for _, dir := range directions {
			if len(out) >= phase.SummonCount {
				return out
			}
			pos := Vec2{X: boss.pos.X + dir.X*candidateRadius, Y: boss.pos.Y + dir.Y*candidateRadius}
			if s.monsterPositionBlocked(pos, 0) {
				continue
			}
			out = append(out, pos)
		}
	}
	return out
}

func (s *Sim) newBossSummonedAdd(monsterDefID string, pos Vec2) *entity {
	def, ok := s.rules.Monsters[monsterDefID]
	if !ok {
		return nil
	}
	add := &entity{
		id:              s.alloc(),
		kind:            monsterEntity,
		pos:             pos,
		spawnPos:        pos,
		hp:              def.MaxHP,
		maxHP:           def.MaxHP,
		monsterDefID:    monsterDefID,
		monsterRarityID: "common",
		lootTable:       "no_drop",
		aiMode:          monsterAIModeIdle,
	}
	if rarity, ok := s.rules.DungeonGeneration.MonsterRarity("common"); ok {
		stats := s.generatedMonsterStats(def, s.currentLevel, rarity)
		add.maxHP = stats.maxHP
		add.hp = add.maxHP
		add.visualScale = rarity.VisualScale
		add.visualTint = rarity.Color
		add.monsterAttackDamage = stats.attackDamage
		add.monsterAttackCooldown = stats.attackCooldown
		add.monsterArmor = stats.armor
		add.monsterHitChance = stats.hitChance
		add.monsterCritChance = stats.critChance
		add.monsterBlockPercent = stats.blockPercent
		add.monsterXPReward = stats.xpReward
	}
	s.applyPartyHPScale(s.activeLevel(), add)
	return add
}

func bossPhaseHitsPlayer(boss, player *entity, phase BossPatternPhase) bool {
	switch phase.Shape {
	case "melee_contact":
		radius := phase.Radius
		if radius <= 0 {
			radius = monsterRadius + playerRadius
		}
		return distance(boss.pos, player.pos) <= radius
	case "circle":
		if phase.Radius <= 0 {
			return false
		}
		return distance(boss.pos, player.pos) <= phase.Radius
	case "line":
		if phase.Radius <= 0 || phase.Width <= 0 || !boss.bossPhaseHasAim {
			return false
		}
		delta := Vec2{X: player.pos.X - boss.pos.X, Y: player.pos.Y - boss.pos.Y}
		projection := delta.X*boss.bossPhaseAim.X + delta.Y*boss.bossPhaseAim.Y
		if projection < 0 || projection > phase.Radius+playerRadius {
			return false
		}
		closest := Vec2{X: boss.pos.X + boss.bossPhaseAim.X*projection, Y: boss.pos.Y + boss.bossPhaseAim.Y*projection}
		return distance(player.pos, closest) <= phase.Width/2+playerRadius
	case "cone":
		if phase.Radius <= 0 || phase.Width <= 0 || !boss.bossPhaseHasAim {
			return false
		}
		delta := Vec2{X: player.pos.X - boss.pos.X, Y: player.pos.Y - boss.pos.Y}
		dist := distance(Vec2{}, delta)
		if dist <= meleeRangeEpsilon {
			return true
		}
		if dist > phase.Radius+playerRadius {
			return false
		}
		dir := Vec2{X: delta.X / dist, Y: delta.Y / dist}
		dot := boss.bossPhaseAim.X*dir.X + boss.bossPhaseAim.Y*dir.Y
		if dot > 1 {
			dot = 1
		}
		if dot < -1 {
			dot = -1
		}
		halfAngle := (phase.Width / 2) * math.Pi / 180
		playerAngleAllowance := math.Asin(math.Min(1, playerRadius/math.Max(dist, playerRadius)))
		return math.Acos(dot) <= halfAngle+playerAngleAllowance
	default:
		return false
	}
}

func (s *Sim) captureBossPhaseAim(boss *entity, phase BossPatternPhase) {
	if phase.HitShape != "line" && phase.Shape != "line" && phase.HitShape != "cone" && phase.Shape != "cone" {
		boss.bossPhaseHasAim = false
		return
	}
	targetState := s.nearestLivingPlayerForMonster(s.activeLevel(), boss)
	if targetState == nil {
		boss.bossPhaseAim = Vec2{X: 1}
		boss.bossPhaseHasAim = true
		return
	}
	target := s.activeLevel().entities[targetState.PlayerID]
	if target == nil {
		boss.bossPhaseAim = Vec2{X: 1}
		boss.bossPhaseHasAim = true
		return
	}
	dir := Vec2{X: target.pos.X - boss.pos.X, Y: target.pos.Y - boss.pos.Y}
	length := distance(Vec2{}, dir)
	if length <= meleeRangeEpsilon {
		boss.bossPhaseAim = Vec2{X: 1}
	} else {
		boss.bossPhaseAim = Vec2{X: dir.X / length, Y: dir.Y / length}
	}
	boss.bossPhaseHasAim = true
}

func bossPhaseEvent(eventType string, boss *entity, runtime bossPhaseRuntime) Event {
	return Event{
		EventType:     eventType,
		EntityID:      idStr(boss.id),
		PatternID:     runtime.patternID,
		PhaseIndex:    intPtr(runtime.index),
		PhaseKind:     runtime.phase.Kind,
		DurationTicks: intPtr(runtime.phase.DurationTicks),
		Telegraph:     bossTelegraphView(runtime.phase),
		HitShape:      bossHitShapeView(runtime.phase),
	}
}

func bossTelegraphView(phase BossPatternPhase) *BossTelegraphView {
	if phase.TelegraphType == "" {
		return nil
	}
	return &BossTelegraphView{
		Type:      phase.TelegraphType,
		FromColor: phase.FromColor,
		ToColor:   phase.ToColor,
		HitShape:  phase.HitShape,
		Radius:    phase.Radius,
		Width:     phase.Width,
	}
}

func bossHitShapeView(phase BossPatternPhase) *BossHitShapeView {
	shape := phase.Shape
	if shape == "" {
		shape = phase.HitShape
	}
	if shape == "" {
		return nil
	}
	return &BossHitShapeView{Shape: shape, Radius: phase.Radius, Width: phase.Width}
}
