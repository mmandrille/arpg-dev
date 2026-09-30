package game

import (
	"math"
	"testing"
)

type skillProgressionGolden struct {
	Progression struct {
		PointsPerLevel int `json:"points_per_level"`
		SkillPoints    struct {
			PointsPerGrant     int `json:"points_per_grant"`
			FirstGrantLevel    int `json:"first_grant_level"`
			SecondGrantLevel   int `json:"second_grant_level"`
			GrantEveryLevels   int `json:"grant_every_levels"`
			GrantEveryMinLevel int `json:"grant_every_min_level"`
		} `json:"skill_points"`
		LevelCases []struct {
			Level                      int `json:"level"`
			ExpectedUnspentStatPoints  int `json:"expected_unspent_stat_points"`
			ExpectedUnspentSkillPoints int `json:"expected_unspent_skill_points"`
		} `json:"level_cases"`
	} `json:"progression"`
	AttackSpeed struct {
		BaseAttackIntervalTicks int     `json:"base_attack_interval_ticks"`
		MinEffectiveAttackSpeed float64 `json:"min_effective_attack_speed"`
		MaxEffectiveAttackSpeed float64 `json:"max_effective_attack_speed"`
		Cases                   []struct {
			Name                         string  `json:"name"`
			DexAttackSpeed               float64 `json:"dex_attack_speed"`
			WeaponAttackSpeed            float64 `json:"weapon_attack_speed"`
			ItemAttackSpeedPercent       int     `json:"item_attack_speed_percent"`
			ExpectedEffectiveAttackSpeed float64 `json:"expected_effective_attack_speed"`
			ExpectedAttackIntervalTicks  int     `json:"expected_attack_interval_ticks"`
			ExpectedMagicBoltCooldown    int     `json:"expected_magic_bolt_cooldown_ticks"`
		} `json:"cases"`
	} `json:"attack_speed"`
	Skill struct {
		SkillID      string `json:"skill_id"`
		MaxRank      int    `json:"max_rank"`
		Requirements struct {
			Level        int            `json:"level"`
			LevelPerRank int            `json:"level_per_rank"`
			Stats        map[string]int `json:"stats"`
			StatsPerRank map[string]int `json:"stats_per_rank"`
		} `json:"requirements"`
		RankRequirementCases []struct {
			Rank  int            `json:"rank"`
			Level int            `json:"level"`
			Stats map[string]int `json:"stats"`
		} `json:"rank_requirement_cases"`
		RankCases []struct {
			Rank     int `json:"rank"`
			ManaCost int `json:"mana_cost"`
			Damage   struct {
				MinPercent int `json:"min_percent"`
				MaxPercent int `json:"max_percent"`
			} `json:"damage"`
		} `json:"rank_cases"`
	} `json:"skill"`
}

// goldenWeaponTemplateBySpeed maps a golden weapon_attack_speed to a classless
// main-hand template that carries it.
var goldenWeaponTemplateBySpeed = map[float64]string{0.7: "great_sword"}

func loadSkillProgressionGolden(t *testing.T) skillProgressionGolden {
	t.Helper()
	var golden skillProgressionGolden
	loadGolden(t, "skill_points_and_magic_bolt.json", &golden)
	return golden
}

