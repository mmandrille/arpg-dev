# GDScript item-visual-resolution test (spec equip-and-see-it §4.6 / acceptance #14).
#
# Server-independent, like test_golden.gd: consumes the SAME shared fixtures the
# Python validator consumes (shared/golden/item_visual_resolution.json +
# shared/assets/item_visuals.v0.json), proving the cross-language visual contract
# holds, then resolves the asset_id through the manifest to an importable res://
# resource. Run with:
#   godot --headless --path client --script res://tests/test_item_visuals.gd
# Exits 0 on success, 1 on failure. Requires no server.
extends SceneTree

const ResolverScript := preload("res://scripts/equipment_visuals.gd")
const MainScript := preload("res://scripts/main.gd")
const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const GroundWallFactoryScript := preload("res://scripts/ground_wall_factory.gd")
const WallRendererScript := preload("res://scripts/wall_renderer.gd")
const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const ScaleProbeScript := preload("res://tests/item_visual_scale_probe.gd")
const EquipmentProbeScript := preload("res://tests/item_visual_equipment_probe.gd")
const InteractableProbeScript := preload("res://tests/item_visual_interactable_probe.gd")
const CharacterScene := preload("res://scenes/character.tscn")

func _initialize() -> void:
	await _run_all()


func _run_all() -> void:
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")
	var assets := ProjectSettings.globalize_path("res://").path_join("../assets")

	# 1. Golden expectations agree with the shared item_visuals metadata.
	var golden := _read(shared.path_join("golden/item_visual_resolution.json"))
	var visuals: Dictionary = _read(shared.path_join("assets/item_visuals.v0.json"))["item_visuals"]
	var item_rules: Dictionary = _read(shared.path_join("rules/items.v0.json"))["items"]
	var item_templates: Dictionary = _read(shared.path_join("rules/item_templates.v0.json"))["templates"]
	var presentation_catalog := _read(shared.path_join("assets/item_presentations.v0.json"))
	var presentations: Dictionary = _resolved_item_presentations(presentation_catalog)

	var def_id := str(golden["item_def_id"])
	if not visuals.has(def_id):
		_fail("item_visuals is missing golden item_def_id %s" % def_id)
		return
	var vis: Dictionary = visuals[def_id]
	if str(vis["asset_id"]) != str(golden["expected_asset_id"]):
		_fail("asset_id %s != golden %s" % [vis["asset_id"], golden["expected_asset_id"]])
		return
	if str(vis["mount_socket"]) != str(golden["expected_mount_socket"]):
		_fail("mount_socket %s != golden %s" % [vis["mount_socket"], golden["expected_mount_socket"]])
		return
	if str(vis["slot"]) != str(golden["expected_slot"]):
		_fail("slot %s != golden %s" % [vis["slot"], golden["expected_slot"]])
		return

	# 2. The manifest resolves asset_id -> runtime_path -> an importable res:// resource.
	var manifest := _read(assets.path_join("manifests/assets.v0.json"))
	var entry = manifest["assets"].get(str(vis["asset_id"]), null)
	if entry == null:
		_fail("asset_id %s not present in asset manifest" % vis["asset_id"])
		return
	var res_path := _res_path(str(entry["runtime_path"]))
	if not ResourceLoader.exists(res_path):
		_fail("runtime asset not importable at %s (run godot --import)" % res_path)
		return
	if load(res_path) == null:
		_fail("failed to load runtime asset %s" % res_path)
		return

	for item_def_id in item_rules.keys():
		if not presentations.has(str(item_def_id)):
			_fail("item_presentations is missing item_def_id %s" % item_def_id)
			return
		var p: Dictionary = presentations[str(item_def_id)]
		if not p.has("family") or not p.has("icon") or not p.has("ground"):
			_fail("item_presentations %s must define icon and ground metadata" % item_def_id)
			return
	for item_def_id in item_templates.keys():
		var template: Dictionary = item_templates[item_def_id]
		if bool(template.get("equippable", false)) and not visuals.has(str(item_def_id)):
			_fail("item_visuals is missing equippable template %s" % item_def_id)
			return

	if not EquipmentProbeScript.new().verify_equipped_fallback_resolver(self, Callable(self, "_fail")):
		return
	var scale_ctx := ScaleProbeScript.new().prepare(self, MainScript, CharacterScene, ResolverScript, Callable(self, "_fail"))
	if scale_ctx.is_empty():
		return
	await process_frame
	await process_frame
	if not ScaleProbeScript.new().verify_transforms(scale_ctx, Callable(self, "_fail")):
		return
	if not EquipmentProbeScript.new().verify_off_hand_weapon_resolver(self, Callable(self, "_fail")):
		return
	if not _verify_loot_label_presentation(item_rules, item_templates, presentations):
		return
	var interactable_probe := InteractableProbeScript.new()
	if not interactable_probe.verify_chest_models(self, Callable(self, "_fail")):
		return
	if not interactable_probe.verify_vendor_models(Callable(self, "_fail")):
		return
	if not interactable_probe.verify_market_board_model(Callable(self, "_fail")):
		return
	if not interactable_probe.verify_stair_models(Callable(self, "_fail")):
		return
	if not _verify_ground_texture_selection():
		return
	if not _verify_town_preview_props():
		return
	if not _verify_wall_texture_material():
		return

	print("[gdtest] PASS: item visual resolution and presentation metadata (manifest -> %s)" % res_path)
	quit(0)

