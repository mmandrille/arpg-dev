# Spec: `gold-autopickup-and-shared-loot-rules`

Status: Draft
Date: 2026-06-10
Branch: `main`
Codename: `gold-autopickup-and-shared-loot-rules`
Slice: v49 - gold auto-pickup and shared floor loot rules
Baseline: v48 `coop-rewards-and-scaling`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - co-op players share one authoritative `Sim`
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - chest loot rolls and open-once behavior
- [`v30_spec-monster-rarity-and-loot-scaling.md`](v30_spec-monster-rarity-and-loot-scaling.md) - monster rarity and loot-depth scaling
- [`v39_spec-ui-currency-and-mana-polish.md`](v39_spec-ui-currency-and-mana-polish.md) - character gold, `gold_picked_up`, and `gold_update`
- [`v48_spec-coop-rewards-and-scaling.md`](v48_spec-coop-rewards-and-scaling.md) - co-op XP sharing and party scaling; loot allocation remains unchanged

## 1. Purpose

Co-op XP and monster scaling are now server-authoritative, but floor loot should remain a classic
shared ARPG resource: a monster or chest rolls one deterministic reward stream from the existing
loot tables, places loot on the floor, and any player who can reach the floor item may pick it up.
Items have no relationship to the player who killed the monster, opened the chest, or received XP.

This slice makes that shared-loot policy explicit and improves gold pickup feel:

- Item and gold drops remain single shared floor entities, visible to everyone on the level.
- Drop rolls remain driven by existing loot rules, monster rarity, dungeon depth, and chest/source
  context.
- Equipment, consumables, quest items, and other non-gold loot still require explicit
  `action_intent` pickup.
- Gold floor entities are automatically picked up by the server when a connected, alive player
  moves into pickup range.
- If multiple players are in range for the same gold on the same tick, the lowest player id wins.
- Gold auto-pickup works on any level, including town.

The client continues to render existing loot entities and consume existing `entity_remove`,
`gold_update`, `character_progression_update`, and `gold_picked_up` messages. No client-side loot
authority or local pickup shortcut is introduced.

## 2. Non-goals

- No personal loot, hidden loot, duplicated per-player drops, loot reservations, ownership timers,
  need/greed rolls, or party loot allocation UI.
- No shared or split gold. Gold is still one floor entity; first valid pickup wins.
- No change to item drop rates, gold amount ranges, treasure classes, rarity weights, loot-depth
  scaling, chest behavior, or monster rarity behavior.
- No item auto-pickup. Only `item_def_id: "gold"` with a positive amount is auto-pickable.
- No new inventory, stash, market, vendor, trade, crafting, loot filter, sorting, or comparison UI.
- No production loot art/audio, pickup sound, floating text redesign, or party HUD.
- No client-authoritative pickup logic and no Godot high-level multiplayer.
- No protocol version bump unless the implementation proves existing v6 messages cannot represent
  the behavior safely.

## 3. Acceptance Criteria

1. Monster, chest, boss, and static loot sources continue to create shared floor loot through the
   existing loot table and treasure-class paths. v49 does not add per-player reward rolls.
2. Item and gold floor entities remain visible to all connected players on the same level through
   existing snapshot and delta entity views.
3. Floor loot entities do not gain player ownership semantics. Do not set `owner_id` on normal loot
   or gold entities as part of this slice.
4. Non-gold floor loot still requires an explicit `action_intent`, validates range or auto-approach
   exactly as today, and mutates only the picking player's inventory on success.
5. Gold floor loot can still be picked up through the existing explicit `action_intent` path when
   the target exists, preserving old protocol/client compatibility.
6. The server also checks for gold auto-pickup every authoritative tick after connected player
   movement and auto-navigation resolve.
7. A player is eligible to auto-pick a gold entity only when all are true:
   - the player is connected,
   - the player is alive,
   - the player is on the same level as the gold entity,
   - the player is within the same pickup range used for ordinary loot pickup.
8. Gold auto-pickup does not require, ack, or synthesize a client intent. It emits authoritative
   state changes/events only.
9. If more than one eligible player is in range of the same gold entity on the same tick, the server
   selects the lowest player id as the deterministic winner.
10. One player can auto-pick multiple gold entities in one tick if multiple gold entities are in
    range; entities are processed in stable entity-id order.
11. When gold is auto-picked, the server removes the gold entity, increments the winner's character
    gold, emits a winner-scoped `gold_update` and `character_progression_update`, and emits a
    `gold_picked_up` event with the picked amount and the winner's total gold.
12. Other same-level players receive the public `entity_remove` for the gold entity but do not
    receive private `gold_update`, private `character_progression_update`, or another player's
    `total_gold` value.
13. Gold auto-pickup persists through the same character gold path as manual gold pickup; reconnect,
    fresh-session state, and `/state` inspection reflect the updated winner wallet.
14. Dead or disconnected players do not auto-pick gold. If a player dies or disconnects before the
    post-movement pickup check, they are ineligible on that tick.
15. Explicit item pickup, gold auto-pickup, monster movement, monster attacks, projectiles, and boss
    phases have a documented deterministic tick ordering pinned by Go tests.
16. Replay reconstructs the same gold pickup winner, removed gold entities, emitted events, final
    wallets, and non-picked item entities from the same seed and ordered inputs.
17. Existing co-op reward/scaling behavior remains unchanged: v48 shared XP does not grant items or
    split floor gold.
18. Existing scenario flows that explicitly picked up gold are updated only where necessary to wait
    for auto-pickup instead of requiring a click; equipment/item pickup scenario steps remain click
    based.
19. Protocol examples, shared validation if touched, Go tests, protocol bot, replay, and `make ci`
    pass.

## 4. Scope And Likely Files

