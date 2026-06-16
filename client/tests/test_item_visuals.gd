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

func _initialize() -> void:
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

	if not _verify_equipped_fallback_resolver():
		return
	if not _verify_off_hand_weapon_resolver():
		return
	if not _verify_loot_label_presentation(item_rules, item_templates, presentations):
		return
	if not _verify_interactable_chest_models():
		return
	if not _verify_interactable_vendor_models():
		return
	if not _verify_market_board_model():
		return
	if not _verify_interactable_stair_models():
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

func _verify_equipped_fallback_resolver() -> bool:
	var mount := _make_mount_root()
	var resolver = ResolverScript.new(mount)
	var inventory := [
		{"item_instance_id": "2001", "item_def_id": "cave_helm", "slot": "head", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2002", "item_def_id": "cave_amulet", "slot": "amulet", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2003", "item_def_id": "cave_mail", "slot": "chest", "equipped": true, "rarity": "common"},
		{"item_instance_id": "2004", "item_def_id": "cave_gloves", "slot": "gloves", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2005", "item_def_id": "cave_belt", "slot": "belt", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2006", "item_def_id": "cave_boots", "slot": "boots", "equipped": true, "rarity": "common"},
		{"item_instance_id": "2007", "item_def_id": "cave_ring", "slot": "ring_left", "equipped": true, "rarity": "magic"},
		{"item_instance_id": "2008", "item_def_id": "cave_ring", "slot": "ring_right", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2009", "item_def_id": "cave_bow", "slot": "main_hand", "equipped": true, "rarity": "rare"},
		{"item_instance_id": "2010", "item_def_id": "cave_shield", "slot": "off_hand", "equipped": true, "rarity": "magic"},
	]
	resolver.apply_snapshot({
		"inventory": inventory,
		"equipped": {
			"head": "2001",
			"amulet": "2002",
			"chest": "2003",
			"gloves": "2004",
			"belt": "2005",
			"boots": "2006",
			"ring_left": "2007",
			"ring_right": "2008",
			"main_hand": "2009",
			"off_hand": "2010",
		},
	})
	var state: Dictionary = resolver.get_debug_state()
	if not state["warnings"].is_empty():
		_fail("resolver emitted warnings for complete equipment fallback map: %s" % state["warnings"])
		return false
	var equipped_visuals: Dictionary = state["equipped_visuals"]
	for slot in ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]:
		if not equipped_visuals.has(slot):
			_fail("resolver did not mount slot %s: %s" % [slot, equipped_visuals])
			return false
		var mounted: Dictionary = equipped_visuals[slot]
		if not bool(mounted.get("visible", false)):
			_fail("resolver mounted invisible slot %s: %s" % [slot, mounted])
			return false
	for slot in ["head", "chest", "boots", "off_hand"]:
		if not bool((equipped_visuals[slot] as Dictionary).get("procedural_fallback", false)):
			_fail("resolver did not use procedural fallback for slot %s: %s" % [slot, equipped_visuals[slot]])
			return false
	if str(equipped_visuals["ring_right"].get("mount_socket", "")) != "ring_right_socket":
		_fail("ring_right mounted to wrong socket: %s" % equipped_visuals["ring_right"])
		return false
	if str(equipped_visuals["head"].get("tint", "")) != "ffd75e":
		_fail("rare head tint mismatch: %s" % equipped_visuals["head"])
		return false
	var head_node := mount.find_child("fallback_equipment_head_v0", true, false) as Node3D
	if head_node == null or head_node.position.y < 0.12:
		_fail("helmet fallback not raised above head socket: %s" % str(head_node.position if head_node != null else null))
		return false
	var chest_node := mount.find_child("fallback_equipment_chest_v0", true, false) as Node3D
	if chest_node == null or chest_node.position.z < 0.10:
		_fail("chest fallback not pushed out from torso: %s" % str(chest_node.position if chest_node != null else null))
		return false
	var boots_node := mount.find_child("fallback_equipment_boots_v0", true, false) as Node3D
	if boots_node == null or absf(boots_node.rotation_degrees.z) > 0.001:
		_fail("boots fallback rotated away from left/right foot layout: %s" % str(boots_node.rotation_degrees if boots_node != null else null))
		return false
	var left_boot := boots_node.find_child("left_boot", true, false) as Node3D
	var right_boot := boots_node.find_child("right_boot", true, false) as Node3D
	if left_boot == null or right_boot == null or left_boot.position.x > -0.5 or right_boot.position.x < 0.5:
		_fail("boots fallback not split across feet: left=%s right=%s" % [
			str(left_boot.position if left_boot != null else null),
			str(right_boot.position if right_boot != null else null),
		])
		return false

	resolver.apply_snapshot({
		"inventory": [{"item_instance_id": "3001", "item_def_id": "future_helmet", "slot": "head", "equipped": true, "rarity": "magic"}],
		"equipped": {"head": "3001"},
	})
	state = resolver.get_debug_state()
	equipped_visuals = state["equipped_visuals"]
	if not equipped_visuals.has("head") or str(equipped_visuals["head"].get("asset_id", "")) != "fallback_equipment_head_v0":
		_fail("unmapped future item did not use head fallback: %s" % equipped_visuals)
		return false
	if str(equipped_visuals["head"].get("tint", "")) != "5aa7ff":
		_fail("unmapped future item magic tint mismatch: %s" % equipped_visuals["head"])
		return false
	mount.queue_free()
	return true

