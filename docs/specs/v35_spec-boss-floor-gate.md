# Spec: `boss-floor-gate`

Status: Draft
Branch: `main`
Slice: v35 - first boss floor, telegraphed attack, and locked depth gate
Baseline: v34 `model-reaction-polish`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0007-animation-state-model.md`](../adr/0007-animation-state-model.md) - client-only event-driven presentation
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - generated dungeon levels and level transitions
- [`../adr/0009-boss-floors-and-timing-mechanics.md`](../adr/0009-boss-floors-and-timing-mechanics.md) - boss floor cadence, timing mechanics, and progression gate
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - plugin adoption checklist for client presentation work
- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) - generated floors and stairs
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - chest interactables and deterministic loot
- [`v31_spec-combat-stat-effects-and-feedback.md`](v31_spec-combat-stat-effects-and-feedback.md) - authoritative combat outcome metadata and client feedback
- [`v34_spec-model-reaction-polish.md`](v34_spec-model-reaction-polish.md) - shared character-like hit/death reactions

## 1. Purpose

Add the first authoritative boss-floor vertical slice. Dungeon level `-5` becomes a milestone floor
with a smaller `30 x 30` footprint, a reward chest, one deterministic boss encounter, and disabled
teleporter/down-stair exits that become usable only after the boss is killed.

The key gameplay promise is skill-based survival. Boss damage must never be inevitable. Every
damaging boss attack must show a clear telegraph before the damage lands. A telegraph can be a
floor/space indicator such as a circle, line, cone, or area, or a body/contact cue such as the boss
charging from its normal tint to full red before a melee hit resolves. The first implementation only
needs one boss template and one attack pattern, but that pattern must prove the full loop:

- The generated level marks `-5` as a boss floor.
- Boss floors are smaller than regular generated dungeon levels; the v35 target footprint is
  `30 x 30`.
- The floor places a chest before the boss arena.
- The down stairs and boss-floor teleporter start disabled/locked.
- The boss reuses the same humanoid/player 3D model for now, with special boss colors and
  `2.0x` character scale.
- Generated monster rarity presentation gains scale metadata: champions render at `1.25x` and
  uniques render at `1.5x`.
- The boss emits authoritative phase events for telegraph, active, and recovery timing.
- The client renders a readable telegraph from those events.
- The active phase applies damage only under the previously telegraphed condition: inside the
  announced area, or still in boss melee/contact range after a contact telegraph.
- A bot can observe the telegraph, move or break contact before the active phase, avoid damage,
  kill the boss, unlock the stairs, and descend.

The proof is: deterministic boss-floor generation -> shared boss pattern rules -> tick-phased
server authority -> client telegraph presentation -> bot/replay/golden coverage.

## 2. Non-goals

- No full boss catalog beyond one first boss template.
- No multiple boss patterns unless the plan finds it trivial after the first pattern is green.
- No enrage phases, summoned adds, arena walls, arena hazards, or co-op boss scaling.
- No boss health bar UI.
- No production boss art, animation, VFX, audio, or camera work. The boss reuses the current
  humanoid/player model path with size and color changes only.
- No new 3D model import or asset pack for the boss.
- No block/parry interaction with boss zones.
- No unavoidable, screen-wide, or instant boss damage.
- No durable boss kill state or durable generated map snapshots across fresh sessions.
- No quest integration, NPC dialog, or depth-band quest rewards.
- No final balance pass. First-pass values should prove the contract and remain semantically tested.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v35_spec-boss-floor-gate.md          - this slice contract
docs/plans/v35_<YYYY-MM-DD>-boss-floor-gate.md  - implementation plan
PROGRESS.md                                - lifecycle update when v35 ships

shared/rules/dungeon_generation.v0.json         - boss-floor generation block plus monster rarity visual scale
shared/rules/dungeon_generation.v0.schema.json  - validation for boss generation fields and rarity visual scale
shared/rules/boss_templates.v0.json             - one boss template catalog entry
shared/rules/boss_templates.v0.schema.json      - template schema
shared/rules/boss_patterns.v0.json              - one telegraphed timing pattern
shared/rules/boss_patterns.v0.schema.json       - phase, duration, telegraph/hit-shape, and damage schema
shared/rules/interactables.v0.json              - locked `stairs_down` state if not already representable
shared/rules/interactables.v0.schema.json       - locked state validation
shared/rules/loot_tables.v0.json                - boss-floor chest/boss loot table references if needed
shared/protocol/messages.v*.schema.json         - new event and rejection payload shapes if schemas require it
shared/protocol/state_delta.v*.schema.json      - boss phase and locked/unlocked event payloads
shared/protocol/session_snapshot.v*.schema.json - boss phase progress if needed for reconnect correctness
shared/protocol/examples/state_delta.json       - boss event examples
shared/golden/boss_floor_-5.json                - generated layout and unlock fixture
shared/golden/boss_floor_-5.v0.schema.json      - fixture schema
shared/golden/boss_pattern_timeline.json        - phase boundary and no-damage-when-dodged fixture
shared/golden/boss_pattern_timeline.v0.schema.json - fixture schema
tools/validate_shared.py                        - boss rules, telegraph guarantee, and golden drift validation

server/internal/game/dungeon_gen.go             - boss-floor detection, compact placement, locked stairs, and disabled teleporter
server/internal/game/rules.go                   - boss template/pattern parsing and validation
server/internal/game/sim.go                     - boss AI phase state, damage predicates, unlock transition
server/internal/game/types.go                   - boss state and event view types if needed
server/internal/game/*_test.go                  - deterministic generation, pattern timing, and unlock tests
server/internal/replay/*                        - replay parity if event/snapshot shapes change
server/internal/http/*_test.go                  - `/state` parity if boss state is exposed

client/scripts/main.gd                          - render telegraphs, locked exits, and monster/boss scale/tint
client/tests/test_golden.gd                     - boss fixture checks if GDScript consumes rules/goldens
client/tests/*                                  - focused telegraph presentation tests if helpers are extracted

tools/bot/run.py                                - timed move/assert helpers if current bot verbs are insufficient
tools/bot/scenarios/24_boss_floor_gate.json     - protocol bot proof
tools/bot/scenarios/client/13_boss_telegraph.json - optional client proof if reliable
```

