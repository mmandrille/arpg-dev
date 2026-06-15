# v31 Plan — Combat Stat Effects and Feedback

Status: Implemented
Goal: Make character and monster combat stats affect authoritative combat, then expose readable combat feedback in Godot.
Architecture: Shared JSON owns combat constants, stat aggregation rules, protocol schemas, and golden fixtures. The Go sim owns every hit, miss, crit, block, mitigation, HP, death, loot, and XP outcome with deterministic seeded RNG. Godot only renders server events, settings, and stat breakdown presentation; it does not own combat math. Python bots prove the protocol path, reconnect, `/state`, replay, and client presentation smoke.
Tech stack: Shared JSON schemas/rules/goldens, Go `server/internal/game`, protocol v1 JSON, Godot 4 GDScript client, Python protocol/client bots, Make CI.

## Baseline and shortcut decision

Baseline is v30 `monster-rarity-and-loot-scaling` complete on `main`. This slice reuses v26 character progression/derived stats, v28 paper-doll equipment and rolled `armor` / `block_percent` / `max_hp`, v12 projectile impact, v21 monster proactive attacks, and v30 monster rarity scaling.
Implementation defaults from the spec:

- Flat armor mitigation: `final_damage = max(1, raw_or_crit_damage - effective_armor)` for successful non-blocked hits.
- Block fully negates damage and is capped at `75%` effective chance.
- Crits multiply raw damage before armor mitigation.
- Misses and blocks do not trigger retaliation, death, loot, or XP.
- Extend protocol v1 schemas in place with optional event/stat-breakdown fields unless implementation proves a version bump is unavoidable.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/combat.v0.schema.json` | Combat constants, block cap, damage floor schema |
| Modify | `shared/rules/combat.v0.json` | First-pass hit/crit/block/mitigation constants |
| Modify | `shared/rules/character_progression.v0.schema.json` | Effective stat and breakdown contract if needed |
| Modify | `shared/rules/character_progression.v0.json` | First-pass formula tuning |
| Modify | `shared/rules/item_templates.v0.schema.json` | Validate combat stat roll keys |
| Modify | `shared/rules/item_templates.v0.json` | Ensure gear rolls aggregate into combat stats |
| Modify | `shared/rules/monsters.v0.schema.json` | Monster hit/crit/armor/block stats |
| Modify | `shared/rules/monsters.v0.json` | Monster combat stat defaults and lab targets |
| Modify | `shared/rules/worlds.v0.json` | Add `combat_stat_lab` if protocol bot needs deterministic lab entities |
| Modify | `shared/rules/worlds.v0.schema.json` | Only if lab needs new world fields |
| Modify | `shared/protocol/session_snapshot.v1.schema.json` | Effective stat breakdowns in character progression |
| Modify | `shared/protocol/state_delta.v1.schema.json` | Combat outcome event metadata and progression updates |
| Modify | `shared/protocol/examples/session_snapshot.json` | Example effective stats/breakdowns |
| Modify | `shared/protocol/examples/state_delta.json` | Example miss/block/crit/mitigated events |
| Create | `shared/golden/combat_stat_effects.json` | Pinned combat stat fixture |
| Create | `shared/golden/combat_stat_effects.v0.schema.json` | Fixture schema |
| Modify | `tools/validate_shared.py` | Validate caps, stat keys, schemas, golden drift |
| Modify | `server/internal/game/rules.go` | Parse and validate combat constants, monster stats |
| Modify | `server/internal/game/types.go` | Event metadata, stat breakdown protocol structs |
| Modify | `server/internal/game/sim.go` | Effective stat aggregation and unified combat resolution |
| Modify | `server/internal/game/game_test.go` | Deterministic combat stat tests and golden checks |
| Modify | `server/internal/replay/replay.go` | Replay parity for updated event/snapshot shapes if needed |
| Modify | `server/internal/http` | `/state` parity if snapshot/event encoding needs updates |
| Modify | `tools/bot/run.py` | Combat event assertions and stat breakdown assertions |
| Create | `tools/bot/scenarios/22_combat_stat_effects.json` | Protocol end-to-end proof |
| Modify | `client/scripts/main.gd` | Render floating combat text from enriched events and settings |
| Modify | `client/scripts/damage_number.gd` | Crit/miss/block display variants |
| Modify | `client/scripts/client_settings.gd` | Persist `floating_combat_text` |
| Modify | `client/scripts/settings_panel.gd` | Toggle setting in menu UI |
| Modify | `client/scripts/character_stats_panel.gd` | Effective stat hover breakdowns |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client-bot assertions/actions for feedback/settings/breakdowns |
| Modify | `client/scripts/bot_controller.gd` | Client-bot actions if new settings/stat hover commands are needed |
| Modify | `client/tests/test_golden.gd` | Shared formula/effective stat golden checks |
| Modify | `client/tests/test_client_bot.gd` | Scenario validation/settings helper tests |
| Create | `tools/bot/scenarios/client/11_combat_feedback.json` | Client presentation proof |
| Modify | `PROGRESS.md` | Lifecycle update when slice ships |

## Task 1 — Shared Combat Contracts and Lab Data

Files:
- Modify: `shared/rules/combat.v0.schema.json`
- Modify: `shared/rules/combat.v0.json`
- Modify: `shared/rules/character_progression.v0.schema.json`
- Modify: `shared/rules/character_progression.v0.json`
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/rules/worlds.v0.schema.json` if needed
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add explicit combat constants for minimum non-blocked damage, block cap percent, and any first-pass default hit/crit behavior to `combat.v0.json` and schema.
```bash
make validate-shared
```

