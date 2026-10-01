# v505 — Dungeon depth mood

- **Status:** Complete; final cache-aware Performance process-p95 gate and combined batch `make ci` passed (see as-built for the earlier miss and limits).
- **Date:** 2026-10-01
- **Codename:** `dungeon-depth-mood`
- **Area:** Client presentation / graphics
- **Depends on:** v498 dungeon light readability, already integrated at the assigned batch base.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D1, D7–D9; [ADR-0001](../adr/0001-technology-stack.md) D2, D4, D6.

## Purpose

Extend the existing shallow-cave, sundered-halls, and deep-vault look beyond the floor and key/ambient lights. Give each existing depth palette its own configurable atmospheric render fog and torch/flame accent so descent changes the whole room mood. Keep the player, active enemies, ground loot, boss warnings, and explored/unexplored boundary readable from the isometric player camera.

At the v505 base, `dungeon_generation.v0.json` already selects three depth palettes and owns their surface and key/ambient light colors. `render_presentation.v0.json` and `dungeon_torch_presentation.v0.json` each apply one dungeon-wide profile. v498 has already tuned the general contrast, fog-of-war interplay, and visible torch pools. This slice adds depth-linked atmospheric and torch variation without retuning that general readability baseline by default.

## Non-goals

- Changing server visibility, line of sight, combat, loot, or replay behavior.
- Changing the gameplay fog-of-war overlay, its explored/unexplored masks, torch reveal radius, or torch placement/count.
- Replacing the current render baseline, quality tiers, or default graphics setting.
- Changing dungeon geometry, biome selection/depth ranges, surface materials, class/town/loot presentation, or boss telegraph design.
- Importing assets, adding a shader pack/plugin, or adding another post-processing pass.

## Acceptance criteria

- [ ] The three existing palette identities (`shallow_cave`, `sundered_halls`, `deep_vault`) resolve to distinct, schema-validated atmospheric fog and torch/flame settings. Selection uses the already-resolved palette identity; the implementation does not duplicate depth cutoffs in client code or add arbitrary tuning literals.
- [ ] Existing `DungeonDepthLighting` palette values continue to supply key and ambient light. Any adjustment to those values is justified by paired live captures and stays in the existing shared catalog.
- [ ] Balanced applies the selected render-fog profile. Performance retains its existing fog-off quality behavior while still applying the selected depth lighting and torch accents. Both tiers keep the same gameplay fog-of-war visibility and interaction cues.
- [ ] Real-renderer before/after captures from the live isometric player camera cover depth 1, depth 5, and depth 8 in both quality tiers, with the HUD and fog-of-war overlay active. The review confirms a readable player/enemy silhouette, findable ground loot, recognizable active warning, visible warm torch pools, and meaningful unexplored darkness. Use the existing shallow loot, depth-5 boss, and depth-8 capture routes where they provide these subjects.
- [ ] Focused tests prove all palette IDs resolve, fog/torch settings reach their runtime nodes, invalid/missing palette settings fail validation or use the documented safe base profile, and quality-tier gating is unchanged. Shared JSON schema validation passes.
- [ ] A matched live-client comparison against the assigned base records frame-interval p50/p95, process p95, draw calls, and first-spawn cost for the same fixture, renderer, machine, resolution, and quality tier. Set the budget before tuning; inherit v498's limits of no more than `max(0.5 ms, 5%)` added frame/process p95, `max(10 calls, 5%)` added draw calls, and `max(20 ms, 10%)` added first-spawn cost. If the control spread exceeds a limit, revise and record the budget before tuning. Do not claim cross-hardware smoothness from this host.

## Likely surfaces

- **Presentation data:** a focused `shared/assets/dungeon_depth_mood.v0.json` catalog keyed by the existing dungeon palette IDs, with its matching schema. It owns render-fog and torch/flame mood overrides; existing biome palette data remains the source for geometry colors and key/ambient light.
- **Client:** `client/scripts/dungeon_depth_lighting.gd` (expose the selected palette ID), `client/scripts/scene_lighting_rig.gd` and `client/scripts/render_environment_presentation.gd` (apply only the selected fog override), plus `client/scripts/dungeon_torch_lights.gd` and/or its loader (apply torch accents without changing placement or count). Prefer a focused loader over adding logic to `main.gd`.
- **Tests:** add a focused depth-mood loader/application test and extend existing `client/tests/test_render_environment_presentation.gd`, `test_dungeon_depth_lighting.gd`, or torch tests only where their current assertions do not cover the contract.
- **Bot/docs:** reuse the v498 real-camera routes if sufficient; otherwise add only a focused scenario for a missing depth/tier/subject. Update the Render baseline row in `docs/CODEMAP.md` if file ownership changes.
- **No server, protocol, golden, or gameplay-rule changes are expected.**

## Focused verification and visual proof

Run shared validation, the focused Godot tests, `make client-unit`, and `make maintainability`. The real renderer is required for the visual gate; headless `gl_compatibility` does not prove Forward+ fog, glow, or torch appearance.

Existing player-camera capture routes to reuse:

```bash
make bot-visual scenario=dungeon_light_readability_shallow HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_balanced HEADLESS=0
make bot-visual scenario=dungeon_light_readability_boss_performance HEADLESS=0
make bot-visual scenario=dungeon_light_readability_deep HEADLESS=0
make bot-visual scenario=dungeon_light_readability_engaged_enemy HEADLESS=0
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_frame_pacing_probe HEADLESS=0
```

The bot scenarios prove their scripted route and captured scene only. Review the saved images at native resolution, retain renderer/tier/viewport metadata, and describe the exact depth coverage of the performance fixture.

## Asset and plugin decision

- **Borrow** the in-repo KayKit dungeon torches, the three existing biome palettes, and current render/torch presentation paths.
- **Reject** external art, shader packs, post-processing packages, and Godot plugins; the slice needs data variants and existing lighting only.

## Open questions

None block planning. Preserve the established warm shallow/central palette and cool deep-vault direction unless paired live captures show that a specific cue is unreadable.