Protocol note: v35 likely needs additive protocol schema changes for boss phase events and locked
stairs rejection feedback. The plan must decide whether this is a `v2` additive extension or a new
protocol version, then update every producer, consumer, example, bot assertion, replay path, and
client parser in the same slice.

## 4. Boss floor generation

### 4.1 Boss floor cadence

The first implementation must support the ADR-0009 cadence:

```text
levelNum < 0 && abs(levelNum) % 5 == 0
```

Level `-5` is the required proof floor. The implementation may support future floors through the
same data path, but tests should focus on `-5` unless broader support is nearly free.

Level `-1` remains a normal entry dungeon floor. Town level `0` is never a boss floor.

Boss floors must use an explicit compact generation footprint. The v35 boss floor target is
`30 x 30`, intentionally smaller than regular generated dungeon floors so the first boss encounter
is easy to find, test, and replay.

### 4.2 Required floor sequence

Generated boss floors must preserve this encounter order on the critical path:

```text
up stairs -> optional trash -> treasure chest -> boss arena -> locked down stairs
```

The first slice does not need a hand-authored arena or wall partitions. It does need deterministic
placement constraints so the chest, boss, and down stairs are not stacked or ambiguous.

The boss-floor golden fixture must assert:

- Level `-5` is classified as a boss floor.
- Exactly one boss entity is generated.
- Exactly one `treasure_chest` is generated.
- Exactly one `stairs_down` is generated in `locked` state.
- Exactly one boss-floor teleporter is generated in disabled/locked state and cannot be used before
  the boss dies.
- The chest is reachable from the up stairs before crossing into boss combat range.
- The down stairs and teleporter become usable after the boss death transition.

### 4.3 Determinism

Boss-floor generation must use per-level deterministic RNG streams and must not consume the main
combat RNG in generation. Same seed and level must produce the same:

- boss template selection,
- chest placement,
- boss placement,
- locked stairs placement,
- disabled/locked teleporter placement,
- optional trash placement,
- loot table references,
- entity IDs after replay reconstruction.

If new labeled RNG streams are introduced, the labels must be stable and documented in the plan.

## 5. Boss rules data

### 5.1 Boss template

Add one first boss template, for example:

```json
{
  "template_id": "cave_warden",
  "base_monster_def_id": "dungeon_mob",
  "pattern_deck": ["ground_slam"],
  "hp_multiplier": 8,
  "damage_multiplier": 1,
  "loot_table": "boss_drop_tier_1",
  "asset_id": "boss_cave_warden",
  "visual": {
    "model": "current_humanoid_player",
    "color": "#b77cff",
    "scale": 2.0
  }
}
```

