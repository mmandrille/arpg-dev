# v334 Spec — Ranged Retreat Engagement

Status: Approved  
Date: 2026-06-24  
Codename: `ranged-retreat-engagement`

## Purpose

Fix frustrating instant kite behavior: ranged monsters with `preferred_min_range` may only
retreat after they have remained in melee attack range of their chase target for a
data-driven minimum duration (default 2 seconds at 10 Hz).

## Non-goals

- No new retreat pathing algorithm, cover seeking, or fog-aware behavior.
- No protocol/schema changes.
- No boss or companion behavior changes unless they share the same ranged retreat helper.

## Acceptance criteria

1. `shared/rules/main_config.v0.json` owns `ranged_retreat_min_melee_engagement_seconds` (default 2.0).
2. Ranged monsters track continuous melee engagement with their chase target; leaving melee range resets the timer.
3. `monsterRangedRetreatGoal` returns no goal until engagement duration elapses.
4. Existing retreat lab still proves eventual repositioning after the gate.
5. New focused test proves no material retreat during the first engagement window.

## Scope and files

- `shared/rules/main_config.v0.json` + schema
- `server/internal/game/rules.go` — config field + tick helper
- `server/internal/game/sim.go` — entity engagement field + tick update hook
- `server/internal/game/monster_ranged_positioning.go` — retreat gate
- `server/internal/game/ranged_monster_positioning_test.go`

## Test proof

```bash
cd server && go test ./internal/game/... -run RangedMonsterRetreat
make validate-shared
```
