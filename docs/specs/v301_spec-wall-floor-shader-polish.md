# v301 Spec - Wall/Floor Shader Polish

Status: Complete
Date: 2026-06-19
Codename: wall-floor-shader-polish

## Purpose

Improve dungeon floor and wall readability with a client-only Godot material pass. The current
client already generates cave floor and wall textures from in-repo palettes; this slice extends that
pipeline with code-native detail/normal maps and material settings so stone floors and walls catch
light differently without changing movement, pathfinding, fog, combat, protocol, or shared rules.

This is the final selected World Detail/Navigation autoloop slice. It intentionally follows the
v295-v300 gameplay obstacle work with presentation polish only.

## Non-goals

- No shared protocol, schema, Go server, dungeon-generation, fog, navigation, movement, collision,
  pathfinding, or bot-runner behavior changes.
- No imported textures, models, shader packages, Godot plugins, external assets, or new asset
  pipeline.
- No production art pass, biome-retuning pass, lighting rebalance, camera change, or UI change.
- No changes to water, holes, rocks, columns, rubble, doors, fog masks, or line-of-sight semantics
  beyond preserving their current rendering behavior.

## Acceptance Criteria

- Dungeon ground materials remain `StandardMaterial3D` and keep their generated albedo texture.
- Dungeon ground materials include generated normal/detail texture data, enabled normal mapping, and
  deterministic material settings.
- Dungeon wall materials remain `StandardMaterial3D`, keep source-based generated/perimeter tinting,
  and include generated normal/detail texture data.
- Town ground keeps its existing simple material behavior so the polish is scoped to dungeon
  floors/walls.
- Generated material texture caches remain deterministic and palette-aware.
- Existing item-visual and wall-rendering tests continue to pass without changing gameplay-visible
  assumptions.
- A focused Godot unit test proves dungeon floor/wall normal maps and material flags are present.
- A headless client bot scenario exercises generated dungeon floor/wall rendering after stairs-down
  transition.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/wall_renderer.gd`
  - `client/tests/test_factories.gd`
- Bot proof:
  - `tools/bot/scenarios/client/78_wall_floor_shader_polish.json`
- Docs:
  - `docs/plans/v301_2026-06-19-wall-floor-shader-polish.md`
  - `docs/as-built/v301_wall-floor-shader-polish.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject imported assets, shader packs, Godot addons, and external texture
libraries. Borrow the existing `GroundWallFactory`, `WallRenderer`, procedural biome palettes, and
`dungeon_wall_rendering` client bot flow.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_factories.gd`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=wall_floor_shader_polish ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=wall_floor_shader_polish
```

Final selected-batch proof after v301 commits:

- `make ci`

## Open Questions and Risks

- No required user questions. Default: use Godot `StandardMaterial3D` normal/detail maps rather than
  replacing the material with `ShaderMaterial`, because existing visual tests and presentation code
  depend on standard material properties.
- Risk: material polish can become a hidden gameplay change if it touches world rules or server
  state. Keep this slice entirely client-side plus docs/bot scenario.
- Risk: `client/tests/test_item_visuals.gd` is already over the maintainability line limit. Do not
  edit it for this slice; add new assertions to `client/tests/test_factories.gd`.
