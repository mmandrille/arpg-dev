# v509 Spec — Boss Lane Telegraphs

- **Status:** Implemented and combined-batch-CI verified; live reconnect and matched render-cost evidence remain partial.
- **Date:** 2026-10-01
- **Codename:** `boss-lane-telegraphs`
- **Batch base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (detached v509 worktree)
- **Depends on:** the completed v250 shape-specific decal path, v282 rectangle hit predicate,
  v287 two-template boss deck, and v498 warning-readability baseline. No unfinished sibling slice
  is a prerequisite; recheck the integrated baseline before execution.

## Purpose

Add a deterministic boss attack whose staged warning marks adjacent floor lanes and leaves one
clearly readable route safe through the entire warning and strike. The server selects and locks the
lane geometry and owns damage. The Godot client uses the existing boss telegraph presentation path
to draw the same geometry. The default proposal is a **stable safe lane during one warning**; the
safe lane may change between separate attacks by a data-authored deterministic sequence.

## Product decision

The owner chose a lane fixed at the telegraph's start position. The warning does not track player
movement. The safe lane may change only between separate attacks through the authored sequence.

## Non-goals

- No new boss arena layout, obstacle generation, boss art, audio, imported VFX, or plugin.
- No client-owned hit detection or damage, random lane shifts, wall-clock scheduling, or hidden
  strike outside the warned lanes.
- No broad retuning of existing boss patterns, stats, or cooldowns; no new boss template.
- No general correction of older line/cone aim decals beyond what the new lane pattern requires.

## Observable acceptance criteria

1. A schema-backed `boss_patterns` entry defines lane count/width/length, warning-stage durations,
   active/recovery durations, damage, and a deterministic safe-lane sequence. Rule validation
   rejects zero-width lanes, invalid indices, a missing safe lane, stage timing below the minimum
   telegraph, and strike geometry that differs from the fully warned danger lanes. The same entry
   is appended to both existing boss decks after their currently bot-proven patterns.
2. At warning start, the server fixes world-space lane origin/orientation and one safe-lane index
   for that attack. The safe corridor stays unchanged through every warning stage and the active
   phase. The full lane layout and safe corridor are visible from the first stage; later stages
   visibly build intensity in their assigned danger lanes without turning the safe corridor
   dangerous. The boss cannot carry the warning or strike zone away by moving. A cast is valid only
   when the chosen safe corridor is walkable within the affected floor area; if no valid frame is
   available among the data-authored aim offsets, the server skips that cast rather than presenting
   an impossible warning.
3. Each active hit uses only the server's locked lane geometry. A player whose collision footprint
   lies fully within the safe corridor takes no damage from this pattern; a player in a warned
   danger lane can be hit once by the normal combat outcome rules. Boundary behavior is defined
   and covered by focused tests. Damage never begins before the configured warning finishes.
4. Authoritative phase events **and** reconnect/snapshot state expose enough geometry, selected
   safe lane, and current stage for the client to reconstruct the exact warning without guessing
   from boss facing, local tick time, or rule constants. Existing phase bars, boss tint, marker
   cleanup, and other boss patterns remain functional. Update the current v8 state-delta/session
   snapshot schemas and wire validation for any added fields; avoid unrelated historical schemas.
5. The real Godot renderer shows the complete lane footprint and unambiguous safe corridor from
   the player camera in the boss-floor lab, in both Balanced and Performance quality tiers. The
   stage change and strike align with server zones. The warning remains legible against v498
   lighting and fog. Reconnect during the warning restores the correct stage and marker.
6. A pinned-seed Go test reproduces the same lane selection, stage/event order, damage, and final
   state under identical ordered inputs. Focused tests cover safe, danger, and edge positions,
   stage timing, replay equivalence, validation failures, and snapshot reconstruction. A protocol
   bot scenario observes the warning and avoids the strike; a Godot client scenario observes the
   staged marker and captures a real-renderer frame.

## Likely surfaces and ownership

| Surface | Likely paths | Ownership |
|---|---|---|
| Shared rules | `shared/rules/boss_patterns.v0.json`, `.schema.json`, `boss_templates.v0.json` | Tunable pattern and stage data |
| Wire contract | `shared/protocol/state_delta.v8.schema.json`, `session_snapshot.v8.schema.json` | Authoritative lane descriptor in event/state |
| Server | `server/internal/game/boss_pattern_rules.go`, `boss_patterns.go`, `rules.go`, `types.go`, focused boss files/tests | Deterministic selection, locked geometry, collision and events |
| Client | `client/scripts/boss_visuals_controller.gd`, focused lane-marker helper/test; existing boss bar path | Render and debug state only |
| Bot and docs | New focused protocol and client scenarios; as-built/progress at closeout | Gameplay and real-camera proof |

The current line/rectangle marker has shape and size but no authoritative aim in its wire view,
and its mesh follows the moving boss. A new lane descriptor must carry a fixed world-space frame
through phase events and snapshots; the client must render it in world space. The implementation
should keep this descriptor narrow to the new pattern.

## Asset and plugin decision

- **Adopt:** `BossVisualsController` marker lifecycle, in-repo code-native MeshInstance3D/material
  patterns, boss bar, and boss-floor lab scenes.
- **Borrow:** existing rectangle telegraph geometry/testing patterns and v498 real-camera captures.
- **Reject:** external assets, Godot plugins, new decal pipeline, and imported effects for this slice.

## Focused verification and visual proof

Planned focused checks: `make validate-shared`, boss-pattern Go tests and determinism lint,
focused Godot lane-marker tests / `make client-unit`, `make bot scenario=boss_lane_telegraphs`,
`make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1`, and `make maintainability`.
The exact real-camera command will be `make bot-visual scenario=boss_lane_telegraphs_visual`.
Inspect the resulting PNGs and record their paths. Compare the boss-warning view across Balanced
and Performance tiers; measure frame/draw-call impact if new meshes make a smoothness claim.
The coordinator owns the combined `make ci` after batch integration.

## Integration risks

- Current boss movement runs before phase advancement and only pauses in active phases. Lock the
  new lane frame independently of subsequent boss movement (or pause this pattern's movement)
  before damage and rendering can diverge.
- Existing boss phase snapshots expose timing but may omit telegraph geometry. Reconnect proof
  must inspect the actual snapshot path, not only live phase events.
- Append deck entries so existing sequence-dependent boss bot scenarios still observe their
  expected earlier patterns; increase scenario budgets only where necessary and rule-derived.
- Sibling presentation slices may touch shared client presentation helpers. Compare their
  integrated paths and behavior before handoff; keep lane work in focused files where possible.
