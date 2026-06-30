## Short radial glow burst when the local player levels up (presentation-only).
class_name LevelUpBurst
extends RefCounted

const BURST_SECONDS := 0.55


static func spawn(parent: Node, world_position: Vector3) -> void:
	if parent == null:
		return
	var root := Node3D.new()
	root.name = "LevelUpBurst"
	root.position = world_position + Vector3(0.0, 0.35, 0.0)
	parent.add_child(root)

	var ring := MeshInstance3D.new()
	var torus := TorusMesh.new()
	torus.inner_radius = 0.35
	torus.outer_radius = 0.95
	ring.mesh = torus
	ring.rotation_degrees.x = 90.0
	var ring_mat := StandardMaterial3D.new()
	ring_mat.albedo_color = Color(1.0, 0.92, 0.42, 0.55)
	ring_mat.emission_enabled = true
	ring_mat.emission = Color("#ffe08a")
	ring_mat.emission_energy_multiplier = 2.2
	ring_mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	ring_mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	ring.material_override = ring_mat
	root.add_child(ring)

	var pillar := MeshInstance3D.new()
	var cyl := CylinderMesh.new()
	cyl.top_radius = 0.55
	cyl.bottom_radius = 0.15
	cyl.height = 1.35
	pillar.mesh = cyl
	pillar.position = Vector3(0.0, 0.65, 0.0)
	var pillar_mat := StandardMaterial3D.new()
	pillar_mat.albedo_color = Color(1.0, 0.96, 0.62, 0.28)
	pillar_mat.emission_enabled = true
	pillar_mat.emission = Color("#fff0a8")
	pillar_mat.emission_energy_multiplier = 1.8
	pillar_mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	pillar_mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	pillar.material_override = pillar_mat
	root.add_child(pillar)

	var tween := root.create_tween()
	tween.set_parallel(true)
	tween.tween_property(ring, "scale", Vector3(2.4, 2.4, 2.4), BURST_SECONDS).set_trans(Tween.TRANS_EXPO).set_ease(Tween.EASE_OUT)
	tween.tween_property(pillar, "scale", Vector3(1.6, 1.6, 1.6), BURST_SECONDS).set_trans(Tween.TRANS_EXPO).set_ease(Tween.EASE_OUT)
	tween.tween_property(ring_mat, "albedo_color:a", 0.0, BURST_SECONDS)
	tween.tween_property(pillar_mat, "albedo_color:a", 0.0, BURST_SECONDS)
	tween.chain().tween_callback(root.queue_free)
