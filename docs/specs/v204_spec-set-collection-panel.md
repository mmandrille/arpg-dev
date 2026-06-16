# v204 Spec: Set Collection Panel

Status: Complete
Date: 2026-06-15
Codename: set-collection-panel

## Purpose

Add a player-facing set collection panel to the Godot client so set items feel like account projects instead of isolated green drops. The panel summarizes enabled set packages from existing item summary lines, showing owned/equipped progress and the active/inactive set bonus ladder without changing server authority or item persistence.

## Baseline

Builds on v203 upgrade pity counter and the existing v181/v194 set item pipeline:

- `shared/rules/set_items.v0.json` defines two enabled five-piece sets.
- The Go server already emits set membership and bonus summary lines on set item views.
- Existing inventory and stash presentations already color set rarity and set tooltip lines.

Asset/plugin decision: reject external assets/plugins for this slice. The panel is a data/UI summary using existing Godot controls, existing set item payloads, and current set rarity colors.

## Non-goals

- No new set items, set drop rates, set art, account-wide discovery persistence, achievements, or collection rewards.
- No server, store, protocol schema, or shared-rule changes unless a current client payload bug blocks the panel.
- No stash capacity/tab work and no changes to set item rarity semantics.
- No exact layout persistence beyond the existing draggable-window behavior if the panel uses it.

## Acceptance criteria

- Inventory UI exposes a set collection panel/control that can be opened while the player has set item data available.
- The panel lists at least Verdant Vanguard and Stormrunner Covenant when their pieces are present in inventory/equipment/chest rows, with readable progress such as `1/5 owned` and `0/5 equipped`.
- The panel marks individual known pieces as owned, equipped, or missing using the set item summary data already present on item rows.
- The panel shows set bonus tiers from summary lines and marks active tiers when the equipped count satisfies the tier.
- The panel updates when inventory/equipment state changes without requiring a reconnect.
- Client bot coverage opens the unique chest/inventory flow and asserts set collection debug state for at least one known set package.

## Scope and likely files

- Client UI:
  - Add a focused `client/scripts/set_collection_panel.gd` or equivalent helper/panel.
  - Integrate from `client/scripts/inventory_panel.gd` with a small button/summary entry.
  - Add focused client unit tests under `client/tests/`.
- Client bot:
  - Extend `client/scripts/bot_scenario_runner.gd` and `client/scripts/bot_step_catalog.gd` with a set collection assertion/action only if needed.
  - Add or update a client scenario under `tools/bot/scenarios/client/`.
- Docs:
  - Add v204 plan/as-built and lifecycle updates when shipped.

## Test and bot proof

- `make client-unit` covers parsing summary lines, progress states, and active bonus rows.
- `SCENARIO=set_collection_panel HEADLESS=1 ./scripts/bot_client_local.sh` or an updated unique chest client scenario proves the panel/debug state with live client data.
- `make maintainability` verifies new files stay under 600 lines and touched grandfathered files remain within the ratchet.
- `make ci` is the final gate.

## Open questions and risks

- Risk: the current server payload encodes set identity only in human summary text. For this slice, parsing that text is acceptable because the scope is presentation-only; a future server contract can add structured set metadata if collection persistence grows.
- Risk: `inventory_panel.gd` is over the line-count target. The plan must keep integration minimal and extract any meaningful set parsing/panel logic into new focused files.
