extends SceneTree
# Headless tests for the v3 animation layer (spec §10). Server-independent.
# Run: godot --headless --path client --script res://tests/test_animation.gd
const ControllerScript := preload("res://scripts/animation_controller.gd")
const ReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const MonsterVisualsLoaderScript := preload("res://scripts/monster_visuals_loader.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")


var _failed: bool = false


func _initialize() -> void:
	_test_controller_locomotion()
	if _failed: quit(1); return
	_test_controller_one_shot_returns()
	if _failed: quit(1); return
	_test_controller_terminal_latches()
	if _failed: quit(1); return
	_test_controller_terminal_reset_restores_locomotion()
	if _failed: quit(1); return
	_test_controller_hit_ignored_after_death()
	if _failed: quit(1); return
	_test_controller_locomotion_change_during_one_shot()
	if _failed: quit(1); return
	await _test_model_reaction_hit_restores()
	if _failed: quit(1); return
	await _test_model_reaction_death_terminal()
	if _failed: quit(1); return
	await _test_model_reaction_terminal_reset_restores_model()
	if _failed: quit(1); return
	# Scene tests exercise runtime _ready (socket attach). add_child() inside
	# _initialize() does NOT enter the tree synchronously in Godot 4.6, so we
	# await a frame to let _ready fire before asserting. This makes _initialize
	# a coroutine; quit() still sets the exit code correctly when it resumes.
	await _test_character_scene()
	if _failed: quit(1); return
	await _test_class_character_models()
	if _failed: quit(1); return
	await _test_player_snapshot_death_pose()
	if _failed: quit(1); return
	await _test_monster_scene()
	if _failed: quit(1); return
	_test_monster_visuals_catalog()
	if _failed: quit(1); return
	print("[gdtest] PASS: animation controller + scenes")
	quit(0)


func _make_player(clips: Array) -> AnimationPlayer:
	var ap := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	for c in clips:
		var a := Animation.new()
		a.length = (0.5 if c != "idle" and c != "walk" else 1.0)
		lib.add_animation(c, a)
	ap.add_animation_library("", lib)
	get_root().add_child(ap)  # animations need the player in-tree
	return ap


func _test_controller_locomotion() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(false)
	_assert(c.current_clip() == "idle", "idle when not moving, got %s" % c.current_clip())
	c.set_locomotion(true)
	_assert(c.current_clip() == "walk", "walk when moving, got %s" % c.current_clip())
	ap.free()


func _test_controller_one_shot_returns() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(true)
	c.play_one_shot("attack")
	_assert(c.current_clip() == "attack", "attack active, got %s" % c.current_clip())
	# Simulate the clip finishing.
	ap.emit_signal("animation_finished", "attack")
	_assert(c.current_clip() == "walk", "returns to locomotion (walk) after one-shot, got %s" % c.current_clip())
	ap.free()


func _test_controller_terminal_latches() -> void:
	var ap := _make_player(["idle", "hit", "death"])
	var c = ControllerScript.new(ap)
	c.enter_terminal("death")
	_assert(c.current_clip() == "death", "death active, got %s" % c.current_clip())
	c.play_one_shot("hit")        # ignored
	c.set_locomotion(true)        # ignored
	_assert(c.current_clip() == "death", "terminal latched, got %s" % c.current_clip())
	_assert(c.get_debug_state()["terminal"] == true, "terminal flag set")
	ap.free()


func _test_controller_terminal_reset_restores_locomotion() -> void:
	var ap := _make_player(["idle", "walk", "death"])
	var c = ControllerScript.new(ap)
	c.enter_terminal("death")
	c.reset_terminal()
	_assert(c.get_debug_state()["terminal"] == false, "terminal reset clears flag")
	_assert(c.current_clip() == "idle", "terminal reset restores idle, got %s" % c.current_clip())
	c.set_locomotion(true)
	_assert(c.current_clip() == "walk", "locomotion works after terminal reset, got %s" % c.current_clip())
	ap.free()


func _test_controller_hit_ignored_after_death() -> void:
	var ap := _make_player(["idle", "hit", "death"])
	var c = ControllerScript.new(ap)
	c.enter_terminal("death")
	c.play_one_shot("hit")
	_assert(c.current_clip() == "death", "hit ignored after terminal death, got %s" % c.current_clip())
	ap.free()


func _test_controller_locomotion_change_during_one_shot() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(false)
	c.play_one_shot("attack")
	c.set_locomotion(true)  # must NOT switch to walk while one-shot is active
	_assert(c.current_clip() == "attack", "locomotion change ignored during one-shot, got %s" % c.current_clip())
	ap.emit_signal("animation_finished", "attack")
	_assert(c.current_clip() == "walk", "fallback honors latched _moving after one-shot, got %s" % c.current_clip())
	ap.free()


