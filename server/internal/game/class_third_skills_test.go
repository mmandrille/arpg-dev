package game

import "testing"

func TestThirdClassSkillsRequirePrerequisites(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		classID string
		skillID string
		prereq  string
		stats   BaseStatsView
	}{
		{classID: "sorcerer", skillID: "arcane_barrage", prereq: "ligthing", stats: BaseStatsView{Str: 3, Dex: 5, Vit: 5, Magic: 11}},
		{classID: "barbarian", skillID: "earthbreaker", prereq: "cleave", stats: BaseStatsView{Str: 8, Dex: 5, Vit: 8, Magic: 5}},
		{classID: "paladin", skillID: "sanctuary", prereq: "holy_shield", stats: BaseStatsView{Str: 6, Dex: 4, Vit: 10, Magic: 10}},
		{classID: "rogue", skillID: "shadow_flurry", prereq: "dash", stats: BaseStatsView{Str: 4, Dex: 12, Vit: 5, Magic: 4}},
		{classID: "ranger", skillID: "split_arrow", prereq: "volley", stats: BaseStatsView{Str: 4, Dex: 14, Vit: 5, Magic: 4}},
	}
	for _, c := range cases {
		t.Run(c.skillID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = c.classID
			state.BaseStats = c.stats
			state.UnspentSkillPoints = 2
			sim, err := NewSimWithWorldProgression("sess_"+c.skillID+"_prereq", "class_third_skill_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			blocked, ok := skillProgressionRow(sim.SkillProgressionView(), c.skillID)
			if !ok || blocked.CanSpend {
				t.Fatalf("%s blocked row = %+v ok=%v, want visible but not spendable before %s", c.skillID, blocked, ok, c.prereq)
			}
			sim.progression.SkillRanks[c.prereq] = 1
			unlocked, ok := skillProgressionRow(sim.SkillProgressionView(), c.skillID)
			if !ok || !unlocked.CanSpend {
				t.Fatalf("%s unlocked row = %+v ok=%v, want spendable after %s rank 1", c.skillID, unlocked, ok, c.prereq)
			}
		})
	}
}
