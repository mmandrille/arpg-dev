# v498 Plan — Dungeon light and readability

- **Status:** Complete; combined `make ci` passed in 10m33s.
- **Spec:** [v498 dungeon light and readability](../specs/v498_spec-dungeon-light-readability.md)
- **Goal:** Give shallow and deep dungeons distinct atmosphere while keeping combat and loot readable from the live player camera.
- **Architecture:** Tune the existing shared presentation catalogs first. `DungeonDepthLighting` supplies biome key and ambient light, `SceneLightingRig` applies the render context, and the torch and fog systems remain separate client presentation layers. Preserve server visibility and combat authority. Add a small, repeatable live capture path only if the existing bot runner cannot collect the required frames.
- **Tech stack:** shared JSON catalogs and schemas, Godot 4.7 client, Python/Godot client bot tooling. No protocol or Go gameplay change is planned.

## Baseline and shortcut decision

`PROGRESS.md` lists v493 as complete and v494–v500 as **To do**. v498 depends on the v497 live frame-pacing result. Begin final tuning only after v497 records its hardware, renderer, dense fixture, quality tier, raw client samples, draw calls, and first-spawn cost. Re-run its control on the same checkout used for v498, since v494–v497 may change the scene.

**Adopt / borrow / reject:** Borrow the in-repo KayKit dungeon, biome palettes, key-light/render, torch, and fog catalogs. Reject new asset packs, shader packs, Godot plugins, and an extra post-processing pass. Keep the default quality tier.

Current-state windowed Godot captures from this v493-based checkout (`a2e8933c`, Godot 4.7.2, project renderer `forward_plus`, Balanced setting, isometric player camera, HUD and fog of war active) are in `.artifacts/v498-baseline/`:

| Frame | What it proves | Limit |
|-------|----------------|-------|
| `shallow-isometric-balanced.png` | Depth 1 spawn, player silhouette, lit floor, fog boundary | No nearby loot or active warning in frame |
| `deep-isometric-balanced.png` | Depth 5 warm palette, boss/enemy and active line/circle warning | Boss arena is not the depth-7+ palette |
| `deep-vault-isometric-balanced.png` | Depth 8 cool palette, player and fog gradient | No loot or active warning in frame |

These are visual observations, not v497 performance controls. The attempted temporary generated-floor loot entry did not spawn (`loot=0`); a starter-weapon drop also failed because the bot character had no equipped main-hand item. Neither failure changed tracked files. The final capture fixture must put actual loot on the dungeon floor through a proven gameplay path and assert its presence before saving an image.

## Pre-tuning budget and capture matrix

Use the same machine, display resolution, Godot build, Forward+ renderer, seed, depth, camera route, entity count, and quality tier for each matched control/result pair. Keep fog of war and HUD enabled. Run at least three valid live-client repetitions per tier and retain raw logs plus the summary; compare the medians of per-run p95 frame and process time, median draw calls, and the first entity-spawn cost. Exclude only the documented warm-up samples, using the v497 report's definition. Record vsync and compositor capping.

Provisional permitted deltas relative to a fresh v497 control on that same topology:

| Measure | Limit for v498 result |
|---------|-----------------------|
| Frame p95 and process p95 | At most `max(0.5 ms, 5%)` above control |
| Median draw calls | At most `max(10 calls, 5%)` above control |
| First-spawn process cost | At most `max(20 ms, 10%)` above control |

Set these limits **before** tuning. If three unchanged control runs already vary more than a limit, document that variance and revise the threshold in this plan before changing presentation data. A quality-tier cost saving cannot compensate for a visibility failure.

