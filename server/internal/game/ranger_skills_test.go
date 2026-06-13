package game

import "testing"

func TestRangerPiercingShotDamagesLineTargets(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_pierce")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	first := addRangerSkillMonster(sim, Vec2{X: 6, Y: 2}, 40)
	second := addRangerSkillMonster(sim, Vec2{X: 9, Y: 2}, 40)
	offLine := addRangerSkillMonster(sim, Vec2{X: 7, Y: 4}, 40)

	cast := sim.Tick([]Input{{
		MessageID:     "pierce",
		CorrelationID: "corr_pierce",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "piercing_shot", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "pierce")
	if !hasEvent(cast, "skill_cast") || !hasSkillDamageEvent(cast, "piercing_shot") {
		t.Fatalf("piercing shot events = %+v", cast.Events)
	}
	if first.hp >= first.maxHP || second.hp >= second.maxHP {
		t.Fatalf("piercing shot hp first=%d/%d second=%d/%d, want both damaged", first.hp, first.maxHP, second.hp, second.maxHP)
	}
	if offLine.hp != offLine.maxHP {
		t.Fatalf("off-line monster hp = %d/%d, want undamaged", offLine.hp, offLine.maxHP)
	}
	if countSkillDamageEvents(cast, "piercing_shot") < 2 {
		t.Fatalf("piercing shot damage events = %+v, want at least two", cast.Events)
	}
}

func TestRangerPinningShotRootsMonsterMovementUntilExpiry(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_pin")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	target := addRangerSkillMonster(sim, Vec2{X: 8, Y: 2}, 40)
	target.monsterDefID = "training_dummy_chase"
	target.aiMode = monsterAIModeChase
	before := target.pos

	cast := sim.Tick([]Input{{
		MessageID:     "pin",
		CorrelationID: "corr_pin",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "pinning_shot", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "pin")
	if !hasEvent(cast, "skill_effect_started") || !containsStringValue(target.effectIDs, "pinning_root") {
		t.Fatalf("pinning shot events/effects = %+v / %v", cast.Events, target.effectIDs)
	}

	for i := 0; i < 5; i++ {
		sim.Tick(nil)
	}
	if target.pos != before {
		t.Fatalf("rooted monster moved from %+v to %+v", before, target.pos)
	}

	for i := 0; i < sim.rules.Skills["pinning_shot"].Root.DurationTicks+1; i++ {
		sim.Tick(nil)
	}
	if containsStringValue(target.effectIDs, "pinning_root") {
		t.Fatalf("pinning root still active after expiry: %v", target.effectIDs)
	}
	expiredBeforeMove := target.pos
	for i := 0; i < 10; i++ {
		sim.Tick(nil)
	}
	if target.pos == expiredBeforeMove {
		t.Fatalf("unpinned monster did not resume movement from %+v", expiredBeforeMove)
	}
}

func TestRangerSkillRulesLoad(t *testing.T) {
	rules := loadRules(t)
	pierce := rules.Skills["piercing_shot"]
	if pierce.Class != "ranger" || pierce.Pierce.MaxHits < 2 || pierce.Projectile.Visual != "piercing_shot_projectile" {
		t.Fatalf("piercing_shot = %+v, want ranger projectile with pierce", pierce)
	}
	pin := rules.Skills["pinning_shot"]
	if pin.Class != "ranger" || pin.Root.EffectID != "pinning_root" || pin.Root.DurationTicks <= 0 || pin.Projectile.Visual != "pinning_shot_projectile" {
		t.Fatalf("pinning_shot = %+v, want ranger projectile with root", pin)
	}
}

func rangerSkillSim(t *testing.T, sessionID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim := MustNewSim(sessionID, sessionID+"_seed", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["ranger"].BaseStats
	sim.progression.SkillRanks["piercing_shot"] = 1
	sim.progression.SkillRanks["pinning_shot"] = 1
	ps := sim.defaultPlayer()
	ps.Progression = sim.progression
	player := sim.activeLevel().entities[sim.playerID]
	player.maxMana = 50
	player.mana = 50
	return sim
}

func addRangerSkillMonster(sim *Sim, pos Vec2, hp int) *entity {
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          pos,
		hp:           hp,
		maxHP:        hp,
		monsterDefID: monsterDefID,
		lootTable:    "no_drop",
	}
	sim.activeLevel().entities[monster.id] = monster
	return monster
}

func countSkillDamageEvents(r TickResult, skillID string) int {
	count := 0
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.SkillID == skillID {
			count++
		}
	}
	return count
}
