# v159 Spec — Kill-gated elite objective

Date: 2026-06-14
Status: Complete
Codename: kill-gated-elite-objective

## Purpose

Turn the v158 generated elite objective from a free bonus chest into a small server-authored side objective: elite-objective chests stay closed until the player has killed at least one generated elite pack leader on that floor.

## Non-goals

- Quest journal UI, named objective text, minimap pins, NPC turn-in flow, or durable quest state.
- New chest visuals, new interactable definitions, or a new reward economy.
- Locking ordinary guarded chests, random quest reward chests, boss chests, or town debug chests.

## Acceptance criteria

- Generated elite-objective chests retain their objective identity when converted into runtime interactable entities.
- Activating an elite-objective chest before any generated pack leader has been killed rejects the action with a stable reason and does not open the chest or drop loot.
- After a generated elite pack leader is killed, the same chest opens through the existing chest path and drops loot.
- Non-objective treasure chests keep their existing open-on-activation behavior.
- The protocol bot proof for `68_dungeon_elite_side_objective` covers the reject-then-kill-then-open flow.

## Scope and likely files

- `server/internal/game/sim.go` / focused interactable helper: preserve generated objective metadata and gate activation.
- `server/internal/game/dungeon_elite_objective_test.go`: focused Go coverage for the objective chest runtime gate.
- `tools/bot/scenarios/68_dungeon_elite_side_objective.json`: update the scenario to prove the kill gate.
- `docs/as-built/v159_kill-gated-elite-objective.md` and `PROGRESS.md`: lifecycle closeout.

## Test and bot proof

- `cd server && go test ./internal/game -run 'TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make ci`

Visual/client verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```

## Open questions and risks

- No blocking product questions. The conservative behavior is server-owned gating with the existing treasure chest presentation and reward table.
- Clearing every generated elite leader remains deferred; v159 is a small “kill one leader, claim side reward” proof.
- `sim.go` is over the 600-line target and close to its baseline allowance; this slice should extract the interactable activation path instead of growing it.
