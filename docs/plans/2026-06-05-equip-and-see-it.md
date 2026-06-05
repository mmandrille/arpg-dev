# Equip and See It (Slice v2) - Implementation Plan

Status: Complete (2026-06-05) — all tasks implemented; `make ci` green incl. extended Godot smoke.
Goal: prove slice v2 from ADR-0001 — an equipped item renders on the player character via a
repeatable shared-metadata → GLB → mount-socket pipeline, without changing server authority.
Architecture: v0 protocol and sim unchanged; new shared asset metadata + asset manifest + Godot
equipment visual resolver; Python asset validation; extended headless smoke.
Tech stack: Go (unchanged), Godot 4.6.3/GDScript, Python, glTF/GLB, JSON Schema.

Related spec: `docs/specs/spec-equip-and-see-it.md`
Baseline: v0 first playable slice (complete — `make ci` green)

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/adr/0006-asset-pipeline.md` | Minimal ADR encoding v2 pipeline contracts |
| Create | `shared/assets/item_visuals.v0.schema.json` | Item → visual metadata schema |
| Create | `shared/assets/item_visuals.v0.json` | `rusty_sword` visual mapping |
| Create | `shared/golden/item_visual_resolution.v0.schema.json` | Golden schema |
| Create | `shared/golden/item_visual_resolution.json` | Cross-language visual resolution fixture |
| Create | `assets/manifests/assets.v0.schema.json` | Asset manifest schema |
| Create | `assets/manifests/assets.v0.json` | Character + weapon manifest entries |
| Create | `assets/characters/base_humanoid/` | Source notes / optional generation script |
| Create | `assets/equipment/weapons/rusty_sword/` | Source notes / optional generation script |
| Create | `client/assets/characters/base_humanoid/base_humanoid.glb` | Runtime character model |
| Create | `client/assets/equipment/weapons/rusty_sword/rusty_sword.glb` | Runtime weapon model |
| Create | `client/scenes/character.tscn` | Character scene with `right_hand_socket` |
| Modify | `client/scenes/main.tscn` | Wire World / PlayerAnchor / CharacterVisual hierarchy |
| Create | `client/scripts/equipment_visuals.gd` | Resolver: metadata → instantiate → mount (mount-root injected) |
| Modify | `client/scripts/net_client.gd` | Add optional `resume_session_id` to `create_session()` |
| Modify | `client/scripts/main.gd` | Use character GLB + delegate equip visuals |
| Modify | `client/scripts/smoke.gd` | Assert debug visuals, move-after-equip, resume |
| Create | `client/tests/test_item_visuals.gd` | GDScript golden test for visual resolution |
| Create | `tools/assets/validate_assets.py` | Manifest + path + GLB node validation |
| Create | `tools/assets/test_validate_assets.py` | Pytest for asset validator |
| Modify | `tools/validate_shared.py` | Include `shared/assets/` + visual golden cross-checks |
| Modify | `Makefile` | Add `validate-assets` target |
| Modify | `scripts/ci.sh` | Run asset validation; renumber steps |
| Modify | `.github/workflows/ci.yml` | Run asset validation in CI |
| Modify | `scripts/client_smoke.sh` | Run `test_item_visuals.gd` before slice smoke |
| Modify | `README.md` | Document `make validate-assets` |

## Task 1: ADR-0006 skeleton and repo layout

Files:
- Create: `docs/adr/0006-asset-pipeline.md`
- Create: `assets/`, `assets/characters/`, `assets/equipment/weapons/`, `assets/manifests/`
- Create: `client/assets/`, `tools/assets/`

- [ ] Step 1.1: Write minimal ADR-0006 covering format (glTF/GLB), manifest contract, mount
  sockets, v2 placeholder-socket decision, and upgrade path to bone attachments.
- [ ] Step 1.2: Create directory layout from spec §3.
- [ ] Step 1.3: Add `.gitkeep` or README stubs under `assets/` source dirs documenting that v2
  commits runtime GLB only (per spec §9 decision #2).

Verification:

```bash
test -f docs/adr/0006-asset-pipeline.md
find assets client/assets tools/assets -type d | sort
```

## Task 2: Shared item visual metadata

Files:
- Create: `shared/assets/item_visuals.v0.schema.json`
- Create: `shared/assets/item_visuals.v0.json`
- Create: `shared/golden/item_visual_resolution.v0.schema.json`
- Create: `shared/golden/item_visual_resolution.json`
- Modify: `tools/validate_shared.py`

- [ ] Step 2.1: Define schema for `item_visuals` including `local_transform` (vec3 position,
  rotation_degrees, scale).
- [ ] Step 2.2: Add v0 data mapping `rusty_sword` → `weapon_rusty_sword_v0` / `right_hand_socket`.
- [ ] Step 2.3: Add golden fixture per spec §4.6.
- [ ] Step 2.4: Extend `validate_shared.py` in **two** functions: `iter_instances()` to glob
  `shared/assets/*.json` (it currently only scans rules/golden/protocol-examples), and
  `schema_for()` to map `item_visuals.v0.json` → `item_visuals.v0.schema.json`. Add cross-checks
  (spec §4.9 #1): every `item_visuals` key exists in `items.v0.json` with a matching `slot`; golden
  fixture matches the metadata. (The new golden `item_visual_resolution.json` is auto-picked-up by
  the existing golden glob + `stem + ".v0.schema.json"` mapping — no extra wiring.)
- [ ] Step 2.5: Add `make validate-shared` still passes with v0 fixtures untouched.

Verification:

```bash
make validate-shared
```

## Task 3: Asset manifest and Python validator

Files:
- Create: `assets/manifests/assets.v0.schema.json`
- Create: `assets/manifests/assets.v0.json`
- Create: `tools/assets/validate_assets.py`
- Create: `tools/assets/test_validate_assets.py`
- Modify: `Makefile`

- [ ] Step 3.1: Define manifest schema (`character` / `equipment`, paths, required_nodes, slot).
- [ ] Step 3.2: Add manifest entries for `character_base_humanoid_v0` and `weapon_rusty_sword_v0`
  pointing at `client/assets/...` runtime paths.
- [ ] Step 3.3: Implement `tools/assets/validate_assets.py`:
  - schema-validate manifest
  - verify every `runtime_path` exists on disk
  - verify every `item_visuals` `asset_id` resolves in manifest
  - verify character `required_nodes` cover all referenced `mount_socket` values
  - optional: parse GLB JSON chunk and confirm `required_nodes` names exist (warn-only if parser
    fails; hard-fail on explicit mismatch)
- [ ] Step 3.4: Add pytest coverage for validator happy path and representative failures.
- [ ] Step 3.5: Add `make validate-assets` Makefile target.

Verification:

```bash
make validate-assets   # fails until GLBs exist — expected before Task 4
make tools && .venv/bin/python -m pytest -q tools/assets
```

## Task 4: Runtime GLB assets

Decision (spec §9 #5): **fetch CC0 low-poly GLBs**, commit them, record provenance. Fallback to a
committed stdlib generator only if no suitable CC0 model is found.

Files:
- Create: `client/assets/characters/base_humanoid/base_humanoid.glb`
- Create: `client/assets/equipment/weapons/rusty_sword/rusty_sword.glb`
- Modify: `assets/manifests/assets.v0.json` (add `provenance`; adjust paths if they differ)

- [ ] Step 4.1: Source a CC0 low-poly humanoid GLB (Kenney.nl / Quaternius). Verify it's glTF-binary
  (`.glb`), embedded materials, < 5 MB, readable under the isometric camera. Convert if the source
  ships glTF-separate or another format.
- [ ] Step 4.2: Source a CC0 one-handed weapon GLB (rusty-sword-equivalent), < 2 MB, handheld scale.
- [ ] Step 4.3: Confirm embedded materials need no network fetches (textures embedded or vertex-color).
- [ ] Step 4.4: Add `provenance` (source URL, `CC0`, `sha256`) to each manifest entry and record the
  same in as-built notes. The asset validator checks `sha256` against the committed file.
- [ ] Step 4.5: Re-run `make validate-assets` until green.
- [ ] **Fallback only** (no CC0 asset found): create `tools/assets/gen_glb.py` (stdlib `struct`,
  no new deps) emitting deterministic primitive `.glb` files; document it as the source-of-truth.

Verification:

```bash
make validate-assets
godot --headless --path client --import
sha256sum client/assets/**/*.glb   # must match manifest provenance entries
```

## Task 5: Godot character scene and socket

Files:
- Create: `client/scenes/character.tscn`
- Modify: `client/scenes/main.tscn`

- [ ] Step 5.1: Create `character.tscn` instancing `res://assets/characters/base_humanoid/base_humanoid.glb`.
- [ ] Step 5.2: Add `ModelRoot/right_hand_socket` as `Node3D` placeholder with transform tuned so
  a weapon appears handheld (adjust in editor or via manifest `local_transform`).
