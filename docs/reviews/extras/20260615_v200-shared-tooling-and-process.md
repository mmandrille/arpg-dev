# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v200**

**Date:** 2026-06-15
**Scope:** Shared protocol/rules/goldens, Python validation and bot tooling, CI orchestration, progress docs, and SDD/review cadence.
**Baseline:** `main` at `9136e410` (`feat: v200: progress doc compaction`); worktree clean before review docs were written. `$refactor` ran first and produced `32c9d347 refactor: derive sanctuary visual radius from rules`.
**Stats:** 72 protocol JSON files, 38 shared rule JSON files, 70 golden JSON files, 88 protocol bot scenarios, 45 client bot scenarios, 53 Python files. Progress docs are now split across `PROGRESS.md` plus `docs/progress/`.
**Overview:** [`../20260615_v200-overview.md`](../20260615_v200-overview.md)

---

## Summary

The shared/tooling story has two sharply different halves. The v200 progress dashboard compaction is a good process direction: it keeps `PROGRESS.md` small and creates focused archives. But the current shared skill catalog is inconsistent with established class progression, and CI is red across Go, protocol bot, and client bot layers. This review should block further feature slices until the shared rule rows and ratchet drift are repaired.

## 1. Architecture

- **[High] Shared skill rules now encode cross-class prerequisite drift.** Barbarian Rage requires Ranger `piercing_shot` (`shared/rules/skills.v0.json:362`), while Ranger `volley` and `split_arrow` have a relationship that conflicts with established tests (`shared/rules/skills.v0.json:848`, `shared/rules/skills.v0.json:951`). This is a shared-contract bug because Go, protocol bot, and Godot all consume the same data.
- **[Strength] Shared validation itself still passes.** `make ci` step 3 passed `make validate-shared`, so schemas catch shape validity; the missing check is semantic class/prerequisite coherence.
- **[Med] Companion AI tuning is still not data-owned.** Companion follow/assist values remain Go constants rather than shared rules (`server/internal/game/companion_ai.go:5`).

## 2. Technical

- **[High] Official CI is red.** `make ci` failed in 8m25s: file-size ratchet, Go tests, protocol bot + replay, and client bot scenarios failed.
- **[Strength] Several independent gates still provide useful evidence.** Extraction-coupling ratchet, shared validation, asset validation, determinism lint, Python tests, most protocol bot scenarios, 44/45 client bot scenarios, and Godot headless smoke passed.
- **[Med] The progress dashboard checker is not independently exposed as a Make target.** `make maintainability` runs `./scripts/check-progress-dashboard.sh`, but `make check-progress-dashboard` has no rule even though the script exists and is useful for v200 process work.

## 3. Maintainability

- **[High] File-size ratchet is currently failing.** The failed files are `client/scripts/main.gd`, `client/tests/test_coop_client.gd`, `server/internal/game/rules.go`, and `skills/showme/scripts/visual_capture.gd`. The ratchet script enforces baseline plus 25 lines (`scripts/check-file-size-ratchet.sh:78`).
- **[Med] Progress archive links need validation.** `docs/progress/slice-lifecycle.md` and `docs/progress/shipped-changelog-archive.md` contain many `](docs/...)` links even though they now live under `docs/progress/`; `rg` found 202 such links.
- **[Strength] `PROGRESS.md` is now a dashboard.** `./scripts/check-progress-dashboard.sh` passes with `179/250` lines, and detailed history now lives in focused archive files.

## 4. Documentation

- **[Strength] v200 has an as-built note and dedicated progress archive files.** The split makes task startup cheaper and aligns with the new instruction to read only current status/open gaps/checklist by default.
- **[High] Current progress metadata overstated health before this review.** `PROGRESS.md` still said `make ci` was green while the review baseline fails the official gate. The review update should make the red state explicit.

## Top 5 shared/tooling/process refactors

1. Fix the shared skill prerequisite drift and add semantic validation that class skills only depend on same-class or explicitly allowed shared skills.
2. Restore `make ci` to green before any new feature slice.
3. Bring file-size ratchet back under control by splitting or explicitly documenting maintenance exceptions for the four failing files.
4. Add or document a `make check-progress-dashboard` target, or keep the script as an internal maintainability substep but avoid suggesting the missing target.
5. Fix or validate relative links in `docs/progress/*` after the progress compaction.

*Evidence: `make ci` failed in 8m25s; `./scripts/check-progress-dashboard.sh` passed; `make check-progress-dashboard` failed with no rule; link count from `rg '\\]\\(docs/' docs/progress/*.md`.*
