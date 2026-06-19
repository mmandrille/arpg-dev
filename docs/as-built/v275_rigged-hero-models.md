# v275 As-Built - Rigged Hero Models

Date: 2026-06-19
Spec: [`docs/specs/v275_spec-rigged-hero-models.md`](../specs/v275_spec-rigged-hero-models.md)
Plan: [`docs/plans/v275_2026-06-19-rigged-hero-models.md`](../plans/v275_2026-06-19-rigged-hero-models.md)

## Shipped

- Added `tools/assets/rig_hero_glbs.py`, a deterministic GLB post-process that preserves the
  supplied hero meshes/materials/textures and appends the shared humanoid skin contract.
- `make gen-assets` now runs the generated placeholder pass and then re-rigs the four supplied
  hero GLBs, so class runtime assets remain reproducible from committed sources.
- Barbarian, paladin, rogue, and sorcerer runtime GLBs now contain skin joints named `root`,
  `spine`, `arm_l`, `hand_l`, `arm_r`, `hand_r`, `leg_l`, and `leg_r`.
- The asset manifest requires those joints again for all four class hero assets and records new
  rigged runtime hashes.
- The Godot animation smoke now rejects static class regressions: each class model must expose a
  `Skeleton3D`, hand sockets must be `BoneAttachment3D`s bound to the correct hand bones, and the
  existing walk/attack/off-hand attack clips must rotate the expected bones.

## Proof

```bash
.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py -q
make gen-assets
.venv/bin/python -m pytest tools/assets/test_rig_hero_glbs.py tools/assets/test_validate_assets.py -q
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
make validate-shared
make client-unit
make ci
```

All commands above passed on 2026-06-19.

## Boundaries

- No server gameplay, class stats, combat, loot, protocol, persistence, or authority changed.
- The automatic weights are rough rigid regions, suitable for proving the old animation clips drive
  the new hero meshes. Final production-quality deformation remains future art/tech-art polish.
- Ranger remains on the generated rigged humanoid asset.
- Paladin keeps the v274 class presentation scale correction because its source GLB is authored at
  a much smaller coordinate scale.
