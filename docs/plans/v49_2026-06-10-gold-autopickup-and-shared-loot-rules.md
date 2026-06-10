# v49 Plan — Gold Auto-Pickup And Shared Floor Loot

Status: Implemented; `make ci` green on 2026-06-10
Goal: Keep floor loot shared while making gold pick up automatically when an eligible player moves into pickup range.
Architecture: The Go sim remains authoritative for loot visibility, gold pickup eligibility, winner selection, wallet mutation, and replay. Item drops stay single shared floor entities with no owner semantics. Gold auto-pickup runs after player movement resolves, uses stable level/entity/player ordering, and emits existing protocol v6 changes/events without a schema bump unless implementation proves otherwise. Private gold/progression changes route by explicit owner when the tick result actor is not the pickup winner.
Tech stack: Go sim and realtime tests, existing protocol v6 JSON, Python protocol bot scenario, replay verification, lifecycle docs.

## Baseline and shortcut decision

Baseline is v48 `coop-rewards-and-scaling` on `main`. Reuse:

- v10 `action_intent` pickup and range behavior for explicit item/gold pickup.
- v11 server-owned movement and auto-navigation.
- v25/v30 shared loot table and rarity/depth drop behavior.
- v39 gold wallet, `gold_picked_up`, `gold_update`, and durable character gold persistence.
- v33/v38/v48 co-op membership, recipient-scoped snapshots/deltas, explicit owner routing, and replay proof patterns.

Godot plugin shortcut decision: **reject / not applicable**. v49 has no new client UI, camera, art, inventory presentation, or placeholder asset work. The existing Godot client should observe existing `entity_remove`, `gold_update`, `character_progression_update`, and `gold_picked_up` deltas. If implementation unexpectedly touches `client/`, rerun the plugin adoption checklist from `docs/researchs/godot-plugins-and-shortcuts.md` before editing client code.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Created | `docs/specs/v49_spec-gold-autopickup-and-shared-loot-rules.md` | Slice spec |
| Create | `docs/plans/v49_2026-06-10-gold-autopickup-and-shared-loot-rules.md` | This implementation plan |
| Modify | `server/internal/game/sim.go` | Auto-pickup pass, shared gold pickup helper, deterministic ordering |
| Modify | `server/internal/game/game_test.go` | Sim tests for auto-pickup, shared loot, ordering, co-op contention |
| Modify | `server/internal/realtime/session_loop.go` | Winner-private gold event filtering if not already sufficient |
| Modify | `server/internal/realtime/session_loop_test.go` | Fanout and persistence tests for auto-pickup winner routing |
| Modify | `server/internal/replay/replay_test.go` | Replay proof for passive gold pickup |
| Audit | `shared/protocol/state_delta.v6.schema.json` | Confirm existing change/event shapes cover auto-pickup |
| Audit | `shared/protocol/session_snapshot.v6.schema.json` | Confirm no new loot ownership field is introduced |
| Modify if needed | `shared/protocol/examples/state_delta.json` | Only if examples need no-correlation auto-pickup coverage |
| Modify | `tools/bot/run.py` | Gold wait helpers and/or dedicated co-op scenario driver |
| Modify if needed | `tools/bot/test_protocol.py` | Bot helper coverage if new helpers are added |
| Create | `tools/bot/scenarios/35_gold_autopickup_shared_loot.json` | Protocol bot proof |
| Modify | `tools/bot/scenarios/client/15_town_vendor_shop_panel.json` | Client bot fixture updated for passive gold pickup |
| Modify | `tools/bot/scenarios/client/16_vendor_item_comparison.json` | Client bot fixture updated for passive gold pickup |
| Modify | `tools/bot/scenarios/client/22_shop_stock_lifecycle.json` | Client bot fixture updated for passive gold pickup |
| Modify | `PROGRESS.md` | Lifecycle update when v49 ships |

## Task 1 — Server Gold Pickup Semantics

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 1.1: Refactor the existing gold branch in `pickUpTarget` into a helper that can pick up one gold loot entity for an explicit `playerID`, optional `correlation_id`, and optional intent ack while preserving existing manual pickup behavior.
- [x] Step 1.2: Ensure the helper updates the winner's player-scoped `gold` and `progression.Gold`, emits public `entity_remove`, and emits winner-owned `gold_update` plus `character_progression_update`.
- [x] Step 1.3: Add a post-movement auto-pickup pass in `TickResults` after all connected players run `applyMovement` and before monster movement, boss phases, monster attacks, and projectiles advance.
- [x] Step 1.4: Scan active levels in sorted level order, gold entities in stable entity-id order, and eligible players in sorted player-id order; choose the lowest eligible player id as the winner.
- [x] Step 1.5: Eligibility must require connected, alive, same-level, and within the same ordinary loot pickup range; disconnected/dead players must not consume gold.
- [x] Step 1.6: Keep non-gold loot unchanged: no owner field, no auto-pickup, explicit `action_intent` required, existing inventory capacity checks preserved.
- [x] Step 1.7: Preserve explicit gold `action_intent` compatibility. A manual click in range still works; an out-of-range click may auto-nav and then be consumed by the passive pass without duplicating gold.

```bash
cd server && go test ./internal/game/... -run 'TestGold|TestDungeonDescendAscendTransitions|TestCoop'
```

## Task 2 — Focused Sim Coverage

