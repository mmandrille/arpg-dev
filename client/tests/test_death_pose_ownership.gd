extends SceneTree
# Death-pose ownership: a visual whose AnimationPlayer has a `death` clip lays itself down, so
# ModelReactionController must not also lean the entity root (that stood corpses on end). Only
# clip-less visuals (training doll silhouette) get the root lean. The darken applies to both.
# Run: godot --headless --path client --script res://tests/test_death_pose_ownership.gd
const ReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const TrainingDamageLogBridgeScript := preload("res://scripts/training_damage_log_bridge.gd")
const MainScript := preload("res://scripts/main.gd")

const SETTLE_SECONDS := 0.5
const BASE_TINT := Color("#8fe8a7")

var _failed := false


func _initialize() -> void:
	await _test_clip_owned_death_keeps_root_upright("kit skeleton", _scene_root("monster_kit_skeleton_warrior"))
	await _test_clip_owned_death_keeps_root_upright("legacy dummy", _scene_root("monster_dummy"))
	await _test_clip_owned_death_keeps_root_upright("kit hero", _kit_hero_root("paladin"))
	await _test_hit_then_death_restores_root_for_clip_owned()
	await _test_clipless_death_still_leans()
	if _failed:
		quit(1)
		return
	print("[gdtest] PASS: test_death_pose_ownership")
	quit(0)


func _scene_root(scene_key: String) -> Node3D:
	var root := Node3D.new()
	root.name = "MonsterVisualRoot"
	root.add_child((load("res://scenes/%s.tscn" % scene_key) as PackedScene).instantiate())
	return root


func _kit_hero_root(class_id: String) -> Node3D:
	var main = MainScript.new()
	var root: Node3D = main._make_remote_player_node({"id": "9001", "character_class": class_id})
	main.free()
	return root


## Mirrors main._upsert_entity: node enters the tree (kit _ready aliases clips), then the
## AnimationController and reaction are built, and terminal death drives both.
func _enter_death(root: Node3D) -> ModelReactionController:
	get_root().add_child(root)
	await process_frame
	var ap := root.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		AnimationControllerScript.new(ap).enter_terminal("death")
	var reaction = ReactionControllerScript.new(root, BASE_TINT)
	reaction.enter_death(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	await create_timer(SETTLE_SECONDS).timeout
	return reaction


func _test_clip_owned_death_keeps_root_upright(label: String, root: Node3D) -> void:
	var ap := root.find_child("AnimationPlayer", true, false) as AnimationPlayer
	_assert(ap != null, "%s must carry an AnimationPlayer" % label)
	var base_rotation := root.rotation
	var reaction := await _enter_death(root)
	_assert(ap.has_animation("death"), "%s must expose a logical death clip" % label)
	_assert(ap.current_animation == "death", "%s death clip should be playing, got %s" % [label, ap.current_animation])
	_assert(root.rotation.distance_to(base_rotation) <= 0.001, "%s death must not lean the root, got %s" % [label, root.rotation])
	var state: Dictionary = reaction.get_debug_state()
	_assert(state["terminal"] == true, "%s death reaction should be terminal" % label)
	_assert(Color(str(state["current_tint"])).v < BASE_TINT.v, "%s death should still darken, got %s" % [label, state["current_tint"]])
	root.free()
	await process_frame


func _test_hit_then_death_restores_root_for_clip_owned() -> void:
	var root := _scene_root("monster_kit_skeleton_warrior")
	get_root().add_child(root)
	await process_frame
	var reaction = ReactionControllerScript.new(root, BASE_TINT)
	reaction.play_hit(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	await create_timer(0.07).timeout
	reaction.enter_death(Vector3(1.0, 0.0, 0.0), Vector3.BACK)
	await create_timer(SETTLE_SECONDS).timeout
	_assert(root.rotation.distance_to(Vector3.ZERO) <= 0.001, "hit lean must not survive into a clip-owned death, got %s" % root.rotation)
	root.free()
	await process_frame


func _test_clipless_death_still_leans() -> void:
	var root := TrainingDamageLogBridgeScript.make_silhouette_root()
	_assert(root.find_child("AnimationPlayer", true, false) == null, "training doll silhouette is the clip-less case")
	await _enter_death(root)
	_assert(absf(root.rotation.x) > 0.1 or absf(root.rotation.z) > 0.1, "clip-less death must lean the root, got %s" % root.rotation)
	root.free()
	await process_frame


func _assert(cond: bool, msg: String) -> void:
	if not cond:
		_failed = true
		printerr("[gdtest] FAIL: %s" % msg)
