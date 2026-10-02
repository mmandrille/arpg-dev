# v522 As-Built — Rooms-First Dungeon Anchors

- **Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
**Spec / plan:** [`v522_spec-rooms-first.md`](../specs/v522_spec-rooms-first.md) / [`v522_2026-10-01-rooms-first.md`](../plans/v522_2026-10-01-rooms-first.md)

## Result

Ordinary room-corridor geometry is now generated before ordinary stairs, cadence teleporters, guarded chests, and quest reward chests. The fixed `rules.PlayerSpawn` remains fixed and is enclosed by a designated spawn room; anchors are selected afterward from room interiors and checked against room-center and player-start reachability. Elite-objective chest placement remains post-layout, and the boss-floor early-return path is unchanged.

Generation uses purpose-separated deterministic RNG streams. The old anchor-cluster pass and fallback retry were removed after the normal room pass succeeded across the wide seed matrix both with and without the fallback compiled. Golden layout positions changed intentionally.

The runtime progression bot fixture now uses the separate `generated_wall_reachability_lab` world preset at the configured spawn point `(4,10)`. The existing `generated_wall_lab` remains unchanged for visual and unrelated tests that intentionally start at `(26,15)`. `TestGeneratedWallLabStairsOffsetMoveGoal` now uses the game's actual interactable approach-goal finder rather than constructing a straight-line offset that can land on the wrong side of a room wall.

## Evidence

- `cd server && ARPG_DUNGEON_SWEEP_SEEDS=600 go test ./internal/game -run '^TestRoomCorridorLayout_SeedSweepAlwaysGenerates$' -count=1 -timeout 60m` passed with fallback compiled: 600 seeds × 10 levels; all ordinary floors passed the normal layout path. Boss floors were checked by their dedicated generation path.
- The exact fallback-free state passed the same 600 × 10 command. Runtime reported `ok` in 448.835s.
- The focused room-layout, anchor reachability, boss-floor, and obstacle golden tests passed after fallback removal.
- `make bot scenario=28_reachable_dungeon_obstacles` passed, including descending from the focused spawn fixture.
- `make validate-shared` passed 2,269 checks and `tools/validate_codemap.py`.
- `go run ./cmd/determinism-lint -baseline ../.maintainability/determinism-baseline.tsv ./internal/game/...` passed: 77 grandfathered map-range sites across 22 files. The one-line baseline reduction records deletion of the old anchor helper's last counted map range.
- `./scripts/check-file-size-ratchet.sh` and `python3 ./scripts/check-extraction-coupling-ratchet.py` passed.
- `cd server && go test ./internal/game -count=1` passed (`ok`, 97.900s).

## Golden changes

`shared/golden/dungeon_obstacles.json` records the intended generated-geometry churn. The stair and teleporter position goldens and the guarded-chest position golden were updated from observed deterministic outputs. The teleporter travel golden records the actual arrival position adjacent to its marker. Replay assertions use the same teleporter golden.

## Changed files

- Updated generation/orchestration: `server/internal/game/dungeon_gen.go`, `server/internal/game/dungeon_room_corridors.go`, `server/internal/game/dungeon_room_layout.go`.
- Added placement ownership: `server/internal/game/dungeon_room_anchor_placement.go`, `server/internal/game/dungeon_room_spawn.go`.
- Removed fallback and anchor clustering: `server/internal/game/dungeon_room_anchor_fallback.go`, `server/internal/game/dungeon_room_anchor_rooms.go`.
- Updated regression coverage: `server/internal/game/dungeon_room_corridor_sweep_test.go`, `server/internal/game/wall_floor_lab_nav_test.go`.
- Updated shared fixtures: `shared/rules/worlds.v0.json`, `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`, and goldens `dungeon_obstacles.json`, `dungeon_stairs.json`, `dungeon_teleporters.json`, `guarded_chest_generation.json`.
- Updated ownership index and determinism baseline: `docs/CODEMAP.md`, `.maintainability/determinism-baseline.tsv`.
- Added this as-built record, the v522 spec, and the v522 plan.

## Verification limits and handoff

No client renderer capture was requested or needed for this server-generation slice. The bot proves the tested server progression flow, not visual quality or performance. Coordinator review, lifecycle/PROGRESS updates, and combined `make ci` are complete; the combined gate passed.

The worker run created `.venv/` and runtime bot artifacts in its detached checkout; those ignored files were removed with the worktree after path comparison. No worker commit, push, `make ci`, or `make ci-full` was run.