The exact IDs can change in the plan, but the template must remain data-driven and deterministic.
The boss should reuse the current humanoid/player 3D model path for v35. Its presentation
must make it read as a boss through special colors and `2.0x` character scale, not through a new
production asset.

### 5.2 Monster visual scale by rarity

Extend generated monster rarity presentation with shared-data visual scale metadata. Minimum v35
values:

| Rarity | Required visual scale |
|--------|-----------------------|
| `common` | `1.0` |
| `champion` | `1.25` |
| `rare` | `1.0` unless the plan deliberately chooses a value |
| `unique` | `1.5` |

The client must treat scale as presentation only. Server collision, attack reach, pathfinding, and
damage must not silently change because a monster is rendered larger. If the plan chooses to make
visual scale affect collision later, that must be a separate explicit gameplay slice.

### 5.3 Boss pattern

Add one first attack pattern with explicit phases:

```json
{
  "pattern_id": "charged_melee",
  "phases": [
    {
      "kind": "telegraph",
      "duration_ticks": 30,
      "telegraph_type": "body_color_charge",
      "from_color": "#b77cff",
      "to_color": "#ff0000",
      "hit_shape": "melee_contact",
      "radius": 1.6
    },
    {
      "kind": "active",
      "duration_ticks": 4,
      "damage": { "min": 4, "max": 6 },
      "shape": "melee_contact",
      "radius": 1.6
    },
    {
      "kind": "recovery",
      "duration_ticks": 20
    }
  ],
  "cooldown_ticks": 40
}
```

The first pattern may be the body-color charged melee above or a simple ground circle if
implementation reliability demands it. The rules/schema should support both categories:

- contact telegraphs: body color charge, windup tint, or similar boss-model cue before a melee-range
  active hit,
- spatial telegraphs: circles, lines, cones, rectangles, or future floor indicators.

Do not overbuild all geometry implementations if only one is used in v35, but avoid naming that
boxes the system into floor circles forever.

### 5.4 Telegraph guarantee

This is a hard rule:

- Every phase that can deal damage must be preceded in the same pattern by a telegraph phase.
- The telegraph must describe the danger clearly enough for the client to render it, either as a
  spatial zone or a contact/body cue.
- The telegraph duration must be at least the configured minimum, defaulting to `20` ticks.
- Active damage must use the hit predicate that was fixed or announced during telegraph.
- Damage must not apply outside the active phase.
- Damage must not apply to players who escaped the announced area or broke contact before the
  active phase.

`tools/validate_shared.py` must reject boss patterns that violate this guarantee.

## 6. Server behavior

### 6.1 Boss phase state machine

The server owns boss timing. The boss AI state machine runs on authoritative sim ticks:

```text
idle/agro -> choose pattern -> telegraph -> active -> recovery -> cooldown -> choose pattern
```

In v35, the boss may choose the only available pattern every cycle. Pattern choice should still go
through the same deterministic data path that can later support a deck.

The phase state must be replay-stable. It must not depend on wall-clock time, map iteration order,
client frame time, or client-side hit detection.

### 6.2 Boss events

The server emits gameplay facts in `state_delta.events`. Minimum event set:

| Event | Required payload |
|-------|------------------|
| `boss_phase_started` | `entity_id`, `pattern_id`, `phase_index`, `phase_kind`, `duration_ticks`, `telegraph`, `hit_shape`, optional `zone`/`aim` |
| `boss_phase_ended` | `entity_id`, `pattern_id`, `phase_index`, `phase_kind` |
| `descend_blocked` or `intent_rejected` detail | `reason: "boss_alive"`, target stairs id or level |
| `interactable_state_changed` or equivalent | down stairs id, `state: "ready"` after boss kill |

Existing `player_damaged`, `player_killed`, and `monster_killed` events should continue to carry
combat outcomes. A dedicated `boss_killed` event is optional. Default: use existing
`monster_killed` with boss metadata unless the implementation finds that a dedicated event is
clearer for bot/client assertions.

### 6.3 Snapshot and reconnect

Reconnect/resume must be correct even if a client attaches mid-pattern. The plan must choose one of
these approaches:

1. Include current boss phase progress in visible entity snapshot data.
2. Reconstruct enough state from replay before snapshot emission so the next delta has a correct
   phase event boundary.

