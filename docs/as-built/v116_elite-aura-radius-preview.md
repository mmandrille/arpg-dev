# v116 As-Built: Elite Aura Radius Preview

**Status:** Complete on `main`

## What shipped

- Monster entity views now include optional generated-pack metadata:
  `monster_pack_id` and `monster_pack_leader`.
- Current v8 snapshot/delta schemas accept the optional pack metadata.
- The Godot client renders a display-only elite command radius preview around visible pack leaders
  when at least one same-pack follower is server-marked with `elite_command`.
- The preview radius is loaded from `shared/rules/dungeon_generation.v0.json`.
- Client debug state exposes pack metadata, preview presence, and radius value for bot assertions.
- Added client bot scenario `37_elite_aura_radius_preview`.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run TestEliteAura -count=1`
- `make client-unit`
- `make bot-client scenario=37_elite_aura_radius_preview`
- `make maintainability`
- `make ci`

## Deferred

Production aura VFX/audio, monster nameplates, aura tooltips, minimap markers, additional aura
types, aura roll tables, and combat tuning remain deferred.

