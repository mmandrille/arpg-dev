# v35 — Boss floor gate

**Proves:** The generated dungeon can introduce a compact skill-gated boss floor with locked exits,
telegraphed damage, boss presentation metadata, and deterministic replay coverage.

- Dungeon level `-5` is generated as a compact `30 x 30` boss floor with fixed up/down stairs,
  one disabled teleporter, one pre-boss chest, one humanoid boss, and reduced trash population.
- Boss floor down stairs and teleporter start locked/disabled with reason `boss_alive`; both
  transition to `ready` and emit state-change events when the boss dies.
- Shared boss template and pattern rules define the first humanoid `cave_warden` boss, visual
  model/tint/scale metadata, and a telegraphed charged melee pattern with active-only damage.
- Go sim owns boss phase timing, hit predicates, locked-exit rejection, unlock, level transition,
  boss loot/chest hooks, and deterministic replay reconstruction.
- Protocol schemas now carry optional boss/visual metadata, boss phase events, lock/unlock events,
  and disabled/locked interactable state without requiring a client intent shape change.
- Godot renders boss entities through the humanoid model path, applies authoritative visual scale
  and tint, shows telegraph tinting, and presents locked/ready exit state from server data.
- Protocol bot scenario `24_boss_floor_gate.json` proves descent to `-5`, compact floor metadata,
  locked exit rejects, boss phase observation, boss kill, exit unlock, descent to `-6`, `/state`,
  reconnect, and replay verification.

**Explicit non-goals:** no additional boss templates, enrage phases, summoned adds, production
boss art/VFX/audio, boss health bar UI, co-op boss scaling, durable boss/map snapshots, quest
integration, or gameplay collision/reach scaling from visual scale.
