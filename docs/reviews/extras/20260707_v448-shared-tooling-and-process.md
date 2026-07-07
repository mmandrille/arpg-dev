# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v448**

**Date:** 2026-07-07
**Scope:** `shared/`, `tools/`, Makefile/CI, SDD cadence — v421–v448 (28 slices) vs v420 baseline.
**Baseline:** `main` @ `ebcbdc9a`. Worktree: **clean**.
**Stats:** **137** protocol bot scenarios; **111** client bot scenarios; CI fast pack: **35** scenarios; `skills.v0.json` **~5,193** lines; `validate_shared.py` **3,156** lines; `run.py` **4,438** lines.
**Overview:** [`../20260707_v448-overview.md`](../20260707_v448-overview.md)

---

## Summary

The v421–v448 run shipped 28 slices dominated by asset pipeline work before returning to gameplay with quest steward hunts (v448). On the shared-contracts side, protocol stayed at v8 (hunt extended in place, no bump needed), three new rule files landed (`quest_steward.v0.json`, `unique_effects.v0.json`, `unique_items.v0.json`), and the rules-as-data boundary held cleanly. The SDD trail is complete for every slice, using the spec-gate exemption correctly for the four client-only art slices (v423, v424, v425, v428). The two structural debts from v420 — `validate_shared.cross_checks()` as a ~2,941-line monolith and `run.py` at the ratchet ceiling — are both unresolved: `cross_checks()` absorbed ~215 new lines, and `run.py` grew 45 lines (4,393 → 4,438). The extraction coupling debt is fully cleared (zero `helpers=globals()` sites). Two items from the v420 top-5 were closed today: v420–v424 spec reconciled, and a `timeout_s` unit test added to `test_movement_runtime.py`. The engineering review cadence is 28 slices overdue; this report closes that gap.

---

## 1. Architecture

### Rules-as-data

**[Strength]** Rules-as-data boundary is healthy. All three new rule files follow the established pattern: compact JSON data with a paired schema, loaded by both Go and client, with tuning values in JSON only. `quest_steward.v0.json` (27 lines) is notably minimal — a positive signal that the hunt's configuration surface is tight.

**[Strength]** `unique_effects.v0.json` (247 lines) and `unique_items.v0.json` (64 lines) both have cross-check coverage in `validate_shared.py` (treasure class refs validated against unique items catalog).

**[Med]** `quest_steward.v0.json` has no cross-check coverage. `trophies[].monster_def_id` is not validated against `monsters.v0.json`, `trophies[].item_def_id` is not validated against `item_templates`, and `reward_families[].template_ids` are not validated against `item_templates`. Broken references will fail silently at runtime rather than at `make validate-shared`.

### Protocol versioning

**[Strength]** Protocol is at v8 across all four schema types (envelope, messages, session_snapshot, state_delta). The v448 as-built explicitly records "Protocol extended in place on v8 (no v9 bump)" — the hunt additions (`steward_hunt_target` flag, trophy metadata, `quest_steward_offers_opened`, `quest_steward_pick_intent`) were backward-compatible additions. Immutable-file-per-version discipline is intact.

### CI tier model

**[Strength]** Three-tier model functions correctly. Fast CI pack: 35 scenarios (22 protocol + 13 client). Extended: `120_steward_hunt_quest` (protocol) and `98_steward_hunt_quest` (client) registered as extended. The fast pack does not include the steward hunt scenario — a valid tradeoff documented in the v448 as-built (seed-pinned world with 10% hunt floor is unsuitable for the fast pack).

**[Strength]** Movement audit: 249 entries, current as of v448. `98_steward_hunt_quest` correctly classified as `extended / contract / 0 movement steps`. No incidental navigation violations.

---

## 2. Technical

### `validate_shared.py`

**[High]** `cross_checks()` spans lines 201–3141 — a **2,941-line single function**. This is unchanged in structure from v420. ~215 new lines of cross-checks were added for unique items and unique effects. The function is functionally correct but is now a ~3,000-line read bottleneck. Adding a new cross-check requires navigating to the correct region, and the lack of named sub-sections makes coverage hard to audit systematically.

### `run.py`

**[High]** `run.py` grew from 4,393 to 4,438 lines (+45). The `quest_steward_pick` action and `steward_hunt_target` selector were added directly to the orchestrator rather than extracted. The ratchet baseline was bumped to 4,438. This continues the pattern of growing the orchestrator — despite 38 Python modules existing alongside it. The extraction coupling debt is cleared (0 `helpers=globals()` sites), so the path to a `quest_runtime.py` extraction is technically clean.

### Movement runtime test

**[Strength]** New `test_movement_runtime.py` (129 lines, added today) enforces extraction independence: verifies `movement_runtime` imports without `tools.bot.run`, tests range candidate ordering, and covers the `timeout_s` early-exit path with a fake loop clock. This is the correct pattern for extracted module proofs and should serve as the template for future extractions.

### SDD artifacts

