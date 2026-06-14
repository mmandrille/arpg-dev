class_name InventoryTransferRouter
extends RefCounted

const KIND_INTENT := "intent"
const KIND_BLACKSMITH_UNSTAGE := "blacksmith_unstage"
const SLOT_KIND_BAG := "bag"
const SLOT_KIND_EQUIP_PREFIX := "equip:"
const DRAG_SOURCE_SHOP_OFFER := "shop_offer"
const DRAG_SOURCE_STASH := "stash"
const DRAG_SOURCE_CORPSE := "corpse"
const DRAG_SOURCE_UNIQUE_CHEST := "unique_chest"
const DRAG_SOURCE_BLACKSMITH_STAGE := "blacksmith_stage"


static func double_click_route(
		item: Dictionary,
		shop_sell_entity_id: String,
		market_context: String,
		blacksmith_context_enabled: bool,
		is_equipped_instance: bool,
		preferred_slot: String,
		weapon_set_payload: Dictionary,
		is_consumable: bool
) -> Dictionary:
	var item_instance_id := str(item.get("item_instance_id", ""))
	if shop_sell_entity_id != "" and not is_equipped_instance:
		return _intent("shop_sell_intent", {
			"shop_entity_id": shop_sell_entity_id,
			"item_instance_id": item_instance_id,
		})
	if market_context != "" and not is_equipped_instance:
		return _intent("market_stage_inventory_item", {
			"context": market_context,
			"item": item.duplicate(true),
		})
	if blacksmith_context_enabled and not is_equipped_instance:
		return _intent("blacksmith_stage_inventory_item", {
			"item": item.duplicate(true),
		})
	if preferred_slot != "":
		var equip_payload := {"item_instance_id": item_instance_id, "slot": preferred_slot}
		equip_payload.merge(weapon_set_payload, true)
		return _intent("equip_intent", equip_payload)
	if is_consumable:
		return _intent("use_intent", {"item_instance_id": item_instance_id})
	return {}


static func shift_click_route(item: Dictionary, is_consumable: bool, first_empty_hotbar_slot: int) -> Dictionary:
	if not is_consumable or first_empty_hotbar_slot < 0:
		return {}
	return _intent("assign_hotbar_intent", {
		"slot_index": first_empty_hotbar_slot,
		"item_instance_id": str(item.get("item_instance_id", "")),
	})


static func drop_route(slot_kind: String, data: Dictionary, can_equip_to_slot: bool, weapon_set_payload: Dictionary = {}) -> Dictionary:
	var item: Dictionary = data.get("item", {})
	if item.is_empty():
		return {}
	var source := str(data.get("source", ""))
	if is_equipment_slot(slot_kind):
		if not can_equip_to_slot:
			return {}
		var slot := slot_from_kind(slot_kind)
		if source == DRAG_SOURCE_STASH:
			return _intent("stash_equip_item_intent", {
				"stash_entity_id": str(data.get("stash_entity_id", "")),
				"stash_item_id": str(data.get("stash_item_id", "")),
				"slot": slot,
			})
		if source == DRAG_SOURCE_CORPSE:
			return _intent("corpse_withdraw_item_intent", {
				"corpse_entity_id": str(data.get("corpse_entity_id", "")),
				"item_instance_id": str(data.get("item_instance_id", "")),
			})
		if source == DRAG_SOURCE_BLACKSMITH_STAGE:
			return {"kind": KIND_BLACKSMITH_UNSTAGE}
		var equip_payload := {"item_instance_id": str(item.get("item_instance_id", "")), "slot": slot}
		equip_payload.merge(weapon_set_payload, true)
		return _intent("equip_intent", equip_payload)
	if slot_kind != SLOT_KIND_BAG:
		return {}
	if source == DRAG_SOURCE_SHOP_OFFER:
		return _intent("shop_buy_intent", {
			"shop_entity_id": str(data.get("shop_entity_id", "")),
			"offer_id": str(data.get("offer_id", "")),
		})
	if source == DRAG_SOURCE_STASH:
		return _intent("stash_withdraw_item_intent", {
			"stash_entity_id": str(data.get("stash_entity_id", "")),
			"stash_item_id": str(data.get("stash_item_id", "")),
		})
	if source == DRAG_SOURCE_CORPSE:
		return _intent("corpse_withdraw_item_intent", {
			"corpse_entity_id": str(data.get("corpse_entity_id", "")),
			"item_instance_id": str(data.get("item_instance_id", "")),
		})
	if source == DRAG_SOURCE_UNIQUE_CHEST:
		return _intent("unique_chest_take_item_intent", {
			"chest_entity_id": str(data.get("stash_entity_id", "")),
			"chest_item_id": str(data.get("stash_item_id", "")),
		})
	if source == DRAG_SOURCE_BLACKSMITH_STAGE:
		return {"kind": KIND_BLACKSMITH_UNSTAGE}
	if is_equipment_slot(source):
		var unequip_slot := slot_from_kind(source)
		var unequip_payload := {"slot": unequip_slot}
		unequip_payload.merge(weapon_set_payload, true)
		return _intent("unequip_intent", unequip_payload)
	return {}


static func is_equipment_slot(kind: String) -> bool:
	return kind.begins_with(SLOT_KIND_EQUIP_PREFIX)


static func slot_from_kind(kind: String) -> String:
	if not is_equipment_slot(kind):
		return ""
	return kind.substr(SLOT_KIND_EQUIP_PREFIX.length())


static func _intent(intent_type: String, payload: Dictionary) -> Dictionary:
	return {
		"kind": KIND_INTENT,
		"intent_type": intent_type,
		"payload": payload,
	}
