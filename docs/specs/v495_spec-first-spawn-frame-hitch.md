# v495 — First-spawn frame hitch

- **Status:** Complete (combined `make ci` passed)
- **Date:** 2026-09-30
- **Codename:** `first-spawn-frame-hitch`
- **Area:** Gameplay smoothness / client performance
- **Depends on:** v494 baseline and its prop count/draw-call measurements.
- **ADRs:** [ADR-0001](../adr/0001-technology-stack.md) D3; [ADR-0018](../adr/0018-art-direction-and-kit-based-visuals.md) D8.

## Purpose

Remove the visible pause when the first group of monsters appears in a live Godot session. The prior post-kit probe recorded a first-spawn `process_ms` spike of about 761 ms, but the current cause and magnitude must be measured again after v494. Fix the dominant measured client-side phase while preserving entity appearance and authoritative state.

## Non-goals

- General wall batching or steady-state draw-call work (v497).
- Combat balance, server tick-rate changes, protocol redesign, or lowering entity counts to hide the hitch.
- Speculative preloading of the entire asset catalog.

## Acceptance criteria

- [ ] A repeatable live-client first-spawn probe records fixture/seed, renderer, quality tier, host, model count, warm/cold state, `process_ms`, frame time, entity-upsert phases, and draw calls before changes.
- [ ] Profiling identifies the largest contributing phase or resource operation; the implementation addresses that cause through bounded preparation, caching, or work distribution. Preparation cannot stall the title screen or every level transition without being counted in the result.
- [ ] On the same host and scenario, at least 10 cold runs before and after show at least a 30% lower first-spawn p95, a lower max, and no new multi-frame catch-up pause. If the fresh baseline p95 is already below 100 ms, stop and document that finding before changing code.
- [ ] Monster count, skins, animations, and combat event presentation remain correct; no entity is silently skipped, and the server-authoritative timeline is unchanged.
- [ ] Steady-state frame p95, memory growth, and resource count do not regress materially against the paired baseline; any retained cache has a bounded lifetime or size.

## Scope and likely files

- **Client:** targeted entity/model creation path found by profiling, likely `client/scripts/kit_piece_library.gd`, `client/scripts/kit_monster_visual.gd`, entity presentation/upsert helpers, or scene setup. Touch only the measured hot path.
- **Tools:** `scripts/benchmark.sh`, `docs/performance/tools.md`, and the client performance report/probe if the present first-spawn sample lacks repeat-run distribution or phase detail.
- **Tests/docs:** focused cache/teardown tests and `docs/as-built/v495_first-spawn-frame-hitch.md` with paired results.
- **Contracts:** no shared gameplay rules, server state, or protocol changes expected.

## Test and bot proof

- Run the same live Godot + bot scenario before and after using `make benchmark` or a smaller equivalent built from its live-client topology. Protocol-only and offline replay timings are insufficient.
- Run affected Godot unit tests, `make client-unit`, `make validate-assets`, and a monster-heavy `make bot-visual scenario=...` selected in the plan.
- Save a real-renderer recording or frame trace of the first group appearing, plus the raw samples used for p50/p95/max.

## Asset/plugin decision

- **Borrow** the existing KayKit model manifests, resource loader, and in-repo performance instrumentation.
- **Reject** new visual assets or plugins; this slice changes how current assets load or instantiate, not their look.

## Open questions and risks

- The old 761 ms result came from an earlier build and one host. Fresh cold-run measurements determine the target and whether v495 still has work.
- Prewarming can move the hitch rather than eliminate it. Include startup/transition cost and memory in the comparison.
- A performance fix that changes ordering must preserve stable visible entity order and avoid changing replay/simulation behavior.
