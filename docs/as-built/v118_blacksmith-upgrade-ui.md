# v118 As-built: Blacksmith Upgrade UI

Date: 2026-06-13
Spec: [`docs/specs/v118_spec-blacksmith-upgrade-ui.md`](../specs/v118_spec-blacksmith-upgrade-ui.md)
Plan: [`docs/plans/v118_2026-06-13-blacksmith-upgrade-ui.md`](../plans/v118_2026-06-13-blacksmith-upgrade-ui.md)

## What shipped

- Added a server-authored `town_blacksmith` interactable with service type `blacksmith`, placed in the
  main town and vendor lab.
- Added a Godot blacksmith panel that lists account-stash equipment, stash gold, current item level,
  next upgrade cost, and one-click upgrade actions.
- Wired the panel to the existing authenticated account-stash item upgrade route; the HTTP response
  updates the matching stash item and stash gold in local client state.
- Added client bot actions and scenario `39_blacksmith_upgrade_ui` to fund stash gold, deposit a rolled
  item, open the blacksmith, upgrade once, and assert item level/gold changes through debug state.
- Fixed the store upgrade helpers to support the current generated item payload shape where rolled
  stats live under `rolled_stats.stats`, while preserving legacy flat rolled stat payloads.
- Recorded a v118 maintenance exception for narrow growth in existing service/UI-bot/store-test
  hotspots after extracting the new blacksmith UI into `client/scripts/blacksmith_panel.gd`.

## Verification

```bash
make validate-shared
make client-unit
cd server && go test ./internal/game/...
cd server && go test ./internal/store/... -run 'TestAccountStashItemUpgrade'
make bot-client scenario=39_blacksmith_upgrade_ui
make maintainability
```

Manual visual proof:

```bash
make bot-visual scenario=39_blacksmith_upgrade_ui
```

## Deferred

Crafting resources, upgrade failure/success chance, recipe tiers, market restrictions for upgraded
items, production blacksmith art/audio, and inventory/equipped item upgrade paths remain deferred.
