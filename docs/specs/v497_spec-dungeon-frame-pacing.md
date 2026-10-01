# v497 — Dungeon frame pacing

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `dungeon-frame-pacing`
- **Area:** Gameplay smoothness / rendering
- **Depends on:** v494 prop layout, v495 first-spawn fix, and the engineering review/refactor gate after v496.
- **ADR:** [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D6–D9.

## Purpose

Keep movement and combat visually steady in dense kit dungeons. The earlier comparison showed roughly four times as many draw calls after kit walls/floors were introduced. Profile the current live renderer, then reduce the measured steady-state wall/floor or shadow cost while preserving the wall geometry, occlusion fade, and pick identity used by gameplay presentation.

## Non-goals

- Lowering monster counts, reducing room complexity, or changing authoritative collision/navigation.
- Hiding the problem by changing the default quality tier or disabling established effects globally.
- First-spawn model work (v495) or a broad engine/render-pipeline rewrite.

## Acceptance criteria

- [ ] Matched live-client baseline and result cover the same dense dungeon seed, entity count, camera route, renderer, quality tier, and hardware. Report draw calls, primitives, frame-time p50/p95/p99, process p95, and server tick p95 separately.
- [ ] A render-profiler breakdown identifies the main source of draw calls before choosing batching, merged geometry, material reuse, or a bounded shadow change.
- [ ] On the selected dense fixture at Balanced quality, draw calls fall by at least 25% relative to the fresh pre-change baseline, with no regression in frame-time p95 or visible wall detail. If profiling shows wall geometry is not the dominant source, revise this target in the plan before implementation rather than optimizing the wrong path.
- [ ] Occlusion fading, wall selection/pick identity, collision alignment, floor coverage, and biome material variation behave as before. Static batching must not merge objects that require independent fade state.
- [ ] The Performance tier continues to show a meaningful cost reduction relative to Balanced; no per-frame full mesh rebuild is introduced.
- [ ] Real-renderer before/after captures at shallow and deep depth show the same scene composition and readable combat.

## Scope and likely files

- **Client:** `client/scripts/dungeon_kit_wall_builder.gd`, `client/scripts/wall_renderer.gd`, `client/scripts/dungeon_kit_floor.gd`, `client/scripts/wall_occlusion_fade.gd`, `client/scripts/kit_piece_library.gd` as indicated by profiling.
- **Data:** a schema-backed batching or quality parameter under `shared/assets/` only if tuning is needed.
- **Tools/tests:** extend the live benchmark/report and focused wall/fade tests; capture both quality tiers. No server or protocol change expected.

## Test and bot proof

- Compare matched `make benchmark` or equivalent live Godot + bot runs and retain raw samples. Offline replay alone cannot prove rendered frame pacing.
- Run `make client-unit`, `make validate-shared`, dungeon wall/floor and wall-occlusion client scenarios, and the relevant `make regen-screenshots SUITE=...` captures.
- The plan names the exact `make bot-visual scenario=...` command for real-window verification of occlusion and dense combat.

## Asset/plugin decision

- **Borrow** the existing KayKit mesh/atlas and Godot batching options already available to the client.
- **Reject** new art packs and renderer plugins; preserve the established ADR-0018 look.

## Open questions and risks

- The fourfold draw-call number is historical. Fresh v496-era data may point to props, particles, or shadows rather than wall pieces.
- Global batching could break per-wall fade and picking. The plan must show how each independently fading wall remains addressable before selecting a batching method.