Capture depth 1 (shallow cave), depth 5 (warm halls with boss warning), and depth 8 (cool deep vault), at Balanced and Performance. For each tier include: torch-lit and unlit ground in one route, a player/enemy silhouette comparison, a deterministic ground loot drop, a warning before impact, and unexplored fog. Pin camera positions or bot steps; take paired before/after frames at the same simulation moment and resolution. Judge warning recognition and loot visibility at native size, not only thumbnail size.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify only if measurements justify | `shared/assets/render_presentation.v0.json` | Dungeon exposure, fog, contrast, and tier settings |
| Modify only if measurements justify | `shared/assets/fog_presentation.v0.json` | Overlay darkness/feather and ambient suppression |
| Modify only if measurements justify | `shared/assets/dungeon_torch_presentation.v0.json` | Torch pool strength, reach, and color |
| Modify only if depth hierarchy requires it | `shared/rules/dungeon_generation.v0.json` | Existing biome light colors and energies |
| Modify with any new field | Matching `*.schema.json` under `shared/assets/` or `shared/rules/` | Bound and validate catalog values |
| Modify only if catalog tuning fails | `client/scripts/scene_lighting_rig.gd`, `dungeon_depth_lighting.gd`, `fog_of_war_overlay.gd`, `dungeon_torch_lights.gd` | Existing presentation behavior; no gameplay visibility changes |
| Extend focused checks | `client/tests/test_render_environment_presentation.gd`, `test_dungeon_depth_lighting.gd`, `test_fog_of_war_overlay.gd`, `test_dungeon_torch_placement.gd` | Data loading, quality tiers, fog/torch invariants |
| Add a focused scenario/capture path if needed | `tools/bot/scenarios/client/dungeon_light_readability_*.json`, `client/scripts/bot_frame_capture.gd`, `bot_wait_handlers.gd`, `bot_action_handlers.gd`, `bot_step_catalog.gd` | Bounded real-camera frames with tier/renderer manifest; loot fixture remains open |
| Finish after acceptance | `docs/as-built/v498_dungeon-light-readability.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Matched comparison and slice status |

## Maintenance ratchet

Target files at or below 600 lines. `fog_of_war_overlay.gd` is already 631 lines; keep it untouched for catalog tuning. If behavior must change there, extract a focused helper or document why an extraction would be riskier and keep the grandfathered file at or below its baseline. Check other touched files against `.maintainability/file-size-baseline.tsv`.

```bash
make maintainability
```

## Task 1 — Lock the v497 control and complete the live fixture

Files: this plan; a focused `tools/bot/scenarios/client/` scenario and capture helper only if necessary.

- [x] Read the shipped v497 as-built and reproduce its matched live performance topology on the v498 starting checkout. Store raw runs with fixture/seed, quality, renderer, resolution, hardware, vsync, and commit.
- [x] Build bounded depth-1/depth-8 and boss-warning visual scenarios. The shallow route withdraws a real unique sword from the town chest, drops it in the dungeon, and asserts `item_dropped`, a loot entity, and its presentation before capture. The CI pack is unchanged.
- [x] Capture from the active main viewport after `RenderingServer.frame_post_draw`, preserving HUD and fog; record renderer and graphics tier in an adjacent JSON manifest under `.artifacts/bot-captures/`. The capture step waits for save completion before the scenario advances. Selected final images will be linked from the as-built after acceptance.

```bash
make bot-client SCENARIO=dungeon_light_readability HEADLESS=0
make bot-client SCENARIO=boss_telegraph_decals HEADLESS=0
```

The exact scenario name above is the planned v498 visual-verification command. Existing focused checks are `make bot-client SCENARIO=dungeon_torch_lights HEADLESS=0` and `make bot-client SCENARIO=fog_of_war_overlay HEADLESS=0`.

## Task 2 — Tune one presentation layer at a time

Files: the listed shared catalogs, matching schemas only for new fields, and focused loader/tier tests.

- [x] Compare the six baseline scenes at native resolution. Label each failed observation: player/enemy contrast, loot, warning, torch falloff, near/far separation, or fog obscurity.
- [x] Try biome light and existing render-context values before torch and fog-overlay values. Keep every tunable value in shared data. Record each candidate's catalog diff and paired frame so a rejected candidate can be reverted cleanly.
- [x] Preserve the Performance tier's intentional omissions of expensive SSAO, render fog, and key-light shadows; check that both tiers keep the same interaction cues and that torch pools do not expose unexplored content.
- [x] Use a script change only for a measured issue catalogs cannot express; state its cost and focused test in the as-built.

```bash
make validate-shared
make client-unit
make bot-client SCENARIO=dungeon_torch_lights HEADLESS=1
make bot-client SCENARIO=fog_of_war_overlay HEADLESS=1
make bot-client SCENARIO=boss_telegraph_decals HEADLESS=1
```

## Task 3 — Pair final visual and performance evidence

Files: capture manifests and raw logs under `.artifacts/`; as-built after acceptance.

- [x] Save before/after frames for shallow and both deep palettes, with enemy, loot, warning, torch pools, HUD, and fog of war. Inspect Balanced and Performance at native size. Keep subjective findings separate from scenario assertions.
- [x] Re-run the v497 live benchmark topology for three matched control/result repetitions per tier. Report frame p50/p95, process p95, draw calls, and first-spawn cost alongside the exact configured budget above; check for cap/noise confounders.
- [ ] If a budget fails, revert or narrow the lighting change and repeat the paired run. Record any residual limitation rather than claiming visual or performance acceptance from a headless pass.

```bash
make benchmark
make bot-client SCENARIO=dungeon_light_readability HEADLESS=0
make bot-client SCENARIO=boss_telegraph_decals HEADLESS=0
```

## Task 4 — Lifecycle and integration handoff

Files: `docs/as-built/v498_dungeon-light-readability.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md`.

- [x] Document catalog changes, paired images, exact commands, raw-sample paths, pass/fail against each acceptance criterion, and any fixture limits.
- [x] Run focused validation and `make maintainability` in this isolated worktree. Leave worktree changes uncommitted for the coordinating task.
- [x] After integration, the coordinating task ran the final combined `make ci` and handled lifecycle and commits. The isolated v498 worktree did not run a separate final `make ci`.

## Deferred scope

No new renderer or asset pipeline, geometry or biome-generation change, combat/protocol behavior change, default-quality change, or fog-of-war replacement. Final tuning and acceptance remain blocked until v497's frame budget is available.

## Provisional presentation trial on the v493 checkout

The current trial changes only isometric fog-of-war ambient suppression from directional/ambient `0.35/0.12` to `0.50/0.22`. A stronger falloff exponent (`2.5` and `3.0`) barely improved the edge silhouette and was reverted. A smaller ambient trial (`0.42/0.18`) raised the near-field more than the distance and was also reverted. The retained values brighten explored ground and hero/enemy separation without changing light count, shadow features, overlay darkness, or tier defaults. Captures in `.artifacts/bot-captures/` are exploratory: the real floor loot is present and visible in the shallow frames, the boss line is recognizable in both tier frames, and the deep cool palette is distinct. Exact before/after animation/tick alignment, an actual torch pool in frame, an engaged enemy within explored space, and v497 performance budget are not yet accepted. The currently visible distant enemies remain mostly obscured by unexplored fog, as intended.

Focused verification on this checkout: `make validate-shared` (2,206 checks plus CODEMAP), `make client-unit`, `make maintainability`, and windowed `make bot-client SCENARIO=dungeon_light_readability HEADLESS=0` (four scenarios) passed. After adding the real-loot route, `make bot-client SCENARIO=dungeon_light_readability_shallow HEADLESS=0` passed separately. These checks do not establish the v497 frame budget or final visual acceptance.

The follow-up fixtures `dungeon_light_readability_torch.json` and `dungeon_light_readability_engaged_enemy.json` cover the two missing visual states. The torch route selects a real mounted torch from the current layout, reaches a room-side approach point, and checks player distance; a windowed run additionally checks its ground-point camera projection. The enemy route waits for a real `monster_aggro` event and then the **same living event subject** within four world units, projected inside the windowed camera. Both routes passed with `HEADLESS=1`; screenshot steps explicitly skip in that mode, so those passes are route proof only. The tightened fixture assertions passed focused Godot tests, and both routes passed with `HEADLESS=0` on Godot 4.7.2 Forward+ Metal at 1280×720 in Balanced and Performance. Run the exact visual commands to repeat:

```bash
make bot-client SCENARIO=dungeon_light_readability_torch HEADLESS=0
make bot-client SCENARIO=dungeon_light_readability_engaged_enemy HEADLESS=0
```

The engaged-enemy before/after comparison passed on the same scenario, seed, renderer, viewport, and quality tiers. The original `0.35/0.12` frames and manifests are `.artifacts/v498-baseline/before-engaged-{balanced,performance}.{png,json}`; the retained `0.50/0.22` pair is `after-engaged-*` there. Baseline and candidate frames were captured at ticks 66–67 (Balanced) and 77 (Performance). The enemy remains visible in explored space in both tiers. The retained pair brightens the ground around the fight and makes the bodies easier to distinguish at native size; the scene positions and animation frames are close but not pixel-identical. The in-frame FPS labels are incidental and do not satisfy the v497 budget.

The subsequent integrated capture found the torch lifecycle fault: `StaticWalls` discarded the torch root during wall replacement. The root now lives under the scene and is cleared explicitly on teardown; the bot reports actual rendered nodes and requires two before capture. The final shallow route rendered 32 torches, with visible flames and warm pools in both tiers. The matched v497-topology A/B met the preset frame, process, draw-call, and first-spawn budgets in both tiers. Exact images, values, and remaining viewpoint limits are in the [as-built](../as-built/v498_dungeon-light-readability.md); the earlier isolated-worktree provisional observations above remain historical notes.
