# v228 Spec: Resource Wallet Panel

Status: Complete
Date: 2026-06-16
Codename: resource-wallet-panel

## Purpose

Show the account resource wallet in the live HUD so upgrade materials are visible before opening the
blacksmith.

## Baseline

Builds on v221 resource wallet foundation and v222 upgrade result preview. The server already owns
wallet balances, emits `resource_wallet_update`, and the client keeps `resource_wallet` state for the
blacksmith. Asset/plugin decision: adopt the existing Godot HUD/control scripts and code-native
labels; reject external UI assets or plugins.

## Non-goals

- No new resource types, recipes, stash material tabs, trading, market restrictions, or server
  ownership changes.
- No inventory footprint changes or item auto-pickup behavior; v229 owns material auto-pickup.
- No production icons or art.

## Acceptance Criteria

- The HUD shows wallet-backed resources and counts while gameplay is active.
- The panel updates after snapshot load and after `resource_wallet_update` deltas.
- Zero/empty wallets hide the count area instead of showing noisy placeholders.
- The blacksmith client-bot scenario proves the wallet count appears after picking up an upgrade
  shard and reaches zero after spending it.

## Scope and Likely Files

- `client/scripts/character_bar.gd`
- `client/scripts/main.gd`
- `client/scripts/bot_assertion_handlers.gd`
- `client/scripts/bot_step_catalog.gd`
- `client/tests/test_character_bar.gd` or an existing client unit harness entry
- `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v228_resource-wallet-panel.md`

## Test and Bot Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1`

Manual visual proof, if desired:

- `make bot-visual scenario=39_blacksmith_upgrade_ui.json`

## Open Questions and Risks

- No blocking questions. The panel uses existing wallet state and does not change authoritative
  resource ownership.
