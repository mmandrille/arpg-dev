# v35 Plan — Boss Floor Gate

Status: Implemented; `make ci` green
Goal: Add the first skill-based boss floor at dungeon level `-5`, with a compact `30 x 30` floor, telegraphed damage, disabled/locked exits until boss death, boss/rarity scale presentation, and bot/replay proof.
Architecture: Shared JSON owns boss templates, boss patterns, compact boss-floor generation, and monster rarity visual scale. The Go sim owns boss generation, phase timing, hit predicates, locked stairs/teleporter exits, and replay determinism. The Godot client renders telegraphs, scale, tint, and disabled/ready exits only from server/rules metadata; it never performs boss hit detection. The Python bot proves the vertical slice through the same protocol path as play.
Tech stack: Shared JSON schemas/goldens, Go authoritative sim/replay, protocol JSON schemas v2 or coordinated bump if needed, Godot 4 GDScript presentation, Python protocol bot.

## Baseline and shortcut decision

Baseline is v34 `model-reaction-polish`: local players, remote players, and monsters already share hit/death reaction handling; remote players reuse the humanoid character model; generated monster rarity already reaches snapshots/deltas and client tinting. v35 reuses v18 dungeon levels/stairs, v25 chest interaction, v30 monster rarity metadata, v31 combat event metadata, and v33/v34 entity presentation paths.

Godot shortcut decision for this client presentation work:

| Candidate | Decision | Reason |
|-----------|----------|--------|
| Existing Godot material/tint tweens and primitive meshes | Borrow/reuse | A contact telegraph can use boss tint charging to full red; spatial telegraphs can use in-repo geometry/materials. |
| Existing humanoid/player model path | Borrow/reuse | The boss intentionally reuses this model at `2.0x` scale with special colors. |
| Existing `ModelReactionController` / animation controller path | Borrow/reuse | Boss hit/death can use the v34 character-like reaction path. |
| Built-in `AnimationTree` | Reject for v35 | No new skeletal animation state is needed. |
| LimboAI / behavior-tree plugin | Reject | Boss timing is authoritative Go sim state, not client AI. |
| New boss asset pack | Reject | Spec requires model reuse, not new production art. |
| New telegraph asset pack | Reject for first pass | Material/tint charge or code-created primitive/decal is enough and easier to test headlessly. |

No branch should be created; work stays on the current checkout.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Boss-floor config, compact `30 x 30` size, and rarity visual scale. |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate boss-floor config, size, and rarity visual scale. |
| Create | `shared/rules/boss_templates.v0.json` | First boss template, visual metadata, pattern deck, loot hooks. |
| Create | `shared/rules/boss_templates.v0.schema.json` | Boss template schema. |
| Create | `shared/rules/boss_patterns.v0.json` | First telegraphed attack pattern. |
| Create | `shared/rules/boss_patterns.v0.schema.json` | Pattern phase, timing, telegraph/hit-shape, and damage schema. |
| Modify | `shared/rules/interactables.v0.json` | Locked/ready down-stair and teleporter state support if needed. |
| Modify | `shared/rules/interactables.v0.schema.json` | Interactable state validation. |
| Modify | `shared/rules/loot_tables.v0.json` | Boss chest/drop table hooks if needed. |
| Modify | `shared/protocol/state_delta.v2.schema.json` | Boss events, stair unlock/reject events, visual scale/boss metadata. |
| Modify | `shared/protocol/session_snapshot.v2.schema.json` | Visible boss phase progress and visual metadata. |
| Modify | `shared/protocol/messages.v2.schema.json` | Envelope examples/types if schema references require update. |
| Modify | `shared/protocol/examples/state_delta.json` | Boss phase and locked-exit examples. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Boss visual/phase metadata example if snapshot schema changes. |
| Create | `shared/golden/boss_floor_-5.json` | Layout, lock, unlock, and boss visual fixture. |
| Create | `shared/golden/boss_floor_-5.v0.schema.json` | Boss-floor fixture schema. |
| Create | `shared/golden/boss_pattern_timeline.json` | Phase timing and dodge/no-damage fixture. |
| Create | `shared/golden/boss_pattern_timeline.v0.schema.json` | Boss-pattern timeline fixture schema. |
| Modify | `tools/validate_shared.py` | Boss schema validation, telegraph guarantee, visual scale checks, golden drift checks. |
| Modify | `server/internal/game/rules.go` | Parse and validate boss templates/patterns and rarity visual scale. |
| Modify | `server/internal/game/dungeon_gen.go` | Boss-floor detection, compact size, placement, chest, locked down stairs/teleporter, boss spawn. |
| Modify | `server/internal/game/types.go` | Boss phase/visual metadata and event view types. |
| Modify | `server/internal/game/sim.go` | Boss phase state machine, active hit predicates, locked exit gate/unlock. |
| Modify | `server/internal/game/game_test.go` | Generation, pattern timing, no-inevitable-damage, unlock, scale tests. |
| Modify | `server/internal/replay/*` | Replay parity if event/snapshot shapes need explicit handling. |
| Modify | `server/internal/http/*_test.go` | `/state` parity if boss state is exposed through inspection. |
| Modify | `client/scripts/main.gd` | Render boss/rarity scale, boss humanoid visual, telegraph cues/zones, locked exit feedback. |
| Modify | `client/tests/test_golden.gd` | Cross-check boss pattern/visual scale data if client consumes shared rules. |
| Modify | `client/tests/*` | Focused presentation tests if helpers are extracted. |
| Modify | `tools/bot/run.py` | Boss-aware assertions, timed dodge helpers, locked-exit reject support. |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for new bot assertions/helpers. |
| Create | `tools/bot/scenarios/24_boss_floor_gate.json` | End-to-end protocol proof. |
| Create/Modify | `tools/bot/scenarios/client/13_boss_telegraph.json` | Optional Godot client proof if reliable. |
| Modify | `docs/PROGRESS.md` | Lifecycle update when slice ships. |

