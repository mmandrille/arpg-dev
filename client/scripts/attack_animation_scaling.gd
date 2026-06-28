class_name AttackAnimationScaling
extends RefCounted

const CombatFeelPresentationLoaderScript := preload("res://scripts/combat_feel_presentation_loader.gd")


static func speed_scale_for(attack_speed: float) -> float:
	CombatFeelPresentationLoaderScript.ensure_loaded()
	var cfg := CombatFeelPresentationLoaderScript.attack_animation()
	if not bool(cfg.get("enabled", true)):
		return 1.0
	var baseline := float(cfg.get("baseline_attack_speed", 1.0))
	if baseline <= 0.0:
		return 1.0
	var min_scale := float(cfg.get("min_speed_scale", 0.75))
	var max_scale := float(cfg.get("max_speed_scale", 2.5))
	var speed := attack_speed if attack_speed > 0.0 else baseline

	return clampf(speed / baseline, min_scale, max_scale)
