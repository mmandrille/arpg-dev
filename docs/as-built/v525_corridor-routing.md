# v525 As Built — Corridor Routing

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).

## Handoff state

- Base commit: `876d02c872db4b465e92f460266f83ea6245c3c8`.
- Dependency state: v522, v523, and v524 were supplied as uncommitted coordinator changes on that base. No dependency commit SHA was supplied.
- Worker handoff: implementation and focused checks passed without a worker commit. The coordinator has since completed integration, lifecycle registration, `/finish`, and combined `make ci`.
- `docs/progress/slice-lifecycle.md` and `PROGRESS.md` were left unchanged.

## Implementation

- Replaced the PCG room-layout scalar corridor width with a non-empty, unique `corridor_widths` rule list. Runtime validation bounds every choice by player diameter and the narrowest room-shape cell.
- Retained the selected topology edges in generated-level internals and builds one deterministic, orthogonal route per edge. Seeded choices select widths and candidate ordering; routes try both L bends and bounded room/floor-edge doglegs, rejecting movement-clearance collisions against existing walls and room perimeters.
- Carried each selected width through its two room doors and emitted corresponding perimeter openings. Final reachability validation also checks each selected edge route against the completed wall set, including later obstacles.
- Added deterministic detour, edge-to-route, configured-width, and invalid-width assertions. Updated the affected geometry goldens for obstacles, stairs, teleporters, and guarded chest placement.
- Updated `docs/CODEMAP.md` for the new route implementation and test files.

The implementation preserves the v522 room-first spawn/anchor contract, v523 shape masks and inner arena boundaries, v524 topology constraints and edge set, and live-grid reachability.

## Verification

| Command | Result |
|---|---|
| `cd server && go test ./internal/game -run 'RoomCorridorRouting|RoomCorridor|Topology|DungeonObstacles|Reachability|StairsGolden|Teleporter|GuardedChestGenerationGolden|BossFloorGeneration' -count=1` | PASS (50.243s) |
| `cd server && go test ./internal/game -run '^TestRoomCorridorRouting_' -count=1` | PASS (2.409s), including blocked-first-choice fallback, bounded impossible-layout failure, and width validation |
| `make validate-shared` | PASS (2,269 shared checks; CODEMAP valid) |
| `make bot scenario=reachable_dungeon_obstacles` | PASS |
| `make bot-visual scenario=wall_floor_dungeon_rollout` | PASS (1/1 visible client scenario) |
| `make bot-visual scenario=28_reachable_dungeon_obstacles` | PASS (protocol recording and visual replay of 80 envelopes; exit code 0) |
| `cd server && go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` | PASS (77 grandfathered sites in 22 files) |
| `make maintainability` | PASS (file-size, extraction-coupling, and progress-dashboard checks) |
| `git diff --check` | PASS |

The visual scenario manifest is `.artifacts/bot-runs/20261002T051101Z-visual.json`. It records the scenario outcome and camera position; this run did not retain a PNG screenshot. The replay verifies that the client consumed the recorded generated-floor state and placed room dressing on floors −1 and −2. It does not prove every corridor's movement clearance or performance. Godot emitted resource-leak warnings during shutdown after the replay completed; the process and bot command exited successfully.

The first pre-refresh obstacle golden run failed with 34 generated walls against the prior 37-wall snapshot. The routed layout intentionally changes wall geometry; after refreshing the deterministic geometry snapshots, the complete focused suite passed. `make ci` and `make ci-full` were not run because combined integration gates are coordinator-owned.

## Changed and untracked paths

### v525 implementation and evidence

