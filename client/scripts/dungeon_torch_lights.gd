## Client-only dungeon perimeter torch meshes and local lights.
class_name DungeonTorchLights
extends RefCounted

const PlacementScript := preload("res://scripts/dungeon_torch_placement.gd")
const LoaderScript := preload("res://scripts/dungeon_torch_presentation_loader.gd")

var _parent: Node3D
var _root: Node3D
var _fog_overlay: FogOfWarOverlay
var _ground_factory: GroundWallFactory
var _wall_renderer: WallRenderer
var _active := false
var _positions: Array = []


func _init(parent: Node3D, fog_overlay: FogOfWarOverlay, ground_factory: GroundWallFactory, wall_renderer: WallRenderer) -> void:
	_parent = parent
	_fog_overlay = fog_overlay
	_ground_factory = ground_factory
	_wall_renderer = wall_renderer


func sync(level: int, walls: Array, dungeon_active: bool) -> void:
	LoaderScript.ensure_loaded()
	var cfg := LoaderScript.config()
	var should_show := dungeon_active and level < 0 and bool(cfg.get("enabled", true))
	var placements := PlacementScript.placements_from_walls(walls, cfg, level) if should_show else []
	_ensure_root()
	if placements.size() == _positions.size() and should_show == _active:
		var same := true
		for i in placements.size():
			if placements[i] != _positions[i]:
				same = false
				break
		if same:
			return
	_clear_torches()
	_positions = placements
	_active = should_show and not placements.is_empty()
	if not _active:
		if _fog_overlay != null:
			_fog_overlay.set_torch_lights([], 0.0, 0.0)

		return
	var wall_height := _mount_wall_height()
	var mount_height := wall_height * float(cfg.get("mount_height_fraction", 0.72))
	for i in placements.size():
		_spawn_torch(i, placements[i] as Vector2, mount_height, cfg)
	if _fog_overlay != null:
		_fog_overlay.set_torch_lights(
			placements,
			float(cfg.get("fog_light_radius", 5.0)),
			float(cfg.get("torch_feather_world", 0.35)),
		)


func get_debug_state() -> Dictionary:
	LoaderScript.ensure_loaded()
	var cfg := LoaderScript.config()

	return {
		"active": _active,
		"count": _positions.size(),
		"light_radius": float(cfg.get("fog_light_radius", 5.0)) if _active else 0.0,
		"shader_torch_cap": int(cfg.get("max_shader_torches", 32)),
	}


func clear() -> void:
	_clear_torches()
	_positions = []
	_active = false
	if _fog_overlay != null:
		_fog_overlay.set_torch_lights([], 0.0, 0.0)


func _ensure_root() -> void:
	if _root != null and is_instance_valid(_root):
		return
	if _parent == null:
		return
	_root = Node3D.new()
	_root.name = "DungeonTorchLights"
	_parent.add_child(_root)


func _clear_torches() -> void:
	if _root == null or not is_instance_valid(_root):
		return
	for child in _root.get_children():
		child.queue_free()


func _spawn_torch(index: int, xz: Vector2, mount_height: float, cfg: Dictionary) -> void:
	if _root == null:
		return
	var torch := Node3D.new()
	torch.name = "Torch_%03d" % index
	torch.position = Vector3(xz.x, mount_height, xz.y)
	var flame_color := Color(str(cfg.get("flame_color", "#ff7a1a")))
	var emission_color := Color(str(cfg.get("flame_emission_color", "#ffd45a")))
	var emission_energy := float(cfg.get("flame_emission_energy", 2.4))
	var bw := float(cfg.get("bracket_width", 0.18))
	var bh := float(cfg.get("bracket_height", 0.38))
	var bracket := MeshInstance3D.new()
	bracket.name = "Bracket"
	var bracket_mesh := BoxMesh.new()
	bracket_mesh.size = Vector3(bw, bh, bw)
	bracket.mesh = bracket_mesh
	var bracket_mat := StandardMaterial3D.new()
	bracket_mat.albedo_color = Color("#4e4030")
	bracket_mat.roughness = 0.95
	bracket.material_override = bracket_mat
	bracket.position = Vector3(0.0, -bh * 0.4, 0.0)
	torch.add_child(bracket)
	var fr_top := float(cfg.get("flame_radius_top", 0.16))
	var fr_bot := float(cfg.get("flame_radius_bottom", 0.22))
	var fh := float(cfg.get("flame_height", 0.50))
	var flame := MeshInstance3D.new()
	flame.name = "Flame"
	var flame_mesh := CylinderMesh.new()
	flame_mesh.top_radius = fr_top
	flame_mesh.bottom_radius = fr_bot
	flame_mesh.height = fh
	flame.mesh = flame_mesh
	var flame_mat := StandardMaterial3D.new()
	flame_mat.albedo_color = flame_color
	flame_mat.emission_enabled = true
	flame_mat.emission = emission_color
	flame_mat.emission_energy_multiplier = emission_energy
	flame.material_override = flame_mat
	flame.position = Vector3(0.0, fh * 0.25, 0.0)
	torch.add_child(flame)
	if bool(cfg.get("omni_light_enabled", false)):
		var light := OmniLight3D.new()
		light.name = "TorchLight"
		light.light_color = Color(str(cfg.get("light_color", "#ff9b3d")))
		light.light_energy = float(cfg.get("omni_energy", 0.85))
		light.omni_range = float(cfg.get("omni_range", 5.0))
		light.omni_attenuation = 1.8
		light.position = Vector3(0.0, 0.12, 0.0)
		torch.add_child(light)
	_root.add_child(torch)


func _mount_wall_height() -> float:
	if _wall_renderer != null:
		return maxf(1.0, _wall_renderer.wall_height())
	if _ground_factory != null and _ground_factory.has_method("dungeon_ceiling_height"):
		return maxf(1.0, float(_ground_factory.dungeon_ceiling_height()))

	return 4.0
