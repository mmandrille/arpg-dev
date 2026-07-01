package game

import (
	"math"
	"testing"
)

func TestPassiveSkillColumnUnlocksByClassLevelAndPrerequisite(t *testing.T) {
	rules := loadRules(t)
	for classID, chain := range passiveSkillColumnTestChains() {
		t.Run(classID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = classID
			state.Level = 1
			state.BaseStats = rules.CharacterProgression.Classes[classID].BaseStats
			state.UnspentSkillPoints = 3
			sim, err := NewSimWithWorldProgression("sess_passive_column_"+classID, "passive_column_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}

			assertPassiveSpendable(t, sim, chain[0], true)
			assertPassiveSpendable(t, sim, chain[1], false)
			assertPassiveSpendable(t, sim, chain[2], false)

			spendFirst := sim.Tick([]Input{{
				MessageID:          "spend_first",
				CorrelationID:      "corr_spend_first",
				Type:               "allocate_skill_point_intent",
				AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: chain[0]},
			}})
			assertAck(t, spendFirst, "spend_first")

			state.Level = 5
			state.SkillRanks = map[string]int{chain[0]: 1}
			state.UnspentSkillPoints = 2
			sim, err = NewSimWithWorldProgression("sess_passive_column_"+classID+"_row2", "passive_column_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new row2 sim: %v", err)
			}
			assertPassiveSpendable(t, sim, chain[1], true)

			spendSecond := sim.Tick([]Input{{
				MessageID:          "spend_second",
				CorrelationID:      "corr_spend_second",
				Type:               "allocate_skill_point_intent",
				AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: chain[1]},
			}})
			assertAck(t, spendSecond, "spend_second")

			state.Level = 10
			state.SkillRanks = map[string]int{chain[0]: 1, chain[1]: 1}
			state.UnspentSkillPoints = 1
			sim, err = NewSimWithWorldProgression("sess_passive_column_"+classID+"_row3", "passive_column_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new row3 sim: %v", err)
			}
			assertPassiveSpendable(t, sim, chain[2], true)
		})
	}
}

func TestPassiveSkillColumnDefinitionsAndStatBonuses(t *testing.T) {
	rules := loadRules(t)
	for classID, chain := range passiveSkillColumnTestChains() {
		t.Run(classID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = classID
			state.Level = 10
			state.BaseStats = rules.CharacterProgression.Classes[classID].BaseStats
			sim, err := NewSimWithWorldProgression("sess_passive_stats_"+classID, "passive_stats_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}

			for index, skillID := range chain {
				def := rules.Skills[skillID]
				wantLevel := []int{1, 5, 10}[index]
				if def.Kind != "passive_stat_bonus" || def.Targeting != "self" || def.MaxRank != 1 {
					t.Fatalf("%s def = %+v, want one-rank self passive_stat_bonus", skillID, def)
				}
				if def.Requirements.Level != wantLevel || def.Requirements.LevelPerRank != 0 || len(def.Requirements.Stats) != 0 || len(def.Requirements.StatsPerRank) != 0 {
					t.Fatalf("%s requirements = %+v, want level %d and no stat requirements", skillID, def.Requirements, wantLevel)
				}
				if index == 0 && len(def.Requirements.Skills) != 0 {
					t.Fatalf("%s prereqs = %+v, want none", skillID, def.Requirements.Skills)
				}
				if index > 0 && (len(def.Requirements.Skills) != 1 || def.Requirements.Skills[0].SkillID != chain[index-1] || def.Requirements.Skills[0].Rank != 1) {
					t.Fatalf("%s prereqs = %+v, want %s rank 1", skillID, def.Requirements.Skills, chain[index-1])
				}
				sim.progression.SkillRanks[skillID] = 1
				for stat, value := range def.PassiveStats.Stats {
					if got := sim.passiveSkillStatTotal(stat); got < value.Base {
						t.Fatalf("%s passive stat %s total = %d, want at least %d", skillID, stat, got, value.Base)
					}
				}
			}
		})
	}
}

