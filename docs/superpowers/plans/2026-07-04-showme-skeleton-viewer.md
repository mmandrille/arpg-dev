# showme skeleton viewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `--focus skeleton` to the showme tool so engineers can inspect every bone position on any character class in a T-spread pose with red dots at every bone and colored labeled spheres at every equipment socket.

**Architecture:** New focused file `showme_skeleton_capture.gd` (following the `showme_item_icons_capture.gd` pattern) handles all skeleton-visualization logic. `visual_capture.gd` gets a single preload + one match arm (staying within its 13-line growth budget). `render_focus.py` gets a new `"skeleton"` choice and size override. SKILL.md gets a catalog row.

**Tech Stack:** GDScript 4 (Godot), Python 3, `Skeleton3D` bone pose API, `BoneAttachment3D` socket nodes.

## Global Constraints

- `visual_capture.gd` baseline is 1042, current is 1054 — may grow by at most 13 more lines total (ratchet ceiling 1067). Do NOT add the skeleton logic inline; it goes in the new file.
- New files must stay under 600 lines.
- No changes to server, protocol, shared rules, or golden fixtures.
- `render_focus.py` choices list must stay in sync with `visual_capture.gd` match block.
- All GDScript loaded as `SceneTree` scripts must `quit(0)` or `quit(1)` — never hang.
- Spec: `docs/superpowers/specs/2026-07-04-showme-skeleton-viewer-design.md`

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `client/scripts/showme/showme_skeleton_capture.gd` | **Create** | All skeleton viz logic: model load, pose spread, bone dots, socket spheres |
| `client/scripts/showme/visual_capture.gd` | **Modify** | +1 preload, +2 lines in match block (within budget) |
| `skills/showme/scripts/render_focus.py` | **Modify** | Add `"skeleton"` choice + 800×600 size override |
| `.claude/skills/showme/SKILL.md` | **Modify** | Add catalog row |
| `tools/test_showme.py` | **Modify** | Add test verifying `"skeleton"` appears in render_focus.py choices |

---

### Task 1: Create `showme_skeleton_capture.gd`

**Files:**
- Create: `client/scripts/showme/showme_skeleton_capture.gd`

**Interfaces:**
- Produces: `static func setup(capture: SceneTree, class_id: String) -> Node3D` — returns the character node (assigned to `_subject` in `visual_capture.gd` for live-mode rotation)

- [ ] **Step 1: Create the file with the full implementation**

Create `client/scripts/showme/showme_skeleton_capture.gd` with the complete content below:

```gdscript
class_name ShowmeSkeletonCapture
extends RefCounted

const CharacterScene := preload("res://scenes/character.tscn")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")

const SOCKET_COLORS := {
	"right_hand_socket": Color("#4488ff"),
	"off_hand_socket": Color("#4488ff"),
	"head_socket": Color("#ffee44"),
	"chest_socket": Color("#44dddd"),
	"belt_socket": Color("#44dddd"),
	"amulet_socket": Color("#44dddd"),
	"boots_socket": Color("#ff8822"),
	"gloves_socket": Color("#ff8822"),
	"ring_left_socket": Color("#cc44ff"),
	"ring_right_socket": Color("#cc44ff"),
}

const SPREAD_BONES := {
	"arm_r": Vector3(0.0, 0.0, -1.0),
	"arm_l": Vector3(0.0, 0.0, 1.0),
	"leg_r": Vector3(0.0, 0.0, -1.0),
	"leg_l": Vector3(0.0, 0.0, 1.0),
}
const ARM_SPREAD_ANGLE := PI / 2.0
const LEG_SPREAD_ANGLE := PI / 5.0


static func setup(capture: SceneTree, class_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "VisualFeedbackSkeleton"
	capture.get_root().add_child(root)

	_add_light(root)
	_add_camera(root)
	_add_floor(root)

	var character := CharacterScene.instantiate() as Node3D
	character.name = "FocusedCharacter"
	root.add_child(character)

	var effective_class := class_id if class_id != "" else "paladin"
	_apply_class_model(character, effective_class)
	await capture.process_frame
	await capture.process_frame

	var skel := character.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel != null:
		_spread_pose(skel)
		await capture.process_frame

		_place_bone_dots(skel, root)
		_place_socket_spheres(character, root)
	else:
		push_warning("[skeleton] no Skeleton3D found for class %s" % effective_class)
		_place_socket_spheres(character, root)

	return character


static func _spread_pose(skel: Skeleton3D) -> void:
	for bone_name in SPREAD_BONES.keys():
		var idx := skel.find_bone(str(bone_name))
		if idx < 0:
			continue
		var axis: Vector3 = SPREAD_BONES[bone_name]
		var angle := ARM_SPREAD_ANGLE if str(bone_name).begins_with("arm") else LEG_SPREAD_ANGLE
		skel.set_bone_pose_rotation(idx, Quaternion(axis.normalized(), angle))


static func _place_bone_dots(skel: Skeleton3D, root: Node3D) -> void:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#e03030")
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	for i in range(skel.get_bone_count()):
		var bone_transform: Transform3D = skel.global_transform * skel.get_bone_global_pose(i)
		var dot := MeshInstance3D.new()
		dot.name = "BoneDot_%s" % skel.get_bone_name(i)
		var mesh := SphereMesh.new()
		mesh.radius = 0.025
		mesh.height = 0.05
		dot.mesh = mesh
		dot.material_override = mat
		root.add_child(dot)
		dot.global_position = bone_transform.origin


static func _place_socket_spheres(character: Node3D, root: Node3D) -> void:
	for socket_name in SOCKET_COLORS.keys():
		var socket := character.find_child(str(socket_name), true, false) as Node3D
		if socket == null:
			continue
		var color: Color = SOCKET_COLORS[socket_name]
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
		var sphere := MeshInstance3D.new()
		sphere.name = "Socket_%s" % socket_name
		var mesh := SphereMesh.new()
		mesh.radius = 0.045
		mesh.height = 0.09
		sphere.mesh = mesh
		sphere.material_override = mat
		root.add_child(sphere)
		sphere.global_position = socket.global_position
		var label := Label3D.new()
		label.name = "Label_%s" % socket_name
		label.text = str(socket_name).replace("_socket", "")
		label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
		label.font_size = 28
		label.modulate = color
		root.add_child(label)
		label.global_position = socket.global_position + Vector3(0.0, 0.12, 0.0)


static func _apply_class_model(character: Node3D, class_id: String) -> void:
	var resolved := ClassPresentationsLoaderScript.resolve(class_id)
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
	if packed == null:
		return
	var old_model := character.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		character.remove_child(old_model)
		old_model.free()
	var model := packed.instantiate() as Node3D
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	ClassIdleStanceScript.apply_to_model(model, class_id)
	character.add_child(model)
	character.move_child(model, 0)
	var ap := character.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	if "class_id" in character:
		character.set("class_id", class_id)
	if character.has_method("_ensure_weapon_socket"):
		character.call("_ensure_weapon_socket")


static func _add_light(root: Node3D) -> void:
	var light := DirectionalLight3D.new()
	light.name = "key_light"
	light.light_energy = 2.2
	light.rotation_degrees = Vector3(-55, -35, 0)
	root.add_child(light)


static func _add_camera(root: Node3D) -> void:
	var camera := Camera3D.new()
	camera.name = "capture_camera"
	camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	camera.size = 3.2
	root.add_child(camera)
	camera.look_at_from_position(Vector3(3.2, 2.2, 4.5), Vector3(0.0, 1.1, 0.0), Vector3.UP)
	camera.current = true


static func _add_floor(root: Node3D) -> void:
	var floor := MeshInstance3D.new()
	floor.name = "reference_floor"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(4.0, 0.04, 4.0)
	floor.mesh = mesh
	floor.position = Vector3(0, -0.03, 0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color("#3f3f3c")
	floor.material_override = mat
	root.add_child(floor)
```

- [ ] **Step 2: Verify the file is syntactically valid (headless preload check)**

Run:
```bash
godot --headless --path /Users/mmandrille/git/arpg-dev/client --script res://tests/test_showme_load.gd 2>&1 | tail -5
```

Expected output contains: `[gdtest] PASS: showme capture scripts preload`

(This test only checks the existing files — next task wires in the new script.)

- [ ] **Step 3: Commit**

```bash
git add client/scripts/showme/showme_skeleton_capture.gd
git commit -m "feat: add ShowmeSkeletonCapture — bone dots + socket spheres in spread pose"
```

---

### Task 2: Wire skeleton focus into `visual_capture.gd` and `render_focus.py`

**Files:**
- Modify: `client/scripts/showme/visual_capture.gd` (lines ~1–8 for preload, ~60–106 for match block)
- Modify: `skills/showme/scripts/render_focus.py` (line ~24 choices, ~76 size overrides)

**Interfaces:**
- Consumes: `ShowmeSkeletonCapture.setup(capture: SceneTree, class_id: String) -> Node3D` from Task 1

- [ ] **Step 1: Add preload and match arm to `visual_capture.gd`**

At the top of `visual_capture.gd` where other `Const` preloads are listed (around line 20), add:

```gdscript
const ShowmeSkeletonCaptureScript := preload("res://scripts/showme/showme_skeleton_capture.gd")
```

In the `match _focus:` block (around line 60–106), add a new arm **before** the `_:` default arm:

```gdscript
		"skeleton":
			_subject = await ShowmeSkeletonCaptureScript.setup(self, _class_id)
```

- [ ] **Step 2: Add `"skeleton"` to `render_focus.py`**

In `skills/showme/scripts/render_focus.py`, the `--focus` choices list is on line ~24. Add `"skeleton"` to it:

```python
    parser.add_argument("--focus", choices=["gear", "classes", "floor-item", "inventory", "corpse", "corpse-inventory", "skills", "item-icons", "shop", "bishop", "market-board", "market-publish", "market-offer", "character-menu", "join-menu", "hud", "stairs", "chests", "vendors", "monsters", "companions", "heal-rain", "town", "skeleton"], default="gear")
```

Then add the size override block in the size-override section (after the existing `if args.focus == "town"` block):

```python
    if args.focus == "skeleton" and (args.width, args.height) == (640, 480):
        width, height = 800, 600
```

- [ ] **Step 3: Run the preload smoke test to confirm wiring**

```bash
godot --headless --path /Users/mmandrille/git/arpg-dev/client --script res://tests/test_showme_load.gd 2>&1 | tail -5
```

Expected: `[gdtest] PASS: showme capture scripts preload`

- [ ] **Step 4: Run the Python layout tests**

```bash
cd /Users/mmandrille/git/arpg-dev && .venv/bin/pytest tools/test_showme.py -v 2>&1 | tail -10
```

Expected: all tests PASS.

- [ ] **Step 5: Capture the first screenshot and inspect it**

```bash
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id paladin
```

The command prints the output path. Open the image and verify:
- Paladin model is visible in a spread pose
- Red dots are present at bone positions
- Colored labeled spheres appear at socket positions

If the arm spread is wrong direction (arms going down instead of out), adjust `SPREAD_BONES` axes in `showme_skeleton_capture.gd`:
- `arm_r` and `arm_l` use Z-axis rotation by default; if they spread the wrong way, try X-axis: `Vector3(1.0, 0.0, 0.0)` / `Vector3(-1.0, 0.0, 0.0)`.
- Likewise for `leg_r` / `leg_l` if they don't visibly separate.

Re-run the capture after any adjustment until the pose is correct.

- [ ] **Step 6: Check maintainability ratchet**

```bash
cd /Users/mmandrille/git/arpg-dev && make maintainability 2>&1 | tail -10
```

Expected: passes with no new violations. `visual_capture.gd` must remain ≤ 1067 lines.

- [ ] **Step 7: Commit**

```bash
git add client/scripts/showme/visual_capture.gd skills/showme/scripts/render_focus.py
git commit -m "feat: wire skeleton focus into showme — visual_capture + render_focus"
```

---

### Task 3: Update SKILL.md catalog and add test

**Files:**
- Modify: `.claude/skills/showme/SKILL.md`
- Modify: `tools/test_showme.py`

- [ ] **Step 1: Add skeleton row to SKILL.md**

In `.claude/skills/showme/SKILL.md`, under the **Equipment and models** table, add a row after the `classes` row:

```markdown
| `skeleton` | Character in spread T-pose; red dot at every bone + colored labeled sphere at each equipment socket. | `--class-id` | 800×600 |
```

Also add a row to the **Quick picker** table:

```markdown
| Skeleton bone positions + socket markers | `skeleton` |
```

Also add to the **Common examples** section:

```bash
# Skeleton bone visualization (default class: paladin).
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id paladin

# Check another class skeleton.
python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id barbarian
```

- [ ] **Step 2: Add test to `tools/test_showme.py`**

Add to `tools/test_showme.py`:

```python
def test_skeleton_focus_in_render_focus_choices() -> None:
    script = (ROOT / "skills" / "showme" / "scripts" / "render_focus.py").read_text(encoding="utf-8")
    assert '"skeleton"' in script, "skeleton focus missing from render_focus.py choices"
    assert "showme_skeleton_capture.gd" in (ROOT / "client" / "scripts" / "showme" / "visual_capture.gd").read_text(encoding="utf-8")
```

- [ ] **Step 3: Run the Python tests to confirm the new test passes**

```bash
cd /Users/mmandrille/git/arpg-dev && .venv/bin/pytest tools/test_showme.py -v 2>&1 | tail -10
```

Expected: all 3 tests PASS.

- [ ] **Step 4: Run CI to confirm nothing is broken**

```bash
cd /Users/mmandrille/git/arpg-dev && make ci 2>&1 | tail -20
```

Expected: green.

- [ ] **Step 5: Commit**

```bash
git add .claude/skills/showme/SKILL.md tools/test_showme.py
git commit -m "docs: register skeleton focus in showme SKILL.md + test coverage"
```