The preferred outcome is that a reconnecting client can render the current telegraph or active
state without waiting for a full new cycle, but this may be narrowed if the bot/replay proof remains
reliable and the plan documents the tradeoff.

### 6.4 Locked exits

On boss floors, `stairs_down` starts in `locked` state and the boss-floor teleporter starts in a
disabled/locked state.

- `descend_intent` targeting locked stairs is rejected with reason `boss_alive`.
- Interacting with or using the disabled boss-floor teleporter is rejected with reason `boss_alive`
  or the equivalent explicit lock reason.
- Killing the boss changes the down stairs and teleporter state to `ready`.
- After unlock, any player on that level may descend or use the teleporter.
- Unlock is level-scoped and session-scoped, not durable across fresh sessions.

Co-op policy for v35 follows ADR-0009: one shared boss kill unlocks the stairs for every player on
the level.

## 7. Client behavior

### 7.1 Telegraph rendering

The Godot client must render a readable telegraph when it receives `boss_phase_started` for a
telegraph phase.

Minimum acceptable v35 presentation:

- body color charge to full red for a contact/melee telegraph, or a circle/area indicator if the
  first pattern uses a spatial telegraph,
- visually distinct from loot, walls, player tints, and rarity tints,
- removed or changed when the telegraph ends,
- no gameplay authority or hit detection in the client.

For a contact/melee telegraph, the boss should visibly charge toward full red during the telegraph
phase, and the server should only hit if the player remains in the announced melee/contact range
when the active phase arrives. The client may reuse tint/material tweens, simple meshes, decals,
transparent materials, or immediate geometry. Production VFX is explicitly out of scope.

### 7.2 Boss and rarity scale rendering

The Godot client must render monster scale metadata from authoritative entity/rules data:

- normal generated monsters keep `1.0x` scale,
- champion monsters render at `1.25x`,
- unique monsters render at `1.5x`,
- the boss renders at `2.0x` character scale,
- boss and rarity tints remain layered with the v34 hit/death reaction tint path.

Scaling is presentation-only in v35. The client must not derive HP, damage, collision, loot, or XP
from scale.

### 7.3 Plugin and shortcut decision

