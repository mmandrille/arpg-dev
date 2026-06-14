# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v170**

**Date:** 2026-06-14
**Scope:** Shared contracts, Python protocol bot/runtime assertions, shared validator splits, maintainability ratchets, and SDD process after v161-v170.
**Baseline:** `main` at `05804d77` (`feat: v170: validate shared catalog split`); worktree clean at review start.
**Stats:** 21520 repo files, 72 `shared/protocol` files, 36 `shared/rules` files, 70 `shared/golden` files, 111 protocol bot scenario JSON files, 167 specs, 167 plans, 168 as-built notes, 33 grandfathered files / 65347 lines, 37 legacy helper-global injections.
**Overview:** [`../20260614_v170-overview.md`](../20260614_v170-overview.md)

---

## Summary

The shared/tooling process continued to improve while feature work shipped. v167 split shop/stash economy runtime assertions out of `runtime_assertions.py`, and v170 split `main_config` gameplay validation out of `validate_shared.py`. Full CI at v170 passed, including 1128 shared validation checks, 111 protocol scenarios, replay, 41 client bot scenarios, and Godot smoke.

## 1. Architecture

[Strength] `validate_shared.py` now delegates another catalog domain. It imports `validate_main_config_gameplay` (`tools/validate_shared.py:30`) and calls it from the main cross-check flow (`tools/validate_shared.py:1155`), while keeping the CLI entrypoint and report shape unchanged (`tools/validate_shared.py:3126`).

[Strength] The extracted main-config validator is independently importable and narrow. `validate_main_config_gameplay` receives explicit inputs and a narrow table resolver (`tools/validate_main_config.py:6`), satisfying the extraction-independence rule without `globals()` forwarding.

[Med] `tools/bot/run.py` remains the broadest Python coordinator at 4294 lines. It still combines action execution such as `attack_until_event` (`tools/bot/run.py:639`) with scenario drive/state/API/replay orchestration (`tools/bot/run.py:3338`).

## 2. Technical

[Strength] Runtime economy assertions now have a focused helper. `runtime_assertions.py` delegates through `handle_runtime_economy_assertion` (`tools/bot/runtime_assertions.py:5`), and the helper owns stash/shop counts, details, appraisals, and events (`tools/bot/runtime_economy_assertions.py:6`).

[Med] Helper-global compatibility is still present in runtime assertions. `run_assertions` requires helper bindings and pulls many functions out of the `helpers` dict (`tools/bot/runtime_assertions.py:29`, `tools/bot/runtime_assertions.py:34`); the ratchet prevents growth but this remains the main process debt.

[Med] `validate_shared.py` is still 3140 lines. The new `main_config` split is good, but combat, dungeon generation, shops, worlds, item visuals, and item presentations still make the script a broad cross-contract coordinator.

## 3. Maintainability

[Strength] The ratchet trend improved across the batch. `make maintainability` reports 33 grandfathered files / 65347 lines, down from 65747 at v160. Helper-global occurrences remained flat at 37.

[Strength] v170 locked in the shared-validator shrink. `validate_shared.py` dropped from 3169 to 3140 lines, with the baseline lowered from 3149 to 3140.

[Med] Scenario count keeps increasing. Protocol scenarios are now 111, and CI still passes within a reasonable time, but new scenario proofs should prefer direct setup and narrow assertions.

## 4. Documentation

[Strength] SDD cadence stayed current. The repo has 167 specs, 167 plans, and 168 as-built notes, with v165-v170 documenting each paydown slice and verification command.

[Low] The review cadence is working. `PROGRESS.md` correctly blocked more feature batches until this v170 review landed.

## Top 5 shared/tooling/process refactors

1. Replace one `runtime_assertions.py` helper-global family with typed context plumbing instead of only splitting by file.
2. Split another `validate_shared.py` catalog domain, preferably shops, dungeon generation, or item presentations.
3. Extract a focused action-execution module from `tools/bot/run.py` when a protocol scenario slice touches action loops.
4. Keep maintainability totals in every review and lower baselines in the same slice as shrinkage.
5. Keep protocol bot additions scenario-budget conscious: direct setup, stable selectors, and assertions scoped to the new contract.

*Evidence: file:line references above; `find`/`wc -l`; `make maintainability`; full `make ci` passed at `05804d77`.*
