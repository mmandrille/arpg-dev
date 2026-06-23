class_name BossArenaPresence
extends RefCounted

const MARKER_NAME := "BossArenaPresence"
const DEFAULT_COLOR := Color(0.92, 0.28, 0.22, 0.22)
const DEFAULT_RADIUS := 2.4

static func sync_for_record(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return

	var is_live_boss := str(rec.get("type", "")) == "monster" and bool(rec.get("is_boss", false)) and int(rec.get("hp", 0)) > 0
	var marker := node.find_child(MARKER_NAME, false, false) as MeshInstance3D
	if not is_live_boss:
		if marker != null:
			marker.queue_free()
		rec["has_boss_arena_presence"] = false
		return

	if marker == null:
		marker = MeshInstance3D.new()
		marker.name = MARKER_NAME
		marker.position = Vector3(0.0, 0.018, 0.0)
		node.add_child(marker)

	var visual_scale := maxf(0.1, float(rec.get("visual_scale", 1.0)))
	var local_radius := DEFAULT_RADIUS / visual_scale
	marker.mesh = _arena_mesh(local_radius)
	marker.material_override = _arena_material(_arena_color(rec))
	rec["has_boss_arena_presence"] = true
	rec["boss_arena_color"] = _arena_color(rec).to_html(true)

static func remove_for_record(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node != null:
		var marker := node.find_child(MARKER_NAME, false, false)
		if marker != null:
			marker.queue_free()
	rec["has_boss_arena_presence"] = false
	rec["boss_arena_color"] = ""

static func _arena_color(rec: Dictionary) -> Color:
	if bool(rec.get("boss_telegraph_active", false)):
		var tint := str(rec.get("telegraph_tint", ""))
		if not tint.is_empty():
			var color := Color("#" + tint)
			color.a = 0.30
			return color

	var phase: Dictionary = rec.get("boss_phase", {}) if typeof(rec.get("boss_phase", {})) == TYPE_DICTIONARY else {}
	if str(phase.get("phase_kind", "")) == "telegraph":
		var telegraph: Dictionary = phase.get("telegraph", {}) if typeof(phase.get("telegraph", {})) == TYPE_DICTIONARY else {}
		var telegraph_color := Color(str(telegraph.get("to_color", "#ff4a32")))
		telegraph_color.a = 0.30
		return telegraph_color

	return DEFAULT_COLOR

static func _arena_mesh(local_radius: float) -> Mesh:
	var torus := TorusMesh.new()
	torus.inner_radius = local_radius * 0.82
	torus.outer_radius = local_radius
	torus.rings = 12
	torus.ring_segments = 24
	return torus

static func _arena_material(color: Color) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	mat.cull_mode = BaseMaterial3D.CULL_DISABLED
	return mat
