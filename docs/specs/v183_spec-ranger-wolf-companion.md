# v183 Spec: Ranger Black Wolf Companion

Status: Complete
Date: 2026-06-15

## Goal

Add a Ranger summon skill that creates one server-owned black wolf companion using the v182 companion AI foundation.

## Requirements

- Add a data-driven Ranger skill, `black_wolf_companion`.
- The skill summons one active wolf companion owned by the casting Ranger.
- Recasting replaces the current wolf for that owner and skill.
- The wolf follows the Ranger, targets nearby monsters, and attacks through server-authoritative companion AI.
- The wolf renders as a black quadruped using the existing monster visual catalog/client monster rendering path.
- Skill cast, cooldown, mana spend, companion spawn, follow, and companion damage are observable through protocol bot assertions.

## Non-Goals

- No companion persistence across sessions.
- No companion command UI, inventory, equipment, leveling, or party panel.
- No Ranger multi-wolf scaling in this slice.
- No new art/plugin dependency.

## Acceptance

- `make validate-shared` passes with the new skill and visual data.
- Focused Go tests prove summon, replacement, ownership, and wolf view fields.
- Bot proof: Ranger casts `black_wolf_companion`; a black wolf companion appears, follows, and damages an enemy.
- Full `make ci` passes before commit.

## Presentation Addendum

- Active local-player companions render in a compact top-left companion row.
- Each companion slot shows a small identity block derived from the companion's shared monster visual metadata and a health bar attached to the bottom of the block.
- The row consumes generic companion entity state (`type`, `owner_id`, `monster_def_id`, `hp`, `max_hp`, visual metadata) so future companion families do not need one-off HUD wiring.
