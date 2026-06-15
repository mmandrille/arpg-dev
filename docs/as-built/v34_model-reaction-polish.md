# v34 — Model reaction polish

**Proves:** Damage/death presentation can be improved for all character-like visible entities while
staying client-only and driven by existing authoritative combat events.

- Godot now attaches a `ModelReactionController` to the local player, remote co-op players, and
  monsters, layering transform/material reactions over the existing `AnimationController`.
- Hit reactions lean away from the resolved attacker when possible, briefly dark-blink the model,
  then restore the entity's base tint.
- Death reactions supersede active hit/locomotion presentation, rotate the model down, and leave a
  persistent darker corpse presentation.
- Snapshot/render paths apply terminal death presentation from `hp <= 0`, so already-dead monsters
  and players do not need the original kill event to look dead.
- Remote co-op players now instantiate the same humanoid character scene as the local player and
  use a readable dark charcoal tint while remaining server-authoritative and unpredicted.
- Client bot debug state exposes local/entity presentation metadata for headless assertions without
  pixel matching.
- Client scenario `12_model_reaction_polish.json` proves monster hit, local-player hit from
  retaliation, and monster terminal death presentation through the real Godot client.

production customization/cosmetics, no monster art replacement, no corpse collision/despawn, and no
respawn/revive behavior.
