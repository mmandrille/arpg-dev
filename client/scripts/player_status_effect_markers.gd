extends RefCounted

const RogueMarkEffectScript := preload("res://scripts/rogue_mark_effect.gd")
const AuraSoftLightsScript := preload("res://scripts/aura_soft_lights.gd")

const HOLY_SHIELD_EFFECT_ID := "holy_shield"
const SANCTUARY_EFFECT_ID := "sanctuary"
const RAGE_EFFECT_ID := "rage"
const ICE_SLOW_EFFECT_ID := "ice_slow"
const BURNING_EFFECT_ID := "everburning_wound"
const ELITE_COMMAND_EFFECT_ID := "elite_command"
const PINNING_ROOT_EFFECT_ID := "pinning_root"
const STUN_EFFECT_ID := "stun"
const DASH_BLEED_EFFECT_ID := "dash_bleed"
const ROGUE_MARK_EFFECT_ID := "rogue_mark"
const STUN_SKILL_IDS := ["leap", "charge"]

const BLEED_MARKER_NAME := "BleedVisualEffect"
const BURNING_MARKER_NAME := "BurningVisualEffect"
const PINNING_ROOT_MARKER_NAME := "PinningRootVisualEffect"
const STUN_MARKER_NAME := "StunStarsVisualEffect"
const ROGUE_MARK_MARKER_NAME := "RogueMarkSkullEffect"


static func has_holy_shield_effect(root: Node3D) -> bool:
	return AuraSoftLightsScript.has_holy_shield_effect(root)


static func has_sanctuary_effect(root: Node3D) -> bool:
	return AuraSoftLightsScript.has_sanctuary_effect(root)


static func has_rage_effect(root: Node3D) -> bool:
	return AuraSoftLightsScript.has_rage_effect(root)


static func has_elite_command_effect(root: Node3D) -> bool:
	return AuraSoftLightsScript.has_elite_command_effect(root)


static func has_elite_command_radius_preview(root: Node3D) -> bool:
	return AuraSoftLightsScript.has_elite_command_radius_preview(root)


static func elite_command_radius_preview_value(root: Node3D) -> float:
	return AuraSoftLightsScript.elite_command_radius_preview_value(root)


static func active_holy_shield_aura_pulse_count(root: Node3D) -> int:
	return AuraSoftLightsScript.active_holy_shield_cast_pulse_count(root)


static func active_holy_shield_target_pulse_count(root: Node3D) -> int:
	return AuraSoftLightsScript.active_holy_shield_target_pulse_count(root)