func _verify_off_hand_weapon_resolver() -> bool:
	var mount := _make_mount_root()
	var resolver = ResolverScript.new(mount)
	resolver.apply_snapshot({
		"inventory": [
			{"item_instance_id": "4001", "item_def_id": "starter_rogue_sword", "slot": "main_hand", "equipped": true, "rarity": "common"},
		],
		"equipped": {"off_hand": "4001"},
	})
	var state: Dictionary = resolver.get_debug_state()
	if not state["warnings"].is_empty():
		_fail("resolver emitted warnings for rogue off-hand sword: %s" % state["warnings"])
		return false
	var equipped_visuals: Dictionary = state["equipped_visuals"]
	if not equipped_visuals.has("off_hand"):
		_fail("rogue starter sword did not mount off hand: %s" % equipped_visuals)
		return false
	var off_hand: Dictionary = equipped_visuals["off_hand"]
	if str(off_hand.get("mount_socket", "")) != "off_hand_socket":
		_fail("rogue starter sword off hand used wrong socket: %s" % off_hand)
		return false
	if bool(off_hand.get("procedural_fallback", false)):
		_fail("rogue starter sword off hand used shield fallback: %s" % off_hand)
		return false
	var node := mount.find_child("weapon_rusty_sword_v0", true, false) as Node3D
	if node == null:
		_fail("rogue starter sword off hand node missing")
		return false
	if absf(node.rotation_degrees.z - 180.0) > 0.01 or node.position.z < 0.07:
		_fail("rogue starter sword off hand transform not mirrored: pos=%s rot=%s" % [str(node.position), str(node.rotation_degrees)])
		return false
	mount.queue_free()
	return true

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
	if label_a.modulate.to_html(false) != "ffd75e":
		_fail("hovered loot label did not use full item color")
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
	if label_a.modulate.to_html(false) != "ffd75e":
		_fail("ALT-hovered loot label did not stay highlighted")
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

