# v502 Plan — Skill-specific impact VFX

- **Status:** Complete; integrated and combined batch `make ci` passed in 11m41s.
- **Spec:** [`docs/specs/v502_spec-skill-specific-vfx.md`](../specs/v502_spec-skill-specific-vfx.md)
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`
- **Prerequisites:** v492 combat VFX foundation already exists at baseline. No slice dependency.

## Goal and execution boundary

Add catalog-driven impact bursts for successful Magic Bolt, Ice Shard, and Lightning hits while retaining `hit_spark` for all other hits and `death_burst` for death. Keep this work in client presentation and shared presentation data. The coordinator assigned execution after transferring the verified v501 prerequisite files. Source and static work can proceed; defer Godot, bot, visual-capture, and timing commands until the coordinator releases the serialized engine slot.

## Ownership and file map

| Surface | Planned files | Boundary |
|---|---|---|
| Effect presets and mapping | `shared/assets/vfx_presentation.v0.json`, `shared/assets/vfx_presentation.v0.schema.json` | Three skill mappings and bounded effect profiles; no gameplay tuning |
| Selection and runtime VFX | `client/scripts/combat_vfx.gd`, `client/scripts/skill_visual_capture.gd`, `client/scripts/skill_visual_capture_runtime.gd`, `client/scripts/main.gd`, `client/scripts/player_camera_controller.gd` | Resolve existing skill IDs through the catalog, capture only opted-in visual-replay hit events, apply configured replay-only camera focus/zoom, and preserve generic and death paths |
| Replay capture directive | `tools/bot/scenarios/44_skill_visual.json`, `tools/bot/skill_visual_runtime.py`, `shared/rules/worlds.v0.json` (`skill_visual_lab` only) | Configure one headless-skippable capture for each allowlisted skill hit, frame the contact area, and stage a ranged target outside unrelated Lightning chain hits while keeping it within cast range |
| Focused checks | `client/tests/test_combat_vfx.gd`, `tools/bot/test_skill_visual.py` | Catalog-derived mapping and parameters, fallback, death behavior, capture allowlist/deduplication, replay camera application, projectile fixture range/isolation, and rank assertions without freezing unrelated spend availability |
| Evidence | `docs/as-built/v502_skill-specific-vfx.md`; `.artifacts/bot-captures/v502_skill_<skill_id>.png` and `.json` as ignored evidence | Record exact commands, capture paths, and proof limits |
| Domain index | `docs/CODEMAP.md` | Add the currently missing VFX builder/catalog/schema/test paths to the existing visuals row; preserve v501 entries elsewhere in the same file |

No server/protocol behavior, new bot runner registration, asset manifest, or CI-pack change is planned. `tools/bot/run.py` already carries the scenario's `visual` object into its replay manifest, and the client reads it only in visual-replay mode. v502 uses that existing path with a skill allowlist, replay-only camera framing, and the existing viewport frame writer; it does not add a protocol client step or change general bot tooling. The existing `skill_visual` scenario is parameterized through `ARPG_SKILL_VISUAL_SKILL_ID` and can showcase each effect.

## Shared-file overlap and integration

The coordinator transferred v501's checksum-verified 11-path prerequisite set into this worktree. Its current manifest changes `gameplay_feedback_presentation.gd`, `model_reaction_controller.gd`, the monster-death loader/data/schema/test, `tools/bot/scenarios/client/103_combat_input_flow_polish.json`, CODEMAP, and v501 docs; it does not modify v502's VFX catalog, tests, or the `44_skill_visual` replay scenario. HEAD remains the original batch SHA and those v501 files are pre-existing worktree changes: preserve them and do not include them in the v502 handoff. Both slices traverse the existing entity-reaction path, so test that skill hits retain v501 behavior and that a death carrying a skill id still selects `death_burst`. The shared CODEMAP edit is an explicit overlap; preserve both slices' entries.

## Ordered tasks

### 1. Revalidate the assigned base and code ownership

- [x] Confirm detached HEAD is the recorded base; classify and preserve the coordinator-transferred v501 changes already present in the worktree.
- [x] Recheck `appendMonsterCombatEvent` still supplies `SkillID` on skill damage events.
- [x] Check maintainability limits for every file that implementation will touch (`make maintainability` passed before and after v502 source/data/CODEMAP edits).

### 2. Extend shared presentation data

- [x] Add `magic_bolt`, `ice_shard`, and `lightning` mappings to `skill_hit_effects` in `vfx_presentation.v0.json`.
- [x] Add three distinct, conservative one-shot effect profiles with readable colors and different spread, lifetime, velocity, gravity, and size ranges.
- [x] Update `vfx_presentation.v0.schema.json` to require and constrain the mapping while retaining strict object validation.
- [x] Validate the resulting data with `make validate-shared` (2,210 checks; CODEMAP valid).

### 3. Select per-skill effects in the client

- [x] In `CombatVfx.spawn_for_reaction`, select a configured per-skill preset only for hit reactions with a mapped `skill_id` and valid effect definition.
- [x] Keep generic hit fallback, damage-type color fallback, parent attachment, quality scaling, one-shot cleanup, and death selection intact.
- [x] Keep the render path free of new file paths or arbitrary text built from event values; resolve identifiers only through the loaded catalog.
- [x] Add catalog-derived cases to `client/tests/test_combat_vfx.gd` covering all three mappings, profile distinctions, unmapped/basic hit fallback, and unchanged death effect.
- [x] Add an opt-in `visual.capture_frame` directive to the existing skill visual replay scenario. Accept only the three configured `monster_damaged` skill hits, derive the burst through the VFX catalog, deduplicate each skill, and save deterministic filenames through the existing `BotFrameCapture` helper.
- [x] Add focused checks for event selection, allowlisting, deterministic names, and one-save-per-skill behavior. The capture is started after event presentation is applied; the existing 40-tick post-cast replay tail remains after it.
- [x] Keep the reusable skill visual assertion focused on seeded rank/max-rank. Runtime recording showed that a hit can grant spendable points, so `can_spend` is not a stable assertion for this visual fixture.
- [x] Apply `visual.camera` focus and zoom through a visual-replay-only `PlayerCameraController` entry point; frame the configured soft target only during visual replay.
- [x] Move projectile showcases within the soft target's contact range so the hit event and capture belong to the intended target rather than a nearer XP dummy.
- [x] Reposition the skill visual lab soft target beyond the one-hit XP dummies' Lightning chain radius while preserving cast range; add a catalog-derived geometry regression check so the level-up/death overlays stay out of the impact capture.

### 4. Focused runtime and visual proof

- [x] Run `make validate-shared` (2,210 checks; CODEMAP valid).
- [x] Run `.venv/bin/pytest tools/bot/test_skill_visual.py -q` (17 passed).
- [x] Run `make maintainability` after moving replay capture orchestration/diagnostics into `SkillVisualCaptureRuntime`; ratchets passed, dashboard at 247/250 lines.
- [x] Run `git diff --check`.
- [x] Run `make client-unit`; the coordinator's complete client-unit suite passed after all slices were integrated.
- [x] Run `make bot-client SCENARIO=client_skill_points_and_magic_bolt HEADLESS=1`; the slice handoff records PASS (1/1).
- [x] Run and inspect real-renderer captures for all three skills:
  - `ARPG_SKILL_VISUAL_SKILL_ID=magic_bolt make bot-visual scenario=skill_visual`
  - `ARPG_SKILL_VISUAL_SKILL_ID=ice_shard make bot-visual scenario=skill_visual`
  - `ARPG_SKILL_VISUAL_SKILL_ID=lightning make bot-visual scenario=skill_visual`
- [x] Confirm each command saves `.artifacts/bot-captures/v502_skill_<skill_id>.png` and its fixture `.json`; inspect all three PNGs.
- [x] Record each artifact path and state that these captures do not establish a performance change.

### 5. Handoff evidence

- [x] Write `docs/as-built/v502_skill-specific-vfx.md` with implemented behavior, capture results, static results, and the remaining Godot gate.
- [x] Report base/dependency revisions, worktree dirty state, complete changed/deleted/untracked list, ignored captures that must be preserved, and the v501 overlap; the coordinator reconciled the handoff manifest against the integrated tree.
- [x] Do not run `make ci`/`make ci-full`, `/finish`, commit, push, or clean this worktree; the worker followed this boundary and the coordinator ran the combined gate.

## Tunable ownership and maintainability

All new effect amounts, durations, velocities, gravity, spread, sizes, and colors belong in the schema-backed `shared/assets/vfx_presentation.v0.json`. Tests must derive expected profile values from this catalog instead of freezing those values as unrelated gameplay assertions. No new over-limit file is planned; check the ratchet before editing and split code only if the catalog-driven selection materially exceeds the current file's budget.

## Review result

Coordinator reviewed the bounded spec and assigned `/execute`. The slice adds no gameplay/protocol behavior and preserves v501 death handling. There is no material product question. v501 currently has no same-path overlap with v502; retain the cross-feature death regression because both effects use the existing reaction dispatcher.
