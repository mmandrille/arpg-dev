# Direct-import test for the extracted tooltip content + pricing modules (v515).
# Deliberately does not preload inventory_panel.gd.
# Run via: godot --headless --path client --script res://tests/test_inventory_tooltip_content.gd
extends SceneTree

const Content := preload("res://scripts/inventory_tooltip_content.gd")
const Pricing := preload("res://scripts/inventory_item_pricing.gd")
const Styles := preload("res://scripts/inventory_panel_styles.gd")
const TooltipPanel := preload("res://scripts/item_tooltip_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	ItemRulesLoader.ensure_loaded()
	_test_header_and_rarity_text()
	_test_requirements_and_comparison()
	_test_set_and_consumable_lines()
	_test_hotbar_assignment()
	_test_pricing_is_display_only_and_nonnegative()
	_test_rarity_border_styles()
	_test_tooltip_header_hierarchy()
	_test_theme_rarity_colors_drive_loot_labels()
	print("[gdtest] PASS: test_inventory_tooltip_content (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _texts(lines: Array) -> Array:
	var out: Array = []
	for line in lines:
		out.append(str((line as Dictionary).get("text", "")) if typeof(line) == TYPE_DICTIONARY else str(line))
	return out


func _has(lines: Array, needle: String) -> bool:
	for text in _texts(lines):
		if str(text).contains(needle):
			return true
	return false


func _ctx(hotbar: Array = []) -> Content.Context:
	return Content.Context.new(hotbar)


func _test_header_and_rarity_text() -> void:
	for rarity in ["common", "magic", "rare", "unique", "set"]:
		var item := {"item_instance_id": "x", "item_def_id": "amulet", "display_name": "Thing", "rarity": rarity}
		var lines := Content.tooltip_lines(item, _ctx())
		_assert_true("%s name first" % rarity, str(_texts(lines)[0]) == "Thing")
		_assert_true("%s rarity word present" % rarity, _has(lines, rarity.capitalize()))
		_assert_true("%s name color is rarity color" % rarity, (lines[0] as Dictionary).get("color") == Content.rarity_color(rarity))
	_assert_true("unknown rarity falls back to common color", Content.rarity_color("bogus") == Content.rarity_color("common"))


func _test_requirements_and_comparison() -> void:
	var item := {
		"item_def_id": "boots", "rarity": "magic", "slot": "boots",
		"requirement_status": [{"stat": "str", "required": 14, "current": 12, "met": false}],
		"comparison": {"deltas": [{"stat": "max_hp", "delta": 3}, {"stat": "armor", "delta": -1.5}]},
		"equip_preview": {"deltas": [{"stat": "armor", "delta": 2}]},
	}
	var req := Content.requirement_lines(item)
	_assert_true("unmet requirement row", req.size() == 1 and (req[0] as Dictionary).get("color") == Content.requirement_color(false))
	var cmp := Content.comparison_entries(item)
	_assert_true("three comparison rows", cmp.size() == 3)
	_assert_true("positive delta uses gain color", _has(cmp, "+3") and (cmp[1] as Dictionary).get("color") == Content.comparison_color(3.0))
	_assert_true("negative delta uses loss color", (cmp[2] as Dictionary).get("color") == Content.comparison_color(-1.5))
	_assert_true("summary text includes requirement", Content.tooltip_text(item, _ctx()).contains("Requires"))
	_assert_true("level requirement fallback", Content.requirement_lines({"requirements": {"level": 5}}).has("Level 5"))


func _test_set_and_consumable_lines() -> void:
	var set_item := {"item_def_id": "gloves", "rarity": "set", "slot": "gloves", "summary_lines": ["Slot: gloves", "Set: Vanguard (1/3)", "2-piece set bonus: +5 life (inactive)"]}
	var lines := Content.tooltip_lines(set_item, _ctx())
	_assert_true("set membership split", _has(lines, "Set: Vanguard") and _has(lines, "(1/3)"))
	_assert_true("set bonus split", _has(lines, "+5 life (inactive)"))
	_assert_true("potion effect", _has(Content.tooltip_lines({"item_def_id": "blue_potion"}, _ctx()), "Restores 3 mana"))


func _test_hotbar_assignment() -> void:
	var item := {"item_instance_id": "p1", "item_def_id": "health_potion", "rarity": "common"}
	_assert_false("unassigned has no hotbar line", _has(Content.tooltip_lines(item, _ctx()), "hotbar"))
	var hotbar := [{"slot_index": 0, "item_instance_id": "p1"}, {"slot_index": 2, "item_instance_id": "p1"}]
	_assert_true("assigned hotbar line", _has(Content.tooltip_lines(item, _ctx(hotbar)), "Assigned to hotbar: 1, 3"))
	_assert_true("slots helper", Content.hotbar_slots_for_item(hotbar, "p1") == [0, 2])
	_assert_true("empty id has no slots", Content.hotbar_slots_for_item(hotbar, "").is_empty())


func _test_pricing_is_display_only_and_nonnegative() -> void:
	_assert_true("explicit sell price wins", Pricing.item_gold_value({"sell_price": 7}) == 7)
	_assert_true("negative clamps to zero", Pricing.item_gold_value({"value": -4}) == 0)
	_assert_true("unknown item has no price", Pricing.item_gold_value({"item_def_id": "no_such_item"}) == -1)


func _test_theme_rarity_colors_drive_loot_labels() -> void:
	# World loot labels read the same UiTheme tokens as the tooltips (single source).
	for rarity in ["common", "magic", "rare", "unique", "set"]:
		_assert_true("%s theme color is the loot label color" % rarity, Styles.rarity_color(rarity) == UiTheme.color("inventory_rarity_" + rarity))
		_assert_true("%s theme border token resolved" % rarity, UiTheme.has_token("spacing", "inventory_border_" + rarity) or UiTheme.spacing("inventory_border_" + rarity) > 0)


func _luminance(c: Color) -> float:
	var ch: Array = []
	for v in [c.r, c.g, c.b]:
		ch.append(v / 12.92 if v <= 0.03928 else pow((v + 0.055) / 1.055, 2.4))
	return 0.2126 * ch[0] + 0.7152 * ch[1] + 0.0722 * ch[2]


func _contrast(a: Color, b: Color) -> float:
	var la := _luminance(a)
	var lb := _luminance(b)
	return (maxf(la, lb) + 0.05) / (minf(la, lb) + 0.05)


func _test_rarity_border_styles() -> void:
	var widths := {}
	for rarity in ["common", "magic", "rare", "unique", "set"]:
		var style: StyleBoxFlat = Styles.item_slot_style(rarity, false)
		widths[rarity] = style.border_width_left
		_assert_true("%s border >= 3:1 against slot background" % rarity, _contrast(style.border_color, style.bg_color) >= 3.0)
		_assert_true("%s hover border stays >= 3:1" % rarity, _contrast(Styles.item_slot_style(rarity, true).border_color, Styles.item_slot_style(rarity, true).bg_color) >= 3.0)
	_assert_true("common is thinnest", widths["common"] < widths["magic"] and widths["magic"] <= widths["rare"] and widths["rare"] < widths["unique"])
	_assert_true("unknown rarity is treated as common weight", Styles.rarity_border_width("bogus") == widths["common"])
	var invalid: StyleBoxFlat = Styles.item_slot_style("common", false, true)
	_assert_true("invalid requirement stays red", invalid.border_color.r > 0.55 and invalid.border_color.g < 0.45)


func _test_tooltip_header_hierarchy() -> void:
	var item := {"item_def_id": "amulet", "display_name": "Thing", "rarity": "rare"}
	var tooltip := TooltipPanel.new()
	tooltip.setup(item, {}, Content.tooltip_lines(item, _ctx()), [], [], 12)
	root.add_child(tooltip)
	_assert_true("header rule present", tooltip.find_child("HeaderRule", true, false) != null)
	var name_label := tooltip.find_children("*", "Label", true, false)[0] as Label
	_assert_true("name larger than body", name_label.get_theme_font_size("font_size") > TooltipPanel.BODY_FONT_SIZE)
	var rarity_label: Label = null
	for node in tooltip.find_children("*", "Label", true, false):
		if (node as Label).text == "Rarity: Rare":
			rarity_label = node
	_assert_true("rarity word still in tooltip, rarity colored", rarity_label != null and rarity_label.get_theme_color("font_color") == Styles.rarity_color("rare"))
	_assert_true("price footer kept", tooltip.debug_gold_value_text() == "12 gold")
	tooltip.queue_free()


func _assert_true(label: String, cond: bool) -> void:
	if cond:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("FAIL: " + label)
		print("FAIL: " + label)


func _assert_false(label: String, cond: bool) -> void:
	_assert_true(label, not cond)
