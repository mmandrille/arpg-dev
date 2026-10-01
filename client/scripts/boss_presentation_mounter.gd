## Mounts catalog attachments and headgear on a boss rig (v512). Bones/assets come from
## shared/assets/boss_presentation.v0.json; code-native headgear is built from data (shape/color/size).
class_name BossPresentationMounter
extends RefCounted

const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const HEADGEAR_NAME := "BossHeadgear"


static func mount(visual: Node3D, cfg: Dictionary) -> void:
	if visual == null:
		return
	if visual.has_method("apply_presentation"):
		visual.call("apply_presentation")  # clip aliases + catalog attachments before ours
	var skel := visual.find_child("Skeleton3D", true, false) as Skeleton3D
	if skel == null:
		push_warning("BossPresentationMounter: no Skeleton3D under %s" % visual.name)
		return
	for raw in cfg.get("attachments", []):
		var spec := raw as Dictionary
		var piece := LibraryScript.instantiate(str(spec.get("asset_id", "")))
		if piece != null:
			_attach(skel, str(spec.get("bone", "")), piece, "BossAttach")
	var headgear: Dictionary = cfg.get("headgear", {})
	var node := _headgear_node(headgear)
	if node != null:
		_attach(skel, str(headgear.get("bone", "head")), node, "BossHead")


static func _attach(skel: Skeleton3D, bone: String, piece: Node3D, prefix: String) -> BoneAttachment3D:
	if skel.find_bone(bone) < 0:
		push_warning("BossPresentationMounter: no bone %s" % bone)
		piece.queue_free()
		return null
	var mount := BoneAttachment3D.new()
	mount.name = "%s_%s" % [prefix, bone.replace(".", "_")]
	mount.bone_name = bone
	skel.add_child(mount)
	mount.add_child(piece)
	return mount


static func _headgear_node(headgear: Dictionary) -> Node3D:
	match str(headgear.get("kind", "none")):
		"asset":
			return LibraryScript.instantiate(str(headgear.get("asset_id", "")))
		"primitive":
			var node := build_primitive(str(headgear.get("shape", "")), Color(str(headgear.get("color", "#ffffff"))), float(headgear.get("size", 1.0)))
			node.position.y = float(headgear.get("lift", 0.0))
			return node
	return null


## Horns or crown from cone/cylinder primitives; head-bone local space (Y up from the bone).
static func build_primitive(shape: String, color: Color, size: float) -> Node3D:
	var root := Node3D.new()
	root.name = HEADGEAR_NAME
	root.scale = Vector3.ONE * maxf(0.05, size)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.metallic = 0.35
	mat.roughness = 0.55
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = 0.35
	match shape:
		"horns":
			for side in [-1.0, 1.0]:
				var horn := _cone(0.045, 0.34, mat)
				horn.position = Vector3(0.11 * side, 0.22, 0.0)
				horn.rotation = Vector3(0.0, 0.0, -0.55 * side)
				root.add_child(horn)
		"crown":
			var band := MeshInstance3D.new()
			var band_mesh := CylinderMesh.new()
			band_mesh.top_radius = 0.135
			band_mesh.bottom_radius = 0.135
			band_mesh.height = 0.05
			band_mesh.radial_segments = 16
			band.mesh = band_mesh
			band.material_override = mat
			band.position = Vector3(0.0, 0.3, 0.0)
			root.add_child(band)
			for i in range(5):
				var angle := TAU * float(i) / 5.0
				var spike := _cone(0.03, 0.15, mat)
				spike.position = Vector3(sin(angle) * 0.125, 0.39, cos(angle) * 0.125)
				root.add_child(spike)
		_:
			push_warning("BossPresentationMounter: unknown headgear shape %s" % shape)
	return root


static func _cone(radius: float, height: float, mat: Material) -> MeshInstance3D:
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.0
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = 8
	var inst := MeshInstance3D.new()
	inst.mesh = mesh
	inst.material_override = mat
	return inst
