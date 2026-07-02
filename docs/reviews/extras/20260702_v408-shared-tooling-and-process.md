# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v408**

**Date:** 2026-07-02  
**Scope:** `shared/`, `tools/`, Makefile/CI, docs/SDD — v399–v408 vs v398 baseline.  
**Baseline:** `main` @ `ffdd3f0b`. Worktree: **clean** at review start.  
**Stats:** 228 bot scenarios; CI pack **35** (22 protocol + 13 client); protocol **v8**; `validate_shared.py` **3,062** lines; 1,909 validation checks.  
**Overview:** [`../20260702_v408-overview.md`](../20260702_v408-overview.md)

---

## Summary

Contract hygiene and the two-tier CI model remain a **strength**. Slices v399–v408 shipped substantive shared/rules work without a protocol version bump. **v398 duplicate scenario ID debt closed** via `CROSS_TREE_SCENARIO_PAIRS` in `ci_pack.py`. **Open debt:** `validate_shared.py` monolith (+29 lines) and systematic tuning-friendly rule-test audit.

**CI note:** `make ci-full` at review start **failed step 1** (`character_stats_panel.gd` ratchet breach). Post-`$refactor` maintainability is green; full-matrix re-run pending.

## 1. Architecture

**[Strength]** Rules-as-data boundary held — upgrade risk, skill decuple, mercenary hire, class-specialist gear, weapon elemental damage/procs in `shared/rules/` with no protocol bump.

**[Strength]** Duplicate scenario IDs now CI-enforced (`tools/bot/ci_pack.py:79–108`).

**[Strength]** v399 closed v398 reconnect proof gap — extended `ws_reconnect_proof`, no protocol change.

**[Med]** v407/v408 weapon elemental procs validated via JSON Schema but **no semantic cross-check** linking proc `effect_id` to `unique_effects.v0.json`.

**[Low]** v402 merged into v401 — spec exists, no lifecycle row.

## 2. Technical

**[Strength]** `validate_skills.py` decuple gate (v401); v400 golden `upgrade_success_chance.json`.

**[Strength]** `make validate-shared` — **1,909 checks** pass (up from 1,818 at v398).

**[Med]** `validate_shared.cross_checks()` ~2,845 lines (`validate_shared.py:201–3046`); +29 lines since v398.

**[Med]** `run.py` at **4,370** lines — at maintainability baseline cap.

**[Low]** `gen-codex` not in CI — stale `codex_index.v0.json` only caught by `test_build_codex.py`.

## 3. Maintainability

**[Strength]** CI pack stable at 35 through 10-slice autoloop batch; extraction-coupling ratchet 0 sites.

| File | Lines | Baseline |
|------|-------|----------|
| `validate_shared.py` | 3,062 | 3,038 |
| `run.py` | 4,370 | 4,370 |

## 4. Documentation

**[Strength]** v399–v408 SDD trail healthy — spec + plan + as-built per shipped slice.

**[Strength]** v392–v394 missing plans from v398 — **closed**.

**[Strength]** `scenario-catalog.md` correctly cites 22 + 13 pack size.

**[Low]** v402 merge-into-v401 documented only in spec header.

## 5. Process / SDD

**[Strength]** Official review cadence at v408 milestone.

**[Strength]** Autoloop batch v399–v408 maintained SDD discipline under time pressure.

**[Med]** `make ci-full` health unverified at v408 review start (step 1 ratchet fail); re-run required post-refactor.

**[Med]** Tuning-friendly rule-test audit remains open (`PROGRESS.md` backlog).

---

## Top 5 shared/tooling/process refactors

1. **[High · Maint]** Decompose `validate_shared.cross_checks()` into importable domain modules.
2. **[Med · Test]** Systematic tuning-friendly rule-test audit — start with validator-owned pins (`validate_shared.py:3035–3039`).
3. **[Med · Validation]** Semantic cross-checks for codex + weapon elemental proc `effect_id` ↔ `unique_effects`.
4. **[Med · Maint]** Typed bot runtime context to stop `run.py` growth at baseline.
5. **[Med · Process]** Confirm `make ci-full` green post-refactor before next feature batch.

*Evidence: `PROGRESS.md`, `tools/bot/ci_pack.json`, `scripts/ci.sh`; comparison to [`20260701_v398-shared-tooling-and-process.md`](20260701_v398-shared-tooling-and-process.md).*
