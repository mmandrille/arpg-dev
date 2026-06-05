# ADR-0006: Asset Pipeline (glTF/GLB → manifest → mount socket)

- **Status:** Accepted (minimal v2 contract; expand with as-built notes as the pipeline grows)
- **Date:** 2026-06-05
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** assets, glTF, GLB, Godot, asset-manifest, attachment-sockets
- **Supersedes/extends:** ADR-0001 D1 (3D low-poly), D7 (glTF-first asset format)
- **Proven by:** `docs/specs/spec-equip-and-see-it.md` (slice v2)

---

## Context

ADR-0001 chose 3D low-poly under an isometric camera (D1) with a **glTF-first** asset format
(D7, marked Tier A / "to be ratified by ADR-0006"). The first playable slice (v0) proved the
authoritative gameplay loop with placeholder primitives only — no real models, no asset pipeline.

Slice v2 ("equip and see it") is the proof obligation for D7: an item equipped through the
authoritative server must render on the player's character via a **repeatable** path from shared
metadata to an imported model to a mounted runtime node. This ADR pins the **durable contract** of
that path. It deliberately does **not** pin the asset-*authorship* tool (Blender, CC0 download,
generator script), which is swappable per slice — the contract is identical regardless of source.

---

## Decisions

### D1 — Runtime format is glTF-binary (`.glb`)

All runtime 3D assets are single-file `.glb` (glTF 2.0 binary): geometry, materials, and (later)
animation in one file with **embedded** textures/materials. No external network fetch is permitted
at build or runtime — the bytes are committed to the repo.

**Rationale:** glTF is text-adjacent, well-documented, AI-generatable, and Godot-native. Single-file
`.glb` keeps the manifest→import→mount path trivial and checkout self-contained.

**Rejected:** glTF-separate (`.gltf` + `.bin` + loose textures) for runtime — more moving parts,
more ways to ship a broken asset; FBX/OBJ — weaker glTF-era tooling and material fidelity.

### D2 — Sources are not committed for v2; runtime `.glb` is

For v2 only the runtime `.glb` is committed under `client/assets/...`. The manifest's `source_path`
documents the *intended* source layout under `assets/...` (e.g. `.blend`), but those source files
need not exist yet. Externally-sourced (e.g. CC0) or generated runtime assets **must** record
provenance — `source_url`/origin, `license`, and `sha256` — in the manifest entry and the spec's
as-built notes.

**Rationale:** keeps checkout small and agent-friendly; matches D7's glTF-first intent. Provenance
+ `sha256` give license traceability and tamper-evidence without committing heavyweight sources.

### D3 — Asset manifest is the single source of truth for asset identity → file

`assets/manifests/assets.v0.json` maps a stable `asset_id` to its `type`
(`character` | `equipment`), `runtime_path` (repo-root-relative, must exist), `format` (`glb`),
declared `required_nodes` (sockets a character must expose), and optional `provenance` + `slot`.
`asset_id` keys are stable and never reused for a different visual.

Gameplay/visual separation is strict: **shared visual metadata**
(`shared/assets/item_visuals.v0.json`) maps an `item_def_id` → `asset_id` + `mount_socket` +
`local_transform`; it carries **no gameplay stats**. The asset manifest maps `asset_id` → file.
The server never sees or decides asset paths (ADR-0001 D2 authority boundary is preserved).

### D4 — Attachment via a named `Node3D` placeholder socket (v2)

The character scene exposes a named mount node (v2 canonical: `right_hand_socket`) as a plain
`Node3D` under the imported model root. Equipment is instanced as a child of that socket; the child
node's name equals its `asset_id` so scene paths are stable across runs. A per-item `local_transform`
(position / rotation_degrees / scale) is applied after mounting.

**Rationale:** v2 has no animation; bone rigging would add scope without proving the
metadata→mount pipeline, which is the actual deliverable. The node **name** and transform semantics
are the contract, not the node class.

**Upgrade path:** when skeletal animation lands, `right_hand_socket` becomes a `BoneAttachment3D`
bound to a hand bone. Consumers key off the socket **name**, so the swap is local to
`character.tscn` and does not change the manifest, shared metadata, or resolver contract.

### D5 — Validation is split by layer (fast Python + authoritative Godot)

- **Python (`make validate-assets`, fast, no engine):** schema-validate the manifest; verify every
  `runtime_path` exists; verify every `item_visuals.asset_id` resolves in the manifest; verify
  character `required_nodes` cover every referenced `mount_socket`; verify `provenance.sha256`
  matches the committed file; best-effort parse of the GLB JSON chunk to confirm declared node
  names exist (warn-only if the parser can't read it; hard-fail on an explicit name mismatch).
- **Godot headless (`make client-smoke`, authoritative for runtime truth):** the engine that
  actually imports and mounts the asset is the source of truth for socket presence and mounted
  visibility. The smoke asserts that authoritative equipped state has a matching **visible** node.
- **Shared (`make validate-shared`):** schema-validates `shared/assets/*` and cross-checks the
  visual metadata against `shared/rules/items.v0.json` and the golden resolution fixture.

### D6 — Path transform: manifest `runtime_path` → Godot `res://`

The Godot project root **is** `client/`. The manifest `runtime_path` is repo-root-relative
(`client/assets/...`, what the Python validator checks on disk). The resolver maps it to a Godot
resource path by **stripping the leading `client/` and prepending `res://`**. The resolver owns this
single transform; nothing else hardcodes it.

---

## Consequences

- A new equippable visual is a data-only change in the common case: add an `item_visuals` entry, an
  `assets` manifest entry, commit the `.glb`. No engine or server code changes.
- The resolver takes its mount-root node as an **injected dependency** (never an absolute
  `/root/Main/...` lookup) so the interactive scene and the headless smoke share one code path.
- Authorship tooling is free to change (CC0 download today, Blender/AI generation later) without
  touching the contract, because the contract is the manifest→import→mount path, not the source.

## Status of D7 (ADR-0001)

This ADR ratifies ADR-0001 D7 (glTF-first) for runtime assets. Animation rigging, a Blender export
pipeline, texture-budget enforcement, and a remote asset patcher remain out of scope here and will
extend this ADR in later slices.
