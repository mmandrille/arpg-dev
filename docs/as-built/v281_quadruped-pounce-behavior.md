# v281 As-Built - Quadruped Pounce Behavior

Date: 2026-06-19
Spec: [`docs/specs/v281_spec-quadruped-pounce-behavior.md`](../specs/v281_spec-quadruped-pounce-behavior.md)
Plan: [`docs/plans/v281_2026-06-19-quadruped-pounce-behavior.md`](../plans/v281_2026-06-19-quadruped-pounce-behavior.md)

## Shipped

- Extended monster `attack_style` to include `pounce`.
- Marked `dungeon_wolf` with `attack_style: "pounce"` and a data-owned `attack_range`.
- Updated Go rules validation and `tools/validate_shared.py` so melee `attack_range` is valid only
  for pounce attackers; ordinary melee attackers still reject it.
- Updated `monsterAttackReach` so pounce monsters use their configured pounce reach.
- Extended server attack-style tests to prove wolves can damage from outside ordinary melee reach
  and emit `attack_style: "pounce"`.
- Updated client event routing so pounce-sourced player damage/kill events play the source
  monster's `pounce` one-shot.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/...
GODOT=/opt/homebrew/bin/godot /opt/homebrew/bin/godot --headless --path client --script res://tests/test_animation.gd
GODOT=/opt/homebrew/bin/godot make client-unit
```

All focused checks above passed on 2026-06-19.

`make maintainability` remains blocked by pre-existing unrelated ratchet debt identified during v278;
the user explicitly directed the autoloop to continue and run `$refactor` after all selected slices.

Manual visual verification command:

```bash
make bot-visual scenario=41_monster_visual_catalog
```

## Boundaries

- No wolf HP, damage, cooldown, loot, spawn table, model, skeleton, or animation asset changed.
- No boss behavior changed in this slice.
