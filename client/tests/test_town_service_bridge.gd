# GDScript town service bridge test.
# Run via: godot --headless --path client --script res://tests/test_town_service_bridge.gd
extends SceneTree

const TownServiceBridgeScript := preload("res://scripts/town_service_bridge.gd")

var _pass_count: int = 0
var _fail_count: int = 0


class FakeInventoryPanel:
	extends Node

	var visible_count: int = 0
	var market_context: String = ""
	var blacksmith_context: bool = false

	func ensure_display_visible() -> void:
		visible_count += 1

	func set_market_context(context: String) -> void:
		market_context = context

	func clear_market_context() -> void:
		market_context = ""

	func set_blacksmith_context(enabled: bool) -> void:
		blacksmith_context = enabled


class FakeMarketPanel:
	extends Node

	var staged_context: String = ""
	var staged_item: Dictionary = {}

	func stage_inventory_item(context: String, item: Dictionary) -> void:
		staged_context = context
		staged_item = item.duplicate(true)


class FakeBlacksmithPanel:
	extends Node

	var staged_item: Dictionary = {}
	var staged_resource: Dictionary = {}

	func stage_inventory_item(item: Dictionary) -> void:
		staged_item = item.duplicate(true)

	func stage_resource_item(item: Dictionary) -> void:
		staged_resource = item.duplicate(true)


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var inventory_panel := FakeInventoryPanel.new()
	var market_panel := FakeMarketPanel.new()
	var blacksmith_panel := FakeBlacksmithPanel.new()
	root.add_child(inventory_panel)
	root.add_child(market_panel)
	root.add_child(blacksmith_panel)

	TownServiceBridgeScript.open_market_inventory_context(inventory_panel)
	_assert_eq("market open shows inventory", inventory_panel.visible_count, 1)
	_assert_eq("market open publish context", inventory_panel.market_context, "publish")

	TownServiceBridgeScript.set_market_inventory_context(inventory_panel, "offer")
	_assert_eq("market context switches", inventory_panel.market_context, "offer")
	TownServiceBridgeScript.set_market_inventory_context(inventory_panel, "")
	_assert_eq("empty market context clears", inventory_panel.market_context, "")

	TownServiceBridgeScript.open_blacksmith_inventory_context(inventory_panel)
	_assert_eq("blacksmith open shows inventory", inventory_panel.visible_count, 2)
	_assert_eq("blacksmith context enabled", inventory_panel.blacksmith_context, true)
	TownServiceBridgeScript.close_blacksmith_inventory_context(inventory_panel)
	_assert_eq("blacksmith context disabled", inventory_panel.blacksmith_context, false)

	var item := {"item_instance_id": "2001", "item_def_id": "bow"}
	var handled_market := TownServiceBridgeScript.route_inventory_stage_intent(
		"market_stage_inventory_item",
		{"context": "offer", "item": item},
		market_panel,
		blacksmith_panel
	)
	_assert_eq("market stage handled", handled_market, true)
	_assert_eq("market stage context", market_panel.staged_context, "offer")
	_assert_eq("market stage item", str(market_panel.staged_item.get("item_instance_id", "")), "2001")

	var handled_blacksmith := TownServiceBridgeScript.route_inventory_stage_intent(
		"blacksmith_stage_inventory_item",
		{"item": item},
		market_panel,
		blacksmith_panel
	)
	_assert_eq("blacksmith stage handled", handled_blacksmith, true)
	_assert_eq("blacksmith stage item", str(blacksmith_panel.staged_item.get("item_instance_id", "")), "2001")
	var shard := {"item_instance_id": "shard1", "item_def_id": "upgrade_shard"}
	var handled_shard := TownServiceBridgeScript.route_inventory_stage_intent(
		"blacksmith_stage_inventory_item",
		{"item": shard},
		market_panel,
		blacksmith_panel
	)
	_assert_eq("blacksmith shard stage handled", handled_shard, true)
	_assert_eq("blacksmith shard staged as resource", str(blacksmith_panel.staged_resource.get("item_instance_id", "")), "shard1")
	_assert_eq(
		"unknown intent not handled",
		TownServiceBridgeScript.route_inventory_stage_intent("shop_sell_intent", {}, market_panel, blacksmith_panel),
		false
	)

	print("[gdtest] PASS: test_town_service_bridge (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
