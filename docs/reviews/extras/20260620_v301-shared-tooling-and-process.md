# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v301**

**Date:** 2026-06-20
**Scope:** Shared contracts/rules/assets, Python tooling, CI orchestration, model preview catalog, progress docs, SDD cadence, and review/refactor handoff after v295-v301.
**Baseline:** `codex/autoloop-movement-fluidity` at `8fdbd38c` (`fix: stabilize movement fluidity batch gates`). Worktree was clean before this review started.
**Stats:** 73 shared protocol files, 40 shared rules files, 70 golden files, 101 protocol bot scenario JSON files, 85 client bot scenario JSON files, `tools` 16,539 Python lines, `docs/specs` 32,662 lines, `docs/plans` 39,137 lines, `docs/as-built` 9,741 lines.
**Overview:** [`../20260620_v301-overview.md`](../20260620_v301-overview.md)

---

## Summary

The process is healthy: v295-v301 each have specs, plans, as-built notes, focused bot/unit proof,
manual visual commands, lifecycle rows, and a final batch proof. The prior v284 model-catalog drift
risk is largely resolved by a generated shared model preview catalog and pytest parity check.

The process risk is CI orchestration, not gameplay correctness. Raw `make ci` is not friendly to
isolated worktrees when another checkout already owns `arpg-postgres`; the agent had to continue
the server-backed steps manually against the existing healthy DB and temporary servers.

## 1. Architecture

- **[Strength] Shared contracts remain authoritative for protocol/rules/assets.** `make validate-shared`
  passed 1,395 checks and `make validate-assets` passed 121 checks during this review.
- **[Strength] Model preview discovery now has a shared generated artifact.** Python owns catalog
  generation to `shared/assets/model_preview_catalog.v0.json` (`tools/assets/model_catalog.py:12`,
  `tools/assets/model_catalog.py:123`, `tools/assets/model_catalog.py:131`), and the Godot model
  viewer reads that generated catalog instead of rebuilding discovery logic
  (`client/scripts/model_viewer.gd:77`, `client/scripts/model_viewer.gd:79`).
- **[Strength] The generated model catalog has drift coverage.** The pytest suite checks that the
  repository catalog matches source data (`tools/assets/test_model_catalog.py:118`,
  `tools/assets/model_catalog.py:138`, `tools/assets/model_catalog.py:142`).

## 2. Technical

- **[Strength] Component gates are green on the current commit.** `make validate-shared`,
  `make validate-assets`, `make lint-determinism`, `cd server && go test ./...`, `cd server && go vet ./...`,
  `.venv/bin/pytest tools`, `make client-unit`, `make maintainability`, post-loop protocol bot +
  replay, post-loop `SCENARIO=all`, and post-loop headless `client_smoke` all passed.
- **[Med] Raw `make ci` has a local infrastructure false-red path.** Step 8 always invokes
  `make db-up` before starting the test server (`scripts/ci.sh:274`, `scripts/ci.sh:278`). That is
  fine in a single checkout, but it failed here because another worktree already had the fixed-name
  Postgres container running. This should be hardened before the next cadence gate.
- **[Low] Presentation-feel tuning lacks a shared home.** The SDD specs documented code ownership
  for several client-only microconstants, but the repository policy still prefers configurable
  presentation tuning (`AGENTS.md:80`, `AGENTS.md:82`). A small catalog or central owner would make
  future tuning less scattered.

## 3. Maintainability

- **[Strength] Extraction-coupling debt remains at zero.** The baseline file says there are no
  grandfathered `helpers=globals()` sites (`.maintainability/extraction-coupling-baseline.tsv:1`),
  and `make maintainability` reported 0 occurrences.
- **[Med] `tools/validate_shared.py` remains a broad validator.** It still owns schema validation,
  class-presentation/model cross-checks, item visuals, goldens, i18n, and many unrelated drift
  guards (`tools/validate_shared.py:1`, `tools/validate_shared.py:552`,
  `tools/validate_shared.py:2870`). At 3,038 lines, future touched domains should keep moving out.
- **[Low] GLB provenance is still local-development only for class models.** Several character
  assets remain `user-provided-unverified` in the manifest (`assets/manifests/assets.v0.json:24`,
  `assets/manifests/assets.v0.json:37`, `assets/manifests/assets.v0.json:50`,
  `assets/manifests/assets.v0.json:63`).

## 4. Documentation

- **[Strength] The SDD trail is complete for the movement batch.** v295-v301 lifecycle rows link
  matching specs, plans, and as-built notes (`docs/progress/slice-lifecycle.md:310`,
  `docs/progress/slice-lifecycle.md:316`).
- **[Strength] The proof limitation is documented instead of hidden.** `PROGRESS.md` explains the
  raw `make ci` step 8 Docker conflict and the CI-equivalent continuation (`PROGRESS.md:28`).
- **[Low] The review cadence should now hand off to `$refactor`.** `PROGRESS.md` correctly says the
  handoff is due, and this review updates it to point at the v301 review set.

## Top 5 shared/tooling/process refactors

1. Make the local CI database/server bootstrap safe for isolated worktrees and already-running
   healthy local Postgres containers.
2. Add a shared presentation-feel config or central code-owner contract before more client feel
   constants spread.
3. Split the next touched validation domain out of `tools/validate_shared.py`.
4. Keep generated model catalog parity in pytest and consider adding `model-catalog-generate --check`
   to a named validation target if model preview scope expands.
5. Resolve or replace `user-provided-unverified` class GLB provenance before production distribution.

*Evidence: current component gates, post-loop final batch proof, current counts from `find`/`wc -l`, and file references from `nl -ba`/`rg -n`.*
