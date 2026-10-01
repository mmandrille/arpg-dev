## Combat VFX (ADR-0018 P5, v492): one-shot GPUParticles3D bursts drawn with the in-repo
## vfx_soft_glow shader, tuned by shared/assets/vfx_presentation.v0.json. Skill-specific hit bursts,
## generic hit sparks, and death bursts come from GameplayFeedbackPresentation reactions; HealRainEffect
## uses the same builder.
## Particle counts scale with the active graphics tier (render_presentation quality_tiers
## particle_scale), which SceneLightingRig.sync hands over.
class_name CombatVfx
extends RefCounted

const RenderPresentationLoaderScript := preload("res://scripts/render_presentation_loader.gd")
const GLOW_SHADER := preload("res://shaders/vfx_soft_glow.gdshader")
const DEFAULT_PATH := "../shared/assets/vfx_presentation.v0.json"
const BURST_PREFIX := "Vfx_"

static var _loaded: bool = false
static var _config: Dictionary = {}
static var _quality: String = ""


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_config = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	if typeof(parsed) == TYPE_DICTIONARY:
		_config = parsed
	else:
		push_warning("CombatVfx: cannot read %s" % path)


static func set_quality(quality: String) -> void:
	_quality = quality


static func effect(effect_id: String) -> Dictionary:
	ensure_loaded()
	var raw = (_config.get("effects", {}) as Dictionary).get(effect_id, {})
	return (raw as Dictionary).duplicate(true) if typeof(raw) == TYPE_DICTIONARY else {}


## A mapped skill gets a dedicated hit burst. Unknown skills retain the generic hit spark.
## Death always uses its own effect, even when the event also carries a skill_id.
static func reaction_effect_id(ev: Dictionary, reaction_name: String) -> String:
	if reaction_name == "death":
		return "death_burst"
	if reaction_name != "hit":
		return "hit_spark"
	ensure_loaded()
	var raw_mappings = _config.get("skill_hit_effects", {})
	if typeof(raw_mappings) == TYPE_DICTIONARY:
		var effect_id := str((raw_mappings as Dictionary).get(str(ev.get("skill_id", "")), ""))
		if not effect(effect_id).is_empty():
			return effect_id
	return "hit_spark"


## Start/end colours for a damage type, or the effect's own when the type is unknown.
static func colors_for(effect_id: String, damage_type: String) -> Array:
	ensure_loaded()
	var cfg := effect(effect_id)
	var by_type: Dictionary = _config.get("damage_type_colors", {})
	var entry = by_type.get(damage_type, null)
	var source: Dictionary = entry if typeof(entry) == TYPE_DICTIONARY else cfg
	return [Color(str(source.get("color", "#ffffff"))), Color(str(source.get("end_color", "#ffffff")))]


static func particle_scale() -> float:
	var tier := RenderPresentationLoaderScript.quality_tier(_quality)
	return float(tier.get("particle_scale", 1.0))


static func scaled_amount(amount: int) -> int:
	return maxi(1, int(round(float(amount) * particle_scale())))


## A one-shot burst. `direction` (world, may be zero) aims the spray; colours override the catalog.
static func make_burst(effect_id: String, direction: Vector3 = Vector3.UP, colors: Array = []) -> GPUParticles3D:
	var cfg := effect(effect_id)
	if cfg.is_empty():
		return null
	var particles := GPUParticles3D.new()
	particles.name = BURST_PREFIX + effect_id
	particles.one_shot = true
	particles.amount = scaled_amount(int(cfg.get("amount", 8)))
	particles.lifetime = float(cfg.get("lifetime", 0.5))
	particles.explosiveness = float(cfg.get("explosiveness", 1.0))
	particles.local_coords = false
	particles.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
	particles.process_material = _process_material(cfg, direction, colors)
	particles.draw_pass_1 = _quad(float(cfg.get("glow_energy", 2.0)))
	particles.emitting = true
	particles.finished.connect(particles.queue_free)
	return particles


