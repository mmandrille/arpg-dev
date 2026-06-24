# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Model policy

Default to **claude-sonnet-4-6** (medium effort) for all tasks in this project.
Escalate to a higher-power model only when explicitly asked or when the task clearly requires it
(e.g. complex architecture decisions, large multi-file refactors). Always confirm before escalating.

## Project progress (read first)

**Before any new feature or slice:** read [`PROGRESS.md`](PROGRESS.md) — **Current status**, **Open
gaps**, and **Agent checklist** only. Slice history and codenames live in [`docs/progress/`](docs/progress/).
Per-slice as-built notes live in [`docs/as-built/`](docs/as-built/).

## Commands

Everything runs through `make`. Run `make help` for the full list.

```bash
# Infrastructure
make db-up           # start local Postgres (required before server or bot)
make db-reset        # destroy + recreate Postgres (drops all data)

# Running
make server          # run Go server (requires db-up first)
make bot             # run Python protocol bot end-to-end (quiet; add VERBOSE=1 for full logs)
make bot-visual      # server + Godot replay (quiet when headless; VERBOSE=1 for full logs)
make play            # Postgres + server + interactive Godot client

# Testing
make test            # unit tests (quiet; add VERBOSE=1 for full logs)
make test-go         # all Go tests  (`cd server && go test ./...`)
make test-py         # Python unit tests (`pytest tools/`)
make client-unit     # Godot headless unit tests (quiet; add VERBOSE=1 for full logs)
make client-smoke    # Godot headless smoke against a running TEST_BASE_URL server
make ci              # full local CI suite (quiet; add VERBOSE=1 for full logs)
make test-all        # test + ci + headless bot-visual (quiet; add VERBOSE=1 for full logs)

# Shared contracts
make validate-shared # validate all shared JSON against their schemas
make validate-assets # validate asset manifest + runtime .glb files

# Assets (regenerate deterministic committed files)
make gen-assets      # regenerate runtime .glb files
make gen-anims       # regenerate AnimationLibrary .tres clips (requires Godot)
make model-list      # list previewable character/monster model asset IDs
make model model=<asset_id>        # open a focused Godot preview for one model
make model model=<asset_id> CHECK=1 # headless model-viewer load check

# Replay
make replay SESSION_ID=<id>   # re-simulate a recorded session and verify output
```

**Single Go test:** `cd server && go test ./internal/game/... -run TestName`

**Single Python test:** `cd <repo-root> && .venv/bin/pytest tools/bot/test_protocol.py::test_name -v`

**Override env vars on the command line:**
```bash
VERBOSE=1 make ci                       # full CI logs (also: V=1)
AUTOPLAY_STEP_DELAY=0.8 make bot-visual   # slower playback
GODOT=/path/to/godot make bot-visual
SESSION_ID=abc123 make replay
```

## Test Locking Policy

Use exact values only when the test intentionally owns a stable contract: protocol/schema shape,
replay determinism for the same seed and ordered inputs, persistence boundaries, formula goldens,
and Go/GDScript evaluator parity. Full snapshots and exact event ordering are appropriate when
replay equality or protocol shape is the subject of the test.

Gameplay tuning values should use semantic, range, derived, or eventual assertions unless a named
golden explicitly owns the value. This includes dungeon size, generated population, movement speed,
timing budgets, damage tuning, loot weights, and generated coordinates.

Examples:
- Dungeon placement: prefer "down stair exists, is within generated bounds, and is reachable from
  spawn" over pinning its exact generated coordinate.
- Population: prefer "at least one champion dungeon mob exists for this pinned rarity scenario" or
  "monster count is within rule-derived bounds" over duplicating a first-pass total in unrelated
  scenarios.
- Movement speed: prefer "monster moves closer / leashes within a timeout derived from current
  rules" over "position equals X after exactly N ticks" unless fixed-tick stepping is the feature.

## Data-Driven Configuration Policy

Gameplay tuning must be data-driven by default. Before adding or changing any balance-sensitive
value, decide whether it belongs in `shared/rules/main_config.v0.json`, another
`shared/rules/*.json` catalog, or a schema-backed content file. This applies to attack speed,
movement speed, drop chance, loot weights, monster stats, class stats, skill costs/cooldowns, shop
pricing, XP curves, generated-content budgets, and similar values. Hardcoding a new tuning value in
Go, GDScript, or Python requires an explicit note in the spec/plan explaining why code ownership is
the right boundary.

