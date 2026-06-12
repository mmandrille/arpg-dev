extends Node3D
class_name ConsumableHealEffect

const LIFETIME := 1.1
const SPARK_COUNT := 16

var _age := 0.0
var _sparks: Array[MeshInstance3D] = []
var _starts: Array[Vector3] = []
var _material: StandardMaterial3D


func _ready() -> void:
	_material = StandardMaterial3D.new()
	_material.albedo_color = Color(0.45, 1.0, 0.58, 0.85)
	_material.emission_enabled = true
	_material.emission = Color(0.35, 1.0, 0.48)
	_material.emission_energy_multiplier = 1.7
	_material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	for i in range(SPARK_COUNT):
		_make_spark(i)


func _process(delta: float) -> void:
	_age += delta
	if _age >= LIFETIME:
		queue_free()
		return
	var t := _age / LIFETIME
	var fade := 1.0 - smoothstep(0.68, 1.0, t)
	for i in range(_sparks.size()):
		var spark := _sparks[i]
		var lift := t * (0.7 + 0.08 * float(i % 4))
		spark.position = _starts[i] + Vector3(0.0, lift, 0.0)
		spark.scale = Vector3.ONE * (1.0 - 0.35 * t)
		spark.transparency = 1.0 - fade
	_material.albedo_color.a = 0.85 * fade


func _make_spark(index: int) -> void:
	var angle := float(index) * 2.399963
	var radius := 0.28 + 0.22 * float(index % 5) / 4.0
	var spark := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.045
	mesh.height = 0.09
	spark.mesh = mesh
	spark.material_override = _material
	spark.position = Vector3(cos(angle) * radius, 0.35 + 0.04 * float(index % 3), sin(angle) * radius)
	add_child(spark)
	_sparks.append(spark)
	_starts.append(spark.position)