## Task 1 — Shared rules and schemas

Files:
- Create: `shared/rules/boss_templates.v0.json`
- Create: `shared/rules/boss_templates.v0.schema.json`
- Create: `shared/rules/boss_patterns.v0.json`
- Create: `shared/rules/boss_patterns.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/interactables.v0.schema.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `boss_floor` config to dungeon generation rules: cadence `5`, required first level `-5`, compact `30 x 30` size, boss chest placement constraints, boss spawn constraints, locked `stairs_down`, disabled/locked boss-floor teleporter, and reduced/explicit boss-floor trash tuning.
```bash
make validate-shared
```

- [x] Step 1.2: Add `visual_scale` to generated monster rarity data, with `common: 1.0`, `champion: 1.25`, `unique: 1.5`, and a deliberate value for `rare` from the implementation decision.
```bash
make validate-shared
```

- [x] Step 1.3: Add the first boss template, reusing `dungeon_mob` or a small derived boss def as the base, with visual metadata `{model: current_humanoid_player, color, scale: 2.0}` and a single-pattern deck.
```bash
make validate-shared
```

- [x] Step 1.4: Add the first boss pattern, likely `charged_melee`, with telegraph, active, recovery, cooldown, body-color charge to full red, melee/contact hit shape, minimum telegraph duration, and damage range. Keep schema open for later spatial circles/lines/cones.
```bash
make validate-shared
```

- [x] Step 1.5: Extend shared validation so damaging phases require a prior telegraph phase, active hit predicates match the announced telegraph data, durations are positive, and telegraph duration is at least the configured floor.
```bash
make validate-shared
```

- [x] Step 1.6: Add locked/ready or disabled/ready state allowance for `stairs_down` and boss-floor teleporters only if the current interactable schema cannot already represent it cleanly.
```bash
make validate-shared
```

## Task 2 — Golden fixtures

Files:
- Create: `shared/golden/boss_floor_-5.json`
- Create: `shared/golden/boss_floor_-5.v0.schema.json`
- Create: `shared/golden/boss_pattern_timeline.json`
- Create: `shared/golden/boss_pattern_timeline.v0.schema.json`
- Modify: `tools/validate_shared.py`
- Modify: `server/internal/game/game_test.go`
- Modify: `client/tests/test_golden.gd` if client consumes the boss pattern fixture

- [x] Step 2.1: Create `boss_floor_-5` fixture covering classification, compact `30 x 30` footprint, one chest, one boss, locked down stairs, disabled/locked teleporter, boss visual scale/color/model, and unlock after boss kill.
```bash
make validate-shared
```

- [x] Step 2.2: Create `boss_pattern_timeline` fixture covering phase order, boundary ticks, telegraph duration, active duration, contact/area hit predicate semantics, and no damage when the player breaks contact or exits the announced danger before active.
```bash
make validate-shared
```

- [x] Step 2.3: Wire fixture validation into `tools/validate_shared.py` and add Go test loaders for both fixtures.
```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestBoss'
```

## Task 3 — Protocol contracts

Files:
- Modify: `shared/protocol/state_delta.v2.schema.json`
- Modify: `shared/protocol/session_snapshot.v2.schema.json`
- Modify: `shared/protocol/messages.v2.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `server/internal/realtime/protocol.go`
- Modify: `client/scripts/net_client.gd` only if parsing assumptions need adjustment

