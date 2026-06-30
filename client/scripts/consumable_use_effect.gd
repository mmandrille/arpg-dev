## Brief flash + spark ring when a consumable restores HP or mana.
class_name ConsumableUseEffect
extends RefCounted

const LIFETIME := 0.75


static func spawn(parent: Node, world_position: Vector3, restores_hp: bool, restores_mana: bool) -> void:
	if parent == null:
		return
	var root := Node3D.new()
	root.name = "ConsumableUseEffect"
	root.position = world_position + Vector3(0.0, 0.55, 0.0)
	parent.add_child(root)

	var color := Color("#58f09a") if restores_hp else Color("#58c7f0")
	if restores_hp and restores_mana:
		color = Color("#8af0c8")

	var ring := MeshInstance3D.new()
	var torus := TorusMesh.new()
	torus.inner_radius = 0.22
	torus.outer_radius = 0.38
	ring.mesh = torus
	ring.rotation_degrees.x = 90.0
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, 0.72)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 1.9
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	ring.material_override = mat
	root.add_child(ring)

	var tween := root.create_tween()
	tween.set_parallel(true)
	tween.tween_property(ring, "scale", Vector3(2.2, 2.2, 2.2), LIFETIME).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	tween.tween_property(mat, "albedo_color:a", 0.0, LIFETIME)
	tween.chain().tween_callback(root.queue_free)
