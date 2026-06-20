# Unit test for the v44 skills panel.
# Run via: godot --headless --path client --script res://tests/test_skills_panel.gd
extends SceneTree

const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const DraggableWindowScript := preload("res://scripts/draggable_window.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_remove_user_file(DraggableWindowScript.layout_storage_path)
	SkillRulesLoaderScript.reset_for_tests()
	SkillRulesLoaderScript.ensure_loaded()
	var magic_bolt_max_rank := _skill_max_rank("magic_bolt")
	var magic_rank1_req := _skill_stat_requirement("magic_bolt", "magic", 0)
	var magic_rank2_req := _skill_stat_requirement("magic_bolt", "magic", 1)
	var magic_bolt_level_req := _skill_level_requirement("magic_bolt", 0)
	var magic_bolt_mana_cost := _skill_mana_cost("magic_bolt", 0)
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
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": magic_rank1_req},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, true),
	})
	panel.set_interactive(true)
	panel.ensure_display_visible()
	var state := panel.get_debug_state()
	_assert_true("panel visible", bool(state.get("visible", false)))
	_assert_eq("panel width is 30 percent larger", int(panel._panel.custom_minimum_size.x), 429)
	_assert_eq("panel height is 30 percent larger", int(panel._panel.custom_minimum_size.y), 650)
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
	_assert_eq("class-filtered skill count", (state.get("skill_ids", []) as Array).size(), 9)
	_assert_eq("sorcerer skill visible", str((state.get("skill_ids", []) as Array)[0]), "magic_bolt")
	_assert_eq("sorcerer second skill visible", str((state.get("skill_ids", []) as Array)[1]), "teleport")
	_assert_eq("sorcerer passive column row 1 visible", str((state.get("skill_ids", []) as Array)[2]), "arcane_focus")
	_assert_eq("sorcerer third skill visible", str((state.get("skill_ids", []) as Array)[3]), "ice_shard")
	_assert_eq("sorcerer fourth skill visible", str((state.get("skill_ids", []) as Array)[4]), "ligthing")
	_assert_eq("sorcerer fifth skill visible", str((state.get("skill_ids", []) as Array)[5]), "revive")
	_assert_eq("sorcerer passive column row 2 visible", str((state.get("skill_ids", []) as Array)[6]), "mana_weaving")
	_assert_eq("sorcerer sixth skill visible", str((state.get("skill_ids", []) as Array)[7]), "arcane_barrage")
	_assert_eq("sorcerer passive column row 3 visible", str((state.get("skill_ids", []) as Array)[8]), "spell_dynamo")
	_assert_eq("skill name from catalog", str(state.get("skill_name", "")), "Magic Bolt")
	_assert_eq("skill icon from presentation", str(state.get("icon_label", "")), "M")
	_assert_eq("skill icon shape from presentation", str(state.get("icon_shape", "")), "orb_projectile")
	_assert_eq("magic bolt rank", int(state.get("rank", -1)), 0)
	_assert_eq("magic bolt max rank", int(state.get("max_rank", -1)), magic_bolt_max_rank)
	_assert_true("spend button enabled at initial magic requirement", bool(state.get("spend_button_enabled", false)))
	_assert_false("spend button is not visible", bool(state.get("spend_button_visible", true)))
	_assert_eq("unbought met requirement is highlighted", str(state.get("visual_state", "")), "highlight")
	_assert_true("requirements met at rule-derived magic", bool(state.get("requirements_met", false)))
	var requirement_status: Array = state.get("requirement_status", [])
	_assert_eq("requirement row count", requirement_status.size(), 2)
	_assert_eq("magic requirement current", int((requirement_status[1] as Dictionary).get("current", -1)), magic_rank1_req)
	_assert_eq("no hovered skill by default", str(state.get("hovered_skill_id", "")), "")
	_assert_false("tooltip hidden by default", bool(state.get("tooltip_visible", true)))
	panel.bot_hover_skill("magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("hovered skill updates", str(state.get("hovered_skill_id", "")), "magic_bolt")
	_assert_true("tooltip visible on hover", bool(state.get("tooltip_visible", false)))
	_assert_false("points label hidden under tooltip", bool(state.get("points_label_visible", true)))
	var tooltip_position: Dictionary = state.get("tooltip_position", {})
	var magic_block := panel._skill_blocks.get("magic_bolt", null) as Control
	_assert_true("magic block exists for tooltip placement", magic_block != null)
	if magic_block != null:
		_assert_true("tooltip opens under hovered skill", float(tooltip_position.get("y", 0.0)) >= magic_block.position.y + magic_block.size.y)
	_assert_eq("tooltip ignores mouse so skill remains clickable", int(state.get("tooltip_mouse_filter", -1)), Control.MOUSE_FILTER_IGNORE)
	_assert_true("tooltip includes mana from catalog", str(state.get("tooltip_body", "")).contains("Mana: %d" % magic_bolt_mana_cost))
	_assert_false("rank 0 tooltip omits next-rank damage", str(state.get("tooltip_body", "")).contains("Damage: 6-9 -> 7-10"))
	_assert_false("tooltip omits next-rank section header", str(state.get("tooltip_body", "")).contains("Next rank:"))
	_assert_false("rank 0 tooltip omits next-rank rich damage", str(state.get("tooltip_rich_text", "")).contains("Damage: 6-9[color=#6fd66f] -> 7-10[/color]"))
	_assert_true("tooltip separates stats from requirements", str(state.get("tooltip_body", "")).contains("\n\nRequires:\nLevel %d\nMagic %d" % [magic_bolt_level_req, magic_rank1_req]))
	_assert_true("ligthing tooltip includes flat cooldown", SkillsPanelScript.tooltip_plain_body("ligthing", 1, panel.skill_progression, panel.character_progression).contains("Cooldown: attack x2 + 1s"))
	var arcane_focus_tooltip := SkillsPanelScript.tooltip_plain_body("arcane_focus", 0, panel.skill_progression, panel.character_progression)
	_assert_true("passive tooltip includes passive summary", arcane_focus_tooltip.contains("Passive max mana"))
	_assert_true("passive tooltip includes stat effect", arcane_focus_tooltip.contains("Max mana: +8"))
	var arcane_state: Dictionary = (state.get("skills", {}) as Dictionary).get("arcane_focus", {})
	_assert_eq("passive icon label from presentation", str(arcane_state.get("icon_label", "")), "A")
	_assert_eq("passive icon shape from presentation", str(arcane_state.get("icon_shape", "")), "orb_projectile")
	panel.bot_leave_skill_tooltip()
	state = panel.get_debug_state()
	_assert_eq("tooltip leave clears hovered skill", str(state.get("hovered_skill_id", "")), "")
	_assert_false("tooltip hides after leaving tooltip", bool(state.get("tooltip_visible", true)))
	_assert_true("points label returns after tooltip hides", bool(state.get("points_label_visible", false)))
	panel.bot_hover_skill("heal")
	state = panel.get_debug_state()
	_assert_eq("non-class skill selection ignored", str(state.get("selected_skill_id", "")), "magic_bolt")
	_assert_false("non-class skill absent from debug state", (state.get("skills", {}) as Dictionary).has("heal"))
	panel.bot_hover_skill("magic_bolt")

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 5,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": magic_rank1_req},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, true),
	})
	state = panel.get_debug_state()
	_assert_true("requirements met at rule-derived magic", bool(state.get("requirements_met", false)))
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
		"character_class": "paladin",
		"level": 5,
		"base_stats": {"str": 5, "dex": 5, "vit": 8, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, false, 0, false, 0, true, 0, true),
	})
	state = panel.get_debug_state()
	var paladin_skill_ids: Array = state.get("skill_ids", [])
	_assert_eq("paladin has four compact visible skills plus passive column", paladin_skill_ids.size(), 7)
	_assert_eq("paladin first local skill", str(paladin_skill_ids[0]), "charge")
	_assert_eq("paladin second local skill", str(paladin_skill_ids[1]), "heal")
	_assert_eq("paladin third local skill", str(paladin_skill_ids[2]), "holy_shield")
	_assert_eq("paladin passive row 1", str(paladin_skill_ids[3]), "vigilant_guard")
	_assert_eq("paladin passive row 2", str(paladin_skill_ids[4]), "faithful_bulwark")
	_assert_eq("paladin fourth local skill", str(paladin_skill_ids[5]), "sanctuary")
	_assert_eq("paladin passive row 3", str(paladin_skill_ids[6]), "consecrated_vitality")
	var heal_tooltip := SkillsPanelScript.tooltip_plain_body("heal", 1, panel.skill_progression, panel.character_progression)
	_assert_true("heal tooltip includes inline next-rank percent", heal_tooltip.contains("Heal: 25% -> 35%"))
	_assert_false("heal tooltip omits next-rank header", heal_tooltip.contains("Next rank:"))
	var heal_block := panel._skill_blocks.get("heal", null) as Control
	var holy_shield_block := panel._skill_blocks.get("holy_shield", null) as Control
	_assert_true("paladin heal block exists", heal_block != null)
	_assert_true("paladin holy shield block exists", holy_shield_block != null)
	if heal_block != null and holy_shield_block != null:
		_assert_true("paladin third-column skill reflows into compact row", heal_block.position.x <= 240.0)
		_assert_true("paladin fourth-column skill remains inside tree", holy_shield_block.position.x + holy_shield_block.size.x <= 395.0)
	_assert_eq("rankable paladin skill is highlighted", str(((state.get("skills", {}) as Dictionary).get("heal", {}) as Dictionary).get("visual_state", "")), "highlight")

	panel.set_character_progression({
		"character_class": "rogue",
		"level": 1,
		"base_stats": {"str": 5, "dex": 8, "vit": 5, "magic": 5},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(0, false, 0, false, 0, false, 0, false, 0, false, 0, false, 0, false, 0, true, 0, true, 0, false, 0, false, 0, false, 0, false, 0, false, 0, true),
	})
	state = panel.get_debug_state()
	var rogue_skill_ids: Array = state.get("skill_ids", [])
	_assert_eq("rogue has two starter skills plus higher row skills and passive column", rogue_skill_ids.size(), 7)
	_assert_eq("rogue first starter skill", str(rogue_skill_ids[0]), "poison_stab")
	_assert_eq("rogue second starter skill", str(rogue_skill_ids[1]), "dash")
	_assert_eq("rogue passive row 1 visible", str(rogue_skill_ids[2]), "quick_hands")
	_assert_eq("rogue higher-row skill visible", str(rogue_skill_ids[3]), "shadow_flurry")
	_assert_eq("rogue passive skill visible", str(rogue_skill_ids[4]), "executioner")
	_assert_eq("rogue passive row 2 visible", str(rogue_skill_ids[5]), "killer_instinct")
	_assert_eq("rogue passive row 3 visible", str(rogue_skill_ids[6]), "evasive_footwork")
	_assert_eq("rogue skill name from catalog", str(state.get("skill_name", "")), "Poison Stab")
	_assert_eq("rogue skill icon from presentation", str(state.get("icon_label", "")), "P")
	_assert_true("rogue poison stab can spend", bool(((state.get("skills", {}) as Dictionary).get("poison_stab", {}) as Dictionary).get("can_spend", false)))
	_assert_true("rogue dash can spend", bool(((state.get("skills", {}) as Dictionary).get("dash", {}) as Dictionary).get("can_spend", false)))
	_assert_false("rogue shadow flurry gated before dash rank", bool(((state.get("skills", {}) as Dictionary).get("shadow_flurry", {}) as Dictionary).get("can_spend", true)))
	_assert_true("rogue executioner can spend", bool(((state.get("skills", {}) as Dictionary).get("executioner", {}) as Dictionary).get("can_spend", false)))

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": magic_rank1_req},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, false),
	})
	state = panel.get_debug_state()
	_assert_false("rank 2 requirement blocks previous-rank magic", bool(state.get("requirements_met", true)))
	_assert_false("rank 2 spend disabled before rule-derived magic", bool(state.get("spend_button_enabled", true)))
	_assert_eq("bought skill missing next-rank req stays normal", str(state.get("visual_state", "")), "normal")
	_assert_true("tooltip includes rank 2 missing magic diff", str(state.get("tooltip_body", "")).contains("Magic %d(-%d)" % [magic_rank2_req, magic_rank2_req - magic_rank1_req]))

	panel.set_character_progression({
		"character_class": "sorcerer",
		"level": 6,
		"base_stats": {"str": 5, "dex": 5, "vit": 5, "magic": magic_rank2_req},
	})
	panel.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1, true),
	})
	state = panel.get_debug_state()
	_assert_true("rank 2 requirements met at rule-derived magic", bool(state.get("requirements_met", false)))
	_assert_true("rank 2 spend enabled at magic 8", bool(state.get("spend_button_enabled", false)))
	panel.bot_leave_skill_tooltip()
	panel.bot_hover_skill("magic_bolt")
	state = panel.get_debug_state()
	var rank_one_magic_tooltip := str(((state.get("skills", {}) as Dictionary).get("magic_bolt", {}) as Dictionary).get("tooltip_body", ""))
	_assert_true("rank 1 tooltip includes next-rank damage", rank_one_magic_tooltip.contains("Damage: 6-9 -> 7-10"))
	panel.set_skill_bindings(["", "magic_bolt", "", "", "", "", "", "", "heal", "", "", "", "", "", "", ""], "magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("assigned key shown", str(state.get("assigned_key", "")), "F2")
	_assert_true("right click assigned state", bool(state.get("right_click_assigned", false)))
	panel.set_skill_bindings(["", "", "", "", "", "", "", "", "magic_bolt", "", "", "", "", "", "", ""], "magic_bolt")
	state = panel.get_debug_state()
	_assert_eq("secondary assigned key shown", str(state.get("assigned_key", "")), "S-F1")
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


func _skill_rows(magic_rank: int, magic_can_spend: bool, rage_rank: int = 0, rage_can_spend: bool = false, heal_rank: int = 0, heal_can_spend: bool = false, holy_shield_rank: int = 0, holy_shield_can_spend: bool = false, cleave_rank: int = 0, cleave_can_spend: bool = false, ice_shard_rank: int = 0, ice_shard_can_spend: bool = false, ligthing_rank: int = 0, ligthing_can_spend: bool = false, poison_stab_rank: int = 0, poison_stab_can_spend: bool = false, dash_rank: int = 0, dash_can_spend: bool = false, earthbreaker_rank: int = 0, earthbreaker_can_spend: bool = false, shadow_flurry_rank: int = 0, shadow_flurry_can_spend: bool = false, split_arrow_rank: int = 0, split_arrow_can_spend: bool = false, arcane_barrage_rank: int = 0, arcane_barrage_can_spend: bool = false, sanctuary_rank: int = 0, sanctuary_can_spend: bool = false, executioner_rank: int = 0, executioner_can_spend: bool = false) -> Array:
	return [
		{"skill_id": "cleave", "rank": cleave_rank, "max_rank": _skill_max_rank("cleave"), "can_spend": cleave_can_spend},
		{"skill_id": "earthbreaker", "rank": earthbreaker_rank, "max_rank": _skill_max_rank("earthbreaker"), "can_spend": earthbreaker_can_spend},
		{"skill_id": "magic_bolt", "rank": magic_rank, "max_rank": _skill_max_rank("magic_bolt"), "can_spend": magic_can_spend},
		{"skill_id": "ice_shard", "rank": ice_shard_rank, "max_rank": _skill_max_rank("ice_shard"), "can_spend": ice_shard_can_spend},
		{"skill_id": "ligthing", "rank": ligthing_rank, "max_rank": _skill_max_rank("ligthing"), "can_spend": ligthing_can_spend},
		{"skill_id": "arcane_barrage", "rank": arcane_barrage_rank, "max_rank": _skill_max_rank("arcane_barrage"), "can_spend": arcane_barrage_can_spend},
		{"skill_id": "rage", "rank": rage_rank, "max_rank": _skill_max_rank("rage"), "can_spend": rage_can_spend},
		{"skill_id": "heal", "rank": heal_rank, "max_rank": _skill_max_rank("heal"), "can_spend": heal_can_spend},
		{"skill_id": "holy_shield", "rank": holy_shield_rank, "max_rank": _skill_max_rank("holy_shield"), "can_spend": holy_shield_can_spend},
		{"skill_id": "sanctuary", "rank": sanctuary_rank, "max_rank": _skill_max_rank("sanctuary"), "can_spend": sanctuary_can_spend},
		{"skill_id": "poison_stab", "rank": poison_stab_rank, "max_rank": _skill_max_rank("poison_stab"), "can_spend": poison_stab_can_spend},
		{"skill_id": "dash", "rank": dash_rank, "max_rank": _skill_max_rank("dash"), "can_spend": dash_can_spend},
		{"skill_id": "shadow_flurry", "rank": shadow_flurry_rank, "max_rank": _skill_max_rank("shadow_flurry"), "can_spend": shadow_flurry_can_spend},
		{"skill_id": "executioner", "rank": executioner_rank, "max_rank": _skill_max_rank("executioner"), "can_spend": executioner_can_spend},
		{"skill_id": "split_arrow", "rank": split_arrow_rank, "max_rank": _skill_max_rank("split_arrow"), "can_spend": split_arrow_can_spend},
	]


func _skill_max_rank(skill_id: String) -> int:
	return int(SkillRulesLoaderScript.skill_definition(skill_id).get("max_rank", 1))


func _skill_level_requirement(skill_id: String, current_rank: int) -> int:
	var req: Dictionary = SkillRulesLoaderScript.skill_definition(skill_id).get("requirements", {})
	return int(req.get("level", 1)) + current_rank * int(req.get("level_per_rank", 0))


func _skill_stat_requirement(skill_id: String, stat: String, current_rank: int) -> int:
	var req: Dictionary = SkillRulesLoaderScript.skill_definition(skill_id).get("requirements", {})
	var stats: Dictionary = req.get("stats", {})
	var stats_per_rank: Dictionary = req.get("stats_per_rank", {})
	return int(stats.get(stat, 0)) + current_rank * int(stats_per_rank.get(stat, 0))


func _skill_mana_cost(skill_id: String, current_rank: int) -> int:
	var cost: Dictionary = SkillRulesLoaderScript.skill_definition(skill_id).get("cost", {})
	var mana: Dictionary = cost.get("mana", {})
	return int(mana.get("base", 0)) + current_rank * int(mana.get("per_rank", 0))
