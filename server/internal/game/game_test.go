package game

import (
	"encoding/json"
	"fmt"
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
	if !r.Items["rusty_sword"].Equippable || r.Items["rusty_sword"].Slot != "weapon" {
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
	if !bow.Equippable || bow.Slot != weaponSlot || bow.AttackMode != attackModeRanged || bow.Damage == nil || bow.Reach == nil || bow.ProjectileSpeed == nil {
		t.Fatalf("training_bow def = %+v, want ranged weapon", bow)
	}
	if r.Interactables["wooden_door"].InitialState != interactableClosed {
		t.Fatalf("wooden_door = %+v, want initially closed", r.Interactables["wooden_door"])
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
	if !item.Equippable || item.Slot != weaponSlot || item.Damage == nil {
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
	equip := sim.Tick([]Input{{MessageID: "equip_bow", CorrelationID: "corr_equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: weaponSlot}}})
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
		if !ok || got != golden.ExpectedItemDefID {
			t.Fatalf("roll %s with seed %d = (%q,%v), want %q", golden.LootTable, seed, got, ok, golden.ExpectedItemDefID)
		}
	}
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
	sim.Tick([]Input{{MessageID: "e1", CorrelationID: "corr_e", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: itemID, Slot: "weapon"}}})

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
		FinalEquipped struct {
			Weapon string `json:"weapon"`
		} `json:"final_equipped"`
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
	// equipped weapon instance must resolve to the expected item_def_id.
	wp := snap.Equipped["weapon"]
	if wp == nil {
		t.Fatal("no weapon equipped")
	}
	if got.ItemInstanceID != *wp || got.ItemDefID != golden.FinalEquipped.Weapon {
		t.Fatalf("equipped weapon = %v (%s), want def %s", *wp, got.ItemDefID, golden.FinalEquipped.Weapon)
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
		Unequip:       &UnequipIntent{Slot: weaponSlot},
	}})
	assertAck(t, r, "unequip")
	if sim.equipped[weaponSlot] != 0 {
		t.Fatalf("equipped weapon = %d, want cleared", sim.equipped[weaponSlot])
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
	if sim.equipped[weaponSlot] != 0 {
		t.Fatalf("equipped weapon = %d, want cleared", sim.equipped[weaponSlot])
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
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: "weapon"}}})
		assertReject(t, r, "x", "not_in_inventory")
	})

	t.Run("equip non-equippable", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		sim.inventory = append(sim.inventory, &invItem{instanceID: 5000, itemDefID: "training_badge"})
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: "weapon"}}})
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
		{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "1004", Slot: "weapon"}},
		{MessageID: "unequip", Type: "unequip_intent", Unequip: &UnequipIntent{Slot: weaponSlot}},
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
		Equip:         &EquipIntent{ItemInstanceID: itemID, Slot: weaponSlot},
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

func hasChange(r TickResult, op string) bool {
	for _, c := range r.Changes {
		if c.Op == op {
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
		Equip:         &EquipIntent{ItemInstanceID: itemID, Slot: weaponSlot},
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
	if sim.currentLevel != -1 {
		t.Fatalf("currentLevel = %d, want -1", sim.currentLevel)
	}
	down := sim.findStair(sim.activeLevel(), stairsDownDefID)
	if down == nil {
		t.Fatal("missing down stairs on level -1")
	}
	sim.entities[sim.playerID].pos = down.pos

	results := sim.TickResults([]Input{{MessageID: "descend", Type: "descend_intent", Descend: &DescendIntent{}}})
	if len(results) != 2 {
		t.Fatalf("descend results = %d, want 2: %+v", len(results), results)
	}
	if results[0].Level != -1 || results[1].Level != -2 {
		t.Fatalf("descend result levels = %d/%d, want -1/-2", results[0].Level, results[1].Level)
	}
	if !hasEntityRemove(results[0], idStr(sim.playerID)) {
		t.Fatalf("from-level result missing player remove: %+v", results[0].Changes)
	}
	if !hasEntitySpawn(results[1], idStr(sim.playerID)) {
		t.Fatalf("to-level result missing player spawn: %+v", results[1].Changes)
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
	if sim.currentLevel != golden.DescendThenAscend.ExpectedLevel {
		t.Fatalf("currentLevel = %d, want %d", sim.currentLevel, golden.DescendThenAscend.ExpectedLevel)
	}
	if got := sim.entities[sim.playerID].pos; got != golden.DescendThenAscend.ExpectedPlayerPosition {
		t.Fatalf("player position after ascend = %+v, want %+v", got, golden.DescendThenAscend.ExpectedPlayerPosition)
	}
	assertLevelChanged(t, results[0], -2, -1)
}

func TestDungeonTeleporterDiscoveryAndTravel(t *testing.T) {
	sim, err := NewSimWithWorld("sess_dungeon_tp", "deadbeefdeadbeef", loadRules(t), "dungeon_levels")
	if err != nil {
		t.Fatalf("new dungeon sim: %v", err)
	}
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
}

func TestTeleportRejectsUndiscoveredTargetLevel(t *testing.T) {
	var golden struct {
		Seed                         string `json:"seed"`
		VisitedUndiscoveredTarget    struct {
			RejectReason            string `json:"reject_reason"`
			DiscoveredTeleporters   []struct {
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
