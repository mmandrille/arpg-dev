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
	if !hasEvent(cast, "skill_cast") || !hasEntitySpawn(cast, idStr(companion.id)) {
		t.Fatalf("revive changes/events = %+v / %+v", cast.Changes, cast.Events)
	}
	view := sim.entityView(companion)
	if view.Type != companionEntity || view.MonsterDefID != "dungeon_wolf" || view.OwnerID != idStr(player.id) {
		t.Fatalf("revived view = %+v, want owned dungeon_wolf companion", view)
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
	if revive.Class != "sorcerer" || revive.Kind != "revive_companion" || revive.Revive.PowerPercentBase != 50 || revive.Revive.PowerPercentPerRank != 10 || revive.Revive.Limit != 1 {
		t.Fatalf("revive = %+v, want sorcerer revive companion scaling", revive)
	}
}

func sorcererReviveSim(t *testing.T, sessionID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim := MustNewSim(sessionID, sessionID+"_seed", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim.progression.BaseStats.Magic = 12
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
	var found *entity
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		entity := sim.activeLevel().entities[id]
		if entity == nil || entity.kind != companionEntity || entity.monsterDefID != monsterDefID {
			continue
		}
		if found != nil {
			t.Fatalf("multiple revived companions: %d and %d", found.id, entity.id)
		}
		found = entity
	}
	if found == nil {
		t.Fatalf("missing revived %s companion", monsterDefID)
	}
	return found
}
