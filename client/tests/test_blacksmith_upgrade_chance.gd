# v400 blacksmith upgrade chance mirror

extends SceneTree

const BlacksmithUpgradeChanceScript := preload("res://scripts/blacksmith_upgrade_chance.gd")
const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")

var _pass_count: int = 0
var _fail_count: int = 0

const CURVE := {
	"safe_target_level_max": 1,
	"level_anchors": [2, 10],
	"failure_chance_percent_anchors": [10, 75],
}


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_assert_eq("safe target level", BlacksmithUpgradeChanceScript.effective_success_percent(0, 1, 1, CURVE, 10), 100)
	_assert_eq("target level 2 base", BlacksmithUpgradeChanceScript.effective_success_percent(1, 1, 1, CURVE, 10), 90)
	_assert_eq("target level 10 base", BlacksmithUpgradeChanceScript.effective_success_percent(9, 9, 9, CURVE, 10), 25)
	_assert_eq("target level 10 over-tier shard", BlacksmithUpgradeChanceScript.effective_success_percent(9, 12, 9, CURVE, 10), 55)
	var item := {
		"rolled_stats": {"stats": {"item_level": 9}, "requirements": {"level": 9, "str": 40}},
	}
	var lines: Array = BlacksmithUpgradePreviewScript.preview_lines(item, {
		"max_level": 0,
		"deepest_dungeon_depth": 140,
		"item_level_levels_per_tier": 10,
		"failure_curve": CURVE,
		"shard_bonus_percent_per_tier": 10,
		"staged_shard_level": 12,
		"resource_required_level": 9,
		"resource_count": 1,
		"resource_inventory_count": 1,
		"resource_name": "Upgrade Shard",
		"wallet_gold": 500,
		"base_cost": 100,
		"growth_cost": 50,
		"pity_failure_threshold": 2,
	})
	_assert_true("preview shows base success", _contains(lines, "Base success: 25%"))
	_assert_true("preview shows shard bonus", _contains(lines, "Shard bonus: +30%"))
	_assert_true("preview shows effective chance", _contains(lines, "Success chance: 55%"))
	_assert_true("preview shows failure outcome", _contains(lines, "shard consumed"))
	_assert_true("preview shows requirements", _contains(lines, "Current requirements: Level 9, Str 40"))
	print("[gdtest] PASS: test_blacksmith_upgrade_chance (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _contains(lines: Array, text: String) -> bool:
	for line in lines:
		if str(line).contains(text):
			return true
	return false


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: got %s want %s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	push_error("%s: expected true" % label)
