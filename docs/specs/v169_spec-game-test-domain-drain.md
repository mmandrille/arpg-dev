# v169 Spec — Game Test Domain Drain

Status: Complete
Date: 2026-06-14
Codename: `game-test-domain-drain`

## Purpose

Move one coherent test domain out of `server/internal/game/game_test.go` into a focused test file
without changing gameplay behavior. This slice targets gold auto-pickup coverage.

## Non-goals

- No production code changes.
- No gameplay, rules, protocol, or golden fixture changes.
- No broad rewrite of `game_test.go`.

## Acceptance Criteria

- Gold auto-pickup tests live in a focused `gold_auto_pickup_test.go` file.
- Shared helpers used outside the domain remain available in `game_test.go`.
- `game_test.go` line count and maintainability baseline are lowered.
- Focused Go tests, `make maintainability`, and `make ci` pass.

## Scope and Files

- Tests: `server/internal/game/game_test.go`, `server/internal/game/gold_auto_pickup_test.go`
- Maintainability: `.maintainability/file-size-baseline.tsv`
- Docs: `PROGRESS.md`, `docs/as-built/v169_game-test-domain-drain.md`

## Test and Bot Proof

Focused proof:

```bash
cd server && go test ./internal/game -run 'Test(GoldAutoPickup|NonGoldLootDoesNotAutoPickup|ManualGoldPickupStillWorksInRange)'
```

Full proof: `make ci`.

## Open Questions and Risks

- Risk: moving helpers can break unrelated tests. Keep broadly shared helpers in `game_test.go`;
  move only the gold-only helper.