- `docs/CODEMAP.md`
- `docs/plans/v525_2026-10-01-corridor-routing.md`
- `docs/specs/v525_spec-corridor-routing.md`
- `docs/as-built/v525_corridor-routing.md`
- `server/internal/game/dungeon_gen.go`
- `server/internal/game/dungeon_generated_types.go`
- `server/internal/game/dungeon_profiles.go`
- `server/internal/game/dungeon_room_corridors.go`
- `server/internal/game/dungeon_room_corridor_routing.go` (new)
- `server/internal/game/dungeon_room_corridor_routing_test.go` (new)
- `server/internal/game/dungeon_room_perimeter_walls.go`
- `shared/golden/dungeon_obstacles.json`
- `shared/golden/dungeon_stairs.json`
- `shared/golden/dungeon_teleporters.json`
- `shared/golden/guarded_chest_generation.json`
- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/dungeon_generation.v0.schema.json`

### Preserved coordinator-supplied v522–v524 overlay

At worker handoff these paths shared the same dirty worktree with v525; the coordinator has since compared the complete path set and integrated it. They are not v525-only changes:

- `.maintainability/determinism-baseline.tsv`
- `server/internal/game/dungeon_reachability_grid.go`
- `server/internal/game/dungeon_room_anchor_fallback.go` (deleted)
- `server/internal/game/dungeon_room_anchor_rooms.go` (deleted)
- `server/internal/game/dungeon_room_anchor_placement.go` (new)
- `server/internal/game/dungeon_room_corridor_sweep_test.go`
- `server/internal/game/dungeon_room_layout.go`
- `server/internal/game/dungeon_room_shapes.go` (new)
- `server/internal/game/dungeon_room_shapes_test.go` (new)
- `server/internal/game/dungeon_room_spawn.go` (new)
- `server/internal/game/dungeon_room_topology.go` (new)
- `server/internal/game/dungeon_room_topology_test.go` (new)
- `server/internal/game/wall_floor_lab_nav_test.go`
- `shared/rules/worlds.v0.json`
- `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`
- `docs/as-built/v522_rooms-first.md`
- `docs/as-built/v523_room-shapes.md`
- `docs/as-built/v524_topology-motifs.md`
- `docs/plans/v522_2026-10-01-rooms-first.md`
- `docs/plans/v523_2026-10-01-room-shapes.md`
- `docs/plans/v524_2026-10-01-topology-motifs.md`
- `docs/specs/v522_spec-rooms-first.md`
- `docs/specs/v523_spec-room-shapes.md`
- `docs/specs/v524_spec-topology-motifs.md`

## Ignored evidence and generated artifacts

- `.artifacts/bot-runs/20261002T051101Z-visual.json` — retained visual-replay manifest.
- `.venv/` and `arpg_tools.egg-info/` — created by `make validate-shared`.
- `client/.godot/` and ignored Godot `.uid` files under `client/scripts/`, `client/tests/`, and `client/shaders/` — created/updated during visual client import and replay. Generated tracked `.glb.import` edits were restored because they were unrelated importer metadata churn.
- `tools/__pycache__/` and `tools/bot/__pycache__/` — Python validation/runtime caches.

## Coordinator integration closeout

The coordinator compared the child worktree path-by-path against the integrated v522–v524 changes, including shared generation flow, room geometry, rules, and goldens. Combined `make ci` passed, and lifecycle/progress records were updated during `/finish`.

### Exact ignored paths at handoff

- `.artifacts/`
- `.venv/`
- `arpg_tools.egg-info/`
- `client/.godot/`
- `client/scripts/armor_look.gd.uid`
- `client/scripts/attack_contact_trace.gd.uid`
- `client/scripts/attack_presentation_loader.gd.uid`
- `client/scripts/bishop_loot_debug_panel.gd.uid`
- `client/scripts/blacksmith_upgrade_chance.gd.uid`
- `client/scripts/boss_bar_frame.gd.uid`
- `client/scripts/boss_intro_banner.gd.uid`
- `client/scripts/boss_lane_marker.gd.uid`
- `client/scripts/boss_presentation_loader.gd.uid`
- `client/scripts/boss_presentation_mounter.gd.uid`
- `client/scripts/bot_action_formatter.gd.uid`
- `client/scripts/bot_boss_intro_banner_assertions.gd.uid`
- `client/scripts/bot_combat_contact_assertions.gd.uid`
- `client/scripts/bot_connection_recovery_assertions.gd.uid`
- `client/scripts/bot_frame_capture.gd.uid`
- `client/scripts/bot_reconnect_proof_actions.gd.uid`
- `client/scripts/bot_remote_player_assertions.gd.uid`
- `client/scripts/bot_steward_hunt_assertions.gd.uid`
- `client/scripts/bot_torch_viewpoint.gd.uid`
- `client/scripts/bot_training_damage_log_actions.gd.uid`
- `client/scripts/bot_transport_delay.gd.uid`
- `client/scripts/character_stat_groups.gd.uid`
- `client/scripts/character_stats_breakdown.gd.uid`
- `client/scripts/character_stats_header.gd.uid`
- `client/scripts/combat_breakdown_format.gd.uid`
- `client/scripts/combat_vfx.gd.uid`
- `client/scripts/connection_overlay.gd.uid`
- `client/scripts/connection_overlay_bridge.gd.uid`
- `client/scripts/connection_recovery.gd.uid`
- `client/scripts/connection_recovery_runtime.gd.uid`
- `client/scripts/dungeon_depth_mood_loader.gd.uid`
- `client/scripts/dungeon_kit_floor.gd.uid`
- `client/scripts/dungeon_kit_presentation_loader.gd.uid`
- `client/scripts/dungeon_kit_props.gd.uid`
- `client/scripts/dungeon_kit_wall_builder.gd.uid`
- `client/scripts/dungeon_room_dressing.gd.uid`
- `client/scripts/dungeon_surface_detail_loader.gd.uid`
- `client/scripts/dungeon_surface_detail_presentation.gd.uid`
- `client/scripts/dungeon_wall_corner_presentation.gd.uid`
- `client/scripts/equipment_display_loader.gd.uid`
- `client/scripts/fog_shadow_polygon_layers.gd.uid`
- `client/scripts/gear_sockets_loader.gd.uid`
- `client/scripts/hero_corpse_visual.gd.uid`
- `client/scripts/hud_globe.gd.uid`
- `client/scripts/hud_layout.gd.uid`
- `client/scripts/hud_style.gd.uid`
- `client/scripts/inventory_item_pricing.gd.uid`
- `client/scripts/inventory_panel_styles.gd.uid`
- `client/scripts/inventory_tooltip_content.gd.uid`
- `client/scripts/inventory_wallet_delta_runtime.gd.uid`
- `client/scripts/item_family_icon_preview.gd.uid`
- `client/scripts/item_icons_catalog.gd.uid`
- `client/scripts/item_model_thumbnail_cache.gd.uid`
- `client/scripts/item_requirement_views.gd.uid`
- `client/scripts/item_tooltip_stat_sections.gd.uid`
- `client/scripts/item_visuals_loader.gd.uid`
- `client/scripts/kit_hero_clips.gd.uid`
- `client/scripts/kit_monster_presentation_loader.gd.uid`
- `client/scripts/kit_monster_visual.gd.uid`
- `client/scripts/kit_piece_library.gd.uid`
- `client/scripts/kit_stairs.gd.uid`
- `client/scripts/live_targeting_trace.gd.uid`
- `client/scripts/loot_quest_badge_shapes.gd.uid`
- `client/scripts/model_detail_tint.gd.uid`
- `client/scripts/model_tint.gd.uid`
- `client/scripts/monster_anim_driver.gd.uid`
- `client/scripts/monster_anim_variants.gd.uid`
- `client/scripts/monster_death_presentation_loader.gd.uid`
- `client/scripts/monster_idle_variation.gd.uid`
- `client/scripts/monster_variant_look.gd.uid`
- `client/scripts/monster_variant_resolver.gd.uid`
- `client/scripts/paper_doll_backdrop.gd.uid`
- `client/scripts/paper_doll_layout.gd.uid`
- `client/scripts/potion_icon_label.gd.uid`
- `client/scripts/projectile_flight_presentation.gd.uid`
- `client/scripts/quest_steward_panel.gd.uid`
- `client/scripts/quest_steward_presentation.gd.uid`
- `client/scripts/ranger_affinity_damage.gd.uid`
- `client/scripts/rarity_cue_loader.gd.uid`
- `client/scripts/rarity_cue_presenter.gd.uid`
- `client/scripts/remote_player_class_sync.gd.uid`
- `client/scripts/render_environment_presentation.gd.uid`
- `client/scripts/render_presentation_loader.gd.uid`
- `client/scripts/scene_lighting_rig.gd.uid`
- `client/scripts/showme/class_silhouette_play_camera_capture.gd.uid`
- `client/scripts/showme/showme_character_screen_capture.gd.uid`
- `client/scripts/showme/showme_gear_matrix_capture.gd.uid`
- `client/scripts/showme/showme_hud_capture.gd.uid`
- `client/scripts/showme/showme_item_asset_capture.gd.uid`
- `client/scripts/showme/showme_item_icon_capture.gd.uid`
- `client/scripts/showme/showme_item_icons_capture.gd.uid`
- `client/scripts/showme/showme_monster_variants_capture.gd.uid`
- `client/scripts/showme/showme_rarity_cues_capture.gd.uid`
- `client/scripts/showme/showme_rarity_ui_fixtures.gd.uid`
- `client/scripts/showme/showme_skeleton_capture.gd.uid`
- `client/scripts/showme/showme_skill_icon_capture.gd.uid`
- `client/scripts/showme/showme_town_play_capture.gd.uid`
- `client/scripts/showme/visual_capture.gd.uid`
- `client/scripts/skill_aim_input.gd.uid`
- `client/scripts/skill_bonus_tooltip.gd.uid`
- `client/scripts/skill_mechanic_tooltip.gd.uid`
- `client/scripts/skill_next_rank_tooltip.gd.uid`
- `client/scripts/skill_rank_scaling.gd.uid`
- `client/scripts/skill_synergy_tooltip.gd.uid`
- `client/scripts/skill_tree_connectors.gd.uid`
- `client/scripts/skill_tree_layout.gd.uid`
- `client/scripts/skill_tree_styles.gd.uid`
- `client/scripts/skill_visual_capture.gd.uid`
- `client/scripts/skill_visual_capture_runtime.gd.uid`
- `client/scripts/steward_hunt_banner.gd.uid`
- `client/scripts/surface_material_loader.gd.uid`
- `client/scripts/surface_material_room_capture.gd.uid`
- `client/scripts/town_ambient_life.gd.uid`
- `client/scripts/town_dressing.gd.uid`
- `client/scripts/town_ground_blend.gd.uid`
- `client/scripts/town_ground_detail.gd.uid`
- `client/scripts/town_nature_landmarks.gd.uid`
- `client/scripts/training_damage_log_bridge.gd.uid`
- `client/scripts/training_damage_log_panel.gd.uid`
- `client/scripts/training_doll_visual.gd.uid`
- `client/scripts/ui_theme.gd.uid`
- `client/scripts/wall_occlusion_fade.gd.uid`
- `client/scripts/wall_occlusion_presentation_loader.gd.uid`
- `client/scripts/wall_occlusion_runtime.gd.uid`
- `client/scripts/weapon_range_tooltip.gd.uid`
- `client/shaders/town_ground_blend.gdshader.uid`
- `client/shaders/vfx_soft_glow.gdshader.uid`
- `client/tests/equipped_gear_fit_probe.gd.uid`
- `client/tests/test_armor_look.gd.uid`
- `client/tests/test_attack_contact_trace.gd.uid`
- `client/tests/test_blacksmith_upgrade_chance.gd.uid`
- `client/tests/test_boss_arena_presence.gd.uid`
- `client/tests/test_boss_intro_banner.gd.uid`
- `client/tests/test_boss_lane_marker.gd.uid`
- `client/tests/test_boss_presentation_loader.gd.uid`
- `client/tests/test_bot_engaged_subject.gd.uid`
- `client/tests/test_bot_frame_capture.gd.uid`
- `client/tests/test_bot_torch_viewpoint.gd.uid`
- `client/tests/test_bot_transport_delay.gd.uid`
- `client/tests/test_character_stat_groups.gd.uid`
- `client/tests/test_character_stats_header.gd.uid`
- `client/tests/test_combat_vfx.gd.uid`
- `client/tests/test_connection_recovery.gd.uid`
- `client/tests/test_connection_recovery_runtime.gd.uid`
- `client/tests/test_death_pose_ownership.gd.uid`
- `client/tests/test_dungeon_depth_mood.gd.uid`
- `client/tests/test_dungeon_kit.gd.uid`
- `client/tests/test_dungeon_kit_props.gd.uid`
- `client/tests/test_dungeon_room_dressing.gd.uid`
- `client/tests/test_first_spawn_trace.gd.uid`
- `client/tests/test_golden_skill_progression.gd.uid`
- `client/tests/test_hud_globe.gd.uid`
- `client/tests/test_hud_layout.gd.uid`
- `client/tests/test_hud_style.gd.uid`
- `client/tests/test_inventory_tooltip_content.gd.uid`
- `client/tests/test_inventory_wallet_delta_runtime.gd.uid`
- `client/tests/test_item_icon_drawer.gd.uid`
- `client/tests/test_item_model_thumbnail_cache.gd.uid`
- `client/tests/test_item_requirement_views.gd.uid`
- `client/tests/test_item_visuals_loader.gd.uid`
- `client/tests/test_kit_monsters.gd.uid`
- `client/tests/test_kit_stairs.gd.uid`
- `client/tests/test_live_targeting_trace.gd.uid`
- `client/tests/test_material_wallet_panel.gd.uid`
- `client/tests/test_monster_anim_variants.gd.uid`
- `client/tests/test_monster_death_presentation.gd.uid`
- `client/tests/test_monster_variant_looks.gd.uid`
- `client/tests/test_paper_doll_backdrop.gd.uid`
- `client/tests/test_player_health_bar.gd.uid`
- `client/tests/test_potion_icon_label.gd.uid`
- `client/tests/test_projectile_flight_presentation.gd.uid`
- `client/tests/test_quest_steward_presentation.gd.uid`
- `client/tests/test_ranger_affinity_damage.gd.uid`
- `client/tests/test_rarity_cues.gd.uid`
- `client/tests/test_remote_player_class.gd.uid`
- `client/tests/test_render_environment_presentation.gd.uid`
- `client/tests/test_showme_load.gd.uid`
- `client/tests/test_skill_bonus_tooltip.gd.uid`
- `client/tests/test_skill_mechanic_tooltip.gd.uid`
- `client/tests/test_skill_rank_scaling.gd.uid`
- `client/tests/test_skill_synergy_tooltip.gd.uid`
- `client/tests/test_skill_tree_layout.gd.uid`
- `client/tests/test_town_ambient_life.gd.uid`
- `client/tests/test_town_dressing.gd.uid`
- `client/tests/test_town_ground_detail.gd.uid`
- `client/tests/test_town_nature_landmarks.gd.uid`
- `client/tests/test_training_damage_log_panel.gd.uid`
- `client/tests/test_ui_theme.gd.uid`
- `client/tests/test_wall_occlusion_fade.gd.uid`
- `client/tests/test_weapon_range_tooltip.gd.uid`
- `tools/__pycache__/`
- `tools/bot/__pycache__/`
