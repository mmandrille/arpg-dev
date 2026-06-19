# v281 Plan - Quadruped Pounce Behavior

Status: Complete
Goal: Make `dungeon_wolf` use a data-driven pounce attack style with extended melee reach and client
pounce playback.
Architecture: Extend the v280 `attack_style` field to include `pounce`; allow melee pounce monsters
to own `attack_range`; use that range in the existing monster attack reach function; route pounce
events to the existing client `pounce` clip.
Tech stack: Shared JSON rules/schema, Go simulation/rules validation, Godot event handling/tests,
docs.

## Baseline and shortcut decision

Builds on v279 pounce animation support and v280 attack-style event metadata. Adopt existing
quadruped assets and v280 routing, reject wolf-specific special cases and external asset work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.schema.json` | Add `pounce` attack style |
| Modify | `shared/rules/monsters.v0.json` | Mark `dungeon_wolf` pounce style/range |
| Modify | `server/internal/game/rules.go` | Validate pounce range ownership |
| Modify | `server/internal/game/sim.go` | Use pounce range in monster reach |
| Modify | `server/internal/game/monster_attack_style_test.go` | Prove validation and pounce reach/event metadata |
| Modify | `client/scripts/main.gd` | Route pounce events to `pounce` clip |
| Modify | `docs/specs/v281_spec-quadruped-pounce-behavior.md` | Mark complete |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Modify | `PROGRESS.md` | Update current status |
| Add | `docs/as-built/v281_quadruped-pounce-behavior.md` | Record proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Decision:
- [x] Keep new server coverage in the focused v280 test file, still under the target.

Verification:

```bash
make maintainability
```

Expected known issue: `make maintainability` is currently blocked by unrelated pre-existing ratchet
debt and will be paid down in the post-loop `$refactor`.

## Task 1 - Pounce rules and reach

Files:
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`

- [x] Add `pounce` to `attack_style`.
- [x] Allow `attack_range` only for ranged attacks or melee pounce attacks.
- [x] Use pounce `attack_range` as monster melee reach.
- [x] Assign `dungeon_wolf` pounce style/range.

```bash
make validate-shared
cd server && go test ./internal/game/...
```

## Task 2 - Pounce proof and client routing

Files:
- Modify: `server/internal/game/monster_attack_style_test.go`
- Modify: `client/scripts/main.gd`

- [x] Prove wolves can damage from pounce range with `attack_style: "pounce"`.
- [x] Prove ordinary melee attackers still cannot own `attack_range`.
- [x] Route `attack_style: "pounce"` to the source monster's `pounce` one-shot.

```bash
cd server && go test ./internal/game/...
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

## Task 3 - Lifecycle docs

Files:
- Modify: `docs/specs/v281_spec-quadruped-pounce-behavior.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v281_quadruped-pounce-behavior.md`

- [x] Mark the spec complete.
- [x] Record focused verification and batch-CI-pending status.
- [x] Add as-built notes and lifecycle row.

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- [x] `GODOT=/opt/homebrew/bin/godot make client-unit`
- [x] `make maintainability` deferred with known unrelated ratchet debt

Manual visual command:

```bash
make bot-visual scenario=41_monster_visual_catalog
```

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
