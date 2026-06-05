# Spec: `animate-and-react`

Status: Implemented (2026-06-05)
Branch: `feature/animate-and-react`
Slice: v3 — skeletal animation + bone-attached weapon + monster reactions
Baseline: slice v2 `equip-and-see-it` (complete — `make ci` green)
Related: ADR-0001 (tech stack), ADR-0006 (asset pipeline — predicts the
`Node3D → BoneAttachment3D` upgrade this slice executes), ADR-0007 (client
animation state model — authored in this slice).

## 1. Purpose

Prove the `rigged GLB → Skeleton3D → state-driven AnimationPlayer` pipeline on
**two** entity types driven by **two** distinct signal sources, demonstrating
that the animation layer is entity-agnostic and works off both client-derived
and authoritative-event state:

- The **local player** plays `idle / walk / attack` driven entirely by
  **client-side signals** already present in `main.gd` (predicted movement,
  local attack input). The equipped weapon stops being tilted by a fake tween
  and instead **rides the `hand_r` bone** via a `BoneAttachment3D` — so the
  attack swing is real skeletal motion.
- The **monster** (training dummy) plays `hit-react / death` driven by the
  **authoritative `monster_damaged` / `monster_killed` events** the server
  already transmits in every `state_delta`. The client simply starts reading
  the `events` array it currently ignores.

The proof is the pipeline and its two drivers, not art quality. Crude generated
rigs are acceptable, exactly as the equip slice treated crude meshes.

## 2. Non-goals

- **No protocol or server change.** The only wire-facing change is that the
  client begins reading the already-transmitted `events` array. No new message
  types, no schema bump, no sim change.
- **No player damage or death.** Server combat is one-directional
  (player → monster only; `handleAttack` rejects any non-monster target).
  Player hit-react/death has no authoritative trigger and is out of scope.
- **No `AnimationTree` / blend spaces.** Discrete clips with a small priority
  state machine only.
- **No respawn.** A killed monster persists at `hp == 0` (the server does not
  remove it), so the death pose simply stays.
- **Loot stays a primitive.** Only the player and the monster adopt the
  animated GLB pipeline; loot remains a runtime `MeshInstance3D`.
- **Art quality is a non-goal.** Deterministic generated rigs, no fetched art.

