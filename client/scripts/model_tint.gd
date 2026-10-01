## Rarity/visual tint for imported models (ADR-0018 P4a fix): multiplies a copy of the mesh's own
## material so albedo textures survive. The previous tint replaced every material with a flat,
## untextured StandardMaterial3D, which is why GLB monsters rendered as solid colors.
class_name ModelTint
extends RefCounted

const MAX_CACHED_TINTS := 64
static var _cached_tints: Dictionary = {}

static func tinted_material(mesh: MeshInstance3D, color: Color) -> StandardMaterial3D:
	var source: Material = mesh.material_override
	if source == null and mesh.mesh != null and mesh.mesh.get_surface_count() > 0:
		source = mesh.get_active_material(0)
	var key := "%s:%s" % [str(source.get_instance_id()) if source != null else "none", color.to_html(true)]
	if _cached_tints.has(key):
		return _cached_tints[key] as StandardMaterial3D
	var mat: StandardMaterial3D
	if source is StandardMaterial3D:
		mat = (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	else:
		mat = StandardMaterial3D.new()
	mat.albedo_color = color
	if _cached_tints.size() >= MAX_CACHED_TINTS:
		_cached_tints.clear()
	_cached_tints[key] = mat
	return mat
