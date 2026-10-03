# v537 As-built — Door, wall and entrance visual proof

- **Status:** Complete — focused slice verification plus the combined `make ci` on the integrated v532–v537 state (CI OK, 11 stages, 20m36s, 2026-10-03). `make ci-full` was not run.
- **Date:** 2026-10-02
- **Spec:** [`v537_spec-door-wall-visual-proof.md`](../specs/v537_spec-door-wall-visual-proof.md) · **Plan:** [`v537_2026-10-02-door-wall-visual-proof.md`](../plans/v537_2026-10-02-door-wall-visual-proof.md)
- **Scope:** two extended client scenarios and retained frames. No code, rules, protocol or golden change.

## What was added

- `tools/bot/scenarios/client/128_generated_entrance_frames.json` (`generated_entrance_frames`): `room_door_interaction_lab`, seed `v526_room_threshold_18` (a room threshold door about 7 units from the up stairs), frames at default, zoomed-in and zoomed-out camera.
- `tools/bot/scenarios/client/129_dungeon_props_deep_frames.json` (`dungeon_props_deep_frames`): `dungeon_dressing_deep_lab`, level −4, frames at default and zoomed-in.
- Both are `ci_tier: extended`, use no navigation (movement audit rows added), and use the existing `capture_frame` step (windowed runs only).

## Frames ([`assets/v537/`](assets/v537/))

| File | Shows |
|---|---|
| `v537_entrance_zoomed_out.png` | New default-plus zoom out: the hero in a hex-tiled room, a wooden room-threshold door in a wall run at the top right with torch light, and walls with kit corners. Entrances are visibly wider than the character. |
| `v537_entrance_default.png`, `v537_entrance_zoomed_in.png` | The hero and floor at the new default and a zoomed-in view; wall corners and torch light at the room edge. |
| `v537_props_deep_default.png`, `v537_props_deep_zoomed_in.png` | A depth-4 floor: the hero at a wall corner/room mouth with the wide corridor opening, torches, and monsters in the corridor. |
| `../v536/fixture-room-props.png` | (v536 evidence) fixture room with barrels, crates and a table drawn from server-style prop walls. |

## What this proves and what it does not

- Proves the real client renders generated floors after v533/v535/v536 without script errors, with kit walls, doors and torches intact, and that a closed room-threshold door sits in a wall run at the new zoom.
- **Does not prove** the props themselves on a real generated floor: the player spawns where props are deliberately kept clear (stairs/spawn clearance) and no incidental movement is allowed by the movement contract, so none of the five frames puts a prop in view. Prop placement and collision are proven by the Go generation/sim tests and the protocol scenario; prop *rendering* by the fixture-room capture and client unit tests. A future slice with an allowed teleport/lab placement could close this.
- Does not prove entrance feel in play, door appearance across many layouts, or performance (FPS overlay values in the frames are single-host snapshots).
- Run notes: the windowed `capture_frame` step is flaky on this host (generated_entrance_frames timed out on the first three attempts and passed on the fourth; `dungeon_props_deep_frames` passed first time and on repeat). The retained client logs contain no Godot leak or ObjectDB warnings.

## Go test runtime and CI timeout

With v535 (wider openings make some searches costlier) and v536 (props) the isolated `go test ./internal/game` package takes about 600 s (596 s measured), against Go's default 10-minute per-package timeout. A full run with other host load timed out once (`panic: test timed out after 10m0s`). Decision: `go test -timeout 20m ./...` in `make test-go` (`make/server.mk`) and the `scripts/ci.sh` Go step. Props generation cost was reduced (the base blocked grid is built once per level and props only mark their own cells), and the new tests use small seed sweeps. The slowest remaining cost is the pre-existing multi-seed generation sweeps (single subtests up to ~77 s).

Strict generation audit on the final code (`ARPG_DUNGEON_GENERATION_AUDIT=1 ..._STRICT=1`, 20 seeds × 12 depths): 240/240 floors generated, 0 invariant findings (815 s).
