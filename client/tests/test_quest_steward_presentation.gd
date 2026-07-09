extends SceneTree

const QuestStewardPresentationScript := preload("res://scripts/quest_steward_presentation.gd")


func _init() -> void:
	_test_label_text()
	_test_count_turn_in_items()
	quit()


func _test_label_text() -> void:
	if QuestStewardPresentationScript.label_text(0) != "":
		_fail("empty count should hide label text")
	if QuestStewardPresentationScript.label_text(1) != "reclaim your reward":
		_fail("single item label mismatch")
	if QuestStewardPresentationScript.label_text(3) != "reclaim your reward x 3":
		_fail("multi item label mismatch")


func _test_count_turn_in_items() -> void:
	var inventory := [
		{"item_def_id": "quest_trophy_bat_wing"},
		{"item_def_id": "rusty_sword"},
	]
	var bag := [
		{"item_def_id": "quest_leaf"},
	]
	var count := QuestStewardPresentationScript.count_turn_in_items(inventory, bag)
	if count != 2:
		_fail("expected 2 quest turn-in items, got %d" % count)


func _fail(message: String) -> void:
	push_error(message)
	quit(1)
