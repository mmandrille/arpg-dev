## Applies a resolved v511 monster variant look to a monster visual root: detail-layer tint
## (shared cached materials), recolored model eye glow, scale multiplier on the model child, and
## a ground aura ring. Everything is built once at node creation; nothing runs per frame.
## Tinted/eye/aura meshes carry SKIP_META so ModelReactionController and the status-tint pass
## leave them alone (their emission/albedo are owned here).
class_name MonsterVariantLook
extends RefCounted

const ResolverScript := preload("res://scripts/monster_variant_resolver.gd")
const DetailTintScript := preload("res://scripts/model_detail_tint.gd")
const MonsterVisualsLoaderScript := preload("res://scripts/monster_visuals_loader.gd")
const AURA_NAME := "MonsterVariantAura"
const SKIP_META := &"presentation_skip_tint"
const MAX_CACHED := 256
const AURA_HEIGHT := 0.06

static var _materials: Dictionary = {}
static var _aura_meshes: Dictionary = {}


static func palette_id_for_level(level: int, factory: GroundWallFactory) -> String:
	return DungeonDepthLighting.palette_id_for_level(level, factory)


## Whether the catalog owns this entity's rarity colour (callers use a neutral albedo base).
static func owns_entity(entity: Dictionary) -> bool:
	if str(entity.get("type", "")) != "monster":
		return false
	var scene := str(MonsterVisualsLoaderScript.resolve(str(entity.get("monster_def_id", "")), str(entity.get("visual_model", ""))).get("scene", ""))
	return ResolverScript.owns_look(entity, scene)


static func apply_for_entity(root: Node3D, entity: Dictionary, palette_id: String) -> void:
	if root == null or not owns_entity(entity):
		return
	var scene := str(MonsterVisualsLoaderScript.resolve(str(entity.get("monster_def_id", "")), str(entity.get("visual_model", ""))).get("scene", ""))
	apply(root, ResolverScript.resolve(scene, str(entity.get("rarity", "common")), palette_id))


static func apply(root: Node3D, look: Dictionary) -> void:
	if root == null or look.is_empty():
		return
	var model := _model_child(root)
	if model != null:
		if not model.has_meta("variant_base_scale"):
			model.set_meta("variant_base_scale", model.scale)
		model.scale = (model.get_meta("variant_base_scale") as Vector3) * float(look.get("scale_multiplier", 1.0))
	_apply_materials(root, look)
	_sync_aura(root, look.get("aura", {}))


static func clear(root: Node3D) -> void:
	apply(root, {"scale_multiplier": 1.0, "tint": {"color": Color.WHITE, "strength": 0.0}, "eye": {}, "aura": {}, "eye_mesh": "", "key": ""} if root != null else {})


## Hides the ground aura of a dying monster (the corpse keeps its model).
static func hide_aura(root: Node3D) -> void:
	var aura := root.find_child(AURA_NAME, false, false) as Node3D if root != null else null
	if aura != null:
		aura.visible = false


static func _model_child(root: Node3D) -> Node3D:
	for child in root.get_children():
		if child is Node3D and child.name != AURA_NAME and not str(child.name).begins_with("MonsterFamilyAccent"):
			return child as Node3D
	return null


static func _apply_materials(node: Node, look: Dictionary) -> void:
	if node is MeshInstance3D and node.name != AURA_NAME:
		var mesh := node as MeshInstance3D
		if not mesh.has_meta("variant_source"):
			mesh.set_meta("variant_source", _current_material(mesh))
		var source := mesh.get_meta("variant_source") as Material
		var eye_mesh := str(look.get("eye_mesh", ""))
		var is_eye := eye_mesh != "" and str(mesh.name).ends_with(eye_mesh)
		var tint: Dictionary = look.get("tint", {})
		if is_eye:
			var eye: Dictionary = look.get("eye", {})
			if not eye.is_empty() and source is StandardMaterial3D:
				mesh.material_override = _eye_material(source as StandardMaterial3D, eye)
				mesh.set_meta(SKIP_META, true)
			else:
				mesh.material_override = source if mesh.material_override != null else null
				mesh.set_meta(SKIP_META, false)
		elif float(tint.get("strength", 0.0)) > 0.0 and source is StandardMaterial3D:
			mesh.material_override = _tinted_material(source as StandardMaterial3D, tint)
		else:
			mesh.material_override = source
	for child in node.get_children():
		_apply_materials(child, look)


static func _current_material(mesh: MeshInstance3D) -> Material:
	if mesh.material_override != null:
		return mesh.material_override
	if mesh.mesh != null and mesh.mesh.get_surface_count() > 0:
		return mesh.get_active_material(0)
	return null


static func _tinted_material(source: StandardMaterial3D, tint: Dictionary) -> StandardMaterial3D:
	var color := tint["color"] as Color
	var strength := float(tint["strength"])
	var key := "t%d:%s:%.3f" % [source.get_instance_id(), color.to_html(false), strength]
	if _materials.has(key):
		return _materials[key] as StandardMaterial3D
	var mat := source.duplicate() as StandardMaterial3D
	mat.detail_enabled = true
	mat.detail_blend_mode = BaseMaterial3D.BLEND_MODE_MIX
	mat.detail_albedo = DetailTintScript.detail_texture(color, strength)
	return _remember(key, mat)


static func _eye_material(source: StandardMaterial3D, eye: Dictionary) -> StandardMaterial3D:
	var color := eye["color"] as Color
	var key := "e%d:%s:%.2f" % [source.get_instance_id(), color.to_html(false), float(eye["energy"])]
	if _materials.has(key):
		return _materials[key] as StandardMaterial3D
	var mat := source.duplicate() as StandardMaterial3D
	mat.albedo_color = color
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = float(eye["energy"])
	return _remember(key, mat)


static func _remember(key: String, mat: StandardMaterial3D) -> StandardMaterial3D:
	if _materials.size() >= MAX_CACHED:
		_materials.clear()
	_materials[key] = mat
	return mat


static func _sync_aura(root: Node3D, aura: Dictionary) -> void:
	var ring := root.find_child(AURA_NAME, false, false) as MeshInstance3D
	if aura.is_empty():
		if ring != null:
			root.remove_child(ring)
			ring.free()
		return
	if ring == null:
		ring = MeshInstance3D.new()
		ring.name = AURA_NAME
		ring.position = Vector3(0.0, AURA_HEIGHT, 0.0)
		ring.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		ring.set_meta(SKIP_META, true)
		root.add_child(ring)
	var radius := float(aura["radius"])
	var mesh_key := "%.3f" % radius
	if not _aura_meshes.has(mesh_key):
		var torus := TorusMesh.new()
		torus.outer_radius = radius
		torus.inner_radius = radius * 0.86
		torus.rings = 24
		torus.ring_segments = 6
		_aura_meshes[mesh_key] = torus
	ring.mesh = _aura_meshes[mesh_key]
	ring.material_override = _aura_material(aura)
	ring.visible = true


static func _aura_material(aura: Dictionary) -> StandardMaterial3D:
	var color := aura["color"] as Color
	var key := "a%s:%.2f:%.2f" % [color.to_html(false), float(aura["alpha"]), float(aura["energy"])]
	if _materials.has(key):
		return _materials[key] as StandardMaterial3D
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, float(aura["alpha"]))
	mat.emission_enabled = true
	mat.emission = color
	mat.emission_energy_multiplier = float(aura["energy"])
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	return _remember(key, mat)
