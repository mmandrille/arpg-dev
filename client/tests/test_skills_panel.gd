# Unit test for the v44 skills panel.
# Run via: godot --headless --path client --script res://tests/test_skills_panel.gd
extends SceneTree

const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_remove_user_file(DraggableWindowScript.layout_storage_path)
	var panel := SkillsPanelScript.new()
	root.add_child(panel)
	await process_frame

	var emitted: Array = []
	panel.allocate_skill_point_requested.connect(func(skill_id: String) -> void:
		emitted.append(skill_id)
	)

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 3,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, true),
	})
	panel.set_interactive(true)
	panel.ensure_display_visible()
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("panel width matches stats", int(panel._panel.custom_minimum_size.x), 330)
	_assert_eq("panel height matches stats", int(panel._panel.custom_minimum_size.y), 500)
	var window: Dictionary = state.get("window", {})
	_assert_eq("skills window title", str(window.get("title", "")), "Skills")
	_assert_true("skills window has close button", bool(window.get("close_visible", false)))
	_assert_true("skills window is draggable", bool(window.get("draggable", false)))
	var stats_panel := CharacterStatsPanelScript.new()
	root.add_child(stats_panel)
	await process_frame
	stats_panel.ensure_display_visible()
	_assert_true("skills opens right of stats", panel._panel.position.x >= stats_panel._panel.position.x + stats_panel._panel.custom_minimum_size.x)
	_assert_true("skills top is on screen", panel._panel.position.y >= 0.0)
	stats_panel.queue_free()
	_assert_eq("unspent skill points", int(state.get("unspent_skill_points", -1)), 1)
	_assert_eq("class-filtered skill count", (state.get("skill_ids", []) as Array).size(), 1)
	_assert_eq("sorcerer skill visible", str((state.get("skill_ids", []) as Array)[0]), "magic_bolt")
	_assert_eq("skill name from catalog", str(state.get("skill_name", "")), "Magic Bolt")
	_assert_eq("skill icon from presentation", str(state.get("icon_label", "")), "M")
	_assert_eq("skill icon shape from presentation", str(state.get("icon_shape", "")), "bolt")
	_assert_eq("magic bolt rank", int(state.get("rank", -1)), 0)
	_assert_eq("magic bolt max rank", int(state.get("max_rank", -1)), 5)
	_assert_true("spend button enabled at initial magic requirement", bool(state.get("spend_button_enabled", false)))
	_assert_eq("unbought met requirement is highlighted", str(state.get("visual_state", "")), "highlight")
	_assert_true("requirements met at magic 5", bool(state.get("requirements_met", false)))
	var requirement_status: Array = state.get("requirement_status", [])
	_assert_eq("requirement row count", requirement_status.size(), 2)
	_assert_eq("magic requirement current", int((requirement_status[1] as Dictionary).get("current", -1)), 5)
	_assert_eq("no hovered skill by default", str(state.get("hovered_skill_id", "")), "")
	_assert_false("tooltip hidden by default", bool(state.get("tooltip_visible", true)))
	panel.bot_hover_skill("magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("hovered skill updates", str(state.get("hovered_skill_id", "")), "magic_bolt")
	_assert_true("tooltip visible on hover", bool(state.get("tooltip_visible", false)))
	_assert_true("tooltip includes mana from catalog", str(state.get("tooltip_body", "")).contains("Mana: 3"))
	_assert_true("tooltip lists requirements", str(state.get("tooltip_body", "")).contains("Requires:\nLevel 1\nMagic 5"))
	panel.bot_hover_skill("heal")
	state = panel.get_debug_state()
	_assert_eq("non-class skill selection ignored", str(state.get("selected_skill_id", "")), "magic_bolt")
	_assert_false("non-class skill absent from debug state", (state.get("skills", {}) as Dictionary).has("heal"))
	panel.bot_hover_skill("magic_bolt")

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 5,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, true),
	})
	state = panel.get_debug_state()
	_assert_true("requirements met at magic 5", bool(state.get("requirements_met", false)))
	_assert_true("spend button enabled after magic requirement", bool(state.get("spend_button_enabled", false)))
	_assert_eq("rankable skill is highlighted", str(state.get("visual_state", "")), "highlight")
	panel.bot_click_skill_button("magic_bolt")
	_assert_eq("spend signal count", emitted.size(), 1)
	_assert_eq("spend signal skill", str(emitted[0]), "magic_bolt")

	panel.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": _skill_rows(1, false),
	})
	state = panel.get_debug_state()
	_assert_eq("rank updates", int(state.get("rank", -1)), 1)
	_assert_false("spend disabled with no points", bool(state.get("spend_button_enabled", true)))
	_assert_eq("bought skill without points stays normal", str(state.get("visual_state", "")), "normal")
	var skills_without_points: Dictionary = state.get("skills", {})
	_assert_false("off-class no-points skill remains hidden", skills_without_points.has("heal"))

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, false),
	})
	state = panel.get_debug_state()
	_assert_false("rank 2 requirement blocks magic 5", bool(state.get("requirements_met", true)))
	_assert_false("rank 2 spend disabled before magic 8", bool(state.get("spend_button_enabled", true)))
	_assert_eq("bought skill missing next-rank req stays normal", str(state.get("visual_state", "")), "normal")
	_assert_true("tooltip includes rank 2 missing magic diff", str(state.get("tooltip_body", "")).contains("Magic 8(-3)"))

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 8},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, true),
	})
	state = panel.get_debug_state()
	_assert_true("rank 2 requirements met at magic 8", bool(state.get("requirements_met", false)))
	_assert_true("rank 2 spend enabled at magic 8", bool(state.get("spend_button_enabled", false)))
	panel.set_skill_bindings(["", "magic_bolt", "", "", "", "", "", ""], "magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("assigned key shown", str(state.get("assigned_key", "")), "F2")
	_assert_true("right click assigned state", bool(state.get("right_click_assigned", false)))
	var drag_start_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	panel.bot_drag_window_by(Vector2(40, 25))
	state = panel.get_debug_state()
	var moved_window: Dictionary = state.get("window", {})
	var moved_position: Dictionary = moved_window.get("position", {})
	_assert_eq("skills drag moved x", int(moved_position.get("x", 0)), int(drag_start_position.get("x", 0)) + 40)
	_assert_eq("skills drag moved y", int(moved_position.get("y", 0)), int(drag_start_position.get("y", 0)) + 25)
	panel.bot_drag_window_by(Vector2(-10000, -10000))
	state = panel.get_debug_state()
	var clamped_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	_assert_eq("skills drag clamps x", int(clamped_position.get("x", -1)), 0)
	_assert_eq("skills drag clamps y", int(clamped_position.get("y", -1)), 0)
	panel.bot_click_close()
	state = panel.get_debug_state()
	_assert_false("skills close button hides panel", bool(state.get("visible", true)))

	panel.queue_free()
	print("[gdtest] PASS: test_skills_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _remove_user_file(path: String) -> void:
	var absolute_path := ProjectSettings.globalize_path(path)
	if FileAccess.file_exists(absolute_path):
		DirAccess.remove_absolute(absolute_path)


func _skill_rows(magic_rank: int, magic_can_spend: bool, rage_rank: int = 0, rage_can_spend: bool = false, heal_rank: int = 0, heal_can_spend: bool = false) -> Array:
	return [
		{"skill_id": "magic_bolt", "rank": magic_rank, "max_rank": 5, "can_spend": magic_can_spend},
		{"skill_id": "rage", "rank": rage_rank, "max_rank": 5, "can_spend": rage_can_spend},
		{"skill_id": "heal", "rank": heal_rank, "max_rank": 5, "can_spend": heal_can_spend},
	]
