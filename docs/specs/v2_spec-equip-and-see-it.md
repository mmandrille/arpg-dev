# Spec: `equip-and-see-it`

Status: Ready for review (2026-06-05)
Branch: `feature/equip-and-see-it`
Related: `docs/adr/0001-technology-stack.md`; `docs/adr/0006-asset-pipeline.md` (Task 1);
`docs/specs/v1_spec-first-playable-vertical-slice.md` (v0 baseline)

## 1. Purpose

Prove slice v2 from ADR-0001: an equipped item is not only backend-owned state, it is visibly
attached to the player's character in the Godot client.

The slice keeps the v0 gameplay loop intact: login, create/resume solo session, kill the training
dummy, pick up the `rusty_sword`, equip it through the authoritative server, and render that item
on the character under the isometric camera. The new proof is the asset path from shared metadata
to imported glTF assets to a mounted runtime visual node.

This slice should establish the smallest durable asset-pipeline contract agents can repeat:

- a base humanoid character model,
- one modular weapon model,
- one named attachment socket,
- one shared visual metadata file mapping `item_def_id` to an asset,
- one automated validation path,
- one Godot smoke assertion that equipped backend state has a matching visible node.

## 2. Non-goals

- Final art quality, animation polish, hit reactions, VFX, UI polish, or sound.
- A complete Blender cleanup/export pipeline.
- Procedural asset generation at runtime.
- Multiple armor slots, layered clothing, dyes, affixes, transmogs, rarity visuals, or cosmetics.
- Server-authoritative visual paths. The server owns item state; the client owns rendering.
- Multiplayer replication of other players' equipment.
- New combat mechanics, attack range, skill systems, or loot table complexity.
- Replacing JSON WebSocket messages or changing the v0 realtime protocol envelope.
- A production asset CDN or remote asset patcher.

## 3. Files to create or modify

```text
assets/                         - Source/export roots and provenance for sourced 3D assets
assets/characters/              - Base character source and exported glTF/.glb files
assets/equipment/weapons/        - Weapon source and exported glTF/.glb files
assets/manifests/                - Asset manifest JSON consumed by validation and Godot import
client/assets/                   - Godot-imported runtime assets or import targets
client/scenes/                   - Character scene with named attachment sockets
client/scripts/                  - Equipment visual resolver and mount/update logic
client/tests/                    - GDScript tests for item visual metadata and mounted equipment
shared/assets/                   - Shared visual metadata schemas/data
shared/golden/                   - Cross-language fixture for item visual resolution
shared/rules/                    - Item data extension or companion reference to visuals
tools/assets/                    - Python validation tooling for manifests, paths, and GLB metadata
scripts/client_smoke.sh          - Extend smoke to assert equipped visual state
scripts/ci.sh                    - Include asset validation and v2 client smoke where available
docs/adr/0006-asset-pipeline.md  - Asset-pipeline ADR, if written before implementation
docs/plans/                      - Implementation plan for this spec
```

## 4. Data shapes

### 4.1 Item visual metadata

Use a companion shared asset metadata file instead of changing the v0 WebSocket payload. The server
continues to emit authoritative item state by `item_def_id`; the client resolves the visual through
shared data.

Minimum file set:

```text
shared/assets/item_visuals.v0.schema.json
shared/assets/item_visuals.v0.json
```

Example:

```json
{
  "version": 0,
  "item_visuals": {
    "rusty_sword": {
      "asset_id": "weapon_rusty_sword_v0",
      "slot": "weapon",
      "mount_socket": "right_hand_socket",
      "local_transform": {
        "position": { "x": 0.0, "y": 0.0, "z": 0.0 },
        "rotation_degrees": { "x": 0.0, "y": 0.0, "z": 0.0 },
        "scale": { "x": 1.0, "y": 1.0, "z": 1.0 }
      }
    }
  }
}
```

Required properties:

| Field | Purpose |
|-------|---------|
| `version` | Schema/data version. |
| `item_visuals` | Map keyed by canonical `item_def_id` from `shared/rules/items.v0.json`. |
| `asset_id` | Stable ID into the asset manifest. |
| `slot` | Equipment slot; v2 only supports `weapon`. |
| `mount_socket` | Named socket node on the character scene. |
| `local_transform` | Per-item transform applied after mounting to the socket. |

