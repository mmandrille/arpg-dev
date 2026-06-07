# Idea: boss floor reward gate

Status: idea / deferred

This note captures a possible dungeon progression slice to consider later. It is not an approved
slice, spec, or implementation plan.

## Concept

Add the first deterministic boss-floor milestone to the dungeon. The initial target is level `-5`,
where the generated floor contains one treasure chest before a boss arena, one boss using a
server-authoritative telegraphed attack pattern, and down stairs that start locked. Opening the
chest grants loot through existing inventory and loot flows. Killing the boss unlocks the stairs
and allows descent to level `-6`.

This follows the direction in `docs/adr/0009-boss-floors-and-timing-mechanics.md`, but that ADR is
still `Proposed`, so its open questions should be resolved before this becomes a spec.

## Suggested first slice shape

- Use a thin vertical slice for level `-5`, not the whole boss system.
- Detect boss floors with the ADR rule: `levelNum < 0 && abs(levelNum) % 5 == 0`.
- Add exactly one `treasure_chest`, one boss, and one locked down stair on level `-5`.
- Open the chest via existing `action_intent`; roll from a simple boss-chest loot table.
- Assemble one boss template from shared rules data.
- Add one boss pattern with `telegraph`, `active`, and `recovery` phases measured in sim ticks.
- Emit authoritative boss phase events for client presentation.
- Keep active-phase damage server-owned and deterministic under replay.
- Reject `descend_intent` against locked stairs with an observable reason.
- Unlock down stairs when the boss dies.
- Add a simple client-side telegraph indicator from server events.
- Add golden fixtures for level `-5` layout and phase tick boundaries.

## Likely touch surfaces

- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/interactables.v0.json`
- `shared/rules/loot_tables.v0.json`
- New shared boss template and boss pattern rule files
- `shared/protocol/` schemas for boss phase events and locked-descend feedback
- `shared/golden/boss_floor_-5.json`
- `server/internal/game/dungeon_gen.go`
- `server/internal/game/sim.go`
- `server/internal/game/types.go`
- `client/scripts/main.gd`
- `client/scripts/bot_scenario_runner.gd`
- `tools/bot/scenarios/17_boss_floor_reward_gate.json`

## Bot scenario idea

Create a protocol bot scenario such as `17_boss_floor_reward_gate.json`:

- Start in town.
- Descend through dungeon levels to level `-5`.
- Open the treasure chest and pick up the reward.
- Trigger one telegraphed boss pattern.
- Move out of the danger zone before the active phase applies damage.
- Kill the boss.
- Verify the down stairs unlock.
- Descend to level `-6`.

The scenario should prove `/state`, reconnect resume, and replay.

## Acceptance criteria ideas

- Level `-5` is generated as a boss floor from deterministic rules.
- Boss floor contains exactly one chest, one boss, and locked down stairs.
- Chest loot follows existing server-authoritative inventory and pickup rules.
- Locked stairs reject descent until the boss is dead.
- Boss phase timing is tick-based and replay-stable.
- Every damaging active phase is preceded by a telegraph phase.
- Client renders a simple floor telegraph from authoritative events.
- Killing the boss unlocks down stairs.
- Golden fixture pins layout and phase tick boundaries.
- `make ci` stays green.

## Open questions

- Is ADR-0009 accepted enough to implement from, or should it be finalized first?
- Should chest loot and boss kill loot use separate tables?
- Should level `-5` include trash mobs, or should the first slice use only the boss?
- Should boss-floor waypoint placement stay standard, move post-boss, or be deferred?
- Should `boss_killed` be a separate event, or is existing `monster_killed` plus boss metadata enough?
- Should reconnect snapshots include compact boss pattern progress, or is replay-derived state enough?

## Non-goals for the first version

- No full procedural boss catalog.
- No enrage phases, multiple patterns, adds, block/parry, or co-op scaling.
- No production boss art, boss health bar polish, audio, or animation set.
- No durable boss map, monsters, corpses, floor drops, HP, or current level across fresh sessions.
- No quest or NPC integration.
- No broad depth scaling beyond the minimum needed for level `-5`.
