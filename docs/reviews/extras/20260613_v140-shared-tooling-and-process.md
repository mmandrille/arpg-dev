# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v140**

**Date:** 2026-06-13
**Scope:** `shared/`, Python tools, protocol/scenario contracts, Make/CI, SDD docs, CODEMAP, and review cadence after v131-v139.
**Baseline:** `main` at `15132acc` (`feat: v139: market expiration read freshness`); worktree clean at review start.
**Stats:** 1164 tracked files, 36 protocol schemas, 36 shared rule JSON files, 70 golden files, 103 protocol bot scenario JSON files, 138 specs, 138 plans, 139 as-built notes.
**Overview:** [`../20260613_v140-overview.md`](../20260613_v140-overview.md)

---

## Summary

The v131-v139 batch acted directly on v130 review recommendations. Unique validation is split out,
unique-effect bot assertions are data-derived, stash assertions moved into a helper, CODEMAP gives
agents a focused-loading index, and the file-size ratchet is now part of CI. The remaining tooling
risks are the same large but now better-contained surfaces: `tools/bot/run.py`, `validate_shared.py`,
and the client bot runner.

## 1. Architecture

[Strength] Shared validation is increasingly modular. `validate_shared.py` imports dedicated
validators for i18n, skills, and unique items (`tools/validate_shared.py:29`,
`tools/validate_shared.py:30`, `tools/validate_shared.py:31`), and unique item validation now lives
in `tools/validate_unique_items.py`.

[Strength] Unique catalog validation now checks behavior-bearing invariants: enabled entries must
be ready, fixed stats/effects must exist, effects must be active, and effects must match the base
template item type (`tools/validate_unique_items.py:37`, `tools/validate_unique_items.py:49`,
`tools/validate_unique_items.py:56`, `tools/validate_unique_items.py:74`,
`tools/validate_unique_items.py:77`).

[Strength] CODEMAP is validated by tooling. The document says every path is checked
(`docs/CODEMAP.md:5`), and `tools/validate_codemap.py` rejects missing path references and domain
rows without paths (`tools/validate_codemap.py:21`, `tools/validate_codemap.py:30`).

## 2. Technical

[Strength] Bot assertions are moving out of the monolith. Stash filtering and event/count assertions
live in `tools/bot/stash_assertions.py` (`tools/bot/stash_assertions.py:11`,
`tools/bot/stash_assertions.py:44`, `tools/bot/stash_assertions.py:74`), and unique-effect coverage
derives expected enabled effects from `shared/rules/unique_effects.v0.json`
(`tools/bot/unique_effect_assertions.py:12`, `tools/bot/unique_effect_assertions.py:24`).

[Med] `tools/bot/run.py` remains broad at 5179 lines. It now imports focused helpers, but runtime
assertion dispatch still contains many domain cases in one function (`tools/bot/run.py:3656`,
`tools/bot/run.py:3693`, `tools/bot/run.py:3735`, `tools/bot/run.py:3782`). The next bot-domain
growth should follow the stash/unique helper pattern.

[Med] `validate_shared.py` remains 3149 lines even after skill and unique splits. It is stable and
green, but the next schema-heavy domain should extract another validator instead of expanding the
wrapper.

## 3. Maintainability

[Strength] The reduction ratchet is now real CI behavior. It checks new source/test/tool files
against the 600-line target, applies grandfathered allowances, detects when baselines should be
lowered, and prints the grandfathered trend (`scripts/check-file-size-ratchet.sh:52`,
`scripts/check-file-size-ratchet.sh:78`, `scripts/check-file-size-ratchet.sh:83`,
`scripts/check-file-size-ratchet.sh:97`).

[Strength] SDD traceability remains strong: v131-v139 all have specs, plans, as-built notes, and
commits, and the current review is being written at the declared cadence gate.

[Low] The repo has generated Python `__pycache__` files under `tools/bot` in the working tree during
inspection output, but they are not tracked by git status at this baseline. Keep avoiding local
artifact staging.

## 4. Documentation

[Strength] `docs/CODEMAP.md` is now the main loading guide for agents and covers market, unique
items, classes, skills, stash, shops, dungeon generation, combat, realtime, replay, persistence,
protocol, bot, assets, and i18n (`docs/CODEMAP.md:7` through `docs/CODEMAP.md:26`).

[Med] Review findings should now feed directly into `/next`: market repo extraction, client bot
runner split, Python bot assertion split, and `main.gd` bot facade are all small enough to become
bounded maintenance slices.

## Top 5 shared/tooling/process refactors

1. Split the next domain out of `tools/bot/run.py` before adding more runtime assertion branches.
2. Extract another domain validator from `tools/validate_shared.py` when the next rules catalog grows.
3. Keep CODEMAP rows current in any slice that creates, moves, or deletes domain files.
4. Treat `make maintainability` failures as design feedback, not a baseline-update reflex.
5. Continue writing review sets every 10 slices and use findings as `/next` input.

*Evidence: line/file counts from `find` and `wc -l`; code references above; `make ci` run after v139.*
