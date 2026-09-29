## Rarity/visual tint for imported models (ADR-0018 P4a fix): multiplies a copy of the mesh's own
## material so albedo textures survive. The previous tint replaced every material with a flat,
## untextured StandardMaterial3D, which is why GLB monsters rendered as solid colors.
class_name ModelTint
extends RefCounted


static func tinted_material(mesh: MeshInstance3D, color: Color) -> StandardMaterial3D:
	var source: Material = mesh.material_override
	if source == null and mesh.mesh != null and mesh.mesh.get_surface_count() > 0:
		source = mesh.get_active_material(0)
	var mat: StandardMaterial3D
	if source is StandardMaterial3D:
		mat = (source as StandardMaterial3D).duplicate() as StandardMaterial3D
	else:
		mat = StandardMaterial3D.new()
	mat.albedo_color = color
	return mat