func _res_path(runtime_path: String) -> String:
	# Manifest runtime_path (client/assets/...) -> Godot res:// (ADR-0006 D6).
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p

func _read(path: String) -> Dictionary:
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		_fail("cannot open %s" % path)
		return {}
	var parsed = JSON.parse_string(f.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		_fail("invalid JSON in %s" % path)
		return {}
	return parsed

func _resolved_item_presentations(catalog: Dictionary) -> Dictionary:
	var families: Dictionary = catalog.get("families", {})
	var items: Dictionary = catalog.get("items", {})
	var resolved := {}
	for item_def_id in items.keys():
		var entry: Dictionary = items.get(item_def_id, {})
		var family_id := str(entry.get("family", ""))
		var family: Dictionary = families.get(family_id, {})
		var presentation := family.duplicate(true)
		for key in ["icon", "ground", "3d_model"]:
			if entry.has(key):
				var value = entry.get(key)
				presentation[key] = (value as Dictionary).duplicate(true) if typeof(value) == TYPE_DICTIONARY else value
		presentation["family"] = family_id
		resolved[str(item_def_id)] = presentation
	return resolved

func _verify_loot_label_presentation(item_rules: Dictionary, item_templates: Dictionary, presentations: Dictionary) -> bool:
	var main = MainScript.new()
	main.item_rules = item_rules.duplicate(true)
	main.item_templates = item_templates.duplicate(true)
	main.item_presentations = presentations.duplicate(true)

	if main._loot_label_color({"item_def_id": "gold"}).to_html(false) != "ffd75e":
		_fail("gold loot label is not yellow")
		main.free()
		return false
	if main._loot_label_color({"item_def_id": "quest_leaf"}).to_html(false) != "6ee68b":
		_fail("quest loot label is not green")
		main.free()
		return false
	if main._loot_label_color({"item_def_id": "red_potion"}).to_html(false) != "ff8f70":
		_fail("consumable loot label is not category-colored")
		main.free()
		return false
	if main._loot_label_color({"item_def_id": "cave_blade", "rarity": "magic"}).to_html(false) != "93c5fd":
		_fail("magic equipment loot label is not magic-colored")
		main.free()
		return false
	if main._loot_label_color({"item_def_id": "cave_blade", "rarity": "rare"}).to_html(false) != "f4d481":
		_fail("rare equipment loot label is not rare-colored")
		main.free()
		return false
	if main._loot_label_color({"item_def_id": "future_item"}).to_html(false) != "e8dcc8":
		_fail("unknown loot label did not fall back to common rarity color")
		main.free()
		return false

	var loot_factory = LootNodeFactoryScript.new({}, ItemRulesLoader.item_presentations)
	var gold_node: Node3D = loot_factory.make_loot_node({"item_def_id": "gold", "rarity": "common", "amount": 7})
	var gold_label := gold_node.find_child("LootLabel", true, false) as Label3D
	if gold_label == null or gold_label.modulate.to_html(false) != "ffd75e":
		_fail("gold loot node label color mismatch")
		gold_node.free()
		main.free()
		return false
	if gold_label.text != "7 gold":
		_fail("gold loot node label text mismatch: %s" % gold_label.text)
		gold_node.free()
		main.free()
		return false
	gold_node.free()

	var root_path := ProjectSettings.globalize_path("res://")
	loot_factory.configure(_read(root_path.path_join("../assets/manifests/assets.v0.json"))["assets"], ItemRulesLoader.item_presentations)
	var sword_node: Node3D = loot_factory.make_loot_node({"item_def_id": "rusty_sword", "rarity": "magic"})
	var sword_model := sword_node.find_child("GroundModel_weapon_rusty_sword_v0", true, false) as Node3D
	if sword_model == null:
		_fail("equipment floor loot did not use manifest-backed ground model")
		sword_node.free()
		main.free()
		return false
	if absf(sword_model.scale.x - 1.0) > 0.001 or absf(sword_model.scale.y - 1.0) > 0.001 or absf(sword_model.scale.z - 1.0) > 0.001:
		_fail("equipment floor loot model was not scaled to 100%%: %s" % str(sword_model.scale))
		sword_node.free()
		main.free()
		return false
	if sword_model.position.y > 0.15 or absf(sword_model.rotation_degrees.x - 90.0) > 0.001:
		_fail("equipment floor loot model is not lying on the floor: pos=%s rot=%s" % [str(sword_model.position), str(sword_model.rotation_degrees)])
		sword_node.free()
		main.free()
		return false
	if sword_node.find_child("RarityBackground", true, false) != null:
		_fail("equipment floor loot model still renders a rarity floor tile")
		sword_node.free()
		main.free()
		return false
	sword_node.free()

	var armor_node: Node3D = loot_factory.make_loot_node({"item_def_id": "cave_helm", "rarity": "rare"})
	if armor_node.find_child("GroundModel_fallback_equipment_head_v0", true, false) != null:
		_fail("armor floor loot loaded fallback sword-backed model")
		armor_node.free()
		main.free()
		return false
	if armor_node.find_child("HelmCap", true, false) == null or armor_node.find_child("HelmBrow", true, false) == null:
		_fail("armor floor loot did not use helmet primitive presentation")
		armor_node.free()
		main.free()
		return false
	if armor_node.find_child("RarityBackground", true, false) == null:
		_fail("armor primitive floor loot is missing rarity background")
		armor_node.free()
		main.free()
		return false
	armor_node.free()

	var shield_node: Node3D = loot_factory.make_loot_node({"item_def_id": "cave_shield", "rarity": "magic"})
	if shield_node.find_child("GroundModel_fallback_equipment_off_hand_v0", true, false) != null:
		_fail("shield floor loot loaded fallback sword-backed model")
		shield_node.free()
		main.free()
		return false
	if shield_node.find_child("ShieldFace", true, false) == null or shield_node.find_child("ShieldBoss", true, false) == null:
		_fail("shield floor loot did not use shield primitive presentation")
		shield_node.free()
		main.free()
		return false
	shield_node.free()

	var loot_a := _make_labelled_loot_node()
	var loot_b := _make_labelled_loot_node()
	main.entities = {
		"loot_a": {"node": loot_a, "type": "loot", "item_def_id": "gold"},
		"loot_b": {"node": loot_b, "type": "loot", "item_def_id": "cave_blade"},
	}
	main.loot_ids = ["loot_a", "loot_b"]
	main.hovered_loot_id = "loot_a"
	main.loot_label_reveal_held = false
	main._refresh_loot_label_visibility()
	var label_a := loot_a.find_child("LootLabel", true, false) as Label3D
	var label_b := loot_b.find_child("LootLabel", true, false) as Label3D
	if label_a == null or label_b == null or not label_a.visible or label_b.visible:
		_fail("hover visibility did not isolate one loot label")
		loot_a.free()
		loot_b.free()
		main.free()
		return false
	if label_a.modulate.to_html(false) != "ffffff":
		_fail("hovered currency loot label did not use white highlight")
		loot_a.free()
		loot_b.free()
		main.free()
		return false
	main.loot_label_reveal_held = true
	main._refresh_loot_label_visibility()
	if not label_a.visible or not label_b.visible:
		_fail("ALT reveal visibility did not show all loot labels")
		loot_a.free()
		loot_b.free()
		main.free()
		return false
	var label_b_full := main._loot_label_color({"item_def_id": "cave_blade"})
	if label_a.modulate.to_html(false) != "ffffff":
		_fail("ALT-hovered currency loot label did not stay white")
		loot_a.free()
		loot_b.free()
		main.free()
		return false
	if label_b.modulate.r >= label_b_full.r or label_b.modulate.g >= label_b_full.g or label_b.modulate.b >= label_b_full.b:
		_fail("ALT-only loot label did not dim below full item color")
		loot_a.free()
		loot_b.free()
		main.free()
		return false
	main.loot_label_reveal_held = false
	main.hovered_loot_id = ""
	main._refresh_loot_label_visibility()
	if label_a.visible or label_b.visible:
		_fail("loot labels remained visible after reveal and hover cleared")
		loot_a.free()
		loot_b.free()
		main.free()
		return false

	loot_a.free()
	loot_b.free()
	main.free()
	return true

func _verify_ground_texture_selection() -> bool:
	var ground_factory = GroundWallFactoryScript.new()
	if ground_factory.ground_texture_id_for_level(0) != ClientConstantsScript.GROUND_TEXTURE_TOWN:
		_fail("town level did not select grass ground texture")
		return false
	for level in [-1, -2, -10, 1]:
		if ground_factory.ground_texture_id_for_level(level) != ClientConstantsScript.GROUND_TEXTURE_DUNGEON:
			_fail("dungeon level %d did not select rock ground texture" % level)
			return false
	var town_a: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_TOWN, 0, 0)
	var town_b: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_TOWN, 17, 29)
	var town_c: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_TOWN, 32, 32)
	var rock_a: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 0, 0)
	var rock_b: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 17, 29)
	var shallow_palette: Dictionary = ground_factory.biome_palette_for_level(-1)
	var deep_palette: Dictionary = ground_factory.biome_palette_for_level(-4)
	if str(shallow_palette.get("id", "")) == str(deep_palette.get("id", "")):
		_fail("dungeon biome palettes did not vary by depth")
		return false
	var shallow_rock: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 0, 0, shallow_palette)
	var deep_rock: Color = ground_factory.ground_texel(ClientConstantsScript.GROUND_TEXTURE_DUNGEON, 0, 0, deep_palette)
	if shallow_rock == deep_rock:
		_fail("dungeon biome palette did not change ground texels")
		return false
	if town_a == rock_a:
		_fail("town and dungeon ground textures share the same base texel")
		return false
	if town_a == town_b or rock_a == rock_b:
		_fail("ground textures are flat colors")
		return false
	if town_a == town_c or town_b == town_c:
		_fail("town ground texture does not expose enough color variation")
		return false
	var mat: StandardMaterial3D = ground_factory.ground_material_for_level(-1)
	if mat.albedo_texture == null:
		_fail("dungeon ground material is missing its generated texture")
		return false
	return true


