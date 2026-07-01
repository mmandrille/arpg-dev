# Headless tests for extracted client presentation factories.
# Run via: godot --headless --path client --script res://tests/test_factories.gd
extends SceneTree

const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const BossArenaPresenceScript := preload("res://scripts/boss_arena_presence.gd")
const BossVisualsContextScript := preload("res://scripts/boss_visuals_context.gd")
const BossVisualsControllerScript := preload("res://scripts/boss_visuals_controller.gd")
const DungeonRoomFloorTintScript := preload("res://scripts/dungeon_room_floor_tint.gd")
const DungeonAmbientMotesScript := preload("res://scripts/dungeon_ambient_motes.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_client_constants()
	_test_ground_wall_factory()
	_test_wall_renderer()
	_test_loot_node_factory()
	_test_town_node_factory()
	_test_boss_visuals()
	_test_dungeon_room_floor_tint()

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
	var town_mat := factory.ground_material_for_level(0)
	var dungeon_mat := factory.ground_material_for_level(-1)
	var dungeon_palette: Dictionary = factory.biome_palette_for_level(-1)
	var dungeon_normal_a := factory.make_ground_normal_texture(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, dungeon_palette)
	var dungeon_normal_b := factory.make_ground_normal_texture(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, dungeon_palette)
	_assert_true("town ground normal remains disabled", not town_mat.normal_enabled)
	_assert_true("dungeon ground normal enabled", dungeon_mat.normal_enabled)
	_assert_true("dungeon ground normal texture exists", dungeon_mat.normal_texture != null)
	_assert_true("dungeon ground normal cache stable", dungeon_normal_a == dungeon_normal_b)
	_assert_eq("ground normal texture cache count", factory.ground_normal_textures.size(), 1)
	_assert_true("dungeon ground normal texel varies", factory.ground_normal_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 0, 0, dungeon_palette) != factory.ground_normal_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 17, 11, dungeon_palette))
	var ground := factory.make_ground_node(0)
	_assert_eq("ground node name", ground.name, "Ground")
	var deep_layout: Dictionary = factory.dungeon_ground_layout(-6)
	_assert_true("deep dungeon ground extends past floor width", float(deep_layout.get("size", Vector2.ZERO).x) > 120.0)
	_assert_true("deep dungeon ground centered on floor", absf(float((deep_layout.get("center", Vector3.ZERO) as Vector3).x) - 60.0) <= 0.001)
	ground.queue_free()


