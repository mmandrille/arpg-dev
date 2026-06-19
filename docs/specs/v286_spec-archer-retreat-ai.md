# v286 Spec: Archer Retreat AI

Status: Implemented
Date: 2026-06-19
Codename: `archer-retreat-ai`

## Purpose

Make ranged chase monsters backpedal when a player closes inside their preferred fighting distance.
The first user-visible target is `dungeon_archer`: if the archer has line of sight but the player is
too close, it should seek a reachable clear-shot cell farther away instead of standing still at
near-melee distance.

This should preserve the existing v52/v273 behavior where blocked archers can move closer to regain
a clear shot.

## Non-goals

- Do not add cover seeking, squad tactics, predictive leading, strafing, or elite archer packs.
- Do not change projectile damage, hit chance, cooldown, attack range, movement speed, loot, or spawn
  composition tuning except for the new ranged preferred-distance field.
- Do not change melee monster chase behavior.
- Do not add client-only archer animation or VFX polish.

## Acceptance Criteria

- Ranged monster rules can declare a `preferred_min_range` value.
- Rule validation rejects `preferred_min_range` on non-ranged monsters and rejects values that do not
  fit inside the monster's attack range.
- `dungeon_archer` declares a preferred minimum range in shared rules.
- When a ranged monster is in chase mode, has the player inside `preferred_min_range`, and a reachable
  clear-shot retreat cell exists, the monster moves toward that cell.
- Retreat goals stay within ranged attack reach and maintain clear line of sight to the target player.
- Existing blocked-shot repositioning still works: an archer that cannot shoot from far away can still
  move closer to acquire a clear shot.
- The behavior is deterministic under the existing sim tick and pathfinding rules.

## Scope And Likely Files

- `shared/rules/monsters.v0.json` adds `preferred_min_range` for `dungeon_archer`.
- `shared/rules/monsters.v0.schema.json` allows the new ranged-only field.
- `server/internal/game/rules.go` loads and validates the new field.
- `server/internal/game/monster_ranged_positioning.go` owns retreat-goal search.
- `server/internal/game/sim.go` consults retreat movement before treating ranged monsters as already
  satisfied in attack range.
- `shared/rules/worlds.v0.json` may add a compact close-start archer lab.
- `server/internal/game/ranged_monster_positioning_test.go` proves retreat and existing blocked-shot
  reposition behavior.
- A protocol bot scenario can prove the authored lab end to end.

## Test And Bot Proof

Focused checks:

```bash
(cd server && go test ./internal/game -run 'TestRangedMonster' -count=1)
make validate-shared
make bot scenario=archer_retreat_ai
make maintainability
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=25_ranged_monster_ai
```

## Asset And Plugin Decision

- Adopt: existing `dungeon_archer` rules, authoritative monster movement/pathing, and current ranged
  bot/client scenarios.
- Borrow: existing ranged positioning helpers and close-shot lab pattern.
- Reject: external assets, plugins, client animation dependencies, cover-object systems, and new
  projectile mechanics.

## Open Questions And Risks

- Retreat must not fight the blocked-shot reposition fallback. The movement decision should prefer
  retreat only when a valid farther clear-shot goal exists; otherwise existing attack or chase logic
  should continue.
- In cramped rooms, no valid retreat cell may exist. That should degrade to current behavior rather
  than leaving the monster with an invalid movement target.
