# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v250**

**Date:** 2026-06-17
**Scope:** Shared protocol/rules/goldens, Python validators and bot tooling, CI orchestration, progress docs, SDD cadence, and review/refactor handoff after the selected v241-v250 autoloop queue.
**Baseline:** `main` at `2f929d09` (`feat: v250: boss-specific telegraph decals`). Worktree was clean before this review started.
**Stats:** `shared/protocol` 14,411 lines, `shared/rules` 7,336 lines, 70 golden JSON files, 160 protocol bot scenarios, 66 client bot scenarios, `tools` 20,814 lines, `docs/specs` 28,876 lines, `docs/plans` 33,329 lines, `docs/as-built` 7,278 lines.
**Overview:** [`../20260617_v250-overview.md`](../20260617_v250-overview.md)

---

## Summary

Shared/tooling/process is strong at v250. The selected queue produced specs, plans, focused checks,
as-built notes, 10 feature commits for v241-v250, and a full batch `make ci` pass. Review-time
`make validate-shared` passed 1,295 checks plus codemap validation, and `make maintainability`
passed all ratchets.

The prior review/refactor ordering drift is fixed in the current entrypoint docs: `PROGRESS.md`,
`CLAUDE.md`, and `AGENTS.md` all say `$review` writes the fresh scorecard before `$refactor` pays
down recommendations. The remaining process debt is narrower: per-slice proof docs still say batch
CI is pending, `tools/bot/run.py` has 14 helper-global injections, and `tools/validate_shared.py`
remains a broad validator.

## 1. Architecture

- **[Strength] Shared contracts remain the integration layer.** `make validate-shared` passed schema
  meta-validation, instance validation, cross-consistency checks, shared goldens, assets, i18n, and
  codemap validation.
- **[Strength] v247 added companion combat stats without protocol churn beyond the current v8 schema.**
  `session_snapshot.v8.schema.json` includes `combat_stats` and a `companion_combat_stats` definition
  (`shared/protocol/session_snapshot.v8.schema.json:121`,
  `shared/protocol/session_snapshot.v8.schema.json:850`).
- **[Med] AI tuning vocabulary is now data-driven but shared between domains.** Companion follow
  and assist values live in `main_config` and are validated by Go, but elite minion formation still
  references companion-named fields. This should become an explicit rules-design task when AI tuning
  is next in scope.

## 2. Technical

- **[Strength] Batch and review gates are green.** The final selected v241-v250 `make ci` passed
  after v250 in 10m27s. During review, `make validate-shared` passed 1,295 checks and
  `make maintainability` passed.
- **[Strength] Client smoke prerequisites are documented.** `make/client.mk` now describes
  `client-smoke` as requiring a running `TEST_BASE_URL` server (`make/client.mk:19`), and
  `CLAUDE.md` repeats that prerequisite (`CLAUDE.md:31`).
- **[Med] Client bot debug progression JSON is useful but undertested at the helper boundary.**
  `scripts/bot_client.sh` forwards the full `debug_progression` object into Godot
  (`scripts/bot_client.sh:264`, `scripts/bot_client.sh:299`), and Godot normalizes nested numeric
  maps before the debug endpoint call (`client/scripts/net_client.gd:160`, `client/scripts/net_client.gd:172`).
  Current proof comes from live boss scenarios; a focused unit would reduce future debugging cost.

## 3. Maintainability

- **[Strength] Ratchets are doing their job.** `make maintainability` passed, reporting 34
  grandfathered files / 63,983 lines and one extraction-coupling baseline file with 14 helper
  injections.
- **[Strength] Helper-global debt dropped materially since v232.** The v232 review recorded 31
  `helpers=globals()`/`globals()` helper injections. The current baseline is 14 in `tools/bot/run.py`
  (`.maintainability/extraction-coupling-baseline.tsv:1`).
- **[Med] `tools/bot/run.py` still has the remaining helper-global cluster.** The remaining sites
  include direct action wrappers (`tools/bot/run.py:474`, `tools/bot/run.py:1328`), wait/pump helper
  calls (`tools/bot/run.py:2168`, `tools/bot/run.py:2199`), runtime assertions
  (`tools/bot/run.py:3186`), and coop helpers (`tools/bot/run.py:3313`).
- **[Med] `tools/validate_shared.py` is still broad.** At 3,030 lines, it remains the natural split
  target when the next rule/golden/content validation domain is touched.

## 4. Documentation

- **[Strength] Agent workflow docs now agree on review/refactor order.** `PROGRESS.md` says `$review`
  first and then `$refactor` (`PROGRESS.md:39`), `CLAUDE.md` says the same for ~10-slice milestones
  (`CLAUDE.md:187`), and `AGENTS.md` aligns the slash-command workflow (`AGENTS.md:44`).
- **[Med] Batch proof wording is stale in slice docs.** The lifecycle rows for v241-v250 still say
  batch `make ci` is pending (`docs/progress/slice-lifecycle.md:257`), and the as-built proof notes
  use the same deferred wording. Since CI passed after v250, this is a low-risk documentation repair.
- **[Low] `PROGRESS.md` remains compact.** The dashboard check reports 185/250 lines.

## Top 5 shared/tooling/process refactors

1. Update v241-v250 lifecycle/as-built proof wording to record the passed batch `make ci`.
2. Add focused GDScript unit coverage for debug progression JSON parsing and integer map coercion.
3. Replace the next `tools/bot/run.py` helper-global cluster with typed context/direct helper calls,
   then lower `.maintainability/extraction-coupling-baseline.tsv`.
4. Split the next touched validator domain out of `tools/validate_shared.py`.
5. Audit tests/scenarios for accidental tuning pins and convert safe cases to rule-derived,
   semantic, range, or eventual assertions.

*Evidence: review-time `make validate-shared` and `make maintainability`; prior batch `make ci`;
current counts from `find`, `wc -l`, and `rg`; code references from `rg -n` and `nl -ba`.*
