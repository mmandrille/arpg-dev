package game

import "testing"

type tier4PassiveCase struct {
	classID      string
	skillID      string
	prerequisite string
	statKey      string
	minStatTotal int
}

func tier4PassiveCases() []tier4PassiveCase {
	return []tier4PassiveCase{
		{classID: "sorcerer", skillID: "arcane_reservoir", prerequisite: "spell_dynamo", statKey: "max_mana_percent", minStatTotal: 12},
		{classID: "barbarian", skillID: "unstoppable_heart", prerequisite: "crushing_force", statKey: "max_hp_percent", minStatTotal: 8},
		{classID: "paladin", skillID: "oathbound_resolve", prerequisite: "consecrated_vitality", statKey: "max_mana_percent", minStatTotal: 10},
		{classID: "ranger", skillID: "wildborn_endurance", prerequisite: "deadeye", statKey: "movement_speed_percent", minStatTotal: 8},
	}
}

func TestTier4PassiveUnlockRequiresLevel15AndPrerequisite(t *testing.T) {
	rules := loadRules(t)
	for _, c := range tier4PassiveCases() {
		t.Run(c.skillID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = c.classID
			state.Level = 14
			state.BaseStats = rules.CharacterProgression.Classes[c.classID].BaseStats
			state.UnspentSkillPoints = 1
			state.SkillRanks = map[string]int{c.prerequisite: 1}
			sim, err := NewSimWithWorldProgression("sess_tier4_locked_"+c.skillID, "tier4_passive_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			assertPassiveSpendable(t, sim, c.skillID, false)

			state.Level = 15
			sim, err = NewSimWithWorldProgression("sess_tier4_unlock_"+c.skillID, "tier4_passive_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new level-15 sim: %v", err)
			}
			assertPassiveSpendable(t, sim, c.skillID, true)

			spend := sim.Tick([]Input{{
				MessageID:          "spend_tier4",
				CorrelationID:      "corr_spend_tier4",
				Type:               "allocate_skill_point_intent",
				AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: c.skillID},
			}})
			assertAck(t, spend, "spend_tier4")
			if sim.progression.SkillRanks[c.skillID] != 1 {
				t.Fatalf("rank after spend = %d, want 1", sim.progression.SkillRanks[c.skillID])
			}
		})
	}
}

func TestTier4PassiveStatBonusesApplyFromRules(t *testing.T) {
	rules := loadRules(t)
	for _, c := range tier4PassiveCases() {
		t.Run(c.skillID, func(t *testing.T) {
			def := rules.Skills[c.skillID]
			want := def.PassiveStats.Stats[c.statKey].Base
			if want < c.minStatTotal {
				t.Fatalf("%s rule base %d < test floor %d", c.statKey, want, c.minStatTotal)
			}

			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = c.classID
			state.Level = 15
			state.BaseStats = rules.CharacterProgression.Classes[c.classID].BaseStats
			state.SkillRanks = map[string]int{c.prerequisite: 1, c.skillID: 1}
			sim, err := NewSimWithWorldProgression("sess_tier4_stats_"+c.skillID, "tier4_passive_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			if got := sim.passiveSkillStatTotal(c.statKey); got < want {
				t.Fatalf("%s passive total = %d, want at least %d from rules", c.statKey, got, want)
			}
		})
	}
}
