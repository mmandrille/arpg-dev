class_name BlacksmithResourceCraftSlot
extends Button

const InventoryTransferRouterScript := preload("res://scripts/inventory_transfer_router.gd")

var panel: BlacksmithPanel
var item: Dictionary = {}


func _draw() -> void:
	if item.is_empty():
		return
	panel._draw_item_icon(self, item)


func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton \
			and event.button_index == MOUSE_BUTTON_LEFT \
			and event.pressed \
			and event.double_click \
			and not item.is_empty():
		panel.unstage_resource()
		accept_event()


func _get_drag_data(_at_position: Vector2) -> Variant:
	if item.is_empty():
		return null
	var data := {
		"source": InventoryTransferRouterScript.DRAG_SOURCE_BLACKSMITH_RESOURCE_STAGE,
		"item": item.duplicate(true),
		"blacksmith_panel": panel,
	}
	set_drag_preview(panel._drag_preview(item))
	return data


func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
	if typeof(data) != TYPE_DICTIONARY or typeof(data.get("item", {})) != TYPE_DICTIONARY:
		return false
	var dragged: Dictionary = data.get("item", {})
	if not BlacksmithPanel.is_craft_resource(dragged):
		return false
	var source := str(data.get("source", ""))
	return source == "bag"


func _drop_data(_at_position: Vector2, data: Variant) -> void:
	panel.stage_resource_item(data.get("item", {}))
