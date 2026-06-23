class_name ClassIdleStance
extends RefCounted

const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")


static func apply_to_model(model: Node3D, class_id: String) -> void:
	if model == null:
		return
	var stance: Dictionary = ClassPresentationsLoaderScript.idle_stance_for_class(class_id)
	model.rotation_degrees.z = float(stance.get("lean_degrees", 0.0))
	var stance_scale := float(stance.get("scale", 1.0))
	if stance_scale > 0.0:
		model.scale *= stance_scale
