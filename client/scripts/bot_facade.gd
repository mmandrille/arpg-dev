class_name BotFacade
extends RefCounted


static func click_shop_buy_offer(main, offer_id: String = "", offer_kind: String = "", offer_index: int = 0) -> void:
	var panel = _member(main, "shop_panel")
	if panel != null and panel.has_method("bot_click_buy_offer"):
		panel.bot_click_buy_offer(offer_id, offer_kind, offer_index)


static func click_shop_sell_item(main, item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	var panel = _member(main, "shop_panel")
	if panel != null and panel.has_method("bot_click_sell_item"):
		panel.bot_click_sell_item(item_def_id, rolled, bag_index)


static func click_shop_reroll(main) -> void:
	var panel = _member(main, "shop_panel")
	if panel != null and panel.has_method("bot_click_reroll"):
		panel.bot_click_reroll()


static func drag_bag_to_stash(main, item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	var panel = _member(main, "stash_panel")
	if panel != null and panel.has_method("bot_drag_bag_to_stash"):
		panel.bot_drag_bag_to_stash(item_def_id, rolled, bag_index)


static func drag_stash_to_bag(main, stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	var panel = _member(main, "stash_panel")
	if panel != null and panel.has_method("bot_drag_stash_to_bag"):
		panel.bot_drag_stash_to_bag(stash_item_id, item_def_id, rolled, stash_index)


static func click_stash_deposit_gold(main, amount: int = 1) -> void:
	var panel = _member(main, "stash_panel")
	if panel != null and panel.has_method("bot_click_deposit_gold"):
		panel.bot_click_deposit_gold(amount)


static func click_stash_withdraw_gold(main, amount: int = 1) -> void:
	var panel = _member(main, "stash_panel")
	if panel != null and panel.has_method("bot_click_withdraw_gold"):
		panel.bot_click_withdraw_gold(amount)


static func click_bishop_respec(main) -> void:
	var panel = _member(main, "bishop_panel")
	if panel != null and panel.has_method("bot_click_respec"):
		panel.bot_click_respec()


static func click_blacksmith_upgrade(main, stash_item_id: String = "", item_def_id: String = "", stash_index: int = 0) -> void:
	var panel = _member(main, "blacksmith_panel")
	if panel != null and panel.has_method("bot_click_upgrade"):
		panel.bot_click_upgrade(stash_item_id, item_def_id, stash_index)


static func click_mercenary_stance(main, stance: String) -> void:
	var panel = _member(main, "mercenary_panel")
	if panel != null and panel.has_method("bot_click_stance"):
		panel.bot_click_stance(stance)


static func set_stash_search(main, text: String) -> void:
	var panel = _member(main, "stash_panel")
	if panel != null and panel.has_method("bot_set_search_text"):
		panel.bot_set_search_text(text)


static func select_stash_sort(main, mode: String) -> void:
	var panel = _member(main, "stash_panel")
	if mode == "unique_chest_sets" and panel != null and panel.has_method("bot_select_unique_chest_tab"):
		panel.bot_select_unique_chest_tab("sets")
		return
	if panel != null and panel.has_method("bot_select_sort_mode"):
		panel.bot_select_sort_mode(mode)


static func set_market_publish_price(main, price_gold: int) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_set_publish_price"):
		panel.bot_set_publish_price(price_gold)


static func click_market_publish_item(main, stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_publish_stash_item"):
		panel.bot_click_publish_stash_item(stash_item_id, item_def_id, rolled, stash_index)


static func click_market_purchase_listing(main, listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_purchase_listing"):
		panel.bot_click_purchase_listing(listing_id, item_def_id, price_gold, listing_index)


static func click_market_view_offers(main, listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_view_offers"):
		panel.bot_click_view_offers(listing_id, item_def_id, price_gold, listing_index)


static func click_market_cancel_listing(main, listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_cancel_listing"):
		panel.bot_click_cancel_listing(listing_id, item_def_id, price_gold, listing_index)


static func click_market_offer_action(main, action: String, offer_id: String = "", offer_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_offer_action"):
		panel.bot_click_offer_action(action, offer_id, offer_index)


static func click_market_accept_offer(main, offer_id: String = "", offer_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_accept_offer"):
		panel.bot_click_accept_offer(offer_id, offer_index)
		return
	click_market_offer_action(main, "accept_offer", offer_id, offer_index)


static func click_market_cancel_offer(main, offer_id: String = "", offer_index: int = 0) -> void:
	var panel = _member(main, "market_panel")
	if panel != null and panel.has_method("bot_click_cancel_offer"):
		panel.bot_click_cancel_offer(offer_id, offer_index)
		return
	click_market_offer_action(main, "cancel_offer", offer_id, offer_index)


static func assign_consumable_hotbar(main, slot_index: int, item_instance_id: String) -> void:
	var bar = _member(main, "consumable_bar")
	if bar != null and bar.has_method("assign_slot"):
		bar.assign_slot(slot_index, item_instance_id)


static func use_consumable_hotbar(main, slot_index: int) -> void:
	var bar = _member(main, "consumable_bar")
	if bar != null and bar.has_method("use_slot"):
		bar.use_slot(slot_index)


static func click_stat_button(main, stat: String) -> void:
	var panel = _member(main, "character_stats_panel")
	if panel != null and panel.has_method("bot_click_stat_button"):
		panel.bot_click_stat_button(stat)


static func click_skill_button(main, skill_id: String = "") -> void:
	var panel = _member(main, "skills_panel")
	if panel != null and panel.has_method("bot_click_skill_button"):
		panel.bot_click_skill_button(skill_id)


static func use_skill_bar(main, skill_id: String = "", target_id: String = "", force_direct: bool = false) -> void:
	if skill_id == "":
		skill_id = SkillRulesLoader.first_skill_id()
	if force_direct or target_id != "":
		if main != null and main.has_method("_send_skill_cast_intent"):
			main._send_skill_cast_intent(skill_id, target_id)
		return
	if skill_id != "" and main != null and main.has_method("_skill_rank") and int(main._skill_rank(skill_id)) > 0:
		main.set("right_click_skill_id", skill_id)
		var bar = _member(main, "skill_bar")
		if bar != null:
			if bar.has_method("set_skill_id"):
				bar.set_skill_id(skill_id)
			if bar.has_method("set_character_progression"):
				bar.set_character_progression(_member(main, "character_progression"))
			if bar.has_method("set_skill_progression"):
				bar.set_skill_progression(_member(main, "skill_progression"))
			if bar.has_method("set_skill_cooldowns"):
				bar.set_skill_cooldowns(_member(main, "skill_cooldowns"))
	if main != null and main.has_method("_sync_skill_bindings_ui"):
		main._sync_skill_bindings_ui()
	var skill_bar = _member(main, "skill_bar")
	if skill_bar != null and skill_bar.has_method("use_slot"):
		skill_bar.use_slot()


static func cast_skill_direction(main, skill_id: String = "", direction: Dictionary = {}) -> void:
	if skill_id == "":
		skill_id = SkillRulesLoader.first_skill_id()
	if main == null:
		return
	if main.has_method("_skill_cast_blocked") and bool(main._skill_cast_blocked(skill_id)):
		return
	var last_facing: Vector2 = _member(main, "_last_facing_direction")
	var dir := Vector2(float(direction.get("x", last_facing.x)), float(direction.get("y", last_facing.y)))
	if dir.length_squared() <= 0.0001:
		dir = Vector2(1.0, 0.0)
	dir = dir.normalized()
	if main.has_method("_face_direction"):
		main._face_direction(dir)
	if main.has_method("_send_skill_cast_intent"):
		main._send_skill_cast_intent(skill_id, "", dir)


static func click_entity_id(main, target_id: String, buffered: bool = false) -> void:
	if main == null:
		return
	var client = _member(main, "client")
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or int(_member(main, "player_hp")) <= 0:
		return
	var entities: Dictionary = _member(main, "entities")
	if target_id == "" or not entities.has(target_id):
		return
	var rec: Dictionary = entities[target_id]
	if buffered:
		if str(rec.get("type", "")) != "monster":
			var attack_buffer = _member(main, "_attack_buffer")
			if attack_buffer != null and attack_buffer.has_method("clear"):
				attack_buffer.clear()
			return
		if main.has_method("_execute_click_pick"):
			main._execute_click_pick({"kind": "monster", "target_id": target_id})
		return
	var typ := str(rec.get("type", ""))
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	if typ == "interactable" and main.has_method("_interactable_should_approach_before_action") and main._interactable_should_approach_before_action(interactable_def_id):
		if main.has_method("_activate_or_approach_interactable"):
			main._activate_or_approach_interactable(target_id, rec)
		return
	if main.has_method("_send_action_intent"):
		main._send_action_intent(target_id)
	main.set("_attack_cooldown", main._basic_attack_cooldown_seconds() if typ == "monster" and main.has_method("_basic_attack_cooldown_seconds") else ClientConstants.SEND_INTERVAL)
	if typ == "monster" and main.has_method("_start_basic_attack_recovery_ui"):
		main._start_basic_attack_recovery_ui(float(_member(main, "_attack_cooldown")))


static func dispatch_action(main, intent_type: String, payload: Dictionary) -> void:
	if main == null:
		return
	var client = _member(main, "client")
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or int(_member(main, "player_hp")) <= 0:
		return
	if main.has_method("_movement_intent_starts_motion") and main._movement_intent_starts_motion(intent_type, payload):
		if main.has_method("_close_gameplay_panels_for_movement"):
			main._close_gameplay_panels_for_movement()
		if main.has_method("_mark_local_player_walking"):
			main._mark_local_player_walking()
	client.send(intent_type, int(_member(main, "last_server_tick")), payload)
	main.set("_attack_cooldown", ClientConstants.SEND_INTERVAL)


static func dispatch_inventory_intent(main, intent_type: String, payload: Dictionary) -> void:
	if main == null:
		return
	var client = _member(main, "client")
	if main.has_method("_input_locked") and main._input_locked():
		return
	if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN and int(_member(main, "player_hp")) > 0:
		client.send(intent_type, int(_member(main, "last_server_tick")), payload)


static func _member(owner, name: String):
	if owner == null:
		return null
	return owner.get(name)
