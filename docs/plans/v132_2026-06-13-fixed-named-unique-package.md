# v132 Plan — Fixed Named Unique Package

Status: Complete
Goal: Make `embercall_blade` a deterministic named unique item package and expose it through the purple test chest.
Architecture: Shared rules own fixed named unique stats and effect ids. Go rules validate and load
the catalog, then the existing server-authored purple chest appends named unique payloads after its
effect coverage rows. Bot proof remains protocol-level; the client only renders existing item fields.
Tech stack: shared JSON/schema, Python validator/bot, Go sim.

## Baseline And Shortcut Decision

Builds on v119 live unique effect reachability and v131 purple chest coverage. Godot plugin
adoption is rejected for this slice because there is no new UI, art, camera, or inventory
presentation work.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/unique_items.v0.schema.json` | Add fixed package fields |
| Modify | `shared/rules/unique_items.v0.json` | Define Embercall Blade package |
| Modify | `tools/validate_shared.py` | Validate fixed package compatibility |
| Modify | `server/internal/game/rules.go` | Load and validate named uniques |
| Modify | `server/internal/game/unique_chest.go` | Build named unique payloads |
| Modify | `server/internal/game/unique_chest_test.go` | Cover named package and chest count |
| Modify | `tools/bot/unique_effect_assertions.py` | Assert named unique presence |
| Modify | `tools/bot/scenarios/61_purple_town_unique_chest.json` | Require Embercall Blade |
| Modify | `tools/bot/test_protocol.py` | Unit-test assertion shape |
| Create | `docs/as-built/v132_fixed-named-unique-package.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle and backlog update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] `tools/validate_shared.py`
- [x] `tools/bot/test_protocol.py`

Decision:
- [x] Defer broad validator and test file splits because v130 already identified the validation split
  as a follow-up; this slice adds only compact checks needed for the named unique package.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Named Package

Files:
- Modify: `shared/rules/unique_items.v0.schema.json`
- Modify: `shared/rules/unique_items.v0.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add `fixed_stats` and `fixed_effect_ids` schema fields.
- [x] Step 1.2: Define Embercall Blade fixed stats and `everburning_wound`.
- [x] Step 1.3: Validate fixed stats are numeric and fixed effects are known, ready, unique, and compatible.

```bash
make validate-shared
```

## Task 2 — Server Payload

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/unique_chest.go`
- Modify: `server/internal/game/unique_chest_test.go`

- [x] Step 2.1: Load named unique definitions into `Rules`.
- [x] Step 2.2: Build deterministic named unique payloads from base template plus fixed package fields.
- [x] Step 2.3: Append named uniques to purple chest grants without changing natural roll odds.
- [x] Step 2.4: Add Go tests for Embercall Blade payload and chest inclusion.

```bash
cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest|TestLoadRules'
```

## Task 3 — Bot Proof

Files:
- Modify: `tools/bot/unique_effect_assertions.py`
- Modify: `tools/bot/scenarios/61_purple_town_unique_chest.json`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 3.1: Add assertion support for required named unique display names/effect ids.
- [x] Step 3.2: Require Embercall Blade in the purple chest scenario.
- [x] Step 3.3: Unit-test the assertion helper through `tools/bot/test_protocol.py`.

```bash
.venv/bin/python -m pytest tools/bot/test_protocol.py -q
make bot scenario=purple_town_unique_chest
```

## Task 4 — Lifecycle Docs And CI

Files:
- Create: `docs/as-built/v132_fixed-named-unique-package.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Record v132 completion and deferred scope.
- [x] Step 4.2: Run final verification.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestNamedUnique|TestUniqueTestChest|TestLoadRules'`
- [x] `.venv/bin/python -m pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=purple_town_unique_chest`
- [x] `make ci`
