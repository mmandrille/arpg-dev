# v250 Plan - Boss-Specific Telegraph Decals

Status: Complete
Goal: Render distinct client-only boss telegraph decal shapes from existing metadata.
Architecture: Extend `BossVisualsController` marker construction with shape-specific mesh helpers,
store the chosen marker shape in entity debug state, and assert it through client tests and bot.
Tech stack: Godot client presentation, client bot, docs.

## Baseline and Asset Decision

Builds on v57 boss phase readability and v240 boss portrait/debug fields. The server already sends
telegraph type, hit shape, radius, width, and pattern id through `boss_phase_started`.

Asset/plugin decision:
- Adopt existing in-repo boss marker controller.
- Borrow lightweight code-native mesh/material construction.
- Reject external assets/plugins and texture decal imports.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/boss_visuals_controller.gd` | Build shape-specific telegraph meshes and debug fields |
| Modify | `client/scripts/main.gd` | Include decal shape in entity presentation debug state without growing the file |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match decal shape in entity presentation assertions |
| Modify | `client/tests/test_factories.gd` | Prove marker shape construction and cleanup |
| Add | `tools/bot/scenarios/client/66_boss_telegraph_decals.json` | Client bot proof |
| Add | `docs/as-built/v250_boss-specific-telegraph-decals.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines, except grandfathered baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep `main.gd` line-neutral by extending the existing debug row.
- [x] Keep new mesh construction inside `boss_visuals_controller.gd`, which is below the limit.

Verification:
```bash
make maintainability
```

## Task 1 - Shape-specific marker meshes

Files:
- Modify: `client/scripts/boss_visuals_controller.gd`
- Modify: `client/tests/test_factories.gd`

- [x] Resolve marker shape from pattern id and telegraph hit shape/type.
- [x] Render distinct circle/summon, line, cone, and melee-contact meshes.
- [x] Preserve color/radius material behavior and cleanup.
- [x] Prove construction and cleanup in a focused Godot test.

```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 2 - Bot-visible debug proof

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Add: `tools/bot/scenarios/client/66_boss_telegraph_decals.json`

- [x] Expose `telegraph_marker_shape` in presentation debug rows.
- [x] Match `telegraph_marker_shape` in bot entity reaction assertions.
- [x] Add a client bot scenario that observes line, summon-circle, and cone decals.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=66_boss_telegraph_decals.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v250_boss-specific-telegraph-decals.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_factories.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `make bot-client scenario=66_boss_telegraph_decals.json HEADLESS=1`
- [x] `make maintainability`
