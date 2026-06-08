# v30 Plan - Monster Rarity and Loot Scaling

Status: Implemented
Goal: Add deterministic generated-dungeon monster rarity that scales challenge, XP, loot depth, and client tint while keeping server authority.
Architecture: Store monster rarity as shared dungeon-generation data and roll it during deterministic level generation. Generated monster entities carry rarity, scaled HP/damage/XP, and a loot table selected from effective loot depth. Protocol v1 already has optional entity `rarity`, so the implementation should emit and test it for monsters without a schema version bump unless a concrete insufficiency appears. Godot remains presentation-only and tints the existing player/enemy models from server-provided rarity.
Tech stack: Shared JSON rules and golden fixtures, Go authoritative sim, protocol v1 examples/schemas, Godot client material tinting, Python protocol bot.

## Baseline and shortcut decision

Baseline is v29 `dungeon-equipment-drop-expansion`: generated dungeon monsters and guarded chests
already resolve depth-banded loot tables, generated monsters store their loot table at generation
time, item/template reachability is validated, and the bot proves real generated dungeon drops.

This slice reuses:

- v21 generated hostile `dungeon_mob` placement and proactive attacks.
- v23 rolled item payloads and item rarity metadata for loot/inventory.
- v25 treasure class rolls and replay-stable source loot.
- v26 XP award path and character progression deltas.
- v29 loot bands: effective depths `3+` continue routing to the existing `3+` band.

Godot plugin adoption: **reject for v30**. The slice only needs material/tint changes on existing
in-repo player and monster visuals. No UI framework, camera controller, asset pack, behavior tree,
or inventory plugin would reduce the core work, and gameplay authority must stay in Go.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Add generated monster rarity rules and validation shape. |
| Modify | `shared/rules/dungeon_generation.v0.json` | Configure rarity weights, colors, HP/damage/XP multipliers, loot offsets, and notes. |
| Create | `shared/golden/monster_rarity.v0.schema.json` | Schema for rarity roll/scaling/effective-depth fixture. |
| Create | `shared/golden/monster_rarity.json` | Pin rarity rules, representative generated outcomes, scaling, and level `-5` unique depth `8`. |
| Modify | `shared/protocol/examples/session_snapshot.json` | Show monster entity rarity in snapshot example. |
| Modify | `shared/protocol/examples/state_delta.json` | Show monster entity rarity in spawn/update example. |
| Modify | `tools/validate_shared.py` | Validate rarity ids, weights, colors, multipliers, offsets, and fixture drift. |
| Modify | `server/internal/game/rules.go` | Parse and validate monster rarity rules; expose lookup helpers. |
| Modify | `server/internal/game/dungeon_gen.go` | Roll rarity in stable generated monster order and resolve effective loot table. |
| Modify | `server/internal/game/sim.go` | Store monster rarity/scaled fields, emit entity rarity, apply scaled attack damage/XP. |
| Modify | `server/internal/game/types.go` | Add internal entity rarity fields only if needed; protocol `EntityView.Rarity` already exists. |
| Modify | `server/internal/game/game_test.go` | Add rarity generation, scaling, effective-depth, and replay determinism coverage. |
| Modify | `server/internal/http/ws_test.go` | Add `/state`/snapshot parity coverage if game tests do not cover HTTP entity rarity. |
| Modify | `server/internal/replay/replay_test.go` | Add replay timeline rarity coverage if game replay tests do not cover it. |
| Modify | `client/scripts/main.gd` | Apply green player tint and rarity tint for generated monsters from entity metadata. |
| Modify/Create | `client/tests/*` | Add tint helper tests if tint logic is extracted; otherwise rely on smoke. |
| Modify | `client/tests/test_golden.gd` | Add data-only rarity fixture checks if the fixture is consumed by Godot tests. |
| Modify | `tools/bot/run.py` | Add monster rarity assertions and effective loot-depth assertions if existing helpers are insufficient. |
| Create | `tools/bot/scenarios/21_monster_rarity_loot_scaling.json` | End-to-end generated dungeon rarity proof. |
| Modify | `docs/specs/v30_spec-monster-rarity-and-loot-scaling.md` | Update status/notes only if implementation decisions differ. |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v30 ships. |

## Task 1 - Shared rarity rules and validation

Files:

- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `monster_rarities` schema with required ids `common`, `champion`, `rare`, and `unique`.

```bash
make validate-shared
```

