# v118 Spec: Blacksmith Upgrade UI

Status: Complete
Date: 2026-06-13
Codename: `blacksmith-upgrade-ui`

## Purpose

Add a town blacksmith NPC that gives players a real client path into the existing account-stash item
upgrade mechanic. The blacksmith is a server-authored town interactable, while upgrade outcomes stay
owned by the existing authenticated `POST /v0/account-stash/items/{stash_item_id}/upgrade` route.

## Non-goals

- No new upgrade formula, success chance, resource type, failure state, recipe system, or item
  bricking.
- No inventory/equipped/hotbar upgrade path; v118 upgrades account-stash items only.
- No market restrictions for upgraded items.
- No production blacksmith art/audio; use code-native placeholder presentation.
- No protocol schema bump unless the current client cannot consume the HTTP response cleanly.

## Acceptance criteria

- Shared interactable data defines `town_blacksmith` as a ready town service and both the main town
  and vendor lab place it at reachable coordinates.
- Godot renders `town_blacksmith` as a distinct NPC and auto-approaches it like other town services.
- Clicking the blacksmith opens an upgrade-focused panel or stash panel mode that lists account-stash
  items, current stash gold, current item level, and the next upgrade cost.
- Pressing Upgrade on an eligible stash item calls the existing HTTP route, updates the local stash
  item and stash gold from the response, and shows a success or error status.
- Upgrade controls reject or disable clearly when there is no eligible stash item, insufficient gold,
  or the item is at the data-driven max level.
- A focused Godot client bot scenario funds stash gold, deposits a rolled item into account stash,
  opens the blacksmith, upgrades the item once, and asserts item level/stat/gold changes through
  debug state.

## Scope and likely files

- Shared rules: `shared/rules/interactables.v0.json`, `shared/rules/worlds.v0.json`
- Client networking: `client/scripts/net_client.gd`
- Client UI/presentation: `client/scripts/main.gd`, `client/scripts/inventory_panel.gd` or a new
  focused blacksmith panel helper
- Client bot: `client/scripts/bot_scenario_runner.gd`, `client/scripts/bot_controller.gd`,
  `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`
- Tests/docs: client unit coverage if helper extraction is used, `PROGRESS.md`,
  `docs/as-built/v118_blacksmith-upgrade-ui.md`

## Test and bot proof

- `make validate-shared`
- `make client-unit`
- `make bot-client scenario=39_blacksmith_upgrade_ui`
- `make ci`

Manual visual proof command:

```bash
make bot-visual scenario=39_blacksmith_upgrade_ui
```
## Open questions and risks

| Risk | Mitigation |
|------|------------|
| `client/scripts/main.gd` and panel files are already large. | Prefer a focused helper/panel file if the upgrade UI cannot be added without broad coordinator growth. |
| The HTTP route returns account-stash item data outside realtime fanout. | Treat the response as the authoritative refresh for the clicked item and stash gold, matching market UI HTTP patterns. |
| Client bot setup can be brittle if funding/deposit steps depend on previous panels. | Reuse existing vendor/stash client-bot steps and add only upgrade-specific assertions/actions. |
