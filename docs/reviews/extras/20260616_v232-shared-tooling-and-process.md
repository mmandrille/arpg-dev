# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v232**

**Date:** 2026-06-16
**Scope:** Shared protocol/rules/goldens, Python validators and bot tooling, CI orchestration, progress docs, SDD cadence, and review/refactor handoff after the v226-v232 autoloop queue.
**Baseline:** `main` at `cfae54ca` (`docs: mark v226-v232 ci green`). Worktree was clean before this review started.
**Stats:** 73 protocol JSON files, 38 shared rule JSON files, 70 golden JSON files, 93 protocol bot scenarios, 49 client bot scenarios, 55 Python files. `PROGRESS.md` is 184 lines.
**Overview:** [`../20260616_v232-overview.md`](../20260616_v232-overview.md)

---

## Summary

Shared/tooling/process remains strong at v232. The v226-v232 queue produced specs, plans, as-built
notes, focused checks, seven feature commits, a batch `make ci` pass, and a docs closeout. Current
review checks passed `make validate-shared`, `.venv/bin/pytest tools`, and `make maintainability`.

The v223 progress-link finding is resolved: `scripts/check-progress-dashboard.sh` now rejects
root-style `](docs/...)` links inside `docs/progress/`. The extraction-coupling baseline also
improved from 37 to 31 `helpers=globals()` calls. Remaining process/tooling debt is narrower:
entrypoint docs disagree on review/refactor order, `tools/bot/run.py` still owns one helper-global
cluster, and `tools/validate_shared.py` remains a broad validator.

## 1. Architecture

- **[Strength] Shared contracts are still the authoritative integration layer.** `make validate-shared`
  passed 1,286 checks plus `tools/validate_codemap.py`, covering schemas, examples, shared rules,
  goldens, assets, text catalogs, and codemap paths.
- **[Strength] Resource wallet and market features used existing protocol/tooling shape.** The
  queue added scenarios without introducing a new protocol version or bypassing server state.
- **[Med] Companion tuning still bypasses shared data.** The same companion/elite-minion constants
  from the backend report remain outside schemas/goldens.

## 2. Technical

- **[Strength] Progress link validation is now enforced.** `scripts/check-progress-dashboard.sh`
  rejects `](docs/...)` links under `docs/progress/` (`scripts/check-progress-dashboard.sh:39`),
  and `make maintainability` wires the script through the ratchet (`make/ci.mk:3`).
- **[Strength] Python tests are green.** `.venv/bin/pytest tools` passed 122 tests.
- **[Med] One standalone smoke command has implicit environment needs.** `make client-smoke` failed
  at live-login without a server on `localhost:18081`; this is not a code regression, but the
  review records it because the target name can be misleading outside CI.

## 3. Maintainability

- **[Strength] Extraction-coupling debt moved down.** The baseline is now one file with 31 helper
  injections (`.maintainability/extraction-coupling-baseline.tsv:1`), down from 37 at v223.
- **[Med] `tools/bot/run.py` still has a helper-global cluster.** Remaining calls pass
  `helpers=globals()` into wait/state/coop helper modules (`tools/bot/run.py:2005`,
  `tools/bot/run.py:2284`, `tools/bot/run.py:3392`).
- **[Med] `tools/validate_shared.py` is still broad.** At 3,040 lines, it validates many unrelated
  catalog/golden domains in one script (`tools/validate_shared.py:1`, `tools/validate_shared.py:2145`,
  `tools/validate_shared.py:2680`).

## 4. Documentation

- **[Strength] SDD cadence is complete for the batch.** v226-v232 have specs, plans, lifecycle rows,
  as-built notes, focused verification, and final CI proof.
- **[Med] Review/refactor ordering is inconsistent across agent docs.** `PROGRESS.md` says `$review`
  first and then `$refactor` (`PROGRESS.md:39`), but `CLAUDE.md` and `AGENTS.md` still include the
  older `$refactor` then `$review` wording.
- **[Low] `PROGRESS.md` stayed compact.** The dashboard is 184/250 lines.

## Top 5 shared/tooling/process refactors

1. Align `CLAUDE.md` and `AGENTS.md` with the current `$review` then `$refactor` workflow.
2. Replace the next `helpers=globals()` wait-runtime cluster with typed context/direct imports and
   lower the extraction-coupling baseline.
3. Split the next touched rarity/item/unique validator out of `tools/validate_shared.py`.
4. Move companion/elite-minion tuning into shared rules with schema/golden validation.
5. Clarify the standalone `make client-smoke` server prerequisite if agents keep running it outside
   CI.

*Evidence: current checks listed in the overview; current counts from `find`, `wc -l`, and `rg`; code references from `nl -ba`.*