**[Strength]** v421–v448 as-built files: 100% coverage (28/28). Spec and plan coverage: 24/28 (the four spec-exempt art slices correctly omitted). As-built quality is consistently structured: Shipped / Key decisions / Deferred / Verification sections.

**[Low]** The four spec-exempt slices (v423, v424, v425, v428) have correct as-built files but none explicitly state "spec-gate exemption applied: client-only presentation." A single sentence in each as-built header would make the audit trail self-explanatory without requiring cross-reference to CLAUDE.md.

---

## 3. Maintainability

| File | v420 lines | v448 lines | Delta | Status |
|---|---|---|---|---|
| `tools/bot/run.py` | 4,393 | 4,438 | +45 | At ceiling — baseline bumped |
| `tools/validate_shared.py` | 3,156 | 3,156 | 0 | Unchanged; `cross_checks()` +215 internally |
| `tools/bot/test_protocol.py` | ~1,434 | 1,455 | +21 | Within ratchet allowance |
| `tools/bot/movement_runtime.py` | ~317 | 317 | 0 | Clean extraction |
| `tools/bot/runtime_assertions.py` | ~504 | 504 | 0 | Clean extraction |
| `tools/bot/test_movement_runtime.py` | 0 | 129 | +129 | New today; under 600 |
| `tools/bot/bot_types.py` | ~101 | 101 | 0 | Under limit |
| `shared/rules/*.json` (total) | ~17,500 est. | ~17,994 | +~500 | 3 new rule files; healthy growth |
| `docs/progress/scenario-movement-audit.tsv` | ~240 est. | 249 | +9 | Current |

---

## 4. Documentation

**[Strength]** Uninterrupted spec/plan/as-built trail for v421–v448. Spec-gate exemption used correctly and consistently for art slices.

**[Strength]** `docs/progress/scenario-catalog.md` documents CI tiers, 10s budget policy, and movement allowlist. v448 steward hunt scenarios appear at the correct tier.

**[Med]** `docs/specs/v420_spec-class-skill-build-branches.md` was reconciled today: batch plan field updated to reflect all 30 skills landing in v420 (v421–v424 were reassigned to asset slices), and acceptance criteria checked off. This closes the v420 review spec/lifecycle drift finding.

**[Low]** The four spec-exempt as-built files do not state the exemption explicitly. Low impact since CLAUDE.md is authoritative, but self-explanatory as-builts reduce review burden.

---

## Top 5 shared/tooling/process refactors

1. **[High · Maint]** Break `validate_shared.cross_checks()` into named validator functions. The 2,941-line single function should become ~10 named sub-functions (`_check_monster_refs`, `_check_skill_stat_cross`, `_check_unique_item_refs`, etc.) called sequentially from `cross_checks()`. Each sub-function takes `(report: Report, rules: dict)` and is independently testable. File total stays at ~3,156; internal structure becomes navigable.

2. **[High · Correctness]** Add `quest_steward.v0.json` cross-checks to `validate_shared`. Three assertions: (a) `trophies[].monster_def_id` exists in `monsters.v0.json`, (b) `trophies[].item_def_id` exists in `item_templates`, (c) `reward_families[].template_ids` entries exist in `item_templates`. Closes a silent-failure risk in the hunt loop.

3. **[Med · Maint]** Extract `quest_steward_pick` action and `steward_hunt_target` selector from `run.py` to `tools/bot/quest_runtime.py`. Follow the `movement_runtime.py` pattern: focused module, `BotContext` interface, paired `test_quest_runtime.py` proving extraction independence. Recovers ~40–60 lines from `run.py` and establishes the quest domain as a first-class tested extraction.

4. **[Med · CI]** Evaluate promoting a trimmed steward hunt proof to the fast CI pack. v448 as-built documents a pinned seed (`v448_steward_probe_17`) that reliably produces a bat-wing trophy on depth 1. If the trimmed happy path (enter floor → kill target → pickup trophy → turn in) runs within the 10-second budget, a pack scenario closes the fast-CI coverage gap for the newest first-class feature.

5. **[Low · Docs]** Add explicit spec-gate exemption notes to v423, v424, v425, v428 as-built files. A single sentence — "Spec-gate exempt: client-only GLB/animation pipeline; no protocol/server/rules/golden changes" — makes the audit trail self-explanatory without requiring cross-reference to CLAUDE.md.

*Evidence: file reads at `ebcbdc9a`; `validate_shared.py` (structure, `cross_checks()` span), `run.py` (line count, quest steward additions), `tools/bot/test_movement_runtime.py` (new today), `quest_steward.v0.json`, `unique_effects.v0.json`, `unique_items.v0.json`, `shared/protocol/` (schema files), `tools/bot/ci_pack.json`, `docs/progress/scenario-catalog.md`, `docs/progress/scenario-movement-audit.tsv`, `docs/as-built/v421–v448`, `docs/reviews/extras/20260703_v420-shared-tooling-and-process.md`.*