Files:
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add a solo test proving walking/moving into range of town gold picks it up without `action_intent`.
- [x] Step 2.2: Add a dungeon-level test proving generated or placed dungeon gold auto-picks and persists in the same wallet path as manual pickup.
- [x] Step 2.3: Add a test proving standing in range of a non-gold loot entity does not auto-pick it and still requires explicit click.
- [x] Step 2.4: Add a manual gold pickup regression test for in-range explicit `action_intent`.
- [x] Step 2.5: Add co-op contention coverage: two connected same-level players in range of one gold entity on the same tick results in the lowest player id winning.
- [x] Step 2.6: Add dead/disconnected exclusion coverage.
- [x] Step 2.7: Add multiple-gold ordering coverage: stable entity-id processing and no map-order-dependent outcome.
- [x] Step 2.8: Add pending auto-navigation coverage: clicking out-of-range gold does not duplicate or resurrect the entity if passive pickup consumes it on arrival.

```bash
cd server && go test ./internal/game/... -run 'Test.*Gold|Test.*Loot|Test.*Coop'
```

## Task 3 — Realtime Routing, Persistence, And Replay

Files:
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/realtime/session_loop_test.go`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 3.1: Audit `filterChangesForClient` for auto-pickup results. `gold_update` and `character_progression_update` must be winner-private using `Change.OwnerPlayerID` when `ActorPlayerID` is `0` or another player.
- [x] Step 3.2: Update `filterEventsForClient` so `gold_picked_up` is private to the event's `entity_id` player, preventing non-winners from receiving another player's `total_gold`.
- [x] Step 3.3: Add realtime fanout tests proving same-level non-winners receive public `entity_remove` but not private wallet/progression changes or `gold_picked_up`.
- [x] Step 3.4: Add persistence tests proving passive gold pickup writes to the winner's account/character when the tick result actor is absent or different.
- [x] Step 3.5: Add replay coverage that advances movement into gold without a pickup input and verifies the same winner, entity removal, wallet, and derived event stream reconstruct.

```bash
cd server && go test ./internal/realtime/... ./internal/replay/...
```

## Task 4 — Protocol Shape Audit

Files:
- Audit: `shared/protocol/state_delta.v6.schema.json`
- Audit: `shared/protocol/session_snapshot.v6.schema.json`
- Modify if needed: `shared/protocol/examples/state_delta.json`

- [x] Step 4.1: Confirm no schema change is required: existing `entity_remove`, `gold_update`, `character_progression_update`, and `gold_picked_up` represent passive pickup.
- [x] Step 4.2: Confirm `gold_picked_up.correlation_id` remains optional so passive pickup can omit it.
- [x] Step 4.3: Confirm loot entity `owner_id` remains unused for shared floor loot.
- [x] Step 4.4: Update protocol examples only if validation or docs need an explicit no-correlation auto-pickup example.

```bash
make validate-shared
```

## Task 5 — Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Modify if needed: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/35_gold_autopickup_shared_loot.json`

- [x] Step 5.1: Add or adjust bot helper behavior so gold pickup can be observed by walking into range and waiting for `gold_picked_up` / wallet change without sending `action_intent`.
- [x] Step 5.2: Keep item pickup assertions strict: non-gold loot must remain on the floor until an explicit `action_intent` succeeds.
- [x] Step 5.3: Add a new scenario `35_gold_autopickup_shared_loot.json` with `world_id: "dungeon_levels"` and stable seed `v49-gold-autopickup-shared-loot`.
- [x] Step 5.4: Drive the scenario through a deterministic drop source that produces both shared floor gold and non-gold loot visible to the actor.
- [x] Step 5.5: Prove walking near gold without clicking increments gold and removes the floor entity.
- [x] Step 5.6: Prove walking near non-gold loot does not add inventory, then explicit click picks the item up.
- [x] Step 5.7: Add a co-op phase or dedicated driver proving two peers see the same floor gold and the lowest-player-id winner receives the wallet update when both are in range.
- [x] Step 5.8: Verify `/state`, reconnect or fresh-session persistence, and replay for the new scenario.
- [x] Step 5.9: Keep v48 co-op rewards/scaling scenario green as regression coverage.

```bash
make bot scenario=35_gold_autopickup_shared_loot.json
make bot scenario=34_coop_rewards_and_scaling.json
```

## Task 6 — Lifecycle Docs And CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v49_2026-06-10-gold-autopickup-and-shared-loot-rules.md`

- [x] Step 6.1: Add v49 to the slice numbering note and lifecycle table when implementation finishes.
- [x] Step 6.2: Add a concise v49 summary under "What each slice proved."
- [x] Step 6.3: Add `gold_autopickup_shared_loot` to the scripted scenario catalog.
- [x] Step 6.4: Move any new deferred scope to Open gaps, keeping shared-loot and no-personal-loot decisions explicit if useful.
- [x] Step 6.5: Keep this plan's checkboxes accurate during execution.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/...`
- [x] `cd server && go test ./internal/realtime/...`
- [x] `cd server && go test ./internal/replay/...`
- [x] `make bot scenario=35_gold_autopickup_shared_loot.json`
- [x] `make bot scenario=34_coop_rewards_and_scaling.json`
- [x] `make bot`
- [x] `make bot-client HEADLESS=1`
- [x] `make ci`

## Deferred scope

- No personal loot, hidden loot, duplicated per-player drops, loot reservations, or loot allocation UI.
- No shared or split gold. Gold is still one shared floor entity; first valid pickup wins.
- No item auto-pickup.
- No drop-rate, treasure-class, rarity, depth-band, chest, or monster-rarity rebalance.
- No client UI/art/audio changes unless implementation discovers a regression that cannot be covered through existing deltas.
