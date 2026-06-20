# Unit test for the character stats panel.
# Run via: godot --headless --path client --script res://tests/test_character_stats_panel.gd
extends SceneTree

const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	await _test_derived_stats_scroll_and_breakdown_tooltips()
	_test_item_tooltip_ignores_mouse()
	print("[gdtest] PASS: test_character_stats_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_derived_stats_scroll_and_breakdown_tooltips() -> void:
	var panel = CharacterStatsPanelScript.new()
	root.add_child(panel)
	await process_frame
	panel.set_progression({
		"derived_stats": {
			"hit_chance": 0.5,
			"crit_chance": 0.06,
			"block_percent": 75,
			"armor": 6,
		},
		"stat_breakdowns": [
			{
				"key": "block_percent",
				"value": 75,
				"uncapped_value": 82,
				"cap": 75,
				"sources": [
					{"label": "Faithful Bulwark", "value": 15, "kind": "passive_skill"},
					{"label": "Cave Shield", "value": 5, "kind": "equipment_base", "item_instance_id": "1004"},
					{"label": "Rolled block", "value": 7, "kind": "equipment_roll", "item_instance_id": "1004"},
					{"label": "Holy Shield", "value": 60, "kind": "skill_effect"},
					{"label": "Shrine Guard", "value": 2, "kind": "buff"},
					{"label": "Cracked Guard", "value": -3, "kind": "debuff"},
					{"label": "Block cap", "value": -11, "kind": "cap"},
				],
			},
		],
	})
	var state: Dictionary = panel.get_debug_state()
	_assert_true("derived stats visible by default", bool(state.get("derived_open", false)))
	_assert_eq("derived title text", str(state.get("derived_title", "")), "Derived")
	_assert_false("derived title is not a button", bool(state.get("derived_title_is_button", true)))
	var scroll: Dictionary = state.get("derived_scroll", {})
	_assert_true("derived scroll visible", bool(scroll.get("visible", false)))
	_assert_true("derived scrollbar is right-side reserved", bool(scroll.get("scrollbar_on_right", false)))
	_assert_eq("derived row count", int(scroll.get("row_count", 0)), CharacterStatsPanelScript.DERIVED_LABELS.keys().size())
	_assert_eq("horizontal scrolling disabled", int(scroll.get("horizontal_scroll_mode", -1)), ScrollContainer.SCROLL_MODE_DISABLED)
	_assert_eq("vertical scrollbar always shown", int(scroll.get("vertical_scroll_mode", -1)), ScrollContainer.SCROLL_MODE_SHOW_ALWAYS)
	_assert_true("derived viewport fits six rows", float(scroll.get("viewport_height", 0.0)) >= 168.0)
	var labels: Dictionary = state.get("derived_labels", {})
	_assert_eq("hit chance displays percent", str(labels.get("hit_chance", "")), "Hit chance  50%")
	_assert_eq("crit chance displays percent", str(labels.get("crit_chance", "")), "Crit chance  6%")
	_assert_eq("block chance displays percent", str(labels.get("block_percent", "")), "Block  75%")
	var mouse_filters: Dictionary = state.get("derived_mouse_filters", {})
	_assert_eq("block row accepts hover for tooltip", int(mouse_filters.get("block_percent", -1)), Control.MOUSE_FILTER_STOP)
	var tooltip_panel: Dictionary = state.get("derived_tooltip_panel", {})
	_assert_true("derived tooltip uses custom panel", bool(tooltip_panel.get("custom", false)))
	_assert_eq("derived tooltip is opaque", float(tooltip_panel.get("background_alpha", 0.0)), 1.0)
	var tooltips: Dictionary = state.get("derived_tooltips", {})
	var block_tip := str(tooltips.get("block_percent", ""))
	_assert_true("tooltip uses formula title", block_tip.contains("Block formula:"))
	_assert_false("tooltip omits source section", block_tip.contains("Sources:"))
	_assert_false("tooltip omits uncapped section", block_tip.contains("Uncapped:"))
	_assert_true("tooltip includes passive formula source", block_tip.contains("+15% (Faithful Bulwark, Passive skill)"))
	_assert_true("tooltip includes item formula source by name only", block_tip.contains("+5% (Cave Shield)"))
	_assert_true("tooltip includes rolled item formula source by item name", block_tip.contains("+7% (Cave Shield)"))
	_assert_false("tooltip omits item source category", block_tip.contains("Item base"))
	_assert_false("tooltip omits item ids", block_tip.contains("item 1004"))
	_assert_true("tooltip includes skill formula source", block_tip.contains("+60% (Holy Shield, Skill effect)"))
	_assert_true("tooltip includes buff formula source", block_tip.contains("+2% (Shrine Guard, Buff)"))
	_assert_true("tooltip includes debuff formula source", block_tip.contains("-3% (Cracked Guard, Debuff)"))
	_assert_true("tooltip includes cap formula source", block_tip.contains("-11% (Block cap, Cap)"))
	_assert_true("tooltip puts formula elements on separate lines", block_tip.contains("+15% (Faithful Bulwark, Passive skill)\n+5% (Cave Shield)"))
	_assert_true("tooltip includes capped final", block_tip.contains("= 75% (cap 75%)"))
	panel.free()


func _test_item_tooltip_ignores_mouse() -> void:
	var tooltip := ItemTooltipPanelScript.new()
	tooltip.setup({}, {}, ["Training Blade"], [], [])
	_assert_eq("item tooltip ignores mouse", int(tooltip.mouse_filter), Control.MOUSE_FILTER_IGNORE)
	var labels := tooltip.find_children("*", "Label", true, false)
	var first_label := labels[0] as Label if not labels.is_empty() else null
	_assert_true("item tooltip has label", first_label != null)
	if first_label != null:
		_assert_eq("item tooltip label ignores mouse", int(first_label.mouse_filter), Control.MOUSE_FILTER_IGNORE)
	tooltip.free()


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
