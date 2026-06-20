# arpg-dev - Shared contracts, Python tooling & SDD process review at slice **v301**

**Date:** 2026-06-20
**Scope:** Shared protocol/rules/assets/goldens, Python validators and bot tooling, Make/CI
orchestration, review cadence, progress docs, and SDD process after the World Detail/Navigation
queue.
**Baseline:** `codex/world-detail-navigation` at `468c055d` (`feat: v301: wall/floor shader polish`). Worktree was clean before this review started.
**Stats:** 73 shared protocol files, 40 shared rules files, 70 golden files, 105 protocol bot
scenario JSON files, 81 client bot scenario JSON files, `tools` 16,545 Python lines, `docs/specs`
32,762 lines, `docs/plans` 39,520 lines, `docs/as-built` 9,631 lines.
**Overview:** [`../20260620_v301-overview.md`](../20260620_v301-overview.md)

---

## Summary

The shared/tooling/process baseline improved. The v284 review's top issue, full-suite client-bot
instability, is resolved at the v301 baseline: full `make ci` is green. The generated model preview
catalog is now a real shared artifact consumed by Python and Godot, closing another v284 follow-up.
The SDD trail is complete for v295-v301, and the lifecycle rows now mark the batch green.

The standout process issue is worktree-safe database orchestration. The repo's Docker Compose file
sets a fixed `container_name: arpg-postgres` (`docker-compose.yml:7`), and `make db-up` probes that
fixed name (`make/db.mk:7`). In the isolated worktree, final CI had to run with
`COMPOSE_PROJECT_NAME=arpg-dev` to reuse the main worktree's healthy container. That is workable but
not self-documenting.

## 1. Architecture

- **[Strength] Shared model preview catalog is now the cross-language source.**
  `shared/assets/model_preview_catalog.v0.json` lists previewable character and monster rows
  (`shared/assets/model_preview_catalog.v0.json:1`). Python loads it when present
  (`tools/assets/model_catalog.py:35`), and Godot reads the same file (`client/scripts/model_viewer.gd:77`).
- **[Strength] The World Detail/Navigation queue kept contracts coordinated.** Water, holes,
  obstacle kinds, flying traversal, Leap crossing, LOS blockers, and wall/floor material proof all
  have specs/plans/as-built notes and bot scenarios.
- **[Med] Local CI orchestration assumes one repo checkout.** Fixed Docker container naming in
  `docker-compose.yml` plus hard-coded `docker exec arpg-postgres` in `make/db.mk` makes parallel
  worktrees collide unless the agent knows to share `COMPOSE_PROJECT_NAME`.

## 2. Technical

- **[Strength] Current quality gates are green.** `COMPOSE_PROJECT_NAME=arpg-dev make ci` passed in
  11m50s. `make maintainability` passed after review docs were updated, and `go vet ./...` passed.
- **[Strength] Bot catalog growth is still covered by broad and focused proof.** The repo now has
  105 protocol scenarios and 81 client scenarios; v301 adds `wall_floor_shader_polish`, and final
  CI ran the entire catalog.
- **[Low] Timing budgets are still hand-authored.** The elite-objective protocol scenario uses
  explicit `max_elapsed_s` and `max_ticks` after generated obstacle growth made the path longer
  (`tools/bot/scenarios/68_dungeon_elite_side_objective.json:5`). This is acceptable, but a small
  guideline or validator warning could prevent future too-tight budgets.

## 3. Maintainability

- **[Strength] The ratchets continue to work.** File-size ratchet, extraction-coupling ratchet, and
  progress-dashboard checks all pass; `PROGRESS.md` is still 188/250 lines.
- **[Med] `tools/validate_shared.py` is still a broad validator.** It is 3,038 lines and continues
  to own many unrelated domains. Future schema/rules changes should prefer focused validators like
  the existing skill and unique-item splits.
- **[Low] Asset provenance remains a production blocker.** Several class GLBs still record
  `user-provided-unverified` licenses (`assets/manifests/assets.v0.json:24`,
  `assets/manifests/assets.v0.json:37`, `assets/manifests/assets.v0.json:50`,
  `assets/manifests/assets.v0.json:63`).

## 4. Documentation

- **[Strength] Review cadence is correct.** `PROGRESS.md` now says the v301 review/refactor handoff
  is due after a green batch CI (`PROGRESS.md:28`, `PROGRESS.md:31`).
- **[Strength] Lifecycle rows are current for v295-v301.** The selected queue rows all show focused
  checks plus batch-CI green status (`docs/progress/slice-lifecycle.md:310`).
- **[Low] Worktree CI workaround is not discoverable enough.** The final full CI proof required
  `COMPOSE_PROJECT_NAME=arpg-dev`; that should either be unnecessary or documented in the local
  workflow instructions.

## Top 5 shared/tooling/process refactors

1. Make `make db-up` / `make ci` worktree-safe without requiring hidden Compose-project knowledge.
2. Add `go vet ./...` to CI now that it is green.
3. Add a short bot-scenario authoring note or validation hint for generated-route timing budgets.
4. Split the next touched validation domain out of `tools/validate_shared.py`.
5. Resolve or replace `user-provided-unverified` GLB provenance before production distribution.

*Evidence: full `COMPOSE_PROJECT_NAME=arpg-dev make ci`, `make maintainability`, `cd server && go
vet ./...`, current counts from `find`/`wc -l`, and cited files above.*
