# v263 Plan - Quest Path Minimap Marker

Status: Complete
Goal: Add a directional minimap marker from the player toward the known active quest objective.
Architecture: Derive a normalized `quest_path` dictionary from the existing objective pin in
`DiscoveryMinimapState`; render it in `DiscoveryMinimap` with code-native lines/triangles; expose
debug fields for unit/bot assertions.
Tech stack: Godot GDScript minimap state/widget, client unit tests, bot scenario assertion.

## Baseline and Shortcut Decision

Builds on v256 discovery minimap, v257 display-mode cycling, and v258 points of interest. This slice
uses the existing objective pin as the source of truth and does not add routefinding or protocol
metadata.

Asset/plugin decision: reject external assets/plugins. Borrow the existing code-native minimap
marker style.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/discovery_minimap_state.gd` | Derive quest-path payload from active objective pin |
| Modify | `client/scripts/discovery_minimap.gd` | Draw quest-path line/arrow and expose debug state |
| Modify | `client/tests/test_discovery_minimap.gd` | Prove active and inactive quest-path marker behavior |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Allow bot assertions for quest-path debug fields |
| Modify | `tools/bot/scenarios/client/45_elite_objective_minimap_pin.json` | Assert quest-path marker in existing objective scenario |
| Modify | `docs/specs/v263_spec-quest-path-minimap-marker.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v263 lifecycle row |
| Add | `docs/as-built/v263_quest-path-minimap-marker.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and selected autoloop queue |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines and grandfathered files stay under their
ratchet allowance.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` was not touched.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep quest-path derivation and rendering inside minimap files.
- [x] Do not touch protocol, server, or `main.gd`.

Verification:
```bash
make maintainability
```

## Task 1 - State Payload

Files:
- Modify: `client/scripts/discovery_minimap_state.gd`
- Modify: `client/tests/test_discovery_minimap.gd`

- [x] Step 1.1: Add `quest_path` payload derived from active objective pin and player-centered map.
- [x] Step 1.2: Keep hidden/complete/missing objective cases inactive.
- [x] Step 1.3: Add unit assertions for active and inactive quest-path state.

```bash
make client-unit
```

## Task 2 - Widget Rendering and Debug

Files:
- Modify: `client/scripts/discovery_minimap.gd`
- Modify: `client/tests/test_discovery_minimap.gd`

- [x] Step 2.1: Render a directional code-native line/arrow in compact and full-screen modes.
- [x] Step 2.2: Add stable debug fields for marker active state and normalized endpoints.
- [x] Step 2.3: Preserve existing objective pin and POI marker debug behavior.

```bash
make client-unit
```

## Task 3 - Bot Scenario and Lifecycle Docs

Files:
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Modify: `tools/bot/scenarios/client/45_elite_objective_minimap_pin.json`
- Modify: `docs/specs/v263_spec-quest-path-minimap-marker.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v263_quest-path-minimap-marker.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Extend bot minimap assertions for quest-path debug fields.
- [x] Step 3.2: Update the elite-objective minimap scenario.
- [x] Step 3.3: Close out spec, lifecycle, as-built, and progress docs.

```bash
HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin`
- [x] `make maintainability`
- [x] Autoloop final batch gate: `make ci`
