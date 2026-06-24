# Movement Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make hero movement speed a fully derived stat (class base × DEX multiplier × gear %), wire a `+% movement speed` rolleable affix onto boots/rings/amulets, fix inertia start-speed (60%), add a direction-correction grace window, fix player slow applying to movement, and update client prediction to use the server-sent derived speed.

**Architecture:** Approach A (multiplier model). `playerEffectiveMovementSpeed() = classBaseMovementSpeed() × formula_multiplier(DEX) × (1 + gearMovementSpeedPercent/100)`. `DerivedStatsView.MovementSpeed` becomes the canonical final tiles/tick value consumed by both server movement and client prediction. Slow debuff applied at move-time (not baked into derived stats).

**Tech Stack:** Go 1.21 (server), GDScript 4 (client), JSON schemas, shared rules data.

## Global Constraints

- Determinism invariant: no `time.Now()`, no `math/rand` without seeded RNG, no bare `for k, v := range map` in game/ hot-path — use `sortedStringKeys` helpers.
- New intent types register in `handlers.go`, not `applyInput`. (Not applicable to this slice — no new intent types.)
- Formula output scale change in `movement_speed` — run `make regen-golden` after Task 3 and commit updated golden fixtures alongside the rule change.
- File-size ratchet: `sim.go` is grandfathered at ~6568 lines. New helpers go into `sim.go` (movement domain already lives there). Net line change must stay ≤ +25 from baseline or require a payback extraction in the same slice.
- `game_test.go` is grandfathered at ~7854 lines. New tests are additions; keep them focused.
- Default character class is `"barbarian"` (see `sim.go:605`). Tests that use `NewSimWithWorld` / `MustNewSim` without a class override operate on a barbarian with `BaseMovementSpeed = 0.75`.

---

### Task 1: Rules data — class base speeds + movement_speed formula

