package game

import (
	"math"
	"testing"
)

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

func TestRangerVolleyDamagesFanTargetsOnce(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_volley")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 2}
	center := addRangerSkillMonster(sim, Vec2{X: 8, Y: 2}, 40)
	upper := addRangerSkillMonster(sim, Vec2{X: 8, Y: 4}, 40)
	lower := addRangerSkillMonster(sim, Vec2{X: 8, Y: 0}, 40)
	behind := addRangerSkillMonster(sim, Vec2{X: 0, Y: 2}, 40)

	cast := sim.Tick([]Input{{
		MessageID:     "volley",
		CorrelationID: "corr_volley",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "volley", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "volley")
	if countSkillDamageEvents(cast, "volley") < 3 {
		t.Fatalf("volley damage events = %+v, want at least three", cast.Events)
	}
	for _, monster := range []*entity{center, upper, lower} {
		if monster.hp >= monster.maxHP {
			t.Fatalf("volley target %d hp=%d/%d, want damaged", monster.id, monster.hp, monster.maxHP)
		}
	}
	if behind.hp != behind.maxHP {
		t.Fatalf("behind monster hp=%d/%d, want undamaged", behind.hp, behind.maxHP)
	}
}

func TestRangerBlackWolfCompanionSummonsAndReplaces(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_black_wolf")
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = Vec2{X: 4, Y: 4}

	firstCast := sim.Tick([]Input{{
		MessageID:     "wolf_1",
		CorrelationID: "corr_wolf_1",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "black_wolf_companion"},
	}})
	assertAck(t, firstCast, "wolf_1")
	firstWolf := onlyRangerWolfCompanion(t, sim)
	if firstWolf.ownerID != player.id || firstWolf.sourceSkillID != "black_wolf_companion" {
		t.Fatalf("wolf owner/source = %d/%s, want %d/black_wolf_companion", firstWolf.ownerID, firstWolf.sourceSkillID, player.id)
	}
	view := sim.entityView(firstWolf)
	if view.Type != companionEntity || view.MonsterDefID != "companion_black_wolf" || view.VisualModel != "monster_wolf" || view.VisualTint != "101014" {
		t.Fatalf("wolf view = %+v, want black wolf companion", view)
	}
	percent := companionHeroStatPercent(sim.rules.Skills["black_wolf_companion"], sim.effectiveSkillRank("black_wolf_companion"))
	if firstWolf.maxHP != scalePositiveInt(player.maxHP, percent) || firstWolf.monsterAttackDamage == nil {
		t.Fatalf("wolf stats hp=%d damage=%+v percent=%d player=%+v", firstWolf.maxHP, firstWolf.monsterAttackDamage, percent, player)
	}
	if math.Abs(view.VisualScale-sim.rules.Skills["black_wolf_companion"].Companion.VisualScale) > 1e-9 {
		t.Fatalf("wolf visual_scale = %.2f, want configured %.2f", view.VisualScale, sim.rules.Skills["black_wolf_companion"].Companion.VisualScale)
	}
	if !hasEvent(firstCast, "skill_cast") || !hasEntitySpawn(firstCast, idStr(firstWolf.id)) {
		t.Fatalf("first wolf cast changes/events = %+v / %+v", firstCast.Changes, firstCast.Events)
	}

	delete(sim.skillCooldowns, "black_wolf_companion")
	secondCast := sim.Tick([]Input{{
		MessageID:     "wolf_2",
		CorrelationID: "corr_wolf_2",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "black_wolf_companion"},
	}})
	assertAck(t, secondCast, "wolf_2")
	secondWolf := onlyRangerWolfCompanion(t, sim)
	if secondWolf.id == firstWolf.id {
		t.Fatalf("recast kept same wolf id %d, want replacement", secondWolf.id)
	}
	if !hasEntityRemove(secondCast, idStr(firstWolf.id)) || !hasEntitySpawn(secondCast, idStr(secondWolf.id)) {
		t.Fatalf("recast changes = %+v, want remove old and spawn new", secondCast.Changes)
	}
}

func TestRangerBlackWolfCompanionFollowsAcrossLevelTravel(t *testing.T) {
	sim := rangerDungeonSkillSim(t, "sess_ranger_wolf_travel")
	player := sim.activeLevel().entities[sim.playerID]
	cast := sim.Tick([]Input{{
		MessageID: "wolf",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: "black_wolf_companion"},
	}})
	assertAck(t, cast, "wolf")
	wolf := onlyRangerWolfCompanion(t, sim)
	wolfID := idStr(wolf.id)

	results := descendFromCurrentLevel(t, sim, "descend_with_wolf")
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want source and destination: %+v", len(results), results)
	}
	if !hasEntityRemove(results[0], wolfID) || !hasEntitySpawn(results[1], wolfID) {
		t.Fatalf("travel changes missing wolf transfer remove/spawn: from=%+v to=%+v", results[0].Changes, results[1].Changes)
	}
	movedWolf := onlyRangerWolfCompanion(t, sim)
	if movedWolf.id != wolf.id || movedWolf.ownerID != player.id || movedWolf.sourceSkillID != "black_wolf_companion" {
		t.Fatalf("moved wolf = %+v, want same owner/source/id %s", movedWolf, wolfID)
	}
	if distance(movedWolf.pos, sim.activeLevel().entities[sim.playerID].pos) > sim.rules.MainConfig.Gameplay.CompanionFollowStop {
		t.Fatalf("moved wolf pos=%+v too far from player=%+v", movedWolf.pos, sim.activeLevel().entities[sim.playerID].pos)
	}
}

