# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v223**

**Date:** 2026-06-16
**Scope:** Shared protocol/rules/goldens, Python validators and bot tooling, CI orchestration, progress docs, SDD cadence, and review/refactor handoff after the v219-v223 autoloop queue.
**Baseline:** `main` at `41d712e3` (`docs: record v219-v223 ci proof`). Worktree was clean before this review started.
**Stats:** 73 protocol JSON files, 38 shared rule JSON files, 70 golden JSON files, 93 protocol bot scenarios, 47 client bot scenarios, 55 Python files. `PROGRESS.md` is 184 lines.
**Overview:** [`../20260616_v223-overview.md`](../20260616_v223-overview.md)

---

## Summary

Shared/tooling/process is green at v223. `make ci` passed in 8m25s after the selected `$autoloop` queue, and `make maintainability` reports the progress dashboard at 184/250 lines. The batch also shows the intended SDD loop working: five player-visible slices landed with specs, plans, as-built notes, bot proof, focused verification, commits, and a final batch CI proof.

The strongest process risk remains the same as v200: moved progress archives still contain file-relative links that point at `docs/...` from inside `docs/progress/`. Tooling also still carries 37 legacy `helpers=globals()` bridge calls in `tools/bot/run.py`, and `tools/validate_shared.py` remains a broad validator.

## 1. Architecture

- **[Strength] Shared rules continue to own build/item behavior.** The new Bloodbound Sigil unique is declared in shared data and proved through server and bot paths, advancing ADR-0014 D5 without protocol churn.
- **[Strength] Protocol bot coverage expanded with feature slices.** Protocol scenario count is now 93, including `unique_non_damage_skill_modifier`; client scenario count is 47, including the expanded unique chest proof.
- **[Med] Companion tuning remains outside shared data.** Companion/elite-minion movement constants are still Go-owned, so balance changes bypass shared schemas and cross-language validation.

## 2. Technical

- **[Strength] CI orchestration is high signal.** The final `make ci` run passed ratchets, shared validation, asset validation, determinism lint, Go tests, Python tests, server boot, protocol bot + replay, client bots, and Godot smoke.
- **[Med] Progress link validation is still narrow.** `make/ci.mk:3` wires `check-progress-dashboard.sh` through `make maintainability`, and the script enforces dashboard size plus required archives, but it does not validate Markdown links (`scripts/check-progress-dashboard.sh:10`, `scripts/check-progress-dashboard.sh:34`).
- **[Med] Moved archive links still appear broken.** `rg '\\]\\(docs/' docs/progress/*.md` finds 224 root-style links from files under `docs/progress/`; examples include `docs/progress/slice-lifecycle.md:20` and `docs/progress/shipped-changelog-archive.md:149`.

## 3. Maintainability

- **[Strength] Ratchets are still doing useful work.** `make maintainability` passed with 34 grandfathered files and 37 legacy helper-global injections.
- **[Med] `tools/bot/run.py` still launders helper namespaces.** The extraction-coupling baseline remains at 37 (`.maintainability/extraction-coupling-baseline.tsv:1`), with wrappers passing `helpers=globals()` into wait/state/coop helpers (`tools/bot/run.py:1981`, `tools/bot/run.py:2284`, `tools/bot/run.py:3392`).
- **[Med] `tools/validate_shared.py` is still broad.** At 3,040 lines, it remains a major tooling coordinator even though focused validators exist.

## 4. Documentation

- **[Strength] SDD cadence is strong.** The v219-v223 queue produced specs, plans, as-built notes, lifecycle rows, and named bot scenarios for each selected slice.
- **[Strength] `PROGRESS.md` stayed compact despite five slices.** The dashboard is 184 lines against the 250-line gate.
- **[Med] Review cadence slipped past the nominal v210 marker.** The review is now being written at v223 because the feature batch continued to completion before handoff. That is acceptable for autoloop throughput, but `PROGRESS.md` should point the next review to v233 after this review lands.

## Top 5 shared/tooling/process refactors

1. Repair `docs/progress/*` links or add link validation to `scripts/check-progress-dashboard.sh`.
2. Replace one legacy `helpers=globals()` cluster in `tools/bot/run.py` with typed/narrow context and lower the extraction-coupling baseline.
3. Move companion/elite-minion tuning into shared rules and validation.
4. Continue splitting `tools/validate_shared.py` by extracting the next touched rarity/item/unique validator.
5. Add a direct `make check-progress-dashboard` target if progress-link validation becomes an agent-facing standalone check.

*Evidence: final `make ci` passed on 2026-06-16 in 8m25s; current counts from `find`, `wc -l`, and `rg`; code references from `nl -ba`.*
