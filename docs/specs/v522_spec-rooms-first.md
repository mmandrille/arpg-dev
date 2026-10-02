# v522 Spec: Rooms-First Dungeon Anchors

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Date:** 2026-10-01
**Codename:** `rooms-first`
**Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8`
**Dependency:** None

## Purpose

Refactor ordinary room-corridor dungeon generation so room geometry is built before anchor placement. The room layout must include a designated spawn room containing the fixed `rules.PlayerSpawn`; place the spawn/up stair there. Place other stairs, teleporters, and chests currently placed before layout inside generated room interiors. This addresses the feasibility gap that required v472's anchor fallback and removes the dependence of room geometry on fixed, pre-rolled random anchor coordinates.

The generation remains deterministic for a given seed, level, and rules. Server-generated collision, reachability, and authoritative positions remain owned by `server/internal/game`. Boss floors retain their dedicated layout and do not enter the ordinary rooms-first pipeline.

## Scope

- Generate and connect ordinary rooms before placing anchors.
- Preserve `rules.PlayerSpawn` as a fixed constraint on the designated spawn room; it is not randomized or moved.
- Place all ordinary-floor stairs (including the player spawn stair), cadence teleporters, guarded chests, and random quest-reward chests in valid room interiors, while honoring their existing shared-rule eligibility, separation, and attempt constraints.
- Preserve server reachability validation and ensure generated anchors have sufficient walkable room clearance.
- Use the established seeded `RNG`/PCG path with stable domain-separated streams; do not consume `Sim.rng` or use `math/rand`.
- Retire the v472 two-pass anchor fallback only if regression evidence demonstrates that the rooms-first algorithm no longer needs it. If it is retained, document the remaining failure class and why rooms-first cannot address it within scope.

## Non-goals

- Change boss-floor placement, content, or rules.
- Change elite-objective chest timing: it remains post-layout and may remain outside rooms, as documented by v472.
- Change dungeon-generation tuning, shared schemas, protocol, replay, client presentation, or asset pipelines.
- Guarantee existing seeds keep their old layouts. Rooms-first changes RNG ownership/order and is expected to change ordinary layouts; update deterministic goldens/fixtures only where the new behavior requires it.
- Add new external assets, plugins, dependencies, or generation systems.

## Decisions and constraints

- **Authority and security:** placement, collision, and reachability remain server-side. Generation accepts no untrusted placement inputs. Keep generated output deterministic and avoid secret/token use of gameplay RNG.
- **RNG:** retain the project `RNG` seeded from stable seed + level + purpose inputs. Independent streams should isolate room generation from anchor and population placement so unrelated call-count changes do not silently reshuffle all content.
- **Rules ownership:** reuse the current `StairPlacement`, `TeleporterPlacement`, `ChestPlacement`, `RoomCorridorPCG`, and clearance rules. Do not introduce hardcoded tuning values or silently widen their existing margins/eligibility.
- **Boss floors:** preserve `generateBossDungeonLevel` behavior and prove it with existing boss-floor tests/goldens.
- **Assets:** adopt the already vendored KayKit Dungeon kit wherever visuals are needed; borrow existing wall/entity presentation and rectangle contracts; reject new external assets, plugins, or pipelines. No visual changes are expected in this slice.

## Acceptance criteria

1. For ordinary room-corridor floors, room layout (including a designated room containing fixed `rules.PlayerSpawn`) is established before placement of the other stairs, teleporters, and pre-layout chests; these anchors are inside room interiors with valid player clearance.
2. Anchor placement continues to honor current shared-rule enabled/chance/cadence/separation/margin behavior, and generated floors pass existing server reachability validation.
3. Repeated generation with the same seed, level, and rules produces identical anchors, rooms, walls, and relevant generated content; generation does not consume `Sim.rng`.
4. The eight pinned v472 seed/level regressions generate, and all relevant anchors including pre-layout chests are inside rooms.
5. The default seed sweep (`audit-00..23`, levels -1 through -10) passes. A 600-seed x 10-level sweep (6,000 floors) passes before removing the fallback. If fallback removal is proposed, the fallback path and tests are then removed and the wide sweep is repeated against that exact state. If retaining it, explain the residual failure mode.
6. Boss-floor generation remains unchanged and passes the focused boss-floor/golden coverage.
7. Update dungeon golden/layout expectations where needed, document intentional seeded-layout churn, and keep determinism lint and maintainability gates clean.

## Likely files and ownership

- `server/internal/game/dungeon_gen.go` — ordinary anchor placement and seeded streams; preserve the independent boss-floor branch.
- `server/internal/game/dungeon_room_corridors.go`, `dungeon_room_anchor_placement.go`, and `dungeon_room_spawn.go` — room generation order, room-contained placement, and the fixed spawn-room constraint.
- `server/internal/game/dungeon_room_corridor_sweep_test.go`, `dungeon_room_layout_test.go`, and dungeon goldens — regression and deterministic behavior proof.
- `docs/CODEMAP.md` — update only if file ownership changes.
- `docs/as-built/v522_rooms-first.md` and progress lifecycle records — handoff evidence and status after implementation.

## Focused verification

- Targeted room-corridor, pinned-regression, anchor containment, deterministic repeat, and boss-floor tests in `server/internal/game`.
- `ARPG_DUNGEON_SWEEP_SEEDS=600` wide sweep over levels -1..-10 if fallback removal is proposed.
- Dungeon golden tests and Go determinism lint.
- Relevant maintainability ratchet checks and CODEMAP validation if indexed ownership changes.
- No renderer capture or bot scenario is required unless implementation changes client-visible behavior; server geometry tests cannot establish rendered appearance.

## Review notes / open planning question

The currently enabled ordinary room-corridor generator is the target. Legacy non-room-corridor placement paths should remain unchanged unless code inspection shows a shared helper must move; any such change needs an acceptance-test rationale. Before implementation, verify the desired purpose-specific RNG boundaries against current golden-test expectations and record the selected streams in the plan.
