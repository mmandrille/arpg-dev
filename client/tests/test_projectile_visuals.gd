# Unit tests for projectile presentation meshes.
#
# Run via: godot --headless --path client --script res://tests/test_projectile_visuals.gd
extends SceneTree

const ProjectileVisualsScript := preload("res://scripts/projectile_visuals.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_training_arrow_is_white_arrow_shape()
	_test_magic_bolt_remains_blue_energy_bolt()

	print("[gdtest] PASS: test_projectile_visuals (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_training_arrow_is_white_arrow_shape() -> void:
	var arrow := ProjectileVisualsScript.make_node("training_arrow")
	_assert_true("arrow shaft exists", arrow.find_child("ArrowShaft", false, false) != null)
	_assert_true("arrow head exists", arrow.find_child("ArrowHead", false, false) != null)
	_assert_true("arrow fletching exists", arrow.find_child("ArrowFletching", false, false) != null)
	var shaft := arrow.find_child("ArrowShaft", false, false) as MeshInstance3D
	var mat := shaft.material_override as StandardMaterial3D
	_assert_true("arrow material is white", mat.albedo_color.r > 0.9 and mat.albedo_color.g > 0.9 and mat.albedo_color.b > 0.9)
	arrow.free()


func _test_magic_bolt_remains_blue_energy_bolt() -> void:
	var bolt := ProjectileVisualsScript.make_node("magic_bolt_projectile")
	_assert_true("magic bolt has no arrow head", bolt.find_child("ArrowHead", false, false) == null)
	var energy := bolt.find_child("EnergyBolt", false, false) as MeshInstance3D
	_assert_true("magic bolt energy exists", energy != null)
	var mat := energy.material_override as StandardMaterial3D
	_assert_true("magic bolt remains blue", mat.albedo_color.b > mat.albedo_color.r)
	bolt.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
