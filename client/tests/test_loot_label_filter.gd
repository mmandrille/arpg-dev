# Headless unit tests for LootLabelFilter (v153).
#
# Pure display-filter logic: rarity ladder ordering, per-mode thresholds,
# cycle wraparound, and off-ladder rarities always allowed. No scene tree.
#
# Run via: godot --headless --path client --script res://tests/test_loot_label_filter.gd
extends SceneTree

const LootLabelFilterScript := preload("res://scripts/loot_label_filter.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	_test_default_mode_allows_all()
	_test_cycle_thresholds()
	_test_cycle_wraps_around()
	_test_restore_mode_label()
	_test_restore_invalid_mode_falls_back()
	_test_off_ladder_always_allowed()
	_test_case_insensitive()
	_test_display_color_dims_unhighlighted()

	if _fail == 0:
		print("[gdtest] PASS: test_loot_label_filter (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_loot_label_filter (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _check(cond: bool, msg: String) -> void:
	if cond:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + msg)
		print("  FAIL: ", msg)


func _test_default_mode_allows_all() -> void:
	var f := LootLabelFilterScript.new()
	_check(f.mode_label() == "All", "default mode is All")
	_check(not f.is_active(), "default mode is inactive")
	for rarity in ["common", "magic", "rare", "unique"]:
		_check(f.allows(rarity), "All allows %s" % rarity)


func _test_cycle_thresholds() -> void:
	var f := LootLabelFilterScript.new()
	f.cycle()  # Magic+
	_check(f.mode_label() == "Magic+", "after one cycle = Magic+")
	_check(f.is_active(), "Magic+ mode is active")
	_check(not f.allows("common"), "Magic+ hides common")
	_check(f.allows("magic") and f.allows("rare") and f.allows("unique"), "Magic+ allows magic and above")

	f.cycle()  # Rare+
	_check(f.mode_label() == "Rare+", "after two cycles = Rare+")
	_check(not f.allows("common") and not f.allows("magic"), "Rare+ hides common and magic")
	_check(f.allows("rare") and f.allows("unique"), "Rare+ allows rare and unique")

	f.cycle()  # Unique
	_check(f.mode_label() == "Unique", "after three cycles = Unique")
	_check(not f.allows("rare"), "Unique hides rare")
	_check(f.allows("unique"), "Unique allows unique")


func _test_cycle_wraps_around() -> void:
	var f := LootLabelFilterScript.new()
	for _i in range(4):
		f.cycle()
	_check(f.mode_label() == "All", "four cycles wraps back to All")
	_check(f.allows("common"), "wrapped All allows common again")


func _test_restore_mode_label() -> void:
	var f := LootLabelFilterScript.new()
	f.set_mode_label("Rare+")
	_check(f.mode_label() == "Rare+", "restore Rare+ mode")
	_check(not f.allows("magic"), "restored Rare+ hides magic")
	_check(f.allows("rare"), "restored Rare+ allows rare")


func _test_restore_invalid_mode_falls_back() -> void:
	var f := LootLabelFilterScript.new()
	f.set_mode_label("Unique")
	f.set_mode_label("legendary")
	_check(f.mode_label() == "All", "invalid restore falls back to All")
	_check(f.allows("common"), "invalid fallback allows common")


func _test_off_ladder_always_allowed() -> void:
	var f := LootLabelFilterScript.new()
	for _i in range(3):  # Unique (strictest)
		f.cycle()
	for rarity in ["currency", "quest", "consumable", "", "mythic_unknown"]:
		_check(f.allows(rarity), "strictest mode still allows off-ladder %s" % rarity)


func _test_case_insensitive() -> void:
	var f := LootLabelFilterScript.new()
	f.cycle()  # Magic+
	_check(not f.allows("COMMON"), "Magic+ hides COMMON (case-insensitive)")
	_check(f.allows("Unique"), "Magic+ allows Unique (case-insensitive)")


func _test_display_color_dims_unhighlighted() -> void:
	var f := LootLabelFilterScript.new()
	var base := Color(0.8, 0.6, 0.4, 1.0)
	_check(f.display_color(base, true) == base, "highlighted label keeps full color")
	var dimmed: Color = f.display_color(base, false)
	_check(dimmed != base, "unhighlighted label is dimmed")
	_check(is_equal_approx(dimmed.r, base.r * 0.58), "dim factor applied to r")
	_check(is_equal_approx(dimmed.a, base.a), "alpha preserved when dimming")
