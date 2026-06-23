class_name MonsterFamilyAccent
extends RefCounted

const MARKER_NAME := "MonsterFamilyAccent"


static func sync_for_monster(node: Node3D, monster_def_id: String) -> void:
	if node == null:
		return
	var visuals: Dictionary = MonsterVisualsLoader.resolve(monster_def_id)
	var accent_hex := str(visuals.get("family_accent", ""))
	var marker := node.find_child(MARKER_NAME, false, false) as MeshInstance3D
	if accent_hex.is_empty():
		if marker != null:
			marker.queue_free()
		return
	if marker == null:
		marker = MeshInstance3D.new()
		marker.name = MARKER_NAME
		marker.position = Vector3(0.0, 0.12, 0.0)
		node.add_child(marker)
	var torus := TorusMesh.new()
	torus.inner_radius = 0.34
	torus.outer_radius = 0.42
	marker.mesh = torus
	var color := Color("#" + accent_hex.trim_prefix("#"))
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, 0.24)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.55
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	marker.material_override = mat
