# v257 Plan - Full-Screen Map Overlay

Status: Complete
Goal: Add a full-screen discovery map mode and cycle map display modes with `TAB`.
Architecture: Keep map state client-presentational and session-local. Extend the existing
`DiscoveryMinimap` widget with display modes and layout changes; `main.gd` only invokes the new
cycle method. No protocol, server, shared rules, replay, persistence, or assets are required.
Tech stack: Godot GDScript client, Godot client bot scenario, docs.

## Baseline and Shortcut Decision

Builds on v256 discovery minimap. The existing minimap already owns explored cells, known walls,
player marker, optional elite-objective pin, and bot debug state.

Asset/plugin decision: reject external assets, imported map art, shader plugins, and Godot addons.
Borrow the existing v256 `DiscoveryMinimap` drawing and debug-state patterns.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/discovery_minimap.gd` | Add hidden/compact/full-screen modes, dynamic map sizing, centered overlay layout, and debug fields |
| Modify | `client/scripts/main.gd` | Use `cycle_display_mode()` for `TAB` |
| Modify | `client/tests/test_discovery_minimap.gd` | Prove mode cycle, full-screen size, and retained debug data |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Allow display-mode string assertions |
| Add | `tools/bot/scenarios/client/70_full_screen_map_overlay.json` | Client bot proof for compact and full-screen modes |
| Modify | `docs/specs/v257_spec-full-screen-map-overlay.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v257 lifecycle row |
| Add | `docs/as-built/v257_full-screen-map-overlay.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep new behavior inside `DiscoveryMinimap`; touch `main.gd` only at the existing TAB handler.

Verification:
```bash
make maintainability
```

## Task 1 - Display Modes and Layout

Files:
- Modify: `client/scripts/discovery_minimap.gd`
- Modify: `client/tests/test_discovery_minimap.gd`

- [x] Step 1.1: Replace binary visibility with hidden, compact, and full-screen display modes.
- [x] Step 1.2: Keep compact mode at the v256 HUD size and place it top-right.
- [x] Step 1.3: Add centered full-screen mode with a much larger map area and wider world radius.
- [x] Step 1.4: Expose debug fields for `display_mode`, `full_screen`, and current map size.
- [x] Step 1.5: Extend unit tests for mode cycling and full-screen sizing.

```bash
make client-unit
```

## Task 2 - TAB Wiring and Bot Proof

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/70_full_screen_map_overlay.json`

- [x] Step 2.1: Change the gameplay `TAB` handler to cycle discovery map modes.
- [x] Step 2.2: Let `assert_discovery_minimap` validate `display_mode`.
- [x] Step 2.3: Add a focused client scenario that asserts hidden, compact, full-screen, then hidden
  after repeated `TAB` presses.

```bash
HEADLESS=1 make bot-visual scenario=70_full_screen_map_overlay
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v257_spec-full-screen-map-overlay.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v257_full-screen-map-overlay.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the v257 spec complete.
- [x] Step 3.2: Add v257 lifecycle and as-built notes.
- [x] Step 3.3: Update `PROGRESS.md` current status and leave marker, active-session memory,
  biome, door, LOS, and quest marker work as remaining selected autoloop scope.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=70_full_screen_map_overlay`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=70_full_screen_map_overlay
```