## 3. Files to create or modify

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/adr/0007-animation-state-model.md` | Durable decision: client-derived vs event-driven state, discrete clips, no protocol change |
| Modify | `docs/adr/0006-asset-pipeline.md` | As-built: `Node3D` socket → `BoneAttachment3D` on `hand_r` (predicted upgrade executed) |
| Modify | `tools/assets/gen_glb.py` | Emit skinned, rigged GLBs (skeleton joints, skin, inverse bind matrices, weights); add `monster_dummy.glb` |
| Modify | `client/assets/characters/base_humanoid/base_humanoid.glb` | Regenerated rigged character (overwrites v2 static mesh) |
| Modify | `client/assets/characters/base_humanoid/base_humanoid.glb.import` | Godot import metadata for the regenerated rig |
| Create | `client/assets/monsters/dummy/monster_dummy.glb` | Rigged training-dummy model |
| Create | `client/assets/monsters/dummy/monster_dummy.glb.import` | Godot import metadata for the dummy rig |
| Modify | `assets/manifests/assets.v0.schema.json` | Extend asset `type` enum with `monster` |
| Modify | `assets/manifests/assets.v0.json` | Add `monster_dummy_v0`; change character `required_nodes` to joint names incl. `hand_r`; keep item visual `mount_socket` as `right_hand_socket`; regenerate `sha256` |
| Modify | `tools/assets/validate_assets.py` | Validate `monster_dummy_v0`; hard-check mount bone `hand_r` + dummy `pivot` among GLB skin joints, not just node names |
| Modify | `tools/assets/test_validate_assets.py` | Cover monster entry + bone-name checks (happy + failure) |
| Modify | `client/scenes/character.tscn` | `right_hand_socket` becomes `BoneAttachment3D` (bone `hand_r`); add `AnimationPlayer` clips `idle/walk/attack` |
| Create | `client/scenes/monster_dummy.tscn` | Dummy GLB + `AnimationPlayer` clips `idle/hit/death` |
| Create | `client/scripts/animation_controller.gd` | Injected `AnimationController` (one script, per-entity instances) |
| Modify | `client/scripts/equipment_visuals.gd` | Delete `play_attack_swing()` + tween machinery; mount logic unchanged |
| Modify | `client/scripts/main.gd` | Instantiate controllers; locomotion from input/prediction; attack one-shot; read `events`; monster nodes use `monster_dummy.tscn`; `entities` entries become `{node, controller, type}` dictionaries |
| Modify | `client/scripts/smoke.gd` | Assert clip state + bone-mounted weapon; drive monster hit/death event path; resume death pose |
| Create | `client/tests/test_animation.gd` | Headless: rig/scene assertions + controller state-machine logic |
| Modify | `scripts/client_smoke.sh` | Run `test_animation.gd` alongside `test_item_visuals.gd` |

## 4. Data shapes

### 4.1 Rigged character GLB

`base_humanoid.glb` (regenerated by `gen_glb.py`) supplies **mesh + skeleton +
skin data + bone names only**. No animation tracks live in the GLB; clips are
authored in Godot (§4.4).

Required skeleton joints (glTF nodes — joints are nodes in glTF):

| Joint | Role |
|-------|------|
| `root` | Skeleton root |
| `spine` | Torso |
| `arm_r` | Right upper arm (drives the attack swing) |
| `hand_r` | **Weapon mount bone** — the `BoneAttachment3D` binds here |
| `leg_l`, `leg_r` | Legs (give the walk clip something to move) |

### 4.2 Rigged monster GLB

`monster_dummy.glb` — a post on a base. Required joints:

| Joint | Role |
|-------|------|
| `root` | Skeleton root / base |
| `pivot` | The bone the `hit` (wobble) and `death` (topple) clips rotate |

### 4.3 Asset manifest

Add a monster entry, update the character's required nodes, and keep the v2
manifest field name `type` (not `kind`):

```json
{
  "assets": {
    "character_base_humanoid_v0": {
      "type": "character",
      "source_path": "assets/characters/base_humanoid/base_humanoid.blend",
      "runtime_path": "client/assets/characters/base_humanoid/base_humanoid.glb",
      "format": "glb",
      "scale_unit": "meters",
      "required_nodes": ["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"],
      "provenance": { "origin": "gen_glb.py", "license": "CC0-1.0", "sha256": "<regenerated>" }
    },
    "weapon_rusty_sword_v0": { "...": "unchanged" },
    "monster_dummy_v0": {
      "type": "monster",
      "source_path": "assets/monsters/dummy/monster_dummy.blend",
      "runtime_path": "client/assets/monsters/dummy/monster_dummy.glb",
      "format": "glb",
      "scale_unit": "meters",
      "required_nodes": ["root", "pivot"],
      "provenance": { "origin": "gen_glb.py", "license": "CC0-1.0", "sha256": "<regenerated>" }
    }
  }
}
```

The manifest schema (`assets.v0.schema.json`) already permits these fields
except the new `type: "monster"` enum value; extend only that enum. The schema
continues to use `additionalProperties: false`, so examples and committed data
must include the required `type`, `runtime_path`, and `format` fields exactly.

`shared/assets/item_visuals.v0.json` remains keyed by the socket name
`right_hand_socket`. That socket name is still the rendering contract for
equipment; the manifest `required_nodes` for the character is upgraded to the
underlying rig joints because the validator now checks the skinned GLB skeleton.

### 4.4 Animation clips (Godot-authored)

Clips live in the `.tscn` / `.tres`, targeting `Skeleton3D` bone-pose tracks.
They are text-reviewable in PRs and deterministic.

| Scene | AnimationPlayer clips | Notes |
|-------|----------------------|-------|
| `character.tscn` | `idle` (loop), `walk` (loop), `attack` (one-shot) | `attack` rotates `arm_r`/`hand_r`; the mounted weapon rides it |
| `monster_dummy.tscn` | `idle` (loop), `hit` (one-shot), `death` (one-shot, terminal pose) | `hit` wobbles `pivot`; `death` topples `pivot` and holds |

### 4.5 `AnimationController` (client)

`class_name AnimationController extends RefCounted`. Constructor takes an
injected `AnimationPlayer` (never an absolute scene-path lookup — same rule as
`EquipmentVisualResolver`). One instance per animated entity; the **same
script** serves player and monster.

State priority (highest wins):

```
terminal (death)  >  one-shot (attack / hit)  >  locomotion (idle / walk)
```

API:

| Method | Behavior |
|--------|----------|
| `set_locomotion(is_moving: bool)` | If no one-shot/terminal active → play `walk` when moving, else `idle` (looped). |
| `play_one_shot(name: String)` | Play `name` once; on `animation_finished` fall back to locomotion. Ignored if terminal latched. Re-triggering restarts it. |
| `enter_terminal(name: String)` | Play `name`, latch `terminal = true`; ignore all further calls (pose persists). |
| `current_clip() -> String` | Currently playing clip name (for assertions). |
| `get_debug_state() -> Dictionary` | `{ current_clip, terminal, is_moving, warnings }` for headless smoke. |

The controller does **not** parse protocol events and does not know entity
types. The event→clip mapping lives in `main.gd` as a small client-only
presentation constant:

```gdscript
const MONSTER_EVENT_CLIPS := {
	"monster_damaged": "hit",
	"monster_killed": "death",
}
```

This deliberately does **not** belong in `shared/`, which is reserved for
cross-language server/client contracts. `AnimationController` tests cover the
generic state machine; `main.gd`/`smoke.gd` tests cover the event mapping.

### 4.6 Protocol impact

None. `state_delta` already carries `events` alongside `changes`
(`server/internal/realtime/protocol.go`, `runner.go`). `monster_damaged` and
`monster_killed` are already on the wire; the client begins consuming the
`events` array it currently ignores in `main.gd::_apply_delta`.

## 5. Architecture and flow

### 5.1 Scene graph (from v2)

```
Main (main.gd)
└─ World
   └─ PlayerAnchor                         # follows authoritative player position
      └─ CharacterVisual (character.tscn)
         └─ ModelRoot (base_humanoid.glb)
            └─ …/Skeleton3D                # joints: root, spine, arm_r, hand_r, leg_l, leg_r
               └─ right_hand_socket        # BoneAttachment3D → bone "hand_r" (was empty Node3D in v2)
         └─ AnimationPlayer                # clips: idle, walk, attack
