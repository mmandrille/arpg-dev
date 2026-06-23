## Client-only dungeon depth / hub lighting profiles.
##
## Presentation-only mood lighting derived from shared biome palette depth bands.
## Does not change fog, line-of-sight, light-radius gameplay, or server authority.
class_name DungeonDepthLighting
extends RefCounted

const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")

const TOWN_PROFILE := {
	"directional_color": "#fff0dc",
	"directional_energy": 1.05,
	"ambient_color": "#d8dce8",
	"ambient_energy": 0.38,
}

static func profile_for_level(level: int, factory: GroundWallFactory) -> Dictionary:
	if level >= 0:
		return TOWN_PROFILE.duplicate(true)

	var palette: Dictionary = factory.biome_palette_for_level(level) if factory != null else {}
	var fallback_directional := str(palette.get("wall_highlight", "#948b7c"))
	var fallback_ambient := str(palette.get("wall_base", "#393b3e"))

	return {
		"directional_color": str(palette.get("directional_color", fallback_directional)),
		"directional_energy": float(palette.get("directional_energy", 1.0)),
		"ambient_color": str(palette.get("ambient_color", fallback_ambient)),
		"ambient_energy": float(palette.get("ambient_energy", 0.30)),
	}

static func apply_profile(
	profile: Dictionary,
	directional: DirectionalLight3D,
	world_environment: WorldEnvironment,
) -> void:
	if directional != null:
		directional.light_color = _color_from_hex(str(profile.get("directional_color", "#ffffff")))
		directional.light_energy = float(profile.get("directional_energy", 1.0))

	if world_environment == null:
		return

	var environment := world_environment.environment
	if environment == null:
		environment = Environment.new()
		environment.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
		world_environment.environment = environment

	environment.ambient_light_color = _color_from_hex(str(profile.get("ambient_color", "#808080")))
	environment.ambient_light_energy = float(profile.get("ambient_energy", 0.30))

static func apply_for_level(
	level: int,
	directional: DirectionalLight3D,
	world_environment: WorldEnvironment,
	factory: GroundWallFactory,
) -> Dictionary:
	var profile := profile_for_level(level, factory)
	apply_profile(profile, directional, world_environment)

	return profile

static func _color_from_hex(hex: String) -> Color:
	var normalized := hex.strip_edges()
	if not normalized.begins_with("#"):
		normalized = "#" + normalized

	return Color(normalized)
