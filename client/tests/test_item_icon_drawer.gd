extends SceneTree

const ItemRulesLoaderScript := preload("res://scripts/item_rules_loader.gd")


func _initialize() -> void:
	ItemRulesLoaderScript.ensure_loaded()
	var presentations: Dictionary = ItemRulesLoaderScript.item_presentations
	_assert_shape(presentations, "long_sword", "blade")
	_assert_shape(presentations, "spear", "spear")
	_assert_shape(presentations, "short_spear", "spear")
	_assert_shape(presentations, "halberd", "halberd")
	_assert_shape(presentations, "mace", "mace")
	_assert_shape(presentations, "hammer", "hammer")
	_assert_shape(presentations, "morningstar", "mace")
	_assert_shape(presentations, "barbarian_axe", "axe")
	print("test_item_icon_drawer: ok")
	quit(0)


func _assert_shape(presentations: Dictionary, def_id: String, expected_shape: String) -> void:
	var icon: Dictionary = presentations.get(def_id, {}).get("icon", {})
	var shape := str(icon.get("shape", ""))
	if shape != expected_shape:
		push_error("expected %s icon shape %s, got %s" % [def_id, expected_shape, shape])
		quit(1)
