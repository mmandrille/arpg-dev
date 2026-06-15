# v50 Plan - Account Stash Storage

Status: Implemented (`make ci` green)
Goal: Add an account-wide town stash for durable item and gold storage shared by all characters on the same account.
Architecture: The Go server remains authoritative for stash visibility, item ownership, capacity checks, gold balance checks, and all mutations. Stash state is account-owned, loaded from an immutable session-start account stash snapshot for replay, and exposed only through recipient-scoped snapshots/deltas/events. Client-facing item/gold changes may be split into inventory and stash ops, but persistence must group each stash transfer into one database transaction so an item or gold amount cannot exist in both places or disappear after partial failure.
Tech stack: shared JSON rules and protocol v7, Go store/game/realtime/replay/http tests, Godot inventory/stash UI, Python protocol bot, Godot client bot, lifecycle docs.

## Baseline and shortcut decision

Baseline is v49 `gold-autopickup-and-shared-loot-rules` on `main`. Reuse:

- v10 `action_intent` range and auto-approach behavior for opening town interactables.
- v36 bag capacity and full-bag rejection rules.
- v39 character gold wallet, `gold_update`, and durable character gold persistence.
- v41/v47 server-authored town vendor panel, private shop events, and refreshed panel payload patterns.
- v33/v38/v48/v49 co-op recipient-scoped snapshot/delta routing and explicit owner-private changes.
- Existing session-start snapshots and replay reconstruction instead of reading live mutable rows during replay.
Key implementation decisions:

