extends SceneTree

const QuestStewardPresentationScript := preload("res://scripts/quest_steward_presentation.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_label_text()
	_test_count_turn_in_items()
	_finish()


func _test_label_text() -> void:
	_assert_eq("empty count hides label text", QuestStewardPresentationScript.label_text(0), "")
	_assert_eq("single item label", QuestStewardPresentationScript.label_text(1), "reclaim your reward")
	_assert_eq("multi item label", QuestStewardPresentationScript.label_text(3), "reclaim your reward x 3")


func _test_count_turn_in_items() -> void:
	var inventory := [
		{"item_def_id": "quest_trophy_bat_wing"},
		{"item_def_id": "rusty_sword"},
	]
	var bag := [
		{"item_def_id": "quest_leaf"},
	]
	var count := QuestStewardPresentationScript.count_turn_in_items(inventory, bag)
	_assert_eq("quest turn-in items across inventory and resource bag", count, 2)


func _assert_eq(label: String, got: Variant, want: Variant) -> void:
	if got == want:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_quest_steward_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_quest_steward_presentation (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
