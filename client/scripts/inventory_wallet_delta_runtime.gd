class_name InventoryWalletDeltaRuntime
extends RefCounted

const WALLET_OPS := [
	"inventory_add",
	"inventory_update",
	"inventory_remove",
	"equipped_update",
	"weapon_set_update",
	"hotbar_update",
	"gold_update",
	"stash_item_add",
	"stash_item_remove",
	"stash_gold_update",
	"resource_wallet_update",
]


static func is_wallet_op(op: String) -> bool:
	return op in WALLET_OPS


static func apply_change(host, change: Dictionary) -> void:
	var op := str(change.get("op", ""))
	match op:
		"inventory_add":
			var inv_item: Dictionary = change.get("item", {})
			host.inventory.append(inv_item)
			host._delta_ui_sync_gate.mark_inventory_dirty()
			if host.resolver != null:
				host.resolver.ingest_inventory_item(inv_item)
		"inventory_update":
			var upd_item: Dictionary = change.get("item", {})
			update_inventory_item(host, upd_item)
			host._delta_ui_sync_gate.mark_inventory_dirty()
			if host.resolver != null:
				host.resolver.ingest_inventory_item(upd_item)
		"inventory_remove":
			remove_inventory_item(host, str(change.get("item_instance_id", "")))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"equipped_update":
			var slot := str(change.get("slot", ""))
			if not slot.is_empty():
				host.equipped[slot] = change.get("item_instance_id")
				if host.resolver != null:
					host.resolver.apply_equipped_update(slot, change.get("item_instance_id"))
			if change.has("inventory_rows"):
				host.inventory_rows = int(change.get("inventory_rows", host.inventory_rows))
			if change.has("inventory_capacity"):
				host.inventory_capacity = int(change.get("inventory_capacity", host.inventory_capacity))
			if change.has("hotbar_capacity"):
				host.hotbar_capacity = int(change.get("hotbar_capacity", host.hotbar_capacity))
				if host.consumable_bar != null:
					host.consumable_bar.set_hotbar_state(host.hotbar_capacity, host.hotbar)
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"weapon_set_update":
			host.active_weapon_set = int(change.get("active_weapon_set", host.active_weapon_set))
			host.weapon_sets = change.get("weapon_sets", host.weapon_sets)
			host._remount_local_equipment_visuals()
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"hotbar_update":
			if change.has("inventory_rows"):
				host.inventory_rows = int(change.get("inventory_rows", host.inventory_rows))
			if change.has("inventory_capacity"):
				host.inventory_capacity = int(change.get("inventory_capacity", host.inventory_capacity))
			apply_hotbar_update(host, int(change.get("slot_index", -1)), change.get("item_instance_id"), change.get("item", {}))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"gold_update":
			host.gold = int(change.get("gold", host.gold))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"stash_item_add":
			upsert_stash_item(host, change.get("item", {}))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"stash_item_remove":
			remove_stash_item(host, str(change.get("stash_item_id", "")))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"stash_gold_update":
			host.stash_gold = int(change.get("stash_gold", host.stash_gold))
			host._delta_ui_sync_gate.mark_inventory_dirty()
		"resource_wallet_update":
			var resource_id := str(change.get("resource_id", ""))
			if resource_id != "":
				host.resource_wallet[resource_id] = max(0, int(change.get("amount", host.resource_wallet.get(resource_id, 0))))
			host._delta_ui_sync_gate.mark_inventory_dirty()


static func update_inventory_item(host, item: Dictionary) -> void:
	for i in range(host.inventory.size()):
		if host.inventory[i]["item_instance_id"] == item["item_instance_id"]:
			host.inventory[i] = item
			return
	host.inventory.append(item)


static func remove_inventory_item(host, item_instance_id: String) -> void:
	for i in range(host.inventory.size() - 1, -1, -1):
		if str(host.inventory[i].get("item_instance_id", "")) == item_instance_id:
			host.inventory.remove_at(i)


static func remove_inventory_items_by_def(host, item_def_id: String, count: int) -> void:
	if item_def_id == "" or count <= 0:
		return
	var removed := 0
	for i in range(host.inventory.size() - 1, -1, -1):
		if str(host.inventory[i].get("item_def_id", "")) == item_def_id:
			host.inventory.remove_at(i)
			removed += 1
			if removed >= count:
				return


static func upsert_stash_item(host, item: Dictionary) -> void:
	var stash_item_id := str(item.get("stash_item_id", ""))
	if stash_item_id == "":
		return
	for i in range(host.stash_items.size()):
		if str(host.stash_items[i].get("stash_item_id", "")) == stash_item_id:
			var merged: Dictionary = host.stash_items[i].duplicate(true)
			merged.merge(item, true)
			host.stash_items[i] = merged
			return
	host.stash_items.append(item.duplicate(true))


static func remove_stash_item(host, stash_item_id: String) -> void:
	for i in range(host.stash_items.size() - 1, -1, -1):
		if str(host.stash_items[i].get("stash_item_id", "")) == stash_item_id:
			host.stash_items.remove_at(i)


static func apply_resource_wallet_snapshot(host, rows: Variant) -> void:
	host.resource_wallet.clear()
	if typeof(rows) != TYPE_ARRAY:
		return
	for value in rows:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var row := value as Dictionary
		var resource_id := str(row.get("resource_id", ""))
		if resource_id != "":
			host.resource_wallet[resource_id] = max(0, int(row.get("amount", 0)))


static func apply_hotbar_update(host, slot_index: int, item_instance_id, item: Dictionary = {}) -> void:
	if slot_index < 0 or slot_index >= 10:
		return
	while host.hotbar.size() < 10:
		host.hotbar.append({"slot_index": host.hotbar.size(), "item_instance_id": null})
	var slot := {"slot_index": slot_index, "item_instance_id": item_instance_id}
	if not item.is_empty():
		slot["item"] = item.duplicate(true)
	host.hotbar[slot_index] = slot
	if host.consumable_bar != null:
		host.consumable_bar.apply_hotbar_update(slot_index, item_instance_id, item)
