extends RefCounted


static func make_node(projectile_def_id: String = "") -> Node3D:
	if projectile_def_id == "training_arrow":
		return _make_training_arrow()
	if projectile_def_id == "piercing_shot_projectile":
		return _make_arrow_projectile(Color(0.48, 1.0, 0.56), Color(0.18, 0.88, 0.32), 1.8, true)
	if projectile_def_id == "pinning_shot_projectile":
		return _make_arrow_projectile(Color(0.72, 1.0, 0.42), Color(0.42, 0.92, 0.16), 2.1, false)
	return _make_energy_projectile(projectile_def_id)


static func _make_training_arrow() -> Node3D:
	return _make_arrow_projectile(Color(0.96, 0.98, 1.0), Color(0.82, 0.90, 1.0), 0.35, false)


static func _make_arrow_projectile(color: Color, emission: Color, emission_energy: float, long_head: bool) -> Node3D:
	var root := Node3D.new()
	root.name = "Projectile"
	var mat := _material(color, emission, emission_energy)

	var shaft := MeshInstance3D.new()
	shaft.name = "ArrowShaft"
	var shaft_mesh := BoxMesh.new()
	shaft_mesh.size = Vector3(0.06, 0.06, 0.82 if long_head else 0.66)
	shaft.mesh = shaft_mesh
	shaft.position = Vector3(0.0, 0.35, 0.02)
	shaft.material_override = mat
	root.add_child(shaft)

	var head := MeshInstance3D.new()
	head.name = "ArrowHead"
	var head_mesh := CylinderMesh.new()
	head_mesh.top_radius = 0.0
	head_mesh.bottom_radius = 0.16 if long_head else 0.13
	head_mesh.height = 0.34 if long_head else 0.24
	head_mesh.radial_segments = 4
	head.mesh = head_mesh
	head.position = Vector3(0.0, 0.35, -0.52 if long_head else -0.43)
	head.rotation_degrees.x = -90.0
	head.material_override = mat
	root.add_child(head)

	for side in [-1.0, 1.0]:
		var feather := MeshInstance3D.new()
		feather.name = "ArrowFletching"
		var feather_mesh := BoxMesh.new()
		feather_mesh.size = Vector3(0.04, 0.12, 0.16)
		feather.mesh = feather_mesh
		feather.position = Vector3(0.07 * side, 0.35, 0.38)
		feather.rotation_degrees.z = 28.0 * side
		feather.material_override = mat
		root.add_child(feather)
	if long_head:
		var wake := MeshInstance3D.new()
		wake.name = "PiercingTrail"
		var wake_mesh := BoxMesh.new()
		wake_mesh.size = Vector3(0.03, 0.03, 1.12)
		wake.mesh = wake_mesh
		wake.position = Vector3(0.0, 0.35, 0.16)
		wake.material_override = _material(Color(color.r, color.g, color.b, 0.28), emission, emission_energy * 0.8)
		root.add_child(wake)
	else:
		var pin := MeshInstance3D.new()
		pin.name = "PinningGlow"
		var pin_mesh := TorusMesh.new()
		pin_mesh.inner_radius = 0.11
		pin_mesh.outer_radius = 0.15
		pin_mesh.ring_segments = 24
		pin.mesh = pin_mesh
		pin.position = Vector3(0.0, 0.35, -0.36)
		pin.rotation_degrees.x = 90.0
		pin.material_override = mat
		root.add_child(pin)
	return root


static func _make_energy_projectile(projectile_def_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "Projectile"
	var mesh := BoxMesh.new()
	var color := Color(0.65, 0.90, 1.0)
	var emission := Color(0.25, 0.55, 0.9)
	if projectile_def_id == "ice_shard_projectile":
		mesh.size = Vector3(0.12, 0.12, 0.9)
		color = Color(0.72, 0.95, 1.0)
		emission = Color(0.35, 0.75, 1.0)
	elif projectile_def_id == "ice_shard_shard":
		mesh.size = Vector3(0.08, 0.08, 0.42)
		color = Color(0.86, 0.98, 1.0)
		emission = Color(0.45, 0.85, 1.0)
	elif projectile_def_id == "ligthing":
		mesh.size = Vector3(0.10, 0.10, 0.85)
		color = Color(1.0, 0.94, 0.28)
		emission = Color(1.0, 0.82, 0.12)
	else:
		mesh.size = Vector3(0.16, 0.16, 0.7)

	var shaft := MeshInstance3D.new()
	shaft.name = "EnergyBolt"
	shaft.mesh = mesh
	shaft.position = Vector3(0.0, 0.35, 0.0)
	shaft.material_override = _material(color, emission, 1.0)
	root.add_child(shaft)
	return root


static func _material(color: Color, emission: Color, emission_energy: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.emission_enabled = true
	mat.emission = emission
	mat.emission_energy_multiplier = emission_energy
	return mat
