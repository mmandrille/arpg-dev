## Data-backed skill rank scaling helpers (mirrors Go rankScaledInt/rankScaledFloat).
class_name SkillRankScaling
extends RefCounted

const PROGRESSION_REL := "../shared/rules/character_progression.v0.json"

static var _progression: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join(PROGRESSION_REL)
	if not FileAccess.file_exists(path):
		_progression = {}
		return
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	_progression = parsed if typeof(parsed) == TYPE_DICTIONARY else {}


static func rank_scaled_int(base: int, per_rank: int, rank: int, curve: Dictionary = {}) -> int:
	var safe_rank := maxi(1, rank)
	var curve_type := str(curve.get("type", "compound_percent"))
	var pct := int(curve.get("percent_per_rank", 8))
	match curve_type:
		"linear":
			return maxi(0, base + per_rank * maxi(0, safe_rank - 1))
		_:
			var factor := pow(1.0 + float(pct) / 100.0, float(safe_rank - 1))
			return maxi(0, int(round(float(base) * factor + float(per_rank) * float(safe_rank - 1))))


static func rank_scaled_float(base: float, per_rank: float, rank: int, curve: Dictionary = {}) -> float:
	var safe_rank := maxi(1, rank)
	var curve_type := str(curve.get("type", "compound_percent"))
	var pct := int(curve.get("percent_per_rank", 8))
	match curve_type:
		"linear":
			return maxf(0.0, base + per_rank * float(maxi(0, safe_rank - 1)))
		_:
			var factor := pow(1.0 + float(pct) / 100.0, float(safe_rank - 1))
			return maxf(0.0, base * factor + per_rank * float(safe_rank - 1))


static func progression_rank_curve() -> Dictionary:
	ensure_loaded()
	var curve: Variant = _progression.get("skill_rank_scaling", {})
	if typeof(curve) != TYPE_DICTIONARY:
		return {"type": "compound_percent", "percent_per_rank": 8}
	return curve


static func progression_mana_curve() -> Dictionary:
	ensure_loaded()
	var curve: Variant = _progression.get("skill_mana_scaling", {})
	if typeof(curve) != TYPE_DICTIONARY:
		return {"type": "compound_percent", "percent_per_rank": 10}
	return curve
