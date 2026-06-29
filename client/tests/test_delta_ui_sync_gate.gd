# Headless unit tests for DeltaUiSyncGate.
extends SceneTree

const DeltaUiSyncGateScript := preload("res://scripts/delta_ui_sync_gate.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	var gate := DeltaUiSyncGateScript.new()
	_assert_true("fresh gate syncs inventory", gate.should_sync_inventory(3))
	gate.note_synced(true, true, true)
	gate.on_delta_tick()
	gate.on_delta_tick()
	gate.on_delta_tick()
	_assert_true("interval triggers inventory sync", gate.should_sync_inventory(3))
	gate.note_synced(true, false, false)
	_assert_false("counter reset after sync", gate.should_sync_inventory(3))
	_assert_true("entity elite marks quest dirty", _quest_dirty_after_entity({"type": "monster", "elite_objective": {}}))
	_finish()


func _quest_dirty_after_entity(entity: Dictionary) -> bool:
	var gate := DeltaUiSyncGateScript.new()
	gate.quest_tracker_dirty = false
	gate.mark_entity_change("entity_update", entity)

	return gate.quest_tracker_dirty


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _finish() -> void:
	if _fail == 0:
		print("[gdtest] PASS: test_delta_ui_sync_gate (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_delta_ui_sync_gate (%d failures)" % _fail)
		quit(1)
