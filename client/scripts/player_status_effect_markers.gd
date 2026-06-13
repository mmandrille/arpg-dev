extends RefCounted

const HOLY_SHIELD_EFFECT_ID := "holy_shield"
const RAGE_EFFECT_ID := "rage"
const ICE_SLOW_EFFECT_ID := "ice_slow"
const BURNING_EFFECT_ID := "everburning_wound"
const ELITE_COMMAND_EFFECT_ID := "elite_command"

const HOLY_SHIELD_MARKER_NAME := "HolyShieldEffect"
const RAGE_MARKER_NAME := "RageVisualEffect"
const BURNING_MARKER_NAME := "BurningVisualEffect"
const ELITE_COMMAND_MARKER_NAME := "EliteCommandVisualEffect"
const ELITE_COMMAND_RADIUS_PREVIEW_NAME := "EliteCommandRadiusPreview"
const HOLY_SHIELD_AURA_PULSE_NAME := "HolyShieldAuraPulse"
const HOLY_SHIELD_TARGET_PULSE_NAME := "HolyShieldTargetPulse"
const AURA_PULSE_SECONDS := 0.30


static func sync_holy_shield_effect(root: Node3D, effect_ids_value) -> void:
	if root == null:
		return
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	var active := effect_ids.has(HOLY_SHIELD_EFFECT_ID)
	var existing := root.find_child(HOLY_SHIELD_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_holy_shield_effect()
		root.add_child(existing)


static func has_holy_shield_effect(root: Node3D) -> bool:
	return root != null and root.find_child(HOLY_SHIELD_MARKER_NAME, false, false) != null


static func sync_rage_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(RAGE_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_rage_effect()
		root.add_child(existing)


static func has_rage_effect(root: Node3D) -> bool:
	return root != null and root.find_child(RAGE_MARKER_NAME, false, false) != null


static func has_ice_slow_effect(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(ICE_SLOW_EFFECT_ID)


static func has_burning_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(BURNING_EFFECT_ID)


static func has_elite_command_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(ELITE_COMMAND_EFFECT_ID)


static func sync_burning_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(BURNING_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_burning_effect()
		root.add_child(existing)


static func has_burning_effect(root: Node3D) -> bool:
	return root != null and root.find_child(BURNING_MARKER_NAME, false, false) != null


static func sync_elite_command_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(ELITE_COMMAND_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_elite_command_effect()
		root.add_child(existing)


static func has_elite_command_effect(root: Node3D) -> bool:
	return root != null and root.find_child(ELITE_COMMAND_MARKER_NAME, false, false) != null


static func sync_elite_command_radius_preview(root: Node3D, active: bool, radius: float) -> void:
	if root == null:
		return
	var existing := root.find_child(ELITE_COMMAND_RADIUS_PREVIEW_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	var safe_radius := maxf(radius, 0.5)
	if existing == null:
		existing = make_elite_command_radius_preview(safe_radius)
		root.add_child(existing)
	existing.set_meta("radius", safe_radius)
	var ring := existing.find_child("EliteCommandRadiusRing", false, false) as MeshInstance3D
	if ring != null:
		ring.scale = Vector3.ONE * safe_radius


static func has_elite_command_radius_preview(root: Node3D) -> bool:
	return root != null and root.find_child(ELITE_COMMAND_RADIUS_PREVIEW_NAME, false, false) != null


static func elite_command_radius_preview_value(root: Node3D) -> float:
	if root == null:
		return 0.0
	var existing := root.find_child(ELITE_COMMAND_RADIUS_PREVIEW_NAME, false, false) as Node3D
	if existing == null:
		return 0.0
	return float(existing.get_meta("radius", 0.0))


static func pulse_holy_shield_aura(root: Node3D, affected_roots: Array, radius: float) -> void:
	if root == null:
		return
	var pulse := make_holy_shield_aura_pulse(maxf(radius, 0.5))
	root.add_child(pulse)
	_fade_and_free(pulse)
	for affected in affected_roots:
		var affected_root := affected as Node3D
		if affected_root == null:
			continue
		var target_pulse := make_holy_shield_target_pulse()
		affected_root.add_child(target_pulse)
		_fade_and_free(target_pulse)


static func make_holy_shield_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = HOLY_SHIELD_MARKER_NAME
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.82, 0.26, 0.54)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.74, 0.22)
	mat.emission_energy_multiplier = 1.25
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var ring := MeshInstance3D.new()
	ring.name = "HolyShieldRing"
	var ring_mesh := TorusMesh.new()
	ring_mesh.inner_radius = 0.64
	ring_mesh.outer_radius = 0.76
	ring_mesh.ring_segments = 72
	ring.mesh = ring_mesh
	ring.position.y = 0.06
	ring.material_override = mat
	marker.add_child(ring)

	var column := MeshInstance3D.new()
	column.name = "HolyShieldShine"
	var column_mesh := CylinderMesh.new()
	column_mesh.top_radius = 0.52
	column_mesh.bottom_radius = 0.52
	column_mesh.height = 1.45
	column_mesh.radial_segments = 36
	column.mesh = column_mesh
	column.position.y = 0.72
	column.material_override = mat
	marker.add_child(column)
	return marker


static func make_holy_shield_aura_pulse(radius: float) -> Node3D:
	var marker := Node3D.new()
	marker.name = HOLY_SHIELD_AURA_PULSE_NAME
	var mat := _holy_shield_pulse_material(0.36)
	var disc := MeshInstance3D.new()
	disc.name = "HolyShieldAuraDisc"
	var disc_mesh := CylinderMesh.new()
	disc_mesh.top_radius = radius
	disc_mesh.bottom_radius = radius
	disc_mesh.height = 0.05
	disc_mesh.radial_segments = 72
	disc.mesh = disc_mesh
	disc.position.y = 0.05
	disc.material_override = mat
	marker.add_child(disc)
	return marker


static func make_holy_shield_target_pulse() -> Node3D:
	var marker := Node3D.new()
	marker.name = HOLY_SHIELD_TARGET_PULSE_NAME
	var mat := _holy_shield_pulse_material(0.92)
	var shine := MeshInstance3D.new()
	shine.name = "HolyShieldTargetShine"
	var shine_mesh := CylinderMesh.new()
	shine_mesh.top_radius = 0.82
	shine_mesh.bottom_radius = 1.02
	shine_mesh.height = 1.85
	shine_mesh.radial_segments = 36
	shine.mesh = shine_mesh
	shine.position.y = 0.86
	shine.material_override = mat
	marker.add_child(shine)
	return marker


static func make_rage_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = RAGE_MARKER_NAME
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.16, 0.05, 0.66)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.10, 0.02)
	mat.emission_energy_multiplier = 1.9
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var ring := MeshInstance3D.new()
	ring.name = "RageRing"
	var ring_mesh := TorusMesh.new()
	ring_mesh.inner_radius = 0.82
	ring_mesh.outer_radius = 1.02
	ring_mesh.ring_segments = 72
	ring.mesh = ring_mesh
	ring.position.y = 0.08
	ring.material_override = mat
	marker.add_child(ring)

	for i in range(3):
		var flame := MeshInstance3D.new()
		flame.name = "RageFlame%d" % i
		var flame_mesh := CylinderMesh.new()
		flame_mesh.top_radius = 0.10
		flame_mesh.bottom_radius = 0.23
		flame_mesh.height = 1.20
		flame_mesh.radial_segments = 18
		flame.mesh = flame_mesh
		var angle := float(i) * TAU / 3.0
		flame.position = Vector3(cos(angle) * 0.72, 0.68, sin(angle) * 0.72)
		flame.material_override = mat
		marker.add_child(flame)
	return marker


static func make_burning_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = BURNING_MARKER_NAME
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.30, 0.04, 0.72)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.18, 0.02)
	mat.emission_energy_multiplier = 2.5
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var core := MeshInstance3D.new()
	core.name = "BurningCore"
	var core_mesh := CylinderMesh.new()
	core_mesh.top_radius = 0.34
	core_mesh.bottom_radius = 0.52
	core_mesh.height = 1.10
	core_mesh.radial_segments = 24
	core.mesh = core_mesh
	core.position.y = 0.56
	core.material_override = mat
	marker.add_child(core)

	for i in range(4):
		var lick := MeshInstance3D.new()
		lick.name = "BurningLick%d" % i
		var lick_mesh := CylinderMesh.new()
		lick_mesh.top_radius = 0.05
		lick_mesh.bottom_radius = 0.14
		lick_mesh.height = 0.74
		lick_mesh.radial_segments = 12
		lick.mesh = lick_mesh
		var angle := float(i) * TAU / 4.0
		lick.position = Vector3(cos(angle) * 0.36, 0.62, sin(angle) * 0.36)
		lick.material_override = mat
		marker.add_child(lick)
	return marker


