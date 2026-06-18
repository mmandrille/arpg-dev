#!/usr/bin/env python3
"""Create a disposable Godot sandbox scene for an imported GLB model.

Run from the arpg-dev repo root:
  python3 skills/3dmodel/scripts/create_model_probe.py --model path/to/model.glb --key wolf --yaw-degrees -90
"""
from __future__ import annotations

import argparse
import json
import math
import os
import re
import shutil
import struct
import subprocess
from pathlib import Path


def sanitize_key(value: str) -> str:
    key = re.sub(r"[^a-z0-9_]+", "_", value.lower()).strip("_")
    return key or "model"


def parse_glb_json(path: Path) -> dict:
    data = path.read_bytes()
    if len(data) < 20 or data[0:4] != b"glTF":
        raise ValueError(f"{path} is not a GLB file")
    offset = 12
    while offset + 8 <= len(data):
        chunk_len, chunk_type = struct.unpack_from("<II", data, offset)
        offset += 8
        chunk = data[offset : offset + chunk_len]
        offset += chunk_len
        if chunk_type == 0x4E4F534A:  # JSON
            return json.loads(chunk.decode("utf-8"))
    raise ValueError(f"{path} has no GLB JSON chunk")


def model_report(gltf: dict) -> str:
    nodes = gltf.get("nodes", [])
    skins = gltf.get("skins", [])
    animations = gltf.get("animations", [])
    mins: list[float] | None = None
    maxs: list[float] | None = None
    for mesh in gltf.get("meshes", []):
        for prim in mesh.get("primitives", []):
            accessor_index = prim.get("attributes", {}).get("POSITION")
            if accessor_index is None:
                continue
            accessor = gltf.get("accessors", [])[accessor_index]
            if "min" not in accessor or "max" not in accessor:
                continue
            if mins is None:
                mins = list(accessor["min"])
                maxs = list(accessor["max"])
            else:
                mins = [min(a, b) for a, b in zip(mins, accessor["min"])]
                maxs = [max(a, b) for a, b in zip(maxs or [], accessor["max"])]
    extents = []
    if mins is not None and maxs is not None:
        extents = [round(maxs[i] - mins[i], 4) for i in range(3)]
    joint_names: list[str] = []
    for skin in skins:
        for idx in skin.get("joints", []):
            if 0 <= idx < len(nodes):
                joint_names.append(str(nodes[idx].get("name", idx)))
    lines = [
        f"nodes={len(nodes)} skins={len(skins)} animations={len(animations)}",
        f"node_names={[str(n.get('name', i)) for i, n in enumerate(nodes[:20])]}",
        f"skin_joints={joint_names[:30]}",
        f"bounds_min={mins} bounds_max={maxs} extents_xyz={extents}",
    ]
    if extents:
        axis = ["x", "y", "z"][max(range(3), key=lambda i: extents[i])]
        lines.append(f"largest_extent_axis={axis} (use yaw checks to find visual nose/front)")
    if not skins:
        lines.append("note=no skin: use node-root transform animation unless replacing with a rigged model")
    if not animations:
        lines.append("note=no embedded animations: generate presentation clips in Godot")
    return "\n".join(lines)


def write_builder(path: Path, out_anim: str, target_node_path: str) -> None:
    path.write_text(
        f'''extends SceneTree

const DEG := PI / 180.0

func _initialize() -> void:
	var lib := AnimationLibrary.new()
	for clip_name in _clips():
		lib.add_animation(clip_name, _make_anim("{target_node_path}", _clips()[clip_name]))
	var err := ResourceSaver.save(lib, "{out_anim}")
	if err != OK:
		printerr("[model-probe] save failed: %s" % err)
		quit(1)
		return
	print("[model-probe] wrote {out_anim}")
	quit(0)

func _make_anim(node_path: String, spec: Dictionary) -> Animation:
	var a := Animation.new()
	a.length = spec.get("length", 1.0)
	a.loop_mode = Animation.LOOP_LINEAR if spec.get("loop", false) else Animation.LOOP_NONE
	if spec.has("positions"):
		var pi := a.add_track(Animation.TYPE_POSITION_3D)
		a.track_set_path(pi, NodePath(node_path))
		for key in spec["positions"]:
			a.position_track_insert_key(pi, key[0], Vector3(key[1], key[2], key[3]))
	if spec.has("rotations"):
		var ri := a.add_track(Animation.TYPE_ROTATION_3D)
		a.track_set_path(ri, NodePath(node_path))
		for key in spec["rotations"]:
			var q := Quaternion.from_euler(Vector3(key[1] * DEG, key[2] * DEG, key[3] * DEG))
			a.rotation_track_insert_key(ri, key[0], q)
	return a

func _clips() -> Dictionary:
	return {{
		"idle": {{"length": 1.0, "loop": true, "positions": [[0.0, 0.0, 0.0, 0.0], [0.5, 0.0, 0.025, 0.0], [1.0, 0.0, 0.0, 0.0]]}},
		"walk": {{"length": 0.55, "loop": true, "positions": [[0.0, 0.0, 0.0, 0.0], [0.1375, 0.0, 0.055, 0.0], [0.275, 0.0, 0.0, 0.0], [0.4125, 0.0, 0.045, 0.0], [0.55, 0.0, 0.0, 0.0]], "rotations": [[0.0, 0.0, 0.0, -4.0], [0.1375, 0.0, 0.0, 4.0], [0.275, 0.0, 0.0, -4.0], [0.4125, 0.0, 0.0, 4.0], [0.55, 0.0, 0.0, -4.0]]}},
		"attack": {{"length": 0.35, "loop": false, "positions": [[0.0, 0.0, 0.0, 0.0], [0.12, 0.0, 0.0, 0.12], [0.35, 0.0, 0.0, 0.0]], "rotations": [[0.0, 0.0, 0.0, 0.0], [0.12, -8.0, 0.0, 0.0], [0.35, 0.0, 0.0, 0.0]]}},
		"death": {{"length": 0.5, "loop": false, "rotations": [[0.0, 0.0, 0.0, 0.0], [0.5, 0.0, 0.0, 82.0]]}},
	}}
''',
        encoding="utf-8",
    )


