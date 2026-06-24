# Movement Overhaul — Design Spec

**Date:** 2026-06-23  
**Slice target:** v332 (next after hero-visibility-lighting)

---

## Problem

`playerMoveSpeed()` ignores character stats entirely — it returns a global flat value from `main_config` for every player regardless of class, DEX, or gear. The `movement_speed` derived stat is computed but never drives actual movement. Inertia starts too low (20% of top speed). Client prediction uses hardcoded base speed, not the character's actual derived value.

---

## Design

### 1. Class base movement speeds

Add `base_movement_speed` (float, tiles/tick at 10 Hz) to each class definition in `character_progression.v0.json`:

| Class | tiles/tick | tiles/sec |
|-------|-----------|-----------|
| Rogue | 0.90 | 9.0 |
| Ranger | 0.85 | 8.5 |
| Barbarian | 0.75 | 7.5 |
| Sorceress | 0.75 | 7.5 |
| Paladin | 0.65 | 6.5 |

`main_config.gameplay.base_movement_speed` is kept as the fallback for unknown/unset classes.

### 2. Movement speed formula — multiplier model

The `movement_speed` formula in `character_progression.v0.json` `derived_stats` changes from a display-scale tile value to a **dimensionless multiplier** (1.0 = no bonus):

```json
"movement_speed": { "type": "linear", "base": 1.0, "per_dex": 0.001, "min": 0.5, "max": 2.0 }
```

- 0 DEX → 1.0× (no bonus)
- 100 DEX → 1.1×
- 200 DEX → 1.2×

### 3. Effective movement speed formula

```
effective_speed_tiles_per_tick =
    class_base_speed
    × movement_speed_multiplier          (from derived_stats formula, DEX-driven)
    × (1 + movement_speed_percent / 100) (accumulated from equipped items)
```

`DerivedStatsView.MovementSpeed` is updated to output this **final tiles/tick value** — the single canonical number both server and client consume.

`playerMoveSpeed()` reads `DerivedStatsView().MovementSpeed`. `applyMovement()` and `applyAutoNav()` both call `playerMoveSpeed()` — both inherit the fix automatically.

### 4. Gear stat: `movement_speed_percent`

New rolleable affix, integer percent, on three slots:

| Slot | Roll range | Min rarity |
|------|-----------|------------|
| Boots | 5–20% | normal |
| Ring | 3–10% | normal |
| Amulet | 3–12% | normal |

Schema: add `"movement_speed_percent": { "type": "integer", "minimum": -50, "maximum": 100 }` to `item_templates.v0.schema.json` rolleable stats enum.

`movement_speed_percent` is accumulated through `playerEffectiveCombatStats()` alongside existing percent stats.

### 5. Inertia feel — min speed factor

`main_config.gameplay.movement_min_speed_factor`: **0.2 → 0.6**  
`main_config.gameplay.movement_acceleration_seconds`: unchanged at 2.0s

Start at 60% of top speed; reach 100% after ~2 seconds of constant movement in the same direction.

### 6. Direction-change grace window

A small direction correction no longer resets the inertia ramp. The dot-product threshold (currently 0.7, meaning >45° change resets) is kept for sharp turns, but a **0.2s grace window** is added: if the player held a direction for < 0.2s before changing, the ramp resets normally; if ≥ 0.2s, small corrections (dot ≥ 0.5) do not reset.

Server: `playerMoveMomentumMultiplier` — add `holdSeconds` tracking alongside `heldTicks` (or derive from ticks) and add the grace condition.  
Client: `PlayerMovementFeel.speed_multiplier` — same logic in GDScript.

### 7. Client prediction — server-derived speed

`PlayerMovementFeel.effective_speed()` currently hardcodes `MainConfigLoader.base_movement_speed()`. Fix:

- Add a `set_server_speed(tiles_per_tick: float)` method to `PlayerMovementFeel`.
- `main.gd` calls it whenever `derived_stats.movement_speed` changes in a state delta.
- `effective_speed()` = `_server_speed × speed_multiplier(inertia_ramp)`, where `_server_speed` defaults to `base_movement_speed` until first server update.

Both isometric and perspective prediction paths call `effective_speed()` — both inherit the fix.

**Bug fixed:** isometric prediction passes `delta` (frame time); perspective passes `ClientConstants.SEND_INTERVAL`. Both now use the same method — the ramp accumulation difference is preserved as-is (minor, acceptable).

### 8. Slow debuff correctness

`unique_survival_effects.go` currently reduces `movement_speed` (the stat) before the base is applied. After this slice, `DerivedStatsView.MovementSpeed` is the final tiles/tick value, so slow percent should be applied as a multiplier on top of that: `effective = playerMoveSpeed() × (1 - slow_percent / 100)`.

The slow is a runtime modifier applied at move-time, not baked into derived stats — this keeps determinism clean.

### 9. Stats panel display

`character_stats_panel.gd` currently shows `movement_speed` as a raw float. Change the formatter for `movement_speed` to render as a percentage of the class base:

```
displayed = round(movement_speed / class_base_speed * 100) → "115%"
```

Or simpler: show `movement_speed` as `"× {:.2f}".format(value)` (e.g. "× 1.15") until a per-class base is available client-side. Final choice left to implementation — either is acceptable.

---

## Files touched

| File | Change |
|------|--------|
| `shared/rules/character_progression.v0.json` | Add `base_movement_speed` per class; update `movement_speed` formula |
| `shared/rules/character_progression.v0.schema.json` | Add `base_movement_speed` to class schema |
| `shared/rules/item_templates.v0.json` | Add `movement_speed_percent` rolleable to boots, rings, amulets |
| `shared/rules/item_templates.v0.schema.json` | Add `movement_speed_percent` to enum + schema |
| `shared/rules/main_config.v0.json` | `movement_min_speed_factor` 0.2 → 0.6 |
| `server/internal/game/sim.go` | `playerMoveSpeed()` uses derived stats; grace window in momentum; slow debuff fix |
| `server/internal/game/derived_stats.go` | `MovementSpeed` output = final tiles/tick |
| `server/internal/game/rules.go` | Parse `base_movement_speed` from class def; accumulate `movement_speed_percent` |
| `client/scripts/player_movement_feel.gd` | `set_server_speed()`; grace window mirror |
| `client/scripts/main.gd` | Call `set_server_speed` on derived_stats update; charge skill speed fix |
| `client/scripts/character_stats_panel.gd` | Format `movement_speed` as multiplier/percent |
| `tools/validate_main_config.py` | Validate new schema fields if needed |

---

## Invariants preserved

- Determinism: no `time.Now()` added; momentum uses tick count only.
- Slow debuff applied at move-time (not baked into derived stats) — replay safe.
- `movement_speed_percent` goes through `playerEffectiveCombatStats()` — same pipeline as `attack_speed_percent`.
- Golden fixtures: `movement_speed` formula output changes scale — if any golden asserts on the old float value it must be updated via `make regen-golden`.
