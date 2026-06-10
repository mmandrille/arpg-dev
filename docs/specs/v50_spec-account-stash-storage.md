# Spec: `account-stash-storage`

Status: Draft
Date: 2026-06-10
Branch: `main`
Codename: `account-stash-storage`
Slice: v50 - account-wide town stash storage
Baseline: v49 `gold-autopickup-and-shared-loot-rules`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared contracts, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - town hub, account/character persistence, co-op level scoping
- [`../adr/0011-player-market-and-multi-item-trade-offers.md`](../adr/0011-player-market-and-multi-item-trade-offers.md) - future market delivery needs town stash ownership rules
- [`../adr/0012-item-upgrades-and-item-levels.md`](../adr/0012-item-upgrades-and-item-levels.md) - future resources/upgrades need durable storage rules
- [`../adr/0013-mystery-seller-and-unidentified-item-offers.md`](../adr/0013-mystery-seller-and-unidentified-item-offers.md) - future mystery purchases may deliver to inventory or stash
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - inventory/stash UI shortcut checklist
- [`v36_spec-inventory-paper-doll-capacity.md`](v36_spec-inventory-paper-doll-capacity.md) - bag capacity and full-bag rejection
- [`v39_spec-ui-currency-and-mana-polish.md`](v39_spec-ui-currency-and-mana-polish.md) - character gold wallet and gold updates
- [`v41_spec-town-vendor-gold-sink.md`](v41_spec-town-vendor-gold-sink.md) - town vendor, buy/sell, durable gold mutation
- [`v47_spec-shop-stock-lifecycle.md`](v47_spec-shop-stock-lifecycle.md) - server-owned shop inventory mutations and client panel refresh
- [`v49_spec-gold-autopickup-and-shared-loot-rules.md`](v49_spec-gold-autopickup-and-shared-loot-rules.md) - shared floor loot and gold pickup rules

## 1. Purpose

Inventory capacity is now meaningful, gold is durable, and the town has a vendor, but players still
have no durable off-character storage. This slice adds an account-wide town stash that preserves
items and gold outside the active character's bag while keeping all ownership transfers
server-authoritative.

The account stash is available from a town interactable on level `0`:

- Any character on the same account can see the same stash contents and stash gold.
- Bag items can be deposited into the stash when they are owned by the actor and are not equipped or
  hotbar-assigned.
- Stash items can be withdrawn into the actor's bag when bag capacity allows.
- Character gold can be deposited into account stash gold.
- Account stash gold can be withdrawn into the current character wallet.
- Stash state persists across fresh sessions and across different characters on the same account.

This is an ownership/storage slice, not a market or crafting slice. The stash establishes the
durable account-level storage surface that future market delivery, mystery-seller overflow, and
upgrade resources can build on.

Client shortcut decision for the spec: **borrow UI patterns, do not adopt gameplay logic**. The
implementation plan must record the adoption checklist from the plugin research doc. GLoot,
Godot-Inventory, or Wyvernbox may be used as UI references for side-by-side inventory/stash
gestures, but item state, capacity, gold, validation, and mutation authority must remain in the Go
server.

## 2. Non-goals

- No account-wide equipment, stats, skill ranks, waypoints, shop stock, or progression. Only stash
  items and stash gold are account-wide in this slice.
- No player market listings, trade offers, item locks/reservations, audit log UI, expiration jobs,
  delivery inbox, or received-trade flow.
- No mystery seller, unidentified items, stash overflow from purchases, or blind-buy stock.
- No crafting, upgrades, resource currencies, material tabs, item level mutation, or item binding.
- No selling, buying, equipping, using, upgrading, or hotbar assignment directly from stash.
- No deposit of equipped items, hotbar-assigned items, floor loot, shop offers, generated offers, or
  buyback rows.
- No stash sorting, filtering, search, tabs, item stacks, multi-cell footprints, or capacity
  upgrades.
- No remote stash access outside town level `0`.
- No production stash art, NPC dialog, audio, stash animation, custom icons, or imported asset pack.
- No real-time cross-session push. If the same account has another already-open session, that
  session may refresh stash state the next time it opens the stash; atomic persistence still must
  prevent invalid double-withdraw/deposit outcomes.
- No backward-compatibility promise for stale protocol schemas beyond coordinated current-dev
  schema, fixture, bot, and client updates.

## 3. Acceptance Criteria

1. Town level `0` contains a `town_stash` interactable visible through `/state`, snapshots, and the
   Godot client.
2. The stash can only be opened by an alive connected player on town level `0` within normal
   interactable range or through existing auto-approach behavior.
3. Opening the stash emits an actor-private `stash_opened` event or equivalent actor-scoped payload
   containing:
   - current account stash item rows,
   - current account stash gold,
   - configured stash item capacity.
4. Account stash item capacity defaults to `50` item slots.
5. Account stash gold is a non-negative integer account balance and has no item-slot capacity cost.
6. Stash contents are account-scoped. Two different characters on the same account see the same
   stash items and stash gold in fresh sessions.
