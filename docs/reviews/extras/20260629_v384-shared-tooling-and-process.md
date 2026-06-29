# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v384**

**Date:** 2026-06-29  
**Scope:** `shared/`, `tools/`, Makefile/CI, SDD docs — v350–v384 combat/exploration perf batch.  
**Baseline:** `main` @ `625f5600`. Worktree: **clean**.  
**Stats:** protocol v8; 20+ rules catalogs; 35 golden fixtures; ~196 bot scenarios (pack + extended); `validate_shared` checks OK; 34 grandfathered files (64,522 lines); new scenario `104_crowded_melee_perf_probe.json`.  
**Overview:** [`../20260629_v384-overview.md`](../20260629_v384-overview.md)

---

## Summary

Contract hygiene remains strong: v370–v384 added data-driven tuning keys (`presentation_lod`, `client_perf`, `monster_overload_*`, fog combat throttle) with schema/golden updates. CI pack policy enforced; maintainability and extraction-coupling ratchets green.

Process drift: **`make ci-full` failed** again (22m03s, **3 extended scenarios**) — different set than v349’s 15 failures (v350 reportedly recovered those). v370–v384 SDD trail is **stub-heavy** (9 slices, minimal as-builts). `PROGRESS.md` claimed ci-full green post-v350; current matrix shows new extended debt.

## 1. Architecture

**[Strength]** Rules-as-data boundary preserved — perf tuning in `main_config.v0.json`, `navigation.v0.json`, `fog_presentation.v0.json`; no protocol version bump.

**[Strength]** Golden schema updated for `monster_overload_live_monster_threshold` in navigation rules.

**[Med]** `validate_shared.py` still ~3,038 lines with monolithic `cross_checks()` — v337 extraction item open.

**[Strength]** New perf probe `tools/bot/scenarios/104_crowded_melee_perf_probe.json` exercises crowded combat path; movement audit row added.

## 2. Technical

**[Strength]** `make validate-shared` — pass (2026-06-29 spot-check during review).

**[Strength]** `make maintainability` — pass; 34 grandfathered files / 64,522 lines; extraction-coupling 0 sites.

**[Strength]** `make ci` merge pack — green (2026-06-29 post-v384 per `PROGRESS.md`).

**[blocker · CI]** `make ci-full` **FAILED** (22m03s, 2026-06-29 review run):
- Step **9/11** protocol bot: **`companion_rank_scaling_and_limits`** (15.73s).
- Step **10/11** client bot: **`client_skill_points_and_magic_bolt`** (step 23 `wait_event` `skill_cooldown_started` — event in `pending_events` but not consumed within 5s); **`interactable_tick_smoothing`** (step 1 `wait_entity` interactable — count stayed 0).
- Steps 1–8, 11 passed (unit gates, headless smoke 74 GDScript tests).

**[Strength]** v349’s 15 extended failures (coop rewards, unique effects marathon, wall_floor, etc.) **passed** in this run — recovery from v350 holds for that cohort.

**[Med]** `movement_presentation` still schema-only validation — no semantic range guards.

## 3. Maintainability

**[Strength]** CI two-tier model clear (`make/ci.mk`, `ARPG_CI_SCENARIO=ci|all`).

**[Med]** `tools/bot/run.py` ~4,258 lines — grandfathered orchestrator; split freeze policy holds.

**[Med]** v370–v384 batch landed 9 slices with thin documentation — agents lack proof commands for regression triage.

**[Strength]** Scenario movement audit gate includes `104_crowded_melee_perf_probe.json`.

## 4. Documentation

**[Strength]** Lifecycle table complete for v370–v384.

**[Med]** v370–v384 specs/plans exist but as-builts are stubs for most perf slices — SDD traceability regression vs v340–v349 presentation batch.

**[Med]** `PROGRESS.md` ci-full narrative (“all 15 recovered at v350; ci-full now green”) is **stale** — 3 new extended failures on 2026-06-29.

**[Strength]** Review skill requires `make ci-full` for periodic reviews — this review recorded full-matrix outcome.

## Top 5 extras refactors

1. **[blocker · Process]** Triage **`make ci-full` failures** — `companion_rank_scaling_and_limits`, `client_skill_points_and_magic_bolt`, `interactable_tick_smoothing`; run `VERBOSE=1 make bot scenario=...` / `make bot-client SCENARIO=...`.

2. **[minor · Process]** Update **`PROGRESS.md`** ci-full status — distinguish v350 recovery (v349 cohort) from current extended debt.

3. **[minor · Schema]** Add `point_light` to fog schema top-level `required` (v337 item still open).

4. **[minor · SDD]** Backfill v372–v378 as-builts with behavior + verification; use v370 as template.

5. **[future · Maint]** Extract `validate_shared.py` validation domains + boss/item presentation validator tests.

*Evidence: `make ci-full` log `/tmp/arpg-ci-full-review.log` (2026-06-29), `make maintainability`, `tools/bot/ci_pack.json`, `PROGRESS.md`, `shared/rules/main_config.v0.json`.*
