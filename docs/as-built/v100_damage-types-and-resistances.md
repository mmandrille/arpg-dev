# v100 As-built - Damage Types and Resistances

Date: 2026-06-12
Codename: `damage-types-and-resistances`

## What shipped

- Added canonical damage types: `force`, `cold`, `poison`, and `lightning`.
- Added optional `damage_type` on skill and item rules; missing values resolve to `force`.
- Marked `magic_bolt` as `force`, `ice_shard` as `cold`, `poison_stab` as `poison`, and the existing `ligthing` skill id as `lightning`.
- Added monster `resistances` maps where positive values reduce damage and negative values increase damage.
- Configured `dungeon_bat` as 50% lightning resistant and `dungeon_wolf` as 50% lightning weak.
- Added authoritative `damage_type` on combat events.
- Applied monster resistance after armor mitigation and before minimum-damage clamping; full resistance can reduce damage to 0 so immunity-style content can still emit zero-damage ticks.
- Added a compact `damage_types_lab` world and `damage_types_and_resistances` protocol bot scenario.

## Key decisions

- The misspelled skill id `ligthing` remains unchanged for compatibility with existing content and scenarios; only its damage type is canonical `lightning`.
- Damage-type math is server-owned. Client presentation remains unchanged and can ignore the new event field.
- New formula code and tests live in focused `damage_types.go` / `damage_types_test.go` files to avoid growing the central combat test file.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestDamageType|TestResistance'
cd server && go test ./internal/game/...
make bot scenario=damage_types_and_resistances
make maintainability
```

Final close-out also ran `make ci`.

## Deferred

- Undead skeleton enemy and poison immunity are intentionally deferred to v101.
- Type-specific client VFX, damage colors, and UI filtering are deferred.
- Renaming `ligthing` to `lightning` is deferred to a deliberate content migration.
