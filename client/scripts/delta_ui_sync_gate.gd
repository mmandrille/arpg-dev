class_name DeltaUiSyncGate
extends RefCounted

var inventory_dirty := true
var quest_tracker_dirty := true
var minimap_dirty := true
var ticks_since_sync := 0


func reset_from_snapshot() -> void:
	inventory_dirty = true
	quest_tracker_dirty = true
	minimap_dirty = true
	ticks_since_sync = 0


func mark_inventory_dirty() -> void:
	inventory_dirty = true


func mark_quest_tracker_dirty() -> void:
	quest_tracker_dirty = true


func mark_minimap_dirty() -> void:
	minimap_dirty = true


func mark_entity_change(op: String, entity: Dictionary) -> void:
	if entity.is_empty():
		return
	if entity.has("elite_objective") or entity.has("quest_reward"):
		quest_tracker_dirty = true
	var entity_type := str(entity.get("type", ""))
	if entity_type == "interactable":
		minimap_dirty = true
	elif op == "entity_spawn" and entity_type in ["monster", "interactable", "loot"]:
		minimap_dirty = true


func mark_entity_removed() -> void:
	minimap_dirty = true
	quest_tracker_dirty = true


func on_delta_tick() -> void:
	ticks_since_sync += 1


func should_sync_inventory(interval_ticks: int) -> bool:
	return inventory_dirty or ticks_since_sync >= interval_ticks


func should_sync_quest_tracker(interval_ticks: int) -> bool:
	return quest_tracker_dirty or ticks_since_sync >= interval_ticks


func should_sync_minimap(interval_ticks: int) -> bool:
	return minimap_dirty or ticks_since_sync >= interval_ticks


func note_synced(synced_inventory: bool, synced_quest: bool, synced_minimap: bool) -> void:
	if synced_inventory:
		inventory_dirty = false
	if synced_quest:
		quest_tracker_dirty = false
	if synced_minimap:
		minimap_dirty = false
	if synced_inventory or synced_quest or synced_minimap:
		ticks_since_sync = 0
