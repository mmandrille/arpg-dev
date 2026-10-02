# v520 Plan - Fog Wall Occlusion

- **Status:** Complete — focused slice gates and combined batch `make ci` passed; the scenario 77 T-face readability limit remains documented.
- **Date:** 2026-10-01
- **Base commit:** `3a132626aa29581da6e3fdb36833c80e22b77b12`
- **Goal:** Remove fog-light leaks at connected wall runs, corners, and T-junctions while
  preserving wall readability and existing visual/gameplay boundaries.
- **Prerequisites:** Existing v255 wall-shadow geometry, v262 supplied door occluders, and v264
  organic fog mask are present at the assigned base. No sibling slice dependency is required.

## Review Gates

### Spec review gate - PASS

- Scope is limited to client-rendered wall shadow geometry and its existing cache.
- Criteria map to deterministic geometry tests, real-camera scenarios 68/77, and maintainability.
- The coordinator's visual follow-up authorizes bounded presentation-schema and test-fixture updates,
  plus a small validated bot camera-zoom action to frame scenario 77. No protocol, server, replay,
  or authoritative gameplay behavior changes are in scope.
- Existing project-native geometry, shader, cache, and bot harness are adopted/borrowed; external
  assets and add-ons are rejected.
- No blocking product decision or sibling-file conflict was found.

### Plan review gate - PASS

- The coordinator approved the plan and emphasized checking connected corner/T shadow coverage
  without filling large open/concave areas, plus detached geometry and door behavior.
- The accepted spec is preserved. Connected components are derived from existing normalized
  LOS rectangles; the implementation must avoid linking detached walls and must test the
  conservative-shadow risk for connected concave shapes.
- The work has a clear owner and runnable proof for every criterion. No generated gameplay values,
  protocol contracts, or authoritative logic are changed.
- The follow-up adds one reusable polygon-layer presenter and remains outside `main.gd`; shared edits
  stay limited to fog presentation values and the explicit blocker-lab fixture.

## File Map and Ownership

| Action | Path | Responsibility |
|---|---|---|
| Modify | `client/scripts/hero_visibility_field.gd` | Group touching/overlapping normalized blockers and produce continuous projected shadows per connected silhouette. |
| Modify | `client/scripts/fog_of_war_overlay.gd` | Add the low-alpha irregular soft rim under the existing gloom/core layers. |
| Modify | `client/scripts/fog_presentation_loader.gd` | Supply safe defaults for new schema-owned shadow-rim values. |
| Add | `client/scripts/fog_shadow_polygon_layers.gd` | Own cached soft-edge, gloom, and core polygon nodes and their draw order. |
| Modify | `client/tests/test_fog_of_war_overlay.gd` | Verify straight runs, corner and T-junction continuity, detached blockers, and door compatibility. |
| Modify | `shared/assets/fog_presentation.v0.json` | Configure rim opacity, scale, and organic variation. |
| Modify | `shared/assets/fog_presentation.v0.schema.json` | Bound the added presentation values. |
| Modify | `shared/rules/worlds.v0.json` | Add one T-shaped wall connection to the existing visual blocker lab fixture. |
| Modify | `client/scripts/bot_action_step_validator.gd`, `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_controller.gd` | Validate and dispatch the bounded visual-test camera zoom action without touching the main coordinator. |
| Modify | `client/tests/test_client_bot.gd` | Cover required, bounded, and numeric zoom-action inputs. |
| Modify | `tools/bot/scenarios/client/68_fog_los_shadow_mask.json` | Save an inspectable real-camera frame for connected dungeon walls. |
| Modify | `tools/bot/scenarios/client/77_line_of_sight_blocker_shadow.json` | Save an inspectable real-camera frame for explicit LOS blockers. |
| Modify | `docs/CODEMAP.md` | Clarify the connected-wall silhouette ownership in the fog-of-war client mapping. |
| Add | `docs/as-built/v520_fog-wall-occlusion.md` | Record implementation, checks, captures, and limitations. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v520 handoff row with current evidence state. |

Do not edit protocol/server/replay/aggro/radius paths or `main.gd`. The existing scenarios did not
save frames, so each receives one windowed-only `capture_frame` step. Scenario 77 also uses the
approved T-shaped blocker fixture and a bounded zoom step so the real client capture can frame it.

## Maintenance and Data Ownership

- No gameplay tuning is introduced or changed.
- Keep any new geometry logic inside the existing presentation owner unless maintainability checks
  show a focused helper extraction is needed.
- Check the touched client files against `.maintainability/file-size-baseline.tsv`; do not grow an
  over-limit file beyond its allowance.

## Ordered Tasks

### 1. Connected blocker geometry

- [x] Add deterministic connected-component derivation for normalized rectangular LOS occluders.
- [x] Use the shared silhouette when deriving a projected shadow for one connected wall run; keep
  detached components independent and preserve door occluder behavior.
- [x] Keep the shadow start at the visible far side of the wall and preserve existing projection,
  height, gloom/core, and debug output conventions.
- [x] Add a low-alpha, deterministic irregular outer rim below the unchanged geometry-driven core;
  use schema-backed fog presentation values and assert the rim stays out of the U-shaped opening.
- [x] Add geometry fixtures for a straight run, right-angle corner, T-junction, detached blockers,
  and supplied door blocker. Assert the immediate behind-joint probe points lie in the generated
  dark shadow and do not link detached objects.
- [x] Run the existing protocol LOS scenario after extending its shared lab with the connected
  crossbar; verify the monster remains hidden and becomes visible from the clear angle.

```bash
make bot scenario=line_of_sight_blockers
```

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_fog_of_war_overlay.gd
```

### 2. Cache regression proof

- [x] Confirm static wall geometry reuses cached shadows.
- [x] Confirm replacing wall layout triggers a complete combined-geometry rebuild.
- [x] Adjust cache tests only if a cache API or invalidation path changes.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_fog_los_shadow_cache.gd
```

### 3. Actual isometric camera proof

- [x] Run scenario 68 and inspect the saved capture for a generated straight/corner wall join,
  behind-wall darkness, and readable wall faces.
- [x] Run scenario 77 with bounded camera zoom, move to the T-junction, and inspect its connected
  shadow edge in a real-camera capture. Record that the town scene remains too dark to establish
  T-wall-face readability conclusively.
- [x] Record exact capture paths and any visual limitation; do not infer performance from captures.

```bash
make bot-visual scenario=68_fog_los_shadow_mask
make bot-visual scenario=77_line_of_sight_blocker_shadow
```

### 4. Handoff documentation and focused gates

- [x] Update CODEMAP ownership description.
- [x] Add as-built evidence and lifecycle row; keep PROGRESS current-status ownership with the
  coordinator unless a v520 integration status update is required there.
- [x] Run focused client suite and maintainability checks; record unrun combined gates as coordinator
  work.
- [x] Report complete changed/deleted/untracked files and ignored evidence, base SHA, exact command
  outcomes, captures, measured limits, unmet criteria, dirty state, and conflicts.

```bash
make client-unit
make maintainability
git diff --check
```

## Final Slice Gate

The worker stops after handoff. Do not run `make ci`, `make ci-full`, `/finish`, commit, or push.
The coordinator integrates the detached worktree and runs the combined batch gate after all accepted
slices are ready.
