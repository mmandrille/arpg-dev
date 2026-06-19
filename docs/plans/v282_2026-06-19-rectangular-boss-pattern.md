# v282 Plan - Rectangular Boss Pattern

Status: Complete
Goal: Add a new Cave Warden `crystal_wall` rectangle pattern with authoritative hit detection and
client rectangular telegraph rendering.
Architecture: Use existing boss phase metadata. Extend aimed shape support to rectangle, add a data
pattern, and cover server/client shape behavior with focused tests.
Tech stack: Shared boss rules, Go sim/rules validation, Godot boss visual controller, docs.

## Baseline and shortcut decision

Builds on existing boss pattern scheduling and the schema's already-declared `rectangle` vocabulary.
Adopt the current telegraph system, borrow line hit-test geometry, and reject external VFX assets.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_patterns.v0.json` | Add `crystal_wall` pattern |
| Modify | `shared/rules/boss_templates.v0.json` | Add pattern to Cave Warden deck |
| Modify | `server/internal/game/boss_pattern_rules.go` | Require rectangle width |
| Modify | `server/internal/game/boss_patterns.go` | Add rectangle aim/hit support |
| Add | `server/internal/game/boss_rectangle_pattern_test.go` | Prove pattern, validation, hit predicate |
| Modify | `client/scripts/boss_visuals_controller.gd` | Render rectangle marker mesh |
| Modify | `PROGRESS.md` | Update current status |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Add | `docs/as-built/v282_rectangular-boss-pattern.md` | Record proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Decision:
- [x] Keep new server coverage in a focused new test file under the target.

Verification:

```bash
make maintainability
```

Expected known issue: `make maintainability` is currently blocked by unrelated pre-existing ratchet
debt and will be paid down in the post-loop `$refactor`.

## Task 1 - Pattern data and validation

Files:
- Modify: `shared/rules/boss_patterns.v0.json`
- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `server/internal/game/boss_pattern_rules.go`

- [x] Add `crystal_wall` rectangle telegraph/active/recovery data.
- [x] Append `crystal_wall` after the already bot-proven deck entries.
- [x] Require positive width for rectangle telegraph phases.

```bash
make validate-shared
cd server && go test ./internal/game/...
```

## Task 2 - Rectangle hit and client marker

Files:
- Modify: `server/internal/game/boss_patterns.go`
- Add: `server/internal/game/boss_rectangle_pattern_test.go`
- Modify: `client/scripts/boss_visuals_controller.gd`

- [x] Lock aim for rectangle phases.
- [x] Reuse forward range/width hit math for rectangle active phases.
- [x] Render rectangle telegraph markers with a rectangular mesh.
- [x] Add tests for inside/outside rectangle cases and marker mesh shape.

```bash
cd server && go test ./internal/game/...
GODOT=/opt/homebrew/bin/godot make client-unit
```

## Task 3 - Lifecycle docs

Files:
- Modify: `docs/specs/v282_spec-rectangular-boss-pattern.md`
- Modify: `docs/plans/v282_2026-06-19-rectangular-boss-pattern.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v282_rectangular-boss-pattern.md`

- [x] Mark the spec and plan complete.
- [x] Record focused verification and batch-CI-pending status.
- [x] Add lifecycle/as-built notes.

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `GODOT=/opt/homebrew/bin/godot make client-unit`
- [x] `make maintainability` deferred with known unrelated ratchet debt

Manual visual command:

```bash
make bot-visual scenario=24_boss_floor_gate
```

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
