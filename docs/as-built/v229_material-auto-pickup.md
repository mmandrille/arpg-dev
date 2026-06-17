# v229 As-Built - Material Auto-Pickup

Date: 2026-06-16

## What shipped

- Generalized the existing server-owned passive pickup scan so wallet-backed resource loot
  auto-picks when an eligible player walks into ordinary loot pickup range.
- Preserved stable level/entity/player ordering: in co-op contention, the lowest eligible connected
  alive same-level player wins the shared floor resource.
- Kept ordinary non-wallet loot click-required and preserved explicit `action_intent` wallet
  resource pickup.
- Routed passive resource pickup through winner-owned `resource_wallet_update` changes and
  `resource_picked_up` events so realtime filtering and persistence use the correct account.
- Extracted passive pickup helpers from `sim.go` into `auto_pickup.go`, lowering the `sim.go`
  line-count baseline from 6643 to 6572.
- Updated the blacksmith client-bot scenario to walk onto the upgrade shard, wait for
  `resource_picked_up`, and assert the v228 HUD wallet readout.

## Proof

```bash
cd server && go test ./internal/game -run 'AutoPickup|ResourceWallet' -count=1
make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=39_blacksmith_upgrade_ui.json
```

## Scope limits

- No new resource types, drop-rate tuning, loot table tuning, recipe changes, stash material tabs,
  market restrictions, or resource trading shipped.
- No auto-pickup for equipment, consumables, quest items, uniques, or other non-wallet loot shipped.
- No protocol schema changes or client-authoritative pickup shortcuts shipped.
