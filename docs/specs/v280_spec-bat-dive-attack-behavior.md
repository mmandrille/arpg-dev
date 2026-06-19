# v280 Spec: Bat Dive Attack Behavior

Status: Complete
Date: 2026-06-19
Codename: bat-dive-attack-behavior

## Purpose

Make `dungeon_bat` attacks read as a distinct dive instead of a generic melee hit. The behavior must
remain server-authored and data-driven: monster rules declare the attack style, combat events expose
that style, and the Godot client uses the event to play a bat dive one-shot.

## Non-goals

- No bat stat damage/cooldown tuning beyond adding a presentation/behavior style field.
- No external assets, plugins, Blender dependency, or replacement bat model.
- No pounce behavior for quadrupeds; that remains the next selected slice.
- No change to ranged monster projectile behavior.

## Asset Decision

- Adopt: existing committed `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb` and its
  v277 skeleton/wing-flap support.
- Borrow: existing Godot animation-library pattern for model-root lunges from v279 quadruped pounce
  support.
- Reject: adding a new bat asset, DCC-authored animation export, or an external animation plugin.

## Acceptance Criteria

- `shared/rules/monsters.v0.schema.json` accepts a schema-backed `attack_style` enum.
- `server/internal/game/rules.go` validates `attack_style`, defaults empty style to `melee`, and
  rejects `dive` unless the monster is a melee chase attacker with `attack_damage`.
- `shared/rules/monsters.v0.json` marks `dungeon_bat` with `attack_style: "dive"`.
- Player damage/miss/kill events caused by a dive attacker include `attack_style: "dive"` while
  non-dive monster attacks omit the field.
- `monster_tiny_flyer_anims.tres` exposes a `dive` one-shot that lunges/lifts the bat model and
  flaps its wings.
- Client event handling plays the source monster's `dive` clip when a player damage event carries
  `attack_style: "dive"`.
- Focused Go and Godot tests cover rule validation, event metadata, and the client bat dive clip.

## Scope and Likely Files

- Shared data/schema: `shared/rules/monsters.v0.json`,
  `shared/rules/monsters.v0.schema.json`.
- Server: `server/internal/game/rules.go`, `server/internal/game/types.go`,
  `server/internal/game/unique_survival_effects.go`, `server/internal/game/game_test.go`.
- Client: `client/scripts/main.gd`, `client/animations/monster_tiny_flyer_anims.tres`,
  `client/tests/test_animation.gd`.
- Docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, and as-built summary at finish.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd`
- `GODOT=/opt/homebrew/bin/godot make client-unit`

Visual scenario for manual verification:

```bash
make bot-visual scenario=41_monster_visual_catalog
```

## Open Questions and Risks

- No blocking questions.
- Risk: source bat wing bone names are imported from the original asset. Existing v277 tests already
  pin those names; this slice extends that coverage to the dive clip.
