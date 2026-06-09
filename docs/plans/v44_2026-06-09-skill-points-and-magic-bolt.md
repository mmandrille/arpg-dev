# v44 Plan - Skill Points and Magic Bolt

Status: Ready for implementation
Goal: Add skill points, a single spendable `magic_bolt` active skill, server-owned cooldowns, and the attack-speed foundation that drives that cooldown.
Architecture: The Go sim remains authoritative for progression grants, skill ranks, mana spend, cooldowns, projectile/damage resolution, and persistence. Shared JSON owns progression cadence, skill tuning, weapon speed, validation fixtures, and cross-language goldens. Godot renders one skills panel and one skill slot from server-owned state, with local cooldown interpolation treated as presentation only.
Tech stack: Shared JSON schemas/rules/goldens, Go sim/store/replay/http, protocol v5 JSON schemas, Godot GDScript UI, Python protocol bot, client bot.

## Baseline and shortcut decision

Baseline is v43 `equipment-requirements-and-preview` on `main`. Reuse v26 character progression and stat panel, v28 hotbar/equipment, v31 effective stat breakdowns, v37 directional attack input patterns, v39 mana payloads, and v12/v37 projectile sweep/resolution paths.

Godot plugin shortcut decision: **reject** RPG/skill-tree/inventory logic plugins for v44 because skill progression and combat authority live in Go and shared rules. **Borrow pattern only if needed** for a cooldown-slot visual, but implement the actual one-slot UI in existing in-repo `Control` scripts so headless client tests remain simple.

Protocol decision: create coordinated **v5** protocol schemas because v44 adds new client intent types and new top-level snapshot/delta state (`skill_progression`, `skill_cooldowns`). Keep compatibility as a coordinated repo update; do not preserve stale v4-only assumptions inside client/bot/server code.

Timing decision: `magic_bolt` requires `rank >= 1`, so the proof levels to 3, spends one skill point, then casts. Normal basic-attack cooldown and animation-speed scaling remain deferred; v44 computes effective attack speed and attack interval for stat views and skill cooldown only.

Attack-speed formula decision for implementation:

```text
dex_speed = character_progression.derived_stats.attack_speed
weapon_speed = equipped primary weapon attack_speed, or 1.0 unarmed
item_speed_percent = sum equipped base/rolled attack_speed_percent
effective_attack_speed = clamp(dex_speed * weapon_speed * (1 + item_speed_percent / 100))
attack_interval_ticks = ceil(combat.base_attack_interval_ticks / effective_attack_speed)
magic_bolt_cooldown_ticks = attack_interval_ticks * 2
```

