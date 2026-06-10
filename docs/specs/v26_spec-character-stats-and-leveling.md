# Spec: `character-stats-and-leveling`

Status: Complete - make ci green on 2026-06-07
Branch: `feature/character-stats-and-leveling`
Slice: v26 - persistent character stats, XP, level-up points, and character sheet UI
Baseline: v25 `treasure-classes-and-guarded-chests`
Related:

- [`v16_spec-use-consumable.md`](v16_spec-use-consumable.md) - HP cap and player healing behavior
- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - dungeon monster kills and proactive player damage
- [`v22_spec-character-scoped-persistence.md`](v22_spec-character-scoped-persistence.md) - durable character state and session-start snapshots
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item stats and requirements
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - current reward-loop baseline
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - character-scoped progression
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

The game now has durable characters, persistent gear, rolled dungeon drops, treasure classes, and
guarded chests. The missing core ARPG loop is character growth: killing monsters should move a
character toward the next level, leveling should grant stat points, and spending those points
should visibly improve the character.

This slice adds the first real character progression model. Each character has a persistent level,
total experience, unspent stat points, and four base stats:

- `str`
- `dex`
- `vit`
- `magic`

Each level grants exactly 5 stat points. The player spends those points from a new left-side
character sheet. The sheet shows base stats, derived substats, unspent points, level, current XP,
and XP required for the next level. A compact XP progress bar appears below the existing hotbar.

v26 should make a small set of stat effects authoritative immediately while keeping the full model
ready for later expansion. `vit` must affect player `max_hp` now. Player damage should resolve
through the derived-stat path so `str` can influence melee weapon damage and `magic` can be reserved
for future spell damage without inventing a second stat system later. The remaining substats may be
computed and displayed from shared rules in v26, but broad combat behavior changes are intentionally
deferred.

## 2. Non-goals

- No passive skill tree.
- No class selection or stat-reset/respec UI.
- No armor equipment slots, jewelry, offhand, or new item families.
- No spell system, mana-consuming abilities, or mana regeneration loop.
- No attack-speed cooldown or animation-speed overhaul unless implementation finds a tiny,
  deterministic, low-risk display-only path.
- No crit, hit-chance, dodge, armor mitigation, or movement-speed gameplay authority in v26 unless
  explicitly approved during planning.
- No item requirement expansion beyond level checks already introduced by v23. Stat requirements
  may be represented in data later but are not enforced in this slice.
- No production character-sheet art or audio.
- No plugin-owned stat logic. Client-side plugins may only be used or borrowed for presentation.

## 3. Files to create or modify

```text
docs/specs/v26_spec-character-stats-and-leveling.md        - this slice contract
docs/plans/v26_2026-06-07-character-stats-and-leveling.md  - implementation plan
shared/rules/character_progression.v0.schema.json          - progression/stat formula contract
shared/rules/character_progression.v0.json                 - defaults, XP curve, stat formulas
shared/rules/monsters.v0.schema.json                       - optional xp_reward field
shared/rules/monsters.v0.json                              - XP rewards for killable monsters
shared/protocol/messages.v1.schema.json                    - add allocate_stat_intent
shared/protocol/session_snapshot.v1.schema.json            - character progression snapshot field
shared/protocol/state_delta.v1.schema.json                 - character_progression_update change/event fields
shared/golden/character_progression.json                   - XP, level-up, derived-stat fixture
tools/validate_shared.py                                   - progression rules and golden drift checks
server/migrations/0005_character_progression_stats.sql     - durable character stats/XP and session-start snapshot
server/internal/store/models.go                            - character progression persistence models
server/internal/store/interfaces.go                        - progression repo methods
server/internal/store/repos.go                             - Postgres implementation
server/internal/store/store_test.go                        - progression persistence tests
server/internal/game/rules.go                              - parse progression and XP rules
server/internal/game/sim.go                                - XP gain, level-up, stat allocation, derived stats
server/internal/game/game_test.go                          - progression and deterministic combat-stat tests
server/internal/realtime/runner.go                         - persist XP/stat mutations by character
server/internal/http/session.go                            - load progression on fresh session create/attach
server/internal/replay/replay.go                           - replay from session-start progression snapshot
client/scripts/main.gd                                     - parse progression snapshots/deltas and send stat intents
client/scripts/character_stats_panel.gd                    - left-side character sheet and stat allocation buttons
client/scripts/consumable_bar.gd                           - or adjacent UI: XP bar below hotbar
client/scripts/bot_scenario_runner.gd                      - client-bot steps/assertions for stats panel if needed
client/tests/test_golden.gd                                - progression golden fixture checks
tools/bot/run.py                                           - protocol bot progression assertions and stat allocation helper
tools/bot/scenarios/18_character_stats_and_leveling.json   - end-to-end protocol proof
tools/bot/scenarios/client/09_character_stats_panel.json   - Godot client UI proof
PROGRESS.md                                           - lifecycle update when v26 ships
```

