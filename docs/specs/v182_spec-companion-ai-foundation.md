# v182 Spec — Companion AI foundation

Status: Complete
Date: 2026-06-15
Codename: `companion-ai-foundation`

## Purpose

Build the first server-owned companion actor foundation so later class skills and mercenaries can reuse the same authoritative follow, target, and melee attack behavior. This slice proves a test companion can exist as a distinct entity, follow its owner, acquire hostile monsters, and damage them without client authority or persistence.

## Non-goals

- No Ranger wolf skill, Sorcerer revive skill, mercenary hiring, persistence, equipment, companion inventory, UI panel, command wheel, XP, loot, or potion behavior.
- No rank scaling or configurable companion limits beyond the minimum lab tuning needed for this proof.
- No client art polish beyond preserving enough protocol identity for the existing renderer and future visual slices.
- No elite minion behavior changes; v186 reuses the foundation after skill companions are proven.

## Acceptance criteria

- A companion entity type is represented in snapshots and deltas separately from `player` and `monster`, with `owner_id`, `monster_def_id`, HP, and max HP.
- The Go sim owns companion follow behavior: when the owner moves away, the companion moves toward a server-selected follow position near that owner.
- The Go sim owns companion target acquisition: a living companion selects a living monster on the same level within its assist radius.
- The Go sim owns deterministic melee attacks: a companion in range damages its target on its attack cadence and emits normal combat events with the companion as source and the monster as target.
- Monsters remain hostile to players only in this slice; companion damage must not produce friendly fire or player-owned client shortcuts.
- A focused protocol bot lab spawns a test companion and proves entity identity, follow movement, and companion damage against a lab monster.

## Scope and likely files

- `server/internal/game/sim.go`, `server/internal/game/types.go`: companion entity kind, entity view fields, follow/target/attack loop, focused helpers.
- `server/internal/game/companion_ai_test.go`: deterministic unit coverage for follow, attack, and protocol identity.
- `shared/protocol/session_snapshot.v8.schema.json`, `shared/protocol/state_delta.v8.schema.json`: allow `companion` entity type.
- `shared/rules/worlds.v0.json`: compact companion lab world with one test companion and one soft target.
- `tools/bot/scenarios/73_companion_ai_foundation.json`: protocol bot proof.
- `tools/bot/run.py`, `tools/bot/runtime_assertions.py`, `tools/bot/test_protocol.py`: only if existing scenario assertions cannot express companion follow/damage.
- Lifecycle docs: `docs/plans/v182_2026-06-15-companion-ai-foundation.md`, `docs/as-built/v182_companion-ai-foundation.md`, `PROGRESS.md`.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run Companion`
- `make bot scenario=73_companion_ai_foundation.json`
- `make ci`

For visual verification after implementation, run:

```bash
make bot-visual scenario=73_companion_ai_foundation.json
```

## Open questions and risks

- Resolved by default: the foundation uses lab-authored companion stats in code/rules only for v182 proof; v185 moves broader HP/damage/attack tuning and limits into companion rank rules.
- Risk: protocol schema type expansion affects both snapshot and delta validators; update examples/tests together if validation exposes stale assumptions.
- Risk: shared monster-target combat events currently assume player sources in some bot helpers; keep new assertions focused and deterministic.

## Godot plugin adoption

Read `docs/researchs/godot-plugins-and-shortcuts.md` on 2026-06-15. Decision: reject plugin adoption for v182 because companion behavior is authoritative server simulation and the existing client can render server entities from snapshots/deltas. Future visual polish can borrow asset patterns when the Ranger wolf visual slice starts.
