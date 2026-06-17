# v229 Plan - Material Auto-Pickup

Status: Complete
Goal: Auto-pick wallet-backed material loot when players walk into range.
Architecture: Generalize the existing server-owned passive pickup pass from gold to wallet-resource
loot while preserving ordinary item pickup and existing protocol shapes.
Tech stack: Go authoritative sim, realtime owner filtering, Godot client bot proof, SDD docs.

## Baseline and shortcut decision

Reuse v49's stable level/entity/player ordering and v221's `pickUpWalletResource` wallet mutation.
No client-side pickup authority is introduced. Asset/plugin decision: adopt existing code-native
events and HUD/debug assertions; reject external assets/plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/sim.go` | Include wallet resources in passive pickup scan |
| Modify | `server/internal/game/resource_wallet.go` | Make wallet pickup usable for explicit and passive winners |
| Modify/Add | `server/internal/game/gold_auto_pickup_test.go` or focused test file | Prove material auto-pickup and non-wallet regression |
| Audit | `server/internal/realtime/session_loop.go` | Confirm `resource_wallet_update` and `resource_picked_up` stay winner-private |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Prove walk-in shard pickup and HUD wallet update |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v229_material-auto-pickup.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/gold_auto_pickup_test.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: passive pickup helpers moved
  from `sim.go` into `auto_pickup.go`, lowering the `sim.go` baseline.
- [ ] Defer extraction with rationale: sim change should only route wallet resources through the
  existing passive pickup pass; any new test can live in the existing gold auto-pickup test domain
  if the file remains below target.

Verification:
```bash
make maintainability
```

## Task 1 - Server passive wallet-resource pickup

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/resource_wallet.go`
- Modify/Add: `server/internal/game/gold_auto_pickup_test.go` or focused test file

- [x] Step 1.1: Auto-pick `isWalletResourceItem` loot in the existing stable passive pickup pass,
  with the same connected/alive/same-level/range winner rules as gold.
- [x] Step 1.2: Preserve explicit click pickup, deterministic co-op winner ordering, and
  click-required behavior for non-wallet loot.
```bash
cd server && go test ./internal/game -run 'AutoPickup|ResourceWallet' -count=1
```

## Task 2 - Client bot proof

Files:
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Step 2.1: Change the blacksmith proof so the player walks into the shard and waits for
  `resource_picked_up` without sending an explicit loot click, then assert the HUD wallet panel.
```bash
make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v229_material-auto-pickup.md`

- [x] Step 3.1: Record v229 as complete with focused proof and note the final batch CI is pending.
```bash
make maintainability
```

## Final verification

- [x] `cd server && go test ./internal/game -run 'AutoPickup|ResourceWallet' -count=1`
- [x] `make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` passed after the selected v226-v232 `$autoloop` queue.
