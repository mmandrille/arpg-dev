# v26 Plan - Character Stats and Leveling

Status: Complete - make ci green on 2026-06-07
Goal: Add persistent character XP, level, base stats, level-up stat points, a server-authoritative stat allocation intent, a character sheet, and an XP bar.
Architecture: Character progression is durable character-owned state with an immutable session-start snapshot for replay, following the v22 item/waypoint boundary. Shared JSON rules define default stats, XP thresholds, points per level, monster XP rewards, and bounded derived-stat formulas; Go owns all authoritative outcomes and GDScript mirrors formulas only for golden/display checks. The first gameplay effects are `vit -> max_hp` and a damage bonus path; armor, crit, attack speed, movement speed, and mana remain display-only in v26.
Tech stack: shared JSON schemas/goldens, Go authoritative sim/store/realtime/replay, protocol v1 schema updates, Python protocol bot scenario, Godot 4 GDScript UI and client bot scenario.

## Baseline and shortcut decision

v26 builds on v25 `treasure-classes-and-guarded-chests`, especially v22 character-scoped persistence, v23 rolled item payloads, v24 named character session start, and v25 dungeon reward loops. The current player HP model is a fixed `playerStartHP` in `server/internal/game/sim.go`, weapon damage already resolves through item rules/rolled stats, and protocol snapshots/deltas already expose player `hp`/`max_hp`, inventory, equipment, and recent events.

Godot shortcut adoption checklist:

- **Decision:** reject new plugin adoption for stat logic and this first character sheet.
- **Reason:** RPG/stat plugins conflict with the authoritative server boundary because they want to own formulas or local character state. The UI is compact enough to build from existing in-repo `Control` patterns.
- **Borrow:** use layout/style patterns from `inventory_panel.gd`, `consumable_bar.gd`, and the current waypoint panel construction. Revisit GLoot/Godot-Inventory only for future stash/inventory complexity, not for v26 stats.

Spec review notes resolved during planning:

- The spec's open questions are locked to their defaults: `str` affects melee/fixed weapon damage first; `magic` is display-only before spells; dungeon mobs get positive XP; UI spends one point per click; menu summaries are deferred; dead players cannot allocate stats; the XP curve is a small explicit table with `level_cap`.
- Protocol changes are additive to the existing v1 JSON schemas and coordinated client/server in the same slice.
- Bot proof is required because the slice changes gameplay, protocol, persistence, replay, and client UI.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Created | `docs/specs/v26_spec-character-stats-and-leveling.md` | Slice contract |
| Create | `shared/rules/character_progression.v0.schema.json` | Progression rule schema |
| Create | `shared/rules/character_progression.v0.json` | Default stats, XP curve, derived stat formulas |
| Modify | `shared/rules/monsters.v0.schema.json` | Add optional `xp_reward` |
| Modify | `shared/rules/monsters.v0.json` | Add positive dungeon mob XP and zero/default lab XP |
| Modify | `shared/protocol/messages.v1.schema.json` | Add `allocate_stat_intent` |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | Add `character_progression` payload shape |
| Modify | `shared/protocol/state_delta.v1.schema.json` | Add progression update change and progression events |
| Create | `shared/golden/character_progression.json` | Cross-language XP/stat/derived fixture |
| Modify | `tools/validate_shared.py` | Validate progression rules, XP rewards, and fixture drift |
| Create | `server/migrations/0005_character_progression_stats.sql` | Durable progression and session-start snapshot tables |
| Modify | `server/internal/store/models.go` | Progression models |
| Modify | `server/internal/store/interfaces.go` | Progression repository methods |
| Modify | `server/internal/store/repos.go` | Postgres progression implementation |
| Modify | `server/internal/store/store_test.go` | Persistence coverage |
| Modify | `server/internal/game/types.go` | Progression protocol views, events, changes, input payloads |
| Modify | `server/internal/game/rules.go` | Load progression rules and monster XP rewards |
| Modify | `server/internal/game/sim.go` | XP gain, level-up, stat allocation, max HP, damage hook |
| Modify | `server/internal/game/game_test.go` | Sim/golden/determinism tests |
| Modify | `server/internal/http/session.go` | Load/create progression at session start and snapshot it |
| Modify | `server/internal/http/ws.go` or realtime decode files | Decode `allocate_stat_intent` |
| Modify | `server/internal/realtime/runner.go` | Persist progression updates |
| Modify | `server/internal/replay/replay.go` | Replay from session-start progression snapshot |
| Modify | `server/internal/http/ws_test.go` | WebSocket/reconnect/progression tests |
| Modify | `server/internal/http/auth_session_test.go` | Fresh-session persistence/ownership tests |
| Modify | `client/scripts/main.gd` | Store progression state, parse deltas, send allocation intents |
| Create | `client/scripts/character_stats_panel.gd` | Left-side stats panel |
| Modify | `client/scripts/consumable_bar.gd` | Add or coordinate XP bar below hotbar |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client bot stat panel steps/assertions |
| Modify | `client/scripts/bot_controller.gd` | Dispatch stat panel actions and expose debug state |
| Modify | `client/tests/test_golden.gd` | Data-only progression golden checks |
| Modify | `client/tests/test_client_bot.gd` | Client bot scenario parser/validation tests |
| Modify | `tools/bot/run.py` | Protocol bot progression helpers/assertions |
| Create | `tools/bot/scenarios/18_character_stats_and_leveling.json` | End-to-end protocol proof |
| Create | `tools/bot/scenarios/client/09_character_stats_panel.json` | Godot client UI proof |
| Modify | `PROGRESS.md` | Lifecycle update when slice ships |

