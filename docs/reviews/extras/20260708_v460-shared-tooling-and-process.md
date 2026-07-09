# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v460**

**Date:** 2026-07-08
**Scope:** `shared/`, `tools/`, Makefile/CI, SDD docs — v449–v460 vs v448 baseline.
**Baseline:** `main` @ `e75e64d0`. v448 baseline: `ebcbdc9a`. Worktree: **clean**.
**Stats:** **262** scenarios (150 protocol + 112 client, up from 248); CI pack **35** (22 protocol + 13 client, unchanged); `run.py` **4,526** lines; `validate_shared.py` **3,186** lines; protocol still **v8**.
**Overview:** [`../20260708_v460-overview.md`](../20260708_v460-overview.md)

---

## Summary

Protocol v8 remained stable while twelve slices extended it in-place (resource bag ops, `skill_damage_burst`, class affinity status, leveled consumable `item_level`). **v448 recommendation #2 closed:** `quest_steward.v0.json` cross-checks fully implemented (`validate_shared.py:3081–3102`). **`validate_unique_items_catalog`** extracted to `tools/validate_unique_items.py` (84 lines) — first structural win from `cross_checks()`.

Gaps widened: **`status_effects.v0.json` is not loaded** by `validate_shared.py` — unique-effect `*_status_id` references are unvalidated. **Leveled potion formulas lack cross-language golden coverage** beyond level 1. **`cross_checks()` body** remains a ~2,971-line monolith. **`run.py` grew +88 lines** to its ceiling. SDD trail has holes: v454–v456 skipped plans, v458 has no as-built/lifecycle row, scenario catalog ~14 entries stale.

---

## 1. Contract/schema health

**[Strength]** Protocol v8 discipline intact; backward-compatible in-place extensions documented in slice as-builts.

**[Strength]** `state_delta.v8.schema.json` covers `skill_damage_burst`, `resource_bag_item_add`, `class_affinity_status_list`.

**[Strength]** `skills.v0.schema.json` constrains `resolution` and `targeting` enums with `if/then` rules.

**[High]** `status_effects.v0.json` not loaded in `validate_shared.py`. `unique_effects.v0.json` references (`dot_status_id`, `burn_status_id`, etc.) have zero catalog resolution — broken IDs fail silently at runtime.

**[Med]** `validate_shared.py` 4 lines above baseline (3,186 vs 3,182).

**[Closed]** Quest steward cross-checks (v448 #2).

---

## 2. Golden coverage

**[High]** `use_consumable.json` covers level-1 red potion only. v460 `3 × item_level` formula and rejuv `max(33%, level%)` have no cross-language golden cases.

**[Med]** No golden for `skill_damage_burst` wire shape or `affinity_scaling` formula (v451).

**[Low]** `dungeon_teleporters` golden drift noted at v451; unfixed.

---

## 3. Python bot/validator shape

**[Strength]** `skill_cast_runtime.py` (89 lines) — clean extraction, no `helpers=globals()`.

**[Med]** `run.py` +88 lines (4,438 → 4,526); ceiling advanced not extracted.

**[Med]** `combat_soak_runtime.py` uses `_combat_soak_helpers()` dict bridge — softer extraction than `skill_cast_runtime.py`.

**[Open]** `quest_steward_pick` still inline at `run.py:1584–1603` (v448 #3).

**[Open]** `cross_checks()` ~2,971 lines; only `validate_unique_items` structurally extracted.

**[Low]** No reusable `assert_status_effect_active()` in `runtime_assertions.py`.

---

## 4. SDD process

**[Strength]** 11/12 slices have as-builts (v449–v460 except v458).

**[Med]** v454, v455, v456 skipped plan phase (combat-stability chain).

**[Med]** v458 (`camera-follow-smoothing-fix`) implemented but no as-built, no lifecycle row; spec still says "awaiting `/finish`".

**[Med]** `docs/progress/scenario-catalog.md` missing ~14 scenarios from v449–v460.

**[Strength]** ADR-0016 and ADR-0017 landed.

**[Open]** Quest system ADR (v448).

---

## 5. CI pack curation

**[Strength]** `validate_ci_pack()` PASS; all new scenarios correctly `extended`.

**[Neutral]** Pack static for 12 slices (35 total). No promotion/demotion.

**[Open]** Three pre-existing extended failures from v448 unchanged.

**[Note]** `ranged_lab` failure observed during this review's `make ci-full` run — verify if new regression vs flaky.

---

## Top 5 shared/tooling refactors

1. **[High · Correctness]** Load `status_effects.v0.json` in `validate_shared` and cross-validate `unique_effects` `*_status_id` refs via `_check_status_effects_refs()`.

2. **[High · Correctness]** Extend `use_consumable.json` golden for leveled potions (levels 5, 10, capped heal, rejuv case); `make regen-golden`.

3. **[Med · Maint]** Decompose `cross_checks()` body into named sub-functions (structural, not file-split).

4. **[Med · Maint]** Extract `quest_steward_pick` to `tools/bot/quest_runtime.py` + `test_quest_runtime.py`.

5. **[Low · SDD]** Finalize v458 as-built/lifecycle row; update scenario catalog for v449–v460 entries.

*Evidence: subagent extras audit + `grep status_effects validate_shared.py` (zero hits); `shops.v0.json:7–9`; `tools/bot/ci_pack.json`.*