// TestSkillPointCadenceGolden asserts the shared skill/stat point cadence rules
// and drives real level-ups through awardExperience to check the unspent points.
func TestSkillPointCadenceGolden(t *testing.T) {
	golden := loadSkillProgressionGolden(t)
	rules := loadRules(t)

	if golden.Progression.PointsPerLevel != rules.CharacterProgression.PointsPerLevel {
		t.Fatalf("points_per_level golden=%d rules=%d", golden.Progression.PointsPerLevel, rules.CharacterProgression.PointsPerLevel)
	}
	cadence := rules.CharacterProgression.SkillPoints
	if golden.Progression.SkillPoints.PointsPerGrant != cadence.PointsPerGrant ||
		golden.Progression.SkillPoints.FirstGrantLevel != cadence.FirstGrantLevel ||
		golden.Progression.SkillPoints.SecondGrantLevel != cadence.SecondGrantLevel ||
		golden.Progression.SkillPoints.GrantEveryLevels != cadence.GrantEveryLevels ||
		golden.Progression.SkillPoints.GrantEveryMinLevel != cadence.GrantEveryMinLevel {
		t.Fatalf("skill point cadence golden=%+v rules=%+v", golden.Progression.SkillPoints, cadence)
	}

	for _, tc := range golden.Progression.LevelCases {
		if got := rules.totalSkillPointGrantsForLevel(tc.Level); got != tc.ExpectedUnspentSkillPoints {
			t.Fatalf("totalSkillPointGrantsForLevel(%d) = %d, want %d", tc.Level, got, tc.ExpectedUnspentSkillPoints)
		}
		// Server-authoritative level-up: reach the level through real XP awards.
		sim := MustNewSim("sess_golden_skill_points", "01", rules)
		if tc.Level > 1 {
			threshold, ok := rules.CharacterProgression.XPThresholds[tc.Level-1]
			if !ok {
				t.Fatalf("no XP threshold to reach level %d", tc.Level)
			}
			res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
			sim.awardExperience(threshold, "corr_golden_levelup", &res)
		}
		view := sim.CharacterProgressionView()
		if view.Level != tc.Level {
			t.Fatalf("level after XP award = %d, want %d", view.Level, tc.Level)
		}
		if view.UnspentStatPoints != tc.ExpectedUnspentStatPoints || view.UnspentSkillPoints != tc.ExpectedUnspentSkillPoints {
			t.Fatalf("level %d unspent stat/skill = %d/%d, want %d/%d",
				tc.Level, view.UnspentStatPoints, view.UnspentSkillPoints, tc.ExpectedUnspentStatPoints, tc.ExpectedUnspentSkillPoints)
		}
		if got := sim.totalEarnedStatPoints(); got != tc.ExpectedUnspentStatPoints {
			t.Fatalf("totalEarnedStatPoints at level %d = %d, want %d", tc.Level, got, tc.ExpectedUnspentStatPoints)
		}
	}
}

// TestAttackSpeedAndMagicBoltCooldownGolden drives the golden dex/weapon/item speed
// inputs through the real derived-stats pipeline (dex speed pinned via a temp-rule
// fixture, weapon and item percent via equipped items) and asserts effective speed,
// attack interval, and the Magic Bolt cooldown.
func TestAttackSpeedAndMagicBoltCooldownGolden(t *testing.T) {
	golden := loadSkillProgressionGolden(t)
	base := loadRules(t)

	if golden.AttackSpeed.BaseAttackIntervalTicks != base.Combat.BaseAttackIntervalTicks ||
		golden.AttackSpeed.MinEffectiveAttackSpeed != base.Combat.MinEffectiveAttackSpeed ||
		golden.AttackSpeed.MaxEffectiveAttackSpeed != base.Combat.MaxEffectiveAttackSpeed {
		t.Fatalf("attack speed rules golden=%+v combat=%+v", golden.AttackSpeed, base.Combat)
	}
	skillDef, ok := base.Skills[golden.Skill.SkillID]
	if !ok {
		t.Fatalf("unknown skill %q", golden.Skill.SkillID)
	}

	// The golden pins the dex-derived speed as an input. Rather than reverse-solving
	// for a stat build, override only the attack_speed derived formula so the
	// character's dex speed equals the golden value.
	for _, tc := range golden.AttackSpeed.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			rules := cloneRules(base)
			formula := rules.CharacterProgression.DerivedStats["attack_speed"]
			formula.PerStr, formula.PerDex, formula.PerVit, formula.PerMagic = 0, 0, 0, 0
			formula.Base = tc.DexAttackSpeed
			rules.CharacterProgression.DerivedStats["attack_speed"] = formula

			sim := MustNewSim("sess_golden_attack_speed", "01", rules)
			if got := sim.characterDerivedStatsView().AttackSpeed; math.Abs(got-tc.DexAttackSpeed) > 1e-9 {
				t.Fatalf("pinned dex attack speed = %v, want %v", got, tc.DexAttackSpeed)
			}

			nextID := uint64(6500)
			equip := func(slot string, item *invItem) {
				assertAck(t, sim.Tick([]Input{{MessageID: "equip-" + slot, Type: "equip_intent",
					Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: slot}}}), "equip-"+slot)
			}
			// Weapon speed comes from item data: a bare 1.0 is the unarmed/default baseline;
			// any other golden weapon speed is realized by an equipped template whose
			// attack_speed matches (verified below, so a rules change fails loudly).
			if tc.WeaponAttackSpeed != 1 {
				templateID, ok := goldenWeaponTemplateBySpeed[tc.WeaponAttackSpeed]
				if !ok {
					t.Fatalf("no golden weapon template registered for attack_speed %v", tc.WeaponAttackSpeed)
				}
				if got := rules.ItemTemplates[templateID].AttackSpeed; got != tc.WeaponAttackSpeed {
					t.Fatalf("template %s attack_speed = %v, want %v", templateID, got, tc.WeaponAttackSpeed)
				}
				weapon := addRolledInventoryItem(t, sim, nextID, templateID, nil)
				nextID++
				equip(mainHandSlot, weapon)
			}
			if tc.ItemAttackSpeedPercent != 0 {
				gloves := addRolledInventoryItem(t, sim, nextID, "gloves", map[string]int{"attack_speed_percent": tc.ItemAttackSpeedPercent})
				equip("gloves", gloves)
			}

			derived := sim.DerivedStatsView()
			if math.Abs(derived.AttackSpeed-tc.ExpectedEffectiveAttackSpeed) > 1e-6 {
				t.Fatalf("effective attack speed = %v, want %v", derived.AttackSpeed, tc.ExpectedEffectiveAttackSpeed)
			}
			if derived.AttackIntervalTicks != tc.ExpectedAttackIntervalTicks {
				t.Fatalf("attack interval = %d, want %d", derived.AttackIntervalTicks, tc.ExpectedAttackIntervalTicks)
			}
			if got := sim.attackIntervalTicksFromSpeed(tc.ExpectedEffectiveAttackSpeed); got != tc.ExpectedAttackIntervalTicks {
				t.Fatalf("attackIntervalTicksFromSpeed(%v) = %d, want %d", tc.ExpectedEffectiveAttackSpeed, got, tc.ExpectedAttackIntervalTicks)
			}
			if got := sim.skillCooldownTicks(skillDef); got != tc.ExpectedMagicBoltCooldown {
				t.Fatalf("%s cooldown = %d, want %d", golden.Skill.SkillID, got, tc.ExpectedMagicBoltCooldown)
			}
		})
	}
}

