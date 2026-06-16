# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v200**

**Date:** 2026-06-15
**Scope:** Shared protocol/rules/goldens, Python validation and bot tooling, CI orchestration, progress docs, and SDD/review cadence.
**Baseline:** `main` at `310b67d4` (`docs: add v200 engineering review`). Worktree was clean before the requested deletion/regeneration of the previous v200 review files; during this review run the only pre-existing dirt was those deleted review files.
**Stats:** 72 protocol JSON files, 38 shared rule JSON files, 70 golden JSON files, 88 protocol bot scenarios, 45 client bot scenarios, 53 Python files. `PROGRESS.md` is 179 lines after v200 compaction.
**Overview:** [`../20260615_v200-overview.md`](../20260615_v200-overview.md)

---

## Summary

The shared/tooling/process baseline is green after the v200 repair commits. The previous skill prerequisite drift was fixed in shared rules and guarded in `tools/validate_skills.py`. `make ci` passed in 7m44s, including schema validation, Python tests, protocol bot, replay, client bot, and smoke.

The main remaining process issue is documentation hygiene from the v200 progress compaction: archive files moved under `docs/progress/`, but many links still appear to use repo-root-relative `docs/...` paths. That is not a CI blocker today, but it will confuse agents following links from the moved files.

## 1. Architecture

- **[Strength] Shared skill prerequisites are stable again.** `tools/validate_skills.py` now checks expected class skill prerequisite chains (`tools/validate_skills.py:123`), including Volley requiring Piercing Shot and Split Arrow requiring Volley (`tools/validate_skills.py:130`).
- **[Strength] Shared validation remains central.** `tools/validate_shared.py` imports focused validators such as `validate_skills`, `validate_item_presentations`, `validate_main_config`, and `validate_unique_items` (`tools/validate_shared.py:22`).
- **[Med] Companion AI tuning has not moved into shared data yet.** Go constants still own follow/assist distances, leaving future tuning outside shared schemas.

## 2. Technical

- **[Strength] Official CI is green.** `make ci` passed in 7m44s.
- **[Strength] The repaired scenarios are green.** Protocol bot now passes `rage_and_heal_skills`, `barbarian_class_foundation`, `ranger_volley_and_visual_showcase`, and `mana_regeneration`; client bot now passes `client_skill_points_and_magic_bolt`.
- **[Med] Progress dashboard validation is present but narrow.** `scripts/check-progress-dashboard.sh` ensures `PROGRESS.md` stays under 250 lines and required archives exist (`scripts/check-progress-dashboard.sh:14`, `scripts/check-progress-dashboard.sh:27`), but it does not validate archive links.

## 3. Maintainability

- **[Strength] File-size ratchet is passing.** `.maintainability/file-size-baseline.tsv` now explicitly records the repaired baselines for the previously failing files (`.maintainability/file-size-baseline.tsv:7`, `.maintainability/file-size-baseline.tsv:13`, `.maintainability/file-size-baseline.tsv:18`, `.maintainability/file-size-baseline.tsv:31`).
- **[Med] Baseline refresh solved the gate, not the underlying size.** `client/scripts/main.gd`, `server/internal/game/rules.go`, `client/tests/test_coop_client.gd`, and `skills/showme/scripts/visual_capture.gd` still need future shrink work.
- **[Med] `tools/validate_shared.py` remains broad.** It is still 3,040 lines and owns item rarity, dungeon rarity, and unique-effect validation sections (`tools/validate_shared.py:904`, `tools/validate_shared.py:1356`, `tools/validate_shared.py:2981`).

## 4. Documentation

- **[Strength] `PROGRESS.md` is now a dashboard.** The v200 compaction split lifecycle history, scenario catalog, and shipped changelog into `docs/progress/`.
- **[Med] Moved archive links likely resolve incorrectly.** `rg '\\]\\(docs/' docs/progress/*.md` finds 202 root-style links from inside `docs/progress/`; for example `docs/progress/slice-lifecycle.md:20` links to `docs/specs/...`.
- **[Low] There is no direct `make check-progress-dashboard` target.** The script is run through `make maintainability`, but an explicit target would make process checks easier to invoke by name.

## Top 5 shared/tooling/process refactors

1. Repair or validate relative links in `docs/progress/*` after the progress compaction.
2. Add archive link validation to the progress dashboard/process checks.
3. Move companion AI follow/assist tuning into shared rules and schema validation.
4. Continue extracting validators out of `tools/validate_shared.py` by domain.
5. Add a direct make target for `scripts/check-progress-dashboard.sh` if agents are expected to run it independently.

*Evidence: `make ci` passed in 7m44s; link count from `rg '\\]\\(docs/' docs/progress/*.md`; current stats from `find` and `wc -l`.*
