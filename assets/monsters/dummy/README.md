# `monster_dummy` — training dummy source notes

- **Runtime artifact:** `client/assets/monsters/dummy/monster_dummy.glb` (committed).
- **Manifest entry:** `monster_dummy_v0` in `assets/manifests/assets.v0.json`.
- **Skeleton bones:** `root` (base slab) and `pivot` (post) — the post is weighted to
  `pivot` so hit/death clips rotate it about the base.

Authored by `tools/assets/gen_glb.py` as a SKINNED rig (v3); runtime bytes are
deterministic. Source/provenance (origin, license, `sha256`) is recorded in the
manifest `provenance` block and in the spec's as-built notes. No `.blend` source is
committed (ADR-0006 D2).
