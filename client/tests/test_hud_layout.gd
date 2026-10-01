# Pure-math tests for HudLayout (v517): no overlap, in-bounds, globes never intrude into the slot cluster.
extends SceneTree

const SIZES := [Vector2(1024, 576), Vector2(1280, 720), Vector2(1600, 900), Vector2(1920, 1080), Vector2(2560, 1440)]
const BOSS_TOP := 58.0
const BOSS_WIDTH := 450.0
const BOSS_HEIGHT := 88.0

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	for vp in SIZES:
		_check_viewport(vp)
	_check_legacy_positions()
	print("[gdtest] PASS: test_hud_layout (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _rects(vp: Vector2) -> Dictionary:
	return {
		"health_globe": HudLayout.health_globe_rect(vp),
		"mana_globe": HudLayout.mana_globe_rect(vp),
		"identity": HudLayout.identity_rect(vp),
		"character_slot": HudLayout.character_slot_rect(vp),
		"hotbar": HudLayout.hotbar_rect(vp),
		"xp_bar": HudLayout.xp_bar_rect(vp),
		"skill_slot": HudLayout.skill_slot_rect(vp),
		"minimap": HudLayout.minimap_compact_rect(vp),
		"boss_bar": HudLayout.boss_bar_rect(vp, BOSS_TOP, BOSS_WIDTH, BOSS_HEIGHT),
	}


func _check_viewport(vp: Vector2) -> void:
	var rects := _rects(vp)
	var screen := Rect2(Vector2.ZERO, vp)
	var names: Array = rects.keys()
	for name in names:
		var rect: Rect2 = rects[name]
		_assert_true("%s in bounds at %s" % [name, str(vp)], screen.encloses(rect))
		_assert_true("%s has area at %s" % [name, str(vp)], rect.size.x > 0.0 and rect.size.y > 0.0)
	for i in range(names.size()):
		for j in range(i + 1, names.size()):
			var a: Rect2 = rects[names[i]]
			var b: Rect2 = rects[names[j]]
			_assert_true("%s does not overlap %s at %s" % [names[i], names[j], str(vp)], not a.intersects(b))
	var d := HudLayout.globe_diameter(vp)
	_assert_true("globe diameter within limits at %s" % str(vp), d >= HudLayout.GLOBE_MIN and d <= HudLayout.GLOBE_MAX)
	_assert_true("globe cluster gap at %s" % str(vp), HudLayout.health_globe_rect(vp).end.x < HudLayout.cluster_rect(vp).position.x)
	_assert_true("mana globe cluster gap at %s" % str(vp), HudLayout.mana_globe_rect(vp).position.x > HudLayout.cluster_rect(vp).end.x)


# The slot cluster must keep the positions the panels used before the helper existed.
func _check_legacy_positions() -> void:
	var vp := Vector2(1920, 1080)
	_assert_eq("hotbar legacy x", HudLayout.hotbar_rect(vp).position.x, (vp.x - 580.0) * 0.5)
	_assert_eq("hotbar legacy y", HudLayout.hotbar_rect(vp).position.y, vp.y - 78.0)
	_assert_eq("character slot legacy x", HudLayout.character_slot_rect(vp).position.x, vp.x * 0.5 - 366.0)
	_assert_eq("skill slot legacy x", HudLayout.skill_slot_rect(vp).position.x, vp.x * 0.5 + 302.0)
	_assert_eq("xp bar legacy y", HudLayout.xp_bar_rect(vp).position.y, vp.y - 12.0)


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
