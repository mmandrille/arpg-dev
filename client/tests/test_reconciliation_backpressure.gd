# Headless unit tests for ReconciliationBackpressure (v378).
extends SceneTree

const ReconciliationBackpressureScript := preload("res://scripts/reconciliation_backpressure.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	MainConfigLoaderScript.ensure_loaded()
	var threshold := MainConfigLoaderScript.reconciliation_backpressure_threshold()
	_assert_false("below threshold keeps pending", ReconciliationBackpressureScript.should_clear_pending_targets(threshold - 0.1, threshold))
	_assert_true("at threshold clears pending", ReconciliationBackpressureScript.should_clear_pending_targets(threshold, threshold))
	_assert_true("above threshold clears pending", ReconciliationBackpressureScript.should_clear_pending_targets(threshold + 1.0, threshold))
	_finish()


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
		print("[gdtest] PASS: test_reconciliation_backpressure (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_reconciliation_backpressure (%d failures)" % _fail)
		quit(1)