The exact clamps and fixture values are pinned in `shared/golden/skill_points_and_magic_bolt.json`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/plans/v44_2026-06-09-skill-points-and-magic-bolt.md` | This implementation plan |
| Modify | `docs/specs/v44_spec-skill-points-and-magic-bolt.md` | Only if planning uncovers accepted spec clarifications |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v44 ships |
| Modify | `shared/rules/character_progression.v0.schema.json` | Skill-point cadence schema and 3 stat points per level |
| Modify | `shared/rules/character_progression.v0.json` | Change `points_per_level` to 3; add skill-point cadence |
| Modify | `shared/rules/combat.v0.schema.json` | Add `base_attack_interval_ticks` and attack-speed clamp defaults if not owned by progression rules |
| Modify | `shared/rules/combat.v0.json` | First-pass base attack interval and clamps |
| Create | `shared/rules/skills.v0.schema.json` | Active skill catalog schema |
| Create | `shared/rules/skills.v0.json` | `magic_bolt` rank, mana, damage, projectile, cooldown rules |
| Modify | `shared/rules/item_templates.v0.schema.json` | Weapon `attack_speed`; signed `attack_speed_percent` stat |
| Modify | `shared/rules/item_templates.v0.json` | Weapon speed values and roll support; short/long bow proof data |
| Modify | `shared/rules/items.v0.schema.json` | Static weapon `attack_speed` |
| Modify | `shared/rules/items.v0.json` | `rusty_sword` / `training_bow` speed values |
| Create | `shared/protocol/envelope.v5.schema.json` | Protocol v5 envelope if current validation expects matched versions |
| Create | `shared/protocol/messages.v5.schema.json` | Skill point and skill cast intents |
| Create | `shared/protocol/session_snapshot.v5.schema.json` | Skill progression/cooldown state |
| Create | `shared/protocol/state_delta.v5.schema.json` | Skill progression/cooldown changes and events |
| Modify | `shared/protocol/examples/session_snapshot.json` | Skill progression/cooldown example |
| Modify | `shared/protocol/examples/state_delta.json` | Skill cast/cooldown/reject examples |
| Create | `shared/golden/skill_points_and_magic_bolt.v0.schema.json` | Golden schema |
| Create | `shared/golden/skill_points_and_magic_bolt.json` | Progression, attack speed, cooldown, rank, damage fixture |
| Modify | `shared/golden/character_progression.json` | Existing stat-point fixture from 5 to 3 |
| Modify | `tools/validate_shared.py` | New schemas, rules validation, speed relationship checks, golden drift |
| Create | `server/migrations/0012_character_skills.sql` | Skill points/ranks and session-start snapshot persistence |
| Modify | `server/internal/store/models.go` | Skill progression models |
| Modify | `server/internal/store/interfaces.go` | Skill progression repo methods |
| Modify | `server/internal/store/repos.go` | Postgres skill progression persistence |
| Modify | `server/internal/store/store_test.go` | Store persistence coverage |
| Modify | `server/internal/game/rules.go` | Parse skills, speed, cadence; validate rule semantics |
| Modify | `server/internal/game/types.go` | Skill/cooldown/attack-speed protocol views |
| Modify | `server/internal/game/sim.go` | Skill points, skill spend, cooldown, magic bolt cast/damage |
| Modify | `server/internal/game/game_test.go` | Sim, golden, cooldown, speed, mutation tests |
| Modify | `server/internal/realtime/runner.go` | Persist skill mutations; decode v5 intents |
| Modify | `server/internal/replay/replay.go` | Reconstruct from session-start skill progression |
| Modify | `server/internal/replay/replay_test.go` | Replay/session-start coverage |
| Modify | `server/internal/http/session.go` | Load skill progression for fresh sessions |
| Modify | `server/internal/http/ws_test.go` | WebSocket protocol/regression tests |
| Modify | `server/internal/inputdecode/inputdecode.go` | Decode v5 skill intents if needed |
| Modify | `tools/bot/run.py` | Skill helpers/assertions and v5 payload handling |
| Modify | `tools/bot/protocol.py` | Protocol version/constants if needed |
| Create | `tools/bot/scenarios/32_skill_points_and_magic_bolt.json` | Protocol proof |
| Modify | `tools/bot/scenarios/18_character_stats_and_leveling.json` | Stat-point expected values from 5 to 3 |
| Modify | `tools/bot/scenarios/*.json` | Audit and update other `unspent_stat_points` expectations |
| Modify | `tools/bot/test_protocol.py` | Bot helper tests |
| Modify | `client/scripts/main.gd` | Skill state parsing, intents, events |
| Create | `client/scripts/skills_panel.gd` | One-skill spend UI |
| Create | `client/scripts/skill_bar.gd` | One skill slot cooldown UI |
| Modify | `client/scripts/character_stats_panel.gd` | Attack interval/speed display if needed |
| Modify | `client/scripts/stat_labels.gd` | Skill point, rank, attack interval labels |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client skill actions/assertions |
| Modify | `client/scripts/player_health_bar.gd` | Mana update compatibility if cast uses entity updates |
| Modify | `client/tests/test_golden.gd` | Skill/attack-speed golden fixture |
| Create | `client/tests/test_skills_panel.gd` | Skill spend UI helper tests |
| Create | `client/tests/test_skill_bar.gd` | Cooldown presentation helper tests |
| Create | `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json` | Godot client proof |

## Task 1 - Shared rules, schemas, and goldens

Files:
- Modify: `shared/rules/character_progression.v0.schema.json`
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/combat.v0.schema.json`
- Modify: `shared/rules/combat.v0.json`
- Create: `shared/rules/skills.v0.schema.json`
- Create: `shared/rules/skills.v0.json`
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/rules/items.v0.schema.json`
- Modify: `shared/rules/items.v0.json`
- Create: `shared/golden/skill_points_and_magic_bolt.v0.schema.json`
- Create: `shared/golden/skill_points_and_magic_bolt.json`
- Modify: `shared/golden/character_progression.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add progression schema fields for `skill_points.points_per_grant`, `grant_every_levels`, and `first_grant_level`; change rule data to `points_per_level: 3`.
- [x] Step 1.2: Add `combat.base_attack_interval_ticks` plus min/max effective attack-speed or interval clamps if the plan keeps them in combat rules.
- [x] Step 1.3: Create declarative skill rules for `magic_bolt`: max rank, target mode, range, projectile speed, mana cost, rank-scaled damage, and `attack_interval_multiplier: 2.0`.
- [x] Step 1.4: Add weapon `attack_speed` to static item and template schemas. Add signed `attack_speed_percent` to `base_stats`, `rolled_stats`, and `rollable_stats`.
- [x] Step 1.5: Assign first-pass weapon speeds: 1H sword baseline, greatsword at `<= 0.70`, short bow faster than long bow, and static weapons covered.
- [x] Step 1.6: Add or rename the minimum bow template data needed to prove short bow vs long bow speed without breaking existing scenario item lookups.
- [x] Step 1.7: Update validation to reject missing weapon speeds, unsupported skill ids, invalid rank scaling, missing skill fixture values, and invalid weapon speed relationships.
- [x] Step 1.8: Add golden fixtures for stat-point grants, skill-point cadence, rank spend, attack speed, attack interval, cooldown ticks, mana cost, and rank-scaled damage.

```bash
make validate-shared
```

## Task 2 - Protocol v5 contracts

Files:
- Create: `shared/protocol/envelope.v5.schema.json`
- Create: `shared/protocol/messages.v5.schema.json`
- Create: `shared/protocol/session_snapshot.v5.schema.json`
- Create: `shared/protocol/state_delta.v5.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `tools/validate_shared.py`

- [x] Step 2.1: Copy v4 schemas to v5 as the starting point, then add `allocate_skill_point_intent` and `cast_skill_intent` to messages.
- [x] Step 2.2: Define `skill_progression` with `unspent_skill_points`, `skills[]`, `rank`, `max_rank`, and `can_spend`.
- [x] Step 2.3: Define `skill_cooldowns[]` with `skill_id`, `remaining_ticks`, and `total_ticks` in snapshots and relevant deltas.
- [x] Step 2.4: Add change ops for `skill_progression_update` and `skill_cooldown_update`, unless implementation cleanly embeds them in existing progression/player updates.
- [x] Step 2.5: Extend event schema for `skill_cast`, `skill_cooldown_started`, `skill_cooldown_rejected`, `skill_rank_updated`, and any projectile/impact metadata needed by client feedback.
- [x] Step 2.6: Add examples and ensure schema validation covers the new message and payload shapes.

```bash
make validate-shared
```

## Task 3 - Store, migrations, and session-start replay boundary

Files:
- Create: `server/migrations/0012_character_skills.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 3.1: Add `unspent_skill_points` to `character_progression` and `session_start_character_progression` with non-negative checks and default/backfill for existing rows.
- [x] Step 3.2: Add `character_skill_ranks(account_id, character_id, skill_id, rank)` with uniqueness and rank non-negative checks.
- [x] Step 3.3: Add `session_start_character_skill_ranks(session_id, account_id, character_id, skill_id, rank)` for immutable replay starts.
- [x] Step 3.4: Extend store models and repo interfaces to load/upsert skill points and ranks with character progression.
- [x] Step 3.5: Extend `CreateSessionStartSnapshot` / load session-start snapshot to include skill points and ranks.
- [x] Step 3.6: Add store tests for initial defaults, spend persistence, rank persistence, and session-start snapshot immutability.

```bash
go test ./internal/store/...
```

## Task 4 - Game rules and effective attack speed

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `client/tests/test_golden.gd`

- [x] Step 4.1: Add Go rule structs for skill-point cadence and `skills.v0.json`; validate `magic_bolt` exists, max rank is positive, and cooldown type is supported.
- [x] Step 4.2: Parse weapon `attack_speed` for static items and templates; support signed `attack_speed_percent` in item stat maps.
- [x] Step 4.3: Extend effective stat aggregation with weapon speed and item speed percent. Preserve existing stat breakdown semantics and add rows for weapon/item speed sources.
- [x] Step 4.4: Add attack interval calculation from effective attack speed and `combat.base_attack_interval_ticks`; expose it through a protocol view or stat breakdown key such as `attack_interval_ticks`.
- [x] Step 4.5: Update golden tests in Go and GDScript for attack speed, attack interval, and the changed stat-point grant.

```bash
go test ./internal/game/... -run 'TestCharacterProgression|TestCombatStat|TestSkill|TestGolden'
make client-unit
```

## Task 5 - Sim progression grants and skill point spend

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/http/session.go`

- [x] Step 5.1: Extend `CharacterProgressionState` and views with `UnspentSkillPoints` and skill rank map/list.
- [x] Step 5.2: Change level-up to grant 3 stat points per level and 1 skill point on levels divisible by 3, starting at level 3.
- [x] Step 5.3: Emit stable progression and skill events when one XP gain crosses multiple levels.
- [x] Step 5.4: Implement `allocate_skill_point_intent` for `magic_bolt`, requiring an unspent skill point and rank below max.
- [x] Step 5.5: Reject invalid skill ids, overspend, max-rank spends, dead actor attempts, and non-owner attempts without mutation.
- [x] Step 5.6: Persist skill-point and rank mutations by character through realtime/session paths.

```bash
go test ./internal/game/... -run 'TestCharacterProgression|TestSkillPoint'
go test ./internal/http/...
```

## Task 6 - Magic bolt cast, cooldown, and damage

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/inputdecode/inputdecode.go`
- Modify: `server/internal/realtime/runner.go`

- [x] Step 6.1: Add decoded `cast_skill_intent` payload with `skill_id` plus target id or direction. Prefer target id plus fallback direction if that fits existing input helpers cleanly.
- [x] Step 6.2: Implement server validation: alive actor, known skill, rank >= 1, enough mana, no active cooldown, valid target/direction, and range/path constraints.
- [x] Step 6.3: Reuse existing projectile movement/sweep code for `magic_bolt` where possible, with `projectile_def_id: "magic_bolt"` and skill-owned damage values.
- [x] Step 6.4: On accepted cast, subtract mana, start cooldown at `attack_interval_ticks * 2`, emit player entity/mana update and skill cooldown update.
- [x] Step 6.5: Resolve `magic_bolt` damage through existing combat event metadata so monster damage, death, loot, XP, and replay behavior remain consistent.
- [x] Step 6.6: Reject recast during cooldown without mana spend, damage, projectile spawn, or cooldown reset.
- [x] Step 6.7: Add deterministic tests for rank-scaled damage, cooldown duration, cooldown recovery, rejection reasons, wall/target behavior, and kill/XP integration.

```bash
go test ./internal/game/... -run 'TestMagicBolt|TestSkillCooldown|TestProjectile'
```

## Task 7 - Replay and HTTP state parity

Files:
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/replay/replay_test.go`
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/http/ws_test.go`

- [x] Step 7.1: Load skill progression from durable character state for fresh sessions and from session-start snapshots for replay/resume.
- [x] Step 7.2: Include skill progression and cooldown state in WebSocket snapshots, `/state`, reconnect, and replay timeline payloads.
- [x] Step 7.3: Add replay tests proving skill spend and `magic_bolt` cast reconstruct rank, mana, cooldown, damage, death/XP, and next message sequencing.
- [x] Step 7.4: Add WebSocket tests for v5 skill intents and rejection paths.

```bash
go test ./internal/replay/...
go test ./internal/http/...
```

## Task 8 - Protocol bot scenario and helper migration

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/protocol.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/32_skill_points_and_magic_bolt.json`
- Modify: `tools/bot/scenarios/18_character_stats_and_leveling.json`
- Modify: `tools/bot/scenarios/*.json`

- [x] Step 8.1: Add bot helpers for `allocate_skill_point`, `cast_skill`, `wait_skill_cooldown`, `assert_skill_progression`, and `assert_skill_cooldown`.
- [x] Step 8.2: Extend runtime state parsing for `skill_progression`, `skill_cooldowns`, v5 changes, and skill events.
- [x] Step 8.3: Update all existing scenario expectations that assume 5 unspent stat points after level-up. Audit with `rg "unspent_stat_points|points_per_level" tools shared docs`.
- [x] Step 8.4: Create `32_skill_points_and_magic_bolt.json`: level to 3, assert 3 stat points per level and 1 skill point, spend `magic_bolt`, cast, assert mana/cooldown, reject immediate recast, wait cooldown, cast again, kill/damage proof, `/state`, reconnect, replay, fresh-session persistence.
- [x] Step 8.5: Add unit tests for new bot assertion helpers.

```bash
make bot scenario=32_skill_points_and_magic_bolt.json
make bot
```

## Task 9 - Godot client skill UI and cooldown presentation

Files:
- Modify: `client/scripts/main.gd`
- Create: `client/scripts/skills_panel.gd`
- Create: `client/scripts/skill_bar.gd`
- Modify: `client/scripts/character_stats_panel.gd`
- Modify: `client/scripts/stat_labels.gd`
- Modify: `client/scripts/player_health_bar.gd`
- Modify: `client/tests/test_golden.gd`
- Create: `client/tests/test_skills_panel.gd`
- Create: `client/tests/test_skill_bar.gd`

- [x] Step 9.1: Parse `skill_progression`, `skill_cooldowns`, skill progression changes, cooldown changes, and skill events in `main.gd`.
- [x] Step 9.2: Add one openable skills panel listing only `Magic Bolt`, rank/max rank, unspent skill points, and a spend button.
- [x] Step 9.3: Send `allocate_skill_point_intent` from the skills panel and update only from authoritative deltas.
- [x] Step 9.4: Add one `magic_bolt` skill slot near the hotbar; send `cast_skill_intent` using current target or facing/direction.
- [x] Step 9.5: Disable the skill slot after accepted cast and draw gradual recovery from `remaining_ticks / total_ticks`; reconcile to authoritative cooldown updates.
- [x] Step 9.6: Add lightweight placeholder skill cast/impact feedback driven by server events. Keep animation/VFX production polish deferred.
- [x] Step 9.7: Show attack speed or attack interval in the character stats panel if the server exposes it there.
- [x] Step 9.8: Add unit tests for skill panel state, spend button enablement, cooldown fill/disabled state, and golden fixture parsing.

```bash
make client-unit
make client-smoke
```

## Task 10 - Client bot scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json`
- Modify: `scripts/bot_client.sh` or related bot runner scripts only if scenario discovery requires it

- [x] Step 10.1: Add client bot actions/assertions for opening skills panel, spending `magic_bolt`, using the skill slot, waiting for cooldown disabled/enabled state, and asserting rank/cooldown debug state.
- [x] Step 10.2: Create client scenario that reaches/uses a character with an available skill point, spends it, casts `magic_bolt`, observes disabled cooldown presentation, waits for recovery, and asserts the slot is enabled again.
- [x] Step 10.3: Keep client assertions data-driven via debug state; do not rely on pixel colors for cooldown proof.

```bash
HEADLESS=1 make bot-client scenario=19_skill_points_and_magic_bolt.json
```

## Task 11 - Lifecycle docs and CI

Files:
- Modify: `docs/PROGRESS.md`
- Modify: `docs/specs/v44_spec-skill-points-and-magic-bolt.md` only if implementation reveals accepted as-built clarifications

- [x] Step 11.1: Update `docs/PROGRESS.md` latest completed slice, slice numbering list, lifecycle row, "what each slice proved", scenario catalog, recently closed notes, and deferred backlog if scope changes.
- [x] Step 11.2: Record as-built deviations in the spec only if the implementation intentionally differs from this plan.
- [x] Step 11.3: Run final gates.

```bash
make validate-shared
go test ./internal/game/...
go test ./internal/store/...
go test ./internal/replay/...
go test ./internal/http/...
make client-unit
make bot
HEADLESS=1 make bot-client scenario=19_skill_points_and_magic_bolt.json
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `go test ./internal/game/...`
- [x] `go test ./internal/store/...`
- [x] `go test ./internal/replay/...`
- [x] `go test ./internal/http/...`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `HEADLESS=1 make bot-client scenario=19_skill_points_and_magic_bolt.json`
- [x] `make ci`

## Deferred scope

- Character classes, class restrictions, class starting stats, and class-specific skill access.
- Skill tree layout, passive skills, multiple active skills, respec/refund, skill tabs, and skill loadouts.
- Global basic-attack cooldown rebalance, animation-speed scaling, and final attack-speed balance.
- Mana regeneration, richer resource systems, buffs, debuffs, AoE, homing, summons, DOTs, and status effects.
- Production skill VFX/audio and polished skill icons.