The implementation plan must run the Godot plugin adoption checklist from
`docs/researchs/godot-plugins-and-shortcuts.md`. Expected decision: reject stat/RPG logic plugins
because the server owns progression; optionally borrow only UI layout ideas.

## 4. Data shapes

### Character progression rules

New file: `shared/rules/character_progression.v0.json`.

Example shape:

```json
{
  "version": 0,
  "base_stats": {
    "str": 5,
    "dex": 5,
    "vit": 5,
    "magic": 5
  },
  "points_per_level": 5,
  "level_cap": 20,
  "experience_curve": {
    "type": "table",
    "levels": [
      { "level": 1, "next_level_total_xp": 20 },
      { "level": 2, "next_level_total_xp": 55 },
      { "level": 3, "next_level_total_xp": 105 }
    ]
  },
  "derived_stats": {
    "damage_min": { "type": "linear", "base": 0, "per_str": 0.2, "per_magic": 0.0 },
    "damage_max": { "type": "linear", "base": 0, "per_str": 0.4, "per_magic": 0.0 },
    "armor": { "type": "linear", "base": 0, "per_vit": 0.25 },
    "attack_speed": { "type": "linear", "base": 1.0, "per_dex": 0.005 },
    "hit_chance": { "type": "linear", "base": 1.0, "per_dex": 0.0, "min": 0.05, "max": 1.0 },
    "crit_chance": { "type": "linear", "base": 0.05, "per_dex": 0.002, "min": 0.0, "max": 0.5 },
    "crit_damage": { "type": "linear", "base": 1.5, "per_str": 0.005 },
    "movement_speed": { "type": "linear", "base": 1.0, "per_dex": 0.002 },
    "max_hp": { "type": "linear", "base": 10, "per_vit": 2.0 },
    "max_mana": { "type": "linear", "base": 10, "per_magic": 3.0 }
  }
}
```

The concrete numbers above are defaults for planning, not balancing targets. The important
contract is that formulas are declarative, bounded, schema-validated, and evaluated equivalently in
Go and GDScript. Avoid a free-form expression language.

`experience` is total lifetime XP, not "XP into current level." XP bar progress is derived from
the previous level threshold, the current total XP, and the next level threshold.

### Monster XP rewards

`shared/rules/monsters.v0.json` gains `xp_reward` for killable monsters:

```json
{
  "dungeon_mob": {
    "max_hp": 8,
    "xp_reward": 10
  }
}
```

Rules validation must require `xp_reward >= 0`. Monsters without an explicit value default to `0`
only if the implementation keeps legacy monsters in non-progression lab worlds. Dungeon monsters
used by the default play loop should award positive XP.

### Persistent character progression

Target logical shape:

```sql
character_progression (
  character_id TEXT PRIMARY KEY REFERENCES characters(id),
  level INTEGER NOT NULL DEFAULT 1,
  experience BIGINT NOT NULL DEFAULT 0,
  unspent_stat_points INTEGER NOT NULL DEFAULT 0,
  stat_str INTEGER NOT NULL,
  stat_dex INTEGER NOT NULL,
  stat_vit INTEGER NOT NULL,
  stat_magic INTEGER NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)
```

Replay needs the same immutable session-start boundary used by v22 item and waypoint snapshots:

```sql
session_start_character_progression (
  session_id TEXT PRIMARY KEY REFERENCES sessions(id),
  character_id TEXT NOT NULL REFERENCES characters(id),
  level INTEGER NOT NULL,
  experience BIGINT NOT NULL,
  unspent_stat_points INTEGER NOT NULL,
  stat_str INTEGER NOT NULL,
  stat_dex INTEGER NOT NULL,
  stat_vit INTEGER NOT NULL,
  stat_magic INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)
```

Implementation may add columns to `characters` instead if that is cleaner, but it must preserve the
same logical contract and session-start snapshot behavior.

### Protocol progression view

`session_snapshot` gains a top-level `character_progression` object:

```json
{
  "level": 2,
  "experience": 25,
  "current_level_xp": 5,
  "next_level_xp": 35,
  "unspent_stat_points": 5,
  "base_stats": {
    "str": 5,
    "dex": 5,
    "vit": 5,
    "magic": 5
  },
  "derived_stats": {
    "damage_min": 1,
    "damage_max": 2,
    "armor": 1,
    "attack_speed": 1.025,
    "hit_chance": 1.0,
    "crit_chance": 0.06,
    "crit_damage": 1.525,
    "movement_speed": 1.01,
    "max_hp": 20,
    "max_mana": 25
  }
}
```

