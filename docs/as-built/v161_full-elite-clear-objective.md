# v161 As-built — Full elite clear objective

Date: 2026-06-14
Status: Complete

## What shipped

- Elite-objective chests now remain locked while any generated pack leader on the active level is
  alive.
- The rejection reason remains `elite_objective_incomplete`; successful completion still uses the
  existing treasure chest open and loot-drop path.
- Added focused Go coverage proving a partial leader clear still rejects and a full leader clear
  opens the chest.
- Updated the protocol bot objective scenario to target pack leaders directly and use an existing
  Barbarian debug progression setup so the proof remains about objective completion rather than
  low-level survivability.

## Proof

- `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresAllLeaderKills|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make maintainability`
- `make ci`

## Deferred

- Quest text, chest presentation polish, minimap pins, and special objective UI remain deferred to
  follow-up presentation work.
