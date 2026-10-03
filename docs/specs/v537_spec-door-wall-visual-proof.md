# v537 — Door, wall and entrance visual proof

- **Status:** Complete (combined `make ci` passed 2026-10-03, 20m36s)
- **Date:** 2026-10-02
- **Codename:** `door-wall-visual-proof`
- **Baseline:** `dab60eb5` plus v532, v533, v535 and v536 in the working tree.
- **Dependencies:** v533 (zoom), v535 (entrances) and v536 (props) are integrated; this slice is evidence only.
- **Source:** v531 review item 10 ("capture the door/wall presentation across representative layouts"), updated for the later playtest changes.

## Purpose

The v531 review found no retained, warning-clean capture of generated doors and walls; the v524–v531 generation work and now v533/v535/v536 change what players see. Take real-renderer frames of real generated floors (server running, windowed client) and record what they do and do not prove.

## Scope

- Add client bot scenarios (extended tier, `runner: godot_client`) that open generated room-corridor floors, wait for the wall layout and entities, and `capture_frame` several named frames: a room entrance and wall run at the new default zoom, the same at zoomed-in, a closed room-threshold door, and a deep floor with props. No movement setup (movement-contract allowlist): use lab worlds and `adjust_camera_zoom`.
- Retain the best frames and a short manifest under `docs/as-built/assets/v537/`.
- Check that the runs exit without Godot shutdown leak warnings, or record exactly which warnings appear and whether they predate this work.
- Closing note: record the Go `game` package runtime against the 10-minute default test timeout and the decision taken.

## Non-goals

No gameplay, rules, protocol or art changes; no performance claim beyond the logged counters; no exhaustive layout sweep.

## Acceptance criteria

1. At least four retained frames from real generated floors show: a room entrance with a wall run, a closed door, the wider entrance next to the player, and props on a deep floor.
2. Each frame names its scenario, seed, level and zoom; the as-built states what each shows and what it cannot prove.
3. The capture runs' warnings are listed; shutdown leak warnings are either absent or attributed (pre-existing vs new) by A/B on the unmodified baseline.
4. New scenarios pass the scenario movement audit and ci-pack validators; no `make ci`.

## Surfaces and verification

`tools/bot/scenarios/client/*.json` (new), `docs/progress/scenario-movement-audit.tsv`, `docs/as-built/assets/v537/`, as-built/lifecycle/PROGRESS. Focused: `make bot-client SCENARIO=<id> HEADLESS=0` per scenario, `pytest tools/test_scenario_movement_audit.py`, `pytest tools` ci-pack checks.

## Risks

Windowed runs depend on the host GPU and window; a transient login or timing failure may need a rerun (recorded, not hidden). A fixed seed's layout can change with later generation edits; frames are evidence for this commit only.
