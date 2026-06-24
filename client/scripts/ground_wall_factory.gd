class_name GroundWallFactory
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

const TOWN_GROUND_SIZE := Vector2(140.0, 90.0)
const TOWN_GROUND_CENTER := Vector3(50.0, -0.02, 25.0)
const DUNGEON_GROUND_MARGIN := 8.0

var ground_textures: Dictionary = {}
var wall_textures: Dictionary = {}
var ground_normal_textures: Dictionary = {}
var wall_normal_textures: Dictionary = {}
var water_textures: Dictionary = {}
var hole_textures: Dictionary = {}
var _dungeon_generation: Dictionary = {}

func make_ground_node(level: int) -> MeshInstance3D:
	var node := MeshInstance3D.new()
	node.name = "Ground"
	var mesh := PlaneMesh.new()
	node.mesh = mesh
	configure_ground_node(node, level)
	node.material_override = ground_material_for_level(level)
	return node


func configure_ground_node(ground_node: MeshInstance3D, level: int) -> void:
	if ground_node == null:
		return
	var mesh := ground_node.mesh as PlaneMesh
	if mesh == null:
		return
	if level >= 0:
		mesh.size = TOWN_GROUND_SIZE
		mesh.subdivide_width = 32
		mesh.subdivide_depth = 20
		ground_node.position = TOWN_GROUND_CENTER
		return
	var layout := dungeon_ground_layout(level)
	mesh.size = layout.size
	mesh.subdivide_width = maxi(8, int(layout.size.x / 4.0))
	mesh.subdivide_depth = maxi(8, int(layout.size.y / 4.0))
	ground_node.position = layout.center


func dungeon_ground_layout(level: int) -> Dictionary:
	var floor_size := floor_size_for_level(level)
	var margin := DUNGEON_GROUND_MARGIN + wall_thickness()
	var size := floor_size + Vector2(margin * 2.0, margin * 2.0)
	return {
		"size": size,
		"center": Vector3(floor_size.x * 0.5, 0.0, floor_size.y * 0.5),
	}


func wall_thickness() -> float:
	_ensure_dungeon_generation()
	return maxf(0.1, float(_dungeon_generation.get("wall_thickness", 1.0)))


func update_ground_material(ground_node: MeshInstance3D, level: int) -> void:
	if ground_node == null:
		return
	configure_ground_node(ground_node, level)
	ground_node.material_override = ground_material_for_level(level)
	if level == 0:
		TownAmbientLife.attach_to_town(ground_node)

func ground_texture_id_for_level(level: int) -> String:
	return ClientConstantsScript.GROUND_TEXTURE_TOWN if level == 0 else ClientConstantsScript.GROUND_TEXTURE_DUNGEON

func ground_material_for_level(level: int) -> StandardMaterial3D:
	var texture_id := ground_texture_id_for_level(level)
	var palette := biome_palette_for_level(level)
	var mat := StandardMaterial3D.new()
	mat.albedo_texture = make_ground_texture(texture_id, palette)
	mat.albedo_color = Color.WHITE
	mat.roughness = 0.88 if level != 0 else 0.92
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.uv1_scale = Vector3(28.0, 18.0, 1.0)
	if level != 0:
		mat.normal_enabled = true
		mat.normal_texture = make_ground_normal_texture(texture_id, palette)
		mat.normal_scale = 0.18
	return mat

