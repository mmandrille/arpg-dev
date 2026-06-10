# Spec: `coop-rewards-and-scaling`

Status: Draft
Date: 2026-06-09
Branch: `main`
Codename: `coop-rewards-and-scaling`
Slice: v48 - co-op proximity XP and logarithmic monster scaling
Baseline: v47 `shop-stock-lifecycle`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - co-op players share one authoritative `Sim`
- [`v30_spec-monster-rarity-and-loot-scaling.md`](v30_spec-monster-rarity-and-loot-scaling.md) - generated monster rarity HP/damage/XP scaling
- [`v33_spec-true-coop-session.md`](v33_spec-true-coop-session.md) - current co-op actor ownership and killing-blow XP baseline
- [`v38_spec-session-browser-and-uncapped-coop-menu.md`](v38_spec-session-browser-and-uncapped-coop-menu.md) - uncapped listed co-op sessions and N-member replay proof
- [`v46_spec-client-join-game-proof.md`](v46_spec-client-join-game-proof.md) - real Godot Join Game proof

## 1. Purpose

Co-op is now player-facing and can be joined through the real Godot `Join Game` path, but combat
rewards and monster challenge still behave like the v33 foundation: the killing-blow actor receives
XP, nearby party members do not, and monsters do not account for party size.

This slice makes co-op combat behave like co-op:

- Nearby eligible party members receive full monster XP when any party member kills a monster.
- Disconnected, dead, different-level, or out-of-radius party members receive no shared XP.
- Monster HP and damage scale logarithmically with active same-level party size, capped at +50%.
- Solo play remains unchanged.

The goal is a thin server-authoritative gameplay slice. The client continues to render existing
XP/progression and combat events; it does not compute eligibility, XP, HP, or damage.

## 2. Non-goals

- No loot allocation changes. The player who picks up an item still owns it.
- No shared gold pickup, shared buyback, shared vendor stock, or shared inventory.
- No explicit party bonus multiplier beyond monster HP/damage scaling.
- No party UI, XP toast redesign, chat, emotes, ready checks, lobby staging, Steam lobby, friend
  flows, or matchmaking.
- No friendly fire, PvP, proximity revive, respawn, or party travel.
- No monster population-count scaling or extra spawned monsters.
- No dynamic re-roll of generated dungeon layouts.
- No client-authoritative gameplay logic or Godot high-level multiplayer.
- No Protobuf migration.

## 3. Acceptance Criteria

1. Shared rules define co-op XP eligibility and party challenge scaling as data, not hardcoded
   literals in gameplay logic.
2. The default co-op XP rule grants each eligible player the full monster XP value. XP is not split.
3. A player is eligible for shared kill XP only when all are true:
   - connected,
   - alive,
   - on the same level as the killed monster,
   - within the configured XP share radius of the killed monster at kill time.
4. The killing player receives XP once, even if also eligible by proximity.
5. Same-level out-of-radius players receive no shared XP.
6. Different-level players receive no shared XP.
7. Disconnected or dead party members receive no shared XP.
8. Shared XP can level up any eligible recipient and can grant stat and skill points through the
   existing progression rules.
9. XP/progression changes are recipient-scoped: each client sees only its own progression update,
   and persistence writes each rewarded character, not only the killing actor.
10. Existing solo monster XP behavior remains unchanged.
11. Loot ownership remains unchanged: a non-killing player may still pick up dropped loot, but XP
    sharing does not grant inventory items.
12. Shared rules define logarithmic party challenge scaling:

    ```text
    party_count = alive connected players on the monster's level
    party_scale_bonus = min(max_bonus, per_double_bonus * log2(max(1, party_count)))
    party_scale_multiplier = 1.0 + party_scale_bonus
    ```

    with defaults `per_double_bonus = 0.25` and `max_bonus = 0.50`.
13. The default scaling yields these effective targets:

    | Alive connected same-level players | Multiplier |
    |------------------------------------|------------|
    | 1 | 1.00 |
    | 2 | about 1.25 |
    | 3 | about 1.40 |
    | 4 | 1.50 |
    | 5+ | 1.50 capped |

