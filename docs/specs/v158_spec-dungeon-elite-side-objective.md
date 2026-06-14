# v158 Spec — Dungeon elite side objective

Date: 2026-06-14
Status: Draft

## Goal

Add a small side-objective hook to generated dungeon floors: when a generated floor contains an elite pack leader, the floor can also contain one extra objective reward chest.

## Behavior

- The objective is configured in `shared/rules/dungeon_generation.v0.json` under `elite_objective`.
- The objective chest reuses `treasure_chest` and existing chest open/loot behavior.
- The generator uses a dedicated deterministic RNG stream so objective placement does not perturb stair, monster, guarded chest, or quest reward rolls.
- Objective chests are placed only on non-boss generated dungeon floors that have at least one elite pack leader.
- The objective chest must be reachable from the generated-floor start.

## Out of scope

- Quest log UI, named objective text, minimap pins, or NPC turn-in flow.
- New interactable type or new reward table economy.
- Mandatory chest lock until elite death; v158 proves the generation hook and lab objective, not a full quest state machine.

## Verification

- Shared validation covers the `elite_objective` rule block.
- Go tests prove forced elite packs create exactly one objective chest, no-elite floors do not, and generation remains deterministic.
- Protocol bot scenario opens the generated objective chest on a pinned elite floor.
