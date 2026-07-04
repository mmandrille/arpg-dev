# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v420**

**Date:** 2026-07-03  
**Scope:** `shared/`, `tools/`, Makefile/CI, SDD cadence — v409–v420 vs v408 baseline.  
**Baseline:** `main` @ `dd037a16`. Worktree: **clean**.  
**Stats:** **239** bot scenarios (136 protocol + 103 client); CI pack **35** scenarios; `skills.v0.json` **~5,194** lines; `validate_shared.py` **3,148** lines; `run.py` **4,393** lines.  
**Overview:** [`../20260703_v420-overview.md`](../20260703_v420-overview.md)

---

## Summary

The v409–v420 autoloop batch strengthened the **skill catalog as shared data** while keeping **protocol v8** and a **stable 35-scenario CI pack**. Python bot tooling is modular on paper (38+ modules, `BotContext`, `movement_runtime` extraction) but **`run.py` remains a 4,393-line grandfathered orchestrator** at the ratchet ceiling (+23 vs baseline). SDD artifacts for v409–v420 are complete; process gaps are **overdue periodic review** (12 slices since v408) and **v420–v424 spec vs single-slice delivery drift**.

**[Strength]** `make ci-full` green after scenario alignment for v420 skill catalog (`936b8c61`, `ca2d9567`, `dd037a16`). **Risks:** `validate_shared.cross_checks()` monolith (~2,900 lines), extended-only skill proofs, tuning-test audit still open.

## 1. Architecture

**[Strength]** Rules-as-data held — no protocol bump v409–v420; Go/GDScript evaluators consume `shared/rules/`.

**[Strength]** Two-tier CI model enforced: pack members in `tools/bot/ci_pack.json` (35) vs `"ci_tier": "extended"` elsewhere; `validate_ci_pack()` gate in CI step 7.

**[Strength]** `class_foundation_coverage.py` — foundation scenarios must reference every non-passive active per class from `skills.v0.json`.

**[Med]** `skills.v0.json` growth increases validator + scenario coupling; catalog churn risks extended-scenario drift without pack promotion policy.

## 2. Technical

**[Strength]** `validate_skills.py` extracted for fastest-growing domain (~463 lines).

**[Strength]** `movement_runtime.timeout_s` on `move_until_entity_in_range` — wall-clock cap prevents multi-minute dungeon walks (`movement_runtime.py:104–120`, `dd037a16`).

**[Strength]** Scenario movement audit TSV + `pytest tools/test_scenario_movement_audit.py` gate incidental navigation.

**[Med]** `validate_shared.cross_checks()` still ~lines 201–3133 in one function — v408 top recommendation unaddressed.

**[Med]** No unit test for `movement_runtime` `timeout_s` path yet.

## 3. Maintainability

| Artifact | Baseline | Actual | Status |
|----------|----------|--------|--------|
| `tools/bot/run.py` | 4,370 | 4,393 | +23 — within +25 |
| `tools/validate_shared.py` | 3,127 | 3,148 | +21 |

Grandfathered ratchet: **PASS**. Extraction coupling: **0** `helpers=globals()` sites.

**[Med]** Typed bot runtime context only partial — `BotContext` exists but `run.py` still orchestrates most dispatch.

## 4. Documentation

**[Strength]** Uninterrupted spec/plan/as-built trail for v409–v420 (12 slices) in `docs/progress/slice-lifecycle.md`.

**[Strength]** `docs/progress/scenario-catalog.md` documents CI tiers, 10s budget policy, movement allowlist.

**[Med]** `docs/specs/v420_spec-class-skill-build-branches.md` plans v420–v424 per-class delivery; as-built shipped all 30 skills in v420 — spec/lifecycle stale for v421–v424.

**[Low]** `PROGRESS.md` CI gate row lagged until this review — now records green `make ci-full`.

---

## Top 5 shared/tooling/process refactors

1. **[High · Maint]** Decompose `validate_shared.cross_checks()` — follow `validate_skills.py` pattern for synergies, item visuals, weapon families (`tools/validate_shared.py`, `tools/validate_skills.py`).
2. **[High · Maint]** Finish typed bot runtime context — stop `run.py` +23 drift at baseline (`tools/bot/run.py`, `tools/bot/bot_context.py`).
3. **[Med · Test]** Start tuning-friendly rule-test audit — prioritize `*_class_foundation.json` and `test_protocol.py` pins duplicating `skills.v0.json` (`PROGRESS.md` backlog).
4. **[Med · SDD]** Reconcile v420–v424 spec with as-built — close or retitle v421–v424 in spec/lifecycle (`docs/specs/v420_spec-class-skill-build-branches.md`).
5. **[Med · CI]** Add unit test for `movement_runtime` `timeout_s`; document when scenarios should set `timeout_s` vs `max_ticks` only.

*Evidence: file reads at `dd037a16`; comparison to [`20260702_v408-shared-tooling-and-process.md`](20260702_v408-shared-tooling-and-process.md).*
