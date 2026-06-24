# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v331**

**Date:** 2026-06-24
**Scope:** `shared/`, `tools/`, `Makefile`, `scripts/`, `docs/specs`, `docs/plans`, ADRs, `PROGRESS.md`. Covers v309–v331.
**Baseline:** main at 7057ecd4, uncommitted shared/assets changes (fog_presentation.v0.json + schema).
**Stats:** 204 shared JSON files total; 70 goldens; 101 protocol/rules/assets schemas; 105 server-bot scenarios, 88 client-bot scenarios; 67 Python tool files; 142 Python test functions across 16 test files; `validate_shared.py` 3,034 lines.
**Overview:** [`../20260624_v331-overview.md`](../20260624_v331-overview.md)

---

## Summary

The shared-contract and tooling layer is in a strong steady state for v309–v331. The three substantial v308 items that were code-addressable are now resolved: `CLAUDE.md` rule #12 now correctly names the 10s default / 15s ceiling, the `dungeon_obstacles.json` Go-only framing is fixed in `CLAUDE.md`, and the `dungeon_obstacles` golden-vs-generation cross-check is fully implemented and unit-tested in `validate_dungeon_goldens.py`. The remaining carry-over — `CODEMAP.md` staleness — is the one open regression: it was not updated for any v309–v331 file. The v331 slice itself (`hero-visibility-lighting`) introduces a new shared presentation contract (`fog_presentation.v0.json`) that is correctly schema-validated by the glob catch-all in `validate_shared.py` but lacks deeper semantic cross-checks linking its tuning ranges to the client compositor. v322–v328 (the presentation polish bundle) ship without individual specs or plans, an acceptable deviation only because those slices were bundled presentation tweaks with no shared-contract surface.

---

## 1. Shared contracts (schemas, rules, goldens)

**[Strength] Schema coverage is comprehensive and auto-expanding.**
`validate_shared.py:123` globs all `shared/assets/*.v0.json` automatically, so every new asset file — including `fog_presentation.v0.json`, `camera_presentations.v0.json`, and `model_preview_catalog.v0.json` — gets schema validation for free on the first commit. No opt-in required. This is the right design.

**[Strength] `fog_presentation.v0.json` schema is well-scoped.**
The `additionalProperties: false` guard, `exclusiveMinimum: 0` on tuning values, and the `"Client-only … No gameplay authority"` description in the schema root (`shared/assets/fog_presentation.v0.schema.json:4`) correctly frame this as presentation data. The uncommitted diff adds `height_fraction`, `min_height`, `shadow_bias`, `shadow_normal_bias` with correct bounds — all schema-correct.

**[Med] No semantic cross-check for `fog_presentation.v0.json` tuning ranges.**
The schema validates type and `minimum`/`maximum`, but nothing checks that `falloff_power` is within a physically plausible range (say 0.5–10) or that `shadow_reach_multiplier` does not wildly exceed `light_radius` scale. For a purely client presentation file this is low-risk, but the pattern established by `validate_dungeon_goldens.py` (goldens cross-linked to generation settings) and `validate_boss_patterns.py` applies here: a future balance edit that sets `falloff_power: 0.01` would pass schema validation and silently produce a black screen. No reference to `fog_presentation` exists in `validate_shared.py:198`'s `cross_checks()`. Worth one cross-check anchoring the key knobs to plausible sentinel bounds.

**[Med] `point_light` section in `fog_presentation.v0.schema.json` lacks `required` entries.**
The schema uses `additionalProperties: false` but the `point_light` object block does not declare `required`. If `shadow_enabled` (a boolean gate for real-time shadows) is absent, GDScript silent-default behavior is undefined. At minimum `shadow_enabled` and `energy` should be in `required`.

**[Low] Golden count steady at 70; no new golden was added for the v331 lighting compositor.**
The compositor's `falloff_power` / `edge_feather_world` / `shadow_reach_multiplier` are pure client presentation, so a server golden is not appropriate. But given the compositor's complexity (OmniLight3D shadow bias, perspective ambient suppression), a dedicated `fog_presentation_defaults.json` golden fixture pinning the committed tuning values would catch accidental numeric drift across balance edits. Cross-language parity is not required here; a single Go or Python fixture read by `validate_shared.py` would suffice.

**[Strength] Protocol schemas remain at v8; no version bump needed for v309–v331.**
All v309–v331 slices are presentation-only on the client side (no new wire fields). Correct discipline.

---

## 2. Python tooling

**[Strength] `validate_dungeon_goldens.py` cross-check is fully implemented and unit-tested.**
`tools/validate_dungeon_goldens.py` implements two validation layers: (1) the v40 wall-layout contract (floor size match, named shape families, positive generated wall count, shape family coverage), and (2) golden-vs-generation achievability (water/hole counts reachable under `obstacle_generation` caps, solid kind weights non-zero). `tools/test_validate_dungeon_goldens.py` has 5 test functions covering it in isolation. This resolves the top v308 code-side carry-over. `validate_shared.py:1450` calls it correctly with already-loaded dicts.

