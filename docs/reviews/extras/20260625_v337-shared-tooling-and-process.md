# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v337**

**Date:** 2026-06-25
**Scope:** `shared/`, `tools/`, `Makefile`, `scripts/`, SDD docs. Covers v335–v337 paydown and v334 `$refactor` completion.
**Baseline:** `main` at `da9a47b4`; worktree clean.
**Stats:** `make validate-shared` 1,557 checks OK; `validate_codemap.py` OK; 193 bot scenarios (36 CI pack); 323 specs / 319 plans / 333 as-builts.
**Overview:** [`../20260625_v337-overview.md`](../20260625_v337-overview.md)

---

## Summary

Shared-contract and Python-tooling health is **strong and improving** since the v334 ad hoc review. The v334 `$refactor` queue is largely complete: fog validator extraction, camera/fog alignment cross-check, `test_validate_main_config.py`, `test_validate_fog_presentation.py`, item-visual probe extraction, and `main.gd` maintainability recovery. `make validate-shared` now runs **1,557 checks** (+2 from fog alignment guards). `make maintainability` passes (34 grandfathered files, 64,302 lines).

Remaining gaps: `validate_shared.py` coordinator scale (~3,038 lines), `run.py` orchestrator (~4,252 lines), SDD traceability on v335/v337 refactor slices (no specs/plans), and batch `make ci` still pending on host.

---

## 1. Shared contracts

**[Strength]** `make validate-shared` passes 1,557 checks including fog semantic range guards and camera/fog organic-edge alignment (`validate_fog_presentation.py`).

**[Strength]** Protocol remains v8; v335–v337 changes are client extraction and server tick context — no unnecessary schema bump.

**[Strength]** `fog_presentation.v0.schema.json` `point_light` sub-object declares required fields — v331 silent-omit risk closed.

**[Strength]** 35 golden fixtures with schemas; cross-language combat/loot/replay contracts intact.

**[Low]** `point_light` not in fog schema top-level `required` — block can be omitted silently; low risk while committed data is complete.

**[Med]** `cross_checks()` in `validate_shared.py` remains ~2,800 lines — coordinator debt, not schema coverage gap.

---

## 2. Python tooling

**[Strength]** `validate_codemap.py` inverse unlisted-file detection — CODEMAP rot guard active.

**[Strength]** Extracted validators with unit tests: `validate_fog_presentation.py` (3 tests), `validate_main_config.py` (3 tests), `validate_dungeon_goldens.py`, `validate_unique_items.py`.

**[Strength]** `tools/bot/test_protocol.py` — 64 tests pass (spot check).

**[Strength]** CI pack curation stable — 36 scenarios in `tools/bot/ci_pack.json`; extended tier for non-gate scenarios.

**[Med]** `validate_shared.py` at 3,038 lines (baseline ratcheted post fog extraction). Further domain extraction remains future-plan scale.

**[Med]** `run.py` at 4,252 lines — grandfathered orchestrator; `bot_types.py` discipline holds.

**[Low]** `validate_boss_patterns.py`, `validate_item_presentations.py`, `validate_i18n.py` lack dedicated `test_*.py` — inconsistent with newer extraction pattern.

---

## 3. Process / SDD

**[Strength]** v332–v334 feature slices have full spec/plan/as-built trails.

**[Strength]** CLAUDE.md spec-gate exemption documents presentation-only slice path.

**[Med]** v335 (`movement-input-presenter`) and v337 (`sim-tick-context`) are refactor paydown slices with as-builts but **no specs/plans** — acceptable under `$refactor` but SDD traceability gap.

**[Med]** v334 plan gap noted in lifecycle table (spec + as-built only).

**[Med]** Batch `make ci` still pending — PROGRESS.md CI gate line unchanged; spot checks green.

**[Low]** `docs/progress/scenario-catalog.md` may need refresh for v336 scenario 78.

---

## 4. Documentation

**[Strength]** CODEMAP current — `validate_codemap.py` passes including inverse check.

**[Strength]** ADR-0015 movement-speed formula accepted; ADR-0001–0008 foundational set intact.

**[Strength]** v337 as-built documents tick-context starter and verification commands.

**[Low]** Engineering review cadence: v334 ad hoc; this v337 review is user-requested before v340 milestone.

---

## Top 5 extras refactors

1. **[future · Maint]** Continue `validate_shared.py` domain extraction — shop/affix cross-check cluster or item-presentation guards; target shrinking `cross_checks()` without `helpers=globals()`.

2. **[minor · Schema]** Add `point_light` to fog schema top-level `required` — closes silent-omit path for hero visibility tuning.

3. **[minor · Test]** Add `test_validate_boss_patterns.py` and/or `test_validate_item_presentations.py` — symmetric unit tests for extracted validators.

4. **[blocker · Process]** Stabilize batch `make ci` and update PROGRESS.md CI gate line — merge confidence depends on full pipeline.

5. **[future · Maint]** Multi-slice `sim.go` coordinator paydown using `simTickCtx` — v337 starter in place; backend-owned but affects shared-rules loader shape indirectly.

**[reject]** CODEMAP inverse check — landed post-v331.
**[reject]** Fog `point_light` sub-object required fields — landed post-v331.
**[reject]** `test_validate_main_config.py` — landed.
**[reject]** Camera/fog cross-check — landed (`validate_camera_fog_mode_alignment`).
**[reject]** `main.gd` ratchet breach — resolved v335 + probe extraction.

*Evidence: `make validate-shared` 2026-06-25; `make maintainability`; `.venv/bin/pytest tools/test_validate_fog_presentation.py tools/test_validate_main_config.py tools/test_validate_codemap.py tools/bot/test_protocol.py -q` (72 passed); v334 extras review comparison.*
