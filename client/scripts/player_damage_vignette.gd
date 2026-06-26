class_name PlayerDamageVignette
extends Control

const DECAY_SPEED := 5.5
const MAX_ALPHA := 0.42

static var _instance: PlayerDamageVignette


static func attach(parent: Node) -> void:
	if _instance != null and is_instance_valid(_instance):
		return
	var layer := CanvasLayer.new()
	layer.layer = 3
	layer.name = "PlayerDamageVignetteLayer"
	parent.add_child(layer)
	_instance = PlayerDamageVignette.new()
	_instance.name = "PlayerDamageVignette"
	layer.add_child(_instance)


static func pulse(damage: int, max_hp: int) -> void:
	if _instance == null or not is_instance_valid(_instance):
		return
	_instance._apply_pulse(damage, max_hp)


static func reset_session() -> void:
	if _instance != null and is_instance_valid(_instance):
		_instance._strength = 0.0
		_instance._sync_alpha()


static func debug_strength() -> float:
	if _instance == null or not is_instance_valid(_instance):
		return 0.0
	return _instance._strength


var _strength: float = 0.0
var _edges: Array[ColorRect] = []


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	set_process(true)
	_build_edges()


func _process(delta: float) -> void:
	if _strength <= 0.0:
		return
	_strength = maxf(0.0, _strength - delta * DECAY_SPEED)
	_sync_alpha()


func _apply_pulse(damage: int, max_hp: int) -> void:
	if damage <= 0:
		return
	var ratio := float(damage) / float(maxi(1, max_hp))
	_strength = clampf(_strength + ratio * 0.55 + 0.12, 0.0, 1.0)
	_sync_alpha()


func _sync_alpha() -> void:
	var alpha := _strength * MAX_ALPHA
	var color := Color(0.55, 0.08, 0.05, alpha)
	for edge in _edges:
		edge.color = color


func _build_edges() -> void:
	_edges = []
	for side in ["top", "bottom", "left", "right"]:
		var edge := ColorRect.new()
		edge.mouse_filter = Control.MOUSE_FILTER_IGNORE
		edge.color = Color(0.55, 0.08, 0.05, 0.0)
		match side:
			"top":
				edge.set_anchors_preset(Control.PRESET_TOP_WIDE)
				edge.offset_bottom = 72.0
			"bottom":
				edge.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
				edge.offset_top = -72.0
			"left":
				edge.set_anchors_preset(Control.PRESET_LEFT_WIDE)
				edge.offset_right = 88.0
			"right":
				edge.set_anchors_preset(Control.PRESET_RIGHT_WIDE)
				edge.offset_left = -88.0
		add_child(edge)
		_edges.append(edge)
