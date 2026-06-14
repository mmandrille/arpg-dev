# v158 Plan — Dungeon elite side objective

Date: 2026-06-14
Spec: `docs/specs/v158_spec-dungeon-elite-side-objective.md`

## Adoption checklist

- Existing plugins/assets: reject. This is server-side generation and existing chest presentation.
- Existing local systems: adopt generated dungeon chests, elite pack metadata, existing treasure chest interaction, and protocol bot entity assertions.

## Tasks

1. Rules and validation
   - Add `elite_objective` to dungeon generation rules/schema.
   - Validate enabled objective loot table and placement values.
   - Extend Python shared validation.

2. Generator
   - Add objective chest metadata to `generatedChest`.
   - Place one objective chest after monster generation only if any generated monster is a pack leader.
   - Use a dedicated `seed|elite_objective|depth` RNG stream.

3. Tests and bot proof
   - Add focused Go tests for forced-elite placement and no-elite absence.
   - Add protocol bot scenario `68_dungeon_elite_side_objective.json`.

4. Close-out
   - Write as-built notes.
   - Run targeted checks, then `make ci`.
