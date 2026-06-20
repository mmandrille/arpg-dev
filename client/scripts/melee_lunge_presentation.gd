class_name MeleeLungePresentation
extends RefCounted

const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")
const ATTACK_CLIPS := ["attack", "attack_off_hand"]
const LUNGE_DISTANCE := CombatFeelConfigScript.MELEE_LUNGE_DISTANCE
const RECOVERY_SECONDS := CombatFeelConfigScript.MELEE_LUNGE_RECOVERY_SECONDS
const SETTLE_EPSILON := CombatFeelConfigScript.MELEE_LUNGE_SETTLE_EPSILON
const META_BASE_POSITION := "_melee_lunge_base_position"
const META_TWEEN := "_melee_lunge_tween"
const META_COUNT := "_melee_lunge_count"


static func start(animation_player: AnimationPlayer, clip_name: String, attack_mode: String) -> bool:
	if attack_mode != "melee" or not (clip_name in ATTACK_CLIPS):
		return false
	var root := _lunge_root(animation_player)
	if root == null:
		return false
	var base := _base_position(root)
	_kill_existing(root)
	root.position = base + Vector3(0.0, 0.0, LUNGE_DISTANCE)
	root.set_meta(META_BASE_POSITION, base)
	root.set_meta(META_COUNT, int(root.get_meta(META_COUNT, 0)) + 1)
	if root.is_inside_tree():
		var tween := root.create_tween()
		tween.tween_property(root, "position", base, RECOVERY_SECONDS).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_OUT)
		root.set_meta(META_TWEEN, tween)
	return true


static func get_debug_state(animation_player: AnimationPlayer) -> Dictionary:
	var root := _lunge_root(animation_player)
	if root == null:
		return {"active": false, "offset_length": 0.0, "offset_z": 0.0, "count": 0}
	var offset := root.position - _base_position(root)
	return {
		"active": offset.length() > SETTLE_EPSILON,
		"offset_length": offset.length(),
		"offset_z": offset.z,
		"count": int(root.get_meta(META_COUNT, 0)),
	}


static func _lunge_root(animation_player: AnimationPlayer) -> Node3D:
	if animation_player == null:
		return null
	var visual := animation_player.get_parent()
	if visual == null:
		return null
	var model := visual.get_node_or_null("ModelRoot") as Node3D
	if model != null:
		return model
	return visual as Node3D


static func _base_position(root: Node3D) -> Vector3:
	if root == null:
		return Vector3.ZERO
	if root.has_meta(META_BASE_POSITION):
		var base = root.get_meta(META_BASE_POSITION)
		if typeof(base) == TYPE_VECTOR3:
			return base
	return root.position


static func _kill_existing(root: Node3D) -> void:
	if root == null or not root.has_meta(META_TWEEN):
		return
	var previous = root.get_meta(META_TWEEN)
	if previous is Tween and is_instance_valid(previous):
		(previous as Tween).kill()
	root.remove_meta(META_TWEEN)
