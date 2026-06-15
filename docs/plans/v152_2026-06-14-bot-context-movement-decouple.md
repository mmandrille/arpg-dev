# v152 Plan — Bot Context Movement Decouple

Status: Complete
Goal: Replace `helpers=globals()` laundering with a typed `BotContext` + direct imports of pure
queries, migrating `movement_runtime` as the first real decoupling (43 → 37 coupling sites).
Architecture: Pure `RuntimeState` queries move to a leaf module everyone imports directly
(`runtime_queries.py`). The one stateful service movement needs (`pump_one`) is injected via a frozen
`BotContext` dataclass. Intra-module movement calls become direct instead of bouncing through
`run.py`'s globals. `run.py` keeps its public wrappers but builds a context instead of passing
`globals()`. Net effect: `movement_runtime` is importable and unit-testable without `tools.bot.run`.
Tech stack: Python protocol bot, pytest, both maintainability ratchets, full bot/CI gates.

## Baseline and shortcut decision

Builds on v151 `extraction-independence-gate`, which baselined 43 `helpers=globals()` sites and froze
new `run.py` splits. This slice is the sanctioned path the v151 policy left open: "unless a future
slice introduces a typed bot runtime context and replaces helper-global wrappers directly." Python

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/bot/runtime_queries.py` | Pure `find_player` / `dict_distance` over `RuntimeState`. |
| Create | `tools/bot/bot_context.py` | Frozen `BotContext` carrying injected runtime services (`pump_one`). |
| Modify | `tools/bot/movement_runtime.py` | Drop `helpers` dict; import pure deps, take `ctx`, call intra-module funcs directly. |
| Modify | `tools/bot/run.py` | Import pure queries; add `_runtime_context()`; movement wrappers pass `ctx`, not `globals()`; remove moved defs. |
| Create | `tools/bot/test_movement_runtime.py` | Prove import/test independence from `run.py`. |
| Modify | `.maintainability/extraction-coupling-baseline.tsv` | Lower `run.py` baseline 43 → 37. |
| Modify | `docs/CODEMAP.md` | Add `runtime_queries.py`, `bot_context.py`, `test_movement_runtime.py` to the Bot row. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |
| Create | `docs/as-built/v152_bot-context-movement-decouple.md` | Close-out proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/run.py` — shrank 4269 → 4260 (within file-size allowance; coupling 43 → 37).
- [x] Other over-limit file: none.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)? Yes —
  `run.py` shrank and its coupling baseline was lowered to 37 in the same slice.

Decision:
- [x] Extract focused helper/module/test file as part of this slice (new `runtime_queries.py`,
  `bot_context.py`, `test_movement_runtime.py`, all well under 600).

Verification:
```bash
make maintainability
```

## Task 1 — Decoupling primitives

Files:
- Create: `tools/bot/runtime_queries.py`, `tools/bot/bot_context.py`

- [x] Step 1.1: Move `find_player` / `dict_distance` into `runtime_queries.py` (pure, depends only on `bot_types`).
- [x] Step 1.2: Define frozen `BotContext(pump_one=...)` in `bot_context.py` (no `run.py` import).
```bash
.venv/bin/python -m py_compile tools/bot/runtime_queries.py tools/bot/bot_context.py
```

## Task 2 — Migrate movement_runtime

Files:
- Modify: `tools/bot/movement_runtime.py`

- [x] Step 2.1: Import `find_player`, `dict_distance` directly; drop `_require_helpers` / `helpers`.
- [x] Step 2.2: Add `ctx: BotContext` to functions using `pump_one`; pure functions take no context.
- [x] Step 2.3: Make intra-module calls direct (`walk_toward`→`move_to_position`, etc.).
```bash
.venv/bin/python -m py_compile tools/bot/movement_runtime.py
```

## Task 3 — Rewire run.py wrappers

Files:
- Modify: `tools/bot/run.py`

- [x] Step 3.1: Add `from tools.bot.runtime_queries import dict_distance, find_player` and `from tools.bot.bot_context import BotContext`; remove the two moved defs.
- [x] Step 3.2: Add `_runtime_context()` returning `BotContext(pump_one=pump_one)`.
- [x] Step 3.3: Six movement wrappers pass `ctx=_runtime_context()` (or nothing for pure fns) instead of `helpers=globals()`.
```bash
.venv/bin/python -c "import tools.bot.run"
```

## Task 4 — Independence test + ratchet baseline

Files:
- Create: `tools/bot/test_movement_runtime.py`
- Modify: `.maintainability/extraction-coupling-baseline.tsv`

- [x] Step 4.1: Subprocess import test asserts `tools.bot.run` absent from `sys.modules`.
- [x] Step 4.2: Unit tests for `range_candidate_positions`, `derived_walk_max_ticks`, and `wait_for_player_move_or_accept` with a stub `BotContext`.
- [x] Step 4.3: Lower coupling baseline 43 → 37.
```bash
.venv/bin/python -m pytest tools/bot/test_movement_runtime.py -q
make maintainability
```

## Task 5 — CODEMAP, lifecycle, CI

Files:
- Modify: `docs/CODEMAP.md`, `PROGRESS.md`
- Create: `docs/as-built/v152_bot-context-movement-decouple.md`

- [x] Step 5.1: Update CODEMAP Bot row.
- [x] Step 5.2: PROGRESS lifecycle + as-built.
```bash
make ci
```

## Final verification

- [x] `.venv/bin/python -m pytest tools/bot/test_movement_runtime.py -q`
- [x] `make maintainability` (file-size + coupling 37/37)
- [x] `make bot`
- [x] `make ci`

## Deferred scope

- Migrate the remaining four runtime modules (`runtime_assertions`, `wait_runtime`, `state_ingest`,
  `coop_runtime`) off `helpers=globals()` to `BotContext` under touch-to-shrink, driving the coupling
  baseline from 37 toward 0.
