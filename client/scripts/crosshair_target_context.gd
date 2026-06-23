## CrosshairTargetContext — narrow data bag for CrosshairTargetSystem.
class_name CrosshairTargetContext
extends RefCounted

var camera: Camera3D
var player_anchor: Node3D
var entities: Dictionary
var inventory: Array
var equipped: Dictionary
var aim_reticle  # AimReticleOverlay or null
var target_in_reach: Callable       ## (target_id: String) -> bool
var revive_hover_enabled: Callable    ## () -> bool
var nearest_loot_at_ground: Callable  ## (ground: Vector3) -> String
var center_ground_point: Callable     ## () -> Vector3
var ray_pick_entity: Callable         ## optional (viewport, world) -> String; empty uses center ray


static func make(
	cam: Camera3D,
	anchor: Node3D,
	entity_map: Dictionary,
	inv: Array,
	equip: Dictionary,
	reticle,
	in_reach: Callable,
	revive_hover: Callable,
	nearest_loot: Callable,
	ground_at_center: Callable,
	ray_pick: Callable = Callable()
) -> RefCounted:
	var ctx: RefCounted = load("res://scripts/crosshair_target_context.gd").new()
	ctx.camera = cam
	ctx.player_anchor = anchor
	ctx.entities = entity_map
	ctx.inventory = inv
	ctx.equipped = equip
	ctx.aim_reticle = reticle
	ctx.target_in_reach = in_reach
	ctx.revive_hover_enabled = revive_hover
	ctx.nearest_loot_at_ground = nearest_loot
	ctx.center_ground_point = ground_at_center
	ctx.ray_pick_entity = ray_pick
	return ctx
