# v63 As-Built: Runtime Sim Error Construction

## What shipped

- `game.NewSim` now returns `(*Sim, error)` for default-world construction instead of panicking.
- `game.MustNewSim` preserves explicit panic-on-invalid-fixture behavior for tests.
- Existing terse Go sim tests use `MustNewSim`, while runtime paths continue to use
  error-returning constructors.
- `TestNewSimReturnsDefaultWorldError` proves invalid default-world setup returns an error without
  crashing the process.
- A stale `dungeon_levels` bot/golden expectation for generated level -2 gold was removed after CI
  exposed that the shared `dungeon_stairs` fixture currently has no level -2 loot; dedicated v49
  scenarios continue to own gold pickup behavior.

## What it proved

The v60 backend review's default sim construction crash path is closed without changing gameplay,
protocol, replay, persistence, or client behavior. Future runtime callers must make construction
errors explicit instead of inheriting a panic-prone default API.

## Deferred

- Broader constructor cleanup remains future work if another panic wrapper appears in runtime code.