The metadata must not contain gameplay stats. It is a rendering contract only.

### 4.2 Asset manifest

Minimum file set:

```text
assets/manifests/assets.v0.schema.json
assets/manifests/assets.v0.json
```

Example:

```json
{
  "version": 0,
  "assets": {
    "character_base_humanoid_v0": {
      "type": "character",
      "source_path": "assets/characters/base_humanoid/base_humanoid.blend",
      "runtime_path": "client/assets/characters/base_humanoid/base_humanoid.glb",
      "format": "glb",
      "scale_unit": "meters",
      "required_nodes": ["right_hand_socket"]
    },
    "weapon_rusty_sword_v0": {
      "type": "equipment",
      "slot": "weapon",
      "source_path": "assets/equipment/weapons/rusty_sword/rusty_sword.blend",
      "runtime_path": "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb",
      "format": "glb",
      "scale_unit": "meters",
      "required_nodes": []
    }
  }
}
```

Required manifest rules:

- `asset_id` keys are stable and never reused for a different visual.
- `runtime_path` is relative to repo root and must exist.
- `source_path` is relative to repo root; if source files are not committed (e.g. a CC0 download),
  the entry's `provenance` (and the plan's as-built notes) must record where the runtime asset came
  from and under what license.
- `format` is `glb` for v2 runtime assets.
- `type` is one of `character` or `equipment`.
- Entries **may** carry an optional `provenance` object (`source_url`, `license`, `sha256`) for
  externally-sourced (e.g. CC0) assets; the manifest schema must allow it. The asset validator
  verifies `sha256` against the committed runtime file when `provenance.sha256` is present.
- Character assets must declare every socket required by `shared/assets/item_visuals.v0.json`.
- Equipment assets must declare the slot they can satisfy.

### 4.3 Attachment sockets

The Godot character scene must expose a named node:

```text
CharacterRoot
+-- SkeletonOrVisualRoot
    +-- right_hand_socket
```

`right_hand_socket` is the v2 canonical weapon mount. The implementation may use a `BoneAttachment3D`
or a plain `Node3D` placeholder, but the node name and transform semantics are part of the contract.

If the imported base character does not have a usable skeleton yet, v2 may attach the sword to a
stable placeholder socket. That fallback must still be a real 3D node under the character and must
be visible under the isometric camera.

### 4.4 Runtime client equipment state

The Godot client should expose equipment visual state through its existing debug surface so tests do
not rely on pixel scraping.

Minimum debug shape:

```json
{
  "equipped_visuals": {
    "weapon": {
      "item_instance_id": "1004",
      "item_def_id": "rusty_sword",
      "asset_id": "weapon_rusty_sword_v0",
      "mount_socket": "right_hand_socket",
      "node_path": "/root/Main/World/PlayerAnchor/CharacterVisual/ModelRoot/right_hand_socket/weapon_rusty_sword_v0",
      "visible": true
    }
  }
}
```

If no item is equipped:

```json
{
  "equipped_visuals": {
    "weapon": null
  }
}
```

Debug state must be machine-readable and stable enough for the headless Godot smoke test.

`node_path` is **environment-dependent**: the interactive client mounts under the full
`main.tscn` tree (`/root/Main/World/PlayerAnchor/...`), while the headless smoke runs a minimal
scene subtree (`smoke.gd` extends `SceneTree` and does not instantiate `main.tscn`). Tests assert
the trailing node name (`asset_id`) and the §4.4 ids/flags, **not** the absolute path verbatim.

### 4.5 Protocol impact

No v2 WebSocket protocol bump is required by default.

The client already receives enough authoritative state:

- `inventory_add` / `inventory_update` include `item_instance_id`, `item_def_id`, `slot`, and
  `equipped`.
- `equipped_update` includes `slot` and `item_instance_id`.
- `session_snapshot.equipped.weapon` carries the currently equipped item instance.

The client resolves the `item_instance_id` back to `item_def_id` from its authoritative inventory
cache, then resolves `item_def_id` to visual metadata from `shared/assets/item_visuals.v0.json`.

If implementation finds this insufficient, the plan must explicitly justify a protocol schema
change before modifying the wire format.

