class_name QuestStewardPresentation
extends RefCounted

const LABEL_NAME := "QuestRewardLabel"
const ACTIVE_COLOR := Color("#e64545")


static func count_turn_in_items(inventory: Array, resource_bag_items: Array) -> int:
	ItemRulesLoader.ensure_loaded()
	var count := 0
	for item in inventory:
		if _is_quest_turn_in(str((item as Dictionary).get("item_def_id", ""))):
			count += 1
	for item in resource_bag_items:
		if _is_quest_turn_in(str((item as Dictionary).get("item_def_id", ""))):
			count += 1

	return count


static func label_text(count: int) -> String:
	if count <= 0:
		return ""
	if count == 1:
		return "reclaim your reward"

	return "reclaim your reward x %d" % count


static func apply_to_steward(steward: Node3D, count: int) -> void:
	if steward == null:
		return
	var label := steward.find_child(LABEL_NAME, true, false) as Label3D
	if label == null:
		return
	var safe_count: int = max(count, 0)
	label.text = label_text(safe_count)
	label.modulate = ACTIVE_COLOR if safe_count > 0 else Color.WHITE
	label.visible = safe_count > 0
	label.set_meta("quest_turn_in_count", safe_count)


static func debug_state(steward: Node3D) -> Dictionary:
	if steward == null:
		return empty_state()
	var label := steward.find_child(LABEL_NAME, true, false) as Label3D
	if label == null:
		return empty_state()

	return {
		"exists": true,
		"count": int(label.get_meta("quest_turn_in_count", 0)),
		"text": label.text,
		"visible": label.visible,
	}


static func empty_state() -> Dictionary:
	return {
		"exists": false,
		"count": 0,
		"text": "",
		"visible": false,
	}


static func _is_quest_turn_in(def_id: String) -> bool:
	if def_id == "":
		return false

	return str(ItemRulesLoader.item_definition(def_id).get("category", "")) == "quest"