Tests must protect configurability. Prefer deriving expectations from loaded rules, using
semantic/range/eventual assertions, or creating focused temp-rule fixtures that modify only the
shared JSON under test and prove gameplay follows it. Do not duplicate current rule values in
unrelated tests; those assertions become accidental tuning locks. Exact numbers remain appropriate
for named goldens, protocol/schema contracts, deterministic replay equality, evaluator parity, and
tests whose stated purpose is to own a formula.

## Maintainability Ratchet

Source, test, and tool files have a target maximum of **600 lines**.

Rules:

1. New source/test/tool files should stay at or below 600 lines.
2. Existing over-limit files are grandfathered in `.maintainability/file-size-baseline.tsv`.
3. Grandfathered files may not grow by more than 25 lines, and their baseline may not remain more
   than 25 lines above the current file size. When a large file shrinks, lower the baseline in the
   same slice, or drop the entry if the file is now at or below 600 lines.
4. touch-to-shrink: a slice that edits a grandfathered file should leave it at or below its current
   baseline. The +25 allowance exists for incidental drift, not for routinely growing coordinators.
5. New gameplay or tooling domains start in their own focused file. Do not add a new domain to a
   large coordinator just because related code already lives there.
6. A documented exception that grows a grandfathered coordinator is allowed only when the same slice
   extracts at least as many lines from that file as it adds. If that net non-positive split is not
   possible, the next slice must be a payback slice that shrinks the touched coordinator.
7. `make maintainability` prints the current grandfathered file count and total line count. Reviews
   should record that trend and expect it to move down over time.
8. `make ci` runs `make maintainability`; the ratchet is a CI gate, not an advisory check.

This is a reduction ratchet, not a repo-wide rewrite mandate: shrink large files opportunistically,
lock in those reductions, and block new large files from appearing.

### Extraction independence

Line-count reduction is not enough to count as architectural decoupling. A newly extracted module
must be importable and unit-testable without importing the file it came from and without receiving
that file's full namespace through `globals()` or equivalent helper laundering. If an extracted
module needs runtime services, pass a typed context object or an explicit narrow helper set and add a
focused test that imports the extracted module directly.

`make maintainability` also runs an extraction-coupling ratchet. Existing `helpers=globals()` sites
are grandfathered as debt; new sites fail CI, and removals must lower the baseline in the same
slice.

## Architecture

```
PROGRESS.md  slice lifecycle + current status (read before new work)
docs/CODEMAP.md domain → files index for loading focused context
client/      Godot 4 (GDScript) — thin client: renders + locally predicts movement/attack feedback
server/      Go authoritative server — owns all game state, loot rolls, inventory, persistence
shared/      Data contracts: JSON schemas, rules-as-data, cross-language golden fixtures
tools/       Python: protocol bot, replay wrapper, shared schema validator, asset pipeline tools
assets/      Source art manifests + asset generation scripts
docs/        ADRs + specs + plans + as-built + reviews (periodic ~every 10 slices — see PROGRESS.md)
```

### The authoritative boundary (ADR-0001 D2)
The client is a renderer + input layer; **the server owns every outcome that matters** (HP, damage, loot rolls, inventory). Even in solo play the client speaks the full production-shaped protocol over WebSocket. There is no local-only path or client-side shortcut.

### Server internals (`server/internal/`)
- **`game/`** — deterministic authoritative simulation (`Sim`). Given the same seed + ordered inputs it always produces identical output. Enforced: seeded PRNG only (`rng.go`), no `time.Now()`, stable entity-ID ordering, **10 Hz live tick** (`server/internal/realtime/protocol.go:17`). Rules loaded from `shared/rules/` at startup (`rules.go`). **CI gate:** `make lint-determinism` fails on `time.Now()`, `math/rand`, or bare map ranges in hot-path files — see `server/cmd/determinism-lint/`.
- **`realtime/`** — WebSocket hub + per-session runner. `Hub.Run()` upgrades the connection, constructs a `Sim`, and enters the session loop.
- **`store/`** — repository interface + Postgres implementation. Sessions, inventory, events all persist here.
- **`auth/`, `http/`, `replay/`** — platform services (auth, REST endpoints, replay command).
- **`cmd/arpg-server`** — main binary (also self-migrates on boot). **`cmd/arpg-replay`** — standalone replay verifier.

### Shared contracts (`shared/`)
- **`protocol/`** — JSON schemas for WebSocket messages (`envelope`, `messages`, `state_delta`, `session_snapshot`). Wire format is JSON over WebSocket in v0; both Go and GDScript consume these.
- **`rules/`** — combat formulas, item definitions, monster definitions, loot tables as JSON data. The Go sim and Godot client both evaluate formulas from this shared catalog — never from language-specific logic.
- **`golden/`** — cross-language fixture files (damage formulas, loot rolls, slice outcomes, retaliation damage). Go tests and Godot `test_golden.gd` both assert against these.

