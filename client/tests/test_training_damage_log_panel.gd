extends SceneTree

const TrainingDamageLogPanelScript := preload("res://scripts/training_damage_log_panel.gd")
const CombatBreakdownFormatScript := preload("res://scripts/combat_breakdown_format.gd")
const TrainingDamageLogBridgeScript := preload("res://scripts/training_damage_log_bridge.gd")
const TrainingDollVisualScript := preload("res://scripts/training_doll_visual.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const MainScript := preload("res://scripts/main.gd")

var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_panel_open_close()
	_test_breakdown_format()
	_test_revive_resets_death_pose()
	print("[gdtest] PASS: test_training_damage_log_panel (%d failed)" % _fail_count)
	quit(1 if _fail_count > 0 else 0)


func _test_panel_open_close() -> void:
	var panel := TrainingDamageLogPanelScript.new()
	root.add_child(panel)
	await process_frame
	var event := {
		"event_type": "monster_damaged",
		"skill_id": "",
		"weapon_slot": "main_hand",
		"damage": 3,
		"outcome": "hit",
		"damage_breakdown": [
			{"label": "Attack", "value": "Basic attack (main_hand)"},
			{"label": "Final damage", "value": "3"},
		],
	}
	panel.on_training_doll_combat_event(event, "town_training_doll")
	if not panel.is_open() or panel.get_entry_count() != 1:
		_fail_count += 1
		push_error("panel should open with one entry")
	panel.bot_click_close()
	await process_frame
	if panel.is_open():
		_fail_count += 1
		push_error("panel should close after X")


func _test_breakdown_format() -> void:
	var event := {
		"skill_id": "magic_bolt",
		"damage": 7,
		"outcome": "hit",
		"damage_breakdown": [{"label": "Final damage", "value": "7"}],
	}
	if CombatBreakdownFormatScript.attack_title(event) != "Magic Bolt":
		_fail_count += 1
		push_error("unexpected attack title")
	if CombatBreakdownFormatScript.outcome_label(event) != "Hit — 7":
		_fail_count += 1
		push_error("unexpected outcome label")


func _test_revive_resets_death_pose() -> void:
	var visual_root := TrainingDollVisualScript.make_node()
	var reaction := ModelReactionControllerScript.new(visual_root, Color.WHITE)
	reaction.enter_death()
	var main = MainScript.new()
	main.entities = {
		"42": {
			"node": visual_root,
			"reaction": reaction,
			"controller": null,
			"max_hp": 100,
		},
	}
	TrainingDamageLogBridgeScript.handle_training_doll_revived(main, "42", {"damage": 100})
	if reaction.is_terminal():
		_fail_count += 1
		push_error("training doll revive should clear death reaction")
	if absf(visual_root.rotation.x) > 0.01 or absf(visual_root.rotation.z) > 0.01:
		_fail_count += 1
		push_error("training doll revive should restore upright pose")
