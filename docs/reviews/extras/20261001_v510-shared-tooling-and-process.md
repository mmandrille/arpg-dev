# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v510**

**Date:** 2026-10-01  
**Scope:** shared contracts/data, validators, bot tooling, CI orchestration, assets, docs, and SDD lifecycle; v501–v510.  
**Baseline:** `main` @ `36d952f5`; clean at review start.  
**Stats:** 35 protocol schema files; 74 golden files; 286 bot scenario definitions; `tools/bot/run.py` 4,534 lines (6 fewer than v500); `tools/validate_shared.py` 3,170 lines (18 fewer). `PROGRESS.md` is 249/250 lines.  
**Overview:** [`../20261001_v510-overview.md`](../20261001_v510-overview.md)

---

## Summary

The v501–v510 batch has complete spec → plan → as-built coverage, a clean integration ledger, and a combined `make ci` pass. Shared catalogs remained schema-backed, the coordinator reconciled multi-slice data changes, and batch worktrees were removed after their integrated changes and evidence were checked. The extended gate provides useful extra coverage but is not green: it reports a single protocol failure (`teleporter_lab`) among 149 scenarios; the same scenario passed when rerun alone. Client bot and headless smoke gates passed.

## 1. Architecture

**Medium — cold CI may silently skip database-backed Go tests.** `scripts/ci.sh:314-322` calls `test_db.sh ensure` before Go tests and unsets the database URL if that fails; `scripts/test_db.sh:36-59` can create a database only through an already-running container. Postgres starts later in CI step 8 (`ci.sh:335-340`). Thus a cold run can print a successful step 6 while omitting store/HTTP database coverage. Start the test DB before those tests and fail the gate if provisioning fails.

**Medium — reconnect snapshots bypass the strict schema check.** The live protocol gate validates scenario snapshot/delta streams (`ci.sh:396-402`), but bot reconnect assertions read a snapshot directly (`tools/bot/run.py:2867-2880`). A malformed reconnect payload can therefore evade the same contract check.

**Medium — protocol Go JSON tags have no drift gate against v8 schemas.** `validate_shared.py` verifies the presence and selected invariants of schema files (`:339-389`), but does not compare Go serialization tags with those schemas. The structural gap could let one side change without the other's validator noticing.

## 2. Technical

**Strength — shared-data and CI coverage.** Schema validation, asset validation, determinism checks, Go/Python tests, protocol replay, 133 Godot bot scenarios, and GDScript smoke all ran. Batch evidence distinguishes a full `make ci` pass from `make ci-full`'s one protocol failure and from unmeasured visual/performance claims.

**Medium — two Python coordinators remain large.** `run.py` is 4,534 lines and retains a large step dispatcher; `validate_shared.py` is 3,170 lines with broad cross-catalog validation. Both shrank slightly from v500 but remain hard to navigate. Extract typed, domain-owned registries with focused tests rather than a mechanical split or reflective `helpers=globals()` pattern.

**Low — final-state visual proof remains incomplete.** v506 memory sampling and v508/v509 matched cost checks remain open in their as-built plans. Bot success establishes behavior for those scenarios, not performance or real-play readability.

## 3. Maintainability

The file-size and extraction-coupling ratchets pass; `make maintainability` passed at 249/250 `PROGRESS.md` lines. The touch-to-shrink gap persists: 25 of 35 grandfathered files are above their own baseline, by 328 lines in total. `client/main.gd` and the two Python coordinators remain extraction candidates; avoid spending additional allowance without shrinking implicated files when they are touched.

## 4. Documentation

All ten slices have corresponding spec, plan, as-built, and lifecycle records. The `PROGRESS.md` line budget leaves only one line, so updates should replace or condense existing entries rather than adding another row. Existing older orphaned slice documentation and environment/remote-CI follow-ups were not re-audited in this pass.

## Top 5 shared/tooling/process refactors

1. Ensure database-backed Go tests cannot report green when the test database was never started.
2. Apply the strict v8 schema validator to reconnect snapshots as well as initial scenario streams.
3. Add a focused Go JSON-tag/schema drift check for the serialized protocol types in scope.
4. Split bot action dispatch and shared-catalog checks into typed domain modules, each with a narrow test seam.
5. Keep v506/v508/v509 evidence limits visible and collect matched, final-state samples before making performance claims.

*Evidence: `make ci` passed in 11m41s. `make ci-full` took 32m03s and failed only `teleporter_lab`; `make bot scenario=teleporter_lab` passed on isolated rerun. The CI step order and reconnect code were inspected directly; the run used an already-ready Postgres instance, so it does not disprove the cold-start skip.*
