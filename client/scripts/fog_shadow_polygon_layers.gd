extends RefCounted

const HeroVisibilityFieldScript := preload("res://scripts/hero_visibility_field.gd")

var _root: Node2D
var _gloom_color: Color
var _core_color: Color
var _soft_edge_color: Color
var _gloom_scale: float
var _soft_edge_scale: float
var _soft_edge_amplitude: float
var _organic_segments: float
var _organic_seed: float
var _soft_edge_polygons: Array[Polygon2D] = []
var _gloom_polygons: Array[Polygon2D] = []
var _core_polygons: Array[Polygon2D] = []


func _init(
	root: Node2D,
	gloom_color: Color,
	core_color: Color,
	soft_edge_color: Color,
	gloom_scale: float,
	soft_edge_scale: float,
	soft_edge_amplitude: float,
	organic_segments: float,
	organic_seed: float,
) -> void:
	_root = root
	_gloom_color = gloom_color
	_core_color = core_color
	_soft_edge_color = soft_edge_color
	_gloom_scale = gloom_scale
	_soft_edge_scale = soft_edge_scale
	_soft_edge_amplitude = soft_edge_amplitude
	_organic_segments = organic_segments
	_organic_seed = organic_seed


func sync(polygons: Array, soft_edges: Array) -> void:
	_ensure_layer_count(polygons.size())
	for i in range(_soft_edge_polygons.size()):
		var soft_edge := _soft_edge_polygons[i]
		if i < polygons.size() and i < soft_edges.size() and bool(soft_edges[i]):
			soft_edge.visible = true
			soft_edge.polygon = PackedVector2Array(HeroVisibilityFieldScript.organic_expanded_polygon(
				polygons[i] as Array,
				_soft_edge_scale,
				_soft_edge_amplitude,
				_organic_segments,
				_organic_seed,
			))
		else:
			_clear(soft_edge)
	for i in range(_gloom_polygons.size()):
		var gloom := _gloom_polygons[i]
		if i < polygons.size():
			gloom.visible = true
			gloom.polygon = PackedVector2Array(HeroVisibilityFieldScript.expanded_polygon(polygons[i] as Array, _gloom_scale))
		else:
			_clear(gloom)
	for i in range(_core_polygons.size()):
		var core := _core_polygons[i]
		if i < polygons.size():
			core.visible = true
			core.polygon = PackedVector2Array(polygons[i])
		else:
			_clear(core)


func hide() -> void:
	for polygon in _soft_edge_polygons + _gloom_polygons + _core_polygons:
		_clear(polygon)


func _ensure_layer_count(count: int) -> void:
	while _soft_edge_polygons.size() < count:
		_soft_edge_polygons.append(_make_polygon(_soft_edge_color, 1))
	while _gloom_polygons.size() < count:
		_gloom_polygons.append(_make_polygon(_gloom_color, 0))
	while _core_polygons.size() < count:
		_core_polygons.append(_make_polygon(_core_color, 2))


func _make_polygon(color: Color, layer: int) -> Polygon2D:
	var polygon := Polygon2D.new()
	polygon.color = color
	polygon.z_index = layer
	_root.add_child(polygon)
	return polygon


func _clear(polygon: Polygon2D) -> void:
	polygon.visible = false
	polygon.polygon = PackedVector2Array()