**[Strength] `validate_shared.py` grows cleanly via sibling modules.**
The `validate_boss_patterns.py`, `validate_dungeon_goldens.py`, `validate_item_presentations.py`, `validate_main_config.py`, `validate_skills.py`, `validate_unique_items.py` split pattern keeps the 3,034-line coordinator from growing further. Each module is importable independently and unit-tested without the parent file. This matches the extraction-independence rule.

**[Med] `validate_shared.py` `cross_checks()` does not load or check `fog_presentation.v0.json`, `camera_presentations.v0.json`, `monster_visuals.v0.json`, or `model_preview_catalog.v0.json`.**
These were added between v308 and v331. They pass schema validation via the glob. But `cross_checks()` (line 198) has no reference to any of them. For `camera_presentations`, specifically, there is a cross-link opportunity: the camera modes named in `camera_presentations.v0.json` should match the mode names referenced in `fog_presentation.v0.json`'s `organic_edge` section (`enabled_isometric`, `enabled_perspective` keys do not use the same naming convention). For `monster_visuals`, there is an existing `item_visuals.v0.json` cross-check pattern (line 2869) that could be replicated. Not blocking, but the gap will widen as new asset files accumulate.

**[Low] Python test count (142 functions) has not grown proportionally to new tooling surface.**
New modules added since v308 — `validate_dungeon_goldens.py`, `validate_item_presentations.py`, `validate_skills.py` — each have their own test files. But `validate_main_config.py` and `validate_boss_patterns.py` lack matching `test_*.py` files. The pattern is inconsistent: some extracted modules have tests, others do not.

**[Low] `validate_codemap.py` still only checks listed paths exist; it does not catch unlisted present files.**
`tools/validate_codemap.py` (full content reviewed) iterates backtick-quoted tokens in `CODEMAP.md` and flags missing files. It has no inverse: it cannot detect that `client/scripts/fog_presentation_loader.gd`, `client/scripts/hero_light_source.gd`, `client/scripts/hero_visibility_field.gd`, `client/scripts/discovery_minimap.gd`, or `client/scripts/discovery_minimap_state.gd` exist on disk but appear nowhere in `CODEMAP.md`. This is the same structural gap flagged in v308. The fix (compare CODEMAP-listed client/scripts against `ls client/scripts/*.gd`) would be five lines of Python.

---

## 3. Process / SDD

**[Strength] SDD discipline holds for structurally significant slices (v309–v321, v329–v331).**
Every structurally new slice has a matching spec and plan: v309 (passive-skill-column), v310–v311 (stat breakdown UI), v312–v313 (stability fixes), v314–v321 (VFX/presentation), v329 (camera modes), v330 (dungeon rooms), v331 (hero visibility lighting). As-built docs exist for all of these. Spec-to-plan-to-as-built is intact.

**[Med] v322–v328 (presentation polish bundle) have no specs or plans.**
`docs/specs/` has entries for v320 (critical hit punch) and v321 (skill rank VFX) but nothing for v322–v328, which are listed in as-built files under `v328_camera-impact-feedback.md` etc. The git commit message `feat: v328: look-and-feel presentation polish (v322-v328)` confirms these were bundled into a single commit retroactively. This is acceptable for pure visual tweaks with no shared-contract surface, but the pattern is undocumented — the current SDD rules do not carve out a "presentation bundle" exception. If this pattern recurs with slices that do touch shared data, the missing-spec gap becomes a contract risk.

**[Strength] Scenario catalog is current at 193 total scenarios (105 server + 88 client).**
`docs/progress/scenario-catalog.md` is 122 lines. Every new slice that adds a scenario has it registered. The bot scenario hygiene established in v308 is holding.

**[Low] `PROGRESS.md` says "Next engineering review: After v330 ships and make ci is green (~v340)" but the review is being written now at v331 — 11 slices after v308.**
This is a process artifact: the review cadence says "~10 slices" but the pointer was set to v340. The current review is happening correctly; the pointer in `PROGRESS.md` just needs updating when the review is complete.

---

## 4. Documentation

**[High] `docs/CODEMAP.md` is stale — not updated since v308.**
The last CODEMAP commit is `303aca56` ("docs: index v302-v308 world-detail files"). v309–v331 added the following production files with no CODEMAP entry:
- `client/scripts/fog_presentation_loader.gd`
- `client/scripts/hero_light_source.gd`
- `client/scripts/hero_visibility_field.gd`
- `client/scripts/discovery_minimap.gd`
- `client/scripts/discovery_minimap_state.gd`
- `shared/assets/fog_presentation.v0.json`
- `shared/assets/camera_presentations.v0.json`