└─ Entities
   └─ <monster> (monster_dummy.tscn)       # was a primitive MeshInstance3D
      └─ ModelRoot (monster_dummy.glb) + Skeleton3D (root, pivot)
      └─ AnimationPlayer                   # clips: idle, hit, death
   └─ <loot> (primitive MeshInstance3D)    # unchanged
```

### 5.2 Player flow (client-derived)

1. `_ready`: build `AnimationController` over the character `AnimationPlayer`
   (found within the injected `character_visual`).
2. `_process`: derive `is_moving` from the local movement input that actually
   advances prediction while the WebSocket is open; call
   `set_locomotion(is_moving)`. If the client is disconnected or no movement
   key is active, set locomotion to idle. Do not infer movement solely from
   reconciliation snaps, or the player may walk while passively receiving a
   snapshot.
3. `_try_attack_toward_mouse`: replace `resolver.play_attack_swing()` with
   `anim.play_one_shot("attack")`.

### 5.3 Monster flow (authoritative-event-driven)

1. `_upsert_entity` for a `monster` instantiates `monster_dummy.tscn` and builds an
   `AnimationController` over its `AnimationPlayer`, starting `idle`. The
   `entities` dictionary entry carries
   `{ "node": Node3D, "controller": AnimationController | null, "type": String }`;
   loot entries keep `controller == null`.
2. `_apply_delta` applies `payload.changes` first, then loops over
   `payload.events`. This ordering ensures a monster spawned/updated in the same
   delta exists before its event is rendered and that `hp == 0` has already
   landed before death is latched:
   - `monster_damaged` → `entities[event.entity_id].controller.play_one_shot("hit")`
   - `monster_killed`  → `entities[event.entity_id].controller.enter_terminal("death")`
   - all other events ignored for visuals.
3. Idempotency: `enter_terminal` latching makes repeated `monster_killed`, or a
   late `monster_damaged` after death, harmless — death wins.

### 5.4 Resume / snapshot consistency

`session_snapshot` carries `recent_events`, but this slice must not depend on
them for visual state. Monster entities carry `hp`; on snapshot, `_upsert_entity`
creates the controller and immediately calls `enter_terminal("death")` when
`type == "monster"` and `hp == 0`. A live monster starts `idle`. This mirrors the
equip slice restoring a pre-equipped weapon from the snapshot alone and avoids
replaying historical events.

### 5.5 Bone attachment & resolver simplification

- `right_hand_socket` keeps its **name** but becomes a `BoneAttachment3D` child
  of `Skeleton3D` bound to `hand_r`.
- `EquipmentVisualResolver` mount logic is **unchanged**: it still
  `find_child("right_hand_socket", true, false)` and `add_child(weapon)`; it
  neither knows nor cares the socket is now a bone attachment.
- **Delete** from the resolver: `play_attack_swing()`, `_swing_tween`,
  `_rest_rotation_degrees`, the `ATTACK_SWING_*` constants, and the tween-kill in
  `_clear_mounted`. The resolver returns to being only a mount resolver.
- `item_visuals.v0.json` `local_transform` (grip offset) is still applied and
  will likely need **re-tuning** now that the weapon hangs off a moving bone —
  a data change in the JSON, not code.

### 5.6 Entity-node typing changes

`main.gd::_make_entity_node` currently returns a `MeshInstance3D`; v3 must widen
that to `Node3D` because monsters become instanced scenes while loot remains a
primitive `MeshInstance3D`. All entity accessors (`_best_monster_in_direction`,
`_remove_entity`, debug counts, smoke helpers) must read through
`entities[id]["node"]` and must not assume the stored value is directly a mesh.

When `monster_dummy.tscn` fails to load, `_make_entity_node("monster")` returns
the same red primitive fallback as v2 plus `controller == null`. Positioning and
target selection still work; only animation warnings are emitted.

## 6. Asset constraints

Same spirit as the equip slice (soft budgets; exceeding requires an as-built
note):

- `base_humanoid.glb` < 5 MB; `monster_dummy.glb` < 2 MB.
- Embedded materials / vertex color only — no network fetches at import/runtime.
- Byte-deterministic output from `gen_glb.py` → stable `sha256` provenance.
- glTF-binary (`.glb`), readable under the existing isometric camera.
- Generated rigs must include a non-empty `skins` array, `skin.joints` pointing
  at the required joint nodes, inverse-bind-matrix data, and mesh attributes
  `JOINTS_0` + `WEIGHTS_0`. A GLB that only contains meshless named nodes is a
  v2 socket placeholder and does **not** satisfy this slice.
- Godot-authored clips may be crude, but they must target imported `Skeleton3D`
  bone poses rather than standalone `Node3D` transforms; otherwise the weapon
  cannot prove it rides `hand_r`.

## 7. Failure behavior

- Unknown / missing clip name: controller emits a structured warning and stays
  in its current state; never crashes.
- Event references an unknown `entity_id`: ignored with a warning (the entity
  may have been culled).
- Monster scene fails to load: fall back to a primitive `MeshInstance3D` with a
  warning, so the entity is still positioned authoritatively (no crash).
- Skeleton/skin import failure is a **hard CI failure** at the rig gate (§10),
  not a silent degrade.
- Missing `AnimationPlayer` or missing `right_hand_socket` in an otherwise
  loaded scene is a hard client test failure. The fallback path is only for an
  unloaded monster scene, not for a malformed rigged scene committed to the repo.
- Repeated events in one delta are processed in array order, with terminal
  precedence enforced by the controller. Example:
  `monster_damaged`, then `monster_killed` in the same `events` array ends in
  `death`; `monster_killed`, then `monster_damaged` also stays in `death`.

## 8. Acceptance criteria

1. `make validate-assets` passes including the new `monster_dummy_v0` and the
   `hand_r` / `pivot` skin-joint checks.
2. `godot --headless --path client --import` imports both rigged GLBs; a
   skeleton-introspection assertion confirms bone `hand_r` (character) and
   `pivot` (dummy) exist, and both imported scenes have non-empty skeletons.
3. `character.tscn` exposes a `BoneAttachment3D` named `right_hand_socket` bound
   to `hand_r`, and an `AnimationPlayer` with clips `idle/walk/attack`.
4. `monster_dummy.tscn` exposes an `AnimationPlayer` with clips `idle/hit/death`.
5. Player: moving toggles `walk`/`idle`; attack input plays `attack` once and
   returns to locomotion.
6. Equipped weapon is mounted under `right_hand_socket` and **visibly rides the
   arm** during `attack` (real skeletal motion; no tween).
7. Monster: a `monster_damaged` event plays `hit`; a `monster_killed` event
   plays `death` and the pose persists (terminal).
8. Resume: a session snapshot with a monster at `hp == 0` shows that monster in
   the `death` pose without any delta.
9. `EquipmentVisualResolver` no longer contains `play_attack_swing` or tween
   machinery; mount/resolve logic is unchanged.
10. `main.gd` still targets monsters correctly after `entities` changes from
    `id -> MeshInstance3D` to `id -> {node, controller, type}`.
11. No server/protocol change: `server` diff is empty; `go test ./...` passes
    unchanged.
12. `make ci` is green including the extended Godot smoke.

## 9. Decisions (resolved)

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Asset authorship | **Generated rigs + Godot-authored clips.** | Keeps byte-deterministic CI and lets the validator hard-check bone names — the same logic the equip slice used to choose the generator over a fetch. A fetched rig would lack predictable bone names and need binary post-processing. |
| 2 | Animation clips in GLB or Godot? | **Godot (`.tscn`/`.tres`).** | Authoring sampler/channel data in raw glTF bytes is brutal; Godot clips are text-reviewable and decouple authorship from the rig. The GLB provides mesh + skeleton + skin data only. |
| 3 | Player reaction states? | **`idle/walk/attack` only.** | Player damage/death has no authoritative trigger (one-directional combat). Out of scope. |
| 4 | Monster reaction trigger | **Authoritative `monster_damaged` / `monster_killed` events** (already on wire). | Real triggers, no server change; proves event-driven animation distinct from client-derived locomotion. |
| 5 | Monster fidelity | **Minimal dummy rig + clips, same pipeline.** | A training dummy is a rigid post; a wobble/topple on a `pivot` bone proves the pipeline generalizes to a second entity without a full humanoid rig. |
| 6 | Socket upgrade | **`Node3D` → `BoneAttachment3D` on `hand_r`, same node name.** | Executes ADR-0006's predicted upgrade; resolver mount logic unchanged because it resolves by name. |
| 7 | Controller architecture | **Injected `AnimationController` (RefCounted), one script per entity.** | Mirrors `EquipmentVisualResolver`: no absolute paths, one code path for `main.gd` + `smoke.gd`, clip state assertable headlessly. |
| 8 | Cross-language golden? | **No.** | Animation is client-only presentation; `shared/` is for cross-language contracts. |
| 9 | Docs | **ADR-0006 as-built + new ADR-0007 + spec.** | Bone attachment was predicted by 0006; the animation state model is a new durable decision worth its own ADR. |

## 10. Testing plan

**Rig gate (fail-fast, runs before any animation work builds on the rig):**
`godot --headless --path client --import` then a skeleton-introspection script
asserting bones `hand_r` (character) and `pivot` (dummy) exist and the skin
imports as real `Skeleton3D`/skinned mesh data (non-empty skeleton; mesh AABB
sane). Catches bad skin math at the source.

**Python (`make validate-assets`, `pytest tools/assets`):** monster entry,
sha256, `type` enum, and skin-joint hard-checks; happy + failure cases.

**Headless GDScript (`client/tests/test_animation.gd`, run in
`client_smoke.sh`):**
- Scene/rig: clip lists, `BoneAttachment3D` bound to `hand_r`, dummy bones.
- Controller logic: locomotion toggle; one-shot returns to locomotion on
  finish; terminal latches and ignores subsequent calls; `hit` after death
  ignored.

**Extended slice smoke (`smoke.gd`):**
- After equip + attack: weapon still mounted under `right_hand_socket`; player
  `current_clip()` is `attack` then settles.
- Drive simulated `monster_damaged` → `monster_killed` through `_apply_delta`'s
  event path against an instanced dummy subtree; assert `hit` → terminal
  `death`.
- Resume: snapshot with monster `hp == 0` ⇒ terminal `death`, no delta.
- Deterministic clip-name/state assertions only — never pixel/pose values
  (same rule the equip slice used for `node_path`).

**Server:** `go test ./...` unchanged; the `server/` diff must be empty.

## 11. As-Built Notes

Status: **Implemented (2026-06-05)**. The slice shipped as designed:
`gen_glb.py` emits skinned, rigged GLBs (skin joints, inverse-bind matrices,
`JOINTS_0`/`WEIGHTS_0`) for both `base_humanoid` and the new `monster_dummy`;
`validate_assets.py` hard-checks the mount bone `hand_r` and the dummy `pivot`
as actual glTF **skin joints** (not just node names); animation is client-only
presentation state with discrete clips driven by an injected
`AnimationController` (`terminal > one-shot > locomotion`); the player derives
`idle/walk/attack` from input/prediction and the monster's `hit/death` from the
authoritative `monster_damaged`/`monster_killed` events already on the wire. No
server or protocol change was made (acceptance #11; empty `server/` diff).

### Deviations from the original design

- **Socket attached in `character_visual.gd`, not declared in `.tscn`.** The
  `BoneAttachment3D` named `right_hand_socket` is created in code and parented to
  the imported `Skeleton3D` at `_ready`, rather than being authored as a node in
  `character.tscn`. This is deliberate robustness against the exact imported
  skeleton node path (which Godot's glTF importer can rename); the resolver still
  finds the socket by name, so the ADR-0006 D4 contract is unchanged. Sanctioned
  by the §9 decision #6 (consumers key off the socket name).
- **Clips authored as committed `.tres`, built by a tool script.** Animation
  clips are not hand-edited inside the `.tscn`; they are produced by
  `client/tools/build_animations.gd` and committed as
  `client/animations/character_anims.tres` and
  `client/animations/monster_anims.tres`, then referenced by the scenes. This
  keeps clip authoring text-reviewable and regenerable, consistent with §9
  decision #2 (Godot-authored clips, not GLB tracks).

### Known deviations recorded honestly

1. **Resume death pose vs. server respawn (spec defect).** Acceptance #8 and
   §5.4 assume the server persists a killed monster at `hp == 0` across resume
   (§2 "no respawn"). The actual server resume path
   (`server/internal/realtime/hub.go`: `game.NewSim(seed)` + `LoadInventory`)
   reconstructs the world from **seed + inventory only** and does **not** persist
   monster death — so a resumed snapshot contains the monster **respawned at full
   hp**. Combined with acceptance #11 (no server change), end-to-end
   server-persisted-death-on-resume is currently **unsatisfiable in this slice**.
   The smoke (`smoke.gd::_resume_monster_from_snapshot`) therefore verifies the
   **client wiring** — a snapshot monster with `hp <= 0` drives the terminal
   `death` clip with no event replay — by forcing `hp = 0` on the snapshot
   monster entity. This is a **server-side follow-up to file** (persist monster
   death across resume); it was **not** done in this slice.

2. **CI gate hardened against a parse-error false-PASS.** Godot exits `0` even on
   a GDScript PARSE/load error when run via `--script`, so a broken gate script
   could print no output yet still "pass" on its exit code. `scripts/client_smoke.sh`
   was hardened: each Godot gate's combined output is captured and the script
   asserts the gate's expected success sentinel is present, failing nonzero
   otherwise. Sentinels: item-visual test → `[gdtest] PASS`; rig gate →
   `[rig-gate] PASS`; animation test →
   `[gdtest] PASS: animation controller + scenes`; slice smoke → `[smoke] PASS`.
   The Godot-absent SKIP path (exit 0 when the runtime is not installed) is
   preserved. This makes a green `make ci` actually meaningful.
