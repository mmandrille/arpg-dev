## Heal skill area rain (v492): glow motes from CombatVfx falling over the heal radius plus a fading
## ground ring. Keeps the pre-v492 API (setup(radius), LIFETIME, self-free) that main.gd and the
## heal-rain tests use.
extends Node3D
class_name HealRainEffect

const LIFETIME := 3.0
const DEFAULT_RADIUS := 4.0
const RING_COLOR := Color(0.35, 1.0, 0.5)

var radius := DEFAULT_RADIUS
var _age := 0.0
var _rain: GPUParticles3D
var _ring_material: StandardMaterial3D


func setup(effect_radius: float = DEFAULT_RADIUS) -> void:
	radius = max(0.8, effect_radius)


func _ready() -> void:
	_make_ground_ring()
	_rain = CombatVfx.make_rain(radius)
	if _rain != null:
		add_child(_rain)


func _process(delta: float) -> void:
	_age += delta
	if _age >= LIFETIME:
		queue_free()
		return
	if _rain != null and _rain.emitting and _age >= LIFETIME - _rain.lifetime:
		_rain.emitting = false  # let the last motes land before the node frees
	var fade := 1.0 - smoothstep(0.66, 1.0, _age / LIFETIME)
	_ring_material.albedo_color.a = 0.85 * fade


func _make_ground_ring() -> void:
	_ring_material = StandardMaterial3D.new()
	_ring_material.albedo_color = Color(RING_COLOR, 0.85)
	_ring_material.emission_enabled = true
	_ring_material.emission = RING_COLOR
	_ring_material.emission_energy_multiplier = 2.2
	_ring_material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	_ring_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	var ring := MeshInstance3D.new()
	ring.name = "HealRainRing"
	var mesh := TorusMesh.new()
	mesh.inner_radius = max(0.05, radius - 0.05)
	mesh.outer_radius = radius + 0.05
	mesh.ring_segments = 96
	ring.mesh = mesh
	ring.material_override = _ring_material
	ring.position.y = 0.04
	add_child(ring)
