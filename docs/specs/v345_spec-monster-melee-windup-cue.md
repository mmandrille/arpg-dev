# v345 — Monster melee windup cue

Codename: monster-melee-windup-cue

## Goal

Give ordinary melee monsters a short data-driven windup before damage lands, with an in-world ring cue on the client.

## Contract

- Optional `attack_windup_ticks` on melee-style monsters in `shared/rules/monsters.v0.json`.
- Server emits `monster_attack_windup` with `source_entity_id`, `target_entity_id`, `remaining_ticks`, `total_ticks`, then applies damage after the windup elapses.
- Client shows a fading ground ring and plays the monster attack clip when the windup event arrives.

## Verification

- `cd server && go test ./internal/game/... -run TestDungeonMobMeleeWindupDelaysDamage`
- `godot --headless --path client --script res://tests/test_look_and_feel_polish.gd`
- `make validate-shared`
