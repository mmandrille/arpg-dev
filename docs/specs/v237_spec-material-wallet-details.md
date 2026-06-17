# v237 Spec - Material Wallet Details

Status: Approved for autoloop
Date: 2026-06-17
Codename: material-wallet-details

## Purpose

Make the compact material wallet HUD readable beyond the abbreviated count. Players should be able
to inspect wallet-backed materials and see the catalog display name, count, category, and account
storage context without opening the blacksmith.

## Non-goals

- No server/protocol changes, new resource types, drop tuning, wallet persistence changes, trading,
  stash material tab, blacksmith recipe changes, icons, or external assets.
- No standalone wallet window or inventory-like material grid.

## Acceptance Criteria

- The existing character HUD wallet readout keeps its compact text for nonzero balances.
- The wallet readout tooltip includes detail lines for each nonzero resource: catalog name, count,
  category when available, and account-wide storage context.
- Character-bar debug state exposes the tooltip/detail lines for tests and bot assertions.
- The resource label comes from shared item rules when available, with the previous fallback kept for
  unknown resource ids.
- A focused unit test proves `upgrade_shard` details are visible and zero balances remain hidden.
- A short client bot scenario auto-picks an upgrade shard and asserts the wallet detail tooltip.

## Scope and Likely Files

- Client: `client/scripts/character_bar.gd`, `client/scripts/bot_assertion_handlers.gd`.
- Unit tests: `client/tests/test_character_bar.gd`.
- Bot/scenario: `tools/bot/scenarios/client/54_material_wallet_details.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_character_bar.gd`
- `make bot-client scenario=54_material_wallet_details.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. This slice rejects external art/plugins and reuses shared item rules for
  names/categories.
- Risk: tooltip-only UI can be missed in headless proof. Debug state and bot `tooltip_contains`
  assertions provide deterministic coverage.
