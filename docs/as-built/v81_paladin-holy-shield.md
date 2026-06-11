# v81 As-built - Paladin Holy Shield

Status: Complete

## Shipped

- Added Paladin `holy_shield` as a data-driven `area_stat_buff` skill with rank-scaled mana,
  requirements, range, radius, duration, and armor/block percent bonuses.
- Extended shared validation and Go rule loading for a closed `area_stat_percent_buff` effect.
- Holy Shield casts now affect living allied players in range, emit start/end events, expose active
  `effect_ids` on authoritative player entity state, and expire deterministically.
- Defensive stat breakdowns include Holy Shield as a skill-effect source while preserving Rage,
  Heal, and Magic Bolt behavior.
- The Godot client renders a code-native gold shield/shine around every entity carrying the
  `holy_shield` effect id and removes it when server state clears the effect.
- Holy Shield appears in the existing Rage-style status-effect/hotbar presentation with its shared
  presentation label.
- Added a protocol bot scenario and focused client/server tests for effect events, visible state,
  defensive stat improvement, expiry, and client presentation.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'HolyShield|Rage|Skill'`
- `make bot scenario=43_paladin_holy_shield.json`
- `make client-unit`
- `make ci`

## Deferred

- Production VFX/audio, absorb shield resources, thorns/reflect, taunt, invulnerability, and a
  generalized buff/debuff taxonomy remain future work.
