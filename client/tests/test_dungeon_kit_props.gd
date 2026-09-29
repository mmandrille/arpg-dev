extends SceneTree
## ADR-0018 P2b (v472): KayKit wall torches and treasure chests.

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const PropsScript := preload("res://scripts/dungeon_kit_props.gd")
const PlacementScript := preload("res://scripts/dungeon_torch_placement.gd")
const TorchLightsScript := preload("res://scripts/dungeon_torch_lights.gd")
const TorchLoaderScript := preload("res://scripts/dungeon_torch_presentation_loader.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_torch_yaw_points_local_z_at_facing()
	_test_mounts_face_the_room_and_match_placements()
	_test_torch_lights_use_kit_body_on_kit_levels()
	_test_torch_lights_keep_bracket_when_kit_disabled()
	_test_kit_chest_contract()
	_test_elite_chest_uses_gold_variant()
	_test_factory_routes_only_treasure_chests()
	_finish()


func _room_walls() -> Array:
	return [
		{"id": "n", "position": {"x": 10.0, "y": -0.5}, "size": {"x": 22.0, "y": 1.0}, "source": "perimeter"},
		{"id": "s", "position": {"x": 10.0, "y": 20.5}, "size": {"x": 22.0, "y": 1.0}, "source": "perimeter"},
		{"id": "w", "position": {"x": -0.5, "y": 10.0}, "size": {"x": 1.0, "y": 20.0}, "source": "perimeter"},
		{"id": "e", "position": {"x": 20.5, "y": 10.0}, "size": {"x": 1.0, "y": 20.0}, "source": "perimeter"},
	]


func _torch_cfg() -> Dictionary:
	TorchLoaderScript.ensure_loaded()
	var cfg := TorchLoaderScript.config()
	cfg["torches_per_segment_min"] = 1  # guarantee mounts on every wall for the test
	return cfg


func _test_torch_yaw_points_local_z_at_facing() -> void:
	for facing in [Vector2(0, 1), Vector2(0, -1), Vector2(1, 0), Vector2(-1, 0)]:
		var z := Basis(Vector3.UP, PropsScript.torch_yaw(facing)) * Vector3(0, 0, 1)
		if not Vector2(z.x, z.z).is_equal_approx(facing):
			_fail("torch yaw for %s points local +Z at %s" % [str(facing), str(z)])
			return
	_pass("torch yaw turns local +Z toward the facing")


func _test_mounts_face_the_room_and_match_placements() -> void:
	var cfg := _torch_cfg()
	var mounts := PlacementScript.mounts_from_walls(_room_walls(), cfg, -2)
	var placements := PlacementScript.placements_from_walls(_room_walls(), cfg, -2)
	if mounts.is_empty() or mounts.size() != placements.size():
		_fail("mounts must exist and match placements (%d vs %d)" % [mounts.size(), placements.size()])
		return
	var center := Vector2(10.0, 10.0)
	for i in mounts.size():
		var mount: Dictionary = mounts[i]
		if mount["position"] != placements[i]:
			_fail("mount %d position differs from placement" % i)
			return
		var to_center := (center - (mount["position"] as Vector2)).normalized()
		if (mount["facing"] as Vector2).dot(to_center) <= 0.0:
			_fail("mount %d faces away from the room" % i)
			return
	_pass("torch mounts face the room and match placements")


func _torch_nodes(level: int) -> Array:
	var parent := Node3D.new()
	get_root().add_child(parent)
	var lights = TorchLightsScript.new(parent, null, GroundWallFactoryScript.new(), null)
	lights.sync(level, _room_walls(), true)
	var torches: Array = []
	for child in parent.get_children():
		for torch in child.get_children():
			torches.append(torch)
	return [parent, torches]


func _test_torch_lights_use_kit_body_on_kit_levels() -> void:
	var result := _torch_nodes(-2)
	var parent: Node3D = result[0]
	var torches: Array = result[1]
	var ok := not torches.is_empty()
	for torch in torches:
		var node := torch as Node3D
		if node.find_child("KitTorch", false, false) == null or node.find_child("Bracket", false, false) != null:
			ok = false
		if node.find_child("Flame", false, false) == null:
			ok = false
	parent.free()
	if not ok:
		_fail("kit levels must spawn KitTorch bodies with flames and no procedural bracket")
		return
	_pass("kit torches replace the procedural bracket")


func _test_torch_lights_keep_bracket_when_kit_disabled() -> void:
	LoaderScript.enabled_override = "off"
	var result := _torch_nodes(-2)
	LoaderScript.enabled_override = ""
	var parent: Node3D = result[0]
	var torches: Array = result[1]
	var ok := not torches.is_empty()
	for torch in torches:
		if (torch as Node3D).find_child("KitTorch", false, false) != null or (torch as Node3D).find_child("Bracket", false, false) == null:
			ok = false
	parent.free()
	if not ok:
		_fail("disabled kit must keep the procedural torch bracket")
		return
	_pass("disabled kit keeps legacy torches")


func _lid_top(chest: Node3D) -> float:
	var pivot := chest.find_child(PropsScript.LID_PIVOT_NAME, true, false) as Node3D
	var top := -INF
	for mi in LibraryScript.mesh_instances(pivot):
		var xf := LibraryScript._transform_to(chest, mi as Node3D)
		top = maxf(top, (xf * (mi as MeshInstance3D).get_aabb()).end.y)
	return top


func _test_kit_chest_contract() -> void:
	var chest := PropsScript.make_chest_node("treasure_chest")
	if chest == null:
		_fail("kit chest must build")
		return
	var pivot := chest.find_child(PropsScript.LID_PIVOT_NAME, true, false) as Node3D
	var glow := chest.find_child(PropsScript.INNER_GLOW_NAME, true, false) as MeshInstance3D
	if chest.name != "TreasureChest" or pivot == null or glow == null or glow.visible:
		_fail("kit chest keeps TreasureChest name, ChestLidPivot and a hidden ChestInnerGlow")
		chest.free()
		return
	var closed_top := _lid_top(chest)
	pivot.rotation.x = deg_to_rad(-68.0)  # same rotation InteractableStatePresentation tweens to
	var open_top := _lid_top(chest)
	chest.free()
	if open_top <= closed_top:
		_fail("opening the kit lid must raise it (closed %.2f, open %.2f)" % [closed_top, open_top])
		return
	_pass("kit chest keeps the open-lid contract")


func _test_elite_chest_uses_gold_variant() -> void:
	var cfg := LoaderScript.chest_config()
	var plain := PropsScript.make_chest_node("treasure_chest", false)
	var elite := PropsScript.make_chest_node("treasure_chest", true)
	var plain_path := str(plain.get_child(0).scene_file_path)
	var elite_path := str(elite.get_child(0).scene_file_path)
	var has_marker := elite.find_child("EliteObjectiveMarker", true, false) != null
	var elite_has_lid := elite.find_child(PropsScript.LID_PIVOT_NAME, true, false) != null
	plain.free()
	elite.free()
	if plain_path != LibraryScript.res_path(str(cfg["asset_id"])) or elite_path != LibraryScript.res_path(str(cfg["elite_objective_asset_id"])):
		_fail("elite objective chest must use the gold kit variant (%s / %s)" % [plain_path, elite_path])
		return
	if not has_marker:
		_fail("elite objective chest keeps its objective marker")
		return
	if not elite_has_lid:
		_fail("elite objective chest must expose its (gold) lid as ChestLidPivot")
		return
	_pass("elite objective chest uses the gold variant")


func _test_factory_routes_only_treasure_chests() -> void:
	var treasure := TownNodeFactoryScript.make_chest_node("treasure_chest")
	var stash := TownNodeFactoryScript.make_chest_node("town_stash")
	var ok := treasure.has_meta("kit_chest") and not stash.has_meta("kit_chest")
	treasure.free()
	stash.free()
	if not ok:
		_fail("only treasure chests use the kit model")
		return
	_pass("factory routes treasure chests to the kit model")


func _pass(msg: String) -> void:
	_pass_count += 1
	print("  ok   %s" % msg)


func _fail(msg: String) -> void:
	_fail_count += 1
	push_error("FAIL: %s" % msg)
	print("  FAIL %s" % msg)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_dungeon_kit_props (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_dungeon_kit_props (%d checks)" % _pass_count)
	quit(0)