- [x] Step 3.1: Decide in implementation whether v35 extends current v2 schemas additively or requires a coordinated protocol bump; prefer additive v2 if current consumers tolerate optional fields.
```bash
make validate-shared
```

- [x] Step 3.2: Add event schema support for `boss_phase_started`, `boss_phase_ended`, telegraph data (`body_color_charge` or spatial zone), hit-shape data, and explicit locked-exit rejection/unlock feedback (`descend_blocked`, `teleport_blocked`, `intent_rejected` detail, or `interactable_state_changed`).
```bash
make validate-shared
```

- [x] Step 3.3: Add entity/snapshot metadata for `is_boss`, `boss_template_id`, `visual_model`, `visual_scale`, optional `visual_tint`, and current boss phase progress if the implementation chooses snapshot mid-phase recovery.
```bash
make validate-shared
```

- [x] Step 3.4: Update protocol examples so schema validation and downstream parser tests have concrete boss phase, locked-stair, and disabled/locked teleporter payloads.
```bash
make validate-shared
```

## Task 4 — Go rules loading

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 4.1: Add Go structs and loader validation for boss templates and boss patterns.
```bash
cd server && go test ./internal/game/... -run 'TestLoadRules|TestBoss'
```

- [x] Step 4.2: Parse monster rarity visual scale and validate the spec-required values for `champion`, `unique`, and boss `2.0x` template scale.
```bash
cd server && go test ./internal/game/... -run 'TestMonsterRarity|TestLoadRules|TestBoss'
```

- [x] Step 4.3: Ensure boss templates reference valid base monster defs, pattern IDs, loot tables if used, and visual metadata.
```bash
cd server && go test ./internal/game/... -run 'TestLoadRules|TestBoss'
```

## Task 5 — Boss floor generation

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 5.1: Add deterministic `isBossFloor(levelNum)` detection using `levelNum < 0 && abs(levelNum)%5 == 0`.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloor'
```

- [x] Step 5.2: Add boss-floor generation path for `-5`: compact `30 x 30` footprint, up stairs, guaranteed pre-boss chest, one boss spawn, locked down stairs, disabled/locked teleporter, optional reduced trash, and no accidental random guarded chest duplication.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloor'
```

- [x] Step 5.3: Use labeled deterministic RNG streams for boss template/placement if needed, keeping geometry/chest/rarity streams stable and documented in tests.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloorDeterminism'
```

- [x] Step 5.4: Spawn boss entity with boss metadata, humanoid visual model metadata, `2.0x` visual scale, special tint, scaled HP/damage, and existing monster combat compatibility.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloor|TestBossVisualMetadata'
```