`state_delta` gains either a `character_progression_update` change or an event with the same
payload. Prefer a change for state replacement and separate events for presentation:

```json
{
  "op": "character_progression_update",
  "character_progression": { "...": "same shape as snapshot" }
}
```

Events:

- `experience_gained` with `amount`, `total_experience`
- `character_leveled` with `level`, `unspent_stat_points`
- `stat_allocated` with `stat`, `amount`, `unspent_stat_points`

### Stat allocation intent

Add `allocate_stat_intent`:

```json
{
  "stat": "vit",
  "points": 1
}
```

Valid stat names are exactly `str`, `dex`, `vit`, and `magic`. `points` must be a positive integer.
The server validates available points and applies the mutation atomically in the Sim. The client
does not pre-spend points locally; it waits for authoritative acceptance and progression update.

## 5. Architecture and flow

### Fresh session bootstrap

```text
authenticate account
select/create character
load durable character items + waypoints
load or initialize durable character progression from shared defaults
persist immutable session-start progression snapshot
start Sim from fresh world/session seed
inject items, waypoints, and character progression into initial snapshot
```

If a fresh character has no progression row, initialize it from
`shared/rules/character_progression.v0.json`.

### XP and level-up flow

```text
player kills monster
  -> Sim emits monster_killed as today
  -> Sim reads monster xp_reward
  -> Sim adds XP to character progression once
  -> Sim loops level-ups while total XP reaches the next threshold
  -> each level grants points_per_level unspent stat points
  -> Sim emits progression update and presentation events
  -> realtime runner persists progression against session.character_id
  -> reconnect and fresh sessions see the updated progression
```

XP must only be granted once per monster death. Replay must reproduce the same XP, level, and event
sequence from the same seed and input stream.

### Stat allocation flow

```text
client sends allocate_stat_intent { stat, points }
  -> server validates stat and available unspent points
  -> Sim increments the base stat and decrements unspent points
  -> Sim recomputes derived stats
  -> if max_hp increases, current HP increases by the gained max HP amount, capped at new max_hp
  -> Sim emits character_progression_update and stat_allocated
  -> realtime runner persists the mutation
```

If `vit` changes `max_hp`, player entity `max_hp` must also update through the authoritative entity
view so the existing health UI remains consistent. If `max_hp` later decreases due to future respec
or equipment behavior, that future slice must define clamping rules; v26 only spends points upward.

### Derived stat authority in v26

The full substat list exists in the progression view:

- damage
- armor
- attack speed
- hit chance
- crit chance
- crit damage
- movement speed
- HP
- mana

Required authoritative effects in v26:

- `max_hp` from `vit`
- player damage bonus through the same damage resolver used by fixed and rolled weapons

Display-only in v26 unless planning explicitly expands scope:

- armor
- attack speed
- hit chance
- crit chance
- crit damage
- movement speed
- max mana

This split keeps the model honest without coupling one slice to a full combat rebalance.

## 6. Client UI

Add a left-side character stats panel, toggled with `C`.

The panel should show:

- level
- total XP and XP to next
- unspent stat points
- base stats with `+` buttons when points are available
- derived substats in a compact read-only list

The panel is display-only except for stat allocation buttons, which send `allocate_stat_intent`.
Buttons are disabled when `unspent_stat_points == 0` or when menus/pause overlays block gameplay
input.

Add an XP progress bar below the existing bottom-center hotbar. The bar must not overlap the hotbar,
health display, inventory panel, waypoint panel, pause/menu overlays, or viewport edges at supported
window sizes. It should update from authoritative snapshots/deltas and expose debug state for the
Godot client bot.

## 7. Architecture and determinism

Progression is authoritative server state. The client may compute display strings from the
authoritative `character_progression` payload, but it must not decide XP gain, level-up, stat
spending, HP changes, or combat damage.

All progression calculations must be deterministic:

- no wall-clock time in `server/internal/game`
- no unseeded randomness
- no map iteration in XP/stat application paths
- stable event/change ordering when a kill causes XP gain and level-up

Session replay must load the session-start character progression snapshot and then replay recorded
inputs. It must not read mutable live character progression rows that may have changed after the
historical session began.

## 8. Acceptance criteria

1. `make validate-shared` validates character progression rules, XP curve monotonicity, base stat
   defaults, derived stat formula references, monster XP rewards, and golden fixtures.
2. Fresh characters start at level 1 with total XP 0, 0 unspent stat points, and base stats from
   shared progression rules.