func _verify_town_preview_props() -> bool:
	var town: Node3D = TownNodeFactoryScript.make_town_preview_scene()
	if town == null:
		_fail("town preview scene was not created")
		return false
	var required := [
		"TownPreviewGround", "TownService_town_vendor", "TownService_town_mystery_seller",
		"TownService_town_stash", "TownService_town_bishop", "TownService_town_market_board",
		"TownCabinWest", "TownCabinEast", "TownCampfire",
	]
	for node_name in required:
		if town.find_child(str(node_name), true, false) == null:
			_fail("town preview missing %s" % node_name)
			town.free()
			return false
	var fire := town.find_child("TownCampfire", true, false) as Node3D
	if fire.find_child("CampfireLight", true, false) == null or fire.find_child("FireFlameInner", true, false) == null:
		_fail("town campfire is missing light or flame parts")
		town.free()
		return false
	var cabin := town.find_child("TownCabinWest", true, false)
	if cabin.find_child("CabinDoor", true, false) == null or cabin.find_child("CabinRoofRidge", true, false) == null:
		_fail("town cabin is missing door or roof parts")
		town.free()
		return false
	var fire_pos := Vector2(fire.position.x, fire.position.z)
	for node_name in required:
		if str(node_name) in ["TownPreviewGround", "TownCampfire"]:
			continue
		var node := town.find_child(str(node_name), true, false) as Node3D
		var distance := fire_pos.distance_to(Vector2(node.position.x, node.position.z))
		if distance < 5.0:
			_fail("town preview %s is too close to campfire: %.2f" % [node_name, distance])
			town.free()
			return false
	town.free()
	return true


