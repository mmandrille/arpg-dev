class_name WallOcclusionRuntime
extends RefCounted


static func sync(
	fade,
	camera: Camera3D,
	wall_layout: Array,
	player_anchor: Node3D,
	entities: Dictionary,
	monster_ids: Array,
	active: bool,
) -> void:
	if fade == null or camera == null:
		return
	fade.sync(camera, wall_layout, collect_targets(player_anchor, entities, monster_ids), active)


static func collect_targets(player_anchor: Node3D, entities: Dictionary, monster_ids: Array) -> Array:
	var targets: Array = []
	if player_anchor != null:
		targets.append(player_anchor.global_position)
	for id in monster_ids:
		var rec: Dictionary = entities.get(id, {})
		if int(rec.get("hp", 0)) <= 0:
			continue
		var node = rec.get("node") as Node3D
		if node != null:
			targets.append(node.global_position)

	return targets