3. Killing a dungeon monster grants XP exactly once from `monster.xp_reward`.
4. Crossing an XP threshold increases character level and grants exactly 5 unspent stat points per
   level gained.
5. Multiple level-ups from one XP gain are handled deterministically if total XP crosses multiple
   thresholds.
6. `allocate_stat_intent` accepts valid spends and rejects invalid stat names, non-positive point
   counts, and overspending.
7. Allocating `vit` recomputes `max_hp`, updates the player entity `max_hp`, and increases current
   HP by the gained max HP amount, capped at the new max.
8. Player damage resolution uses the derived damage bonus path in addition to current fixed/rolled
   weapon damage.
9. Character progression persists across fresh sessions for the same account/character.
10. Same-session reconnect restores current progression from recorded session state.
11. Replay uses session-start progression snapshots plus recorded inputs and does not drift after
    later fresh-session stat changes.
12. `/state`, WebSocket `session_snapshot`, and replay timeline expose the same character
    progression view.
13. Godot displays the character stats panel on the left side and sends stat allocation intents
    only through server protocol.
14. Godot displays an XP progress bar below the hotbar without UI overlap at supported window sizes.
15. Protocol bot scenario `18_character_stats_and_leveling.json` proves XP gain, level-up, stat
    allocation, `/state`, reconnect, replay, and fresh-session persistence.
16. Godot client bot scenario `09_character_stats_panel.json` proves panel visibility, XP bar debug
    state, and stat allocation UI.
17. `make ci` green.

## 9. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'CharacterProgression|Experience|Level|Stat|Damage|HP'`
3. `cd server && go test ./internal/store/... -run 'CharacterProgression|SessionStart'`
4. `cd server && go test ./internal/http/... ./internal/replay/... -run 'CharacterProgression|Experience|Replay|State'`
5. `make client-unit`
6. `make bot` - includes `18_character_stats_and_leveling.json`
7. `make bot-client` - includes `09_character_stats_panel.json`
8. `make ci`
9. Manual: `make play`, kill dungeon mobs until level-up, open the character panel, spend points,
   verify HP/damage values update, return to main menu, continue the same character, and confirm the
   progression persists.

## 10. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | v26 introduces `str`, `dex`, `vit`, and `magic` as base stats. | Matches the requested ARPG stat model and gives future passive skills/items a stable target. |
| 2 | Each level grants 5 unspent stat points. | Keeps the reward immediately understandable and matches the requested design. |
| 3 | XP is total lifetime XP; level thresholds are total XP thresholds. | Avoids subtracting/resetting XP and makes replay/fresh-session inspection simpler. |
| 4 | `vit -> max_hp` is authoritative in v26. | HP already exists in the sim and UI, making this the safest first real stat effect. |
| 5 | Damage gets the first offensive derived-stat hook. | Gear already modifies damage, so this extends an existing resolver instead of inventing a new combat path. |
| 6 | Armor, crit, hit chance, attack speed, movement speed, and mana are computed/displayed but mostly deferred for gameplay. | Prevents one slice from becoming a full combat rebalance while preserving the target data shape. |
| 7 | Stat spending is an intent, not a REST-only character API. | Keeps the same authoritative realtime input/replay model used by gameplay actions. |
| 8 | Character progression has a session-start snapshot. | Preserves v22's replay boundary and prevents historical replay drift. |
| 9 | Client UI is custom or borrowed presentation only; no plugin-owned stat logic. | Preserves ADR-0001 server authority and shared-rules discipline. |

## 11. Open questions

| # | Question | Default if unanswered |
|---|----------|----------------------|
| Q-1 | Should `str` affect only melee damage or all weapon damage in v26? | Affect melee/fixed weapon damage first; leave ranged/spell scaling for later tuning if needed. |
| Q-2 | Should `magic` affect anything authoritative before spells exist? | No. Display `max_mana`; reserve magic damage for the spell slice. |
| Q-3 | Should XP rewards come from all monsters or only dungeon monsters? | Dungeon monsters positive; static training/lab monsters may be `0` unless a scenario needs level-up. |
| Q-4 | Should stat allocation allow spending more than one point per click? | Protocol supports `points`, UI starts with one-point `+` buttons. |
| Q-5 | Should character summaries in the main menu show level? | Defer unless cheap after protocol/store changes; the in-game panel is required. |
| Q-6 | Should death block stat allocation? | Default: reject allocation while dead to keep the same dead-player gameplay intent posture. |
| Q-7 | Should the XP curve be finite or generated? | Use a small explicit table with `level_cap` for v26; add generated/infinite curve later when balancing begins. |
