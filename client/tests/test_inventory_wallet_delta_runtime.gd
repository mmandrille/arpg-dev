# Direct unit tests for InventoryWalletDeltaRuntime (extraction independence).
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const InventoryWalletDeltaRuntimeScript := preload("res://scripts/inventory_wallet_delta_runtime.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	var host := MainScript.new()
	host.inventory = [{"item_instance_id": "ii_1", "hp": 10}]
	InventoryWalletDeltaRuntimeScript.update_inventory_item(host, {"item_instance_id": "ii_1", "hp": 3})
	_assert_eq("update_inventory_item patches existing row", int(host.inventory[0].get("hp", -1)), 3)
	InventoryWalletDeltaRuntimeScript.apply_change(host, {"op": "gold_update", "gold": 99})
	_assert_eq("apply_change gold_update", host.gold, 99)
	if _fail == 0:
		print("[gdtest] PASS: test_inventory_wallet_delta_runtime (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_inventory_wallet_delta_runtime (%d failures)" % _fail)
		quit(1)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
		return
	_fail += 1
	print("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])
