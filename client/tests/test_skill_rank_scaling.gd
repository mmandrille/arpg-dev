extends SceneTree

const SkillRankScalingScript := preload("res://scripts/skill_rank_scaling.gd")


func _initialize() -> void:
	_test_compound_rank_scaling_increases()
	_test_matches_go_golden_sample()
	_finish()


func _test_compound_rank_scaling_increases() -> void:
	var curve := {"type": "compound_percent", "percent_per_rank": 8}
	var prev := SkillRankScalingScript.rank_scaled_int(100, 10, 1, curve)
	for rank in range(2, 11):
		var next := SkillRankScalingScript.rank_scaled_int(100, 10, rank, curve)
		if next <= prev:
			push_error("rank %d value %d should exceed rank %d value %d" % [rank, next, rank - 1, prev])
			return
		prev = next
	print("[gdtest] PASS: compound rank scaling increases")


func _test_matches_go_golden_sample() -> void:
	SkillRankScalingScript.ensure_loaded()
	var rank_curve := SkillRankScalingScript.progression_rank_curve()
	var mana_curve := SkillRankScalingScript.progression_mana_curve()
	if int(rank_curve.get("percent_per_rank", 0)) != 8:
		push_error("rank curve percent = %s, want 8" % str(rank_curve.get("percent_per_rank")))
		return
	if int(mana_curve.get("percent_per_rank", 0)) != 10:
		push_error("mana curve percent = %s, want 10" % str(mana_curve.get("percent_per_rank")))
		return
	if SkillRankScalingScript.rank_scaled_int(300, 50, 2, rank_curve) != 374:
		push_error("magic bolt rank 2 min sample mismatch")
		return
	print("[gdtest] PASS: progression curve sample")


func _finish() -> void:
	quit(0)
