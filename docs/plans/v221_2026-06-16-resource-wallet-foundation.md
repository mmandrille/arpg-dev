# v221 Plan - Resource Wallet Foundation

Status: Complete
Goal: move upgrade shards into an account-wide wallet and spend that wallet at the blacksmith.
Architecture: model resources as account-owned durable balances, load them into session snapshots,
and mutate them through authoritative sim changes. The sim treats the configured upgrade resource
as wallet currency on pickup; HTTP blacksmith upgrades consume the same account wallet. Client UI
continues using the existing blacksmith panel, with one additional wallet count field.
Tech stack: Go store/HTTP/realtime/sim, JSON protocol schemas/examples, Godot client, Python and
Godot bot scenarios, lifecycle docs.

## Baseline and Shortcut Decision

Builds on v180 upgrade shard drops, v202 blacksmith resource consumption, and v220 clean baseline.
Asset/plugin decision: borrow existing blacksmith panel resource row and bot assertions; reject new
art, plugins, or a standalone materials panel.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/migrations/0027_account_resource_wallet.sql` | Durable account resource balances and session-start snapshot table. |
| Modify | `server/internal/store/models.go` | Add resource wallet model. |
| Modify | `server/internal/store/interfaces.go` | Add narrow wallet read/add/spend methods and snapshot parameters. |
| Modify | `server/internal/store/repos.go` | Implement wallet persistence and session snapshot load/save. |
| Modify | `server/internal/store/store_test.go` | Cover account-scoped add/spend/reject and snapshot load. |
| Modify | `server/internal/http/account_stash.go` | Spend wallet resource for inventory upgrades and return wallet count. |
| Modify | `server/internal/http/auth_session_test.go` | Cover wallet-funded blacksmith upgrade and missing wallet rejection. |
| Modify | `server/internal/http/session.go` | Include wallet in session-start snapshot creation. |
| Modify | `server/internal/game/types.go` | Add wallet views and `resource_wallet_update` change op. |
| Modify | `server/internal/game/sim.go` | Deposit upgrade-shard pickup into wallet instead of inventory. |
| Modify | `server/internal/game/sim_load.go` | Load account resources into sim state. |
| Modify | `server/internal/game/sim_players.go` | Save/use wallet state per player. |
| Modify | `server/internal/realtime/runner.go` | Persist wallet updates for solo runner changes. |
| Modify | `server/internal/realtime/session_loop.go` | Persist wallet updates for session loop changes. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Add `resource_wallet`. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Add `resource_wallet_update`. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Include wallet example. |
| Modify | `shared/protocol/examples/state_delta.json` | Include wallet update example. |
| Modify | `client/scripts/blacksmith_panel.gd` | Display/use wallet count for upgrade resource availability. |
| Modify | `client/scripts/main.gd` | Store wallet state, apply deltas, and pass it to blacksmith panel. |
| Review | `client/scripts/net_client.gd` | Existing JSON response path already returns wallet fields; no code change needed. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Assert wallet resource counts in client scenarios. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register wallet-count assertion fields. |
| Modify | `client/tests/test_shop_panel.gd` | Cover wallet resource count and enabled state. |
| Modify | `tools/bot/state_ingest.py` | Ingest wallet snapshot/deltas. |
| Modify | `tools/bot/runtime_assertions.py` | Add protocol wallet count assertion. |
| Modify | `tools/bot/run.py` | Add helper assertion bridge if needed. |
| Modify | `tools/bot/scenarios/72_upgrade_resource_drop.json` | Assert pickup enters wallet, not bag inventory. |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Assert wallet count before/after upgrade. |
| Add | `docs/as-built/v221_resource-wallet-foundation.md` | Record proof and deferred scope. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/store/repos.go`
- [x] `server/internal/http/auth_session_test.go`
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: checked before finish.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice where feasible; keep unavoidable
  coordinator hooks to small state plumbing.

Verification:

```bash
make maintainability
```

## Task 1 - Store and HTTP Wallet

Files:
- Add: `server/migrations/0027_account_resource_wallet.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/internal/http/account_stash.go`
- Modify: `server/internal/http/auth_session_test.go`
- Modify: `server/internal/http/session.go`

- [x] Add account wallet and session-start wallet persistence.
- [x] Add account-scoped add/spend/read store methods with non-negative validation.
- [x] Consume wallet resources from inventory blacksmith upgrades and report remaining count.
- [x] Prove wallet-funded upgrade and missing-wallet rejection.

```bash
cd server && go test ./internal/store ./internal/http -run 'ResourceWallet|Upgrade' -count=1
```

## Task 2 - Protocol and Sim Wallet

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/sim_load.go`
- Modify: `server/internal/game/sim_players.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`

- [x] Add wallet view and `resource_wallet_update`.
- [x] Deposit configured upgrade resource pickups into wallet state.
- [x] Persist realtime wallet changes to the account wallet table.
- [x] Validate protocol examples.

```bash
make validate-shared
make bot scenario=upgrade_resource_drop
```

## Task 3 - Client and Bot Proof

Files:
- Modify: `client/scripts/blacksmith_panel.gd`
- Modify: `client/scripts/main.gd`
- Review: `client/scripts/net_client.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `tools/bot/state_ingest.py`
- Modify: `tools/bot/runtime_assertions.py`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/scenarios/72_upgrade_resource_drop.json`
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Ingest wallet state in Python and Godot clients.
- [x] Make blacksmith availability read wallet count.
- [x] Update resource drop and blacksmith scenarios to prove wallet increment/spend.

```bash
make client-unit
make bot-client scenario=blacksmith_upgrade_ui
```

## Task 4 - Lifecycle Docs

Files:
- Add: `docs/as-built/v221_resource-wallet-foundation.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/specs/v221_spec-resource-wallet-foundation.md`

- [x] Mark the spec complete and record focused verification.
- [x] Update current status and lifecycle row.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/store ./internal/http -run 'ResourceWallet|Upgrade' -count=1`
- [x] `make bot scenario=upgrade_resource_drop`
- [x] `make client-unit`
- [x] `make bot-client scenario=blacksmith_upgrade_ui`
- [x] `make maintainability`

Final batch `make ci` is deferred to the enclosing `$autoloop` gate.
