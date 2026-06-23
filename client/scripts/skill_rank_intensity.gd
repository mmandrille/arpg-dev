## Data-backed skill rank presentation intensity helpers.
class_name SkillRankIntensity
extends RefCounted

const CAST_BURST_NAME := "SkillCastBurst"


static func resolve(presentation: Dictionary, rank: int) -> Dictionary:
	var cfg: Dictionary = presentation.get("rank_intensity", {}) if typeof(presentation.get("rank_intensity", {})) == TYPE_DICTIONARY else {}
	var safe_rank := maxi(0, rank)
	return {
		"accent_width": float(cfg.get("accent_width_base", 2.0)) + float(cfg.get("accent_width_per_rank", 0.35)) * float(safe_rank),
		"glow_ring_count": int(cfg.get("glow_ring_count_base", 0)) + int(cfg.get("glow_ring_count_per_rank", 1)) * safe_rank,
		"cast_burst_scale": float(cfg.get("cast_burst_scale_base", 0.85)) + float(cfg.get("cast_burst_scale_per_rank", 0.12)) * float(safe_rank),
	}


static func spawn_cast_burst(anchor: Node3D, skill_id: String, rank: int) -> void:
	if anchor == null or skill_id == "":
		return
	var presentation := SkillRulesLoader.skill_presentation(skill_id)
	var intensity := resolve(presentation, rank)
	var burst := MeshInstance3D.new()
	burst.name = CAST_BURST_NAME
	var torus := TorusMesh.new()
	var scale := float(intensity.get("cast_burst_scale", 1.0))
	torus.inner_radius = 0.42 * scale
	torus.outer_radius = 0.62 * scale
	burst.mesh = torus
	burst.position = Vector3(0.0, 0.08, 0.0)
	var icon: Dictionary = presentation.get("icon", {})
	var color := Color(str(icon.get("accent", "#e8f7ff")))
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, 0.34)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.9 + float(rank) * 0.12
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	burst.material_override = mat
	anchor.add_child(burst)
	var timer := anchor.get_tree().create_timer(0.28)
	timer.timeout.connect(func() -> void:
		if is_instance_valid(burst):
			burst.queue_free()
	)