// TestMagicBoltRankScalingGolden asserts the Magic Bolt requirement and rank
// scaling contract (max rank, per-rank level/stat requirements, mana cost, and
// weapon-percent damage).
func TestMagicBoltRankScalingGolden(t *testing.T) {
	golden := loadSkillProgressionGolden(t)
	rules := loadRules(t)
	def, ok := rules.Skills[golden.Skill.SkillID]
	if !ok {
		t.Fatalf("unknown skill %q", golden.Skill.SkillID)
	}

	if golden.Skill.MaxRank != def.MaxRank {
		t.Fatalf("max_rank golden=%d rules=%d", golden.Skill.MaxRank, def.MaxRank)
	}
	req := def.Requirements
	if golden.Skill.Requirements.Level != req.Level ||
		golden.Skill.Requirements.LevelPerRank != req.LevelPerRank ||
		golden.Skill.Requirements.Stats["magic"] != req.Stats["magic"] ||
		golden.Skill.Requirements.StatsPerRank["magic"] != req.StatsPerRank["magic"] {
		t.Fatalf("requirements golden=%+v rules=%+v", golden.Skill.Requirements, req)
	}

	for _, tc := range golden.Skill.RankRequirementCases {
		got := skillRequirementsForRank(req, tc.Rank)
		if got["level"] != tc.Level {
			t.Fatalf("rank %d required level = %d, want %d", tc.Rank, got["level"], tc.Level)
		}
		for stat, want := range tc.Stats {
			if got[stat] != want {
				t.Fatalf("rank %d required %s = %d, want %d (got %+v)", tc.Rank, stat, got[stat], want, got)
			}
		}
	}

	for _, tc := range golden.Skill.RankCases {
		if got := skillManaCost(rules, def, tc.Rank); got != tc.ManaCost {
			t.Fatalf("rank %d mana cost = %d, want %d", tc.Rank, got, tc.ManaCost)
		}
		if got := skillWeaponMultiplierPercent(rules, def, tc.Rank, true); got != tc.Damage.MinPercent {
			t.Fatalf("rank %d min damage percent = %d, want %d", tc.Rank, got, tc.Damage.MinPercent)
		}
		if got := skillWeaponMultiplierPercent(rules, def, tc.Rank, false); got != tc.Damage.MaxPercent {
			t.Fatalf("rank %d max damage percent = %d, want %d", tc.Rank, got, tc.Damage.MaxPercent)
		}
	}
}
