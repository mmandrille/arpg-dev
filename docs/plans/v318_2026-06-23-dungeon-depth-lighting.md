# v318 Plan - Dungeon Depth Lighting

Status: Ready for implementation
Goal: Apply depth/biome mood lighting on the client using data-backed palette fields.
Architecture: Extend existing `biome_palettes` depth bands; keep fog/LoS and server authority unchanged.
Tech stack: Godot 4 GDScript, shared dungeon generation rules JSON.

## Baseline and Shortcut Decision

Builds on v317. Asset/plugin decision: adopt code-native scene lighting; reject external assets/plugins.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add lighting fields per biome palette |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Optional lighting properties on `biome_palette` |
| Create | `client/scripts/dungeon_depth_lighting.gd` | Resolve and apply lighting profiles |
| Modify | `client/scripts/main.gd` | Own scene lights and refresh on level changes |
| Create | `client/tests/test_dungeon_depth_lighting.gd` | Headless lighting profile proof |
| Modify | `scripts/client_smoke.sh` | Register new unit test |
| Create | `docs/as-built/v318_dungeon-depth-lighting.md` | Completion proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `main.gd` (grandfathered) — lighting hook only; no net growth beyond allowance

Decision:
- [x] New focused `dungeon_depth_lighting.gd` module; minimal `main.gd` hook.

## Tasks

- [x] Add palette lighting fields to shared dungeon generation rules + schema.
- [x] Implement `DungeonDepthLighting` profile resolver/applier.
- [x] Wire `main.gd` scene lights and level-change refresh.
- [x] Add headless unit test and smoke registration.
- [x] Update lifecycle docs and as-built proof.

## Verification

- [x] `make validate-shared`
- [x] `godot --headless --path client --script res://tests/test_dungeon_depth_lighting.gd`
- [x] `make client-unit`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
