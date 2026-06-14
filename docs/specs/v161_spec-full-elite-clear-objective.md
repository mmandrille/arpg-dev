# v161 Spec — Full elite clear objective

Date: 2026-06-14
Status: Complete
Codename: full-elite-clear-objective

## Purpose

Strengthen generated elite side objectives so an elite-objective chest unlocks only after every
generated elite pack leader on that floor is dead. This turns the v159 "kill any leader" proof into
a clearer floor objective while keeping the same chest reward, rejection reason, and protocol shape.

## Non-goals

- No quest journal, minimap pins, objective text UI, special chest art, or durable quest state.
- No changes to elite pack generation, monster rarity tuning, chest loot tables, or ordinary chest
  behavior.
- No new protocol schema fields.
- No broad Python bot runtime refactor; the protocol proof may use a narrow leader selector on the
  existing `kill_monsters` action.

## Acceptance criteria

- An elite-objective chest rejects with `elite_objective_incomplete` while any generated elite pack
  leader on that level is still alive.
- Killing one leader no longer unlocks the chest when another generated elite pack leader remains.
- Once all generated elite pack leaders on the level are dead, the same chest opens through the
  existing treasure chest path and drops loot.
- Ordinary non-objective treasure chests keep their existing behavior.
- The protocol bot proof for `68_dungeon_elite_side_objective` targets generated pack leaders and
  remains green for the clear-all objective contract.

## Scope and likely files

- `server/internal/game/interactables.go`: update elite objective lock check.
- `server/internal/game/dungeon_elite_objective_test.go`: focused multi-leader gate coverage.
- `tools/bot/scenarios/68_dungeon_elite_side_objective.json`: update description/assertions if
  needed to match the new objective wording.
- `docs/as-built/v161_full-elite-clear-objective.md` and `PROGRESS.md`: lifecycle closeout.

## Test and bot proof

- `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresAllLeaderKills|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make ci`

Visual/client verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```

## Open questions and risks

- No blocking product questions. The conservative rule is server-owned clear-all completion using
  existing generated pack leader metadata.
- Scenario runtime should stay bounded; the bot proof continues to use the pinned objective seed
  and the Go test owns the multi-leader edge.
