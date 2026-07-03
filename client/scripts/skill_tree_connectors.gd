class_name SkillTreeConnectors
extends Control

const SkillTreeLayoutScript := preload("res://scripts/skill_tree_layout.gd")

const MET_COLOR := Color(0.45, 0.40, 0.32, 0.95)
const UNMET_COLOR := Color(0.18, 0.17, 0.16, 0.75)
const LINE_WIDTH := 4.0

var _edges: Array = []
var _origin := SkillTreeLayoutScript.DEFAULT_ORIGIN
var _spacing := SkillTreeLayoutScript.DEFAULT_SPACING
var _block_size := SkillTreeLayoutScript.DEFAULT_BLOCK_SIZE


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE


func configure(origin: Vector2, spacing: Vector2, block_size: Vector2) -> void:
	_origin = origin
	_spacing = spacing
	_block_size = block_size
	queue_redraw()


func set_edges(edges: Array) -> void:
	_edges = edges.duplicate(true)
	queue_redraw()


func get_debug_state() -> Dictionary:
	var out: Array = []
	for edge in _edges:
		if typeof(edge) != TYPE_DICTIONARY:
			continue
		var rec := edge as Dictionary
		out.append({
			"from": str(rec.get("from", "")),
			"to": str(rec.get("to", "")),
			"met": bool(rec.get("met", false)),
		})

	return {"connections": out}


func _draw() -> void:
	for edge in _edges:
		if typeof(edge) != TYPE_DICTIONARY:
			continue
		var rec := edge as Dictionary
		var from_id := str(rec.get("from", ""))
		var to_id := str(rec.get("to", ""))
		if from_id == "" or to_id == "":
			continue
		var met := bool(rec.get("met", false))
		var color := MET_COLOR if met else UNMET_COLOR
		var points := _orthogonal_path(from_id, to_id)
		if points.size() < 2:
			continue
		for i in range(points.size() - 1):
			draw_line(points[i], points[i + 1], color, LINE_WIDTH, true)


func _orthogonal_path(from_id: String, to_id: String) -> PackedVector2Array:
	var from_center := SkillTreeLayoutScript.block_center(from_id, _origin, _spacing, _block_size)
	var to_center := SkillTreeLayoutScript.block_center(to_id, _origin, _spacing, _block_size)
	var from_bottom := Vector2(from_center.x, from_center.y + _block_size.y * 0.5)
	var to_top := Vector2(to_center.x, to_center.y - _block_size.y * 0.5)
	if is_equal_approx(from_bottom.x, to_top.x):
		return PackedVector2Array([from_bottom, to_top])
	var elbow := Vector2(to_top.x, from_bottom.y)

	return PackedVector2Array([from_bottom, elbow, to_top])
