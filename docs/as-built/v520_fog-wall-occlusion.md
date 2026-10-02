# v520 As-Built - Fog Wall Occlusion

Date: 2026-10-01
Spec: [`docs/specs/v520_spec-fog-wall-occlusion.md`](../specs/v520_spec-fog-wall-occlusion.md)
Plan: [`docs/plans/v520_2026-10-01-fog-wall-occlusion.md`](../plans/v520_2026-10-01-fog-wall-occlusion.md)
Base: `3a132626aa29581da6e3fdb36833c80e22b77b12`

## Implemented in the slice worktree

- Connected touching/overlapping rectangular LOS blockers into components before projecting fog
  shadows. Compact straight, corner, and T silhouettes share a continuous projected shadow;
  detached blockers stay independent. Large concave groups fall back to individual blockers so a
  U-shaped opening remains visible.
- Added a deterministic low-alpha irregular rim for accepted compact silhouettes. The existing
  gloom and geometry-derived core remain authoritative; concave fallback components do not receive
  the rim. Polygon node ownership and draw order live in `fog_shadow_polygon_layers.gd`.
- Tuned only shared fog presentation values (`soft_edge_alpha` 0.20, scale 1.12, amplitude 0.025,
  core alpha 0.76) with schema bounds and matching loader defaults. No server, protocol, replay,
  aggro, or authoritative LOS behavior changed.
- Added a T connection to the existing visual blocker lab and a bounded bot camera-zoom action used
  by scenario 77. The action rejects missing, nonnumeric, non-finite, or out-of-range deltas and
  calls the existing camera controller; `main.gd` was not changed.
- Geometry coverage includes straight, corner, T, detached, U-shaped, and supplied-door fixtures.

## Verification

| Command | Result |
|---|---|
| `godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_fog_of_war_overlay.gd` | PASS, 147 assertions |
| `godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_fog_los_shadow_cache.gd` | PASS, 12 assertions |
| `godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_client_bot.gd` | PASS, 347 assertions |
| `make client-unit` | PASS after polygon-layer extraction |
| `make validate-shared` | PASS, 2,260 checks; codemap validation passed |
| `make maintainability` | PASS after extracting polygon layers and trimming the fog test file below its limit |
| `make bot-visual scenario=68_fog_los_shadow_mask` | PASS; visible client capture saved |
| `make bot-visual scenario=77_line_of_sight_blocker_shadow` | PASS; visible client capture saved with zoom/reframe |
| `make bot scenario=line_of_sight_blockers` | PASS; protocol LOS fixture still hides/reveals the monster as expected |
| `git diff --check` and `python3 tools/validate_codemap.py` | PASS |
| Combined batch `make ci` | PASS — 11/11 stages, 7m50s; `make ci-full` was not run |

The captures use the actual isometric client renderer (Forward+, Metal, 3840×2100). Scenario 68
shows the level −1 generated dungeon and fog presentation. Scenario 77 now moves beside the
connected T-shaped blocker and zooms the real camera to its configured minimum. The updated capture
shows the light-to-shadow boundary and a low-contrast outer band, but the Town scene is still too dark and
the T wall faces are not distinct enough to claim strong visual readability. The capture proves the
scenario ran and that the shadow transition is inspectable; deterministic geometry fixtures remain
the proof for connected T coverage. These images do not establish performance.

The final captures and manifests are preserved with this report:

- `docs/as-built/assets/v520/v520_fog_los_shadow_mask.png` and `.json`
- `docs/as-built/assets/v520/v520_line_of_sight_blocker_shadow.png` and `.json`

The combined CI result is from the coordinator's fully integrated v518–v520 tree. `make ci-full`
was not run.
