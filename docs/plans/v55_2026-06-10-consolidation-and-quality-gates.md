# v55 Plan — Consolidation and Quality Gates

Status: Complete
Goal: Pay down the four god-file monoliths and add enforcement gates for determinism and test
coverage, as directed by the v53 engineering review.
Architecture: No new features or protocol changes. Pure structural refactoring across all four
layers (Go, GDScript, Python, CI) plus targeted bug fixes and new CI gates.
Tech stack: Go game package, GDScript client, Python tools, bash CI scripts.

## Baseline and approach

Triggered by the v53 engineering review (docs/reviews/20260610_v53-overview.md). Treats the
review's top 10 recommendations as a ranked backlog and executes the highest-leverage items
that are safe to implement without spec-level design:

- Review #1 (CLAUDE.md staleness): already resolved in prior PROGRESS.md work.
- Review #2 (CI determinism lint): implemented as Tier 3 item 1.
- Review #3 (handler registry): implemented as Tier 2 item 1.
- Review #4 (ItemRulesLoader): implemented as Tier 2 item 4.
- Review #5 (15-vs-20 Hz): documented in CLAUDE.md; CLAUDE.md fix deferred.
- Review #6 (architecture doc): deferred.
- Review #7 (regen-golden): implemented as Tier 3 item 2.
- Review #8 (BotMainProxy isolation): partially deferred — bot_types.py extracted but
  BotMainProxy scene-tree work deferred; run.py monolith remains.
- Review #9 (panic(err), push_warning): push_warning implemented in Tier 1; panic(err) skipped
  because NewSim is test-only (production uses NewSimWithWorldProgression which already returns
  errors).
- Review #10 (protocol deprecation policy): deferred.

