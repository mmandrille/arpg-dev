# Spec: `consolidation-and-quality-gates`

Status: Complete
Date: 2026-06-10
Codename: `consolidation-and-quality-gates`
Slice: v55 — consolidation and quality gates

## Purpose

The v53 engineering review scored the project 8.1/10 and identified a single dominant risk: four
god-files (sim.go 7,036 LOC, main.gd 4,410 LOC, run.py 4,824 LOC, validate_shared.py 2,779 LOC)
growing monotonically with each slice, plus two determinism violations and zero unit coverage on
the highest-risk client code. The review verdict: *"The next phase of work should be consolidation,
not just accretion — pay down the four monoliths and add the enforcement gates while the boundaries
are still clean enough to make it cheap."*

This slice executes that verdict. It introduces no new gameplay features. Its goal is to prove that
the architecture can absorb future feature slices without the maintenance tax compounding further.

**No protocol changes, no new game mechanics, no database migrations.**

## Non-goals

- No new gameplay or protocol features.
- Full UIOrchestrator or BotMainProxy extraction from main.gd (complex scene-tree dependencies;
  deferred to a dedicated slice).
- Complete three-file split of validate_shared.py cross_checks (requires 2,600-line surgery with
  no unit tests; the targeted ShopRNG extraction and handler registry cover most of the value).
- Backfilling `-update` support to all golden tests (added to the two most formula-sensitive ones;
  others can be extended when formulas change).

## Acceptance criteria

**Safety fixes:**
- `sim.go:handleDrop` and `sim.go:equipPreviewForItemWithSlot` iterate `s.equipped` via
  `sortedStringKeys()` instead of bare range. `go vet` and all 265 Go tests remain green.
- `_handle_message` in `main.gd` has a default `_:` arm that calls `push_warning` so unknown
  server message types are never silently dropped.

**Cross-check:**
- `validate_shared.py` asserts that `health_regen_per_10_seconds` appears in item/shop rules AND
  `health_regen_per_second` appears in character_progression derived stats. 600 checks pass.

**Handler registry (sim.go → handlers.go):**
- All 22 `handle*` methods and the `inputHandlers` registry live in `handlers.go`.
- `applyInput` in `sim.go` is ≤ 12 lines: one dead-player guard and one registry dispatch.
- sim.go shrinks by > 1,000 lines. All 265 Go tests pass.
- New intents are added by registering one entry in `handlers.go`; `sim.go` is not edited.

**ShopRNG / seed_to_uint64 (validate_shared.py):**
- Both moved to module level with docstrings. The `cross_checks()` function no longer contains
  nested class definitions. `validate_shared.py`: 600 checks pass; pytest tools/: 59 tests green.

**ItemRulesLoader (GDScript):**
- `item_rules_loader.gd` uses `class_name ItemRulesLoader` with static vars and `ensure_loaded()`.
  No Godot autoload registration required (works headless without `--import`).
- `main.gd`, `inventory_panel.gd`, `shop_panel.gd`, `stash_panel.gd`, `consumable_bar.gd` no
  longer contain `_load_item_rules()`, `_load_item_templates()`, or `_load_item_presentations()`.
- All 15 Godot client-unit tests pass (including `test_item_visuals.gd` which preloads main.gd).

**bot_types.py:**
- `Scenario`, `RuntimeState`, `CoopPeer`, and `DEFAULT_WORLD_ID` live in `tools/bot/bot_types.py`.
- `run.py` imports from `bot_types`; `test_protocol.py` imports `RuntimeState` from `bot_types`.
- pytest tools/: 59 tests green.

**Determinism lint:**
- `server/cmd/determinism-lint/main.go` fails CI on `time.Now()`, `math/rand` imports, and bare
  map ranges (key+value) in `sim.go` / `handlers.go`.
- `make lint-determinism` is clean. The lint is step 3/9 in `scripts/ci.sh`.
- Known-safe map clones annotated with `//nolint:determinism` plus a WHY comment.

