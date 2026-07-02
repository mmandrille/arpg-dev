class_name BlacksmithUpgradeChance
extends RefCounted

const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")
const _LOG10 := 2.302585092994046


static func upgrade_target_level(current_level: int) -> int:
	return maxi(0, current_level) + 1


static func effective_success_percent(current_level: int, shard_level: int, min_shard_level: int, curve: Dictionary, shard_bonus_percent_per_tier: int) -> int:
	var target_level := upgrade_target_level(current_level)
	var failure := failure_chance_percent(target_level, curve)
	var success := 100 - failure
	if shard_bonus_percent_per_tier > 0 and shard_level > min_shard_level:
		success += shard_bonus_percent_per_tier * (shard_level - min_shard_level)

	return clamp_percent(success)


static func failure_chance_percent(target_level: int, curve: Dictionary) -> int:
	var safe_max := int(curve.get("safe_target_level_max", 1))
	if target_level <= safe_max:
		return 0
	var anchors: Array = curve.get("level_anchors", [])
	var failures: Array = curve.get("failure_chance_percent_anchors", [])
	if anchors.is_empty() or failures.is_empty() or anchors.size() != failures.size():
		return 0
	if target_level <= int(anchors[0]):
		return clamp_percent(int(failures[0]))
	var last_index := anchors.size() - 1
	if target_level >= int(anchors[last_index]):
		if last_index == 0:
			return clamp_percent(int(failures[0]))
		return clamp_percent(_extrapolate_log_failure(
			target_level,
			int(anchors[last_index - 1]),
			int(failures[last_index - 1]),
			int(anchors[last_index]),
			int(failures[last_index])
		))
	for i in range(1, anchors.size()):
		if target_level > int(anchors[i]):
			continue
		return clamp_percent(_interpolate_log_failure(
			target_level,
			int(anchors[i - 1]),
			int(failures[i - 1]),
			int(anchors[i]),
			int(failures[i])
		))

	return clamp_percent(int(failures[last_index]))


static func _interpolate_log_failure(target: int, left_level: int, left_failure: int, right_level: int, right_failure: int) -> int:
	var left_log := log(float(left_level)) / _LOG10
	var right_log := log(float(right_level)) / _LOG10
	var target_log := log(float(target)) / _LOG10
	if right_log <= left_log:
		return left_failure
	var weight := (target_log - left_log) / (right_log - left_log)
	return int(round(float(left_failure) + weight * float(right_failure - left_failure)))


static func _extrapolate_log_failure(target: int, left_level: int, left_failure: int, right_level: int, right_failure: int) -> int:
	var left_log := log(float(left_level)) / _LOG10
	var right_log := log(float(right_level)) / _LOG10
	var target_log := log(float(target)) / _LOG10
	if right_log <= left_log:
		return right_failure
	var slope := float(right_failure - left_failure) / (right_log - left_log)
	return int(round(float(right_failure) + slope * (target_log - right_log)))


static func clamp_percent(value: int) -> int:
	return clampi(value, 0, 100)


static func staged_shard_level(staged_resource: Dictionary, resource_def_id: String, required_level: int) -> int:
	if staged_resource.is_empty():
		return required_level
	if str(staged_resource.get("item_def_id", "")) != resource_def_id:
		return required_level
	return maxi(required_level, BlacksmithUpgradePreviewScript.shard_level(staged_resource))


static func panel_success_chance(current_level: int, staged_resource: Dictionary, resource_def_id: String, required_level: int, curve: Dictionary, shard_bonus_percent_per_tier: int, fallback_percent: int) -> int:
	if curve.is_empty():
		return fallback_percent
	var shard_level := staged_shard_level(staged_resource, resource_def_id, required_level)
	return BlacksmithUpgradePreviewScript.effective_success_chance_for_item(
		current_level,
		shard_level,
		required_level,
		curve,
		shard_bonus_percent_per_tier,
		fallback_percent
	)
