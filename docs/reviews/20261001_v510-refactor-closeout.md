# v510 review refactor closeout

**Baseline:** `main` at `9171686f` (v510 review commit).  
**Result:** One low-risk CI reliability fix landed as `16eab682` (`fix: start Postgres before database tests`). The final post-refactor `make ci` passed in 6m56s. No gameplay or protocol behavior changed.

The database-order recommendation was implemented by moving `make db-up` and `test_db.sh ensure` into CI step 6 before Go tests, exporting the database URL for that test process, and removing the path that unset it and silently skipped DB-backed coverage. Shell syntax, diff checks, the maintainability gate, and the final combined CI gate passed.

## Overview recommendations

| # | Classification | Outcome |
|---:|---|---|
| 1 | `future-plan` | Define durable input acceptance and append-failure semantics; add an end-to-end replay regression. |
| 2 | `minor-commit` | Complete: `16eab682`; final `make ci` passed. |
| 3 | `future-plan` | Add replay cancellation and select a maximum timeline horizon. |
| 4 | `future-plan` | Persist the rules/content revision and debug-mode input required for deterministic replay. |
| 5 | `future-plan` | Validate reconnect snapshots against the strict v8 schema. |
| 6 | `future-plan` | Add a Go JSON-tag/schema drift check. |
| 7 | `future-slice` | Prove active boss-warning restoration across reconnect. |
| 8 | `future-plan` | Extract a typed event router from `main.gd:_apply_delta` with ordering/state-sync tests. |
| 9 | `future-plan` | Extract typed domain modules from bot dispatch and shared validation. |
| 10 | `future-slice` | Complete v506/v508/v509 memory, final-cost, readability, and reconnect evidence. |

## Companion-report recommendations

**Backend:** #1 maps to overview #1 (`future-plan`); #2 maps to #3 (`future-plan`); #3 maps to #4 (`future-plan`); #4 bounded DB contexts/lock ownership is `future-plan`; #5 further `sim.go` extraction is `future-plan`.

**Client:** #1 maps to overview #8 (`future-plan`); #2 maps to #7 (`future-slice`); #3 final v508/v509 render-cost capture and #4 v506 memory/hands-on town proof map to #10 (`future-slice`); #5 typed `BotFacade` replacement is `future-plan`.

**Shared/tooling/process:** #1 maps to overview #2 (complete); #2 maps to #5 (`future-plan`); #3 maps to #6 (`future-plan`); #4 maps to #9 (`future-plan`); #5 maps to #10 (`future-slice`).

The remaining queue needs either behavior/acceptance decisions, coordinated architecture work, or additional gameplay/performance evidence. No further small refactor was justified without broadening scope.