The v35 plan must run the adoption checklist in
[`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md)
because this slice touches Godot presentation.

Expected v35 decision:

| Candidate | Decision | Reason |
|-----------|----------|--------|
| Existing Godot material/tint tweens and primitive geometry | Borrow/reuse | Enough for a body-color charge or simple floor telegraph without dependency cost. |
| Existing `AnimationController` and reaction path | Borrow/reuse | Boss hit/death can reuse the v34 character-like reaction path in v35. |
| Built-in `AnimationTree` | Reject for v35 unless required | No new skeletal boss animation is needed. |
| LimboAI or external behavior-tree plugin | Reject | Boss combat timing is server-authoritative and deterministic. |
| New boss model or asset pack | Reject | v35 boss reuses the current humanoid/player model with `2.0x` scale and special colors. |
| New telegraph asset pack | Reject unless plan proves a small CC0 decal/mesh is useful | The first telegraph can be a material/tint charge or generated in-code primitive. |

### 7.4 Locked stairs feedback

The client should make locked exits understandable enough for testing:

- locked stairs have a visible disabled/locked state, or
- boss-floor teleporters have a visible disabled/locked state, or
- attempting to descend produces clear feedback from the server rejection, or
- attempting to use the disabled teleporter produces clear feedback from the server rejection, or
- both, if cheap.

This is not a full UI polish requirement. Bot assertions remain authoritative.

## 8. Bot and golden proof

### 8.1 Protocol bot scenario

Add `tools/bot/scenarios/24_boss_floor_gate.json`.

Required flow:

1. Start in `dungeon_levels`.
2. Reach level `-5`.
3. Assert boss-floor layout: compact `30 x 30` floor, one chest, one boss, locked down stairs, and
   disabled/locked teleporter.
4. Open the chest.
5. Attempt to descend before killing the boss and assert `boss_alive` rejection.
6. Trigger or wait for one boss telegraph.
7. Move out of the telegraphed zone or break boss contact before the active phase.
8. Assert the player avoids damage from that attack.
9. Kill the boss.
10. Assert the down stairs and teleporter transition to `ready`.
11. Descend to level `-6`.
12. Verify `/state`, reconnect resume, and replay.

The no-damage assertion should not depend on exact tuning beyond the named boss-pattern fixture.
It should assert the semantic contract: player was inside risk before/during telegraph, moved or
broke contact before active damage, and took no damage from that attack.

### 8.2 Golden fixtures

Required fixtures:

- `shared/golden/boss_floor_-5.json`: generation classification, `30 x 30` footprint, layout
  entities, locked stairs, disabled/locked teleporter, and unlock after boss death.
- `shared/golden/boss_pattern_timeline.json`: phase boundary ticks, telegraph duration, active
  duration, hit predicate semantics, and no damage when the player escapes the announced danger.

Go tests must consume both fixtures. GDScript tests should consume the pattern fixture if the client
needs to parse shared zone shapes or if doing so prevents drift in telegraph rendering.

## 9. Acceptance criteria

1. Level `-5` is generated as a boss floor from the shared cadence rule.
2. Boss floor generation uses the compact `30 x 30` footprint and deterministically places exactly
   one pre-boss chest, one boss, and one locked down stairs.
3. The first boss template and first boss pattern are loaded from shared rules data.
4. The boss reuses the current humanoid/player 3D model path with special colors and `2.0x`
   character scale.
5. Champion generated monsters render at `1.25x` and unique generated monsters render at `1.5x`
   from shared-data presentation metadata.
6. Monster/boss visual scale remains presentation-only and does not change server collision, reach,
   HP, damage, XP, or loot.
7. Shared validation rejects any boss pattern that can apply damage without a prior telegraph phase.
8. Shared validation enforces a minimum telegraph duration for damaging attacks.
9. Boss phase progression is authoritative, tick-based, replay-stable, and independent of wall-clock
   time.
10. The server emits `boss_phase_started` and `boss_phase_ended` events with enough telegraph and
   hit-shape data for the client to render the telegraph.
11. The client renders a readable telegraph before the boss attack can damage the player.
12. Active boss damage applies only during the active phase and only to players who remain inside
   the announced danger condition, such as inside an area or within contact range.
13. The boss bot proof dodges at least one telegraphed attack and asserts no damage from that attack.
14. Attempting to descend through locked boss-floor stairs before the boss dies is rejected with
    `boss_alive` or an equivalent explicit reason.
15. Attempting to use the disabled boss-floor teleporter before the boss dies is rejected with
    `boss_alive` or an equivalent explicit reason.
16. Killing the boss unlocks the down stairs and teleporter for the level.
17. After unlock, the bot can descend from `-5` to `-6`.
18. `/state`, reconnect resume, and replay preserve boss-floor layout, locked/unlocked exit state,
    and deterministic boss events.
19. Golden fixtures cover boss-floor layout and one full boss-pattern timeline.
20. The implementation plan records adopt/borrow/reject decisions for Godot telegraph presentation
    shortcuts.
21. `make validate-shared` passes.
22. `make test-go` passes.
23. `make client-unit` passes if client helper tests are added.
24. `make bot` passes with the new boss-floor scenario.
25. `make ci` passes.

## 10. Open questions

Resolved defaults for v35:

| # | Decision |
|---|----------|
| Q-1 | Bosses must be skill-beatable. No inevitable damage. Every damaging boss attack must telegraph before landing. Telegraphs may be spatial indicators such as areas, lines, cones, or circles, or contact/body cues such as the boss charging to full red before a melee hit. |
| Q-2 | Use existing `monster_killed` with boss metadata unless a dedicated `boss_killed` event materially improves assertions or client clarity. |
| Q-3 | Use a simple material/tint charge or floor decal/mesh/primitive telegraph for v35. Production VFX is deferred. |
| Q-4 | The v35 boss reuses the current humanoid/player 3D model with special colors at `2.0x` character scale. Champion monsters render at `1.25x`; unique monsters render at `1.5x`. |

## 11. Implementation notes for planning

- Keep the slice thin: one boss floor, one boss template, one pattern, one reliable bot proof.
- Favor semantic tests over brittle generated coordinates except in named golden fixtures that own
  generation contracts.
- Call out any protocol schema bump explicitly and update all examples and consumers together.
- Be careful with deterministic RNG consumption. Boss generation, boss pattern choice, and combat
  damage rolls should have stable ownership and fixture coverage.
- If a client reconnects mid-telegraph, prefer snapshot boss phase metadata over waiting for the
  next event cycle, but do not expand the slice into a full boss UI state model.
- Treat all client telegraph visuals as presentation only. The Go sim owns every damage outcome.
