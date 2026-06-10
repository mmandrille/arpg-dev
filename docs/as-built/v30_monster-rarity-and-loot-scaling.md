# v30 — Monster rarity and loot scaling

**Proves:** Generated dungeon monster population can roll server-authoritative rarity that changes
challenge, XP, loot depth, protocol state, replay, bot assertions, and Godot presentation.

- `shared/rules/dungeon_generation.v0.json` now declares generated monster rarities:
  `common`, `champion`, `rare`, and `unique`, with weights `100/15/6/3`, pastel colors, challenge
  multipliers, and loot-depth offsets `+0/+1/+2/+3`.
- Shared/golden validation pins rarity tuning, scaled `dungeon_mob` HP/damage/XP, seeded generated
  roll order, and the unique `level -5 -> effective depth 8 -> 3+ loot band` case.
- Go generation rolls rarity from a separate deterministic rarity RNG stream so existing floor and
  chest layout streams do not drift.
- Generated monsters store rarity, scaled HP, scaled proactive attack damage, scaled XP reward, and
  a monster loot table selected from `abs(level) + loot_depth_offset`.
- Static/lab/world-preset monsters remain unscaled and do not emit v30 generated rarity.
- Existing protocol v1 entity `rarity` now carries monster rarity through snapshots, deltas,
  `/state`, reconnect, and replay timelines.
- Godot keeps server authority unchanged and applies a green player tint plus rarity tints on the
  existing monster model: pastel white, blue, red, and golden.
- Protocol bot scenario `21_monster_rarity_loot_scaling.json` descends into a generated dungeon,
  observes a champion mob, kills it, picks up rolled loot, and proves `/state`, reconnect, replay,
  and fresh-session persistence.
- Existing character leveling bot coverage now pins a generated-dungeon seed because v30-scaled XP
  changes the expected XP total from generated mobs.

**Explicit non-goals:** unique/set item catalogs, unique monster special drops, affixes, named
elite packs, minions, aura modifiers, boss floors, Magic Find, final item-level/depth economy,
chest rarity, production monster art/VFX/audio, and colorblind/accessibility-safe rarity treatment.
