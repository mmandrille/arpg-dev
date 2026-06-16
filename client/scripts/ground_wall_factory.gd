class_name GroundWallFactory
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

var ground_textures: Dictionary = {}
var wall_textures: Dictionary = {}

func make_ground_node(level: int) -> MeshInstance3D:
	var node := MeshInstance3D.new()
	node.name = "Ground"
	var mesh := PlaneMesh.new()
	mesh.size = Vector2(140.0, 90.0)
	mesh.subdivide_width = 32
	mesh.subdivide_depth = 20
	node.mesh = mesh
	node.position = Vector3(50.0, -0.02, 25.0)
	node.material_override = ground_material_for_level(level)
	return node

func update_ground_material(ground_node: MeshInstance3D, level: int) -> void:
	if ground_node == null:
		return
	ground_node.material_override = ground_material_for_level(level)

func ground_texture_id_for_level(level: int) -> String:
	return ClientConstantsScript.GROUND_TEXTURE_TOWN if level == 0 else ClientConstantsScript.GROUND_TEXTURE_DUNGEON

func ground_material_for_level(level: int) -> StandardMaterial3D:
	var texture_id := ground_texture_id_for_level(level)
	var mat := StandardMaterial3D.new()
	mat.albedo_texture = make_ground_texture(texture_id)
	mat.albedo_color = Color.WHITE
	mat.roughness = 0.92
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.uv1_scale = Vector3(28.0, 18.0, 1.0)
	return mat

func make_ground_texture(texture_id: String) -> ImageTexture:
	if ground_textures.has(texture_id):
		return ground_textures[texture_id] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, ground_texel(texture_id, x, y))
	var texture := ImageTexture.create_from_image(image)
	ground_textures[texture_id] = texture
	return texture

func ground_texel(texture_id: String, x: int, y: int) -> Color:
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
	var rock := Color("#3c3f43").lerp(Color("#73706b"), float(n) / 16.0)
	if abs((x % 16) - (y % 16)) <= 1:
		rock = rock.lerp(Color("#25282c"), 0.35)
	if ((x * 3 + y * 5) % 29) == 0:
		rock = rock.lerp(Color("#a09a8e"), 0.28)
	return rock

func make_wall_texture(texture_id: String) -> ImageTexture:
	if wall_textures.has(texture_id):
		return wall_textures[texture_id] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, wall_texel(texture_id, x, y))
	var texture := ImageTexture.create_from_image(image)
	wall_textures[texture_id] = texture
	return texture

func wall_texel(_texture_id: String, x: int, y: int) -> Color:
	var brick_w := 16
	var brick_h := 12
	var row := int(y / brick_h)
	var offset := 8 if row % 2 == 1 else 0
	var local_x := int((x + offset) % brick_w)
	var local_y := int(y % brick_h)
	var noise := int((x * 29 + y * 43 + row * 17 + int((x + offset) / brick_w) * 13) % 23)
	var stone := Color("#34363a").lerp(Color("#6b6255"), float(noise) / 22.0)
	if local_x <= 1 or local_y <= 1:
		return stone.lerp(Color("#17191c"), 0.62)
	if local_x >= brick_w - 2 or local_y >= brick_h - 2:
		stone = stone.lerp(Color("#202226"), 0.34)
	if ((x * 5 + y * 7) % 31) == 0:
		stone = stone.lerp(Color("#9b9386"), 0.32)
	if ((x - y) % 19) == 0:
		stone = stone.lerp(Color("#22252a"), 0.30)
	return stone
