# `base_humanoid` — base character source notes

- **Runtime artifact:** `client/assets/characters/base_humanoid/base_humanoid.glb` (committed).
- **Manifest entry:** `character_base_humanoid_v0` in `assets/manifests/assets.v0.json`.
- **Required sockets:** `right_hand_socket` on `hand_r` and `off_hand_socket` on `hand_l`
  (v2/v99 canonical weapon mounts — ADR-0006 D4).

Source/provenance (origin, license, `sha256`) is recorded in the manifest `provenance` block and in
the spec's as-built notes (`docs/specs/v2_spec-equip-and-see-it.md`). No `.blend` source is committed
for v2 (ADR-0006 D2).