## Continuous rain (HealRainEffect): the heal_rain effect emitted from a disc-sized box at
## spawn_height over `radius`, falling. The caller stops emission and frees it.
static func make_rain(radius: float) -> GPUParticles3D:
	var cfg := effect("heal_rain")
	if cfg.is_empty():
		return null
	var particles := GPUParticles3D.new()
	particles.name = BURST_PREFIX + "heal_rain"
	particles.amount = scaled_amount(int(cfg.get("amount", 64)))
	particles.lifetime = float(cfg.get("lifetime", 1.0))
	particles.preprocess = particles.lifetime  # the column is already falling on the first frame
	particles.explosiveness = float(cfg.get("explosiveness", 0.0))
	particles.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
	var mat := _process_material(cfg, Vector3.DOWN, [])
	mat.emission_shape = ParticleProcessMaterial.EMISSION_SHAPE_BOX
	mat.emission_box_extents = Vector3(radius * 0.72, 0.15, radius * 0.72)
	particles.process_material = mat
	particles.draw_pass_1 = _quad(float(cfg.get("glow_energy", 2.0)))
	particles.position = Vector3(0.0, float(cfg.get("spawn_height", 4.0)), 0.0)
	particles.visibility_aabb = AABB(Vector3(-radius, -float(cfg.get("spawn_height", 4.0)) - 1.0, -radius), Vector3(radius * 2.0, float(cfg.get("spawn_height", 4.0)) + 2.0, radius * 2.0))
	particles.emitting = true
	return particles


static func _process_material(cfg: Dictionary, direction: Vector3, colors: Array) -> ParticleProcessMaterial:
	var mat := ParticleProcessMaterial.new()
	var dir := direction.normalized() if direction.length_squared() > 0.0001 else Vector3.UP
	mat.direction = dir
	mat.spread = float(cfg.get("spread_degrees", 45.0))
	mat.initial_velocity_min = float(cfg.get("velocity_min", 1.0))
	mat.initial_velocity_max = float(cfg.get("velocity_max", 2.0))
	mat.gravity = Vector3(0.0, float(cfg.get("gravity", -9.8)), 0.0)
	mat.damping_min = float(cfg.get("damping", 0.0))
	mat.damping_max = float(cfg.get("damping", 0.0))
	mat.scale_min = float(cfg.get("size_min", 0.1))
	mat.scale_max = float(cfg.get("size_max", 0.1))
	var start: Color = colors[0] if colors.size() >= 2 else Color(str(cfg.get("color", "#ffffff")))
	var end: Color = colors[1] if colors.size() >= 2 else Color(str(cfg.get("end_color", "#ffffff")))
	var gradient := Gradient.new()
	gradient.set_color(0, start)
	gradient.set_color(1, Color(end.r, end.g, end.b, 0.0))
	var ramp := GradientTexture1D.new()
	ramp.gradient = gradient
	mat.color_ramp = ramp
	return mat


static func _quad(glow_energy: float) -> QuadMesh:
	var quad := QuadMesh.new()
	var shader_mat := ShaderMaterial.new()
	shader_mat.shader = GLOW_SHADER
	shader_mat.set_shader_parameter("glow_energy", glow_energy)
	quad.material = shader_mat
	return quad


## Hit sparks (away from the source, coloured by damage_type) or a death burst at `world_pos`, added to
## `parent` so the burst outlives the entity node on death.
static func spawn_for_reaction(parent: Node, world_pos: Vector3, source_pos: Vector3, ev: Dictionary, reaction_name: String) -> GPUParticles3D:
	if parent == null:
		return null
	var effect_id := reaction_effect_id(ev, reaction_name)
	var cfg := effect(effect_id)
	if cfg.is_empty():
		return null
	var direction := Vector3.UP
	var surface_offset_direction := Vector3.ZERO
	if bool(cfg.get("direction_from_source", false)):
		var away := world_pos - source_pos
		away.y = 0.0
		if away.length_squared() > 0.0001:
			direction = away.normalized() + Vector3.UP * 0.6
			if source_pos.length_squared() < 1.0e20:
				surface_offset_direction = away.normalized()
	var colors := colors_for(effect_id, str(ev.get("damage_type", ""))) if effect_id == "hit_spark" else []
	var burst := make_burst(effect_id, direction, colors)
	burst.position = world_pos + Vector3(0.0, float(cfg.get("spawn_height", 0.0)), 0.0) \
		+ surface_offset_direction * float(cfg.get("spawn_surface_offset", 0.0))
	parent.add_child(burst)
	return burst