func _test_wall_renderer() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	var ground_factory = GroundWallFactoryScript.new()
	var renderer = WallRendererScript.new(root, ground_factory)
	renderer.set_level(-4)
	var walls := renderer.render_wall_layout([{
		"id": "test_wall",
		"position": {"x": 2.0, "y": 3.0},
		"size": {"x": 4.0, "y": 1.5},
		"source": "generated",
	}])
	_assert_eq("wall layout count", walls.size(), 1)
	_assert_eq("wall child count", root.get_child_count(), 2)
	var wall_body := root.get_child(0) as StaticBody3D
	_assert_eq("wall child name", wall_body.name, "Wall_test_wall")
	var wall := wall_body.get_child(1) as MeshInstance3D
	_assert_true("dungeon wall height", absf((wall.mesh as BoxMesh).size.y - (ground_factory.dungeon_ceiling_height() + 0.08)) <= 0.001)
	var ceiling := root.get_node_or_null("DungeonCeiling") as MeshInstance3D
	_assert_true("dungeon ceiling node exists", ceiling != null)
	var ceiling_mat := ceiling.material_override as StandardMaterial3D
	_assert_true("dungeon ceiling normal enabled", ceiling_mat.normal_enabled)
	_assert_true("dungeon ceiling normal texture exists", ceiling_mat.normal_texture != null)
	renderer.set_ceiling_visible(false)
	_assert_true("dungeon ceiling can be hidden for isometric", not ceiling.visible)
	renderer.set_ceiling_visible(true)
	_assert_true("dungeon ceiling can be shown for perspective", ceiling.visible)
	_assert_true("wall material has palette texture", (wall.material_override as StandardMaterial3D).albedo_texture != null)
	_assert_true("wall material normal enabled", (wall.material_override as StandardMaterial3D).normal_enabled)
	_assert_true("wall material has normal texture", (wall.material_override as StandardMaterial3D).normal_texture != null)
	_assert_eq("wall normal texture cache count", ground_factory.wall_normal_textures.size(), 1)
	var water_walls := renderer.render_wall_layout([{
		"id": "test_water",
		"position": {"x": 7.0, "y": 8.0},
		"size": {"x": 5.0, "y": 2.0},
		"source": "generated",
		"kind": "water",
	}])
	_assert_eq("water layout count", water_walls.size(), 1)
	_assert_eq("water layout kind", str((water_walls[0] as Dictionary).get("kind", "")), "water")
	_assert_eq("water child count", root.get_child_count(), 2)
	var water := root.get_child(0) as MeshInstance3D
	_assert_eq("water child name", water.name, "Water_test_water")
	_assert_eq("water metadata kind", str(water.get_meta("kind", "")), "water")
	_assert_true("water mesh is plane", water.mesh is PlaneMesh)
	_assert_true("water material has texture", (water.material_override as StandardMaterial3D).albedo_texture != null)
	_assert_true("water motion overlay exists", water.find_child("WaterMotionBands", false, false) != null)
	var hole_walls := renderer.render_wall_layout([{
		"id": "test_hole",
		"position": {"x": 9.0, "y": 4.0},
		"size": {"x": 3.0, "y": 2.0},
		"source": "generated",
		"kind": "hole",
	}])
	_assert_eq("hole layout count", hole_walls.size(), 1)
	_assert_eq("hole layout kind", str((hole_walls[0] as Dictionary).get("kind", "")), "hole")
	_assert_eq("hole child count", root.get_child_count(), 2)
	var hole := root.get_child(0) as MeshInstance3D
	_assert_eq("hole child name", hole.name, "Hole_test_hole")
	_assert_eq("hole metadata kind", str(hole.get_meta("kind", "")), "hole")
	_assert_true("hole mesh is plane", hole.mesh is PlaneMesh)
	_assert_true("hole material has texture", (hole.material_override as StandardMaterial3D).albedo_texture != null)
	_assert_true("hole parallax overlay exists", hole.find_child("HoleParallaxBands", false, false) != null)
	var rock_walls := renderer.render_wall_layout([{
		"id": "test_rock",
		"position": {"x": 3.0, "y": 6.0},
		"size": {"x": 2.0, "y": 2.5},
		"source": "generated",
		"kind": "rock",
		"blocks_line_of_sight": true,
	}])
	_assert_eq("rock layout count", rock_walls.size(), 1)
	_assert_eq("rock layout kind", str((rock_walls[0] as Dictionary).get("kind", "")), "rock")
	_assert_true("rock layout LOS metadata", bool((rock_walls[0] as Dictionary).get("blocks_line_of_sight", false)))
	_assert_eq("rock child count", root.get_child_count(), 2)
	var rock := root.get_child(0) as Node3D
	_assert_eq("rock child name", rock.name, "Rock_test_rock")
	_assert_eq("rock metadata kind", str(rock.get_meta("kind", "")), "rock")
	_assert_true("rock has non-rectangular chunks", rock.get_child_count() >= 3)
	var rock_mat := (rock.get_child(0) as MeshInstance3D).material_override as StandardMaterial3D
	_assert_true("rock obstacle normal enabled", rock_mat.normal_enabled)
	_assert_true("rock obstacle normal texture exists", rock_mat.normal_texture != null)
	var column_walls := renderer.render_wall_layout([{
		"id": "test_column",
		"position": {"x": 4.0, "y": 6.0},
		"size": {"x": 1.2, "y": 5.0},
		"source": "generated",
		"kind": "column",
	}])
	_assert_eq("column layout count", column_walls.size(), 1)
	_assert_eq("column layout kind", str((column_walls[0] as Dictionary).get("kind", "")), "column")
	var column := root.get_child(0) as Node3D
	_assert_eq("column child name", column.name, "Column_test_column")
	_assert_eq("column metadata kind", str(column.get_meta("kind", "")), "column")
	_assert_true("column has pillars", column.get_child_count() >= 2)
	_assert_true("column first mesh is cylinder", (column.get_child(0) as MeshInstance3D).mesh is CylinderMesh)
	var rubble_walls := renderer.render_wall_layout([{
		"id": "test_rubble",
		"position": {"x": 6.0, "y": 6.0},
		"size": {"x": 3.0, "y": 2.0},
		"source": "generated",
		"kind": "rubble",
	}])
	_assert_eq("rubble layout count", rubble_walls.size(), 1)
	_assert_eq("rubble layout kind", str((rubble_walls[0] as Dictionary).get("kind", "")), "rubble")
	var rubble := root.get_child(0) as Node3D
	_assert_eq("rubble child name", rubble.name, "Rubble_test_rubble")
	_assert_eq("rubble metadata kind", str(rubble.get_meta("kind", "")), "rubble")
	_assert_true("rubble has chunks", rubble.get_child_count() >= 5)
	renderer.set_level(0)
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
	var variety_walls := renderer.render_world_walls("obstacle_variety_lab")
	var variety_rock_count := 0
	var variety_column_count := 0
	var variety_rubble_count := 0
	for rendered_wall in variety_walls:
		var rendered: Dictionary = rendered_wall
		match str(rendered.get("kind", "wall")):
			"rock":
				variety_rock_count += 1
			"column":
				variety_column_count += 1
			"rubble":
				variety_rubble_count += 1
	_assert_eq("variety lab wall layout count", variety_walls.size(), 7)
	_assert_eq("variety lab rock layout count", variety_rock_count, 1)
	_assert_eq("variety lab column layout count", variety_column_count, 1)
	_assert_eq("variety lab rubble layout count", variety_rubble_count, 1)
	_assert_eq("variety lab child count", root.get_child_count(), 7)
	_assert_true("variety lab rock node exists", root.find_child("Rock_obstacle_variety_lab_wall_004", false, false) != null)
	_assert_true("variety lab column node exists", root.find_child("Column_obstacle_variety_lab_wall_005", false, false) != null)
	_assert_true("variety lab rubble node exists", root.find_child("Rubble_obstacle_variety_lab_wall_006", false, false) != null)
	root.queue_free()


