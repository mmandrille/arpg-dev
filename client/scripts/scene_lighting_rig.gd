## Owns the world key light + WorldEnvironment and keeps them in sync with the current level:
## DungeonDepthLighting sets light color/energy and ambient; RenderEnvironmentPresentation layers
## the ADR-0018 D7 render baseline (shadows, tonemap, SSAO, glow, fog, AA) on top.
## Shared by main.gd and focused capture scripts so screenshots use the runtime lighting path.
class_name SceneLightingRig
extends RefCounted

const DungeonDepthLightingScript := preload("res://scripts/dungeon_depth_lighting.gd")
const RenderPresentationLoaderScript := preload("res://scripts/render_presentation_loader.gd")
const RenderEnvironmentPresentationScript := preload("res://scripts/render_environment_presentation.gd")

var directional: DirectionalLight3D
var world_environment: WorldEnvironment
var _parent: Node3D


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
	RenderEnvironmentPresentationScript.apply(
		RenderPresentationLoaderScript.context(context_id),
		RenderPresentationLoaderScript.quality_tier(quality),
		RenderPresentationLoaderScript.key_light(),
		directional,
		world_environment,
		_viewport(),
	)
	return profile


func _viewport() -> Viewport:
	if _parent == null or not _parent.is_inside_tree():
		return null
	return _parent.get_viewport()
