# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v531**

**Date:** 2026-10-02  
**Scope:** Shared protocol/rules/schemas/goldens, Python validators and bot tooling, CI/build orchestration, SDD lifecycle and audit evidence.  
**Baseline:** `main` @ `afe691fb` (`docs: preserve v531 worker audit evidence`); clean at review start.  
**Stats:** `tools/validate_shared.py` 3,179 LOC; `tools/bot/run.py` 4,534 LOC; 44 shared rule catalogs and 74 shared goldens. The validator reports 2,271 checks.  
**Verification:** Coordinator reports `make ci` PASS (11/11, 21m23s); `make ci-full` not run. Review-side `make validate-shared` PASS (2,271 checks + CODEMAP); `make lint-determinism` PASS (77 grandfathered map-range sites); `make maintainability` FAILS on `server/internal/game/dungeon_encounter_composition.go` at 1,055 LOC vs 600-line new-file target.  
**Overview:** [`../20261002_v531-overview.md`](../20261002_v531-overview.md)

---

## Summary

The batch has a complete spec → plan → implementation → as-built record for v522–v531. Shared schemas and validators cover the new dungeon generation contracts, and the CODEMAP lists the added helpers. The 240-floor audit repeated byte-identically in the coordinator's integrated run, while a separate preserved worker report transparently records initial pack-placement failures and their later resolution.

The audit is a diagnostic report-only test, not a failing gate. A future CI-facing validator could turn its `failures` and `findings` into an explicit regression signal. Cross-language drift between Go JSON tags and shared schemas is still checked only by review/one-off tooling; the main shared validator verifies JSON schemas and instances but does not inspect Go structs. Exact floor-level audit JSON remains under ignored `.artifacts/`, so clone-portability is limited. `validate_shared.py` and `tools/bot/run.py` remain large coordinators.

## 1. Architecture

- **[Strength] Rules-as-data and schemas own dungeon tuning/contracts.** The generator's rule catalog and schema were added alongside its Go consumers, and `docs/CODEMAP.md:16` now indexes the helper family.
- **[Med] Go struct/schema parity is not automated.** `validate_shared.py` meta-validates schemas and validates instances (`tools/validate_shared.py:133-166`), but does not compare schema members against Go `json:` tags. The earlier v486 as-built notes a one-off check that found `boss_phase.duration_ticks` and `character_class` drift (`docs/as-built/v486_live-payload-schema-gate.md:105-106`).
- **[Med-Low] The validator still centralizes schema routing and dispatch.** Schema groups and mapping live near `tools/validate_shared.py:106-145,321-370`; the long cross-check coordinator runs through `:3051-3162`.

## 2. Technical

- **[Med] Audit reports do not fail on collected issues.** `TestDungeonGenerationAuditReport` explicitly says it is report-only (`server/internal/game/dungeon_generation_audit_test.go:20-22`); it records generation failures and invariant findings (`:43-55`) then writes the report without asserting either collection empty (`:59-70`). This keeps a diagnostic artifact complete, but cannot by itself serve as a gate. The v531 as-built correctly documents that boundary (`docs/as-built/v531_generation-audit.md:15-16`).
- **[Low] Per-floor audit data is local-only.** The report JSON is kept under ignored `.artifacts/` and the as-built describes the path and digest (`docs/specs/v531_spec-generation-audit.md:32-35`, `docs/plans/v531_2026-10-01-generation-audit.md:41-49`, `docs/as-built/v531_generation-audit.md:65`). The committed record has summary distributions/hash, but a clean clone cannot inspect or compare each floor.
- **[Strength] Shared validation passes 2,271 checks and CODEMAP validation.** That establishes schema/instance/catalog consistency for covered validators; it does not prove Go struct parity or clone-portable audit details.

## 3. Maintainability

- **[Med] `validate_shared.py` is 3,179 LOC.** It imports domain validators, maps schemas, performs cross-catalog checks, and dispatches validation in one manual coordinator (`:22-45,106-145,3051-3169`). Extract focused domain registration/dispatch while preserving order and diagnostics.
- **[Med] `tools/bot/run.py` is 4,534 LOC.** Keep new scenarios and shared execution logic modular rather than expanding the runner further.
- **[Med] The maintainability gate catches a new over-limit Go file.** `make maintainability` fails on `dungeon_encounter_composition.go` (1,055 LOC); split it before the next baseline.
- **[Strength] Prior v520 cleanup is present.** UiTheme token scanning is recursive and has a nested-script fixture (`tools/validate_ui_theme.py:38-53`, `tools/test_validate_ui_theme.py:121-129`); the CODEMAP includes `boss_bar_frame.gd` (`docs/CODEMAP.md:42`).

## 4. Documentation

All ten batch slices have specs, plans, as-built entries, and lifecycle tracking; progress now lists v531 and its audit limits. The root audit evidence includes the repeated integrated report digest and the preserved worker handoff report, but the report contents are ignored local artifacts. `make ci-full` was not run and is recorded as such.

## Top 5 shared/tooling/process refactors

1. **[minor-commit · Med] Add an explicit CI-friendly validator for audit reports.** Keep report generation diagnostic and complete; separately fail when a selected report contains generation failures/findings, with fixtures proving both cases.
2. **[future-plan · Med] Add static Go `json:` tag to schema member parity.** Include negative fixtures for missing and extra members, beginning with live protocol types.
3. **[future-plan · Med-Low] Extract `validate_shared.py` registration and domain dispatch.** Preserve deterministic check order, labels, and failure output.
4. **[future-plan · Med] Reduce bot-runner coordination concentration.** Extract shared execution primitives from `tools/bot/run.py` as a bounded tooling plan.
5. **[future-plan · Low] Preserve compact clone-portable audit receipts.** Store enough seed/depth/result or a validated compact digest to reproduce claims without committing bulky raw JSON.

*Evidence: citations above and current baseline files. Review checks and the exact `make ci`/`make ci-full` boundary are in the header.*
