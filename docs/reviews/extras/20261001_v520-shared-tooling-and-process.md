# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v520**

**Date:** 2026-10-01
**Scope:** Shared protocol/rules/assets/goldens, Python validators and bot tools, build/CI orchestration, specs/plans/as-built cadence, ADRs, and the v518–v520 batch closeout.
**Baseline:** `main` @ `09678fc4` (`docs: close out v518-v520 batch`); worktree clean at review start.
**Stats:** 35 protocol schemas, 44 rule JSON files, 74 goldens, 291 bot scenarios (138 client), 42 top-level Python modules. `tools/validate_shared.py` is 3,179 LOC and `tools/bot/run.py` 4,534 LOC. Batch adds no protocol schemas or goldens.
**Overview:** [`../20261001_v520-overview.md`](../20261001_v520-overview.md)

---

## Summary

The batch follows the documented SDD sequence and keeps new rarity, fog, and theme values in schema-backed catalogs. Specs, plans, as-builts, lifecycle rows, and committed images are present for all accepted slices. The coordinator's combined `make ci` passed 11/11 stages in 7m50s; `make ci-full` was not run. Review-side checks passed: 2,267 shared validations plus CODEMAP, maintainability ratchets, and determinism lint.

## 1. Architecture

- **[Strength] Presentation data stays data-driven.** v518 rarity and item visuals, v519 UI tokens, and v520 fog parameters have schemas and semantic validation; shared rule values remain separate from client-only display assets.
- **[Strength] Batch lifecycle evidence is complete.** Each accepted slice has spec, plan, as-built, lifecycle row, and preserved captures. The v520 as-built distinguishes geometric coverage from visual readability.
- **[Med] Shared validator coordination remains monolithic.** `tools/validate_shared.py` is a manually wired coordinator with imports and dispatch in the large file (`:22`, `:3051`); the v517 extraction recommendation remains open.

## 2. Technical

- **[Strength] Shared validation passes.** `make validate-shared` reports 2,267 checks and CODEMAP passes. `make maintainability` passes size, extraction-coupling, and progress-dashboard ratchets. `make lint-determinism` passes with 78 grandfathered sites.
- **[Med] UI token scanning omits nested scripts.** `tools/validate_ui_theme.py:38-44` uses `scripts_dir.glob("*.gd")`, skipping nested paths; tests in `tools/test_validate_ui_theme.py:91-118` do not demonstrate nested discovery. Ownership prefixes and unused tokens are also not checked. This narrows the guarantee of the new token-reference gate.
- **[Low] CODEMAP entry is incomplete.** The UI-theme row includes `skill_tree_styles.gd` but omits `boss_bar_frame.gd`, which calls `UiTheme.frame(...)` (`docs/CODEMAP.md:33`, `client/scripts/boss_bar_frame.gd:9`).

## 3. Maintainability

- **[Med] The test/validation coordinator and bot runner remain large.** `validate_shared.py` (3,179 LOC) and `tools/bot/run.py` (4,534 LOC) are stable v517 hotspots. Any extraction should keep validator ordering and bot-run determinism easy to audit.
- **[Low] Several old evidence links still depend on ignored local artifacts.** v518–v520 captures are committed, while some earlier as-built notes refer to ignored `.artifacts` paths. Those files may not exist on a fresh clone.

## 4. Documentation

- **[Strength] v518–v520 lifecycle and scope records are accurate.** The accepted three slices are complete; character-sheet redesign remains outside this batch.
- **[Med] `PROGRESS.md` open gaps need reconciliation after this review.** The v517 UI-token/scenario-catalog follow-ups are partly complete; preserve only the nested-scan, CODEMAP, extended-visual, and older evidence-provenance limits that still apply.

## Top 5 shared/tooling/process recommendations

1. **[minor-commit · Low] Complete token-gate and CODEMAP hygiene.** Switch literal-token discovery to recursive traversal, add a nested-script fixture, add `boss_bar_frame.gd` to CODEMAP, and update `PROGRESS.md` to mark completed portions of the v517 follow-up. `tools/validate_ui_theme.py:38-44`, `tools/test_validate_ui_theme.py`, `docs/CODEMAP.md:33`, `PROGRESS.md:89`.
2. **[future-plan · Med] Extract a typed validator registry from `validate_shared.py`.** Preserve existing ordering and error reporting, add registration tests, then lower the file-size baseline. `tools/validate_shared.py:22,3051`.
3. **[future-slice · Med] Close remaining visual proof gaps.** Run `v511_monster_variant_pack` on the integrated tree and capture a readable T-wall face if that acceptance goal remains; report these separately from `make ci-full`. `tools/bot/ci_pack.json:41-44`, `docs/as-built/v520_fog-wall-occlusion.md`.
4. **[future-plan · Low-Med] Make historical evidence clone-reproducible.** Preserve essential old captures in tracked assets or clearly label local-only evidence without presenting ignored paths as portable. Check the v512/v513/v516/v517 as-builts.
5. **[future-plan · Med] Decompose Python tooling at ownership seams.** Start with validator registration and keep focused tests per domain; do not add another generic dispatch abstraction.

*Evidence: `make validate-shared` (2,267 checks and CODEMAP), `make maintainability` (pass, dashboard 244/250), `make lint-determinism` (pass), and direct reads of the validator, tests, CODEMAP, scenario catalog, CI pack, lifecycle records, and ADR-0001/0006/0007/0018. Combined `make ci` PASS 11/11, 7m50s is the batch closeout result; `make ci-full` was not run. This review did not rerun bot scenarios, Python unit tests, screenshot regeneration, or `make ci`.*