**Files:**
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/character_progression.v0.schema.json`
- Modify: `server/internal/game/rules.go:186-190`

**Interfaces:**
- Produces: `CharacterClassDef.BaseMovementSpeed float64` (used by Task 3's `classBaseMovementSpeed()`)
- Produces: updated `movement_speed` formula — output is now a dimensionless multiplier in range `[0.5, 2.0]` (used by Task 3's `playerEffectiveMovementSpeed()`)

- [ ] **Step 1: Update character_progression.v0.schema.json — add base_movement_speed to class def**

In `shared/rules/character_progression.v0.schema.json`, find the `character_class` `$defs` entry and add `base_movement_speed`:

```json
"$defs": {
  "character_class": {
    "type": "object",
    "required": ["name", "light_radius", "base_stats"],
    "additionalProperties": false,
    "properties": {
      "name": { "type": "string", "minLength": 1 },
      "light_radius": { "type": "number", "exclusiveMinimum": 0 },
      "base_movement_speed": { "type": "number", "exclusiveMinimum": 0 },
      "base_stats": { "$ref": "#/$defs/base_stats" }
    }
  }
}
```

Note: `base_movement_speed` is intentionally not in `required` — missing means "fall back to main_config base".

- [ ] **Step 2: Update character_progression.v0.json — add base_movement_speed per class and update formula**

In `shared/rules/character_progression.v0.json`, add `"base_movement_speed"` to each class entry:

```json
"classes": {
  "rogue":      { "name": "Rogue",      "light_radius": 3.5, "base_movement_speed": 0.90, "base_stats": { ... } },
  "ranger":     { "name": "Ranger",     "light_radius": 3.5, "base_movement_speed": 0.85, "base_stats": { ... } },
  "barbarian":  { "name": "Barbarian",  "light_radius": 4.0, "base_movement_speed": 0.75, "base_stats": { ... } },
  "sorcerer":   { "name": "Sorceress",  "light_radius": 4.5, "base_movement_speed": 0.75, "base_stats": { ... } },
  "paladin":    { "name": "Paladin",    "light_radius": 4.0, "base_movement_speed": 0.65, "base_stats": { ... } }
}
```

Keep all other fields in each class unchanged. Only add `"base_movement_speed"`.

Also update the `movement_speed` formula in `derived_stats` (change from display-scale to multiplier):

```json
"derived_stats": {
  ...
  "movement_speed": { "type": "linear", "base": 1.0, "per_dex": 0.001, "min": 0.5, "max": 2.0 },
  ...
}
```

Old value was `{ "type": "linear", "base": 0.49, "per_dex": 0.0014, "min": 0.35, "max": 1.4 }`.

- [ ] **Step 3: Update rules.go — add BaseMovementSpeed to CharacterClassDef**

In `server/internal/game/rules.go`, update the struct at line 186:

```go
type CharacterClassDef struct {
	Name              string        `json:"name"`
	LightRadius       float64       `json:"light_radius"`
	BaseStats         BaseStatsView `json:"base_stats"`
	BaseMovementSpeed float64       `json:"base_movement_speed"`
}
```

No validation needed — zero means "use fallback", handled in Task 3.

- [ ] **Step 4: Verify rules load cleanly**

```bash
cd server && go test ./internal/game/... -run TestLoadRules -v 2>&1 | head -20
```

If `TestLoadRules` doesn't exist, run a broader check:

```bash
cd server && go build ./... 2>&1
```

Expected: no compilation errors, no test failures.

- [ ] **Step 5: Validate shared JSON**

```bash
make validate-shared
```

Expected: all validations pass.

- [ ] **Step 6: Commit**

```bash
git add shared/rules/character_progression.v0.json shared/rules/character_progression.v0.schema.json server/internal/game/rules.go
git commit -m "feat: add per-class base_movement_speed; movement_speed formula → multiplier scale"
```

---

### Task 2: movement_speed_percent gear stat — schema + item templates + accumulation

**Files:**
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `server/internal/game/sim.go` (lines ~235-252, ~5663-5820)
- Modify: `server/internal/game/game_test.go`

**Interfaces:**
- Consumes: nothing from prior tasks
- Produces: `effectiveCombatStats.MovementSpeedPercent float64` — accumulated total `+%` movement speed from all equipped items (used by Task 3's `playerEffectiveMovementSpeed()`)

- [ ] **Step 1: Write a failing test for movement_speed_percent accumulation**

In `server/internal/game/game_test.go`, add this test (e.g. after `TestMovement`):

```go
func TestMovementSpeedPercentFromGear(t *testing.T) {
	sim := MustNewSim("sess_ms_gear", "seed1", loadRules(t))
	baseline := sim.DerivedStatsView().MovementSpeed

	// Give the player a cave_boots item with +15% movement speed rolled.
	item := &invItem{
		instanceID:     "boots1",
		itemDefID:      "cave_boots",
		ItemTemplateID: "cave_boots",
		rolledStats:    map[string]int{"movement_speed_percent": 15},
	}
	sim.progression.Inventory = append(sim.progression.Inventory, item)
	if err := sim.equipItem("boots1", "boots"); err != nil {
		t.Fatalf("equip boots: %v", err)
	}

	boosted := sim.DerivedStatsView().MovementSpeed
	wantFactor := 1.15
	wantBoosted := baseline * wantFactor
	if math.Abs(boosted-wantBoosted) > 0.001 {
		t.Fatalf("movement_speed with +15%% boots = %.4f, want %.4f (baseline=%.4f)",
			boosted, wantBoosted, baseline)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./internal/game/... -run TestMovementSpeedPercentFromGear -v 2>&1 | tail -5
```

Expected: FAIL — `effectiveCombatStats` has no `MovementSpeedPercent` field yet / `DerivedStatsView` ignores gear percent.

- [ ] **Step 3: Add movement_speed_percent to the item schema enum**

In `shared/rules/item_templates.v0.schema.json`, find the rolleable `stat` enum and append `"movement_speed_percent"`:

```json
"stat": { "type": "string", "enum": [
  "damage_min", "damage_max", "str", "dex", "vit", "magic",
  "all_skills", "max_hp", "max_mana", "armor", "block_percent",
  "attack_speed_percent", "hit_chance", "crit_chance", "evade_chance",
  "health_regen_per_10_seconds", "mana_regen_per_10_seconds",
  "skill_damage_percent", "skill_cooldown_reduction_percent",
  "skill_mana_cost_reduction", "magic_find_percent", "light_radius",
  "hotbar_slots", "inventory_rows",
  "movement_speed_percent"
] }
```

Also add the property definition in the `rolleable_stats` item properties section alongside `attack_speed_percent`:

```json
"movement_speed_percent": { "type": "integer", "minimum": -50, "maximum": 100 }
```

- [ ] **Step 4: Add rolleable_stats entries to cave_boots, cave_ring, cave_amulet**

In `shared/rules/item_templates.v0.json`, update (or add) `rolleable_stats` for these three items:

```json
"cave_boots": {
  "name": "Cave Boots",
  "item_type": "boots",
  "slot": "boots",
  "rolleable_stats": [
    { "stat": "movement_speed_percent", "min": 5, "max": 20, "weight": 3 }
  ]
},
"cave_ring": {
  "name": "Cave Ring",
  "item_type": "ring",
  "slot": "ring",
  "rolleable_stats": [
    ...(existing stats)...,
    { "stat": "movement_speed_percent", "min": 3, "max": 10, "weight": 2 }
  ]
},
"cave_amulet": {
  "name": "Cave Amulet",
  "item_type": "amulet",
  "slot": "amulet",
  "rolleable_stats": [
    ...(existing stats)...,
    { "stat": "movement_speed_percent", "min": 3, "max": 12, "weight": 2 }
  ]
}
```

Keep all existing `rolleable_stats` entries — only append the new `movement_speed_percent` entry.

- [ ] **Step 5: Add MovementSpeedPercent to effectiveCombatStats**

In `server/internal/game/sim.go`, update the struct at line ~235:

```go
type effectiveCombatStats struct {
	DamageMin            float64
	DamageMax            float64
	HitChance            float64
	CritChance           float64
	CritDamage           float64
	EvadeChance          float64
	Armor                float64
	BlockPercent         float64
	AttackSpeed          float64
	AttackIntervalTicks  int
	MaxHP                float64
	MaxMana              float64
	HealthRegenPerSecond float64
	ManaRegenPerSecond   float64
	MagicFindPercent     float64
	LightRadius          float64
	MovementSpeedPercent float64
}
```

- [ ] **Step 6: Accumulate movement_speed_percent in playerEffectiveCombatStatsFor**

In `server/internal/game/sim.go`, inside `playerEffectiveCombatStatsFor` (around line 5663), add a local accumulator and the accumulation logic:

Add the local var near the other percent accumulators (around line 5740 where `itemSpeedPercent` and similar are declared):

```go
var moveSpeedPercent float64
```

Inside the item loop where `attack_speed_percent` is read (around line 5784), add immediately after the last `rolledStats` read in the loop:

```go
if value := rolledStats["movement_speed_percent"]; value != 0 {
    moveSpeedPercent += float64(value)
}
```

In the `effective := effectiveCombatStats{...}` literal (line ~5859), add:

```go
MovementSpeedPercent: moveSpeedPercent,
```

- [ ] **Step 7: Run the test to verify it passes**

```bash
cd server && go test ./internal/game/... -run TestMovementSpeedPercentFromGear -v 2>&1 | tail -5
```

Expected: FAIL still — `DerivedStatsView()` doesn't use `MovementSpeedPercent` yet (fixed in Task 3). This is expected; the accumulation is now in place. Confirm `effectiveCombatStats.MovementSpeedPercent` equals 15 by adding a temporary `t.Logf` if needed, then remove it.

- [ ] **Step 8: Validate shared JSON**

```bash
make validate-shared
```

Expected: all validations pass.

- [ ] **Step 9: Commit**

```bash
git add shared/rules/item_templates.v0.schema.json shared/rules/item_templates.v0.json server/internal/game/sim.go server/internal/game/game_test.go
git commit -m "feat: movement_speed_percent rolleable on boots/rings/amulets; accumulate in effectiveCombatStats"
```

---

### Task 3: Server effective speed pipeline

**Files:**
- Modify: `server/internal/game/sim.go` (~lines 2676-2707, ~5186-5200)
- Modify: `server/internal/game/derived_stats.go`
- Modify: `server/internal/game/game_test.go`

**Interfaces:**
- Consumes: `CharacterClassDef.BaseMovementSpeed` (Task 1), `effectiveCombatStats.MovementSpeedPercent` (Task 2)
- Produces: `DerivedStatsView.MovementSpeed` = final tiles/tick (class_base × DEX_multiplier × gear_factor); `playerMoveSpeed()` = same × slow_multiplier

- [ ] **Step 1: Write failing tests**

In `server/internal/game/game_test.go`, add these two tests:

```go
func TestMovementSpeedDerivedFromClass(t *testing.T) {
	// Rogue should be faster than paladin at equal progression.
	rogueRules := loadRules(t)
	rogueProgression := rogueRules.DefaultCharacterProgressionState()
	rogueProgression.CharacterClass = "rogue"
	rogueProgression.BaseStats = rogueRules.CharacterProgression.Classes["rogue"].BaseStats
	rogueSim, err := NewSimWithWorldProgression("sess_rogue", "r1", rogueRules, "collision_lab", rogueProgression)
	if err != nil {
		t.Fatalf("rogue sim: %v", err)
	}

	paladinRules := loadRules(t)
	paladinProgression := paladinRules.DefaultCharacterProgressionState()
	paladinProgression.CharacterClass = "paladin"
	paladinProgression.BaseStats = paladinRules.CharacterProgression.Classes["paladin"].BaseStats
	paladinSim, err := NewSimWithWorldProgression("sess_paladin", "p1", paladinRules, "collision_lab", paladinProgression)
	if err != nil {
		t.Fatalf("paladin sim: %v", err)
	}

	rogueSpeed := rogueSim.DerivedStatsView().MovementSpeed
	paladinSpeed := paladinSim.DerivedStatsView().MovementSpeed
	if rogueSpeed <= paladinSpeed {
		t.Fatalf("rogue speed %.4f should exceed paladin speed %.4f", rogueSpeed, paladinSpeed)
	}
}

func TestMovementSpeedDexScaling(t *testing.T) {
	sim := MustNewSim("sess_dex", "d1", loadRules(t))
	lowDex := sim.DerivedStatsView().MovementSpeed

	sim.progression.AllocatedStats.Dex += 100
	highDex := sim.DerivedStatsView().MovementSpeed

	if highDex <= lowDex {
		t.Fatalf("100 extra DEX should increase movement speed: got %.4f (was %.4f)", highDex, lowDex)
	}
	// 100 DEX → +10% (per_dex = 0.001), so highDex ≈ lowDex * 1.10 / 1.0
	// (formula goes from base 1.0 to 1.1)
	expectedRatio := 1.10 / (1.0 + float64(sim.progression.AllocatedStats.Dex-100)*0.001)
	actualRatio := highDex / lowDex
	if math.Abs(actualRatio-expectedRatio) > 0.01 {
		t.Fatalf("DEX speed ratio = %.4f, want ~%.4f", actualRatio, expectedRatio)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd server && go test ./internal/game/... -run "TestMovementSpeedDerivedFromClass|TestMovementSpeedDexScaling" -v 2>&1 | tail -10
```

Expected: FAIL — `DerivedStatsView.MovementSpeed` still returns the old formula output (not class-aware).

- [ ] **Step 3: Add classBaseMovementSpeed() helper to sim.go**

In `server/internal/game/sim.go`, add after the existing `playerMoveSpeed()` function (line ~2681):

```go
func (s *Sim) classBaseMovementSpeed() float64 {
	if s.rules != nil {
		if classDef, ok := s.rules.CharacterProgression.Classes[s.progression.CharacterClass]; ok {
			if classDef.BaseMovementSpeed > 0 {
				return classDef.BaseMovementSpeed
			}
		}
		if s.rules.MainConfig.Gameplay.BaseMovementSpeed > 0 {
			return s.rules.MainConfig.Gameplay.BaseMovementSpeed
		}
	}
	return defaultMoveSpeed
}
```

- [ ] **Step 4: Add playerEffectiveMovementSpeed() helper to sim.go**

```go
func (s *Sim) playerEffectiveMovementSpeed() float64 {
	classBase := s.classBaseMovementSpeed()
	character := s.characterDerivedStatsView()
	effective, _ := s.playerEffectiveCombatStats()
	return classBase * character.MovementSpeed * (1 + effective.MovementSpeedPercent/100)
}
```

- [ ] **Step 5: Add playerSlowMultiplier() helper to sim.go**

```go
func (s *Sim) playerSlowMultiplier() float64 {
	slowPercent := 0
	for _, stateKey := range sortedStringKeys(s.skillEffects) {
		effect := s.skillEffects[stateKey]
		if effect.EndsTick <= s.tick {
			continue
		}
		if effect.TargetID != 0 && effect.TargetID != s.playerID {
			continue
		}
		if !containsStringValue(effect.Stats, "movement_speed") || effect.Percent <= slowPercent {
			continue
		}
		slowPercent = effect.Percent
	}
	if slowPercent <= 0 {
		return 1.0
	}
	if slowPercent > 95 {
		slowPercent = 95
	}
	return 1.0 - float64(slowPercent)/100.0
}
```

- [ ] **Step 6: Replace playerMoveSpeed() to use the new pipeline**

Replace the existing `playerMoveSpeed()` function body (lines ~2676-2681):

```go
func (s *Sim) playerMoveSpeed() float64 {
	return s.playerEffectiveMovementSpeed() * s.playerSlowMultiplier()
}
```

Delete the old body (the `if s.rules != nil && s.rules.MainConfig.Gameplay.BaseMovementSpeed > 0` block and `return defaultMoveSpeed`).

- [ ] **Step 7: Update DerivedStatsView() in derived_stats.go**

In `server/internal/game/derived_stats.go`, `DerivedStatsView()` currently sets `MovementSpeed: character.MovementSpeed` (line ~39). Replace that line:

```go
MovementSpeed: s.playerEffectiveMovementSpeed(),
```

The `character` variable is still used for other fields — only `MovementSpeed` changes source.

- [ ] **Step 8: Update movementDistanceForTicks test helper**

In `server/internal/game/game_test.go`, replace the helper at line ~4191:

```go
func movementDistanceForTicks(sim *Sim, ticks int) float64 {
	base := sim.DerivedStatsView().MovementSpeed
	accel := sim.rules.MainConfig.Gameplay.MovementAccelerationSeconds
	minFactor := sim.rules.MainConfig.Gameplay.MovementMinSpeedFactor
	if accel <= 0 {
		return float64(ticks) * base
	}
	wantTicks := int(accel * simulationTickHz)
	if wantTicks < 1 {
		wantTicks = 1
	}
	total := 0.0
	for held := 1; held <= ticks; held++ {
		mult := float64(held) / float64(wantTicks)
		if mult > 1 {
			mult = 1
		}
		if mult < minFactor {
			mult = minFactor
		}
		total += base * mult
	}
	return total
}
```

Update the two call sites that pass `sim.rules` to pass `sim` instead:
- Line ~363: `movementDistanceForTicks(sim.rules, 2)` → `movementDistanceForTicks(sim, 2)`
- Line ~4231: `movementDistanceForTicks(sim.rules, 3)` → `movementDistanceForTicks(sim, 3)`

- [ ] **Step 9: Run all movement and derived-stats tests**

```bash
cd server && go test ./internal/game/... -run "TestMovement|TestDerived|TestMovementSpeed" -v 2>&1 | tail -20
```

Expected: all pass, including `TestMovementSpeedPercentFromGear` (now the full pipeline is wired).

- [ ] **Step 10: Regenerate goldens**

The `movement_speed` field in `shared/golden/character_progression.json` has stale values (e.g. `0.497`). The formula scale changed; golden values will be different.

```bash
cd server && go test ./internal/game/... -run TestGolden -update -v 2>&1 | tail -10
```

Inspect the diff to verify `movement_speed` values changed to sensible per-class multiples:

```bash
git diff shared/golden/character_progression.json | grep movement_speed
```

Expected: values like `0.75`, `0.90`, `0.85`, `0.65` (class base × 1.0 at 0 bonus DEX) instead of `0.497`.

- [ ] **Step 11: Run full test suite**

```bash
cd server && go test ./internal/game/... 2>&1 | tail -5
```

Expected: all pass.

- [ ] **Step 12: Commit**

```bash
git add server/internal/game/sim.go server/internal/game/derived_stats.go server/internal/game/game_test.go shared/golden/character_progression.json
git commit -m "feat: playerMoveSpeed() uses class base × DEX multiplier × gear%; fix player slow"
```

---

### Task 4: Inertia config + direction grace window

**Files:**
- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `client/scripts/main_config_loader.gd`
- Modify: `client/scripts/player_movement_feel.gd`
- Modify: `client/tests/test_player_movement_feel.gd`

**Interfaces:**
- Produces: `movement_min_speed_factor = 0.6` (both server and client read this), `movement_direction_grace_seconds = 0.2` (client reads via `MainConfigLoader`)

- [ ] **Step 1: Update main_config.v0.json**

In `shared/rules/main_config.v0.json`, change:

```json
"movement_min_speed_factor": 0.6,
"movement_direction_grace_seconds": 0.2,
```

(`movement_acceleration_seconds` stays at 2.0.)

- [ ] **Step 2: Update main_config.v0.schema.json**

In `shared/rules/main_config.v0.schema.json`, in the `gameplay` properties block (near `movement_min_speed_factor`), add:

```json
"movement_direction_grace_seconds": { "type": "number", "minimum": 0 }
```

Do NOT add it to the `required` array — it has a default (0.2) and is optional.

- [ ] **Step 3: Add movement_direction_grace_seconds() to main_config_loader.gd**

In `client/scripts/main_config_loader.gd`, add after `movement_min_speed_factor()`:

```gdscript
static func movement_direction_grace_seconds() -> float:
    ensure_loaded()
    return float(gameplay.get("movement_direction_grace_seconds", 0.2))
```

- [ ] **Step 4: Write failing tests for grace window**

In `client/tests/test_player_movement_feel.gd`, add two new tests and register them in `_initialize()`:

```gdscript
func _initialize() -> void:
    MainConfigLoaderScript.ensure_loaded()
    _test_starts_at_min_speed()
    _test_reaches_full_speed_after_accel_window()
    _test_direction_change_resets_ramp()
    _test_small_correction_within_grace_does_not_reset()
    _test_sharp_turn_always_resets()
    if _fail_count == 0:
        print("[gdtest] PASS: test_player_movement_feel (%d passed, %d failed)" % [_pass_count, _fail_count])
        quit(0)
    else:
        print("[gdtest] FAIL: test_player_movement_feel (%d passed, %d failed)" % [_pass_count, _fail_count])
        quit(1)
```

Add the two test functions:

```gdscript
func _test_small_correction_within_grace_does_not_reset() -> void:
    # After holding RIGHT past the grace window, a small correction (dot > 0.5) should NOT reset.
    var feel := PlayerMovementFeelScript.new()
    var grace := MainConfigLoaderScript.movement_direction_grace_seconds()
    # Hold right past grace threshold
    feel.effective_speed(Vector2.RIGHT, grace + 0.05)
    # Small correction: slightly up-right (dot with RIGHT ≈ 0.866, well above 0.5)
    var slight_up_right := Vector2(0.866, 0.5).normalized()
    var speed_after := feel.effective_speed(slight_up_right, 0.01)
    # Speed should NOT have reset to min — should be above min
    var min_speed := MainConfigLoaderScript.base_movement_speed() * MainConfigLoaderScript.movement_min_speed_factor()
    if speed_after <= min_speed + 0.001:
        _fail("small correction should not reset ramp", speed_after, min_speed)
        return
    _pass("small correction within grace does not reset")


func _test_sharp_turn_always_resets() -> void:
    # Even after holding long past grace, a sharp turn (dot < 0.5) resets.
    var feel := PlayerMovementFeelScript.new()
    var accel := MainConfigLoaderScript.movement_acceleration_seconds()
    # Hold right until fully ramped
    for _i in range(15):
        feel.effective_speed(Vector2.RIGHT, accel / 10.0)
    # Sharp turn: UP (90 degrees from RIGHT, dot = 0)
    var after_sharp := feel.effective_speed(Vector2.UP, 0.01)
    var expected_min := MainConfigLoaderScript.base_movement_speed() * MainConfigLoaderScript.movement_min_speed_factor()
    if absf(after_sharp - expected_min) > 0.001:
        _fail("sharp turn should reset ramp", after_sharp, expected_min)
        return
    _pass("sharp turn always resets ramp")
```

- [ ] **Step 5: Run tests to verify new ones fail**

```bash
make client-unit 2>&1 | grep -A3 "test_player_movement_feel"
```

Expected: `_test_small_correction_within_grace_does_not_reset` FAILS (grace logic not implemented yet). `_test_direction_change_resets_ramp` still passes (90° turn resets without grace).

- [ ] **Step 6: Update speed_multiplier() in player_movement_feel.gd for grace window**

Replace the full `speed_multiplier()` function in `client/scripts/player_movement_feel.gd`:

```gdscript
func speed_multiplier(direction: Vector2, delta: float) -> float:
    if direction == Vector2.ZERO:
        on_stop()
        return 0.0
    var normalized := direction.normalized()
    if _last_dir != Vector2.ZERO:
        var dot := normalized.dot(_last_dir)
        var grace := MainConfigLoaderScript.movement_direction_grace_seconds()
        # Small correction (dot >= 0.5) after holding past grace window → no reset.
        # Sharp turn (dot < 0.5) or not yet past grace → reset.
        var is_small_correction := dot >= 0.5 and _hold_seconds >= grace
        if not is_small_correction and dot < 0.7:
            _hold_seconds = 0.0
    _last_dir = normalized
    _hold_seconds += maxf(delta, 0.0)
    var accel_seconds := MainConfigLoaderScript.movement_acceleration_seconds()
    if accel_seconds <= 0.0:
        return 1.0
    var min_factor := MainConfigLoaderScript.movement_min_speed_factor()
    var ramp := clampf(_hold_seconds / accel_seconds, min_factor, 1.0)
    return ramp
```

- [ ] **Step 7: Run all client unit tests**

```bash
make client-unit 2>&1 | grep -E "PASS|FAIL|test_player"
```

Expected: all pass, including both new grace window tests.

- [ ] **Step 8: Validate shared JSON**

```bash
make validate-shared
```

Expected: all validations pass.

- [ ] **Step 9: Commit**

```bash
git add shared/rules/main_config.v0.json shared/rules/main_config.v0.schema.json client/scripts/main_config_loader.gd client/scripts/player_movement_feel.gd client/tests/test_player_movement_feel.gd
git commit -m "feat: inertia start 60%; direction-correction grace window 0.2s"
```

---

### Task 5: Client prediction — feed server-sent derived speed to PlayerMovementFeel

**Files:**
- Modify: `client/scripts/player_movement_feel.gd`
- Modify: `client/scripts/main.gd`

**Interfaces:**
- Consumes: `DerivedStatsView.MovementSpeed` tiles/tick (from server via `character_progression.derived_stats.movement_speed`)
- Produces: `PlayerMovementFeel.effective_speed()` uses the server-sent speed instead of the global base

- [ ] **Step 1: Add _server_speed and set_server_speed() to PlayerMovementFeel**

In `client/scripts/player_movement_feel.gd`, add the state variable and setter:

```gdscript
var _hold_seconds: float = 0.0
var _last_dir := Vector2.ZERO
var _server_speed: float = 0.0  # tiles/tick from server derived_stats; 0 = use config fallback
```

Add the setter method:

```gdscript
func set_server_speed(tiles_per_tick: float) -> void:
    _server_speed = tiles_per_tick
```

- [ ] **Step 2: Update effective_speed() to use _server_speed**

Replace the existing `effective_speed()` function:

```gdscript
func effective_speed(direction: Vector2, delta: float) -> float:
    var base_speed := _server_speed if _server_speed > 0.0 else MainConfigLoaderScript.base_movement_speed()
    return base_speed * speed_multiplier(direction, delta)
```

- [ ] **Step 3: Feed server speed from session_snapshot in main.gd**

In `client/scripts/main.gd`, find line ~1025 where `character_progression` is assigned from `session_snapshot`:

```gdscript
character_progression = p.get("character_progression", {})
```

Add immediately after:

```gdscript
var _ms := float((character_progression.get("derived_stats", {}) as Dictionary).get("movement_speed", 0.0))
if _ms > 0.0:
    _player_movement_feel.set_server_speed(_ms)
```

- [ ] **Step 4: Feed server speed from character_progression_update in main.gd**

In `client/scripts/main.gd`, find line ~1127 the `"character_progression_update"` handler:

```gdscript
"character_progression_update":
    character_progression = c.get("character_progression", {})
    _apply_local_player_class_model()
    ...
```

Add after `character_progression = c.get(...)`:

```gdscript
var _ms := float((character_progression.get("derived_stats", {}) as Dictionary).get("movement_speed", 0.0))
if _ms > 0.0:
    _player_movement_feel.set_server_speed(_ms)
```

- [ ] **Step 5: Update existing client unit tests for set_server_speed**

The existing tests in `test_player_movement_feel.gd` create a fresh `PlayerMovementFeel` without calling `set_server_speed()`, so `_server_speed = 0` and `effective_speed` falls back to `base_movement_speed()`. All existing tests should pass unchanged — verify:

```bash
make client-unit 2>&1 | grep -E "PASS|FAIL|test_player"
```

Expected: all pass.

- [ ] **Step 6: Add a test for set_server_speed in test_player_movement_feel.gd**

Register and add:

```gdscript
func _initialize() -> void:
    ...
    _test_server_speed_overrides_config()
    ...
```

```gdscript
func _test_server_speed_overrides_config() -> void:
    var feel := PlayerMovementFeelScript.new()
    var custom_speed := 1.2  # tiles/tick, higher than config base
    feel.set_server_speed(custom_speed)
    # Hold direction past full ramp
    for _i in range(25):
        feel.effective_speed(Vector2.RIGHT, 0.2)
    var at_full := feel.effective_speed(Vector2.RIGHT, 0.2)
    if absf(at_full - custom_speed) > 0.001:
        _fail("server speed override at full ramp", at_full, custom_speed)
        return
    _pass("set_server_speed overrides config base at full ramp")
```

- [ ] **Step 7: Run all client unit tests**

```bash
make client-unit 2>&1 | grep -E "PASS|FAIL|test_player"
```

Expected: all pass including `_test_server_speed_overrides_config`.

- [ ] **Step 8: Commit**

```bash
git add client/scripts/player_movement_feel.gd client/scripts/main.gd client/tests/test_player_movement_feel.gd
git commit -m "feat: client movement prediction uses server-sent derived movement_speed"
```

---

### Task 6: Stats panel — format movement_speed as tiles/sec

**Files:**
- Modify: `client/scripts/character_stats_panel.gd`

**Interfaces:**
- Consumes: `DerivedStatsView.MovementSpeed` tiles/tick (e.g. 0.75 for default barbarian)
- Produces: UI shows "7.5 t/s" instead of "0.75"

- [ ] **Step 1: Add TILES_PER_TICK_STATS constant**

In `client/scripts/character_stats_panel.gd`, after the existing `WHOLE_PERCENT_STATS` constant (line ~31):

```gdscript
const TILES_PER_TICK_STATS := ["movement_speed"]
```

- [ ] **Step 2: Update _format_stat_value() to handle tiles/tick stats**

In `client/scripts/character_stats_panel.gd`, find `_format_stat_value()` (line ~341):

```gdscript
func _format_stat_value(key: String, value: float) -> String:
    if key in FRACTION_PERCENT_STATS:
        return "%s%%" % _format_number(value * 100.0)
    if key in WHOLE_PERCENT_STATS:
        return "%s%%" % _format_number(value)
    return _format_number(value)
```

Replace with:

```gdscript
func _format_stat_value(key: String, value: float) -> String:
    if key in FRACTION_PERCENT_STATS:
        return "%s%%" % _format_number(value * 100.0)
    if key in WHOLE_PERCENT_STATS:
        return "%s%%" % _format_number(value)
    if key in TILES_PER_TICK_STATS:
        return "%.1f t/s" % (value * 10.0)
    return _format_number(value)
```

A barbarian at default DEX (0.75 tiles/tick) now shows "7.5 t/s". A rogue with fast boots shows e.g. "10.4 t/s".

- [ ] **Step 3: Run full CI**

```bash
make ci 2>&1 | tail -15
```

Expected: all steps pass — Go tests, GDScript unit tests, bot smoke, schema validation, lint-determinism, maintainability.

- [ ] **Step 4: Commit**

```bash
git add client/scripts/character_stats_panel.gd
git commit -m "feat: stats panel shows movement_speed as tiles/sec (e.g. '7.5 t/s')"
```

---

## Self-Review

**Spec coverage:**
- ✅ Hero movement speed is a derivative stat — `playerMoveSpeed() = classBase × DEX_mult × gearFactor × slowMult` (Tasks 1+2+3)
- ✅ Both camera modes: `applyMovement()` and `applyAutoNav()` both call `playerMoveSpeed()` — both get the fix automatically (Task 3)
- ✅ `+% movement speed` rolleable on boots (5–20%), rings (3–10%), amulets (3–12%) (Task 2)
- ✅ Inertia start at 60% (`movement_min_speed_factor = 0.6`) ramping over 2s (Task 4)
- ✅ Direction-correction grace window 0.2s (Task 4)
- ✅ Client prediction uses server-sent `derived_stats.movement_speed` (Task 5)
- ✅ Slow debuff applied at move-time via `playerSlowMultiplier()` (Task 3)
- ✅ Stats panel shows tiles/sec (Task 6)
- ✅ Golden fixtures regenerated via `make regen-golden` (Task 3, Step 10)

**Placeholder scan:** No TBDs. All code blocks are complete.

**Type consistency:**
- `playerEffectiveMovementSpeed() float64` defined in Task 3, called in Task 3 (`playerMoveSpeed`, `DerivedStatsView`).
- `MovementSpeedPercent float64` on `effectiveCombatStats` defined in Task 2 Step 5, used in Task 3 Step 4.
- `set_server_speed(tiles_per_tick: float)` defined in Task 5 Step 1, called in Task 5 Steps 3+4.
- `movement_direction_grace_seconds()` defined in Task 4 Step 3, called in Task 4 Step 6.
- `movementDistanceForTicks(sim *Sim, ticks int)` signature updated in Task 3 Step 8; call sites updated in same step. ✓
