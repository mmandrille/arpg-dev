package game

import "testing"

func TestClassSkillEngineSkillsRegistered(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		skillID string
		classID string
		prereq  string
	}{
		{skillID: "rend", classID: "barbarian", prereq: "ground_slam"},
		{skillID: "retribution", classID: "paladin", prereq: "hammer_of_light"},
		{skillID: "predators_mark", classID: "rogue", prereq: "eviscerate"},
	}
	for _, c := range cases {
		def, ok := rules.Skills[c.skillID]
		if !ok {
			t.Fatalf("missing skill %s", c.skillID)
		}
		if def.Class != c.classID {
			t.Fatalf("%s class = %s, want %s", c.skillID, def.Class, c.classID)
		}
		if len(def.Requirements.Skills) != 1 || def.Requirements.Skills[0].SkillID != c.prereq {
			t.Fatalf("%s prereq = %+v, want %s", c.skillID, def.Requirements.Skills, c.prereq)
		}
		if c.skillID == "predators_mark" && def.Mark.DurationTicks <= 0 {
			t.Fatalf("predators_mark missing mark payload: %+v", def.Mark)
		}
	}
}

func TestRendAppliesBleed(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.SkillRanks = map[string]int{"rend": 1, "ground_slam": 1, "cleave": 1}
	state.BaseStats = BaseStatsView{Str: 14, Dex: 5, Vit: 10, Magic: 5}
	state.Level = 8
	sim, err := NewSimWithWorldProgression("sess_rend", "rend_seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatalf("NewSimWithWorldProgression: %v", err)
	}
	player := sim.entities[sim.playerID]
	player.mana = 50
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if target == nil {
		t.Fatal("missing combat_lab_soft_target")
	}

	cast := sim.Tick([]Input{{
		MessageID:     "rend_cast",
		CorrelationID: "corr_rend",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "rend", TargetID: idStr(target.id)},
	}})
	assertAck(t, cast, "rend_cast")
	if !eventListHasSkillEffect(cast.Events, "skill_effect_started", "rend") {
		t.Fatalf("rend cast missing bleed start: %+v", cast.Events)
	}
	if len(sim.bleedDots) < 1 {
		t.Fatalf("bleed dots = %d, want at least 1", len(sim.bleedDots))
	}
}

func TestRetributionReflectsOnBlock(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "paladin"
	state.Level = 8
	state.BaseStats = BaseStatsView{Str: 12, Dex: 6, Vit: 12, Magic: 10}
	state.SkillRanks = map[string]int{
		"retribution": 1, "hammer_of_light": 1, "charge": 1, "radiant_bolt": 1,
	}
	sim, err := NewSimWithWorldProgression("sess_retribution", "01", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatalf("NewSimWithWorldProgression: %v", err)
	}
	player := sim.entities[sim.playerID]
	player.mana = player.maxMana
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{
		MessageID:     "cast_retribution",
		CorrelationID: "corr_retribution",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "retribution"},
	}})
	assertAck(t, cast, "cast_retribution")
	effectState := sim.skillEffects["retribution"]
	if !effectState.ReflectOnBlock || effectState.Percent != 40 {
		t.Fatalf("retribution state = %+v, want reflect 40%%", effectState)
	}

	monster := findMonsterByDef(sim, "combat_lab_soft_target")
	if monster == nil {
		t.Fatal("missing combat_lab_soft_target")
	}
	beforeHP := monster.hp
	res := &TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.trySkillReflectOnPlayerBlock(player, monster, DamageRange{Min: 20, Max: 20}, "retribution_hit", res)
	if !eventListHasDamage(res.Events, "monster_damaged", "retribution") {
		t.Fatalf("reflect events = %+v", res.Events)
	}
	if monster.hp >= beforeHP {
		t.Fatalf("monster hp = %d, want reflected damage below %d", monster.hp, beforeHP)
	}
}

func TestPredatorsMarkOnProjectileHit(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "rogue"
	state.SkillRanks = map[string]int{"predators_mark": 1, "eviscerate": 1, "poison_stab": 1}
	state.BaseStats = BaseStatsView{Str: 4, Dex: 14, Vit: 5, Magic: 4}
	state.Level = 8
	sim, err := NewSimWithWorldProgression("sess_mark", "mark_seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatalf("NewSimWithWorldProgression: %v", err)
	}
	target := findMonsterByDef(sim, "combat_lab_soft_target")
	if target == nil {
		t.Fatal("missing combat_lab_soft_target")
	}
	def := rules.Skills["predators_mark"]
	projectile := &entity{
		kind:             projectileEntity,
		ownerID:          sim.playerID,
		sourceSkillID:    "predators_mark",
		sourceCorrID:     "corr_mark",
		damageRange:      sim.skillDamageRange(def, 1),
		sourceDamageType: sim.skillDamageType(def),
	}
	res := &TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.resolveSkillProjectileMonsterHit(projectile, target, res)
	if len(sim.rogueMarks) != 1 {
		t.Fatalf("rogue marks = %d, want 1", len(sim.rogueMarks))
	}
	if !eventListHasSkillEffect(res.Events, "skill_effect_started", "predators_mark") {
		t.Fatalf("mark events = %+v", res.Events)
	}
}

func eventListHasSkillEffect(events []Event, eventType, skillID string) bool {
	for _, ev := range events {
		if ev.EventType == eventType && ev.SkillID == skillID {
			return true
		}
	}

	return false
}

func eventListHasDamage(events []Event, eventType, skillID string) bool {
	for _, ev := range events {
		if ev.EventType == eventType && ev.SkillID == skillID && ev.Damage != nil && *ev.Damage > 0 {
			return true
		}
	}

	return false
}
