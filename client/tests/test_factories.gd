# Headless tests for extracted client presentation factories.
# Run via: godot --headless --path client --script res://tests/test_factories.gd
extends SceneTree

const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const BossVisualsContextScript := preload("res://scripts/boss_visuals_context.gd")
const BossVisualsControllerScript := preload("res://scripts/boss_visuals_controller.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_client_constants()
	_test_ground_wall_factory()
	_test_wall_renderer()
	_test_loot_node_factory()
	_test_town_node_factory()
	_test_boss_visuals()

	if _fail_count > 0:
		print("[gdtest] FAIL: test_factories (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
	print("[gdtest] PASS: test_factories (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(0)


func _test_client_constants() -> void:
	_assert_eq("player start hp", ClientConstantsScript.PLAYER_START_HP, 10)
	_assert_true("rarity color present", ClientConstantsScript.LOOT_LABEL_RARITY_COLORS.has("rare"))


func _test_ground_wall_factory() -> void:
	var factory = GroundWallFactoryScript.new()
	var town_texture := factory.make_ground_texture(ClientConstantsScript.GROUND_TEXTURE_TOWN)
	var dungeon_texture := factory.make_ground_texture(ClientConstantsScript.GROUND_TEXTURE_DUNGEON)
	var water_texture := factory.make_water_texture()
	var hole_texture := factory.make_hole_texture()
	_assert_true("town ground texture exists", town_texture != null)
	_assert_true("dungeon ground texture exists", dungeon_texture != null)
	_assert_true("water texture exists", water_texture != null)
	_assert_true("hole texture exists", hole_texture != null)
	_assert_eq("ground texture cache count", factory.ground_textures.size(), 2)
	_assert_eq("water texture cache count", factory.water_textures.size(), 1)
	_assert_eq("hole texture cache count", factory.hole_textures.size(), 1)
	var ground := factory.make_ground_node(0)
	_assert_eq("ground node name", ground.name, "Ground")
	ground.queue_free()


func _test_wall_renderer() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	var renderer = WallRendererScript.new(root, GroundWallFactoryScript.new())
	renderer.set_level(-4)
	var walls := renderer.render_wall_layout([{
		"id": "test_wall",
		"position": {"x": 2.0, "y": 3.0},
		"size": {"x": 4.0, "y": 1.5},
		"source": "generated",
	}])
	_assert_eq("wall layout count", walls.size(), 1)
	_assert_eq("wall child count", root.get_child_count(), 1)
	var wall := root.get_child(0) as MeshInstance3D
	_assert_eq("wall child name", wall.name, "Wall_test_wall")
	_assert_true("wall material has palette texture", (wall.material_override as StandardMaterial3D).albedo_texture != null)
	var water_walls := renderer.render_wall_layout([{
		"id": "test_water",
		"position": {"x": 7.0, "y": 8.0},
		"size": {"x": 5.0, "y": 2.0},
		"source": "generated",
		"kind": "water",
	}])
	_assert_eq("water layout count", water_walls.size(), 1)
	_assert_eq("water layout kind", str((water_walls[0] as Dictionary).get("kind", "")), "water")
	_assert_eq("water child count", root.get_child_count(), 1)
	var water := root.get_child(0) as MeshInstance3D
	_assert_eq("water child name", water.name, "Water_test_water")
	_assert_eq("water metadata kind", str(water.get_meta("kind", "")), "water")
	_assert_true("water mesh is plane", water.mesh is PlaneMesh)
	_assert_true("water material has texture", (water.material_override as StandardMaterial3D).albedo_texture != null)
	var hole_walls := renderer.render_wall_layout([{
		"id": "test_hole",
		"position": {"x": 9.0, "y": 4.0},
		"size": {"x": 3.0, "y": 2.0},
		"source": "generated",
		"kind": "hole",
	}])
	_assert_eq("hole layout count", hole_walls.size(), 1)
	_assert_eq("hole layout kind", str((hole_walls[0] as Dictionary).get("kind", "")), "hole")
	_assert_eq("hole child count", root.get_child_count(), 1)
	var hole := root.get_child(0) as MeshInstance3D
	_assert_eq("hole child name", hole.name, "Hole_test_hole")
	_assert_eq("hole metadata kind", str(hole.get_meta("kind", "")), "hole")
	_assert_true("hole mesh is plane", hole.mesh is PlaneMesh)
	_assert_true("hole material has texture", (hole.material_override as StandardMaterial3D).albedo_texture != null)
	var lab_walls := renderer.render_world_walls("flying_navigation_lab")
	var lab_water_count := 0
	var lab_hole_count := 0
	for rendered_wall in lab_walls:
		var rendered: Dictionary = rendered_wall
		match str(rendered.get("kind", "wall")):
			"water":
				lab_water_count += 1
			"hole":
				lab_hole_count += 1
	_assert_eq("flying lab wall layout count", lab_walls.size(), 6)
	_assert_eq("flying lab water layout count", lab_water_count, 1)
	_assert_eq("flying lab hole layout count", lab_hole_count, 1)
	_assert_eq("flying lab child count", root.get_child_count(), 6)
	_assert_true("flying lab water node exists", root.find_child("Water_flying_navigation_lab_wall_004", false, false) != null)
	_assert_true("flying lab hole node exists", root.find_child("Hole_flying_navigation_lab_wall_005", false, false) != null)
	root.queue_free()


func _test_loot_node_factory() -> void:
	ItemRulesLoader.ensure_loaded()
	var factory = LootNodeFactoryScript.new({}, ItemRulesLoader.item_presentations)
	_assert_eq("gold label text", factory.loot_label_text({"item_def_id": "gold", "amount": 7}), "7 gold")
	_assert_eq("known loot name", factory.generic_loot_name("rusty_sword"), "Sword")
	var node := factory.make_loot_node({"item_def_id": "rusty_sword", "rarity": "common"})
	_assert_eq("loot node name", node.name, "Loot_rusty_sword")
	_assert_true("loot label exists", node.find_child("LootLabel", true, false) != null)
	node.queue_free()


func _test_town_node_factory() -> void:
	var door := TownNodeFactoryScript.make_door_node()
	_assert_true("door pivot exists", door.find_child("DoorPivot", true, false) != null)
	door.queue_free()
	var chest := TownNodeFactoryScript.make_chest_node("town_stash")
	_assert_true("chest lid pivot exists", chest.find_child("ChestLidPivot", true, false) != null)
	chest.queue_free()
	var preview := TownNodeFactoryScript.make_town_preview_scene()
	_assert_eq("town preview name", preview.name, "TownPreview")
	preview.queue_free()


func _test_boss_visuals() -> void:
	var ctx = BossVisualsContextScript.new()
	ctx.entities = {
		"boss_b": {"type": "monster", "is_boss": true, "hp": 0},
		"boss_a": {"type": "monster", "is_boss": true, "hp": 7, "max_hp": 10, "boss_template_id": "cave_warden"},
	}
	var controller = BossVisualsControllerScript.new(ctx, null)
	_assert_eq("active boss id", controller.active_boss_entity_id(), "boss_a")
	_assert_eq("boss title", controller.boss_health_bar_title("cave_warden"), "Cave Warden")
	var root := Node3D.new()
	get_root().add_child(root)
	var rec := {"node": root, "visual_scale": 1.0, "boss_phase": {"pattern_id": "stone_lance"}}
	controller.sync_boss_telegraph_marker(rec, {"hit_shape": "line", "radius": 7.5, "width": 1.0, "to_color": "#79b8ff"})
	_assert_eq("line marker shape", str(rec.get("telegraph_marker_shape", "")), "line")
	_assert_true("line marker exists", root.find_child(ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME, false, false) != null)
	rec["boss_phase"] = {"pattern_id": "shard_fan"}
	controller.sync_boss_telegraph_marker(rec, {"hit_shape": "cone", "radius": 5.8, "width": 70.0, "to_color": "#ffd166"})
	_assert_eq("cone marker shape", str(rec.get("telegraph_marker_shape", "")), "cone")
	rec["boss_phase"] = {"pattern_id": "crystal_wall"}
	controller.sync_boss_telegraph_marker(rec, {"hit_shape": "rectangle", "radius": 4.8, "width": 2.0, "to_color": "#64f4ff"})
	_assert_eq("rectangle marker shape", str(rec.get("telegraph_marker_shape", "")), "rectangle")
	var marker := root.find_child(ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME, false, false) as MeshInstance3D
	_assert_true("rectangle marker mesh", marker != null and marker.mesh is BoxMesh)
	rec["boss_phase"] = {"pattern_id": "summon_wolves"}
	controller.sync_boss_telegraph_marker(rec, {"hit_shape": "circle", "radius": 3.2, "to_color": "#4fd18b"})
	_assert_eq("summon marker shape", str(rec.get("telegraph_marker_shape", "")), "summon_circle")
	controller.remove_boss_telegraph_marker(rec)
	_assert_eq("marker cleanup", bool(rec.get("has_boss_telegraph_marker", true)), false)
	root.queue_free()


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