- [x] Step 1.2: Configure the first-pass rarity weights exactly as `common: 100`, `champion: 15`, `rare: 6`, and `unique: 3`.

```bash
make validate-shared
```

- [x] Step 1.3: Configure pastel tint colors, HP multipliers, damage multipliers, XP multipliers, and loot depth offsets `0`, `1`, `2`, and `3`.

```bash
make validate-shared
```

- [x] Step 1.4: Extend shared validation for stable lowercase ids, positive weights, hex colors, positive multipliers, non-negative offsets, and `common` offset `0`.

```bash
make validate-shared
```

- [x] Step 1.5: Document in `dungeon_generation.v0.json` that monster rarity is a v30 first-pass hook and does not imply unique item drops.

```bash
make validate-shared
```

## Task 2 - Golden fixtures and protocol examples

Files:

- Create: `shared/golden/monster_rarity.v0.schema.json`
- Create: `shared/golden/monster_rarity.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `tools/validate_shared.py`
- Modify: `client/tests/test_golden.gd` only if Godot consumes the fixture

- [x] Step 2.1: Add `monster_rarity.json` with cases for each rarity tier, including weights, colors, multipliers, offsets, and expected scaled values from `dungeon_mob`.

```bash
make validate-shared
```

- [x] Step 2.2: Pin the level `-5` + `unique` case so effective loot depth is `8` and currently routes to the v29 `3+` monster loot band.

```bash
make validate-shared
```

- [x] Step 2.3: Add one or more pinned generated-level cases that prove seeded rarity roll order for generated monsters.

```bash
make validate-shared
```

- [x] Step 2.4: Update protocol examples to show `rarity` on a monster entity, while keeping item rarity examples intact.

```bash
make validate-shared
```

- [x] Step 2.5: If the new fixture is parsed by Godot tests, add `client/tests/test_golden.gd` data-only assertions for rarity ids/colors/offsets.

```bash
make client-unit
```

## Task 3 - Go rules loader and dungeon generation

Files:

- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Add typed Go structures for monster rarity rules and validate the same invariants enforced by `tools/validate_shared.py`.

```bash
cd server && go test ./internal/game/... -run 'Rules|Rarity'
```

- [x] Step 3.2: Add deterministic helper methods for rarity lookup, weighted rarity roll, scaling, and effective loot depth.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Dungeon'
```

- [x] Step 3.3: Roll rarity during `placeDungeonMonsters` after position selection and before appending the generated monster, preserving stable append order.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Dungeon'
```

- [x] Step 3.4: Resolve generated monster loot table from `abs(level) + loot_depth_offset`, reusing v29 loot bands so depths `3+` route to the `3+` band.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Loot|Depth'
```

- [x] Step 3.5: Add tests proving static world-preset/lab monsters do not get rarity or scaled stats in v30.

```bash
cd server && go test ./internal/game/... -run 'Rarity|World|Dungeon'
```

## Task 4 - Go sim scaling, entity views, and replay

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/http/ws_test.go` if needed
- Modify: `server/internal/replay/replay_test.go` if needed

- [x] Step 4.1: Store generated monster rarity on the internal entity and emit it through existing `EntityView.Rarity` for monster snapshots and deltas.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Snapshot|Delta'
```

- [x] Step 4.2: Apply scaled `hp` / `max_hp` at generated monster spawn time using the plan-chosen deterministic rounding rule.

```bash
cd server && go test ./internal/game/... -run 'Rarity|HP|Dungeon'
```

- [x] Step 4.3: Apply scaled attack damage for proactive monster attacks without changing the base `dungeon_mob` definition.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Damage|Monster'
```

- [x] Step 4.4: Apply scaled XP reward exactly once on monster kill and emit existing progression deltas.

```bash
cd server && go test ./internal/game/... -run 'Rarity|XP|Progression'
```

- [x] Step 4.5: Add replay/determinism coverage proving the same seed and inputs reproduce rarity, scaled state, effective loot table, loot drops, and entity order.

```bash
cd server && go test ./internal/game/... -run 'Rarity|Replay|Loot'
```

- [x] Step 4.6: Add HTTP/replay timeline coverage if game-level tests do not directly prove `/state`, reconnect snapshots, and replay timelines include monster rarity.

```bash
cd server && go test ./internal/http/... ./internal/replay/... -run 'Rarity|Replay|State'
```

## Task 5 - Protocol bot scenario

Files:

- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/21_monster_rarity_loot_scaling.json`
- Modify: `tools/bot/test_protocol.py` if new assertion schema coverage is needed