Confirmed sufficient for v2: `session_snapshot.equipped.weapon` carries the **item instance id**
(decimal string), not `item_def_id`. The client resolves instance → def via its authoritative
inventory cache (same data as `/v0/sessions/{id}/state`), then def → visual via shared asset metadata.

Resume is **already server-supported with no protocol change**: `POST /v0/sessions` accepts an
optional `resume_session_id`, and the WS handler rehydrates persisted inventory **and equipped
state** (`ListInventory` → `Sim.LoadInventory`) before sending the initial `session_snapshot`. The
one client-side gap: `net_client.gd::create_session()` currently hardcodes `{"mode":"solo"}` and
must gain an optional `resume_session_id` argument so the smoke (acceptance #8) can rejoin the same
session instead of minting a fresh one.

### 4.6 Golden fixture (cross-language)

Add a golden fixture consumed by Python validation, GDScript tests, and documentation:

```text
shared/golden/item_visual_resolution.v0.schema.json
shared/golden/item_visual_resolution.json
```

Example:

```json
{
  "description": "rusty_sword resolves to the v2 weapon asset and mount socket via shared metadata + manifest.",
  "item_def_id": "rusty_sword",
  "expected_asset_id": "weapon_rusty_sword_v0",
  "expected_mount_socket": "right_hand_socket",
  "expected_slot": "weapon"
}
```

### 4.7 Client debug export (machine-readable)

v0 exposes human-readable on-screen debug only. v2 adds a structured export for automation without
pixel scraping.

Minimum contract:

- `EquipmentVisualResolver.get_debug_state() -> Dictionary` returns the shape from §4.4 plus an
  optional `warnings` array.
- Headless smoke calls this in-process after equip (same pattern as v0 `smoke.gd` driving the
  protocol directly, not via HTTP).
- Warning entries use stable codes:

```json
{ "code": "unknown_item_def_id", "item_def_id": "missing_item" }
{ "code": "unknown_asset_id", "asset_id": "missing_asset" }
{ "code": "missing_mount_socket", "mount_socket": "right_hand_socket" }
```

`node_path` in `equipped_visuals.weapon` is the Godot scene-tree path to the mounted asset child.
The mounted child node name must equal `asset_id` (e.g. `weapon_rusty_sword_v0`) so paths are
stable across runs.

### 4.8 Shared and manifest data loading

Follow the v0 cross-language pattern established in `client/tests/test_golden.gd`:

- Godot resolves repo-root shared JSON via
  `ProjectSettings.globalize_path("res://").path_join("../shared/...")`.
- Godot loads runtime `.glb` files from `client/assets/...` as `res://assets/...` after import.
  The Godot project root **is** `client/`, so the manifest `runtime_path` (`client/assets/...`,
  repo-root-relative, what the Python validator checks on disk) maps to the Godot resource path by
  stripping the leading `client/` and prepending `res://`. The resolver owns this transform.
- Python validators resolve paths relative to the repository root.
- Do not duplicate `shared/assets/*.json` inside `client/`; read from `shared/` at runtime and in
  CI validation.
- The resolver must **not** hardcode an absolute scene path (e.g. `/root/Main/...`). It takes the
  mount-root node (the character's socket parent) as an injected dependency so `main.gd` (full
  `main.tscn`) and `smoke.gd` (minimal `SceneTree` subtree) drive one shared code path.

### 4.9 Cross-consistency rules (validation)

Asset and shared validation must enforce:

1. Every key in `item_visuals` exists in `shared/rules/items.v0.json` with a matching `slot`.
2. Every `asset_id` referenced in `item_visuals` exists in `assets/manifests/assets.v0.json`.
3. Every character manifest entry lists all `mount_socket` values referenced by item visuals.
4. Equipment manifest entries declare a `slot` matching the item visuals that reference them.
5. For v2 scope, only `rusty_sword` requires a visual entry; other items may remain unmapped until
   a later slice.

`make validate-shared` covers `shared/assets/` schemas, instances, and golden cross-checks.
`make validate-assets` covers manifests, runtime path existence, and GLB node-name checks where
inspectable.

## 5. Architecture and flow

```text
Shared data
  -> item_visuals.v0.json maps rusty_sword -> weapon_rusty_sword_v0 + right_hand_socket
  -> assets.v0.json maps weapon_rusty_sword_v0 -> client runtime .glb path

Validation
  -> schema-validate shared asset metadata
  -> schema-validate asset manifest
  -> verify referenced runtime paths exist
  -> verify character assets expose required sockets where tooling can inspect them

Godot client
  -> loads base character scene with right_hand_socket
  -> runs existing v0 auth/session/WebSocket flow
  -> receives authoritative inventory/equipped state
  -> resolves item_def_id through shared asset metadata
  -> instantiates weapon GLB under right_hand_socket
  -> updates debug state: equipped_visuals.weapon.visible == true

Server
  -> unchanged authority boundary
  -> continues to own pickup/equip/inventory state
  -> does not decide asset paths or render behavior

Python bot
  -> continues to verify authoritative backend state
  -> may remain blind to visuals

Godot smoke
  -> verifies both authoritative equip state and client-visible mounted equipment state
```

### 5.1 Client scene refactor (from v0 placeholders)

v0 builds the player as a runtime `CapsuleMesh` in `main.gd`. v2 introduces a durable scene graph:

```text
Main (Node3D)                         # res://scenes/main.tscn
+-- World (Node3D)
|   +-- PlayerAnchor (Node3D)         # follows authoritative player position
|       +-- CharacterVisual (instance of res://scenes/character.tscn)
|           +-- ModelRoot             # imported base_humanoid.glb scene root
|               +-- right_hand_socket # Node3D placeholder in v2 (see §9 decisions)
+-- Entities (Node3D)                 # monsters, loot — may remain v0 primitives
+-- Camera / Light / UI               # unchanged isometric setup
```

Monsters and loot may stay placeholder primitives for v2. Only the **local player** adopts the base
humanoid GLB plus mounted weapon. Other entities are out of scope (non-goal: multiplayer equipment
replication).

Equipment mounting is owned by `EquipmentVisualResolver` (`client/scripts/equipment_visuals.gd`),
invoked from `main.gd` on `session_snapshot`, `inventory_*`, and `equipped_update` changes.

## 6. Asset constraints for v2

The goal is a repeatable proof, not final art.

Minimum visual constraints:

- Low-poly, readable silhouette under the existing isometric/orthographic camera.
- One base humanoid character and one rusty sword or equivalent one-handed weapon.
- Runtime format is `.glb`.
- Assets use a consistent scale where the weapon appears handheld, not oversized or microscopic.
- Materials must render without external network dependencies.
- Committed assets must be **CC0** (or otherwise license-clean for redistribution); provenance
  (source URL + license + `sha256`) is recorded per asset in the manifest entry / as-built notes.
- The character and weapon must be visible on a blank/local development machine after checkout.
- The mounted item must remain visible while the player is idle and after movement reconciliation.

Budget targets:

| Asset | Target |
|-------|--------|
| Base humanoid runtime `.glb` | Under 5 MB. |
| Rusty sword runtime `.glb` | Under 2 MB. |
| Texture files | PNG; under 2048 px per side unless justified. |

These are soft budgets for v2. Exceeding them requires an as-built note.

## 7. Failure behavior

- Missing or invalid shared visual metadata fails `make validate-shared` or the new asset
  validation command.
- Missing runtime asset files fail the asset validation command.
- Missing required sockets fail asset validation when inspectable; if static inspection cannot
  inspect the socket, the Godot test must fail if the socket is absent at runtime.
- Unknown `item_def_id` at runtime renders no equipment, emits a structured client debug warning,
  and does not crash the client.
- Unknown `asset_id` at runtime renders no equipment, emits a structured client debug warning, and
  does not crash the client.
- Equipping a different weapon later must remove or replace the previously mounted weapon node.
  v2 only needs one weapon item, but the mount logic must not duplicate stale child nodes.

## 8. Acceptance criteria

1. `shared/assets/item_visuals.v0.schema.json` and `shared/assets/item_visuals.v0.json` exist and
   map `rusty_sword` to one weapon visual.
2. `assets/manifests/assets.v0.schema.json` and `assets/manifests/assets.v0.json` exist and map the
   base character plus one weapon to committed runtime `.glb` paths.
3. Asset validation confirms every referenced runtime path exists and every required item visual
   has a manifest entry.
4. The Godot character scene exposes `right_hand_socket`.
5. The existing v0 server flow remains authoritative for pickup/equip; the server does not expose
   client asset paths.
6. After the client receives an authoritative `equipped_update` for `rusty_sword`, Godot mounts the
   weapon asset under `right_hand_socket`.
7. Godot debug state reports `equipped_visuals.weapon.visible == true`, with the expected
   `item_instance_id`, `item_def_id`, `asset_id`, and `mount_socket`.
8. Reconnecting/resuming the same session with the item already equipped restores the mounted
   visual from `session_snapshot`, not only from the live `equipped_update` delta.
9. The Python protocol bot still completes the v0 authoritative flow without any visual-specific
   assumptions.
10. `make ci` includes shared validation, Go tests, Python bot/replay, asset validation, and the
    Godot visual smoke where the Godot runtime is available.
11. Invalid or missing visual metadata fails validation before runtime.
12. Runtime visual-resolution failures produce structured client debug warnings and do not crash
    the server or client.
13. After equip, a move intent and reconciliation leave `equipped_visuals.weapon.visible == true`.
14. `shared/golden/item_visual_resolution.json` exists and is consumed by both Python validation
    and a GDScript unit test.

## 9. Decisions (resolved)

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | ADR-0006 timing | Write a **minimal ADR-0006** in implementation Task 1 encoding the v2 pipeline **contract** (format, manifest shape, sockets, import + validation rules); expand with as-built notes after assets land. | ADR-0001 assigns D7 Tier A to ADR-0006; this slice is the proof. The ADR pins the durable contract, **not** the asset-authorship tool (which is swappable per slice — see #5). |
| 2 | Commit `.blend` sources? | **Runtime `.glb` only** for v2. Manifest `source_path` documents intended source layout; provenance (source URL + license + `sha256`) recorded in plan as-built notes. | Keeps checkout small and agent-friendly; matches glTF-first ADR-0001 D7. |
| 3 | Bone attachment vs placeholder socket? | **`Node3D` placeholder** named `right_hand_socket` parented under the imported character root. | No animation in v2; bone rigging adds scope without proving the metadata→mount pipeline. Upgrade path documented in ADR-0006. |
| 4 | Python GLB inspection vs Godot? | **Both, split by layer**: Python checks manifest paths + optional GLB node-name inspection; Godot headless smoke is authoritative for runtime socket presence and mounted visibility. | Fast CI feedback from Python; runtime truth from the engine that actually mounts assets. |
| 5 | Asset authorship for v2 | **Fetch CC0 low-poly GLBs** (e.g. Kenney.nl, Quaternius) for the base humanoid and a one-handed weapon; commit them to `client/assets/...` and record provenance (URL + `CC0` + `sha256`) in the manifest entry / as-built notes. **Fallback:** a committed stdlib-only generator script if no suitable CC0 model is found. No external CDN; bytes are committed, never fetched at build/runtime. | Rehearses the real long-run pipeline (acquire → record license/provenance → import → manifest → mount) and gives a readable silhouette now. CC0 = no attribution/copyleft risk. Authorship is **not** the proof — the manifest→import→mount contract is, and that is identical regardless of source. Art quality remains a non-goal. |

## 10. Testing plan

1. Shared validation:

```bash
make validate-shared
```

2. Asset validation command, to be added by the implementation plan:

```bash
make validate-assets
```

3. Server regression tests:

```bash
make test-go
```

4. Python protocol bot and replay remain authoritative-state regressions:

```bash
make bot
make replay SESSION_ID=<recorded-session-id>
```

5. GDScript unit test for item visual resolution (extends v0 golden test pattern):

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
```

6. Godot headless smoke extends the v0 flow and asserts visible mounted equipment through debug
   state. Scenarios covered in one run:

   - Complete kill → pickup → equip → assert server `/state` **and** `get_debug_state()`.
   - Send at least one move intent after equip; assert visual still mounted.
   - Disconnect WebSocket, call `create_session` resume on the same session, assert visual restored
     from `session_snapshot` without a new `equipped_update`.

```bash
make client-smoke
```

7. Full local gate:

```bash
make ci
```
