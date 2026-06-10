# v26 — Character stats and leveling

**Proves:** Character-owned XP, levels, stat points, and derived substats can be durable,
server-authoritative progression while the Godot client remains a renderer/input surface.

- Shared `character_progression.v0.json` defines base stats, a table XP curve, points per level,
  and bounded derived-stat formulas for damage, armor, attack speed, hit chance, crit chance,
  crit damage, movement speed, HP, and mana.
- Dungeon mobs now award positive XP; monster kill XP applies exactly once, crosses level
  thresholds in order, and grants 5 unspent stat points per level.
- `character_progression` persists per character, and session-start progression snapshots preserve
  deterministic reconnect/replay boundaries.
- `allocate_stat_intent` is server-authoritative; invalid stats, dead-player allocation, and
  overspending reject without mutating state.
- `vit` allocation updates derived `max_hp` and raises current HP by the gained max; the first
  `str` damage hook adds derived damage to melee/fixed weapon damage.
- Armor, crit, hit chance, attack speed, movement speed, max mana, and magic damage are computed
  and displayed but remain gameplay-deferred.
- Godot adds a left-side `C`-toggle character sheet with stat `+` buttons and a compact XP bar
  below the hotbar; client state only updates after authoritative snapshots/deltas.
- Protocol bot scenario `18_character_stats_and_leveling.json` proves XP, level-up points, VIT
  allocation, overspend rejection, `/state`, replay, reconnect, and fresh-session persistence.
- Client bot scenario `09_character_stats_panel.json` proves the stats panel, XP bar, pause/menu
  allocation lock, VIT spend through the `+` button, and max HP UI update.

**Explicit non-goals:** no passive skill tree, no respec, no class selection, no stat requirements,
no mana consumers, no armor/crit/hit/attack-speed gameplay, and no main-menu character summaries.