Execution order: Tier 1 (trivial fixes) → Tier 2 (structural splits) → Tier 3 (infrastructure),
with test-and-commit between each step.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/sim.go` | Fix bare map ranges; replace applyInput switch with registry; nolint annotations |
| Add    | `server/internal/game/handlers.go` | Handler registry + all 22 handle* methods |
| Modify | `server/internal/game/game_test.go` | -update flag, writeGolden, update paths in TestItemRollsGolden + TestTreasureClassRollsGolden |
| Add    | `server/cmd/determinism-lint/main.go` | CI determinism lint tool |
| Modify | `make/server.mk` | lint-determinism and regen-golden make targets |
| Add    | `tools/bot/bot_types.py` | Scenario, RuntimeState, CoopPeer, DEFAULT_WORLD_ID |
| Modify | `tools/bot/run.py` | Remove class defs, import from bot_types |
| Modify | `tools/bot/test_protocol.py` | Import RuntimeState from bot_types |
| Modify | `tools/validate_shared.py` | Lift ShopRNG + seed_to_uint64 to module level; add health_regen cross-check |
| Add    | `client/scripts/item_rules_loader.gd` | Static singleton for shared item data |
| Modify | `client/scripts/main.gd` | Property getters/setters, ensure_loaded(), .get() payload guards, push_warning |
| Modify | `client/scripts/inventory_panel.gd` | Delegate item data to ItemRulesLoader |
| Modify | `client/scripts/shop_panel.gd` | Delegate item data to ItemRulesLoader |
| Modify | `client/scripts/stash_panel.gd` | Delegate item data to ItemRulesLoader |
| Modify | `client/scripts/consumable_bar.gd` | Delegate item data to ItemRulesLoader |
| Add    | `client/tests/test_delta_apply.gd` | 15 headless unit tests for _apply_delta/_apply_snapshot |
| Modify | `scripts/ci.sh` | Renumber steps 1–9; add lint-determinism as step 3/9 |
| Modify | `shared/golden/item_rolls.json` | Regenerated with effect_ids: [] instead of null |
| Modify | `PROGRESS.md` | v55 lifecycle row + current status |
| Modify | `CLAUDE.md` | New invariants and agent rules from this slice |
| Add    | `docs/specs/v55_spec-consolidation-and-quality-gates.md` | This slice's spec |
| Add    | `docs/plans/v55_2026-06-10-consolidation-and-quality-gates.md` | This plan |

## Tier 1 — Trivial safety fixes

### Task 1.1 — Bare map range fixes (sim.go)
- [x] Step 1.1.1: Replace `for slot, instanceID := range s.equipped` in `handleDrop` (was :3150)
  with `for _, slot := range sortedStringKeys(s.equipped)`. Closes hot-path determinism hole.
- [x] Step 1.1.2: Same fix in `equipPreviewForItemWithSlot` (was :5382).
- [x] Step 1.1.3: `go vet ./...; go test ./internal/game/...` green.

### Task 1.2 — push_warning for unknown message types (main.gd)
- [x] Step 1.2.1: Add `_: push_warning(...)` default arm to `_handle_message` match statement.
- [x] Step 1.2.2: `make client-unit` green.

### Task 1.3 — health_regen cross-check (validate_shared.py)
- [x] Step 1.3.1: Add cross-check near end of `cross_checks()` that asserts
  `health_regen_per_10_seconds` in item_templates/shop stat_weights and
  `health_regen_per_second` in character_progression derived_stats.
- [x] Step 1.3.2: `validate_shared.py`: 600 checks pass; `pytest tools/`: 59 green.

```bash
cd server && go vet ./... && go test ./internal/game/... -count=1
make client-unit
.venv/bin/python tools/validate_shared.py
```

## Tier 2 — Structural splits

### Task 2.1 — Handler registry (sim.go → handlers.go)
- [x] Step 2.1.1: Use Python to extract all 22 `func (s *Sim) handle*` methods from sim.go into
  a new `handlers.go` (same package). Add `import "fmt"` to handlers.go.
- [x] Step 2.1.2: Define `inputHandlerFunc`, `wrapLevelTravel`, `handleClientReady`, and the
  `inputHandlers` registry map in handlers.go.
- [x] Step 2.1.3: Replace the 57-line `applyInput` switch in sim.go with a 9-line registry
  dispatch + dead-player guard.
- [x] Step 2.1.4: `go build ./...; go test ./internal/game/...` green.
  sim.go: 7,036 → 5,980 lines (−1,056). handlers.go: 1,067 lines.

### Task 2.2 — ShopRNG and seed_to_uint64 (validate_shared.py)
- [x] Step 2.2.1: Move `class ShopRNG` and `def seed_to_uint64` from nested inside
  `cross_checks()` to module level. Add docstrings. Update indentation.
- [x] Step 2.2.2: `validate_shared.py`: 600 checks; `pytest tools/`: 59 tests.

### Task 2.3 — bot_types.py extraction (run.py)
- [x] Step 2.3.1: Create `tools/bot/bot_types.py` with `Scenario`, `RuntimeState`, `CoopPeer`,
  and `DEFAULT_WORLD_ID`. Add module docstring.
- [x] Step 2.3.2: Remove class defs from run.py; replace `from dataclasses import dataclass, field`
  with `from tools.bot.bot_types import ...`.
- [x] Step 2.3.3: Update `test_protocol.py` to import `RuntimeState` from `bot_types`.
- [x] Step 2.3.4: `pytest tools/`: 59 green.

### Task 2.4 — ItemRulesLoader (GDScript)
- [x] Step 2.4.1: Create `client/scripts/item_rules_loader.gd` with `class_name ItemRulesLoader`,
  `extends RefCounted`, static vars, `ensure_loaded()` guard, `item_definition()` helper, and
  three `_load_*` static methods.
- [x] Step 2.4.2: In each of the 5 scripts: replace local var declarations with property
  getters/setters that delegate to `ItemRulesLoader.*`; add `ItemRulesLoader.ensure_loaded()`
  to `_ready()`; remove the three `_load_*` functions.
  Key: properties need setters (not just getters) because `test_item_visuals.gd` injects fixture
  data via `main.item_rules = ...`.
- [x] Step 2.4.3: `make client-unit` (15/15 tests pass).
  Note: do NOT use Godot autoload registration for this — `class_name` resolution works headless
  without `--import`; autoload would fail tests that `preload()` main.gd at compile time.

```bash
cd server && go build ./... && go test ./internal/game/... -count=1
.venv/bin/pytest tools/ -q
.venv/bin/python tools/validate_shared.py
make client-unit
```

## Tier 3 — Infrastructure

### Task 3.1 — Determinism lint CI gate
- [x] Step 3.1.1: Write `server/cmd/determinism-lint/main.go` using only stdlib (go/ast,
  go/parser, go/token, go/types, go/importer). Three checks:
  (a) `math/rand` import in any game/ file, (b) `time.Now()` call in any game/ file,
  (c) bare map range (key+value, non-blank key) in hot-path files sim.go + handlers.go only.
- [x] Step 3.1.2: Add `//nolint:determinism` comment mechanism; annotate 5 known-safe
  map-clone patterns in sim.go (output is a map, iteration order irrelevant).