func _verify_interactable_chest_models() -> bool:
	var main = MainScript.new()
	get_root().add_child(main)
	var stash := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_stash"})
	var chest := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest"})
	var objective := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest", "elite_objective": true})
	var quest := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest", "quest_reward": true})
	if stash == null or stash.name != "TownStashChest" or stash.find_child("ChestStashCrest", true, false) == null:
		_fail("town stash did not use fortified chest model")
		main.free()
		return false
	if chest == null or chest.name != "TreasureChest" or chest.find_child("ChestLockPlate", true, false) == null:
		_fail("treasure chest did not use chest model")
		stash.free()
		main.free()
		return false
	var glow := chest.find_child("ChestInnerGlow", true, false) as MeshInstance3D
	var lid := chest.find_child("ChestLidPivot", true, false) as Node3D
	if glow == null or lid == null or glow.visible:
		_fail("treasure chest missing closed lid/glow state")
		stash.free()
		chest.free()
		main.free()
		return false
	if objective == null or objective.find_child("EliteObjectiveMarker", true, false) == null:
		_fail("objective treasure chest did not expose marker")
		stash.free()
		chest.free()
		quest.free()
		main.free()
		return false
	if quest == null or quest.find_child("QuestRewardMarker", true, false) == null:
		_fail("quest reward treasure chest did not expose marker")
		stash.free()
		chest.free()
		objective.free()
		main.free()
		return false
	main.add_child(chest)
	main._set_interactable_state("chest_1", {"node": chest, "interactable_def_id": "treasure_chest", "state": "closed"}, "open")
	if not glow.visible:
		_fail("opened treasure chest did not reveal inner glow")
		stash.free()
		main.free()
		return false
	var objective_marker := objective.find_child("EliteObjectiveMarker", true, false) as MeshInstance3D
	main.add_child(objective)
	main._set_interactable_state("objective_chest_1", {"node": objective, "interactable_def_id": "treasure_chest", "elite_objective": true, "state": "closed"}, "open")
	if objective_marker == null or not objective_marker.visible:
		_fail("opened objective treasure chest did not keep marker visible")
		stash.free()
		main.free()
		return false
	var quest_marker := quest.find_child("QuestRewardMarker", true, false) as MeshInstance3D
	main.add_child(quest)
	main._set_interactable_state("quest_chest_1", {"node": quest, "interactable_def_id": "treasure_chest", "quest_reward": true, "state": "closed"}, "open")
	if quest_marker == null or not quest_marker.visible:
		_fail("opened quest reward treasure chest did not keep marker visible")
		stash.free()
		main.free()
		return false
	stash.free()
	main.free()
	return true

func _verify_interactable_vendor_models() -> bool:
	var main = MainScript.new()
	var vendor := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_vendor"})
	var mystery := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_mystery_seller"})
	if vendor == null or vendor.name != "ShopVendor" or vendor.find_child("VendorSign", true, false) == null:
		_fail("town vendor did not use merchant model")
		main.free()
		return false
	if mystery == null or mystery.name != "MysterySeller" or mystery.find_child("CrystalOrb", true, false) == null:
		_fail("mystery seller did not use dark-violet merchant model")
		vendor.free()
		main.free()
		return false
	var vendor_body := vendor.find_child("Body", true, false) as MeshInstance3D
	var mystery_body := mystery.find_child("Body", true, false) as MeshInstance3D
	if vendor_body == null or (vendor_body.material_override as StandardMaterial3D).albedo_color.to_html(false) != "e2b92e":
		_fail("town vendor body is not yellow")
		vendor.free()
		mystery.free()
		main.free()
		return false
	if mystery_body == null or (mystery_body.material_override as StandardMaterial3D).albedo_color.to_html(false) != "2b124a":
		_fail("mystery seller body is not dark violet")
		vendor.free()
		mystery.free()
		main.free()
		return false
	vendor.free()
	mystery.free()
	main.free()
	return true

func _verify_market_board_model() -> bool:
	var main = MainScript.new()
	var board := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_market_board"})
	if board == null or board.name != "MarketBoard":
		_fail("market board did not use board model")
		main.free()
		return false
	if board.find_child("IncomingBidCount", true, false) == null:
		_fail("market board missing incoming bid counter")
		board.free()
		main.free()
		return false
	if board.find_child("PublishedListingCount", true, false) == null:
		_fail("market board missing published listing counter")
		board.free()
		main.free()
		return false
	board.free()
	main.free()
	return true