## Task 1 - Shared progression contracts

Files:
- Create: `shared/rules/character_progression.v0.schema.json`
- Create: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Create: `shared/golden/character_progression.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Define `character_progression.v0.schema.json` with `base_stats`, `points_per_level`, `level_cap`, explicit table XP curve, and bounded `linear` derived-stat formulas.
- [x] Step 1.2: Add `character_progression.v0.json` with defaults: `str=5`, `dex=5`, `vit=5`, `magic=5`, `points_per_level=5`, and a small level table tuned so the protocol bot can level within one scenario.
- [x] Step 1.3: Include derived stat keys for damage, armor, attack speed, hit chance, crit chance, crit damage, movement speed, hp, and mana.
- [x] Step 1.4: Extend monster schema and rules with `xp_reward`; set `dungeon_mob` positive and preserve lab/training monsters as `0` unless needed by a test fixture.
- [x] Step 1.5: Create `shared/golden/character_progression.json` pinning initial progression, one XP gain below threshold, one level-up granting 5 points, one multi-level gain, one `vit` allocation, and one damage-derived case.
- [x] Step 1.6: Extend `tools/validate_shared.py` to validate formula references, non-negative rewards, monotonic XP thresholds, positive `points_per_level`, level-cap/table consistency, and golden fixture drift.

```bash
make validate-shared
```

## Task 2 - Store and session-start progression snapshots

Files:
- Create: `server/migrations/0005_character_progression_stats.sql`
- Modify: `server/internal/store/models.go`
- Modify: `server/internal/store/interfaces.go`
- Modify: `server/internal/store/repos.go`
- Modify: `server/internal/store/store_test.go`

- [x] Step 2.1: Add `character_progression` table keyed by `character_id`, with `level`, `experience`, `unspent_stat_points`, and integer columns for `stat_str`, `stat_dex`, `stat_vit`, and `stat_magic`.
- [x] Step 2.2: Add `session_start_character_progression` keyed by `session_id` with the same progression fields and `created_at`.
- [x] Step 2.3: Add Go models for durable progression, session-start progression, and stat maps/views.
- [x] Step 2.4: Add repo methods to get-or-create character progression from shared defaults, update progression atomically, create the session-start snapshot, and load the session-start snapshot.
- [x] Step 2.5: Keep account/character ownership checks at the repo or caller boundary; another account must not load or mutate progression for a foreign character.
- [x] Step 2.6: Add store tests for initialization defaults, persistence update, session-start snapshot immutability, and ownership-scoped loads.

```bash
cd server && go test ./internal/store -run 'CharacterProgression|SessionStart'
```

## Task 3 - Protocol and Go progression views

Files:
- Modify: `shared/protocol/messages.v1.schema.json`
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`
- Modify: `server/internal/game/types.go`
- Modify: realtime input decode files under `server/internal/http/` or `server/internal/realtime/`

- [x] Step 3.1: Add `allocate_stat_intent` to `messages.v1.schema.json`, with payload `{ stat, points }`.
- [x] Step 3.2: Define reusable schema defs for `character_progression`, `base_stats`, and `derived_stats`; add `character_progression` to `session_snapshot` required fields.
- [x] Step 3.3: Add `character_progression_update` as a state delta change.
- [x] Step 3.4: Add `experience_gained`, `character_leveled`, and `stat_allocated` event fields with required payload validation.
- [x] Step 3.5: Add Go protocol structs for `CharacterProgressionView`, `BaseStats`, `DerivedStats`, and `AllocateStatIntent`.
- [x] Step 3.6: Decode and validate `allocate_stat_intent` in the same path as existing WebSocket gameplay intents.
- [x] Step 3.7: Ensure unknown stats, non-positive points, and malformed payloads reject cleanly without mutating Sim state.

