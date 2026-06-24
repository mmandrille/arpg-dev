# ADR-0015: Player Movement Speed Formula

**Status:** Accepted (v332, 2026-06-24)
**Context:** v332 movement-overhaul batch

---

## Context

Before v332 the effective player movement speed was a single constant pulled directly from
`main_config.v0.json → gameplay.base_movement_speed`. There was no per-class variation,
no stat scaling, and no momentum ramp-up. v332 extracted the helpers into `movement.go` and
introduced a layered formula to support per-class base speeds and stat-influenced scaling.

---

## Decision

Player movement speed is a multiplicative pipeline evaluated each tick:

```
playerMoveSpeed() =
    classBaseMovementSpeed()
    × character.MovementSpeed            (from derived-stats formula, default 1.0)
    × (1 + effective.MovementSpeedPercent / 100)
    × playerSlowMultiplier()
```

### Layer 1 — class base (`classBaseMovementSpeed`)

Priority order:
1. `character_progression.classes[charClass].base_movement_speed` (if > 0)
2. `main_config.gameplay.base_movement_speed` (if > 0)
3. `defaultMoveSpeed` constant (0.75) as a last resort

This allows each class to have a distinct base, while preserving the old single-constant
path for worlds or tests that don't set a class.

### Layer 2 — DEX scaling (`character.MovementSpeed`)

Evaluated via the `derived_stats["movement_speed"]` progression formula:

```json
{ "type": "linear", "base": 1.0, "per_dex": 0.001, "min": 0.5, "max": 2.0 }
```

For a fresh character with no DEX investment this resolves to exactly **1.0**, leaving the
class base unchanged. DEX gear or stat investments can push the multiplier above 1.

### Layer 3 — gear percent (`effective.MovementSpeedPercent`)

Accumulated from rolleable `movement_speed_percent` stats on boots, rings, and amulets.
At launch this is additive across all sources and capped implicitly by equipment variety.

### Layer 4 — slow multiplier (`playerSlowMultiplier`)

Applied last to preserve the semantic contract: skills that "slow the player" reduce the
final speed regardless of how high the other layers are. Capped at 95% slow to prevent
complete immobilization.

### Momentum ramp (key-held movement only)

`playerMoveMomentumMultiplier` applies **only to `move_intent` (key-held) movement**, not to
`autoNav` path-following. The ramp uses:
- `main_config.gameplay.movement_acceleration_seconds` — ticks to reach full speed
- `main_config.gameplay.movement_min_speed_factor` — floor fraction (default 0.6)

autoNav path steps are normalized to unit vectors then scaled by `playerMoveSpeed()`, so
diagonal and cardinal navigation steps cover the same distance per tick.

---

## Consequences

### Replay determinism

`playerMoveSpeed()` is evaluated every tick inside `applyAutoNav` and `applyMovement`.
Both call the same deterministic formula stack (no `time.Now()`, no bare `rand`). Replays
that use the same rules and seed will produce identical positions.

### Bot timing

Bot scenarios and `appendMoveToAndAdvance` test helpers depend on how many ticks the player
needs to traverse a distance. A rule change to `base_movement_speed`, `per_dex`, or
`movement_acceleration_seconds` will shift arrival times and may require updating tick
budgets in tests and bot scenario `max_elapsed_s` values.

### Tests that use `move_intent` directly

Because `move_intent` applies the momentum ramp, the first tick covers only
`classBase × 0.6 = 0.45` units (at default 0.75 class base). Tests that assume
`0.75/tick` for the first step will fail after this formula landed — see the v332 test
update commits.

### Future changes

- Any change to `derived_stats["movement_speed"]` formula affects all classes uniformly.
- Adding `base_movement_speed` to a new class requires checking existing bot scenarios for
  timing regressions.
- A future `movement_speed_percent` cap or diminishing-returns formula would insert a new
  layer between layers 3 and 4.
