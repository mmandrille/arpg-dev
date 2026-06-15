# v198 Spec: Mercenary Foundation

## Goal

Introduce a server-authored mercenary companion archetype that reuses the existing companion AI, creating a concrete foundation for future hiring, persistence, equipment, and commands.

## Player-visible behavior

- A mercenary companion can exist as an owned companion entity.
- The mercenary follows the owning player and assists against nearby hostile monsters.
- Mercenary identity, presentation, and combat stats are data-authored in shared rules/assets/text catalogs.

## Scope

- Add `mercenary_guard` monster data with no drops or XP.
- Add mercenary visual and localized names.
- Add a compact `mercenary_foundation_lab` world with an owned mercenary and a target monster.
- Add focused server and protocol bot coverage for identity, movement, and companion-sourced damage.

## Out of Scope

- Hiring UI, town NPC services, gold cost, persistence, roster management, equipment, commands, death recovery, and leveling.
- New art assets or plugins.

## Acceptance Criteria

- Shared validation accepts the new monster, visual, i18n, and world references.
- Server tests prove the mercenary appears as an owned companion with authored stats.
- Protocol bot scenario `86_mercenary_foundation.json` proves the mercenary follows and damages a target.
- `make validate-shared`, focused server tests, the bot scenario, and final `make ci` pass.

## Godot Plugin Adoption

Rejected for this slice. The mercenary uses existing companion rendering and `monster_dummy` presentation; no new client UI or art asset is introduced.
