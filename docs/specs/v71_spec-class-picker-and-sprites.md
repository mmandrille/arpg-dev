# v71 Spec — Class Picker and Sprites

Status: Complete
Date: 2026-06-11
Codename: class-picker-and-sprites

## Purpose

Expose character classes in the Godot client. Creating a new character lets the player select
Barbarian, Sorcerer, or Paladin with visible class blocks, sprite-like icons, and class stat/skill
tooltips. Existing character rows show the class icon first so players can identify heroes quickly.

## Non-goals

- No new server class rules or gameplay gates; v69/v70 own authority.
- No production bitmap art pipeline; v71 uses in-repo drawn placeholder sprites matching the
  current skill icon approach.
- No class-specific player model changes, animations, passives, or skill trees.
- No protocol/schema version bump; character APIs already carry `character_class`.

## Acceptance Criteria

- Create-character UI shows three selectable class blocks under the name input.
- Exactly one class is selected at a time; default selection is `barbarian`.
- Hovering a class block shows tooltip data including class name, starting stats, and class skill.
- Create requests include the selected `character_class`.
- Character rows in the picker display the class sprite/icon at the first part of the row before
  the textual summary.
- Debug/test state exposes selected class and row class ids/icons for client tests and bot proof.
- Existing create/reuse flows continue to work without a server change.

## Scope and Likely Files

- Client UI: `client/scripts/character_select_panel.gd`, a small class icon/control helper, and
  `client/scripts/main.gd` / `client/scripts/net_client.gd` create request plumbing.
- Client tests: `client/tests/test_coop_client.gd` character panel coverage.
- Client bot: scenario steps and `client/scripts/bot_scenario_runner.gd` if needed to select/assert
  class UI.
- Docs: plan, as-built, `PROGRESS.md`.

## Test and Bot Proof

- `make client-unit`
- `make client-smoke`
- `make ci`

Client bot proof should update a menu/create scenario to choose a non-default class and assert the
created character can start normally.

## Open Questions and Risks

- The UI must stay compact enough for the existing 430px-wide panel. If needed, widen only the
  character select panel rather than introducing a new page.
- Class presentation data can be hardcoded in a focused client helper for v71; a shared class
  presentation catalog remains a likely follow-up if class art expands.
