# v472 As-Built — Room-Corridor Anchor Fallback

Date: 2026-09-28
Status: Complete
Commit: pending

## What Shipped

- `placeRoomCorridorLayout` now runs two passes of `room_corridor_pcg.max_attempts`: the
  unchanged normal pass, then an **anchor fallback pass** on its own RNG stream
  (`|room_corridor_anchor_fallback|`). The fallback runs only when the normal pass is exhausted.
- Fallback anchor rooms (`dungeon_room_anchor_fallback.go`):
  - clusters that cannot be placed merge with their nearest cluster and placement restarts
    (terminates because each merge strictly reduces the cluster count);
  - rooms grow past `room_size_max` to the cluster's padded bounding box;
  - rooms may sit flush with the perimeter wall, giving margin-line stairs/teleporters interior
    clearance (`playerRadius + 0.1`, degrading to the existing 0 inset).
- `ensureAnchorRooms` loop extracted to `placeAnchorClusterRooms` with identical RNG consumption.
- Regression tests in `dungeon_room_corridor_sweep_test.go`: 8 pinned failing seed/level pairs,
  a parallel 24-seed x 10-level sweep (widen with `ARPG_DUNGEON_SWEEP_SEEDS`), and fallback
  determinism.

## Root causes (proved with a temporary per-attempt probe)

1. `audit-09/-1`, `audit-13/-3`, `audit-17/-7`, `audit-20/-3`: all 96 attempts failed in
   anchor-room placement (distance-cluster dead zone, cluster larger than `room_size_max`, fixed
   spawn room blocking a nearby stair).
2. `audit-102/-4`, `audit-111/-8`, `audit-168/-9`, `audit-181/-3` (found by the wider sweep): all
   attempts failed reachability because a stair/teleporter on the `margin_from_wall` = 2.0 line
   landed on its own room wall, 1 unit from the perimeter.

Neither cause is fixable by retuning `shared/rules`; see the spec's decision record.

## Proof

- Before: ~3% of floors in levels -1..-10 failed (7/240 in the `audit-00..23` set; 39/1248 floors
  across the 208 seeds referenced by bot scenarios, levels -1..-6).
- After: 6000/6000 floors generate (`ARPG_DUNGEON_SWEEP_SEEDS=600`).
- Layout fingerprint of 240 floors before/after: only the 7 previously failing floors changed. No
  floor that generated before changed, so `shared/golden/dungeon_obstacles.json` was not
  regenerated and bot scenarios cannot drift.

## Validation

```bash
cd server && go test ./internal/game/ -run 'TestRoomCorridorLayout_|TestPlaceRoomCorridorLayout|Golden' -count=1
cd server && ARPG_DUNGEON_SWEEP_SEEDS=600 go test ./internal/game/ -run '^TestRoomCorridorLayout_SeedSweepAlwaysGenerates$' -count=1 -timeout 60m
cd server && ARPG_DUNGEON_SWEEP_SEEDS=3 go test -race ./internal/game/ -run 'TestRoomCorridorLayout_' -count=1
cd server && go run ./cmd/determinism-lint ./internal/game/...
cd server && CGO_ENABLED=0 go test ./...
./scripts/check-file-size-ratchet.sh && python3 ./scripts/check-extraction-coupling-ratchet.py && ./scripts/check-progress-dashboard.sh
python3 tools/validate_codemap.py && .venv/bin/pytest tools/ -q
```

Not run: `make ci` / live bot scenarios (host `make` blocked by the Xcode license; Docker/Postgres
not running). Covered by the fingerprint proof above: no floor that generated before changed.

## Deferred

- Rooms-first generation (place anchors inside generated rooms). This removes the fallback's
  reason to exist, but changes every layout.
- The elite-objective chest is placed after the layout and can sit outside rooms.
- `TestDungeonWallGridAudit` (landed on main with v471) stays report-only; its
  `generation_failures` list is now empty for its 20-seed x 10-level set, and the v472 sweep test is
  the pass/fail gate.
