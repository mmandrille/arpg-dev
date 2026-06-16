# Unit tests for local status-effect presentation markers.
#
# Run via: godot --headless --path client --script res://tests/test_status_effect_presentation.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const HealRainEffectScript := preload("res://scripts/heal_rain_effect.gd")
const ConsumableHealEffectScript := preload("res://scripts/consumable_heal_effect.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_rage_effect_started_drives_world_aura()
	_test_rage_icon_expiry_clears_world_aura()
	_test_holy_shield_started_blinks_models_in_range()
	_test_holy_shield_ended_clears_local_world_effect()
	_test_sanctuary_started_and_ended_updates_local_dome()
	_test_unique_burn_started_and_ended_updates_monster_cue()
	_test_pinning_root_started_and_ended_updates_monster_cue()
	_test_stun_started_and_ended_updates_monster_cue_for_leap_and_charge()
	_test_rogue_mark_effect_id_updates_monster_skull()
	_test_monster_death_clears_elite_aura_markers()
	_test_potion_heal_uses_personal_effect()
	_test_paladin_heal_uses_area_rain()

	print("[gdtest] PASS: test_status_effect_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _make_main():
	var main = MainScript.new()
	main.player_anchor = Node3D.new()
	main.entities_root = Node3D.new()
	main.walls_root = Node3D.new()
	get_root().add_child(main.player_anchor)
	get_root().add_child(main.entities_root)
	get_root().add_child(main.walls_root)
	return main


func _test_rage_effect_started_drives_world_aura() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{"event_type": "skill_effect_started", "entity_id": "1001", "skill_id": "rage", "amount": 25}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local rage marker active", bool(local_state.get("has_rage_effect", false)))
	_assert_float("local rage visual scale active", float(local_state.get("visual_scale", 0.0)), 1.25)

	main._apply_delta({"events": [{"event_type": "skill_effect_ended", "entity_id": "1001", "skill_id": "rage"}], "changes": []})
	local_state = main._bot_local_player_presentation()
	_assert_true("local rage marker removed", not bool(local_state.get("has_rage_effect", true)))
	_assert_float("local rage visual scale reset", float(local_state.get("visual_scale", 0.0)), 1.0)
	_free_main(main)


func _test_rage_icon_expiry_clears_world_aura() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{"event_type": "skill_effect_started", "entity_id": "1001", "skill_id": "rage", "amount": 25, "remaining_ticks": 1, "total_ticks": 1}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local rage marker active before icon expiry", bool(local_state.get("has_rage_effect", false)))

	main._on_status_effect_expired("rage")
	local_state = main._bot_local_player_presentation()
	_assert_true("local rage marker removed by icon expiry", not bool(local_state.get("has_rage_effect", true)))
	_assert_float("local rage visual scale reset by icon expiry", float(local_state.get("visual_scale", 0.0)), 1.0)
	_free_main(main)


func _test_holy_shield_started_blinks_models_in_range() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main.player_anchor.position = Vector3(2.0, 0.0, 3.0)
	main._upsert_entity({
		"id": "1002",
		"type": "player",
		"position": {"x": 5.0, "y": 3.0},
		"hp": 10,
		"max_hp": 10,
		"character_id": "char_near",
	})
	main._upsert_entity({
		"id": "1003",
		"type": "monster",
		"position": {"x": 13.0, "y": 3.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
	})
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "1001",
		"source_entity_id": "1001",
		"target_entity_id": "1001",
		"skill_id": "holy_shield",
		"correlation_id": "corr_holy_shield_blink",
	}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_eq("local holy shield aura pulse", int(local_state.get("holy_shield_aura_pulses", 0)), 1)
	_assert_eq("local holy shield target pulse", int(local_state.get("holy_shield_target_pulses", 0)), 1)
	var remote_state: Array = main._bot_entities_presentation_debug()
	var near_pulses := -1
	var far_pulses := -1
	for row in remote_state:
		var rec: Dictionary = row
		if str(rec.get("id", "")) == "1002":
			near_pulses = int(rec.get("holy_shield_target_pulses", 0))
		if str(rec.get("id", "")) == "1003":
			far_pulses = int(rec.get("holy_shield_target_pulses", 0))
	_assert_eq("near model holy shield target pulse", near_pulses, 1)
	_assert_eq("far model holy shield target pulse", far_pulses, 0)
	_free_main(main)


func _test_holy_shield_ended_clears_local_world_effect() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1001",
		"type": "player",
		"position": {"x": 0.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"effect_ids": ["holy_shield"],
	})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local holy shield marker active before end event", bool(local_state.get("has_holy_shield_effect", false)))

	main._apply_delta({"events": [{"event_type": "skill_effect_ended", "entity_id": "1001", "skill_id": "holy_shield"}], "changes": []})
	local_state = main._bot_local_player_presentation()
	_assert_true("local holy shield marker removed by end event", not bool(local_state.get("has_holy_shield_effect", true)))
	_free_main(main)


func _test_sanctuary_started_and_ended_updates_local_dome() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "1001",
		"skill_id": "sanctuary",
		"remaining_ticks": 60,
		"total_ticks": 60,
	}], "changes": []})
	var local_state := main._bot_local_player_presentation()
	_assert_true("local sanctuary dome active", bool(local_state.get("has_sanctuary_effect", false)))
	_assert_true("local sanctuary effect id visible", (local_state.get("effect_ids", []) as Array).has("sanctuary"))
	var marker := main.player_anchor.find_child("SanctuaryDomeEffect", false, false) as Node3D
	var dome := marker.find_child("SanctuaryDome", false, false) as MeshInstance3D
	var ground := marker.find_child("SanctuaryGround", false, false) as MeshInstance3D
	var expected_radius := _skill_effect_radius("sanctuary", 5.0)
	_assert_float("sanctuary dome radius follows skill rule", (dome.mesh as SphereMesh).radius, expected_radius)
	_assert_float("sanctuary ground top radius follows skill rule", (ground.mesh as CylinderMesh).top_radius, expected_radius)
	_assert_float("sanctuary ground bottom radius follows skill rule", (ground.mesh as CylinderMesh).bottom_radius, expected_radius)
	_assert_float("sanctuary dome alpha is subtle", (dome.material_override as StandardMaterial3D).albedo_color.a, 0.10)

	main._apply_delta({"events": [{"event_type": "skill_effect_ended", "entity_id": "1001", "skill_id": "sanctuary"}], "changes": []})
	local_state = main._bot_local_player_presentation()
	_assert_true("local sanctuary dome removed by end event", not bool(local_state.get("has_sanctuary_effect", true)))
	_assert_true("local sanctuary effect id removed", not (local_state.get("effect_ids", []) as Array).has("sanctuary"))
	_free_main(main)


