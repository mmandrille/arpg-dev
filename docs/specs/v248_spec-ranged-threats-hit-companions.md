# v248 Spec - Ranged Threats Hit Companions

Status: Complete
Date: 2026-06-17
Codename: ranged-threats-hit-companions

## Purpose

Let ranged monsters punish companions that are actively engaging them. Melee monsters already hit
engaged companions; ranged monsters should be able to aim their projectiles at those same companion
targets instead of always shooting the player.

## Non-goals

- No companion threat table, taunt system, ranged AI retarget policy beyond the existing engaged
  companion rule, projectile avoidance, new monster types, damage tuning, or client VFX.
- No protocol schema changes; existing projectile target IDs and companion combat events are enough.

## Client Asset / Plugin Decision

- **Adopt:** Existing projectile, companion damage/death, and bot combat-event proof paths.
- **Borrow:** Existing `monsterEngagedCompanionTarget` rule from melee attacks.
- **Reject:** External assets/plugins and visual projectile changes.

## Acceptance Criteria

- A ranged monster picks an engaged companion target when that companion is in ranged attack range.
- Monster-owned projectile entity views target that companion instead of the player.
- Monster-owned projectiles collide with the targeted companion and resolve through existing
  `damageCompanionByMonster`.
- Existing player-targeted ranged monster shots continue to work.
- Focused Go tests prove projectile targeting and companion damage.
- A protocol bot scenario observes an archer-sourced companion combat event.

## Scope and Likely Files

- Server: `server/internal/game/sim.go`, `server/internal/game/monster_companion_combat.go` or a
  small helper file, focused test file.
- Shared worlds/scenario: add a compact ranged companion lab if needed, and a protocol bot scenario.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `cd server && go test ./internal/game -run 'RangedMonster.*Companion|RangedMonsterProjectile' -count=1`
- `make bot scenario=91_ranged_threats_hit_companions.json`
- `make validate-shared`
- `make maintainability`

## Open Questions and Risks

- No blocking questions.
- Risk: projectile hit ordering must not let monster-owned projectiles hit unintended companions.
  This slice only allows the targeted companion for monster-owned projectiles.
