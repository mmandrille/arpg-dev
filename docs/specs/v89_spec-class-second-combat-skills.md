# Spec: `class-second-combat-skills`

Status: Complete
Date: 2026-06-12
Codename: `class-second-combat-skills`
Slice: v89 - class second combat skills
Baseline: v88 `skill-visual-rank-seeding`

## Purpose

Add the missing second active combat skill for Barbarian and Sorcerer:

- `cleave`: a Barbarian cone attack that damages all enemies up to 3 world units in front of the
  caster within a 50 degree cone, then pushes each hit enemy back 1 to 3 world units.
- `ice_shard`: a Sorcerer cold projectile that always hits when its projectile collides, slows hit
  enemies by 25% for 3 seconds, stacks repeated slows down to a 75% movement-speed reduction cap,
  and shatters on impact into 2 to 5 deterministic shards that can hit other enemies.

All active skills should have 100% combat hit chance by default unless the skill definition
explicitly opts into miss/block combat resolution. Existing skills that already produce damage
should keep their intended behavior, but they should not fail due to player hit chance.

The slice should make both skills visible in the existing skill catalog, skill visual matrix, and
bot proof path so future class-skill additions follow a repeatable data-driven pattern.

## Non-goals

- No final balance pass across all skills.
- No new explicit ground-targeting intent shape.
- No player-vs-player behavior.
- No persistent monster debuff storage beyond live session/replay state.
- No skill tree restructuring beyond adding the new first-row class skills.
- No new art assets; use code-native Godot presentation helpers and existing projectile primitives.

## Acceptance criteria

1. `shared/rules/skills.v0.json` contains `cleave` as a Barbarian skill and `ice_shard` as a
   Sorcerer skill, with rule-owned range, cone angle, push distance range, shard count range, slow
   amount/duration/cap, mana cost, cooldown, requirements, and rank-scaled damage where applicable.
2. Skill rule schemas validate new closed skill/effect capabilities and reject malformed cone,
   push, slow, and shard definitions.
3. Skill presentation data contains icons/summaries for both skills; `make skill-visual-list`
   lists both skills dynamically.
4. Server skill casts use 100% hit chance by default for skill damage. A skill must only miss or
   be blocked if its data explicitly opts into normal combat resolution.
5. Cleave resolves targets deterministically by distance/entity id, damages every live enemy inside
   the caster-facing cone, emits combat events for each hit, pushes hit enemies away from the caster
   by a rule-derived distance, and emits push/presentation events or entity position updates that
   clients and bots can observe.
6. Ice Shard projectile impact damages the first hit enemy, applies or refreshes a cold slow status,
   emits visible effect metadata, and spawns 2 to 5 deterministic shard projectiles or shard hit
   resolutions from the impacted enemy.
7. Ice Shard shard damage is `floor(original_damage / shard_count)` with a minimum of 1 when the
   original damage is positive; shard hit ordering and random directions are deterministic from the
   sim RNG stream.
8. Repeated Ice Shard hits can stack slow severity to a 75% movement-speed reduction cap and refresh
   duration; slowed monsters visibly move slower in the sim.
9. The Godot client renders Cleave as a short red cone/area from the caster and shows pushed enemies
   moving from authoritative position updates.
10. The Godot client renders Ice Shard impacted/slowed enemies light blue while the slow status is
    active and removes the tint on expiry or state refresh.
11. Protocol bot coverage learns and casts Cleave and Ice Shard, proving cone multi-hit/push,
    shard secondary hits, slow effect application, slow stacking cap, and 100% skill-hit behavior.
12. Skill visual tooling can run both new skills with `make skill-visual skill=cleave` and
    `make skill-visual skill=ice_shard`.
13. Existing Magic Bolt, Rage, Heal, and Holy Shield behavior remains covered and green.

## Scope and likely files

- Shared rules/assets:
  - `shared/rules/skills.v0.json`
  - `shared/rules/skills.v0.schema.json`
  - `shared/assets/skill_presentations.v0.json`
  - `shared/protocol/state_delta.v*.schema.json`
  - `shared/protocol/session_snapshot.v*.schema.json`
- Server:
  - `server/internal/game/rules.go`
  - `server/internal/game/handlers.go`
  - focused helper files if needed for skills/status effects/projectiles
  - focused Go tests outside the overlarge `game_test.go` where possible
- Bot/tools:
  - `tools/bot/scenarios/45_class_second_combat_skills.json`
  - `tools/bot/run.py` and `tools/bot/test_protocol.py` if new assertions are needed
  - `tools/bot/test_skill_demo.py`
  - `tools/bot/test_skill_visual.py`
- Client:
  - `client/scripts/main.gd` or extracted skill-presentation helper
  - `client/scripts/player_status_effect_markers.gd` or a new focused helper
  - focused GDScript tests
- Docs:
  - `docs/plans/v89_2026-06-12-class-second-combat-skills.md`
  - `docs/as-built/v89_class-second-combat-skills.md`
  - `PROGRESS.md`

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'Cleave|IceShard|Skill'`
- `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_demo.py tools/bot/test_skill_visual.py -q`
- `make bot scenario=45_class_second_combat_skills.json`
- `make skill-visual-list`
- `make skill-visual skill=cleave DRY_RUN=1`
- `make skill-visual skill=ice_shard DRY_RUN=1`
- `make client-unit`
- `make ci`

For visual verification after implementation:

```bash
make bot-visual scenario=45_class_second_combat_skills.json
```

## Open questions and risks

- Damage defaults were not specified. This slice uses data-owned conservative defaults: Cleave uses
  current effective melee/weapon damage; Ice Shard uses rank-linear projectile damage comparable to
  Magic Bolt unless plan/code discovers a better existing formula hook.
- "Tiles" are interpreted as current world units because the sim and client already use continuous
  world coordinates.
- Ice Shard random shard directions must consume only the deterministic sim RNG, never wall-clock or
  language-global randomness.
- `server/internal/game/sim.go`, `server/internal/game/game_test.go`, `client/scripts/main.gd`, and
  `tools/bot/run.py` are over the maintainability target. The plan should prefer focused helper/test
  extraction and document any unavoidable small growth.

## Shortcut decision