- [x] Step 1.2: Add monster combat stat fields to monster schema/rules: `hit_chance`, `crit_chance`, `crit_damage`, `armor`, and `block_percent`, with safe defaults for legacy monsters.
```bash
make validate-shared
```

- [x] Step 1.3: Validate existing item roll keys used by v31 combat aggregation: `armor`, `block_percent`, `max_hp`, `damage_min`, and `damage_max`. Add no new affix grammar.
```bash
make validate-shared
```

- [x] Step 1.4: Add a deterministic `combat_stat_lab` world if needed, with simple preset monsters for miss, crit, armor floor, block, monster crit, and monster block proofs. Keep positions reachable by current bot actions.
```bash
make validate-shared
```

- [x] Step 1.5: Extend `tools/validate_shared.py` to reject invalid chances, negative armor, block caps above `75`, invalid combat stat keys, and unreachable lab references.
```bash
make validate-shared
```

## Task 2 — Protocol Schemas and Golden Fixtures

Files:
- Modify: `shared/protocol/session_snapshot.v1.schema.json`
- Modify: `shared/protocol/state_delta.v1.schema.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Create: `shared/golden/combat_stat_effects.json`
- Create: `shared/golden/combat_stat_effects.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 2.1: Extend combat events with optional metadata: `source_entity_id`, `target_entity_id` or clarified `entity_id`, `outcome`, `raw_damage`, `mitigated_damage`, `blocked`, and `critical`.
```bash
make validate-shared
```

- [x] Step 2.2: Extend `character_progression.derived_stats` or adjacent progression view with effective stat breakdowns that include final value, uncapped value, cap, and source rows.
```bash
make validate-shared
```

- [x] Step 2.3: Update protocol examples to include one normal damage, one crit, one miss, one block, and one capped block stat breakdown.
```bash
make validate-shared
```

- [x] Step 2.4: Create `combat_stat_effects` golden fixture for player miss, player crit, monster armor floor, player armor floor, player block, 75% block cap, monster crit, monster block, and projectile impact.
```bash
make validate-shared
```

- [x] Step 2.5: Update shared validation to schema-check the new golden and verify expected constants match shared rules.
```bash
make validate-shared
```

## Task 3 — Go Rules Parsing and Effective Stat Aggregation

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Parse and validate combat constants and monster combat stats in `rules.go`, preserving defaults for existing worlds.
```bash
cd server && go test ./internal/game/... -run TestLoadRules -count=1
```

- [x] Step 3.2: Add Go protocol structs for combat event metadata and effective stat breakdown source rows.
```bash
cd server && go test ./internal/game/... -run Test -count=1
```

- [x] Step 3.3: Replace character-only `DerivedStatsView` usage with player effective stats that aggregate derived formulas plus equipped base/rolled item stats.
```bash
cd server && go test ./internal/game/... -run 'Progression|Equipment|FullEquipment|Strength' -count=1
```

- [x] Step 3.4: Add monster effective stat aggregation from monster rules plus existing rarity HP/damage scaling, leaving rarity hit/crit/armor/block scaling data-driven and optional.
```bash
cd server && go test ./internal/game/... -run 'Monster|Rarity' -count=1
```

