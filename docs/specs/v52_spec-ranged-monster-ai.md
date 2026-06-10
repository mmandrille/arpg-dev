# Spec: `ranged-monster-ai`

Status: Implemented
Date: 2026-06-10
Branch: `main`
Codename: `ranged-monster-ai`
Slice: v52 - ranged dungeon monster AI
Baseline: v51 `mystery-seller-core`
Related:

- [`../PROGRESS.md`](../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, deterministic replay, shared rules as data
- [`../adr/0007-animation-state-model.md`](../adr/0007-animation-state-model.md) - client-only presentation and event-driven reactions
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - generated dungeon levels and town safety
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client presentation shortcut checklist
- [`v12_spec-ranged-projectile-combat.md`](v12_spec-ranged-projectile-combat.md) - server-authoritative player projectile baseline
- [`v17_spec-monster-chase-movement.md`](v17_spec-monster-chase-movement.md) - server-authoritative chase/leash baseline
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - generated dungeon monster attacks
- [`v30_spec-monster-rarity-and-loot-scaling.md`](v30_spec-monster-rarity-and-loot-scaling.md) - generated monster rarity scaling
- [`v37_spec-combat-control-and-boss-ai-fixes.md`](v37_spec-combat-control-and-boss-ai-fixes.md) - projectile aggro and combat-control fixes

## 1. Purpose

Generated dungeon monsters are currently melee-only chase attackers. This slice adds the first
ranged dungeon monster variant: an archer-style monster that spawns in normal generated dungeon
floors, keeps distance when possible, fires server-owned projectiles at the player, and looks
minimally different in the Godot client by carrying a bow marker.

The proof should be a thin vertical slice:

- Shared rules define a `dungeon_archer` monster and declarative dungeon spawn mix.
- Normal negative dungeon levels include at least one generated `dungeon_archer`.
- Ranged monsters use the existing authoritative projectile entity lifecycle where possible.
- A monster projectile can be blocked by walls/barriers, expire, or hit a player and resolve combat
  through the same server-owned hit/crit/block/stat path as other monster damage.
- The client only renders and labels the distinction; it does not decide range, line of sight,
  projectile collision, damage, or AI state.

Client shortcut decision for the spec: reject external animation or AI plugin adoption for v52.
The existing `main.gd` monster presentation, primitive projectile rendering, and client bot debug
surface are sufficient for a minimal bow marker. The implementation plan must record this
adopt/borrow/reject decision.

## 2. Non-goals

- No full monster AI rewrite, behavior tree system, LimboAI integration, flocking, strafing,
  predictive leading, retreat, cover seeking, or coordinated ranged packs.
- No new projectile catalog beyond a rule field or existing `training_arrow`/arrow-style id.
- No piercing, AoE, homing, DOT, status effects, traps, summons, ammunition, or monster skill bar.
- No production archer model, production bow art, animation retargeting, VFX, sound, or bespoke
  attack animation. A simple procedural bow marker is enough.
- No ranged bosses or boss pattern deck changes.
- No town monsters, safe-zone combat changes, respawn, healing, or balance pass.
- No loot table redesign, unique archer drops, affix changes, or final depth scaling.
- No protocol schema bump unless implementation discovers an unavoidable wire-shape change.

## 3. Acceptance Criteria

1. Shared monster rules define `dungeon_archer` with chase behavior, positive ranged attack damage,
   positive cooldown, positive attack range greater than melee reach, positive projectile speed,
   and an arrow-style projectile id.
2. Shared dungeon generation rules define a monster spawn mix as data, not hardcoded Go constants,
   while preserving existing melee `dungeon_mob` generation.
3. Normal generated dungeon levels below town include at least one generated `dungeon_archer`
   whenever the configured monster count is positive. Boss floors may keep their existing special
   boss-floor population.
4. Dungeon archer rarity, HP, damage, XP, loot table, visual tint, and visual scale compose with
   the existing generated monster rarity rules from v30.
5. Static lab monsters and existing melee dungeon mobs keep their current melee attack behavior
   unless explicitly configured otherwise.
6. Ranged monster AI remains server-authoritative: the Go sim decides aggro, leash, standoff
   movement, line of sight, cooldown, projectile spawn, projectile movement, collision, combat
   outcome, HP mutation, player death, events, and replay state.
7. A ranged monster in aggro range tries to stop at a ranged standoff distance instead of walking
   into melee contact when a valid standoff exists.
8. A ranged monster only fires when the target player is in configured range and the straight
   projectile path is clear of generated walls and closed barriers.
9. If line of sight is blocked but the monster can path to a clear standoff, it moves before
   firing. If no clear standoff exists, it does not damage the player through cover.
10. On a valid ranged attack, the server spawns a wire-visible `projectile` entity owned by the
    monster and targeted at the player, advances it deterministically by tick, and removes it on
    hit, block, or expiry.
11. Monster-owned projectiles can hit living players on the same level. A hit resolves through the
    existing monster combat stat path and emits `player_damaged`, `player_killed`,
    `attack_missed`, or `attack_blocked` as appropriate.
12. Monster-owned projectiles do not damage monsters or unrelated players on other levels.
13. Projectile iteration, target selection, spawn ids, and collision tie-breaks are deterministic
    and stable under replay. No wall-clock time, unseeded randomness, or map iteration order is
    introduced in `server/internal/game`.
14. Existing player-owned ranged projectile behavior remains unchanged, including wall blocking,
    target damage, `projectile_busy`, replay, and bot coverage.
15. Protocol bot proof descends into a generated dungeon, asserts at least one live
    `dungeon_archer`, waits for an archer-owned projectile and ranged `player_damaged` or
    equivalent combat event, and proves the player was damaged without sending a player attack
    intent.
16. A focused Go test proves a blocked archer shot does not damage the player through a wall or
    closed barrier.
17. The Godot client renders `dungeon_archer` differently from melee `dungeon_mob` by adding a
    visible bow marker or equivalent minimal attachment to the monster node.
18. The Godot client bot can descend into the dungeon, observe a `dungeon_archer`, assert the bow
    marker through presentation debug state, and observe the projectile/player-damage path.
19. Shared validation, Go tests, protocol bot, client unit tests, client bot, replay coverage, and
    `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v52_spec-ranged-monster-ai.md - this spec
docs/plans/v52_2026-06-10-ranged-monster-ai.md - implementation plan
docs/PROGRESS.md - lifecycle update when v52 ships

shared/rules/monsters.v0.json - add dungeon_archer and ranged attack fields
shared/rules/monsters.v0.schema.json - validate ranged monster fields
shared/rules/dungeon_generation.v0.json - declare melee/ranged dungeon spawn mix
shared/rules/dungeon_generation.v0.schema.json - validate spawn mix if schema exists separately
shared/golden/ranged_monster_ai.json - deterministic archer spawn/attack fixture if useful
shared/golden/ranged_monster_ai.v0.schema.json - fixture schema if added
tools/validate_shared.py - validate ranged monster rule/golden drift

server/internal/game/rules.go - parse/validate ranged monster fields and dungeon spawn mix
server/internal/game/dungeon_gen.go - roll or guarantee archer spawns from rule data
server/internal/game/sim.go - ranged monster standoff, line of sight, projectile spawn/hit player
server/internal/game/types.go - reuse existing projectile entity fields unless a new field is needed
server/internal/game/game_test.go - spawn mix, ranged attack, blocked shot, replay/determinism tests
server/internal/replay/replay_test.go - replay proof if existing game replay tests do not cover it

client/scripts/main.gd - attach a minimal bow marker to dungeon_archer and expose debug state
client/scripts/bot_scenario_runner.gd - assert bow marker/presentation if current helper is insufficient
client/tests/test_client_bot.gd - validate new client scenario step/assertion shape
client/tests/test_golden.gd - validate ranged monster golden/rule references if a fixture is added

tools/bot/run.py - protocol assertions for archer/projectile owner/player damage if current helpers are insufficient
tools/bot/test_protocol.py - helper tests for new assertions
tools/bot/scenarios/38_ranged_monster_ai.json - protocol proof
tools/bot/scenarios/client/25_ranged_monster_ai.json - Godot client proof
```

Protocol note: existing entity views already include `monster_def_id` for monsters and
`owner_id`, `target_id`, and `projectile_def_id` for projectiles. Existing combat events already
cover player damage, kills, misses, and blocks. The plan should avoid a protocol bump unless the
implementation requires a new public field.

Spawn-mix note: the exact rule shape is plan-level detail. The important contract is that the
choice is data-driven, deterministic, validated, and gives the bot a reliable generated archer on
normal dungeon floors.

## 5. Test And Bot Proof

- Shared validation proves `dungeon_archer`, ranged monster fields, dungeon spawn mix, and any
  golden fixture are valid.
- Go tests cover ranged monster rule validation, generated archer placement, rarity scaling
  composition, ranged standoff, projectile spawn/update/remove, player hit resolution, blocked-shot
  no-damage behavior, and unchanged player projectile behavior.
- Protocol bot scenario `38_ranged_monster_ai.json` descends into `dungeon_levels`, observes a
  generated `dungeon_archer`, waits without attacking, asserts an archer-owned projectile appears,
  and asserts player HP decreases from a ranged monster combat event.
- Client bot scenario `25_ranged_monster_ai.json` descends through the real Godot client, asserts
  the archer presentation has a bow marker, and observes the ranged damage path.
- Existing dungeon monster, monster rarity, combat feedback, ranged projectile, boss floor, and
  mystery seller scenarios remain green.

## 6. Open Questions And Risks

- The exact spawn mix weight is a tuning detail. Default for planning: keep melee mobs dominant and
  guarantee one archer on normal generated dungeon floors so the feature is visible and testable.
- Current projectile collision is monster-target oriented. Implementation must add player-hit
  handling carefully without changing player-owned projectile behavior or allowing friendly fire.
- Ranged standoff can become brittle around generated obstacles. Tests should assert semantic
  behavior such as "does not damage through blocked line of sight" rather than pinning exact path
  coordinates.
- The bow marker should be deliberately small and procedural. Production archer art belongs in a
  later content/presentation slice.
