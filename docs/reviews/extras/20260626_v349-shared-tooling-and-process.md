# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v349**

**Date:** 2026-06-26  
**Scope:** `shared/`, `tools/`, Makefile/CI, SDD docs, agent skills — v340–v349 batch.  
**Baseline:** `main` @ `82422216`. Worktree: uncommitted `skills/review/SKILL.md` (now requires `make ci-full` for reviews).  
**Stats:** protocol v8; 20 rules catalogs; 35 golden fixtures; 196 bot scenarios (36 pack / ~160 extended); `validate_shared` 1,564 checks; `validate_assets` orphan gate; 34 grandfathered files (64,429 lines).  
**Overview:** [`../20260626_v349-overview.md`](../20260626_v349-overview.md)

---

## Summary

Contract hygiene remains a project strength: **protocol v8 unchanged** across v348–v349 presentation work, **`movement_presentation.v0.json`** added with schema validation, and **`monster_rarity` golden** updated intentionally at v340 for depth pressure. CI pack policy (`validate_ci_pack`) is enforced; **`validate_assets` step [7]** now rejects unmanifested `client/assets` GLB/import sidecars (landed `82422216`).

Process drift is the headline: **`make ci` is green** on the merge pack, but **`make ci-full` failed** (37m05s) with 12 extended protocol + 3 extended client-bot failures. `PROGRESS.md` still gates periodic reviews on `make ci` while the review skill now requires `ci-full` — reconcile on this review landing. SDD traceability gaps persist for v340–v347 presentation slices (specs/as-builts without plans).

## 1. Architecture

**[Strength]** Rules-as-data boundary preserved — v348/v349 touched presentation assets only; combat formulas and protocol unchanged.

**[Strength]** Asset manifest is source of truth; orphan client asset gate prevents Godot import debris from re-entering the repo.

**[Med]** `validate_shared.py` still ~3,038 lines with monolithic `cross_checks()` — v337 extraction item open.

## 2. Technical

**[Strength]** `make validate-shared` — **1,564 checks passed** (2026-06-26 review run).

**[Strength]** `make maintainability` — pass; grandfathered count **34 files / 64,429 lines** (trending down vs v337’s breach narrative).

**[Med]** `movement_presentation` has schema-only validation — unlike `fog_presentation`, no semantic range/coupling guards yet.

**[blocker · CI]** `make ci-full` **FAILED** (37m05s):
- Step **9/11** protocol bot: **12 extended scenario failures** (unique effects, coop rewards, pack aggro, dungeon elite, mercenary foundation, etc.).
- Step **10/11** client bot: **3 failures** — `character_stats_panel`, `blacksmith_armor_recipe`, `wall_floor_dungeon_rollout`.
- Steps 1–8, 11 passed (unit gates, pack-equivalent early steps, headless smoke).

**[Strength]** Fast `make ci` (36-scenario pack) green per `PROGRESS.md` — appropriate for merge gate, insufficient for full-matrix confidence.

## 3. Maintainability

**[Strength]** CI two-tier model clear in `make/ci.mk` + `scripts/ci.sh` (same 11 steps; `ARPG_CI_SCENARIO=ci|all`).

**[Med]** `tools/bot/run.py` ~4,252 lines — grandfathered orchestrator; `run.py` split freeze still policy.

**[Med]** `docs/progress/scenario-catalog.md` missing v348–v349 entries (`entity_tick_smoothing`, render scenarios).

## 4. Documentation

**[Strength]** v348/v349 full SDD trail (spec, plan, as-built).

**[Med]** v340–v346: specs + as-builts but many **missing plans**; v343/v344/v346 spec-gate exempt without uniform lifecycle annotation.

**[Med]** `PROGRESS.md` stale items: “batch `make ci` pending” in v337 follow-ups vs current “CI green”; review cadence still says v340 target while at v349.

**[Strength]** Review skill updated to require `make ci-full` — aligns official review gate with full matrix (pending PROGRESS.md sync).

## Top 5 extras refactors

1. **[blocker · Process]** Triage **`make ci-full` failures** — classify flake vs regression; fix or quarantine extended scenarios with explicit ownership.
2. **[blocker · Process]** Update **`PROGRESS.md`** review cadence: Last review → v349; Next → ~v359; gate text → `make ci-full` for periodic reviews.
3. **[minor · Schema]** Add `point_light` to fog schema top-level `required` (v337 item still open).
4. **[future · Maint]** Extract `validate_shared.py` validation domains + add `test_validate_boss_patterns.py` / `test_validate_item_presentations.py`.
5. **[minor · SDD]** Backfill plans or exemption notes for v340–v347; refresh `scenario-catalog.md` for v348–v349.

*Evidence: `make validate-shared`, `make maintainability`, `tools/bot/ci_pack.py`, `validate_assets.py:235-252`, `/tmp/arpg-ci-full-review.log`, `PROGRESS.md`, `skills/review/SKILL.md`.*
