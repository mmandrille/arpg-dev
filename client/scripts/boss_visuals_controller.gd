class_name BossVisualsController
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

var ctx: RefCounted
var boss_health_bar

func _init(context: RefCounted = null, health_bar = null) -> void:
	ctx = context
	boss_health_bar = health_bar

func hide_boss_health_bar() -> void:
	if boss_health_bar != null:
		boss_health_bar.hide_boss()

func sync_boss_health_bar() -> void:
	if boss_health_bar == null or ctx == null:
		return
	var boss_id := active_boss_entity_id()
	if boss_id == "":
		boss_health_bar.hide_boss()
		return
	var rec: Dictionary = ctx.entities[boss_id]
	var hp := int(rec.get("hp", 0))
	var max_hp := int(rec.get("max_hp", hp))
	var template_id := str(rec.get("boss_template_id", ""))
	boss_health_bar.show_boss(boss_id, template_id, boss_health_bar_title(template_id), hp, max_hp)
	var phase := boss_phase_for_display(rec)
	if phase.is_empty():
		boss_health_bar.clear_phase_state()
	else:
		boss_health_bar.set_phase_state(phase)

func advance_boss_phase_display(delta: float) -> void:
	if ctx == null:
		return
	var changed := false
	for id in ctx.entities.keys():
		var rec: Dictionary = ctx.entities[id]
		var phase := boss_phase_for_display(rec)
		if phase.is_empty():
			continue
		var remaining := float(phase.get("remaining_ticks_float", float(phase.get("remaining_ticks", 0))))
		remaining = maxf(0.0, remaining - maxf(0.0, delta) * ClientConstantsScript.BOSS_PHASE_TICK_RATE)
		phase["remaining_ticks_float"] = remaining
		phase["remaining_ticks"] = int(ceil(remaining))
		rec["boss_phase"] = phase
		changed = true
	if changed:
		sync_boss_health_bar()

func boss_phase_for_display(rec: Dictionary) -> Dictionary:
	var raw = rec.get("boss_phase", {})
	if typeof(raw) != TYPE_DICTIONARY:
		return {}
	var phase: Dictionary = (raw as Dictionary).duplicate(true)
	var kind := str(phase.get("phase_kind", ""))
	if kind == "":
		return {}
	var duration := maxi(0, int(phase.get("duration_ticks", 0)))
	phase["duration_ticks"] = duration
	var remaining := int(phase.get("remaining_ticks", -1))
	if remaining < 0:
		var started_tick := int(phase.get("started_tick", ctx.last_server_tick if ctx != null else 0))
		remaining = max(0, duration - max(0, (ctx.last_server_tick if ctx != null else 0) - started_tick))
	phase["remaining_ticks"] = clampi(remaining, 0, duration)
	phase["remaining_ticks_float"] = clampf(float(phase.get("remaining_ticks_float", phase["remaining_ticks"])), 0.0, float(duration))
	rec["boss_phase"] = phase
	return phase

func active_boss_entity_id() -> String:
	if ctx == null:
		return ""
	var candidates: Array = []
	for id in ctx.entities.keys():
		var rec: Dictionary = ctx.entities[id]
		if str(rec.get("type", "")) != "monster":
			continue
		if not bool(rec.get("is_boss", false)):
			continue
		if int(rec.get("hp", 0)) <= 0:
			continue
		candidates.append(str(id))
	candidates.sort()
	return str(candidates[0]) if not candidates.is_empty() else ""

func boss_health_bar_title(template_id: String) -> String:
	if template_id == "":
		return "Boss"
	var pieces := template_id.replace("-", "_").split("_", false)
	var words := PackedStringArray()
	for piece in pieces:
		var word := str(piece)
		if word == "":
			continue
		words.append(word.substr(0, 1).to_upper() + word.substr(1).to_lower())
	return " ".join(words)

