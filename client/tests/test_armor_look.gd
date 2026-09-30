extends SceneTree
## ADR-0018 D5 / P3b (v483): armor recolours the kit hero's own body parts and toggles headgear.
## Node-render component test (Agent rule 9 exception): real kit heroes + EquipmentVisualResolver.

const CharacterScene := preload("res://scenes/character.tscn")
const ResolverScript := preload("res://scripts/equipment_visuals.gd")
const ArmorLookScript := preload("res://scripts/armor_look.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const ReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")

const CLASS_IDS := ["barbarian", "paladin", "rogue", "ranger", "sorcerer"]

var _failed := false
var _checks := 0


func _initialize() -> void:
	for class_id in CLASS_IDS:
		await _test_full_loadout_tints_regions(class_id)
	await _test_unequip_clears_and_hides()
	await _test_belt_only_tints_body_without_chest()
	await _test_tint_survives_hit_flash_and_regear()
	if not _failed:
		print("[gdtest] PASS: test_armor_look (%d checks)" % _checks)
	quit(1 if _failed else 0)


func _test_full_loadout_tints_regions(class_id: String) -> void:
	var hero := await _hero(class_id)
	var resolver = ResolverScript.new(hero)
	resolver.set_character_class(class_id)
	resolver.apply_snapshot(_snapshot({"chest": "mail", "gloves": "gloves", "boots": "boots", "head": "helm"}))
	var state: Dictionary = resolver.get_debug_state()
	_check((state["warnings"] as Array).is_empty(), "%s full loadout warns: %s" % [class_id, state["warnings"]])
	_check(_region_colour(hero, "_Body") == _look_colour("mail"), "%s body should carry the mail colour" % class_id)
	_check(_region_colour(hero, "_ArmLeft") == _look_colour("gloves") and _region_colour(hero, "_ArmRight") == _look_colour("gloves"), "%s arms should carry the gloves colour" % class_id)
	_check(_region_colour(hero, "_LegLeft") == _look_colour("boots"), "%s legs should carry the boots colour" % class_id)
	_check(_region_colour(hero, "_Head") == "", "%s face must never be tinted" % class_id)
	for mesh in _headgear(hero):
		_check(mesh.visible, "%s headgear %s should show with a helm" % [class_id, mesh.name])
	var look: Dictionary = state["armor_look"]
	_check(int(look.get("headgear_meshes", 0)) == _headgear(hero).size(), "%s headgear count mismatch: %s" % [class_id, look])
	for socket_name in ["head_socket", "chest_socket", "gloves_socket", "boots_socket"]:
		var socket := hero.find_child(socket_name, true, false)
		_check(socket == null or socket.get_child_count() == 0, "%s %s must not carry an armor mesh" % [class_id, socket_name])
	hero.queue_free()
	await process_frame


func _test_unequip_clears_and_hides() -> void:
	var hero := await _hero("paladin")
	var resolver = ResolverScript.new(hero)
	resolver.set_character_class("paladin")
	resolver.apply_snapshot(_snapshot({"chest": "mail", "head": "helm"}))
	_check(not _headgear(hero).is_empty(), "paladin (Knight) must expose helmet meshes")
	resolver.apply_equipped_update("head", null)
	resolver.apply_equipped_update("chest", null)
	for mesh in _headgear(hero):
		_check(not mesh.visible, "helmet %s must hide when the head slot empties" % mesh.name)
	_check(_region_colour(hero, "_Body") == "", "body tint must clear when the chest empties")
	hero.queue_free()
	await process_frame


func _test_belt_only_tints_body_without_chest() -> void:
	var hero := await _hero("barbarian")
	var resolver = ResolverScript.new(hero)
	resolver.set_character_class("barbarian")
	resolver.apply_snapshot(_snapshot({"chest": "leather_vest", "belt": "war_girdle"}))
	_check(_region_colour(hero, "_Body") == _look_colour("leather_vest"), "chest must win the body over the belt")
	resolver.apply_equipped_update("chest", null)
	_check(_region_colour(hero, "_Body") == _look_colour("war_girdle"), "belt must tint the body when no chest is worn")
	hero.queue_free()
	await process_frame


func _test_tint_survives_hit_flash_and_regear() -> void:
	var hero := await _hero("paladin")
	var resolver = ResolverScript.new(hero)
	resolver.set_character_class("paladin")
	resolver.apply_snapshot(_snapshot({"chest": "full_plate", "boots": "boots"}))
	var reaction = ReactionControllerScript.new(hero, Color.WHITE)
	reaction.play_hit()
	for i in 12:
		await process_frame
	_check(_region_colour(hero, "_Body") == _look_colour("full_plate"), "hit flash must not erase the chest tint")
	resolver.apply_equipped_update("boots", "9002")
	resolver.ingest_inventory_item({"item_instance_id": "9002", "item_def_id": "plate_boots", "slot": "boots", "equipped": true})
	_check(_region_colour(hero, "_LegLeft") == _look_colour("plate_boots"), "re-gearing after a hit must retint the legs")
	reaction.enter_death()
	for i in 12:
		await process_frame
	_check(_region_colour(hero, "_LegLeft") == _look_colour("plate_boots"), "death darken must keep the boots tint")
	reaction.dispose()
	hero.queue_free()
	await process_frame


func _snapshot(by_slot: Dictionary) -> Dictionary:
	var inventory: Array = []
	var equipped := {}
	var next_id := 9100
	for slot in by_slot.keys():
		var iid := str(next_id)
		next_id += 1
		inventory.append({"item_instance_id": iid, "item_def_id": by_slot[slot], "slot": slot, "equipped": true, "rarity": "common"})
		equipped[slot] = iid
	return {"inventory": inventory, "equipped": equipped}


func _hero(class_id: String) -> Node3D:
	var hero := CharacterScene.instantiate() as Node3D
	get_root().add_child(hero)
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
	var old := hero.find_child("ModelRoot", false, false)
	if old != null:
		hero.remove_child(old)
		old.free()
	var model := packed.instantiate() as Node3D
	model.name = "ModelRoot"
	hero.add_child(model)
	hero.move_child(model, 0)
	if "class_id" in hero:
		hero.set("class_id", class_id)
	if hero.has_method("refresh_gear_sockets"):
		hero.call("refresh_gear_sockets")
	await process_frame
	return hero


## Detail-layer colour on the first mesh whose name ends with `suffix` ("" when untinted).
func _region_colour(hero: Node, suffix: String) -> String:
	for node in hero.find_children("*", "MeshInstance3D", true, false):
		if not str(node.name).ends_with(suffix):
			continue
		var mat := (node as MeshInstance3D).material_override as StandardMaterial3D
		if mat == null or not mat.detail_enabled or mat.detail_albedo == null:
			return ""
		return mat.detail_albedo.get_image().get_pixel(0, 0).to_html(false)
	return "<no %s mesh>" % suffix


func _look_colour(item_def_id: String) -> String:
	var look := ArmorLookScript.item_look(item_def_id)
	var colour: Color = look.get("color", Color.BLACK)
	return Color(colour.r, colour.g, colour.b).to_html(false)


func _headgear(hero: Node) -> Array:
	var out := []
	for node in hero.find_children("*", "MeshInstance3D", true, false):
		for suffix in ["_Helmet", "_HelmetVisor", "_BearHat", "_Hat"]:
			if str(node.name).ends_with(suffix):
				out.append(node)
	return out


func _check(cond: bool, msg: String) -> void:
	_checks += 1
	if not cond:
		_failed = true
		printerr("[gdtest] FAIL: ", msg)
