# v152 As-Built: Bot Context Movement Decouple

Date: 2026-06-14
Codename: `bot-context-movement-decouple`
Spec: [`v152_spec-bot-context-movement-decouple.md`](../specs/v152_spec-bot-context-movement-decouple.md)
Plan: [`v152_2026-06-14-bot-context-movement-decouple.md`](../plans/v152_2026-06-14-bot-context-movement-decouple.md)

## What shipped

The first real paydown of the `helpers=globals()` debt v151 baselined. `movement_runtime` no longer
receives `run.py`'s namespace; it now depends only on leaf modules and a typed context.

- New `tools/bot/runtime_queries.py` — pure `find_player` / `dict_distance` over `RuntimeState`.
- New `tools/bot/bot_context.py` — frozen `BotContext(pump_one=...)`, the decoupling primitive the
  v151 policy named ("typed bot runtime context") in place of `globals()` laundering.
- `movement_runtime.py` migrated: pure deps imported directly, the WebSocket pump injected as
  `ctx: BotContext`, and the four intra-module calls that previously bounced out to `run.py` and back
  now call directly. The module imports nothing from `tools.bot.run`.
- `run.py` keeps its public movement wrappers (unchanged signatures, no scenario churn) but builds
  `BotContext(pump_one=pump_one)` via `_runtime_context()`; `find_player`/`dict_distance` are now
  imported and re-exported so unmigrated modules' `globals()` lookups still resolve.
- New `tools/bot/test_movement_runtime.py` enforces the property the coupling ratchet's regex cannot:
  a subprocess import asserts `tools.bot.run` is absent from `sys.modules`, and runtime functions are
  exercised with a stub `BotContext`.

## Proof

- `coupled helper injections: 1 files, 37 occurrences` (down from 43); coupling baseline lowered to 37.
- `tools/bot/run.py`: 4269 → 4260 lines; grandfathered trend 65592 → 65583.
- `.venv/bin/python -c "import tools.bot.run"` clean (no circular import).
- `tools/bot/test_movement_runtime.py`: 4 passed (incl. subprocess import-independence).
- `make maintainability` and `make ci` green.

## Why it matters / deferred

This converts v151's frozen debt into an actively-shrinking baseline with a proven recipe. The other
four runtime modules (`runtime_assertions`, `wait_runtime`, `state_ingest`, `coop_runtime`) keep the
remaining 37 `helpers=globals()` sites and migrate to `BotContext` under touch-to-shrink as future
slices touch them — driving the baseline toward 0. `BotContext` stays minimal (one field) until a
migrated module needs more.
