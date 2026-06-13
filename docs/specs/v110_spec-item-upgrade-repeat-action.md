# v110 Spec: Item Upgrade Repeat Action

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `item-upgrade-repeat-action`

## Purpose

Extend the v94 account-stash upgrade route so stash equipment can be upgraded repeatedly up to a
data-driven maximum level. Each upgrade remains server-owned, gold-funded, deterministic, and
limited to one existing rolled numeric stat, but the gold cost now scales with the item's current
`item_level`.

## Non-goals

- No Godot blacksmith UI, town NPC, protocol bot, resource drops, failure chance, item bricking,
  recipe tiers, or random affix addition.
- No inventory/equipped/hotbar upgrade path; only account stash items remain mutable.
- No market restriction for upgraded items; upgraded items remain listable/tradeable for now.
- No new audit/history table.

## Acceptance criteria

- `shared/rules/main_config.v0.json` owns the repeat-upgrade tuning:
  `item_upgrade_cost_gold`, `item_upgrade_cost_growth_per_level`, and `item_upgrade_max_level`.
- The existing `POST /v0/account-stash/items/{stash_item_id}/upgrade` route computes cost as:
  `base_cost + current_item_level * growth_per_level`.
- Successful repeated upgrades spend the computed stash gold cost, increment `item_level`, preserve
  unrelated rolled payload keys, and increase one existing numeric rolled stat by 1 using stable
  deterministic stat-key order.
- The upgrade rejects missing/foreign items, insufficient stash gold, non-equipment/no-roll items,
  and items at `item_upgrade_max_level`.
- Store and HTTP tests prove first upgrade, second upgrade with higher cost, persisted stats, gold
  deduction, and max-level rejection.

## Scope and likely files

- Shared rules: `shared/rules/main_config.v0.json`,
  `shared/rules/main_config.v0.schema.json`
- Server store: `server/internal/store/repos.go`, `server/internal/store/interfaces.go`,
  `server/internal/store/store_test.go`
- Server HTTP: `server/internal/http/account_stash.go`,
  `server/internal/http/auth_session_test.go`
- Lifecycle docs: `PROGRESS.md`, `docs/as-built/v110_item-upgrade-repeat-action.md`

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/store -run TestAccountStashItemUpgrade -count=1`
- `cd server && go test ./internal/http -run TestAccountStashItemUpgrade -count=1`
- `make test-go`
- `make maintainability`
- `make ci`

No protocol bot is required because this slice extends an authenticated HTTP account-stash route and
does not change realtime gameplay protocol, world generation, combat, inventory intents, or client
presentation.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Repeat cost formula? | Use the conservative linear formula `base + current_level * growth`, owned by main config. |
| Q-2 | Market eligibility? | Preserve v94 behavior: upgraded stash items remain listable/tradeable until a dedicated market-restriction slice. |
| R-1 | Accidental tuning locks in tests. | Tests should derive base/growth/max from loaded rules where possible, or use focused arguments at store boundary. |
| R-2 | JSON mutation can drop rolled payload fields. | Preserve unrelated keys while changing only `item_level` and the selected numeric stat. |

## ADR alignment

- ADR-0012: advances item-level mutation through a deterministic, server-owned upgrade contract
  while deferring resources, success rates, and recipe tiers.
- ADR-0014 D3/D12: keeps gold valuable and leaves upgraded items feeding stash/market goals.
