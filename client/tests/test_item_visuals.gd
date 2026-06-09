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


func _initialize() -> void:
	var shared := ProjectSettings.globalize_path("res://").path_join("../shared")
	var assets := ProjectSettings.globalize_path("res://").path_join("../assets")

	# 1. Golden expectations agree with the shared item_visuals metadata.
	var golden := _read(shared.path_join("golden/item_visual_resolution.json"))
	var visuals: Dictionary = _read(shared.path_join("assets/item_visuals.v0.json"))["item_visuals"]
	var item_rules: Dictionary = _read(shared.path_join("rules/items.v0.json"))["items"]
	var item_templates: Dictionary = _read(shared.path_join("rules/item_templates.v0.json"))["templates"]
	var presentations: Dictionary = _read(shared.path_join("assets/item_presentations.v0.json"))["items"]

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
		if not p.has("icon") or not p.has("ground"):
			_fail("item_presentations %s must define icon and ground metadata" % item_def_id)
			return
	for item_def_id in item_templates.keys():
		var template: Dictionary = item_templates[item_def_id]
		if bool(template.get("equippable", false)) and not visuals.has(str(item_def_id)):
			_fail("item_visuals is missing equippable template %s" % item_def_id)
			return

	if not _verify_equipped_fallback_resolver():
		return
	if not _verify_loot_label_presentation(item_rules, item_templates, presentations):
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

	var gold_node := main._make_loot_node({"item_def_id": "gold", "rarity": "common", "amount": 7})
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
