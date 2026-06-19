# v280 As-Built - Bat Dive Attack Behavior

Date: 2026-06-19
Spec: [`docs/specs/v280_spec-bat-dive-attack-behavior.md`](../specs/v280_spec-bat-dive-attack-behavior.md)
Plan: [`docs/plans/v280_2026-06-19-bat-dive-attack-behavior.md`](../plans/v280_2026-06-19-bat-dive-attack-behavior.md)

## Shipped

- Added schema-backed monster `attack_style` rules with default `melee` and validated `dive` as a
  melee chase attacker style.
- Marked `dungeon_bat` with `attack_style: "dive"`.
- Added optional `attack_style` to authoritative combat events and attached it to direct bat attacks
  while leaving normal monster attacks omitted.
- Added a focused `monster_attack_style_test.go` server test file for validation and event metadata.
- Added a bat `dive` one-shot to `monster_tiny_flyer_anims.tres`; it lifts/lunges the model and
  flares both wing bones.
- Updated client event handling so player damage/kill events with `attack_style: "dive"` play the
  source monster's `dive` clip when that source is present.
- Extended the Godot animation smoke to require and inspect the bat `dive` clip.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/...
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

All focused checks above passed on 2026-06-19.

Post-loop `$refactor` paid down the ratchet debt and the selected batch passed full `make ci` on
2026-06-19.

Manual visual verification command:

```bash
make bot-visual scenario=41_monster_visual_catalog
```

## Boundaries

- No bat damage, HP, movement speed, cooldown, loot, spawn table, or asset source changed.
- No quadruped pounce behavior changed in this slice; it remains next in the selected queue.
