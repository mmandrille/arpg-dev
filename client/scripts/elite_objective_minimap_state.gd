class_name EliteObjectiveMinimapState
extends RefCounted

const EDGE_PADDING := 0.10
const WORLD_TO_MAP_SCALE := 0.055


static func from_entities(entities: Dictionary, player_position: Vector3) -> Dictionary:
	for rec in entities.values():
		var row: Dictionary = rec
		if not bool(row.get("elite_objective", false)):
			continue
		var open := str(row.get("state", "")) == "open"
		if open:
			return _hidden("complete")
		var node := row.get("node", null) as Node3D
		if node == null:
			return _hidden("hidden")
		var delta := node.position - player_position
		return {
			"visible": true,
			"has_pin": true,
			"status": "active",
			"pin_x": clampf(0.5 + delta.x * WORLD_TO_MAP_SCALE, EDGE_PADDING, 1.0 - EDGE_PADDING),
			"pin_y": clampf(0.5 + delta.z * WORLD_TO_MAP_SCALE, EDGE_PADDING, 1.0 - EDGE_PADDING),
		}
	return _hidden("hidden")


static func _hidden(status: String) -> Dictionary:
	return {
		"visible": false,
		"has_pin": false,
		"status": status,
		"pin_x": 0.5,
		"pin_y": 0.5,
	}