7. Stash contents are private to the owning account. Other co-op accounts cannot see stash items,
   stash gold, or stash mutation events.
8. Depositing a bag item validates all of the following before mutation:
   - actor owns the item through the current character inventory,
   - actor is in town and has an open or reachable town stash interaction,
   - item is in the bag, not equipped, and not on any hotbar slot,
   - account stash item capacity has a free slot.
9. Successful item deposit atomically removes the item from the character bag, persists it in the
   account stash with its full item payload, emits actor-private `inventory_remove` and stash-add
   changes, and refreshes the visible stash panel if open.
10. Failed item deposit emits a rejection/ack reason and does not mutate character inventory or
    account stash.
11. Withdrawing a stash item validates all of the following before mutation:
    - item exists in the actor account stash,
    - actor is in town and has an open or reachable town stash interaction,
    - actor's bag has free capacity.
12. Successful item withdrawal atomically removes the item from account stash, adds it to the
    current character bag with the same item identity/payload, emits actor-private stash-remove and
    `inventory_add` changes, and refreshes the visible inventory/stash panels if open.
13. Full-bag withdrawal rejects without removing the stash item.
14. Depositing gold validates positive amount and sufficient current character gold before
    mutation.
15. Successful gold deposit atomically subtracts from the current character wallet, adds to account
    stash gold, persists both balances, and emits actor-private character `gold_update`,
    `character_progression_update`, and stash-gold update changes/events.
16. Withdrawing gold validates positive amount and sufficient account stash gold before mutation.
17. Successful gold withdrawal atomically subtracts from account stash gold, adds to the current
    character wallet, persists both balances, and emits actor-private stash-gold update plus
    character `gold_update` / `character_progression_update`.
18. Invalid gold amounts, insufficient wallet gold, and insufficient stash gold reject without
    partial mutation.
19. Stash item and gold mutations persist across reconnect, `/state` inspection if exposed there,
    replay reconstruction, and fresh session creation.
20. Stash opening and mutations are replay-safe: given the same session seed, session-start account
    stash snapshot, and ordered inputs, replay derives the same inventory, stash, gold, and event
    stream.
21. Co-op remains level-scoped for world state. A player opening or mutating their stash does not
    create public floor loot, public entity changes, or private deltas for other accounts.
22. The Godot client presents a side stash panel alongside the existing inventory panel when the
    stash is open, reusing existing item row/grid and tooltip presentation where practical.
23. The Godot client can deposit and withdraw items through explicit UI gestures, and can deposit
    and withdraw gold through explicit numeric controls or bot-callable UI actions.
24. The stash UI shows character gold and stash gold as distinct balances.
25. Existing inventory, equipment, hotbar, vendor, buyback, gold auto-pickup, and co-op scenarios
    remain green.
26. Protocol examples, shared validation, Go tests, client tests, protocol bot, client bot, replay,
    and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v50_spec-account-stash-storage.md - this spec
docs/plans/v50_2026-06-10-account-stash-storage.md - implementation plan
PROGRESS.md - lifecycle update when v50 ships

shared/protocol/envelope.v7.schema.json - protocol version bump for stash intents/events if needed
shared/protocol/messages.v7.schema.json - stash open/deposit/withdraw item/gold intents
shared/protocol/session_snapshot.v7.schema.json - account stash snapshot fields if included on session attach
shared/protocol/state_delta.v7.schema.json - stash item/gold changes and stash events
shared/protocol/examples/state_delta.json - stash open, item deposit/withdraw, gold deposit/withdraw examples

