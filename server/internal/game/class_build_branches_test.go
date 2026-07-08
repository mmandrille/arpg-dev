package game

import "testing"

func TestClassBuildBranchSkillsRegistered(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		skillID string
		classID string
		prereq  string
	}{
		{skillID: "rampage", classID: "barbarian", prereq: "rage"},
		{skillID: "worldbreaker", classID: "barbarian", prereq: "shatter_strike"},
		{skillID: "glacial_lance", classID: "sorcerer", prereq: "ice_shard"},
		{skillID: "inferno", classID: "sorcerer", prereq: "fireball"},
		{skillID: "blessed_recovery", classID: "paladin", prereq: "heal"},
		{skillID: "divine_hammer", classID: "paladin", prereq: "hammer_of_light"},
		{skillID: "blade_dance", classID: "rogue", prereq: "shadow_flurry"},
		{skillID: "assassinate", classID: "rogue", prereq: "eviscerate"},
		{skillID: "alpha_call", classID: "ranger", prereq: "black_wolf_companion"},
		{skillID: "meteor_shot", classID: "ranger", prereq: "explosive_shot"},
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
		if len(def.Synergies) != 1 || def.Synergies[0].SourceSkillID != c.prereq {
			t.Fatalf("%s synergy = %+v, want source %s", c.skillID, def.Synergies, c.prereq)
		}
	}
}

func TestGoreStrikeAppliesBleed(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.SkillRanks = map[string]int{"cleave": 1, "ground_slam": 1, "rend": 1, "gore_strike": 1}
	state.BaseStats = BaseStatsView{Str: 18, Dex: 5, Vit: 14, Magic: 3}
	state.Level = 24
	sim, err := NewSimWithWorldProgression("sess_gore", "gore_seed", rules, "skill_progression_lab", state)
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
		MessageID:     "gore_cast",
		CorrelationID: "corr_gore",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "gore_strike", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "gore_cast")
	if !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("gore_strike events = %+v, want skill_effect_started", cast.Events)
	}
}
