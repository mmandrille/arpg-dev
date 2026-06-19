# v273 Spec - Crocodile Archer Model

Status: Complete
Date: 2026-06-18
Codename: crocodile-archer-model

## Purpose

Use the supplied `assets/monsters/archer/crocodile_archer.glb` as the regular ranged enemy
presentation for `dungeon_archer`, without changing server-owned ranged monster gameplay.

## Asset Decision

- Adopt: the user-provided local GLB at `assets/monsters/archer/crocodile_archer.glb`.
- Borrow: the existing ADR-0006 manifest/runtime-copy pattern and node-root animation approach used
  by other user-provided static monster GLBs.
- Reject: external asset/plugin lookup, a new monster gameplay definition, and any client-owned
  navigation/combat behavior.

Probe result: the GLB imports cleanly, has no skin and no embedded animations, and reports bounds of
about `1.78 x 1.80 x 1.40` meters. It should use authored presentation clips on the scene node.

## Goals

- Register a manifest asset for the crocodile archer runtime GLB and extracted texture.
- Add a `monster_crocodile_archer` Godot scene with idle/walk/hit/death clips compatible with the
  existing monster animation controller.
- Point `dungeon_archer` in `shared/assets/monster_visuals.v0.json` at the new scene/asset.
- Preserve the existing ranged archer bot contract, including the bow-marker presentation assertion.

## Non-Goals

- No new monster stats, loot, AI, projectile rules, or pathing behavior.
- No client-authoritative monster movement/combat.
- No rigging/Blender export work or production attack animation pass.

## Acceptance

- Shared and asset validation resolve `dungeon_archer -> monster_crocodile_archer_v0`.
- Godot scene tests instantiate the new archer scene and verify required clips.
- The client ranged monster bot proves the regular ranged enemy appears through the live client flow.
- The protocol ranged monster bot still proves server-owned ranged combat behavior.
