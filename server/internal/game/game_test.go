package game

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden fixtures with current sim output")

// --- shared fixture helpers -------------------------------------------------

func sharedDir(t *testing.T) string {
	t.Helper()
	rulesDir, err := FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate shared/rules: %v", err)
	}
	return filepath.Dir(rulesDir) // .../shared
}

func loadRules(t *testing.T) *Rules {
	t.Helper()
	rulesDir, err := FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate rules: %v", err)
	}
	rules, err := LoadRules(rulesDir)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return rules
}

func loadRulesWithMainGameplay(t *testing.T, overrides map[string]any) *Rules {
	t.Helper()
	sourceDir, err := FindSharedRulesDir()
	if err != nil {
		t.Fatalf("locate rules: %v", err)
	}
	sourceSharedDir := filepath.Dir(sourceDir)
	targetSharedDir := t.TempDir()
	targetRulesDir := filepath.Join(targetSharedDir, "rules")
	if err := copyTree(sourceDir, targetRulesDir); err != nil {
		t.Fatalf("copy rules: %v", err)
	}
	if err := copyTree(filepath.Join(sourceSharedDir, "content"), filepath.Join(targetSharedDir, "content")); err != nil {
		t.Fatalf("copy content: %v", err)
	}

	configPath := filepath.Join(targetRulesDir, "main_config.v0.json")
	var config map[string]any
	b, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read temp main_config: %v", err)
	}
	if err := json.Unmarshal(b, &config); err != nil {
		t.Fatalf("parse temp main_config: %v", err)
	}
	gameplay, ok := config["gameplay"].(map[string]any)
	if !ok {
		t.Fatal("temp main_config gameplay missing")
	}
	for key, value := range overrides {
		gameplay[key] = value
	}
	b, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("marshal temp main_config: %v", err)
	}
	if err := os.WriteFile(configPath, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write temp main_config: %v", err)
	}

	rules, err := LoadRules(targetRulesDir)
	if err != nil {
		t.Fatalf("load temp rules: %v", err)
	}
	return rules
}

func copyTree(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(targetDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func loadGolden(t *testing.T, name string, v any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(sharedDir(t), "golden", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("parse golden %s: %v", name, err)
	}
}

func writeGolden(t *testing.T, name string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden %s: %v", name, err)
	}
	path := filepath.Join(sharedDir(t), "golden", name)
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write golden %s: %v", name, err)
	}
	t.Logf("updated golden: %s", name)
}

// --- rules ------------------------------------------------------------------

func TestLoadRules(t *testing.T) {
	r := loadRules(t)
	if r.MainConfig.Gameplay.BaseAttackIntervalTicks != 14 ||
		r.MainConfig.Gameplay.BaseMovementSpeed != 1.0 ||
		r.MainConfig.Gameplay.BaseDropRatePercent != 25 ||
		r.MainConfig.Gameplay.RespecCostGold != 250 {
		t.Fatalf("main config gameplay = %+v", r.MainConfig.Gameplay)
	}
	if r.Combat.PlayerDamage.Min != 2 || r.Combat.PlayerDamage.Max != 4 {
		t.Fatalf("combat player_damage = %+v, want {2,4}", r.Combat.PlayerDamage)
	}
	dummy := r.Monsters[monsterDefID]
	if dummy.MaxHP != 3 {
		t.Fatalf("training_dummy max_hp = %d, want 3", dummy.MaxHP)
	}
	if dummy.RetaliationDamage == nil || dummy.RetaliationDamage.Min != 1 || dummy.RetaliationDamage.Max != 1 {
		t.Fatalf("training_dummy retaliation_damage = %+v, want {1,1}", dummy.RetaliationDamage)
	}
	if !r.Items["rusty_sword"].Equippable || r.Items["rusty_sword"].Slot != "main_hand" {
		t.Fatalf("rusty_sword def = %+v", r.Items["rusty_sword"])
	}
	if r.Items["rusty_sword"].Damage == nil || r.Items["rusty_sword"].Damage.Min != 3 || r.Items["rusty_sword"].Damage.Max != 5 {
		t.Fatalf("rusty_sword damage = %+v, want {3,5}", r.Items["rusty_sword"].Damage)
	}
	if r.Items["rusty_sword"].Reach == nil || *r.Items["rusty_sword"].Reach != 1.5 {
		t.Fatalf("rusty_sword reach = %+v, want 1.5", r.Items["rusty_sword"].Reach)
	}
	if r.Combat.UnarmedReach != 1.0 {
		t.Fatalf("unarmed reach = %v, want 1.0", r.Combat.UnarmedReach)
	}
	if r.Items["quest_leaf"].Equippable || r.Items["quest_leaf"].Slot != "" {
		t.Fatalf("quest_leaf def = %+v, want non-equippable without slot", r.Items["quest_leaf"])
	}
	if r.Items["quest_leaf"].Damage != nil {
		t.Fatalf("quest_leaf damage = %+v, want nil", r.Items["quest_leaf"].Damage)
	}
	if r.Items["gold"].Category != "currency" || r.Items["gold"].Gold == nil || r.Items["gold"].Gold.Min != 6 || r.Items["gold"].Gold.Max != 12 {
		t.Fatalf("gold def = %+v, want currency amount {6,12}", r.Items["gold"])
	}
	potion := r.Items["red_potion"]
	if potion.Category != "consumable" || potion.Heal == nil || potion.Heal.Min != 5 || potion.Heal.Max != 5 {
		t.Fatalf("red_potion def = %+v, want consumable heal {5,5}", potion)
	}
	manaPotion := r.Items["blue_potion"]
	if manaPotion.Category != "consumable" || manaPotion.ManaRestore == nil || manaPotion.ManaRestore.Min != 5 || manaPotion.ManaRestore.Max != 5 {
		t.Fatalf("blue_potion def = %+v, want consumable mana_restore {5,5}", manaPotion)
	}
	if _, ok := r.Worlds[DefaultWorldID]; !ok {
		t.Fatalf("missing default world %q", DefaultWorldID)
	}
	if _, ok := r.Worlds["gear_before_combat"]; !ok {
		t.Fatal("missing gear_before_combat world")
	}
	if _, ok := r.Worlds["collision_lab"]; !ok {
		t.Fatal("missing collision_lab world")
	}
	if _, ok := r.Worlds["door_lab"]; !ok {
		t.Fatal("missing door_lab world")
	}
	if _, ok := r.Worlds["ranged_lab"]; !ok {
		t.Fatal("missing ranged_lab world")
	}
	if _, ok := r.Worlds["inventory_lab"]; !ok {
		t.Fatal("missing inventory_lab world")
	}
	bow := r.Items["training_bow"]
	if !bow.Equippable || bow.Slot != "main_hand" || bow.AttackMode != attackModeRanged || bow.Damage == nil || bow.Reach == nil || bow.ProjectileSpeed == nil {
		t.Fatalf("training_bow def = %+v, want ranged weapon", bow)
	}
	if r.Items["barbarian_axe"].ClassRequired != "barbarian" || r.Items["sorcerer_staff"].ClassRequired != "sorcerer" || r.Items["paladin_mace"].ClassRequired != "paladin" || r.Items["ranger_shortbow"].ClassRequired != "ranger" {
		t.Fatalf("class weapon requirements = axe:%q staff:%q mace:%q bow:%q", r.Items["barbarian_axe"].ClassRequired, r.Items["sorcerer_staff"].ClassRequired, r.Items["paladin_mace"].ClassRequired, r.Items["ranger_shortbow"].ClassRequired)
	}
	if r.Interactables["wooden_door"].InitialState != interactableClosed {
		t.Fatalf("wooden_door = %+v, want initially closed", r.Interactables["wooden_door"])
	}
	if r.CharacterProgression.PointsPerLevel != 3 || r.CharacterProgression.LevelCap != 100 {
		t.Fatalf("character progression = %+v, want points_per_level 3 level_cap 100", r.CharacterProgression)
	}
	if len(r.CharacterProgression.Classes) != 5 ||
		r.CharacterProgression.Classes["barbarian"].BaseStats.Str != 5 ||
		r.CharacterProgression.Classes["sorcerer"].BaseStats.Str != 3 ||
		r.CharacterProgression.Classes["paladin"].BaseStats.Vit != 8 ||
		r.CharacterProgression.Classes["rogue"].BaseStats.Dex != 8 ||
		r.CharacterProgression.Classes["ranger"].BaseStats.Dex != 8 ||
		r.CharacterProgression.Classes["ranger"].BaseStats.Magic != 3 {
		t.Fatalf("character classes = %+v, want barbarian/sorcerer/paladin/rogue/ranger starting stats", r.CharacterProgression.Classes)
	}
	if r.CharacterProgression.SkillPoints.PointsPerGrant != 1 || r.CharacterProgression.SkillPoints.GrantEveryLevels != 3 || r.CharacterProgression.SkillPoints.FirstGrantLevel != 1 {
		t.Fatalf("skill point cadence = %+v, want 1 point every 3 levels starting at 1", r.CharacterProgression.SkillPoints)
	}
	if skill := r.Skills[magicBoltSkillID]; skill.Class != "sorcerer" || skill.MaxRank != 5 || skill.Kind != "projectile_attack" || skill.Cooldown.Type != "attack_interval_multiplier" || skill.Requirements.Stats["magic"] != 5 || skill.Requirements.LevelPerRank != 1 || skill.Requirements.StatsPerRank["magic"] != 3 {
		t.Fatalf("magic_bolt skill = %+v, want projectile_attack max rank 5 magic 5 +3/rank level +1/rank attack interval cooldown", skill)
	}
	if scaling := r.Skills[magicBoltSkillID].Damage.MagicScaling; scaling.Stat != "magic" || scaling.PercentPerPoint <= 0 || scaling.MaxBonusPercent <= 0 || !scaling.UseRequirementBaseline {
		t.Fatalf("magic_bolt magic scaling = %+v, want capped requirement-based magic scaling", scaling)
	}
	if skill := r.Skills["ice_shard"]; len(skill.Requirements.Skills) != 1 || skill.Requirements.Skills[0].SkillID != magicBoltSkillID || skill.Requirements.Skills[0].Rank != 1 {
		t.Fatalf("ice_shard prerequisites = %+v, want magic_bolt rank 1", skill.Requirements.Skills)
	}
	if skill := r.Skills["ligthing"]; skill.Class != "sorcerer" || skill.Kind != "chain_projectile_attack" || len(skill.Requirements.Skills) != 1 || skill.Requirements.Skills[0].SkillID != magicBoltSkillID || skill.Chain.RangeMultiplier != 0.8 || skill.Projectile.Speed != r.Skills[magicBoltSkillID].Projectile.Speed || skill.Cooldown.FlatTicks != 10 {
		t.Fatalf("ligthing skill = %+v, want sorcerer chain_projectile_attack with magic_bolt prerequisite, magic_bolt speed, 0.8 chain, and +1s cooldown", skill)
	}
	if skill := r.Skills["rage"]; skill.Class != "barbarian" || skill.MaxRank != 5 || skill.Kind != "self_buff" || skill.Targeting != "self" || skill.Requirements.Stats["str"] != 5 || skill.Requirements.Stats["vit"] != 5 || skill.Requirements.StatsPerRank["str"] != 1 || skill.Requirements.StatsPerRank["vit"] != 1 || len(skill.Effects) != 1 || skill.Effects[0].Type != "stat_percent_buff" || skill.Effects[0].DurationTicks != 450 {
		t.Fatalf("rage skill = %+v, want self_buff STR/VIT 5 +1/rank requirements and 450 tick effect", skill)
	}
	if skill := r.Skills["heal"]; skill.Class != "paladin" || skill.MaxRank != 5 || skill.Kind != "area_heal" || skill.Targeting != "direction_or_target_area" || skill.Requirements.Stats["magic"] != 5 || skill.Requirements.StatsPerRank["magic"] != 3 || len(skill.Effects) != 1 || skill.Effects[0].Type != "area_percent_heal" || skill.Effects[0].Range != 9.0 || skill.Effects[0].Radius != 4.0 || skill.Effects[0].DurationTicks != 30 {
		t.Fatalf("heal skill = %+v, want area_heal magic 5 +3/rank requirements and enlarged range/radius effect", skill)
	}
	if scaling := r.Skills["heal"].Effects[0].MagicScaling; scaling.Stat != "magic" || scaling.PercentPerPoint <= 0 || scaling.MaxBonusPercent <= 0 || !scaling.UseRequirementBaseline {
		t.Fatalf("heal magic scaling = %+v, want capped requirement-based magic scaling", scaling)
	}
	if r.Monsters["dungeon_mob"].XPReward <= 0 {
		t.Fatalf("dungeon_mob xp_reward = %d, want positive", r.Monsters["dungeon_mob"].XPReward)
	}
	if _, ok := r.Worlds["combat_control_lab"]; !ok {
		t.Fatal("missing combat_control_lab world")
	}
}

func TestItemTemplateFamilyStatRequirements(t *testing.T) {
	rules := loadRules(t)
	expectedByType := map[string]string{
		"sword":      "str",
		"greatsword": "str",
		"axe":        "str",
		"staff":      "magic",
		"bow":        "dex",
		"shield":     "str",
		"helm":       "str",
		"chest":      "str",
		"gloves":     "str",
		"belt":       "str",
		"boots":      "str",
		"ring":       "magic",
		"amulet":     "magic",
	}

	for templateID, template := range rules.ItemTemplates {
		if !template.Equippable {
			continue
		}
		stat, ok := expectedByType[template.ItemType]
		if !ok {
			t.Fatalf("template %s has unmapped item_type %q", templateID, template.ItemType)
		}
		if template.Requirements["level"] < 1 {
			t.Fatalf("template %s requirements = %+v, want level requirement", templateID, template.Requirements)
		}
		if template.Requirements[stat] < 1 {
			t.Fatalf("template %s item_type %s requirements = %+v, want %s requirement", templateID, template.ItemType, template.Requirements, stat)
		}
	}
}

func TestMainConfigAttackIntervalOverridesCombatMirror(t *testing.T) {
	rules := loadRulesWithMainGameplay(t, map[string]any{
		"base_attack_interval_ticks": 10,
	})
	if rules.Combat.BaseAttackIntervalTicks != 10 {
		t.Fatalf("combat base attack interval = %d, want main_config value 10", rules.Combat.BaseAttackIntervalTicks)
	}
	sim := MustNewSim("sess_main_config_attack", "01", rules)
	expectedInterval := sim.attackIntervalTicksFromSpeed(sim.characterDerivedStatsView().AttackSpeed)
	if got := sim.DerivedStatsView().AttackIntervalTicks; got != expectedInterval {
		t.Fatalf("derived attack interval = %d, want configured interval derived from character speed %d", got, expectedInterval)
	}
}

func TestMainConfigMovementSpeedDrivesPlayerMovement(t *testing.T) {
	rules := loadRulesWithMainGameplay(t, map[string]any{
		"base_movement_speed": 0.5,
	})
	sim, err := NewSimWithWorld("sess_main_config_move", "abcd", rules, "gear_before_combat")
	if err != nil {
		t.Fatalf("gear world: %v", err)
	}
	start := sim.entities[sim.playerID].pos
	sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1, Y: 0}, DurationTicks: 2}}})
	sim.Tick(nil)
	got := sim.entities[sim.playerID].pos
	wantX := start.X + 1.0
	if got.X != wantX || got.Y != start.Y {
		t.Fatalf("player pos = %+v, want x=%v", got, wantX)
	}
}

func TestMonsterRarityGolden(t *testing.T) {
	var golden struct {
		MonsterDefID string `json:"monster_def_id"`
		Rarities     []struct {
			ID                       string  `json:"id"`
			Weight                   int     `json:"weight"`
			Color                    string  `json:"color"`
			HPMultiplier             float64 `json:"hp_multiplier"`
			DamageMultiplier         float64 `json:"damage_multiplier"`
			XPMultiplier             float64 `json:"xp_multiplier"`
			ArmorMultiplier          float64 `json:"armor_multiplier"`
			ArmorBonus               float64 `json:"armor_bonus"`
			HitChanceBonus           float64 `json:"hit_chance_bonus"`
			CritChanceBonus          float64 `json:"crit_chance_bonus"`
			BlockPercentBonus        float64 `json:"block_percent_bonus"`
			AttackCooldownMultiplier float64 `json:"attack_cooldown_multiplier"`
			LootDepthOffset          int     `json:"loot_depth_offset"`
			Expected                 struct {
				MaxHP               int         `json:"max_hp"`
				AttackDamage        DamageRange `json:"attack_damage"`
				Armor               float64     `json:"armor"`
				HitChance           float64     `json:"hit_chance"`
				CritChance          float64     `json:"crit_chance"`
				BlockPercent        float64     `json:"block_percent"`
				AttackCooldownTicks int         `json:"attack_cooldown_ticks"`
				XPReward            int         `json:"xp_reward"`
			} `json:"expected"`
		} `json:"rarities"`
		EffectiveDepthCases []struct {
			Level                    int    `json:"level"`
			Rarity                   string `json:"rarity"`
			ExpectedEffectiveDepth   int    `json:"expected_effective_depth"`
			ExpectedMonsterLootTable string `json:"expected_monster_loot_table"`
		} `json:"effective_depth_cases"`
	}
	loadGolden(t, "monster_rarity.json", &golden)
	rules := loadRules(t)
	base := rules.Monsters[golden.MonsterDefID]
	for _, c := range golden.Rarities {
		rarity, ok := rules.DungeonGeneration.MonsterRarity(c.ID)
		if !ok {
			t.Fatalf("rarity %s missing", c.ID)
		}
		if rarity.Weight != c.Weight || rarity.Color != c.Color || rarity.LootDepthOffset != c.LootDepthOffset {
			t.Fatalf("rarity %s = %+v, want golden weight/color/offset", c.ID, rarity)
		}
		if rarity.ArmorMultiplier != c.ArmorMultiplier || rarity.ArmorBonus != c.ArmorBonus || rarity.HitChanceBonus != c.HitChanceBonus || rarity.CritChanceBonus != c.CritChanceBonus || rarity.BlockPercentBonus != c.BlockPercentBonus || rarity.AttackCooldownMultiplier != c.AttackCooldownMultiplier {
			t.Fatalf("rarity %s stat modifiers = %+v, want golden", c.ID, rarity)
		}
		sim := MustNewSim("sess_monster_rarity_"+c.ID, "monster_rarity_"+c.ID, rules)
		stats := sim.generatedMonsterStats(base, -1-int(rarity.LootDepthOffset), rarity)
		if stats.maxHP != c.Expected.MaxHP {
			t.Fatalf("rarity %s max hp = %d, want %d", c.ID, stats.maxHP, c.Expected.MaxHP)
		}
		if stats.attackDamage == nil || *stats.attackDamage != c.Expected.AttackDamage {
			t.Fatalf("rarity %s damage = %+v, want %+v", c.ID, stats.attackDamage, c.Expected.AttackDamage)
		}
		if stats.armor != c.Expected.Armor || math.Abs(stats.hitChance-c.Expected.HitChance) > 1e-12 || math.Abs(stats.critChance-c.Expected.CritChance) > 1e-12 || stats.blockPercent != c.Expected.BlockPercent || stats.attackCooldown != c.Expected.AttackCooldownTicks {
			t.Fatalf("rarity %s generated stats = %+v, want armor %.3f hit %.3f crit %.3f block %.3f cooldown %d", c.ID, stats, c.Expected.Armor, c.Expected.HitChance, c.Expected.CritChance, c.Expected.BlockPercent, c.Expected.AttackCooldownTicks)
		}
		if stats.xpReward != c.Expected.XPReward {
			t.Fatalf("rarity %s xp = %d, want %d", c.ID, stats.xpReward, c.Expected.XPReward)
		}
	}
	for _, c := range golden.EffectiveDepthCases {
		rarity, ok := rules.DungeonGeneration.MonsterRarity(c.Rarity)
		if !ok {
			t.Fatalf("rarity %s missing", c.Rarity)
		}
		effectiveDepth := absInt(c.Level) + rarity.LootDepthOffset
		if effectiveDepth != c.ExpectedEffectiveDepth {
			t.Fatalf("level %d rarity %s effective depth = %d, want %d", c.Level, c.Rarity, effectiveDepth, c.ExpectedEffectiveDepth)
		}
		band, ok := rules.DungeonGeneration.LootBandForDepth(effectiveDepth)
		if !ok {
			t.Fatalf("effective depth %d has no band", effectiveDepth)
		}
		if band.MonsterLootTable != c.ExpectedMonsterLootTable {
			t.Fatalf("effective depth %d loot table = %s, want %s", effectiveDepth, band.MonsterLootTable, c.ExpectedMonsterLootTable)
		}
	}
}

func TestCoopRewardsAndScalingGolden(t *testing.T) {
	var golden struct {
		XPShare struct {
			Radius                  float64 `json:"radius"`
			FullXPPerEligiblePlayer bool    `json:"full_xp_per_eligible_player"`
			IncludeDeadPlayers      bool    `json:"include_dead_players"`
			IncludeDisconnected     bool    `json:"include_disconnected_players"`
		} `json:"xp_share"`
		PartyChallenge struct {
			PerDoubleBonus float64 `json:"per_double_bonus"`
			MaxBonus       float64 `json:"max_bonus"`
			Cases          []struct {
				PartyCount         int     `json:"party_count"`
				ExpectedMultiplier float64 `json:"expected_multiplier"`
			} `json:"cases"`
		} `json:"party_challenge"`
	}
	loadGolden(t, "coop_rewards_and_scaling.json", &golden)
	rules := loadRules(t)
	if rules.Combat.Coop.XPShare.Radius != golden.XPShare.Radius ||
		rules.Combat.Coop.XPShare.FullXPPerEligiblePlayer != golden.XPShare.FullXPPerEligiblePlayer ||
		rules.Combat.Coop.XPShare.IncludeDeadPlayers != golden.XPShare.IncludeDeadPlayers ||
		rules.Combat.Coop.XPShare.IncludeDisconnected != golden.XPShare.IncludeDisconnected {
		t.Fatalf("xp share rules = %+v, want golden %+v", rules.Combat.Coop.XPShare, golden.XPShare)
	}
	if rules.Combat.Coop.PartyChallenge.PerDoubleBonus != golden.PartyChallenge.PerDoubleBonus ||
		rules.Combat.Coop.PartyChallenge.MaxBonus != golden.PartyChallenge.MaxBonus {
		t.Fatalf("party challenge rules = %+v, want golden %+v", rules.Combat.Coop.PartyChallenge, golden.PartyChallenge)
	}
	for _, c := range golden.PartyChallenge.Cases {
		got := rules.Combat.Coop.PartyChallenge.Multiplier(c.PartyCount)
		if math.Abs(got-c.ExpectedMultiplier) > 1e-12 {
			t.Fatalf("party count %d multiplier = %.15f, want %.15f", c.PartyCount, got, c.ExpectedMultiplier)
		}
	}
}

func TestValidateCoopCombatRules(t *testing.T) {
	rules := loadRules(t).Combat.Coop
	if err := validateCoopCombatRules(rules); err != nil {
		t.Fatalf("default coop rules invalid: %v", err)
	}
	rules.XPShare.Radius = 0
	if err := validateCoopCombatRules(rules); err == nil {
		t.Fatal("zero xp share radius unexpectedly valid")
	}
	rules = loadRules(t).Combat.Coop
	rules.PartyChallenge.MaxBonus = rules.PartyChallenge.PerDoubleBonus / 2
	if err := validateCoopCombatRules(rules); err == nil {
		t.Fatal("max bonus below per-double bonus unexpectedly valid")
	}
}

func TestBossRulesAndGoldens(t *testing.T) {
	rules := loadRules(t)

	var floorGolden struct {
		Seed        string           `json:"seed"`
		Level       int              `json:"level"`
		IsBossFloor bool             `json:"is_boss_floor"`
		FloorSize   DungeonFloorSize `json:"floor_size"`
		Expected    struct {
			BossCount              int    `json:"boss_count"`
			ChestCount             int    `json:"chest_count"`
			StairsDownCount        int    `json:"stairs_down_count"`
			TeleporterCount        int    `json:"teleporter_count"`
			StairsDownInitialState string `json:"stairs_down_initial_state"`
			TeleporterInitialState string `json:"teleporter_initial_state"`
			UnlockedState          string `json:"unlocked_state"`
			LockedReason           string `json:"locked_reason"`
			Boss                   struct {
				TemplateID       string  `json:"template_id"`
				BaseMonsterDefID string  `json:"base_monster_def_id"`
				VisualModel      string  `json:"visual_model"`
				VisualColor      string  `json:"visual_color"`
				VisualScale      float64 `json:"visual_scale"`
			} `json:"boss"`
		} `json:"expected"`
	}
	loadGolden(t, "boss_floor_-5.json", &floorGolden)
	if !floorGolden.IsBossFloor || floorGolden.Level != rules.DungeonGeneration.BossFloor.FirstLevel {
		t.Fatalf("boss floor golden level/classification = %d/%v", floorGolden.Level, floorGolden.IsBossFloor)
	}
	if floorGolden.FloorSize != rules.DungeonGeneration.BossFloor.FloorSize {
		t.Fatalf("boss floor size = %+v, want %+v", floorGolden.FloorSize, rules.DungeonGeneration.BossFloor.FloorSize)
	}
	if floorGolden.Expected.BossCount != 1 || floorGolden.Expected.ChestCount != 1 || floorGolden.Expected.StairsDownCount != 1 || floorGolden.Expected.TeleporterCount != 0 {
		t.Fatalf("boss floor entity counts = %+v", floorGolden.Expected)
	}
	if floorGolden.Expected.StairsDownInitialState != interactableLocked || floorGolden.Expected.TeleporterInitialState != "absent" || floorGolden.Expected.UnlockedState != interactableReady {
		t.Fatalf("boss floor exit states = %+v", floorGolden.Expected)
	}
	if floorGolden.Expected.LockedReason != rules.DungeonGeneration.BossFloor.LockedExitReason {
		t.Fatalf("boss floor locked reason = %s, want %s", floorGolden.Expected.LockedReason, rules.DungeonGeneration.BossFloor.LockedExitReason)
	}
	template, ok := rules.BossTemplates[floorGolden.Expected.Boss.TemplateID]
	if !ok {
		t.Fatalf("boss template %s missing", floorGolden.Expected.Boss.TemplateID)
	}
	if template.BaseMonsterDefID != floorGolden.Expected.Boss.BaseMonsterDefID {
		t.Fatalf("boss base monster = %s, want %s", template.BaseMonsterDefID, floorGolden.Expected.Boss.BaseMonsterDefID)
	}
	if template.Visual.Color != floorGolden.Expected.Boss.VisualColor || template.Visual.Scale != floorGolden.Expected.Boss.VisualScale {
		t.Fatalf("boss visual = %+v, want golden %+v", template.Visual, floorGolden.Expected.Boss)
	}
	if len(template.Visual.ModelPool) == 0 {
		t.Fatalf("boss visual model pool is empty")
	}
	if !stringSliceContains(template.Visual.ModelPool, floorGolden.Expected.Boss.VisualModel) {
		t.Fatalf("boss visual model %s not in pool %+v", floorGolden.Expected.Boss.VisualModel, template.Visual.ModelPool)
	}

	var timelineGolden struct {
		PatternID             string `json:"pattern_id"`
		MinimumTelegraphTicks int    `json:"minimum_telegraph_ticks"`
		Timeline              []struct {
			PhaseIndex    int          `json:"phase_index"`
			Kind          string       `json:"kind"`
			StartTick     int          `json:"start_tick"`
			EndTick       int          `json:"end_tick"`
			DurationTicks int          `json:"duration_ticks"`
			TelegraphType string       `json:"telegraph_type"`
			HitShape      string       `json:"hit_shape"`
			Shape         string       `json:"shape"`
			Radius        float64      `json:"radius"`
			Damage        *DamageRange `json:"damage"`
		} `json:"timeline"`
		CooldownTicks int `json:"cooldown_ticks"`
		DodgeCase     struct {
			PlayerStartsInContact  bool `json:"player_starts_in_contact"`
			BreakContactBeforeTick int  `json:"break_contact_before_tick"`
			ExpectedDamage         int  `json:"expected_damage"`
		} `json:"dodge_case"`
	}
	loadGolden(t, "boss_pattern_timeline.json", &timelineGolden)
	pattern, ok := rules.BossPatterns[timelineGolden.PatternID]
	if !ok {
		t.Fatalf("boss pattern %s missing", timelineGolden.PatternID)
	}
	if pattern.CooldownTicks != timelineGolden.CooldownTicks {
		t.Fatalf("cooldown = %d, want %d", pattern.CooldownTicks, timelineGolden.CooldownTicks)
	}
	if len(pattern.Phases) != len(timelineGolden.Timeline) {
		t.Fatalf("phase count = %d, want %d", len(pattern.Phases), len(timelineGolden.Timeline))
	}
	cursor := 0
	for i, want := range timelineGolden.Timeline {
		got := pattern.Phases[i]
		if want.PhaseIndex != i || got.Kind != want.Kind || got.DurationTicks != want.DurationTicks {
			t.Fatalf("phase %d = %+v, want %+v", i, got, want)
		}
		if want.StartTick != cursor || want.EndTick != cursor+got.DurationTicks-1 {
			t.Fatalf("phase %d bounds = %d..%d, want %d..%d", i, want.StartTick, want.EndTick, cursor, cursor+got.DurationTicks-1)
		}
		if want.TelegraphType != "" && got.TelegraphType != want.TelegraphType {
			t.Fatalf("phase %d telegraph type = %s, want %s", i, got.TelegraphType, want.TelegraphType)
		}
		if want.HitShape != "" && got.HitShape != want.HitShape {
			t.Fatalf("phase %d hit shape = %s, want %s", i, got.HitShape, want.HitShape)
		}
		if want.Shape != "" && got.Shape != want.Shape {
			t.Fatalf("phase %d shape = %s, want %s", i, got.Shape, want.Shape)
		}
		if want.Radius != 0 && got.Radius != want.Radius {
			t.Fatalf("phase %d radius = %v, want %v", i, got.Radius, want.Radius)
		}
		if want.Damage != nil && (got.Damage == nil || *got.Damage != *want.Damage) {
			t.Fatalf("phase %d damage = %+v, want %+v", i, got.Damage, want.Damage)
		}
		cursor += got.DurationTicks
	}
	if !timelineGolden.DodgeCase.PlayerStartsInContact || timelineGolden.DodgeCase.ExpectedDamage != 0 {
		t.Fatalf("invalid dodge case = %+v", timelineGolden.DodgeCase)
	}
	if timelineGolden.DodgeCase.BreakContactBeforeTick >= timelineGolden.Timeline[1].StartTick {
		t.Fatalf("dodge breaks contact at %d, active starts at %d", timelineGolden.DodgeCase.BreakContactBeforeTick, timelineGolden.Timeline[1].StartTick)
	}
}

func TestCharacterProgressionGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		Cases []struct {
			Name                      string        `json:"name"`
			Experience                int           `json:"experience"`
			BaseStats                 BaseStatsView `json:"base_stats"`
			StartingUnspentStatPoints int           `json:"starting_unspent_stat_points"`
			AllocatedStat             string        `json:"allocated_stat"`
			AllocatedPoints           int           `json:"allocated_points"`
			Expected                  struct {
				Level              int              `json:"level"`
				CurrentLevelXP     int              `json:"current_level_xp"`
				NextLevelXP        int              `json:"next_level_xp"`
				UnspentStatPoints  int              `json:"unspent_stat_points"`
				UnspentSkillPoints int              `json:"unspent_skill_points"`
				BaseStats          BaseStatsView    `json:"base_stats"`
				DerivedStats       DerivedStatsView `json:"derived_stats"`
			} `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "character_progression.json", &golden)

	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			sim := MustNewSim("sess_progression_"+tc.Name, "01", rules)
			sim.progression.BaseStats = tc.BaseStats
			sim.progression.UnspentStatPoints = tc.StartingUnspentStatPoints
			res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
			sim.awardExperience(tc.Experience, "corr_progression", &res)
			if tc.AllocatedStat != "" {
				sim.handleAllocateStat(Input{
					MessageID:     "alloc",
					CorrelationID: "corr_alloc",
					Type:          "allocate_stat_intent",
					AllocateStat:  &AllocateStatIntent{Stat: tc.AllocatedStat, Points: tc.AllocatedPoints},
				}, &res)
			}

			view := sim.CharacterProgressionView()
			if view.Level != tc.Expected.Level || view.Experience != tc.Experience || view.UnspentStatPoints != tc.Expected.UnspentStatPoints {
				t.Fatalf("progression = %+v, want level %d exp %d unspent %d", view, tc.Expected.Level, tc.Experience, tc.Expected.UnspentStatPoints)
			}
			if view.UnspentSkillPoints != tc.Expected.UnspentSkillPoints {
				t.Fatalf("unspent skill points = %d, want %d; progression=%+v", view.UnspentSkillPoints, tc.Expected.UnspentSkillPoints, view)
			}
			if view.ExperienceToNextLevel == nil || *view.ExperienceToNextLevel != tc.Expected.NextLevelXP-tc.Expected.CurrentLevelXP {
				t.Fatalf("experience_to_next_level = %v, want %d", view.ExperienceToNextLevel, tc.Expected.NextLevelXP-tc.Expected.CurrentLevelXP)
			}
			if view.BaseStats != tc.Expected.BaseStats {
				t.Fatalf("base stats = %+v, want %+v", view.BaseStats, tc.Expected.BaseStats)
			}
			assertDerivedStats(t, view.DerivedStats, tc.Expected.DerivedStats)
		})
	}
}

func TestExperienceGainAndLevelUpFromMonsterKill(t *testing.T) {
	rules := cloneRules(loadRules(t))
	def := rules.Monsters["dungeon_mob"]
	def.XPReward = 20
	rules.Monsters["dungeon_mob"] = def
	sim := MustNewSim("sess_xp_kill", "01", rules)
	player := sim.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 0.5, Y: player.pos.Y},
		hp:           1,
		maxHP:        1,
		monsterDefID: "dungeon_mob",
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster

	res := sim.Tick([]Input{{MessageID: "kill_xp", CorrelationID: "corr_xp", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, res, "kill_xp")
	if !hasEvent(res, "monster_killed") || !hasEvent(res, "experience_gained") || !hasEvent(res, "character_leveled") {
		t.Fatalf("missing kill/xp/level events: %+v", res.Events)
	}
	view := sim.CharacterProgressionView()
	if view.Experience != 20 || view.Level != 2 || view.UnspentStatPoints != 3 {
		t.Fatalf("progression after kill = %+v, want exp 20 level 2 unspent 3", view)
	}
	if !hasProgressionChange(res) {
		t.Fatalf("missing progression update change: %+v", res.Changes)
	}

	reject := sim.Tick([]Input{{MessageID: "kill_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, reject, "kill_again", "invalid_target")
	if sim.CharacterProgressionView().Experience != 20 {
		t.Fatalf("dead monster granted XP twice: %+v", sim.CharacterProgressionView())
	}
}

func TestCharacterProgressionCapsAtLevel100(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_level_cap", "01", rules)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	requiredXP := rules.CharacterProgression.XPThresholds[rules.CharacterProgression.LevelCap-1]

	sim.awardExperience(requiredXP+1000, "corr_level_cap", &res)

	view := sim.CharacterProgressionView()
	if view.Level != 100 || view.LevelCap != 100 {
		t.Fatalf("progression level cap = %+v, want level 100 cap 100", view)
	}
	if view.ExperienceToNextLevel != nil {
		t.Fatalf("experience_to_next_level = %v, want nil at cap", view.ExperienceToNextLevel)
	}
}

func TestStatAllocationVitHPAndRejects(t *testing.T) {
	sim := MustNewSim("sess_stat_alloc", "01", loadRules(t))
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.awardExperience(20, "corr_xp", &res)
	player := sim.entities[sim.playerID]
	player.hp = 7

	sim.handleAllocateStat(Input{MessageID: "vit", CorrelationID: "corr_vit", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "vit", Points: 1}}, &res)
	assertAck(t, res, "vit")
	view := sim.CharacterProgressionView()
	if view.BaseStats.Vit != 6 || view.UnspentStatPoints != 2 || player.maxHP != 11 || player.hp != 8 {
		t.Fatalf("vit allocation progression=%+v player=%+v, want vit 6 unspent 2 hp 8/11", view, player)
	}
	if !hasEvent(res, "stat_allocated") || !hasProgressionChange(res) {
		t.Fatalf("missing stat allocation outputs: changes=%+v events=%+v", res.Changes, res.Events)
	}

	overspend := sim.Tick([]Input{{MessageID: "overspend", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "str", Points: 99}}})
	assertReject(t, overspend, "overspend", "not_enough_stat_points")
	invalid := sim.Tick([]Input{{MessageID: "invalid_stat", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "luck", Points: 1}}})
	assertReject(t, invalid, "invalid_stat", "invalid_payload")
}

func TestSkillPointCadenceAndSpend(t *testing.T) {
	sim := MustNewSim("sess_skill_points", "01", loadRules(t))
	sim.progression.CharacterClass = "sorcerer"
	sim.savePlayer(sim.defaultPlayer())
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.awardExperience(157, "corr_skill_xp", &res)
	sim.savePlayer(sim.defaultPlayer())

	view := sim.CharacterProgressionView()
	if view.Level != 5 || view.UnspentStatPoints != 12 || view.UnspentSkillPoints != 2 {
		t.Fatalf("progression after level 5 = %+v, want level 5, 12 stat points, 2 skill points", view)
	}
	if !hasEvent(res, "skill_point_gained") {
		t.Fatalf("missing skill_point_gained event: %+v", res.Events)
	}
	skillView := sim.SkillProgressionView()
	magicSkill, ok := skillProgressionRow(skillView, magicBoltSkillID)
	if !ok || magicSkill.Rank != 0 || !magicSkill.CanSpend {
		t.Fatalf("skill progression at baseline requirements = %+v, want rank 0 and spendable", skillView)
	}

	spend := sim.Tick([]Input{{
		MessageID:          "spend_magic",
		CorrelationID:      "corr_spend",
		Type:               "allocate_skill_point_intent",
		AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: magicBoltSkillID},
	}})
	assertAck(t, spend, "spend_magic")
	if !hasEvent(spend, "skill_rank_updated") || skillProgressionUpdate(spend) == nil {
		t.Fatalf("missing skill spend outputs: changes=%+v events=%+v", spend.Changes, spend.Events)
	}
	skillView = sim.SkillProgressionView()
	magicSkill, _ = skillProgressionRow(skillView, magicBoltSkillID)
	if skillView.UnspentSkillPoints != 1 || magicSkill.Rank != 1 || magicSkill.CanSpend {
		t.Fatalf("skill progression after spend = %+v, want rank 1, 1 point, and blocked rank 2 requirements", skillView)
	}

	overspend := sim.Tick([]Input{{
		MessageID:          "overspend_skill",
		Type:               "allocate_skill_point_intent",
		AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: magicBoltSkillID},
	}})
	assertReject(t, overspend, "overspend_skill", "skill_requirements_not_met")
	magicSkill, _ = skillProgressionRow(sim.SkillProgressionView(), magicBoltSkillID)
	if magicSkill.Rank != 1 {
		t.Fatalf("overspend mutated rank: %+v", sim.SkillProgressionView())
	}

	rankTwoXP := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.awardExperience(76, "corr_rank_two_xp", &rankTwoXP)
	sim.savePlayer(sim.defaultPlayer())
	view = sim.CharacterProgressionView()
	if view.Level != 6 || view.UnspentStatPoints != 15 || view.UnspentSkillPoints != 1 {
		t.Fatalf("progression after level 6 = %+v, want level 6, 15 stat points, 1 skill point", view)
	}
	skillView = sim.SkillProgressionView()
	magicSkill, _ = skillProgressionRow(skillView, magicBoltSkillID)
	if magicSkill.CanSpend {
		t.Fatalf("rank 2 skill progression at magic 5 = %+v, want blocked until magic 8", skillView)
	}
	rankTwoRejected := sim.Tick([]Input{{
		MessageID:          "spend_rank_two_rejected",
		CorrelationID:      "corr_spend_rank_two_rejected",
		Type:               "allocate_skill_point_intent",
		AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: magicBoltSkillID},
	}})
	assertReject(t, rankTwoRejected, "spend_rank_two_rejected", "skill_requirements_not_met")
	magicSkill, _ = skillProgressionRow(sim.SkillProgressionView(), magicBoltSkillID)
	if magicSkill.Rank != 1 || sim.progression.UnspentSkillPoints != 1 {
		t.Fatalf("rank 2 requirement rejection mutated skill progression: %+v", sim.SkillProgressionView())
	}

	allocateRankTwoMagic := sim.Tick([]Input{{
		MessageID:     "allocate_rank_two_magic",
		CorrelationID: "corr_allocate_rank_two_magic",
		Type:          "allocate_stat_intent",
		AllocateStat:  &AllocateStatIntent{Stat: "magic", Points: 3},
	}})
	assertAck(t, allocateRankTwoMagic, "allocate_rank_two_magic")
	if sim.progression.BaseStats.Magic != 8 || sim.progression.UnspentStatPoints != 12 {
		t.Fatalf("rank 2 magic allocation progression = %+v, want magic 8 and 12 unspent stat points", sim.CharacterProgressionView())
	}
	skillView = sim.SkillProgressionView()
	magicSkill, _ = skillProgressionRow(skillView, magicBoltSkillID)
	if !magicSkill.CanSpend {
		t.Fatalf("rank 2 skill progression after magic 8 = %+v, want spendable", skillView)
	}
	rankTwoSpend := sim.Tick([]Input{{
		MessageID:          "spend_rank_two",
		CorrelationID:      "corr_spend_rank_two",
		Type:               "allocate_skill_point_intent",
		AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: magicBoltSkillID},
	}})
	assertAck(t, rankTwoSpend, "spend_rank_two")
	skillView = sim.SkillProgressionView()
	magicSkill, _ = skillProgressionRow(skillView, magicBoltSkillID)
	if skillView.UnspentSkillPoints != 0 || magicSkill.Rank != 2 || magicSkill.CanSpend {
		t.Fatalf("skill progression after rank 2 spend = %+v, want rank 2 and no spendable points", skillView)
	}
}

func TestSetSkillBindingsIntentUpdatesSnapshotAndDelta(t *testing.T) {
	sim := MustNewSim("sess_skill_bindings", "01", loadRules(t))
	update := sim.Tick([]Input{{
		MessageID: "bind_skills",
		Type:      "set_skill_bindings_intent",
		SetSkillBindings: &SetSkillBindingsIntent{
			FunctionKeys:      []string{magicBoltSkillID, "heal"},
			RightClickSkillID: "heal",
		},
	}})
	assertAck(t, update, "bind_skills")
	var got *SkillBindingsView
	for i := range update.Changes {
		if update.Changes[i].Op == OpSkillBindingsUpdate {
			got = update.Changes[i].SkillBindings
			break
		}
	}
	if got == nil || len(got.FunctionKeys) != skillFunctionKeyCount || got.FunctionKeys[0] != magicBoltSkillID || got.FunctionKeys[1] != "heal" || got.RightClickSkillID != "heal" {
		t.Fatalf("skill binding update = %+v changes=%+v", got, update.Changes)
	}
	snap := sim.Snapshot()
	if len(snap.SkillBindings.FunctionKeys) != skillFunctionKeyCount || snap.SkillBindings.FunctionKeys[0] != magicBoltSkillID || snap.SkillBindings.RightClickSkillID != "heal" {
		t.Fatalf("snapshot skill bindings = %+v", snap.SkillBindings)
	}

	rejected := sim.Tick([]Input{{
		MessageID: "bad_bind",
		Type:      "set_skill_bindings_intent",
		SetSkillBindings: &SetSkillBindingsIntent{
			FunctionKeys: []string{"not_a_skill"},
		},
	}})
	assertReject(t, rejected, "bad_bind", "unknown_skill")
}

func TestLoadInventoryAppliesEquippedResourceStats(t *testing.T) {
	sim := MustNewSim("sess_resume_equipped_stats", "01", loadRules(t))
	player := sim.entities[sim.playerID]
	if player.maxHP != 10 || player.hp != 10 {
		t.Fatalf("base player hp = %d/%d, want 10/10", player.hp, player.maxHP)
	}
	payload := ItemRollPayload{
		ItemTemplateID: "cave_ring",
		DisplayName:    "Resume Ring",
		Rarity:         "test",
		Stats:          map[string]int{"max_hp": 4},
		Requirements:   map[string]int{"level": 1},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	sim.LoadInventory([]PersistedItem{{
		InstanceID:  "6415",
		ItemDefID:   "cave_ring",
		Slot:        ringLeftSlot,
		Equipped:    true,
		RolledStats: raw,
	}})

	player = sim.entities[sim.playerID]
	if player.maxHP != 14 || player.hp != 14 {
		t.Fatalf("loaded equipped ring hp = %d/%d, want 14/14", player.hp, player.maxHP)
	}
	snap := sim.Snapshot()
	var playerView *EntityView
	for i := range snap.Entities {
		if snap.Entities[i].ID == idStr(sim.playerID) {
			playerView = &snap.Entities[i]
			break
		}
	}
	if playerView == nil || playerView.MaxHP == nil || *playerView.MaxHP != 14 || playerView.HP == nil || *playerView.HP != 14 {
		t.Fatalf("snapshot player after load = %+v, want hp/maxHP 14/14", playerView)
	}
	if snap.CharacterProgression.DerivedStats.MaxHP != 14 {
		t.Fatalf("snapshot derived max hp = %v, want 14", snap.CharacterProgression.DerivedStats.MaxHP)
	}
}

func TestClassSkillAccessGatesSpendabilityAndLearning(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		classID string
		allowed string
		blocked string
	}{
		{classID: "barbarian", allowed: "rage", blocked: magicBoltSkillID},
		{classID: "sorcerer", allowed: magicBoltSkillID, blocked: "rage"},
		{classID: "paladin", allowed: "heal", blocked: magicBoltSkillID},
	}
	for _, c := range cases {
		t.Run(c.classID, func(t *testing.T) {
			state := rules.DefaultCharacterProgressionState()
			state.CharacterClass = c.classID
			state.UnspentSkillPoints = 2
			classDef := rules.CharacterProgression.Classes[c.classID]
			state.BaseStats = classDef.BaseStats
			sim, err := NewSimWithWorldProgression("sess_class_skill_"+c.classID, "class_skill_seed", rules, DefaultWorldID, state)
			if err != nil {
				t.Fatalf("new sim: %v", err)
			}
			allowed, ok := skillProgressionRow(sim.SkillProgressionView(), c.allowed)
			if !ok || !allowed.CanSpend {
				t.Fatalf("%s allowed skill row = %+v ok=%v, want spendable", c.classID, allowed, ok)
			}
			blocked, ok := skillProgressionRow(sim.SkillProgressionView(), c.blocked)
			if !ok || blocked.CanSpend {
				t.Fatalf("%s blocked skill row = %+v ok=%v, want not spendable", c.classID, blocked, ok)
			}
			rejected := sim.Tick([]Input{{
				MessageID:          "spend_blocked",
				CorrelationID:      "corr_spend_blocked",
				Type:               "allocate_skill_point_intent",
				AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: c.blocked},
			}})
			assertReject(t, rejected, "spend_blocked", "skill_class_not_allowed")
			if sim.progression.SkillRanks[c.blocked] != 0 {
				t.Fatalf("blocked skill rank mutated: %+v", sim.progression.SkillRanks)
			}
			sim.progression.SkillRanks[c.blocked] = 1
			castRejected := sim.Tick([]Input{{
				MessageID:     "cast_blocked",
				CorrelationID: "corr_cast_blocked",
				Type:          "cast_skill_intent",
				CastSkill:     &CastSkillIntent{SkillID: c.blocked, Direction: &Vec2{X: 1}},
			}})
			assertReject(t, castRejected, "cast_blocked", "skill_class_not_allowed")
			sim.progression.SkillRanks[c.blocked] = 0
			spend := sim.Tick([]Input{{
				MessageID:          "spend_allowed",
				CorrelationID:      "corr_spend_allowed",
				Type:               "allocate_skill_point_intent",
				AllocateSkillPoint: &AllocateSkillPointIntent{SkillID: c.allowed},
			}})
			assertAck(t, spend, "spend_allowed")
			if sim.progression.SkillRanks[c.allowed] != 1 {
				t.Fatalf("allowed skill rank = %d, want 1", sim.progression.SkillRanks[c.allowed])
			}
		})
	}
}

func TestClassRequiredWeaponsGateEquip(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_class_weapon_barbarian", "class_weapon_seed", rules, DefaultWorldID, state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	sim.inventory = append(sim.inventory,
		&invItem{instanceID: 2001, itemDefID: "sorcerer_staff"},
		&invItem{instanceID: 2002, itemDefID: "barbarian_axe"},
	)
	sim.savePlayer(sim.defaultPlayer())
	rejected := sim.Tick([]Input{{
		MessageID:     "equip_staff",
		CorrelationID: "corr_equip_staff",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: "2001", Slot: mainHandSlot},
	}})
	assertReject(t, rejected, "equip_staff", "class_requirement_not_met")
	if sim.equipped[mainHandSlot] != 0 {
		t.Fatalf("blocked class weapon equipped: %+v", sim.equipped)
	}
	equipped := sim.Tick([]Input{{
		MessageID:     "equip_axe",
		CorrelationID: "corr_equip_axe",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: "2002", Slot: mainHandSlot},
	}})
	assertAck(t, equipped, "equip_axe")
	if sim.equipped[mainHandSlot] != 2002 {
		t.Fatalf("class weapon equipped slot = %d, want 2002", sim.equipped[mainHandSlot])
	}
}

func TestMagicBoltCastCooldownAndProjectileDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	zero := 0.0
	crit := rules.CharacterProgression.DerivedStats["crit_chance"]
	crit.Base = 0
	crit.PerDex = 0
	crit.Min = &zero
	crit.Max = &zero
	rules.CharacterProgression.DerivedStats["crit_chance"] = crit

	sim := MustNewSim("sess_magic_bolt", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 15
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 5, Y: player.pos.Y},
		hp:           20,
		maxHP:        20,
		monsterDefID: monsterDefID,
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster

	cast := sim.Tick([]Input{{
		MessageID:     "cast_magic",
		CorrelationID: "corr_cast",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(monster.id)},
	}})
	assertAck(t, cast, "cast_magic")
	if player.mana != 9 {
		t.Fatalf("mana after cast = %d, want 9", player.mana)
	}
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("missing cast/cooldown events: %+v", cast.Events)
	}
	cooldowns := skillCooldownUpdate(cast)
	expectedCooldown := sim.skillCooldownTicks(rules.Skills[magicBoltSkillID])
	if len(cooldowns) != 1 || cooldowns[0].SkillID != magicBoltSkillID || cooldowns[0].RemainingTicks != expectedCooldown || cooldowns[0].TotalTicks != expectedCooldown {
		t.Fatalf("cooldown update = %+v, want magic_bolt %d/%d", cooldowns, expectedCooldown, expectedCooldown)
	}
	spawn := firstChangeEntityByType(cast, projectileEntity)
	if spawn == nil || spawn.ProjectileDefID != magicBoltSkillID || spawn.TargetID != idStr(monster.id) {
		t.Fatalf("magic bolt projectile spawn = %+v", spawn)
	}

	recast := sim.Tick([]Input{{
		MessageID: "recast_magic",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(monster.id)},
	}})
	assertReject(t, recast, "recast_magic", "skill_on_cooldown")
	if player.mana != 9 {
		t.Fatalf("recast spent mana: %d", player.mana)
	}
	if !hasEvent(recast, "skill_cooldown_rejected") {
		t.Fatalf("missing cooldown rejected event: %+v", recast.Events)
	}

	var impact TickResult
	for i := 0; i < 30; i++ {
		impact = sim.Tick(nil)
		if hasEvent(impact, "monster_damaged") {
			break
		}
	}
	damage := eventDamage(impact, "monster_damaged")
	if damage < 6 || damage > 9 {
		t.Fatalf("magic bolt impact damage = %d, want rank 1 damage 6..9; events=%+v", damage, impact.Events)
	}

	for i := 0; i < 50; i++ {
		if _, active := sim.skillCooldownRemaining(magicBoltSkillID); !active {
			break
		}
		sim.Tick(nil)
	}
	if remaining, active := sim.skillCooldownRemaining(magicBoltSkillID); active {
		t.Fatalf("cooldown still active with %d ticks remaining", remaining)
	}
	second := sim.Tick([]Input{{
		MessageID: "cast_magic_again",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: magicBoltSkillID, Direction: &Vec2{X: 1}},
	}})
	assertAck(t, second, "cast_magic_again")
}

func TestCleaveDamagesConeTargetsAndPushes(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_cleave", "01", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.BaseStats.Str = 10
	sim.progression.BaseStats.Vit = 10
	sim.progression.SkillRanks["cleave"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 20, Y: 20}
	player.mana = player.maxMana
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	front := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 1, Y: player.pos.Y}, hp: 20, maxHP: 20, monsterDefID: monsterDefID, lootTable: "no_drop"}
	side := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 2, Y: player.pos.Y + 2}, hp: 20, maxHP: 20, monsterDefID: monsterDefID, lootTable: "no_drop"}
	sim.entities[front.id] = front
	sim.entities[side.id] = side
	beforeX := front.pos.X

	cast := sim.Tick([]Input{{
		MessageID:     "cast_cleave",
		CorrelationID: "corr_cleave",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "cleave", Direction: &Vec2{X: 1, Y: 0}},
	}})
	assertAck(t, cast, "cast_cleave")
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "monster_damaged") || !hasEvent(cast, "monster_pushed") {
		t.Fatalf("cleave missing cast/damage/push events: %+v", cast.Events)
	}
	if front.hp >= 20 {
		t.Fatalf("front monster hp = %d, want damaged", front.hp)
	}
	if front.pos.X <= beforeX {
		t.Fatalf("front monster x = %.2f, want pushed beyond %.2f", front.pos.X, beforeX)
	}
	if side.hp != 20 {
		t.Fatalf("side monster hp = %d, want untouched outside 50 degree cone", side.hp)
	}
	var castEvent *Event
	for i := range cast.Events {
		if cast.Events[i].EventType == "skill_cast" {
			castEvent = &cast.Events[i]
			break
		}
	}
	if castEvent == nil || castEvent.Direction == nil || castEvent.Range == nil || castEvent.AngleDegrees == nil {
		t.Fatalf("cleave cast event missing cone geometry: %+v", cast.Events)
	}
}

func TestIceShardAlwaysHitsAppliesSlowAndSpawnsShards(t *testing.T) {
	rules := cloneRules(loadRules(t))
	skill := rules.Skills["ice_shard"]
	skill.Damage.MinBase = 8
	skill.Damage.MaxBase = 8
	skill.Shatter.MinShards = 3
	skill.Shatter.MaxShards = 3
	rules.Skills["ice_shard"] = skill
	sim := MustNewSim("sess_ice_shard", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 15
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	sim.progression.SkillRanks["ice_shard"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	monster := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 5, Y: player.pos.Y}, hp: 30, maxHP: 30, monsterDefID: "combat_lab_blocking_target", lootTable: "no_drop"}
	sim.entities[monster.id] = monster

	cast := sim.Tick([]Input{{
		MessageID:     "cast_ice",
		CorrelationID: "corr_ice",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "ice_shard", TargetID: idStr(monster.id)},
	}})
	assertAck(t, cast, "cast_ice")
	var impact TickResult
	impactEvents := []Event{}
	impactChanges := []Change{}
	for i := 0; i < 30; i++ {
		impact = sim.Tick(nil)
		impactEvents = append(impactEvents, impact.Events...)
		impactChanges = append(impactChanges, impact.Changes...)
		if hasEvent(impact, "skill_effect_started") {
			break
		}
	}
	if !eventListHas(impactEvents, "monster_damaged") || !eventListHas(impactEvents, "skill_effect_started") {
		t.Fatalf("ice shard impact missing damage/slow: %+v", impactEvents)
	}
	if !sameStringSlice(monster.effectIDs, []string{"ice_slow"}) {
		t.Fatalf("monster effect ids = %v, want ice_slow", monster.effectIDs)
	}
	shards := 0
	originOverlap := 0
	for _, change := range impactChanges {
		if change.Op == OpEntitySpawn && change.Entity != nil && change.Entity.Type == projectileEntity && change.Entity.ProjectileDefID == "ice_shard_shard" {
			shards++
			if distance(change.Entity.Position, monster.pos) <= monsterRadius+projectileRadius {
				originOverlap++
			}
		}
	}
	if shards != 3 {
		t.Fatalf("ice shard spawned %d shards, want 3; changes=%+v", shards, impactChanges)
	}
	if originOverlap != 0 {
		t.Fatalf("ice shard spawned %d shards inside impact target radius; changes=%+v", originOverlap, impactChanges)
	}
}

func TestLigthingChainsToNearestTargetsWithShrinkingRange(t *testing.T) {
	rules := cloneRules(loadRules(t))
	skill := rules.Skills["ligthing"]
	skill.Damage.MinBase = 5
	skill.Damage.MaxBase = 5
	skill.Chain.RangeMultiplier = 0.8
	skill.Chain.MaxJumps = 8
	rules.Skills["ligthing"] = skill
	sim := MustNewSim("sess_ligthing", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 15
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	sim.progression.SkillRanks["ligthing"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	for id, e := range sim.entities {
		if e != nil && e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	first := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 5, Y: player.pos.Y}, hp: 30, maxHP: 30, monsterDefID: monsterDefID, lootTable: "no_drop"}
	second := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 12, Y: player.pos.Y}, hp: 30, maxHP: 30, monsterDefID: monsterDefID, lootTable: "no_drop"}
	third := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 20, Y: player.pos.Y}, hp: 30, maxHP: 30, monsterDefID: monsterDefID, lootTable: "no_drop"}
	tooFar := &entity{id: sim.alloc(), kind: monsterEntity, pos: Vec2{X: player.pos.X + 30, Y: player.pos.Y}, hp: 30, maxHP: 30, monsterDefID: monsterDefID, lootTable: "no_drop"}
	for _, monster := range []*entity{first, second, third, tooFar} {
		sim.entities[monster.id] = monster
	}

	cast := sim.Tick([]Input{{
		MessageID:     "cast_ligthing",
		CorrelationID: "corr_ligthing",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "ligthing", TargetID: idStr(first.id)},
	}})
	assertAck(t, cast, "cast_ligthing")
	cooldowns := skillCooldownUpdate(cast)
	expectedCooldown := sim.skillCooldownTicks(rules.Skills["ligthing"])
	if len(cooldowns) != 1 || cooldowns[0].SkillID != "ligthing" || cooldowns[0].TotalTicks != expectedCooldown {
		t.Fatalf("ligthing cooldown update = %+v, want ligthing total %d", cooldowns, expectedCooldown)
	}
	var impactEvents []Event
	for i := 0; i < 30; i++ {
		res := sim.Tick(nil)
		impactEvents = append(impactEvents, res.Events...)
		if eventListHas(res.Events, "skill_chain_hit") {
			break
		}
	}
	damageByTarget := map[string]int{}
	chainHits := 0
	for _, ev := range impactEvents {
		if ev.EventType == "monster_damaged" && ev.TargetEntityID != "" && ev.Damage != nil {
			damageByTarget[ev.TargetEntityID] += *ev.Damage
		}
		if ev.EventType == "skill_chain_hit" {
			chainHits++
		}
	}
	if chainHits != 2 {
		t.Fatalf("ligthing chain hits = %d, want 2; events=%+v", chainHits, impactEvents)
	}
	for _, monster := range []*entity{first, second, third} {
		if damageByTarget[idStr(monster.id)] == 0 || monster.hp >= monster.maxHP {
			t.Fatalf("monster %d damage=%d hp=%d/%d events=%+v", monster.id, damageByTarget[idStr(monster.id)], monster.hp, monster.maxHP, impactEvents)
		}
	}
	if damageByTarget[idStr(tooFar.id)] != 0 || tooFar.hp != tooFar.maxHP {
		t.Fatalf("too-far monster was chained: damage=%d hp=%d/%d events=%+v", damageByTarget[idStr(tooFar.id)], tooFar.hp, tooFar.maxHP, impactEvents)
	}
}

func TestMagicBoltAutoNavigatesToCastRange(t *testing.T) {
	rules := cloneRules(loadRules(t))
	skill := rules.Skills[magicBoltSkillID]
	skill.Projectile.Range = 4
	rules.Skills[magicBoltSkillID] = skill

	sim := MustNewSim("sess_magic_bolt_auto_nav", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats.Magic = 15
	sim.progression.SkillRanks[magicBoltSkillID] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 2, Y: 5}
	for id, e := range sim.entities {
		if e.kind == monsterEntity {
			delete(sim.entities, id)
		}
	}
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: 12, Y: 5},
		hp:           20,
		maxHP:        20,
		monsterDefID: monsterDefID,
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster

	cast := sim.Tick([]Input{{
		MessageID:     "cast_far_magic",
		CorrelationID: "corr_far_magic",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(monster.id)},
	}})
	assertAck(t, cast, "cast_far_magic")
	if firstChangeEntityByType(cast, projectileEntity) != nil {
		t.Fatal("far magic bolt spawned before auto-navigation completed")
	}
	if player.mana != player.maxMana {
		t.Fatalf("far magic bolt spent mana before reaching cast range: mana=%d max=%d", player.mana, player.maxMana)
	}
	nav := sim.activeLevel().autoNav
	if nav == nil || nav.pendingSkill == nil || nav.pendingSkill.SkillID != magicBoltSkillID {
		t.Fatalf("far magic bolt did not queue pending skill auto-nav: %+v", nav)
	}

	sawProjectile := false
	var castDistance float64
	for i := 0; i < 100 && !sawProjectile; i++ {
		r := sim.Tick(nil)
		for _, c := range r.Changes {
			if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == projectileEntity {
				sawProjectile = true
				castDistance = distance(c.Entity.Position, monster.pos)
				if c.Entity.ProjectileDefID != magicBoltSkillID || c.Entity.TargetID != idStr(monster.id) {
					t.Fatalf("auto-nav magic bolt projectile = %+v", c.Entity)
				}
			}
		}
	}
	if !sawProjectile {
		t.Fatal("auto-navigated magic bolt never spawned projectile")
	}
	if castDistance > skill.Projectile.Range+meleeRangeEpsilon {
		t.Fatalf("magic bolt cast distance = %.3f, want within range %.3f", castDistance, skill.Projectile.Range)
	}
	minExpected := skill.Projectile.Range - rules.Navigation.CellSize - meleeRangeEpsilon
	if castDistance < minExpected {
		t.Fatalf("magic bolt cast distance = %.3f, want farthest reachable point near range >= %.3f", castDistance, minExpected)
	}
}

func TestMagicBoltCastRequiresMagicRequirement(t *testing.T) {
	sim := MustNewSim("sess_magic_bolt_requirement", "01", loadRules(t))
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.SkillRanks[magicBoltSkillID] = 2
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 5, Y: player.pos.Y},
		hp:           20,
		maxHP:        20,
		monsterDefID: monsterDefID,
		lootTable:    "no_drop",
	}
	sim.entities[monster.id] = monster

	beforeMana := player.mana
	cast := sim.Tick([]Input{{
		MessageID: "cast_requirement",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: magicBoltSkillID, TargetID: idStr(monster.id)},
	}})
	assertReject(t, cast, "cast_requirement", "skill_requirements_not_met")
	if player.mana != beforeMana {
		t.Fatalf("requirement rejection spent mana: got %d want %d", player.mana, beforeMana)
	}
	if len(sim.SkillCooldownViews()) != 0 {
		t.Fatalf("requirement rejection started cooldown: %+v", sim.SkillCooldownViews())
	}
	if firstChangeEntityByType(cast, projectileEntity) != nil || hasEvent(cast, "monster_damaged") {
		t.Fatalf("requirement rejection spawned/damaged: changes=%+v events=%+v", cast.Changes, cast.Events)
	}
}

func TestRageBuffAppliesStatsVisualScaleAndExpires(t *testing.T) {
	sim := MustNewSim("sess_rage", "01", loadRules(t))
	player := sim.entities[sim.playerID]
	sim.progression.BaseStats.Str = 10
	sim.progression.BaseStats.Vit = 10
	sim.progression.SkillRanks["rage"] = 1
	player.maxHP = sim.currentMaxHP()
	player.hp = player.maxHP
	player.mana = player.maxMana
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{
		MessageID:     "cast_rage",
		CorrelationID: "corr_rage",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "rage", Direction: &Vec2{X: 1}},
	}})
	assertAck(t, cast, "cast_rage")
	if player.mana != 5 {
		t.Fatalf("rage mana after cast = %d, want 5", player.mana)
	}
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_effect_started") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("rage missing cast/effect/cooldown events: %+v", cast.Events)
	}
	if player.maxHP != 16 || player.hp != 15 {
		t.Fatalf("rage hp/max = %d/%d, want 15/16 without free heal", player.hp, player.maxHP)
	}
	if math.Abs(player.visualScale-1.1) > 0.000001 {
		t.Fatalf("rage visual scale = %v, want 1.1", player.visualScale)
	}
	if got := int(math.Round(sim.DerivedStatsView().MaxHP)); got != 16 {
		t.Fatalf("rage derived max hp = %d, want 16", got)
	}

	var expired TickResult
	for i := 0; i < 450; i++ {
		expired = sim.Tick(nil)
	}
	if !hasEvent(expired, "skill_effect_ended") {
		t.Fatalf("rage expiry missing event: %+v", expired.Events)
	}
	if player.maxHP != 15 || player.visualScale != 0 {
		t.Fatalf("rage after expiry hp max/scale = %d/%v, want max 15 and no scale", player.maxHP, player.visualScale)
	}
}

func TestHealAreaSkillHealsAlliesAndAllowsFullHPNoop(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_heal_skill", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.savePlayer(sim.defaultPlayer())
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	sim.usePlayer(sim.players[hostID])
	player := sim.entities[hostID]
	guest := sim.entities[guestID]
	guest.pos = Vec2{X: player.pos.X + 1, Y: player.pos.Y}
	sim.progression.BaseStats.Magic = 10
	sim.progression.SkillRanks["heal"] = 1
	player.hp = 5
	guest.hp = 3
	player.mana = player.maxMana
	sim.savePlayer(sim.players[hostID])

	cast := sim.Tick([]Input{{
		MessageID:     "cast_heal",
		CorrelationID: "corr_heal",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "heal", TargetID: idStr(hostID)},
	}})
	assertAck(t, cast, "cast_heal")
	wantMana := player.maxMana - skillManaCost(rules.Skills["heal"], 1)
	if player.mana != wantMana {
		t.Fatalf("heal mana after cast = %d, want %d", player.mana, wantMana)
	}
	if player.hp != 7 || guest.hp != 5 {
		t.Fatalf("heal hp host/guest = %d/%d, want 7/5", player.hp, guest.hp)
	}
	healed := 0
	for _, ev := range cast.Events {
		if ev.EventType == "player_healed" {
			healed++
			if ev.SkillID != "heal" || ev.Heal == nil || *ev.Heal != 2 {
				t.Fatalf("heal event = %+v, want skill heal for 2", ev)
			}
		}
	}
	if healed != 2 || !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("heal events = %+v, want two heals plus cast/cooldown", cast.Events)
	}
	guest.pos = Vec2{X: player.pos.X + 10, Y: player.pos.Y}
	guest.hp = 3
	for i := 0; i < healRainPulseIntervalTicks-1; i++ {
		noPulse := sim.Tick(nil)
		if hasEvent(noPulse, "player_healed") {
			t.Fatalf("heal rain pulsed before one second: tick=%d events=%+v", noPulse.Tick, noPulse.Events)
		}
	}
	guest.pos = Vec2{X: player.pos.X + 1, Y: player.pos.Y}
	pulse := sim.Tick(nil)
	if player.hp != 9 || guest.hp != 5 {
		t.Fatalf("heal rain first pulse hp host/guest = %d/%d, want 9/5", player.hp, guest.hp)
	}
	if countEvents(pulse, "player_healed") != 2 {
		t.Fatalf("heal rain first pulse events = %+v, want two player_healed", pulse.Events)
	}
	guest.pos = Vec2{X: player.pos.X + 10, Y: player.pos.Y}
	guest.hp = 3
	for i := 0; i < healRainPulseIntervalTicks; i++ {
		sim.Tick(nil)
	}
	if guest.hp != 3 {
		t.Fatalf("heal rain healed outside ally to %d, want unchanged 3", guest.hp)
	}
	guest.pos = Vec2{X: player.pos.X + 1, Y: player.pos.Y}
	for i := 0; i < 50; i++ {
		sim.Tick(nil)
	}
	if guest.hp != 3 {
		t.Fatalf("expired heal rain healed late entrant to %d, want unchanged 3", guest.hp)
	}

	for i := 0; i < 50; i++ {
		sim.Tick(nil)
	}
	sim.usePlayer(sim.players[hostID])
	monster := firstEntityByKind(sim, monsterEntity)
	if monster == nil {
		t.Fatal("missing monster for heal targeting regression")
	}
	monster.pos = Vec2{X: player.pos.X + 8, Y: player.pos.Y}
	player.mana = player.maxMana
	player.hp = 5
	guest.hp = guest.maxHP
	sim.savePlayer(sim.players[hostID])
	selfHeal := sim.Tick([]Input{{
		MessageID:     "cast_heal_monster_target",
		CorrelationID: "corr_heal_monster_target",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "heal", TargetID: idStr(monster.id)},
	}})
	assertAck(t, selfHeal, "cast_heal_monster_target")
	if player.hp != 7 {
		t.Fatalf("monster-targeted heal hp = %d, want immediate caster heal while area is centered on target", player.hp)
	}
	if guest.hp != guest.maxHP {
		t.Fatalf("monster-targeted heal guest hp = %d, want unchanged full hp %d", guest.hp, guest.maxHP)
	}

	for i := 0; i < 50; i++ {
		sim.Tick(nil)
	}
	sim.usePlayer(sim.players[hostID])
	player.pos = Vec2{X: 2, Y: 5}
	monster.pos = Vec2{X: player.pos.X + rules.Skills["heal"].Effects[0].Range + 2, Y: player.pos.Y}
	player.mana = player.maxMana
	beforeMana := player.mana
	sim.savePlayer(sim.players[hostID])
	outOfRange := sim.Tick([]Input{{
		MessageID:     "cast_heal_monster_target_far",
		CorrelationID: "corr_heal_monster_target_far",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "heal", TargetID: idStr(monster.id)},
	}})
	assertAck(t, outOfRange, "cast_heal_monster_target_far")
	if player.mana != beforeMana {
		t.Fatalf("out-of-range heal auto-nav spent mana before cast: got %d want %d", player.mana, beforeMana)
	}
	if hasEvent(outOfRange, "skill_cast") || hasEvent(outOfRange, "player_healed") {
		t.Fatalf("out-of-range heal emitted cast/heal before auto-nav finished: %+v", outOfRange.Events)
	}
	var completed TickResult
	for steps := 0; steps < 200; steps++ {
		completed = sim.Tick(nil)
		if hasEvent(completed, "skill_cast") || len(completed.Rejects) > 0 {
			break
		}
	}
	if !hasEvent(completed, "skill_cast") || !hasEvent(completed, "skill_cooldown_started") {
		t.Fatalf("auto-nav heal completion missing cast/cooldown events: %+v", completed.Events)
	}
	wantAutoNavMana := beforeMana - skillManaCost(rules.Skills["heal"], 1)
	if player.mana != wantAutoNavMana {
		t.Fatalf("auto-nav heal mana after cast = %d, want %d", player.mana, wantAutoNavMana)
	}

	for i := 0; i < 50; i++ {
		sim.Tick(nil)
	}
	sim.usePlayer(sim.players[hostID])
	player.mana = player.maxMana
	player.hp = player.maxHP
	guest.hp = guest.maxHP
	sim.savePlayer(sim.players[hostID])
	noop := sim.Tick([]Input{{
		MessageID: "cast_heal_full",
		Type:      "cast_skill_intent",
		CastSkill: &CastSkillIntent{SkillID: "heal"},
	}})
	assertAck(t, noop, "cast_heal_full")
	wantNoopMana := player.maxMana - skillManaCost(rules.Skills["heal"], 1)
	if player.mana != wantNoopMana {
		t.Fatalf("full-hp heal mana = %d, want %d", player.mana, wantNoopMana)
	}
	if hasEvent(noop, "player_healed") {
		t.Fatalf("full-hp heal events = %+v, want no player_healed", noop.Events)
	}
	if !hasEvent(noop, "skill_cast") || !hasEvent(noop, "skill_cooldown_started") {
		t.Fatalf("full-hp heal events = %+v, want cast plus cooldown", noop.Events)
	}
}

func TestMagicStatScalesSkillDamageHealAndArea(t *testing.T) {
	rules := loadRules(t)
	magicBolt := rules.Skills[magicBoltSkillID]
	heal := rules.Skills["heal"]

	base := MustNewSim("sess_magic_skill_base", "01", rules)
	base.progression.CharacterClass = "sorcerer"
	base.progression.BaseStats.Magic = skillStatRequirementForRank(magicBolt, "magic", 1)
	baseRange := base.scaleSkillDamageForMagic(magicBolt, 1, skillDamageRange(magicBolt, 1))

	scaled := MustNewSim("sess_magic_skill_scaled", "01", rules)
	scaled.progression.CharacterClass = "sorcerer"
	scaled.progression.BaseStats.Magic = skillStatRequirementForRank(magicBolt, "magic", 1) + 50
	scaledRange := scaled.scaleSkillDamageForMagic(magicBolt, 1, skillDamageRange(magicBolt, 1))

	if scaledRange.Min <= baseRange.Min || scaledRange.Max <= baseRange.Max {
		t.Fatalf("magic-scaled skill damage = %+v, want above base %+v", scaledRange, baseRange)
	}

	sim := MustNewSim("sess_magic_heal_scaled", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.savePlayer(sim.defaultPlayer())
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest_magic", "char_guest_magic", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	sim.usePlayer(sim.players[hostID])
	player := sim.entities[hostID]
	guest := sim.entities[guestID]
	effect := heal.Effects[0]
	player.hp = player.maxHP / 2
	guest.hp = guest.maxHP / 2
	playerHPBefore := player.hp
	guestHPBefore := guest.hp
	player.mana = player.maxMana
	sim.progression.BaseStats.Magic = skillStatRequirementForRank(heal, "magic", 1) + 50
	sim.progression.SkillRanks["heal"] = 1
	basePercent := skillEffectPercent(effect, 1)
	scaledPercent := sim.scaleSkillPercentForMagic(heal, 1, effect, basePercent)
	if scaledPercent <= basePercent {
		t.Fatalf("magic-scaled heal percent = %d, want above base %d", scaledPercent, basePercent)
	}
	guest.pos = Vec2{X: player.pos.X + effect.Radius + 0.5, Y: player.pos.Y}
	sim.savePlayer(sim.players[hostID])

	cast := sim.Tick([]Input{{
		MessageID:     "cast_magic_scaled_heal",
		CorrelationID: "corr_magic_scaled_heal",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "heal", TargetID: idStr(hostID)},
	}})
	assertAck(t, cast, "cast_magic_scaled_heal")

	expectedHeal := healAmountForTarget(&entity{hp: playerHPBefore, maxHP: player.maxHP}, scaledPercent)
	if player.hp < playerHPBefore+expectedHeal {
		t.Fatalf("magic-scaled heal hp = %d, want at least %d from scaled percent %d", player.hp, playerHPBefore+expectedHeal, scaledPercent)
	}
	if guest.hp <= guestHPBefore {
		t.Fatalf("magic-scaled heal radius did not reach guest at %.2f from %.2f: hp=%d", effect.Radius+0.5, effect.Radius, guest.hp)
	}
}

func TestHolyShieldAreaBuffAppliesDefenseVisualStateAndExpires(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_holy_shield", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	sim.progression.BaseStats.Vit = 8
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["holy_shield"] = 1
	player.mana = player.maxMana
	beforeMana := player.mana
	sim.savePlayer(sim.defaultPlayer())

	before, _ := sim.playerEffectiveCombatStats()
	cast := sim.Tick([]Input{{
		MessageID:     "cast_holy_shield",
		CorrelationID: "corr_holy_shield",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "holy_shield"},
	}})
	assertAck(t, cast, "cast_holy_shield")
	if player.mana != beforeMana-5 {
		t.Fatalf("holy shield mana after cast = %d, want %d", player.mana, beforeMana-5)
	}
	if !hasEvent(cast, "skill_cast") || !hasEvent(cast, "skill_effect_started") || !hasEvent(cast, "skill_cooldown_started") {
		t.Fatalf("holy shield missing cast/effect/cooldown events: %+v", cast.Events)
	}
	if !sameStringSlice(player.effectIDs, []string{"holy_shield"}) {
		t.Fatalf("player effect ids = %v, want holy_shield", player.effectIDs)
	}
	after, breakdowns := sim.playerEffectiveCombatStats()
	if after.Armor <= before.Armor || after.BlockPercent <= before.BlockPercent {
		t.Fatalf("holy shield stats before=%+v after=%+v", before, after)
	}
	if armor := findStatBreakdown(breakdowns, "armor"); armor == nil || !statBreakdownHasSourceKind(*armor, "skill_effect") {
		t.Fatalf("holy shield armor breakdown missing skill source: %+v", armor)
	}
	if block := findStatBreakdown(breakdowns, "block_percent"); block == nil || !statBreakdownHasSourceKind(*block, "skill_effect") {
		t.Fatalf("holy shield block breakdown missing skill source: %+v", block)
	}

	var expired TickResult
	for i := 0; i < 300; i++ {
		expired = sim.Tick(nil)
	}
	if !hasEvent(expired, "skill_effect_ended") {
		t.Fatalf("holy shield expiry missing event: %+v", expired.Events)
	}
	if len(player.effectIDs) != 0 {
		t.Fatalf("holy shield effect ids after expiry = %v, want empty", player.effectIDs)
	}
	finalStats, _ := sim.playerEffectiveCombatStats()
	if math.Abs(finalStats.Armor-before.Armor) > 0.000001 || math.Abs(finalStats.BlockPercent-before.BlockPercent) > 0.000001 {
		t.Fatalf("holy shield stats after expiry=%+v want before=%+v", finalStats, before)
	}
}

func TestHolyShieldIgnoresAimAndAlwaysBuffsCaster(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_holy_shield_caster_centered", "01", rules)
	sim.progression.CharacterClass = "paladin"
	sim.progression.BaseStats.Vit = 8
	sim.progression.BaseStats.Magic = 8
	sim.progression.SkillRanks["holy_shield"] = 1
	sim.savePlayer(sim.defaultPlayer())
	player := sim.entities[sim.playerID]
	player.mana = player.maxMana
	beforeMana := player.mana
	sim.savePlayer(sim.defaultPlayer())

	cast := sim.Tick([]Input{{
		MessageID:     "cast_holy_shield_bad_target",
		CorrelationID: "corr_holy_shield_bad_target",
		Type:          "cast_skill_intent",
		CastSkill:     &CastSkillIntent{SkillID: "holy_shield", TargetID: "999999", Direction: &Vec2{X: 1, Y: 0}},
	}})
	assertAck(t, cast, "cast_holy_shield_bad_target")
	if !hasEvent(cast, "skill_effect_started") {
		t.Fatalf("holy shield with bad target missing caster buff event: %+v", cast.Events)
	}
	if !sameStringSlice(player.effectIDs, []string{"holy_shield"}) {
		t.Fatalf("player effect ids = %v, want holy_shield", player.effectIDs)
	}
	if player.mana != beforeMana-5 {
		t.Fatalf("holy shield mana after bad-target cast = %d, want %d", player.mana, beforeMana-5)
	}
}

func TestStrengthDamageBonusAdjustsMeleeDamageRange(t *testing.T) {
	rules := loadRules(t)
	base := MustNewSim("sess_damage_base", "01", rules)
	strong, err := NewSimWithWorldProgression("sess_damage_str", "01", rules, DefaultWorldID, CharacterProgressionState{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		BaseStats:         BaseStatsView{Str: 10, Dex: 5, Vit: 5, Magic: 5},
	})
	if err != nil {
		t.Fatalf("new strong sim: %v", err)
	}
	if got := base.resolvePlayerAttackDamage(); got != (DamageRange{Min: 2, Max: 4}) {
		t.Fatalf("base damage range = %+v, want {2 4}", got)
	}
	if got := strong.resolvePlayerAttackDamage(); got != (DamageRange{Min: 3, Max: 6}) {
		t.Fatalf("strong damage range = %+v, want {3 6}", got)
	}
}

func TestNewSimWithWorldSpawnsPresets(t *testing.T) {
	rules := loadRules(t)

	vertical, err := NewSimWithWorld("sess_vertical", "01", rules, DefaultWorldID)
	if err != nil {
		t.Fatalf("vertical world: %v", err)
	}
	vsnap := vertical.Snapshot()
	if len(vsnap.Entities) != 2 {
		t.Fatalf("vertical entities = %d, want 2: %+v", len(vsnap.Entities), vsnap.Entities)
	}
	assertEntity(t, vsnap, "1001", playerEntity, "", "", Vec2{X: 10, Y: 5})
	assertEntity(t, vsnap, "1002", monsterEntity, monsterDefID, "", Vec2{X: 12, Y: 5})

	gear, err := NewSimWithWorld("sess_gear", "01", rules, "gear_before_combat")
	if err != nil {
		t.Fatalf("gear world: %v", err)
	}
	gsnap := gear.Snapshot()
	if len(gsnap.Entities) != 3 {
		t.Fatalf("gear entities = %d, want 3: %+v", len(gsnap.Entities), gsnap.Entities)
	}
	assertEntity(t, gsnap, "1001", playerEntity, "", "", Vec2{X: 2, Y: 5})
	assertEntity(t, gsnap, "1002", lootEntity, "", "rusty_sword", Vec2{X: 6, Y: 5})
	assertEntity(t, gsnap, "1003", monsterEntity, "training_dummy_reward", "", Vec2{X: 12, Y: 5})

	collision, err := NewSimWithWorld("sess_collision", "01", rules, "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}
	csnap := collision.Snapshot()
	if len(csnap.Entities) != 2 {
		t.Fatalf("collision entities = %d, want 2 mutable entities: %+v", len(csnap.Entities), csnap.Entities)
	}
	if len(collision.walls) != 5 {
		t.Fatalf("collision walls = %d, want 5", len(collision.walls))
	}
	assertEntity(t, csnap, "1001", playerEntity, "", "", Vec2{X: 2, Y: 2})
	assertEntity(t, csnap, "1002", monsterEntity, "training_dummy_reward", "", Vec2{X: 8, Y: 5})

	door, err := NewSimWithWorld("sess_door", "01", rules, "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}
	dsnap := door.Snapshot()
	if len(dsnap.Entities) != 3 {
		t.Fatalf("door entities = %d, want player+door+loot: %+v", len(dsnap.Entities), dsnap.Entities)
	}
	if len(door.walls) != 5 {
		t.Fatalf("door walls = %d, want 5", len(door.walls))
	}
	assertEntity(t, dsnap, "1001", playerEntity, "", "", Vec2{X: 2, Y: 2})
	assertInteractable(t, dsnap, "1002", "wooden_door", interactableClosed, Vec2{X: 4, Y: 2})
	assertEntity(t, dsnap, "1003", lootEntity, "", "quest_leaf", Vec2{X: 8, Y: 2})

	ranged, err := NewSimWithWorld("sess_ranged", "01", rules, "ranged_lab")
	if err != nil {
		t.Fatalf("ranged world: %v", err)
	}
	rsnap := ranged.Snapshot()
	if len(rsnap.Entities) != 3 {
		t.Fatalf("ranged entities = %d, want player+bow+monster: %+v", len(rsnap.Entities), rsnap.Entities)
	}
	if len(ranged.walls) != 5 {
		t.Fatalf("ranged walls = %d, want 5", len(ranged.walls))
	}
	assertEntity(t, rsnap, "1001", playerEntity, "", "", Vec2{X: 2, Y: 2})
	assertEntity(t, rsnap, "1002", lootEntity, "", "training_bow", Vec2{X: 3, Y: 2})
	assertEntity(t, rsnap, "1003", monsterEntity, "training_dummy_ranged", "", Vec2{X: 12, Y: 5})
}

func TestNewSimReturnsDefaultWorldError(t *testing.T) {
	rules := loadRules(t)
	broken := *rules
	broken.Worlds = map[string]WorldDef{}

	if sim, err := NewSim("sess_missing_default", "01", &broken); err == nil {
		t.Fatalf("NewSim returned nil error and sim=%v for missing default world", sim)
	}
}

func assertEntity(t *testing.T, snap Snapshot, id, typ, monsterDefID, itemDefID string, pos Vec2) {
	t.Helper()
	for _, e := range snap.Entities {
		if e.ID != id {
			continue
		}
		if e.Type != typ || e.MonsterDefID != monsterDefID || e.ItemDefID != itemDefID || e.Position != pos {
			t.Fatalf("entity %s = %+v", id, e)
		}
		return
	}
	t.Fatalf("missing entity %s in %+v", id, snap.Entities)
}

func assertInteractable(t *testing.T, snap Snapshot, id, defID, state string, pos Vec2) {
	t.Helper()
	for _, e := range snap.Entities {
		if e.ID != id {
			continue
		}
		if e.Type != interactableEntity || e.InteractableDefID != defID || e.State != state || e.Position != pos {
			t.Fatalf("interactable %s = %+v", id, e)
		}
		return
	}
	t.Fatalf("missing interactable %s in %+v", id, snap.Entities)
}

// --- cross-language golden fixtures (criterion 7) ---------------------------

func TestDamageFormulaGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		PlayerDamage DamageRange `json:"player_damage"`
		Cases        []struct {
			Draw           int `json:"draw"`
			ExpectedDamage int `json:"expected_damage"`
		} `json:"cases"`
	}
	loadGolden(t, "damage_formula.json", &golden)

	if golden.PlayerDamage != r.Combat.PlayerDamage {
		t.Fatalf("golden player_damage %+v != rules %+v", golden.PlayerDamage, r.Combat.PlayerDamage)
	}
	span := r.Combat.PlayerDamage.Max - r.Combat.PlayerDamage.Min + 1
	for _, c := range golden.Cases {
		got := r.Combat.PlayerDamage.Min + (c.Draw % span)
		if got != c.ExpectedDamage {
			t.Fatalf("draw %d: damage = %d, want %d", c.Draw, got, c.ExpectedDamage)
		}
	}
}

func TestRetaliationDamageGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		RetaliationDamage DamageRange `json:"retaliation_damage"`
		Cases             []struct {
			Draw           int `json:"draw"`
			ExpectedDamage int `json:"expected_damage"`
		} `json:"cases"`
	}
	loadGolden(t, "retaliation_damage.json", &golden)

	dummy := r.Monsters[monsterDefID]
	if dummy.RetaliationDamage == nil {
		t.Fatal("training_dummy missing retaliation_damage")
	}
	if golden.RetaliationDamage != *dummy.RetaliationDamage {
		t.Fatalf("golden retaliation_damage %+v != rules %+v", golden.RetaliationDamage, *dummy.RetaliationDamage)
	}
	span := dummy.RetaliationDamage.Max - dummy.RetaliationDamage.Min + 1
	for _, c := range golden.Cases {
		got := dummy.RetaliationDamage.Min + (c.Draw % span)
		if got != c.ExpectedDamage {
			t.Fatalf("draw %d: retaliation damage = %d, want %d", c.Draw, got, c.ExpectedDamage)
		}
	}
}

func TestEquippedWeaponDamageGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		ItemDefID string      `json:"item_def_id"`
		Damage    DamageRange `json:"damage"`
		Cases     []struct {
			Draw           int `json:"draw"`
			ExpectedDamage int `json:"expected_damage"`
		} `json:"cases"`
	}
	loadGolden(t, "equipped_weapon_damage.json", &golden)

	item := r.Items[golden.ItemDefID]
	if !item.Equippable || item.Slot != mainHandSlot || item.Damage == nil {
		t.Fatalf("golden item %s = %+v, want equippable weapon with damage", golden.ItemDefID, item)
	}
	if golden.Damage != *item.Damage {
		t.Fatalf("golden damage %+v != rules %+v", golden.Damage, *item.Damage)
	}
	span := item.Damage.Max - item.Damage.Min + 1
	for _, c := range golden.Cases {
		got := item.Damage.Min + (c.Draw % span)
		if got != c.ExpectedDamage {
			t.Fatalf("draw %d: weapon damage = %d, want %d", c.Draw, got, c.ExpectedDamage)
		}
	}
}

func TestAutoPathGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name    string `json:"name"`
			WorldID string `json:"world_id"`
		} `json:"cases"`
	}
	loadGolden(t, "auto_path.json", &golden)
	rules := loadRules(t)
	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			sim, err := NewSimWithWorld("sess_auto_path_golden", "01", rules, tc.WorldID)
			if err != nil {
				t.Fatalf("world: %v", err)
			}
			target := firstEntityByKind(sim, monsterEntity)
			if target == nil {
				t.Fatal("missing monster target")
			}
			end, steps, ok := sim.findMeleeApproachGoal(target)
			if !ok {
				t.Fatal("findMeleeApproachGoal ok=false")
			}
			if len(steps) == 0 {
				t.Fatal("findMeleeApproachGoal returned empty path")
			}
			if len(steps) > rules.Navigation.MaxAutoSteps {
				t.Fatalf("path len %d exceeds max_auto_steps %d", len(steps), rules.Navigation.MaxAutoSteps)
			}
			if !meleeInRange(distance(end, target.pos), sim.playerMeleeReach(), sim.targetInteractionRadius(target)) {
				t.Fatalf("path end %+v is not in melee reach of target %+v", end, target.pos)
			}
		})
	}
}

func TestRangedProjectileGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name                       string   `json:"name"`
			WorldID                    string   `json:"world_id"`
			Seed                       string   `json:"seed"`
			BaseHitChance              *float64 `json:"base_hit_chance"`
			PlayerPosition             Vec2     `json:"player_position"`
			ExpectedEvent              string   `json:"expected_event"`
			ExpectedMonsterHPUnchanged bool     `json:"expected_monster_hp_unchanged"`
			ExpectedPlayerHP           int      `json:"expected_player_hp"`
			ExpectedMonsterDead        bool     `json:"expected_monster_dead"`
		} `json:"cases"`
	}
	loadGolden(t, "ranged_projectile.json", &golden)
	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			rules := loadRules(t)
			if tc.BaseHitChance != nil {
				rules = rulesCopyWithHitChance(rules, *tc.BaseHitChance)
			}
			sim := rangedLabWithEquippedBow(t, rules, tc.Seed)
			sim.entities[sim.playerID].pos = tc.PlayerPosition
			monster := firstEntityByKind(sim, monsterEntity)
			initialMonsterHP := monster.hp
			fire := sim.Tick([]Input{{MessageID: "fire", CorrelationID: "corr_ranged", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
			assertAck(t, fire, "fire")
			if firstEntityByKind(sim, projectileEntity) == nil && sim.autoNav == nil && !hasEvent(fire, tc.ExpectedEvent) {
				t.Fatalf("no projectile spawned, auto-nav queued, or expected event on fire tick: %+v", fire)
			}
			var impact TickResult
			resolved := false
			for i := 0; i < 80; i++ {
				r := sim.Tick(nil)
				if len(r.Events) > 0 {
					impact = r
					resolved = true
					break
				}
			}
			if !resolved {
				t.Fatal("projectile scenario did not resolve within tick budget")
			}
			if tc.ExpectedEvent != "" && !hasEvent(impact, tc.ExpectedEvent) {
				t.Fatalf("impact events = %+v, want %s", impact.Events, tc.ExpectedEvent)
			}
			if tc.ExpectedMonsterHPUnchanged && monster.hp != initialMonsterHP {
				t.Fatalf("monster hp = %d, want unchanged %d", monster.hp, initialMonsterHP)
			}
			if tc.ExpectedMonsterDead && monster.hp != 0 {
				t.Fatalf("monster hp = %d, want dead", monster.hp)
			}
			if tc.ExpectedPlayerHP != 0 {
				player := sim.entities[sim.playerID]
				if player.hp != tc.ExpectedPlayerHP {
					t.Fatalf("player hp = %d, want %d", player.hp, tc.ExpectedPlayerHP)
				}
			}
		})
	}
}

func rangedLabWithEquippedBow(t *testing.T, rules *Rules, seed string) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_ranged", seed, rules, "ranged_lab")
	if err != nil {
		t.Fatalf("ranged_lab world: %v", err)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_bow", CorrelationID: "corr_pick", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, pickup, "pick_bow")
	equip := sim.Tick([]Input{{MessageID: "equip_bow", CorrelationID: "corr_equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
	return sim
}

func combatControlLabWithEquippedBow(t *testing.T, rules *Rules, seed string) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_combat_control", seed, rules, "combat_control_lab")
	if err != nil {
		t.Fatalf("combat_control_lab world: %v", err)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_bow", CorrelationID: "corr_pick", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, pickup, "pick_bow")
	equip := sim.Tick([]Input{{MessageID: "equip_bow", CorrelationID: "corr_equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
	return sim
}

func equipStaticBow(t *testing.T, sim *Sim) {
	t.Helper()
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "training_bow"})
	equip := sim.Tick([]Input{{MessageID: "equip_bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
	assertAck(t, equip, "equip_bow")
}

func rulesWithTrainingBowReach(t *testing.T, reach float64) *Rules {
	t.Helper()
	base := loadRules(t)
	copyRules := *base
	items := make(map[string]ItemDef, len(base.Items))
	for k, v := range base.Items {
		items[k] = v
	}
	bow := items["training_bow"]
	bow.Reach = &reach
	items["training_bow"] = bow
	copyRules.Items = items
	return &copyRules
}

func rulesWithHitChance(t *testing.T, chance float64) *Rules {
	t.Helper()
	return rulesCopyWithHitChance(loadRules(t), chance)
}

func rulesCopyWithHitChance(base *Rules, chance float64) *Rules {
	copyRules := *base
	copyRules.Combat = base.Combat
	copyRules.Combat.BaseHitChance = chance
	progress := base.CharacterProgression
	derived := make(map[string]LinearStatFormula, len(base.CharacterProgression.DerivedStats))
	for key, formula := range base.CharacterProgression.DerivedStats {
		derived[key] = formula
	}
	hit := derived["hit_chance"]
	hit.Type = "linear"
	hit.Base = chance
	hit.PerStr = 0
	hit.PerDex = 0
	hit.PerVit = 0
	hit.PerMagic = 0
	hit.Stat = ""
	hit.Scale = 0
	hit.Offset = 0
	hit.Denominator = 0
	hit.Min = &chance
	hit.Max = &chance
	derived["hit_chance"] = hit
	progress.DerivedStats = derived
	copyRules.CharacterProgression = progress
	return &copyRules
}

func TestProjectileBusyRejectsSecondFire(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	first := sim.Tick([]Input{{MessageID: "fire1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, first, "fire1")
	second := sim.Tick([]Input{{MessageID: "fire2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, second, "fire2", "projectile_busy")
}

func TestBasicAttackCooldownRejectsRapidMelee(t *testing.T) {
	sim := MustNewSim("sess_basic_attack_cooldown", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = sim.entities[sim.playerID].pos
	monster.hp = 50
	monster.maxHP = 50
	interval := sim.DerivedStatsView().AttackIntervalTicks

	first := sim.Tick([]Input{{MessageID: "hit1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, first, "hit1")
	second := sim.Tick([]Input{{MessageID: "hit2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, second, "hit2", "basic_attack_on_cooldown")
	for i := 0; i < interval-1; i++ {
		sim.Tick(nil)
	}
	third := sim.Tick([]Input{{MessageID: "hit3", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, third, "hit3")
}

func TestDirectionalAttackRejectsInvalidDirection(t *testing.T) {
	sim := MustNewSim("sess_directional_invalid", "01", loadRules(t))
	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{}}}})
	assertReject(t, r, "dir", "invalid_direction")
}

func TestDirectionalMeleeHitsMonsterInFront(t *testing.T) {
	sim := MustNewSim("sess_directional_melee", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = Vec2{X: 11.2, Y: 5}
	monster.hp = 10
	monster.maxHP = 10

	r := sim.Tick([]Input{{MessageID: "dir", CorrelationID: "corr_dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if !hasEvent(r, "monster_damaged") {
		t.Fatalf("directional melee events = %+v, want monster_damaged", r.Events)
	}
	if monster.hp >= monster.maxHP {
		t.Fatalf("monster hp = %d, want reduced", monster.hp)
	}
}

func TestDirectionalMeleeMissesBehindAndOutsideCapsule(t *testing.T) {
	t.Run("behind", func(t *testing.T) {
		sim := MustNewSim("sess_directional_behind", "01", loadRules(t))
		monster := firstEntityByKind(sim, monsterEntity)
		monster.pos = Vec2{X: 9.2, Y: 5}
		initialHP := monster.hp
		r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, r, "dir")
		if len(r.Events) != 0 {
			t.Fatalf("behind swing emitted events: %+v", r.Events)
		}
		if monster.hp != initialHP {
			t.Fatalf("behind monster hp = %d, want %d", monster.hp, initialHP)
		}
	})

	t.Run("outside capsule", func(t *testing.T) {
		sim := MustNewSim("sess_directional_lateral", "01", loadRules(t))
		monster := firstEntityByKind(sim, monsterEntity)
		monster.pos = Vec2{X: 11.0, Y: 6.2}
		initialHP := monster.hp
		r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, r, "dir")
		if len(r.Events) != 0 {
			t.Fatalf("outside capsule swing emitted events: %+v", r.Events)
		}
		if monster.hp != initialHP {
			t.Fatalf("outside capsule monster hp = %d, want %d", monster.hp, initialHP)
		}
	})
}

func TestDirectionalMeleeTieBreaksByEntityID(t *testing.T) {
	sim := MustNewSim("sess_directional_tie", "01", loadRules(t))
	first := firstEntityByKind(sim, monsterEntity)
	first.pos = Vec2{X: 11, Y: 4.8}
	first.hp = 10
	first.maxHP = 10
	second := addTestMonster(sim, "training_dummy", Vec2{X: 11, Y: 5.2}, 10)

	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if first.hp >= first.maxHP {
		t.Fatalf("first monster hp = %d, want damaged", first.hp)
	}
	if second.hp != second.maxHP {
		t.Fatalf("second monster hp = %d, want unchanged %d", second.hp, second.maxHP)
	}
}

func TestDirectionalMeleeStopsMovementAndAcksEmptySwing(t *testing.T) {
	sim := MustNewSim("sess_directional_stop", "01", loadRules(t))
	monster := firstEntityByKind(sim, monsterEntity)
	monster.pos = Vec2{X: 9, Y: 5}
	move := sim.Tick([]Input{{MessageID: "move", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	assertAck(t, move, "move")
	beforeAttack := sim.entities[sim.playerID].pos

	r := sim.Tick([]Input{{MessageID: "dir", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, r, "dir")
	if len(r.Events) != 0 {
		t.Fatalf("empty directional swing emitted events: %+v", r.Events)
	}
	if sim.move != nil {
		t.Fatalf("directional attack did not clear movement: %+v", sim.move)
	}
	if sim.entities[sim.playerID].pos != beforeAttack {
		t.Fatalf("directional attack moved player from %+v to %+v", beforeAttack, sim.entities[sim.playerID].pos)
	}
}

func TestDirectionalRangedFreeShotHitsAndOmitsTargetID(t *testing.T) {
	sim := combatControlLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 3, Y: 5}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	initialDistance := distance(monster.pos, player.pos)

	fire := sim.Tick([]Input{{MessageID: "dir_fire", CorrelationID: "corr_dir_fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, fire, "dir_fire")
	spawn := firstChangeEntityByType(fire, projectileEntity)
	if spawn == nil {
		t.Fatalf("directional ranged did not spawn projectile: %+v", fire.Changes)
	}
	if spawn.TargetID != "" {
		t.Fatalf("free-shot projectile target_id = %q, want omitted", spawn.TargetID)
	}

	var impact TickResult
	for i := 0; i < 20; i++ {
		impact = sim.Tick(nil)
		if hasEvent(impact, "monster_damaged") || hasEvent(impact, "monster_killed") || hasEvent(impact, "attack_missed") {
			break
		}
	}
	if !hasEvent(impact, "monster_damaged") {
		t.Fatalf("directional ranged impact events = %+v, want monster_damaged", impact.Events)
	}
	if !hasEvent(impact, "monster_aggro") {
		t.Fatalf("directional ranged impact events = %+v, want monster_aggro", impact.Events)
	}
	if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
		t.Fatalf("monster ai target/mode = %d/%s, want %d/%s", monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
	}

	moved := false
	for i := 0; i < 10; i++ {
		sim.Tick(nil)
		if distance(monster.pos, player.pos) < initialDistance-0.01 {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatalf("aggroed monster did not move toward player: start dist %.3f now %.3f", initialDistance, distance(monster.pos, player.pos))
	}
}

func TestDirectionalRangedProjectileBusy(t *testing.T) {
	sim := combatControlLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	first := sim.Tick([]Input{{MessageID: "fire1", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, first, "fire1")
	second := sim.Tick([]Input{{MessageID: "fire2", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertReject(t, second, "fire2", "projectile_busy")
}

func TestDirectionalRangedProjectileBlockedAndExpires(t *testing.T) {
	t.Run("closed interactable blocks", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_directional_blocked", "01", loadRules(t), "door_lab")
		if err != nil {
			t.Fatalf("door world: %v", err)
		}
		equipStaticBow(t, sim)
		fire := sim.Tick([]Input{{MessageID: "fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, fire, "fire")
		var resolved TickResult
		for i := 0; i < 10; i++ {
			resolved = sim.Tick(nil)
			if hasEvent(resolved, "projectile_blocked") {
				break
			}
		}
		if !hasEvent(resolved, "projectile_blocked") {
			t.Fatalf("blocked projectile events = %+v, want projectile_blocked", resolved.Events)
		}
	})

	t.Run("expires without hit", func(t *testing.T) {
		rules := rulesWithTrainingBowReach(t, 2.0)
		sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
		sim.entities[sim.playerID].pos = Vec2{X: 3, Y: 5}
		fire := sim.Tick([]Input{{MessageID: "fire", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{Y: 1}}}})
		assertAck(t, fire, "fire")
		var resolved TickResult
		for i := 0; i < 10; i++ {
			resolved = sim.Tick(nil)
			if hasEvent(resolved, "projectile_expired") {
				break
			}
		}
		if !hasEvent(resolved, "projectile_expired") {
			t.Fatalf("expired projectile events = %+v, want projectile_expired", resolved.Events)
		}
	})
}

func TestAggroOnHitDirectionalRangedMovesFromOutsidePassiveRadius(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 3, Y: 5}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	if distance(player.pos, monster.pos) <= sim.rules.Monsters[monster.monsterDefID].AggroRadius {
		t.Fatalf("setup inside passive aggro radius: player=%+v monster=%+v", player.pos, monster.pos)
	}

	fire := sim.Tick([]Input{{MessageID: "fire", CorrelationID: "corr_aggro", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	assertAck(t, fire, "fire")
	var impact TickResult
	for i := 0; i < 20; i++ {
		impact = sim.Tick(nil)
		if hasEvent(impact, "monster_aggro") {
			break
		}
	}
	if !hasEvent(impact, "monster_aggro") {
		t.Fatalf("impact events = %+v, want monster_aggro", impact.Events)
	}
	before := monster.pos
	sim.Tick(nil)
	if distance(monster.pos, player.pos) >= distance(before, player.pos)-0.01 {
		t.Fatalf("monster did not chase aggro target: before=%+v after=%+v player=%+v", before, monster.pos, player.pos)
	}
}

func TestAggroOnHitPrefersAttackingPlayerInCoop(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	sim := combatControlLabWithEquippedBow(t, rules, "cafebabecafebabe")
	hostID := sim.playerID
	sim.SetPlayerMetadata(hostID, "acct_host", "char_host", "Host", "host")
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 20
	monster.maxHP = 20
	sim.entities[hostID].pos = Vec2{X: 3, Y: 5}
	sim.entities[guestID].pos = Vec2{X: 12.4, Y: 5}
	sim.savePlayer(sim.players[hostID])
	sim.savePlayer(sim.players[guestID])
	sim.usePlayer(sim.players[hostID])

	fire := sim.TickResults([]Input{{MessageID: "fire", ActorPlayerID: hostID, CorrelationID: "corr_aggro", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
	if len(fire) == 0 {
		t.Fatal("directional fire produced no results")
	}
	assertAck(t, fire[0], "fire")
	var impact TickResult
	for i := 0; i < 20; i++ {
		for _, res := range sim.TickResults(nil) {
			if hasEvent(res, "monster_damaged") || hasEvent(res, "attack_missed") {
				impact = res
				break
			}
		}
		if hasEvent(impact, "monster_damaged") || hasEvent(impact, "attack_missed") {
			break
		}
	}
	if !hasEvent(impact, "monster_damaged") && !hasEvent(impact, "attack_missed") {
		t.Fatalf("projectile impact events = %+v, want monster_damaged or attack_missed", impact.Events)
	}
	if monster.aiTargetPlayerID != hostID {
		t.Fatalf("monster ai target = %d, want host %d", monster.aiTargetPlayerID, hostID)
	}
	targetPlayer := sim.nearestLivingPlayerForMonster(sim.activeLevel(), monster)
	if targetPlayer == nil || targetPlayer.PlayerID != hostID {
		t.Fatalf("target player = %+v, want host %d", targetPlayer, hostID)
	}
}

func TestAggroOnHitPropagatesToNearbyMonsterGroup(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_group_aggro", "group_aggro", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon world: %v", err)
	}
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range level.entities {
		if candidate.kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	near := addTestMonster(sim, "dungeon_mob", Vec2{X: 25, Y: 10}, 20)
	chained := addTestMonster(sim, "dungeon_mob", Vec2{X: 30, Y: 10}, 20)
	far := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_group", &res)

	for _, monster := range []*entity{primary, near, chained} {
		if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
			t.Fatalf("monster %d target/mode = %d/%s, want %d/%s", monster.id, monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
		}
	}
	if far.aiTargetPlayerID != 0 || far.aiMode != monsterAIModeIdle {
		t.Fatalf("far monster target/mode = %d/%s, want idle outside group radius", far.aiTargetPlayerID, far.aiMode)
	}

	aggroEvents := map[string]bool{}
	for _, ev := range res.Events {
		if ev.EventType == "monster_aggro" {
			aggroEvents[ev.EntityID] = true
		}
	}
	for _, monster := range []*entity{primary, near, chained} {
		if !aggroEvents[idStr(monster.id)] {
			t.Fatalf("missing monster_aggro for %d in events %+v", monster.id, res.Events)
		}
	}
	if aggroEvents[idStr(far.id)] {
		t.Fatalf("unexpected far monster_aggro for %d in events %+v", far.id, res.Events)
	}
}

func TestAggroOnLethalHitPropagatesToNearbyMonsterGroup(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_lethal_group_aggro", "lethal_group_aggro", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon world: %v", err)
	}
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range level.entities {
		if candidate.kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 1)
	near := addTestMonster(sim, "dungeon_mob", Vec2{X: 25, Y: 10}, 20)
	far := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	outcome := sim.damageMonsterByPlayer(primary, sim.playerID, "corr_lethal_group", &res, DamageRange{Min: 99, Max: 99})
	if !outcome.Hit || outcome.Blocked || primary.hp != 0 {
		t.Fatalf("setup expected lethal hit outcome=%+v primary_hp=%d events=%+v", outcome, primary.hp, res.Events)
	}
	if primary.aiMode != monsterAIModeIdle || primary.aiTargetPlayerID != 0 {
		t.Fatalf("dead primary target/mode = %d/%s, want no aggro on dead source", primary.aiTargetPlayerID, primary.aiMode)
	}
	if near.aiTargetPlayerID != sim.playerID || near.aiMode != monsterAIModeChase {
		t.Fatalf("near monster target/mode = %d/%s, want %d/%s", near.aiTargetPlayerID, near.aiMode, sim.playerID, monsterAIModeChase)
	}
	if far.aiTargetPlayerID != 0 || far.aiMode != monsterAIModeIdle {
		t.Fatalf("far monster target/mode = %d/%s, want idle outside group radius", far.aiTargetPlayerID, far.aiMode)
	}
	if eventForEntity(res, "monster_aggro", primary.id) {
		t.Fatalf("dead primary should not emit monster_aggro: %+v", res.Events)
	}
	if !eventForEntity(res, "monster_aggro", near.id) {
		t.Fatalf("missing nearby monster_aggro after lethal hit: %+v", res.Events)
	}
	if eventForEntity(res, "monster_aggro", far.id) {
		t.Fatalf("unexpected far monster_aggro after lethal hit: %+v", res.Events)
	}
}

func TestAggroOnHitAlsoAggrosMonstersWithAttackerInRange(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_attack_range_aggro", "range_aggro", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon world: %v", err)
	}
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range level.entities {
		if candidate.kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	sim.syncCompatibilityFields()

	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	attackerRange := addTestMonster(sim, "dungeon_mob", Vec2{X: 7, Y: 5}, 20)
	outsideBoth := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}

	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_attacker_range", &res)

	for _, monster := range []*entity{primary, attackerRange} {
		if monster.aiTargetPlayerID != sim.playerID || monster.aiMode != monsterAIModeChase {
			t.Fatalf("monster %d target/mode = %d/%s, want %d/%s", monster.id, monster.aiTargetPlayerID, monster.aiMode, sim.playerID, monsterAIModeChase)
		}
	}
	if outsideBoth.aiTargetPlayerID != 0 || outsideBoth.aiMode != monsterAIModeIdle {
		t.Fatalf("outside monster target/mode = %d/%s, want idle outside attacker and group radius", outsideBoth.aiTargetPlayerID, outsideBoth.aiMode)
	}

	aggroEvents := map[string]bool{}
	for _, ev := range res.Events {
		if ev.EventType == "monster_aggro" {
			aggroEvents[ev.EntityID] = true
		}
	}
	for _, monster := range []*entity{primary, attackerRange} {
		if !aggroEvents[idStr(monster.id)] {
			t.Fatalf("missing monster_aggro for %d in events %+v", monster.id, res.Events)
		}
	}
	if aggroEvents[idStr(outsideBoth.id)] {
		t.Fatalf("unexpected outside monster_aggro for %d in events %+v", outsideBoth.id, res.Events)
	}
}

func TestAggroOnHitUsesAssistRadiusWithoutGrowingPassiveAggro(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_assist_radius_aggro", "assist_radius_aggro", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("dungeon world: %v", err)
	}
	level, err := sim.ensureDungeonLevel(-1)
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range level.entities {
		if candidate.kind == monsterEntity {
			delete(level.entities, id)
		}
	}
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 2, Y: 5})
	player := level.entities[sim.playerID]
	sim.syncCompatibilityFields()

	def := sim.rules.Monsters["dungeon_mob"]
	if def.effectiveAssistRadius() <= def.AggroRadius {
		t.Fatalf("test requires assist radius > aggro radius, got aggro %.2f assist %.2f", def.AggroRadius, def.effectiveAssistRadius())
	}
	primary := addTestMonster(sim, "dungeon_mob", Vec2{X: 20, Y: 10}, 20)
	assistOnly := addTestMonster(sim, "dungeon_mob", Vec2{X: player.pos.X + def.AggroRadius + 1, Y: player.pos.Y}, 20)
	outside := addTestMonster(sim, "dungeon_mob", Vec2{X: 45, Y: 10}, 20)

	passive := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.updateMonsterAIMode(assistOnly, player, def, assistOnly.aiMode, &passive)
	if assistOnly.aiMode != monsterAIModeIdle || eventForEntity(passive, "monster_aggro", assistOnly.id) {
		t.Fatalf("assist-only monster passively aggroed at distance beyond aggro radius: mode=%s events=%+v", assistOnly.aiMode, passive.Events)
	}

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.aggroMonsterOnHit(primary, sim.playerID, "corr_assist_radius", &res)
	if assistOnly.aiTargetPlayerID != sim.playerID || assistOnly.aiMode != monsterAIModeChase {
		t.Fatalf("assist-only monster target/mode = %d/%s, want %d/%s", assistOnly.aiTargetPlayerID, assistOnly.aiMode, sim.playerID, monsterAIModeChase)
	}
	if outside.aiTargetPlayerID != 0 || outside.aiMode != monsterAIModeIdle {
		t.Fatalf("outside monster target/mode = %d/%s, want idle outside assist radius", outside.aiTargetPlayerID, outside.aiMode)
	}
	if !eventForEntity(res, "monster_aggro", assistOnly.id) {
		t.Fatalf("missing assist radius monster_aggro: %+v", res.Events)
	}
	if eventForEntity(res, "monster_aggro", outside.id) {
		t.Fatalf("unexpected outside monster_aggro: %+v", res.Events)
	}
}

func TestRangedAutoApproachThenFire(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	monster.pos = Vec2{X: 12, Y: 5}
	r := sim.Tick([]Input{{MessageID: "far_fire", CorrelationID: "corr_far", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "far_fire")
	sawProjectile := false
	sawImpact := false
	for i := 0; i < 80 && !sawImpact; i++ {
		r := sim.Tick(nil)
		for _, c := range r.Changes {
			if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == projectileEntity {
				sawProjectile = true
			}
		}
		if hasEvent(r, "monster_damaged") || hasEvent(r, "attack_missed") || hasEvent(r, "projectile_blocked") {
			sawImpact = true
		}
	}
	if !sawProjectile {
		t.Fatal("auto-approach did not spawn projectile")
	}
	if !sawImpact {
		t.Fatal("auto-approach projectile did not resolve")
	}
}

func TestRangedDummyDropsSeparatedLootItems(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 1
	r := sim.Tick([]Input{{MessageID: "kill", CorrelationID: "corr_kill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "kill")
	for i := 0; i < 20 && !hasEvent(r, "monster_killed"); i++ {
		r = sim.Tick(nil)
	}
	if !hasEvent(r, "monster_killed") {
		t.Fatalf("ranged kill did not resolve: %+v", r.Events)
	}

	want := map[string]bool{
		"gold":        false,
		"quest_leaf":  false,
		"red_potion":  false,
		"blue_potion": false,
	}
	positions := map[Vec2]string{}
	for _, c := range r.Changes {
		if c.Op != OpEntitySpawn || c.Entity == nil || c.Entity.Type != lootEntity {
			continue
		}
		itemDefID := c.Entity.ItemDefID
		if _, ok := want[itemDefID]; !ok {
			t.Fatalf("unexpected ranged loot %s in %+v", itemDefID, r.Changes)
		}
		if positions[c.Entity.Position] != "" {
			t.Fatalf("loot overlap at %+v: %s and %s", c.Entity.Position, positions[c.Entity.Position], itemDefID)
		}
		if sim.lootDropBlocked(c.Entity.Position) {
			t.Fatalf("loot spawned inside blocked geometry at %+v", c.Entity.Position)
		}
		positions[c.Entity.Position] = itemDefID
		want[itemDefID] = true
	}
	for itemDefID, seen := range want {
		if !seen {
			t.Fatalf("missing ranged loot %s in %+v", itemDefID, r.Changes)
		}
	}
}

func TestRangedBowLootRequiresMeleeReach(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, rulesWithHitChance(t, 1.0), "cafebabecafebabe")
	if sim.playerActionReach() != 12.0 {
		t.Fatalf("playerActionReach = %v, want bow reach 12.0", sim.playerActionReach())
	}
	if sim.playerMeleeReach() != sim.rules.Combat.UnarmedReach {
		t.Fatalf("playerMeleeReach = %v, want unarmed %v", sim.playerMeleeReach(), sim.rules.Combat.UnarmedReach)
	}

	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	monster := firstEntityByKind(sim, monsterEntity)
	monster.hp = 1
	fire := sim.Tick([]Input{{MessageID: "kill", CorrelationID: "corr_kill", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, fire, "kill")
	var loot *entity
	for i := 0; i < 20; i++ {
		r := sim.Tick(nil)
		if hasEvent(r, "monster_killed") {
			for _, c := range r.Changes {
				if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == lootEntity {
					loot = sim.findEntity(c.Entity.ID)
					break
				}
			}
		}
		if loot != nil {
			break
		}
	}
	if loot == nil {
		t.Fatal("missing loot after ranged kill")
	}
	if sim.inMeleeRange(loot) {
		t.Fatalf("player at %+v should not be in melee range of loot at %+v with bow equipped", sim.entities[sim.playerID].pos, loot.pos)
	}

	pickup := sim.Tick([]Input{{MessageID: "loot_pick", CorrelationID: "corr_loot", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertAck(t, pickup, "loot_pick")
	if sim.autoNav == nil {
		t.Fatal("loot pickup from range should queue auto-nav, not dispatch immediately")
	}
	pickupEvent := "item_picked_up"
	if loot.itemDefID == goldItemDefID {
		pickupEvent = "gold_picked_up"
	}
	if hasEvent(pickup, pickupEvent) {
		t.Fatal("loot picked up instantly from ranged distance")
	}

	picked := false
	for i := 0; i < 80; i++ {
		r := sim.Tick(nil)
		if hasEvent(r, pickupEvent) {
			picked = true
			break
		}
	}
	if !picked {
		t.Fatal("auto-nav did not complete loot pickup within tick budget")
	}
}

func TestGoldAutoPickupWalkingIntoTownGold(t *testing.T) {
	sim := MustNewSim("sess_gold_auto_town", "v49_gold_auto_town", loadRules(t))
	player := sim.entities[sim.playerID]
	gold := addTestGoldLoot(sim, Vec2{X: player.pos.X + 2, Y: player.pos.Y}, 7)

	results := sim.TickResults([]Input{{MessageID: "move_gold", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})

	assertAckInResults(t, results, "move_gold")
	if sim.gold != 7 || sim.players[sim.playerID].Gold != 7 || sim.players[sim.playerID].Progression.Gold != 7 {
		t.Fatalf("gold after auto-pickup = sim %d player %d progression %d, want 7", sim.gold, sim.players[sim.playerID].Gold, sim.players[sim.playerID].Progression.Gold)
	}
	if sim.entities[gold.id] != nil {
		t.Fatalf("gold entity %d still present after auto-pickup", gold.id)
	}
	if !hasEventInResults(results, "gold_picked_up") {
		t.Fatalf("missing gold_picked_up in results: %+v", results)
	}
	if owners := changeOwnersForOpInResults(results, OpGoldUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("gold_update owners = %v, want player %d", owners, sim.playerID)
	}
	if owners := changeOwnersForOpInResults(results, OpCharacterProgressionUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("progression owners = %v, want player %d", owners, sim.playerID)
	}
}

func TestGoldAutoPickupWorksOnDungeonLevelWallet(t *testing.T) {
	sim, err := NewSimWithWorld("sess_gold_auto_dungeon", "v49_gold_auto_dungeon", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	hostID := sim.playerID
	descendFromCurrentLevel(t, sim, "descend_1")
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	player := sim.entities[hostID]
	gold := addTestGoldLoot(sim, player.pos, 13)

	res := sim.Tick(nil)

	if sim.entities[gold.id] != nil {
		t.Fatalf("dungeon gold entity %d still present", gold.id)
	}
	if sim.players[hostID].Gold != 13 || sim.players[hostID].Progression.Gold != 13 {
		t.Fatalf("dungeon wallet gold=%d progression=%d, want 13", sim.players[hostID].Gold, sim.players[hostID].Progression.Gold)
	}
	if !hasEvent(res, "gold_picked_up") || !hasChange(res, OpGoldUpdate) || !hasProgressionChange(res) {
		t.Fatalf("dungeon pickup result missing wallet/progression: %+v", res)
	}
}

func TestNonGoldLootDoesNotAutoPickup(t *testing.T) {
	sim := MustNewSim("sess_item_no_auto", "v49_item_no_auto", loadRules(t))
	player := sim.entities[sim.playerID]
	loot := addTestFloorLoot(sim, "quest_leaf", player.pos)

	res := sim.Tick(nil)

	if sim.entities[loot.id] == nil {
		t.Fatalf("non-gold loot %d was auto-picked", loot.id)
	}
	if len(sim.inventory) != 0 || hasEvent(res, "item_picked_up") {
		t.Fatalf("non-gold loot auto-mutated inventory/events: inv=%+v result=%+v", sim.inventory, res)
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_leaf", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertAck(t, pickup, "pick_leaf")
	if len(sim.inventory) != 1 || sim.inventory[0].itemDefID != "quest_leaf" {
		t.Fatalf("explicit item pickup inventory = %+v, want quest_leaf", sim.inventory)
	}
}

func TestManualGoldPickupStillWorksInRange(t *testing.T) {
	sim := MustNewSim("sess_gold_manual", "v49_gold_manual", loadRules(t))
	player := sim.entities[sim.playerID]
	gold := addTestGoldLoot(sim, player.pos, 9)

	res := sim.Tick([]Input{{MessageID: "pick_gold", CorrelationID: "corr_gold", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(gold.id)}}})

	assertAck(t, res, "pick_gold")
	if sim.gold != 9 || sim.entities[gold.id] != nil {
		t.Fatalf("manual gold pickup gold=%d entity=%+v, want gold 9 and removed", sim.gold, sim.entities[gold.id])
	}
	if owners := changeOwnersForOp(res, OpGoldUpdate); !sameUint64Set(owners, []uint64{sim.playerID}) {
		t.Fatalf("manual gold_update owners = %v, want player %d", owners, sim.playerID)
	}
	foundCorr := false
	for _, ev := range res.Events {
		if ev.EventType == "gold_picked_up" && ev.CorrelationID == "corr_gold" {
			foundCorr = true
		}
	}
	if !foundCorr {
		t.Fatalf("manual gold pickup missing correlation: %+v", res.Events)
	}
}

func TestGoldAutoPickupCoopLowestPlayerIDWins(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_gold_coop_winner", "v49_gold_coop_winner", rules)
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 11)
	sim.entities[hostID].pos = gold.pos
	sim.entities[guestID].pos = gold.pos

	results := sim.TickResults(nil)

	if sim.players[hostID].Gold != 11 || sim.players[guestID].Gold != 0 {
		t.Fatalf("coop gold host=%d guest=%d, want host winner", sim.players[hostID].Gold, sim.players[guestID].Gold)
	}
	if owners := changeOwnersForOpInResults(results, OpGoldUpdate); !sameUint64Set(owners, []uint64{hostID}) {
		t.Fatalf("coop gold_update owners = %v, want host %d", owners, hostID)
	}
	if !eventForEntityInResults(results, "gold_picked_up", hostID) {
		t.Fatalf("coop pickup event missing host winner: %+v", results)
	}
}

func TestGoldAutoPickupSkipsDeadAndDisconnectedPlayers(t *testing.T) {
	t.Run("dead lower player", func(t *testing.T) {
		rules := loadRules(t)
		sim := MustNewSim("sess_gold_dead_skip", "v49_gold_dead_skip", rules)
		hostID := sim.playerID
		guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
		if err != nil {
			t.Fatalf("add guest: %v", err)
		}
		gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 5)
		sim.entities[hostID].pos = gold.pos
		sim.entities[guestID].pos = gold.pos
		sim.entities[hostID].hp = 0

		sim.TickResults(nil)

		if sim.players[hostID].Gold != 0 || sim.players[guestID].Gold != 5 {
			t.Fatalf("dead-player skip gold host=%d guest=%d, want guest", sim.players[hostID].Gold, sim.players[guestID].Gold)
		}
	})
	t.Run("disconnected lower player", func(t *testing.T) {
		rules := loadRules(t)
		sim := MustNewSim("sess_gold_disconnected_skip", "v49_gold_disconnected_skip", rules)
		hostID := sim.playerID
		guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
		if err != nil {
			t.Fatalf("add guest: %v", err)
		}
		gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 6}, 6)
		sim.entities[hostID].pos = gold.pos
		sim.entities[guestID].pos = gold.pos
		sim.players[hostID].Connected = false

		sim.TickResults(nil)

		if sim.players[hostID].Gold != 0 || sim.players[guestID].Gold != 6 {
			t.Fatalf("disconnected-player skip gold host=%d guest=%d, want guest", sim.players[hostID].Gold, sim.players[guestID].Gold)
		}
	})
}

func TestGoldAutoPickupMultipleGoldStableEntityOrder(t *testing.T) {
	sim := MustNewSim("sess_gold_order", "v49_gold_order", loadRules(t))
	player := sim.entities[sim.playerID]
	first := addTestGoldLoot(sim, player.pos, 3)
	second := addTestGoldLoot(sim, player.pos, 5)
	if first.id >= second.id {
		t.Fatalf("test setup ids not increasing: first=%d second=%d", first.id, second.id)
	}

	res := sim.Tick(nil)

	if sim.gold != 8 {
		t.Fatalf("gold after multiple pickup = %d, want 8", sim.gold)
	}
	amounts := goldPickupAmounts(res)
	if len(amounts) != 2 || amounts[0] != 3 || amounts[1] != 5 {
		t.Fatalf("gold pickup amounts = %v, want [3 5]", amounts)
	}
}

func TestGoldAutoPickupPendingAutoNavDoesNotDuplicate(t *testing.T) {
	sim := MustNewSim("sess_gold_pending_nav", "v49_gold_pending_nav", loadRules(t))
	hostID := sim.playerID
	sim.entities[hostID].pos = Vec2{X: 2, Y: 2}
	gold := addTestGoldLoot(sim, Vec2{X: 6, Y: 2}, 10)

	queue := sim.TickResults([]Input{{MessageID: "queue_gold", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(gold.id)}}})
	assertAckInResults(t, queue, "queue_gold")
	if sim.players[hostID].Gold != 0 {
		t.Fatalf("out-of-range click picked gold immediately: %d", sim.players[hostID].Gold)
	}

	pickupEvents := countEventsInResults(queue, "gold_picked_up")
	for i := 0; i < 20 && sim.players[hostID].Gold == 0; i++ {
		results := sim.TickResults(nil)
		pickupEvents += countEventsInResults(results, "gold_picked_up")
	}
	if sim.players[hostID].Gold != 10 {
		t.Fatalf("pending auto-nav gold = %d, want 10", sim.players[hostID].Gold)
	}
	if sim.entities[gold.id] != nil {
		t.Fatalf("pending auto-nav gold entity %d still present", gold.id)
	}
	for i := 0; i < 5; i++ {
		results := sim.TickResults(nil)
		pickupEvents += countEventsInResults(results, "gold_picked_up")
		if sim.players[hostID].Gold != 10 {
			t.Fatalf("gold duplicated after pickup: %d", sim.players[hostID].Gold)
		}
	}
	if pickupEvents != 1 {
		t.Fatalf("gold_picked_up events = %d, want exactly 1", pickupEvents)
	}
}

func TestRangedBlockedLineAutoMovesUntilClearThenFires(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "deadbeefdeadbeef")
	monster := firstEntityByKind(sim, monsterEntity)
	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 8}
	if sim.hasClearRangedShot(sim.entities[sim.playerID].pos, monster) {
		t.Fatal("test setup has clear shot; want wall-blocked line")
	}

	r := sim.Tick([]Input{{MessageID: "covered_fire", CorrelationID: "corr_covered", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, r, "covered_fire")
	if sim.autoNav == nil {
		t.Fatal("blocked ranged click fired immediately; want auto-nav")
	}
	if firstEntityByKind(sim, projectileEntity) != nil {
		t.Fatal("projectile spawned before line was clear")
	}

	sawProjectile := false
	sawImpact := false
	for i := 0; i < 80 && !sawImpact; i++ {
		r := sim.Tick(nil)
		for _, c := range r.Changes {
			if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == projectileEntity {
				sawProjectile = true
				player := sim.entities[sim.playerID]
				if !sim.hasClearRangedShot(player.pos, monster) {
					t.Fatalf("projectile spawned without clear shot from %+v to %+v", player.pos, monster.pos)
				}
				playerMonsterDistance := distance(player.pos, monster.pos)
				if meleeInRange(playerMonsterDistance, sim.rules.Combat.UnarmedReach, monsterRadius) {
					t.Fatalf("ranged auto-nav entered melee range at %+v", player.pos)
				}
			}
		}
		if hasEvent(r, "monster_damaged") || hasEvent(r, "monster_killed") || hasEvent(r, "attack_missed") || hasEvent(r, "projectile_blocked") {
			sawImpact = true
			if hasEvent(r, "projectile_blocked") {
				t.Fatalf("projectile was still blocked after auto-nav: %+v", r.Events)
			}
		}
	}
	if !sawProjectile {
		t.Fatal("auto-nav never spawned projectile")
	}
	if !sawImpact {
		t.Fatal("auto-nav projectile did not resolve")
	}
}

func TestActionIntentAutoApproachAndAttack(t *testing.T) {
	sim, err := NewSimWithWorld("sess_path_maze", "01", loadRules(t), "path_maze")
	if err != nil {
		t.Fatalf("path_maze world: %v", err)
	}
	target := firstEntityByKind(sim, monsterEntity)
	r := sim.Tick([]Input{{MessageID: "maze_action", CorrelationID: "corr_maze", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
	assertAck(t, r, "maze_action")
	for i := 0; i < 100 && target.hp > 0; i++ {
		sim.Tick(nil)
	}
	if target.hp != 0 {
		t.Fatalf("target hp = %d, want killed by queued action", target.hp)
	}
}

func TestMoveToIntentArrivesAndManualMoveCancels(t *testing.T) {
	sim, err := NewSimWithWorld("sess_move_to", "01", loadRules(t), "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}
	r := sim.Tick([]Input{{MessageID: "go", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: Vec2{X: 3, Y: 5}}}})
	assertAck(t, r, "go")
	sim.Tick(nil)
	manual := sim.Tick([]Input{{MessageID: "manual", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{Y: 1}, DurationTicks: 1}}})
	assertAck(t, manual, "manual")
	if sim.autoNav != nil {
		t.Fatal("manual move did not clear autoNav")
	}
}

func TestStopMovementIntentCancelsActiveMove(t *testing.T) {
	sim, err := NewSimWithWorld("sess_stop_move", "abcd", loadRules(t), "gear_before_combat")
	if err != nil {
		t.Fatalf("gear world: %v", err)
	}
	start := sim.entities[sim.playerID].pos
	move := sim.Tick([]Input{{MessageID: "move", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	assertAck(t, move, "move")
	moved := sim.entities[sim.playerID].pos
	if moved.X <= start.X {
		t.Fatalf("setup failed: player did not move from %+v to %+v", start, moved)
	}

	stop := sim.Tick([]Input{{MessageID: "stop", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{}, DurationTicks: 1}}})
	assertAck(t, stop, "stop")
	if sim.move != nil {
		t.Fatalf("stop did not clear active move: %+v", sim.move)
	}
	if sim.entities[sim.playerID].pos != moved {
		t.Fatalf("stop tick moved player from %+v to %+v", moved, sim.entities[sim.playerID].pos)
	}

	sim.Tick(nil)
	if sim.entities[sim.playerID].pos != moved {
		t.Fatalf("player moved after stop from %+v to %+v", moved, sim.entities[sim.playerID].pos)
	}
}

func TestStopMovementIntentCancelsAutoNavAndPendingAction(t *testing.T) {
	t.Run("move_to", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_stop_nav", "01", loadRules(t), "collision_lab")
		if err != nil {
			t.Fatalf("collision world: %v", err)
		}
		goNav := sim.Tick([]Input{{MessageID: "go", Type: "move_to_intent", MoveTo: &MoveToIntent{Position: Vec2{X: 7, Y: 2}}}})
		assertAck(t, goNav, "go")
		if sim.autoNav == nil {
			t.Fatal("setup failed: move_to did not queue autoNav")
		}
		beforeStop := sim.entities[sim.playerID].pos
		stop := sim.Tick([]Input{{MessageID: "stop", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{}, DurationTicks: 1}}})
		assertAck(t, stop, "stop")
		if sim.autoNav != nil {
			t.Fatalf("stop did not clear autoNav: %+v", sim.autoNav)
		}
		if sim.entities[sim.playerID].pos != beforeStop {
			t.Fatalf("stop tick advanced autoNav from %+v to %+v", beforeStop, sim.entities[sim.playerID].pos)
		}
	})

	t.Run("pending action", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_stop_action", "01", loadRules(t), "path_maze")
		if err != nil {
			t.Fatalf("path_maze world: %v", err)
		}
		target := firstEntityByKind(sim, monsterEntity)
		queue := sim.Tick([]Input{{MessageID: "attack", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(target.id)}}})
		assertAck(t, queue, "attack")
		if sim.autoNav == nil || sim.autoNav.pendingAction == nil {
			t.Fatal("setup failed: action did not queue pending autoNav")
		}
		beforeStop := sim.entities[sim.playerID].pos
		stop := sim.Tick([]Input{{MessageID: "stop", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{}, DurationTicks: 1}}})
		assertAck(t, stop, "stop")
		if sim.autoNav != nil {
			t.Fatalf("stop did not clear pending action autoNav: %+v", sim.autoNav)
		}
		if sim.entities[sim.playerID].pos != beforeStop {
			t.Fatalf("stop tick advanced pending action from %+v to %+v", beforeStop, sim.entities[sim.playerID].pos)
		}
		for i := 0; i < 20; i++ {
			r := sim.Tick(nil)
			if hasEvent(r, "monster_damaged") || hasEvent(r, "monster_killed") {
				t.Fatalf("canceled pending action still attacked on tick %d: %+v", i, r.Events)
			}
		}
	})
}

func firstEntityByKind(sim *Sim, kind string) *entity {
	for _, id := range sortedEntityIDs(sim.entities) {
		if sim.entities[id].kind == kind {
			return sim.entities[id]
		}
	}
	return nil
}

func firstChangeEntityByType(r TickResult, kind string) *EntityView {
	for _, c := range r.Changes {
		if c.Entity != nil && c.Entity.Type == kind {
			return c.Entity
		}
	}
	return nil
}

func addTestMonster(sim *Sim, monsterDefID string, pos Vec2, hp int) *entity {
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          pos,
		spawnPos:     pos,
		hp:           hp,
		maxHP:        hp,
		monsterDefID: monsterDefID,
		lootTable:    sim.rules.Monsters[monsterDefID].LootTable,
		aiMode:       monsterAIModeIdle,
	}
	sim.activeLevel().entities[monster.id] = monster
	sim.syncCompatibilityFields()
	return monster
}

func TestMeleeReachGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name         string  `json:"name"`
			Reach        float64 `json:"reach"`
			TargetRadius float64 `json:"target_radius"`
			Distance     float64 `json:"distance"`
			InRange      bool    `json:"in_range"`
		} `json:"cases"`
	}
	loadGolden(t, "melee_reach.json", &golden)

	for _, c := range golden.Cases {
		got := meleeInRange(c.Distance, c.Reach, c.TargetRadius)
		if got != c.InRange {
			t.Fatalf("%s: meleeInRange(%v,%v,%v) = %v, want %v", c.Name, c.Distance, c.Reach, c.TargetRadius, got, c.InRange)
		}
	}
}

func TestLootRollGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		LootTable         string `json:"loot_table"`
		ExpectedItemDefID string `json:"expected_item_def_id"`
	}
	loadGolden(t, "loot_roll.json", &golden)

	// Single-entry table must yield the expected item for any draw.
	for seed := uint64(0); seed < 50; seed++ {
		rng := NewRNG(seed)
		got, ok := r.RollLoot(golden.LootTable, rng)
		if !ok || got.ItemDefID != golden.ExpectedItemDefID || got.ItemTemplateID != "" {
			t.Fatalf("roll %s with seed %d = (%q,%v), want %q", golden.LootTable, seed, got, ok, golden.ExpectedItemDefID)
		}
	}
}

func TestItemRollsGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		TemplateID string `json:"template_id"`
		Cases      []struct {
			Name     string          `json:"name"`
			Seed     string          `json:"seed"`
			Expected ItemRollPayload `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "item_rolls.json", &golden)

	for i, c := range golden.Cases {
		sim := MustNewSim("sess_item_roll_"+c.Name, c.Seed, r)
		got, ok := sim.rollItemTemplate(golden.TemplateID, 1)
		if !ok {
			t.Fatalf("%s: rollItemTemplate returned false", c.Name)
		}
		if *update {
			// Normalize nil slices to empty so GDScript "as Array" casts work.
			if got.EffectIDs == nil {
				got.EffectIDs = []string{}
			}
			golden.Cases[i].Expected = got
			continue
		}
		if got.ItemTemplateID != c.Expected.ItemTemplateID ||
			got.DisplayName != c.Expected.DisplayName ||
			got.Rarity != c.Expected.Rarity ||
			!sameIntMap(got.Stats, c.Expected.Stats) ||
			!sameIntMap(got.Requirements, c.Expected.Requirements) ||
			!sameStringSlice(got.EffectIDs, c.Expected.EffectIDs) {
			t.Fatalf("%s: rolled payload = %+v, want %+v", c.Name, got, c.Expected)
		}
	}
	if *update {
		writeGolden(t, "item_rolls.json", golden)
	}
}

func TestDepthScaledAttributeAffixRanges(t *testing.T) {
	minValue, maxValue := scaledAttributeRollRange(1)
	if minValue != 1 || maxValue != 3 {
		t.Fatalf("depth 1 attribute range = %d-%d, want 1-3", minValue, maxValue)
	}
	_, maxValue = scaledAttributeRollRange(100)
	if maxValue != 50 {
		t.Fatalf("depth 100 attribute max = %d, want 50", maxValue)
	}
	if _, _, ok := scaledAllSkillsRollRange(9); ok {
		t.Fatal("all_skills should not roll before depth 10")
	}
	if minValue, maxValue, ok := scaledAllSkillsRollRange(25); !ok || minValue != 1 || maxValue != 2 {
		t.Fatalf("depth 25 all_skills range = %d-%d ok=%v, want 1-2 true", minValue, maxValue, ok)
	}
	if minValue, maxValue, ok := scaledAllSkillsRollRange(100); !ok || minValue != 1 || maxValue != 10 {
		t.Fatalf("depth 100 all_skills range = %d-%d ok=%v, want 1-10 true", minValue, maxValue, ok)
	}
}

func TestGeneratedMagicAndRareAffixPools(t *testing.T) {
	rules := loadRules(t)
	base := []RollableStatDef{}
	magic := rules.rollableStatsForRarity(base, "magic", 1)
	for _, stat := range []string{"str", "dex", "vit", "magic"} {
		roll, ok := findRollableStat(magic, stat)
		if !ok {
			t.Fatalf("magic pool missing %s", stat)
		}
		if roll.Min != 1 || roll.Max != 3 {
			t.Fatalf("magic %s range = %d-%d, want 1-3", stat, roll.Min, roll.Max)
		}
	}
	if _, ok := findRollableStat(magic, "all_skills"); ok {
		t.Fatal("magic pool should not include all_skills")
	}
	rare := rules.rollableStatsForRarity(base, "rare", 20)
	roll, ok := findRollableStat(rare, "all_skills")
	if !ok {
		t.Fatal("rare depth 20 pool missing all_skills")
	}
	if roll.Min != 1 || roll.Max != 2 {
		t.Fatalf("rare depth 20 all_skills range = %d-%d, want 1-2", roll.Min, roll.Max)
	}
}

func TestJewelryTemplatesCanRollInventoryRows(t *testing.T) {
	rules := loadRules(t)
	for _, templateID := range []string{"cave_ring", "cave_amulet"} {
		template := rules.ItemTemplates[templateID]
		roll, ok := findRollableStat(template.RollableStats, "inventory_rows")
		if !ok {
			t.Fatalf("%s rollable stats missing inventory_rows", templateID)
		}
		if roll.Min < 1 || roll.Max < roll.Min {
			t.Fatalf("%s inventory_rows roll range = %d-%d, want positive valid range", templateID, roll.Min, roll.Max)
		}
	}
}

func findRollableStat(stats []RollableStatDef, stat string) (RollableStatDef, bool) {
	for _, roll := range stats {
		if roll.Stat == stat {
			return roll, true
		}
	}
	return RollableStatDef{}, false
}

func TestUniqueEffectRollsRespectItemTypeCompatibility(t *testing.T) {
	r := loadRules(t)
	shield := r.ItemTemplates["cave_shield"]
	if shield.ItemType != "shield" {
		t.Fatalf("cave_shield item_type = %q, want shield", shield.ItemType)
	}
	for seed := uint64(0); seed < 25; seed++ {
		got, ok := r.rollUniqueEffectForTemplate(shield, NewRNG(seed))
		if !ok {
			t.Fatalf("seed %d: expected compatible unique effect for shield", seed)
		}
		effect := r.UniqueEffects[got]
		if !containsString(effect.CompatibleItemTypes, shield.ItemType) {
			t.Fatalf("seed %d: effect %s is not compatible with %s", seed, got, shield.ItemType)
		}
		if got == "echoing_finish" {
			t.Fatalf("seed %d: shield selected weapon-only effect %s", seed, got)
		}
	}
}

func TestAllEnabledUniqueEffectsReachACompatibleTemplateRoll(t *testing.T) {
	r := loadRules(t)
	for effectID, effect := range r.UniqueEffects {
		if !effect.Enabled || effect.Status != "ready" {
			continue
		}
		if !uniqueEffectReachableByTemplateRoll(r, effectID, effect) {
			t.Fatalf("enabled unique effect %s was not selected by any compatible template roll", effectID)
		}
	}
}

func uniqueEffectReachableByTemplateRoll(r *Rules, effectID string, effect UniqueEffectDef) bool {
	for _, templateID := range sortedStringKeys(r.ItemTemplates) {
		template := r.ItemTemplates[templateID]
		if !containsString(effect.CompatibleItemTypes, template.ItemType) {
			continue
		}
		for seed := uint64(0); seed < 4096; seed++ {
			got, ok := r.rollUniqueEffectForTemplate(template, NewRNG(seed))
			if ok && got == effectID {
				return true
			}
		}
	}
	return false
}

func TestTreasureClassRollsGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		TreasureClassID string `json:"treasure_class_id"`
		Cases           []struct {
			Name          string     `json:"name"`
			Seed          string     `json:"seed"`
			ExpectedDrops []LootDrop `json:"expected_drops"`
		} `json:"cases"`
	}
	loadGolden(t, "treasure_class_rolls.json", &golden)

	for i, c := range golden.Cases {
		got := rules.RollTreasureClass(golden.TreasureClassID, NewRNG(SeedToUint64(c.Seed)))
		if *update {
			golden.Cases[i].ExpectedDrops = got
			continue
		}
		if len(got) != len(c.ExpectedDrops) {
			t.Fatalf("%s: drops = %+v, want %+v", c.Name, got, c.ExpectedDrops)
		}
		for j := range got {
			if got[j] != c.ExpectedDrops[j] {
				t.Fatalf("%s: drop %d = %+v, want %+v", c.Name, j, got[j], c.ExpectedDrops[j])
			}
		}
	}
	if *update {
		writeGolden(t, "treasure_class_rolls.json", golden)
	}
}

func TestDungeonEquipmentDropsGolden(t *testing.T) {
	rules := loadRules(t)
	var golden struct {
		WorldID           string   `json:"world_id"`
		RequiredTemplates []string `json:"required_templates"`
		Bands             []struct {
			Level            int    `json:"level"`
			Depth            int    `json:"depth"`
			MonsterLootTable string `json:"monster_loot_table"`
			ChestLootTable   string `json:"chest_loot_table"`
		} `json:"bands"`
		Cases []struct {
			Name            string     `json:"name"`
			Level           int        `json:"level"`
			Source          string     `json:"source"`
			LootTable       string     `json:"loot_table"`
			TreasureClassID string     `json:"treasure_class_id"`
			Seed            string     `json:"seed"`
			ExpectedDrops   []LootDrop `json:"expected_drops"`
		} `json:"cases"`
	}
	loadGolden(t, "dungeon_equipment_drops.json", &golden)

	if golden.WorldID != "dungeon_levels" {
		t.Fatalf("world_id = %s, want dungeon_levels", golden.WorldID)
	}
	for _, bandGolden := range golden.Bands {
		band, ok := rules.DungeonGeneration.LootBandForLevel(bandGolden.Level)
		if !ok {
			t.Fatalf("level %d missing loot band", bandGolden.Level)
		}
		if got := absInt(bandGolden.Level); got != bandGolden.Depth {
			t.Fatalf("level %d depth = %d, want %d", bandGolden.Level, got, bandGolden.Depth)
		}
		if band.MonsterLootTable != bandGolden.MonsterLootTable || band.ChestLootTable != bandGolden.ChestLootTable {
			t.Fatalf("level %d band = %+v, want monster=%s chest=%s", bandGolden.Level, band, bandGolden.MonsterLootTable, bandGolden.ChestLootTable)
		}
	}

	reachable := rules.templatesReachableFromLootTable("dungeon_mob_drop_depth_3_plus")
	for templateID := range rules.templatesReachableFromLootTable("guarded_chest_drop_depth_3_plus") {
		reachable[templateID] = true
	}
	for _, templateID := range golden.RequiredTemplates {
		if !reachable[templateID] {
			t.Fatalf("3+ dungeon sources cannot reach required template %s", templateID)
		}
	}

	for _, c := range golden.Cases {
		table := rules.LootTables[c.LootTable]
		if table.TreasureClassID != c.TreasureClassID {
			t.Fatalf("%s: treasure class = %s, want %s", c.Name, table.TreasureClassID, c.TreasureClassID)
		}
		band, ok := rules.DungeonGeneration.LootBandForLevel(c.Level)
		if !ok {
			t.Fatalf("%s: missing band for level %d", c.Name, c.Level)
		}
		wantTable := band.MonsterLootTable
		if c.Source == "chest" {
			wantTable = band.ChestLootTable
		}
		if c.LootTable != wantTable {
			t.Fatalf("%s: loot table = %s, want %s", c.Name, c.LootTable, wantTable)
		}
		got := rules.LootDrops(c.LootTable, NewRNG(SeedToUint64(c.Seed)))
		if len(got) != len(c.ExpectedDrops) {
			t.Fatalf("%s: drops = %+v, want %+v", c.Name, got, c.ExpectedDrops)
		}
		for i := range got {
			if got[i] != c.ExpectedDrops[i] {
				t.Fatalf("%s: drop %d = %+v, want %+v", c.Name, i, got[i], c.ExpectedDrops[i])
			}
		}
	}
}

func TestDungeonMonsterLootRate(t *testing.T) {
	rules := loadRules(t)
	assertDungeonMonsterLootRate(t, rules, float64(rules.MainConfig.Gameplay.BaseDropRatePercent)/100.0)
}

func TestMainConfigDropRateDrivesDungeonMonsterLoot(t *testing.T) {
	rules := loadRulesWithMainGameplay(t, map[string]any{
		"base_drop_rate_percent": 35,
	})
	assertDungeonMonsterLootRate(t, rules, 0.35)
}

func assertDungeonMonsterLootRate(t *testing.T, rules *Rules, want float64) {
	t.Helper()
	tableIDs := []string{"dungeon_mob_drop"}
	for _, band := range rules.DungeonGeneration.LootBands {
		tableIDs = append(tableIDs, band.MonsterLootTable)
	}
	seen := map[string]bool{}
	for _, tableID := range tableIDs {
		if seen[tableID] {
			continue
		}
		seen[tableID] = true
		table := rules.LootTables[tableID]
		got := treasureClassAtLeastOneDropChance(rules.TreasureClasses[table.TreasureClassID])
		if math.Abs(got-want) > 0.000001 {
			t.Fatalf("%s at-least-one drop chance = %.4f, want %.4f", tableID, got, want)
		}
	}
}

func treasureClassAtLeastOneDropChance(tc TreasureClassDef) float64 {
	noDropChance := 1.0
	for _, attempt := range tc.Attempts {
		total := attempt.SuccessWeight + attempt.NoDropWeight
		if total <= 0 {
			continue
		}
		noDropChance *= float64(attempt.NoDropWeight) / float64(total)
	}
	return 1 - noDropChance
}

func TestRolledTemplateLootTransfersToInventory(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.TreasureClasses["test_rolled_tc"] = TreasureClassDef{Attempts: []TreasureAttemptDef{{
		AttemptID:     "rolled",
		SuccessWeight: 1,
		NoDropWeight:  0,
		Entries: []TreasureClassEntry{{
			ItemTemplateID: "cave_blade",
			Weight:         1,
		}},
	}}}
	rules.LootTables["test_rolled_drop"] = LootTable{TreasureClassID: "test_rolled_tc"}
	sim := MustNewSim("sess_rolled_loot", "0000000000000004", rules)
	player := sim.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 0.5, Y: player.pos.Y},
		hp:           1,
		maxHP:        1,
		monsterDefID: "dungeon_mob",
		lootTable:    "test_rolled_drop",
	}
	sim.entities[monster.id] = monster

	kill := sim.Tick([]Input{{
		MessageID:     "kill_rolled",
		CorrelationID: "corr_rolled",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(monster.id)},
	}})
	assertAck(t, kill, "kill_rolled")
	if !hasEvent(kill, "loot_dropped") {
		t.Fatalf("missing loot_dropped: %+v", kill.Events)
	}
	var loot *entity
	for _, e := range sim.entities {
		if e.kind == lootEntity && e.rollPayload != nil {
			loot = e
			break
		}
	}
	if loot == nil {
		t.Fatalf("missing rolled loot entity: %+v", sim.entities)
	}
	if loot.itemDefID != "cave_blade" || loot.rollPayload.ItemTemplateID != "cave_blade" || loot.rollPayload.Rarity == "" {
		t.Fatalf("rolled loot payload = itemDefID %q payload %+v", loot.itemDefID, loot.rollPayload)
	}
	lootView := loot.view()
	if lootView.ItemTemplateID != "cave_blade" || lootView.DisplayName == "" || lootView.RolledStats["damage_max"] == 0 {
		t.Fatalf("loot view missing rolled fields: %+v", lootView)
	}

	pickup := sim.Tick([]Input{{
		MessageID:     "pickup_rolled",
		CorrelationID: "corr_pickup",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(loot.id)},
	}})
	assertAck(t, pickup, "pickup_rolled")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory size = %d, want 1", len(sim.inventory))
	}
	got := sim.inventory[0].view()
	if got.ItemDefID != "cave_blade" || got.ItemTemplateID != "cave_blade" || got.DisplayName != loot.rollPayload.DisplayName {
		t.Fatalf("inventory rolled view = %+v, loot payload %+v", got, loot.rollPayload)
	}
	if !sameIntMap(got.RolledStats, loot.rollPayload.Stats) {
		t.Fatalf("inventory rolled stats = %+v, want %+v", got.RolledStats, loot.rollPayload.Stats)
	}
}

func TestLegacyLootHasNoRolledPayload(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_legacy_loot", "01", rules)
	player := sim.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 0.5, Y: player.pos.Y},
		hp:           1,
		maxHP:        1,
		monsterDefID: "training_dummy",
		lootTable:    "basic_drop",
	}
	sim.entities[monster.id] = monster
	res := sim.Tick([]Input{{MessageID: "kill_legacy", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, res, "kill_legacy")
	for _, e := range sim.entities {
		if e.kind == lootEntity && e.itemDefID == "rusty_sword" {
			if e.rollPayload != nil {
				t.Fatalf("legacy loot has rolled payload: %+v", e.rollPayload)
			}
			return
		}
	}
	t.Fatal("missing legacy rusty_sword loot")
}

func TestRolledWeaponDamageOverridesStaticFallback(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_rolled_damage", "01", rules)
	player := sim.entities[sim.playerID]
	item := &invItem{
		instanceID: 5000,
		itemDefID:  "cave_blade",
		slot:       mainHandSlot,
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "cave_blade",
			DisplayName:    "Test Cave Blade",
			Rarity:         "rare",
			Stats:          map[string]int{"damage_min": 7, "damage_max": 7, "max_hp": 3},
			Requirements:   map[string]int{"level": 1},
			EffectIDs:      []string{},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped[mainHandSlot] = item.instanceID
	sim.savePlayer(sim.defaultPlayer())
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 0.5, Y: player.pos.Y},
		hp:           10,
		maxHP:        10,
		monsterDefID: "training_dummy_reward",
		lootTable:    "reward_drop",
	}
	sim.entities[monster.id] = monster
	res := sim.Tick([]Input{{MessageID: "rolled_hit", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, res, "rolled_hit")
	assertEventDamage(t, res, "monster_damaged", 7)
	if player.hp != playerStartHP-1 {
		t.Fatalf("rolled max_hp should be display-only; player hp = %d", player.hp)
	}
}

func TestRolledBaseStatsAndAllSkillsApplyWhenEquipped(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_rolled_base_stats", "01", rules)
	sim.progression.SkillRanks["magic_bolt"] = 1
	beforeStats := sim.effectiveBaseStatsView()
	beforeDerived := sim.characterDerivedStatsView()
	item := &invItem{
		instanceID: 5200,
		itemDefID:  "cave_ring",
		slot:       ringLeftSlot,
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "cave_ring",
			DisplayName:    "Rare Test Ring",
			Rarity:         "rare",
			Stats: map[string]int{
				"magic":      10,
				"vit":        8,
				"all_skills": 2,
			},
			Requirements: map[string]int{"level": 1},
			EffectIDs:    []string{},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped[ringLeftSlot] = item.instanceID

	afterStats := sim.effectiveBaseStatsView()
	if afterStats.Magic != beforeStats.Magic+10 || afterStats.Vit != beforeStats.Vit+8 {
		t.Fatalf("effective base stats = %+v, before %+v", afterStats, beforeStats)
	}
	afterDerived := sim.characterDerivedStatsView()
	if afterDerived.MaxMana <= beforeDerived.MaxMana || afterDerived.MaxHP <= beforeDerived.MaxHP {
		t.Fatalf("derived stats did not improve: before %+v after %+v", beforeDerived, afterDerived)
	}
	if rank := sim.effectiveSkillRank("magic_bolt"); rank != 3 {
		t.Fatalf("effective magic_bolt rank = %d, want 3", rank)
	}
	view := sim.SkillProgressionView()
	var shownRank int
	for _, skill := range view.Skills {
		if skill.SkillID == "magic_bolt" {
			shownRank = skill.Rank
			break
		}
	}
	if shownRank != 3 {
		t.Fatalf("skill progression magic_bolt rank = %d, want 3", shownRank)
	}
}

func TestFullEquipmentSlotsGolden(t *testing.T) {
	var golden struct {
		EquipmentSlots []string `json:"equipment_slots"`
	}
	loadGolden(t, "full_equipment.json", &golden)
	if len(golden.EquipmentSlots) != len(equipmentSlots) {
		t.Fatalf("equipment slots = %v, want %v", equipmentSlots, golden.EquipmentSlots)
	}
	for i, want := range golden.EquipmentSlots {
		if equipmentSlots[i] != want {
			t.Fatalf("equipment slot[%d] = %q, want %q", i, equipmentSlots[i], want)
		}
	}
	snap := MustNewSim("sess_equipment_slots", "01", loadRules(t)).Snapshot()
	for _, slot := range golden.EquipmentSlots {
		if _, ok := snap.Equipped[slot]; !ok {
			t.Fatalf("snapshot missing equipped slot %q: %+v", slot, snap.Equipped)
		}
	}
}

func TestEquipmentSlotCompatibilityAndRings(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_equipment_slots", "01", rules)
	cases := []struct {
		templateID string
		slot       string
	}{
		{"cave_helm", "head"},
		{"cave_amulet", "amulet"},
		{"cave_mail", "chest"},
		{"cave_gloves", "gloves"},
		{"cave_belt", "belt"},
		{"cave_boots", "boots"},
		{"cave_ring", ringLeftSlot},
		{"cave_ring", ringRightSlot},
		{"cave_blade", mainHandSlot},
		{"cave_shield", offHandSlot},
	}
	for i, tc := range cases {
		item := addRolledInventoryItem(t, sim, uint64(6000+i), tc.templateID, nil)
		res := sim.Tick([]Input{{
			MessageID: "equip_" + tc.templateID + "_" + tc.slot,
			Type:      "equip_intent",
			Equip:     &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: tc.slot},
		}})
		assertAck(t, res, "equip_"+tc.templateID+"_"+tc.slot)
		if sim.equipped[tc.slot] != item.instanceID || !item.equipped || item.slot != tc.slot {
			t.Fatalf("%s in %s equipped=%d item=%+v", tc.templateID, tc.slot, sim.equipped[tc.slot], item)
		}
	}
}

func TestEquipmentWrongSlotRejects(t *testing.T) {
	sim := MustNewSim("sess_wrong_slot", "01", loadRules(t))
	shield := addRolledInventoryItem(t, sim, 6100, "cave_shield", nil)
	res := sim.Tick([]Input{{
		MessageID: "wrong",
		Type:      "equip_intent",
		Equip:     &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: mainHandSlot},
	}})
	assertReject(t, res, "wrong", "wrong_slot")
}

func TestRogueOffhandWeaponEquipRules(t *testing.T) {
	rules := loadRules(t)
	rogueState := rules.DefaultCharacterProgressionState()
	rogueState.CharacterClass = "rogue"
	rogueState.BaseStats = rules.CharacterProgression.Classes["rogue"].BaseStats
	if req := rules.ItemTemplates["starter_rogue_sword"].Requirements["str"]; rogueState.BaseStats.Str < req {
		rogueState.BaseStats.Str = req
	}
	if req := rules.ItemTemplates["cave_blade"].Requirements["str"]; rogueState.BaseStats.Str < req {
		rogueState.BaseStats.Str = req
	}
	rogue, err := NewSimWithWorldProgression("sess_rogue_offhand", "rogue_offhand_seed", rules, DefaultWorldID, rogueState)
	if err != nil {
		t.Fatalf("new rogue sim: %v", err)
	}
	main := addRolledInventoryItem(t, rogue, 6110, "starter_rogue_sword", nil)
	off := addRolledInventoryItem(t, rogue, 6111, "cave_blade", nil)
	mainResult := rogue.Tick([]Input{{MessageID: "main", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(main.instanceID), Slot: mainHandSlot}}})
	assertAck(t, mainResult, "main")
	offResult := rogue.Tick([]Input{{MessageID: "off", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(off.instanceID), Slot: offHandSlot}}})
	assertAck(t, offResult, "off")
	if rogue.equipped[mainHandSlot] != main.instanceID || rogue.equipped[offHandSlot] != off.instanceID {
		t.Fatalf("rogue equipped = %+v, want main %d off %d", rogue.equipped, main.instanceID, off.instanceID)
	}
	if off.slot != offHandSlot {
		t.Fatalf("offhand weapon slot = %q, want %q", off.slot, offHandSlot)
	}

	barbarian := MustNewSim("sess_barbarian_offhand", "barbarian_offhand_seed", rules)
	barbarianWeapon := addRolledInventoryItem(t, barbarian, 6120, "cave_blade", nil)
	barbarianResult := barbarian.Tick([]Input{{MessageID: "barb_off", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(barbarianWeapon.instanceID), Slot: offHandSlot}}})
	assertReject(t, barbarianResult, "barb_off", "wrong_slot")

	twoHand := addRolledInventoryItem(t, rogue, 6130, "starter_sorcerer_staff", nil)
	twoHandResult := rogue.Tick([]Input{{MessageID: "two_hand_off", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(twoHand.instanceID), Slot: offHandSlot}}})
	assertReject(t, twoHandResult, "two_hand_off", "wrong_slot")
}

func TestEquipmentRequirementsRejectAndPreview(t *testing.T) {
	var golden struct {
		TemplateID      string                     `json:"template_id"`
		Requirements    map[string]int             `json:"requirements"`
		FreshCharacter  equipmentRequirementGolden `json:"fresh_character"`
		AfterAllocation equipmentRequirementGolden `json:"after_allocation"`
		ExpectedReject  string                     `json:"expected_reject"`
	}
	loadGolden(t, "equipment_requirements.json", &golden)

	sim := MustNewSim("sess_equipment_requirements", "01", loadRules(t))
	item := addRolledInventoryItem(t, sim, 6110, golden.TemplateID, map[string]int{"damage_min": 5, "damage_max": 8})
	view := sim.itemView(item)
	assertRequirementStatus(t, view.RequirementStatus, golden.FreshCharacter.Status)
	if view.RequirementsMet == nil || *view.RequirementsMet != golden.FreshCharacter.RequirementsMet {
		t.Fatalf("fresh requirements_met = %v, want %v", view.RequirementsMet, golden.FreshCharacter.RequirementsMet)
	}
	if view.EquipPreview == nil || view.EquipPreview.Slot != mainHandSlot || view.EquipPreview.RequirementsMet {
		t.Fatalf("fresh equip preview = %+v, want unmet main-hand preview", view.EquipPreview)
	}
	if findPreviewDelta(view.EquipPreview.Deltas, "damage_max") == nil {
		t.Fatalf("fresh equip preview missing damage_max delta: %+v", view.EquipPreview.Deltas)
	}

	beforeSlot := item.slot
	beforeEquipped := item.equipped
	reject := sim.Tick([]Input{{
		MessageID: "equip_unmet",
		Type:      "equip_intent",
		Equip:     &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: mainHandSlot},
	}})
	assertReject(t, reject, "equip_unmet", golden.ExpectedReject)
	if item.equipped != beforeEquipped || item.slot != beforeSlot || sim.equipped[mainHandSlot] != 0 {
		t.Fatalf("unmet equip mutated item=%+v equipped=%v", item, sim.equipped)
	}

	sim.progression.Level = golden.AfterAllocation.Level
	sim.progression.UnspentStatPoints = 1
	sim.savePlayer(sim.defaultPlayer())

	allocate := sim.Tick([]Input{{
		MessageID: "allocate_str",
		Type:      "allocate_stat_intent",
		AllocateStat: &AllocateStatIntent{
			Stat:   "str",
			Points: 1,
		},
	}})
	assertAck(t, allocate, "allocate_str")
	update := inventoryUpdateForItem(allocate, item.instanceID)
	if update == nil {
		t.Fatalf("stat allocation did not refresh inventory item: %+v", allocate.Changes)
	}
	assertRequirementStatus(t, update.RequirementStatus, golden.AfterAllocation.Status)
	if update.RequirementsMet == nil || !*update.RequirementsMet || update.EquipPreview == nil || !update.EquipPreview.RequirementsMet {
		t.Fatalf("allocation inventory update = %+v", update)
	}

	view = sim.itemView(item)
	assertRequirementStatus(t, view.RequirementStatus, golden.AfterAllocation.Status)
	if view.RequirementsMet == nil || !*view.RequirementsMet || view.EquipPreview == nil || !view.EquipPreview.RequirementsMet {
		t.Fatalf("met item view = %+v", view)
	}
	equip := sim.Tick([]Input{{
		MessageID: "equip_met",
		Type:      "equip_intent",
		Equip:     &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: mainHandSlot},
	}})
	assertAck(t, equip, "equip_met")
	if sim.equipped[mainHandSlot] != item.instanceID || !item.equipped {
		t.Fatalf("met equip failed item=%+v equipped=%v", item, sim.equipped)
	}
}

func TestHandOccupancyAndPrimaryWeaponGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name     string `json:"name"`
			Expected struct {
				Equipped      map[string]*string `json:"equipped"`
				OccupiesHands []string           `json:"occupies_hands"`
				RolledStats   map[string]int     `json:"rolled_stats"`
				CombatEffect  string             `json:"combat_effect"`
			} `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "full_equipment.json", &golden)
	expected := map[string]struct {
		equipped     map[string]*string
		occupies     []string
		rolledStats  map[string]int
		combatEffect string
	}{}
	for _, c := range golden.Cases {
		expected[c.Name] = struct {
			equipped     map[string]*string
			occupies     []string
			rolledStats  map[string]int
			combatEffect string
		}{c.Expected.Equipped, c.Expected.OccupiesHands, c.Expected.RolledStats, c.Expected.CombatEffect}
	}

	t.Run("one hand sword and shield can coexist", func(t *testing.T) {
		sim := MustNewSim("sess_one_hand_shield", "01", loadRules(t))
		sword := addRolledInventoryItem(t, sim, 6200, "cave_blade", map[string]int{"damage_min": 4, "damage_max": 5})
		shield := addRolledInventoryItem(t, sim, 6201, "cave_shield", map[string]int{"armor": 3, "block_percent": 8})
		assertAck(t, sim.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
		assertAck(t, sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}), "shield")
		want := expected["one hand sword and shield can coexist"].equipped
		assertEquippedTemplate(t, sim, mainHandSlot, *want[mainHandSlot])
		assertEquippedTemplate(t, sim, offHandSlot, *want[offHandSlot])
		if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 4, Max: 5}) {
			t.Fatalf("primary attack damage = %+v, want rolled sword 4..5", got)
		}
	})

	t.Run("two handed sword clears offhand", func(t *testing.T) {
		sim := MustNewSim("sess_two_hand_clear", "01", loadRules(t))
		shield := addRolledInventoryItem(t, sim, 6210, "cave_shield", nil)
		greatsword := addRolledInventoryItem(t, sim, 6211, "cave_greatsword", nil)
		assertAck(t, sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}), "shield")
		res := sim.Tick([]Input{{MessageID: "greatsword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}})
		assertAck(t, res, "greatsword")
		want := expected["two handed sword clears offhand"].equipped
		assertEquippedTemplate(t, sim, mainHandSlot, *want[mainHandSlot])
		if want[offHandSlot] != nil || sim.equipped[offHandSlot] != 0 || shield.equipped {
			t.Fatalf("offhand after two-hand equip = %d shield=%+v", sim.equipped[offHandSlot], shield)
		}
		if !hasEquippedUpdate(res, offHandSlot, nil) {
			t.Fatalf("missing offhand clear change: %+v", res.Changes)
		}
	})

	t.Run("two handed sword clears occupied hands", func(t *testing.T) {
		sim := MustNewSim("sess_two_hand_clear_both", "01", loadRules(t))
		sword := addRolledInventoryItem(t, sim, 6212, "cave_blade", nil)
		shield := addRolledInventoryItem(t, sim, 6213, "cave_shield", nil)
		greatsword := addRolledInventoryItem(t, sim, 6214, "cave_greatsword", nil)
		assertAck(t, sim.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
		assertAck(t, sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}), "shield")

		res := sim.Tick([]Input{{MessageID: "greatsword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}})
		assertAck(t, res, "greatsword")
		assertEquippedTemplate(t, sim, mainHandSlot, "cave_greatsword")
		if sword.equipped || shield.equipped || sim.equipped[offHandSlot] != 0 {
			t.Fatalf("2H equip did not clear occupied hands: sword=%+v shield=%+v equipped=%v", sword, shield, sim.equipped)
		}
		if !hasEquippedUpdate(res, mainHandSlot, stringPtr(idStr(greatsword.instanceID))) || !hasEquippedUpdate(res, offHandSlot, nil) {
			t.Fatalf("missing hand slot updates for 2H replacement: %+v", res.Changes)
		}
	})

	t.Run("two handed sword full bag reject preserves occupied hands", func(t *testing.T) {
		sim := MustNewSim("sess_two_hand_full_bag_preserve", "01", loadRules(t))
		sword := addRolledInventoryItem(t, sim, 6215, "cave_blade", nil)
		shield := addRolledInventoryItem(t, sim, 6216, "cave_shield", nil)
		greatsword := addRolledInventoryItem(t, sim, 6217, "cave_greatsword", nil)
		assertAck(t, sim.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
		assertAck(t, sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}), "shield")
		for i := 0; i < inventoryCapacityForRows(baseInventoryRows)-1; i++ {
			addStaticInventoryItem(sim, uint64(6218+i), "quest_leaf")
		}
		if got := sim.bagOccupancyCount(); got != inventoryCapacityForRows(baseInventoryRows) {
			t.Fatalf("setup bag occupancy = %d, want full capacity", got)
		}

		res := sim.Tick([]Input{{MessageID: "greatsword_full", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}})
		assertReject(t, res, "greatsword_full", "capacity_would_overflow")
		if !sword.equipped || !shield.equipped || greatsword.equipped || sim.equipped[mainHandSlot] != sword.instanceID || sim.equipped[offHandSlot] != shield.instanceID {
			t.Fatalf("full-bag 2H reject mutated hands: sword=%+v shield=%+v greatsword=%+v equipped=%v", sword, shield, greatsword, sim.equipped)
		}
	})

	t.Run("offhand blocked by two handed main hand", func(t *testing.T) {
		sim := MustNewSim("sess_two_hand_block", "01", loadRules(t))
		bow := addRolledInventoryItem(t, sim, 6220, "cave_bow", nil)
		shield := addRolledInventoryItem(t, sim, 6221, "cave_shield", nil)
		assertAck(t, sim.Tick([]Input{{MessageID: "bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "bow")
		res := sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}})
		assertReject(t, res, "shield", "hands_blocked")
	})

	t.Run("bow occupies both hands", func(t *testing.T) {
		sim := MustNewSim("sess_bow_occupies", "01", loadRules(t))
		bow := addRolledInventoryItem(t, sim, 6230, "cave_bow", map[string]int{"damage_min": 2, "damage_max": 2})
		assertAck(t, sim.Tick([]Input{{MessageID: "bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "bow")
		want := expected["bow occupies both hands"]
		assertEquippedTemplate(t, sim, mainHandSlot, *want.equipped[mainHandSlot])
		if sim.equipped[offHandSlot] != 0 {
			t.Fatalf("offhand = %d, want empty for bow", sim.equipped[offHandSlot])
		}
		if got := sim.itemOccupiesHands(bow); !sameStrings(got, want.occupies) {
			t.Fatalf("bow occupies_hands = %v, want %v", got, want.occupies)
		}
		if mode := sim.playerAttackMode(); mode != attackModeRanged {
			t.Fatalf("bow attack mode = %q, want ranged", mode)
		}
	})

	t.Run("shield display rolls do not affect combat yet", func(t *testing.T) {
		sim := MustNewSim("sess_shield_display", "01", loadRules(t))
		sword := addRolledInventoryItem(t, sim, 6240, "cave_blade", map[string]int{"damage_min": 4, "damage_max": 4})
		shieldStats := expected["shield display rolls do not affect combat yet"].rolledStats
		shield := addRolledInventoryItem(t, sim, 6241, "cave_shield", shieldStats)
		assertAck(t, sim.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
		assertAck(t, sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}}), "shield")
		if got := shield.view().RolledStats; got["armor"] != shieldStats["armor"] || got["block_percent"] != shieldStats["block_percent"] {
			t.Fatalf("shield display stats = %v, want %v", got, shieldStats)
		}
		if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 4, Max: 4}) {
			t.Fatalf("shield affected attack damage = %+v", got)
		}
	})
}

func TestCombatStatBreakdownsIncludeEquipmentAndCap(t *testing.T) {
	sim := MustNewSim("sess_combat_breakdown", "01", loadRules(t))
	sword := addRolledInventoryItem(t, sim, 6250, "cave_blade", map[string]int{"damage_min": 4, "damage_max": 6})
	shield := addRolledInventoryItem(t, sim, 6251, "cave_shield", map[string]int{"armor": 5, "block_percent": 82})
	ring := addRolledInventoryItem(t, sim, 6252, "cave_ring", map[string]int{"max_hp": 4})
	assertAck(t, sim.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
	shieldResult := sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}})
	assertAck(t, shieldResult, "shield")
	assertAck(t, sim.Tick([]Input{{MessageID: "ring", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(ring.instanceID), Slot: ringLeftSlot}}}), "ring")

	shieldUpdate := characterProgressionUpdate(shieldResult)
	if shieldUpdate == nil {
		t.Fatalf("shield equip did not publish character progression update: %+v", shieldResult.Changes)
	}
	shieldBlock := findStatBreakdown(shieldUpdate.StatBreakdowns, "block_percent")
	if shieldBlock == nil || !hasBreakdownSource(shieldBlock.Sources, "equipment_base") || !hasBreakdownSource(shieldBlock.Sources, "equipment_roll") {
		t.Fatalf("shield equip progression block breakdown = %+v", shieldBlock)
	}

	view := sim.CharacterProgressionView()
	if view.DerivedStats.DamageMin != 4 || view.DerivedStats.DamageMax != 6 {
		t.Fatalf("effective damage = %v..%v, want 4..6", view.DerivedStats.DamageMin, view.DerivedStats.DamageMax)
	}
	if view.DerivedStats.Armor != 6 || view.DerivedStats.MaxHP != 14 {
		t.Fatalf("effective armor/maxHP = %v/%v, want 6/14", view.DerivedStats.Armor, view.DerivedStats.MaxHP)
	}

	block := findStatBreakdown(view.StatBreakdowns, "block_percent")
	if block == nil {
		t.Fatalf("missing block breakdown: %+v", view.StatBreakdowns)
	}
	if block.Value != 75 || block.UncappedValue != 82 || block.Cap == nil || *block.Cap != 75 {
		t.Fatalf("block breakdown = %+v, want capped 75 from uncapped 82", block)
	}
	if !hasBreakdownSource(block.Sources, "equipment_base") || !hasBreakdownSource(block.Sources, "equipment_roll") || !hasBreakdownSource(block.Sources, "cap") {
		t.Fatalf("block breakdown sources = %+v, want base, roll, cap", block.Sources)
	}
}

func TestCombatStatEffectsGolden(t *testing.T) {
	var golden struct {
		Cases []struct {
			Name            string `json:"name"`
			Outcome         string `json:"outcome"`
			RawDamage       int    `json:"raw_damage"`
			MitigatedDamage int    `json:"mitigated_damage"`
			FinalDamage     int    `json:"final_damage"`
			Blocked         bool   `json:"blocked"`
			Critical        bool   `json:"critical"`
		} `json:"cases"`
		StatBreakdowns []StatBreakdownView `json:"stat_breakdowns"`
	}
	loadGolden(t, "combat_stat_effects.json", &golden)

	sim := MustNewSim("sess_combat_stat_golden", "01", loadRules(t))
	for _, c := range golden.Cases {
		attacker, defender, damageRange := combatGoldenStats(c.Name)
		got := sim.resolveCombat(attacker, defender, damageRange)
		if got.Outcome != c.Outcome || got.RawDamage != c.RawDamage || got.MitigatedDamage != c.MitigatedDamage ||
			got.Damage != c.FinalDamage || got.Blocked != c.Blocked || got.Critical != c.Critical {
			t.Fatalf("%s outcome = %+v, want outcome=%s raw=%d mitigated=%d final=%d blocked=%v critical=%v",
				c.Name, got, c.Outcome, c.RawDamage, c.MitigatedDamage, c.FinalDamage, c.Blocked, c.Critical)
		}
	}

	block := findStatBreakdown(golden.StatBreakdowns, "block_percent")
	if block == nil || block.Value != 75 || block.UncappedValue <= block.Value || block.Cap == nil || *block.Cap != 75 {
		t.Fatalf("golden block breakdown = %+v, want capped 75", block)
	}
}

func TestMonsterCombatStatsEffective(t *testing.T) {
	sim := MustNewSim("sess_monster_combat_stats", "01", loadRules(t))
	monster := &entity{
		id:           7001,
		kind:         monsterEntity,
		maxHP:        12,
		hp:           12,
		monsterDefID: "combat_lab_blocking_target",
	}
	stats := sim.monsterEffectiveCombatStats(monster, DamageRange{Min: 3, Max: 5})
	if stats.DamageMin != 3 || stats.DamageMax != 5 {
		t.Fatalf("monster damage = %v..%v, want 3..5", stats.DamageMin, stats.DamageMax)
	}
	if stats.HitChance != 0.5 || stats.CritChance != 0 || stats.CritDamage != 1.5 || stats.Armor != 0 {
		t.Fatalf("monster chance/crit/armor stats = %+v", stats)
	}
	if stats.BlockPercent != 75 {
		t.Fatalf("monster block percent = %v, want capped 75", stats.BlockPercent)
	}
}

func combatGoldenStats(name string) (effectiveCombatStats, effectiveCombatStats, DamageRange) {
	alwaysHit := effectiveCombatStats{HitChance: 1, CritDamage: 1.5}
	switch name {
	case "player_miss":
		return effectiveCombatStats{HitChance: 0, CritDamage: 1.5}, effectiveCombatStats{}, DamageRange{Min: 5, Max: 5}
	case "player_crit":
		return effectiveCombatStats{HitChance: 1, CritChance: 1, CritDamage: 2}, effectiveCombatStats{}, DamageRange{Min: 5, Max: 5}
	case "monster_armor_minimum_damage":
		return alwaysHit, effectiveCombatStats{Armor: 99}, DamageRange{Min: 8, Max: 8}
	case "player_armor_minimum_damage":
		return alwaysHit, effectiveCombatStats{Armor: 6}, DamageRange{Min: 2, Max: 2}
	case "player_block", "block_cap_75", "monster_block":
		return alwaysHit, effectiveCombatStats{BlockPercent: 100}, DamageRange{Min: 2, Max: 2}
	case "monster_crit":
		return effectiveCombatStats{HitChance: 1, CritChance: 1, CritDamage: 2}, effectiveCombatStats{}, DamageRange{Min: 2, Max: 2}
	case "projectile_impact":
		return alwaysHit, effectiveCombatStats{}, DamageRange{Min: 5, Max: 5}
	default:
		return alwaysHit, effectiveCombatStats{}, DamageRange{Min: 1, Max: 1}
	}
}

// --- scripted slice ---------------------------------------------------------

// runSlice drives a sim through the full vertical-slice flow and returns it.
func runSlice(t *testing.T, seed string) *Sim {
	t.Helper()
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	sim := MustNewSim("sess_test", seed, rules)

	// Move into unarmed reach of the monster.
	sim.Tick([]Input{{MessageID: "m1", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 2}}})

	// Attack until the monster is dead.
	monsterID := "1002"
	for i := 0; i < 10; i++ {
		if e := sim.findEntity(monsterID); e == nil || e.hp == 0 {
			break
		}
		sim.Tick([]Input{{MessageID: "a" + itoa(i), CorrelationID: "corr_a", Type: "action_intent", Action: &ActionIntent{TargetID: monsterID}}})
		advanceBasicAttackCooldown(sim)
	}
	if e := sim.findEntity(monsterID); e == nil || e.hp != 0 {
		t.Fatalf("monster not dead after attacks: %+v", e)
	}

	// Find the dropped loot entity and pick it up.
	lootID := ""
	for _, ev := range sim.Snapshot().Entities {
		if ev.Type == lootEntity {
			lootID = ev.ID
		}
	}
	if lootID == "" {
		t.Fatal("no loot entity after kill")
	}
	sim.Tick([]Input{{MessageID: "p1", CorrelationID: "corr_p", Type: "action_intent", Action: &ActionIntent{TargetID: lootID}}})
	for i := 0; i < 10 && len(sim.Snapshot().Inventory) == 0; i++ {
		sim.Tick(nil)
	}

	// Equip the picked-up item.
	snap := sim.Snapshot()
	if len(snap.Inventory) != 1 {
		t.Fatalf("inventory size = %d, want 1", len(snap.Inventory))
	}
	itemID := snap.Inventory[0].ItemInstanceID
	sim.Tick([]Input{{MessageID: "e1", CorrelationID: "corr_e", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: itemID, Slot: mainHandSlot}}})

	return sim
}

func advanceBasicAttackCooldown(sim *Sim) {
	interval := sim.DerivedStatsView().AttackIntervalTicks
	for i := 1; i < interval; i++ {
		sim.Tick(nil)
	}
}

func TestScriptedSliceMatchesGolden(t *testing.T) {
	var golden struct {
		PinnedSeed       string `json:"pinned_seed"`
		MonsterDefID     string `json:"monster_def_id"`
		DroppedItemDefID string `json:"dropped_item_def_id"`
		FinalPlayerHP    int    `json:"final_player_hp"`
		FinalMonsterHP   int    `json:"final_monster_hp"`
		FinalInventory   []struct {
			ItemDefID string `json:"item_def_id"`
			Slot      string `json:"slot"`
			Equipped  bool   `json:"equipped"`
		} `json:"final_inventory"`
		FinalEquipped map[string]string `json:"final_equipped"`
	}
	loadGolden(t, "slice_outcome.json", &golden)

	sim := runSlice(t, golden.PinnedSeed)
	snap := sim.Snapshot()

	var player, monster *EntityView
	for i := range snap.Entities {
		switch snap.Entities[i].Type {
		case playerEntity:
			player = &snap.Entities[i]
		case monsterEntity:
			monster = &snap.Entities[i]
		}
	}
	if player == nil || *player.HP != golden.FinalPlayerHP {
		t.Fatalf("player hp mismatch: %+v want %d", player, golden.FinalPlayerHP)
	}
	if monster == nil || *monster.HP != golden.FinalMonsterHP {
		t.Fatalf("monster hp mismatch: %+v want %d", monster, golden.FinalMonsterHP)
	}
	if len(snap.Inventory) != len(golden.FinalInventory) {
		t.Fatalf("inventory size %d want %d", len(snap.Inventory), len(golden.FinalInventory))
	}
	got := snap.Inventory[0]
	want := golden.FinalInventory[0]
	if got.ItemDefID != want.ItemDefID || got.Slot != want.Slot || got.Equipped != want.Equipped {
		t.Fatalf("inventory item = %+v, want %+v", got, want)
	}
	// equipped main-hand instance must resolve to the expected item_def_id.
	wp := snap.Equipped[mainHandSlot]
	if wp == nil {
		t.Fatal("no main_hand equipped")
	}
	if got.ItemInstanceID != *wp || got.ItemDefID != golden.FinalEquipped[mainHandSlot] {
		t.Fatalf("equipped main_hand = %v (%s), want def %s", *wp, got.ItemDefID, golden.FinalEquipped[mainHandSlot])
	}
}

func TestSuccessfulHitRetaliatesAndPreservesKillOrder(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	sim := MustNewSim("sess_retaliate", "deadbeefdeadbeef", rules)
	sim.entities[sim.playerID].pos = Vec2{X: 11, Y: 5}
	r := sim.Tick([]Input{{
		MessageID:     "a1",
		CorrelationID: "corr_hit",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})

	assertAck(t, r, "a1")
	if len(r.Changes) != 3 {
		t.Fatalf("changes len = %d, want 3: %+v", len(r.Changes), r.Changes)
	}
	if r.Changes[0].Op != OpEntityUpdate || r.Changes[0].Entity == nil || r.Changes[0].Entity.Type != monsterEntity {
		t.Fatalf("first change = %+v, want monster entity_update", r.Changes[0])
	}
	if r.Changes[1].Op != OpEntitySpawn || r.Changes[1].Entity == nil || r.Changes[1].Entity.Type != lootEntity {
		t.Fatalf("second change = %+v, want loot entity_spawn", r.Changes[1])
	}
	if r.Changes[2].Op != OpEntityUpdate || r.Changes[2].Entity == nil || r.Changes[2].Entity.Type != playerEntity {
		t.Fatalf("third change = %+v, want player entity_update", r.Changes[2])
	}
	if r.Changes[2].Entity.HP == nil || *r.Changes[2].Entity.HP != 9 {
		t.Fatalf("player hp update = %+v, want hp 9", r.Changes[2].Entity)
	}

	wantEvents := []string{"monster_damaged", "monster_killed", "loot_dropped", "player_damaged"}
	if len(r.Events) != len(wantEvents) {
		t.Fatalf("events len = %d, want %d: %+v", len(r.Events), len(wantEvents), r.Events)
	}
	for i, want := range wantEvents {
		if r.Events[i].EventType != want || r.Events[i].CorrelationID != "corr_hit" {
			t.Fatalf("event[%d] = %+v, want %s corr_hit", i, r.Events[i], want)
		}
	}
	assertEventDamageAtLeast(t, r, "monster_damaged", 3)
	assertEventDamage(t, r, "player_damaged", 1)
	if hasEvent(r, "player_killed") {
		t.Fatalf("unexpected player_killed event: %+v", r.Events)
	}
}

func TestEquippedWeaponOneShotsRewardDummy(t *testing.T) {
	sim := gearBeforeCombatWithEquippedSword(t, loadRules(t))

	r := sim.Tick([]Input{{
		MessageID:     "a1",
		CorrelationID: "corr_weapon",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1003"},
	}})

	assertAck(t, r, "a1")
	monster := sim.findEntity("1003")
	if monster == nil || monster.hp != 0 {
		t.Fatalf("reward dummy hp = %+v, want dead", monster)
	}
	if !hasEvent(r, "monster_damaged") || !hasEvent(r, "monster_killed") || !hasEvent(r, "loot_dropped") {
		t.Fatalf("missing equipped attack events: %+v", r.Events)
	}
	assertEventDamageAtLeast(t, r, "monster_damaged", 3)
	if !hasLootSpawn(r, "gold") {
		t.Fatalf("missing gold loot spawn: %+v", r.Changes)
	}
	lootPos, ok := lootSpawnPosition(r, "gold")
	if !ok {
		t.Fatalf("missing gold loot spawn position: %+v", r.Changes)
	}
	if lootPos == monster.pos {
		t.Fatalf("loot spawned on monster body at %+v", lootPos)
	}
	if distance(lootPos, monster.pos) < monsterRadius+lootInteractionRadius {
		t.Fatalf("loot overlaps monster body: loot=%+v monster=%+v", lootPos, monster.pos)
	}
}

func TestEquippedWeaponWithoutDamageFallsBackToBaseDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sword := rules.Items["rusty_sword"]
	sword.Damage = nil
	rules.Items["rusty_sword"] = sword
	rules.Combat.PlayerDamage = DamageRange{Min: 2, Max: 2}
	sim := gearBeforeCombatWithEquippedSword(t, rules)

	r := sim.Tick([]Input{{
		MessageID:     "a1",
		CorrelationID: "corr_base",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1003"},
	}})

	assertAck(t, r, "a1")
	monster := sim.findEntity("1003")
	if monster == nil || monster.hp != 1 {
		t.Fatalf("reward dummy hp = %+v, want hp 1 from base damage fallback", monster)
	}
	if hasEvent(r, "monster_killed") || hasEvent(r, "loot_dropped") {
		t.Fatalf("fallback base hit should not kill reward dummy: %+v", r.Events)
	}
}

func TestDamageEventReportsRolledDamageNotClampedHPDelta(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Combat.PlayerDamage = DamageRange{Min: 5, Max: 5}
	sim := MustNewSim("sess_overkill_damage_event", "deadbeefdeadbeef", rules)
	sim.entities[sim.playerID].pos = Vec2{X: 11, Y: 5}
	sim.findEntity("1002").hp = 1

	r := sim.Tick([]Input{{
		MessageID:     "a1",
		CorrelationID: "corr_overkill",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})

	assertAck(t, r, "a1")
	monster := sim.findEntity("1002")
	if monster == nil || monster.hp != 0 {
		t.Fatalf("monster hp = %+v, want dead", monster)
	}
	assertEventDamage(t, r, "monster_damaged", 5)
}

func TestMissedAttackDoesNotRetaliate(t *testing.T) {
	rules := loadRules(t)
	forceCharacterHitChance(rules, 0)
	sim := MustNewSim("sess_miss", "deadbeefdeadbeef", rules)
	sim.entities[sim.playerID].pos = Vec2{X: 11, Y: 5}
	r := sim.Tick([]Input{{
		MessageID:     "a1",
		CorrelationID: "corr_miss",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})

	assertAck(t, r, "a1")
	if !hasEvent(r, "attack_missed") {
		t.Fatalf("expected attack_missed: %+v", r.Events)
	}
	if !hasEvent(r, "monster_aggro") {
		t.Fatalf("missed attack did not aggro target: %+v", r.Events)
	}
	if hasEvent(r, "player_damaged") || hasEvent(r, "player_killed") || hasPlayerUpdate(r) {
		t.Fatalf("miss retaliated unexpectedly: changes=%+v events=%+v", r.Changes, r.Events)
	}
	if sim.entities[sim.playerID].hp != playerStartHP {
		t.Fatalf("player hp = %d, want %d", sim.entities[sim.playerID].hp, playerStartHP)
	}
}

func TestBlockedAttackAggrosWithoutDamage(t *testing.T) {
	rules := loadRules(t)
	forceCharacterHitChance(rules, 1.0)
	rules.Combat.BlockCap = 100
	def := rules.Monsters[monsterDefID]
	def.BlockPercent = 100
	rules.Monsters[monsterDefID] = def
	sim := MustNewSim("sess_block_aggro", "deadbeefdeadbeef", rules)
	sim.entities[sim.playerID].pos = Vec2{X: 11, Y: 5}
	monster := sim.findEntity("1002")
	if monster == nil {
		t.Fatal("missing target monster")
	}
	beforeHP := monster.hp
	r := sim.Tick([]Input{{
		MessageID:     "blocked",
		CorrelationID: "corr_block",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})

	assertAck(t, r, "blocked")
	blocked := false
	for _, ev := range r.Events {
		if ev.EventType == "monster_damaged" && ev.Blocked != nil && *ev.Blocked {
			blocked = true
			break
		}
	}
	if !blocked {
		t.Fatalf("expected blocked monster_damaged event: %+v", r.Events)
	}
	if !hasEvent(r, "monster_aggro") {
		t.Fatalf("blocked attack did not aggro target: %+v", r.Events)
	}
	if monster.hp != beforeHP {
		t.Fatalf("blocked attack changed hp from %d to %d", beforeHP, monster.hp)
	}
}

func TestPlayerKilledByRetaliation(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	dummy := rules.Monsters[monsterDefID]
	dummy.MaxHP = 100
	rules.Monsters[monsterDefID] = dummy

	sim := MustNewSim("sess_player_death", "deadbeefdeadbeef", rules)
	sim.entities[sim.playerID].pos = Vec2{X: 11, Y: 5}
	damaged, killed := 0, 0
	for i := 0; i < playerStartHP+2; i++ {
		r := sim.Tick([]Input{{
			MessageID:     "a" + itoa(i),
			CorrelationID: "corr_death",
			Type:          "action_intent",
			Action:        &ActionIntent{TargetID: "1002"},
		}})
		for _, ev := range r.Events {
			switch ev.EventType {
			case "player_damaged":
				damaged++
			case "player_killed":
				killed++
				if hasEvent(r, "player_damaged") {
					t.Fatalf("fatal retaliation emitted paired player_damaged: %+v", r.Events)
				}
			}
		}
		if sim.entities[sim.playerID].hp == 0 {
			break
		}
		advanceBasicAttackCooldown(sim)
	}

	if sim.entities[sim.playerID].hp != 0 {
		t.Fatalf("player hp = %d, want 0", sim.entities[sim.playerID].hp)
	}
	if sim.entities[sim.playerID].hp < 0 {
		t.Fatalf("player hp went negative: %d", sim.entities[sim.playerID].hp)
	}
	if damaged == 0 || killed != 1 {
		t.Fatalf("player events damaged=%d killed=%d, want damaged>0 killed=1", damaged, killed)
	}
}

func TestUnequipWeapon(t *testing.T) {
	rules := loadRules(t)
	sim, itemID := inventoryLabEquippedSword(t, rules)

	r := sim.Tick([]Input{{
		MessageID:     "unequip",
		CorrelationID: "corr_unequip",
		Type:          "unequip_intent",
		Unequip:       &UnequipIntent{Slot: mainHandSlot},
	}})
	assertAck(t, r, "unequip")
	if sim.equipped[mainHandSlot] != 0 {
		t.Fatalf("equipped main_hand = %d, want cleared", sim.equipped[mainHandSlot])
	}
	item := sim.findItem(itemID)
	if item == nil || item.equipped {
		t.Fatalf("item after unequip = %+v, want present and unequipped", item)
	}
	if !hasEvent(r, "item_unequipped") {
		t.Fatalf("missing item_unequipped: %+v", r.Events)
	}
}

func TestDropInventoryItem(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_drop_badge", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatalf("new inventory lab: %v", err)
	}
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "quest_leaf", slot: "", equipped: false})

	r := sim.Tick([]Input{{
		MessageID:     "drop",
		CorrelationID: "corr_drop",
		Type:          "drop_intent",
		Drop:          &DropIntent{ItemInstanceID: "5000"},
	}})
	assertAck(t, r, "drop")
	if len(sim.inventory) != 0 {
		t.Fatalf("inventory after drop = %+v, want empty", sim.inventory)
	}
	loot := findLootByDef(sim, "quest_leaf")
	if loot == nil {
		t.Fatal("missing dropped quest_leaf loot")
	}
	player := sim.entities[sim.playerID]
	if distance(loot.pos, player.pos) < playerRadius+lootInteractionRadius {
		t.Fatalf("loot too close to player: player=%+v loot=%+v", player.pos, loot.pos)
	}
	if !hasEvent(r, "item_dropped") {
		t.Fatalf("missing item_dropped: %+v", r.Events)
	}
}

func TestDropEquippedWeapon(t *testing.T) {
	rules := loadRules(t)
	sim, itemID := inventoryLabEquippedSword(t, rules)

	r := sim.Tick([]Input{{
		MessageID:     "drop",
		CorrelationID: "corr_drop",
		Type:          "drop_intent",
		Drop:          &DropIntent{ItemInstanceID: itemID},
	}})
	assertAck(t, r, "drop")
	if sim.equipped[mainHandSlot] != 0 {
		t.Fatalf("equipped main_hand = %d, want cleared", sim.equipped[mainHandSlot])
	}
	if sim.findItem(itemID) != nil {
		t.Fatalf("dropped item %s still in inventory", itemID)
	}
	if findLootByDef(sim, "rusty_sword") == nil {
		t.Fatal("missing dropped rusty_sword loot")
	}
	if !hasChange(r, OpInventoryRemove) || !hasChange(r, OpEquippedUpdate) {
		t.Fatalf("drop missing inventory_remove/equipped_update changes: %+v", r.Changes)
	}
}

func TestDropThenPickup(t *testing.T) {
	rules := loadRules(t)
	sim, itemID := inventoryLabEquippedSword(t, rules)
	drop := sim.Tick([]Input{{
		MessageID: "drop",
		Type:      "drop_intent",
		Drop:      &DropIntent{ItemInstanceID: itemID},
	}})
	assertAck(t, drop, "drop")
	loot := findLootByDef(sim, "rusty_sword")
	if loot == nil {
		t.Fatal("missing dropped loot")
	}

	pickup := sim.Tick([]Input{{
		MessageID: "pickup",
		Type:      "action_intent",
		Action:    &ActionIntent{TargetID: idStr(loot.id)},
	}})
	assertAck(t, pickup, "pickup")
	if len(sim.inventory) != 1 || sim.inventory[0].itemDefID != "rusty_sword" {
		t.Fatalf("inventory after re-pickup = %+v, want rusty_sword", sim.inventory)
	}
}

func TestDropNoSpace(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_drop_no_space", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatalf("new inventory lab: %v", err)
	}
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "quest_leaf"})
	player := sim.entities[sim.playerID]
	level := sim.activeLevel()
	for ring := 1; ring <= 6; ring++ {
		for _, offset := range adjacentUnitOffsets() {
			level.walls = append(level.walls, wallObstacle{
				pos:  Vec2{X: player.pos.X + offset.X*float64(ring), Y: player.pos.Y + offset.Y*float64(ring)},
				size: Vec2{X: 1, Y: 1},
			})
		}
	}
	sim.syncCompatibilityFields()

	r := sim.Tick([]Input{{MessageID: "drop", Type: "drop_intent", Drop: &DropIntent{ItemInstanceID: "5000"}}})
	assertReject(t, r, "drop", "no_drop_space")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory mutated on rejected drop: %+v", sim.inventory)
	}
	if findLootByDef(sim, "quest_leaf") != nil {
		t.Fatal("rejected drop spawned loot")
	}
}

func TestUseConsumableHealLab(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceMonsterHitChance(rules, "training_dummy_heal", 1.0)
	sim, err := NewSimWithWorld("sess_heal_lab", "01", rules, "heal_lab")
	if err != nil {
		t.Fatalf("new heal lab: %v", err)
	}
	monster := findMonsterByDef(sim, "training_dummy_heal")
	if monster == nil {
		t.Fatal("missing heal_lab training_dummy_heal")
	}

	for i := 0; i < 2; i++ {
		attack := sim.Tick([]Input{{
			MessageID: "attack",
			Type:      "action_intent",
			Action:    &ActionIntent{TargetID: idStr(monster.id)},
		}})
		assertAck(t, attack, "attack")
		if monster.hp == 0 {
			break
		}
	}
	player := sim.entities[sim.playerID]
	if player.hp != 5 {
		t.Fatalf("player hp after combat = %d, want 5", player.hp)
	}
	redLoots := findAllLootByDef(sim, "red_potion")
	blueLoots := findAllLootByDef(sim, "blue_potion")
	if len(redLoots) != 1 || len(blueLoots) != 1 {
		t.Fatalf("loot drops red=%+v blue=%+v, want one red_potion and one blue_potion", redLoots, blueLoots)
	}

	for _, itemDefID := range []string{"red_potion", "blue_potion"} {
		loot := findLootByDef(sim, itemDefID)
		if loot == nil {
			t.Fatalf("missing %s loot pickup", itemDefID)
		}
		messageID := "pickup-" + itemDefID
		pickup := sim.Tick([]Input{{
			MessageID: messageID,
			Type:      "action_intent",
			Action:    &ActionIntent{TargetID: idStr(loot.id)},
		}})
		assertAck(t, pickup, messageID)
		if sim.autoNav != nil {
			for step := 0; step < 30 && findLootByDef(sim, itemDefID) != nil; step++ {
				sim.Tick(nil)
			}
		}
	}
	if len(sim.inventory) != 2 {
		t.Fatalf("inventory after pickups = %+v, want two items", sim.inventory)
	}

	redPotion := findItemByDef(sim, "red_potion")
	if redPotion == nil {
		t.Fatalf("missing red_potion in inventory %+v", sim.inventory)
	}
	firstID := idStr(redPotion.instanceID)
	use1 := sim.Tick([]Input{{
		MessageID: "use1",
		Type:      "use_intent",
		Use:       &UseIntent{ItemInstanceID: firstID},
	}})
	assertAck(t, use1, "use1")
	if player.hp != 10 {
		t.Fatalf("player hp after first use = %d, want 10", player.hp)
	}
	assertEventHeal(t, use1, "player_healed", 5)

	bluePotion := findItemByDef(sim, "blue_potion")
	if bluePotion == nil {
		t.Fatalf("missing blue_potion in inventory %+v", sim.inventory)
	}
	player.mana = 5
	secondID := idStr(bluePotion.instanceID)
	use2 := sim.Tick([]Input{{
		MessageID: "use2",
		Type:      "use_intent",
		Use:       &UseIntent{ItemInstanceID: secondID},
	}})
	assertAck(t, use2, "use2")
	if player.mana != 10 {
		t.Fatalf("player mana after blue potion = %d, want 10", player.mana)
	}
	assertEventMana(t, use2, "player_mana_restored", 5)
	if len(sim.inventory) != 0 {
		t.Fatalf("inventory after potion uses = %+v, want empty", sim.inventory)
	}
}

func TestUseConsumableRejectsFullHP(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_use_full", "01", rules, "heal_lab")
	if err != nil {
		t.Fatalf("new heal lab: %v", err)
	}
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "red_potion", equipped: false})

	r := sim.Tick([]Input{{MessageID: "use", Type: "use_intent", Use: &UseIntent{ItemInstanceID: "5000"}}})
	assertReject(t, r, "use", "already_full_hp")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory mutated on rejected use: %+v", sim.inventory)
	}
}

func TestUseConsumableRejectsNonConsumable(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_use_badge", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatalf("new inventory lab: %v", err)
	}
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "quest_leaf", equipped: false})
	sim.entities[sim.playerID].hp = 5

	r := sim.Tick([]Input{{MessageID: "use", Type: "use_intent", Use: &UseIntent{ItemInstanceID: "5000"}}})
	assertReject(t, r, "use", "not_consumable")
}

func TestHotbarCapacityAndBelt(t *testing.T) {
	sim := MustNewSim("sess_hotbar_capacity", "01", loadRules(t))
	snap := sim.Snapshot()
	if snap.HotbarCapacity != 2 || len(snap.Hotbar) != 10 {
		t.Fatalf("base hotbar capacity=%d len=%d, want 2/10", snap.HotbarCapacity, len(snap.Hotbar))
	}

	belt := addRolledInventoryItem(t, sim, 7000, "cave_belt", map[string]int{"hotbar_slots": 6})
	equipBelt := sim.Tick([]Input{{MessageID: "belt", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(belt.instanceID), Slot: "belt"}}})
	assertAck(t, equipBelt, "belt")
	if !hasEquippedUpdateCapacity(equipBelt, "belt", 6) {
		t.Fatalf("belt equip missing capacity delta changes=%+v", equipBelt.Changes)
	}
	if got := sim.Snapshot().HotbarCapacity; got != 6 {
		t.Fatalf("rolled belt capacity = %d, want 6", got)
	}

	unequipBelt := sim.Tick([]Input{{MessageID: "unbelt", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}})
	assertAck(t, unequipBelt, "unbelt")
	if !hasEquippedUpdateCapacity(unequipBelt, "belt", 2) {
		t.Fatalf("belt unequip missing capacity delta changes=%+v", unequipBelt.Changes)
	}
	fallbackBelt := addRolledInventoryItem(t, sim, 7001, "cave_belt", map[string]int{"armor": 1})
	assertAck(t, sim.Tick([]Input{{MessageID: "fallback", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(fallbackBelt.instanceID), Slot: "belt"}}}), "fallback")
	if got := sim.Snapshot().HotbarCapacity; got != 3 {
		t.Fatalf("belt base-stat fallback capacity = %d, want 3", got)
	}

	assertAck(t, sim.Tick([]Input{{MessageID: "unbelt2", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}}), "unbelt2")
	maxBelt := addRolledInventoryItem(t, sim, 7002, "cave_belt", map[string]int{"hotbar_slots": 99})
	assertAck(t, sim.Tick([]Input{{MessageID: "max", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(maxBelt.instanceID), Slot: "belt"}}}), "max")
	if got := sim.Snapshot().HotbarCapacity; got != 10 {
		t.Fatalf("clamped belt capacity = %d, want 10", got)
	}
}

func TestInventoryCapacityBaseItemBonusAndGolden(t *testing.T) {
	var golden struct {
		BaseInventoryRows int    `json:"base_inventory_rows"`
		Columns           int    `json:"columns"`
		BaseCapacity      int    `json:"base_capacity"`
		RowGrantingItem   string `json:"row_granting_item"`
		RowItemBonus      int    `json:"row_item_bonus"`
		RowItemCapacity   int    `json:"row_item_capacity"`
	}
	loadGolden(t, "inventory_capacity.json", &golden)
	if golden.BaseInventoryRows != baseInventoryRows || golden.Columns != inventoryColumns || golden.BaseCapacity != inventoryCapacityForRows(baseInventoryRows) {
		t.Fatalf("inventory capacity constants = rows %d columns %d cap %d, want golden %+v", baseInventoryRows, inventoryColumns, inventoryCapacityForRows(baseInventoryRows), golden)
	}

	sim := MustNewSim("sess_inventory_capacity", "01", loadRules(t))
	snap := sim.Snapshot()
	if snap.InventoryRows != golden.BaseInventoryRows || snap.InventoryCapacity != golden.BaseCapacity {
		t.Fatalf("base snapshot rows/capacity = %d/%d, want %d/%d", snap.InventoryRows, snap.InventoryCapacity, golden.BaseInventoryRows, golden.BaseCapacity)
	}
	if sim.bagOccupancyCount() != 0 {
		t.Fatalf("empty bag occupancy = %d, want 0", sim.bagOccupancyCount())
	}

	belt := addRolledInventoryItem(t, sim, 7300, golden.RowGrantingItem, map[string]int{"inventory_rows": golden.RowItemBonus})
	equip := sim.Tick([]Input{{MessageID: "pack_belt", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(belt.instanceID), Slot: "belt"}}})
	assertAck(t, equip, "pack_belt")
	if !hasEquippedUpdateInventoryCapacity(equip, "belt", golden.BaseInventoryRows+golden.RowItemBonus, golden.RowItemCapacity) {
		t.Fatalf("pack belt equip missing inventory capacity update: %+v", equip.Changes)
	}
	snap = sim.Snapshot()
	if snap.InventoryRows != golden.BaseInventoryRows+golden.RowItemBonus || snap.InventoryCapacity != golden.RowItemCapacity {
		t.Fatalf("pack belt snapshot rows/capacity = %d/%d, want %d/%d", snap.InventoryRows, snap.InventoryCapacity, golden.BaseInventoryRows+golden.RowItemBonus, golden.RowItemCapacity)
	}

	sim.progression.UnspentStatPoints = 1
	sim.savePlayer(sim.defaultPlayer())
	assertAck(t, sim.Tick([]Input{{MessageID: "stat", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "str", Points: 1}}}), "stat")
	snap = sim.Snapshot()
	if snap.InventoryRows != golden.BaseInventoryRows+golden.RowItemBonus || snap.InventoryCapacity != golden.RowItemCapacity {
		t.Fatalf("stat allocation changed inventory capacity to %d/%d", snap.InventoryRows, snap.InventoryCapacity)
	}
}

func TestInventoryCapacityOccupancyExemptsEquippedAndHotbar(t *testing.T) {
	sim := MustNewSim("sess_inventory_occupancy", "01", loadRules(t))
	sword := addStaticInventoryItem(sim, 7310, "rusty_sword")
	potion := addStaticInventoryItem(sim, 7311, "red_potion")
	badge := addStaticInventoryItem(sim, 7312, "quest_leaf")

	if got := sim.bagOccupancyCount(); got != 3 {
		t.Fatalf("bag occupancy = %d, want 3", got)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "equip_sword")
	if got := sim.bagOccupancyCount(); got != 2 {
		t.Fatalf("bag occupancy after equip = %d, want 2", got)
	}
	assign := sim.Tick([]Input{{MessageID: "assign_potion", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(potion.instanceID))}}})
	assertAck(t, assign, "assign_potion")
	if !hasHotbarUpdateInventoryCapacity(assign, 0, baseInventoryRows, inventoryCapacityForRows(baseInventoryRows)) {
		t.Fatalf("hotbar assignment missing inventory capacity update: %+v", assign.Changes)
	}
	if got := sim.bagOccupancyCount(); got != 1 {
		t.Fatalf("bag occupancy after hotbar assign = %d, want 1", got)
	}
	snap := sim.Snapshot()
	if findSnapshotItemByID(snap.Inventory, potion.instanceID) != nil {
		t.Fatalf("hotbar potion still appears in snapshot inventory: %+v", snap.Inventory)
	}
	if len(snap.Hotbar) == 0 || snap.Hotbar[0].Item == nil || snap.Hotbar[0].Item.ItemDefID != "red_potion" {
		t.Fatalf("hotbar snapshot missing assigned item view: %+v", snap.Hotbar)
	}
	if badge.equipped {
		t.Fatal("badge unexpectedly equipped")
	}
}

func TestInventoryCapacityPickupRejectsFullBagBeforeMutation(t *testing.T) {
	sim := MustNewSim("sess_inventory_full_pickup", "01", loadRules(t))
	for i := 0; i < inventoryCapacityForRows(baseInventoryRows); i++ {
		addStaticInventoryItem(sim, uint64(7400+i), "quest_leaf")
	}
	loot := &entity{id: sim.alloc(), kind: lootEntity, pos: sim.entities[sim.playerID].pos, itemDefID: "quest_leaf"}
	sim.entities[loot.id] = loot
	beforeNextID := sim.nextID

	res := sim.Tick([]Input{{MessageID: "pick_full", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertReject(t, res, "pick_full", "inventory_full")
	if sim.entities[loot.id] == nil {
		t.Fatalf("full-bag pickup removed loot")
	}
	if len(sim.inventory) != inventoryCapacityForRows(baseInventoryRows) || sim.nextID != beforeNextID {
		t.Fatalf("full-bag pickup mutated inventory/ids: len=%d next=%d want next=%d", len(sim.inventory), sim.nextID, beforeNextID)
	}
}

func TestInventoryCapacityUnequipAndShrinkRejectBeforeMutation(t *testing.T) {
	sim := MustNewSim("sess_inventory_unequip_full", "01", loadRules(t))
	sword := addStaticInventoryItem(sim, 7500, "rusty_sword")
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "equip_sword")
	for i := 0; i < inventoryCapacityForRows(baseInventoryRows); i++ {
		addStaticInventoryItem(sim, uint64(7510+i), "quest_leaf")
	}
	rejectUnequip := sim.Tick([]Input{{MessageID: "unequip_full", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: mainHandSlot}}})
	assertReject(t, rejectUnequip, "unequip_full", "capacity_would_overflow")
	if !sword.equipped || sim.equipped[mainHandSlot] != sword.instanceID {
		t.Fatalf("rejected unequip mutated sword/equipped: item=%+v equipped=%v", sword, sim.equipped)
	}

	sim = MustNewSim("sess_inventory_shrink_full", "01", loadRules(t))
	pack := addRolledInventoryItem(t, sim, 7600, "cave_pack_belt", map[string]int{"inventory_rows": 1})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip_pack", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(pack.instanceID), Slot: "belt"}}}), "equip_pack")
	for i := 0; i < inventoryCapacityForRows(baseInventoryRows)+1; i++ {
		addStaticInventoryItem(sim, uint64(7610+i), "quest_leaf")
	}
	rejectShrink := sim.Tick([]Input{{MessageID: "unequip_pack", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}})
	assertReject(t, rejectShrink, "unequip_pack", "capacity_would_overflow")
	if !pack.equipped || sim.equipped["belt"] != pack.instanceID || sim.Snapshot().InventoryCapacity != 20 {
		t.Fatalf("rejected shrink mutated pack/equipped/capacity: item=%+v equipped=%v snap=%+v", pack, sim.equipped, sim.Snapshot())
	}

	normal := addRolledInventoryItem(t, sim, 7700, "cave_belt", map[string]int{"armor": 1})
	rejectReplace := sim.Tick([]Input{{MessageID: "replace_pack", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(normal.instanceID), Slot: "belt"}}})
	assertReject(t, rejectReplace, "replace_pack", "capacity_would_overflow")
	if sim.equipped["belt"] != pack.instanceID || normal.equipped {
		t.Fatalf("rejected replacement mutated belts: equipped=%v normal=%+v", sim.equipped, normal)
	}
}

func TestHotbarAssignUseDirectUseAndReenable(t *testing.T) {
	sim := MustNewSim("sess_hotbar_use", "01", loadRules(t))
	player := sim.entities[sim.playerID]
	player.hp = 4
	first := addStaticInventoryItem(sim, 7100, "red_potion")
	assign := sim.Tick([]Input{{MessageID: "assign_disabled", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 5, ItemInstanceID: stringPtr(idStr(first.instanceID))}}})
	assertAck(t, assign, "assign_disabled")
	if sim.hotbar[5] != first.instanceID || !hasHotbarUpdate(assign, 5, stringPtr(idStr(first.instanceID))) {
		t.Fatalf("disabled assignment failed hotbar=%v changes=%+v", sim.hotbar, assign.Changes)
	}

	disabled := sim.Tick([]Input{{MessageID: "use_disabled", Type: "use_hotbar_intent", UseHotbar: &UseHotbarIntent{SlotIndex: 5}}})
	assertReject(t, disabled, "use_disabled", "hotbar_slot_disabled")
	direct := sim.Tick([]Input{{MessageID: "direct", Type: "use_intent", Use: &UseIntent{ItemInstanceID: idStr(first.instanceID)}}})
	assertAck(t, direct, "direct")
	if sim.hotbar[5] != 0 || !hasHotbarUpdate(direct, 5, nil) {
		t.Fatalf("direct use did not clear disabled hotbar slot: hotbar=%v changes=%+v", sim.hotbar, direct.Changes)
	}

	player.hp = 4
	second := addStaticInventoryItem(sim, 7101, "red_potion")
	assign = sim.Tick([]Input{{MessageID: "assign_second", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 5, ItemInstanceID: stringPtr(idStr(second.instanceID))}}})
	assertAck(t, assign, "assign_second")
	belt := addRolledInventoryItem(t, sim, 7102, "cave_belt", map[string]int{"hotbar_slots": 6})
	assertAck(t, sim.Tick([]Input{{MessageID: "belt", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(belt.instanceID), Slot: "belt"}}}), "belt")
	assertAck(t, sim.Tick([]Input{{MessageID: "unbelt", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: "belt"}}}), "unbelt")
	if got := sim.Snapshot().HotbarCapacity; got != 2 {
		t.Fatalf("capacity after belt unequip = %d, want 2", got)
	}
	assertAck(t, sim.Tick([]Input{{MessageID: "rebelt", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(belt.instanceID), Slot: "belt"}}}), "rebelt")
	use := sim.Tick([]Input{{MessageID: "use_hotbar", Type: "use_hotbar_intent", UseHotbar: &UseHotbarIntent{SlotIndex: 5}}})
	assertAck(t, use, "use_hotbar")
	if sim.findItemByID(second.instanceID) != nil || sim.hotbar[5] != 0 {
		t.Fatalf("hotbar use did not consume/clear item=%+v hotbar=%v", sim.findItemByID(second.instanceID), sim.hotbar)
	}
}

func TestHotbarAssignClearAndReplaceMovesItemsBetweenBagViewAndHotbar(t *testing.T) {
	sim := MustNewSim("sess_hotbar_bag_view", "01", loadRules(t))
	red := addStaticInventoryItem(sim, 7110, "red_potion")
	blue := addStaticInventoryItem(sim, 7111, "blue_potion")

	assignRed := sim.Tick([]Input{{MessageID: "assign_red", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(red.instanceID))}}})
	assertAck(t, assignRed, "assign_red")
	if !hasInventoryRemove(assignRed, red.instanceID) || !hasHotbarUpdateItem(assignRed, 0, red.instanceID, "red_potion") {
		t.Fatalf("assign red did not move potion to hotbar view: %+v", assignRed.Changes)
	}
	if findSnapshotItemByID(sim.Snapshot().Inventory, red.instanceID) != nil {
		t.Fatalf("assigned red still visible in bag snapshot: %+v", sim.Snapshot().Inventory)
	}

	clearRed := sim.Tick([]Input{{MessageID: "clear_red", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: nil}}})
	assertAck(t, clearRed, "clear_red")
	if !hasInventoryAdd(clearRed, red.instanceID) || !hasHotbarUpdate(clearRed, 0, nil) {
		t.Fatalf("clear red did not return potion to bag view: %+v", clearRed.Changes)
	}
	if findSnapshotItemByID(sim.Snapshot().Inventory, red.instanceID) == nil {
		t.Fatalf("cleared red missing from bag snapshot: %+v", sim.Snapshot().Inventory)
	}

	assertAck(t, sim.Tick([]Input{{MessageID: "assign_red_again", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(red.instanceID))}}}), "assign_red_again")
	replaceBlue := sim.Tick([]Input{{MessageID: "replace_blue", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(blue.instanceID))}}})
	assertAck(t, replaceBlue, "replace_blue")
	if !hasInventoryAdd(replaceBlue, red.instanceID) || !hasInventoryRemove(replaceBlue, blue.instanceID) || !hasHotbarUpdateItem(replaceBlue, 0, blue.instanceID, "blue_potion") {
		t.Fatalf("replace did not swap bag/hotbar views: %+v", replaceBlue.Changes)
	}
	snap := sim.Snapshot()
	if findSnapshotItemByID(snap.Inventory, red.instanceID) == nil || findSnapshotItemByID(snap.Inventory, blue.instanceID) != nil {
		t.Fatalf("replace snapshot inventory = %+v", snap.Inventory)
	}
}

func TestHotbarRejectsAndDropClears(t *testing.T) {
	sim := MustNewSim("sess_hotbar_rejects", "01", loadRules(t))
	potion := addStaticInventoryItem(sim, 7200, "red_potion")
	badge := addStaticInventoryItem(sim, 7201, "quest_leaf")

	assertReject(t, sim.Tick([]Input{{MessageID: "bad_index", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 10, ItemInstanceID: stringPtr(idStr(potion.instanceID))}}}), "bad_index", "invalid_payload")
	assertReject(t, sim.Tick([]Input{{MessageID: "missing", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr("9999")}}}), "missing", "not_in_inventory")
	assertReject(t, sim.Tick([]Input{{MessageID: "non_consumable", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(badge.instanceID))}}}), "non_consumable", "not_consumable")
	assertReject(t, sim.Tick([]Input{{MessageID: "empty", Type: "use_hotbar_intent", UseHotbar: &UseHotbarIntent{SlotIndex: 0}}}), "empty", "slot_empty")

	assertAck(t, sim.Tick([]Input{{MessageID: "assign0", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 0, ItemInstanceID: stringPtr(idStr(potion.instanceID))}}}), "assign0")
	assertAck(t, sim.Tick([]Input{{MessageID: "assign1", Type: "assign_hotbar_intent", AssignHotbar: &AssignHotbarIntent{SlotIndex: 1, ItemInstanceID: stringPtr(idStr(potion.instanceID))}}}), "assign1")
	drop := sim.Tick([]Input{{MessageID: "drop", Type: "drop_intent", Drop: &DropIntent{ItemInstanceID: idStr(potion.instanceID)}}})
	assertAck(t, drop, "drop")
	if sim.hotbar[0] != 0 || sim.hotbar[1] != 0 || !hasHotbarUpdate(drop, 0, nil) || !hasHotbarUpdate(drop, 1, nil) {
		t.Fatalf("drop did not clear hotbar refs: hotbar=%v changes=%+v", sim.hotbar, drop.Changes)
	}
}

func TestAdjacentLootDropSpreadsAndAvoidsWalls(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_multi_drop", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatalf("new inventory lab: %v", err)
	}
	source := Vec2{X: 10, Y: 10}
	blockedFirstCandidate := Vec2{X: source.X + 1, Y: source.Y}
	sim.walls = append(sim.walls, wallObstacle{pos: blockedFirstCandidate, size: Vec2{X: 1, Y: 1}})

	positions := map[Vec2]bool{}
	for i := 0; i < 12; i++ {
		pos, ok := sim.findEntityLootDropPosition(source, monsterRadius)
		if !ok {
			t.Fatalf("drop %d had no placement", i)
		}
		if positions[pos] {
			t.Fatalf("drop %d overlapped existing loot at %+v", i, pos)
		}
		if sim.lootDropBlocked(pos) {
			t.Fatalf("drop %d placed inside blocked geometry at %+v", i, pos)
		}
		if circlesOverlap(pos, lootInteractionRadius, source, monsterRadius) {
			t.Fatalf("drop %d overlaps source body: %+v", i, pos)
		}
		positions[pos] = true
		loot := &entity{kind: lootEntity, pos: pos, itemDefID: "quest_leaf"}
		loot.id = sim.alloc()
		sim.entities[loot.id] = loot
	}
	if positions[blockedFirstCandidate] {
		t.Fatalf("drop placed inside wall at %+v", blockedFirstCandidate)
	}
}

func TestInventoryDropGolden(t *testing.T) {
	var golden struct {
		WorldID              string `json:"world_id"`
		ItemDefID            string `json:"item_def_id"`
		ExpectedLootPosition Vec2   `json:"expected_loot_position"`
	}
	loadGolden(t, "inventory_drop.json", &golden)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_inventory_drop_golden", "01", rules, golden.WorldID)
	if err != nil {
		t.Fatalf("new golden world: %v", err)
	}
	addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: golden.ItemDefID})
	r := sim.Tick([]Input{{
		MessageID: "drop",
		Type:      "drop_intent",
		Drop:      &DropIntent{ItemInstanceID: "5000"},
	}})
	assertAck(t, r, "drop")
	loot := findLootByDef(sim, golden.ItemDefID)
	if loot == nil {
		t.Fatalf("missing dropped loot for %s", golden.ItemDefID)
	}
	if loot.pos != golden.ExpectedLootPosition {
		t.Fatalf("drop position = %+v, want %+v", loot.pos, golden.ExpectedLootPosition)
	}
}

// --- determinism ------------------------------------------------------------

func TestDeterministicReplayAndStableIDs(t *testing.T) {
	a := runSlice(t, "cafef00dcafef00d")
	b := runSlice(t, "cafef00dcafef00d")

	ja, _ := json.Marshal(a.Snapshot())
	jb, _ := json.Marshal(b.Snapshot())
	if string(ja) != string(jb) {
		t.Fatalf("snapshots diverged for identical seed+inputs:\n a=%s\n b=%s", ja, jb)
	}

	// Stable, reproducible entity ids matching the spec examples.
	snap := a.Snapshot()
	var player, monster *EntityView
	for i := range snap.Entities {
		switch snap.Entities[i].Type {
		case playerEntity:
			player = &snap.Entities[i]
		case monsterEntity:
			monster = &snap.Entities[i]
		}
	}
	if player.ID != "1001" || monster.ID != "1002" {
		t.Fatalf("entity ids = player %s monster %s, want 1001/1002", player.ID, monster.ID)
	}
	if snap.Inventory[0].ItemInstanceID != "1004" {
		t.Fatalf("item instance id = %s, want 1004", snap.Inventory[0].ItemInstanceID)
	}
}

func TestDifferentSeedsStillProduceItem(t *testing.T) {
	// The slice succeeds regardless of seed (single-entry loot, base_hit 1.0).
	for _, seed := range []string{"00", "0102030405060708", "ffffffffffffffff"} {
		sim := runSlice(t, seed)
		snap := sim.Snapshot()
		if len(snap.Inventory) != 1 || !snap.Inventory[0].Equipped {
			t.Fatalf("seed %s: inventory = %+v", seed, snap.Inventory)
		}
	}
}

func TestSnapshotIncludesWallLayout(t *testing.T) {
	sim, err := NewSimWithWorld("sess_wall_snapshot", "v40_obstacles", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend_wall_snapshot")
	snap := sim.Snapshot()
	if snap.CurrentLevel != -1 {
		t.Fatalf("snapshot level = %d, want -1", snap.CurrentLevel)
	}
	if len(snap.Walls) < 5 {
		t.Fatalf("snapshot walls = %d, want perimeter plus generated walls", len(snap.Walls))
	}
	if countWallSource(snap.Walls, "generated") == 0 {
		t.Fatalf("snapshot walls missing generated source: %+v", snap.Walls)
	}
	for i, wall := range snap.Walls {
		wantID := wallID(snap.CurrentLevel, i)
		if wall.ID != wantID {
			t.Fatalf("wall %d id = %s, want %s", i, wall.ID, wantID)
		}
		if wall.Size.X <= 0 || wall.Size.Y <= 0 {
			t.Fatalf("wall %d has invalid size: %+v", i, wall)
		}
	}
}

func TestLevelTransitionWallLayoutUpdatePrecedesSpawns(t *testing.T) {
	sim, err := NewSimWithWorld("sess_wall_delta", "v40_obstacles", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	results := descendFromCurrentLevel(t, sim, "descend_wall_delta")
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2: %+v", len(results), results)
	}
	dest := results[1]
	if len(dest.Changes) == 0 {
		t.Fatal("destination delta has no changes")
	}
	if dest.Changes[0].Op != OpWallLayoutUpdate {
		t.Fatalf("first destination change = %s, want %s in %+v", dest.Changes[0].Op, OpWallLayoutUpdate, dest.Changes)
	}
	if countWallSource(dest.Changes[0].Walls, "generated") == 0 {
		t.Fatalf("wall layout update missing generated walls: %+v", dest.Changes[0].Walls)
	}
	spawnIndex := -1
	for i, change := range dest.Changes {
		if change.Op == OpEntitySpawn {
			spawnIndex = i
			break
		}
	}
	if spawnIndex <= 0 {
		t.Fatalf("entity spawns did not follow wall layout update: %+v", dest.Changes)
	}
}

func countWallSource(walls []WallView, source string) int {
	count := 0
	for _, wall := range walls {
		if wall.Source == source {
			count++
		}
	}
	return count
}

// --- movement ---------------------------------------------------------------

func TestMovement(t *testing.T) {
	sim, err := NewSimWithWorld("sess_move", "abcd", loadRules(t), "gear_before_combat")
	if err != nil {
		t.Fatalf("gear world: %v", err)
	}
	start := sim.entities[sim.playerID].pos

	r := sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1, Y: 0}, DurationTicks: 3}}})
	if !hasPlayerUpdate(r) {
		t.Fatal("expected player entity_update on move tick")
	}
	sim.Tick(nil)
	sim.Tick(nil)
	// 3 ticks of configured input movement in +x.
	got := sim.entities[sim.playerID].pos
	wantX := start.X + 3*sim.rules.MainConfig.Gameplay.BaseMovementSpeed
	if got.X != wantX || got.Y != start.Y {
		t.Fatalf("player pos = %+v, want x=%v", got, wantX)
	}
	// Movement is exhausted; a 4th tick must not move.
	sim.Tick(nil)
	if sim.entities[sim.playerID].pos.X != got.X {
		t.Fatal("player moved after duration exhausted")
	}
}

func TestTickResultSlicesNeverNil(t *testing.T) {
	// A movement-only tick must still carry non-nil Changes/Events so the
	// state_delta marshals arrays, not null (regression guard).
	sim := MustNewSim("s", "01", loadRules(t))
	r := sim.Tick(nil)
	if r.Changes == nil || r.Events == nil {
		t.Fatalf("nil slices in tick result: %+v", r)
	}
	if b, _ := json.Marshal(r.Events); string(b) != "[]" {
		t.Fatalf("events marshaled as %s, want []", b)
	}
	if b, _ := json.Marshal(r.Changes); string(b) != "[]" {
		t.Fatalf("changes marshaled as %s, want []", b)
	}
}

func hasPlayerUpdate(r TickResult) bool {
	for _, c := range r.Changes {
		if c.Op == OpEntityUpdate && c.Entity != nil && c.Entity.Type == playerEntity {
			return true
		}
	}
	return false
}

func TestCollisionBlocksLiveMonster(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_monster", "01", loadRules(t), "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}

	sim.entities[sim.playerID].pos = Vec2{X: 7, Y: 5}
	sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	for i := 0; i < 2; i++ {
		sim.Tick(nil)
	}

	player := sim.entities[sim.playerID]
	monster := sim.findEntity("1002")
	if player.pos.X >= 8 {
		t.Fatalf("player pos = %+v, want stopped before live monster at x=8", player.pos)
	}
	if circlesOverlap(player.pos, playerRadius, monster.pos, monsterRadius) {
		t.Fatalf("player overlaps live monster: player=%+v monster=%+v", player.pos, monster.pos)
	}
}

func TestCollisionIgnoresDeadMonster(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_dead_monster", "01", loadRules(t), "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}
	sim.findEntity("1002").hp = 0

	sim.entities[sim.playerID].pos = Vec2{X: 7, Y: 5}
	sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})
	sim.Tick(nil)

	if got := sim.entities[sim.playerID].pos; got != (Vec2{X: 8, Y: 5}) {
		t.Fatalf("player pos = %+v, want able to enter dead monster position", got)
	}
}

func TestCollisionBlocksWallAndAllowsRoute(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_wall", "01", loadRules(t), "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}

	sim.entities[sim.playerID].pos = Vec2{X: 3, Y: 5}
	blocked := sim.Tick([]Input{{MessageID: "push_wall", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	for i := 0; i < 2; i++ {
		sim.Tick(nil)
	}
	if got := sim.entities[sim.playerID].pos; got.X >= 4 {
		t.Fatalf("player passed solid divider at y=5: pos=%+v", got)
	}
	if hasPlayerUpdate(blocked) {
		t.Fatalf("blocked wall push emitted player update: %+v", blocked.Changes)
	}

	sim.entities[sim.playerID].pos = Vec2{X: 2, Y: 2}
	moveTicks(sim, "through_bottom_gap", Vec2{X: 1}, 5)
	if got := sim.entities[sim.playerID].pos; got.X < 5 || got.Y > 3 {
		t.Fatalf("player did not route through bottom gap: pos=%+v", got)
	}
}

func TestActionAutoApproachQueuesWhenOutOfRange(t *testing.T) {
	rules := loadRules(t)

	t.Run("monster", func(t *testing.T) {
		sim := MustNewSim("sess_range_monster", "01", rules)
		r := sim.Tick([]Input{{MessageID: "a", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "a")
	})

	t.Run("loot", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_loot", "01", rules, "gear_before_combat")
		if err != nil {
			t.Fatalf("gear world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "p", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "p")
	})

	t.Run("door", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_door", "01", rules, "door_lab")
		if err != nil {
			t.Fatalf("door world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "d", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertAck(t, r, "d")
	})
}

func TestDoorLabClosedDoorPreventsPassageUntilActivated(t *testing.T) {
	sim, err := NewSimWithWorld("sess_door_passage", "01", loadRules(t), "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}

	sim.Tick([]Input{{MessageID: "push_closed", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 7}}})
	for i := 0; i < 6; i++ {
		sim.Tick(nil)
	}
	if got := sim.entities[sim.playerID].pos; got.X >= 4 {
		t.Fatalf("player passed closed door: pos=%+v", got)
	}
	open := sim.Tick([]Input{{MessageID: "open", CorrelationID: "corr_door", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
	assertAck(t, open, "open")
	if !hasEvent(open, "interactable_activated") {
		t.Fatalf("missing interactable_activated: %+v", open.Events)
	}
	door := sim.findEntity("1002")
	if door == nil || door.state != interactableOpen {
		t.Fatalf("door state = %+v, want open", door)
	}

	sim.Tick([]Input{{MessageID: "through", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 6}}})
	for i := 0; i < 5; i++ {
		sim.Tick(nil)
	}
	if got := sim.entities[sim.playerID].pos; got.X <= 4 {
		t.Fatalf("player did not pass open door: pos=%+v", got)
	}
	pickup := sim.Tick([]Input{{MessageID: "loot", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}}})
	assertAck(t, pickup, "loot")
	if !hasEvent(pickup, "item_picked_up") {
		t.Fatalf("missing item_picked_up after door passage: %+v", pickup.Events)
	}
}

func TestTreasureChestOpensOnceAndDropsLoot(t *testing.T) {
	sim, err := NewSimWithWorld("sess_chest_open", "chest_seed_22", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	var chest *entity
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == treasureChestDefID {
			chest = e
			break
		}
	}
	if chest == nil {
		t.Fatalf("missing generated chest: %+v", sim.activeLevel().entities)
	}
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	beforeLoot := countEntitiesByKind(sim.activeLevel(), lootEntity)
	beforeGold := countLootByItemDef(sim.activeLevel(), goldItemDefID)
	open := sim.Tick([]Input{{MessageID: "open_chest", CorrelationID: "corr_chest", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, open, "open_chest")
	if !hasEvent(open, "interactable_activated") || !hasEvent(open, "loot_dropped") {
		t.Fatalf("open chest events = %+v", open.Events)
	}
	if chest.state != interactableOpen {
		t.Fatalf("chest state = %s, want open", chest.state)
	}
	afterLoot := countEntitiesByKind(sim.activeLevel(), lootEntity)
	if afterLoot <= beforeLoot {
		t.Fatalf("loot count after open = %d, before %d", afterLoot, beforeLoot)
	}
	if got := countLootByItemDef(sim.activeLevel(), goldItemDefID); got != beforeGold+1 {
		t.Fatalf("gold drops after chest open = %d, want %d", got, beforeGold+1)
	}
	again := sim.Tick([]Input{{MessageID: "open_chest_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, again, "open_chest_again", "invalid_target")
	if got := countEntitiesByKind(sim.activeLevel(), lootEntity); got != afterLoot {
		t.Fatalf("reopen changed loot count = %d, want %d", got, afterLoot)
	}
}

func TestChestSeed22AllMonstersApproachable(t *testing.T) {
	sim, err := NewSimWithWorld("sess_chest_approach", "chest_seed_22", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	player := sim.entities[sim.playerID]
	reachableMonsters := []*entity{}
	unreachableMonsters := []*entity{}
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		e := sim.activeLevel().entities[id]
		if e.kind != monsterEntity || e.hp <= 0 {
			continue
		}
		_, steps, ok := sim.findApproachGoal(e)
		if !ok {
			unreachableMonsters = append(unreachableMonsters, e)
			t.Logf("unreachable monster %d pos=%+v player=%+v", id, e.pos, player.pos)
			continue
		}
		reachableMonsters = append(reachableMonsters, e)
		t.Logf("monster %d pos=%+v steps=%d", id, e.pos, len(steps))
	}
	for _, blocked := range unreachableMonsters {
		def := sim.rules.Monsters[blocked.monsterDefID]
		coveredByAssist := false
		for _, reachable := range reachableMonsters {
			if distance(blocked.pos, reachable.pos) <= def.effectiveAssistRadius() {
				coveredByAssist = true
				break
			}
		}
		if !coveredByAssist {
			t.Fatalf("unapproachable monster %d at %+v is not assist-covered by an approachable packmate", blocked.id, blocked.pos)
		}
	}
}

// --- rejections (criterion 12) ----------------------------------------------

func TestRejections(t *testing.T) {
	rules := loadRules(t)

	t.Run("invalid attack target", func(t *testing.T) {
		sim := MustNewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "action_intent", Action: &ActionIntent{TargetID: "9999"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("pickup non-loot", func(t *testing.T) {
		sim := MustNewSim("s", "01", rules)
		sim.findEntity("1002").hp = 0
		r := sim.Tick([]Input{{MessageID: "x", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("equip not in inventory", func(t *testing.T) {
		sim := MustNewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
		assertReject(t, r, "x", "not_in_inventory")
	})

	t.Run("equip non-equippable", func(t *testing.T) {
		sim := MustNewSim("s", "01", rules)
		addTestInventoryItem(sim, &invItem{instanceID: 5000, itemDefID: "quest_leaf"})
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
		assertReject(t, r, "x", "not_equippable")
	})

	t.Run("unknown type", func(t *testing.T) {
		sim := MustNewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "bogus_intent"}})
		assertReject(t, r, "x", "unknown_type")
	})

	t.Run("duplicate pickup", func(t *testing.T) {
		sim := runSlice(t, "0011223344556677")
		// The loot was already picked up during runSlice; picking up 1003 again rejects.
		r := sim.Tick([]Input{{MessageID: "dup", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}}})
		assertReject(t, r, "dup", "invalid_target")
	})
}

func TestDeadPlayerRejectsIntentsAndStopsActiveMovement(t *testing.T) {
	rules := loadRules(t)

	cases := []Input{
		{MessageID: "move", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}},
		{MessageID: "directional", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}},
		{MessageID: "attack", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}},
		{MessageID: "pickup", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}},
		{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}},
		{MessageID: "unequip", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: mainHandSlot}},
		{MessageID: "drop", Type: "drop_intent", Drop: &DropIntent{ItemInstanceID: "1004"}},
	}
	for _, in := range cases {
		sim := MustNewSim("sess_dead_"+in.MessageID, "01", rules)
		sim.entities[sim.playerID].hp = 0
		r := sim.Tick([]Input{in})
		assertReject(t, r, in.MessageID, "player_dead")
	}

	sim := MustNewSim("sess_dead_move", "01", rules)
	start := sim.entities[sim.playerID].pos
	sim.Tick([]Input{{MessageID: "move", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
	afterFirst := sim.entities[sim.playerID].pos
	if afterFirst.X == start.X {
		t.Fatal("setup failed: player did not move on first active movement tick")
	}
	sim.entities[sim.playerID].hp = 0
	r := sim.Tick(nil)
	if hasPlayerUpdate(r) {
		t.Fatalf("dead active movement emitted player update: %+v", r.Changes)
	}
	if sim.entities[sim.playerID].pos != afterFirst {
		t.Fatalf("dead player moved from %+v to %+v", afterFirst, sim.entities[sim.playerID].pos)
	}
	if sim.move != nil {
		t.Fatalf("active movement not cleared for dead player: %+v", sim.move)
	}
}

func assertReject(t *testing.T, r TickResult, msgID, reason string) {
	t.Helper()
	for _, rej := range r.Rejects {
		if rej.MessageID == msgID {
			if rej.Reason != reason {
				t.Fatalf("reject reason = %q, want %q", rej.Reason, reason)
			}
			return
		}
	}
	t.Fatalf("expected reject of %q with reason %q; rejects=%+v acks=%+v", msgID, reason, r.Rejects, r.Acks)
}

func assertAck(t *testing.T, r TickResult, msgID string) {
	t.Helper()
	for _, ack := range r.Acks {
		if ack.MessageID == msgID {
			return
		}
	}
	t.Fatalf("expected ack of %q; rejects=%+v acks=%+v", msgID, r.Rejects, r.Acks)
}

func assertAckInResults(t *testing.T, results []TickResult, msgID string) {
	t.Helper()
	for _, result := range results {
		for _, ack := range result.Acks {
			if ack.MessageID == msgID {
				return
			}
		}
	}
	t.Fatalf("expected ack of %q in results: %+v", msgID, results)
}

func hasEventInResults(results []TickResult, eventType string) bool {
	return countEventsInResults(results, eventType) > 0
}

func countEventsInResults(results []TickResult, eventType string) int {
	count := 0
	for _, result := range results {
		for _, ev := range result.Events {
			if ev.EventType == eventType {
				count++
			}
		}
	}
	return count
}

func eventForEntityInResults(results []TickResult, eventType string, entityID uint64) bool {
	for _, result := range results {
		if eventForEntity(result, eventType, entityID) {
			return true
		}
	}
	return false
}

func changeOwnersForOpInResults(results []TickResult, op string) []uint64 {
	var out []uint64
	for _, result := range results {
		out = append(out, changeOwnersForOp(result, op)...)
	}
	return out
}

func goldPickupAmounts(r TickResult) []int {
	var out []int
	for _, ev := range r.Events {
		if ev.EventType == "gold_picked_up" && ev.Amount != nil {
			out = append(out, *ev.Amount)
		}
	}
	return out
}

type equipmentRequirementGolden struct {
	Level           int                     `json:"level"`
	BaseStats       BaseStatsView           `json:"base_stats"`
	RequirementsMet bool                    `json:"requirements_met"`
	Status          []RequirementStatusView `json:"status"`
}

func assertRequirementStatus(t *testing.T, got, want []RequirementStatusView) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("requirement status = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("requirement status[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func findPreviewDelta(deltas []EquipPreviewDeltaView, stat string) *EquipPreviewDeltaView {
	for i := range deltas {
		if deltas[i].Stat == stat {
			return &deltas[i]
		}
	}
	return nil
}

func inventoryUpdateForItem(r TickResult, instanceID uint64) *ItemView {
	want := idStr(instanceID)
	for i := range r.Changes {
		change := r.Changes[i]
		if change.Op == OpInventoryUpdate && change.Item != nil && change.Item.ItemInstanceID == want {
			return change.Item
		}
	}
	return nil
}

func inventoryLabEquippedSword(t *testing.T, rules *Rules) (*Sim, string) {
	t.Helper()
	sim, err := NewSimWithWorld("sess_inventory_lab", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatalf("new inventory lab: %v", err)
	}
	pickup := sim.Tick([]Input{{
		MessageID:     "pickup",
		CorrelationID: "corr_pickup",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})
	assertAck(t, pickup, "pickup")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory size = %d, want 1", len(sim.inventory))
	}
	itemID := idStr(sim.inventory[0].instanceID)
	equip := sim.Tick([]Input{{
		MessageID:     "equip",
		CorrelationID: "corr_equip",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: itemID, Slot: mainHandSlot},
	}})
	assertAck(t, equip, "equip")
	return sim, itemID
}

func findLootByDef(sim *Sim, itemDefID string) *entity {
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e.kind == lootEntity && e.itemDefID == itemDefID {
			return e
		}
	}
	return nil
}

func addTestFloorLoot(sim *Sim, itemDefID string, pos Vec2) *entity {
	loot := &entity{kind: lootEntity, pos: pos, itemDefID: itemDefID}
	loot.id = sim.alloc()
	sim.activeLevel().entities[loot.id] = loot
	return loot
}

func addTestGoldLoot(sim *Sim, pos Vec2, amount int) *entity {
	loot := addTestFloorLoot(sim, goldItemDefID, pos)
	loot.goldAmount = amount
	return loot
}

func findAllLootByDef(sim *Sim, itemDefID string) []*entity {
	var out []*entity
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e.kind == lootEntity && e.itemDefID == itemDefID {
			out = append(out, e)
		}
	}
	return out
}

func findMonsterByDef(sim *Sim, monsterDefID string) *entity {
	for _, id := range sortedEntityIDs(sim.entities) {
		e := sim.entities[id]
		if e.kind == monsterEntity && e.monsterDefID == monsterDefID {
			return e
		}
	}
	return nil
}

func findItemByDef(sim *Sim, itemDefID string) *invItem {
	for _, item := range sim.inventory {
		if item.itemDefID == itemDefID {
			return item
		}
	}
	return nil
}

func addRolledInventoryItem(t *testing.T, sim *Sim, instanceID uint64, templateID string, stats map[string]int) *invItem {
	t.Helper()
	template, ok := sim.rules.ItemTemplates[templateID]
	if !ok {
		t.Fatalf("missing item template %s", templateID)
	}
	payload := &ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    template.Name,
		Rarity:         "test",
		Stats:          cloneIntMap(template.BaseStats),
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      []string{},
	}
	if payload.Stats == nil && len(stats) > 0 {
		payload.Stats = map[string]int{}
	}
	for key, value := range stats {
		payload.Stats[key] = value
	}
	item := &invItem{
		instanceID:  instanceID,
		itemDefID:   templateID,
		slot:        template.Slot,
		rollPayload: payload,
	}
	addTestInventoryItem(sim, item)
	return item
}

func addStaticInventoryItem(sim *Sim, instanceID uint64, itemDefID string) *invItem {
	item := &invItem{instanceID: instanceID, itemDefID: itemDefID}
	addTestInventoryItem(sim, item)
	return item
}

func addTestInventoryItem(sim *Sim, item *invItem) {
	sim.inventory = append(sim.inventory, item)
	sim.savePlayer(sim.defaultPlayer())
}

func stringPtr(v string) *string {
	return &v
}

func assertEquippedTemplate(t *testing.T, sim *Sim, slot, templateID string) {
	t.Helper()
	item := sim.findItemByID(sim.equipped[slot])
	if item == nil || item.rollPayload == nil || item.rollPayload.ItemTemplateID != templateID {
		t.Fatalf("equipped[%s] = %+v, want template %s", slot, item, templateID)
	}
}

func hasChange(r TickResult, op string) bool {
	for _, c := range r.Changes {
		if c.Op == op {
			return true
		}
	}
	return false
}

func hasPlayerResourceUpdate(r TickResult, hp, mana int) bool {
	for _, c := range r.Changes {
		if c.Op != OpEntityUpdate || c.Entity == nil || c.Entity.Type != playerEntity {
			continue
		}
		if c.Entity.HP == nil || *c.Entity.HP != hp || c.Entity.Mana == nil || *c.Entity.Mana != mana {
			continue
		}
		return true
	}
	return false
}

func hasHotbarUpdate(r TickResult, slotIndex int, itemInstanceID *string) bool {
	for _, c := range r.Changes {
		if c.Op != OpHotbarUpdate || c.SlotIndex != slotIndex {
			continue
		}
		if itemInstanceID == nil {
			return c.ItemInstanceID == nil
		}
		return c.ItemInstanceID != nil && *c.ItemInstanceID == *itemInstanceID
	}
	return false
}

func hasHotbarUpdateItem(r TickResult, slotIndex int, instanceID uint64, itemDefID string) bool {
	for _, c := range r.Changes {
		if c.Op != OpHotbarUpdate || c.SlotIndex != slotIndex || c.ItemInstanceID == nil {
			continue
		}
		if *c.ItemInstanceID != idStr(instanceID) || c.Item == nil || c.Item.ItemDefID != itemDefID {
			continue
		}
		return true
	}
	return false
}

func hasInventoryAdd(r TickResult, instanceID uint64) bool {
	for _, c := range r.Changes {
		if c.Op == OpInventoryAdd && c.Item != nil && c.Item.ItemInstanceID == idStr(instanceID) {
			return true
		}
	}
	return false
}

func hasInventoryRemove(r TickResult, instanceID uint64) bool {
	for _, c := range r.Changes {
		if c.Op == OpInventoryRemove && c.ItemInstanceID != nil && *c.ItemInstanceID == idStr(instanceID) {
			return true
		}
	}
	return false
}

func findSnapshotItemByID(items []ItemView, instanceID uint64) *ItemView {
	for i := range items {
		if items[i].ItemInstanceID == idStr(instanceID) {
			return &items[i]
		}
	}
	return nil
}

func hasEquippedUpdate(r TickResult, slot string, itemInstanceID *string) bool {
	for _, c := range r.Changes {
		if c.Op != OpEquippedUpdate || c.Slot != slot {
			continue
		}
		if itemInstanceID == nil && c.ItemInstanceID == nil {
			return true
		}
		if itemInstanceID != nil && c.ItemInstanceID != nil && *c.ItemInstanceID == *itemInstanceID {
			return true
		}
	}
	return false
}

func hasEquippedUpdateCapacity(r TickResult, slot string, capacity int) bool {
	for _, c := range r.Changes {
		if c.Op == OpEquippedUpdate && c.Slot == slot && c.HotbarCapacity != nil && *c.HotbarCapacity == capacity {
			return true
		}
	}
	return false
}

func hasEquippedUpdateInventoryCapacity(r TickResult, slot string, rows, capacity int) bool {
	for _, c := range r.Changes {
		if c.Op == OpEquippedUpdate && c.Slot == slot &&
			c.InventoryRows != nil && *c.InventoryRows == rows &&
			c.InventoryCap != nil && *c.InventoryCap == capacity {
			return true
		}
	}
	return false
}

func hasHotbarUpdateInventoryCapacity(r TickResult, slotIndex int, rows, capacity int) bool {
	for _, c := range r.Changes {
		if c.Op == OpHotbarUpdate && c.SlotIndex == slotIndex &&
			c.InventoryRows != nil && *c.InventoryRows == rows &&
			c.InventoryCap != nil && *c.InventoryCap == capacity {
			return true
		}
	}
	return false
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func gearBeforeCombatWithEquippedSword(t *testing.T, rules *Rules) *Sim {
	t.Helper()
	sim, err := NewSimWithWorld("sess_gear_weapon", "deadbeefdeadbeef", rules, "gear_before_combat")
	if err != nil {
		t.Fatalf("new gear sim: %v", err)
	}
	moveTicks(sim, "to_sword", Vec2{X: 1}, 5)

	pickup := sim.Tick([]Input{{
		MessageID:     "p1",
		CorrelationID: "corr_pickup",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: "1002"},
	}})
	assertAck(t, pickup, "p1")

	snap := sim.Snapshot()
	if len(snap.Inventory) != 1 {
		t.Fatalf("inventory size = %d, want 1", len(snap.Inventory))
	}
	itemID := snap.Inventory[0].ItemInstanceID
	equip := sim.Tick([]Input{{
		MessageID:     "e1",
		CorrelationID: "corr_equip",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: itemID, Slot: mainHandSlot},
	}})
	assertAck(t, equip, "e1")
	moveTicks(sim, "to_dummy", Vec2{X: 1}, 6)
	return sim
}

func moveTicks(sim *Sim, messageID string, dir Vec2, ticks int) {
	sim.Tick([]Input{{MessageID: messageID, Type: "move_intent", Move: &MoveIntent{Direction: dir, DurationTicks: ticks}}})
	for i := 1; i < ticks; i++ {
		sim.Tick(nil)
	}
}

func cloneRules(r *Rules) *Rules {
	out := *r
	out.Items = make(map[string]ItemDef, len(r.Items))
	for id, def := range r.Items {
		out.Items[id] = def
	}
	out.ItemTemplates = make(map[string]ItemTemplateDef, len(r.ItemTemplates))
	for id, def := range r.ItemTemplates {
		out.ItemTemplates[id] = def
	}
	out.Skills = make(map[string]SkillDef, len(r.Skills))
	for id, def := range r.Skills {
		out.Skills[id] = def
	}
	out.CharacterProgression.DerivedStats = make(map[string]LinearStatFormula, len(r.CharacterProgression.DerivedStats))
	for id, def := range r.CharacterProgression.DerivedStats {
		out.CharacterProgression.DerivedStats[id] = def
	}
	out.Monsters = make(map[string]MonsterDef, len(r.Monsters))
	for id, def := range r.Monsters {
		out.Monsters[id] = def
	}
	out.LootTables = make(map[string]LootTable, len(r.LootTables))
	for id, def := range r.LootTables {
		out.LootTables[id] = def
	}
	out.TreasureClasses = make(map[string]TreasureClassDef, len(r.TreasureClasses))
	for id, def := range r.TreasureClasses {
		out.TreasureClasses[id] = def
	}
	out.Interactables = make(map[string]InteractableDef, len(r.Interactables))
	for id, def := range r.Interactables {
		out.Interactables[id] = def
	}
	out.Worlds = make(map[string]WorldDef, len(r.Worlds))
	for id, def := range r.Worlds {
		out.Worlds[id] = def
	}
	return &out
}

func TestMonsterChaseGolden(t *testing.T) {
	var golden struct {
		Seed  string `json:"seed"`
		Cases []struct {
			Name            string `json:"name"`
			WorldID         string `json:"world_id"`
			IdlePlayerTicks int    `json:"idle_player_ticks"`
			PlayerKiteSteps []struct {
				Direction     Vec2 `json:"direction"`
				DurationTicks int  `json:"duration_ticks"`
				Ticks         int  `json:"ticks"`
			} `json:"player_kite_steps"`
			WaitTicksAfterKite      int      `json:"wait_ticks_after_kite"`
			ExpectedMonsterPosition *Vec2    `json:"expected_monster_position"`
			ExpectedNearSpawn       bool     `json:"expected_monster_final_near_spawn"`
			ExpectedEvents          []string `json:"expected_events"`
		} `json:"cases"`
	}
	loadGolden(t, "monster_chase.json", &golden)
	rules := loadRules(t)
	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			worldID := tc.WorldID
			if worldID == "" {
				worldID = "chase_maze"
			}
			sim, err := NewSimWithWorld("sess_monster_chase", golden.Seed, rules, worldID)
			if err != nil {
				t.Fatalf("world: %v", err)
			}
			monster := firstEntityByKind(sim, monsterEntity)
			spawn := monster.spawnPos
			seen := map[string]bool{}
			record := func(res TickResult) {
				for _, ev := range res.Events {
					seen[ev.EventType] = true
				}
			}
			for i := 0; i < tc.IdlePlayerTicks; i++ {
				record(sim.Tick(nil))
			}
			for _, step := range tc.PlayerKiteSteps {
				dir := step.Direction
				duration := step.DurationTicks
				if duration == 0 {
					duration = 1
				}
				repeats := step.Ticks
				if repeats == 0 {
					repeats = 1
				}
				for i := 0; i < repeats; i++ {
					record(sim.Tick([]Input{{
						MessageID: fmt.Sprintf("kite-%d", i),
						Type:      "move_intent",
						Move:      &MoveIntent{Direction: dir, DurationTicks: duration},
					}}))
				}
			}
			for i := 0; i < tc.WaitTicksAfterKite; i++ {
				record(sim.Tick(nil))
			}
			for _, want := range tc.ExpectedEvents {
				if !seen[want] {
					t.Fatalf("missing event %s; saw %v", want, seen)
				}
			}
			monster = firstEntityByKind(sim, monsterEntity)
			if tc.ExpectedMonsterPosition != nil {
				want := *tc.ExpectedMonsterPosition
				if distance(monster.pos, want) > 0.001 {
					t.Fatalf("monster position = %+v, want %+v", monster.pos, want)
				}
			}
			if tc.ExpectedNearSpawn {
				nav := rules.Navigation
				if distance(monster.pos, spawn) > nav.StopDistance+0.001 {
					t.Fatalf("monster %+v not near spawn %+v (dist=%.3f)", monster.pos, spawn, distance(monster.pos, spawn))
				}
			}
		})
	}
}

func TestMonsterChaseStaticDefault(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "01", rules, DefaultWorldID)
	if err != nil {
		t.Fatal(err)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	before := monster.pos
	for i := 0; i < 10; i++ {
		sim.Tick(nil)
	}
	monster = firstEntityByKind(sim, monsterEntity)
	if monster.pos != before {
		t.Fatalf("static monster moved from %+v to %+v", before, monster.pos)
	}
}

func TestMonsterChaseOpenField(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "01", rules, "chase_lab")
	if err != nil {
		t.Fatal(err)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	before := monster.pos
	var aggro bool
	for i := 0; i < 40; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "monster_aggro") {
			aggro = true
		}
	}
	monster = firstEntityByKind(sim, monsterEntity)
	player := sim.entities[sim.playerID]
	if !aggro {
		t.Fatal("expected monster_aggro")
	}
	if distance(monster.pos, before) < 0.5 {
		t.Fatalf("monster did not move enough: before=%+v after=%+v", before, monster.pos)
	}
	if distance(monster.pos, player.pos) > 1.5 {
		t.Fatalf("monster not within player distance: dist=%.3f max=1.5", distance(monster.pos, player.pos))
	}
}

func TestMonsterChaseLeashLastsAtLeastTwentyFiveTiles(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "01", rules, "chase_lab")
	if err != nil {
		t.Fatal(err)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	player := sim.entities[sim.playerID]
	def := rules.Monsters[monster.monsterDefID]
	if def.LeashRadius >= minimumChaseLeashTiles*rules.Navigation.CellSize {
		t.Fatalf("test requires a monster leash below the minimum; got %.3f", def.LeashRadius)
	}

	aggro := TickResult{}
	player.pos = Vec2{X: monster.pos.X, Y: monster.pos.Y + def.AggroRadius - 1}
	sim.updateMonsterAIMode(monster, player, def, monster.aiMode, &aggro)
	if !hasEvent(aggro, "monster_aggro") || monster.aiMode != monsterAIModeChase || monster.aiTargetPlayerID != sim.playerID {
		t.Fatalf("monster did not acquire aggro target: mode=%s target=%d events=%+v", monster.aiMode, monster.aiTargetPlayerID, aggro.Events)
	}

	prevMode := monster.aiMode
	player.pos = Vec2{X: monster.spawnPos.X + minimumChaseLeashTiles*rules.Navigation.CellSize - 1, Y: monster.spawnPos.Y}
	inside := TickResult{}
	sim.updateMonsterAIMode(monster, player, def, prevMode, &inside)
	if hasEvent(inside, "monster_leashed") || monster.aiMode != monsterAIModeChase {
		t.Fatalf("monster leashed before 25 tiles: mode=%s events=%+v", monster.aiMode, inside.Events)
	}

	prevMode = monster.aiMode
	player.pos = Vec2{X: monster.spawnPos.X + minimumChaseLeashTiles*rules.Navigation.CellSize + 1, Y: monster.spawnPos.Y}
	outside := TickResult{}
	sim.updateMonsterAIMode(monster, player, def, prevMode, &outside)
	if !hasEvent(outside, "monster_leashed") || monster.aiMode != monsterAIModeReturn {
		t.Fatalf("monster did not leash after passing 25 tiles: mode=%s events=%+v", monster.aiMode, outside.Events)
	}
}

func TestMonsterChaseStopsWhenMeleeAdjacent(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "01", rules, "chase_maze")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 30; i++ {
		sim.Tick(nil)
	}
	monster := firstEntityByKind(sim, monsterEntity)
	player := sim.entities[sim.playerID]
	at := monster.pos
	for i := 0; i < 10; i++ {
		res := sim.Tick(nil)
		for _, ch := range res.Changes {
			if ch.Op == OpEntityUpdate && ch.Entity != nil && ch.Entity.Type == "monster" {
				t.Fatalf("tick %d: monster still moving at %+v after reaching player at dist=%.3f",
					sim.CurrentTick(), monster.pos, distance(at, player.pos))
			}
		}
	}
	monster = firstEntityByKind(sim, monsterEntity)
	if monster.pos != at {
		t.Fatalf("monster drifted from %+v to %+v while adjacent to player", at, monster.pos)
	}
}

func TestMonsterPackFindsAttackSlotsAroundOffsetPlayer(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 4.49, Y: 5.49}
	player.hp = playerStartHP
	for _, pos := range []Vec2{
		{X: 8.3, Y: 5.49},
		{X: 8.3, Y: 4.49},
		{X: 8.3, Y: 6.49},
	} {
		addTestMonster(sim, "dungeon_mob", pos, rules.Monsters["dungeon_mob"].MaxHP)
	}

	attackers := map[string]bool{}
	for i := 0; i < 160 && len(attackers) < 2; i++ {
		res := sim.Tick(nil)
		for _, ev := range res.Events {
			if ev.EventType == "player_damaged" || ev.EventType == "player_killed" {
				attackers[ev.SourceEntityID] = true
			}
		}
	}
	if len(attackers) < 2 {
		t.Fatalf("monster pack did not surround and attack; attackers=%v player=%+v monsters=%+v",
			attackers, player.pos, dungeonMonsterDebugPositions(sim.activeLevel()))
	}
}

func TestDungeonStairsGolden(t *testing.T) {
	var golden struct {
		Seed   string `json:"seed"`
		Levels map[string]struct {
			StairsDown *Vec2 `json:"stairs_down"`
			StairsUp   *Vec2 `json:"stairs_up"`
			Teleporter *Vec2 `json:"teleporter"`
			Loot       []struct {
				ItemDefID string `json:"item_def_id"`
				Position  Vec2   `json:"position"`
			} `json:"loot"`
		} `json:"levels"`
	}
	loadGolden(t, "dungeon_stairs.json", &golden)
	rules := loadRules(t)

	level1, err := GenerateDungeonLevel(golden.Seed, -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate -1: %v", err)
	}
	if got := generatedStairPos(level1, stairsDownDefID); golden.Levels["-1"].StairsDown == nil || got != *golden.Levels["-1"].StairsDown {
		t.Fatalf("level -1 stairs_down = %+v, want %+v", got, golden.Levels["-1"].StairsDown)
	}
	if got := generatedStairPos(level1, stairsUpDefID); golden.Levels["-1"].StairsUp == nil || got != *golden.Levels["-1"].StairsUp {
		t.Fatalf("level -1 stairs_up = %+v, want %+v", got, golden.Levels["-1"].StairsUp)
	}
	if len(level1.teleporters) != 0 || golden.Levels["-1"].Teleporter != nil {
		t.Fatalf("level -1 teleporter count = %d, golden=%+v; want absent", len(level1.teleporters), golden.Levels["-1"].Teleporter)
	}

	level2, err := GenerateDungeonLevel(golden.Seed, -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate -2: %v", err)
	}
	if got := generatedStairPos(level2, stairsUpDefID); golden.Levels["-2"].StairsUp == nil || got != *golden.Levels["-2"].StairsUp {
		t.Fatalf("level -2 stairs_up = %+v, want %+v", got, golden.Levels["-2"].StairsUp)
	}
	if got := generatedStairPos(level2, stairsDownDefID); golden.Levels["-2"].StairsDown == nil || got != *golden.Levels["-2"].StairsDown {
		t.Fatalf("level -2 stairs_down = %+v, want %+v", got, golden.Levels["-2"].StairsDown)
	}
	if len(level2.teleporters) != 0 || golden.Levels["-2"].Teleporter != nil {
		t.Fatalf("level -2 teleporter count = %d, golden=%+v; want absent", len(level2.teleporters), golden.Levels["-2"].Teleporter)
	}
	if len(golden.Levels["-2"].Loot) != 0 || len(level2.loot) != 0 {
		t.Fatalf("level -2 generated loot = %+v, golden=%+v; want none", level2.loot, golden.Levels["-2"].Loot)
	}

	level3, err := GenerateDungeonLevel(golden.Seed, -3, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate -3: %v", err)
	}
	if got := generatedStairPos(level3, stairsUpDefID); golden.Levels["-3"].StairsUp == nil || got != *golden.Levels["-3"].StairsUp {
		t.Fatalf("level -3 stairs_up = %+v, want %+v", got, golden.Levels["-3"].StairsUp)
	}
	if got := generatedStairPos(level3, stairsDownDefID); golden.Levels["-3"].StairsDown == nil || got != *golden.Levels["-3"].StairsDown {
		t.Fatalf("level -3 stairs_down = %+v, want %+v", got, golden.Levels["-3"].StairsDown)
	}
	if got := generatedTeleporterPos(level3); golden.Levels["-3"].Teleporter == nil || got != *golden.Levels["-3"].Teleporter {
		t.Fatalf("level -3 teleporter = %+v, want %+v", got, golden.Levels["-3"].Teleporter)
	}
}

func TestBossFloorGenerationGolden(t *testing.T) {
	var golden struct {
		Seed      string           `json:"seed"`
		Level     int              `json:"level"`
		FloorSize DungeonFloorSize `json:"floor_size"`
		Expected  struct {
			BossCount              int    `json:"boss_count"`
			ChestCount             int    `json:"chest_count"`
			StairsDownCount        int    `json:"stairs_down_count"`
			TeleporterCount        int    `json:"teleporter_count"`
			StairsDownInitialState string `json:"stairs_down_initial_state"`
			TeleporterInitialState string `json:"teleporter_initial_state"`
			Boss                   struct {
				TemplateID       string  `json:"template_id"`
				BaseMonsterDefID string  `json:"base_monster_def_id"`
				VisualModel      string  `json:"visual_model"`
				VisualColor      string  `json:"visual_color"`
				VisualScale      float64 `json:"visual_scale"`
			} `json:"boss"`
		} `json:"expected"`
	}
	loadGolden(t, "boss_floor_-5.json", &golden)
	rules := loadRules(t)
	level, err := GenerateDungeonLevel(golden.Seed, golden.Level, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor: %v", err)
	}
	again, err := GenerateDungeonLevel(golden.Seed, golden.Level, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor again: %v", err)
	}
	if len(level.stairs) != len(again.stairs) || len(level.teleporters) != len(again.teleporters) || len(level.chests) != len(again.chests) || len(level.monsters) != len(again.monsters) {
		t.Fatalf("repeat boss floor counts changed")
	}
	for i := range level.monsters {
		if level.monsters[i].defID != again.monsters[i].defID || level.monsters[i].rarityID != again.monsters[i].rarityID || level.monsters[i].bossTemplate != again.monsters[i].bossTemplate || level.monsters[i].pos != again.monsters[i].pos {
			t.Fatalf("repeat boss monster %d = %+v, want %+v", i, again.monsters[i], level.monsters[i])
		}
	}
	if !isBossFloor(golden.Level, rules.DungeonGeneration) {
		t.Fatalf("level %d not classified as boss floor", golden.Level)
	}
	if rules.DungeonGeneration.BossFloor.FloorSize != golden.FloorSize {
		t.Fatalf("boss floor size = %+v, want %+v", rules.DungeonGeneration.BossFloor.FloorSize, golden.FloorSize)
	}
	if len(level.chests) != golden.Expected.ChestCount {
		t.Fatalf("chests = %d, want %d", len(level.chests), golden.Expected.ChestCount)
	}
	bosses := 0
	for _, monster := range level.monsters {
		if monster.isBoss {
			bosses++
			if monster.bossTemplate != golden.Expected.Boss.TemplateID {
				t.Fatalf("boss template = %s, want %s", monster.bossTemplate, golden.Expected.Boss.TemplateID)
			}
		}
	}
	if bosses != golden.Expected.BossCount {
		t.Fatalf("boss count = %d, want %d", bosses, golden.Expected.BossCount)
	}
	downCount := 0
	for _, stair := range level.stairs {
		if stair.defID == stairsDownDefID {
			downCount++
			if stair.state != golden.Expected.StairsDownInitialState {
				t.Fatalf("stairs_down state = %s, want %s", stair.state, golden.Expected.StairsDownInitialState)
			}
		}
	}
	if downCount != golden.Expected.StairsDownCount {
		t.Fatalf("stairs_down count = %d, want %d", downCount, golden.Expected.StairsDownCount)
	}
	if len(level.teleporters) != golden.Expected.TeleporterCount {
		t.Fatalf("teleporter count = %d, want %d", len(level.teleporters), golden.Expected.TeleporterCount)
	}
	if len(level.teleporters) > 0 && level.teleporters[0].state != golden.Expected.TeleporterInitialState {
		t.Fatalf("teleporter state = %s, want %s", level.teleporters[0].state, golden.Expected.TeleporterInitialState)
	}

	sim, err := NewSimWithWorld("sess", golden.Seed, rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	levelState, err := sim.ensureDungeonLevel(golden.Level)
	if err != nil {
		t.Fatalf("ensure boss floor: %v", err)
	}
	var bossView *EntityView
	for _, id := range sortedEntityIDs(levelState.entities) {
		view := levelState.entities[id].view()
		if view.IsBoss {
			bossView = &view
			break
		}
	}
	if bossView == nil {
		t.Fatal("missing boss entity view")
	}
	if bossView.BossTemplateID != golden.Expected.Boss.TemplateID || bossView.MonsterDefID != golden.Expected.Boss.BaseMonsterDefID {
		t.Fatalf("boss view ids = template %s def %s", bossView.BossTemplateID, bossView.MonsterDefID)
	}
	if bossView.VisualModel != golden.Expected.Boss.VisualModel || bossView.VisualTint != golden.Expected.Boss.VisualColor || bossView.VisualScale != golden.Expected.Boss.VisualScale {
		t.Fatalf("boss view visual = model %s tint %s scale %v", bossView.VisualModel, bossView.VisualTint, bossView.VisualScale)
	}
}

func TestBossPhaseTimingAndDodge(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	boss := findBossEntity(t, level)
	player := level.entities[sim.playerID]
	player.pos = boss.pos
	start := sim.Tick(nil)
	if !hasEvent(start, "boss_phase_started") || hasEvent(start, "player_damaged") {
		t.Fatalf("boss telegraph start events = %+v", start.Events)
	}
	for i := 0; i < 28; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
			t.Fatalf("player damaged during telegraph tick %d: %+v", i, res.Events)
		}
	}
	player.pos = Vec2{X: boss.pos.X - 5, Y: boss.pos.Y}
	activeStart := sim.Tick(nil)
	if !hasEvent(activeStart, "boss_phase_ended") || !hasEvent(activeStart, "boss_phase_started") {
		t.Fatalf("missing telegraph end/active start: %+v", activeStart.Events)
	}
	for i := 0; i < 4; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
			t.Fatalf("player damaged after breaking contact tick %d: %+v", i, res.Events)
		}
	}
}

func TestBossPatternDeckCycles(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_boss_deck", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	boss := findBossEntity(t, level)
	player := level.entities[sim.playerID]
	player.pos = Vec2{X: boss.pos.X - 6, Y: boss.pos.Y}

	first := sim.Tick(nil)
	if !hasBossPatternStart(first, "charged_melee") {
		t.Fatalf("first boss phase events = %+v, want charged_melee start", first.Events)
	}
	if boss.bossPatternDeckIndex != 0 || boss.bossPatternID != "charged_melee" {
		t.Fatalf("initial boss deck state = index %d pattern %s", boss.bossPatternDeckIndex, boss.bossPatternID)
	}

	waitForBossPatternStart(t, sim, "ground_slam", 180)
	if boss.bossPatternDeckIndex != 1 || boss.bossPatternID != "ground_slam" {
		t.Fatalf("second boss deck state = index %d pattern %s", boss.bossPatternDeckIndex, boss.bossPatternID)
	}

	waitForBossPatternStart(t, sim, "charged_melee", 220)
	if boss.bossPatternDeckIndex != 0 || boss.bossPatternID != "charged_melee" {
		t.Fatalf("wrapped boss deck state = index %d pattern %s", boss.bossPatternDeckIndex, boss.bossPatternID)
	}
}

func TestBossGroundSlamCircleHit(t *testing.T) {
	rules := loadRules(t)
	pattern, ok := rules.BossPatterns["ground_slam"]
	if !ok {
		t.Fatal("missing ground_slam pattern")
	}
	if len(pattern.Phases) < 2 {
		t.Fatalf("ground_slam phases = %d, want at least 2", len(pattern.Phases))
	}
	phase := pattern.Phases[1]
	if phase.Kind != "active" || phase.Shape != "circle" || phase.Radius <= 0 {
		t.Fatalf("ground_slam active phase = %+v, want circle with positive radius", phase)
	}
	boss := &entity{pos: Vec2{X: 10, Y: 10}}
	inside := &entity{pos: Vec2{X: boss.pos.X + phase.Radius - 0.01, Y: boss.pos.Y}}
	outside := &entity{pos: Vec2{X: boss.pos.X + phase.Radius + 0.01, Y: boss.pos.Y}}
	if !bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("ground_slam did not hit inside-radius player at distance %.2f radius %.2f", distance(boss.pos, inside.pos), phase.Radius)
	}
	if bossPhaseHitsPlayer(boss, outside, phase) {
		t.Fatalf("ground_slam hit outside-radius player at distance %.2f radius %.2f", distance(boss.pos, outside.pos), phase.Radius)
	}
}

func TestBossMovesDuringTelegraphAndPausesDuringActive(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_boss_move", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	boss := findBossEntity(t, level)
	player := level.entities[sim.playerID]
	player.pos = Vec2{X: boss.pos.X - 4, Y: boss.pos.Y}
	before := boss.pos
	start := sim.Tick(nil)
	if !hasEvent(start, "boss_phase_started") {
		t.Fatalf("boss start events = %+v, want boss_phase_started", start.Events)
	}
	if boss.pos == before {
		t.Fatalf("boss did not move during initial telegraph tick from %+v", before)
	}

	for guard := 0; guard < 40 && boss.bossPhaseKind != "active"; guard++ {
		sim.Tick(nil)
	}
	if boss.bossPhaseKind != "active" {
		t.Fatalf("boss phase = %s, want active", boss.bossPhaseKind)
	}
	activePos := boss.pos
	sim.Tick(nil)
	if boss.pos != activePos {
		t.Fatalf("boss moved during active phase from %+v to %+v", activePos, boss.pos)
	}
}

func TestBossAggroPreferredTargetWinsOverNearestPlayer(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_boss_preferred", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	guest := sim.levels[townLevel].entities[guestID]
	delete(sim.levels[townLevel].entities, guestID)
	guest.pos = Vec2{X: 0, Y: 0}
	level.entities[guestID] = guest
	sim.players[guestID].CurrentLevel = -5
	sim.savePlayer(sim.players[guestID])
	sim.usePlayer(sim.players[hostID])
	boss := findBossEntity(t, level)
	level.entities[hostID].pos = Vec2{X: boss.pos.X - 4, Y: boss.pos.Y}
	level.entities[guestID].pos = Vec2{X: boss.pos.X, Y: boss.pos.Y + 0.8}
	boss.aiTargetPlayerID = hostID
	boss.aiMode = monsterAIModeChase

	targetPlayer := sim.nearestLivingPlayerForMonster(level, boss)
	if targetPlayer == nil || targetPlayer.PlayerID != hostID {
		t.Fatalf("boss target = %+v, want host %d", targetPlayer, hostID)
	}
	beforeHostDist := distance(boss.pos, level.entities[hostID].pos)
	beforeGuestDist := distance(boss.pos, level.entities[guestID].pos)
	sim.Tick(nil)
	if distance(boss.pos, level.entities[hostID].pos) >= beforeHostDist-0.01 {
		t.Fatalf("boss did not move toward preferred host: before %.3f after %.3f", beforeHostDist, distance(boss.pos, level.entities[hostID].pos))
	}
	if distance(boss.pos, level.entities[guestID].pos) < beforeGuestDist-0.01 {
		t.Fatalf("boss moved toward nearer guest despite preferred host: before %.3f after %.3f", beforeGuestDist, distance(boss.pos, level.entities[guestID].pos))
	}
}

func TestBossDamagesStationaryPlayer(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_boss_damage", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	boss := findBossEntity(t, level)
	player := level.entities[sim.playerID]
	player.pos = boss.pos
	startHP := player.hp
	sawDamage := false
	for i := 0; i < 60; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
			sawDamage = true
			break
		}
	}
	if !sawDamage {
		t.Fatal("boss did not damage stationary player during active phase")
	}
	if player.hp >= startHP {
		t.Fatalf("player hp = %d, want below %d", player.hp, startHP)
	}
}

func TestBossFloorExitsUnlockAfterBossKill(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess", "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	player := level.entities[sim.playerID]
	down := sim.findStair(level, stairsDownDefID)
	if down == nil {
		t.Fatal("missing boss-floor down stairs")
	}
	player.pos = down.pos
	blockedDescend := sim.Tick([]Input{{MessageID: "blocked_descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	assertReject(t, blockedDescend, "blocked_descend", rules.DungeonGeneration.BossFloor.LockedExitReason)
	if !hasEvent(blockedDescend, "descend_blocked") {
		t.Fatalf("missing descend_blocked: %+v", blockedDescend.Events)
	}
	if teleporter := sim.findTeleporter(level); teleporter != nil {
		t.Fatalf("boss floor has unexpected teleporter: %+v", teleporter)
	}

	boss := findBossEntity(t, level)
	player.pos = boss.pos
	boss.hp = 1
	kill := sim.Tick([]Input{{MessageID: "kill_boss", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(boss.id)}}})
	assertAck(t, kill, "kill_boss")
	if !hasEvent(kill, "monster_killed") || !hasEvent(kill, "boss_killed") || !hasEvent(kill, "interactable_state_changed") {
		t.Fatalf("missing boss kill/unlock events: %+v", kill.Events)
	}
	var bossKilled Event
	for _, ev := range kill.Events {
		if ev.EventType == "boss_killed" {
			bossKilled = ev
			break
		}
	}
	if bossKilled.BossTemplateID != "cave_warden" || bossKilled.TargetEntityID != idStr(boss.id) {
		t.Fatalf("boss_killed event = %+v", bossKilled)
	}
	if down.state != interactableReady {
		t.Fatalf("down state = %s, want %s", down.state, interactableReady)
	}
	player.pos = down.pos
	descend := sim.TickResults([]Input{{MessageID: "descend_after_boss", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(descend) != 2 {
		t.Fatalf("descend after boss results = %d, want 2: %+v", len(descend), descend)
	}
	assertAck(t, descend[0], "descend_after_boss")
	assertLevelChanged(t, descend[0], -5, -6)
}

func TestDungeonMonsterGeneration(t *testing.T) {
	rules := loadRules(t)
	placement := rules.DungeonGeneration.MonsterPlacement
	allowedDefs := map[string]bool{placement.MonsterDefID: true}
	for _, entry := range placement.MonsterPool {
		allowedDefs[entry.MonsterDefID] = true
	}
	for _, levelNum := range []int{-1, -2} {
		t.Run(fmt.Sprintf("level_%d", levelNum), func(t *testing.T) {
			level, err := GenerateDungeonLevel("dungeon_monster_generation", levelNum, rules.DungeonGeneration)
			if err != nil {
				t.Fatalf("generate %d: %v", levelNum, err)
			}
			again, err := GenerateDungeonLevel("dungeon_monster_generation", levelNum, rules.DungeonGeneration)
			if err != nil {
				t.Fatalf("generate again %d: %v", levelNum, err)
			}
			if len(level.monsters) < placement.Count {
				t.Fatalf("level %d monsters = %d, want at least %d", levelNum, len(level.monsters), placement.Count)
			}
			if len(again.monsters) != len(level.monsters) {
				t.Fatalf("repeat level %d monsters = %d, want %d", levelNum, len(again.monsters), len(level.monsters))
			}
			archers := 0
			for i, monster := range level.monsters {
				if !allowedDefs[monster.defID] {
					t.Fatalf("level %d monster %d defID = %s, want one of %v", levelNum, i, monster.defID, allowedDefs)
				}
				if monster.defID == "dungeon_archer" {
					archers++
				}
				if monster != again.monsters[i] {
					t.Fatalf("level %d monster %d = %+v, repeat %+v", levelNum, i, monster, again.monsters[i])
				}
				if distance(monster.pos, rules.DungeonGeneration.PlayerSpawn) < placement.MinSpawnDistance {
					t.Fatalf("level %d monster %d too close to player spawn: %+v", levelNum, i, monster.pos)
				}
				if dungeonMonsterPositionBlocked(monster.pos, rules.DungeonGeneration, levelWithoutMonsterIndex(level, i)) {
					t.Fatalf("level %d monster %d blocked at %+v", levelNum, i, monster.pos)
				}
			}
			if archers == 0 {
				t.Fatalf("level %d generated no dungeon_archer: %+v", levelNum, level.monsters)
			}
		})
	}
}

func TestDungeonMonsterGenerationCreatesDeterministicPacks(t *testing.T) {
	rules := loadRules(t)
	placement := rules.DungeonGeneration.MonsterPlacement
	level, err := GenerateDungeonLevel("v76_pack_generation", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	again, err := GenerateDungeonLevel("v76_pack_generation", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate again: %v", err)
	}
	if len(level.monsters) != len(again.monsters) {
		t.Fatalf("repeat monster count = %d, want %d", len(again.monsters), len(level.monsters))
	}
	for i := range level.monsters {
		if level.monsters[i] != again.monsters[i] {
			t.Fatalf("monster %d = %+v, repeat %+v", i, level.monsters[i], again.monsters[i])
		}
	}

	packs := map[string][]generatedMonster{}
	for _, monster := range level.monsters {
		if monster.packID == "" {
			continue
		}
		packs[monster.packID] = append(packs[monster.packID], monster)
	}
	if len(packs) < placement.PackCount.Min || len(packs) > placement.PackCount.Max {
		t.Fatalf("pack count = %d, want %d..%d in %+v", len(packs), placement.PackCount.Min, placement.PackCount.Max, packs)
	}
	for packID, monsters := range packs {
		if len(monsters) < placement.PackSize.Min || len(monsters) > placement.PackSize.Max {
			t.Fatalf("%s size = %d, want %d..%d", packID, len(monsters), placement.PackSize.Min, placement.PackSize.Max)
		}
		frontlineCount := 0
		rangedCount := 0
		leaderCount := 0
		for _, monster := range monsters {
			switch rules.DungeonGeneration.MonsterRole(monster.defID) {
			case "frontline":
				frontlineCount++
			case "ranged":
				rangedCount++
			}
			if monster.packLeader {
				leaderCount++
				if monster.rarityID != "champion" {
					t.Fatalf("%s leader rarity = %s, want champion", packID, monster.rarityID)
				}
			}
		}
		if frontlineCount < placement.PackComposition.FrontlineMin {
			t.Fatalf("%s frontline count = %d, want at least %d", packID, frontlineCount, placement.PackComposition.FrontlineMin)
		}
		if rangedCount > placement.PackComposition.RangedMax {
			t.Fatalf("%s ranged count = %d, want at most %d", packID, rangedCount, placement.PackComposition.RangedMax)
		}
		if leaderCount > 1 {
			t.Fatalf("%s leader count = %d, want at most 1", packID, leaderCount)
		}
		for i := range monsters {
			for j := i + 1; j < len(monsters); j++ {
				if got := distance(monsters[i].pos, monsters[j].pos); got > placement.PackMemberRadius*2+0.000001 {
					t.Fatalf("%s monsters too far apart: %.3f > %.3f: %+v %+v", packID, got, placement.PackMemberRadius*2, monsters[i], monsters[j])
				}
			}
		}
	}
}

func TestDungeonMonsterGenerationCanForceElitePackLeaders(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 100
	level, err := GenerateDungeonLevel("v79_forced_elite_packs", -1, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	packLeaders := map[string]int{}
	for _, monster := range level.monsters {
		if monster.packID == "" || !monster.packLeader {
			continue
		}
		packLeaders[monster.packID]++
		if monster.rarityID != "champion" {
			t.Fatalf("%s leader rarity = %s, want champion", monster.packID, monster.rarityID)
		}
	}
	if len(packLeaders) < rules.DungeonGeneration.MonsterPlacement.PackCount.Min {
		t.Fatalf("elite leaders = %+v, want one per generated pack", packLeaders)
	}
	for packID, count := range packLeaders {
		if count != 1 {
			t.Fatalf("%s leader count = %d, want 1", packID, count)
		}
	}
}

func TestDungeonObstacleGeneration(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("v40_obstacles", -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	again, err := GenerateDungeonLevel("v40_obstacles", -2, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate again: %v", err)
	}
	if len(level.walls) != len(again.walls) {
		t.Fatalf("repeat walls = %d, want %d", len(again.walls), len(level.walls))
	}
	for i := range level.walls {
		if level.walls[i] != again.walls[i] {
			t.Fatalf("wall %d = %+v, repeat %+v", i, level.walls[i], again.walls[i])
		}
	}
	generatedCount := 0
	shapeFamilies := map[string]bool{}
	for _, wall := range level.walls {
		if wall.source != "generated" {
			continue
		}
		generatedCount++
		shapeFamilies[wall.shapeFamily] = true
		if !wallInsideDungeonFloor(wall, rules.DungeonGeneration) {
			t.Fatalf("generated wall outside floor: %+v", wall)
		}
		if !wallClearsGeneratedTargets(wall, rules.DungeonGeneration, level) {
			t.Fatalf("generated wall overlaps target clearance: %+v", wall)
		}
	}
	if generatedCount < rules.DungeonGeneration.ObstacleGeneration.TargetGroupCount.Min {
		t.Fatalf("generated walls = %d, want at least %d", generatedCount, rules.DungeonGeneration.ObstacleGeneration.TargetGroupCount.Min)
	}
	if len(shapeFamilies) < 2 {
		t.Fatalf("shape families = %+v, want at least two", shapeFamilies)
	}
}

func TestDungeonObstaclesGolden(t *testing.T) {
	var golden struct {
		Seed     string `json:"seed"`
		Level    int    `json:"level"`
		Expected struct {
			FloorSize                 DungeonFloorSize `json:"floor_size"`
			MinimumWallCount          int              `json:"minimum_wall_count"`
			MinimumGeneratedWallCount int              `json:"minimum_generated_wall_count"`
			ShapeFamilies             []string         `json:"shape_families"`
			Walls                     []struct {
				ID          string `json:"id"`
				Position    Vec2   `json:"position"`
				Size        Vec2   `json:"size"`
				Source      string `json:"source"`
				ShapeFamily string `json:"shape_family"`
			} `json:"walls"`
			ReachableTargets []struct {
				Kind     string `json:"kind"`
				DefID    string `json:"def_id"`
				Position Vec2   `json:"position"`
			} `json:"reachable_targets"`
		} `json:"expected"`
	}
	loadGolden(t, "dungeon_obstacles.json", &golden)
	rules := loadRules(t)
	if rules.DungeonGeneration.FloorSize != golden.Expected.FloorSize {
		t.Fatalf("floor_size = %+v, want %+v", rules.DungeonGeneration.FloorSize, golden.Expected.FloorSize)
	}
	level, err := GenerateDungeonLevel(golden.Seed, golden.Level, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(level.walls) != len(golden.Expected.Walls) {
		t.Fatalf("wall count = %d, want golden %d", len(level.walls), len(golden.Expected.Walls))
	}
	if len(level.walls) < golden.Expected.MinimumWallCount {
		t.Fatalf("wall count = %d, want at least %d", len(level.walls), golden.Expected.MinimumWallCount)
	}
	generatedCount := 0
	shapeFamilies := map[string]bool{}
	for i, want := range golden.Expected.Walls {
		got := level.walls[i]
		if got.source == "generated" {
			generatedCount++
			shapeFamilies[got.shapeFamily] = true
		}
		if id := wallID(level.levelNum, i); id != want.ID {
			t.Fatalf("wall %d id = %s, want %s", i, id, want.ID)
		}
		if got.pos != want.Position || got.size != want.Size || got.source != want.Source || got.shapeFamily != want.ShapeFamily {
			t.Fatalf("wall %d = pos %+v size %+v source %s shape %s, want %+v", i, got.pos, got.size, got.source, got.shapeFamily, want)
		}
	}
	if generatedCount < golden.Expected.MinimumGeneratedWallCount {
		t.Fatalf("generated walls = %d, want at least %d", generatedCount, golden.Expected.MinimumGeneratedWallCount)
	}
	if len(shapeFamilies) != len(golden.Expected.ShapeFamilies) {
		t.Fatalf("shape families = %+v, want %+v", shapeFamilies, golden.Expected.ShapeFamilies)
	}
	for _, want := range golden.Expected.ShapeFamilies {
		if !shapeFamilies[want] {
			t.Fatalf("missing shape family %s in %+v", want, shapeFamilies)
		}
	}
	targets := generatedReachabilityTargets(level)
	start := generatedReachabilityStart(rules.DungeonGeneration, level)
	for i, got := range targets {
		if !generatedTargetReachableFrom(rules.DungeonGeneration, level, start, got.pos) {
			t.Fatalf("target %d %s at %+v unreachable from %+v", i, got.kind, got.pos, start)
		}
	}
}

func goldenReachabilityKindAndDefID(kind string) (string, string) {
	if strings.HasPrefix(kind, "loot:") {
		return "loot", strings.TrimPrefix(kind, "loot:")
	}
	if strings.HasPrefix(kind, "monster:") {
		return "monster", strings.TrimPrefix(kind, "monster:")
	}
	if kind == stairsUpDefID || kind == stairsDownDefID || kind == teleporterDefID {
		return kind, ""
	}
	return "chest", kind
}

func TestGeneratedDungeonTargetsReachable(t *testing.T) {
	rules := loadRules(t)
	for _, levelNum := range []int{-1, -2, -3, -4} {
		level, err := GenerateDungeonLevel("v40_reachability", levelNum, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("level %d generate: %v", levelNum, err)
		}
		if err := validateGeneratedDungeonReachability(rules.DungeonGeneration, level); err != nil {
			t.Fatalf("level %d reachability: %v", levelNum, err)
		}
	}
}

func TestGeneratedDungeonUnreachableTuningFailsClearly(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.DungeonGeneration.ObstacleGeneration.MaxAttempts = 2
	rules.DungeonGeneration.ObstacleGeneration.TargetGroupCount = IntRange{Min: 1, Max: 1}
	rules.DungeonGeneration.ObstacleGeneration.ShapeWeights = ObstacleShapeWeights{Block: 1}
	rules.DungeonGeneration.ObstacleGeneration.SolidBlock = SolidBlockRules{
		MinSize: Vec2{X: 98, Y: 48},
		MaxSize: Vec2{X: 98, Y: 48},
	}
	_, err := GenerateDungeonLevel("v40_bad_obstacles", -2, rules.DungeonGeneration)
	if err == nil {
		t.Fatal("generate succeeded with impossible obstacle tuning")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "could not place reachable obstacles") {
		t.Fatalf("error = %q, want clear obstacle placement failure", got)
	}
}

func TestBossFloorExcludesGeneratedObstacles(t *testing.T) {
	rules := loadRules(t)
	level, err := GenerateDungeonLevel("boss_floor_gate", -5, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate boss floor: %v", err)
	}
	for _, wall := range level.walls {
		if wall.source == "generated" {
			t.Fatalf("boss floor has generated obstacle wall: %+v", wall)
		}
	}
	if len(level.walls) != 4 {
		t.Fatalf("boss floor walls = %d, want perimeter only", len(level.walls))
	}
}

func TestGeneratedObstacleCollisionPaths(t *testing.T) {
	t.Run("player movement blocks", func(t *testing.T) {
		sim := MustNewSim("sess_generated_wall_move", "01", loadRules(t))
		player := sim.entities[sim.playerID]
		player.pos = Vec2{X: 10, Y: 5}
		addGeneratedWallToActiveLevel(sim, Vec2{X: 10.5, Y: 5}, Vec2{X: 1, Y: 4})
		blocked := sim.Tick([]Input{{MessageID: "push_generated_wall", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 3}}})
		assertAck(t, blocked, "push_generated_wall")
		if player.pos.X > 10.01 {
			t.Fatalf("player moved through generated wall to %+v", player.pos)
		}
	})

	t.Run("pathing routes around wall", func(t *testing.T) {
		nav := testNav()
		wall := wallObstacle{pos: Vec2{X: 1, Y: 0}, size: Vec2{X: 1, Y: 3}, source: "generated"}
		blocked := func(gx, gy int) bool {
			center := gridToWorld(nav, gridCell{x: gx, y: gy})
			return circleIntersectsAABB(center, playerRadius, wall.pos, wall.size)
		}
		steps, ok := PlanPath(nav, Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 0}, blocked)
		if !ok {
			t.Fatal("path around generated wall failed")
		}
		if len(steps) <= 3 {
			t.Fatalf("path steps = %v, want routed path longer than direct line", steps)
		}
	})

	t.Run("projectile sweep blocks", func(t *testing.T) {
		sim := combatControlLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
		player := sim.entities[sim.playerID]
		player.pos = Vec2{X: 3, Y: 5}
		addGeneratedWallToActiveLevel(sim, Vec2{X: 6, Y: 5}, Vec2{X: 1, Y: 4})
		fire := sim.Tick([]Input{{MessageID: "fire_generated_wall", Type: "directional_attack_intent", DirectionalAttack: &DirectionalAttackIntent{Direction: Vec2{X: 1}}}})
		assertAck(t, fire, "fire_generated_wall")
		var resolved TickResult
		for i := 0; i < 20; i++ {
			resolved = sim.Tick(nil)
			if hasEvent(resolved, "projectile_blocked") || hasEvent(resolved, "monster_damaged") {
				break
			}
		}
		if !hasEvent(resolved, "projectile_blocked") {
			t.Fatalf("projectile events = %+v, want projectile_blocked by generated wall", resolved.Events)
		}
		if hasEvent(resolved, "monster_damaged") {
			t.Fatalf("projectile damaged monster through generated wall: %+v", resolved.Events)
		}
	})

	t.Run("loot drop avoids wall", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_generated_wall_drop", "01", loadRules(t), "inventory_lab")
		if err != nil {
			t.Fatalf("new inventory lab: %v", err)
		}
		source := Vec2{X: 10, Y: 10}
		blockedFirstCandidate := Vec2{X: source.X + 1, Y: source.Y}
		addGeneratedWallToActiveLevel(sim, blockedFirstCandidate, Vec2{X: 1, Y: 1})
		pos, ok := sim.findEntityLootDropPosition(source, monsterRadius)
		if !ok {
			t.Fatal("no loot drop position")
		}
		if pos == blockedFirstCandidate || sim.lootDropBlocked(pos) {
			t.Fatalf("drop position = %+v, want generated-wall-free tile", pos)
		}
	})

	t.Run("travel arrival avoids wall", func(t *testing.T) {
		sim := MustNewSim("sess_generated_wall_arrival", "01", loadRules(t))
		level := sim.activeLevel()
		marker := Vec2{X: 10, Y: 5}
		firstCandidate := Vec2{X: 11, Y: 5}
		addGeneratedWallToActiveLevel(sim, firstCandidate, Vec2{X: 1, Y: 1})
		got := sim.travelArrivalPosition(level, marker, sim.playerID)
		if got == firstCandidate || sim.travelArrivalBlocked(level, got, sim.playerID) {
			t.Fatalf("arrival = %+v, want generated-wall-free tile", got)
		}
	})
}

func addGeneratedWallToActiveLevel(sim *Sim, pos, size Vec2) {
	level := sim.activeLevel()
	level.walls = append(level.walls, wallObstacle{pos: pos, size: size, source: "generated", shapeFamily: "line"})
	sim.syncCompatibilityFields()
}

func TestChampionMonstersSpawnWithCommonMinions(t *testing.T) {
	rules := loadRules(t)
	var level generatedDungeonLevel
	var err error
	championIndex := -1
	for i := 0; i < 64 && championIndex < 0; i++ {
		level, err = GenerateDungeonLevel(fmt.Sprintf("v30_monster_rarity_%02d", i), -1, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("generate seed %d: %v", i, err)
		}
		for j, monster := range level.monsters {
			if monster.rarityID == "champion" && championHasCommonMinions(level.monsters, j) {
				championIndex = j
				break
			}
		}
	}
	if championIndex < 0 {
		t.Fatalf("missing champion in generated monsters: %+v", level.monsters)
	}
	if championIndex+championCommonMinionCount >= len(level.monsters) {
		t.Fatalf("champion at %d does not have enough following minions in %+v", championIndex, level.monsters)
	}
	champion := level.monsters[championIndex]
	for i := 1; i <= championCommonMinionCount; i++ {
		minion := level.monsters[championIndex+i]
		if minion.rarityID != "common" {
			t.Fatalf("champion minion %d rarity = %s, want common", i, minion.rarityID)
		}
		if distance(champion.pos, minion.pos) > 3.0 {
			t.Fatalf("champion minion %d too far: champion %+v minion %+v", i, champion.pos, minion.pos)
		}
	}
}

func championHasCommonMinions(monsters []generatedMonster, championIndex int) bool {
	if championIndex+championCommonMinionCount >= len(monsters) {
		return false
	}
	champion := monsters[championIndex]
	for i := 1; i <= championCommonMinionCount; i++ {
		minion := monsters[championIndex+i]
		if minion.rarityID != "common" || distance(champion.pos, minion.pos) > 3.0 {
			return false
		}
	}
	return true
}

func TestGuardedChestGenerationGolden(t *testing.T) {
	var golden struct {
		Level             int `json:"level"`
		BaseMonsterCount  int `json:"base_monster_count"`
		MonsterCountBonus int `json:"monster_count_bonus"`
		Cases             []struct {
			Name                 string `json:"name"`
			Seed                 string `json:"seed"`
			ExpectedMonsterCount int    `json:"expected_monster_count"`
			ExpectedChest        *struct {
				InteractableDefID string `json:"interactable_def_id"`
				LootTable         string `json:"loot_table"`
				Position          Vec2   `json:"position"`
			} `json:"expected_chest"`
		} `json:"cases"`
	}
	loadGolden(t, "guarded_chest_generation.json", &golden)
	rules := loadRules(t)
	for _, c := range golden.Cases {
		level, err := GenerateDungeonLevel(c.Seed, golden.Level, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("%s: generate: %v", c.Name, err)
		}
		if c.ExpectedChest == nil {
			wantCount := rules.DungeonGeneration.MonsterPlacement.Count
			if len(level.monsters) < wantCount {
				t.Fatalf("%s: monsters = %d, want at least rule-derived base count %d", c.Name, len(level.monsters), wantCount)
			}
			if len(level.chests) != 0 {
				t.Fatalf("%s: chests = %+v, want none", c.Name, level.chests)
			}
			continue
		}
		wantCount := rules.DungeonGeneration.MonsterPlacement.Count + rules.DungeonGeneration.ChestPlacement.MonsterCountBonus
		if len(level.monsters) < wantCount {
			t.Fatalf("%s: monsters = %d, want at least rule-derived guarded count %d", c.Name, len(level.monsters), wantCount)
		}
		if len(level.chests) != 1 {
			t.Fatalf("%s: chests = %+v, want one", c.Name, level.chests)
		}
		got := level.chests[0]
		if got.defID != c.ExpectedChest.InteractableDefID || got.lootTable != c.ExpectedChest.LootTable || got.pos != c.ExpectedChest.Position {
			t.Fatalf("%s: chest = %+v, want %+v", c.Name, got, *c.ExpectedChest)
		}
	}
}

func TestGeneratedDungeonSourcesUseDepthLootTables(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.DungeonGeneration.ChestPlacement.ChanceWeight = 1
	rules.DungeonGeneration.ChestPlacement.NoChestWeight = 0
	cases := []struct {
		levelNum       int
		chestLootTable string
	}{
		{-1, "guarded_chest_drop_depth_1"},
		{-2, "guarded_chest_drop_depth_2"},
		{-3, "guarded_chest_drop_depth_3_plus"},
		{-4, "guarded_chest_drop_depth_3_plus"},
	}
	for _, c := range cases {
		level, err := GenerateDungeonLevel("v29_source_tables", c.levelNum, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("level %d generate: %v", c.levelNum, err)
		}
		if len(level.monsters) == 0 {
			t.Fatalf("level %d: missing generated monsters", c.levelNum)
		}
		for _, monster := range level.monsters {
			rarity, ok := rules.DungeonGeneration.MonsterRarity(monster.rarityID)
			if !ok {
				t.Fatalf("level %d monster rarity %q missing from rules", c.levelNum, monster.rarityID)
			}
			effectiveDepth := absInt(c.levelNum) + rarity.LootDepthOffset
			band, ok := rules.DungeonGeneration.LootBandForDepth(effectiveDepth)
			if !ok {
				t.Fatalf("level %d effective depth %d has no loot band", c.levelNum, effectiveDepth)
			}
			if monster.lootTable != band.MonsterLootTable {
				t.Fatalf("level %d rarity %s monster lootTable = %s, want %s", c.levelNum, monster.rarityID, monster.lootTable, band.MonsterLootTable)
			}
		}
		if len(level.chests) != 1 {
			t.Fatalf("level %d: chests = %+v, want one", c.levelNum, level.chests)
		}
		if got := level.chests[0].lootTable; got != c.chestLootTable {
			t.Fatalf("level %d chest lootTable = %s, want %s", c.levelNum, got, c.chestLootTable)
		}
	}
}

func TestGeneratedDungeonMonsterRarityGolden(t *testing.T) {
	var golden struct {
		GeneratedCases []struct {
			Name             string `json:"name"`
			Seed             string `json:"seed"`
			Level            int    `json:"level"`
			ExpectedMonsters []struct {
				Index               int         `json:"index"`
				Rarity              string      `json:"rarity"`
				LootTable           string      `json:"loot_table"`
				MaxHP               int         `json:"max_hp"`
				AttackDamage        DamageRange `json:"attack_damage"`
				Armor               float64     `json:"armor"`
				HitChance           float64     `json:"hit_chance"`
				CritChance          float64     `json:"crit_chance"`
				BlockPercent        float64     `json:"block_percent"`
				AttackCooldownTicks int         `json:"attack_cooldown_ticks"`
				XPReward            int         `json:"xp_reward"`
			} `json:"expected_monsters"`
		} `json:"generated_cases"`
	}
	loadGolden(t, "monster_rarity.json", &golden)
	rules := cloneRules(loadRules(t))
	rules.DungeonGeneration.MonsterPlacement.ElitePackChance = 0
	for _, c := range golden.GeneratedCases {
		level, err := GenerateDungeonLevel(c.Seed, c.Level, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("%s: generate: %v", c.Name, err)
		}
		for _, expected := range c.ExpectedMonsters {
			if expected.Index >= len(level.monsters) {
				t.Fatalf("%s: missing monster index %d in %+v", c.Name, expected.Index, level.monsters)
			}
			got := level.monsters[expected.Index]
			if got.rarityID != expected.Rarity || got.lootTable != expected.LootTable {
				t.Fatalf("%s monster %d rarity/table = %s/%s, want %s/%s", c.Name, expected.Index, got.rarityID, got.lootTable, expected.Rarity, expected.LootTable)
			}
		}

		sim, err := NewSimWithWorld("sess_"+c.Name, c.Seed, rules, "dungeon_levels")
		if err != nil {
			t.Fatalf("%s: new sim: %v", c.Name, err)
		}
		for levelNum := 0; levelNum > c.Level; levelNum-- {
			results := descendFromCurrentLevel(t, sim, fmt.Sprintf("%s_descend_%d", c.Name, -levelNum+1))
			assertLevelChanged(t, results[0], levelNum, levelNum-1)
		}
		monsters := liveDungeonMonsters(sim.activeLevel())
		for _, expected := range c.ExpectedMonsters {
			if expected.Index >= len(monsters) {
				t.Fatalf("%s: missing live monster index %d in %+v", c.Name, expected.Index, monsters)
			}
			got := monsters[expected.Index]
			if got.monsterRarityID != expected.Rarity || got.lootTable != expected.LootTable {
				t.Fatalf("%s entity %d rarity/table = %s/%s, want %s/%s", c.Name, expected.Index, got.monsterRarityID, got.lootTable, expected.Rarity, expected.LootTable)
			}
			if got.maxHP != expected.MaxHP || got.hp != expected.MaxHP {
				t.Fatalf("%s entity %d hp = %d/%d, want %d", c.Name, expected.Index, got.hp, got.maxHP, expected.MaxHP)
			}
			if got.monsterAttackDamage == nil || *got.monsterAttackDamage != expected.AttackDamage {
				t.Fatalf("%s entity %d attack = %+v, want %+v", c.Name, expected.Index, got.monsterAttackDamage, expected.AttackDamage)
			}
			if got.monsterArmor != expected.Armor || math.Abs(got.monsterHitChance-expected.HitChance) > 1e-12 || math.Abs(got.monsterCritChance-expected.CritChance) > 1e-12 || got.monsterBlockPercent != expected.BlockPercent || got.monsterAttackCooldown != expected.AttackCooldownTicks {
				t.Fatalf("%s entity %d stats = armor %.3f hit %.3f crit %.3f block %.3f cooldown %d, want %.3f %.3f %.3f %.3f %d", c.Name, expected.Index, got.monsterArmor, got.monsterHitChance, got.monsterCritChance, got.monsterBlockPercent, got.monsterAttackCooldown, expected.Armor, expected.HitChance, expected.CritChance, expected.BlockPercent, expected.AttackCooldownTicks)
			}
			if got.monsterXPReward != expected.XPReward {
				t.Fatalf("%s entity %d xp = %d, want %d", c.Name, expected.Index, got.monsterXPReward, expected.XPReward)
			}
			view := got.view()
			if view.Rarity != expected.Rarity {
				t.Fatalf("%s entity %d view rarity = %s, want %s", c.Name, expected.Index, view.Rarity, expected.Rarity)
			}
		}
	}
}

func TestStaticWorldMonstersDoNotGetGeneratedRarity(t *testing.T) {
	sim, err := NewSimWithWorld("sess_static_rarity", "static_rarity_seed", loadRules(t), "vertical_slice")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	monster := findMonsterByDef(sim, monsterDefID)
	if monster == nil {
		t.Fatal("missing vertical slice monster")
	}
	if monster.monsterRarityID != "" || monster.monsterAttackDamage != nil || monster.monsterAttackCooldown != 0 || monster.monsterArmor != 0 || monster.monsterHitChance != 0 || monster.monsterCritChance != 0 || monster.monsterBlockPercent != 0 || monster.monsterXPReward != 0 {
		t.Fatalf("static monster has generated rarity fields: %+v", monster)
	}
	if monster.view().Rarity != "" {
		t.Fatalf("static monster view rarity = %q, want empty", monster.view().Rarity)
	}
}

func TestCoopPlayersHaveIndependentLevelsMovementAndVisibility(t *testing.T) {
	sim, err := NewSimWithWorld("sess_coop_levels", "coop_levels_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	sim.SetPlayerMetadata(hostID, "acct_host", "char_host", "Host", "host")

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if down == nil {
		t.Fatal("missing town down stair")
	}
	sim.entities[hostID].pos = down.pos
	results := sim.TickResults([]Input{{MessageID: "host_descend", ActorPlayerID: hostID, Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("host descend results = %d, want 2", len(results))
	}
	if sim.players[hostID].CurrentLevel != -1 {
		t.Fatalf("host level = %d, want -1", sim.players[hostID].CurrentLevel)
	}

	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", sim.rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	if sim.players[guestID].CurrentLevel != townLevel {
		t.Fatalf("guest level = %d, want town", sim.players[guestID].CurrentLevel)
	}
	if countPlayers(sim.SnapshotForPlayer(hostID).Entities) != 1 {
		t.Fatalf("host should not see town guest while in dungeon")
	}
	if countPlayers(sim.SnapshotForPlayer(guestID).Entities) != 1 {
		t.Fatalf("guest should not see dungeon host while in town")
	}

	sim.usePlayer(sim.players[guestID])
	down = sim.findStair(sim.activeLevel(), stairsDownDefID)
	if down == nil {
		t.Fatal("missing guest town down stair")
	}
	sim.entities[guestID].pos = down.pos
	sim.savePlayer(sim.players[guestID])
	results = sim.TickResults([]Input{{MessageID: "guest_descend", ActorPlayerID: guestID, Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("guest descend results = %d, want 2", len(results))
	}
	if got := countPlayers(sim.SnapshotForPlayer(hostID).Entities); got != 2 {
		t.Fatalf("same-level host player count = %d, want 2", got)
	}
	if got := countPlayers(sim.SnapshotForPlayer(guestID).Entities); got != 2 {
		t.Fatalf("same-level guest player count = %d, want 2", got)
	}

	hostBefore := sim.levels[-1].entities[hostID].pos
	guestBefore := sim.levels[-1].entities[guestID].pos
	sim.TickResults([]Input{{MessageID: "guest_move", ActorPlayerID: guestID, Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})
	hostAfter := sim.levels[-1].entities[hostID].pos
	guestAfter := sim.levels[-1].entities[guestID].pos
	if hostAfter != hostBefore {
		t.Fatalf("host moved from %+v to %+v after guest input", hostBefore, hostAfter)
	}
	if guestAfter == guestBefore {
		t.Fatalf("guest did not move from %+v", guestBefore)
	}
}

func TestCoopGuestsSpawnOnFreeTownTiles(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_coop_spawn", "coop_spawn_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	guestAID, err := sim.AddGuestPlayer("acct_guest_a", "char_guest_a", "Guest A", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest A: %v", err)
	}
	guestBID, err := sim.AddGuestPlayer("acct_guest_b", "char_guest_b", "Guest B", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest B: %v", err)
	}

	level := sim.levels[townLevel]
	positions := map[uint64]Vec2{
		hostID:   level.entities[hostID].pos,
		guestAID: level.entities[guestAID].pos,
		guestBID: level.entities[guestBID].pos,
	}
	for id, pos := range positions {
		for otherID, otherPos := range positions {
			if id >= otherID {
				continue
			}
			if circlesOverlap(pos, playerRadius, otherPos, playerRadius) {
				t.Fatalf("players %d and %d overlap at %+v / %+v", id, otherID, pos, otherPos)
			}
		}
	}
}

func TestRespawnPlayerInTownUsesFreeTile(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_coop_respawn", "coop_respawn_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	level := sim.levels[townLevel]
	hostPos := level.entities[hostID].pos
	delete(level.entities, guestID)

	if err := sim.RespawnPlayerInTown(guestID); err != nil {
		t.Fatalf("respawn guest: %v", err)
	}
	guestPos := sim.levels[townLevel].entities[guestID].pos
	if circlesOverlap(hostPos, playerRadius, guestPos, playerRadius) {
		t.Fatalf("respawned guest overlaps host at %+v / %+v", guestPos, hostPos)
	}
}

func TestCoopThreePlayersSameLevelVisibilityAndActorMovement(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_three_coop", "cafebabecafebabe", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	sim.SetPlayerMetadata(hostID, "acct_host", "char_host", "Host", "host")
	guestAID, err := sim.AddGuestPlayer("acct_guest_a", "char_guest_a", "Guest A", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest A: %v", err)
	}
	guestBID, err := sim.AddGuestPlayer("acct_guest_b", "char_guest_b", "Guest B", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest B: %v", err)
	}

	for _, playerID := range []uint64{hostID, guestAID, guestBID} {
		sim.usePlayer(sim.players[playerID])
		down := sim.findStair(sim.activeLevel(), stairsDownDefID)
		if down == nil {
			t.Fatalf("missing down stair for player %d", playerID)
		}
		sim.entities[playerID].pos = down.pos
		sim.savePlayer(sim.players[playerID])
		results := sim.TickResults([]Input{{MessageID: fmt.Sprintf("descend_%d", playerID), ActorPlayerID: playerID, Type: "descend_intent", Descend: &DescendIntent{}}})
		if len(results) != 2 {
			t.Fatalf("descend results for %d = %d, want 2", playerID, len(results))
		}
	}

	if got := countPlayers(sim.SnapshotForPlayer(hostID).Entities); got != 3 {
		t.Fatalf("host same-level player count = %d, want 3", got)
	}
	if got := countPlayers(sim.SnapshotForPlayer(guestAID).Entities); got != 3 {
		t.Fatalf("guest A same-level player count = %d, want 3", got)
	}
	if got := countPlayers(sim.SnapshotForPlayer(guestBID).Entities); got != 3 {
		t.Fatalf("guest B same-level player count = %d, want 3", got)
	}
	hostBefore := sim.levels[-1].entities[hostID].pos
	guestABefore := sim.levels[-1].entities[guestAID].pos
	guestBBefore := sim.levels[-1].entities[guestBID].pos
	sim.TickResults([]Input{{MessageID: "guest_b_move", ActorPlayerID: guestBID, Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 1}}})
	if got := sim.levels[-1].entities[hostID].pos; got != hostBefore {
		t.Fatalf("host moved from %+v to %+v after third-player input", hostBefore, got)
	}
	if got := sim.levels[-1].entities[guestAID].pos; got != guestABefore {
		t.Fatalf("guest A moved from %+v to %+v after third-player input", guestABefore, got)
	}
	if got := sim.levels[-1].entities[guestBID].pos; got == guestBBefore {
		t.Fatalf("guest B did not move from %+v", guestBBefore)
	}
}

func TestCoopActorScopedLootExperienceAndMonsterTargeting(t *testing.T) {
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	forceMonsterHitChance(rules, monsterDefID, 1.0)
	dmg := DamageRange{Min: 1, Max: 1}
	dummy := rules.Monsters[monsterDefID]
	dummy.AttackDamage = &dmg
	dummy.AttackCooldown = 1
	dummy.RetaliationDamage = nil
	dummy.MaxHP = 1
	dummy.XPReward = 10
	rules.Monsters[monsterDefID] = dummy

	sim, err := NewSimWithWorld("sess_coop_rewards", "coop_rewards_seed", rules, "vertical_slice")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	sim.SetPlayerMetadata(hostID, "acct_host", "char_host", "Host", "host")
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}

	monster := findMonsterByDef(sim, monsterDefID)
	if monster == nil {
		t.Fatal("missing monster")
	}
	sim.entities[hostID].pos = monster.pos
	sim.entities[guestID].pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}
	kill := sim.Tick([]Input{{MessageID: "host_kill", ActorPlayerID: hostID, Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, kill, "host_kill")
	if !hasEvent(kill, "monster_killed") || !hasEvent(kill, "experience_gained") {
		t.Fatalf("kill events = %+v", kill.Events)
	}
	if sim.players[hostID].Progression.Experience != 10 || sim.players[guestID].Progression.Experience != 10 {
		t.Fatalf("xp host=%d guest=%d, want both nearby players", sim.players[hostID].Progression.Experience, sim.players[guestID].Progression.Experience)
	}
	if owners := changeOwnersForOp(kill, OpCharacterProgressionUpdate); !sameUint64Set(owners, []uint64{hostID, guestID}) {
		t.Fatalf("progression update owners = %v, want host+guest", owners)
	}
	if owners := changeOwnersForOp(kill, OpSkillProgressionUpdate); !sameUint64Set(owners, []uint64{hostID, guestID}) {
		t.Fatalf("skill progression update owners = %v, want host+guest", owners)
	}
	var loot *entity
	for _, e := range sim.entities {
		if e.kind == lootEntity {
			loot = e
			break
		}
	}
	if loot == nil {
		t.Fatal("missing loot after kill")
	}
	sim.entities[guestID].pos = loot.pos
	pickup := sim.Tick([]Input{{MessageID: "guest_pickup", ActorPlayerID: guestID, Type: "action_intent", Action: &ActionIntent{TargetID: idStr(loot.id)}}})
	assertAck(t, pickup, "guest_pickup")
	if len(sim.players[hostID].Inventory) != 0 || len(sim.players[guestID].Inventory) != 1 {
		t.Fatalf("inventory host=%d guest=%d, want guest pickup only", len(sim.players[hostID].Inventory), len(sim.players[guestID].Inventory))
	}

	target := &entity{kind: monsterEntity, pos: Vec2{X: 10, Y: 10}, hp: 5, maxHP: 5, monsterDefID: monsterDefID, aiMode: monsterAIModeChase}
	target.id = sim.alloc()
	sim.entities[target.id] = target
	sim.entities[hostID].pos = Vec2{X: 20, Y: 20}
	sim.entities[guestID].pos = Vec2{X: 10.4, Y: 10}
	hostHP := sim.entities[hostID].hp
	guestHP := sim.entities[guestID].hp
	sim.TickResults(nil)
	if sim.entities[hostID].hp != hostHP {
		t.Fatalf("host hp changed from %d to %d", hostHP, sim.entities[hostID].hp)
	}
	if sim.entities[guestID].hp >= guestHP {
		t.Fatalf("guest hp = %d, want below %d", sim.entities[guestID].hp, guestHP)
	}
}

func TestCoopExperienceEligibilityRules(t *testing.T) {
	cases := []struct {
		name        string
		setup       func(sim *Sim, hostID, guestID uint64, monster *entity)
		wantGuestXP int
	}{
		{
			name: "nearby alive connected same level",
			setup: func(sim *Sim, hostID, guestID uint64, monster *entity) {
				sim.entities[guestID].pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}
			},
			wantGuestXP: 10,
		},
		{
			name: "out of radius",
			setup: func(sim *Sim, hostID, guestID uint64, monster *entity) {
				sim.entities[guestID].pos = Vec2{X: monster.pos.X + 11, Y: monster.pos.Y}
			},
		},
		{
			name: "different level",
			setup: func(sim *Sim, hostID, guestID uint64, monster *entity) {
				guest := sim.entities[guestID]
				delete(sim.entities, guestID)
				otherLevel := newLevelState(-1, &sim.rules.Navigation)
				otherLevel.entities[guestID] = guest
				sim.levels[-1] = otherLevel
				sim.players[guestID].CurrentLevel = -1
				guest.pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}
			},
		},
		{
			name: "disconnected",
			setup: func(sim *Sim, hostID, guestID uint64, monster *entity) {
				sim.entities[guestID].pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}
				sim.RemovePlayerEntity(guestID)
			},
		},
		{
			name: "dead",
			setup: func(sim *Sim, hostID, guestID uint64, monster *entity) {
				sim.entities[guestID].pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}
				sim.entities[guestID].hp = 0
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sim, hostID, guestID, monster := newCoopRewardTestSim(t, "sess_coop_xp_"+strings.ReplaceAll(tc.name, " ", "_"))
			tc.setup(sim, hostID, guestID, monster)
			kill := killCoopRewardMonster(t, sim, hostID, monster)
			if sim.players[hostID].Progression.Experience != 10 {
				t.Fatalf("host xp = %d, want 10", sim.players[hostID].Progression.Experience)
			}
			if sim.players[guestID].Progression.Experience != tc.wantGuestXP {
				t.Fatalf("guest xp = %d, want %d; events=%+v changes=%+v", sim.players[guestID].Progression.Experience, tc.wantGuestXP, kill.Events, kill.Changes)
			}
		})
	}
}

func TestCoopExperienceRecipientLevelUpAndSoloUnchanged(t *testing.T) {
	sim, hostID, guestID, monster := newCoopRewardTestSim(t, "sess_coop_xp_level")
	sim.players[guestID].Progression.Experience = 10
	sim.entities[guestID].pos = Vec2{X: monster.pos.X + 4, Y: monster.pos.Y}

	kill := killCoopRewardMonster(t, sim, hostID, monster)
	if sim.players[guestID].Progression.Experience != 20 || sim.players[guestID].Progression.Level != 2 || sim.players[guestID].Progression.UnspentStatPoints != 3 {
		t.Fatalf("guest progression = %+v, want level 2 at 20 xp with 3 stat points", sim.players[guestID].Progression)
	}
	if !eventForEntity(kill, "character_leveled", guestID) {
		t.Fatalf("missing guest level-up event: %+v", kill.Events)
	}
	if owners := changeOwnersForOp(kill, OpCharacterProgressionUpdate); !sameUint64Set(owners, []uint64{hostID, guestID}) {
		t.Fatalf("progression update owners = %v, want host+guest", owners)
	}

	rules := coopRewardTestRules(t)
	solo := MustNewSim("sess_solo_xp", "solo_xp_seed", rules)
	soloMonster := findMonsterByDef(solo, monsterDefID)
	if soloMonster == nil {
		t.Fatal("solo missing monster")
	}
	soloMonster.hp = 1
	soloMonster.maxHP = 1
	soloMonster.lootTable = "no_drop"
	kill = killCoopRewardMonster(t, solo, solo.playerID, soloMonster)
	if solo.players[solo.playerID].Progression.Experience != 10 {
		t.Fatalf("solo xp = %d, want 10; events=%+v", solo.players[solo.playerID].Progression.Experience, kill.Events)
	}
}

func TestCoopPartyChallengeScaling(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_party_scaling", "party_scaling_seed", rules)
	levelNum := sim.currentLevel
	for count := 1; count <= 5; count++ {
		for len(sim.players) < count {
			if _, err := sim.AddGuestPlayer(fmt.Sprintf("acct_%d", len(sim.players)), fmt.Sprintf("char_%d", len(sim.players)), "Guest", rules.DefaultCharacterProgressionState()); err != nil {
				t.Fatalf("add guest: %v", err)
			}
		}
		got := sim.partyChallengeMultiplierForLevel(levelNum)
		want := rules.Combat.Coop.PartyChallenge.Multiplier(count)
		if math.Abs(got-want) > 1e-12 {
			t.Fatalf("count %d multiplier = %.15f, want %.15f", count, got, want)
		}
	}
}

func TestCoopPartyChallengeHPDamageAndLateJoin(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_party_hp_damage", "party_hp_damage_seed", rules)
	if _, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState()); err != nil {
		t.Fatalf("add guest: %v", err)
	}
	base := rules.Monsters["dungeon_mob"]
	var rarity MonsterRarityDef
	for _, candidate := range rules.DungeonGeneration.MonsterRarities {
		if candidate.HPMultiplier > 1 {
			rarity = candidate
			break
		}
	}
	if rarity.ID == "" {
		t.Fatal("missing monster rarity with hp multiplier > 1")
	}
	rarityScaledHP := roundPositive(float64(base.MaxHP) * rarity.HPMultiplier)
	monster := &entity{kind: monsterEntity, hp: rarityScaledHP, maxHP: rarityScaledHP, monsterDefID: "dungeon_mob"}
	sim.applyPartyHPScale(sim.activeLevel(), monster)
	wantHP := roundPositive(float64(rarityScaledHP) * rules.Combat.Coop.PartyChallenge.Multiplier(2))
	if monster.maxHP != wantHP || monster.hp != wantHP {
		t.Fatalf("party-scaled rarity hp = %d/%d, want %d", monster.hp, monster.maxHP, wantHP)
	}

	solo := MustNewSim("sess_party_late_join", "party_late_join_seed", rules)
	lateMonster := &entity{kind: monsterEntity, hp: 20, maxHP: 20, monsterDefID: "dungeon_mob"}
	solo.applyPartyHPScale(solo.activeLevel(), lateMonster)
	if lateMonster.maxHP != 20 {
		t.Fatalf("solo spawn hp = %d, want unscaled 20", lateMonster.maxHP)
	}
	if _, err := solo.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState()); err != nil {
		t.Fatalf("late add guest: %v", err)
	}
	if lateMonster.maxHP != 20 {
		t.Fatalf("late join retro-scaled hp to %d, want 20", lateMonster.maxHP)
	}
	gotDamage := solo.scaleMonsterDamageForParty(solo.currentLevel, DamageRange{Min: 4, Max: 4})
	if gotDamage != (DamageRange{Min: 5, Max: 5}) {
		t.Fatalf("two-player damage scale = %+v, want {5,5}", gotDamage)
	}
}

func TestCoopMonsterAttackDamageScalesAtResolution(t *testing.T) {
	rules := cloneRules(loadRules(t))
	damage := DamageRange{Min: 4, Max: 4}
	dummy := rules.Monsters[monsterDefID]
	dummy.AttackDamage = &damage
	dummy.AttackCooldown = 1
	dummy.HitChance = floatPtr(1)
	dummy.CritChance = floatPtr(0)
	dummy.RetaliationDamage = nil
	rules.Monsters[monsterDefID] = dummy

	sim := MustNewSim("sess_party_attack_damage", "party_attack_damage_seed", rules)
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	monster := findMonsterByDef(sim, monsterDefID)
	if monster == nil {
		t.Fatal("missing monster")
	}
	monster.pos = Vec2{X: 10, Y: 10}
	sim.entities[hostID].pos = Vec2{X: 10.4, Y: 10}
	sim.entities[guestID].pos = Vec2{X: 9, Y: 10}
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.advanceMonsterAttack(&res)
	ev := eventForTarget(res, "player_damaged", hostID)
	if ev == nil || ev.RawDamage == nil || *ev.RawDamage != 5 {
		t.Fatalf("player_damaged = %+v, events=%+v, want raw damage 5 from two-player scaling", ev, res.Events)
	}
}

func TestCoopDisconnectRemovalAndReconnectTownRespawn(t *testing.T) {
	sim, err := NewSimWithWorld("sess_coop_reconnect", "coop_reconnect_seed", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", sim.rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	if countPlayers(sim.SnapshotForPlayer(hostID).Entities) != 2 {
		t.Fatalf("host should initially see guest in town")
	}
	sim.RemovePlayerEntity(guestID)
	if sim.players[guestID].Connected {
		t.Fatal("guest still connected after removal")
	}
	if countPlayers(sim.SnapshotForPlayer(hostID).Entities) != 1 {
		t.Fatalf("host should not see disconnected guest")
	}
	if err := sim.RespawnPlayerInTown(guestID); err != nil {
		t.Fatalf("respawn guest: %v", err)
	}
	if !sim.players[guestID].Connected || sim.players[guestID].CurrentLevel != townLevel {
		t.Fatalf("guest reconnect state = %+v", sim.players[guestID])
	}
	if sim.levels[townLevel].entities[guestID] == nil {
		t.Fatal("guest entity missing after respawn")
	}
}

func countPlayers(entities []EntityView) int {
	count := 0
	for _, e := range entities {
		if e.Type == playerEntity {
			count++
		}
	}
	return count
}

func TestDungeonEquipmentLootDeterminism(t *testing.T) {
	first := dungeonEquipmentKillLootSequence(t, "v29_replay_equipment_16")
	second := dungeonEquipmentKillLootSequence(t, "v29_replay_equipment_16")
	if !sameStrings(first, second) {
		t.Fatalf("same-seed loot sequence drifted: %v != %v", first, second)
	}
	if len(first) == 0 {
		t.Fatal("expected at least one dungeon drop")
	}
}

func TestDungeonMonsterProactiveAttackGolden(t *testing.T) {
	var golden struct {
		SessionSeed                   string `json:"session_seed"`
		Level                         int    `json:"level"`
		MonsterDefID                  string `json:"monster_def_id"`
		AttackCooldownTicks           int    `json:"attack_cooldown_ticks"`
		MaxTicksForFirstPlayerDamaged int    `json:"max_ticks_for_first_player_damaged"`
		Damage                        int    `json:"damage"`
		PlayerHPAfter                 int    `json:"player_hp_after"`
	}
	loadGolden(t, "dungeon_monster_attack.json", &golden)
	rules := loadRules(t)
	def := rules.Monsters[golden.MonsterDefID]
	if def.AttackCooldown != golden.AttackCooldownTicks {
		t.Fatalf("%s attack cooldown = %d, want golden %d", golden.MonsterDefID, def.AttackCooldown, golden.AttackCooldownTicks)
	}
	sim, err := NewSimWithWorld("sess_dungeon_monster_attack", golden.SessionSeed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	results := descendFromCurrentLevel(t, sim, "descend")
	assertLevelChanged(t, results[0], 0, golden.Level)
	monster := firstEntityByKind(sim, monsterEntity)
	if monster == nil {
		t.Fatal("missing generated monster")
	}
	sim.entities[sim.playerID].pos = sim.travelArrivalPosition(sim.activeLevel(), monster.pos, sim.playerID)

	firstDamage, ok := waitForPlayerDamage(sim, golden.MaxTicksForFirstPlayerDamaged)
	if !ok {
		t.Fatalf("expected proactive player_damaged event; player=%+v monsters=%+v", sim.activeLevel().entities[sim.playerID].pos, dungeonMonsterDebugPositions(sim.activeLevel()))
	}
	if eventDamage(firstDamage, "player_damaged") != golden.Damage {
		t.Fatalf("first damage = %d, want %d", eventDamage(firstDamage, "player_damaged"), golden.Damage)
	}
	player := sim.activeLevel().entities[sim.playerID]
	if player.hp != golden.PlayerHPAfter {
		t.Fatalf("player hp = %d, want %d", player.hp, golden.PlayerHPAfter)
	}
	if len(liveDungeonMonsters(sim.activeLevel())) < rules.DungeonGeneration.MonsterPlacement.Count {
		t.Fatalf("live generated monster count below base population")
	}
}

func TestDungeonMonsterAttackCooldownAndDeterminism(t *testing.T) {
	var golden struct {
		SessionSeed                   string `json:"session_seed"`
		AttackCooldownTicks           int    `json:"attack_cooldown_ticks"`
		MaxTicksForFirstPlayerDamaged int    `json:"max_ticks_for_first_player_damaged"`
	}
	loadGolden(t, "dungeon_monster_attack.json", &golden)
	rules := loadRules(t)
	if rules.Monsters["dungeon_mob"].AttackCooldown != golden.AttackCooldownTicks {
		t.Fatalf("dungeon_mob attack cooldown = %d, want golden %d", rules.Monsters["dungeon_mob"].AttackCooldown, golden.AttackCooldownTicks)
	}
	sim, err := NewSimWithWorld("sess_dungeon_monster_cooldown", golden.SessionSeed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	monster := findMonsterByDef(sim, "dungeon_mob")
	if monster == nil {
		t.Fatal("missing generated dungeon_mob")
	}
	killOtherMonstersForTest(sim.activeLevel(), monster.id)
	sim.entities[sim.playerID].pos = sim.travelArrivalPosition(sim.activeLevel(), monster.pos, sim.playerID)
	firstDamage, ok := waitForPlayerDamage(sim, golden.MaxTicksForFirstPlayerDamaged)
	if !ok {
		t.Fatal("expected first proactive damage")
	}
	for i := 0; i < rules.Monsters["dungeon_mob"].AttackCooldown-1; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
			t.Fatalf("unexpected attack before cooldown on tick %d", res.Tick)
		}
	}
	second := sim.Tick(nil)
	if !hasEvent(second, "player_damaged") && !hasEvent(second, "player_killed") {
		t.Fatalf("missing attack after cooldown; first tick %d second result %+v", firstDamage.Tick, second)
	}

	replay, err := NewSimWithWorld("sess_dungeon_monster_cooldown_replay", golden.SessionSeed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new replay sim: %v", err)
	}
	descendFromCurrentLevel(t, replay, "descend")
	replayMonster := findMonsterByDef(replay, "dungeon_mob")
	if replayMonster == nil {
		t.Fatal("missing replay generated dungeon_mob")
	}
	killOtherMonstersForTest(replay.activeLevel(), replayMonster.id)
	replay.entities[replay.playerID].pos = replay.travelArrivalPosition(replay.activeLevel(), replayMonster.pos, replay.playerID)
	replayDamage, ok := waitForPlayerDamage(replay, golden.MaxTicksForFirstPlayerDamaged)
	if !ok {
		t.Fatal("expected replay proactive damage")
	}
	if replayDamage.Tick != firstDamage.Tick || eventDamage(replayDamage, "player_damaged") != eventDamage(firstDamage, "player_damaged") {
		t.Fatalf("replay first damage = tick %d damage %d, want tick %d damage %d", replayDamage.Tick, eventDamage(replayDamage, "player_damaged"), firstDamage.Tick, eventDamage(firstDamage, "player_damaged"))
	}
}

func TestRangedMonsterProjectileDamagesPlayer(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_ranged_monster_hit", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 4, Y: 5}
	player.hp = playerStartHP
	archer := addTestMonster(sim, "dungeon_archer", Vec2{X: 10, Y: 5}, rules.Monsters["dungeon_archer"].MaxHP)

	spawn := sim.Tick(nil)
	projectile := firstChangeEntityByType(spawn, projectileEntity)
	if projectile == nil {
		t.Fatalf("ranged monster did not spawn projectile: %+v", spawn)
	}
	if projectile.OwnerID != idStr(archer.id) || projectile.TargetID != idStr(player.id) || projectile.ProjectileDefID != rules.Monsters["dungeon_archer"].ProjectileDefID {
		t.Fatalf("projectile view = %+v, want archer owner/player target", projectile)
	}
	if distance(archer.pos, player.pos) <= rules.Combat.UnarmedReach+playerRadius {
		t.Fatalf("archer moved into melee contact at %+v vs player %+v", archer.pos, player.pos)
	}

	damage, ok := waitForPlayerDamage(sim, 20)
	if !ok {
		t.Fatalf("expected ranged monster player damage; player=%+v archer=%+v", player.pos, archer.pos)
	}
	if eventSource(damage, "player_damaged") != idStr(archer.id) && eventSource(damage, "player_killed") != idStr(archer.id) {
		t.Fatalf("ranged damage source = %+v, want archer %s", damage.Events, idStr(archer.id))
	}
	if player.hp >= playerStartHP {
		t.Fatalf("player hp = %d, want reduced by archer projectile", player.hp)
	}
}

func TestRangedMonsterBlockedShotDoesNotDamageThroughWall(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_ranged_monster_blocked", "01", rules, "inventory_lab")
	if err != nil {
		t.Fatal(err)
	}
	for id, candidate := range sim.activeLevel().entities {
		if candidate.kind == monsterEntity {
			delete(sim.activeLevel().entities, id)
		}
	}
	player := sim.entities[sim.playerID]
	player.pos = Vec2{X: 4, Y: 5}
	player.hp = playerStartHP
	archer := addTestMonster(sim, "dungeon_archer", Vec2{X: 10, Y: 5}, rules.Monsters["dungeon_archer"].MaxHP)
	archer.aiMode = monsterAIModeChase
	sim.activeLevel().walls = append(sim.activeLevel().walls, wallObstacle{pos: Vec2{X: 7, Y: 5}, size: Vec2{X: 1, Y: 4}, source: "generated"})

	res := TickResult{Tick: sim.tick, Level: sim.currentLevel}
	sim.advanceMonsterAttack(&res)
	if firstChangeEntityByType(res, projectileEntity) != nil {
		t.Fatalf("ranged monster fired through blocked line: %+v", res.Changes)
	}
	if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
		t.Fatalf("ranged monster damaged through blocked line: %+v", res.Events)
	}
	if player.hp != playerStartHP {
		t.Fatalf("player hp = %d, want unchanged", player.hp)
	}
}

func TestDungeonDescendAscendTransitions(t *testing.T) {
	var golden struct {
		Seed              string `json:"seed"`
		DescendThenAscend struct {
			ExpectedLevel int `json:"expected_level"`
		} `json:"descend_then_ascend"`
	}
	loadGolden(t, "dungeon_stairs.json", &golden)
	sim, err := NewSimWithWorld("sess_dungeon", golden.Seed, loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	if sim.currentLevel != townLevel {
		t.Fatalf("currentLevel = %d, want town", sim.currentLevel)
	}
	if _, ok := sim.levels[-1]; ok {
		t.Fatal("level -1 was generated before first descend")
	}
	townDown := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if townDown == nil || townDown.pos != (Vec2{X: 11, Y: 8}) {
		t.Fatalf("town down stair = %+v, want {11 8}", townDown)
	}

	results := descendFromCurrentLevel(t, sim, "descend_town")
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2: %+v", len(results), results)
	}
	if results[0].Level != townLevel || results[1].Level != -1 {
		t.Fatalf("town descend result levels = %d/%d, want 0/-1", results[0].Level, results[1].Level)
	}
	if !hasEntityRemove(results[0], idStr(sim.playerID)) {
		t.Fatalf("from-level result missing player remove: %+v", results[0].Changes)
	}
	if !hasEntitySpawn(results[1], idStr(sim.playerID)) {
		t.Fatalf("to-level result missing player spawn: %+v", results[1].Changes)
	}
	assertLevelChanged(t, results[0], townLevel, -1)

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if down == nil {
		t.Fatal("missing down stairs on level -1")
	}
	sim.entities[sim.playerID].pos = down.pos
	results = sim.TickResults([]Input{{MessageID: "descend_2", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("descend to -2 results = %d, want 2: %+v", len(results), results)
	}
	if results[0].Level != -1 || results[1].Level != -2 {
		t.Fatalf("descend result levels = %d/%d, want -1/-2", results[0].Level, results[1].Level)
	}
	assertLevelChanged(t, results[0], -1, -2)

	up := sim.findStair(sim.activeLevel(), stairsUpDefID)
	if up == nil {
		t.Fatal("missing up stairs on level -2")
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, up.pos)
	if coin := findLootByDef(sim, "gold"); coin != nil {
		t.Fatalf("found generated dungeon gold coin %+v; gold should drop from monsters, chests, or other loot sources", coin)
	}

	sim.entities[sim.playerID].pos = up.pos
	results = sim.TickResults([]Input{{MessageID: "ascend", Type: "ascend_intent", Ascend: &AscendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("ascend results = %d, want 2: %+v", len(results), results)
	}
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, down.pos)
	assertLevelChanged(t, results[0], -2, -1)

	up = sim.findStair(sim.activeLevel(), stairsUpDefID)
	if up == nil {
		t.Fatal("missing up stairs on level -1")
	}
	sim.entities[sim.playerID].pos = up.pos
	results = sim.TickResults([]Input{{MessageID: "ascend_town", Type: "ascend_intent", Ascend: &AscendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("ascend to town results = %d, want 2: %+v", len(results), results)
	}
	if sim.currentLevel != golden.DescendThenAscend.ExpectedLevel {
		t.Fatalf("currentLevel after town ascend = %d, want %d", sim.currentLevel, golden.DescendThenAscend.ExpectedLevel)
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, townDown.pos)
	assertLevelChanged(t, results[0], -1, townLevel)
}

func TestDungeonTownBootstrapAndDeepDescent(t *testing.T) {
	sim, err := NewSimWithWorld("sess_dungeon_town", "deadbeefdeadbeef", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	if sim.currentLevel != townLevel {
		t.Fatalf("currentLevel = %d, want town", sim.currentLevel)
	}
	if len(sim.levels) != 1 {
		t.Fatalf("levels at start = %v, want only town", sim.levels)
	}
	if !sim.discoveredTeleporters[townLevel] {
		t.Fatal("town teleporter should start discovered")
	}
	if sim.findStair(sim.activeLevel(), stairsDownDefID) == nil {
		t.Fatal("town missing down stair")
	}
	if sim.findTeleporter(sim.activeLevel()) == nil {
		t.Fatal("town missing teleporter")
	}

	for want := -1; want >= -3; want-- {
		results := descendFromCurrentLevel(t, sim, fmt.Sprintf("descend_%d", -want))
		if len(results) != 2 {
			t.Fatalf("descend to %d results = %d, want 2: %+v", want, len(results), results)
		}
		if sim.currentLevel != want {
			t.Fatalf("currentLevel after descend = %d, want %d", sim.currentLevel, want)
		}
		if _, ok := sim.levels[want]; !ok {
			t.Fatalf("level %d not generated", want)
		}
		tp := sim.findTeleporter(sim.activeLevel())
		if want == -3 && tp == nil {
			t.Fatalf("level %d missing cadence teleporter", want)
		}
		if want != -3 && tp != nil {
			t.Fatalf("level %d has unexpected teleporter %+v", want, tp)
		}
	}
}

func TestDungeonTownNavigationCoversVisibleHubFloor(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_dungeon_town_nav", "deadbeefdeadbeef", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	nav := sim.activeNav()
	if nav.GridBounds.MaxY <= rules.Navigation.GridBounds.MaxY {
		t.Fatalf("town nav max_y = %d, want expanded beyond global %d", nav.GridBounds.MaxY, rules.Navigation.GridBounds.MaxY)
	}
	target := Vec2{X: 4, Y: 16}
	if !sim.positionInNavigationBounds(nav, target) {
		t.Fatalf("town floor target %+v is outside nav bounds %+v", target, nav.GridBounds)
	}
	result := sim.Tick([]Input{{
		MessageID: "walk_town_floor",
		Type:      "move_to_intent",
		MoveTo:    &MoveToIntent{Position: target},
	}})
	assertAck(t, result, "walk_town_floor")
	if sim.activeLevel().autoNav == nil {
		t.Fatal("town floor move did not queue auto-navigation")
	}
}

func TestDungeonTeleporterDiscoveryAndTravel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_dungeon_tp", "deadbeefdeadbeef", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	townTeleporter := sim.findTeleporter(sim.activeLevel())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	for level := -1; level >= -2; level-- {
		descendFromCurrentLevel(t, sim, fmt.Sprintf("descend_to_%d", level))
		if tp := sim.findTeleporter(sim.activeLevel()); tp != nil {
			t.Fatalf("level %d has unexpected teleporter %+v", level, tp)
		}
	}
	descendFromCurrentLevel(t, sim, "descend_to_3")
	level3Teleporter := sim.findTeleporter(sim.activeLevel())
	if level3Teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	sim.entities[sim.playerID].pos = level3Teleporter.pos

	discover3 := sim.Tick([]Input{{MessageID: "discover_3", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level3Teleporter.id)}}})
	assertAck(t, discover3, "discover_3")
	if !hasTeleporterDiscoveryUpdate(discover3, -3) || !hasTeleporterDiscoveredEvent(discover3, -3) {
		t.Fatalf("missing level -3 discovery result: changes=%+v events=%+v", discover3.Changes, discover3.Events)
	}
	if !sim.discoveredTeleporters[-3] {
		t.Fatal("level -3 teleporter not marked discovered")
	}

	results := sim.TickResults([]Input{{MessageID: "tp_to_town", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: townLevel}}})
	if len(results) != 2 {
		t.Fatalf("teleport to town results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], -3, townLevel)
	if sim.currentLevel != townLevel {
		t.Fatalf("currentLevel = %d, want town", sim.currentLevel)
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, townTeleporter.pos)

	results = sim.TickResults([]Input{{MessageID: "tp_to_3", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -3}}})
	if len(results) != 2 {
		t.Fatalf("teleport results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], townLevel, -3)
	if sim.currentLevel != -3 {
		t.Fatalf("currentLevel = %d, want -3", sim.currentLevel)
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, level3Teleporter.pos)
}

func TestLoadDiscoveredTeleportersAllowsFreshSessionWaypointTravel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_loaded_waypoint", "feedfacefeedface", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	sim.LoadDiscoveredTeleporters([]int{-1, -3})
	if _, ok := sim.levels[-3]; ok {
		t.Fatal("level -3 should not be generated by loading waypoint discovery")
	}

	view := sim.Snapshot().DiscoveredTeleporters
	assertTeleporterDiscoveryView(t, view, []struct {
		Level      int  `json:"level"`
		Discovered bool `json:"discovered"`
	}{{Level: -3, Discovered: true}, {Level: 0, Discovered: true}})

	townTeleporter := sim.findTeleporter(sim.activeLevel())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	sim.entities[sim.playerID].pos = townTeleporter.pos
	results := sim.TickResults([]Input{{MessageID: "tp_loaded_waypoint", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -3}}})
	if len(results) != 2 {
		t.Fatalf("teleport results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], townLevel, -3)
	if sim.currentLevel != -3 {
		t.Fatalf("currentLevel = %d, want -3", sim.currentLevel)
	}
	if _, ok := sim.levels[-3]; !ok {
		t.Fatal("teleport to loaded waypoint did not generate level -3")
	}
}

func TestTeleportRejectsUndiscoveredTargetLevel(t *testing.T) {
	var golden struct {
		Seed                      string `json:"seed"`
		VisitedUndiscoveredTarget struct {
			RejectReason          string `json:"reject_reason"`
			DiscoveredTeleporters []struct {
				Level      int  `json:"level"`
				Discovered bool `json:"discovered"`
			} `json:"discovered_teleporters"`
		} `json:"visited_undiscovered_target"`
	}
	loadGolden(t, "dungeon_teleporters.json", &golden)

	sim, err := NewSimWithWorld("sess_dungeon_tp_reject", golden.Seed, loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	for level := -1; level >= -3; level-- {
		descendFromCurrentLevel(t, sim, fmt.Sprintf("descend_to_%d", level))
	}
	level3Teleporter := sim.findTeleporter(sim.activeLevel())
	if level3Teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	sim.entities[sim.playerID].pos = level3Teleporter.pos
	discover3 := sim.Tick([]Input{{MessageID: "discover_3", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level3Teleporter.id)}}})
	assertAck(t, discover3, "discover_3")

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	sim.entities[sim.playerID].pos = down.pos
	results := sim.TickResults([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2", len(results))
	}
	if hasTeleporterDiscoveryUpdateWith(results[1], -4, false) {
		t.Fatalf("descend arrival unexpectedly advertised non-teleporter level -4: %+v", results[1].Changes)
	}

	up := sim.findStair(sim.activeLevel(), stairsUpDefID)
	if up == nil {
		t.Fatal("missing up stairs on level -2")
	}
	sim.entities[sim.playerID].pos = up.pos
	results = sim.TickResults([]Input{{MessageID: "ascend", Type: "ascend_intent", Ascend: &AscendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("ascend results = %d, want 2", len(results))
	}

	level3Teleporter = sim.findTeleporter(sim.activeLevel())
	sim.entities[sim.playerID].pos = level3Teleporter.pos
	assertTeleporterDiscoveryView(t, sim.teleporterDiscoveryView(), golden.VisitedUndiscoveredTarget.DiscoveredTeleporters)

	reject := sim.Tick([]Input{{MessageID: "tp_undiscovered_target", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -4}}})
	assertReject(t, reject, "tp_undiscovered_target", golden.VisitedUndiscoveredTarget.RejectReason)
}

func TestDungeonTeleportersGolden(t *testing.T) {
	var golden struct {
		Seed                    string `json:"seed"`
		WorldID                 string `json:"world_id"`
		DiscoverDescendTeleport struct {
			ExpectedLevel          int  `json:"expected_level"`
			ExpectedPlayerPosition Vec2 `json:"expected_player_position"`
			DiscoveredTeleporters  []struct {
				Level      int  `json:"level"`
				Discovered bool `json:"discovered"`
			} `json:"discovered_teleporters"`
		} `json:"discover_descend_teleport"`
	}
	loadGolden(t, "dungeon_teleporters.json", &golden)

	sim, err := NewSimWithWorld("sess_dungeon_tp_golden", golden.Seed, loadRules(t), golden.WorldID)
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	for level := -1; level >= -3; level-- {
		descendFromCurrentLevel(t, sim, fmt.Sprintf("descend_to_%d", level))
	}
	level3Teleporter := sim.findTeleporter(sim.activeLevel())
	if level3Teleporter == nil {
		t.Fatal("missing level -3 teleporter")
	}
	sim.entities[sim.playerID].pos = level3Teleporter.pos
	assertAck(t, sim.Tick([]Input{{MessageID: "discover_3", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level3Teleporter.id)}}}), "discover_3")

	sim.TickResults([]Input{{MessageID: "tp_to_town", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: townLevel}}})
	townTeleporter := sim.findTeleporter(sim.activeLevel())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	sim.entities[sim.playerID].pos = townTeleporter.pos
	sim.TickResults([]Input{{MessageID: "tp_to_3", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -3}}})

	want := golden.DiscoverDescendTeleport
	if sim.currentLevel != want.ExpectedLevel {
		t.Fatalf("currentLevel = %d, want %d", sim.currentLevel, want.ExpectedLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != want.ExpectedPlayerPosition {
		t.Fatalf("player position = %+v, want %+v", got, want.ExpectedPlayerPosition)
	}
	assertTravelArrivalNextToMarker(t, sim, sim.activeLevel(), sim.entities[sim.playerID].pos, level3Teleporter.pos)
	assertTeleporterDiscoveryView(t, sim.teleporterDiscoveryView(), want.DiscoveredTeleporters)
}

func assertTeleporterDiscoveryView(t *testing.T, got []TeleporterDiscoveryView, want []struct {
	Level      int  `json:"level"`
	Discovered bool `json:"discovered"`
}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("discovery view len = %d, want %d: got=%+v", len(got), len(want), got)
	}
	for i, row := range want {
		if got[i].Level != row.Level || got[i].Discovered != row.Discovered {
			t.Fatalf("discovery[%d] = %+v, want level=%d discovered=%v", i, got[i], row.Level, row.Discovered)
		}
	}
}

func descendFromCurrentLevel(t *testing.T, sim *Sim, messageID string) []TickResult {
	t.Helper()
	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if down == nil {
		t.Fatalf("missing down stairs on level %d", sim.currentLevel)
	}
	sim.entities[sim.playerID].pos = down.pos
	return sim.TickResults([]Input{{MessageID: messageID, Type: "descend_intent", Descend: &DescendIntent{}}})
}

func assertTravelArrivalNextToMarker(t *testing.T, sim *Sim, level *LevelState, got, marker Vec2) {
	t.Helper()
	if got == marker {
		t.Fatalf("arrival position = marker %+v, want adjacent free tile", marker)
	}
	nav := sim.navigationForLevel(level)
	step := nav.CellSize
	if step <= 0 {
		step = 1.0
	}
	if distance(got, marker) > math.Sqrt2*step+1e-9 {
		t.Fatalf("arrival position = %+v, want adjacent to marker %+v", got, marker)
	}
	if sim.travelArrivalBlocked(level, got, sim.playerID) {
		t.Fatalf("arrival position = %+v, want unblocked free tile near %+v", got, marker)
	}
}

func generatedStairPos(level generatedDungeonLevel, defID string) Vec2 {
	for _, stair := range level.stairs {
		if stair.defID == defID {
			return stair.pos
		}
	}
	return Vec2{}
}

func generatedTeleporterPos(level generatedDungeonLevel) Vec2 {
	for _, teleporter := range level.teleporters {
		if teleporter.defID == teleporterDefID {
			return teleporter.pos
		}
	}
	return Vec2{}
}

func liveDungeonMonsters(level *LevelState) []*entity {
	out := []*entity{}
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate != nil && candidate.kind == monsterEntity && !candidate.isBoss && candidate.hp > 0 {
			out = append(out, candidate)
		}
	}
	return out
}

func killOtherMonstersForTest(level *LevelState, keepID uint64) {
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate != nil && candidate.kind == monsterEntity && candidate.id != keepID {
			candidate.hp = 0
		}
	}
}

func findBossEntity(t *testing.T, level *LevelState) *entity {
	t.Helper()
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate != nil && candidate.kind == monsterEntity && candidate.isBoss {
			return candidate
		}
	}
	t.Fatal("missing boss entity")
	return nil
}

func placeDefaultPlayerOnLevel(t *testing.T, sim *Sim, level *LevelState, pos Vec2) {
	t.Helper()
	playerID := sim.playerID
	player := (*entity)(nil)
	for _, existing := range sim.levels {
		if existing == nil {
			continue
		}
		if candidate := existing.entities[playerID]; candidate != nil {
			player = candidate
			delete(existing.entities, playerID)
		}
	}
	if player == nil {
		t.Fatalf("missing default player %d", playerID)
	}
	player.pos = pos
	level.entities[playerID] = player
	if ps := sim.players[playerID]; ps != nil {
		ps.CurrentLevel = level.levelNum
	}
	sim.currentLevel = level.levelNum
	sim.usePlayer(sim.players[playerID])
}

func dungeonEquipmentKillLootSequence(t *testing.T, seed string) []string {
	t.Helper()
	sim, err := NewSimWithWorld("sess_"+seed, seed, loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend_1")
	descendFromCurrentLevel(t, sim, "descend_2")
	descendFromCurrentLevel(t, sim, "descend_3")
	if sim.currentLevel != -3 {
		t.Fatalf("currentLevel = %d, want -3", sim.currentLevel)
	}
	var monster *entity
	for _, id := range sortedEntityIDs(sim.activeLevel().entities) {
		candidate := sim.activeLevel().entities[id]
		if candidate.kind == monsterEntity && candidate.monsterDefID == "dungeon_mob" {
			monster = candidate
			break
		}
	}
	if monster == nil {
		t.Fatal("missing depth-3 dungeon mob")
	}
	player := sim.activeLevel().entities[sim.playerID]
	player.pos = monster.pos
	monster.hp = 1
	res := sim.Tick([]Input{{
		MessageID:     "kill_depth3",
		CorrelationID: "corr_depth3",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(monster.id)},
	}})
	assertAck(t, res, "kill_depth3")
	if !hasEvent(res, "monster_killed") {
		t.Fatalf("missing monster_killed: %+v", res.Events)
	}
	out := []string{}
	for _, change := range res.Changes {
		if change.Op != OpEntitySpawn || change.Entity == nil || change.Entity.Type != lootEntity {
			continue
		}
		out = append(out, change.Entity.ItemTemplateID+":"+change.Entity.ItemDefID)
	}
	return out
}

func levelWithoutMonsterIndex(level generatedDungeonLevel, skip int) generatedDungeonLevel {
	out := level
	out.monsters = make([]generatedMonster, 0, len(level.monsters)-1)
	for i, monster := range level.monsters {
		if i == skip {
			continue
		}
		out.monsters = append(out.monsters, monster)
	}
	return out
}

func waitForPlayerDamage(sim *Sim, maxTicks int) (TickResult, bool) {
	for i := 0; i < maxTicks; i++ {
		res := sim.Tick(nil)
		if hasEvent(res, "player_damaged") || hasEvent(res, "player_killed") {
			return res, true
		}
	}
	return TickResult{}, false
}

func eventDamage(r TickResult, eventType string) int {
	for _, event := range r.Events {
		if event.EventType == eventType && event.Damage != nil {
			return *event.Damage
		}
	}
	return 0
}

func eventSource(r TickResult, eventType string) string {
	for _, event := range r.Events {
		if event.EventType == eventType {
			return event.SourceEntityID
		}
	}
	return ""
}

func countLiveMonstersByDef(level *LevelState, defID string) int {
	count := 0
	for _, entity := range level.entities {
		if entity.kind == monsterEntity && entity.hp > 0 && entity.monsterDefID == defID {
			count++
		}
	}
	return count
}

func countEntitiesByKind(level *LevelState, kind string) int {
	count := 0
	for _, entity := range level.entities {
		if entity.kind == kind {
			count++
		}
	}
	return count
}

func countLootByItemDef(level *LevelState, itemDefID string) int {
	count := 0
	for _, entity := range level.entities {
		if entity.kind == lootEntity && entity.itemDefID == itemDefID {
			count++
		}
	}
	return count
}

func dungeonMonsterDebugPositions(level *LevelState) []Vec2 {
	positions := []Vec2{}
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity.kind == monsterEntity && !entity.isBoss {
			positions = append(positions, entity.pos)
		}
	}
	return positions
}

func hasEntityRemove(r TickResult, entityID string) bool {
	for _, ch := range r.Changes {
		if ch.Op == OpEntityRemove && ch.EntityID == entityID {
			return true
		}
	}
	return false
}

func hasEntitySpawn(r TickResult, entityID string) bool {
	for _, ch := range r.Changes {
		if ch.Op == OpEntitySpawn && ch.Entity != nil && ch.Entity.ID == entityID {
			return true
		}
	}
	return false
}

func hasTeleporterDiscoveryUpdate(r TickResult, level int) bool {
	return hasTeleporterDiscoveryUpdateWith(r, level, true)
}

func hasTeleporterDiscoveryUpdateWith(r TickResult, level int, discovered bool) bool {
	for _, ch := range r.Changes {
		if ch.Op == OpTeleporterDiscoveryUpdate && ch.Level == level && ch.Discovered == discovered {
			return true
		}
	}
	return false
}

func hasTeleporterDiscoveredEvent(r TickResult, level int) bool {
	for _, ev := range r.Events {
		if ev.EventType == "teleporter_discovered" && ev.Level != nil && *ev.Level == level {
			return true
		}
	}
	return false
}

func assertLevelChanged(t *testing.T, r TickResult, fromLevel, toLevel int) {
	t.Helper()
	for _, ev := range r.Events {
		if ev.EventType != "level_changed" {
			continue
		}
		if ev.FromLevel == nil || ev.ToLevel == nil || *ev.FromLevel != fromLevel || *ev.ToLevel != toLevel {
			t.Fatalf("level_changed = %+v, want %d -> %d", ev, fromLevel, toLevel)
		}
		return
	}
	t.Fatalf("missing level_changed in %+v", r.Events)
}

func hasEvent(r TickResult, eventType string) bool {
	for _, ev := range r.Events {
		if ev.EventType == eventType {
			return true
		}
	}
	return false
}

func countEvents(r TickResult, eventType string) int {
	count := 0
	for _, ev := range r.Events {
		if ev.EventType == eventType {
			count++
		}
	}
	return count
}

func eventListHas(events []Event, eventType string) bool {
	for _, ev := range events {
		if ev.EventType == eventType {
			return true
		}
	}
	return false
}

func hasBossPatternStart(r TickResult, patternID string) bool {
	for _, ev := range r.Events {
		if ev.EventType == "boss_phase_started" && ev.PatternID == patternID && ev.PhaseKind == "telegraph" {
			return true
		}
	}
	return false
}

func waitForBossPatternStart(t *testing.T, sim *Sim, patternID string, maxTicks int) Event {
	t.Helper()
	for i := 0; i < maxTicks; i++ {
		res := sim.Tick(nil)
		for _, ev := range res.Events {
			if ev.EventType == "boss_phase_started" && ev.PatternID == patternID && ev.PhaseKind == "telegraph" {
				return ev
			}
		}
	}
	t.Fatalf("missing boss_phase_started for pattern %s within %d ticks", patternID, maxTicks)
	return Event{}
}

func hasProgressionChange(r TickResult) bool {
	for _, change := range r.Changes {
		if change.Op == OpCharacterProgressionUpdate && change.Progression != nil {
			return true
		}
	}
	return false
}

func changeOwnersForOp(r TickResult, op string) []uint64 {
	var out []uint64
	for _, change := range r.Changes {
		if change.Op == op {
			out = append(out, change.OwnerPlayerID)
		}
	}
	return out
}

func sameUint64Set(got, want []uint64) bool {
	if len(got) != len(want) {
		return false
	}
	counts := make(map[uint64]int, len(got))
	for _, v := range got {
		counts[v]++
	}
	for _, v := range want {
		counts[v]--
		if counts[v] < 0 {
			return false
		}
	}
	return true
}

func eventForEntity(r TickResult, eventType string, entityID uint64) bool {
	for _, ev := range r.Events {
		if ev.EventType == eventType && ev.EntityID == idStr(entityID) {
			return true
		}
	}
	return false
}

func eventForTarget(r TickResult, eventType string, targetID uint64) *Event {
	for i := range r.Events {
		ev := &r.Events[i]
		if ev.EventType == eventType && ev.TargetEntityID == idStr(targetID) {
			return ev
		}
	}
	return nil
}

func coopRewardTestRules(t *testing.T) *Rules {
	t.Helper()
	rules := cloneRules(loadRules(t))
	forceCharacterHitChance(rules, 1.0)
	dummy := rules.Monsters[monsterDefID]
	dummy.MaxHP = 1
	dummy.XPReward = 10
	dummy.RetaliationDamage = nil
	rules.Monsters[monsterDefID] = dummy
	return rules
}

func forceCharacterHitChance(rules *Rules, chance float64) {
	hit := rules.CharacterProgression.DerivedStats["hit_chance"]
	hit.Base = chance
	hit.PerDex = 0
	hit.PerStr = 0
	hit.PerVit = 0
	hit.PerMagic = 0
	hit.Min = &chance
	hit.Max = &chance
	rules.CharacterProgression.DerivedStats["hit_chance"] = hit
}

func forceMonsterHitChance(rules *Rules, monsterDefID string, chance float64) {
	def := rules.Monsters[monsterDefID]
	def.HitChance = &chance
	rules.Monsters[monsterDefID] = def
}

func newCoopRewardTestSim(t *testing.T, sessionID string) (*Sim, uint64, uint64, *entity) {
	t.Helper()
	rules := coopRewardTestRules(t)
	sim, err := NewSimWithWorld(sessionID, "coop_reward_seed", rules, "vertical_slice")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	hostID := sim.playerID
	guestID, err := sim.AddGuestPlayer("acct_guest", "char_guest", "Guest", rules.DefaultCharacterProgressionState())
	if err != nil {
		t.Fatalf("add guest: %v", err)
	}
	monster := findMonsterByDef(sim, monsterDefID)
	if monster == nil {
		t.Fatal("missing monster")
	}
	monster.hp = 1
	monster.maxHP = 1
	monster.lootTable = "no_drop"
	return sim, hostID, guestID, monster
}

func killCoopRewardMonster(t *testing.T, sim *Sim, hostID uint64, monster *entity) TickResult {
	t.Helper()
	sim.entities[hostID].pos = monster.pos
	kill := sim.Tick([]Input{{MessageID: "host_kill", ActorPlayerID: hostID, Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, kill, "host_kill")
	if !hasEvent(kill, "monster_killed") || !hasEvent(kill, "experience_gained") {
		t.Fatalf("kill events = %+v", kill.Events)
	}
	return kill
}

func assertDerivedStats(t *testing.T, got, want DerivedStatsView) {
	t.Helper()
	assertFloat := func(name string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 0.000001 {
			t.Fatalf("%s = %v, want %v; got=%+v want=%+v", name, got, want, got, want)
		}
	}
	assertFloat("damage_min", got.DamageMin, want.DamageMin)
	assertFloat("damage_max", got.DamageMax, want.DamageMax)
	assertFloat("armor", got.Armor, want.Armor)
	assertFloat("attack_speed", got.AttackSpeed, want.AttackSpeed)
	if got.AttackIntervalTicks != want.AttackIntervalTicks {
		t.Fatalf("attack_interval_ticks = %d, want %d; got=%+v want=%+v", got.AttackIntervalTicks, want.AttackIntervalTicks, got, want)
	}
	assertFloat("hit_chance", got.HitChance, want.HitChance)
	assertFloat("crit_chance", got.CritChance, want.CritChance)
	assertFloat("crit_damage", got.CritDamage, want.CritDamage)
	assertFloat("movement_speed", got.MovementSpeed, want.MovementSpeed)
	assertFloat("max_hp", got.MaxHP, want.MaxHP)
	assertFloat("max_mana", got.MaxMana, want.MaxMana)
	assertFloat("health_regen_per_second", got.HealthRegenPerSecond, want.HealthRegenPerSecond)
	assertFloat("mana_regen_per_second", got.ManaRegenPerSecond, want.ManaRegenPerSecond)
}

func findStatBreakdown(rows []StatBreakdownView, key string) *StatBreakdownView {
	for i := range rows {
		if rows[i].Key == key {
			return &rows[i]
		}
	}
	return nil
}

func statBreakdownHasSourceKind(row StatBreakdownView, kind string) bool {
	for _, source := range row.Sources {
		if source.Kind == kind {
			return true
		}
	}
	return false
}

func characterProgressionUpdate(r TickResult) *CharacterProgressionView {
	for i := range r.Changes {
		if r.Changes[i].Op == OpCharacterProgressionUpdate {
			return r.Changes[i].Progression
		}
	}
	return nil
}

func skillProgressionUpdate(r TickResult) *SkillProgressionView {
	for i := range r.Changes {
		if r.Changes[i].Op == OpSkillProgressionUpdate {
			return r.Changes[i].SkillProgression
		}
	}
	return nil
}

func skillProgressionRow(view SkillProgressionView, skillID string) (SkillProgressionSkillView, bool) {
	for _, row := range view.Skills {
		if row.SkillID == skillID {
			return row, true
		}
	}
	return SkillProgressionSkillView{}, false
}

func skillCooldownUpdate(r TickResult) []SkillCooldownView {
	for i := range r.Changes {
		if r.Changes[i].Op == OpSkillCooldownUpdate {
			return r.Changes[i].SkillCooldowns
		}
	}
	return nil
}

func hasBreakdownSource(rows []StatBreakdownSourceView, kind string) bool {
	for _, row := range rows {
		if row.Kind == kind {
			return true
		}
	}
	return false
}

func assertEventHeal(t *testing.T, r TickResult, eventType string, want int) {
	t.Helper()
	for _, ev := range r.Events {
		if ev.EventType != eventType {
			continue
		}
		if ev.Heal == nil || *ev.Heal != want {
			t.Fatalf("%s heal = %v, want %d in events %+v", eventType, ev.Heal, want, r.Events)
		}
		return
	}
	t.Fatalf("missing event %s in %+v", eventType, r.Events)
}

func assertEventMana(t *testing.T, r TickResult, eventType string, want int) {
	t.Helper()
	for _, ev := range r.Events {
		if ev.EventType != eventType {
			continue
		}
		if ev.Mana == nil || *ev.Mana != want {
			t.Fatalf("%s mana = %v, want %d in events %+v", eventType, ev.Mana, want, r.Events)
		}
		return
	}
	t.Fatalf("missing event %s in %+v", eventType, r.Events)
}

func assertEventDamage(t *testing.T, r TickResult, eventType string, want int) {
	t.Helper()
	for _, ev := range r.Events {
		if ev.EventType != eventType {
			continue
		}
		if ev.Damage == nil || *ev.Damage != want {
			t.Fatalf("%s damage = %v, want %d in events %+v", eventType, ev.Damage, want, r.Events)
		}
		return
	}
	t.Fatalf("missing event %s in %+v", eventType, r.Events)
}

func assertEventDamageAtLeast(t *testing.T, r TickResult, eventType string, min int) {
	t.Helper()
	for _, ev := range r.Events {
		if ev.EventType != eventType {
			continue
		}
		if ev.Damage == nil || *ev.Damage < min {
			t.Fatalf("%s damage = %v, want >= %d in events %+v", eventType, ev.Damage, min, r.Events)
		}
		return
	}
	t.Fatalf("missing event %s in %+v", eventType, r.Events)
}

func sameIntMap(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for key, av := range a {
		if b[key] != av {
			return false
		}
	}
	return true
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func hasLootSpawn(r TickResult, itemDefID string) bool {
	for _, c := range r.Changes {
		if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == lootEntity && c.Entity.ItemDefID == itemDefID {
			return true
		}
	}
	return false
}

func lootSpawnPosition(r TickResult, itemDefID string) (Vec2, bool) {
	for _, c := range r.Changes {
		if c.Op == OpEntitySpawn && c.Entity != nil && c.Entity.Type == lootEntity && c.Entity.ItemDefID == itemDefID {
			return c.Entity.Position, true
		}
	}
	return Vec2{}, false
}

func adjacentUnitOffsets() []Vec2 {
	return []Vec2{
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 0, Y: -1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: 1, Y: -1},
	}
}

func stringSliceContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
