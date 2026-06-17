# Headless tests for market item comparison.
# Run via: godot --headless --path client --script res://tests/test_market_item_comparison.gd
extends SceneTree

const MarketItemComparisonScript := preload("res://scripts/market_item_comparison.gd")
const MarketPanelScript := preload("res://scripts/market_panel.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_comparison_against_equipped_item()
	await _test_market_panel_debug_and_tooltip_comparison()
	if _fail == 0:
		print("[gdtest] PASS: test_market_item_comparison (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_market_item_comparison (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_comparison_against_equipped_item() -> void:
	var listed := _mail_listing("listing-1", 6)
	var equipped_item := {
		"item_instance_id": "equipped-chest",
		"item_def_id": "cave_mail",
		"item_template_id": "cave_mail",
		"slot": "chest",
		"rolled_stats": {"armor": 3},
	}
	var comparison := MarketItemComparisonScript.comparison_for_item(listed, [equipped_item], {"chest": "equipped-chest"})
	_assert_eq("comparison slot", str(comparison.get("slot", "")), "chest")
	_assert_eq("equipped id", str(comparison.get("equipped_item_instance_id", "")), "equipped-chest")
	var lines := MarketItemComparisonScript.text_lines_for_comparison(comparison)
	_assert_true("comparison line includes armor delta", _array_contains_text(lines, "+3 Armor vs equipped"))

	var no_equipped := MarketItemComparisonScript.comparison_for_item(listed, [], {})
	_assert_true("no equipped still compares against zero", _array_contains_text(MarketItemComparisonScript.text_lines_for_comparison(no_equipped), "+6 Armor vs equipped"))


func _test_market_panel_debug_and_tooltip_comparison() -> void:
	var panel := MarketPanelScript.new()
	root.add_child(panel)
	await process_frame
	var listed := _mail_listing("listing-1", 6)
	var equipped_item := {
		"item_instance_id": "equipped-chest",
		"item_def_id": "cave_mail",
		"item_template_id": "cave_mail",
		"slot": "chest",
		"rolled_stats": {"armor": 3},
	}
	panel.show_market("market-1", [listed], [equipped_item], "me", "Active listings", {"chest": "equipped-chest"})
	var rows: Array = panel.get_debug_state().get("listing_rows", [])
	_assert_eq("debug row count", rows.size(), 1)
	var row: Dictionary = rows[0] if not rows.is_empty() else {}
	_assert_eq("debug comparison count", int(row.get("comparison_count", 0)), 1)
	_assert_true("debug comparison text", _array_contains_text(row.get("comparison_lines", []), "+3 Armor vs equipped"))
	var tooltip := panel._make_item_tooltip(listed)
	_assert_true("tooltip comparison text", _array_contains_text(tooltip.debug_main_line_font_sizes(), "+3 Armor vs equipped"))
	tooltip.queue_free()
	panel.queue_free()


func _mail_listing(listing_id: String, armor: int) -> Dictionary:
	return {
		"listing_id": listing_id,
		"item_def_id": "cave_mail",
		"item_template_id": "cave_mail",
		"display_name": "Cave Mail",
		"seller_account_id": "seller",
		"price_gold": 63,
		"rolled_stats": {"armor": armor},
	}


func _array_contains_text(values: Array, needle: String) -> bool:
	for value in values:
		if typeof(value) == TYPE_DICTIONARY:
			if str((value as Dictionary).get("text", "")).find(needle) >= 0:
				return true
		elif str(value).find(needle) >= 0:
			return true
	return false


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + label)


func _assert_eq(label: String, got, want) -> void:
	_assert_true("%s got=%s want=%s" % [label, str(got), str(want)], got == want)
