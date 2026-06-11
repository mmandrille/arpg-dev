# Unit test for the v44 skills panel.
# Run via: godot --headless --path client --script res://tests/test_skills_panel.gd
extends SceneTree

const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel := SkillsPanelScript.new()
	root.add_child(panel)
	await process_frame

	var emitted: Array = []
	panel.allocate_skill_point_requested.connect(func(skill_id: String) -> void:
		emitted.append(skill_id)
	)

	panel.set_character_progression({
		"level": 3,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, false),
	})
	panel.set_interactive(true)
	panel.ensure_display_visible()
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("panel width matches stats", int(panel._panel.custom_minimum_size.x), 330)
	_assert_eq("panel height matches stats", int(panel._panel.custom_minimum_size.y), 500)
	var stats_panel := CharacterStatsPanelScript.new()
	root.add_child(stats_panel)
	await process_frame
	stats_panel.ensure_display_visible()
	_assert_true("skills opens right of stats", panel._panel.position.x >= stats_panel._panel.position.x + stats_panel._panel.custom_minimum_size.x)
	_assert_eq("skills aligns with stats top", int(panel._panel.position.y), int(stats_panel._panel.position.y))
	stats_panel.queue_free()
	_assert_eq("unspent skill points", int(state.get("unspent_skill_points", -1)), 1)
	_assert_eq("first-row skill count", (state.get("skill_ids", []) as Array).size(), 3)
	_assert_eq("skill name from catalog", str(state.get("skill_name", "")), "Magic Bolt")
	_assert_eq("skill icon from presentation", str(state.get("icon_label", "")), "M")
	_assert_eq("magic bolt rank", int(state.get("rank", -1)), 0)
	_assert_eq("magic bolt max rank", int(state.get("max_rank", -1)), 5)
	_assert_false("spend button disabled before magic requirement", bool(state.get("spend_button_enabled", true)))
	_assert_eq("unbought missing requirement is disabled", str(state.get("visual_state", "")), "disabled")
	_assert_false("requirements not met before magic 15", bool(state.get("requirements_met", true)))
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
	_assert_true("tooltip lists requirements", str(state.get("tooltip_body", "")).contains("Requires:\nLevel 1\nMagic 15(-10)"))
	panel.bot_hover_skill("heal")
	state = panel.get_debug_state()
	_assert_eq("heal selection updates", str(state.get("selected_skill_id", "")), "heal")
	_assert_eq("heal icon from presentation", str(state.get("icon_label", "")), "H")
	_assert_true("heal tooltip includes mana", str(state.get("tooltip_body", "")).contains("Mana: 10"))
	panel.bot_hover_skill("magic_bolt")

	panel.set_character_progression({
		"level": 5,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 15},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, true),
	})
	state = panel.get_debug_state()
	_assert_true("requirements met after magic 15", bool(state.get("requirements_met", false)))
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
	_assert_eq("unbought no-points skill is disabled", str((skills_without_points.get("heal", {}) as Dictionary).get("visual_state", "")), "disabled")

	panel.set_character_progression({
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 15},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, false),
	})
	state = panel.get_debug_state()
	_assert_false("rank 2 requirement blocks magic 15", bool(state.get("requirements_met", true)))
	_assert_false("rank 2 spend disabled before magic 20", bool(state.get("spend_button_enabled", true)))
	_assert_eq("bought skill missing next-rank req stays normal", str(state.get("visual_state", "")), "normal")
	_assert_true("tooltip includes rank 2 missing magic diff", str(state.get("tooltip_body", "")).contains("Magic 20(-5)"))

	panel.set_character_progression({
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": 20},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, true),
	})
	state = panel.get_debug_state()
	_assert_true("rank 2 requirements met at magic 20", bool(state.get("requirements_met", false)))
	_assert_true("rank 2 spend enabled at magic 20", bool(state.get("spend_button_enabled", false)))
	panel.set_skill_bindings(["", "magic_bolt", "", "", "", "", "", ""], "magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("assigned key shown", str(state.get("assigned_key", "")), "F2")
	_assert_true("right click assigned state", bool(state.get("right_click_assigned", false)))

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


func _skill_rows(magic_rank: int, magic_can_spend: bool, rage_rank: int = 0, rage_can_spend: bool = false, heal_rank: int = 0, heal_can_spend: bool = false) -> Array:
	return [
		{"skill_id": "magic_bolt", "rank": magic_rank, "max_rank": 5, "can_spend": magic_can_spend},
		{"skill_id": "rage", "rank": rage_rank, "max_rank": 5, "can_spend": rage_can_spend},
		{"skill_id": "heal", "rank": heal_rank, "max_rank": 5, "can_spend": heal_can_spend},
	]