14. Monster HP/max HP scale at monster spawn time using the party count for that spawn context.
    Existing live monsters are not retroactively healed or rebalanced when a player joins or leaves.
15. Monster attack damage scales at attack resolution time using current alive connected same-level
    party count, so late joiners still increase danger even for already-spawned monsters.
16. Party scaling composes deterministically with existing monster rarity scaling from v30. Rarity
    scaling and party scaling both apply, with deterministic rounding documented in tests.
17. The scale calculation does not use wall-clock time, unseeded randomness, map iteration order, or
    client state.
18. Replay reconstructs the same XP events, progression changes, monster HP, monster attack damage,
    and final character progression from the same seed and ordered inputs.
19. Existing co-op session/list/join proofs remain green.
20. Protocol examples, shared validation, Go tests, protocol bot, replay, and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v48_spec-coop-rewards-and-scaling.md - this spec
docs/plans/v48_2026-06-09-coop-rewards-and-scaling.md - implementation plan
PROGRESS.md - lifecycle update when v48 ships

shared/rules/combat.v0.json - co-op XP and party challenge rule data
shared/rules/combat.v0.schema.json - validate co-op rule fields
shared/golden/coop_rewards_and_scaling.json - XP eligibility and multiplier fixture
shared/golden/coop_rewards_and_scaling.v0.schema.json - fixture schema
tools/validate_shared.py - validate co-op combat rule/golden drift

server/internal/game/rules.go - parse/validate co-op combat rules
server/internal/game/sim.go - XP eligibility, multi-recipient XP award, party challenge multiplier
server/internal/game/types.go - private-change routing metadata only if needed
server/internal/game/game_test.go - focused XP/scaling/determinism tests
server/internal/realtime/session_loop.go - persist and fan out multi-recipient progression changes
server/internal/realtime/session_loop_test.go - recipient-scoped filtering/persistence routing tests
server/internal/replay/replay.go - replay parity if result routing changes
server/internal/replay/replay_test.go - co-op XP replay proof

tools/bot/run.py - co-op rewards/scaling scenario helper if current helpers are insufficient
tools/bot/test_protocol.py - helper/scenario loader coverage if new scenario metadata is added
tools/bot/scenarios/34_coop_rewards_and_scaling.json - protocol proof
```

Protocol note: v48 should not require a protocol schema bump if existing v6 `experience_gained`,
`character_leveled`, `skill_point_gained`, `character_progression_update`, and combat events can be
emitted as recipient-scoped deltas. If implementation requires private change ownership metadata on
the wire, stop and revise the spec/plan before changing protocol shape.

Client note: no Godot UI, art, camera, or inventory presentation work is expected. The plan does not
need a plugin adoption decision unless implementation discovers a real client-visible change.

## 5. Data And Behavior Draft

### 5.1 Shared combat rules

Extend `combat.v0.json` with a declarative co-op block. Suggested shape:

```json
{
  "coop": {
    "xp_share": {
      "enabled": true,
      "radius": 10.0,
      "full_xp_per_eligible_player": true,
      "include_dead_players": false,
      "include_disconnected_players": false
    },
    "party_challenge": {
      "enabled": true,
      "per_double_bonus": 0.25,
      "max_bonus": 0.50,
      "hp_scales_at_spawn": true,
      "damage_scales_at_attack": true
    }
  }
}
```

The exact field names are plan-level detail, but validation must enforce positive radius,
non-negative bonus values, `max_bonus >= per_double_bonus`, and a deterministic formula that does
not depend on map iteration or wall-clock time.

### 5.2 XP sharing

On monster death, the server resolves the base XP exactly as today, including any generated monster
rarity XP multiplier already baked into the monster entity. It then builds the eligible player list
from the monster's level in stable player-id order:

```text
eligible = alive connected players
  where player.current_level == monster.level
  and distance(player.position, monster.position) <= xp_share.radius
