# Unit tests for ground-loot presentation polish.
extends SceneTree

const LootNodeFactoryScript := preload("res://scripts/loot_node_factory.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_common_gold_has_glow_marker_and_label()
	_test_rare_equipment_keeps_model_and_rarity_glow()
	_test_rare_equipment_has_pickup_beam()
	print("[gdtest] PASS: test_loot_node_factory (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_common_gold_has_glow_marker_and_label() -> void:
	var factory = LootNodeFactoryScript.new({}, {})
	var node := factory.make_loot_node({"id": "loot_gold", "type": "loot", "item_def_id": "gold", "rarity": "common", "amount": 12})
	_assert_true("common loot glow exists", node.find_child("RarityGlow", false, false) != null)
	_assert_true("common spawn pop exists", node.find_child("SpawnPopRing", false, false) != null)
	var label := node.find_child("LootLabel", false, false) as Label3D
	_assert_true("gold label exists", label != null)
	_assert_eq("gold label text", label.text if label != null else "", "12 gold")
	node.free()


func _test_rare_equipment_keeps_model_and_rarity_glow() -> void:
	var factory = LootNodeFactoryScript.new({}, {
		"starter_paladin_sword": {
			"ground": {"shape": "blade", "color": "#b8c7d8", "accent": "#f6e8b1", "scale": 1.0},
			"3d_model": "fallback_equipment_main_hand_v0",
		},
	})
	var node := factory.make_loot_node({"id": "loot_sword", "type": "loot", "item_def_id": "starter_paladin_sword", "rarity": "rare"})
	_assert_true("rare loot glow exists", node.find_child("RarityGlow", false, false) != null)
	_assert_true("rare spawn pop exists", node.find_child("SpawnPopRing", false, false) != null)
	_assert_true("rare primitive remains", node.find_child("Blade", false, false) != null)
	var glow := node.find_child("RarityGlow", false, false) as MeshInstance3D
	var mat := glow.material_override as StandardMaterial3D
	_assert_true("rare glow uses warm rarity color", mat != null and mat.albedo_color.r > mat.albedo_color.b)
	node.free()


func _test_rare_equipment_has_pickup_beam() -> void:
	var factory = LootNodeFactoryScript.new({}, {
		"starter_paladin_sword": {
			"ground": {"shape": "blade", "color": "#b8c7d8", "accent": "#f6e8b1", "scale": 1.0},
			"3d_model": "fallback_equipment_main_hand_v0",
		},
	})
	var node := factory.make_loot_node({"id": "loot_sword", "type": "loot", "item_def_id": "starter_paladin_sword", "rarity": "rare"})
	_assert_true("rare pickup beam exists", node.find_child("PickupBeam", false, false) != null)
	var common := factory.make_loot_node({"id": "loot_gold", "type": "loot", "item_def_id": "gold", "rarity": "common", "amount": 3})
	_assert_true("common loot has no pickup beam", common.find_child("PickupBeam", false, false) == null)
	node.free()
	common.free()


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
