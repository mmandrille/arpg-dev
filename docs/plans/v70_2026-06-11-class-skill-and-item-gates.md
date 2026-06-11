# v70 Plan — Class Skill and Item Gates

Status: Complete
Goal: Enforce class identity for skills and class-required weapons.
Architecture: The sim carries `CharacterClass` inside progression state and snapshots. Shared skill/item rules declare class requirements; server handlers reject cross-class skill spend/cast and equipment. UI remains display-only until v71.
Tech stack: Shared JSON/schema, Go sim/store/realtime/replay, Go tests.

## Baseline and Shortcut Decision

Builds on v69. Godot plugin decision: reject for this slice because no client UI/art is in scope.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Map skills to class ids |
| Modify | `shared/rules/items.v0.json`, schema | Add class-required fixed weapons |
| Modify | `tools/validate_shared.py` | Validate class ids in skills/items |
| Add | `server/migrations/0018_session_start_character_class.sql` | Snapshot class identity |
| Modify | `server/internal/game/*` | Carry class state and enforce gates |
| Modify | `server/internal/store/*`, realtime/replay | Persist/load class in snapshots |
| Modify | `server/internal/game/game_test.go` | Gameplay proof |
| Modify | docs lifecycle files | Close-out |

## Task 1 — Shared Rules

- [x] Change skill `class` values to real class ids.
- [x] Add `class_required` to fixed class weapons.
- [x] Validate class ids in skills/items.

```bash
make validate-shared
```

## Task 2 — Server Authority

- [x] Carry character class in progression state/view and session-start snapshots.
- [x] Reject cross-class skill spend/cast.
- [x] Reject cross-class weapon equip.

```bash
cd server && go test ./internal/game ./internal/store ./internal/http ./internal/realtime ./internal/replay
```

## Task 3 — Tests and Docs

- [x] Add Go tests for class skill and weapon gates.
- [x] Add as-built and progress updates.

```bash
make test-go
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game ./internal/store ./internal/http ./internal/realtime ./internal/replay`
- [x] `make test-go`
- [x] `make ci`
