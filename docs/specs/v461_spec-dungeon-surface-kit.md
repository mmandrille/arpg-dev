# v461 Spec - Dungeon Surface Kit

Status: Draft
Date: 2026-07-09
Codename: dungeon-surface-kit
Baseline: v460 `leveled-potions`

## Purpose

Make dungeon floors feel cleaner and more intentional by upgrading the existing code-native surface
pipeline for walls, floors, water, and holes while removing rubble/debris clutter from generated
dungeon obstacles. The slice should make dungeon spaces read as stronger stone spaces with clearer
surface contrast, better biome separation, and less visual noise.

This slice stays inside the existing procedural Godot presentation path and shared dungeon
generation data. It does not introduce external textures, imported assets, plugins, or a new asset
pipeline.

## Non-goals

- No protocol, replay, persistence, combat, fog-of-war rules, pathfinding algorithm, or movement
  changes.
- No new dungeon props, trim meshes, decorative debris sets, or imported art packs.
- No boss-floor-specific art pass.
- No dungeon audio pass.
- No full biome/content expansion beyond the current depth bands.
- No removal of rock or column obstacles; only rubble/debris is removed from the dungeon obstacle
  mix in this slice.

## Acceptance Criteria

- Dungeon biome palettes define three clearly different depth bands for surface presentation:
  entry, mid-depth, and deep vault.
- Dungeon ground materials keep the existing code-native procedural texture path but improve
  contrast/readability between floor, walls, water, and holes for each biome band.
- Water and hole surfaces remain plane-based authoritative blockers, but their materials read more
  distinctly against surrounding floor tiles through palette/data tuning and client-only surface
  treatment.
- Generated dungeon obstacle rules no longer select `rubble`; generated floors use walls, rocks,
  and columns only.
- The obstacle-variety lab and relevant client/tests no longer expect rubble/debris visuals.
- The slice does not add any new decorative trim/debris layer on top of the dungeon presentation.
- Focused client tests prove biome palette differentiation and the absence of rubble generation in
  the updated rules/test surfaces.
- A focused visual client-bot scenario proves the refreshed dungeon surface kit is active on a
  generated dungeon floor.

## Scope and Likely Files

- Shared rules/data:
  - `shared/rules/dungeon_generation.v0.json`
  - `shared/rules/dungeon_generation.v0.schema.json`
- Server gameplay/tests:
  - `server/internal/game/dungeon_obstacle_variety.go`
  - `server/internal/game/dungeon_obstacle_variety_test.go`
- Client presentation:
  - `client/scripts/ground_wall_factory.gd`
  - `client/scripts/wall_renderer.gd`
  - `client/tests/test_factories.gd`
- Bot proof:
  - `tools/bot/scenarios/client/99_dungeon_surface_kit.json`
  - possibly `tools/bot/scenarios/101_obstacle_variety_pack.json` if the rubble expectation is
    still part of scenario proof
- Docs:
  - `docs/plans/v461_2026-07-09-dungeon-surface-kit.md`
  - `docs/as-built/v461_dungeon-surface-kit.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

### Asset / plugin decision

- Adopt: existing procedural biome palette and `GroundWallFactory` / `WallRenderer` path.
- Borrow: current depth-lighting, torch, and wall/floor test scaffolding where useful.
- Reject: external texture packs, shader packs, decorators, trim kits, or Godot addons.

## Test and Bot Proof

```bash
make validate-shared
cd server && go test ./internal/game -run TestSolidObstacleKindWeights
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=99_dungeon_surface_kit
```

Manual visual proof:

```bash
make bot-visual scenario=99_dungeon_surface_kit
```

## Open Questions and Risks

- Risk: removing rubble from generated rules can break obstacle-variety proof surfaces if labs or
  assertions still require it. Update those expectations together instead of leaving half-removed
  debt.
- Risk: stronger palette contrast can fight torch/fog presentation. Keep the slice focused on base
  surfaces and verify composition with existing dungeon torch and wall/floor presentation tests.
- Default decision: treat the user's "delete debris/trim" request as removing dungeon rubble/debris
  presentation and avoiding new decorative trim in this slice, not as a repo-wide removal of every
  non-dungeon trim usage.
