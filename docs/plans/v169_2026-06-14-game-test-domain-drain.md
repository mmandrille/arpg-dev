# v169 Plan — Game Test Domain Drain

Status: Complete
Goal: Move gold auto-pickup tests out of `game_test.go`.
Architecture: Behavior-preserving test extraction only. The focused file stays in package `game`
so it can reuse existing package-local test helpers.
Tech stack: Go tests and maintainability ratchet.

## Baseline and shortcut decision

Builds on v168. No Godot plugin decision is required; this is Go test-only maintenance.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/internal/game/gold_auto_pickup_test.go` | Gold auto-pickup test domain |
| Modify | `server/internal/game/game_test.go` | Remove moved test block and gold-only helper |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `game_test.go` baseline |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v169_game-test-domain-drain.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/game_test.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Extract gold auto-pickup tests

Files:
- Add: `server/internal/game/gold_auto_pickup_test.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 1.1: Move gold auto-pickup tests and the gold-only helper.
- [x] Step 1.2: Keep shared helpers used by other domains in `game_test.go`.
```bash
cd server && go test ./internal/game -run 'Test(GoldAutoPickup|NonGoldLootDoesNotAutoPickup|ManualGoldPickupStillWorksInRange)'
```

## Task 2 — Lifecycle docs and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v169_game-test-domain-drain.md`

- [x] Step 2.1: Lower the `game_test.go` baseline.
- [x] Step 2.2: Update lifecycle docs and write the as-built note.
- [x] Step 2.3: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `cd server && go test ./internal/game -run 'Test(GoldAutoPickup|NonGoldLootDoesNotAutoPickup|ManualGoldPickupStillWorksInRange)'`
- [x] `make maintainability`
- [x] `make ci`