static func has_ice_slow_effect(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(ICE_SLOW_EFFECT_ID)


static func has_burning_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(BURNING_EFFECT_ID)


static func has_elite_command_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(ELITE_COMMAND_EFFECT_ID)


static func has_pinning_root_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(PINNING_ROOT_EFFECT_ID)


static func has_rogue_mark_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(ROGUE_MARK_EFFECT_ID)


static func has_bleed_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(DASH_BLEED_EFFECT_ID)


static func is_stun_skill_id(skill_id: String) -> bool:
	return STUN_SKILL_IDS.has(skill_id)


static func has_stun_effect_id(effect_ids_value) -> bool:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return effect_ids.has(STUN_EFFECT_ID) or effect_ids.has("leap_stun") or effect_ids.has("charge_stun")


static func strip_combat_status_effect_ids(effect_ids_value) -> Array:
	var strip_ids := {
		ICE_SLOW_EFFECT_ID: true,
		BURNING_EFFECT_ID: true,
		PINNING_ROOT_EFFECT_ID: true,
		STUN_EFFECT_ID: true,
		"leap_stun": true,
		"charge_stun": true,
		DASH_BLEED_EFFECT_ID: true,
		ROGUE_MARK_EFFECT_ID: true,
		ELITE_COMMAND_EFFECT_ID: true,
	}
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	var kept: Array = []
	for id in effect_ids:
		if not strip_ids.has(str(id)):
			kept.append(id)
	return kept


static func clear_combat_markers(root: Node3D) -> void:
	sync_burning_effect(root, false)
	sync_bleed_effect(root, false)
	sync_pinning_root_effect(root, false)
	sync_stun_effect(root, false)
	sync_rogue_mark_effect(root, false)


static func sync_bleed_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(BLEED_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_bleed_effect()
		root.add_child(existing)


static func has_bleed_effect(root: Node3D) -> bool:
	return root != null and root.find_child(BLEED_MARKER_NAME, false, false) != null


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


static func sync_pinning_root_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(PINNING_ROOT_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_pinning_root_effect()
		root.add_child(existing)


static func has_pinning_root_effect(root: Node3D) -> bool:
	return root != null and root.find_child(PINNING_ROOT_MARKER_NAME, false, false) != null


static func sync_stun_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(STUN_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = make_stun_effect()
		root.add_child(existing)


static func has_stun_effect(root: Node3D) -> bool:
	return root != null and root.find_child(STUN_MARKER_NAME, false, false) != null


static func sync_rogue_mark_effect(root: Node3D, active: bool) -> void:
	if root == null:
		return
	var existing := root.find_child(ROGUE_MARK_MARKER_NAME, false, false) as Node3D
	if not active:
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	if existing == null:
		existing = RogueMarkEffectScript.new()
		root.add_child(existing)


static func has_rogue_mark_effect(root: Node3D) -> bool:
	return root != null and root.find_child(ROGUE_MARK_MARKER_NAME, false, false) != null


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


static func make_pinning_root_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = PINNING_ROOT_MARKER_NAME
	var root_mat := _pinning_root_material(Color(0.32, 0.95, 0.18, 0.9), Color(0.14, 0.72, 0.08), 2.8)
	var floor_mat := _pinning_root_material(Color(0.45, 1.0, 0.24, 0.38), Color(0.20, 0.90, 0.10), 2.0)

	var ring := MeshInstance3D.new()
	ring.name = "PinningRootRing"
	var ring_mesh := TorusMesh.new()
	ring_mesh.inner_radius = 0.58
	ring_mesh.outer_radius = 0.72
	ring_mesh.ring_segments = 72
	ring.mesh = ring_mesh
	ring.position.y = 0.045
	ring.material_override = floor_mat
	marker.add_child(ring)

	var inner_ring := MeshInstance3D.new()
	inner_ring.name = "PinningRootInnerRing"
	var inner_mesh := TorusMesh.new()
	inner_mesh.inner_radius = 0.26
	inner_mesh.outer_radius = 0.32
	inner_mesh.ring_segments = 48
	inner_ring.mesh = inner_mesh
	inner_ring.position.y = 0.055
	inner_ring.material_override = floor_mat
	marker.add_child(inner_ring)

	for i in range(6):
		var tendril := MeshInstance3D.new()
		tendril.name = "PinningRootTendril%d" % i
		var tendril_mesh := BoxMesh.new()
		tendril_mesh.size = Vector3(0.08, 0.06, 0.72)
		tendril.mesh = tendril_mesh
		var angle := float(i) * TAU / 6.0
		tendril.position = Vector3(cos(angle) * 0.28, 0.06, sin(angle) * 0.28)
		tendril.rotation.y = -angle
		tendril.material_override = root_mat
		marker.add_child(tendril)

	for i in range(4):
		var stake := MeshInstance3D.new()
		stake.name = "PinningRootStake%d" % i
		var stake_mesh := CylinderMesh.new()
		stake_mesh.top_radius = 0.035
		stake_mesh.bottom_radius = 0.10
		stake_mesh.height = 0.82
		stake_mesh.radial_segments = 6
		stake.mesh = stake_mesh
		var angle := float(i) * TAU / 4.0 + PI / 4.0
		stake.position = Vector3(cos(angle) * 0.48, 0.40, sin(angle) * 0.48)
		stake.rotation_degrees.z = 16.0 * (1.0 if i % 2 == 0 else -1.0)
		stake.material_override = root_mat
		marker.add_child(stake)
	return marker


static func make_stun_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = STUN_MARKER_NAME
	marker.position.y = 1.62
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.90, 0.22, 0.92)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.78, 0.12)
	mat.emission_energy_multiplier = 3.2
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	var orbit := Node3D.new()
	orbit.name = "StunStarsOrbit"
	marker.add_child(orbit)
	for i in range(5):
		var star := MeshInstance3D.new()
		star.name = "StunStar%d" % i
		var star_mesh := CylinderMesh.new()
		star_mesh.top_radius = 0.13
		star_mesh.bottom_radius = 0.13
		star_mesh.height = 0.035
		star_mesh.radial_segments = 5
		star.mesh = star_mesh
		var angle := float(i) * TAU / 5.0
		star.position = Vector3(cos(angle) * 0.48, 0.07 * sin(angle * 2.0), sin(angle) * 0.48)
		star.rotation_degrees = Vector3(90.0, 0.0, 36.0)
		star.material_override = mat
		orbit.add_child(star)

	var tween := orbit.create_tween()
	tween.set_loops()
	tween.tween_property(orbit, "rotation:y", TAU, 0.95).from(0.0)
	return marker


static func make_bleed_effect() -> Node3D:
	var marker := Node3D.new()
	marker.name = BLEED_MARKER_NAME
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.82, 0.04, 0.08, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(0.62, 0.02, 0.05)
	mat.emission_energy_multiplier = 2.2
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED

	for i in range(5):
		var drip := MeshInstance3D.new()
		drip.name = "BleedDrip%d" % i
		var drip_mesh := CylinderMesh.new()
		drip_mesh.top_radius = 0.03
		drip_mesh.bottom_radius = 0.08
		drip_mesh.height = 0.42
		drip_mesh.radial_segments = 10
		drip.mesh = drip_mesh
		var angle := float(i) * TAU / 5.0
		drip.position = Vector3(cos(angle) * 0.28, 0.48, sin(angle) * 0.28)
		drip.material_override = mat
		marker.add_child(drip)
	return marker


static func _pinning_root_material(color: Color, emission: Color, emission_energy: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.emission_enabled = true
	mat.emission = emission
	mat.emission_energy_multiplier = emission_energy
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.blend_mode = BaseMaterial3D.BLEND_MODE_ADD
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	mat.no_depth_test = true
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	return mat
