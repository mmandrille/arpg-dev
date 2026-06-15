package game

import "testing"

func TestSorcererReviveCreatesScaledCompanionFromDeadMonster(t *testing.T) {
	sim := sorcererReviveSim(t, "sess_sorcerer_revive")
	player := sim.activeLevel().entities[sim.playerID]
	target := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 0)
	originalID := target.id

	cast := sim.Tick([]Input{{
		MessageID:     "revive",
		CorrelationID: "corr_revive",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "revive")
	if sim.findEntity(idStr(originalID)) != nil || !hasEntityRemove(cast, idStr(originalID)) {
		t.Fatalf("dead monster was not consumed: entity=%+v changes=%+v", sim.findEntity(idStr(originalID)), cast.Changes)
	}
	companion := onlyRevivedCompanion(t, sim, "dungeon_wolf")
	if companion.ownerID != player.id || companion.sourceSkillID != "revive" {
		t.Fatalf("revived owner/source = %d/%s, want %d/revive", companion.ownerID, companion.sourceSkillID, player.id)
	}
	def := sim.rules.Monsters["dungeon_wolf"]
	if companion.maxHP != scalePositiveInt(def.MaxHP, 50) || companion.hp != companion.maxHP {
		t.Fatalf("revived hp = %d/%d, want 50%% of %d", companion.hp, companion.maxHP, def.MaxHP)
	}
	if companion.monsterAttackDamage == nil || companion.monsterAttackDamage.Min != 1 || companion.monsterAttackDamage.Max != 1 {
		t.Fatalf("revived damage = %+v, want 50%% scaled wolf damage", companion.monsterAttackDamage)
	}
	if companion.totalDurationTicks != 600 || companion.expiresTick != sim.tick+600 {
		t.Fatalf("revived duration ticks total/expires = %d/%d at tick %d, want 600/%d", companion.totalDurationTicks, companion.expiresTick, sim.tick, sim.tick+600)
	}
	if !hasEvent(cast, "skill_cast") || !hasEntitySpawn(cast, idStr(companion.id)) {
		t.Fatalf("revive changes/events = %+v / %+v", cast.Changes, cast.Events)
	}
	view := sim.entityView(companion)
	if view.Type != companionEntity || view.MonsterDefID != "dungeon_wolf" || view.OwnerID != idStr(player.id) {
		t.Fatalf("revived view = %+v, want owned dungeon_wolf companion", view)
	}
	if view.RemainingTicks == nil || *view.RemainingTicks != 600 || view.TotalTicks == nil || *view.TotalTicks != 600 {
		t.Fatalf("revived timer view = remaining %v total %v, want 600/600", view.RemainingTicks, view.TotalTicks)
	}
}

func TestSorcererReviveRankScalingAllowsMultipleCompanions(t *testing.T) {
	sim := sorcererReviveSim(t, "sess_sorcerer_revive_rank4")
	sim.progression.SkillRanks["revive"] = 4
	player := sim.activeLevel().entities[sim.playerID]
	first := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 0)
	second := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 0)

	firstCast := sim.Tick([]Input{{
		MessageID:     "revive_rank4_1",
		CorrelationID: "corr_revive_rank4_1",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(first.id)},
	}})
	assertAck(t, firstCast, "revive_rank4_1")
	delete(sim.skillCooldowns, "revive")
	secondCast := sim.Tick([]Input{{
		MessageID:     "revive_rank4_2",
		CorrelationID: "corr_revive_rank4_2",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(second.id)},
	}})
	assertAck(t, secondCast, "revive_rank4_2")

	companions := revivedCompanions(sim, "dungeon_wolf")
	if len(companions) != 2 {
		t.Fatalf("rank 4 revived companions = %d, want 2: %+v", len(companions), companions)
	}
	def := sim.rules.Monsters["dungeon_wolf"]
	wantHP := scalePositiveInt(def.MaxHP, 80)
	for _, companion := range companions {
		if companion.maxHP != wantHP || companion.monsterAttackDamage == nil || companion.monsterAttackDamage.Max != 2 {
			t.Fatalf("rank 4 companion stats hp=%d damage=%+v, want scaled hp=%d and damage max=2", companion.maxHP, companion.monsterAttackDamage, wantHP)
		}
		if companion.totalDurationTicks != 900 {
			t.Fatalf("rank 4 companion duration = %d, want 900", companion.totalDurationTicks)
		}
	}
}

