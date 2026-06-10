# v31 — Combat stat effects and feedback

**Proves:** Player and monster combat stats now drive deterministic authoritative combat outcomes
and Godot renders those outcomes from server event metadata.

- Shared combat rules now define base hit/crit values, minimum non-blocked damage, and the global
  `75%` block cap.
- Monster rules support hit chance, crit chance, crit damage, armor, and block chance, with explicit
  combat-lab targets for miss, crit, armor-floor, block, and monster-side proofs.
- Go combat uses one deterministic resolution path for melee, projectiles, proactive monster
  attacks, and retaliation: hit roll, block roll, damage roll, crit roll, armor mitigation, and
  minimum damage.
- Misses and blocks emit combat events but do not mutate HP, trigger retaliation, kill entities,
  drop loot, or award XP; successful non-blocked hits always deal at least `1`.
- Protocol v1 combat events now expose source/target ids, outcome, raw and mitigated damage,
  `blocked`, and `critical`; progression snapshots/deltas expose effective stat breakdown rows.
- Equipped base stats, rolled equipment stats, derived character formulas, caps, and clamps are
  visible through server-owned stat breakdowns and the Godot character stats panel.
- Godot floating combat text now renders normal damage, crits, misses, and blocks from authoritative
  events, with a persisted settings toggle to suppress the presentation only.
- Protocol scenario `22_combat_stat_effects.json` and client scenario `11_combat_feedback.json`
  prove the complete path through `/state`, reconnect, replay, and headless Godot presentation.

**Explicit non-goals:** attack-speed gameplay, movement-speed gameplay, spells, mana consumers,
status effects, affix grammar, polished comparison UI, enemy equipment inventories, production
combat VFX/audio, and Protobuf migration.
