# Unit tests for BotFacade's panel/action adapter helpers.
# Run via: godot --headless --path client --script res://tests/test_bot_facade.gd
extends SceneTree

const BotFacadeScript := preload("res://scripts/bot_facade.gd")

var _pass_count: int = 0
var _fail_count: int = 0


class FakePanel:
	extends Node

	var calls: Array = []

	func bot_click_buy_offer(offer_id: String, offer_kind: String, offer_index: int) -> void:
		calls.append(["buy", offer_id, offer_kind, offer_index])

	func bot_click_sell_item(item_def_id: String, rolled: Variant, bag_index: int) -> void:
		calls.append(["sell", item_def_id, rolled, bag_index])

	func bot_click_reroll() -> void:
		calls.append(["reroll"])

	func bot_drag_bag_to_stash(item_def_id: String, rolled: Variant, bag_index: int) -> void:
		calls.append(["bag_to_stash", item_def_id, rolled, bag_index])

	func bot_drag_stash_to_bag(stash_item_id: String, item_def_id: String, rolled: Variant, stash_index: int) -> void:
		calls.append(["stash_to_bag", stash_item_id, item_def_id, rolled, stash_index])

	func bot_click_deposit_gold(amount: int) -> void:
		calls.append(["deposit", amount])

	func bot_click_withdraw_gold(amount: int) -> void:
		calls.append(["withdraw", amount])

	func bot_click_respec() -> void:
		calls.append(["respec"])

	func bot_click_upgrade(stash_item_id: String, item_def_id: String, stash_index: int) -> void:
		calls.append(["upgrade", stash_item_id, item_def_id, stash_index])

	func bot_set_search_text(text: String) -> void:
		calls.append(["search", text])

	func bot_select_sort_mode(mode: String) -> void:
		calls.append(["sort", mode])

	func bot_set_publish_price(price_gold: int) -> void:
		calls.append(["price", price_gold])

	func bot_click_publish_stash_item(stash_item_id: String, item_def_id: String, rolled: Variant, stash_index: int) -> void:
		calls.append(["publish", stash_item_id, item_def_id, rolled, stash_index])

	func bot_click_purchase_listing(listing_id: String, item_def_id: String, price_gold: int, listing_index: int) -> void:
		calls.append(["purchase", listing_id, item_def_id, price_gold, listing_index])

	func bot_click_view_offers(listing_id: String, item_def_id: String, price_gold: int, listing_index: int) -> void:
		calls.append(["offers", listing_id, item_def_id, price_gold, listing_index])

	func bot_click_accept_offer(offer_id: String, offer_index: int) -> void:
		calls.append(["accept", offer_id, offer_index])

	func bot_click_stat_button(stat: String) -> void:
		calls.append(["stat", stat])

	func bot_click_skill_button(skill_id: String) -> void:
		calls.append(["skill", skill_id])


class FakeHotbar:
	extends Node

	var calls: Array = []

	func assign_slot(slot_index: int, item_instance_id: String) -> void:
		calls.append(["assign", slot_index, item_instance_id])

	func use_slot(slot_index: int) -> void:
		calls.append(["use", slot_index])


class FakeSkillBar:
	extends Node

	var calls: Array = []

	func set_skill_id(skill_id: String) -> void:
		calls.append(["skill_id", skill_id])

	func set_character_progression(progression: Dictionary) -> void:
		calls.append(["character_progression", progression.duplicate(true)])

	func set_skill_progression(progression: Dictionary) -> void:
		calls.append(["skill_progression", progression.duplicate(true)])

	func set_skill_cooldowns(cooldowns: Dictionary) -> void:
		calls.append(["cooldowns", cooldowns.duplicate(true)])

	func use_slot() -> void:
		calls.append(["use_slot"])