func _test_unique_burn_started_and_ended_updates_monster_cue() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1002",
		"type": "monster",
		"position": {"x": 2.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
	})
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "1002",
		"source_entity_id": "1001",
		"target_entity_id": "1002",
		"skill_id": "everburning_wound",
		"damage_type": "fire",
	}], "changes": []})
	var rows: Array = main._bot_entities_presentation_debug()
	var has_burning := false
	for row in rows:
		var rec: Dictionary = row
		if str(rec.get("id", "")) == "1002":
			has_burning = bool(rec.get("has_burning_effect", false))
	_assert_true("monster burning marker active", has_burning)

	main._apply_delta({"events": [{
		"event_type": "skill_effect_ended",
		"entity_id": "1002",
		"source_entity_id": "1001",
		"target_entity_id": "1002",
		"skill_id": "everburning_wound",
	}], "changes": []})
	rows = main._bot_entities_presentation_debug()
	has_burning = true
	for row in rows:
		var rec: Dictionary = row
		if str(rec.get("id", "")) == "1002":
			has_burning = bool(rec.get("has_burning_effect", true))
	_assert_true("monster burning marker removed", not has_burning)
	_free_main(main)


func _test_pinning_root_started_and_ended_updates_monster_cue() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1002",
		"type": "monster",
		"position": {"x": 2.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
	})
	main._apply_delta({"events": [{
		"event_type": "skill_effect_started",
		"entity_id": "1002",
		"source_entity_id": "1001",
		"target_entity_id": "1002",
		"skill_id": "pinning_shot",
	}], "changes": []})
	var rows: Array = main._bot_entities_presentation_debug()
	var before: Dictionary = _presentation_row(rows, "1002")
	_assert_true("monster pinning root marker active", bool(before.get("has_pinning_root_effect", false)))
	_assert_true("monster pinning root effect id tracked", (before.get("effect_ids", []) as Array).has("pinning_root"))

	main._apply_delta({"events": [{
		"event_type": "skill_effect_ended",
		"entity_id": "1002",
		"source_entity_id": "1001",
		"target_entity_id": "1002",
		"skill_id": "pinning_shot",
	}], "changes": []})
	rows = main._bot_entities_presentation_debug()
	var after: Dictionary = _presentation_row(rows, "1002")
	_assert_true("monster pinning root marker removed", not bool(after.get("has_pinning_root_effect", true)))
	_assert_true("monster pinning root effect id removed", not (after.get("effect_ids", []) as Array).has("pinning_root"))
	_free_main(main)


