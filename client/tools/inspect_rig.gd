extends SceneTree
# Rig gate (spec §10): confirm both rigged GLBs import as real skinned scenes.
func _initialize() -> void:
	_check("res://assets/characters/base_humanoid/base_humanoid.glb", ["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"])
	_check("res://assets/monsters/dummy/monster_dummy.glb", ["root", "pivot"])
	print("[rig-gate] PASS")
	quit(0)

func _check(path: String, expected_bones: Array) -> void:
	var packed = load(path)
	if packed == null:
		_fail("cannot load %s" % path)
		return
	var scene = (packed as PackedScene).instantiate()
	var skel := scene.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		_fail("no Skeleton3D in %s (skin import failed)" % path)
		return
	var names := []
	for i in range(skel.get_bone_count()):
		names.append(skel.get_bone_name(i))
	for b in expected_bones:
		if not names.has(b):
			_fail("%s missing bone %s (have %s)" % [path, b, names])
			return
	if scene.find_child("*", true, false) == null:
		_fail("%s has no descendant nodes (trivial tree)" % path)
		return
	print("[rig-gate] %s bones=%s skeleton_path=%s" % [path, names, scene.get_path_to(skel)])

func _fail(msg: String) -> void:
	printerr("[rig-gate] FAIL: ", msg)
	quit(1)
