# v94 Spec: Item Upgrade Starter

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-12
Codename: `item-upgrade-starter`

## Purpose

Add the first server-owned item upgrade action for account-stash equipment. A player can spend
account stash gold to upgrade one owned stash item. The v94 starter upgrade is guaranteed, increases
one existing rolled stat by 1, and increments `item_level` in the item's rolled payload.

## Non-goals

- No Godot upgrade UI, blacksmith NPC, protocol bot, resource drops, failure chance, item bricking,
  recipe tiers, or random affix addition.
- No inventory/equipped/hotbar upgrade path; only account stash items are mutable.
- No market restriction for upgraded items; upgraded items remain listable/tradeable.
- Advanced levels with special resources and failure chances are deferred.

## Acceptance criteria

- `shared/rules/main_config.v0.json` owns `item_upgrade_cost_gold` and `item_upgrade_max_level`.
- `POST /v0/account-stash/items/{stash_item_id}/upgrade` spends account stash gold and upgrades one
  owned stash item.
- The upgrade rejects missing items, insufficient stash gold, non-equipment/no-roll items, and items
  at max upgrade level.
- Upgraded rolled payloads include `item_level`, preserve existing stats, and increase one existing
  numeric stat by 1 using deterministic stat-key order.
- Store and HTTP tests prove success, persistence, insufficient-gold rejection, and max-level
  rejection.

## Test proof

- `make validate-shared`
- `cd server && go test ./internal/store -run TestAccountStashItemUpgrade`
- `cd server && go test ./internal/http -run TestAccountStashItemUpgrade`
- `make test-go`
- `make ci`

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Upgrade outcome? | Owner accepted guaranteed gold-only upgrade of one existing rolled stat plus `item_level`. |
| Q-2 | Future advanced costs/failure? | Documented as deferred: later advanced levels need special resources and failure chances. |
| Q-3 | Market eligibility? | Upgraded items remain tradeable/listable for now. |
| R-1 | Item mutation can lose rolled data. | Mutate JSON maps in store transaction and preserve all unrelated keys. |

## ADR alignment

- ADR-0012: establishes item-level mutation and server-owned cost/payment while deferring advanced
  resources, success chances, and recipe tiers.
- ADR-0014 D3/D12: keeps gold valuable and leaves upgraded items feeding market value.
