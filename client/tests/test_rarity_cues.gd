## v508: one catalog drives visible rarity cues without rewriting authoritative item data.
extends SceneTree

const Loader := preload("res://scripts/rarity_cue_loader.gd")
const Presenter := preload("res://scripts/rarity_cue_presenter.gd")
const LootFactory := preload("res://scripts/loot_node_factory.gd")
const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")

var _passed := 0
var _failed := 0


func _initialize() -> void:
	ItemRulesLoaderScript.ensure_loaded()
	_test_all_five_rarities()
	_test_exclusions()
	_test_small_slot_bounds()
	print("[gdtest] PASS: test_rarity_cues (%d passed, %d failed)" % [_passed, _failed])
	quit(1 if _failed else 0)


func _test_all_five_rarities() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/item_templates.v0.json")
	var rules: Dictionary = JSON.parse_string(FileAccess.get_file_as_string(path))
	var rarities: Dictionary = rules.get("rarities", {})
	var entries: Dictionary = Loader.catalog().get("rarities", {})
	_assert("catalog has exact rarity count", entries.size() == rarities.size())
	var factory = LootFactory.new({}, {})
	for rarity in rarities.keys():
		var item := {"item_def_id": "long_sword", "rarity": rarity, "display_name": "Long Sword"}
		var original := item.duplicate(true)
		var cue := Loader.cue_for_item(item)
		_assert("%s cue maps to catalog" % rarity, cue == entries.get(rarity, {}))
		_assert("%s uppercase normalized" % rarity, Loader.cue_for_rarity(str(rarity).to_upper()) == cue)
		var node: Node3D = factory.make_loot_node(item)
		var marker := node.find_child("RarityCue", false, false) as Label3D
		var label := node.find_child("LootLabel", false, false) as Label3D
		_assert("%s persistent marker" % rarity, marker != null and marker.visible and marker.text == str(cue.get("world_symbol", "")))
		_assert("%s revealed full label" % rarity, label != null and not label.visible and label.text == "%s · Long Sword" % cue.get("name", ""))
		_assert("%s revealed label sits above marker" % rarity, label != null and marker != null and label.position.y > marker.position.y)
		_assert("%s input unchanged" % rarity, item == original)
		node.free()
	var named := {"item_def_id": "long_sword", "rarity": "rare", "display_name": "Rare Long Sword"}
	_assert("existing rarity prefix is not duplicated", factory.loot_label_text(named) == "Rare Long Sword")
	_assert("template-only trade row resolves", Loader.cue_for_item({"item_template_id": "long_sword", "rarity": "rare"}) == entries.get("rare", {}))


func _test_exclusions() -> void:
	var factory = LootFactory.new({}, {})
	for def_id in ["gold", "red_potion", "quest_leaf", "stat_badge"]:
		var item := {"item_def_id": def_id, "rarity": "rare"}
		_assert("%s category excluded" % def_id, Loader.cue_for_item(item).is_empty())
		var node: Node3D = factory.make_loot_node(item)
		_assert("%s ground marker absent" % def_id, node.find_child("RarityCue", false, false) == null)
		node.free()
	for item in [
		{"item_def_id": "long_sword"},
		{"item_def_id": "long_sword", "rarity": ""},
		{"item_def_id": "long_sword", "rarity": "unknown"},
		{"item_def_id": "missing", "rarity": "rare"},
		{"item_def_id": "long_sword", "rarity": "rare", "concealed": true},
		{"item_def_id": "long_sword", "rarity": "rare", "kind": "mystery"},
		{"item_def_id": "long_sword", "rarity": "rare", "offer_id": "mystery:offer"},
	]:
		_assert("unknown or concealed item excluded: %s" % str(item), Loader.cue_for_item(item).is_empty())
	var unknown := factory.make_loot_node({"item_def_id": "long_sword", "rarity": "unrecognized"})
	_assert("unknown rarity has no ground marker", unknown.find_child("RarityCue", false, false) == null)
	unknown.free()


func _test_small_slot_bounds() -> void:
	var minimum := float(Loader.catalog().get("slot", {}).get("minimum_size_px", 0))
	for size in [Vector2(50, 50), Vector2(52, 52), Vector2(68, 54), Vector2(96, 58), Vector2(84, 84)]:
		var slot := Rect2(Vector2.ZERO, size)
		var badge := Presenter.slot_badge_rect(slot)
		_assert("badge fits slot %s" % size, slot.encloses(badge) and badge.size.x >= minimum and badge.size.y >= minimum)


func _assert(label: String, condition: bool) -> void:
	if condition:
		_passed += 1
	else:
		_failed += 1
		push_error("[gdtest] FAIL: %s" % label)
