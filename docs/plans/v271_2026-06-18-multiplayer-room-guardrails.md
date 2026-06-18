# v271 Plan - Multiplayer Room Guardrails

Status: Complete
Goal: Add authoritative per-session tick budget guardrails and overload degradation behavior for
future multiplayer rooms.
Architecture: Realtime owns wall-clock tick budget detection and warning logs. The deterministic
game sim owns the degradation state that affects monster movement. Clients remain presentation-only:
they may predict or interpolate, but never decide AI, navigation, combat, loot, or persistence truth.
Tech stack: Go realtime session loop, Go sim movement LOD, shared navigation rules/schema, protocol
and visual bot probe.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/navigation.v0.json` | Add overload degradation duration |
| Modify | `shared/rules/navigation.v0.schema.json` | Validate overload duration |
| Modify | shared golden navigation mirrors | Keep golden fixtures schema-valid |
| Modify | `server/internal/game/rules.go` / navigation loader | Load degradation duration |
| Modify | `server/internal/game/sim.go` | Store transient overload degradation deadline |
| Modify | `server/internal/game/monster_movement_lod.go` | Apply overload degradation to low-priority monsters |
| Add | `server/internal/game/monster_overload_guardrails_test.go` | Prove server-owned degradation behavior |
| Add | `server/internal/realtime/tick_guardrails.go` | Evaluate tick budget and log warnings |
| Modify | `server/internal/realtime/session_tick.go` | Apply guardrail decisions after each tick |
| Add/Modify | realtime tests | Prove warning payloads and budget decisions |
| Add | `docs/as-built/v271_multiplayer-room-guardrails.md` | Record authority model and proof |
| Modify | progress docs | Advance lifecycle/current status |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go` is at the current ratchet allowance; add only the transient
  field and remove an equivalent blank line if needed.
- [x] `server/internal/game/rules.go` kept the navigation loader extraction.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Extract focused helper/module/test files as part of this slice.
- [ ] Defer extraction with rationale.

Verification:

```bash
make maintainability
```

## Task 1 - Data-driven overload degradation

Files:
- Modify: shared navigation rules/schema/goldens
- Modify: `server/internal/game/navigation_rules.go`
- Modify: `server/internal/game/rules.go`

- [x] Add `monster_overload_degrade_ticks`.
- [x] Validate it as non-negative.
- [x] Keep it short enough to recover quickly but long enough to break over-budget feedback loops.

```bash
make validate-shared
```

## Task 2 - Server-owned degradation state

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/monster_movement_lod.go`
- Add: focused game test

- [x] Add a transient overload-degradation deadline to `Sim`.
- [x] Add a method realtime can call after an over-budget tick.
- [x] During overload degradation, low-priority monsters skip movement/path work; important or nearby
  monsters remain precise.
- [x] Test that degradation affects server movement eligibility only and does not expose any client
  authority path.

```bash
cd server && go test -count=1 ./internal/game
```

## Task 3 - Realtime guardrail warnings

Files:
- Add: `server/internal/realtime/tick_guardrails.go`
- Modify: `server/internal/realtime/session_tick.go`
- Add/Modify: focused realtime tests

- [x] Evaluate the tick budget every session tick.
- [x] Log over-budget warnings with enough bounded fields to diagnose one hot room.
- [x] Apply sim degradation only when the authoritative backend tick exceeds budget and path or
  monster movement counters show room pressure.

```bash
cd server && go test -count=1 ./internal/realtime
```

## Task 4 - Proof and docs

Files:
- Add: `docs/as-built/v271_multiplayer-room-guardrails.md`
- Modify: spec/plan/progress docs

- [x] Run crowded lightning protocol and visual probes.
- [x] Document authority model and overload behavior.
- [x] Mark spec/plan complete.

```bash
ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test -count=1 ./internal/game ./internal/realtime`
- [x] `ARPG_PERF_DEBUG=1 make bot scenario=crowded_lightning_perf_probe`
- [x] `ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=crowded_lightning_perf_probe`
- [x] `make maintainability`
- [x] `HEADLESS=1 make bot-visual scenario=11_combat_feedback`
- [x] `HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle`
- [x] `make ci`

Final full `make ci` passed as the enclosing `$autoloop` batch gate on 2026-06-18.
