# v37 — Combat control and boss AI fixes

**Proves:** Force-stand directional attacks, aggro-on-hit, nearby group aggro, and boss chase/damage
behavior can be repaired while keeping combat authority in the Go sim.

- Shared protocol now accepts `directional_attack_intent` with a direction-only payload, validates
  an example, and documents zero-vector `move_intent` as the authoritative stop/cancel path.
- Server movement cancel semantics clear active movement and auto-approach before the player advances,
  and directional attacks also cancel movement before resolving melee or ranged combat.
- Directional melee selects targets through a deterministic server-owned forward capsule; directional
  ranged shots reuse authoritative projectile spawning and swept collision while omitting `target_id`
  for free shots.
- Player-to-monster damage now centralizes damage events, kill/drop/XP follow-up, and aggro-on-hit.
  A damaged live chase-capable monster records the attacking player as its preferred target, and aggro
  also wakes other chase-capable monsters whose own aggro radius contains the attacking player.
  Propagation then chains through nearby live chase-capable monsters on the same level so close packs
  wake as a group while monsters outside attacker and group radius stay idle.
- Bosses now participate in hostile movement during idle, cooldown, telegraph, and recovery, pause
  during active damage ticks, prefer the player that aggroed them, and damage a stationary failed-dodge
  target during active phases.
- Locked boss-floor exits remain server-owned; disabled boss teleporters now reject immediately with
  `boss_alive` instead of starting an unusable auto-approach.
- Godot `SHIFT` force-stand sends the stop intent, suppresses movement while held, and `SHIFT+LMB`
  sends repeated direction-only attacks using existing facing/attack presentation.
- Protocol bot scenarios prove directional ranged aggro/movement and the repaired boss floor path.
  Godot helper/unit tests cover force-stand and directional hold behavior; the full headless
  modifier/mouse client bot scenario remains deferred because the current bot controller fallback is
  not reliable for that proof.

**Explicit non-goals:** no final attack-speed/cooldown gameplay, skill bar, mana system, active
ability catalog, homing/target-prediction projectiles, client hit detection, PvP/friendly fire, new
boss templates, boss enrage/adds, production boss/combat VFX/audio/art, co-op boss scaling, broad
monster AI rewrite, Protobuf migration, or a reliable full-scene headless `SHIFT+LMB` client bot proof.