```bash
make validate-shared
cd server && go test ./internal/http/... -run 'AllocateStat|Protocol|WebSocket'
```

## Task 4 - Go rules, Sim progression, HP, and damage

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/replay/replay.go`

- [x] Step 4.1: Load `character_progression.v0.json` into `Rules` and add typed formula evaluators for the bounded `linear` catalog.
- [x] Step 4.2: Add progression state to `Sim`, initialized from durable/session-start progression rather than hard-coded defaults.
- [x] Step 4.3: Replace hard-coded player `maxHP` initialization with derived `max_hp` from progression; preserve current HP behavior by starting fresh sessions at max HP.
- [x] Step 4.4: On monster death, award `xp_reward` exactly once, apply all crossed XP thresholds in order, grant `points_per_level` for each level gained, and append stable progression changes/events after the monster kill/drop events.
- [x] Step 4.5: Implement `allocate_stat_intent`: reject invalid/dead/overspend cases; on valid spend, increment the stat, decrement unspent points, recompute derived stats, and emit `character_progression_update` plus `stat_allocated`.
- [x] Step 4.6: For `vit` allocations, update player entity `max_hp` and current `hp` by the gained max HP amount, capped at the new max.
- [x] Step 4.7: Add the first damage hook: apply the derived damage bonus to melee/fixed weapon damage while preserving seeded damage-roll order.
- [x] Step 4.8: Keep armor, crit, hit chance changes, attack speed, movement speed, max mana, and magic damage display-only in this slice.
- [x] Step 4.9: Add Go tests for the golden fixture, XP gain once per kill, level-up and multi-level-up, valid/rejected stat allocation, `vit` HP behavior, damage bonus, reconnect snapshot shape, and replay determinism.

```bash
cd server && go test ./internal/game/... -run 'CharacterProgression|Experience|Level|Stat|Damage|HP'
```

## Task 5 - Session bootstrap, realtime persistence, and replay boundary

Files:
- Modify: `server/internal/http/session.go`
- Modify: `server/internal/realtime/hub.go`
- Modify: `server/internal/realtime/runner.go`
- Modify: `server/internal/replay/replay.go`
- Modify: `server/internal/http/ws_test.go`
- Modify: `server/internal/http/auth_session_test.go`

- [x] Step 5.1: During fresh session creation, load or initialize durable character progression before creating the session-start snapshot.
- [x] Step 5.2: Persist the immutable session-start progression snapshot alongside existing item/waypoint session-start snapshots.
- [x] Step 5.3: Feed fresh WebSocket attach from durable progression and same-session replay/reconnect from session-start progression plus recorded inputs.
- [x] Step 5.4: Persist progression mutations from `character_progression_update` changes or explicit Sim result fields in `runner.go`.
- [x] Step 5.5: Ensure `/state`, WebSocket `session_snapshot`, and replay timeline expose the same progression view.
- [x] Step 5.6: Add HTTP/WebSocket tests for cross-session progression persistence, selected-character isolation, reconnect, replay after later character changes, and `/state` parity.

```bash
cd server && go test ./internal/http/... ./internal/replay/... -run 'CharacterProgression|Experience|Replay|State|Session'
```

## Task 6 - Protocol bot scenario

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py` if helper parsing needs tests
- Create: `tools/bot/scenarios/18_character_stats_and_leveling.json`

- [x] Step 6.1: Add bot assertions for `character_progression.level`, `experience`, `unspent_stat_points`, base stats, derived stats, and player `max_hp`.
- [x] Step 6.2: Add an `allocate_stat` scenario step that sends `allocate_stat_intent`.
- [x] Step 6.3: Create scenario `18_character_stats_and_leveling.json` using a deterministic character/session and enough dungeon monster kills to cross the first XP threshold.
- [x] Step 6.4: Assert XP increases, level increments, unspent points increases by 5, and `vit` allocation decreases points and increases base `vit`.
- [x] Step 6.5: Assert player `max_hp` and progression `derived_stats.max_hp` increase after the `vit` allocation.
- [x] Step 6.6: Include a reject case for overspending or invalid stat where it can be done without destabilizing the scenario.
- [x] Step 6.7: Run `/state`, reconnect resume, replay verification, and fresh-session persistence assertions.