func _make_reaction_root(color: Color = Color("#8fe8a7")) -> Node3D:
	var root := Node3D.new()
	var mesh_node := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3.ONE
	mesh_node.mesh = mesh
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mesh_node.material_override = mat
	root.add_child(mesh_node)
	get_root().add_child(root)
	return root


func _reaction_mesh_color(root: Node3D) -> Color:
	var mesh_node := root.get_child(0) as MeshInstance3D
	var mat := mesh_node.material_override as StandardMaterial3D
	return mat.albedo_color


func _test_model_reaction_hit_restores() -> void:
	var root := _make_reaction_root(Color("#8fe8a7"))
	await process_frame
	var c = ReactionControllerScript.new(root, Color("#8fe8a7"))
	c.play_hit(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	await create_timer(0.50).timeout
	var state: Dictionary = c.get_debug_state()
	_assert(state["terminal"] == false, "hit reaction should not be terminal")
	_assert(str(state["last_reaction"]) == "hit", "last reaction should be hit")
	_assert(root.rotation.distance_to(Vector3.ZERO) <= 0.001, "hit reaction restores rotation, got %s" % root.rotation)
	var color := _reaction_mesh_color(root)
	_assert(color.is_equal_approx(Color("#8fe8a7")), "hit reaction restores color, got %s" % color)
	root.free()
	await process_frame


func _test_model_reaction_death_terminal() -> void:
	var root := _make_reaction_root(Color("#8fe8a7"))
	await process_frame
	var c = ReactionControllerScript.new(root, Color("#8fe8a7"))
	c.play_hit(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	c.enter_death(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	c.play_hit(Vector3(-1.0, 0.0, 0.0), Vector3.FORWARD)
	await create_timer(0.50).timeout
	var state: Dictionary = c.get_debug_state()
	_assert(state["terminal"] == true, "death reaction should be terminal")
	_assert(str(state["last_reaction"]) == "death", "hit after death ignored")
	_assert(absf(root.rotation.x) > 0.1 or absf(root.rotation.z) > 0.1, "death reaction rotates model down, got %s" % root.rotation)
	var color := _reaction_mesh_color(root)
	_assert(color.r < Color("#8fe8a7").r and color.g < Color("#8fe8a7").g, "death reaction darkens color, got %s" % color)
	root.free()
	await process_frame


func _test_model_reaction_terminal_reset_restores_model() -> void:
	var base_color := Color("#8fe8a7")
	var root := _make_reaction_root(base_color)
	await process_frame
	var c = ReactionControllerScript.new(root, base_color)
	c.enter_death(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	await create_timer(0.25).timeout
	c.reset_terminal()
	var state: Dictionary = c.get_debug_state()
	_assert(state["terminal"] == false, "reaction terminal reset clears flag")
	_assert(str(state["last_reaction"]) == "", "reaction terminal reset clears last reaction")
	_assert(root.rotation.distance_to(Vector3.ZERO) <= 0.001, "reaction terminal reset restores rotation, got %s" % root.rotation)
	var color := _reaction_mesh_color(root)
	_assert(color.is_equal_approx(base_color), "reaction terminal reset restores color, got %s" % color)
	c.play_hit(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	_assert(str(c.get_debug_state()["last_reaction"]) == "hit", "hit works after reaction terminal reset")
	root.free()
	await process_frame


func _test_character_scene() -> void:
	var s = (load("res://scenes/character.tscn") as PackedScene).instantiate()
	get_root().add_child(s)  # _ready attaches the socket
	await process_frame      # node enters tree + _ready fires next frame
	var sock = s.find_child("right_hand_socket", true, false)
	_assert(sock is BoneAttachment3D, "right_hand_socket must be a BoneAttachment3D")
	if sock is BoneAttachment3D:
		_assert(sock.bone_name == "hand_r", "socket bound to hand_r, got %s" % sock.bone_name)
	var off_sock = s.find_child("off_hand_socket", true, false)
	_assert(off_sock is BoneAttachment3D, "off_hand_socket must be a BoneAttachment3D")
	if off_sock is BoneAttachment3D:
		_assert(off_sock.bone_name == "hand_l", "off_hand_socket bound to hand_l, got %s" % off_sock.bone_name)
	var ap := s.find_child("AnimationPlayer", true, false) as AnimationPlayer
	_assert(ap != null, "character AnimationPlayer missing")
	if ap != null:
		for clip in ["idle", "walk", "attack", "attack_off_hand", "hit", "death"]:
			_assert(ap.has_animation(clip), "character missing clip %s" % clip)
	s.free()
	await process_frame


func _test_class_character_models() -> void:
	for class_id in ["barbarian", "sorcerer", "paladin", "rogue"]:
		var resolved := ClassPresentationsLoaderScript.resolve(class_id)
		_assert(str(resolved.get("asset_id", "")) == "character_%s_v0" % class_id, "%s model asset mismatch: %s" % [class_id, resolved])
		var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
		_assert(packed != null, "%s model packed scene missing" % class_id)
		if packed == null:
			continue
		var model := packed.instantiate() as Node3D
		get_root().add_child(model)
		await process_frame
		var skel := model.find_child("Skeleton3D", true, false) as Skeleton3D
		_assert(skel != null, "%s model missing Skeleton3D" % class_id)
		if skel != null:
			for bone in ["root", "spine", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"]:
				_assert(skel.find_bone(bone) >= 0, "%s model missing bone %s" % [class_id, bone])
		model.free()
		await process_frame
	var fallback := ClassPresentationsLoaderScript.resolve("necromancer")
	_assert(str(fallback.get("asset_id", "")) == "character_base_humanoid_v0", "unknown class should use base humanoid fallback: %s" % fallback)


func _test_player_snapshot_death_pose() -> void:
	var s = (load("res://scenes/character.tscn") as PackedScene).instantiate()
	get_root().add_child(s)
	await process_frame
	var ap := s.find_child("AnimationPlayer", true, false) as AnimationPlayer
	_assert(ap != null, "character AnimationPlayer missing for player snapshot death")
	if ap != null:
		var player_controller = ControllerScript.new(ap)
		var e := {"id": "1001", "type": "player", "position": {"x": 10, "y": 5}, "hp": 0, "max_hp": 10}
		if str(e.get("type", "")) == "player" and int(e.get("hp", 1)) <= 0:
			player_controller.enter_terminal("death")
		_assert(player_controller.get_debug_state()["terminal"] == true, "player snapshot hp<=0 enters terminal death")
		_assert(player_controller.current_clip() == "death", "player snapshot death clip active, got %s" % player_controller.current_clip())
	s.free()
	await process_frame


func _test_monster_scene() -> void:
	for scene_path in [
		"res://scenes/monster_dummy.tscn",
		"res://scenes/monster_dark_purple.tscn",
		"res://scenes/monster_crocodile_archer.tscn",
		"res://scenes/monster_quadruped.tscn",
		"res://scenes/monster_wolf.tscn",
		"res://scenes/monster_tiny_flyer.tscn",
		"res://scenes/monster_skeleton.tscn",
	]:
		var s = (load(scene_path) as PackedScene).instantiate()
		get_root().add_child(s)
		await process_frame
		var ap := s.find_child("AnimationPlayer", true, false) as AnimationPlayer
		_assert(ap != null, "%s AnimationPlayer missing" % scene_path)
		if ap != null:
			for clip in ["idle", "walk", "hit", "death"]:
				_assert(ap.has_animation(clip), "%s missing clip %s" % [scene_path, clip])
			if scene_path == "res://scenes/monster_wolf.tscn":
				var model_root := s.find_child("ModelRoot", false, false) as Node3D
				_assert(model_root != null, "%s ModelRoot missing" % scene_path)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s ModelRoot should correct GLB nose to parent +Z, got y=%s" % [scene_path, model_root.rotation.y])
				var wolf_model := s.find_child("WolfModel", true, false) as Node3D
				_assert(wolf_model != null, "%s WolfModel missing" % scene_path)
				ap.play("walk")
				ap.seek(0.1375, true)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s walk clip must preserve ModelRoot yaw correction, got y=%s" % [scene_path, model_root.rotation.y])
				_assert(wolf_model.position.y > 0.0, "%s walk clip should bob WolfModel, got y=%s" % [scene_path, wolf_model.position.y])
			if scene_path == "res://scenes/monster_quadruped.tscn":
				var model_root := s.find_child("ModelRoot", false, false) as Node3D
				_assert(model_root != null, "%s ModelRoot missing" % scene_path)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s ModelRoot should correct GLB nose to parent +Z, got y=%s" % [scene_path, model_root.rotation.y])
				var quadruped_model := s.find_child("QuadrupedModel", true, false) as Node3D
				_assert(quadruped_model != null, "%s QuadrupedModel missing" % scene_path)
				ap.play("walk")
				ap.seek(0.1375, true)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s walk clip must preserve ModelRoot yaw correction, got y=%s" % [scene_path, model_root.rotation.y])
				_assert(quadruped_model.position.y > 0.0, "%s walk clip should bob QuadrupedModel, got y=%s" % [scene_path, quadruped_model.position.y])
			if scene_path == "res://scenes/monster_dark_purple.tscn":
				var model_root := s.find_child("ModelRoot", false, false) as Node3D
				_assert(model_root != null, "%s ModelRoot missing" % scene_path)
				_assert(absf(model_root.rotation.y) <= 0.001, "%s ModelRoot should face parent +Z after right-facing correction, got y=%s" % [scene_path, model_root.rotation.y])
				var model := s.find_child("Model", true, false) as Node3D
				_assert(model != null, "%s Model missing" % scene_path)
				_assert(model.scale.is_equal_approx(Vector3(0.75, 0.75, 0.75)), "%s Model should be scaled to 75%%, got %s" % [scene_path, model.scale])
				ap.play("walk")
				ap.seek(0.1375, true)
				_assert(absf(model_root.rotation.y) <= 0.001, "%s walk clip must preserve ModelRoot yaw correction, got y=%s" % [scene_path, model_root.rotation.y])
				_assert(model.position.y > 0.0, "%s walk clip should bob Model, got y=%s" % [scene_path, model.position.y])
			if scene_path == "res://scenes/monster_crocodile_archer.tscn":
				var marker := s.find_child("ArcherBowMarker", true, false) as Node3D
				_assert(marker != null, "%s should expose the ranged archer marker" % scene_path)
				var model_root := s.find_child("ModelRoot", false, false) as Node3D
				_assert(model_root != null, "%s ModelRoot missing" % scene_path)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s ModelRoot should correct GLB left-facing axis to parent +Z, got y=%s" % [scene_path, model_root.rotation.y])
				var model := s.find_child("Model", true, false) as Node3D
				_assert(model != null, "%s Model missing" % scene_path)
				ap.play("walk")
				ap.seek(0.1375, true)
				_assert(absf(model_root.rotation.y + PI * 0.5) <= 0.001, "%s walk clip must preserve ModelRoot yaw correction, got y=%s" % [scene_path, model_root.rotation.y])
				_assert(model.position.y > 0.0, "%s walk clip should bob Model, got y=%s" % [scene_path, model.position.y])
		s.free()
		await process_frame


func _test_monster_visuals_catalog() -> void:
	var mob := MonsterVisualsLoaderScript.resolve("dungeon_mob")
	_assert(str(mob.get("scene", "")) == "monster_dark_purple", "dungeon_mob scene = %s" % mob.get("scene", ""))
	_assert(str(mob.get("asset_id", "")) == "monster_dark_purple_v0", "dungeon_mob asset = %s" % mob.get("asset_id", ""))
	var wolf := MonsterVisualsLoaderScript.resolve("dungeon_wolf")
	_assert(str(wolf.get("scene", "")) == "monster_quadruped", "dungeon_wolf scene = %s" % wolf.get("scene", ""))
	_assert(str(wolf.get("asset_id", "")) == "monster_quadruped_predator_v0", "dungeon_wolf asset = %s" % wolf.get("asset_id", ""))
	var archer := MonsterVisualsLoaderScript.resolve("dungeon_archer")
	_assert(str(archer.get("scene", "")) == "monster_crocodile_archer", "dungeon_archer scene = %s" % archer.get("scene", ""))
	_assert(str(archer.get("asset_id", "")) == "monster_crocodile_archer_v0", "dungeon_archer asset = %s" % archer.get("asset_id", ""))
	var companion_wolf := MonsterVisualsLoaderScript.resolve("companion_black_wolf", "monster_wolf")
	_assert(str(companion_wolf.get("scene", "")) == "monster_wolf", "companion_black_wolf scene = %s" % companion_wolf.get("scene", ""))
	var bat := MonsterVisualsLoaderScript.resolve("dungeon_bat")
	_assert(str(bat.get("scene", "")) == "monster_tiny_flyer", "dungeon_bat scene = %s" % bat.get("scene", ""))
	_assert(float(bat.get("height_offset", 0.0)) > 0.0, "dungeon_bat must hover above ground")
	var undead := MonsterVisualsLoaderScript.resolve("dungeon_undead")
	_assert(str(undead.get("scene", "")) == "monster_skeleton", "dungeon_undead scene = %s" % undead.get("scene", ""))
	var boss := MonsterVisualsLoaderScript.resolve("dungeon_mob", "monster_tiny_flyer")
	_assert(str(boss.get("scene", "")) == "monster_tiny_flyer", "boss visual_model should select flyer scene")


func _assert(cond: bool, msg: String) -> void:
	if not cond:
		printerr("[gdtest] FAIL: ", msg)
		_failed = true
		quit(1)