func _test_stun_started_and_ended_updates_monster_cue_for_leap_and_charge() -> void:
	for skill_id in ["leap", "charge", "dash"]:
		var main = _make_main()
		main.player_id = "1001"
		main._upsert_entity({
			"id": "1002",
			"type": "monster",
			"position": {"x": 2.0, "y": 0.0},
			"hp": 10,
			"max_hp": 10,
			"monster_def_id": "training_dummy",
		})
		main._apply_delta({"events": [{
			"event_type": "skill_effect_started",
			"entity_id": "1002",
			"source_entity_id": "1001",
			"target_entity_id": "1002",
			"skill_id": skill_id,
		}], "changes": []})
		var before: Dictionary = _presentation_row(main._bot_entities_presentation_debug(), "1002")
		_assert_true("%s stun marker active" % skill_id, bool(before.get("has_stun_effect", false)))
		_assert_true("%s uses shared stun effect id" % skill_id, (before.get("effect_ids", []) as Array).has("stun"))

		main._apply_delta({"events": [{
			"event_type": "skill_effect_ended",
			"entity_id": "1002",
			"source_entity_id": "1001",
			"target_entity_id": "1002",
			"skill_id": skill_id,
		}], "changes": []})
		var after: Dictionary = _presentation_row(main._bot_entities_presentation_debug(), "1002")
		_assert_true("%s stun marker removed" % skill_id, not bool(after.get("has_stun_effect", true)))
		_assert_true("%s shared stun effect id removed" % skill_id, not (after.get("effect_ids", []) as Array).has("stun"))
		_free_main(main)


func _test_rogue_mark_effect_id_updates_monster_skull() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1002",
		"type": "monster",
		"position": {"x": 2.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
		"effect_ids": ["rogue_mark"],
	})
	var before: Dictionary = _presentation_row(main._bot_entities_presentation_debug(), "1002")
	_assert_true("rogue mark skull marker active", bool(before.get("has_rogue_mark_effect", false)))
	var monster: Dictionary = main.entities.get("1002", {})
	var node := monster.get("node", null) as Node3D
	var marker := node.find_child("RogueMarkSkullEffect", false, false) as Node3D
	_assert_true("rogue mark skull node exists", marker != null)
	if marker != null:
		_assert_float("rogue mark skull floats higher", marker.position.y, 3.05)
		var skull := marker.find_child("RogueMarkSkull", false, false) as Label3D
		_assert_true("rogue mark skull label exists", skull != null)
		if skull != null:
			_assert_eq("rogue mark skull text", skull.text, "☠")
			_assert_eq("rogue mark skull font size", skull.font_size, 192)
			_assert_eq("rogue mark skull color", skull.modulate.to_html(false), "ff140d")

	main._apply_delta({"events": [], "changes": [{
		"op": "entity_update",
		"entity": {
			"id": "1002",
			"type": "monster",
			"position": {"x": 2.0, "y": 0.0},
			"hp": 10,
			"max_hp": 10,
			"monster_def_id": "training_dummy",
			"effect_ids": [],
		},
	}]})
	var after: Dictionary = _presentation_row(main._bot_entities_presentation_debug(), "1002")
	_assert_true("rogue mark skull marker removed", not bool(after.get("has_rogue_mark_effect", true)))
	_free_main(main)


