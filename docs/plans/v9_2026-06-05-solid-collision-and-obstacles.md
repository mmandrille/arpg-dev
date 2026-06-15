# Solid Collision and Obstacles - Implementation Plan

Goal: Add server-authoritative collision against live monsters and static walls, with a bot scenario
that proves blocked and routed movement.

Architecture: Shared world rules define static wall obstacles; the Go sim owns collision; the
Godot client renders simple local wall props and reconciles to server positions.

Tech stack: Go sim/tests, shared JSON schemas, Python protocol bot, Godot GDScript smoke.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v9_spec-solid-collision-and-obstacles.md` | Feature contract |
| Create | `docs/plans/v9_2026-06-05-solid-collision-and-obstacles.md` | Implementation checklist |
| Modify | `shared/rules/worlds.v0.schema.json` | Allow `wall` obstacle entries |
| Modify | `shared/rules/worlds.v0.json` | Add `collision_lab` preset |
| Modify | `server/internal/game/rules.go` | Parse and validate wall sizes |
| Modify | `server/internal/game/sim.go` | Block/slide movement against obstacles |
| Modify | `server/internal/game/game_test.go` | Collision regression tests |
| Create | `tools/bot/scenarios/03_collision_lab.json` | End-to-end collision scenario |
| Modify | `tools/bot/run.py` | Movement assertion actions |
| Modify | `client/scripts/main.gd` | Simple wall visuals from shared world rules |
| Modify | `PROGRESS.md` | v9 completion record |
## Task 1: Shared World Data

- [x] Step 1.1: Spec and plan v9.
- [x] Step 1.2: Extend world schema with `wall` entries and `size`.
- [x] Step 1.3: Add `collision_lab` world preset.
- [x] Step 1.4: Validate shared rules.

## Task 2: Authoritative Collision

- [x] Step 2.1: Add parsed wall data and validation in `rules.go`.
- [x] Step 2.2: Add collision constants and helper functions in `sim.go`.
- [x] Step 2.3: Change movement to block or axis-slide when candidate position collides.
- [x] Step 2.4: Add Go tests for wall blocking, monster blocking, dead monster non-blocking, and routing.

## Task 3: Bot Scenario

- [x] Step 3.1: Add `03_collision_lab.json`.
- [x] Step 3.2: Add bot actions for movement attempts and position assertions.
- [x] Step 3.3: Run protocol bot.

## Task 4: Client Visuals

- [x] Step 4.1: Load static walls from shared world data for current world.
- [x] Step 4.2: Render simple box wall props in the Godot scene.
- [x] Step 4.3: Keep prediction reconciliation working when the server blocks movement.

## Task 5: Documentation and Verification

- [x] Step 5.1: Update `PROGRESS.md` with v9 as-built notes.
- [x] Step 5.2: Run focused Go tests.
- [x] Step 5.3: Run `make validate-shared`.
- [x] Step 5.4: Run `make bot`.
- [x] Step 5.5: Run `make client-smoke`.
- [x] Step 5.6: Run `make ci`.