- [ ] Step 5.3: Refactor `main.tscn` / `main.gd` scene graph per spec §5.1 (`World/PlayerAnchor/
  CharacterVisual`). Keep monster/loot as v0 primitives under `Entities`.
- [ ] Step 5.4: Player position reconciliation moves `PlayerAnchor`, not a raw capsule mesh.

Verification:

```bash
godot --headless --path client --import
# Manual: run client, confirm humanoid visible under isometric camera
```

## Task 6: Equipment visual resolver

Files:
- Create: `client/scripts/equipment_visuals.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/net_client.gd` (add optional `resume_session_id` to `create_session()` —
  server already supports it; needed by Task 8.4 / acceptance #8)

- [ ] Step 6.1: Implement `EquipmentVisualResolver` loading `shared/assets/item_visuals.v0.json` and
  `assets/manifests/assets.v0.json` via repo-relative paths (spec §4.8). The resolver takes the
  **mount-root node injected as a dependency** — never `get_node("/root/Main/...")` — so `main.gd`
  (full scene) and `smoke.gd` (minimal `SceneTree` subtree) share one code path.
- [ ] Step 6.2: On equip changes, resolve `item_instance_id` → `item_def_id` from inventory cache,
  then metadata → `asset_id` → manifest `runtime_path` → `res://` GLB path. The path transform:
  strip the leading `client/` from `runtime_path` and prepend `res://` (project root is `client/`).
- [ ] Step 6.3: Mount weapon under named socket; child node name = `asset_id`; apply
  `local_transform`.
- [ ] Step 6.4: Replace prior weapon child on re-equip (no duplicate stale nodes — spec §7).
- [ ] Step 6.5: Handle unknown def/asset/socket: emit structured warning, render nothing, no crash.
- [ ] Step 6.6: Implement `get_debug_state()` per spec §4.4 / §4.7.
- [ ] Step 6.7: Wire resolver from `session_snapshot`, `inventory_add`, `inventory_update`, and
  `equipped_update` handlers in `main.gd`.
- [ ] Step 6.8: On resume snapshot with pre-equipped weapon, mount without waiting for a new delta.

Verification:

```bash
# With server running: kill → pickup → equip → Q; weapon visible on character
make server &
make bot   # server authority unchanged
```

## Task 7: GDScript tests

Files:
- Create: `client/tests/test_item_visuals.gd`
- Modify: `scripts/client_smoke.sh`

- [ ] Step 7.1: Add headless test reading `shared/golden/item_visual_resolution.json` and
  `shared/assets/item_visuals.v0.json`; assert expected asset_id and mount_socket.
- [ ] Step 7.2: Optionally assert manifest resolves `runtime_path` to an existing
  `res://assets/...` import.
- [ ] Step 7.3: Wire test into `scripts/client_smoke.sh` before slice smoke.

Verification:

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
```

## Task 8: Extended headless smoke

Files:
- Modify: `client/scripts/smoke.gd`

- [ ] Step 8.1: Instantiate `EquipmentVisualResolver` (or minimal scene subtree) in smoke harness.
- [ ] Step 8.2: After equip, assert server `/state` (existing) **and** `get_debug_state()`:
  `equipped_visuals.weapon.visible == true`, matching ids and socket.
- [ ] Step 8.3: Send move intent after equip; assert visual still mounted (spec acceptance #13).
- [ ] Step 8.4: Disconnect WS, resume the **same** session via `create_session(resume_session_id=…)`
  (Task 6 net_client change), reconnect WS, wait for `session_snapshot`, assert visual restored
  (spec acceptance #8). Server rehydrates equipped state via `ListInventory`→`Sim.LoadInventory`, so
  the resumed snapshot carries `equipped.weapon` — no `equipped_update` delta is needed.
- [ ] Step 8.5: Keep timeout and failure messages actionable.

Verification:

```bash
make db-up
make server &
make client-smoke
```

## Task 9: CI and docs integration

Files:
- Modify: `scripts/ci.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `Makefile` (if not done in Task 3)
- Modify: `README.md`

- [ ] Step 9.1: Insert `make validate-assets` into `scripts/ci.sh` after shared validation.
- [ ] Step 9.2: Add asset validation step to GitHub Actions workflow.
- [ ] Step 9.3: Document `make validate-assets` in README dev commands table.
- [ ] Step 9.4: Update ADR-0006 and spec as-built section if implementation deviates.

Verification:

```bash
make ci
```

## Sequencing notes

- Tasks 1–3 can proceed before GLB files exist; `validate-assets` fails until Task 4 — that is
  expected.
- Do not change WebSocket schemas or server equip logic unless a proven gap appears (spec §4.5).
- Do not block the Python bot on visuals; bot remains authoritative-state-only.
- Keep monster/loot as v0 primitives; only the local player adopts the GLB pipeline in v2.
- Prefer extracting visual logic into `equipment_visuals.gd` so smoke and main share one code path.

## Final verification

```bash
make validate-shared
make validate-assets
cd server && go test ./...
make tools && .venv/bin/python -m pytest -q tools
make ci
```

Expected: full CI green including extended Godot smoke when Godot is installed locally.

## As-Built Notes

### Decisions locked before implementation (2026-06-05 plan review)

- **Asset authorship (spec §9 #5):** fetch CC0 low-poly GLBs (Kenney.nl / Quaternius), commit them,
  record provenance (`source_url` + `CC0` + `sha256`) in the manifest. Stdlib generator is fallback
  only. Rationale: rehearses the real long-run pipeline; the manifest→import→mount contract is the
  proof, not the authorship tool. ADR-0006 encodes the contract, treats authorship as swappable.
- **Resume (acceptance #8) is server-ready:** `POST /v0/sessions` accepts `resume_session_id` and the
  WS handler rehydrates inventory **and** equipped state (`ListInventory`→`Sim.LoadInventory`,
  verified in `server/internal/http/session.go`, `realtime/hub.go`, `game/sim.go`). The only change
  is client-side: `net_client.gd::create_session()` gains an optional `resume_session_id`. No
  protocol/server change.
- **Resolver decoupling:** `EquipmentVisualResolver` takes the mount-root node as an injected
  dependency (no absolute `/root/Main/...` lookups) so the full scene and the headless `SceneTree`
  smoke share one path. `node_path` in debug state is therefore environment-dependent and is not
  asserted verbatim.
- **Path transform:** manifest `runtime_path` (`client/assets/...`, repo-root-relative) → Godot
  resource path by stripping `client/` and prepending `res://`.

### Implementation log

Implemented Tasks 1–9 on branch `feature/equip-and-see-it`, one commit per task,
testing progressively. Final gate `make ci` is green including the extended Godot smoke.

Deviations from the plan (all sanctioned by the spec/ADR decisions):

- **Asset authorship (Task 4): used the stdlib generator (`tools/assets/gen_glb.py`), not a CC0
  fetch.** This is spec §9 #5's documented fallback, chosen deliberately for two reasons beyond
  sandbox network constraints: (a) the asset validator hard-checks `required_nodes` against the GLB,
  and a fetched humanoid would not contain a node named `right_hand_socket` — so the CC0 path would
  require post-processing the binary anyway, erasing the "no authorship tool" benefit; (b) a
  generator produces **byte-deterministic** output, giving a stable `provenance.sha256` and
  reproducible CI. The generator is the recorded source-of-truth; provenance is
  `origin=gen_glb.py / license=CC0-1.0 / sha256`. The manifest→import→mount contract is identical
  regardless of source, which is the actual proof.
- **Socket lives in the GLB, not as a scene placeholder (Task 5, Step 5.2).** The generator embeds
  `right_hand_socket` as an empty node in `base_humanoid.glb`, so `character.tscn` instances the GLB
  as `ModelRoot` and the socket arrives with it — no duplicate `Node3D` is added (which would create
  an ambiguous `right_hand_socket`). The §4.4 path `CharacterVisual/ModelRoot/right_hand_socket`
  holds; verified headless. This also makes the Python GLB node-name check a hard pass, not warn-only.
- **`node_path` in debug state is environment-dependent (spec §4.4/§4.7).** In the headless smoke's
  minimal `SceneTree` subtree the mounted node reports `node_path == ""`; tests assert the
  `asset_id`/ids/`visible`/socket fields, never the absolute path. In the interactive `main.tscn`
  tree it is populated.

Verification highlights:
- `make validate-shared` (39 checks) and `make validate-assets` (9 checks) green.
- `tools/assets/test_validate_assets.py` (9 cases) green.
- Godot smoke proved acceptance #6/#7 (server `/state` + `get_debug_state().weapon.visible`),
  #13 (move-after-equip still mounted), and #8 (resume restores the visual from `session_snapshot`
  alone via a fresh client + resolver — no `equipped_update` delta). No server/protocol changes.