The existing "Fog of war / line of sight" row (`docs/CODEMAP.md:16`) lists only `fog_of_war_overlay.gd` and `interactable_rules_loader.gd` on the client side. At minimum five new GDScript files and one new shared asset file belong there. `validate_codemap.py` passes cleanly because it only checks that listed paths exist — it cannot detect unlisted files. This is the same high-priority finding from v308, now 23 slices overdue for remediation.

**[Strength] `CLAUDE.md` is accurate on three previously-wrong items.**
Rule #12 (`CLAUDE.md:291`) now reads "10-second budget (15s hard ceiling)" matching `tools/bot/run.py:54` (`MAX_SCENARIO_ELAPSED_S = 15.0`). The golden fixtures section (`CLAUDE.md:211-212`) now explicitly names `dungeon_obstacles.json` as "a Go-only dungeon-layout determinism contract with no GDScript consumer." Both corrections are precise and match the code. The v308 framing bug is closed.

**[Low] ADR-0008 addendum for LoS-gated fog / camera-mode fog policy still absent.**
The v308 review flagged this as a future-plan item. v329 (camera modes) and v331 (hero-visibility-lighting) both make decisions (perspective fog enabled, ambient suppression per camera mode) that belong in an ADR-0008 addendum or a new ADR-0015. The spec (`docs/specs/v331_spec-hero-visibility-lighting.md`) captures the decisions inline, but there is no durable ADR record for "camera-mode fog compositor behavior." As the client presentation grows more complex this becomes harder to reconstruct from specs.

---

## 5. v308 carry-over resolution

| Item | Status | Evidence |
|------|--------|---------|
| `docs/CODEMAP.md` stale (v302–v308 files unlisted) | **OPEN — worsened** | Last CODEMAP commit `303aca56` (v308). v309–v331 added at least 7 new production files with no CODEMAP entries. `validate_codemap.py` cannot catch unlisted files. |
| CLAUDE.md rule #12 contradicts `run.py` (said "10s" vs actual 15s ceiling) | **RESOLVED** | `CLAUDE.md:291–296` now says "10-second budget (15s hard ceiling)" with explicit `MAX_SCENARIO_ELAPSED_S = 15.0` citation. |
| `dungeon_obstacles.json` mislabeled as "cross-language" | **RESOLVED** | `CLAUDE.md:211–212` now explicitly labels it "a Go-only dungeon-layout determinism contract with no GDScript consumer." |
| Add `cross_checks()` linking `dungeon_obstacles` golden to `dungeon_generation.v0.json` | **RESOLVED** | `tools/validate_dungeon_goldens.py` implements two-layer validation (wall-layout contract + achievability under generation caps). `tools/test_validate_dungeon_goldens.py` has 5 tests. Called from `validate_shared.py:1450`. |

---

## Top 5 shared/tooling/process improvements

1. **[High] Update `docs/CODEMAP.md` for v309–v331.** The "Fog of war / line of sight" row (`CODEMAP.md:16`) must add `fog_presentation_loader.gd`, `hero_light_source.gd`, `hero_visibility_field.gd`, `discovery_minimap.gd`, `discovery_minimap_state.gd`, and `shared/assets/fog_presentation.v0.json`. This is the only v308 carry-over still open. Every review has flagged it.

2. **[Med] Extend `validate_codemap.py` to catch unlisted present files.** Add an inverse check: for each directory CODEMAP tracks (e.g. `client/scripts/*.gd`), compare disk contents against CODEMAP tokens and report files present on disk but absent from the map. This makes the CI gate actually catch CODEMAP drift instead of requiring a human reviewer to notice it every 10 slices.

3. **[Med] Add `required` entries to `fog_presentation.v0.schema.json`'s `point_light` object block.** `shadow_enabled`, `energy`, and `attenuation` are semantically mandatory (a missing `shadow_enabled` produces undefined GDScript behavior). The schema's `additionalProperties: false` is correct but insufficient without `required` (`shared/assets/fog_presentation.v0.schema.json:99–120`).

4. **[Med] Add a `cross_checks()` entry for `fog_presentation.v0.json` sentinel ranges.** `falloff_power` should be checked to be in a physically meaningful range (e.g. 0.5–8); `shadow_reach_multiplier` should be checked to be at least 1.0. This follows the pattern already established for dungeon generation and boss patterns and prevents silent-screen bugs from accidental near-zero tuning edits.

5. **[Low] Document the "presentation bundle" SDD exception or eliminate it.** v322–v328 are retroactively bundled into one commit with no specs or plans. If this is an accepted pattern for pure VFX slices with no shared-contract surface, add a one-paragraph exception to the SDD section of `CLAUDE.md` with the condition (no shared JSON change, no new GDScript singleton, no protocol field). If it is not accepted, require at minimum a brief "scope note" doc in `docs/specs/` for bundled slices.