func TestSorcererReviveCompanionExpiresAfterDuration(t *testing.T) {
	sim := sorcererReviveSim(t, "sess_sorcerer_revive_expire")
	player := sim.activeLevel().entities[sim.playerID]
	target := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 0)

	cast := sim.Tick([]Input{{
		MessageID:     "revive_expire",
		CorrelationID: "corr_revive_expire",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "revive_expire")
	companion := onlyRevivedCompanion(t, sim, "dungeon_wolf")
	companionID := idStr(companion.id)

	var last TickResult
	for i := 0; i < 601; i++ {
		last = sim.Tick(nil)
	}
	if sim.findEntity(companionID) != nil || !hasEntityRemove(last, companionID) {
		t.Fatalf("revived companion did not expire: entity=%+v last changes=%+v", sim.findEntity(companionID), last.Changes)
	}
}

func TestSorcererReviveRejectsBossAndLivingTargets(t *testing.T) {
	sim := sorcererReviveSim(t, "sess_sorcerer_revive_reject")
	player := sim.activeLevel().entities[sim.playerID]
	living := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 4)
	boss := addReviveTestMonster(sim, "dungeon_wolf", Vec2{X: player.pos.X + 3, Y: player.pos.Y}, 0)
	boss.isBoss = true

	livingCast := sim.Tick([]Input{{
		MessageID:     "revive_living",
		CorrelationID: "corr_revive_living",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(living.id)},
	}})
	assertReject(t, livingCast, "revive_living", "target_not_dead")

	bossCast := sim.Tick([]Input{{
		MessageID:     "revive_boss",
		CorrelationID: "corr_revive_boss",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "revive", TargetID: idStr(boss.id)},
	}})
	assertReject(t, bossCast, "revive_boss", "cannot_revive_boss")
}

func TestSorcererReviveRulesLoad(t *testing.T) {
	rules := loadRules(t)
	revive := rules.Skills["revive"]
	if revive.Class != "sorcerer" || revive.Kind != "revive_companion" || revive.Revive.PowerPercentBase != 50 || revive.Revive.PowerPercentPerRank != 10 || reviveDurationTicks(revive, 1) != 600 || reviveDurationTicks(revive, 4) != 900 || companionLimitAtRank(revive.Revive.Limit, 4) != 2 {
		t.Fatalf("revive = %+v, want sorcerer revive companion scaling", revive)
	}
}

func sorcererReviveSim(t *testing.T, sessionID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim := MustNewSim(sessionID, sessionID+"_seed", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.Level = 7
	sim.progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim.progression.BaseStats.Magic = 18
	sim.progression.SkillRanks["magic_bolt"] = 1
	sim.progression.SkillRanks["revive"] = 1
	ps := sim.defaultPlayer()
	ps.Progression = sim.progression
	player := sim.activeLevel().entities[sim.playerID]
	player.maxMana = 50
	player.mana = 50
	return sim
}

func addReviveTestMonster(sim *Sim, monsterDefID string, pos Vec2, hp int) *entity {
	def := sim.rules.Monsters[monsterDefID]
	monster := &entity{
		id:                    sim.alloc(),
		kind:                  monsterEntity,
		pos:                   pos,
		spawnPos:              pos,
		hp:                    hp,
		maxHP:                 def.MaxHP,
		monsterDefID:          monsterDefID,
		lootTable:             "no_drop",
		speed:                 def.MoveSpeed,
		monsterAttackDamage:   def.AttackDamage,
		monsterAttackCooldown: def.AttackCooldown,
		aiMode:                monsterAIModeIdle,
	}
	sim.activeLevel().entities[monster.id] = monster
	return monster
}

func onlyRevivedCompanion(t *testing.T, sim *Sim, monsterDefID string) *entity {
	t.Helper()
	found := revivedCompanions(sim, monsterDefID)
	if len(found) != 1 {
		t.Fatalf("revived %s companions = %d, want 1", monsterDefID, len(found))
	}
	return found[0]
}

func revivedCompanions(sim *Sim, monsterDefID string) []*entity {
	out := []*entity{}
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		entity := sim.activeLevel().entities[id]
		if entity == nil || entity.kind != companionEntity || entity.monsterDefID != monsterDefID {
			continue
		}
		out = append(out, entity)
	}
	return out
}
