# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v190**

**Date:** 2026-06-15
**Scope:** Shared protocol/rules/goldens, Python validation and bot tooling, CI orchestration, SDD/review cadence.
**Baseline:** `main` at `0ec2d8c6` (`docs: remove stale Godot plugin research requirement`); worktree clean before review docs were written.
**Stats:** 36 protocol schemas, 38 shared rule JSON files, 70 golden files, 82 protocol bot scenarios, 45 client bot scenarios, 54 Python files.
**Overview:** [`../20260615_v190-overview.md`](../20260615_v190-overview.md)

---

## Summary

Shared contracts and process are in good shape. v181-v190 expanded the rule surface with set items, companion world entries, rarity roll ranges, rare/skill affixes, and Paladin defensive effects while keeping `make validate-shared` and full `make ci` green. The pre-review `$refactor` pass also repaired two process issues: it lowered a stale maintainability baseline and removed a deleted Godot plugin research requirement from active workflow docs.

The largest risks are validator concentration and a few tuning locks that still live in implementation code rather than schema/golden-owned checks.

## 1. Architecture

- **[Strength] Protocol v8 explicitly admits companions as stateful server entities.** The v8 snapshot schema includes `companion` in entity types (`shared/protocol/session_snapshot.v8.schema.json:80`), and world presets can seed companion entities (`shared/rules/worlds.v0.schema.json:36`, `shared/rules/worlds.v0.json:861`).
- **[Strength] Skill and item tuning continue to live in shared JSON.** Holy Shield/Sanctuary effect radius/duration/cooldown live in `shared/rules/skills.v0.json:515` and `shared/rules/skills.v0.json:579`; rarity roll ranges and `set.random_rollable=false` live in `shared/rules/item_templates.v0.json:3`.
- **[Med] Not all gameplay tuning is equally data-owned yet.** Companion follow/assist radii live in Go constants, and monster-rarity expected weights/colors/scales are hardcoded in loader validation (`server/internal/game/companion_ai.go:5`, `server/internal/game/rules.go:2247`).

## 2. Technical

- **[Strength] CI reporting improved materially before review.** `scripts/ci.sh` now runs all steps without fast-failing and prints durations/summaries (`scripts/ci.sh:2`, `scripts/ci.sh:58`, `scripts/ci.sh:95`), which made this review's `make ci` evidence much easier to interpret.
- **[Strength] Scenario coverage tracks the latest feature batch.** Protocol bot scenarios now include companion foundation/rank/elite-minion/rarity/rare-affix/skill-affix coverage (`tools/bot/scenarios/73_companion_ai_foundation.json:1`, `tools/bot/scenarios/76_companion_rank_scaling_and_limits.json:1`, `tools/bot/scenarios/77_elite_minion_pack_ai.json:1`, `tools/bot/scenarios/78_rarity_roll_pools.json:1`).
- **[Med] `tools/validate_shared.py` remains the broadest validation coordinator.** It imports some extracted validators (`tools/validate_shared.py:29`), but still contains item rarity, shop, monster rarity, golden, and unique-effect cross-checks in a 3,041-line file (`tools/validate_shared.py:906`, `tools/validate_shared.py:1357`, `tools/validate_shared.py:2982`).

## 3. Maintainability

- **[Strength] Maintainability ratchets are active and useful.** The pre-review `3f5f8a47` commit lowered `game_test.go`'s baseline after the file dropped to 7,888 lines; `make maintainability` now reports 33 grandfathered files and 63,565 grandfathered lines.
- **[Med] Coupled helper injections remain unchanged.** `make maintainability` reports one file with 37 coupled helper injections. This is grandfathered debt and likely larger than a minor `$refactor` task, but it should remain visible in future review scorecards.
- **[Low] Pycache artifacts exist locally under `tools/__pycache__`, but they are not tracked.** This did not affect review output; keep them ignored and out of commits.

## 4. Documentation

- **[Strength] SDD cadence was followed for the official review gate.** `PROGRESS.md` marked v190 review due, `$refactor` ran first, and this review updates the cadence to v200.
- **[Strength] Active workflow docs no longer reference the deleted Godot plugin research note.** The docs now require checking existing in-repo assets/code and recording an adopt/borrow/reject decision.

## Top 5 shared/tooling/process refactors

1. Extract a `tools/validate_set_items.py` or `tools/validate_item_rarity.py` module when the next itemization slice touches those rules.
2. Move companion AI tuning into shared rules and validate it through schema plus a semantic bot/unit test.
3. Convert hardcoded monster-rarity expected values in Go validation into explicit golden/schema checks.
4. Continue reducing `tools/validate_shared.py` by domain instead of adding new cross-checks directly.
5. Keep the CI summary behavior; it is now useful review evidence and should not regress to fast-fail-only output.

*Evidence: `make ci` passed in 7m03s; `make maintainability` passed inside CI with 33 grandfathered files / 63,565 lines and 37 coupled helper injections; stale Godot plugin references are gone from `AGENTS.md` and `skills/`.*