class FakeMain:
	extends Node

	var shop_panel: FakePanel
	var stash_panel: FakePanel
	var bishop_panel: FakePanel
	var blacksmith_panel: FakePanel
	var market_panel: FakePanel
	var consumable_bar: FakeHotbar
	var character_stats_panel: FakePanel
	var skills_panel: FakePanel
	var skill_bar: FakeSkillBar
	var right_click_skill_id: String = ""
	var character_progression: Dictionary = {"level": 3}
	var skill_progression: Dictionary = {"magic_bolt": 1}
	var skill_cooldowns: Dictionary = {"magic_bolt": 0}
	var _last_facing_direction: Vector2 = Vector2(0.0, 1.0)
	var sync_count: int = 0
	var cast_calls: Array = []
	var faced_direction: Vector2 = Vector2.ZERO
	var blocked: bool = false

	func _skill_rank(skill_id: String) -> int:
		return int(skill_progression.get(skill_id, 0))

	func _sync_skill_bindings_ui() -> void:
		sync_count += 1

	func _skill_cast_blocked(_skill_id: String = "") -> bool:
		return blocked

	func _face_direction(dir: Vector2) -> void:
		faced_direction = dir

	func _send_skill_cast_intent(skill_id: String, target_id: String = "", direction: Vector2 = Vector2.ZERO, use_nearest_fallback: bool = true) -> bool:
		cast_calls.append([skill_id, target_id, direction, use_nearest_fallback])
		return true


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_panel_adapters()
	_test_hotbar_and_skill_adapters()
	print("[gdtest] PASS: test_bot_facade (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_panel_adapters() -> void:
	var main := FakeMain.new()
	main.shop_panel = FakePanel.new()
	main.stash_panel = FakePanel.new()
	main.bishop_panel = FakePanel.new()
	main.blacksmith_panel = FakePanel.new()
	main.market_panel = FakePanel.new()
	main.character_stats_panel = FakePanel.new()
	main.skills_panel = FakePanel.new()

	BotFacadeScript.click_shop_sell_item(main, "rusty_sword", {"damage_min": 1}, 2)
	_assert_eq("shop sell routed", main.shop_panel.calls[0], ["sell", "rusty_sword", {"damage_min": 1}, 2])
	BotFacadeScript.click_shop_reroll(main)
	_assert_eq("shop reroll routed", main.shop_panel.calls[1], ["reroll"])
	BotFacadeScript.drag_stash_to_bag(main, "stash_1", "red_potion", null, 1)
	_assert_eq("stash to bag routed", main.stash_panel.calls[0], ["stash_to_bag", "stash_1", "red_potion", null, 1])
	BotFacadeScript.click_bishop_respec(main)
	_assert_eq("bishop routed", main.bishop_panel.calls[0], ["respec"])
	BotFacadeScript.click_blacksmith_upgrade(main, "stash_2", "bow", 3)
	_assert_eq("blacksmith routed", main.blacksmith_panel.calls[0], ["upgrade", "stash_2", "bow", 3])
	BotFacadeScript.set_market_publish_price(main, 77)
	_assert_eq("market price routed", main.market_panel.calls[0], ["price", 77])
	BotFacadeScript.click_market_accept_offer(main, "offer_1", 4)
	_assert_eq("market accept routed", main.market_panel.calls[1], ["accept", "offer_1", 4])
	BotFacadeScript.click_stat_button(main, "vit")
	_assert_eq("stat routed", main.character_stats_panel.calls[0], ["stat", "vit"])
	BotFacadeScript.click_skill_button(main, "magic_bolt")
	_assert_eq("skill routed", main.skills_panel.calls[0], ["skill", "magic_bolt"])


func _test_hotbar_and_skill_adapters() -> void:
	var main := FakeMain.new()
	main.consumable_bar = FakeHotbar.new()
	main.skill_bar = FakeSkillBar.new()

	BotFacadeScript.assign_consumable_hotbar(main, 2, "101")
	BotFacadeScript.use_consumable_hotbar(main, 2)
	_assert_eq("hotbar assign routed", main.consumable_bar.calls[0], ["assign", 2, "101"])
	_assert_eq("hotbar use routed", main.consumable_bar.calls[1], ["use", 2])

	BotFacadeScript.use_skill_bar(main, "magic_bolt", "", false)
	_assert_eq("skill selected", main.right_click_skill_id, "magic_bolt")
	_assert_eq("skill bindings synced", main.sync_count, 1)
	_assert_eq("skill bar used", main.skill_bar.calls.back(), ["use_slot"])

	BotFacadeScript.use_skill_bar(main, "magic_bolt", "1002", true)
	_assert_eq("direct skill cast routed", main.cast_calls[0][0], "magic_bolt")
	_assert_eq("direct skill target routed", main.cast_calls[0][1], "1002")

	BotFacadeScript.cast_skill_direction(main, "magic_bolt", {"x": 3.0, "y": 0.0})
	_assert_eq("direction faced", main.faced_direction, Vector2(1.0, 0.0))
	_assert_eq("direction cast skill", main.cast_calls[1][0], "magic_bolt")
	_assert_eq("direction cast vector", main.cast_calls[1][2], Vector2(1.0, 0.0))


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])
