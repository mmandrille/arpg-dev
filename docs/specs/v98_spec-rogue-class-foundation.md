# v98 Spec - Rogue class foundation

Status: Complete
Date: 2026-06-12
Codename: rogue-class-foundation

## Purpose

Introduce Rogue as the fourth playable class. Rogues are smaller and thinner than the existing
baseline hero, start with two common one-handed swords plus one health potion and one mana potion,
and can equip a one-handed weapon in `off_hand` as the foundation for dual-wield combat.

This slice makes Rogue selectable, persistent, visible, and properly equipped through the existing
authoritative server and Godot client paths. It deliberately creates the clean class/equipment base
that Poison Stab and Dash can build on in follow-up slices.

## Non-goals

- No Poison Stab damage-over-time implementation yet.
- No Dash movement-through-enemies skill implementation yet.
- No separate off-hand attack cadence, animation, or damage events yet; v98 only allows Rogue to
  equip a one-handed weapon in `off_hand` and preserves current main-hand basic attack behavior.
- No existing-character backfill or class-change UI.
- No external Godot plugin or asset-pack adoption.

## Acceptance Criteria

- Shared progression rules define `rogue` with Rogue display name and dexterity-leaning base stats.
- Character creation accepts `character_class: "rogue"` and session-start progression uses Rogue base
  stats.
- Godot character creation UI exposes Rogue as a fourth class option, emits Rogue on create, and
  character rows resolve Rogue names/tooltips instead of falling back to Barbarian.
- Shared class presentation data maps Rogue to a dedicated `character_rogue_v0` model, and the
  generated model is visibly smaller/thinner than the baseline humanoid while retaining the required
  animation bones.
- New Rogue characters receive durable starter equipment: a common one-handed sword in `main_hand`,
  a common one-handed sword in `off_hand`, one `red_potion`, and one `blue_potion`.
- Rogue can equip a one-handed sword-class weapon in `off_hand`; non-Rogue classes still cannot equip
  weapons in `off_hand`, and two-handed main-hand equipment still blocks `off_hand`.
- Shared validation, focused Go tests, client unit tests, and protocol bot proof cover the new class
  and starter kit.

## Scope and Likely Files

- `shared/rules/character_progression.v0.json` and related validation/goldens.
- `shared/rules/item_templates.v0.json`, `shared/assets/item_presentations.v0.json`, and class/item
  schemas as needed for the starter Rogue sword and Rogue icon shape.
- `server/internal/http/starter_loadout.go` and tests for Rogue starter inventory.
- `server/internal/game/sim.go` or adjacent equipment helpers/tests for Rogue off-hand weapon rules.
- `client/scripts/character_select_panel.gd`, client tests, class presentation data, generated GLB
  assets, and the asset manifest for Rogue presentation.
- `tools/bot/scenarios/47_rogue_class_foundation.json` and protocol bot tests.
- `docs/plans/`, `docs/as-built/`, and `PROGRESS.md` lifecycle docs.

## Test and Bot Proof

- `make gen-assets`
- `make validate-shared`
- Focused Go tests for Rogue class stats, off-hand equip rules, and starter loadout.
- Focused client unit tests for the fourth class option and Rogue model resolution.
- `make bot scenario=47_rogue_class_foundation`
- Final `make maintainability` and `make ci`.

## Open Questions and Risks

- Rogue base stats are intentionally provisional: high DEX, normal VIT/MAGIC, lower STR. Future
  combat tuning can adjust the JSON rules without code changes.
- The user-requested off-hand 1.5x attack speed is deferred with Poison Stab/Dash because it needs
  independent attack timing state and combat event proof. v98 keeps the equipment contract clean
  first.
- Poison Stab and Dash remain high-priority follow-ups after v98.
