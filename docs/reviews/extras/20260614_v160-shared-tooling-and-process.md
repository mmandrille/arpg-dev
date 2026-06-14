# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v160**

**Date:** 2026-06-14
**Scope:** Shared protocol/rules/goldens, Python protocol bot, validation tooling, maintainability gates, and SDD process after v151-v159.
**Baseline:** `main` at `4a46229e` (`feat: v159: kill-gated elite objective`); worktree clean at review start.
**Stats:** 3971 repo files, 72 `shared/protocol` files, 36 `shared/rules` files, 70 `shared/golden` files, 110 protocol bot scenario JSON files, 40 client bot scenario JSON files, 156 specs, 156 plans, 157 as-built notes.
**Overview:** [`../20260614_v160-overview.md`](../20260614_v160-overview.md)

---

## Summary

The process held up through a mixed maintenance and gameplay batch. v151 added an extraction-coupling ratchet, v152 reduced bot movement context coupling, and v153-v159 delivered feature slices with specs, plans, as-builts, targeted checks, and full CI closeout. The main tooling risks remain Python concentration: `run.py`, `runtime_assertions.py`, and `validate_shared.py` still centralize too many domains.

## 1. Architecture

[Strength] Maintainability now checks both file size and extraction independence. `make maintainability` reports 33 grandfathered files / 65747 lines and 37 legacy helper-global injections, giving future extraction slices two concrete trends to improve.

[Strength] SDD traceability is current. The repo now has 156 specs, 156 plans, and 157 as-built notes, and the recent v159 docs explain the objective gate rule, deferred full elite clearing, and exact protocol bot proof.

[Med] `tools/bot/run.py` is still the broad protocol bot coordinator at 4294 lines. It continues to own scenario execution, action handling, session orchestration, and compatibility wrappers (`tools/bot/run.py:639`, `tools/bot/run.py:3338`).

## 2. Technical

[Strength] Protocol bot coverage kept pace with the feature batch. The scenario catalog grew to 110 protocol JSON scenarios, including `68_dungeon_elite_side_objective`, which now proves objective chest rejection, leader kill, and successful opening.

[Med] `runtime_assertions.py` needs another domain split. `run_assertions` and `run_runtime_assertions` bind many helper functions through a single `helpers` object and then branch across inventory, gold, stash, walls, entities, hotbar, progression, movement, events, combat, levels, and skills (`tools/bot/runtime_assertions.py:6`, `tools/bot/runtime_assertions.py:65`, `tools/bot/runtime_assertions.py:209`).

[Med] `validate_shared.py` remains a broad all-contract validator at 3169 lines. It has modular validators for some domains (`tools/validate_shared.py:23`), but central schema discovery and cross-contract checks still live in one script (`tools/validate_shared.py:84`).

## 3. Maintainability

[Strength] The ratchets are influencing day-to-day work. v159 lowered the `sim.go` file-size baseline and trimmed unrelated `inventory_panel.gd` drift rather than leaving maintainability output stale.

[Med] Scenario runtime budget is worth watching. The v159 protocol scenario stayed inside the budget, but the objective proof is naturally travel-and-combat heavy. Future scenario additions should use stable selectors and shortest-path setup rather than adding long waits.

[Med] Helper-global compatibility remains the main extraction smell. v151 correctly froze the pattern by ratchet, but true improvement still requires typed context movement rather than more namespace forwarding.

## 4. Documentation

[Strength] `PROGRESS.md` is current on completed slices, deferred non-goals, and the review cadence. The autoloop queue is visible, but the v160 review should steer the next batch before more feature work proceeds.

[Low] CODEMAP updates should remain mandatory for any extraction that changes where agents should start reading. Small adjacent files like v159 `interactables.go` can ride the existing domain map, but larger moves should update it.

## Top 5 shared/tooling/process refactors

1. Split protocol bot runtime assertions by domain, starting with inventory/shop/stash or movement/combat.
2. Replace helper-global bot runtime wrappers with typed context plumbing in the next real bot extraction slice.
3. Extract one catalog-specific validator from `validate_shared.py` when the next shared data family grows.
4. Keep protocol scenarios below the timeout budget by preferring direct setup, stable selectors, and narrow assertions.
5. Continue recording maintainability totals and lowering baselines in the same slice as any shrinkage.

*Evidence: line/file counts from `find` and `wc -l`; `make maintainability`; code references above; v159 closeout `make ci` green.*
