extends RefCounted
class_name TownServiceBridge


static func open_market_inventory_context(inventory_panel: Node) -> void:
	if inventory_panel == null:
		return
	inventory_panel.call("ensure_display_visible")
	inventory_panel.call("set_market_context", "publish")


static func set_market_inventory_context(inventory_panel: Node, context: String) -> void:
	if inventory_panel == null:
		return
	if context == "":
		inventory_panel.call("clear_market_context")
	else:
		inventory_panel.call("set_market_context", context)


static func close_market_inventory_context(inventory_panel: Node) -> void:
	if inventory_panel == null:
		return
	inventory_panel.call("clear_market_context")


static func open_blacksmith_inventory_context(inventory_panel: Node) -> void:
	if inventory_panel == null:
		return
	inventory_panel.call("ensure_display_visible")
	inventory_panel.call("set_blacksmith_context", true)


static func close_blacksmith_inventory_context(inventory_panel: Node) -> void:
	if inventory_panel == null:
		return
	inventory_panel.call("set_blacksmith_context", false)


static func route_inventory_stage_intent(intent_type: String, payload: Dictionary, market_panel: Node, blacksmith_panel: Node) -> bool:
	if intent_type == "market_stage_inventory_item":
		if market_panel != null:
			market_panel.call("stage_inventory_item", str(payload.get("context", "")), payload.get("item", {}))
		return true
	if intent_type == "blacksmith_stage_inventory_item":
		if blacksmith_panel != null:
			var item: Dictionary = payload.get("item", {})
			var def_id := str(item.get("item_def_id", ""))
			if def_id == "upgrade_shard" or def_id == "renew_stone":
				blacksmith_panel.call("stage_resource_item", item)
			else:
				blacksmith_panel.call("stage_inventory_item", item)
		return true
	return false