static func make_elite_command_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = ELITE_COMMAND_MARKER_NAME
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.36, 0.78, 1.0, 0.70)
	mat.emission_enabled = true
	mat.emission = Color(0.24, 0.72, 1.0)
	mat.emission_energy_multiplier = 2.2
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var ring := MeshInstance3D.new()
	ring.name = "EliteCommandRing"
	var ring_mesh := TorusMesh.new()
	ring_mesh.inner_radius = 0.78
	ring_mesh.outer_radius = 0.92
	ring_mesh.ring_segments = 72
	ring.mesh = ring_mesh
	ring.position.y = 0.10
	ring.material_override = mat
	marker.add_child(ring)

	var crown := MeshInstance3D.new()
	crown.name = "EliteCommandCrown"
	var crown_mesh := CylinderMesh.new()
	crown_mesh.top_radius = 0.20
	crown_mesh.bottom_radius = 0.42
	crown_mesh.height = 0.28
	crown_mesh.radial_segments = 5
	crown.mesh = crown_mesh
	crown.position.y = 1.60
	crown.material_override = mat
	marker.add_child(crown)
	return marker


static func make_elite_command_radius_preview(radius: float) -> Node3D:
	var marker := Node3D.new()
	marker.name = ELITE_COMMAND_RADIUS_PREVIEW_NAME
	marker.set_meta("radius", maxf(radius, 0.5))
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.32, 0.74, 1.0, 0.18)
	mat.emission_enabled = true
	mat.emission = Color(0.20, 0.64, 1.0)
	mat.emission_energy_multiplier = 1.6
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var ring := MeshInstance3D.new()
	ring.name = "EliteCommandRadiusRing"
	var ring_mesh := TorusMesh.new()
	ring_mesh.inner_radius = 0.98
	ring_mesh.outer_radius = 1.0
	ring_mesh.ring_segments = 96
	ring.mesh = ring_mesh
	ring.position.y = 0.06
	ring.scale = Vector3.ONE * maxf(radius, 0.5)
	ring.material_override = mat
	marker.add_child(ring)
	return marker


static func active_holy_shield_aura_pulse_count(root: Node3D) -> int:
	return _active_pulse_count(root, HOLY_SHIELD_AURA_PULSE_NAME)


static func active_holy_shield_target_pulse_count(root: Node3D) -> int:
	return _active_pulse_count(root, HOLY_SHIELD_TARGET_PULSE_NAME)


static func _holy_shield_pulse_material(alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.90, 0.32, alpha)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.86, 0.28)
	mat.emission_energy_multiplier = 3.4
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	return mat


static func _fade_and_free(node: Node3D) -> void:
	node.scale = Vector3.ONE * 0.55
	var tween := node.create_tween()
	tween.set_parallel(true)
	tween.tween_property(node, "scale", Vector3.ONE * 1.08, AURA_PULSE_SECONDS)
	tween.chain().tween_callback(node.queue_free)


static func _active_pulse_count(root: Node3D, pulse_name: String) -> int:
	if root == null:
		return 0
	var count := 0
	for child in root.get_children():
		if child is Node3D and str(child.name) == pulse_name:
			count += 1
	return count
