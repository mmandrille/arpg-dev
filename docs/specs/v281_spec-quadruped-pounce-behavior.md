# v281 Spec: Quadruped Pounce Behavior

Status: Complete
Date: 2026-06-19
Codename: quadruped-pounce-behavior

## Purpose

Turn the pounce-ready quadruped presentation from v279 into a server-authored `dungeon_wolf`
behavior. Wolves should use a data-driven pounce attack style with longer melee reach, emit
authoritative `attack_style: "pounce"` combat events, and trigger the existing client `pounce`
animation.

## Non-goals

- No new quadruped asset, skeleton, or animation library; v279 already supplied them.
- No wolf damage, HP, loot, spawn-table, or cooldown retuning beyond a schema-backed pounce reach.
- No changes to ranged projectile monsters.
- No boss pattern or summon behavior changes.

## Asset Decision

- Adopt: existing v279 `monster_quadruped.tscn` and `monster_quadruped_fox_anims.tres` `pounce`
  clip.
- Borrow: v280 `attack_style` event metadata and client source-monster clip routing.
- Reject: adding an external animation system, DCC-authored export, or wolf-specific code path.

## Acceptance Criteria

- `attack_style` accepts `pounce` in the shared monster schema and Go rules validation.
- `dungeon_wolf` declares `attack_style: "pounce"` and a positive `attack_range` owned by shared
  rules data.
- Rules validation allows `attack_range` for melee `pounce` attackers and still rejects it for
  ordinary melee attackers.
- Server attack range calculation uses the pounce range for melee pounce monsters.
- A wolf can damage the player from outside ordinary unarmed reach but inside its pounce range, and
  the combat event carries `attack_style: "pounce"`.
- Client event handling maps `attack_style: "pounce"` to the source monster's `pounce` one-shot.
- Existing quadruped pounce animation smoke remains green.

## Scope and Likely Files

- Shared data/schema: `shared/rules/monsters.v0.json`,
  `shared/rules/monsters.v0.schema.json`.
- Server: `server/internal/game/rules.go`, `server/internal/game/sim.go`,
  `server/internal/game/monster_attack_style_test.go`.
- Client: `client/scripts/main.gd`.
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
- Risk: longer wolf reach can change dungeon difficulty. This slice keeps damage/cooldown unchanged
  and scopes the reach to the pounce style so future tuning remains data-owned.
