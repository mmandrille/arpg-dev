# v258 Plan - Minimap Points of Interest

Status: Complete
Goal: Draw discovered stairs, waypoint, town-service, and objective markers on the discovery map.
Architecture: Keep marker derivation client-presentational and based only on known client entity
records. `DiscoveryMinimapState` owns marker classification and explored-cell gating; the widget
draws code-native marker shapes and exposes debug counts. No server/protocol/shared-rule changes.
Tech stack: Godot GDScript client, Godot client bot scenario, docs.

## Baseline and Shortcut Decision

Builds on v257 full-screen map overlay. Both compact and full-screen modes already use the same
state projection, so marker drawing belongs in the existing widget.

Asset/plugin decision: reject external icon art, shaders, plugins, and Godot addons. Borrow the
existing `DiscoveryMinimap` code-native shape drawing and bot debug-state pattern.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/discovery_minimap_state.gd` | Classify explored POI markers from known entities |
| Modify | `client/scripts/discovery_minimap.gd` | Draw marker shapes and expose marker debug counts |
| Modify | `client/tests/test_discovery_minimap.gd` | Prove marker derivation and debug counts |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Allow marker count assertions |
| Add | `tools/bot/scenarios/client/71_minimap_points_of_interest.json` | Client bot proof for service markers |
| Modify | `docs/specs/v258_spec-minimap-points-of-interest.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v258 lifecycle row |
| Add | `docs/as-built/v258_minimap-points-of-interest.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`: not planned
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep marker derivation in focused minimap scripts; no `main.gd` touch expected.

Verification:
```bash
make maintainability
```

## Task 1 - Marker State and Rendering

Files:
- Modify: `client/scripts/discovery_minimap_state.gd`
- Modify: `client/scripts/discovery_minimap.gd`
- Modify: `client/tests/test_discovery_minimap.gd`

- [x] Step 1.1: Derive marker records for explored stairs, waypoint, town services, and objective.
- [x] Step 1.2: Gate interactable markers by explored cell membership.
- [x] Step 1.3: Draw distinct code-native marker shapes/colors in compact and full-screen modes.
- [x] Step 1.4: Expose total and per-kind marker counts in debug state.
- [x] Step 1.5: Extend unit tests for marker classification and explored gating.

```bash
make client-unit
```

## Task 2 - Bot Proof

Files:
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/71_minimap_points_of_interest.json`

- [x] Step 2.1: Allow `assert_discovery_minimap` to check marker counts.
- [x] Step 2.2: Add a focused client scenario that opens the map in `vendor_lab` and asserts town
  service markers are present without changing server state.

```bash
HEADLESS=1 make bot-visual scenario=71_minimap_points_of_interest
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v258_spec-minimap-points-of-interest.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v258_minimap-points-of-interest.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the v258 spec complete.
- [x] Step 3.2: Add v258 lifecycle and as-built notes.
- [x] Step 3.3: Update `PROGRESS.md` current status and leave active-session memory, biome, door,
  LOS, and quest marker work as remaining selected autoloop scope.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=71_minimap_points_of_interest`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=71_minimap_points_of_interest
```
