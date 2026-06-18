# v260 Spec - Biome Color Palettes

Status: Complete
Date: 2026-06-18
Codename: biome-color-palettes

## Purpose

Add data-driven dungeon biome color palettes so dungeon depths can visually shift floor and wall
tints without changing gameplay generation. The Godot client should read palette bands from shared
dungeon-generation rules and apply them to generated floor and wall textures by depth.

## Non-goals

- No server gameplay, dungeon layout, monster, loot, fog authority, protocol, replay, persistence,
  or database change.
- No imported art, shader plugin, Godot addon, texture pipeline, or asset generation change.
- No biome-specific monster/loot tables, obstacles, sounds, lighting VFX, or fog rules.

## Acceptance Criteria

- `shared/rules/dungeon_generation.v0.json` defines schema-backed `biome_palettes` depth bands.
- Palette bands include stable ids and hex colors for ground low/high/crack/highlight and wall
  base/mortar/highlight tones.
- Shared validation accepts the new palette data and rejects structurally invalid palette rows.
- `GroundWallFactory` selects the correct palette by dungeon depth and keeps town visuals unchanged.
- Dungeon ground and wall textures use palette-owned colors, with different colors for shallow and
  deeper bands.
- Existing wall rendering and generated texture tests remain green.

## Scope and Likely Files

- Shared rules:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
- Client presentation:
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/wall_renderer.gd`
  - `client/scripts/main.gd`
- Tests:
  - `client/tests/test_item_visuals.gd`
  - `client/tests/test_factories.gd`
- Docs:
  - `docs/plans/v260_2026-06-18-biome-color-palettes.md`
  - `docs/as-built/v260_biome-color-palettes.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported texture packs, shader plugins, and Godot
addons. Borrow existing code-native procedural ground/wall texture generation.

## Test and Bot Proof

- `make validate-shared`
- `make client-unit`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```

## Open Questions and Risks

- No required questions.
- Risk: palette values can become accidental test locks. Tests should assert palette selection and
  structural color differences, not exact final texel values beyond data ownership.
