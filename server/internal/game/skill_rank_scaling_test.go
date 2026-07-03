package game

import (
	"fmt"
	"testing"
)

func TestRankScaledIntCompoundIncreasesEveryRank(t *testing.T) {
	curve := RankScalingCurve{Type: "compound_percent", PercentPerRank: 8}
	prev := rankScaledInt(curve, 100, 10, 1)
	for rank := 2; rank <= 15; rank++ {
		next := rankScaledInt(curve, 100, 10, rank)
		if next <= prev {
			t.Fatalf("rank %d value %d should exceed rank %d value %d", rank, next, rank-1, prev)
		}
		prev = next
	}
}

func TestRankScaledIntLinearFallback(t *testing.T) {
	curve := RankScalingCurve{Type: "linear"}
	if got := rankScaledInt(curve, 5, 2, 3); got != 9 {
		t.Fatalf("linear rank 3 = %d, want 9", got)
	}
}

func TestSkillPointGrantLevelsV417Cadence(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		level int
		want  bool
	}{
		{level: 1, want: true},
		{level: 2, want: true},
		{level: 3, want: false},
		{level: 4, want: true},
		{level: 5, want: false},
		{level: 6, want: true},
		{level: 100, want: true},
		{level: 99, want: false},
	}
	for _, c := range cases {
		if got := rules.skillPointGrantLevel(c.level); got != c.want {
			t.Fatalf("skillPointGrantLevel(%d) = %v, want %v", c.level, got, c.want)
		}
	}
	if rules.totalSkillPointGrantsForLevel(100) != 51 {
		t.Fatalf("total grants at 100 = %d, want 51", rules.totalSkillPointGrantsForLevel(100))
	}
	if rules.totalSkillPointGrantsForLevel(6) != 4 {
		t.Fatalf("total grants at 6 = %d, want 4", rules.totalSkillPointGrantsForLevel(6))
	}
}

func TestSkillRankScalingMagicBoltDamageGrowsWithRank(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_rank_scaling_bolt", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	def := rules.Skills["magic_bolt"]
	prev := sim.skillDamageRange(def, 1)
	for rank := 2; rank <= 10; rank++ {
		next := sim.skillDamageRange(def, rank)
		if next.Min <= prev.Min || next.Max <= prev.Max {
			t.Fatalf("magic_bolt rank %d damage %+v should exceed rank %d %+v", rank, next, rank-1, prev)
		}
		prev = next
	}
}

func TestAllocateCastThenFourAllocatesReachRankFive(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "sorcerer"
	state.Level = 20
	state.UnspentSkillPoints = 10
	state.BaseStats = BaseStatsView{Str: 5, Dex: 5, Vit: 5, Magic: 20}
	rules.applyClassLevelStatGrowthFloor(&state)

	sim, err := NewSimWithWorldProgression("sess_alloc_cast", "seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatal(err)
	}

	alloc := sim.Tick([]Input{{
		MessageID:          "alloc_0",
		Type:               "allocate_skill_point_intent",
		AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: "magic_bolt"},
	}})
	if len(alloc.Rejects) > 0 {
		t.Fatalf("first allocate rejected: %+v", alloc.Rejects)
	}

	cast := sim.Tick([]Input{{
		MessageID: "cast_0",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: "magic_bolt", Direction: &Vec2{X: 1, Y: 0}},
	}})
	if len(cast.Rejects) > 0 {
		t.Fatalf("cast rejected: %+v", cast.Rejects)
	}

	for i := 1; i <= 4; i++ {
		res := sim.Tick([]Input{{
			MessageID:          fmt.Sprintf("alloc_%d", i),
			Type:               "allocate_skill_point_intent",
			AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: "magic_bolt"},
		}})
		if len(res.Rejects) > 0 {
			t.Fatalf("allocate %d rejected: %+v", i, res.Rejects)
		}
	}

	view := sim.SkillProgressionView()
	magic, ok := skillProgressionRow(view, "magic_bolt")
	if !ok || magic.Rank != 5 || view.UnspentSkillPoints != 5 {
		t.Fatalf("after allocate/cast/4x allocate: rank=%d unspent=%d base=%d",
			magic.Rank, view.UnspentSkillPoints, sim.progression.SkillRanks["magic_bolt"])
	}
}

func TestFiveAllocatesReachRankFive(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "sorcerer"
	state.Level = 20
	state.UnspentSkillPoints = 10
	state.BaseStats = BaseStatsView{Str: 5, Dex: 5, Vit: 5, Magic: 20}
	rules.applyClassLevelStatGrowthFloor(&state)

	sim, err := NewSimWithWorldProgression("sess_five_alloc", "seed", rules, "skill_progression_lab", state)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		res := sim.Tick([]Input{{
			MessageID:          fmt.Sprintf("alloc_%d", i),
			Type:               "allocate_skill_point_intent",
			AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: "magic_bolt"},
		}})
		if len(res.Rejects) > 0 {
			t.Fatalf("allocate %d rejected: %+v", i, res.Rejects)
		}
	}

	view := sim.SkillProgressionView()
	magic, ok := skillProgressionRow(view, "magic_bolt")
	if !ok || magic.Rank != 5 || view.UnspentSkillPoints != 5 {
		t.Fatalf("after 5 allocates: rank=%d unspent=%d base=%d view=%+v",
			magic.Rank, view.UnspentSkillPoints, sim.progression.SkillRanks["magic_bolt"], view)
	}
}

func TestEffectiveRankAboveMaxStillScales(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_rank_scaling_gear", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.SkillRanks["magic_bolt"] = 10
	item := &invItem{
		instanceID: 9501,
		itemDefID:  "amulet",
		slot:       "amulet",
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "amulet",
			DisplayName:    "Rare Amulet",
			Rarity:         "rare",
			SkillLevelBonuses: []SkillLevelBonusRoll{
				{SkillID: "magic_bolt", Value: 3},
			},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped["amulet"] = item.instanceID
	def := rules.Skills["magic_bolt"]
	base := sim.skillDamageRange(def, 10)
	boosted := sim.skillDamageRange(def, sim.effectiveSkillRank("magic_bolt"))
	if boosted.Min <= base.Min || boosted.Max <= base.Max {
		t.Fatalf("effective rank 13 damage %+v should exceed rank 10 %+v", boosted, base)
	}
}
