class_name BossArenaPresence
extends RefCounted

const MARKER_NAME := "BossArenaPresence"
const DEFAULT_COLOR := Color(0.92, 0.28, 0.22, 0.22)
const DEFAULT_RADIUS := 2.4
const DISC_NAME := "BossAuraDisc"
const LIGHT_NAME := "BossAuraLight"
const PULSE_META := "boss_aura_pulse_hz"
const LoaderScript := preload("res://scripts/boss_presentation_loader.gd")
const RenderPresentationLoaderScript := preload("res://scripts/render_presentation_loader.gd")
const CombatVfxScript := preload("res://scripts/combat_vfx.gd")

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
	var aura := _aura_config(rec)
	var radius := float(aura.get("radius", DEFAULT_RADIUS))
	var local_radius := radius / visual_scale
	var color := _arena_color(rec)
	var key := "%.3f|%s|%s" % [local_radius, color.to_html(true), str(_full_quality())]
	if str(marker.get_meta("aura_key", "")) != key or marker.mesh == null:
		marker.set_meta("aura_key", key)  # sync runs every tick; rebuild meshes only on change
		marker.mesh = _arena_mesh(local_radius)
		marker.material_override = _arena_material(color)
		_sync_aura_extras(node, marker, aura, local_radius)
	rec["has_boss_arena_presence"] = true
	rec["boss_arena_color"] = color.to_html(true)
	rec["boss_aura_radius"] = radius
	rec["boss_aura_full"] = _full_quality()

static func remove_for_record(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node != null:
		var marker := node.find_child(MARKER_NAME, false, false)
		if marker != null:
			marker.queue_free()
	rec["has_boss_arena_presence"] = false
	rec["boss_arena_color"] = ""

static func _aura_config(rec: Dictionary) -> Dictionary:
	var cfg := LoaderScript.entry(str(rec.get("boss_template_id", "")))
	var aura = cfg.get("aura", {})
	return aura if typeof(aura) == TYPE_DICTIONARY else {}

static func _full_quality() -> bool:
	var tier := RenderPresentationLoaderScript.quality_tier(CombatVfxScript.quality())
	return float(tier.get("particle_scale", 1.0)) >= 1.0

## Inner disc, soft light and slow pulse. The Performance tier keeps the ring only.
static func _sync_aura_extras(node: Node3D, marker: MeshInstance3D, aura: Dictionary, local_radius: float) -> void:
	var disc := marker.find_child(DISC_NAME, false, false) as MeshInstance3D
	var light := marker.find_child(LIGHT_NAME, false, false) as OmniLight3D
	if aura.is_empty():
		_free_extra(disc)
		_free_extra(light)
		return
	var color := Color(str(aura.get("color", "#ff4a32")))
	var inner_alpha := float(aura.get("inner_alpha", 0.0))
	if inner_alpha > 0.0:
		if disc == null:
			disc = MeshInstance3D.new()
			disc.name = DISC_NAME
			disc.position = Vector3(0.0, 0.004, 0.0)
			marker.add_child(disc)
		var mesh := CylinderMesh.new()
		mesh.top_radius = local_radius * 0.82
		mesh.bottom_radius = local_radius * 0.82
		mesh.height = 0.01
		mesh.radial_segments = 32
		disc.mesh = mesh
		disc.material_override = _arena_material(Color(color.r, color.g, color.b, inner_alpha))
	else:
		_free_extra(disc)
	var full := _full_quality()
	var energy := float(aura.get("light_energy", 0.0))
	if full and energy > 0.0:
		if light == null:
			light = OmniLight3D.new()
			light.name = LIGHT_NAME
			light.position = Vector3(0.0, 0.8, 0.0)
			light.shadow_enabled = false
			marker.add_child(light)
		light.light_color = color
		light.light_energy = energy
		light.omni_range = float(aura.get("light_range", 5.0))
	else:
		_free_extra(light)
	var hz := float(aura.get("pulse_hz", 0.0)) if full else 0.0
	_sync_pulse(marker, hz)

static func _free_extra(extra: Node) -> void:
	if extra != null:
		extra.get_parent().remove_child(extra)
		extra.queue_free()

static func _sync_pulse(marker: MeshInstance3D, hz: float) -> void:
	if is_equal_approx(float(marker.get_meta(PULSE_META, 0.0)), hz):
		return
	marker.set_meta(PULSE_META, hz)
	marker.scale = Vector3.ONE
	if hz <= 0.0 or not marker.is_inside_tree():
		marker.set_meta(PULSE_META, 0.0)
		return
	var half := 0.5 / hz
	var tween := marker.create_tween().set_loops()
	tween.tween_property(marker, "scale", Vector3(1.05, 1.0, 1.05), half).set_trans(Tween.TRANS_SINE)
	tween.tween_property(marker, "scale", Vector3(0.97, 1.0, 0.97), half).set_trans(Tween.TRANS_SINE)

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

	var aura := _aura_config(rec)
	if aura.has("color"):
		var themed := Color(str(aura["color"]))
		themed.a = DEFAULT_COLOR.a
		return themed
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
