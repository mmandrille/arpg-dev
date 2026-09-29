## Applies the ADR-0018 D7 render baseline (tonemap, SSAO, glow, fog, adjustments, key-light
## shadows, viewport anti-aliasing) from RenderPresentationLoader data. Pure presentation: it never
## touches light color/energy or ambient, which DungeonDepthLighting owns.
class_name RenderEnvironmentPresentation
extends RefCounted

const TONEMAP_MODES := {
	"linear": Environment.TONE_MAPPER_LINEAR,
	"reinhard": Environment.TONE_MAPPER_REINHARDT,
	"filmic": Environment.TONE_MAPPER_FILMIC,
	"aces": Environment.TONE_MAPPER_ACES,
	"agx": Environment.TONE_MAPPER_AGX,
}
const MSAA_MODES := {
	"disabled": Viewport.MSAA_DISABLED,
	"2x": Viewport.MSAA_2X,
	"4x": Viewport.MSAA_4X,
	"8x": Viewport.MSAA_8X,
}
const SCREEN_SPACE_AA_MODES := {
	"none": Viewport.SCREEN_SPACE_AA_DISABLED,
	"fxaa": Viewport.SCREEN_SPACE_AA_FXAA,
}


static func configure_key_light(light: DirectionalLight3D, key_light: Dictionary) -> void:
	if light == null:
		return
	var rot: Dictionary = key_light.get("rotation_degrees", {})
	light.rotation_degrees = Vector3(float(rot.get("x", -50.0)), float(rot.get("y", -40.0)), float(rot.get("z", 0.0)))
	light.directional_shadow_max_distance = float(key_light.get("shadow_max_distance", light.directional_shadow_max_distance))
	light.shadow_blur = float(key_light.get("shadow_blur", light.shadow_blur))
	light.shadow_bias = float(key_light.get("shadow_bias", light.shadow_bias))
	light.shadow_normal_bias = float(key_light.get("shadow_normal_bias", light.shadow_normal_bias))
	light.shadow_opacity = float(key_light.get("shadow_opacity", light.shadow_opacity))


## `context_cfg` is one catalog context; `tier` gates the expensive features on top of it.
static func apply(
	context_cfg: Dictionary,
	tier: Dictionary,
	key_light: Dictionary,
	light: DirectionalLight3D,
	world_environment: WorldEnvironment,
	viewport: Viewport = null,
) -> void:
	if light != null:
		light.shadow_enabled = bool(key_light.get("shadow_enabled", false)) and bool(tier.get("key_light_shadows", false))
	if viewport != null:
		viewport.msaa_3d = MSAA_MODES.get(str(tier.get("msaa_3d", "disabled")), Viewport.MSAA_DISABLED)
		viewport.screen_space_aa = SCREEN_SPACE_AA_MODES.get(str(tier.get("screen_space_aa", "none")), Viewport.SCREEN_SPACE_AA_DISABLED)
	if world_environment == null:
		return
	var env := world_environment.environment
	if env == null:
		env = Environment.new()
		env.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
		world_environment.environment = env
	_apply_tonemap(env, context_cfg.get("tonemap", {}))
	_apply_ssao(env, context_cfg.get("ssao", {}), bool(tier.get("ssao", false)))
	_apply_glow(env, context_cfg.get("glow", {}), bool(tier.get("glow", false)))
	_apply_fog(env, context_cfg.get("fog", {}), bool(tier.get("fog", false)))
	_apply_adjustments(env, context_cfg.get("adjustments", {}))


static func _apply_tonemap(env: Environment, cfg: Dictionary) -> void:
	env.tonemap_mode = TONEMAP_MODES.get(str(cfg.get("mode", "linear")), Environment.TONE_MAPPER_LINEAR)
	env.tonemap_exposure = float(cfg.get("exposure", 1.0))
	env.tonemap_white = float(cfg.get("white", 1.0))


static func _apply_ssao(env: Environment, cfg: Dictionary, tier_allows: bool) -> void:
	env.ssao_enabled = tier_allows and bool(cfg.get("enabled", false))
	env.ssao_radius = float(cfg.get("radius", env.ssao_radius))
	env.ssao_intensity = float(cfg.get("intensity", env.ssao_intensity))
	env.ssao_power = float(cfg.get("power", env.ssao_power))
	env.ssao_detail = float(cfg.get("detail", env.ssao_detail))


static func _apply_glow(env: Environment, cfg: Dictionary, tier_allows: bool) -> void:
	env.glow_enabled = tier_allows and bool(cfg.get("enabled", false))
	env.glow_intensity = float(cfg.get("intensity", env.glow_intensity))
	env.glow_strength = float(cfg.get("strength", env.glow_strength))
	env.glow_bloom = float(cfg.get("bloom", env.glow_bloom))
	env.glow_hdr_threshold = float(cfg.get("hdr_threshold", env.glow_hdr_threshold))


static func _apply_fog(env: Environment, cfg: Dictionary, tier_allows: bool) -> void:
	env.fog_enabled = tier_allows and bool(cfg.get("enabled", false))
	env.fog_mode = Environment.FOG_MODE_DEPTH if str(cfg.get("mode", "exponential")) == "depth" else Environment.FOG_MODE_EXPONENTIAL
	env.fog_light_color = Color(str(cfg.get("color", "#000000")))
	env.fog_light_energy = float(cfg.get("light_energy", env.fog_light_energy))
	env.fog_density = float(cfg.get("density", env.fog_density))
	env.fog_sky_affect = float(cfg.get("sky_affect", env.fog_sky_affect))
	env.fog_depth_begin = float(cfg.get("depth_begin", env.fog_depth_begin))
	env.fog_depth_end = float(cfg.get("depth_end", env.fog_depth_end))


static func _apply_adjustments(env: Environment, cfg: Dictionary) -> void:
	env.adjustment_enabled = bool(cfg.get("enabled", false))
	env.adjustment_brightness = float(cfg.get("brightness", 1.0))
	env.adjustment_contrast = float(cfg.get("contrast", 1.0))
	env.adjustment_saturation = float(cfg.get("saturation", 1.0))
