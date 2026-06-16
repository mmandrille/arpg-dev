# v221 Spec - Resource Wallet Foundation

Codename: `resource-wallet-foundation`
Status: Complete
Baseline: v220 `mercenary-death-loss`

## Purpose

Move the configured upgrade resource (`upgrade_shard`) out of character inventory and into an
account-wide resource wallet. Picking up an upgrade shard should increase an account-scoped balance,
the blacksmith panel should use that balance for upgrade availability, and accepted upgrades should
spend from the wallet.

The slice follows the `$autoloop` clarification that resource wallet balances default to
account-wide ownership.

## Non-goals

- No multi-resource recipe UI, resource exchange, market restrictions, stash material tabs, or
  resource trading.
- No migration of pre-existing inventory shards into wallet balances.
- No new resource art, plugins, or production wallet panel.
- No changes to gold wallets, item upgrade chance, pity, or item-level formulas.
- No account-wide wallet for arbitrary non-upgrade items.

## Adopt / Borrow / Reject

- **Adopt:** Existing `upgrade_shard` item/rules and blacksmith resource requirement config.
- **Borrow:** Account stash gold persistence pattern: durable account-owned balance, session-start
  snapshot, state delta update, and client-bot assertion style.
- **Reject:** New assets/plugins and a broad material inventory interface.

## Acceptance Criteria

1. `upgrade_shard` pickup removes the loot entity and increments `resource_wallet[upgrade_shard]`
   instead of adding a bag item.
2. Session snapshots and state deltas expose the account resource wallet to clients.
3. The realtime persistence layer saves wallet increments to an account-wide durable table.
4. Inventory blacksmith upgrades require and spend the configured wallet resource count.
5. The blacksmith panel displays the wallet resource count and enables upgrades from the wallet.
6. Store and HTTP tests prove account scoping, spend/reject behavior, and blacksmith consumption.
7. A client/protocol proof picks up an upgrade shard, sees wallet count, upgrades, and sees the
   wallet count decrease.

## Scope and Files Likely Touched

- `server/migrations/0027_account_resource_wallet.sql`
- `server/internal/store/models.go`
- `server/internal/store/interfaces.go`
- `server/internal/store/repos.go`
- `server/internal/store/store_test.go`
- `server/internal/http/account_stash.go`
- `server/internal/http/auth_session_test.go`
- `server/internal/http/session.go`
- `server/internal/game/types.go`
- `server/internal/game/sim.go`
- `server/internal/game/sim_load.go`
- `server/internal/game/sim_players.go`
- `server/internal/realtime/runner.go`
- `server/internal/realtime/session_loop.go`
- `shared/protocol/session_snapshot.v8.schema.json`
- `shared/protocol/state_delta.v8.schema.json`
- `shared/protocol/examples/session_snapshot.json`
- `shared/protocol/examples/state_delta.json`
- `client/scripts/blacksmith_panel.gd`
- `client/scripts/main.gd`
- `client/scripts/net_client.gd`
- `client/scripts/bot_scenario_runner.gd`
- `client/scripts/bot_step_catalog.gd`
- `client/tests/test_shop_panel.gd`
- `tools/bot/state_ingest.py`
- `tools/bot/runtime_assertions.py`
- `tools/bot/run.py`
- `tools/bot/scenarios/72_upgrade_resource_drop.json`
- `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`
- `docs/as-built/v221_resource-wallet-foundation.md`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/store ./internal/http -run 'ResourceWallet|Upgrade' -count=1`
- `make bot scenario=upgrade_resource_drop`
- `make client-unit`
- `make bot-client scenario=blacksmith_upgrade_ui`
- `make maintainability`

Manual visual check, if desired:

```bash
make bot-visual scenario=blacksmith_upgrade_ui
```

## Open Questions and Risks

- Risk: this adds a protocol field and change op but keeps the version at v8 because the project is
  in active development and the current client/server are updated together.
- Risk: existing inventory shards are not migrated. The active development policy favors the clean
  wallet model over compatibility for old local state.
