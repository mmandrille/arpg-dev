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
	if r.Monsters[monsterDefID].MaxHP != 3 {
		t.Fatalf("training_dummy max_hp = %d, want 3", r.Monsters[monsterDefID].MaxHP)
	}
	if !r.Items["rusty_sword"].Equippable || r.Items["rusty_sword"].Slot != "weapon" {
		t.Fatalf("rusty_sword def = %+v", r.Items["rusty_sword"])
	}
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

func TestLootRollGolden(t *testing.T) {
	r := loadRules(t)
	var golden struct {
		LootTable          string `json:"loot_table"`
		ExpectedItemDefID  string `json:"expected_item_def_id"`
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

	// Move toward the monster for a few ticks (exercises movement; combat does
	// not gate on range in v0).
	sim.Tick([]Input{{MessageID: "m1", Type: "move_intent", Move: &MoveIntent{Direction: Vec2{X: 1}, DurationTicks: 2}}})

	// Attack until the monster is dead.
	monsterID := "1002"
	for i := 0; i < 10; i++ {
		if e := sim.findEntity(monsterID); e == nil || e.hp == 0 {
			break
		}
		sim.Tick([]Input{{MessageID: "a" + itoa(i), CorrelationID: "corr_a", Type: "attack_intent", Attack: &AttackIntent{TargetID: monsterID}}})
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
	sim.Tick([]Input{{MessageID: "p1", CorrelationID: "corr_p", Type: "pick_up_intent", PickUp: &PickUpIntent{EntityID: lootID}}})

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
		MonsterDefID    string `json:"monster_def_id"`
		DroppedItemDefID string `json:"dropped_item_def_id"`
		FinalPlayerHP   int    `json:"final_player_hp"`
		FinalMonsterHP  int    `json:"final_monster_hp"`
		FinalInventory  []struct {
			ItemDefID string `json:"item_def_id"`
			Slot      string `json:"slot"`
			Equipped  bool   `json:"equipped"`
		} `json:"final_inventory"`
		FinalEquipped struct {
			Weapon string `json:"weapon"`
		} `json:"final_equipped"`
	}
	loadGolden(t, "slice_outcome.json", &golden)

	sim := runSlice(t, "deadbeefdeadbeef")
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
	sim := NewSim("sess_move", "abcd", loadRules(t))
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

// --- rejections (criterion 12) ----------------------------------------------

func TestRejections(t *testing.T) {
	rules := loadRules(t)

	t.Run("invalid attack target", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "attack_intent", Attack: &AttackIntent{TargetID: "9999"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("pickup non-loot", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "pick_up_intent", PickUp: &PickUpIntent{EntityID: "1002"}}})
		assertReject(t, r, "x", "invalid_target")
	})

	t.Run("equip not in inventory", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: "5000", Slot: "weapon"}}})
		assertReject(t, r, "x", "not_in_inventory")
	})

	t.Run("unknown type", func(t *testing.T) {
		sim := NewSim("s", "01", rules)
		r := sim.Tick([]Input{{MessageID: "x", Type: "teleport_intent"}})
		assertReject(t, r, "x", "unknown_type")
	})

	t.Run("duplicate pickup", func(t *testing.T) {
		sim := runSlice(t, "0011223344556677")
		// The loot was already picked up during runSlice; picking up 1003 again rejects.
		r := sim.Tick([]Input{{MessageID: "dup", Type: "pick_up_intent", PickUp: &PickUpIntent{EntityID: "1003"}}})
		assertReject(t, r, "dup", "invalid_target")
	})
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
