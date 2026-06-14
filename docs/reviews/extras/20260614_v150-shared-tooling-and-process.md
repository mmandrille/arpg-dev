# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v150**

**Date:** 2026-06-14
**Scope:** Shared schemas/rules/goldens, Python protocol bot tooling, maintainability ratchet, CODEMAP, Make/CI, and SDD process after v141-v149.
**Baseline:** `main` at `1c7e50d7` (`feat: v149: bot coop runtime split`); worktree clean at review start.
**Stats:** 1216 repo files, 36 protocol schemas, 36 shared rule JSON files, 70 golden files, 103 protocol bot scenario JSON files, 40 client bot scenario JSON files, 147 specs, 147 plans, 148 as-built notes.
**Overview:** [`../20260614_v150-overview.md`](../20260614_v150-overview.md)

---

## Summary

The tooling/process story is stronger than at v140. The ratchet is now producing visible reductions, CODEMAP stayed current through the splits, and the Python bot has focused modules for assertion dispatch, movement, waits, state ingestion, and co-op peer helpers. The main remaining tooling risks are predictable: `tools/bot/run.py` is still broad, `runtime_assertions.py` contains long assertion chains, and `validate_shared.py` remains a large all-contract wrapper.

## 1. Architecture

[Strength] The Python bot is now decomposed along runtime responsibilities. `run.py` imports shared types and domain assertion helpers (`tools/bot/run.py:33`, `tools/bot/run.py:37`, `tools/bot/run.py:46`), while movement lives in `movement_runtime.py` (`tools/bot/movement_runtime.py:21`), waits/pumping live in `wait_runtime.py` (`tools/bot/wait_runtime.py:18`, `tools/bot/wait_runtime.py:65`), state ingestion lives in `state_ingest.py` (`tools/bot/state_ingest.py:14`, `tools/bot/state_ingest.py:197`), and co-op peer operations live in `coop_runtime.py` (`tools/bot/coop_runtime.py:21`, `tools/bot/coop_runtime.py:48`, `tools/bot/coop_runtime.py:78`).

[Strength] The compatibility-wrapper pattern let the bot shrink without invalidating scenario code. The co-op wrappers in `run.py` forward to `tools.bot.coop_runtime` with helper bindings (`tools/bot/run.py:3351`, `tools/bot/run.py:3363`, `tools/bot/run.py:3369`, `tools/bot/run.py:3375`, `tools/bot/run.py:3381`, `tools/bot/run.py:3411`).

[Strength] Shared validation is modular where recent domains needed it. `validate_shared.py` imports dedicated i18n, skills, and unique item validators (`tools/validate_shared.py:29`, `tools/validate_shared.py:30`, `tools/validate_shared.py:31`).

[Med] `run.py` remains the protocol bot coordinator at 4269 lines. It is smaller than v140 but still owns scenario bodies, helper binding glue, top-level orchestration, and compatibility wrappers.

## 2. Technical

[Strength] State ingestion moved into a focused parser/mutator module. It handles snapshots, intent accepts/rejects, state deltas, level changes, events, inventory, hotbar, stash, progression, cooldowns, and runtime distance refresh in one named place (`tools/bot/state_ingest.py:29`, `tools/bot/state_ingest.py:32`, `tools/bot/state_ingest.py:51`, `tools/bot/state_ingest.py:127`, `tools/bot/state_ingest.py:152`, `tools/bot/state_ingest.py:168`, `tools/bot/state_ingest.py:176`, `tools/bot/state_ingest.py:184`, `tools/bot/state_ingest.py:190`, `tools/bot/state_ingest.py:194`).

[Med] Python runtime assertions need a second split. `run_runtime_assertions` still initializes many domain helpers and then branches across movement, events, combat, level, inventory, stash, shop, progression, and skill assertions in one function (`tools/bot/runtime_assertions.py:209`, `tools/bot/runtime_assertions.py:239`, `tools/bot/runtime_assertions.py:243`, `tools/bot/runtime_assertions.py:269`, `tools/bot/runtime_assertions.py:280`, `tools/bot/runtime_assertions.py:286`, `tools/bot/runtime_assertions.py:402`).

[Med] `validate_shared.py` is stable but large at 3149 lines. It still owns schema discovery, schema mapping, cross-checks, and contract validation for shared protocol/rules/goldens/assets/content/i18n in one tool (`tools/validate_shared.py:2`, `tools/validate_shared.py:84`, `tools/validate_shared.py:112`, `tools/validate_shared.py:116`).

## 3. Maintainability

[Strength] The file-size ratchet is doing what v138 intended. The script blocks stale high baselines after shrinkage (`scripts/check-file-size-ratchet.sh:52`), blocks over-baseline growth (`scripts/check-file-size-ratchet.sh:78`), blocks new over-600 source/test/tool files (`scripts/check-file-size-ratchet.sh:83`), and prints the trend (`scripts/check-file-size-ratchet.sh:97`). At v150 it reports 33 grandfathered files / 65592 lines, down from 68778 at v140.

[Strength] SDD traceability stayed high. v141-v149 all have specs, plans, as-built notes, CODEMAP updates where needed, targeted checks, and full `make ci` closeout commits.

[Med] Review cadence should stay strict. v150 came at the correct gate after a concentrated architecture batch; the next review should be v160 before another long feature batch.

## 4. Documentation

[Strength] CODEMAP now covers the new backend/client/Python split files, making the loading path narrower for future agents (`docs/CODEMAP.md:7`, `docs/CODEMAP.md:17`, `docs/CODEMAP.md:24`).

[Low] The progress backlog should now convert review findings into future `/next` options rather than leaving the review as an archive.

## Top 5 shared/tooling/process refactors

1. Split runtime assertions again, starting with movement/combat assertions or inventory/shop/stash assertions.
2. Keep `run.py` as orchestration plus compatibility wrappers; move new bot runtime domains into focused modules.
3. Extract a catalog-specific validator from `validate_shared.py` when the next shared-data domain grows.
4. Keep lowering file-size baselines in every shrink slice and record the trend in reviews.
5. Keep CODEMAP updates mandatory for file moves and domain splits.

*Evidence: line/file counts from `find` and `wc -l`; `make maintainability` output; code references above; v150 closeout `make ci` green.*
