# v101 Spec — Undead Skeleton Poison Immunity

Status: Complete
Date: 2026-06-12
Codename: `undead-skeleton-poison-immunity`

## Purpose

Add an undead enemy that appears as a skeleton in the Godot client and proves v100 damage
resistances can express full poison immunity. The undead should still receive the poison status
from poison-applying attacks, but poison damage must mitigate to `0` because the monster has
`100%` poison resistance.

## Non-goals

- Do not add a new enemy AI, dungeon family system, or spawn progression pass.
- Do not replace the existing poison status presentation beyond whatever the client already shows
  for poisoned monsters and damage numbers.
- Do not import external art packs; use the existing deterministic generated GLB asset pipeline.
- Do not add new damage types beyond the v100 catalog.

## Acceptance Criteria

- Shared monster rules define an undead monster with `resistances.poison = 1.0`.
- The undead monster is localized in English and Spanish text catalogs.
- The monster visual catalog maps the undead to a new `monster_skeleton` scene backed by a
  deterministic committed skeleton GLB and asset manifest entry.
- The Godot client can load and spawn the skeleton monster scene through the same visual catalog
  path used by existing monsters.
- Poison Stab can apply the poisoned effect to a poison-immune undead when the attack connects.
- Poison damage events against the undead carry `damage_type: "poison"` and `damage: 0`; the
  undead's HP does not drop from poison ticks.
- A protocol bot scenario proves the undead can be poisoned and receives zero poison damage.

## Scope and Likely Files

- Shared rules: `shared/rules/monsters.v0.json`, `shared/rules/worlds.v0.json`, locale catalogs.
- Shared assets: `shared/assets/monster_visuals.v0.json` and schema, `assets/manifests/assets.v0.json`.
- Asset pipeline: `tools/assets/gen_glb.py`, generated runtime GLB under `client/assets/monsters/`.
- Server combat: poison effect application and poison DOT damage mitigation in `server/internal/game/`.
- Client presentation: monster visual loader, main scene registry, skeleton `.tscn`, animation tests.
- Bot proof: new scenario under `tools/bot/scenarios/` and any missing assertion support in
  `tools/bot/run.py`.
- Docs: v101 plan, as-built, and `PROGRESS.md` close-out.

## Test and Bot Proof

- `make validate-shared`
- `make validate-assets`
- Focused Go test for poison immunity behavior in `server/internal/game`.
- `make bot scenario=undead_skeleton_poison_immunity`
- `make client-unit` or `make client-smoke` to prove the new scene/catalog entry loads.
- Final `make maintainability` and `make ci`.
- Visual verification command to report after implementation:
  `make bot-visual scenario=undead_skeleton_poison_immunity`

deterministic generated-GLB pipeline, asset manifest, and monster visual catalog so the new
skeleton remains repo-native and CI-validatable.

## Open Questions and Risks

- Poison application currently may be gated on final damage being above zero. If so, this slice
  should change the poison status gate to use a connected hit while keeping the server-authoritative
  damage amount resistance-mitigated.
- The skeleton art is deliberately low-poly/generated; production art fidelity is deferred.