- [x] Step 5.5: Assert `boss_floor_-5` golden layout semantically: compact `30 x 30` footprint, one chest, one boss, one locked down stair, disabled/locked teleporter, reachable/ordered placement, and deterministic output for pinned seed.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloorGolden'
```

## Task 6 — Boss phase state machine and hit predicates

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 6.1: Add per-boss phase state: pattern id, phase index, phase kind, started tick, remaining ticks, announced telegraph data, hit predicate, cooldown, and deterministic pattern choice.
```bash
cd server && go test ./internal/game/... -run 'TestBossPattern'
```

- [x] Step 6.2: Emit `boss_phase_started` and `boss_phase_ended` events at exact phase boundaries.
```bash
cd server && go test ./internal/game/... -run 'TestBossPatternTimeline'
```

- [x] Step 6.3: Apply boss damage only during active phase, only to players satisfying the announced hit predicate, and never during telegraph/recovery/cooldown. For `charged_melee`, this means only players still in boss melee/contact range when active starts can be hit.
```bash
cd server && go test ./internal/game/... -run 'TestBossPatternDamage'
```

- [x] Step 6.4: Prove dodge semantics from the golden: player starts in contact/risk during telegraph, breaks contact before active phase, and takes no damage from that attack.
```bash
cd server && go test ./internal/game/... -run 'TestBossPatternTimelineGolden'
```

- [x] Step 6.5: Keep boss phase logic deterministic: no wall-clock time, no unseeded randomness, no map-order-dependent target selection.
```bash
cd server && go test ./internal/game/... -run 'TestBossPatternDeterminism'
```

## Task 7 — Locked exits and boss unlock

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/http/*_test.go` if `/state` parity needs explicit coverage

- [x] Step 7.1: Make boss-floor `stairs_down` reject `descend_intent` while locked and boss-floor teleporter reject use while disabled/locked, both with explicit `boss_alive` reason or equivalent event detail.
```bash
cd server && go test ./internal/game/... -run 'TestBossLockedStairs'
```

- [x] Step 7.2: On boss death, transition the down stair and teleporter state to `ready` and emit observable state-change events.
```bash
cd server && go test ./internal/game/... -run 'TestBossUnlocksStairs'
```

- [x] Step 7.3: Allow descent to `-6` and teleporter use after unlock, and ensure level transition remains actor-scoped for co-op sessions.
```bash
cd server && go test ./internal/game/... -run 'TestBossUnlocksStairs|TestLevelTransition'
```

- [x] Step 7.4: Include locked/unlocked exit state and boss metadata in `/state`, snapshot, reconnect, and replay reconstruction.
```bash
go test ./server/internal/... 
```

## Task 8 — Replay and realtime integration

Files:
- Modify: `server/internal/realtime/*`
- Modify: `server/internal/replay/*`
- Modify: `server/internal/http/*`
- Modify: `server/internal/game/*_test.go`

- [x] Step 8.1: Ensure realtime snapshots/deltas include boss phase/visual metadata and boss events without breaking existing v2 clients.
```bash
cd server && go test ./internal/realtime/... ./internal/http/...
```

- [x] Step 8.2: Ensure replay reconstructs boss floor generation, boss phase events, locked-exit rejects, boss death, exit unlock, and descent/teleporter use.
```bash
cd server && go test ./internal/replay/... ./internal/game/...
```

- [x] Step 8.3: Add or update replay parity tests for a pinned boss-floor input sequence.
```bash
cd server && go test ./internal/replay/... -run 'TestBoss'
```