func apply_boss_phase_started(entity_id: String, ev: Dictionary) -> void:
	if ctx == null:
		return
	var rec: Dictionary = ctx.entities.get(entity_id, {})
	if rec.is_empty():
		return
	var duration := maxi(0, int(ev.get("duration_ticks", 0)))
	rec["boss_phase"] = {
		"pattern_id": str(ev.get("pattern_id", "")),
		"phase_index": int(ev.get("phase_index", -1)),
		"phase_kind": str(ev.get("phase_kind", "")),
		"duration_ticks": duration,
		"remaining_ticks": duration,
		"remaining_ticks_float": float(duration),
		"telegraph": ev.get("telegraph", {}),
		"hit_shape": ev.get("hit_shape", {}),
	}
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var phase_kind := str(ev.get("phase_kind", ""))
	if phase_kind == "telegraph":
		var telegraph: Dictionary = ev.get("telegraph", {})
		var tint := Color(str(telegraph.get("to_color", "#ff0000")))
		rec["boss_telegraph_active"] = true
		rec["telegraph_tint"] = tint.to_html(false)
		if ctx.apply_model_tint.is_valid():
			ctx.apply_model_tint.call(node, tint)
		sync_boss_telegraph_marker(rec, telegraph)
	else:
		rec["boss_telegraph_active"] = false
		remove_boss_telegraph_marker(rec)
		if ctx.apply_entity_status_tint.is_valid():
			ctx.apply_entity_status_tint.call(rec)
	sync_boss_health_bar()

func apply_boss_phase_ended(entity_id: String, _ev: Dictionary) -> void:
	if ctx == null:
		return
	var rec: Dictionary = ctx.entities.get(entity_id, {})
	if rec.is_empty():
		return
	rec["boss_telegraph_active"] = false
	rec.erase("boss_phase")
	remove_boss_telegraph_marker(rec)
	if ctx.apply_entity_status_tint.is_valid():
		ctx.apply_entity_status_tint.call(rec)
	sync_boss_health_bar()

func normalize_boss_phase_metadata(rec: Dictionary) -> void:
	var phase := boss_phase_for_display(rec)
	if phase.is_empty():
		return
	if str(phase.get("phase_kind", "")) == "telegraph":
		var telegraph: Dictionary = phase.get("telegraph", {})
		if not telegraph.is_empty():
			rec["boss_telegraph_active"] = true
			rec["telegraph_tint"] = Color(str(telegraph.get("to_color", "#ff0000"))).to_html(false)

func sync_boss_telegraph_marker_from_record(rec: Dictionary) -> void:
	var phase := boss_phase_for_display(rec)
	if phase.is_empty() or str(phase.get("phase_kind", "")) != "telegraph":
		remove_boss_telegraph_marker(rec)
		return
	var telegraph: Dictionary = phase.get("telegraph", {})
	if telegraph.is_empty():
		remove_boss_telegraph_marker(rec)
		return
	sync_boss_telegraph_marker(rec, telegraph)

func sync_boss_telegraph_marker(rec: Dictionary, telegraph: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var marker := node.find_child(ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME, false, false) as MeshInstance3D
	if marker == null:
		marker = MeshInstance3D.new()
		marker.name = ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME
		marker.position = Vector3(0.0, 0.035, 0.0)
		node.add_child(marker)
	var radius := maxf(0.1, float(telegraph.get("radius", 1.0)))
	var visual_scale := maxf(0.1, float(rec.get("visual_scale", 1.0)))
	var local_radius := radius / visual_scale
	var mesh := CylinderMesh.new()
	mesh.top_radius = local_radius
	mesh.bottom_radius = local_radius
	mesh.height = 0.035
	mesh.radial_segments = 48
	marker.mesh = mesh
	var color := Color(str(telegraph.get("to_color", "#ff4a32")))
	color.a = 0.34
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	marker.material_override = mat
	rec["boss_telegraph_active"] = true
	rec["telegraph_tint"] = color.to_html(false)
	rec["has_boss_telegraph_marker"] = true
	rec["telegraph_radius"] = radius
	rec["telegraph_marker_color"] = color.to_html(false)

func remove_boss_telegraph_marker(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node != null:
		var marker := node.find_child(ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME, false, false)
		if marker != null:
			marker.queue_free()
	rec["has_boss_telegraph_marker"] = false
	rec["telegraph_radius"] = 0.0
	rec["telegraph_marker_color"] = ""
