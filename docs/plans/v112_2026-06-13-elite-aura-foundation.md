# v112 Plan — Elite Aura Foundation

Status: Ready for implementation
Goal: Add a data-driven generated-pack leader aura that increases same-pack follower monster damage
while the leader is alive and nearby.
Architecture: Keep the feature inside authoritative Go sim combat. Preserve generated pack metadata
on live entities, validate aura tuning from shared rules, and avoid protocol/client changes.
Tech stack: Shared JSON/schema, Go rules, Go sim, focused Go tests, lifecycle docs.

## Baseline and shortcut decision

Builds on v79 generated elite pack roles and the existing dungeon pack metadata in
`server/internal/game/dungeon_gen.go`. No client UI, camera, inventory presentation, or art work is
in scope, so the Godot plugin adoption checklist is not required.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Define the default elite command aura tuning. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate the aura object. |
| Modify | `server/internal/game/rules.go` | Load and validate aura rules. |
| Modify | `server/internal/game/sim.go` | Preserve generated pack metadata on live monster entities. |
| Create | `server/internal/game/elite_aura.go` | Apply the live leader aura to follower damage. |
| Create | `server/internal/game/elite_aura_test.go` | Prove aura application/rejection and live metadata transfer. |
| Modify | `PROGRESS.md` | Mark v112 complete during finish. |
| Create | `docs/as-built/v112_elite-aura-foundation.md` | Record shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/rules.go`

Decision:
- [x] Keep sim/rules edits surgical and put new behavior/tests in focused files.
- [x] Run `make maintainability` before final CI.

Verification:
```bash
make maintainability
```

## Task 1 — Aura rules

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `monster_placement.elite_aura` with id, radius, and damage bonus percent.
- [x] Step 1.2: Add schema and Go validation for positive radius and bounded non-negative damage
  bonus.
```bash
make validate-shared
```

## Task 2 — Live pack metadata

Files:
- Modify: `server/internal/game/sim.go`
- Create/modify: `server/internal/game/elite_aura_test.go`

- [x] Step 2.1: Add pack id and leader fields to live monster entities.
- [x] Step 2.2: Populate live generated monsters from dungeon-generation metadata.
- [x] Step 2.3: Test that generated dungeon monsters preserve pack metadata after spawning.
```bash
cd server && go test ./internal/game -run TestGeneratedDungeonMonstersPreservePackMetadata -count=1
```

## Task 3 — Aura combat effect

Files:
- Create: `server/internal/game/elite_aura.go`
- Modify: `server/internal/game/unique_survival_effects.go`
- Create/modify: `server/internal/game/elite_aura_test.go`

- [x] Step 3.1: Apply aura damage bonus before monster-to-player combat resolution.
- [x] Step 3.2: Require living same-pack leader within configured radius.
- [x] Step 3.3: Exclude leaders, no-pack monsters, other-pack monsters, dead-leader packs, and
  out-of-radius followers.
```bash
cd server && go test ./internal/game -run TestEliteAura -count=1
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/plans/v112_2026-06-13-elite-aura-foundation.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v112_elite-aura-foundation.md`

- [x] Step 4.1: Mark plan tasks complete as they pass.
- [x] Step 4.2: Update `PROGRESS.md` latest slice, next slice, lifecycle row, and recently closed
  note.
- [x] Step 4.3: Add the v112 as-built note.
```bash
make ci
```

## Bot proof deferral

No protocol bot scenario is required in v112 because aura state is not visible through snapshots or
events, and deterministic damage/radius behavior is fully owned by focused Go tests. A future
client/protocol aura-readability slice should add a bot scenario once it exposes aura state or
presentation.

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestEliteAura|TestGeneratedDungeonMonstersPreservePackMetadata' -count=1`
- [x] `make maintainability`
- [x] `make test-go`
- [x] `make ci`
