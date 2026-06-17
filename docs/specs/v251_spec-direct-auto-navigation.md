# v251 Spec - Direct Auto Navigation

Status: Implemented
Date: 2026-06-17
Codename: direct-auto-navigation

## Purpose

Improve server-authored auto-navigation so click-to-move and auto-approach paths look more direct
and credible. Shortest diagonal routes should keep using diagonals, while obstacle detours should
prefer stable straight runs and clean corner turns instead of equal-length zig-zag routes.

## Non-goals

- No protocol changes, client-side path authority, new movement speed tuning, new navigation data
  fields, obstacle generation changes, or visual-only movement smoothing.
- No external assets/plugins. This is authoritative simulation behavior only.

## Acceptance Criteria

- Open-field unequal-axis paths still use diagonal steps for the shortest portion of the route.
- Equal-length obstacle detours prefer fewer direction changes before diagonal count tie-breaks.
- Player `move_to_intent` auto-nav continues toward the exact clicked target after the grid route,
  avoiding row/column snap drift when the target is inside the same navigation cell band.
- Pending action and skill auto-approach still dispatch after reaching their planned approach goal.
- Focused Go tests cover diagonal preference, corner detour shape, and exact-target completion.

## Scope and Likely Files

- Server: `server/internal/game/pathfind.go`, `server/internal/game/auto_nav.go`,
  `server/internal/game/sim.go`, `server/internal/game/handlers.go`.
- Tests: `server/internal/game/pathfind_test.go`, `server/internal/game/auto_nav_test.go`.
- Docs: plan and as-built/progress if the slice is finished.

## Test Proof

- `cd server && go test ./internal/game -run 'TestPlanPath|TestMoveTo'`
- `make maintainability`

## Open Questions and Risks

- No blocking questions.
- Risk: `PlanPath` is shared by monsters and companions, so path tie-breaking must preserve shortest
  path length as the primary cost to avoid AI regressions.