func _test_loot_node_factory() -> void:
	ItemRulesLoader.ensure_loaded()
	var factory = LootNodeFactoryScript.new({}, ItemRulesLoader.item_presentations)
	_assert_eq("gold label text", factory.loot_label_text({"item_def_id": "gold", "amount": 7}), "7 gold")
	_assert_eq("known loot name", factory.generic_loot_name("rusty_sword"), "Sword")
	_assert_eq("staff loot name", factory.generic_loot_name("starter_sorcerer_staff"), "Staff")
	_assert_eq("greatsword loot name", factory.generic_loot_name("great_sword"), "Greatsword")
	_assert_eq("rolled loot label", factory.loot_label_text({
		"item_def_id": "starter_sorcerer_staff",
		"display_name": "Magic Starter Sorcerer Staff",
	}), "Magic Starter Sorcerer Staff")
	_assert_eq("template loot label", factory.loot_label_text({"item_def_id": "starter_sorcerer_staff"}), "Sorcerer Staff")
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
	ctx.entities["boss_a"]["node"] = root
	ctx.entities["boss_a"]["visual_scale"] = 1.0
	var rec: Dictionary = ctx.entities["boss_a"]
	rec["boss_phase"] = {"pattern_id": "stone_lance"}
	controller.sync_boss_arena_presence()
	_assert_true("boss arena ring exists", root.find_child(BossArenaPresenceScript.MARKER_NAME, false, false) != null)
	controller.sync_boss_telegraph_marker(rec, {"hit_shape": "line", "radius": 7.5, "width": 1.0, "to_color": "#79b8ff"})
	_assert_true("telegraph and arena coexist", root.find_child(ClientConstantsScript.BOSS_TELEGRAPH_MARKER_NAME, false, false) != null and root.find_child(BossArenaPresenceScript.MARKER_NAME, false, false) != null)
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
	BossArenaPresenceScript.remove_for_record(rec)
	_assert_eq("arena cleanup", bool(rec.get("has_boss_arena_presence", true)), false)
	root.queue_free()


func _test_dungeon_room_floor_tint() -> void:
	var ground_factory = GroundWallFactoryScript.new()
	var ground := ground_factory.make_ground_node(-2)
	get_root().add_child(ground)
	var walls := [
		{"id": "divider_h", "position": {"x": 50.0, "y": 25.0}, "size": {"x": 80.0, "y": 1.0}, "source": "room_divider"},
		{"id": "divider_v", "position": {"x": 50.0, "y": 25.0}, "size": {"x": 1.0, "y": 40.0}, "source": "room_divider"},
	]
	var entities := {
		"chest_1": {
			"type": "interactable",
			"interactable_def_id": "treasure_chest",
			"position": {"x": 20.0, "y": 20.0},
		},
	}
	DungeonRoomFloorTintScript.sync(ground, ground_factory, -2, walls, entities)
	var tint_root := ground.get_node_or_null("DungeonRoomFloorTint")
	_assert_true("room tint root exists", tint_root != null)
	if tint_root != null:
		_assert_true("room tint overlays created", tint_root.get_child_count() >= 2)
		var treasure_found := false
		for child in tint_root.get_children():
			if str(child.name).begins_with("RoomTint_treasure"):
				treasure_found = true
		_assert_true("treasure room tint present", treasure_found)
	DungeonAmbientMotesScript.sync(ground, -2, ground_factory.floor_size_for_level(-2))
	_assert_true("ambient motes root exists", ground.get_node_or_null("DungeonAmbientMotes") != null)
	ground.queue_free()


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
