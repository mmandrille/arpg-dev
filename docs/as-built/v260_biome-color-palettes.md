# v260 As-built - Biome Color Palettes

Date: 2026-06-18

## What shipped

- Added schema-backed `biome_palettes` to `shared/rules/dungeon_generation.v0.json`.
- Defined shallow cave and deep vault depth bands with stable ids and palette-owned ground/wall
  hex colors.
- `GroundWallFactory` loads dungeon-generation rules, selects palettes by absolute dungeon depth,
  and keeps town level visuals on the existing town texture path.
- Procedural dungeon ground and wall texels now use palette-owned low/high/crack/highlight,
  base/mortar/highlight tones.
- `WallRenderer` receives current level context so obstacle wall textures match the active dungeon
  palette.
- Client unit coverage proves shallow and deeper dungeon palettes differ and that wall rendering
  accepts the palette context.

## Proof

```bash
make validate-shared
make client-unit
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```

## Scope limits

- No server gameplay, dungeon layout, monster, loot, fog authority, protocol, replay, persistence,
  or database change shipped.
- No imported art, shader plugin, Godot addon, texture pipeline, asset generation, biome-specific
  monster/loot, obstacles, audio, lighting VFX, or fog rules shipped.
