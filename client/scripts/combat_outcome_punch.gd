class_name CombatOutcomePunch
extends RefCounted

const DamageTypeCombatTextScript := preload("res://scripts/damage_type_combat_text.gd")

const NODE_NAME := "CombatOutcomePunch"


static func should_spawn(ev: Dictionary) -> bool:
	var outcome := _normalized_outcome(ev)
	return outcome in ["hit", "kill", "miss", "block", "immune", "crit"]


static func make_node(ev: Dictionary) -> Node3D:
	var outcome := _normalized_outcome(ev)
	var root := Node3D.new()
	root.name = NODE_NAME
	root.set_meta("outcome", outcome)
	var color := _color_for_outcome(ev, outcome)
	root.add_child(_ring_node(color, _ring_scale(outcome)))
	for i in range(_spark_count(outcome)):
		root.add_child(_spark_node(color, i, outcome))
	return root


static func _normalized_outcome(ev: Dictionary) -> String:
	var outcome := str(ev.get("outcome", "")).to_lower()
	if bool(ev.get("critical", false)):
		return "crit"
	var event_type := str(ev.get("event_type", ""))
	if event_type == "monster_killed":
		return "kill"
	if event_type == "monster_damaged" and outcome == "":
		return "hit"
	return outcome


static func _color_for_outcome(ev: Dictionary, outcome: String) -> Color:
	if outcome == "crit":
		var presentation := DamageTypeCombatTextScript.number_for_event(ev, Color(1.0, 0.58, 0.22))
		return presentation.get("color", Color(1.0, 0.58, 0.22))
	if outcome == "kill":
		return Color(1.0, 0.72, 0.26)
	if outcome == "hit":
		var damage_presentation := DamageTypeCombatTextScript.number_for_event(ev, Color(1.0, 0.92, 0.76))
		return damage_presentation.get("color", Color(1.0, 0.92, 0.76))
	var special := DamageTypeCombatTextScript.special_outcome(outcome)
	if not special.is_empty():
		return special.get("color", Color.WHITE)
	return Color.WHITE


static func _ring_scale(outcome: String) -> float:
	match outcome:
		"crit":
			return 1.18
		"kill":
			return 1.26
		"immune":
			return 1.08
		"hit":
			return 0.96
		"block":
			return 1.0
		_:
			return 0.92


static func _spark_count(outcome: String) -> int:
	match outcome:
		"crit":
			return 8
		"kill":
			return 7
		"immune":
			return 6
		"hit":
			return 5
		"block":
			return 4
		_:
			return 3


static func _ring_node(color: Color, scale: float) -> MeshInstance3D:
	var ring := MeshInstance3D.new()
	ring.name = "OutcomeRing"
	ring.position = Vector3(0.0, 0.55, 0.0)
	var torus := TorusMesh.new()
	torus.inner_radius = 0.34 * scale
	torus.outer_radius = 0.48 * scale
	ring.mesh = torus
	ring.material_override = _material(color, 0.42 if scale > 1.0 else 0.34)
	return ring


static func _spark_node(color: Color, index: int, outcome: String) -> MeshInstance3D:
	var spark := MeshInstance3D.new()
	spark.name = "OutcomeSpark_%d" % index
	var mesh := BoxMesh.new()
	var height := 0.22 + float(index % 2) * 0.06
	if outcome == "crit":
		height += 0.08
	elif outcome == "kill":
		height += 0.12
	mesh.size = Vector3(0.03, 0.03, height)
	spark.mesh = mesh
	var angle := (TAU / float(_spark_count(outcome))) * float(index)
	var radius := 0.14 if outcome == "miss" else 0.24 if outcome == "kill" else 0.20
	spark.position = Vector3(cos(angle) * radius, 0.58 + float(index) * 0.02, sin(angle) * radius)
	spark.rotation = Vector3(0.36, angle, 0.18)
	spark.material_override = _material(color, 0.78 - float(index % 3) * 0.08)
	return spark


static func _material(color: Color, alpha: float) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, alpha)
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 1.1
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	return mat