- [x] Step 3.5: Ensure effective stat breakdowns include source rows for character formula contribution, equipped base stats, rolled stats, and caps/clamps.
```bash
cd server && go test ./internal/game/... -run 'Progression|CombatStat' -count=1
```

## Task 4 — Unified Deterministic Combat Resolution

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/replay.go` if event/snapshot replay comparisons need updates

- [x] Step 4.1: Introduce one combat resolution helper for attacker stats, defender stats, damage range, hit roll, block roll, raw damage, crit roll, armor mitigation, min damage, and event metadata.
```bash
cd server && go test ./internal/game/... -run 'Damage|CombatStat' -count=1
```

- [x] Step 4.2: Route player melee attacks through the helper, preserving reach checks, death, loot, XP, and retaliation semantics.
```bash
cd server && go test ./internal/game/... -run 'Attack|Damage|Retaliation|Experience' -count=1
```

- [x] Step 4.3: Route projectile impact through the same helper at impact time, preserving projectile blocked/expired behavior.
```bash
cd server && go test ./internal/game/... -run 'Projectile|Ranged|CombatStat' -count=1
```

- [x] Step 4.4: Route monster proactive attacks and retaliation through the helper, including player armor, player block, monster hit, and monster crit.
```bash
cd server && go test ./internal/game/... -run 'DungeonMonster|Retaliation|PlayerDamage|CombatStat' -count=1
```

- [x] Step 4.5: Enforce semantics: misses and blocks do not trigger retaliation, death, loot, or XP; non-blocked successful hits deal at least `1`; block is the only `0` damage outcome.
```bash
cd server && go test ./internal/game/... -run 'CombatStat|Block|Crit|Miss|MinimumDamage' -count=1
```

- [x] Step 4.6: Pin RNG draw order in Go tests and update older damage/retaliation/projectile goldens intentionally if draw order changes.
```bash
cd server && go test ./internal/game/... -run 'Golden|Replay|CombatStat' -count=1
```

## Task 5 — Go Golden, Replay, and State Parity

Files:
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/replay.go` if needed
- Modify: `server/internal/http` if needed
- Modify: existing `shared/golden/*.json` if old fixtures intentionally drift

- [x] Step 5.1: Add Go golden checks for all `combat_stat_effects.json` cases.
```bash
cd server && go test ./internal/game/... -run TestCombatStatEffectsGolden -count=1
```

- [x] Step 5.2: Add focused tests for effective player stat aggregation from equipped base stats and rolled stats.
```bash
cd server && go test ./internal/game/... -run 'Effective.*Stat|Equipment.*Stat|Progression' -count=1
```

- [x] Step 5.3: Add focused tests for monster hit/crit/armor/block stats and generated rarity compatibility.
```bash
cd server && go test ./internal/game/... -run 'Monster.*CombatStat|Rarity' -count=1
```

- [x] Step 5.4: Verify replay reconstruction and `/state` include the enriched event/progression shape without losing determinism.
```bash
cd server && go test ./internal/replay/... ./internal/http/... -count=1
```

- [x] Step 5.5: Run the complete Go suite after server changes settle.
```bash
make test-go
```