func make_ground_texture(texture_id: String, palette: Dictionary = {}) -> ImageTexture:
	var cache_key := "%s:%s" % [texture_id, str(palette.get("id", "default"))]
	if ground_textures.has(cache_key):
		return ground_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, ground_texel(texture_id, x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	ground_textures[cache_key] = texture
	return texture

func make_ground_normal_texture(texture_id: String, palette: Dictionary = {}) -> ImageTexture:
	var cache_key := "%s:%s" % [texture_id, str(palette.get("id", "default"))]
	if ground_normal_textures.has(cache_key):
		return ground_normal_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, ground_normal_texel(texture_id, x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	ground_normal_textures[cache_key] = texture
	return texture

func ground_texel(texture_id: String, x: int, y: int, palette: Dictionary = {}) -> Color:
	var n := int((x * 37 + y * 19 + ((x / 8) * 11) + ((y / 8) * 23)) % 17)
	if texture_id == ClientConstantsScript.GROUND_TEXTURE_TOWN:
		var base := Color("#2f6136").lerp(Color("#79aa58"), float(n) / 16.0)
		var dirt_patch := int((x * 9 + y * 5 + ((x / 8) * 17) + ((y / 8) * 29)) % 41)
		if dirt_patch < 8:
			base = base.lerp(Color("#8f7447"), 0.50)
		elif dirt_patch < 13:
			base = base.lerp(Color("#6f5f39"), 0.30)
		if ((x * 5 + y * 3) % 23) == 0:
			base = base.lerp(Color("#b7b56c"), 0.34)
		if ((x * 11 + y * 7) % 31) == 0:
			base = base.lerp(Color("#274c2b"), 0.36)
		if ((x / 8 + y / 8) % 5) == 0 and ((x * 13 + y * 17) % 19) < 3:
			base = base.lerp(Color("#456f35"), 0.20)
		return base
	var rock := _palette_color(palette, "ground_low", Color("#3c3f43")).lerp(_palette_color(palette, "ground_high", Color("#73706b")), float(n) / 16.0)
	if abs((x % 16) - (y % 16)) <= 1:
		rock = rock.lerp(_palette_color(palette, "ground_crack", Color("#25282c")), 0.35)
	if ((x * 3 + y * 5) % 29) == 0:
		rock = rock.lerp(_palette_color(palette, "ground_highlight", Color("#a09a8e")), 0.28)
	return rock

func ground_normal_texel(texture_id: String, x: int, y: int, palette: Dictionary = {}) -> Color:
	var left := ground_detail_height(texture_id, (x + 63) % 64, y, palette)
	var right := ground_detail_height(texture_id, (x + 1) % 64, y, palette)
	var down := ground_detail_height(texture_id, x, (y + 63) % 64, palette)
	var up := ground_detail_height(texture_id, x, (y + 1) % 64, palette)
	return _normal_texel(right - left, up - down, 1.35)

func ground_detail_height(texture_id: String, x: int, y: int, _palette: Dictionary = {}) -> float:
	var n := float(int((x * 37 + y * 19 + ((x / 8) * 11) + ((y / 8) * 23)) % 17)) / 16.0
	if texture_id == ClientConstantsScript.GROUND_TEXTURE_TOWN:
		var dirt := 0.35 if int((x * 9 + y * 5 + ((x / 8) * 17) + ((y / 8) * 29)) % 41) < 8 else 0.0
		return clampf(n * 0.38 + dirt, 0.0, 1.0)
	var crack := 0.46 if abs((x % 16) - (y % 16)) <= 1 else 0.0
	var highlight := 0.22 if ((x * 3 + y * 5) % 29) == 0 else 0.0
	return clampf(0.42 + n * 0.34 - crack + highlight, 0.0, 1.0)

func make_wall_texture(texture_id: String, palette: Dictionary = {}) -> ImageTexture:
	var cache_key := "%s:%s" % [texture_id, str(palette.get("id", "default"))]
	if wall_textures.has(cache_key):
		return wall_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, wall_texel(texture_id, x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	wall_textures[cache_key] = texture
	return texture

func make_wall_normal_texture(texture_id: String, palette: Dictionary = {}) -> ImageTexture:
	var cache_key := "%s:%s" % [texture_id, str(palette.get("id", "default"))]
	if wall_normal_textures.has(cache_key):
		return wall_normal_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, wall_normal_texel(texture_id, x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	wall_normal_textures[cache_key] = texture
	return texture

func wall_material_for_level(level: int, source: String, size: Dictionary, wall_height: float = 1.0) -> StandardMaterial3D:
	var palette := biome_palette_for_level(level)
	var mat := StandardMaterial3D.new()
	mat.albedo_texture = make_wall_texture(ClientConstantsScript.WALL_TEXTURE_CAVE, palette)
	mat.normal_enabled = true
	mat.normal_texture = make_wall_normal_texture(ClientConstantsScript.WALL_TEXTURE_CAVE, palette)
	mat.normal_scale = 0.24
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.94
	mat.uv1_scale = Vector3(maxf(1.0, float(size.get("x", 1.0)) / 2.0), maxf(1.0, wall_height / 2.0), 1.0)
	match source:
		"generated":
			mat.albedo_color = Color(0.92, 0.86, 0.76)
		"perimeter":
			mat.albedo_color = Color(0.62, 0.64, 0.68)
		_:
			mat.albedo_color = Color(0.78, 0.80, 0.82)
	return mat

func make_water_texture(palette: Dictionary = {}) -> ImageTexture:
	var cache_key := str(palette.get("id", "default"))
	if water_textures.has(cache_key):
		return water_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, water_texel(x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	water_textures[cache_key] = texture
	return texture

func water_texel(x: int, y: int, palette: Dictionary = {}) -> Color:
	var low := _palette_color(palette, "water_low", Color("#1f5f7a"))
	var high := _palette_color(palette, "water_high", Color("#5fa8bb"))
	var ripple := int((x * 17 + y * 31 + ((x / 8) * 13) + ((y / 8) * 7)) % 19)
	var water := low.lerp(high, float(ripple) / 18.0)
	if abs((x % 16) - (y % 16)) <= 1:
		water = water.lerp(_palette_color(palette, "water_foam", Color("#a9d8df")), 0.24)
	if ((x * 11 + y * 5) % 37) == 0:
		water = water.lerp(_palette_color(palette, "water_foam", Color("#c2eef1")), 0.34)
	return water

func make_hole_texture(palette: Dictionary = {}) -> ImageTexture:
	var cache_key := str(palette.get("id", "default"))
	if hole_textures.has(cache_key):
		return hole_textures[cache_key] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, hole_texel(x, y, palette))
	var texture := ImageTexture.create_from_image(image)
	hole_textures[cache_key] = texture
	return texture

func hole_texel(x: int, y: int, palette: Dictionary = {}) -> Color:
	var void_low := _palette_color(palette, "hole_low", Color("#090a0d"))
	var void_high := _palette_color(palette, "hole_high", Color("#1b1d22"))
	var edge := _palette_color(palette, "hole_edge", Color("#46413a"))
	var noise := int((x * 23 + y * 41 + ((x / 8) * 19) + ((y / 8) * 11)) % 23)
	var hole := void_low.lerp(void_high, float(noise) / 22.0)
	var border: int = mini(mini(x, y), mini(63 - x, 63 - y))
	if border < 7:
		hole = hole.lerp(edge, 0.55 - float(border) * 0.06)
	if abs((x % 18) - (y % 18)) <= 1:
		hole = hole.lerp(_palette_color(palette, "hole_crack", Color("#26282d")), 0.24)
	return hole

func wall_texel(_texture_id: String, x: int, y: int, palette: Dictionary = {}) -> Color:
	var brick_w := 16
	var brick_h := 12
	var row := int(y / brick_h)
	var offset := 8 if row % 2 == 1 else 0
	var local_x := int((x + offset) % brick_w)
	var local_y := int(y % brick_h)
	var noise := int((x * 29 + y * 43 + row * 17 + int((x + offset) / brick_w) * 13) % 23)
	var stone := _palette_color(palette, "wall_base", Color("#34363a")).lerp(_palette_color(palette, "wall_highlight", Color("#6b6255")), float(noise) / 22.0)
	if local_x <= 1 or local_y <= 1:
		return stone.lerp(_palette_color(palette, "wall_mortar", Color("#17191c")), 0.62)
	if local_x >= brick_w - 2 or local_y >= brick_h - 2:
		stone = stone.lerp(_palette_color(palette, "wall_mortar", Color("#202226")), 0.34)
	if ((x * 5 + y * 7) % 31) == 0:
		stone = stone.lerp(_palette_color(palette, "wall_highlight", Color("#9b9386")), 0.32)
	if ((x - y) % 19) == 0:
		stone = stone.lerp(_palette_color(palette, "wall_mortar", Color("#22252a")), 0.30)
	return stone

func wall_normal_texel(_texture_id: String, x: int, y: int, palette: Dictionary = {}) -> Color:
	var left := wall_detail_height((x + 63) % 64, y, palette)
	var right := wall_detail_height((x + 1) % 64, y, palette)
	var down := wall_detail_height(x, (y + 63) % 64, palette)
	var up := wall_detail_height(x, (y + 1) % 64, palette)
	return _normal_texel(right - left, up - down, 1.55)

func wall_detail_height(x: int, y: int, _palette: Dictionary = {}) -> float:
	var brick_w := 16
	var brick_h := 12
	var row := int(y / brick_h)
	var offset := 8 if row % 2 == 1 else 0
	var local_x := int((x + offset) % brick_w)
	var local_y := int(y % brick_h)
	var noise := float(int((x * 29 + y * 43 + row * 17 + int((x + offset) / brick_w) * 13) % 23)) / 22.0
	if local_x <= 1 or local_y <= 1:
		return 0.08
	if local_x >= brick_w - 2 or local_y >= brick_h - 2:
		return 0.26 + noise * 0.18
	var chip := 0.18 if ((x * 5 + y * 7) % 31) == 0 else 0.0
	var vein := 0.12 if ((x - y) % 19) == 0 else 0.0
	return clampf(0.46 + noise * 0.32 + chip - vein, 0.0, 1.0)

func biome_palette_for_level(level: int) -> Dictionary:
	var depth: int = abs(level)
	if depth <= 0:
		return {}
	_ensure_dungeon_generation()
	for raw in _dungeon_generation.get("biome_palettes", []):
		var palette := raw as Dictionary
		var min_depth := int(palette.get("min_depth", 1))
		var max_value = palette.get("max_depth", null)
		if depth < min_depth:
			continue
		if max_value != null and depth > int(max_value):
			continue
		return palette
	return {}


func dungeon_ceiling_height() -> float:
	_ensure_dungeon_generation()
	return float(_dungeon_generation.get("ceiling_height", 4.0))


func floor_size_for_level(level: int) -> Vector2:
	_ensure_dungeon_generation()
	if level >= 0:
		return Vector2.ZERO
	if _is_boss_floor(level):
		var boss: Dictionary = _dungeon_generation.get("boss_floor", {})
		var boss_size: Dictionary = boss.get("floor_size", {})
		return Vector2(float(boss_size.get("width", 30.0)), float(boss_size.get("height", 30.0)))
	var size: Dictionary = _dungeon_generation.get("floor_size", {})
	var width := float(size.get("width", 100.0))
	var height := float(size.get("height", 50.0))
	var depth := absi(level)
	for raw in _dungeon_generation.get("floor_profiles", []):
		var profile := raw as Dictionary
		var min_depth := int(profile.get("min_depth", 1))
		var max_value = profile.get("max_depth", null)
		if depth < min_depth:
			continue
		if max_value != null and depth > int(max_value):
			continue
		var profile_size: Dictionary = profile.get("floor_size", {})
		width = float(profile_size.get("width", width))
		height = float(profile_size.get("height", height))
		break
	return Vector2(width, height)


func make_ceiling_node(level: int) -> MeshInstance3D:
	var floor_size := floor_size_for_level(level)
	var ceiling_height := dungeon_ceiling_height()
	var node := MeshInstance3D.new()
	node.name = "DungeonCeiling"
	node.set_meta("kind", "ceiling")
	var mesh := PlaneMesh.new()
	mesh.size = floor_size
	node.mesh = mesh
	node.position = Vector3(floor_size.x * 0.5, ceiling_height, floor_size.y * 0.5)
	node.rotation_degrees = Vector3(180.0, 0.0, 0.0)
	node.material_override = ceiling_material_for_level(level)
	return node


func ceiling_material_for_level(level: int) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	var palette: Dictionary = biome_palette_for_level(level)
	if palette.is_empty():
		mat.albedo_color = Color(0.22, 0.24, 0.26)
	else:
		mat.albedo_color = _palette_color(palette, "wall_base", Color(0.22, 0.24, 0.26)).darkened(0.18)
	if not palette.is_empty():
		mat.albedo_texture = make_wall_texture(ClientConstantsScript.WALL_TEXTURE_CAVE, palette)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.98
	var floor_size := floor_size_for_level(level)
	mat.uv1_scale = Vector3(maxf(1.0, floor_size.x / 8.0), maxf(1.0, floor_size.y / 8.0), 1.0)
	return mat


func _is_boss_floor(level: int) -> bool:
	if level >= 0:
		return false
	var boss: Dictionary = _dungeon_generation.get("boss_floor", {})
	var cadence := int(boss.get("cadence", 0))
	if cadence <= 0:
		return false
	return absi(level) % cadence == 0

func _ensure_dungeon_generation() -> void:
	if not _dungeon_generation.is_empty():
		return
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/dungeon_generation.v0.json")
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_dungeon_generation = parsed

func _palette_color(palette: Dictionary, key: String, fallback: Color) -> Color:
	if not palette.has(key):
		return fallback
	return Color(str(palette.get(key, "#" + fallback.to_html(false))))

func _normal_texel(dx: float, dy: float, strength: float) -> Color:
	var normal := Vector3(-dx * strength, -dy * strength, 1.0).normalized()
	return Color(normal.x * 0.5 + 0.5, normal.y * 0.5 + 0.5, normal.z * 0.5 + 0.5)
