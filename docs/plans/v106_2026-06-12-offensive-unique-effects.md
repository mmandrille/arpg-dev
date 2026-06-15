# v106 Plan — Offensive Unique Effects

Status: Ready for implementation
Goal: Make `stormbound_echo`, `executioners_mark`, and `hunger_of_the_deep` work in the authoritative sim.
Architecture: Extend the existing unique-effect module with offensive effect state and hook helpers. Direct damage paths call the helpers with source metadata; the helpers read tuning from `shared/rules/unique_effects.v0.json`, use seeded RNG, and emit existing combat/effect events. No protocol schema bump is planned.
Tech stack: Go sim, shared JSON rules, Python bot scenario, lifecycle docs.

## Baseline and shortcut decision

adoption applies because this slice does not add client UI, camera, art, or presentation systems.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/unique_effects.go` | Offensive effect state, triggers, deterministic target search |
| Modify | `server/internal/game/sim.go` | Persist new state and pass attack/skill metadata into hooks |
| Modify | `server/internal/game/unique_effects_test.go` | Focused tests for storm echo, execution mark, hunger stacks |
| Modify | `shared/rules/worlds.v0.json` | Add or reuse a lab world for bot proof |
| Create | `tools/bot/scenarios/54_offensive_unique_effects.json` | Protocol bot proof |
| Modify | `PROGRESS.md` | Slice lifecycle update |
| Create | `docs/as-built/v106_offensive-unique-effects.md` | As-built summary |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: `sim.go` only receives minimal state and hook wiring; behavior lives in `unique_effects.go` and tests live in `unique_effects_test.go`.

Verification:
```bash
make maintainability
```

## Task 1 — Offensive state and hook metadata

Files:
- Modify: `server/internal/game/unique_effects.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 1.1: Add persisted state for execution marks and hunger stacks to `playerState` and `Sim`.
- [x] Step 1.2: Extend the hero-damage hook to know whether the source was a basic attack and which target was damaged.
- [x] Step 1.3: Keep `everburning_wound` behavior unchanged.

```bash
cd server && go test ./internal/game/... -run TestUniqueBurn
```

## Task 2 — Stormbound Echo

Files:
- Modify: `server/internal/game/unique_effects.go`
- Modify: `server/internal/game/unique_effects_test.go`

- [x] Step 2.1: Implement seeded 25% proc roll using `trigger_chance_percent`.
- [x] Step 2.2: Search sorted nearby monsters using `search_radius_tiles`, excluding the primary target.
- [x] Step 2.3: Apply `chain_damage_percent` with `lightning` resistance and emit `monster_damaged` with `skill_id: "stormbound_echo"`.
- [x] Step 2.4: Add tests for trigger and non-trigger from skill damage.

```bash
cd server && go test ./internal/game/... -run TestOffensiveUniqueStormbound
```

## Task 3 — Executioner's Mark

Files:
- Modify: `server/internal/game/unique_effects.go`
- Modify: `server/internal/game/unique_effects_test.go`

- [x] Step 3.1: Mark low-health damaged monsters using `target_hp_percent_threshold`, `mark_duration_seconds`, and `mark_status_id`.
- [x] Step 3.2: On marked monster death, pulse nearby monsters for `pulse_damage_percent_of_marking_hit`.
- [x] Step 3.3: Expire marks deterministically and emit start/end events.
- [x] Step 3.4: Add tests for pulse and expiration.

```bash
cd server && go test ./internal/game/... -run TestOffensiveUniqueExecutionersMark
```

## Task 4 — Hunger of the Deep

Files:
- Modify: `server/internal/game/unique_effects.go`
- Modify: `server/internal/game/unique_effects_test.go`

- [x] Step 4.1: Track same-target stacks per player using catalog stack size and expiration.
- [x] Step 4.2: Apply stack bonus before the triggering hit is finalized, then update stack state after damage.
- [x] Step 4.3: Reset stacks on target change and expiration.
- [x] Step 4.4: Add tests for ramp, target-change reset, and idle expiration.

```bash
cd server && go test ./internal/game/... -run TestOffensiveUniqueHunger
```

## Task 5 — Bot scenario proof

Files:
- Modify: `shared/rules/worlds.v0.json`
- Create: `tools/bot/scenarios/54_offensive_unique_effects.json`

- [x] Step 5.1: Add a compact lab world if existing labs cannot reliably prove an offensive unique.
- [x] Step 5.2: Add a scenario that equips a deterministic unique-effect item and observes the expected event.

```bash
make validate-shared
ARPG_BOT_SCENARIO=offensive_unique_effects VERBOSE=1 make bot
```

## Task 6 — Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v106_offensive-unique-effects.md`
- Modify: `docs/plans/v106_2026-06-12-offensive-unique-effects.md`

- [x] Step 6.1: Mark plan tasks complete.
- [x] Step 6.2: Update `PROGRESS.md` current status, lifecycle index, and deferred work.
- [x] Step 6.3: Write as-built notes.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'TestUniqueBurn|TestOffensiveUnique|TestUniqueEffect'`
- [x] `ARPG_BOT_SCENARIO=offensive_unique_effects VERBOSE=1 make bot`
- [x] `make ci`
