# Headless unit tests for LocalPlayerAuthoritativeSync.
extends SceneTree

const LocalPlayerAuthoritativeSyncScript := preload("res://scripts/local_player_authoritative_sync.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	_assert_false("matching hp unchanged", LocalPlayerAuthoritativeSyncScript.hp_changed({"hp": 10, "max_hp": 10}, 10, 10))
	_assert_true("hp delta detected", LocalPlayerAuthoritativeSyncScript.hp_changed({"hp": 9, "max_hp": 10}, 10, 10))
	_assert_false("matching mana unchanged", LocalPlayerAuthoritativeSyncScript.mana_changed({"mana": 5, "max_mana": 10}, 5, 10))
	_assert_true("mana delta detected", LocalPlayerAuthoritativeSyncScript.mana_changed({"mana": 4, "max_mana": 10}, 5, 10))
	_assert_false("matching effects unchanged", LocalPlayerAuthoritativeSyncScript.effect_ids_changed(["a", "b"], ["a", "b"]))
	_assert_true("effect order change detected", LocalPlayerAuthoritativeSyncScript.effect_ids_changed(["b", "a"], ["a", "b"]))
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
		print("[gdtest] PASS: test_local_player_authoritative_sync (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_local_player_authoritative_sync (%d failures)" % _fail)
		quit(1)