```bash
make bot
```

## Task 7 - Godot golden checks and progression UI

Files:
- Modify: `client/scripts/main.gd`
- Create: `client/scripts/character_stats_panel.gd`
- Modify: `client/scripts/consumable_bar.gd`
- Modify: `client/tests/test_golden.gd`
- Modify: `client/scripts/net_client.gd` if helper send wrappers are useful

- [x] Step 7.1: Load `character_progression.v0.json` in Godot test/display code and add `test_golden.gd` coverage for `shared/golden/character_progression.json`.
- [x] Step 7.2: Store `character_progression` in `main.gd`, update it from snapshots and `character_progression_update` deltas, and expose it in `get_bot_state()`.
- [x] Step 7.3: Add `CharacterStatsPanel` on the left side, toggled by `C`, with level, XP text, unspent points, base stat rows, `+` buttons, and compact derived stat rows.
- [x] Step 7.4: Send `allocate_stat_intent` from `+` buttons only; do not mutate local stats until the authoritative update arrives.
- [x] Step 7.5: Disable stat buttons when points are unavailable, player is dead, WebSocket is closed, or menus/pause/settings overlays block gameplay input.
- [x] Step 7.6: Add a compact XP bar below the bottom-center hotbar, reusing or coordinating with `consumable_bar.gd` positioning so supported window sizes do not overlap existing UI.
- [x] Step 7.7: Update health bar/player HP handling so `max_hp` changes from snapshots/deltas flow into existing UI.
- [x] Step 7.8: Keep visual replay and automation paths rendering progression updates without requiring user interaction.

```bash
make client-unit
make client-smoke
```

## Task 8 - Godot client bot stats panel proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/tests/test_client_bot.gd`
- Create: `tools/bot/scenarios/client/09_character_stats_panel.json`

- [x] Step 8.1: Add client bot state fields for `character_stats_panel_visible`, `character_progression`, XP bar progress/debug data, and stat button enabled states.
- [x] Step 8.2: Add client bot actions/assertions for pressing `C`, asserting panel visibility, clicking a stat `+` button, and asserting progression fields after server confirmation.
- [x] Step 8.3: Add parser/validation tests in `test_client_bot.gd` for all new client bot step types.
- [x] Step 8.4: Create `09_character_stats_panel.json`: start a menu/automation session, earn or load unspent points through the scenario path, open the panel, allocate `vit`, assert panel values and XP bar state, and ensure pause/menu input locking blocks allocation.
- [x] Step 8.5: Ensure existing client scenarios `01`-`08` keep passing with the new `C` key handling and XP bar mounted.

```bash
make client-unit
make db-up
make bot-client
```

## Task 9 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v26_spec-character-stats-and-leveling.md`
- Modify: `docs/plans/v26_2026-06-07-character-stats-and-leveling.md`

- [x] Step 9.1: When implementation ships, mark the v26 spec `Complete - make ci green on <date>`.
- [x] Step 9.2: Add v26 to the `PROGRESS.md` slice numbering note and lifecycle table.
- [x] Step 9.3: Add a concise "What v26 proved" section covering durable XP, level-up points, stat allocation, max HP/damage effects, character sheet, XP bar, and bot proofs.
- [x] Step 9.4: Update deferred gaps: passive skills, armor/crit/hit/attack speed gameplay, mana consumers, stat requirements, respec, class selection, main-menu character summaries, and deeper XP curve/balancing.
- [x] Step 9.5: Mark this plan complete only after final CI is green.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/store -run 'CharacterProgression|SessionStart'`
- [x] `cd server && go test ./internal/game/... -run 'CharacterProgression|Experience|Level|Stat|Damage|HP'`
- [x] `cd server && go test ./internal/http/... ./internal/replay/... -run 'CharacterProgression|Experience|Replay|State|Session'`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `make bot-client`
- [x] `make ci`
- [ ] `make bot-client`
- [ ] `make ci`

## Deferred scope

- Passive skill tree and passive stat/effect sources.
- Full armor mitigation, hit chance, crit chance, crit damage, attack-speed cooldowns, movement-speed authority, and mana-consuming skills.
- Spell system and magic damage authority.
- Stat requirements for items beyond existing level requirements.
- Class selection, respec/stat reset, character delete/rename, and menu character summaries.
- Infinite/generated XP curve and real balance pass beyond the v26 explicit table.
- Production character-sheet art/audio and external UI plugin adoption.
