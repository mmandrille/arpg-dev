# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v537**

**Date:** 2026-10-03
**Scope:** `shared/rules`, `shared/protocol`, `shared/golden`, `shared/assets`, `tools/`, `make/`, `scripts/`, SDD docs for v532–v537.
**Baseline:** `main` @ `fca53fcf`, clean. `make validate-shared` 2,273 checks pass; `make validate-assets` 464 pass; `pytest tools` 324 pass; `make maintainability` pass; `make ci` PASS 20m36s; `make ci-full` not run.
**Overview:** [`../20261003_v537-overview.md`](../20261003_v537-overview.md)

---

## Summary

Contract changes were additive and in place, per project policy, and the new cross-stack invariant (server prop catalog ↔ client presentation) got a validator in a new small module rather than growing `validate_shared.py`. Process was sound but ran standalone: the host could not create the separate sessions the batch workflow requires, so the six accepted slices were executed one at a time in the main chat with the same spec → plan → execute artifacts.

## 1. Contracts and rules-as-data

- **[Strength] Data-owned floor for openings.** `dungeon_generation.passage_clearance.min_player_diameters` (schema-required) rejects narrow corridors, legacy widths and door gaps at load; tests derive from loaded rules and cover each rejected case.
- **[Strength] Props are fully data-driven** (`obstacle_generation.props`: bands, clearances, catalog with footprints); presentation (`dungeon_kit_presentation.v0.json`) is keyed by `prop_id`.
- **[Med] Wire parity is still manual.** `kind: prop` and `prop_id` were added by hand to `state_delta.v8`, `session_snapshot.v8` and the golden schema; the Go-struct-tag ↔ schema check carried from v486/v531 remains open, so a missed schema would be caught only by the live payload gate on a scenario that exercises it (`126_dungeon_props` now does).
- **[Low] Golden schema enum edits** were needed twice (shape families, wall kind); props are excluded from `shape_families`/`solid_kinds` because they are not scatter groups.

## 2. Tooling and CI

- **[Med] Go test runtime hit the default timeout.** `go test ./...` timed out once at 10m during the batch and the package measures ~596 s isolated; CI now uses `-timeout 20m` (`scripts/ci.sh:326`, `make/server.mk:10`). The `ci.sh` step label still reads `go test ./...` (cosmetic).
- **[Med] Seed-pinned scenarios are brittle.** `124_room_doors` passed on only 2 of 6 seeds on baseline `main`; it was re-pinned to `_03`. Prefer choosing a seed by search at scenario build time or asserting properties on any door-bearing seed.
- **[Low] `dungeon_frame_pacing_probe` is stale** (times out at `wait_wall_layout` on baseline); its pinned fixture predates v522–v531 generation changes. Perf claims for v533 therefore rest on one live snapshot per side.
- **[Strength] New tooling is small and tested:** `tools/validate_dungeon_props.py` (2 checks), `item-tooltip` showme suite with catalog tests, two extended client scenarios registered in the movement audit.
- **[Low] `PROGRESS.md` is at 248/250 lines**, the dashboard guard's ceiling; the next batch must trim before adding.

## 3. SDD process

- Specs, plans and as-built exist for v532, v533, v535, v536, v537 with explicit non-goals, recorded decisions (signature kept for the tooltip; widths chosen empirically) and limits. Plan checkboxes are ticked; statuses say "Complete" only after the combined gate.
- **[Med] Spec numbers were renumbered mid-batch** (v535 pace cancelled, later slices shifted down one) before any session existed, which is fine, but the offer text and the files must stay in sync; the lifecycle table is the authority.
- **[Med] Evidence discipline:** the as-built notes separate runtime proof, static reading and unproven claims (e.g. "props not visible in real-floor frames", "not a controlled A/B").

## 4. Documentation

Review set, as-built notes and the retained frames are coherent; CODEMAP lists every new file and `make validate-shared` verifies it.

## Top 5 shared/tooling refactors

1. **[future-plan · Med]** Automate Go struct-tag ↔ v8 schema parity with a negative fixture (open since v486).
2. **[minor-commit · Med]** Replace seed-pinned door scenarios with a deterministic seed finder or property assertion.
3. **[future-plan · Med]** Rebuild `dungeon_frame_pacing_probe` on a fixture the current generator guarantees.
4. **[minor-commit · Low]** Update the CI step label and add a runtime budget note for the game package.
5. **[minor-commit · Low]** Trim `PROGRESS.md` (move shipped prose to lifecycle) to restore dashboard headroom.

*Evidence: `docs/reviews` prior set; `docs/progress/*`, `tools/validate_dungeon_props.py`, `shared/rules/dungeon_generation.v0.json`; validators and CI results above.*
