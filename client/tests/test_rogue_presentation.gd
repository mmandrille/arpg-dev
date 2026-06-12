extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const DamageNumberScript := preload("res://scripts/damage_number.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_poison_dot_shows_green_floating_text()
	_test_poison_status_tints_monster_until_end()
	_test_off_hand_combat_event_uses_off_hand_attack_clip()
	_test_rogue_visual_replay_dwell_events()
	print("[gdtest] PASS: test_rogue_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	root.add_child(main.player_anchor)
	root.add_child(main.entities_root)
	root.add_child(main.walls_root)
	return main


func _test_poison_dot_shows_green_floating_text() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.client_settings = ClientSettingsScript.new()
	main.damage_numbers_layer = CanvasLayer.new()
	main._camera = Camera3D.new()
	var monster := Node3D.new()
	monster.position = Vector3(4.0, 0.0, 4.0)
	main.entities_root.add_child(monster)
	main.entities["2001"] = {"node": monster, "type": "monster", "hp": 5}
	root.add_child(main.damage_numbers_layer)
	root.add_child(main._camera)
	main._camera.look_at_from_position(Vector3(4.0, 12.0, 14.0), monster.position, Vector3.UP)
	main._apply_delta({"events": [{
		"event_type": "monster_damaged",
		"entity_id": "2001",
		"target_entity_id": "2001",
		"source_entity_id": "1001",
		"skill_id": "poison_stab",
		"outcome": "hit",
		"damage": 2
	}], "changes": []})
	var numbers := main._bot_damage_numbers()
	_assert_eq("poison floating text count", numbers.size(), 1)
	_assert_eq("poison floating text", str((numbers[0] as Dictionary).get("text", "")), "2")
	_assert_eq("poison floating text variant", str((numbers[0] as Dictionary).get("variant", "")), "poison")
	_assert_eq("poison floating text color", str((numbers[0] as Dictionary).get("color", "")), "55e66f")
	main.damage_numbers_layer.queue_free()
	main._camera.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_poison_status_tints_monster_until_end() -> void:
	var main = _make_main()
	var monster := Node3D.new()
	var mesh := MeshInstance3D.new()
	mesh.mesh = BoxMesh.new()
	monster.add_child(mesh)
	main.entities_root.add_child(monster)
	var base_tint := MainScript.MONSTER_RARITY_TINTS["common"]
	main.entities["2001"] = {
		"node": monster,
		"type": "monster",
		"base_tint": base_tint.to_html(false),
		"reaction": ModelReactionControllerScript.new(monster, base_tint),
	}
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "2001",
		"skill_id": "poison_stab"
	}], "changes": []})
	_assert_eq("poison status tint active", _mesh_tint(mesh).to_html(false), MainScript.POISON_TINT.to_html(false))
	main._apply_delta({"events": [{
		"event_type": "monster_damaged",
		"entity_id": "2001",
		"target_entity_id": "2001",
		"source_entity_id": "1001",
		"skill_id": "poison_stab",
		"outcome": "hit",
		"damage": 1
	}], "changes": []})
	var reaction_debug: Dictionary = (main.entities["2001"] as Dictionary).get("reaction").get_debug_state()
	_assert_eq("poison reaction base tint", str(reaction_debug.get("base_tint", "")), MainScript.POISON_TINT.to_html(false))
	main._apply_delta({"events": [{
		"event_type": "skill_effect_ended",
		"entity_id": "2001",
		"skill_id": "poison_stab"
	}], "changes": []})
	_assert_eq("poison status tint restored", _mesh_tint(mesh).to_html(false), base_tint.to_html(false))
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_off_hand_combat_event_uses_off_hand_attack_clip() -> void:
	var main = _make_main()
	var ap := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	for clip in ["idle", "attack", "attack_off_hand"]:
		var anim := Animation.new()
		anim.length = 0.35
		lib.add_animation(clip, anim)
	ap.add_animation_library("", lib)
	root.add_child(ap)
	main.player_anim = AnimationControllerScript.new(ap)
	main._play_local_attack_animation_for_event({"weapon_slot": "off_hand"})
	_assert_eq("off hand attack clip", main.player_anim.current_clip(), "attack_off_hand")
	main._play_local_player_reaction_animation("hit")
	_assert_eq("off hand not interrupted by hit", main.player_anim.current_clip(), "attack_off_hand")
	main._play_local_attack_animation_for_event({"weapon_slot": "main_hand"})
	_assert_eq("main hand attack clip", main.player_anim.current_clip(), "attack")
	ap.queue_free()
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _test_rogue_visual_replay_dwell_events() -> void:
	var main = _make_main()
	main.autoplay_step_delay = 0.45
	var poison_delay := main._visual_replay_delay_for({
		"type": "state_delta",
		"payload": {"events": [{"event_type": "skill_effect_started", "skill_id": "poison_stab"}], "changes": []}
	})
	var offhand_delay := main._visual_replay_delay_for({
		"type": "state_delta",
		"payload": {"events": [{"event_type": "attack_blocked", "weapon_slot": "off_hand"}], "changes": []}
	})
	_assert_true("poison replay delay dwells", poison_delay > main.autoplay_step_delay)
	_assert_true("off hand replay delay dwells", offhand_delay > main.autoplay_step_delay)
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _mesh_tint(mesh: MeshInstance3D) -> Color:
	var mat := mesh.material_override as StandardMaterial3D
	if mat == null:
		return Color.TRANSPARENT
	return mat.albedo_color


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: got %s want %s" % [label, str(got), str(want)])


func _assert_true(label: String, got: bool) -> void:
	if got:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: got false want true" % label)