func _verify_interactable_stair_models() -> bool:
	var main = MainScript.new()
	var up := main._make_entity_node({"type": "interactable", "interactable_def_id": "stairs_up"})
	var down := main._make_entity_node({"type": "interactable", "interactable_def_id": "stairs_down"})
	if up == null or up.find_child("UpHighLanding", true, false) == null or up.find_child("UpBackWall", true, false) == null:
		_fail("stairs_up did not use raised stair model")
		main.free()
		return false
	if down == null or down.find_child("DownPitOpening", true, false) == null or down.find_child("DownBackWall", true, false) == null:
		_fail("stairs_down did not use descending pit model")
		up.free()
		main.free()
		return false
	var first_down_step := down.find_child("DownStep0", true, false) as Node3D
	var last_down_step := down.find_child("DownStep4", true, false) as Node3D
	if first_down_step == null or last_down_step == null or first_down_step.position.y <= last_down_step.position.y:
		_fail("stairs_down steps do not descend into the opening")
		up.free()
		down.free()
		main.free()
		return false
	var first_up_step := up.find_child("UpStep0", true, false) as Node3D
	var last_up_step := up.find_child("UpStep4", true, false) as Node3D
	if first_up_step == null or last_up_step == null or first_up_step.position.y >= last_up_step.position.y:
		_fail("stairs_up steps do not rise to the landing")
		up.free()
		down.free()
		main.free()
		return false
	up.free()
	down.free()
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
	var generated: MeshInstance3D = wall_renderer.make_wall_node({
		"id": "test_generated",
		"position": {"x": 4.0, "y": 5.0},
		"size": {"x": 8.0, "y": 2.0},
		"source": "generated",
	})
	var perimeter: MeshInstance3D = wall_renderer.make_wall_node({
		"id": "test_perimeter",
		"position": {"x": 4.0, "y": 5.0},
		"size": {"x": 8.0, "y": 2.0},
		"source": "perimeter",
	})
	var generated_mat := generated.material_override as StandardMaterial3D
	var perimeter_mat := perimeter.material_override as StandardMaterial3D
	if generated_mat == null or generated_mat.albedo_texture == null:
		_fail("generated cave wall material is missing its stone texture")
		generated.free()
		perimeter.free()
		return false
	if perimeter_mat == null or perimeter_mat.albedo_texture == null:
		_fail("perimeter cave wall material is missing its stone texture")
		generated.free()
		perimeter.free()
		return false
	var wall_factory = GroundWallFactoryScript.new()
	var wall_a: Color = wall_factory.wall_texel(ClientConstantsScript.WALL_TEXTURE_CAVE, 0, 0)
	var wall_b: Color = wall_factory.wall_texel(ClientConstantsScript.WALL_TEXTURE_CAVE, 17, 19)
	if wall_a == wall_b:
		_fail("cave wall texture is flat")
		generated.free()
		perimeter.free()
		return false
	if generated_mat.albedo_color == perimeter_mat.albedo_color:
		_fail("generated and perimeter wall materials use the same tint")
		generated.free()
		perimeter.free()
		return false
	generated.free()
	perimeter.free()
	return true


func _make_mount_root() -> Node3D:
	var mount := Node3D.new()
	mount.name = "CharacterVisual"
	for socket_name in [
		"right_hand_socket",
		"off_hand_socket",
		"head_socket",
		"amulet_socket",
		"chest_socket",
		"gloves_socket",
		"belt_socket",
		"boots_socket",
		"ring_left_socket",
		"ring_right_socket",
	]:
		var socket := Node3D.new()
		socket.name = str(socket_name)
		mount.add_child(socket)
	get_root().add_child(mount)
	return mount


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
