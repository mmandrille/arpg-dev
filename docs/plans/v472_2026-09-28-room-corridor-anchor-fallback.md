# v472 Plan — Room-Corridor Anchor Fallback

Status: Complete
Goal: Every generated dungeon floor builds for every seed, without changing floors that build today.
Spec: [`v472_spec-room-corridor-anchor-fallback.md`](../specs/v472_spec-room-corridor-anchor-fallback.md)
Tech stack: Go server (`internal/game`), SDD docs. No shared rules, protocol, client, or bot changes.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/dungeon_room_corridors.go` | Two-pass attempt loop (normal, then anchor fallback) with a separate RNG stream per pass; thread the fallback flag into room packing. |
| Modify | `server/internal/game/dungeon_room_anchor_rooms.go` | Extract `placeAnchorClusterRooms` (same RNG consumption as before); route fallback clusters to perimeter-flush oversized rooms. |
| Create | `server/internal/game/dungeon_room_anchor_fallback.go` | Nearest-cluster merge loop and `oversizedRoomContainingPoints`. |
| Create | `server/internal/game/dungeon_room_corridor_sweep_test.go` | Pinned failing seeds, parallel seed sweep gate, fallback determinism. |
| Modify | `docs/CODEMAP.md`, `docs/progress/*`, `PROGRESS.md` | Index the new files and slice. |

## Steps

1. Reproduce with a temporary probe around each attempt: classify failures as anchor placement,
   room count, or reachability. Result: original 4 = anchor placement (all 96 attempts); sweep
   extras = reachability of perimeter-edge anchors (all 96 attempts).
2. Write failing tests (pinned seeds + sweep + determinism).
3. Implement the fallback pass; verify pinned + sweep green.
4. Fingerprint 240 floors (`audit-00..23` x -1..-10) before/after: only previously failing floors
   may change.
5. Wide sweep (`ARPG_DUNGEON_SWEEP_SEEDS=600`, 6000 floors) to look for a third failure mode.
6. Determinism lint, goldens, maintainability, CODEMAP, full Go suite, Python tools tests.

## Verification

```bash
cd server && go test ./internal/game/ -run 'TestRoomCorridorLayout_|TestPlaceRoomCorridorLayout|Golden' -count=1
cd server && ARPG_DUNGEON_SWEEP_SEEDS=600 go test ./internal/game/ -run '^TestRoomCorridorLayout_SeedSweepAlwaysGenerates$' -count=1 -timeout 60m
cd server && go run ./cmd/determinism-lint ./internal/game/...
cd server && go test ./...
./scripts/check-file-size-ratchet.sh && python3 ./scripts/check-extraction-coupling-ratchet.py && ./scripts/check-progress-dashboard.sh
python3 tools/validate_codemap.py
```

`make` is unavailable on this host until the Xcode license is accepted; the commands above are the
underlying targets.
