extends Node3D
class_name HealRainEffect

const LIFETIME := 1.15
const DROP_COUNT := 34
const DEFAULT_RADIUS := 4.0
const FALL_HEIGHT := 3.2

var radius := DEFAULT_RADIUS
var _age := 0.0
var _drops: Array[Node3D] = []
var _starts: Array[Vector3] = []
var _speeds: Array[float] = []
var _material: StandardMaterial3D


func setup(effect_radius: float = DEFAULT_RADIUS) -> void:
	radius = max(0.8, effect_radius)


func _ready() -> void:
	_material = StandardMaterial3D.new()
	_material.albedo_color = Color(0.24, 1.0, 0.42, 0.78)
	_material.emission_enabled = true
	_material.emission = Color(0.16, 1.0, 0.35)
	_material.emission_energy_multiplier = 1.45
	_material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	_make_ground_ring()
	for i in range(DROP_COUNT):
		_make_drop(i)


func _process(delta: float) -> void:
	_age += delta
	if _age >= LIFETIME:
		queue_free()
		return
	var t := _age / LIFETIME
	var fade := 1.0 - smoothstep(0.66, 1.0, t)
	for i in range(_drops.size()):
		var drop := _drops[i]
		var speed := _speeds[i]
		var fall := fmod(_age * speed + float(i) * 0.13, 1.0)
		drop.position = _starts[i] + Vector3(0.0, FALL_HEIGHT * (1.0 - fall), 0.0)
		drop.scale = Vector3.ONE * (1.0 + 0.18 * (1.0 - fall))
		for child in drop.get_children():
			if child is GeometryInstance3D:
				(child as GeometryInstance3D).transparency = 1.0 - fade
	_material.albedo_color.a = 0.78 * fade


func _make_drop(index: int) -> void:
	var angle := float(index) * 2.399963
	var distance := radius * sqrt((float(index % DROP_COUNT) + 0.5) / float(DROP_COUNT))
	var start := Vector3(cos(angle) * distance, 0.25, sin(angle) * distance)
	var drop := Node3D.new()
	drop.rotation.y = angle + PI * 0.25
	var vertical := MeshInstance3D.new()
	var vertical_mesh := BoxMesh.new()
	vertical_mesh.size = Vector3(0.06, 0.50, 0.06)
	vertical.mesh = vertical_mesh
	vertical.material_override = _material
	drop.add_child(vertical)
	var horizontal := MeshInstance3D.new()
	var horizontal_mesh := BoxMesh.new()
	horizontal_mesh.size = Vector3(0.36, 0.06, 0.06)
	horizontal.mesh = horizontal_mesh
	horizontal.material_override = _material
	drop.add_child(horizontal)
	drop.position = start + Vector3(0.0, FALL_HEIGHT, 0.0)
	add_child(drop)
	_drops.append(drop)
	_starts.append(start)
	_speeds.append(0.86 + 0.05 * float(index % 7))


func _make_ground_ring() -> void:
	var ring := MeshInstance3D.new()
	var mesh := TorusMesh.new()
	mesh.inner_radius = max(0.05, radius - 0.06)
	mesh.outer_radius = radius + 0.06
	mesh.ring_segments = 96
	ring.mesh = mesh
	ring.material_override = _material
	ring.position.y = 0.04
	add_child(ring)