```text
docs/specs/v49_spec-gold-autopickup-and-shared-loot-rules.md - this spec
docs/plans/v49_2026-06-10-gold-autopickup-and-shared-loot-rules.md - implementation plan
PROGRESS.md - lifecycle update when v49 ships

server/internal/game/sim.go - post-movement gold auto-pickup, deterministic winner/order, shared loot invariants
server/internal/game/types.go - only if an internal owner marker is needed for private gold changes
server/internal/game/game_test.go - focused auto-pickup, shared loot, co-op contention, replay-order tests
server/internal/realtime/session_loop.go - private gold/progression routing if auto-pickup results are not actor-owned
server/internal/realtime/session_loop_test.go - winner-scoped gold fanout and persistence routing
server/internal/replay/replay_test.go - auto-pickup replay proof

shared/protocol/state_delta.v6.schema.json - likely unchanged; only touch if existing event/change schemas are insufficient
shared/protocol/session_snapshot.v6.schema.json - likely unchanged
shared/protocol/examples/state_delta.json - update only if examples need to show auto-pickup without correlation id

tools/bot/run.py - helper/wait behavior for passive gold pickup
tools/bot/test_protocol.py - helper/scenario tests if new bot actions are added
tools/bot/scenarios/35_gold_autopickup_shared_loot.json - protocol proof
```

Protocol note: existing v6 schema already has public loot entities, `entity_remove`,
`gold_update`, `character_progression_update`, and `gold_picked_up`. v49 should not need a protocol
bump if auto-pickup uses those shapes and omits `correlation_id` when no client intent caused the
pickup. If implementation needs a new field to express pickup source or winner privacy, stop and
revise the spec/plan before changing protocol shape.

Client note: no Godot UI changes are required. Existing clients should observe gold disappearing
and the local gold counter updating through the same delta path as manual pickup. Client bot
scenarios may need updated assertions, but presentation code should remain unchanged unless a real
regression is discovered.

## 5. Data And Behavior Draft

### 5.1 Shared floor loot policy

All drop sources continue to resolve one server-owned loot stream:

```text
source context -> loot table / treasure class -> one or more shared floor entities
```

The source context may include monster definition, monster rarity, dungeon level, boss/chest source,
and existing loot-depth rules. It must not include "which player gets a personal roll" or "which
party member was eligible for XP." Co-op members compete for the same visible floor entities.

### 5.2 Gold auto-pickup tick pass

After all connected player movement and auto-navigation have resolved for a tick, the sim scans
active levels for gold loot. The plan may choose the most local implementation shape, but behavior
must be equivalent to:

```text
for level in sorted active levels:
  for gold entity in stable entity-id order:
    eligible = alive connected players on that level within pickup range
    if eligible is not empty:
      winner = lowest player id in eligible
      remove gold entity
      add amount to winner character gold
      emit public entity_remove for the level
      emit winner-private gold/progression changes and gold_picked_up event
```

This pass is authoritative server logic. It must not depend on wall-clock time, unseeded randomness,
Go map iteration order, client hover state, client click state, or rendering.

### 5.3 Explicit pickup compatibility

Manual `action_intent` against a gold entity keeps the existing behavior when the target still
exists. A clicked out-of-range gold may still queue auto-navigation through the existing path, but
the gold can be consumed automatically as soon as the player enters pickup range. If another player
auto-picks or manually picks the same gold first, later actions against the removed target may
reject as `invalid_target`; they must not duplicate gold or recreate the entity.

## 6. Test And Bot Proof

Expected coverage:

- Go sim tests for:
  - item drops remain shared, visible, and click-required;
  - gold auto-picks when the player moves into pickup range without `action_intent`;
  - gold auto-pickup works on town and dungeon levels;
  - explicit gold pickup still works when already in range;
  - non-gold loot is not auto-picked while standing in range;
  - dead and disconnected players do not auto-pick gold;
  - same-tick co-op contention selects the lowest player id;
  - multiple gold entities in range are processed in stable entity-id order;
  - pending auto-navigation to a clicked gold does not duplicate pickup when the passive pass wins.
- Realtime tests proving only the winning player receives private gold/progression changes and
  `total_gold` event data, while all same-level players receive the gold entity removal.
- Replay test proving auto-pickup reconstructs identical event streams and wallets.
- Protocol bot scenario `35_gold_autopickup_shared_loot.json` proving:
  - host and guest see the same shared item/gold floor entities after a deterministic drop source;
  - moving near gold without clicking picks it up;
  - standing near a non-gold item does not pick it up;
  - explicit item click still picks up the item;
  - co-op contention gives the gold to the deterministic winner;
  - reconnect or fresh state reflects the winner's durable gold;
  - replay verification passes.

Expected verification commands:

```bash
make validate-shared
cd server && go test ./internal/game/... ./internal/realtime/... ./internal/replay/...
make bot scenario=35_gold_autopickup_shared_loot.json
make bot scenario=34_coop_rewards_and_scaling.json
make ci
```

## 7. Open Questions And Risks

No product questions are blocking planning. Defaults confirmed:

- Gold auto-pickup checks every authoritative tick after movement resolves.
- Same-tick co-op contention is resolved by lowest player id.
- Auto-pickup applies on any level, including town.

Implementation risks to handle in the plan:

- Auto-pickup events have no originating client intent, so `correlation_id` should be omitted unless
  a manual `action_intent` caused the pickup.
- Winner-scoped gold/progression updates may need the same explicit owner routing pattern v48 used
  for shared XP when the tick result actor is not the winner.
- Existing bot scenarios that click gold may become timing-sensitive if the gold disappears from
  passive pickup before the scripted click runs; update bot helpers to wait for either passive
  `gold_picked_up` or explicit pickup, without weakening item pickup assertions.