### Client (`client/`)
- `scripts/main.gd` — top-level coordinator: manages `net_client`, entity map (`{node, controller, type}`), routes server events to entity animation controllers, reads `state_delta.events` array for authoritative triggers.
- `scripts/animation_controller.gd` — injected per-entity. Priority state machine: terminal (death) > one-shot (hit/attack) > locomotion (idle/walk). Drives `AnimationPlayer`; never crosses the wire.
- `scripts/net_client.gd` — WebSocket connection, serializes/deserializes protocol envelopes.
- `tests/test_golden.gd` — headless cross-language golden checks run via `make client-smoke`.

### Python tools (`tools/`)
- `bot/run.py` — protocol bot: authenticates, creates a session, sends move/attack/pickup/equip intents over the same WebSocket the real client uses, asserts on authoritative state. Primary agent-playability path.
- `replay/run.py` — replay wrapper (invoked via `make replay`).
- `validate_shared.py` — validates all JSON in `shared/` against their schemas.
- `assets/gen_glb.py` — generates committed runtime `.glb` files deterministically.
- `assets/validate_assets.py` — validates the asset manifest + GLB node presence.

### Animation model (ADR-0007)
Animation is **client-only presentation state** — never on the wire. Player locomotion (`idle/walk/attack`) is driven by client input/prediction signals. Player damage/death and monster reactions (`hit/death`) are driven by authoritative events (`player_damaged`, `player_killed`, `monster_damaged`, `monster_killed`) already present in every `state_delta.events`. No server or protocol change is needed to add new reactions — only a server event emission and a client mapping.

## SDD Process

This project uses Spec-Driven Development. Before touching code for any new feature:

1. Read [`PROGRESS.md`](PROGRESS.md) — confirm baseline slice, branch, open gaps, and invariants.
2. Read or write the spec under `docs/specs/vN_spec-<feature>.md` (`N` = next execution order from `PROGRESS.md` and existing specs/plans).
3. Write or check the plan under `docs/plans/vN_<date>-<feature>.md`.
4. Consult the relevant ADRs in `docs/adr/` — especially ADR-0001 (foundational) and any feature-specific ones.
5. When the slice completes, update `PROGRESS.md` (lifecycle table, new gaps) and
   `docs/as-built/vN_<codename>.md` (what it proved).
6. At ~10-slice milestones, write the repo-wide engineering review under `docs/reviews/` first, then run `$refactor` against that fresh scorecard for minor verified cleanup (see `PROGRESS.md` → **Periodic engineering reviews**).

## Key Invariants

- **Determinism in the Go sim is non-negotiable.** No `time.Now()`, `rand.Intn()` without the
  seeded `RNG`, or bare map ranges with key+value in game logic (`game/` package hot-path files).
  Enforced by `make lint-determinism` (CI step 3/9). Known-safe map clones that output maps are
  annotated `//nolint:determinism` with a WHY comment. Any violation breaks replay.
- **New intents register in `handlers.go`, not in `applyInput`.** Add one entry to `inputHandlers`
  map in `handlers.go`. Never edit `applyInput` in `sim.go` for a new intent type.
- **Shared rules are data, not code.** Formula types live in `shared/rules/`; Go and GDScript each
  implement the same small evaluator set. Never add executable logic to shared data files.
- **Golden fixtures are cross-language contracts.** Any change to combat formulas or loot logic
  must update `shared/golden/` and both Go and GDScript golden tests. Use `make regen-golden`
  after intentional formula changes (runs `go test -update -run Golden`). Cross-language applies to
  formula/loot goldens consumed by both stacks; a few goldens are intentionally single-language —
  e.g. `shared/golden/dungeon_obstacles.json` is a Go-only dungeon-layout *determinism* contract with
  no GDScript consumer, which is valid replay-determinism locking under the Test Locking Policy.
  When writing golden update paths: normalize nil Go slices to `[]T{}` before `writeGolden` —
  Go `null` JSON breaks GDScript `as Array` casts.
- **Protocol JSON schemas are versioned.** Changes to `shared/protocol/` require a schema version
  bump and must remain backward-compatible or require coordinated client+server update.
- **Asset manifest is the source of truth for asset identity.** `assets/manifests/assets.v0.json`
  maps `asset_id → runtime_path`. `shared/assets/item_visuals.v0.json` maps
  `item_def_id → asset_id + mount_socket`. These two files are the only canonical link between
  gameplay and visuals.
