## Owns the world key light + WorldEnvironment and keeps them in sync with the current level:
## DungeonDepthLighting sets light color/energy and ambient; RenderEnvironmentPresentation layers
## the ADR-0018 D7 render baseline (shadows, tonemap, SSAO, glow, fog, AA) on top.
## Shared by main.gd and focused capture scripts so screenshots use the runtime lighting path.
class_name SceneLightingRig
extends RefCounted

const DungeonDepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const RenderPresentationLoaderScript := preload("res://scripts/render_presentation_loader.gd")
const RenderEnvironmentPresentationScript := preload("res://scripts/render_environment_presentation.gd")
const DungeonDepthMoodLoaderScript := preload("res://scripts/dungeon_depth_mood_loader.gd")

var directional: DirectionalLight3D
var world_environment: WorldEnvironment
var _parent: Node3D
# Private cache: consumers pass this config to read-only presentation application code.
var _has_cached_dungeon_context := false
var _cached_dungeon_palette_id := ""
var _cached_dungeon_context_cfg: Dictionary = {}


func attach(parent: Node3D) -> void:
	_parent = parent
	directional = DirectionalLight3D.new()
	directional.name = "KeyLight"
	RenderEnvironmentPresentationScript.configure_key_light(directional, RenderPresentationLoaderScript.key_light())
	parent.add_child(directional)
	world_environment = WorldEnvironment.new()
	world_environment.name = "WorldEnvironment"
	parent.add_child(world_environment)


## Returns the DungeonDepthLighting profile that was applied (for debug/tests).
func sync(
	level: int,
	factory: GroundWallFactory,
	quality: String,
	suppress_ambient: bool = false,
	fog_suppression: Dictionary = {},
	town_fog_active: bool = false,
) -> Dictionary:
	var profile := DungeonDepthLightingScript.apply_for_level(
		level, directional, world_environment, factory, suppress_ambient, fog_suppression, town_fog_active
	)
	CombatVfx.set_quality(quality)  # v492: particle counts follow the graphics tier
	var context_id := RenderPresentationLoaderScript.context_for_level(level, town_fog_active)
	var context_cfg: Dictionary
	if level < 0:
		context_cfg = _dungeon_context_for_palette_id(str(profile.get("palette_id", "")))
	else:
		context_cfg = RenderPresentationLoaderScript.context(context_id)
	RenderEnvironmentPresentationScript.apply(
		context_cfg,
		RenderPresentationLoaderScript.quality_tier(quality),
		RenderPresentationLoaderScript.key_light(),
		directional,
		world_environment,
		_viewport(),
	)
	return profile


func _dungeon_context_for_palette_id(palette_id: String) -> Dictionary:
	if not _has_cached_dungeon_context or palette_id != _cached_dungeon_palette_id:
		var base_context := RenderPresentationLoaderScript.context(RenderPresentationLoaderScript.CONTEXT_DUNGEON)
		_cached_dungeon_context_cfg = DungeonDepthMoodLoaderScript.context_with_fog(base_context, palette_id)
		_cached_dungeon_palette_id = palette_id
		_has_cached_dungeon_context = true
	return _cached_dungeon_context_cfg


func _viewport() -> Viewport:
	if _parent == null or not _parent.is_inside_tree():
		return null
	return _parent.get_viewport()