```

Each eligible player receives the full XP amount. The killing actor must not receive duplicate XP if
they also satisfy the proximity rule.

Each rewarded player must produce actor/recipient-scoped progression changes so realtime fanout,
persistence, replay, and reconnect snapshots agree. The implementation may represent that as one
`TickResult` per rewarded player or another explicit server-internal ownership marker; it must not
expose another player's private progression to unrelated clients.

### 5.3 Monster HP scaling

At monster spawn time, compute party scale from alive connected players on the monster's level for
that spawn context:

```text
scaled_max_hp = round_positive(base_scaled_max_hp * party_scale_multiplier)
current_hp = scaled_max_hp
```

`base_scaled_max_hp` means the monster max HP after existing rarity/depth/template scaling but
before party scaling. Scaling applies to generated dungeon mobs, boss-floor bosses, and any future
runtime monster spawns that use the normal monster spawn path. Preset lab/static monsters may remain
unscaled unless the implementation can route them through the same spawn helper without changing
old test intent.

Existing live monsters are not retroactively rebalanced when party membership or same-level presence
changes after spawn. This avoids healing or damaging half-fought monsters because a player joins,
disconnects, descends, or dies.

### 5.4 Monster damage scaling

At monster attack resolution time, compute the same party scale using current alive connected
players on the attacking monster's level. Scale the already-rarity-scaled monster damage range:

```text
scaled_min = round_positive(base_attack_min * party_scale_multiplier)
scaled_max = round_positive(base_attack_max * party_scale_multiplier)
```

The plan must choose one rounding rule and pin it in tests. Default: deterministic nearest integer
with minimum `1` for positive HP/damage values.

Damage scaling applies to proactive monster melee attacks, retaliation-style monster damage if it
is represented by monster attack rules, boss active-phase damage, and any future server-owned
monster damage path that uses the shared monster damage resolver. If a path is intentionally
excluded in implementation, the plan must document that exclusion and test it.

## 6. Test And Bot Proof

Expected coverage:

- Shared validation for `combat.v0.json` co-op rule fields and the co-op golden fixture.
- Go sim tests for:
  - full XP to killer plus nearby eligible party member,
  - no XP for out-of-radius, different-level, disconnected, and dead players,
  - level-up/stat/skill point events for a non-killing eligible recipient,
  - unchanged solo XP,
  - unchanged loot pickup ownership,
  - 1/2/3/4/5-player scaling multiplier values,
  - HP scaling at spawn,
  - damage scaling at attack,
  - rarity scaling composed with party scaling.
- Realtime tests proving multi-recipient private progression changes fan out only to each owner and
  persist under the correct account/character.
- Replay test proving the co-op XP/scaling event stream and final progression reconstruct.
- Protocol bot scenario `34_coop_rewards_and_scaling.json` proving:
  - host creates listed or private co-op,
  - guest joins and reaches the same dungeon level,
  - a nearby non-killing guest gains XP from a host kill,
  - a far or different-level peer does not gain XP,
  - monster attack damage is higher when two alive connected players are on the same level,
  - reconnect or fresh state reflects both rewarded characters' durable progression.

Expected verification commands:

```bash
make validate-shared
cd server && go test ./internal/game/... ./internal/realtime/... ./internal/replay/...
make bot scenario=34_coop_rewards_and_scaling.json
make bot scenario=23_true_coop_session.json
make bot scenario=27_session_browser_uncapped_coop.json
make ci
```

## 7. Open Questions And Risks

No planning blockers. Defaults accepted:

- Full XP for every eligible player.
- XP radius comes from shared rules; default `10.0`.
- Dead nearby players do not receive XP.
- No party bonus beyond monster HP/damage scaling.

Risks:

- Multi-recipient progression from one kill must not be persisted only to the killing actor.
- Realtime private-change filtering may need an internal owner marker or separate per-recipient
  results; do not leak one player's progression to another client.
- HP-at-spawn scaling means a player joining an already-generated floor does not retroactively make
  existing monsters tougher in HP. Damage still scales dynamically at attack time.
- Rounding and rarity composition can create brittle tests if exact values are asserted outside the
  named golden/focused unit tests.