- **GDScript shared singletons: use `class_name` with static vars, not Godot autoload.**
  Autoload names are resolved at runtime, not at GDScript compile time. Any test that `preload()`s
  a script depending on an autoload will fail with "Identifier not found" in headless mode.
  Pattern: `class_name Foo extends RefCounted` + `static var _loaded: bool` + `ensure_loaded()`.

## Agent rules (from v55 consolidation)

These rules emerged from paying down the god-file debt. Agents should follow them on every slice:

### Go server

1. **Handler registry discipline.** `applyInput` dispatches via `inputHandlers` map. A new intent
   type adds one line to `handlers.go`. Never add a new `case` to `applyInput`.

2. **Map range in game/.** When you write `for k, v := range someMap` in `sim.go` or
   `handlers.go`, you must either: (a) use a `sorted*` helper before iterating, or (b) add
   `//nolint:determinism` with a one-line comment explaining WHY the result is order-independent
   (e.g., "output is a map", "commutative sum", "bool existence check").
   The lint will fail CI otherwise.

3. **Golden updates.** After any formula or loot rule change, run `make regen-golden` and commit
   the updated fixtures alongside the rule change. When adding a new `-update` path to a test,
   normalize nil slices: `if got.SomeSlice == nil { got.SomeSlice = []T{} }` before writing.

### Python tools

4. **bot_types.py is the type home.** `Scenario`, `RuntimeState`, `CoopPeer`, and
   `DEFAULT_WORLD_ID` live in `tools/bot/bot_types.py`. Do not re-add them to `run.py` or any
   other module. New shared types go to `bot_types.py`, not inline in `run.py`.

5. **validate_shared cross-checks.** When adding a new stat key that appears under two different
   names or units across rules/protocol/goldens, add a `cross_checks()` assertion that verifies
   both names are present. The `health_regen_per_10_seconds` ↔ `health_regen_per_second` example
   is the template.

6. **run.py split freeze.** Do not run another mechanical `tools/bot/run.py` split that passes
   `helpers=globals()` into the extracted module. Keep `run.py` as a grandfathered orchestrator
   unless a future slice introduces a typed bot runtime context and replaces helper-global wrappers
   directly.

### GDScript client

7. **ItemRulesLoader pattern for shared data.** Any data loaded from `shared/` that is needed by
   multiple scripts goes through a `class_name Foo extends RefCounted` static singleton with
   `ensure_loaded()`. Do NOT duplicate file-load code across scripts.

8. **Delta payload access.** Never use `c["key"]` on a delta change dictionary — always use
   `c.get("key", default)`. Direct access crashes on partial/malformed server messages; `.get()`
   degrades gracefully and makes the malformed-delta test pass cleanly.

9. **Headless unit tests avoid `main.gd` scene-tree paths.** Tests in `tests/` using
   `extends SceneTree` that exercise `MainScript.new()` test pure state mutations only; they must not
   call paths that access `$Node` children, `get_parent()`, or require `entities_root`/`_walls_root`/etc.
   Those `main.gd` paths are covered indirectly by bot scenarios; unit tests cover state correctness.
   Exception — **node-render component tests** (e.g. `test_factories.gd`, `test_fog_of_war_overlay.gd`):
   a test whose subject *is* a self-contained node may `add_child` that node and `await process_frame`
   to exercise its rendering/geometry. This is an established, CI-gated pattern; the prohibition is
   specifically about reaching into `main.gd`'s scene graph from a `MainScript.new()` state test.
   Register every such test in `scripts/client_smoke.sh` so it actually runs.

### CI / process

10. **Test and commit between independent refactoring steps.** Each structural change should be
   committed with a passing test suite before the next one begins. This keeps `git bisect` clean
   and avoids entangling unrelated failures.

11. **Trivials first, then structural splits, then infrastructure.** Safety bugs (bare ranges,
    missing guards) should land before monolith splits so the split is on proven-correct code.

12. **Protocol bot scenarios default to a 10-second budget (15s hard ceiling).** `tools/bot/run.py`
    fails any scenario whose full run exceeds its budget (`MAX_SCENARIO_ELAPSED_S = 15.0`; default and
    per-scenario `max_elapsed_s` override at `tools/bot/run.py`). Keep scenarios at the 10.0s default;
    raise `max_elapsed_s` toward the 15s ceiling **only** when the generated-world route is the
    behavior under test, and pair it with step-level budgets (`max_ticks`/`timeout_s`) so failures
    still point at the slow navigation or wait. Otherwise shorten to the contract it is really
    proving: compact lab worlds, focused setup, or lower-level Go/Python tests for exhaustive
    traversal/timing coverage instead of waiting through unrelated dungeon walks or combat cycles.
    See `docs/progress/scenario-catalog.md`.