- Add protocol v7 because clients need explicit stash item/gold transfer intents and actor-private stash item/gold views.
- Open the stash through existing `action_intent` against a `town_stash` interactable, matching the vendor/open-interactable flow. Do not add a separate open protocol unless implementation proves the existing flow cannot express auto-approach.
- Add account-owned stash tables instead of overloading `character_item_instances.location = 'stash'`. That existing constant is character-scoped and must not become the account-wide ownership model.
- Give stash rows a `stash_item_id` row key used by stash transfer intents. Do not key account stash rows only by character `item_instance_id`, because current durable item ids are unique per character, not per account.
- Keep stash item payloads compatible with current inventory/shop item views. Withdrawal should return a server-authored `inventory_add` item view; if an implementation cannot safely preserve a prior character item id across characters, it must preserve item definition, roll payload, requirements, effects, rarity, and visible tooltip data while using the new authoritative item id returned by the server.
- Snapshot account stash items and stash gold at session creation for every member. Replay must load the frozen session-start account stash snapshot, not live account stash rows.
- Stash payloads, stash gold, and stash mutation events are private to the owning account/player in co-op.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/plans/v50_2026-06-10-account-stash-storage.md` | This implementation plan |
| Modify | `docs/specs/v50_spec-account-stash-storage.md` | Only if execution discovers approved spec corrections |
| Modify | `PROGRESS.md` | Lifecycle update when v50 ships |
| Modify | `shared/rules/interactables.v0.json` | Add `town_stash` definition |
| Modify | `shared/rules/interactables.v0.schema.json` | Allow ready stash interactables |
| Modify | `shared/rules/worlds.v0.json` | Place `town_stash` on town level `0` for `dungeon_levels` |
| Modify if needed | `shared/rules/worlds.v0.schema.json` | Only if world validation needs stash-specific constraints |
| Add | `shared/protocol/envelope.v7.schema.json` | Protocol version bump and stash intent types |
| Add | `shared/protocol/messages.v7.schema.json` | Stash deposit/withdraw item/gold intents |
| Add | `shared/protocol/session_snapshot.v7.schema.json` | Recipient-scoped stash item/gold/capacity snapshot fields |
| Add | `shared/protocol/state_delta.v7.schema.json` | Stash item/gold changes and stash events |
| Modify | `shared/protocol/examples/state_delta.json` | Stash open, deposit, withdraw, gold transfer examples |
| Modify | `tools/validate_shared.py` | Validate v7 schemas, examples, and new interactable rules |
| Add | `server/migrations/0014_account_stash.sql` | Account stash item/gold tables and session-start stash snapshots |
| Modify | `server/internal/store/models.go` | Account stash models and session-start snapshot fields |
| Modify | `server/internal/store/interfaces.go` | Account stash repository methods |
| Modify | `server/internal/store/repos.go` | Atomic stash item/gold transfers and snapshot load/save |
| Modify | `server/internal/store/store_test.go` | Account scope, transaction, session-start, and collision coverage |
| Modify | `server/internal/store/stale_sessions_test.go` | Cleanup coverage for stash session-start rows if needed |
| Modify | `server/internal/http/session.go` | Include account stash in session-start snapshot creation |
| Modify | `server/internal/http/*_test.go` | State/session visibility and same-account character coverage |
| Modify | `server/internal/game/rules.go` | Parse stash interactable/rule data if needed |
| Modify | `server/internal/game/types.go` | Stash intents, views, change ops, events, and persisted rows |
| Modify | `server/internal/game/sim.go` | Stash open, validation, mutation, capacity, and private result generation |
| Modify | `server/internal/game/game_test.go` | Stash open, transfer, rejection, co-op privacy, deterministic order tests |
| Modify | `server/internal/realtime/hub.go` | Convert account stash store rows into sim persisted rows |
| Modify | `server/internal/realtime/session_loop.go` | Load stash state, private fanout, grouped atomic persistence |
| Modify | `server/internal/realtime/session_loop_test.go` | Owner routing, persistence grouping, reconnect/state tests |
| Modify | `server/internal/replay/replay.go` | Load account stash session-start snapshot during reconstruction |
| Modify | `server/internal/replay/replay_test.go` | Stash replay, fake repo interface, event stream coverage |
| Modify | `client/scripts/main.gd` | Route stash events/changes and panel lifecycle |
| Modify | `client/scripts/inventory_panel.gd` | Reuse item grid/tooltips and expose transfer hooks if needed |
| Add | `client/scripts/stash_panel.gd` | Side stash panel, stash gold, item and gold controls |
| Add | `client/tests/test_stash_panel.gd` | Panel state, item transfer, gold transfer, debug-state tests |
| Modify | `client/tests/test_golden.gd` | Protocol/golden shape coverage if v7 data is mirrored client-side |
| Modify | `client/scripts/bot_controller.gd` | Bot-callable stash actions and debug state |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client-bot stash steps and assertions |
| Modify | `client/tests/test_client_bot.gd` | Stash client-bot action/assertion coverage |
| Modify | `tools/bot/run.py` | Protocol stash helpers, same-account character flow, replay assertions |
| Modify | `tools/bot/test_protocol.py` | Helper tests for stash assertions/intents |
| Add | `tools/bot/scenarios/36_account_stash_storage.json` | Protocol bot proof |
| Add | `tools/bot/scenarios/client/23_account_stash_panel.json` | Godot client proof |

## Task 1 - Shared Rules And Protocol V7

Files:
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/interactables.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify if needed: `shared/rules/worlds.v0.schema.json`
- Add: `shared/protocol/envelope.v7.schema.json`
- Add: `shared/protocol/messages.v7.schema.json`
- Add: `shared/protocol/session_snapshot.v7.schema.json`
- Add: `shared/protocol/state_delta.v7.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Extend `interactables.v0.schema.json` with a stash-ready shape such as `stash_id`, and update the ready interactable `oneOf` so `transition`, `shop_id`, and `stash_id` are mutually valid destination types.
```bash
make validate-shared
```

- [x] Step 1.2: Add `town_stash` to `interactables.v0.json` with `initial_state: "ready"` and a stable stash id such as `account_stash`.
```bash
make validate-shared
```

- [x] Step 1.3: Place one `town_stash` entity in `shared/rules/worlds.v0.json` on town level `0` for `world_id: "dungeon_levels"`, with a stable entity id and non-overlapping position near the existing town vendor.
```bash
make validate-shared
```

- [x] Step 1.4: Copy protocol v6 schemas to v7 and add envelope/message support for `stash_deposit_item_intent`, `stash_withdraw_item_intent`, `stash_deposit_gold_intent`, and `stash_withdraw_gold_intent`. Opening remains existing `action_intent` against the stash entity unless implementation proves otherwise.
```bash
make validate-shared
```

- [x] Step 1.5: Define stash transfer intent payloads. Item deposit uses `{ stash_entity_id, item_instance_id }`; item withdraw uses `{ stash_entity_id, stash_item_id }`; gold transfers use `{ stash_entity_id, amount }` with positive integer validation.
```bash
make validate-shared
```

- [x] Step 1.6: Extend `session_snapshot.v7` with recipient-scoped `stash_items`, `stash_gold`, and `stash_capacity` fields. `stash_items[]` must include `stash_item_id` plus item view payload fields compatible with inventory/shop tooltip rendering.
```bash
make validate-shared
```

- [x] Step 1.7: Extend `state_delta.v7` with actor-private stash changes and events: `stash_item_add`, `stash_item_remove`, `stash_gold_update`, `stash_opened`, `stash_item_deposited`, `stash_item_withdrawn`, `stash_gold_deposited`, and `stash_gold_withdrawn`. Keep existing `inventory_add`, `inventory_remove`, `gold_update`, and `character_progression_update` for the character side of each transfer.
```bash
make validate-shared
```

- [ ] Step 1.8: Add rejection/ack reason coverage for stash-specific failures: `not_in_town`, `out_of_range`, `stash_not_open_or_reachable`, `item_not_owned`, `item_not_in_bag`, `item_equipped`, `item_hotbar_assigned`, `stash_full`, `stash_item_not_found`, `inventory_full`, `invalid_amount`, `insufficient_character_gold`, and `insufficient_stash_gold`.
```bash
make validate-shared
```

- [x] Step 1.9: Update protocol examples with stash open, item deposit, item withdraw, gold deposit, and gold withdraw payloads, each showing only actor-private stash data.
```bash
make validate-shared
```

- [x] Step 1.10: Update `tools/validate_shared.py` so v7 is the current schema set for session snapshots, state deltas, messages, examples, and current protocol file presence checks.
```bash
make validate-shared
```

## Task 2 - Account Stash Persistence And Session-Start Snapshots

Files:
- Add: `server/migrations/0014_account_stash.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`
- Modify: `server/internal/store/stale_sessions_test.go`

- [x] Step 2.1: Add `account_stash_items` with `account_id`, `stash_item_id`, source/last character metadata if useful, item payload fields (`item_def_id`, slot/category as currently persisted, equipped false, rolled stats), timestamps, and a primary key that is account-scoped by `stash_item_id`.
```bash
cd server && go test ./internal/store/... -run TestMigrations -count=1
```

- [x] Step 2.2: Add `account_stash_gold` with `account_id` primary key, non-negative `gold`, and timestamps.
```bash
cd server && go test ./internal/store/... -run 'TestMigrations|AccountStashGold' -count=1
```

- [x] Step 2.3: Add `session_start_account_stash_items` and `session_start_account_stash_gold` keyed by `session_id` plus `account_id`, so replay freezes account stash item rows and stash gold per member at session creation.
```bash
cd server && go test ./internal/store/... -run 'SessionStart|AccountStash' -count=1
```

- [x] Step 2.4: Add store models such as `AccountStashItem`, `AccountStashSnapshot`, and `AccountStashGold`, and extend `SessionStartSnapshot` with account stash fields.
```bash
cd server && go test ./internal/store/... -run 'AccountStash|SessionStart' -count=1
```

- [x] Step 2.5: Add repository methods to list account stash items, get/create account stash gold, write session-start account stash snapshots, and load session-start account stash snapshots for a member/session.
```bash
cd server && go test ./internal/store/... -run 'AccountStash|SessionStart' -count=1
```

- [x] Step 2.6: Add atomic item transfer methods for character bag -> account stash and account stash -> character bag. These methods must validate account ownership, lock affected rows, remove from the source location, insert into the destination location, and reject without partial mutation on missing rows, capacity conflicts, or item id collisions.
```bash
cd server && go test ./internal/store/... -run 'AccountStash.*Item|CharacterItem' -count=1
```

- [x] Step 2.7: Add atomic gold transfer methods for character wallet -> stash gold and stash gold -> character wallet. These methods must lock both balances, validate positive amounts and sufficient funds, and commit both balance updates together.
```bash
cd server && go test ./internal/store/... -run 'AccountStash.*Gold|CharacterProgression' -count=1
```

- [x] Step 2.8: Add store tests proving two characters on the same account see the same account stash rows and stash gold, while a different account cannot list or mutate them.
```bash
cd server && go test ./internal/store/... -run 'AccountStash.*Account|AccountStash.*Privacy' -count=1
```

- [ ] Step 2.9: Add store tests proving the account stash does not depend on `character_item_instances.location = 'stash'`, and that two same-account characters with colliding character-local item ids can both deposit without overwriting each other.
```bash
cd server && go test ./internal/store/... -run 'AccountStash.*Collision|AccountStash.*Location' -count=1
```

- [x] Step 2.10: Update stale-session cleanup tests if session-start account stash rows need explicit cleanup when stale sessions are deleted.
```bash
cd server && go test ./internal/store/... -run 'Stale|SessionStart|AccountStash' -count=1
```

## Task 3 - Session Creation, Sim State, And Stash Mutations

Files:
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/http/*_test.go`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Update session creation and join flows to load live account stash rows/gold and pass them into `CreateSessionStartSnapshot` for the actor account alongside character items, waypoints, hotbar, shop stock, and progression.
```bash
cd server && go test ./internal/http/... -run 'CreateSession|Join|SessionStart|AccountStash' -count=1
```

- [x] Step 3.2: Add game types for stash persisted rows, `StashItemView`, stash snapshot fields, stash transfer intents, stash change ops, and stash events. Include internal owner/private fields without leaking account ids through protocol payloads.
```bash
cd server && go test ./internal/game/... -run 'Test.*Snapshot|Test.*Protocol' -count=1
```

- [x] Step 3.3: Add sim-owned stash state loaded per player/account from session-start rows. This state should include item rows, stash gold, default capacity `50`, and deterministic ordering by stash item id.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Load|Test.*Snapshot' -count=1
```

- [x] Step 3.4: Add `town_stash` interactable handling in `activateInteractable`. Alive, connected, in-town, in-range actors should receive an actor-private `stash_opened` view; out-of-range actors should use the existing auto-approach/retry behavior where possible.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Open|Test.*Interactable' -count=1
```

- [x] Step 3.5: Implement item deposit validation in the sim: actor owns the item in the current character bag, actor is in town and can interact with the stash, item is not equipped, item is not hotbar-assigned, and stash capacity has a free slot.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Deposit|Test.*Hotbar|Test.*Equipped' -count=1
```

- [x] Step 3.6: Implement successful item deposit changes/events: actor-private `inventory_remove`, actor-private `stash_item_add`, `stash_item_deposited`, ack, and refreshed stash state. Generate or assign stash row identity deterministically and never allow duplicate ownership in sim state.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Deposit|Test.*Inventory' -count=1
```

- [x] Step 3.7: Implement item withdraw validation in the sim: stash row exists for the actor account, actor is in town and can interact with stash, and current character bag capacity has space.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Withdraw|Test.*InventoryCapacity' -count=1
```

- [x] Step 3.8: Implement successful item withdrawal changes/events: actor-private `stash_item_remove`, actor-private `inventory_add`, `stash_item_withdrawn`, ack, and refreshed inventory/stash state.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Withdraw|Test.*Inventory' -count=1
```

- [x] Step 3.9: Implement gold deposit and withdraw validation in the sim: positive amount, actor in town and can interact with stash, sufficient character gold for deposit, and sufficient stash gold for withdraw.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Gold|Test.*Gold' -count=1
```

- [x] Step 3.10: Implement successful gold transfer changes/events: actor-private `stash_gold_update`, existing actor-private `gold_update`, existing `character_progression_update`, stash gold transfer event, and ack.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Gold|Test.*CharacterProgression' -count=1
```

- [ ] Step 3.11: Add sim tests for invalid amount, insufficient wallet gold, insufficient stash gold, full stash deposit, full bag withdraw, equipped deposit reject, hotbar-assigned deposit reject, dead/disconnected reject, non-town reject, and out-of-range behavior.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Reject|Test.*Stash.*Gold' -count=1
```

- [ ] Step 3.12: Add co-op sim tests proving another account sees the public `town_stash` entity but receives no stash payloads, stash gold, stash item changes, or stash mutation events from the actor.
```bash
cd server && go test ./internal/game/... -run 'Test.*Stash.*Coop|Test.*Privacy' -count=1
```

## Task 4 - Realtime Persistence, Private Fanout, And Replay

Files:
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/realtime/session_loop.go`
- Modify: `server/internal/realtime/session_loop_test.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/http/*_test.go`

- [x] Step 4.1: Update realtime session build and late-join paths to load each member's account stash session-start snapshot into that member's sim player state.
```bash
cd server && go test ./internal/realtime/... -run 'Test.*Session.*Snapshot|Test.*Join|Test.*Stash' -count=1
```

- [x] Step 4.2: Extend `SnapshotForPlayer`, snapshot envelopes, and `/state` responses so each recipient gets only their own `stash_items`, `stash_gold`, and `stash_capacity`.
```bash
cd server && go test ./internal/realtime/... ./internal/http/... -run 'Test.*Snapshot|Test.*State|Test.*Stash' -count=1
```

- [x] Step 4.3: Update change filtering so `stash_item_add`, `stash_item_remove`, `stash_gold_update`, and character-side changes produced by stash transfer route only to the owner player.
```bash
cd server && go test ./internal/realtime/... -run 'Test.*Stash.*Fanout|Test.*Private' -count=1
```

- [x] Step 4.4: Update event filtering so `stash_opened`, `stash_item_deposited`, `stash_item_withdrawn`, `stash_gold_deposited`, and `stash_gold_withdrawn` are actor-private and never delivered to other co-op accounts.
```bash
cd server && go test ./internal/realtime/... -run 'Test.*Stash.*Event|Test.*Private' -count=1
```

- [x] Step 4.5: Update `persistChanges` to group each stash item transfer and call the new store atomic item transfer method once. It must not persist the paired `inventory_add`/`inventory_remove` generically in a way that can create partial or duplicate ownership.
```bash
cd server && go test ./internal/realtime/... -run 'Test.*Stash.*Persist|Test.*Inventory.*Persist' -count=1
```

- [x] Step 4.6: Update `persistChanges` to group each stash gold transfer and call the new store atomic gold transfer method once. It must not persist `gold_update` and `stash_gold_update` as independent commits for the same transfer.
```bash
cd server && go test ./internal/realtime/... -run 'Test.*Stash.*Gold.*Persist|Test.*Gold.*Persist' -count=1
```

- [ ] Step 4.7: Add realtime tests proving reconnect and fresh `/state` show account stash item/gold changes for the owner after successful mutation.
```bash
cd server && go test ./internal/realtime/... ./internal/http/... -run 'Test.*Stash.*Reconnect|Test.*Stash.*State' -count=1
```

- [ ] Step 4.8: Add same-account different-character HTTP/realtime coverage: deposit item/gold with character A, create or select character B on the same account, start a fresh session, and verify B sees and can withdraw the same account stash rows.
```bash
cd server && go test ./internal/http/... ./internal/realtime/... -run 'Test.*Stash.*Character|Test.*AccountWide' -count=1
```

- [x] Step 4.9: Update replay reconstruction to load account stash from session-start snapshots for host and co-op members, then derive stash mutations from ordered inputs.
```bash
cd server && go test ./internal/replay/... -run 'Test.*Stash|Test.*SessionStart' -count=1
```

- [ ] Step 4.10: Add replay tests proving a historical session reconstructs the same stash item rows, stash gold, character inventory, character gold, and private event stream even if live account stash rows changed after the session started.
```bash
cd server && go test ./internal/replay/... -run 'Test.*Stash.*Replay|Test.*Stash.*Live' -count=1
```

## Task 5 - Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Add: `tools/bot/scenarios/36_account_stash_storage.json`

- [x] Step 5.1: Add protocol bot helpers to find/open `town_stash`, wait for `stash_opened`, send stash item/gold transfer intents, and assert stash item/gold deltas.
```bash
pytest tools/bot/test_protocol.py -q
```

- [ ] Step 5.2: Add helper support for same-account character flow: create/select a second character or use existing character endpoints, start a fresh session, and assert account stash continuity.
```bash
pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.3: Add `36_account_stash_storage.json` using `world_id: "dungeon_levels"` and stable seed `v50-account-stash-storage`.
```bash
make bot scenario=36_account_stash_storage.json
```

- [x] Step 5.4: Drive the scenario to acquire a real dungeon item and gold, return to town, open the stash, deposit one bag item, and deposit a positive gold amount.
```bash
make bot scenario=36_account_stash_storage.json
```

- [x] Step 5.5: Verify a fresh session on the same account sees the stored stash item and stash gold before any new pickup or vendor action.
```bash
make bot scenario=36_account_stash_storage.json
```

- [ ] Step 5.6: Verify another character on the same account sees the same stash item/gold and can withdraw item and gold into that character's bag/wallet.
```bash
make bot scenario=36_account_stash_storage.json
```

- [ ] Step 5.7: Fill the active character bag, attempt stash item withdrawal, and assert `inventory_full` rejection without removing the stash item.
```bash
make bot scenario=36_account_stash_storage.json
```

- [ ] Step 5.8: Add privacy proof in a co-op phase or companion connection: another account sees the public stash entity but not the owner stash payloads, stash gold, or mutation events.
```bash
make bot scenario=36_account_stash_storage.json
```

- [x] Step 5.9: Verify reconnect, `/state`, and replay for the scenario.
```bash
make bot scenario=36_account_stash_storage.json
```

- [x] Step 5.10: Keep v49 and vendor regressions green because stash touches loot, inventory, gold, and town UI flows.
```bash
make bot scenario=35_gold_autopickup_shared_loot.json
make bot scenario=33_shop_stock_lifecycle.json
```

## Task 6 - Godot Client Stash UI

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/inventory_panel.gd`
- Add: `client/scripts/stash_panel.gd`
- Add: `client/tests/test_stash_panel.gd`
- Modify: `client/tests/test_golden.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/23_account_stash_panel.json`

- [x] Step 6.1: Implement `StashPanel` as a side panel that reuses current item row/grid and tooltip patterns, displays stash capacity, stash gold, and current character gold, and never owns transfer outcomes locally.
```bash
make client-unit
```

- [x] Step 6.2: Route `stash_opened`, `stash_item_add`, `stash_item_remove`, `stash_gold_update`, `inventory_add`, `inventory_remove`, `gold_update`, and `character_progression_update` through `main.gd` so inventory, stash, and HUD balances stay synchronized.
```bash
make client-unit
```

- [x] Step 6.3: Add item deposit/withdraw UI gestures. Deposit sends `stash_deposit_item_intent` for a bag item; withdraw sends `stash_withdraw_item_intent` for a stash row id. Equipped and hotbar-assigned items should show server rejection without optimistic mutation.
```bash
make client-unit
```

- [ ] Step 6.4: Add explicit numeric gold deposit/withdraw controls with validation for positive integer input, while still relying on server rejection for authoritative balance checks.
```bash
make client-unit
```

- [x] Step 6.5: Add debug-state and bot hooks for opening stash, depositing/withdrawing a specific item, depositing/withdrawing gold, and reading visible stash/inventory/gold state.
```bash
make client-unit
```

- [x] Step 6.6: Add `test_stash_panel.gd` coverage for open state, item row rendering, item transfer signals, gold transfer signals, balance updates, capacity text, and rejection display.
```bash
make client-unit
```

- [x] Step 6.7: Add client bot scenario `23_account_stash_panel.json` proving Godot opens stash in town, item panels synchronize after deposit/withdraw, and character/stash gold balances update after deposit/withdraw.
```bash
make bot-client scenario=23_account_stash_panel.json HEADLESS=1
```

- [x] Step 6.8: Run existing client bot regressions for inventory, vendor, shop stock, and gold auto-pickup.
```bash
make bot-client HEADLESS=1
```

## Task 7 - Lifecycle Docs And CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v50_2026-06-10-account-stash-storage.md`

- [x] Step 7.1: Add v50 to the slice lifecycle table when implementation finishes.
```bash
make ci
```

- [x] Step 7.2: Add a concise v50 summary under "What each slice proved", including account-wide item storage, stash gold, privacy, and replay-safe session-start stash snapshots.
```bash
make ci
```

- [x] Step 7.3: Add `36_account_stash_storage.json` and `23_account_stash_panel.json` to the scripted scenario catalog.
```bash
make ci
```

- [x] Step 7.4: Record any deferred stash scope in Open gaps, especially sorting/filtering/tabs/capacity upgrades/crafting/market delivery if still deferred.
```bash
make ci
```

- [x] Step 7.5: Keep this plan's checkboxes accurate during execution and update the status line to implemented only after CI is green.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/store/...`
- [x] `cd server && go test ./internal/game/...`
- [x] `cd server && go test ./internal/realtime/...`
- [x] `cd server && go test ./internal/replay/...`
- [x] `cd server && go test ./internal/http/...`
- [x] `make client-unit`
- [x] `make bot scenario=36_account_stash_storage.json`
- [x] `make bot scenario=35_gold_autopickup_shared_loot.json`
- [x] `make bot scenario=33_shop_stock_lifecycle.json`
- [x] `make bot-client scenario=23_account_stash_panel.json HEADLESS=1`
- [x] `make bot`
- [x] `make bot-client HEADLESS=1`
- [x] `make ci`

## Deferred scope

- No account-wide equipment, stats, skills, waypoints, shop stock, or progression.
- No remote stash access outside town level `0`.
- No stash sorting, filtering, search, tabs, item stacks, multi-cell footprints, or capacity upgrades.
- No selling, buying, equipping, using, upgrading, or hotbar assignment directly from stash.
- No player market, trade delivery, mystery-seller overflow, crafting, upgrades, resource currencies, or material tabs.
- No real-time cross-session push for already-open sessions on the same account. Store transactions still must prevent invalid duplicate withdraw/deposit outcomes.
- No production stash art, NPC dialog, audio, stash animation, custom icons, or imported asset pack.
- Explicit same-account second-character protocol flow, co-op privacy bot proof, full-bag bot proof, and arbitrary numeric gold entry remain deferred. v50 covers the mechanics through store/sim tests, private realtime filtering, fixed one-gold UI controls, and protocol/client bot persistence scenarios.