def write_test(path: Path, scene_res: str, expected_yaw: float) -> None:
    path.write_text(
        f'''extends SceneTree

func _initialize() -> void:
	var packed := load("{scene_res}") as PackedScene
	if packed == null:
		_fail("missing scene")
		return
	var scene := packed.instantiate() as Node3D
	get_root().add_child(scene)
	await process_frame
	var ap := scene.find_child("AnimationPlayer", true, false) as AnimationPlayer
	var root := scene.find_child("ModelRoot", false, false) as Node3D
	var model := scene.find_child("Model", true, false) as Node3D
	if ap == null or root == null or model == null:
		_fail("missing AnimationPlayer/ModelRoot/Model")
		return
	for clip in ["idle", "walk", "attack", "death"]:
		if not ap.has_animation(clip):
			_fail("missing clip " + clip)
			return
	if absf(root.rotation.y - {expected_yaw:.8f}) > 0.001:
		_fail("yaw correction changed before animation")
		return
	ap.play("walk")
	ap.seek(0.1375, true)
	if absf(root.rotation.y - {expected_yaw:.8f}) > 0.001:
		_fail("walk overwrote ModelRoot yaw correction")
		return
	if model.position.y <= 0.0:
		_fail("walk did not bob child model")
		return
	ap.play("attack")
	ap.seek(0.12, true)
	if model.position.z <= 0.0:
		_fail("attack did not lunge child model")
		return
	print("[model-probe] PASS {scene_res}")
	quit(0)

func _fail(msg: String) -> void:
	printerr("[model-probe] FAIL: " + msg)
	quit(1)
''',
        encoding="utf-8",
    )


def run(cmd: list[str], cwd: Path) -> None:
    subprocess.run(cmd, cwd=cwd, check=True)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", required=True, help="Source .glb path")
    parser.add_argument("--key", help="Probe key; defaults to model stem")
    parser.add_argument("--yaw-degrees", type=float, default=0.0, help="ModelRoot yaw correction to test")
    parser.add_argument("--repo", default=".", help="Repo root")
    parser.add_argument("--no-run", action="store_true", help="Write files but skip Godot import/test")
    args = parser.parse_args()

    repo = Path(args.repo).resolve()
    model = Path(args.model).resolve()
    key = sanitize_key(args.key or model.stem)
    yaw = math.radians(args.yaw_degrees)

    gltf = parse_glb_json(model)
    print(model_report(gltf))

    rel_asset = Path("assets/_model_probe") / key / f"{key}.glb"
    asset_path = repo / "client" / rel_asset
    asset_path.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(model, asset_path)

    scene_name = f"model_probe_{key}.tscn"
    anim_name = f"model_probe_{key}_anims.tres"
    builder_name = f"model_probe_{key}_build.gd"
    test_name = f"model_probe_{key}_test.gd"
    scene_path = repo / "client/scenes" / scene_name
    anim_res = f"res://animations/{anim_name}"
    builder_path = repo / "client/tools" / builder_name
    test_path = repo / "client/tools" / test_name

    write_builder(builder_path, anim_res, "ModelRoot/Model")
    scene_path.write_text(
        f'''[gd_scene load_steps=3 format=3]

[ext_resource type="PackedScene" path="res://{rel_asset.as_posix()}" id="1_glb"]
[ext_resource type="AnimationLibrary" path="{anim_res}" id="2_anims"]

[node name="ModelProbe{key.title().replace("_", "")}" type="Node3D"]

[node name="ModelRoot" type="Node3D" parent="."]
rotation = Vector3(0, {yaw:.8f}, 0)

[node name="Model" parent="ModelRoot" instance=ExtResource("1_glb")]

[node name="AnimationPlayer" type="AnimationPlayer" parent="."]
root_node = NodePath("..")
libraries = {{
"": ExtResource("2_anims")
}}
''',
        encoding="utf-8",
    )
    write_test(test_path, f"res://scenes/{scene_name}", yaw)

    godot = os.environ.get("GODOT") or shutil.which("godot") or shutil.which("godot4")
    if args.no_run:
        godot = None
    if godot:
        run([godot, "--headless", "--path", "client", "--import"], repo)
        run([godot, "--headless", "--path", "client", "--script", f"res://tools/{builder_name}"], repo)
        run([godot, "--headless", "--path", "client", "--import"], repo)
        run([godot, "--headless", "--path", "client", "--script", f"res://tools/{test_name}"], repo)
    else:
        print("[model-probe] Godot not found or --no-run set; wrote files without running test")

    print(f"[model-probe] asset={asset_path.relative_to(repo)}")
    print(f"[model-probe] scene={scene_path.relative_to(repo)}")
    print(f"[model-probe] animation=client/animations/{anim_name}")
    print(f"[model-probe] yaw_degrees={args.yaw_degrees}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