## Task 9 — Godot client presentation

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_golden.gd`
- Modify: `client/tests/*` if helper scripts are extracted
- Modify: `tools/bot/scenarios/client/13_boss_telegraph.json` if client bot proof is reliable

- [x] Step 9.1: Apply `visual_scale` to generated monster nodes from authoritative entity metadata or shared rarity rules: `champion 1.25x`, `unique 1.5x`, normal `1.0x`.
```bash
make client-unit
```

- [x] Step 9.2: Render boss entities using the current humanoid/player model path, `2.0x` scale, and special boss tint while preserving v34 hit/death reaction tint layering.
```bash
make client-unit
```

- [x] Step 9.3: Render a readable telegraph from `boss_phase_started`: for `body_color_charge`, ramp the boss tint toward full red during telegraph and restore/transition on phase end; for spatial telegraphs, render/remove the indicated zone.
```bash
make client-unit
```

- [x] Step 9.4: Render locked/ready down stairs and disabled/ready boss-floor teleporter distinctly enough for smoke/debug, or surface server rejection feedback through existing debug/UI pathways.
```bash
make client-unit
```

- [x] Step 9.5: Extend bot/debug presentation state so a headless client scenario can assert boss visual model, visual scale, rarity scale, and telegraph cue presence/color if reliable.
```bash
make client-unit
```

- [x] Step 9.6: Add `tools/bot/scenarios/client/13_boss_telegraph.json` only if it can reliably start, reach `-5`, observe boss scale/telegraph cue, and exit under headless automation; otherwise document that protocol bot plus client-unit covers the presentation contract. Decision: do not add a client scenario in v35; protocol bot plus `make client-unit` covers the presentation contract without adding a brittle headless route-to-`-5` visual test.
```bash
make client-smoke
```

## Task 10 — Bot scenario and assertions

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Create: `tools/bot/scenarios/24_boss_floor_gate.json`

- [x] Step 10.1: Add bot filters/assertions for boss entities, `is_boss`, visual model/scale metadata, compact boss-floor size, locked/ready stairs and teleporter state, and boss phase events.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -v
```

- [x] Step 10.2: Add `use_stair`/`action_entity` paths that can expect locked-stair and disabled-teleporter rejection with reason `boss_alive`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -v
```

- [x] Step 10.3: Add a timed boss dodge helper or compose existing `wait_for_event`, `move_to_position`, `wait_ticks`, and `assert_player_hp` steps so the scenario can prove telegraph -> break contact or move out -> no damage. Decision: the no-inevitable-damage contract is covered by the shared golden and focused Go boss phase test; the protocol bot observes boss phase and proves the same live protocol path for gating/kill/unlock.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -v
```

- [x] Step 10.4: Create `24_boss_floor_gate.json`: descend to `-5`, assert compact `30 x 30` layout, open chest, verify locked stair and teleporter rejects, break contact or dodge one telegraph with no damage, kill boss, assert exit unlock, descend to `-6`.
```bash
make bot
```

- [x] Step 10.5: Ensure scenario assertions cover `/state`, reconnect resume, replay, and the semantic no-inevitable-damage contract.
```bash
make bot
```

## Task 11 — Integration hardening

Files:
- Modify: `server/internal/game/*_test.go`
- Modify: `server/internal/replay/*_test.go`
- Modify: `client/tests/*`
- Modify: `tools/bot/test_protocol.py`
- Modify: `shared/protocol/examples/*`

- [x] Step 11.1: Run focused shared, Go, Python, and client checks after the first full implementation pass.
```bash
make validate-shared
make test-go
.venv/bin/pytest tools/bot/test_protocol.py -v
make client-unit
```

- [x] Step 11.2: Run protocol bot and replay-heavy integration checks, then fix drift in event ordering or semantic assertions.
```bash
make bot
```

- [x] Step 11.3: Run client smoke if telegraph/scale presentation changed runtime scenes.
```bash
make client-smoke
```

## Task 12 — Lifecycle docs and CI

Files:
- Modify: `docs/PROGRESS.md`
- Modify: `docs/specs/v35_spec-boss-floor-gate.md` only if implementation forces a clarified as-built note
- Modify: `docs/plans/v35_2026-06-08-boss-floor-gate.md` if task scope changes during implementation

- [x] Step 12.1: Update `docs/PROGRESS.md` lifecycle table with v35, slice numbering note, current status, what v35 proved, and any newly deferred boss-floor gaps.
```bash
rg -n "v35|boss-floor-gate|Latest completed slice|Open gaps" docs/PROGRESS.md
```

- [x] Step 12.2: Record any as-built deviations from this plan in the plan or spec before finish.
```bash
git diff -- docs/specs/v35_spec-boss-floor-gate.md docs/plans/v35_2026-06-08-boss-floor-gate.md docs/PROGRESS.md
```

- [x] Step 12.3: Run final CI.
```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `make test-go`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -v`
- [x] `make client-unit`
- [x] `make bot scenario=boss_floor_gate`
- [x] `make client-smoke`
- [x] `make ci`

## Deferred scope

- Additional boss templates, pattern decks, full spatial shape implementations beyond schema readiness, enrage phases, summoned adds, elaborate boss arena geometry beyond the compact `30 x 30` generation footprint, co-op boss scaling, boss health bar UI, production boss art/VFX/audio, block/parry boss-zone interaction, durable boss kill/map snapshots, quest integration, and final balance pass.
- Visual scale remains presentation-only in v35; any future collision/reach scaling must be a separate explicit gameplay slice with server tests and bot proof.
