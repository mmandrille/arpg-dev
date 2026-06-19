# v280 Plan - Bat Dive Attack Behavior

Status: Complete
Goal: Give `dungeon_bat` a server-authored dive attack style with matching client animation.
Architecture: Add an optional monster `attack_style` rules field, carry that style on monster combat
events, and route dive-sourced player damage events to a bat `dive` one-shot.
Tech stack: Shared JSON rules/schema, Go simulation/rules validation, Godot animation/event handling,
client tests, docs.

## Baseline and shortcut decision

Builds on v277 bat skeleton/wing animation support and v279 model-lunge animation coverage. Adopt the
existing bat runtime GLB, borrow the animation-library lunge pattern, and reject new external assets
or plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.schema.json` | Add `attack_style` enum |
| Modify | `shared/rules/monsters.v0.json` | Mark `dungeon_bat` as `dive` |
| Modify | `server/internal/game/rules.go` | Parse/default/validate monster attack style |
| Modify | `server/internal/game/types.go` | Expose `attack_style` on events |
| Modify | `server/internal/game/unique_survival_effects.go` | Attach direct monster attack style to player combat events |
| Modify | `server/internal/game/game_test.go` | Prove bat dive event metadata and normal attacks omit it |
| Modify | `client/scripts/main.gd` | Play source monster attack clip for player damage events |
| Modify | `client/animations/monster_tiny_flyer_anims.tres` | Add `dive` one-shot |
| Modify | `client/tests/test_animation.gd` | Assert bat `dive` clip motion/wing flap |
| Modify | `docs/specs/v280_spec-bat-dive-attack-behavior.md` | Mark complete |
| Modify | `docs/progress/slice-lifecycle.md` | Add lifecycle row |
| Modify | `PROGRESS.md` | Update current status |
| Add | `docs/as-built/v280_bat-dive-attack-behavior.md` | Record proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Decision:
- [x] Keep implementation scoped to focused files; new server coverage lives in a dedicated
  `monster_attack_style_test.go` file.

Verification:

```bash
make maintainability
```

Expected known issue: `make maintainability` is currently blocked by unrelated pre-existing ratchet
debt and will be paid down in the post-loop `$refactor`.

## Task 1 - Data-driven attack style

Files:
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `server/internal/game/rules.go`

- [x] Add `attack_style` enum with empty/default `melee`.
- [x] Validate `dive` only for melee chase attackers with attack damage/cooldown.
- [x] Assign `dungeon_bat` the `dive` style.

```bash
make validate-shared
cd server && go test ./internal/game/...
```

## Task 2 - Server combat event metadata

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/unique_survival_effects.go`
- Add: `server/internal/game/monster_attack_style_test.go`

- [x] Add optional `attack_style` to combat events.
- [x] Attach `dive` only for direct bat attacks, including miss/kill paths.
- [x] Prove non-dive monster attacks omit the field.

```bash
cd server && go test ./internal/game/...
```

## Task 3 - Client dive animation routing

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/animations/monster_tiny_flyer_anims.tres`
- Modify: `client/tests/test_animation.gd`

- [x] Add a bat `dive` clip that lunges/lifts the model and flaps wings.
- [x] Route `player_damaged`/`player_killed` events with `attack_style: "dive"` to the source
  monster's `dive` clip.
- [x] Keep missing-source events harmless.

```bash
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

## Task 4 - Lifecycle docs

Files:
- Modify: `docs/specs/v280_spec-bat-dive-attack-behavior.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v280_bat-dive-attack-behavior.md`

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
