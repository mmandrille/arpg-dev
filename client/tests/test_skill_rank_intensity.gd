extends SceneTree

const SkillRankIntensityScript := preload("res://scripts/skill_rank_intensity.gd")
const SkillIconScript := preload("res://scripts/skill_icon.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_resolve_scales_with_rank()
	_test_icon_draw_state()
	_finish()


func _test_resolve_scales_with_rank() -> void:
	var presentation := {
		"rank_intensity": {
			"accent_width_base": 2.0,
			"accent_width_per_rank": 0.5,
			"glow_ring_count_base": 0,
			"glow_ring_count_per_rank": 1,
			"cast_burst_scale_base": 0.8,
			"cast_burst_scale_per_rank": 0.2,
		},
	}
	var low: Dictionary = SkillRankIntensityScript.resolve(presentation, 1)
	var high: Dictionary = SkillRankIntensityScript.resolve(presentation, 4)
	if float(high.get("accent_width", 0.0)) <= float(low.get("accent_width", 0.0)):
		_fail("accent width should grow with rank")
		return
	if int(high.get("glow_ring_count", 0)) <= int(low.get("glow_ring_count", 0)):
		_fail("glow ring count should grow with rank")
		return
	_pass("rank intensity resolves")


func _test_icon_draw_state() -> void:
	SkillRulesLoader.ensure_loaded()
	var icon := SkillIconScript.new()
	icon.size = Vector2(48.0, 48.0)
	icon.configure("magic_bolt", SkillRulesLoader.skill_presentation("magic_bolt"), 3)
	if icon.glow_ring_count < 2:
		_fail("icon should inherit rank glow rings")
		return
	if icon.accent_width <= 2.0:
		_fail("icon accent width should increase with rank")
		return
	_pass("icon rank intensity state")


func _pass(_label: String) -> void:
	_pass_count += 1


func _fail(label: String) -> void:
	_fail_count += 1
	push_error(label)


func _finish() -> void:
	if _fail_count == 0:
		print("[gdtest] PASS: test_skill_rank_intensity (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(0)
		return
	print("[gdtest] FAIL: test_skill_rank_intensity (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1)