func _verify_wall_texture_material() -> bool:
	var wall_renderer = WallRendererScript.new(null, GroundWallFactoryScript.new())
	var generated_body := wall_renderer.make_wall_node({
		"id": "test_generated",
		"position": {"x": 4.0, "y": 5.0},
		"size": {"x": 8.0, "y": 2.0},
		"source": "generated",
	})
	var perimeter_body := wall_renderer.make_wall_node({
		"id": "test_perimeter",
		"position": {"x": 4.0, "y": 5.0},
		"size": {"x": 8.0, "y": 2.0},
		"source": "perimeter",
	})
	var generated := _wall_mesh_from_body(generated_body)
	var perimeter := _wall_mesh_from_body(perimeter_body)
	if generated == null or perimeter == null:
		_fail("wall renderer did not create mesh children")
		generated_body.free()
		perimeter_body.free()
		return false
	var generated_mat := generated.material_override as StandardMaterial3D
	var perimeter_mat := perimeter.material_override as StandardMaterial3D
	if generated_mat == null or generated_mat.albedo_texture == null:
		_fail("generated cave wall material is missing its stone texture")
		generated_body.free()
		perimeter_body.free()
		return false
	if perimeter_mat == null or perimeter_mat.albedo_texture == null:
		_fail("perimeter cave wall material is missing its stone texture")
		generated_body.free()
		perimeter_body.free()
		return false
	var wall_factory = GroundWallFactoryScript.new()
	var wall_a: Color = wall_factory.wall_texel(ClientConstantsScript.WALL_TEXTURE_CAVE, 0, 0)
	var wall_b: Color = wall_factory.wall_texel(ClientConstantsScript.WALL_TEXTURE_CAVE, 17, 19)
	if wall_a == wall_b:
		_fail("cave wall texture is flat")
		generated_body.free()
		perimeter_body.free()
		return false
	if generated_mat.albedo_color == perimeter_mat.albedo_color:
		_fail("generated and perimeter wall materials use the same tint")
		generated_body.free()
		perimeter_body.free()
		return false
	generated_body.free()
	perimeter_body.free()
	return true


func _wall_mesh_from_body(body: Node3D) -> MeshInstance3D:
	for child in body.get_children():
		if child is MeshInstance3D:
			return child as MeshInstance3D
	return null


func _make_labelled_loot_node() -> Node3D:
	var node := Node3D.new()
	var label := Label3D.new()
	label.name = "LootLabel"
	label.visible = false
	node.add_child(label)
	return node


func _fail(msg: String) -> void:
	printerr("[gdtest] FAIL: ", msg)
	quit(1)