server/migrations/*_account_stash.sql - account stash item rows and account stash gold balance
server/internal/store/models.go - account stash models
server/internal/store/interfaces.go - account stash repository surface
server/internal/store/repos.go - atomic stash item/gold persistence operations
server/internal/store/store_test.go - account-wide persistence and transaction coverage

server/internal/game/types.go - stash item/gold views, intents, changes/events
server/internal/game/sim.go - town stash interactable, validation, mutation, tick/replay behavior
server/internal/game/game_test.go - item/gold deposit/withdraw, rejection, co-op privacy, replay-order tests
server/internal/realtime/session_loop.go - load account stash, persist mutations, private fanout
server/internal/realtime/session_loop_test.go - owner-scoped stash routing and persistence tests
server/internal/replay/replay_test.go - stash replay reconstruction from session-start snapshot
server/internal/http/* - `/state` or character/session inspection if stash state is exposed

client/scripts/main.gd - stash event/change routing and panel lifecycle
client/scripts/inventory_panel.gd - shared inventory/stash item presentation if reused
client/scripts/stash_panel.gd - stash panel if a separate component is cleaner
client/tests/test_stash_panel.gd - panel state, item transfer, gold transfer debug tests
client/scripts/bot_controller.gd - bot hooks for stash open/deposit/withdraw if needed
client/scripts/bot_scenario_runner.gd - client-bot stash assertions/actions if needed

tools/bot/run.py - protocol stash helpers and assertions
tools/bot/test_protocol.py - helper tests if new protocol helpers are added
tools/bot/scenarios/36_account_stash_storage.json - protocol proof
tools/bot/scenarios/client/23_account_stash_panel.json - Godot client proof
```

Protocol note: v50 likely needs protocol v7 because clients must send stash item/gold intents and
receive actor-private stash item/gold views. The plan may choose the exact event/change names, but
the contract must clearly distinguish character gold from account stash gold and must not expose
stash payloads to other accounts.

Persistence note: account-wide stash state should not be tied to one character id. Existing
character-owned item rows may be moved, copied, or represented through a new account-level stash
table as the plan determines, but ownership must remain single-location and atomic: an item cannot
exist in both a character bag and the account stash after a successful mutation.

## 5. Data And Behavior Draft

### 5.1 Stash ownership model

The first stash model is account-wide:

```text
account_id -> stash_items[] + stash_gold
character_id -> inventory/equipment/hotbar/progression/gold
```

Item transfers move an item between exactly one character bag and the account stash. Gold transfers
move integer balance between the current character wallet and account stash gold. The server is the
only writer for both directions.

### 5.2 Stash opening

The town stash is an interactable like vendor, stairs, chests, and waypoints. Opening it should use
existing range/auto-approach conventions where possible. The opened view is actor-private and
contains server-authored item payloads compatible with current inventory/shop tooltip rendering:
item identity, rarity, slot/category, rolled stats, requirements, equip preview/comparison fields
where already available or locally derivable from existing helpers.

### 5.3 Item deposit and withdrawal

Item deposit and withdrawal are explicit intents. They do not run as passive pickup and they do not
touch floor loot. Deposit accepts only bag items to keep ownership rules simple. Withdraw adds the
item to the active character bag only when capacity allows.

Hotbar-assigned items are rejected on deposit by default. A future slice may add a UI that clears
hotbar references before deposit, but v50 should keep mutation semantics obvious.

### 5.4 Gold deposit and withdrawal

Gold transfer is explicit and amount-based:

```text
deposit_gold(amount): character.gold -= amount; account_stash.gold += amount
withdraw_gold(amount): account_stash.gold -= amount; character.gold += amount
```

Amounts must be positive integers. The server validates balances and persists both sides
atomically. Stash gold changes are private account state; character gold changes continue through
the existing character wallet/progression path so the HUD and inventory panel remain consistent.

## 6. Test And Bot Proof

Expected coverage:

- Shared protocol validation for stash intents, stash item/gold views, examples, and private deltas.
- Store tests proving:
  - account stash items are account-scoped, not character-scoped;
  - account stash gold persists independently of character gold;
  - item and gold transfers are atomic;
  - invalid withdraw/deposit cases do not partially mutate.
- Go sim/realtime tests proving:
  - town stash opens only in town/range through existing interactable behavior;
  - item deposit/withdraw success and rejection paths;
  - gold deposit/withdraw success and rejection paths;
  - full-bag withdraw rejection;
  - equipped and hotbar-assigned item deposit rejection;
  - same-account different-character visibility;
  - different-account co-op privacy;
  - replay reconstruction from a session-start account stash snapshot.
- Godot unit tests proving the stash panel renders item rows, character gold, stash gold, and emits
  the correct item/gold intents.
- Protocol bot scenario `36_account_stash_storage.json` proving:
  - acquire a real dungeon item and gold;
  - return to town and open stash;
  - deposit item and gold;
  - fresh session on the same account sees the stored item/gold;
  - another character on the same account sees the same stash item/gold;
  - withdraw item and gold;
  - full-bag withdraw rejects without losing the stash item;
  - replay verification passes.
- Client bot scenario `23_account_stash_panel.json` proving:
  - Godot opens stash in town;
  - visible inventory/stash panels stay synchronized after deposit/withdraw;
  - visible character/stash gold balances update after deposit/withdraw.

Expected verification commands:

```bash
make validate-shared
cd server && go test ./internal/store/... ./internal/game/... ./internal/realtime/... ./internal/replay/...
make bot scenario=36_account_stash_storage.json
make bot-client scenario=23_account_stash_panel.json HEADLESS=1
make ci
```

## 7. Open Questions And Risks

No planning-blocking product questions remain from the v50 brief:

- Account-wide stash: **confirmed**.
- Item slot capacity: default **50**.
- Hotbar-assigned item deposit: default reject.
- UI direction: reuse inventory panel pattern with a side stash panel.
- Gold storage: include account stash gold deposit/withdraw.

Implementation risks for the plan:

- Protocol v7 must preserve private fanout. Stash payloads and stash gold must never broadcast to
  other accounts in co-op.
- Account-wide item ownership is the first durable item state not scoped to one character. The plan
  must choose a clean persistence model and atomic transfer path instead of overloading character
  inventory in a way that leaves duplicate ownership.
- Replay needs a session-start account stash snapshot or equivalent stable baseline. Replaying
  against live current account stash would make historical sessions nondeterministic.
- Client UI work touches inventory presentation. The plan must record the Godot plugin shortcut
  checklist result before editing client UI code.