func _test_monster_death_clears_elite_aura_markers() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._upsert_entity({
		"id": "1002",
		"type": "monster",
		"position": {"x": 2.0, "y": 0.0},
		"hp": 10,
		"max_hp": 10,
		"monster_def_id": "training_dummy",
		"effect_ids": ["elite_command"],
		"monster_pack_id": "pack_1",
		"monster_pack_leader": true,
	})
	main._upsert_entity({
		"id": "1003",
		"type": "monster",
		"position": {"x": 2.5, "y": 0.0},
		"hp": 5,
		"max_hp": 5,
		"monster_def_id": "training_dummy",
		"effect_ids": ["elite_command"],
		"monster_pack_id": "pack_1",
	})
	var rows: Array = main._bot_entities_presentation_debug()
	var before: Dictionary = _presentation_row(rows, "1002")
	var minion_before: Dictionary = _presentation_row(rows, "1003")
	_assert_true("elite command marker active before death", bool(before.get("has_elite_command_effect", false)))
	_assert_true("elite command radius active before death", bool(before.get("has_elite_command_radius_preview", false)))
	_assert_true("elite minion command marker active before leader death", bool(minion_before.get("has_elite_command_effect", false)))

	main.entities["1002"]["reaction"] = null
	main._apply_delta({"events": [{
		"event_type": "monster_killed",
		"entity_id": "1002",
		"source_entity_id": "1001",
		"target_entity_id": "1002",
	}], "changes": []})
	rows = main._bot_entities_presentation_debug()
	var after: Dictionary = _presentation_row(rows, "1002")
	var minion_after: Dictionary = _presentation_row(rows, "1003")
	_assert_eq("elite monster hp after death", int(after.get("hp", 1)), 0)
	_assert_true("elite command marker removed on death", not bool(after.get("has_elite_command_effect", true)))
	_assert_true("elite command radius removed on death", not bool(after.get("has_elite_command_radius_preview", true)))
	_assert_true("elite minion command marker removed after leader death", not bool(minion_after.get("has_elite_command_effect", true)))
	_free_main(main)


func _test_potion_heal_uses_personal_effect() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{
		"event_type": "player_healed",
		"entity_id": "1001",
		"item_instance_id": "potion_1",
		"heal": 5,
	}], "changes": []})
	_assert_eq("potion heal does not spawn area rain", _child_count(main, HealRainEffectScript), 0)
	_assert_eq("potion heal spawns personal effect", _child_count(main, ConsumableHealEffectScript), 1)
	_free_main(main)


func _test_paladin_heal_uses_area_rain() -> void:
	var main = _make_main()
	main.player_id = "1001"
	main._apply_delta({"events": [{
		"event_type": "player_healed",
		"entity_id": "1001",
		"skill_id": "heal",
		"heal": 5,
	}], "changes": []})
	_assert_eq("paladin heal spawns area rain", _child_count(main, HealRainEffectScript), 1)
	_assert_eq("paladin heal does not spawn potion effect", _child_count(main, ConsumableHealEffectScript), 0)
	_free_main(main)


func _free_main(main) -> void:
	main.player_anchor.queue_free()
	main.entities_root.queue_free()
	main.walls_root.queue_free()
	main.free()


func _child_count(root: Node, script: Script) -> int:
	var count := 0
	for child in root.get_children():
		if child.get_script() == script:
			count += 1
	return count


func _presentation_row(rows: Array, entity_id: String) -> Dictionary:
	for row in rows:
		var rec: Dictionary = row
		if str(rec.get("id", "")) == entity_id:
			return rec
	return {}


func _skill_effect_radius(skill_id: String, fallback: float) -> float:
	var def := SkillRulesLoaderScript.skill_definition(skill_id)
	for effect in def.get("effects", []):
		var row: Dictionary = effect
		if str(row.get("effect_id", skill_id)) == skill_id and row.has("radius"):
			return maxf(float(row.get("radius", fallback)), 0.5)
	return fallback


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_float(label: String, got: float, want: float, epsilon: float = 0.001) -> void:
	if abs(got - want) <= epsilon:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
