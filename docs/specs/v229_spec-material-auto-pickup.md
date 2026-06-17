# v229 Spec: Material Auto-Pickup

Status: Complete
Date: 2026-06-16
Codename: material-auto-pickup

## Purpose

Make wallet-backed upgrade materials pick up automatically when an eligible player walks into loot
range, matching gold's quality-of-life behavior while preserving ordinary shared item loot.

## Baseline

Builds on v49 gold auto-pickup, v221 resource wallet foundation, and v228 wallet HUD display. The
server already owns upgrade-shard pickup, deposits it into `resource_wallet`, emits
`resource_picked_up`, and persists wallet increments. ADR alignment: ADR-0001 keeps pickup
authoritative on the Go server; ADR-0012 and ADR-0014 support upgrade materials as valuable
wallet-backed resources without adding a new material type.

Asset/plugin decision: adopt existing server rules, protocol events, and code-native client/bot
assertions; reject external assets or plugins.

## Non-goals

- No new material/resource types, drop-rate tuning, loot table tuning, item upgrade recipe changes,
  stash material tabs, resource trading, or market restrictions.
- No auto-pickup for equipment, potions, quest items, uniques, or ordinary non-wallet loot.
- No client-authoritative pickup shortcut and no protocol schema change unless implementation proves
  the existing `resource_wallet_update` / `resource_picked_up` shapes are insufficient.

## Acceptance Criteria

- Wallet resource loot auto-picks when an alive connected player is on the same level and within
  the same pickup range used by gold/manual loot pickup.
- The pickup removes the shared floor entity, increments the winner's wallet, emits
  `resource_picked_up`, and sends a private `resource_wallet_update`.
- Non-wallet loot still requires explicit `action_intent`.
- Co-op contention is deterministic and follows the existing lowest-eligible-player ordering used
  by gold auto-pickup.
- A client-bot scenario proves walking into an upgrade shard picks it up without clicking and the
  v228 HUD wallet readout updates.

## Scope and Likely Files

- `server/internal/game/sim.go`
- `server/internal/game/resource_wallet.go`
- `server/internal/game/gold_auto_pickup_test.go` or a focused resource wallet pickup test
- `server/internal/realtime/session_loop.go` / tests only if owner routing requires adjustment
- `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` or a new focused client scenario
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v229_material-auto-pickup.md`

## Test and Bot Proof

- `cd server && go test ./internal/game -run 'AutoPickup|ResourceWallet' -count=1`
- `make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

- `make bot-visual scenario=39_blacksmith_upgrade_ui.json`

## Open Questions and Risks

- No blocking questions. The conservative behavior is to auto-pick only `isWalletResourceItem`
  loot, reuse the same eligibility/range ordering as gold, and keep every other item click-required.
