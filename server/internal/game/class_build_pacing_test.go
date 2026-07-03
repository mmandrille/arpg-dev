package game

import "testing"

func TestSkillPointGrantLevels(t *testing.T) {
	rules := loadRules(t)
	if rules.totalSkillPointGrantsForLevel(6) != 4 {
		t.Fatalf("total grants at 6 = %d, want 4", rules.totalSkillPointGrantsForLevel(6))
	}
}

func TestClassLevelStatGrowth(t *testing.T) {
	rules := loadRules(t)
	growth := rules.classLevelStatGrowthTotal("barbarian", 6)
	if growth.Str != 5 || growth.Dex != 0 {
		t.Fatalf("barbarian growth at 6 = %+v, want +5 STR", growth)
	}
	grown := rules.classGrownBaseStats("sorcerer", 4)
	if grown.Magic != 8 {
		t.Fatalf("sorcerer grown magic at 4 = %d, want 8 (5 base + 3 growth)", grown.Magic)
	}

	sim := MustNewSim("sess_class_growth", "01", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim.savePlayer(sim.defaultPlayer())
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.awardExperience(52, "corr_growth", &res)
	if sim.progression.Level != 3 || sim.progression.BaseStats.Str != 7 {
		t.Fatalf("barbarian after level 3 = level %d stats %+v, want level 3 STR 7", sim.progression.Level, sim.progression.BaseStats)
	}
}

func TestClassBuildPacingSkillsRegistered(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		skillID string
		classID string
		tier    int
		column  int
		prereq  string
	}{
		{skillID: "rain_of_arrows", classID: "ranger", tier: 3, column: 1, prereq: "volley"},
		{skillID: "explosive_shot", classID: "ranger", tier: 3, column: 4, prereq: "snipe"},
		{skillID: "fireball", classID: "sorcerer", tier: 3, column: 1, prereq: "ice_shard"},
		{skillID: "energy_ward", classID: "sorcerer", tier: 3, column: 3, prereq: "teleport"},
		{skillID: "war_cry", classID: "barbarian", tier: 3, column: 4, prereq: "leap"},
		{skillID: "hammer_of_light", classID: "paladin", tier: 3, column: 2, prereq: "charge"},
		{skillID: "smoke_screen", classID: "rogue", tier: 3, column: 4, prereq: "shadowstep"},
	}
	for _, c := range cases {
		def, ok := rules.Skills[c.skillID]
		if !ok {
			t.Fatalf("missing skill %s", c.skillID)
		}
		if def.Class != c.classID || def.Tree.Tier != c.tier || def.Tree.Column != c.column {
			t.Fatalf("%s placement = class %s tier %d col %d, want %s T%d C%d", c.skillID, def.Class, def.Tree.Tier, def.Tree.Column, c.classID, c.tier, c.column)
		}
		if len(def.Requirements.Skills) != 1 || def.Requirements.Skills[0].SkillID != c.prereq {
			t.Fatalf("%s prereq = %+v, want %s", c.skillID, def.Requirements.Skills, c.prereq)
		}
	}
}

func TestRainOfArrowsCast(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "ranger"
	state.BaseStats = BaseStatsView{Str: 4, Dex: 16, Vit: 5, Magic: 3}
	state.Level = 8
	state.SkillRanks = map[string]int{
		"piercing_shot": 1, "volley": 1, "snipe": 1, "rain_of_arrows": 1,
	}
	sim, err := NewSimWithWorldProgression("sess_rain", "rain_seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.defaultPlayer()
	entity := sim.activeLevel().entities[sim.playerID]
	entity.mana = 50
	sim.savePlayer(player)

	cast := sim.Tick([]Input{{
		MessageID:     "rain_cast",
		CorrelationID: "corr_rain",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "rain_of_arrows", Direction: &Vec2{X: 1, Y: 0}},
	}})
	assertAck(t, cast, "rain_cast")
	if !hasEvent(cast, "skill_cast") {
		t.Fatalf("missing skill_cast: %+v", cast.Events)
	}
}

func TestSmokeScreenEvadeBuff(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "rogue"
	state.BaseStats = BaseStatsView{Str: 4, Dex: 14, Vit: 5, Magic: 4}
	state.Level = 6
	state.UnspentSkillPoints = 3
	state.SkillRanks = map[string]int{
		"shadowstep":   1,
		"smoke_screen": 1,
	}
	sim, err := NewSimWithWorldProgression("sess_smoke", "smoke_seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	player := sim.defaultPlayer()
	entity := sim.activeLevel().entities[sim.playerID]
	entity.mana = 50
	sim.savePlayer(player)

	cast := sim.Tick([]Input{{
		MessageID:     "smoke_cast",
		CorrelationID: "corr_smoke",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "smoke_screen"},
	}})
	assertAck(t, cast, "smoke_cast")
	stats := sim.DerivedStatsView()
	if stats.EvadeChance < 0.29 {
		t.Fatalf("evade after smoke = %v, want at least 0.30", stats.EvadeChance)
	}
}
