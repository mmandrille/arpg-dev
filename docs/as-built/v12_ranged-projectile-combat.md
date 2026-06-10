# v12 — Ranged projectile combat

**Proves:** Ranged weapons can use server-owned traveling projectile entities with deterministic
impact-time collision, hit, damage, replay, and client presentation.

- `training_bow` declares `attack_mode: "ranged"`, weapon damage, reach, and projectile speed in
  shared item rules, with schema and validation guards.
- Ranged monster `action_intent` spawns a wire-visible `projectile` entity; melee combat, loot, and
  interactables keep their existing behavior.
- Projectile flight advances at 20 Hz and sweeps against inflated wall/door AABBs and live monster
  circles using nearest-hit selection with deterministic tie-breaks.
- Ranged hit chance and damage roll only at impact; miss emits `attack_missed` without retaliation.
- `ranged_projectile.json` pins gap kill, wall block, and miss/no-retaliation cases for Go and
  GDScript fixture checks.
- `ranged_lab` plus bot scenario `06_ranged_lab.json` proves bow pickup/equip, ranged kill beyond
  melee range through a wall gap, `/state`, reconnect resume, and replay.
- Godot renders placeholder projectile entities from authoritative spawn/update/remove deltas.
- `make ci` green on 2026-06-05.

**Explicit non-goals for v12:** no spells, piercing, homing/AoE, monster ranged AI, predictive
leading, ranged pickup/door activation, production bow art, inventory UI, or projectile catalog.
