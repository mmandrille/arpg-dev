# v535 — Wider room entrances

- **Status:** Complete (combined `make ci` passed 2026-10-03, 20m36s)
- **Date:** 2026-10-02
- **Codename:** `wider-entrances`
- **Baseline:** `dab60eb5` plus v532 and v533 in the working tree.
- **Dependencies:** none. v536 (`solid-dungeon-props`) must implement after this slice because both edit `dungeon_generation.v0.json` and the dungeon goldens; v537 captures after it.
- **ADRs:** ADR-0001 D2 (server owns collision), ADR-0008 (procedural dungeon), Data-Driven Configuration Policy.

## Purpose

Playtest feedback: the character is almost as wide as a room entrance, so passing through feels bad. `playerRadius` is 0.45 (0.9 wide); room-corridor widths were 1.5 and 2.0, and obstacle-door gaps 1.6, leaving 0.6–0.7 of slack. Widen them, and add a data-owned floor so a later retune cannot bring the problem back.

## Scope

1. New rule `dungeon_generation.passage_clearance.min_player_diameters` (schema-required, ≥ 1). The minimum opening is that many player diameters (`2 × playerRadius`). Rules load fails if any `room_corridor_pcg.corridor_widths`, `room_layout.corridor_width`, or enabled `obstacle_generation.doors.gap_width` is narrower.
2. Retune the three width values in `shared/rules/dungeon_generation.v0.json` to clear the floor with real slack. Room entrances are the corridor width (`doorA/doorB.width = width`), so `corridor_widths` is the main fix; `gap_width` covers wall doors.
3. If widening starves small-room floors of generation budget, resolve it in data (e.g. floor-profile room sizes) and record why; never raise the search-node cap to hide it.

## Non-goals

Player/monster collision radius, pathfinding code, wall thickness, room shape rules, props (v536), speed (cancelled).

## Acceptance criteria

1. The default rules satisfy the floor with at least 2.25 units (2.5 diameters) at every opening; a rules-derived Go test asserts this from loaded rules and over generated routes without pinning the numbers, and a fixture test proves each narrow case (below-floor corridor widths, legacy width, door gap, `min_player_diameters` < 1) is rejected.
2. Schema validation and `make validate-shared` pass; the schema requires the new rule.
3. `TestRoomCorridorLayout_SeedSweepAlwaysGenerates` (multi-seed, deep levels) passes: wider openings do not make any audited floor fail to generate.
4. Dungeon goldens, replay and rule-sensitive Go tests are updated deliberately (`make regen-golden`, test fixtures derived from rules) and pass; the Go-only obstacle golden stays a determinism contract.
5. A bot scenario proves a player walks through a generated room-to-corridor threshold with a monster nearby, and a real-camera capture shows the wider entrance.

## Surfaces and verification

`shared/rules/dungeon_generation.v0.json` and schema; `server/internal/game/dungeon_passage_clearance.go` (+ test), `rules.go`, `dungeon_generation_rules.go`; goldens under `shared/golden/`; `docs/CODEMAP.md`. Focused: `go test ./internal/game` (full package, since generation changes), `make validate-shared`, `make lint-determinism`, one bot scenario, one capture. No per-slice `make ci`.

## Risks

Wider corridors consume room interior (clearance zones), shrinking pack-placement options; deep floors with 10×8 rooms hit the 20M-node search cap. Layout goldens move for every seed.
