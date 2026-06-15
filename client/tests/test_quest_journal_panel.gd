# Unit test for the quest journal panel.
extends SceneTree

const QuestJournalPanelScript := preload("res://scripts/quest_journal_panel.gd")

var _failures: int = 0


func _initialize() -> void:
	var panel: QuestJournalPanel = QuestJournalPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	var empty := panel.get_debug_state()
	_assert_eq("empty summary", str(empty.get("summary", "")), "No active quests on this floor.")
	_assert_eq("empty count", int(empty.get("count", -1)), 0)
	panel.set_objectives([{"id": "reward_chest", "title": "Open the marked reward chest", "complete": false}])
	panel.ensure_display_visible()
	var active := panel.get_debug_state()
	_assert_true("panel visible", bool(active.get("visible", false)))
	_assert_eq("active count", int(active.get("count", -1)), 1)
	_assert_true("active objective incomplete", not bool((active.get("objectives", [])[0] as Dictionary).get("complete", true)))
	panel.set_objectives([{"id": "reward_chest", "title": "Open the marked reward chest", "complete": true}])
	var complete := panel.get_debug_state()
	_assert_true("complete objective", bool((complete.get("objectives", [])[0] as Dictionary).get("complete", false)))
	panel.toggle()
	_assert_true("toggle hides panel", not bool(panel.get_debug_state().get("visible", true)))
	panel.free()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_quest_journal_panel")
	quit(0)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
