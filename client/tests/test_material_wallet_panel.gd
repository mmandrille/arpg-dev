# Unit test for resource bag grid in the material wallet panel.
extends SceneTree

const MaterialWalletPanelScript := preload("res://scripts/material_wallet_panel.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var panel = MaterialWalletPanelScript.new()
	get_root().add_child(panel)
	await process_frame

	panel.show_wallet({}, [])
	var empty_state := panel.get_debug_state()
	_assert_eq("empty starts at base rows", int(empty_state.get("wallet_rows", 0)), MaterialWalletPanelScript.BASE_WALLET_ROWS)

	var bag_items: Array = []
	for i in range(MaterialWalletPanelScript.WALLET_COLUMNS * MaterialWalletPanelScript.BASE_WALLET_ROWS):
		bag_items.append({
			"stash_item_id": "bag_%d" % i,
			"item_def_id": "upgrade_shard",
			"rolled_stats": {"item_level": 1},
		})
	panel.show_wallet({}, bag_items)
	var full_state := panel.get_debug_state()
	_assert_eq("full base grid rows", int(full_state.get("wallet_rows", 0)), MaterialWalletPanelScript.BASE_WALLET_ROWS)
	_assert_eq("full base grid capacity", int(full_state.get("wallet_capacity", 0)), MaterialWalletPanelScript.WALLET_COLUMNS * MaterialWalletPanelScript.BASE_WALLET_ROWS)

	bag_items.append({
		"stash_item_id": "bag_overflow",
		"item_def_id": "renew_stone",
		"rolled_stats": {"item_level": 2},
	})
	panel.set_wallet({}, bag_items)
	var grown_state := panel.get_debug_state()
	_assert_eq("auto-grow rows", int(grown_state.get("wallet_rows", 0)), MaterialWalletPanelScript.BASE_WALLET_ROWS + 1)
	_assert_eq("bag item count", int(grown_state.get("bag_item_count", 0)), bag_items.size())

	panel.free()
	print("[gdtest] PASS: test_material_wallet_panel (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
