# Unit tests for ProjectileFlightPresentation.
extends SceneTree

const CombatReachScript := preload("res://scripts/combat_reach.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")
const ProjectileFlightPresentationScript := preload("res://scripts/projectile_flight_presentation.gd")
const ProjectilePresentationCapScript := preload("res://scripts/projectile_presentation_cap.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_visual_id_maps_skill_presentation()
	_test_flight_path_uses_owner_to_target()
	_test_motion_segment_uses_from_to_positions()
	_test_presentation_cap_keeps_flight_authority_hidden()
	_test_local_player_reach_wrapper()
	print("[gdtest] PASS: test_projectile_flight_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_visual_id_maps_skill_presentation() -> void:
	var visual := ProjectileFlightPresentationScript._visual_id_for_def("magic_bolt")
	_assert_eq("magic bolt uses presentation visual", visual, "magic_bolt_projectile")


func _test_flight_path_uses_owner_to_target() -> void:
	var owner_node := Node3D.new()
	owner_node.position = Vector3(1.0, 0.0, 2.0)
	var target_node := Node3D.new()
	target_node.position = Vector3(5.0, 0.0, 2.0)
	var entities := {
		"owner": {"type": "monster", "node": owner_node, "monster_def_id": "dungeon_archer"},
		"target": {"type": "monster", "node": target_node},
	}
	var entity := {
		"owner_id": "owner",
		"target_id": "target",
		"projectile_def_id": "training_arrow",
		"position": {"x": 1.0, "y": 2.0},
	}
	var path := ProjectileFlightPresentationScript._flight_path_from_entity(
		entity,
		entities,
		[],
		{},
		"",
		"",
		null,
	)
	_assert_true("flight path computed", not path.is_empty())
	_assert_true("flight starts ahead of owner", float(path["start"].x) > 1.0)
	_assert_true("flight ends toward target", float(path["finish"].x) > float(path["start"].x))
	owner_node.queue_free()
	target_node.queue_free()


func _test_motion_segment_uses_from_to_positions() -> void:
	ItemRulesLoaderScript.ensure_loaded()
	var inventory := [{
		"item_instance_id": "bow-1",
		"item_def_id": "training_bow",
	}]
	var equipped := {"main_hand": "bow-1"}
	var entities_root := Node3D.new()
	var tween_parent := Node.new()
	get_root().add_child(tween_parent)
	get_root().add_child(entities_root)
	var entity := {
		"projectile_def_id": "training_arrow",
		"owner_id": "player-1",
		"position": {"x": 2.5, "y": 0.0},
	}
	ProjectileFlightPresentationScript.spawn_from_motion_segment(
		tween_parent,
		entities_root,
		entity,
		{},
		Vector3.ZERO,
		Vector3(2.5, 0.0, 0.0),
		2.5,
		inventory,
		equipped,
		"ranger",
		"player-1",
		0.8,
		false,
	)
	var flight := entities_root.get_node_or_null("ProjectileFlight_training_arrow")
	_assert_true("motion segment spawns flight visual", flight != null)
	if flight != null:
		_assert_true("flight visual starts near muzzle offset", float(flight.position.x) >= 0.4)
		flight.queue_free()
	entities_root.queue_free()
	tween_parent.queue_free()


func _test_presentation_cap_keeps_flight_authority_hidden() -> void:
	var node := Node3D.new()
	node.visible = false
	var entities := {
		"p1": {"type": "projectile", "node": node, "use_flight_visual": true},
	}
	ProjectilePresentationCapScript.apply(entities, Vector3.ZERO)
	_assert_true("flight authority node stays hidden", not node.visible)
	node.queue_free()


func _test_local_player_reach_wrapper() -> void:
	ItemRulesLoaderScript.ensure_loaded()
	var inventory := [{
		"item_instance_id": "staff-1",
		"item_def_id": "sorcerer_staff",
	}]
	var equipped := {"main_hand": "staff-1"}
	var reach := CombatReachScript.local_player_attack_reach(inventory, equipped, "sorcerer")
	_assert_true("staff reach is positive", reach > 0.0)


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s (got %s want %s)" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL: %s" % label)
