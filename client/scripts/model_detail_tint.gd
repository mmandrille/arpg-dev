## Colour a textured model through the material DETAIL layer (a 1x1 texture, Mix blend, alpha =
## strength), which lerps the atlas toward the colour and keeps its shading. A multiply
## (albedo_color) can only darken a coloured atlas. albedo_color stays free for other owners
## (rarity tint, ModelReactionController hit flash); every helper edits the mesh's current
## material_override in place, so they compose. Extracted from ArmorLook (v483) in v490 for the
## kit potion ground models.
class_name ModelDetailTint
extends RefCounted

static var _detail_textures: Dictionary = {}  # "rrggbb:strength" -> ImageTexture


static func set_detail(mesh: MeshInstance3D, color: Color, strength: float) -> void:
	var mat := override_material(mesh)
	if mat == null:
		return
	mat.detail_enabled = true
	mat.detail_blend_mode = BaseMaterial3D.BLEND_MODE_MIX
	mat.detail_albedo = detail_texture(color, strength)


static func clear_detail(mesh: MeshInstance3D) -> void:
	var mat := mesh.material_override as StandardMaterial3D
	if mat != null:
		mat.detail_enabled = false
		mat.detail_albedo = null


## The mesh's current override, created from its surface material when there is none yet.
static func override_material(mesh: MeshInstance3D) -> StandardMaterial3D:
	if mesh.material_override is StandardMaterial3D:
		return mesh.material_override as StandardMaterial3D
	var source: Material = null
	if mesh.mesh != null and mesh.mesh.get_surface_count() > 0:
		source = mesh.get_active_material(0)
	var mat: StandardMaterial3D
	if source is StandardMaterial3D:
		mat = (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	else:
		mat = StandardMaterial3D.new()
	mesh.material_override = mat
	return mat


static func detail_texture(color: Color, strength: float) -> ImageTexture:
	var key := "%s:%.3f" % [color.to_html(false), strength]
	if _detail_textures.has(key):
		return _detail_textures[key]
	var image := Image.create(1, 1, false, Image.FORMAT_RGBA8)
	image.set_pixel(0, 0, Color(color.r, color.g, color.b, strength))
	var texture := ImageTexture.create_from_image(image)
	_detail_textures[key] = texture
	return texture
