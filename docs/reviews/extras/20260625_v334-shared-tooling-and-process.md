# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v334**

**Date:** 2026-06-25
**Scope:** `shared/`, `tools/`, `Makefile`, `scripts/`, SDD docs. Covers v332–v334 and v331 `$refactor` paydown verification.
**Baseline:** `main` at `9d02179b`; uncommitted bot scenario JSON + nav test WIP.
**Stats:** `make validate-shared` 1,555 checks OK; `validate_codemap.py` OK (inverse unlisted-file check active); 193 bot scenarios; ADR-0015 accepted.
**Overview:** [`../20260625_v334-overview.md`](../20260625_v334-overview.md)

---

## Summary

Most v331 extras recommendations **landed** between reviews: `validate_codemap.py` now fails on unlisted `server/internal/game/*.go` and `client/scripts/*.gd` files; `fog_presentation.v0.schema.json` has `required` on `point_light`; `_validate_fog_presentation_ranges()` guards tuning in `validate_shared.py`; ADR-0015 records movement-speed boundary; CLAUDE.md documents spec-gate exemption for presentation-only slices. CODEMAP index includes v309–v331 domains (hero light, fog, minimap, room layout).

Remaining gaps: `rules.go` type-cluster extraction (backend-owned but affects shared-rules loader shape), `main.gd` maintainability ratchet failure (blocks clean CI), and optional deeper cross-checks for `camera_presentations.v0.json` mode naming vs fog compositor keys.

---

## 1. Shared contracts

**[Strength]** `make validate-shared` passes 1,555 checks including fog semantic range guards (`falloff_power`, `point_light.energy`, `range_multiplier`).

**[Strength]** `fog_presentation.v0.schema.json` `point_light` block now declares `required: ["energy", "range_multiplier", "shadow_enabled", "color"]` — closes v331 silent-omit risk.

**[Strength]** Protocol remains v8; v332–v334 changes are rules tuning and server sim only where applicable — no unnecessary schema bump.

**[Low]** No golden pins fog presentation defaults — acceptable for client-only tuning; schema + semantic guards sufficient for now.

**[Med]** `camera_presentations.v0.json` mode names vs `fog_presentation` `organic_edge` keys still lack a `cross_checks()` link (v331 carry-over, low risk).

---

## 2. Python tooling

**[Strength]** `validate_codemap.py:43-59` implements inverse unlisted-file detection — resolves the third consecutive v331 CODEMAP rot flag.

**[Strength]** `tools/bot/test_protocol.py` — 64 tests pass (2026-06-25 spot check).

**[Med]** `validate_shared.py` at ~3,000+ lines; sibling modules (`validate_dungeon_goldens.py`, etc.) keep growth contained but a dedicated validation domain extraction remains future-plan scale.

**[Low]** Some extracted validators (`validate_main_config.py`, `validate_boss_patterns.py`) still lack matching `test_*.py` — inconsistent pattern from v331.

---

## 3. Process / SDD

**[Strength]** v332–v334 feature slices have as-built docs (`v332_pathfinding-cell-accuracy`, `v333_room-spawn-awareness`, `v334_ranged-retreat-engagement`).

**[Strength]** Spec-gate exemption documented in CLAUDE.md:198–201 for presentation-only work without protocol/rules changes.

**[Med]** This review is **ad hoc** at v334; PROGRESS.md cadence pointed to v340. User-requested early pass captures post-v331 stabilization; cadence pointer updates to v340 on landing.

**[Low]** PROGRESS.md still notes "batch `make ci` pending" — spot checks green (`go test`, `validate-shared`, `validate_codemap`); full `make ci` not re-run in this review.

---

## 4. Documentation

**[Strength]** CODEMAP current — `validate_codemap.py` passes including inverse check.

**[Strength]** ADR-0015 movement-speed formula accepted.

**[Low]** `docs/progress/scenario-catalog.md` may need refresh when wall-floor scenario 78 lands.

---

## Top 5 extras refactors

1. **[minor · Test]** Add `test_validate_main_config.py` mirroring `test_validate_dungeon_goldens.py` pattern for consistency.

2. **[minor · Schema]** Add `cross_checks()` entry linking `camera_presentations` mode IDs to fog `organic_edge` enabled flags.

3. **[future]** Extract validation domain from `validate_shared.py` coordinator.

4. **[reject]** CODEMAP inverse check — landed post-v331.

5. **[reject]** Fog `point_light` required fields — landed post-v331.

*Evidence: `tools/validate_codemap.py`; `tools/validate_shared.py:3019-3059`; `shared/assets/fog_presentation.v0.schema.json:97-99`; `docs/adr/0015-movement-speed-formula.md`; `make validate-shared` output.*
