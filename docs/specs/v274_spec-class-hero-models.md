# v274 Spec: Class Hero Models

## Summary

Replace the generated placeholder visuals for the barbarian, paladin, rogue, and sorcerer hero classes with the supplied GLB assets while keeping class gameplay, networking, and server authority unchanged.

## Goals

- Barbarian uses `assets/characters/barbarian/goliath_barbarian.glb`.
- Paladin uses `assets/characters/paladin/knight.glb`.
- Rogue uses `assets/characters/rogue/assasine.glb`.
- Sorcerer uses `assets/characters/sorcerer/mage.glb`.
- The existing `character_<class>_v0` asset IDs remain stable for class selection, saved snapshots, and bot debug output.
- Static class models still expose right-hand, off-hand, and fallback equipment sockets so item presentation does not disappear.
- Paladin scale correction is data-driven through class presentation metadata, not hardcoded in gameplay code.

## Non-Goals

- Do not change class stats, combat, skills, loot, collision, or server authority.
- Do not add client-authoritative player movement or combat truth.
- Do not require the supplied static GLBs to satisfy the generated humanoid skeleton contract.
- Do not replace the ranger model in this slice.

## Asset Decision

- Adopt: the four user-provided local GLBs as class body visuals.
- Borrow: the existing asset manifest, class presentation loader, `ModelRoot` replacement path, and character socket conventions.
- Reject: a new external model pipeline, retargeting pipeline, or fake skeleton metadata for static meshes. These models have no skins or embedded animations, so the slice makes static mesh support explicit and keeps generated/root animation behavior as the fallback.

## Acceptance

- `make validate-assets` accepts explicitly static character entries with empty `required_nodes` while continuing to require hand bones for rigged character entries.
- `godot --headless --path client --script res://tests/test_animation.gd` loads all four class models, verifies a visible mesh, applies class scale metadata, and confirms equipment sockets exist after class model replacement.
- A client visual scenario remains available for manual inspection: `make bot-visual scenario=20_menu_create_join_flow`.
