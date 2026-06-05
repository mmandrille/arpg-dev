package game

import (
	"encoding/json"
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
	assertEntity(t, gsnap, "1001", playerEntity, "", "", Vec2{X: 0, Y: 5})
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
	if len(collision.walls) != 2 {
		t.Fatalf("collision walls = %d, want 2", len(collision.walls))
	}
	assertEntity(t, csnap, "1001", playerEntity, "", "", Vec2{X: 0, Y: 5})
	assertEntity(t, csnap, "1002", monsterEntity, "training_dummy_reward", "", Vec2{X: 8, Y: 5})

	door, err := NewSimWithWorld("sess_door", "01", rules, "door_lab")
	if err != nil {
		t.Fatalf("door world: %v", err)
	}
	dsnap := door.Snapshot()
	if len(dsnap.Entities) != 3 {
		t.Fatalf("door entities = %d, want player+door+loot: %+v", len(dsnap.Entities), dsnap.Entities)
	}
	if len(door.walls) != 2 {
		t.Fatalf("door walls = %d, want 2", len(door.walls))
	}
	assertEntity(t, dsnap, "1001", playerEntity, "", "", Vec2{X: 0, Y: 5})
	assertInteractable(t, dsnap, "1002", "wooden_door", interactableClosed, Vec2{X: 4, Y: 5})
	assertEntity(t, dsnap, "1003", lootEntity, "", "training_badge", Vec2{X: 8, Y: 5})
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

	sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 10}}})
	for i := 0; i < 9; i++ {
		sim.Tick(nil)
	}

	player := sim.entities[sim.playerID]
	monster := sim.findEntity("1002")
	if player.pos != (Vec2{X: 7, Y: 5}) {
		t.Fatalf("player pos = %+v, want stopped at x=7 before monster after wall gap", player.pos)
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

	sim.Tick([]Input{{MessageID: "m", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 8}}})
	for i := 0; i < 7; i++ {
		sim.Tick(nil)
	}

	if got := sim.entities[sim.playerID].pos; got != (Vec2{X: 8, Y: 5}) {
		t.Fatalf("player pos = %+v, want able to enter dead monster position", got)
	}
}

func TestCollisionBlocksWallAndAllowsRoute(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_wall", "01", loadRules(t), "collision_lab")
	if err != nil {
		t.Fatalf("collision world: %v", err)
	}
	moveTicks(sim, "to_gap_edge", Vec2{X: 1}, 3)
	moveTicks(sim, "up_to_wall", Vec2{Y: -1}, 1)
	blocked := sim.Tick([]Input{{MessageID: "push_wall", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 2}}})
	sim.Tick(nil)
	if got := sim.entities[sim.playerID].pos; got != (Vec2{X: 3, Y: 4}) {
		t.Fatalf("player pos after wall push = %+v, want blocked at x=3,y=4", got)
	}
	if hasPlayerUpdate(blocked) {
		t.Fatalf("blocked wall push emitted player update: %+v", blocked.Changes)
	}

	moveTicks(sim, "back_to_gap", Vec2{Y: 1}, 1)
	moveTicks(sim, "through_gap", Vec2{X: 1}, 4)
	if got := sim.entities[sim.playerID].pos; got != (Vec2{X: 7, Y: 5}) {
		t.Fatalf("player gap route pos = %+v, want x=7,y=5 before monster", got)
	}
}

func TestActionRejectsOutOfRange(t *testing.T) {
	rules := loadRules(t)

	t.Run("monster", func(t *testing.T) {
		sim := NewSim("sess_range_monster", "01", rules)
		r := sim.Tick([]Input{{MessageID: "a", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertReject(t, r, "a", "out_of_range")
	})

	t.Run("loot", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_loot", "01", rules, "gear_before_combat")
		if err != nil {
			t.Fatalf("gear world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "p", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertReject(t, r, "p", "out_of_range")
	})

	t.Run("door", func(t *testing.T) {
		sim, err := NewSimWithWorld("sess_range_door", "01", rules, "door_lab")
		if err != nil {
			t.Fatalf("door world: %v", err)
		}
		r := sim.Tick([]Input{{MessageID: "d", Type: "action_intent", Action: &ActionIntent{TargetID: "1002"}}})
		assertReject(t, r, "d", "out_of_range")
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
	r := sim.Tick([]Input{{MessageID: "closed_loot", Type: "action_intent", Action: &ActionIntent{TargetID: "1003"}}})
	assertReject(t, r, "closed_loot", "out_of_range")

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
		r := sim.Tick([]Input{{MessageID: "x", Type: "teleport_intent"}})
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

func hasEvent(r TickResult, eventType string) bool {
	for _, ev := range r.Events {
		if ev.EventType == eventType {
			return true
		}
	}
	return false
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