- [x] Step 5.1: Add bot assertions for monster entity rarity in current state, reconnect snapshot, and replay timeline if existing assertions cannot express this.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.2: Keep exact effective loot-depth assertions in Go/golden tests and make the bot assert non-common generated rarity plus rolled loot pickup.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 5.3: Create `21_monster_rarity_loot_scaling.json` using `world_id: "dungeon_levels"` and a pinned seed that generates at least one non-common dungeon monster.

```bash
make bot
```

- [x] Step 5.4: Scenario steps should descend, observe a generated monster rarity, kill that monster, pick up its loot, assert rolled loot metadata, then run the standard `/state`, reconnect, replay, and fresh-session persistence checks.

```bash
make bot
```

- [x] Step 5.5: Prefer a pinned seed with a `unique` monster. If that makes the scenario brittle or too long, keep the scenario non-common and rely on Go/golden tests for the `unique` level `-5` -> depth `8` requirement.

```bash
make bot
```

## Task 6 - Godot client rarity tinting

Files:

- Modify: `client/scripts/main.gd`
- Modify/Create: `client/tests/*` if tint helpers are extracted

- [x] Step 6.1: Add a small in-repo rarity color map matching shared data defaults for presentation, with missing monster rarity defaulting to `common`.

```bash
make client-unit
```

- [x] Step 6.2: Apply a green material tint to the player visual during client setup without changing server state or protocol.

```bash
make client-unit
```

- [x] Step 6.3: Apply monster material tint on spawn/update from server-provided entity `rarity`; ensure existing loot item `rarity` metadata remains unaffected.

```bash
make client-unit
```

- [x] Step 6.4: Add client smoke coverage if tinting changes scene setup enough to risk headless startup or visual replay.

```bash
make client-smoke
```

- [x] Step 6.5: Run client bot only if the Godot client bot assertions are extended to observe rarity/tint data. No tint-specific Godot client bot assertion was added.

```bash
make bot-client
```

## Task 7 - Regression checks

Files:

- Modify: existing scenario JSON only if pinned seeds or loot expectations intentionally change.
- Modify: `tools/bot/test_protocol.py` only if scenario catalog validation needs updates.

- [x] Step 7.1: Confirm existing dungeon and loot scenarios still pass with rarity defaulting/presentation added.

```bash
make bot
```

- [x] Step 7.2: Confirm `20_dungeon_equipment_drops.json` still proves v29 depth-banded drops with any adjusted generated monster loot table expectations.

```bash
make bot
```

- [x] Step 7.3: Confirm client smoke still handles snapshots/deltas containing monster rarity.

```bash
make client-smoke
```

## Task 8 - Lifecycle docs and CI

Files:

- Modify: `docs/specs/v30_spec-monster-rarity-and-loot-scaling.md`
- Modify: `docs/PROGRESS.md`

- [x] Step 8.1: When implementation is complete, mark the v30 spec as implemented and record any final rounding or data-placement decisions.

```bash
rg -n "Status: Draft|round|monster-rarity-and-loot-scaling|v30" docs/specs/v30_spec-monster-rarity-and-loot-scaling.md docs/PROGRESS.md
```

- [x] Step 8.2: Add v30 to the `PROGRESS.md` lifecycle table, slice numbering note, summary, scripted scenario catalog, and recently closed gaps.

```bash
rg -n "v30|monster-rarity|Latest completed slice|Next slice" docs/PROGRESS.md
```

- [x] Step 8.3: Run full CI and keep the current branch; do not create a feature branch.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run 'Rarity|Dungeon|Loot|XP|Replay'`
- [x] `cd server && go test ./internal/http/... ./internal/replay/... -run 'Rarity|Replay|State'`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make bot`
- [x] `make bot-client` if Godot client bot assertions are extended
- [x] `make ci`

## Deferred scope

- Unique/set item catalogs and unique drop rules.
- Final item-level/depth economy beyond v29 coarse `1`, `2`, `3+` bands.
- Named rare/unique monsters, rare packs, minions, aura modifiers, and boss floors.
- Chest rarity, Magic Find, and player-driven loot rarity modifiers.
- Production enemy art, textures, VFX, audio, and colorblind/accessibility-safe rarity presentation.