## Task 6 — Protocol Bot Scenario

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/22_combat_stat_effects.json`
- Modify: `tools/bot/test_protocol.py` if helper validation needs unit coverage

- [x] Step 6.1: Add bot assertions for combat event metadata: outcome, damage, raw damage, mitigated damage, critical, blocked, source id, and target id.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 6.2: Add bot assertions for effective stat values and breakdown source rows in `character_progression`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 6.3: Create `22_combat_stat_effects.json` to prove miss, crit, armor minimum damage, player block, monster crit, monster block, projectile impact, `/state`, reconnect, replay, and character persistence.
```bash
make bot
```

- [x] Step 6.4: Run all protocol scenarios and fix any intentional assertion drift from combat RNG changes.
```bash
make bot
```

## Task 7 — Godot Shared Golden and Floating Combat Text

Files:
- Modify: `client/tests/test_golden.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/damage_number.gd`
- Modify: `client/tests/test_client_bot.gd` or focused client tests if helpers are extracted

- [x] Step 7.1: Add Godot golden checks for effective stat formula/display-equivalence parts of `combat_stat_effects.json`.
```bash
make client-unit
```

- [x] Step 7.2: Teach `main.gd` to choose floating text by authoritative event outcome: normal damage, crit, miss, block, player damage, and monster damage.
```bash
make client-unit
```

- [x] Step 7.3: Extend `damage_number.gd` with display variants for crit size/tilt, miss text, and block text. Keep stable dimensions and avoid gameplay logic.
```bash
make client-unit
```

- [x] Step 7.4: Add helper tests for formatting/variant selection if the logic is extracted from `main.gd`.
```bash
make client-unit
```

## Task 8 — Godot Settings Toggle

Files:
- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 8.1: Add `floating_combat_text: true` to local settings load/save with backward-compatible defaults for missing existing settings files.
```bash
make client-unit
```

- [x] Step 8.2: Add a settings panel toggle and wire it through `main.gd` so events are processed but floating text is suppressed when disabled.
```bash
make client-unit
```

- [x] Step 8.3: Add client-bot action/assertion support for toggling and observing the setting.
```bash
make client-unit
```

## Task 9 — Godot Stat Breakdown Hover

Files:
- Modify: `client/scripts/character_stats_panel.gd`
- Modify: `client/scripts/main.gd` if progression view plumbing changes
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 9.1: Parse server-provided effective stat breakdowns from `character_progression`.
```bash
make client-unit
```

- [x] Step 9.2: Add hover tooltip/panel rows for damage, armor, block, and max HP, including source labels and cap/clamp notes.
```bash
make client-unit
```

- [x] Step 9.3: Add client-bot support or focused unit coverage for asserting one breakdown with equipment contribution and one capped block breakdown.
```bash
make client-unit
```

## Task 10 — Client Bot Scenario

Files:
- Create: `tools/bot/scenarios/client/11_combat_feedback.json`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 10.1: Add client scenario steps to reach the combat stat lab and trigger a normal hit, crit, miss, and block using server events.
```bash
make client-unit
```

- [x] Step 10.2: Assert floating combat text appears for normal/crit/block/miss when enabled.
```bash
make bot-client
```

- [x] Step 10.3: Toggle floating combat text off in settings and assert new combat events do not spawn floating text.
```bash
make bot-client
```

- [x] Step 10.4: Assert the stats panel exposes at least one effective stat breakdown with equipment contribution.
```bash
make bot-client
```

## Task 11 — Regression Updates

Files:
- Modify: existing bot scenarios under `tools/bot/scenarios/`
- Modify: existing client scenarios under `tools/bot/scenarios/client/`
- Modify: existing golden fixtures under `shared/golden/`
- Modify: `client/scripts/smoke.gd` if event assertions assume old damage-only shape

- [x] Step 11.1: Update existing protocol bot assertions that assume exact old damage values, especially character leveling, dungeon monster attack, rarity scaling, and full equipment scenarios.
```bash
make bot
```

- [x] Step 11.2: Update existing client smoke/bot assertions that expect only numeric damage popups.
```bash
make client-smoke
```

- [x] Step 11.3: Re-run shared, Go, bot, and client narrow suites after regression updates.
```bash
make validate-shared
make test-go
make bot
make client-unit
```

## Task 12 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v31_spec-combat-stat-effects-and-feedback.md` only if implementation resolves details differently
- Modify: `docs/plans/v31_2026-06-07-combat-stat-effects-and-feedback.md` if tasks are materially revised during execution

- [x] Step 12.1: Update the v31 spec status from Draft to Implemented only after implementation and CI are complete.
```bash
rg -n "Status:|v31|combat-stat-effects" docs/specs/v31_spec-combat-stat-effects-and-feedback.md PROGRESS.md
```

- [x] Step 12.2: Add v31 to the `PROGRESS.md` lifecycle table, numbering note, current status, summary, scenario catalog, and recently closed/open deferred sections.
```bash
rg -n "v31|combat_stat|Combat" PROGRESS.md
```

- [x] Step 12.3: Run final local CI.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `make client-unit`
- [x] `make bot`
- [x] `make bot-client`
- [x] `make client-smoke`
- [x] `make ci`

## Deferred scope

- Attack-speed gameplay and animation-speed scaling remain deferred.
- Movement-speed gameplay and pathing/prediction speed changes remain deferred.
- Dodge, resistances, status effects, lifesteal, thorns, spell systems, mana consumers, and affix grammar remain deferred.
- Full item comparison UI and polished production combat VFX/audio remain deferred.
- Enemy equipment inventories and player-facing monster inspection panels remain deferred.
