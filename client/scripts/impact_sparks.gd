class_name ImpactSparks
extends RefCounted

const DamageTypeCombatTextScript := preload("res://scripts/damage_type_combat_text.gd")


static func color_for_event(ev: Dictionary, fallback: Color) -> Color:
	var presentation := DamageTypeCombatTextScript.number_for_event(ev, fallback)
	if presentation.has("color"):
		return presentation.get("color", fallback)
	return fallback


static func should_spawn(ev: Dictionary) -> bool:
	var outcome := str(ev.get("outcome", "")).to_lower()
	if outcome in ["miss", "block", "immune"]:
		return false
	return ev.has("damage") or str(ev.get("event_type", "")).ends_with("_killed")


static func make_node(ev: Dictionary, fallback: Color = Color("#f4d481")) -> Node3D:
	var root := Node3D.new()
	root.name = "ImpactSparks"
	root.set_meta("event_type", str(ev.get("event_type", "")))
	root.set_meta("damage_type", str(ev.get("damage_type", "")))
	var color := color_for_event(ev, fallback)
	for i in range(5):
		var spark := MeshInstance3D.new()
		spark.name = "Spark_%d" % i
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.035, 0.035, 0.28 + float(i % 2) * 0.08)
		spark.mesh = mesh
		var angle := (TAU / 5.0) * float(i)
		spark.position = Vector3(cos(angle) * 0.16, 0.62 + float(i) * 0.018, sin(angle) * 0.16)
		spark.rotation = Vector3(0.42, angle, 0.22)
		spark.material_override = _material(color, 0.72 - float(i) * 0.06)
		root.add_child(spark)
	return root


static func _material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.9
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	return mat
