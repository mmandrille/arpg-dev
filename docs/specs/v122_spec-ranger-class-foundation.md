# v122 Spec - Ranger class foundation

Status: Complete
Date: 2026-06-13
Codename: ranger-class-foundation

## Purpose

Introduce Ranger as a playable bow class. Rangers are tall, thin, hooded physical ranged attackers
with dexterity-leaning base stats, a green bow class logo, and a starter bow equipped through the
same authoritative character creation path used by the existing classes.

This slice makes Ranger selectable, persistent, visible, and properly equipped. It deliberately
defers Ranger active skills to the next two slices while proving that the class can use the existing
ranged basic attack path from creation.

## Non-goals

- No Ranger active skills yet.
- No new projectile protocol fields.
- No external Godot plugin or asset-pack adoption.
- No existing-character backfill or class-change UI.

## Acceptance Criteria

- Shared progression rules define `ranger` with display name `Ranger` and base stats
  `str: 4`, `dex: 8`, `vit: 5`, `magic: 3`.
- Character creation accepts `character_class: "ranger"` and session-start progression uses Ranger
  base stats.
- New Ranger characters receive durable starter equipment: `starter_ranger_bow` equipped in
  `main_hand`, one `red_potion`, and one `blue_potion`.
- `starter_ranger_bow` is a two-handed ranged weapon backed by shared rules, item presentation, and
  item visuals.
- Shared class presentation maps Ranger to `character_ranger_v0`, a generated tall, thin hooded
  model with required animation bones, and a green bow icon.
- Godot character creation UI exposes Ranger as a fifth class option and resolves Ranger names in
  character rows.
- Protocol bot proof creates a Ranger, asserts class stats and starter equipment, then observes a
  ranged basic-attack damage event.

## Scope and Likely Files

- `shared/rules/character_progression.v0.json`, `shared/rules/items.v0.json`,
  `shared/rules/item_templates.v0.json`, and related schemas/validation.
- `shared/assets/class_presentations.v0.json`, `shared/assets/item_presentations.v0.json`,
  `shared/assets/item_visuals.v0.json`, and `assets/manifests/assets.v0.json`.
- `tools/assets/gen_glb.py` and generated Ranger GLB.
- `server/internal/http/starter_loadout.go` and tests.
- `server/internal/game/rules.go`, `server/internal/game/game_test.go`.
- `client/scripts/character_select_panel.gd`, `client/scripts/class_icon.gd`, client tests.
- `tools/bot/scenarios/58_ranger_class_foundation.json` and protocol bot discovery tests.

## Test and Bot Proof

- `make gen-assets`
- `make validate-shared`
- Focused Go tests for class stats, class weapon rules, and starter loadout.
- `make client-unit`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=58_ranger_class_foundation`
- Final `make maintainability` and `make ci`.

## Open Questions and Risks

- v121 is already reserved by an approved draft not reflected in `PROGRESS.md`; Ranger uses v122 to
  avoid duplicate slice numbers.
- Ranger combat tuning is intentionally provisional and data-driven.
- The generated hooded model is deterministic placeholder art, not final production art.
