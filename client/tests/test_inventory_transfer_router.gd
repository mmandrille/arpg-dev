# Unit test for inventory transfer routing.
# Run via: godot --headless --path client --script res://tests/test_inventory_transfer_router.gd
extends SceneTree

const Router := preload("res://scripts/inventory_transfer_router.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var item := {"item_instance_id": "2001", "item_def_id": "long_sword"}
	_assert_intent("double shop sell", Router.double_click_route(item, "shop_1", "", false, false, "", {}, false), "shop_sell_intent")
	_assert_eq("double shop entity", _payload(Router.double_click_route(item, "shop_1", "", false, false, "", {}, false)).get("shop_entity_id", ""), "shop_1")
	_assert_intent("double market stage", Router.double_click_route(item, "", "offer", false, false, "", {}, false), "market_stage_inventory_item")
	_assert_eq("double market context", _payload(Router.double_click_route(item, "", "offer", false, false, "", {}, false)).get("context", ""), "offer")
	_assert_intent("double blacksmith stage", Router.double_click_route(item, "", "", true, false, "", {}, false), "blacksmith_stage_inventory_item")
	_assert_intent("double equip", Router.double_click_route(item, "", "", false, false, "main_hand", {"weapon_set": 1}, false), "equip_intent")
	_assert_eq("double equip weapon set", int(_payload(Router.double_click_route(item, "", "", false, false, "main_hand", {"weapon_set": 1}, false)).get("weapon_set", -1)), 1)
	_assert_intent("double use", Router.double_click_route(item, "", "", false, false, "", {}, true), "use_intent")
	_assert_eq("equipped item ignores shop context", Router.double_click_route(item, "shop_1", "", false, true, "", {}, false), {})

	_assert_eq("non-consumable shift ignored", Router.shift_click_route(item, false, 0), {})
	_assert_eq("full hotbar shift ignored", Router.shift_click_route(item, true, -1), {})
	_assert_intent("shift assigns hotbar", Router.shift_click_route(item, true, 2), "assign_hotbar_intent")
	_assert_eq("shift slot", int(_payload(Router.shift_click_route(item, true, 2)).get("slot_index", -1)), 2)
	_assert_eq("equipment slot recognized", Router.is_equipment_slot("equip:main_hand"), true)
	_assert_eq("bag is not equipment slot", Router.is_equipment_slot("bag"), false)
	_assert_eq("equipment slot parsed", Router.slot_from_kind("equip:off_hand"), "off_hand")
	_assert_eq("non-equipment slot parsed empty", Router.slot_from_kind("bag"), "")

	_assert_intent("shop offer to bag", Router.drop_route("bag", {
		"source": "shop_offer", "shop_entity_id": "shop_1", "offer_id": "offer_1", "item": item,
	}, false), "shop_buy_intent")
	_assert_intent("stash to bag", Router.drop_route("bag", {
		"source": "stash", "stash_entity_id": "stash_1", "stash_item_id": "stash_item_1", "item": item,
	}, false), "stash_withdraw_item_intent")
	_assert_intent("corpse to bag", Router.drop_route("bag", {
		"source": "corpse", "corpse_entity_id": "corpse_1", "item_instance_id": "dead_item_1", "item": item,
	}, false), "corpse_withdraw_item_intent")
	_assert_intent("unique chest to bag", Router.drop_route("bag", {
		"source": "unique_chest", "stash_entity_id": "chest_1", "stash_item_id": "chest_item_1", "item": item,
	}, false), "unique_chest_take_item_intent")
	_assert_eq("blacksmith stage to bag", str(Router.drop_route("bag", {
		"source": "blacksmith_stage", "item": item,
	}, false).get("kind", "")), Router.KIND_BLACKSMITH_UNSTAGE)
	_assert_eq("blacksmith resource stage to bag", str(Router.drop_route("bag", {
		"source": "blacksmith_resource_stage", "item": {"item_def_id": "upgrade_shard"},
	}, false).get("unstage", "")), "resource")
	_assert_intent("equipment to bag", Router.drop_route("bag", {
		"source": "equip:main_hand", "item": item,
	}, false, {"weapon_set": 1}), "unequip_intent")

	_assert_eq("ineligible equip ignored", Router.drop_route("equip:main_hand", {
		"source": "bag", "item": item,
	}, false), {})
	_assert_intent("bag to equipment", Router.drop_route("equip:main_hand", {
		"source": "bag", "item": item,
	}, true, {"weapon_set": 1}), "equip_intent")
	_assert_intent("stash to equipment", Router.drop_route("equip:main_hand", {
		"source": "stash", "stash_entity_id": "stash_1", "stash_item_id": "stash_item_1", "item": item,
	}, true), "stash_equip_item_intent")
	_assert_intent("corpse to equipment", Router.drop_route("equip:main_hand", {
		"source": "corpse", "corpse_entity_id": "corpse_1", "item_instance_id": "dead_item_1", "item": item,
	}, true), "corpse_withdraw_item_intent")
	_assert_eq("blacksmith stage to equipment", str(Router.drop_route("equip:main_hand", {
		"source": "blacksmith_stage", "item": item,
	}, true).get("kind", "")), Router.KIND_BLACKSMITH_UNSTAGE)

	print("[gdtest] PASS: test_inventory_transfer_router (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _payload(decision: Dictionary) -> Dictionary:
	return decision.get("payload", {})


func _assert_intent(label: String, decision: Dictionary, expected_type: String) -> void:
	_assert_eq("%s kind" % label, str(decision.get("kind", "")), Router.KIND_INTENT)
	_assert_eq("%s type" % label, str(decision.get("intent_type", "")), expected_type)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
