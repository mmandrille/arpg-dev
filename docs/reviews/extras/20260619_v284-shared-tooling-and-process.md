# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v284**

**Date:** 2026-06-19
**Scope:** Shared protocol/rules/assets/goldens, Python validators and bot tooling, asset generation/rigging tools, Makefile/CI orchestration, progress docs, SDD cadence, and review/refactor handoff after v252-v284.
**Baseline:** `main` at `7812c09b` (`chore: add Model Viewer Tool`). Worktree was clean before this review started.
**Stats:** 73 shared protocol files, 38 shared rules files, 70 golden files, 96 protocol bot scenario JSON files, 74 client bot scenario JSON files, `tools` 16,399 Python lines, `docs/specs` 31,222 lines, `docs/plans` 36,786 lines, `docs/as-built` 8,772 lines.
**Overview:** [`../20260619_v284-overview.md`](../20260619_v284-overview.md)

---

## Summary

Shared/tooling/process improved since v250. The extraction-coupling ratchet is now at zero `helpers=globals()` sites, the progress-dashboard check stays compact, asset validation covers increasingly real GLB provenance and skin joints, and v284 added a small dynamic model catalog plus Godot viewer. The SDD trail is complete for v252-v284 with specs, plans, as-built notes, lifecycle rows, and focused proofs.

The review-time issue is process-facing: the official cadence `make ci` failed in the full client bot step even though both failed scenarios passed individually afterward. That means the next maintenance work should prioritize suite stability, because the project relies on `make ci` as the review/refactor gate.

## 1. Architecture

- **[Strength] Shared contracts still own cross-language facts.** `make validate-shared` passed inside review-time CI, covering schemas, examples, cross-checks, goldens, assets, i18n, and codemap validation.
- **[Strength] Asset identity remains manifest-driven.** Character, equipment, and monster assets are resolved through `assets/manifests/assets.v0.json`, with runtime paths and required nodes in one place (`assets/manifests/assets.v0.json:3`). Class model references are cross-checked against character assets in `validate_shared.py` (`tools/validate_shared.py:552`).
- **[Strength] CI includes asset and model-viewer gates.** `scripts/ci.sh` runs shared validation, asset validation, determinism lint, Go tests, Python tests, protocol bot/replay, client bot scenarios, and Godot smoke (`scripts/ci.sh:248`, `scripts/ci.sh:257`, `scripts/ci.sh:260`, `scripts/ci.sh:263`, `scripts/ci.sh:266`, `scripts/ci.sh:269`).
- **[Low] Previewable model catalog discovery is duplicated across Python and GDScript.** The Python catalog reads manifest/class/monster JSON (`tools/assets/model_catalog.py:33`) and the Godot viewer repeats equivalent discovery (`client/scripts/model_viewer.gd:77`). This is manageable now, but should get a parity check before item/equipment previews expand the surface.

## 2. Technical

- **[Strength] Maintainability ratchets are stronger than v250.** Review-time checks passed with 34 grandfathered files / 64,006 lines, 0 helper-global coupling occurrences, and `PROGRESS.md` at 187/250 lines. `.maintainability/extraction-coupling-baseline.tsv:1` now records no grandfathered helper-global sites.
- **[Strength] Python tool tests cover the new model catalog.** `tools/assets/test_model_catalog.py` proves character/monster discovery, multi-use grouping, equipment exclusion, and unknown-asset CLI guidance (`tools/assets/test_model_catalog.py:90`, `tools/assets/test_model_catalog.py:105`, `tools/assets/test_model_catalog.py:112`, `tools/assets/test_model_catalog.py:123`).
- **[Med] Full CI did not pass despite focused reruns being green.** The client bot step reported 72 passed / 2 failed, then focused reruns for `49_mercenary_recovery_ui` and `73_door_fog_toggle` both passed. This is a flake/stability problem in the gate itself, not a reason to mark the cadence review green.

## 3. Maintainability

- **[Strength] Prior review follow-ups are materially closed.** Stale v241-v250 CI wording is fixed in lifecycle rows, the client debug-progression unit gap was closed by `client/tests/test_net_client.gd`, and helper-global coupling is gone.
- **[Med] `tools/validate_shared.py` remains broad.** At 3,042 lines, it still owns many unrelated validation domains, including class presentation assets (`tools/validate_shared.py:552`) and item visual/golden checks (`tools/validate_shared.py:2874`). Future schema/rules work should keep extracting focused validators.
- **[Low] Asset provenance is not production-ready for several supplied GLBs.** The manifest records `user-provided-unverified` licenses for multiple class source models (`assets/manifests/assets.v0.json:24`, `assets/manifests/assets.v0.json:37`, `assets/manifests/assets.v0.json:50`, `assets/manifests/assets.v0.json:63`). That is workable for local development but must be resolved before distribution.

## 4. Documentation

- **[Strength] SDD cadence is intact.** The lifecycle rows for v252-v284 point to matching specs, plans, and as-built notes (`docs/progress/slice-lifecycle.md:267`, `docs/progress/slice-lifecycle.md:299`).
- **[Med] `PROGRESS.md` should now steer `$next` toward CI stability.** The current dashboard correctly says the review is due (`PROGRESS.md:31`), but after this review the active follow-up should prioritize full-suite client bot stability before new features.
- **[Low] Deferred backlog language needs a cleanup pass.** `PROGRESS.md:102` and nearby presentation rows still describe first-pass bat dive, quadruped pounce, boss summons, and boss pattern additions as deferred even though v280-v283 shipped those first passes.

## Top 5 shared/tooling/process refactors

1. Stabilize the client bot full-suite flakes and rerun full `make ci` to restore a green cadence baseline.
2. Add model-catalog parity coverage or generate one shared preview catalog consumed by both Python and Godot.
3. Split the next touched validator domain out of `tools/validate_shared.py`.
4. Resolve or replace `user-provided-unverified` class GLB provenance before production distribution.
5. Refresh deferred backlog wording after v280-v283 so future `$next` candidates target remaining work, not already-shipped first passes.

*Evidence: review-time `make ci`; focused client scenario reruns; review-time maintainability commands; current counts from `find`/`wc -l`; code references from `rg -n` and `nl -ba`.*