**Golden regen:**
- `game_test.go` has a `-update` flag and `writeGolden` helper.
- `TestItemRollsGolden` and `TestTreasureClassRollsGolden` support `-update`.
- `make regen-golden` regenerates those fixtures and all 265 tests pass.
- Nil `[]string` fields are normalized to `[]` on write so GDScript `as Array` casts never see null.

**Delta unit tests + payload guards:**
- `client/tests/test_delta_apply.gd` has 15 tests / 20 assertions covering: gold_update,
  inventory_add/remove/update, equipped_update, hotbar_update, stash_gold_update,
  stash_item_add/remove, snapshot field assignment, and malformed-delta robustness.
- `_apply_delta` in `main.gd` replaces `c["entity"]`, `c["item"]`, `c["item_instance_id"]` direct
  access with `.get()` guards; a malformed partial op no longer logs a script error.
- `client_smoke.sh` registers the test as gate 2m.
- All 15 Godot client-unit tests pass.

**Final:**
- `make ci` green (all 9 phases including new determinism-lint step).

## Scope and files changed

### Go server
- `server/internal/game/sim.go` — bare map range fixes, nolint annotations, applyInput replaced
- `server/internal/game/handlers.go` (new) — handler registry + 22 handle* methods
- `server/internal/game/game_test.go` — `update` flag, `writeGolden`, update paths in 2 tests
- `server/cmd/determinism-lint/main.go` (new) — CI lint tool
- `make/server.mk` — `lint-determinism` and `regen-golden` targets

### Python tools
- `tools/bot/bot_types.py` (new) — Scenario, RuntimeState, CoopPeer, DEFAULT_WORLD_ID
- `tools/bot/run.py` — removes class defs, imports from bot_types
- `tools/bot/test_protocol.py` — imports RuntimeState from bot_types
- `tools/validate_shared.py` — ShopRNG + seed_to_uint64 lifted, health_regen cross-check added

### GDScript client
- `client/scripts/item_rules_loader.gd` (new) — static singleton
- `client/scripts/main.gd` — property getters/setters, ensure_loaded(), payload guards
- `client/scripts/inventory_panel.gd` — delegates to ItemRulesLoader
- `client/scripts/shop_panel.gd` — delegates to ItemRulesLoader
- `client/scripts/stash_panel.gd` — delegates to ItemRulesLoader
- `client/scripts/consumable_bar.gd` — delegates to ItemRulesLoader
- `client/tests/test_delta_apply.gd` (new) — 15 unit tests

### CI / shared
- `scripts/ci.sh` — steps renumbered 1/9–9/9, lint-determinism added as 3/9
- `shared/golden/item_rolls.json` — regenerated (effect_ids: null → [])

## Test commands

```bash
# Go
cd server && go vet ./...
cd server && go test ./internal/game/... -count=1
cd server && go run ./cmd/determinism-lint ./internal/game/...
make regen-golden

# Python
.venv/bin/python tools/validate_shared.py
.venv/bin/pytest tools/ -q

# Godot
make client-unit

# Full
make ci
```

## Open questions and risks

- Q1: Should map ranges in rules.go be covered by the determinism lint?
  - Decision: No. rules.go ranges run only inside LoadRules (startup validation), not the hot-path
    tick loop. Excluding the file is the right scope; the two documented hot-path files are covered.
- Q2: Should ShopRNG share the same implementation as rng.go?
  - Decision: No for v55. ShopRNG lives in validate_shared.py as a Python mirror for cross-language
    validation. A future slice can add a test that cross-checks ShopRNG against rng.go output.
- Risk: GDScript static vars are shared across all instances in a Godot session. Tests that inject
  fixture data into ItemRulesLoader will overwrite the loaded data. Accepted: the property setters
  write through to the static vars, which is what tests need; test isolation is by test ordering,
  not instance isolation.