func TestRangerSkillRulesLoad(t *testing.T) {
	rules := loadRules(t)
	pierce := rules.Skills["piercing_shot"]
	if pierce.Class != "ranger" || pierce.Tree.Tier != 1 || pierce.Pierce.MaxHits < 2 || pierce.Projectile.Visual != "piercing_shot_projectile" {
		t.Fatalf("piercing_shot = %+v, want ranger projectile with pierce", pierce)
	}
	pin := rules.Skills["pinning_shot"]
	if pin.Class != "ranger" || pin.Tree.Tier != 2 || len(pin.Requirements.Skills) != 1 || pin.Requirements.Skills[0].SkillID != "piercing_shot" || pin.Root.EffectID != "pinning_root" || pin.Root.DurationTicks <= 0 || pin.Projectile.Visual != "pinning_shot_projectile" {
		t.Fatalf("pinning_shot = %+v, want tier 2 ranger projectile with piercing prereq and root", pin)
	}
	split := rules.Skills["split_arrow"]
	if split.Class != "ranger" || split.Tree.Tier != 3 || len(split.Requirements.Skills) != 1 || split.Requirements.Skills[0].SkillID != "volley" {
		t.Fatalf("split_arrow = %+v, want tier 3 with volley prereq", split)
	}
	volley := rules.Skills["volley"]
	if volley.Class != "ranger" || volley.Tree.Tier != 2 || len(volley.Requirements.Skills) != 1 || volley.Requirements.Skills[0].SkillID != "piercing_shot" || volley.Volley.ArrowCount < 3 || volley.Volley.SpreadDegrees <= 0 || volley.Projectile.Visual != "volley_arrow_projectile" {
		t.Fatalf("volley = %+v, want tier 2 ranger projectile with piercing prereq and fan", volley)
	}
	wolf := rules.Skills["black_wolf_companion"]
	if wolf.Class != "ranger" || wolf.Tree.Tier != 1 || wolf.Requirements.Stats["magic"] != 8 || wolf.Kind != "summon_companion" || wolf.Companion.MonsterDefID != "companion_black_wolf" || companionHeroStatPercent(wolf, 1) != 70 || companionHeroStatPercent(wolf, 2) != 85 || companionLimitAtRank(wolf.Companion.Limit, 5) != 1 {
		t.Fatalf("black_wolf_companion = %+v, want tier 1 magic-scaled black wolf summon", wolf)
	}
	if wolf.Cooldown.FixedTicks != 1200 || wolf.Cooldown.MagicReductionTicksPerPoint <= 0 {
		t.Fatalf("black_wolf_companion cooldown = %+v, want 120s base with magic reduction", wolf.Cooldown)
	}
}

func TestRangerBlackWolfCooldownReducedByExtraMagic(t *testing.T) {
	sim := rangerSkillSim(t, "sess_ranger_wolf_cooldown")
	def := sim.rules.Skills["black_wolf_companion"]
	sim.progression.BaseStats.Magic = skillStatRequirementForRank(def, "magic", 1)
	if got := sim.skillCooldownTicks(def); got != 1200 {
		t.Fatalf("wolf baseline cooldown = %d, want 1200", got)
	}
	sim.progression.BaseStats.Magic += 5
	if got := sim.skillCooldownTicks(def); got != 1150 {
		t.Fatalf("wolf magic-reduced cooldown = %d, want 1150", got)
	}
}

func rangerSkillSim(t *testing.T, sessionID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim := MustNewSim(sessionID, sessionID+"_seed", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["ranger"].BaseStats
	sim.progression.BaseStats.Dex = 14
	sim.progression.BaseStats.Magic = 12
	sim.progression.SkillRanks["piercing_shot"] = 1
	sim.progression.SkillRanks["pinning_shot"] = 1
	sim.progression.SkillRanks["volley"] = 1
	sim.progression.SkillRanks["split_arrow"] = 1
	sim.progression.SkillRanks["black_wolf_companion"] = 1
	ps := sim.defaultPlayer()
	ps.Progression = sim.progression
	player := sim.activeLevel().entities[sim.playerID]
	player.maxMana = 50
	player.mana = 50
	return sim
}

func rangerDungeonSkillSim(t *testing.T, sessionID string) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorld(sessionID, sessionID+"_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon ranger sim: %v", err)
	}
	sim.progression.CharacterClass = "ranger"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["ranger"].BaseStats
	sim.progression.BaseStats.Dex = 14
	sim.progression.BaseStats.Magic = 12
	sim.progression.SkillRanks["black_wolf_companion"] = 1
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

func onlyRangerWolfCompanion(t *testing.T, sim *Sim) *entity {
	t.Helper()
	var found *entity
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		entity := sim.activeLevel().entities[id]
		if entity == nil || entity.kind != companionEntity || entity.monsterDefID != "companion_black_wolf" {
			continue
		}
		if found != nil {
			t.Fatalf("multiple black wolf companions: %d and %d", found.id, entity.id)
		}
		found = entity
	}
	if found == nil {
		t.Fatalf("missing black wolf companion")
	}
	return found
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
