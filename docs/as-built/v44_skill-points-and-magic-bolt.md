# v44 — Skill points and Magic Bolt

**Proves:** The first active skill loop is server-authoritative from progression grant through
rank spend, mana/cooldown mutation, projectile cast, replay, persistence, and Godot presentation.

- Character progression now grants 3 stat points per level and 1 skill point every 3 levels
  starting at level 3.
- Protocol v5 adds skill point allocation, skill casting, skill progression snapshots/deltas,
  cooldown views, and skill events.
- `magic_bolt` is data-driven from `shared/rules/skills.v0.json`; rank increases damage, cast
  spends mana, and cooldown is `2x` the current server-authored attack interval.
- Effective attack speed now combines DEX-derived speed, weapon `attack_speed`, and signed
  equipment `attack_speed_percent`; character stats expose `attack_interval_ticks`.
- Character skill points and ranks persist durably and are frozen into session-start snapshots for
  deterministic replay.
- Godot adds a one-skill panel (`K`) and one Magic Bolt slot (`Q`) with local cooldown interpolation
  reconciled from server `skill_cooldowns`.
- Protocol bot scenario `32_skill_points_and_magic_bolt.json` proves level-to-3, VIT allocation,
  skill spend, cast, cooldown rejection/recovery, damage, `/state`, reconnect, replay, and
  fresh-session persistence.
- Client bot scenario `19_skill_points_and_magic_bolt.json` proves the real Godot skill panel,
  spend button, skill bar disabled/recovery state, and directional Magic Bolt cast/reject path.
- The protocol bot `wait_ticks` helper now falls back to oscillating one-tick movement pulses when
  zero-direction pulses do not produce authoritative ticks, preserving older idle-settle scenarios.

**Explicit non-goals:** no classes, skill tree, respec/refund, passive skills, multiple active
skills, basic-attack cooldown rebalance, animation-speed scaling, mana regeneration, buffs/debuffs,
AoE/homing/summons/DOT/status effects, production skill VFX/audio, or final combat balance.
