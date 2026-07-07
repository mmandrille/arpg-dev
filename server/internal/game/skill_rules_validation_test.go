package game

import (
	"strings"
	"testing"
)

func TestValidateSkillRulesRequiresBuffCooldownLongerThanDuration(t *testing.T) {
	skills := map[string]SkillDef{
		"short_buff": {
			Name:      "Short Buff",
			Class:     "barbarian",
			Tree:      SkillTreeDef{Tier: 1, Column: 1},
			Kind:      "self_buff",
			MaxRank:   1,
			Targeting: "self",
			Requirements: SkillRequirementDef{
				Level:        1,
				Stats:        map[string]int{},
				StatsPerRank: map[string]int{},
			},
			Effects: []SkillEffectDef{{
				Type:          "stat_percent_buff",
				Stats:         []string{"str"},
				PercentBase:   10,
				DurationTicks: 100,
			}},
			Cooldown: SkillCooldownDef{Type: "attack_interval_multiplier", Multiplier: 1, FlatTicks: 120},
		},
	}
	if err := validateSkillRules(skills, nil, 10); err == nil || !strings.Contains(err.Error(), "must be at least 150% of duration") {
		t.Fatalf("validateSkillRules error = %v, want buff cooldown ratio failure", err)
	}
	skill := skills["short_buff"]
	skill.Cooldown.FlatTicks = 140
	skills["short_buff"] = skill
	if err := validateSkillRules(skills, nil, 10); err != nil {
		t.Fatalf("validateSkillRules valid buff cooldown: %v", err)
	}
}
