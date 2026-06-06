extends SceneTree
# Source-of-truth for the committed AnimationLibrary .tres clips (spec §4.4).
# Builds rotation_3d bone-pose tracks against the imported skeleton and saves a
# library next to client/animations/. Run via `make gen-anims`. Deterministic
# for a pinned Godot. Clip motion is crude on purpose (art is a non-goal).
const DEG := PI / 180.0

func _initialize() -> void:
	_build(
		"res://assets/characters/base_humanoid/base_humanoid.glb",
		"res://animations/character_anims.tres",
		_character_clips())
	_build(
		"res://assets/monsters/dummy/monster_dummy.glb",
		"res://animations/monster_anims.tres",
		_monster_clips())
	print("[build-anims] PASS")
	quit(0)

func _build(glb_path: String, out_path: String, clips: Dictionary) -> void:
	var packed = load(glb_path)
	if packed == null:
		_fail("cannot load %s" % glb_path)
		return
	var scene = (packed as PackedScene).instantiate()
	var skel := scene.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		_fail("no Skeleton3D in %s" % glb_path)
		return
	var skel_path := str(scene.get_path_to(skel))  # e.g. "Skeleton3D"
	var lib := AnimationLibrary.new()
	for clip_name in clips:
		lib.add_animation(clip_name, _make_anim(skel_path, clips[clip_name]))
	var err := ResourceSaver.save(lib, out_path)
	if err != OK:
		_fail("save %s failed: %d" % [out_path, err])
		return
	print("[build-anims] wrote %s (skeleton=%s)" % [out_path, skel_path])

func _make_anim(skel_path: String, spec: Dictionary) -> Animation:
	# spec: { "length": float, "loop": bool, "bones": { bone: [[t, x,y,z(deg)], ...] } }
	var a := Animation.new()
	a.length = spec.get("length", 1.0)
	a.loop_mode = Animation.LOOP_LINEAR if spec.get("loop", false) else Animation.LOOP_NONE
	for bone in spec["bones"]:
		var ti := a.add_track(Animation.TYPE_ROTATION_3D)
		a.track_set_path(ti, NodePath("%s:%s" % [skel_path, bone]))
		for key in spec["bones"][bone]:
			var t: float = key[0]
			var q := Quaternion.from_euler(Vector3(key[1] * DEG, key[2] * DEG, key[3] * DEG))
			a.rotation_track_insert_key(ti, t, q)
	return a

func _character_clips() -> Dictionary:
	return {
		# idle: a near-still pose (one identity key so the track exists).
		"idle": {"length": 1.0, "loop": true, "bones": {"spine": [[0.0, 0, 0, 0]]}},
		# walk: alternate the legs back/forth.
		"walk": {"length": 0.8, "loop": true, "bones": {
			"leg_l": [[0.0, 25, 0, 0], [0.4, -25, 0, 0], [0.8, 25, 0, 0]],
			"leg_r": [[0.0, -25, 0, 0], [0.4, 25, 0, 0], [0.8, -25, 0, 0]],
		}},
		# attack: swing the right arm down and back (the weapon rides hand_r).
		"attack": {"length": 0.35, "loop": false, "bones": {
			"arm_r": [[0.0, 0, 0, 0], [0.12, -110, 0, 0], [0.35, 0, 0, 0]],
		}},
		# hit: brief backward wobble on authoritative player damage.
		"hit": {"length": 0.25, "loop": false, "bones": {
			"spine": [[0.0, 0, 0, 0], [0.08, -14, 0, 0], [0.25, 0, 0, 0]],
		}},
		# death: terminal topple pose held at the clip end.
		"death": {"length": 0.6, "loop": false, "bones": {
			"spine": [[0.0, 0, 0, 0], [0.6, -72, 0, 0]],
		}},
	}

func _monster_clips() -> Dictionary:
	return {
		"idle": {"length": 1.0, "loop": true, "bones": {"pivot": [[0.0, 0, 0, 0]]}},
		"walk": {"length": 0.8, "loop": true, "bones": {
			"pivot": [[0.0, 0, 0, 0], [0.4, 0, 0, 8], [0.8, 0, 0, 0]],
		}},
		# hit: a quick wobble about the base.
		"hit": {"length": 0.3, "loop": false, "bones": {
			"pivot": [[0.0, 0, 0, 0], [0.1, 0, 0, 18], [0.3, 0, 0, 0]],
		}},
		# death: topple over and hold (terminal pose).
		"death": {"length": 0.6, "loop": false, "bones": {
			"pivot": [[0.0, 0, 0, 0], [0.6, 0, 0, 88]],
		}},
	}

func _fail(msg: String) -> void:
	printerr("[build-anims] FAIL: ", msg)
	quit(1)