func TestPassiveSkillColumnAffectsDerivedStatsAndCannotCast(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		classID string
		skillID string
		statKey string
	}{
		{classID: "sorcerer", skillID: "arcane_focus", statKey: "max_mana"},
		{classID: "barbarian", skillID: "iron_hide", statKey: "max_hp"},
		{classID: "paladin", skillID: "vigilant_guard", statKey: "armor"},
		{classID: "rogue", skillID: "quick_hands", statKey: "attack_speed"},
		{classID: "ranger", skillID: "trail_sense", statKey: "light_radius"},
	}
	for _, c := range cases {
		t.Run(c.skillID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = c.classID
			state.Level = 10
			state.BaseStats = rules.CharacterProgression.Classes[c.classID].BaseStats
			sim, err := NewSimWithWorldProgression("sess_passive_effect_"+c.skillID, "passive_effect_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			before := findStatBreakdown(sim.StatBreakdownViews(), c.statKey)
			if before == nil {
				t.Fatalf("missing %s breakdown before passive", c.statKey)
			}

			sim.progression.SkillRanks[c.skillID] = 1
			after := findStatBreakdown(sim.StatBreakdownViews(), c.statKey)
			if after == nil || after.Value <= before.Value || !statBreakdownHasSourceKind(*after, "passive_skill") {
				t.Fatalf("%s after passive = %+v, before %+v; want increased passive_skill source", c.statKey, after, before)
			}

			castRejected := sim.Tick([]Input{{
				MessageID:     "cast_passive",
				CorrelationID: "corr_cast_passive",
				Type:          "cast_skill_intent",
				CastSkill:     &CastSkillIntent{SkillID: c.skillID, Direction: &Vec2{X: 1}},
			}})
			assertReject(t, castRejected, "cast_passive", "passive_skill_not_castable")
		})
	}
}

func TestPassiveSkillPercentStatsApplyAfterFlatStatsAndDoNotCompound(t *testing.T) {
	rules := loadRules(t)
	tempo := rules.Skills["battle_tempo"]
	tempo.PassiveStats.Stats = map[string]SkillRankValueDef{
		"max_hp_percent": {Base: 5, PerRank: 0},
	}
	rules.Skills["battle_tempo"] = tempo

	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.Level = 10
	state.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_passive_percent_order", "passive_percent_order_seed", rules, DefaultWorldID, state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	ring := addRolledInventoryItem(t, sim, 7310, "ring", map[string]int{"max_hp": 20})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_hp_ring", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "equip_hp_ring")
	flatOnly, _ := sim.playerEffectiveCombatStats()

	sim.progression.SkillRanks["iron_hide"] = 1
	sim.progression.SkillRanks["battle_tempo"] = 1
	withPercent, _ := sim.playerEffectiveCombatStats()
	want := flatOnly.MaxHP * 1.15
	if math.Abs(withPercent.MaxHP-want) > 0.000001 {
		t.Fatalf("max hp with passives = %v, want flat %v * summed percent 1.15 = %v", withPercent.MaxHP, flatOnly.MaxHP, want)
	}
	compounded := flatOnly.MaxHP * 1.10 * 1.05
	if math.Abs(withPercent.MaxHP-compounded) < 0.000001 {
		t.Fatalf("max hp compounded percent sources: got %v, compounded %v", withPercent.MaxHP, compounded)
	}
}

func passiveSkillColumnTestChains() map[string][]string {
	return map[string][]string{
		"sorcerer":  {"arcane_focus", "mana_weaving", "spell_dynamo"},
		"barbarian": {"iron_hide", "battle_tempo", "crushing_force"},
		"paladin":   {"vigilant_guard", "faithful_bulwark", "consecrated_vitality"},
		"rogue":     {"quick_hands", "killer_instinct", "evasive_footwork"},
		"ranger":    {"trail_sense", "precision_draw", "deadeye"},
	}
}

func assertPassiveSpendable(t *testing.T, sim *Sim, skillID string, want bool) {
	t.Helper()
	row, ok := skillProgressionRow(sim.SkillProgressionView(), skillID)
	if !ok {
		t.Fatalf("%s missing from skill progression", skillID)
	}
	if row.CanSpend != want {
		t.Fatalf("%s can_spend = %v, want %v; row=%+v", skillID, row.CanSpend, want, row)
	}
}
