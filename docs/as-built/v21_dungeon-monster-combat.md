# v21 — Dungeon monster combat

**Proves:** Generated dungeon floors can be dangerous without changing the authoritative
client/server boundary or adding client-side combat authority.

- Shared monster rules define `dungeon_mob` as a chase monster with proactive melee attack damage
  and tick-based cooldown.
- Dungeon generation places deterministic `dungeon_mob` entities on negative dungeon levels only;
  town level `0` remains monster-free.
- Server `Sim.Tick` advances monster chase, then proactive monster attacks, then projectiles,
  preserving deterministic replay order.
- Proactive attacks emit existing `player_damaged` / `player_killed` events, so current Godot
  player hit/death reactions work without a protocol schema change.
- `shared/golden/dungeon_monster_attack.json` pins seed, level, monster def, first damage tick,
  damage, and resulting HP for Go and Godot golden checks.
- Bot scenario `14_dungeon_monsters.json` proves descend, passive damage, dungeon mob kill,
  `/state`, reconnect resume, and replay.

**Explicit non-goals:** no monster loot drops beyond `no_drop`, no monster attack animation,
no depth scaling, no ranged/AoE monsters, no protocol-level town safe-zone guard, no
character-scoped persistence, and no production monster art.
