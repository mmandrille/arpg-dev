# v348 Spec: Forward Plus Renderer

Status: Complete  
Date: 2026-06-26  
Codename: `forward-plus-renderer`  
Baseline: v347 `dungeon-render-performance`

## Purpose

Migrate the interactive Godot client default renderer from `gl_compatibility` to `forward_plus`
so dungeon combat and fog presentation use the modern 3D pipeline on player machines, while
headless CI/bot paths keep an explicit `gl_compatibility` override where required.

v347 proved the remaining combat FPS bottleneck is client rendering (not mesh count). This slice
evaluates whether `forward_plus` is a safe default for `make play` / `make bot-visual` without
breaking fog shaders, materials, or headless gates.

## Non-goals

- No server, protocol, shared rules, or golden changes.
- No gameplay tuning, fog cache logic changes, or graphics-quality preset redesign (v347 owns those).
- No mobile / Vulkan / `mobile` renderer support matrix in this slice.
- No per-user renderer toggle in Settings unless migration proves a hard blocker (prefer single default).
- No CI pack promotion of perf scenarios.

## Acceptance criteria

- `client/project.godot` sets `renderer/rendering_method="forward_plus"` for interactive play.
- `make play` and `make bot-visual` (non-headless) launch with forward_plus as the project default.
- Headless unit/smoke/bot-client entrypoints that invoke Godot directly continue to pass using
  `--rendering-method gl_compatibility` (or equivalent project override) so CI does not require GPU features.
- Visual regressions pass on headless client bot scenarios:
  - `67_fog_of_war_overlay`
  - `68_fog_los_shadow_mask`
  - `73_door_fog_toggle`
- Extended rendered check: `ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe`
  completes without protocol/client failures (FPS captured in as-built, not CI-locked).
- Document adopt/borrow/reject: **reject** external render plugins; **borrow** existing fog overlay,
  depth lighting, and GLB materials; no new asset pipeline.
- As-built records forward_plus vs v347 `gl_compatibility` FPS sample on the implementation host.

## Scope and likely files

| Area | Files |
|------|-------|
| Project default | `client/project.godot` |
| Launch scripts | `scripts/bot_client.sh`, `scripts/client_smoke.sh`, `Makefile` targets if Godot flags need centralizing |
| Headless tests | `client/tests/*.gd` invocations via `client/scripts/client_smoke.sh` (keep `gl_compatibility` override) |
| Docs | `docs/as-built/v348_forward-plus-renderer.md`, lifecycle, `PROGRESS.md` |
| Evaluation matrix | as-built table: fog, lighting, materials, perf probe |

Unlikely: `fog_of_war_overlay.gd` shader edits unless forward_plus exposes a concrete breakage.

## Test and bot proof

Focused:

```bash
make client-unit
make maintainability
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=73_door_fog_toggle
```

Extended / manual:

```bash
ARPG_PERF_DEBUG=1 make bot-visual scenario=dungeon_combat_perf_probe
make play
```

## Open questions and risks

- **Fog fullscreen shader:** `FogOfWarOverlay` uses a custom `Shader` on `ColorRect`; verify under
  forward_plus (primary regression risk).
- **Headless CI:** Godot 4 headless may not support forward_plus on all hosts; keep compatibility
  renderer on automated paths.
- **Material compatibility:** committed `.glb` runtime assets must render without pink materials;
  spot-check player + dungeon_mob in play mode.
- **Perf portability:** forward_plus may help or hurt depending on GPU; as-built captures samples only.

No product blockers assumed: default is migrate project default + preserve headless override pattern.
