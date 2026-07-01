package game

import (
	"path/filepath"
	"testing"
)

func TestTier2Column3SkillExpansion(t *testing.T) {
	rules, err := LoadRules(filepath.Join("..", "..", "..", "shared", "rules"))
	if err != nil {
		t.Fatalf("LoadRules: %v", err)
	}
	want := map[string]string{
		"ground_slam":    "barbarian",
		"arcane_orb":     "sorcerer",
		"radiant_bolt":   "paladin",
		"fan_of_blades":  "rogue",
		"snipe":          "ranger",
	}
	for skillID, classID := range want {
		skill, ok := rules.Skills[skillID]
		if !ok {
			t.Fatalf("missing skill %s", skillID)
		}
		if skill.Class != classID {
			t.Fatalf("skill %s class = %q, want %q", skillID, skill.Class, classID)
		}
		if skill.Tree.Tier != 2 || skill.Tree.Column != 3 {
			t.Fatalf("skill %s tree = tier %d column %d, want tier 2 column 3", skillID, skill.Tree.Tier, skill.Tree.Column)
		}
		if skill.Kind == "passive_stat_bonus" {
			t.Fatalf("skill %s should be active, got passive", skillID)
		}
		if len(skill.Requirements.Skills) == 0 {
			t.Fatalf("skill %s should require a prerequisite skill", skillID)
		}
	}
}
