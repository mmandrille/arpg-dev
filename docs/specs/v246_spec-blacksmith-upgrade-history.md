# v246 Spec - Blacksmith Upgrade History

Status: Complete
Date: 2026-06-17
Codename: blacksmith-upgrade-history

## Purpose

Give players immediate blacksmith feedback after upgrade attempts by showing a compact recent
history in the blacksmith panel. Each entry should name the recipe, item, result, and gold spent so
players can see what just happened without relying only on the transient status line.

## Non-goals

- No server/protocol changes, durable audit log, account-wide history, timestamps, filtering,
  pagination, analytics, or market/trade receipt integration.
- No recipe balance changes, material tuning, new recipes, icons, art, or external assets.

## Client Asset / Plugin Decision

- **Adopt:** Existing blacksmith panel, `DraggableWindow`, and text-first UI style.
- **Borrow:** Existing upgrade response fields (`success`, `cost_gold`) and selected recipe label.
- **Reject:** External assets/plugins and production history icons.

## Acceptance Criteria

- After each blacksmith attempt, the panel shows a recent-history section.
- Each history entry includes selected recipe label, item display name, success/failure wording, and
  gold spent.
- History keeps the newest entry first and caps itself to a small fixed number of entries.
- History is exposed through blacksmith debug state for tests.
- A focused Godot unit test proves history recording, ordering, cap, and hidden empty state.
- A client bot scenario performs an upgrade and verifies the panel remains usable after the history
  entry is recorded.

## Scope and Likely Files

- Client: `client/scripts/blacksmith_upgrade_history.gd`, `client/scripts/blacksmith_panel.gd`.
- Tests: `client/tests/test_blacksmith_panel.gd`.
- Bot/scenario: `tools/bot/scenarios/client/63_blacksmith_upgrade_history.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `make bot-client scenario=63_blacksmith_upgrade_history.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The slice keeps history local to the currently running client.
- Risk: `blacksmith_panel.gd` is near 600 lines. Keep history rendering in a new focused helper and
  avoid growing `main.gd`.
