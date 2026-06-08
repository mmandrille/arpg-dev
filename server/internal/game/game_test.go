package game

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

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

// --- rules ------------------------------------------------------------------

func TestLoadRules(t *testing.T) {
	r := loadRules(t)
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
	if r.Items["training_badge"].Equippable || r.Items["training_badge"].Slot != "" {
		t.Fatalf("training_badge def = %+v, want non-equippable without slot", r.Items["training_badge"])
	}
	if r.Items["training_badge"].Damage != nil {
		t.Fatalf("training_badge damage = %+v, want nil", r.Items["training_badge"].Damage)
	}
	potion := r.Items["red_potion"]
	if potion.Category != "consumable" || potion.Heal == nil || potion.Heal.Min != 5 || potion.Heal.Max != 5 {
		t.Fatalf("red_potion def = %+v, want consumable heal {5,5}", potion)
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
	if r.Interactables["wooden_door"].InitialState != interactableClosed {
		t.Fatalf("wooden_door = %+v, want initially closed", r.Interactables["wooden_door"])
	}
	if r.CharacterProgression.PointsPerLevel != 5 || r.CharacterProgression.LevelCap != 20 {
		t.Fatalf("character progression = %+v, want points_per_level 5 level_cap 20", r.CharacterProgression)
	}
	if r.Monsters["dungeon_mob"].XPReward <= 0 {
		t.Fatalf("dungeon_mob xp_reward = %d, want positive", r.Monsters["dungeon_mob"].XPReward)
	}
}

func TestMonsterRarityGolden(t *testing.T) {
	var golden struct {
		MonsterDefID string `json:"monster_def_id"`
		Rarities     []struct {
			ID               string  `json:"id"`
			Weight           int     `json:"weight"`
			Color            string  `json:"color"`
			HPMultiplier     float64 `json:"hp_multiplier"`
			DamageMultiplier float64 `json:"damage_multiplier"`
			XPMultiplier     float64 `json:"xp_multiplier"`
			LootDepthOffset  int     `json:"loot_depth_offset"`
			Expected         struct {
				MaxHP        int         `json:"max_hp"`
				AttackDamage DamageRange `json:"attack_damage"`
				XPReward     int         `json:"xp_reward"`
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
		if roundPositive(float64(base.MaxHP)*rarity.HPMultiplier) != c.Expected.MaxHP {
			t.Fatalf("rarity %s max hp mismatch", c.ID)
		}
		if got := scaleDamageRange(*base.AttackDamage, rarity.DamageMultiplier); got != c.Expected.AttackDamage {
			t.Fatalf("rarity %s damage = %+v, want %+v", c.ID, got, c.Expected.AttackDamage)
		}
		if roundPositive(float64(base.XPReward)*rarity.XPMultiplier) != c.Expected.XPReward {
			t.Fatalf("rarity %s xp mismatch", c.ID)
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
				Level             int              `json:"level"`
				CurrentLevelXP    int              `json:"current_level_xp"`
				NextLevelXP       int              `json:"next_level_xp"`
				UnspentStatPoints int              `json:"unspent_stat_points"`
				BaseStats         BaseStatsView    `json:"base_stats"`
				DerivedStats      DerivedStatsView `json:"derived_stats"`
			} `json:"expected"`
		} `json:"cases"`
	}
	loadGolden(t, "character_progression.json", &golden)

	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			sim := NewSim("sess_progression_"+tc.Name, "01", rules)
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
	sim := NewSim("sess_xp_kill", "01", rules)
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
	if view.Experience != 20 || view.Level != 2 || view.UnspentStatPoints != 5 {
		t.Fatalf("progression after kill = %+v, want exp 20 level 2 unspent 5", view)
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

func TestStatAllocationVitHPAndRejects(t *testing.T) {
	sim := NewSim("sess_stat_alloc", "01", loadRules(t))
	res := TickResult{Tick: sim.tick, Level: sim.currentLevel, Changes: []Change{}, Events: []Event{}}
	sim.awardExperience(20, "corr_xp", &res)
	player := sim.entities[sim.playerID]
	player.hp = 7

	sim.handleAllocateStat(Input{MessageID: "vit", CorrelationID: "corr_vit", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "vit", Points: 1}}, &res)
	assertAck(t, res, "vit")
	view := sim.CharacterProgressionView()
	if view.BaseStats.Vit != 6 || view.UnspentStatPoints != 4 || player.maxHP != 11 || player.hp != 8 {
		t.Fatalf("vit allocation progression=%+v player=%+v, want vit 6 unspent 4 hp 8/11", view, player)
	}
	if !hasEvent(res, "stat_allocated") || !hasProgressionChange(res) {
		t.Fatalf("missing stat allocation outputs: changes=%+v events=%+v", res.Changes, res.Events)
	}

	overspend := sim.Tick([]Input{{MessageID: "overspend", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "str", Points: 99}}})
	assertReject(t, overspend, "overspend", "not_enough_stat_points")
	invalid := sim.Tick([]Input{{MessageID: "invalid_stat", Type: "allocate_stat_intent", AllocateStat: &AllocateStatIntent{Stat: "luck", Points: 1}}})
	assertReject(t, invalid, "invalid_stat", "invalid_payload")
}

func TestStrengthDamageBonusAdjustsMeleeDamageRange(t *testing.T) {
	rules := loadRules(t)
	base := NewSim("sess_damage_base", "01", rules)
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
	assertEntity(t, dsnap, "1003", lootEntity, "", "training_badge", Vec2{X: 8, Y: 2})

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
				rulesCopy := *rules
				rulesCopy.Combat = rules.Combat
				rulesCopy.Combat.BaseHitChance = *tc.BaseHitChance
				rules = &rulesCopy
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

func TestProjectileBusyRejectsSecondFire(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	monster := firstEntityByKind(sim, monsterEntity)
	first := sim.Tick([]Input{{MessageID: "fire1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertAck(t, first, "fire1")
	second := sim.Tick([]Input{{MessageID: "fire2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(monster.id)}}})
	assertReject(t, second, "fire2", "projectile_busy")
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

func TestRangedDummyDropsThreeSeparatedLootItems(t *testing.T) {
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
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
		"training_badge": false,
		"quest_leaf":     false,
		"red_potion":     false,
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
	sim := rangedLabWithEquippedBow(t, loadRules(t), "cafebabecafebabe")
	if sim.playerActionReach() != 16.0 {
		t.Fatalf("playerActionReach = %v, want bow reach 16.0", sim.playerActionReach())
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
	if hasEvent(pickup, "item_picked_up") {
		t.Fatal("loot picked up instantly from ranged distance")
	}

	picked := false
	for i := 0; i < 80; i++ {
		r := sim.Tick(nil)
		if hasEvent(r, "item_picked_up") {
			picked = true
			break
		}
	}
	if !picked {
		t.Fatal("auto-nav did not complete loot pickup within tick budget")
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

func firstEntityByKind(sim *Sim, kind string) *entity {
	for _, id := range sortedEntityIDs(sim.entities) {
		if sim.entities[id].kind == kind {
			return sim.entities[id]
		}
	}
	return nil
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

	for _, c := range golden.Cases {
		sim := NewSim("sess_item_roll_"+c.Name, c.Seed, r)
		got, ok := sim.rollItemTemplate(golden.TemplateID)
		if !ok {
			t.Fatalf("%s: rollItemTemplate returned false", c.Name)
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

	for _, c := range golden.Cases {
		got := rules.RollTreasureClass(golden.TreasureClassID, NewRNG(SeedToUint64(c.Seed)))
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
	sim := NewSim("sess_rolled_loot", "0000000000000004", rules)
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
	sim := NewSim("sess_legacy_loot", "01", rules)
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
	sim := NewSim("sess_rolled_damage", "01", rules)
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
	sim.inventory = append(sim.inventory, item)
	sim.equipped[mainHandSlot] = item.instanceID
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
	snap := NewSim("sess_equipment_slots", "01", loadRules(t)).Snapshot()
	for _, slot := range golden.EquipmentSlots {
		if _, ok := snap.Equipped[slot]; !ok {
			t.Fatalf("snapshot missing equipped slot %q: %+v", slot, snap.Equipped)
		}
	}
}

func TestEquipmentSlotCompatibilityAndRings(t *testing.T) {
	rules := loadRules(t)
	sim := NewSim("sess_equipment_slots", "01", rules)
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
	sim := NewSim("sess_wrong_slot", "01", loadRules(t))
	shield := addRolledInventoryItem(t, sim, 6100, "cave_shield", nil)
	res := sim.Tick([]Input{{
		MessageID: "wrong",
		Type:      "equip_intent",
		Equip:     &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: mainHandSlot},
	}})
	assertReject(t, res, "wrong", "wrong_slot")
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
		sim := NewSim("sess_one_hand_shield", "01", loadRules(t))
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
		sim := NewSim("sess_two_hand_clear", "01", loadRules(t))
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

	t.Run("offhand blocked by two handed main hand", func(t *testing.T) {
		sim := NewSim("sess_two_hand_block", "01", loadRules(t))
		bow := addRolledInventoryItem(t, sim, 6220, "cave_bow", nil)
		shield := addRolledInventoryItem(t, sim, 6221, "cave_shield", nil)
		assertAck(t, sim.Tick([]Input{{MessageID: "bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "bow")
		res := sim.Tick([]Input{{MessageID: "shield", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(shield.instanceID), Slot: offHandSlot}}})
		assertReject(t, res, "shield", "hands_blocked")
	})

	t.Run("bow occupies both hands", func(t *testing.T) {
		sim := NewSim("sess_bow_occupies", "01", loadRules(t))
		bow := addRolledInventoryItem(t, sim, 6230, "cave_bow", map[string]int{"damage_min": 8, "damage_max": 8})
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
		sim := NewSim("sess_shield_display", "01", loadRules(t))
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

// --- scripted slice ---------------------------------------------------------

// runSlice drives a sim through the full vertical-slice flow and returns it.
func runSlice(t *testing.T, seed string) *Sim {
	t.Helper()
	sim := NewSim("sess_test", seed, loadRules(t))

	// Move into unarmed reach of the monster.
	sim.Tick([]Input{{MessageID: "m1", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 2}}})

	// Attack until the monster is dead.
	monsterID := "1002"
	for i := 0; i < 10; i++ {
		if e := sim.findEntity(monsterID); e == nil || e.hp == 0 {
			break
		}
		sim.Tick([]Input{{MessageID: "a" + itoa(i), CorrelationID: "corr_a", Type: "action_intent", Action: &ActionIntent{TargetID: monsterID}}})
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
	sim := NewSim("sess_retaliate", "deadbeefdeadbeef", loadRules(t))
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
	if !hasLootSpawn(r, "training_badge") {
		t.Fatalf("missing training_badge loot spawn: %+v", r.Changes)
	}
	lootPos, ok := lootSpawnPosition(r, "training_badge")
	if !ok {
		t.Fatalf("missing training_badge loot spawn position: %+v", r.Changes)
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
	sim := NewSim("sess_overkill_damage_event", "deadbeefdeadbeef", rules)
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
	rules.Combat.BaseHitChance = 0
	sim := NewSim("sess_miss", "deadbeefdeadbeef", rules)
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
	if hasEvent(r, "player_damaged") || hasEvent(r, "player_killed") || hasPlayerUpdate(r) {
		t.Fatalf("miss retaliated unexpectedly: changes=%+v events=%+v", r.Changes, r.Events)
	}
	if sim.entities[sim.playerID].hp != playerStartHP {
		t.Fatalf("player hp = %d, want %d", sim.entities[sim.playerID].hp, playerStartHP)
	}
}

func TestPlayerKilledByRetaliation(t *testing.T) {
	rules := loadRules(t)
	dummy := rules.Monsters[monsterDefID]
	dummy.MaxHP = 100
	rules.Monsters[monsterDefID] = dummy

	sim := NewSim("sess_player_death", "deadbeefdeadbeef", rules)
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
	sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "training_badge", slot: "", equipped: false})

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
	loot := findLootByDef(sim, "training_badge")
	if loot == nil {
		t.Fatal("missing dropped training_badge loot")
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
	sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "training_badge"})
	player := sim.entities[sim.playerID]
	for ring := 1; ring <= 6; ring++ {
		for _, offset := range adjacentUnitOffsets() {
			sim.walls = append(sim.walls, wallObstacle{
				pos:  Vec2{X: player.pos.X + offset.X*float64(ring), Y: player.pos.Y + offset.Y*float64(ring)},
				size: Vec2{X: 1, Y: 1},
			})
		}
	}

	r := sim.Tick([]Input{{MessageID: "drop", Type: "drop_intent", Drop: &DropIntent{ItemInstanceID: "5000"}}})
	assertReject(t, r, "drop", "no_drop_space")
	if len(sim.inventory) != 1 {
		t.Fatalf("inventory mutated on rejected drop: %+v", sim.inventory)
	}
	if findLootByDef(sim, "training_badge") != nil {
		t.Fatal("rejected drop spawned loot")
	}
}

func TestUseConsumableHealLab(t *testing.T) {
	rules := loadRules(t)
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
	if player.hp != 4 {
		t.Fatalf("player hp after combat = %d, want 4", player.hp)
	}
	loots := findAllLootByDef(sim, "red_potion")
	if len(loots) != 2 {
		t.Fatalf("loot drops = %+v, want two red_potion", loots)
	}

	for i := 0; i < 2; i++ {
		loot := findLootByDef(sim, "red_potion")
		if loot == nil {
			t.Fatalf("missing red_potion loot pickup %d", i)
		}
		pickup := sim.Tick([]Input{{
			MessageID: fmt.Sprintf("pickup-%d", i),
			Type:      "action_intent",
			Action:    &ActionIntent{TargetID: idStr(loot.id)},
		}})
		assertAck(t, pickup, fmt.Sprintf("pickup-%d", i))
		if sim.autoNav != nil {
			for step := 0; step < 30 && findLootByDef(sim, "red_potion") != nil; step++ {
				sim.Tick(nil)
			}
		}
	}
	if len(sim.inventory) != 2 {
		t.Fatalf("inventory after pickups = %+v, want two items", sim.inventory)
	}

	firstID := idStr(sim.inventory[0].instanceID)
	use1 := sim.Tick([]Input{{
		MessageID: "use1",
		Type:      "use_intent",
		Use:       &UseIntent{ItemInstanceID: firstID},
	}})
	assertAck(t, use1, "use1")
	if player.hp != 9 {
		t.Fatalf("player hp after first use = %d, want 9", player.hp)
	}
	assertEventHeal(t, use1, "player_healed", 5)

	secondID := idStr(sim.inventory[0].instanceID)
	use2 := sim.Tick([]Input{{
		MessageID: "use2",
		Type:      "use_intent",
		Use:       &UseIntent{ItemInstanceID: secondID},
	}})
	assertAck(t, use2, "use2")
	if player.hp != 10 {
		t.Fatalf("player hp after second use = %d, want 10", player.hp)
	}
	if len(sim.inventory) != 0 {
		t.Fatalf("inventory after second use = %+v, want empty", sim.inventory)
	}
	assertEventHeal(t, use2, "player_healed", 1)
}

func TestUseConsumableRejectsFullHP(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_use_full", "01", rules, "heal_lab")
	if err != nil {
		t.Fatalf("new heal lab: %v", err)
	}
	sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "red_potion", equipped: false})

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
	sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "training_badge", equipped: false})
	sim.entities[sim.playerID].hp = 5

	r := sim.Tick([]Input{{MessageID: "use", Type: "use_intent", Use: &UseIntent{ItemInstanceID: "5000"}}})
	assertReject(t, r, "use", "not_consumable")
}

func TestHotbarCapacityAndBelt(t *testing.T) {
	sim := NewSim("sess_hotbar_capacity", "01", loadRules(t))
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

func TestHotbarAssignUseDirectUseAndReenable(t *testing.T) {
	sim := NewSim("sess_hotbar_use", "01", loadRules(t))
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

func TestHotbarRejectsAndDropClears(t *testing.T) {
	sim := NewSim("sess_hotbar_rejects", "01", loadRules(t))
	potion := addStaticInventoryItem(sim, 7200, "red_potion")
	badge := addStaticInventoryItem(sim, 7201, "training_badge")

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
		loot := &entity{kind: lootEntity, pos: pos, itemDefID: "training_badge"}
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
	sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: golden.ItemDefID})
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
	// 3 ticks of speed 1 in +x.
	got := sim.entities[sim.playerID].pos
	if got.X != start.X+3*moveSpeed || got.Y != start.Y {
		t.Fatalf("player pos = %+v, want x=%v", got, start.X+3*moveSpeed)
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
	sim := NewSim("s", "01", loadRules(t))
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
		sim := NewSim("sess_range_monster", "01", rules)
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
	again := sim.Tick([]Input{{MessageID: "open_chest_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, again, "open_chest_again", "invalid_target")
	if got := countEntitiesByKind(sim.activeLevel(), lootEntity); got != afterLoot {
		t.Fatalf("reopen changed loot count = %d, want %d", got, afterLoot)
	}
}

// --- rejections (criterion 12) ----------------------------------------------

func TestRejections(t *testing.T) {
	rules := loadRules(t)

	t.Run("invalid attack target", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "action_intent", Action: &ActionIntent{TargetID: "9999"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("pickup non-loot", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		sim.findEntity("1002").hp = 0
		r := sim.Tick([]Input{{MessageID: "x", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("equip not in inventory", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
		assertReject(t, r, "x", "not_in_inventory")
	})

	t.Run("equip non-equippable", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "training_badge"})
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: mainHandSlot}}})
		assertReject(t, r, "x", "not_equippable")
	})

	t.Run("unknown type", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
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
		{MessageID: "attack", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}},
		{MessageID: "pickup", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}},
		{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: mainHandSlot}},
		{MessageID: "unequip", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: mainHandSlot}},
		{MessageID: "drop", Type: "drop_intent", Drop: &DropIntent{ItemInstanceID: "1004"}},
	}
	for _, in := range cases {
		sim := NewSim("sess_dead_"+in.MessageID, "01", rules)
		sim.entities[sim.playerID].hp = 0
		r := sim.Tick([]Input{in})
		assertReject(t, r, in.MessageID, "player_dead")
	}

	sim := NewSim("sess_dead_move", "01", rules)
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
	sim.inventory = append(sim.inventory, item)
	return item
}

func addStaticInventoryItem(sim *Sim, instanceID uint64, itemDefID string) *invItem {
	item := &invItem{instanceID: instanceID, itemDefID: itemDefID}
	sim.inventory = append(sim.inventory, item)
	return item
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

func hasEquippedUpdate(r TickResult, slot string, itemInstanceID *string) bool {
	for _, c := range r.Changes {
		if c.Op != OpEquippedUpdate || c.Slot != slot {
			continue
		}
		if itemInstanceID == nil {
			return c.ItemInstanceID == nil
		}
		return c.ItemInstanceID != nil && *c.ItemInstanceID == *itemInstanceID
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
	if got := generatedTeleporterPos(level1); golden.Levels["-1"].Teleporter == nil || got != *golden.Levels["-1"].Teleporter {
		t.Fatalf("level -1 teleporter = %+v, want %+v", got, golden.Levels["-1"].Teleporter)
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
	if got := generatedTeleporterPos(level2); golden.Levels["-2"].Teleporter == nil || got != *golden.Levels["-2"].Teleporter {
		t.Fatalf("level -2 teleporter = %+v, want %+v", got, golden.Levels["-2"].Teleporter)
	}
	if len(golden.Levels["-2"].Loot) != 1 {
		t.Fatalf("level -2 golden loot = %+v, want one entry", golden.Levels["-2"].Loot)
	}
	wantLoot := golden.Levels["-2"].Loot[0]
	if got, ok := generatedLootPos(level2, wantLoot.ItemDefID); !ok || got != wantLoot.Position {
		t.Fatalf("level -2 loot %s = %+v/%v, want %+v", wantLoot.ItemDefID, got, ok, wantLoot.Position)
	}
	if distance(wantLoot.Position, *golden.Levels["-2"].StairsUp) < dungeonCoinStairDistance {
		t.Fatalf("level -2 coin distance from stairs_up = %v, want at least %v", distance(wantLoot.Position, *golden.Levels["-2"].StairsUp), dungeonCoinStairDistance)
	}
}

func TestDungeonMonsterGeneration(t *testing.T) {
	rules := loadRules(t)
	placement := rules.DungeonGeneration.MonsterPlacement
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
			if len(level.monsters) != placement.Count {
				t.Fatalf("level %d monsters = %d, want %d", levelNum, len(level.monsters), placement.Count)
			}
			if len(again.monsters) != len(level.monsters) {
				t.Fatalf("repeat level %d monsters = %d, want %d", levelNum, len(again.monsters), len(level.monsters))
			}
			for i, monster := range level.monsters {
				if monster.defID != placement.MonsterDefID {
					t.Fatalf("level %d monster %d defID = %s, want %s", levelNum, i, monster.defID, placement.MonsterDefID)
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
		})
	}
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
	if golden.BaseMonsterCount != rules.DungeonGeneration.MonsterPlacement.Count {
		t.Fatalf("base monster count = %d, want rules %d", golden.BaseMonsterCount, rules.DungeonGeneration.MonsterPlacement.Count)
	}
	if golden.MonsterCountBonus != rules.DungeonGeneration.ChestPlacement.MonsterCountBonus {
		t.Fatalf("monster bonus = %d, want rules %d", golden.MonsterCountBonus, rules.DungeonGeneration.ChestPlacement.MonsterCountBonus)
	}
	for _, c := range golden.Cases {
		level, err := GenerateDungeonLevel(c.Seed, golden.Level, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("%s: generate: %v", c.Name, err)
		}
		if len(level.monsters) != c.ExpectedMonsterCount {
			t.Fatalf("%s: monsters = %d, want %d", c.Name, len(level.monsters), c.ExpectedMonsterCount)
		}
		if c.ExpectedChest == nil {
			if len(level.chests) != 0 {
				t.Fatalf("%s: chests = %+v, want none", c.Name, level.chests)
			}
			continue
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
		{-10, "guarded_chest_drop_depth_3_plus"},
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
				Index        int         `json:"index"`
				Rarity       string      `json:"rarity"`
				LootTable    string      `json:"loot_table"`
				MaxHP        int         `json:"max_hp"`
				AttackDamage DamageRange `json:"attack_damage"`
				XPReward     int         `json:"xp_reward"`
			} `json:"expected_monsters"`
		} `json:"generated_cases"`
	}
	loadGolden(t, "monster_rarity.json", &golden)
	rules := loadRules(t)
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
	if monster.monsterRarityID != "" || monster.monsterAttackDamage != nil || monster.monsterXPReward != 0 {
		t.Fatalf("static monster has generated rarity fields: %+v", monster)
	}
	if monster.view().Rarity != "" {
		t.Fatalf("static monster view rarity = %q, want empty", monster.view().Rarity)
	}
}

func TestDungeonEquipmentLootDeterminism(t *testing.T) {
	first := dungeonEquipmentKillLootSequence(t, "v29_replay_equipment_0")
	second := dungeonEquipmentKillLootSequence(t, "v29_replay_equipment_0")
	if !sameStrings(first, second) {
		t.Fatalf("same-seed loot sequence drifted: %v != %v", first, second)
	}
	if len(first) == 0 {
		t.Fatal("expected at least one dungeon equipment drop")
	}
	foundEquipment := false
	for _, drop := range first {
		if drop == "cave_bow:cave_bow" {
			foundEquipment = true
			break
		}
	}
	if !foundEquipment {
		t.Fatalf("loot sequence = %v, want rolled cave_bow equipment", first)
	}
}

func TestDungeonMonsterProactiveAttackGolden(t *testing.T) {
	var golden struct {
		SessionSeed              string `json:"session_seed"`
		Level                    int    `json:"level"`
		MonsterDefID             string `json:"monster_def_id"`
		TickOfFirstPlayerDamaged uint64 `json:"tick_of_first_player_damaged"`
		Damage                   int    `json:"damage"`
		PlayerHPAfter            int    `json:"player_hp_after"`
	}
	loadGolden(t, "dungeon_monster_attack.json", &golden)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_dungeon_monster_attack", golden.SessionSeed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	results := descendFromCurrentLevel(t, sim, "descend")
	assertLevelChanged(t, results[0], 0, golden.Level)

	firstDamage, ok := waitForPlayerDamage(sim, 240)
	if !ok {
		t.Fatalf("expected proactive player_damaged event; player=%+v monsters=%+v", sim.activeLevel().entities[sim.playerID].pos, dungeonMonsterDebugPositions(sim.activeLevel()))
	}
	if firstDamage.Tick != golden.TickOfFirstPlayerDamaged {
		t.Fatalf("first damage tick = %d, want %d", firstDamage.Tick, golden.TickOfFirstPlayerDamaged)
	}
	if eventDamage(firstDamage, "player_damaged") != golden.Damage {
		t.Fatalf("first damage = %d, want %d", eventDamage(firstDamage, "player_damaged"), golden.Damage)
	}
	player := sim.activeLevel().entities[sim.playerID]
	if player.hp != golden.PlayerHPAfter {
		t.Fatalf("player hp = %d, want %d", player.hp, golden.PlayerHPAfter)
	}
	if countLiveMonstersByDef(sim.activeLevel(), golden.MonsterDefID) != rules.DungeonGeneration.MonsterPlacement.Count {
		t.Fatalf("live %s count mismatch", golden.MonsterDefID)
	}
}

func TestDungeonMonsterAttackCooldownAndDeterminism(t *testing.T) {
	var golden struct {
		SessionSeed string `json:"session_seed"`
	}
	loadGolden(t, "dungeon_monster_attack.json", &golden)
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_dungeon_monster_cooldown", golden.SessionSeed, rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	descendFromCurrentLevel(t, sim, "descend")
	firstDamage, ok := waitForPlayerDamage(sim, 240)
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
	replayDamage, ok := waitForPlayerDamage(replay, 240)
	if !ok {
		t.Fatal("expected replay proactive damage")
	}
	if replayDamage.Tick != firstDamage.Tick || eventDamage(replayDamage, "player_damaged") != eventDamage(firstDamage, "player_damaged") {
		t.Fatalf("replay first damage = tick %d damage %d, want tick %d damage %d", replayDamage.Tick, eventDamage(replayDamage, "player_damaged"), firstDamage.Tick, eventDamage(firstDamage, "player_damaged"))
	}
}

func TestDungeonDescendAscendTransitions(t *testing.T) {
	var golden struct {
		Seed              string `json:"seed"`
		DescendThenAscend struct {
			ExpectedLevel          int  `json:"expected_level"`
			ExpectedPlayerPosition Vec2 `json:"expected_player_position"`
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
	if townDown == nil || townDown.pos != (Vec2{X: 8, Y: 10}) {
		t.Fatalf("town down stair = %+v, want {8 10}", townDown)
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
	if got := sim.entities[sim.playerID].pos; got != up.pos {
		t.Fatalf("player position after descend = %+v, want up stair %+v", got, up.pos)
	}
	coin := findLootByDef(sim, "training_badge")
	if coin == nil {
		t.Fatal("missing dungeon training_badge coin")
	}
	pickup := sim.Tick([]Input{{MessageID: "pick_coin", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(coin.id)}}})
	assertAck(t, pickup, "pick_coin")
	if len(sim.inventory) != 0 {
		t.Fatalf("coin picked up without leaving stair: %+v", sim.inventory)
	}
	pickupTicks := 1
	for ; pickupTicks < 20 && len(sim.inventory) == 0; pickupTicks++ {
		sim.Tick(nil)
	}
	if len(sim.inventory) != 1 || sim.inventory[0].itemDefID != "training_badge" {
		t.Fatalf("inventory after coin pickup = %+v, want training_badge", sim.inventory)
	}
	if pickupTicks < 5 {
		t.Fatalf("coin pickup took %d ticks from stair, want at least 5", pickupTicks)
	}

	sim.entities[sim.playerID].pos = up.pos
	results = sim.TickResults([]Input{{MessageID: "ascend", Type: "ascend_intent", Ascend: &AscendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("ascend results = %d, want 2: %+v", len(results), results)
	}
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != down.pos {
		t.Fatalf("player position after ascend = %+v, want %+v", got, down.pos)
	}
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
	if got := sim.entities[sim.playerID].pos; got != golden.DescendThenAscend.ExpectedPlayerPosition {
		t.Fatalf("player position after town ascend = %+v, want %+v", got, golden.DescendThenAscend.ExpectedPlayerPosition)
	}
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
	descendFromCurrentLevel(t, sim, "descend_town")
	level1Teleporter := sim.findTeleporter(sim.activeLevel())
	if level1Teleporter == nil {
		t.Fatal("missing level -1 teleporter")
	}
	sim.entities[sim.playerID].pos = level1Teleporter.pos

	discover1 := sim.Tick([]Input{{MessageID: "discover_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level1Teleporter.id)}}})
	assertAck(t, discover1, "discover_1")
	if !hasTeleporterDiscoveryUpdate(discover1, -1) || !hasTeleporterDiscoveredEvent(discover1, -1) {
		t.Fatalf("missing level -1 discovery result: changes=%+v events=%+v", discover1.Changes, discover1.Events)
	}
	if !sim.discoveredTeleporters[-1] {
		t.Fatal("level -1 teleporter not marked discovered")
	}

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	sim.entities[sim.playerID].pos = down.pos
	results := sim.TickResults([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2", len(results))
	}
	level2Teleporter := sim.findTeleporter(sim.activeLevel())
	if level2Teleporter == nil {
		t.Fatal("missing level -2 teleporter")
	}
	sim.entities[sim.playerID].pos = level2Teleporter.pos

	reject := sim.Tick([]Input{{MessageID: "tp_before_discover", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -1}}})
	assertReject(t, reject, "tp_before_discover", "teleporter_not_discovered")

	discover2 := sim.Tick([]Input{{MessageID: "discover_2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level2Teleporter.id)}}})
	assertAck(t, discover2, "discover_2")
	if !hasTeleporterDiscoveryUpdate(discover2, -2) || !hasTeleporterDiscoveredEvent(discover2, -2) {
		t.Fatalf("missing level -2 discovery result: changes=%+v events=%+v", discover2.Changes, discover2.Events)
	}

	results = sim.TickResults([]Input{{MessageID: "tp_to_1", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -1}}})
	if len(results) != 2 {
		t.Fatalf("teleport results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], -2, -1)
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != level1Teleporter.pos {
		t.Fatalf("player position after teleport = %+v, want %+v", got, level1Teleporter.pos)
	}

	results = sim.TickResults([]Input{{MessageID: "tp_to_town", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: townLevel}}})
	if len(results) != 2 {
		t.Fatalf("teleport to town results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], -1, townLevel)
	if sim.currentLevel != townLevel {
		t.Fatalf("currentLevel = %d, want town", sim.currentLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != townTeleporter.pos {
		t.Fatalf("player position after town teleport = %+v, want %+v", got, townTeleporter.pos)
	}
}

func TestLoadDiscoveredTeleportersAllowsFreshSessionWaypointTravel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_loaded_waypoint", "feedfacefeedface", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
	sim.LoadDiscoveredTeleporters([]int{-1})
	if _, ok := sim.levels[-1]; ok {
		t.Fatal("level -1 should not be generated by loading waypoint discovery")
	}

	view := sim.Snapshot().DiscoveredTeleporters
	assertTeleporterDiscoveryView(t, view, []struct {
		Level      int  `json:"level"`
		Discovered bool `json:"discovered"`
	}{{Level: -1, Discovered: true}, {Level: 0, Discovered: true}})

	townTeleporter := sim.findTeleporter(sim.activeLevel())
	if townTeleporter == nil {
		t.Fatal("missing town teleporter")
	}
	sim.entities[sim.playerID].pos = townTeleporter.pos
	results := sim.TickResults([]Input{{MessageID: "tp_loaded_waypoint", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -1}}})
	if len(results) != 2 {
		t.Fatalf("teleport results = %d, want 2: %+v", len(results), results)
	}
	assertLevelChanged(t, results[0], townLevel, -1)
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	if _, ok := sim.levels[-1]; !ok {
		t.Fatal("teleport to loaded waypoint did not generate level -1")
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
	descendFromCurrentLevel(t, sim, "descend_town")
	level1Teleporter := sim.findTeleporter(sim.activeLevel())
	if level1Teleporter == nil {
		t.Fatal("missing level -1 teleporter")
	}
	sim.entities[sim.playerID].pos = level1Teleporter.pos
	discover1 := sim.Tick([]Input{{MessageID: "discover_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level1Teleporter.id)}}})
	assertAck(t, discover1, "discover_1")

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	sim.entities[sim.playerID].pos = down.pos
	results := sim.TickResults([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2", len(results))
	}
	if !hasTeleporterDiscoveryUpdateWith(results[1], -2, false) {
		t.Fatalf("descend arrival missing undiscovered teleporter update for -2: %+v", results[1].Changes)
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

	level1Teleporter = sim.findTeleporter(sim.activeLevel())
	sim.entities[sim.playerID].pos = level1Teleporter.pos
	assertTeleporterDiscoveryView(t, sim.teleporterDiscoveryView(), golden.VisitedUndiscoveredTarget.DiscoveredTeleporters)

	reject := sim.Tick([]Input{{MessageID: "tp_undiscovered_target", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -2}}})
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
	descendFromCurrentLevel(t, sim, "descend_town")
	level1Teleporter := sim.findTeleporter(sim.activeLevel())
	if level1Teleporter == nil {
		t.Fatal("missing level -1 teleporter")
	}
	sim.entities[sim.playerID].pos = level1Teleporter.pos
	assertAck(t, sim.Tick([]Input{{MessageID: "discover_1", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level1Teleporter.id)}}}), "discover_1")

	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	sim.entities[sim.playerID].pos = down.pos
	sim.TickResults([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})

	level2Teleporter := sim.findTeleporter(sim.activeLevel())
	if level2Teleporter == nil {
		t.Fatal("missing level -2 teleporter")
	}
	sim.entities[sim.playerID].pos = level2Teleporter.pos
	assertAck(t, sim.Tick([]Input{{MessageID: "discover_2", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(level2Teleporter.id)}}}), "discover_2")

	sim.TickResults([]Input{{MessageID: "tp_to_1", Type: "teleport_intent", Teleport: &TeleportIntent{TargetLevel: -1}}})

	want := golden.DiscoverDescendTeleport
	if sim.currentLevel != want.ExpectedLevel {
		t.Fatalf("currentLevel = %d, want %d", sim.currentLevel, want.ExpectedLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != want.ExpectedPlayerPosition {
		t.Fatalf("player position = %+v, want %+v", got, want.ExpectedPlayerPosition)
	}
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

func generatedLootPos(level generatedDungeonLevel, itemDefID string) (Vec2, bool) {
	for _, loot := range level.loot {
		if loot.itemDefID == itemDefID {
			return loot.pos, true
		}
	}
	return Vec2{}, false
}

func liveDungeonMonsters(level *LevelState) []*entity {
	out := []*entity{}
	for _, id := range sortedEntityIDs(level.entities) {
		candidate := level.entities[id]
		if candidate != nil && candidate.kind == monsterEntity && candidate.monsterDefID == "dungeon_mob" && candidate.hp > 0 {
			out = append(out, candidate)
		}
	}
	return out
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

func dungeonMonsterDebugPositions(level *LevelState) []Vec2 {
	positions := []Vec2{}
	for _, id := range sortedEntityIDs(level.entities) {
		entity := level.entities[id]
		if entity.kind == monsterEntity && entity.monsterDefID == "dungeon_mob" {
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

func hasProgressionChange(r TickResult) bool {
	for _, change := range r.Changes {
		if change.Op == OpCharacterProgressionUpdate && change.Progression != nil {
			return true
		}
	}
	return false
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
	assertFloat("hit_chance", got.HitChance, want.HitChance)
	assertFloat("crit_chance", got.CritChance, want.CritChance)
	assertFloat("crit_damage", got.CritDamage, want.CritDamage)
	assertFloat("movement_speed", got.MovementSpeed, want.MovementSpeed)
	assertFloat("max_hp", got.MaxHP, want.MaxHP)
	assertFloat("max_mana", got.MaxMana, want.MaxMana)
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