- [x] Step 3.1.3: Add `lint-determinism` target to `make/server.mk`.
- [x] Step 3.1.4: Insert lint as step 3/9 in `scripts/ci.sh`; renumber remaining steps.
- [x] Step 3.1.5: `make lint-determinism` → `determinism-lint: OK`.

### Task 3.2 — make regen-golden
- [x] Step 3.2.1: Add `var update = flag.Bool("update", ...)` and `writeGolden()` helper to
  `server/internal/game/game_test.go`.
- [x] Step 3.2.2: Add `-update` path to `TestItemRollsGolden`: recompute actual,
  normalize nil EffectIDs to `[]string{}` (Go nil → null in JSON → GDScript `as Array` crash),
  write back via `writeGolden`.
- [x] Step 3.2.3: Same for `TestTreasureClassRollsGolden`.
- [x] Step 3.2.4: Add `regen-golden` target to `make/server.mk`.
- [x] Step 3.2.5: Run `make regen-golden` → item_rolls.json regenerated with `effect_ids: []`.
- [x] Step 3.2.6: `go test ./internal/game/... -count=1` green with updated golden.

### Task 3.3 — Delta unit tests + payload guards
- [x] Step 3.3.1: Write `client/tests/test_delta_apply.gd` covering pure state-mutation ops:
  gold_update, inventory_add/remove/update, equipped_update, hotbar_update, stash_gold_update,
  stash_item_add/remove, snapshot field assignment, and malformed delta robustness.
  Exclude entity ops (require scene-tree nodes).
- [x] Step 3.3.2: Fix `c["entity"]`, `c["item"]`, `c["item_instance_id"]` direct access in
  `_apply_delta` with `.get()` guards (as flagged in the v53 review and confirmed by the
  malformed-delta test).
- [x] Step 3.3.3: Register test as gate 2m in `scripts/client_smoke.sh`.
- [x] Step 3.3.4: `make client-unit` (15/15 pass, including new delta test).

```bash
make lint-determinism
make regen-golden
cd server && go test ./internal/game/... -count=1
make client-unit
```

## Final verification

- [x] `cd server && go vet ./...`
- [x] `cd server && go test ./internal/game/... -count=1` (265 tests)
- [x] `cd server && go run ./cmd/determinism-lint ./internal/game/...` → OK
- [x] `.venv/bin/python tools/validate_shared.py` → 600 checks
- [x] `.venv/bin/pytest tools/ -q` → 59 passing
- [x] `make client-unit` → 15/15 PASS
- [x] `make ci` → 9/9 phases green

## Deferred scope

- Full UIOrchestrator and BotMainProxy extraction (scene-tree wiring, needs dedicated slice).
- Complete `cross_checks()` decomposition into sub-functions (2,659-line function; safe once
  bot+replay coverage is better; add `-update` path to remaining golden tests first).
- Protocol version deprecation policy and `shared/protocol/archive/` pruning.
- ARCHITECTURE.md as-built doc.
- ADRs 0002–0005 (wire protocol, shared-rules contract, determinism enforcement, netcode).
- `-update` paths for remaining golden tests beyond item_rolls and treasure_class_rolls.
- BotMainProxy: move ~220 lines of autoplay_*/visual_replay_*/bot_* state out of main.gd.
- Additional entity/scene test coverage for _apply_delta (entity_spawn/entity_update paths require
  a wired scene tree; covered indirectly by bot scenarios, but not by unit tests).
