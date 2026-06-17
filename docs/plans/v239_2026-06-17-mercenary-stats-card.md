# v239 Plan - Mercenary Stats Card

Status: Complete
Goal: Add a compact active-companion stats card to the mercenary panel.
Architecture: Render derived client-side details from existing companion state; leave server
contracts and companion AI unchanged.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v207 mercenary roster UI, v208 stance commands, and v232 recovery UI. Asset/plugin
decision: reject external assets/plugins; this is a text stats card in the existing draggable panel.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/mercenary_panel.gd` | Render stats card and expose debug state |
| Modify | `client/scripts/bot_mercenary_panel_assertions.gd` | Assert card text |
| Modify | `client/scripts/bot_step_catalog.gd` | Allow card assertion fields |
| Modify | `client/tests/test_mercenary_panel.gd` | Prove card content and hide behavior |
| Add | `tools/bot/scenarios/client/56_mercenary_stats_card.json` | Client proof |
| Add | `docs/as-built/v239_mercenary-stats-card.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] None expected; focused mercenary files are below 600 lines.

Decision:
- [x] Use existing companion state only; do not expose attack/armor through client contracts.
- [x] Extend the focused mercenary bot assertion helper instead of the large scenario runner.

Verification:
```bash
make maintainability
```

## Task 1 - Stats card UI/debug state

Files:
- Modify: `client/scripts/mercenary_panel.gd`

- [x] Add a compact stats card below the roster.
- [x] Render name, HP, stance, state, and id for the first active companion.
- [x] Hide the card when there are no companions.
- [x] Expose `stats_card_visible`, `stats_card_text`, and `stats_card_lines`.

## Task 2 - Tests and bot proof

Files:
- Modify: `client/tests/test_mercenary_panel.gd`
- Modify: `client/scripts/bot_mercenary_panel_assertions.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Add: `tools/bot/scenarios/client/56_mercenary_stats_card.json`

- [x] Assert card content in the focused unit.
- [x] Assert the card hides after loss clears the roster.
- [x] Add `stats_card_contains` to mercenary panel bot assertions.
- [x] Add a scenario that hires a mercenary and asserts HP/stance in the stats card.

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot-client scenario=56_mercenary_stats_card.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `docs/progress/slice-codename-index.md`
- Add: `docs/as-built/v239_mercenary-stats-card.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- [x] `make bot-client scenario=56_mercenary_stats_card.json HEADLESS=1`
- [x] `make maintainability`
