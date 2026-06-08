# Interactive client scene (ADR-0001 D3/D4): a thin renderer over the
# authoritative server. The client predicts the player's movement locally and
# reconciles to authoritative snapshots/deltas; the server owns all combat,
# loot, and inventory outcomes. Visuals are placeholder primitives (slice v1).
extends Node3D

const WaypointPanelConfig := preload("res://scripts/waypoint_panel_config.gd")

const NetClientScript := preload("res://scripts/net_client.gd")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")
const DamageNumberScript := preload("res://scripts/damage_number.gd")
const MonsterHealthBarScript := preload("res://scripts/monster_health_bar.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ConsumableBarScript := preload("res://scripts/consumable_bar.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const InputShadowOverlayScript := preload("res://scripts/input_shadow_overlay.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const MainMenuScript := preload("res://scripts/main_menu.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const PauseMenuScript := preload("res://scripts/pause_menu.gd")
const SustainedClickInputScript := preload("res://scripts/sustained_click_input.gd")
const CharacterScene := preload("res://scenes/character.tscn")
const MonsterDummyScene := preload("res://scenes/monster_dummy.tscn")
const MONSTER_EVENT_CLIPS := {
	"monster_damaged": "hit",
	"monster_killed": "death",
}
const PLAYER_EVENT_CLIPS := {
	"player_damaged": "hit",
	"player_killed": "death",
}
const PLAYER_START_HP := 10
const INTERACTABLE_ACTIVATION_RANGE := 1.5
const PLAYER_TINT := Color("#8fe8a7")
const REMOTE_PLAYER_TINT := Color("#202934")
const MONSTER_RARITY_TINTS := {
	"common": Color("#f2f2ec"),
	"champion": Color("#9fc7ff"),
	"rare": Color("#ff9b9b"),
	"unique": Color("#ffd978"),
}

var client: NetClient
var resolver: EquipmentVisualResolver
var player_anim: AnimationController
var player_reaction: ModelReactionController
var entities: Dictionary = {}        # id (String) -> {node:Node3D, controller:AnimationController|null, type:String}
var player_id: String = ""
var party: Array = []
var player_hp: int = PLAYER_START_HP
var player_max_hp: int = PLAYER_START_HP
var predicted_pos := Vector3.ZERO    # client-predicted player position
var reconciliation_delta: float = 0.0
var last_server_tick: int = 0
var inventory: Array = []
var equipped: Dictionary = {}
var hotbar_capacity: int = 2
var hotbar: Array = []
var character_progression: Dictionary = {}
var item_rules: Dictionary = {}
var item_presentations: Dictionary = {}
var dungeon_generation: Dictionary = {}
var loot_ids: Array = []
var monster_ids: Array = []
var interactable_ids: Array = []
var current_world_id: String = "vertical_slice"
var current_level: int = 0
var discovered_teleporters: Dictionary = {}
var pending_interactable_action: Dictionary = {}
var pending_waypoint_target_level: int = 0
var pending_waypoint_travel: bool = false
var ready_sent: bool = false
var item_to_equip: String = ""
var bot_mode: bool = false
var _bot_logged_snapshot: bool = false
var _bot_pending_events: Array = []
var autoplay_enabled: bool = false
var autoplay_phase: String = "idle"
var autoplay_timer: float = 0.0
var autoplay_attack_cooldown: float = 0.0
var autoplay_move_sent: bool = false
var autoplay_pickup_sent: bool = false
var autoplay_equip_sent: bool = false
var autoplay_step_delay: float = 0.35
var visual_replay_enabled: bool = false
var visual_replay_manifest_path: String = ""
var visual_replay_scenarios: Array = []
var visual_replay_index: int = -1
var visual_replay_envelopes: Array = []
var visual_replay_envelope_index: int = 0
var visual_replay_timer: float = 0.0
var visual_replay_title: String = ""
var visual_replay_debug_token: String = ""
var visual_replay_dev_token: String = ""
var visual_replay_exit_on_complete: bool = false
var visual_replay_exit_requested: bool = false
var waypoint_panel: PanelContainer
var waypoint_rows: VBoxContainer
var visual_replay_exit_timer: float = 0.0
var visual_replay_show_inventory: bool = false
var client_settings: ClientSettings
var menu_layer: CanvasLayer
var main_menu: MainMenu
var character_panel: CharacterSelectPanel
var settings_panel: SettingsPanel
var pause_menu: PauseMenu
var gameplay_active: bool = false
var settings_return_target: String = "main"

const INVENTORY_REPLAY_EVENT_HINTS := {
	"item_picked_up": "Pickup",
	"item_equipped": "Equip (double-click / drag)",
	"item_unequipped": "Unequip (drag to bag)",
	"item_dropped": "Drop (drag outside panel)",
}

# Slice v2 scene graph (spec §5.1): the local player is a humanoid under a
# PlayerAnchor that follows authoritative position; monsters/loot live under
# Entities. These are defined in main.tscn and cached on ready.
var player_anchor: Node3D
var character_visual: Node3D
var entities_root: Node3D
var damage_numbers_layer: CanvasLayer
var health_bars_layer: CanvasLayer
var monster_health_bars: Dictionary = {} # id (String) -> MonsterHealthBar
var walls_root: Node3D
var inventory_panel: InventoryPanel
var consumable_bar: ConsumableBar
var character_stats_panel: CharacterStatsPanel
var input_shadow: InputShadowOverlay
var _health_bar: PlayerHealthBar

var _send_cooldown: float = 0.0
var _attack_cooldown: float = 0.0
var _sustained_click: SustainedClickInput = SustainedClickInputScript.new()
var _debug_label: Label
var _level_label: Label
var _camera: Camera3D

const SEND_INTERVAL := 0.1
const PLAYER_SPEED := 2.8
const CAMERA_ZOOM_DEFAULT := 20.0
const CAMERA_ZOOM_STEP := 1.5
const CAMERA_ZOOM_MIN := 8.0
const CAMERA_ZOOM_MAX := 40.0
const CAMERA_FOLLOW_OFFSET := Vector3(9.0, 20.0, 15.0)
const PROJECTILE_LERP_SECONDS := 0.10


func _ready() -> void:
	player_anchor = $World/PlayerAnchor
	character_visual = $World/PlayerAnchor/CharacterVisual
	entities_root = $Entities
	# Mount-root is injected (spec §4.8): the resolver finds the named socket
	# within CharacterVisual, never via an absolute scene path.
	resolver = EquipmentResolverScript.new(character_visual)
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)
	_apply_model_tint(character_visual, PLAYER_TINT)
	player_reaction = ModelReactionControllerScript.new(character_visual, PLAYER_TINT)
	_build_scene()
	client_settings = ClientSettingsScript.new()
	client_settings.load()
	client_settings.apply()
	_sync_settings_panel()
	_load_item_rules()
	_load_item_presentations()
	_load_dungeon_generation()
	var base_url := _env("ARPG_BASE_URL", "http://localhost:8080")
	var dev_token := _env("ARPG_DEV_TOKEN", "local-dev-token")

	client = NetClientScript.new(base_url)
	var bot_client_run := _truthy_env("ARPG_BOT_CLIENT")
	var bot_menu_run := bot_client_run and _bot_uses_menu()
	if not client.login(_env("ARPG_EMAIL", "client@example.test"), dev_token):
		if bot_client_run:
			printerr("[bot-client] login failed base_url=%s" % base_url)
		_debug("login failed")
		return
	if bot_client_run:
		print("[bot-client] login ok")
	visual_replay_manifest_path = _env("ARPG_VISUAL_REPLAY_MANIFEST", "")
	visual_replay_enabled = visual_replay_manifest_path != ""
	if visual_replay_enabled:
		visual_replay_debug_token = _env("ARPG_DEBUG_TOKEN", "local-debug-token")
		visual_replay_dev_token = dev_token
		visual_replay_exit_on_complete = _truthy_text(_env("ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE", "1"))
		autoplay_step_delay = maxf(0.05, float(_env("ARPG_AUTOPLAY_STEP_DELAY", "0.35")))
		if not _load_visual_replay_manifest(visual_replay_manifest_path):
			_debug("visual replay manifest failed: %s" % visual_replay_manifest_path)
			return
		_debug("visual replay playlist loaded: %d scenario(s)" % visual_replay_scenarios.size())
		_start_next_visual_replay()
		return
	if bot_menu_run:
		bot_mode = true
		_show_main_menu()
		_mount_bot_controller()
		return
	var resume_session_id := _env("ARPG_SESSION_ID", "")
	var requested_world_id := _env("ARPG_WORLD_ID", "")
	var requested_seed := _env("ARPG_SEED", "")
	if requested_world_id == "" and not bot_client_run:
		requested_world_id = "dungeon_levels"
	if bot_client_run or resume_session_id != "" or _truthy_env("ARPG_AUTOSTART"):
		if not _start_automation_session(resume_session_id, requested_world_id, requested_seed, bot_client_run):
			return
		if bot_client_run:
			_mount_bot_controller()
		return
	_show_main_menu()


func _bot_uses_menu() -> bool:
	if _truthy_env("ARPG_BOT_MENU"):
		return true
	var scenario_path := _env("ARPG_BOT_SCENARIO", "")
	return scenario_path.get_file().begins_with("08_main_menu_flow")


func _mount_bot_controller() -> void:
	if input_shadow != null and DisplayServer.get_name() != "headless":
		input_shadow.set_active(true)
	else:
		Input.set_mouse_mode(Input.MOUSE_MODE_HIDDEN)
	var bot := preload("res://scripts/bot_controller.gd").new()
	add_child(bot)


func _start_automation_session(resume_session_id: String, requested_world_id: String, requested_seed: String, bot_client_run: bool) -> bool:
	if not client.create_session(resume_session_id, requested_world_id, "", requested_seed):
		if bot_client_run:
			printerr("[bot-client] session failed world_id=%s resume=%s" % [requested_world_id, resume_session_id])
		_debug("session failed")
		return false
	if bot_client_run:
		print("[bot-client] session ok id=%s world=%s" % [client.session_id, client.world_id])
	bot_mode = bot_client_run
	_begin_gameplay_connection(_truthy_env("ARPG_AUTOPLAY"))
	if bot_client_run:
		print("[bot-client] ws connect requested session=%s" % client.session_id)
	return true


func _begin_gameplay_connection(enable_autoplay: bool = false) -> void:
	_hide_all_menus()
	gameplay_active = true
	current_world_id = client.world_id
	_render_world_walls(client.world_id)
	autoplay_enabled = enable_autoplay
	if autoplay_enabled:
		autoplay_phase = "move"
		autoplay_step_delay = maxf(0.05, float(_env("ARPG_AUTOPLAY_STEP_DELAY", "0.35")))
		_debug("visual bot enabled for session %s" % client.session_id)
	predicted_pos = Vector3.ZERO
	ready_sent = false
	client.connect_ws()
	_debug("connecting session %s" % client.session_id)


func _show_main_menu() -> void:
	_hide_all_menus()
	gameplay_active = false
	if main_menu != null:
		main_menu.show_menu()


func _hide_all_menus() -> void:
	if main_menu != null:
		main_menu.visible = false
	if character_panel != null:
		character_panel.hide_panel()
	if settings_panel != null:
		settings_panel.hide_panel()
	if pause_menu != null:
		pause_menu.hide_pause()
	if character_stats_panel != null:
		character_stats_panel.hide_display()


func _on_continue_pressed() -> void:
	var characters := client.list_characters()
	if character_panel != null:
		main_menu.visible = false
		character_panel.show_continue(characters)


func _on_new_game_pressed() -> void:
	if character_panel != null:
		main_menu.visible = false
		character_panel.show_new_game()


func _on_character_create_requested(name: String) -> void:
	var character := client.create_character(name)
	if character.is_empty():
		character_panel.set_error("Could not create character")
		return
	_start_character_session(str(character.get("character_id", "")))


func _on_character_delete_requested(character_id: String) -> void:
	if not client.delete_character(character_id):
		if character_panel != null:
			character_panel.set_error("Could not delete character")
		return
	if character_panel != null:
		character_panel.show_continue(client.list_characters())


func _start_character_session(character_id: String) -> void:
	if character_id == "":
		if character_panel != null:
			character_panel.set_error("Could not start character")
		return
	_teardown_gameplay_state(false)
	if not client.create_session("", "dungeon_levels", character_id):
		if character_panel != null:
			character_panel.set_error("Could not start session")
		return
	bot_mode = false
	_begin_gameplay_connection(false)


func _on_settings_from_main() -> void:
	settings_return_target = "main"
	main_menu.visible = false
	if settings_panel != null:
		settings_panel.show_settings(ClientSettingsScript.size_label(client_settings.window_size), client_settings.floating_combat_text)


func _on_settings_from_pause() -> void:
	settings_return_target = "pause"
	if pause_menu != null:
		pause_menu.hide_pause()
	if settings_panel != null:
		settings_panel.show_settings(ClientSettingsScript.size_label(client_settings.window_size), client_settings.floating_combat_text)


func _on_settings_back() -> void:
	if settings_panel != null:
		settings_panel.hide_panel()
	if settings_return_target == "pause" and pause_menu != null:
		pause_menu.show_pause()
	elif main_menu != null:
		main_menu.show_menu()


func _on_window_size_selected(label: String) -> void:
	client_settings.set_window_size_label(label)
	_sync_settings_panel()


func _on_floating_combat_text_toggled(enabled: bool) -> void:
	if client_settings == null:
		return
	client_settings.set_floating_combat_text(enabled)
	_sync_settings_panel()


func _sync_settings_panel() -> void:
	if settings_panel != null and client_settings != null:
		settings_panel.set_selected_size_label(ClientSettingsScript.size_label(client_settings.window_size))
		settings_panel.set_floating_combat_text_enabled(client_settings.floating_combat_text)


func _show_pause_menu() -> void:
	if gameplay_active and pause_menu != null:
		pause_menu.show_pause()


func _resume_from_pause() -> void:
	if pause_menu != null:
		pause_menu.hide_pause()


func _return_to_main_menu() -> void:
	if client != null:
		if client.session_id != "":
			client.end_session()
		client.close()
	_teardown_gameplay_state(true)
	_show_main_menu()


func _exit_game() -> void:
	if client != null:
		if gameplay_active and client.session_id != "":
			client.end_session()
		client.close()
	get_tree().quit(0)


func _teardown_gameplay_state(clear_session: bool) -> void:
	gameplay_active = false
	ready_sent = false
	player_id = ""
	party = []
	player_hp = PLAYER_START_HP
	player_max_hp = PLAYER_START_HP
	predicted_pos = Vector3.ZERO
	reconciliation_delta = 0.0
	last_server_tick = 0
	inventory = []
	equipped = {}
	character_progression = {}
	loot_ids.clear()
	monster_ids.clear()
	interactable_ids.clear()
	discovered_teleporters.clear()
	pending_interactable_action.clear()
	pending_waypoint_target_level = 0
	pending_waypoint_travel = false
	autoplay_enabled = false
	autoplay_phase = "idle"
	autoplay_timer = 0.0
	autoplay_attack_cooldown = 0.0
	autoplay_move_sent = false
	autoplay_pickup_sent = false
	autoplay_equip_sent = false
	_bot_pending_events.clear()
	_bot_logged_snapshot = false
	_clear_level_entities()
	if walls_root != null:
		for child in walls_root.get_children():
			child.queue_free()
	if resolver != null:
		resolver.apply_snapshot({"inventory": [], "equipped": {}})
	_refresh_inventory_ui()
	if _health_bar != null:
		_health_bar.update_hp(player_hp, player_max_hp)
	if character_stats_panel != null:
		character_stats_panel.hide_display()
	_refresh_progression_ui()
	_hide_waypoint_panel()
	if player_anchor != null:
		player_anchor.position = Vector3.ZERO
	if clear_session and client != null:
		client.session_id = ""
		client.seed = ""
		client.world_id = ""
		client.ws_url = ""


func _process(delta: float) -> void:
	if client == null:
		return

	var state := client.ready_state()
	if state == WebSocketPeer.STATE_OPEN and not ready_sent:
		if bot_mode:
			print("[bot-client] ws open, sending client_ready tick=%d" % last_server_tick)
		client.send("client_ready", last_server_tick, {"client_version": "godot", "last_seen_tick": last_server_tick})
		ready_sent = true

	if gameplay_active or visual_replay_enabled:
		for env in client.poll():
			_handle_message(env)
	_sync_progression_interactivity()
	_try_complete_pending_interactable_action()
	_try_complete_pending_waypoint_travel()

	if visual_replay_enabled:
		_handle_visual_replay(delta)
	elif autoplay_enabled:
		_handle_autoplay(delta)
	else:
		_handle_input(delta)
	_sync_waypoint_panel_reach()
	if player_anim != null:
		var moving := client.ready_state() == WebSocketPeer.STATE_OPEN \
			and player_hp > 0 \
			and not _user_input_blocked() \
			and (Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_A) \
			or Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_D))
		player_anim.set_locomotion(moving)
	if not _user_input_blocked():
		_update_facing_toward_mouse()
	_update_debug()


# --- message handling -------------------------------------------------------

func _handle_message(env: Dictionary) -> void:
	last_server_tick = max(last_server_tick, int(env.get("tick", 0)))
	match env.get("type", ""):
		"session_snapshot":
			_apply_snapshot(env["payload"])
		"state_delta":
			_apply_delta(env["payload"])
		"intent_rejected":
			pending_interactable_action.clear()
			pending_waypoint_travel = false
			_debug("rejected: %s" % env["payload"].get("reason", "?"))
		"error":
			_debug("error: %s" % env["payload"].get("message", "?"))


func _snapshot_local_player_id(p: Dictionary) -> String:
	var explicit := str(p.get("local_player_id", ""))
	if explicit != "":
		return explicit
	for e in p.get("entities", []):
		if str(e.get("type", "")) == "player":
			return str(e.get("id", ""))
	return player_id


func _event_subject_entity_id(ev: Dictionary) -> String:
	var event_type := str(ev.get("event_type", ""))
	if event_type in ["monster_damaged", "monster_killed", "player_damaged", "player_killed"]:
		var target_id := str(ev.get("target_entity_id", ""))
		if target_id != "":
			return target_id
	return str(ev.get("entity_id", ""))


func _apply_snapshot(p: Dictionary) -> void:
	current_level = int(p.get("current_level", 0))
	var local_id := _snapshot_local_player_id(p)
	if local_id != "":
		player_id = local_id
	party = p.get("party", [])
	pending_interactable_action.clear()
	pending_waypoint_travel = false
	_apply_teleporter_snapshot(p.get("discovered_teleporters", []))
	_clear_level_entities()
	_render_world_walls(current_world_id)
	_update_level_hud()
	_refresh_waypoint_panel()
	# (player is the PlayerAnchor/CharacterVisual, not a per-snapshot entity node)
	for e in p.get("entities", []):
		_upsert_entity(e)
	inventory = p.get("inventory", [])
	equipped = p.get("equipped", {})
	hotbar_capacity = int(p.get("hotbar_capacity", 2))
	hotbar = p.get("hotbar", [])
	character_progression = p.get("character_progression", {})
	if resolver != null:
		resolver.apply_snapshot(p)
	_refresh_inventory_ui()
	_refresh_progression_ui()
	_reconcile_player()
	if bot_mode and not _bot_logged_snapshot:
		_bot_logged_snapshot = true
		print("[bot-client] snapshot applied entities=%d monsters=%d loot=%d hp=%d" % [
			entities.size(), monster_ids.size(), loot_ids.size(), player_hp
		])


func _apply_delta(p: Dictionary) -> void:
	for ev in p.get("events", []):
		if str(ev.get("event_type", "")) == "level_changed":
			current_level = int(ev.get("to_level", current_level))
			pending_interactable_action.clear()
			pending_waypoint_travel = false
			_clear_level_entities()
			_render_world_walls(current_world_id)
			_update_level_hud()
			_hide_waypoint_panel()
	var changes: Array = p.get("changes", [])
	for c in changes:
		match c.get("op", ""):
			"entity_spawn", "entity_update":
				_upsert_entity(c["entity"])
			"entity_remove":
				_remove_entity(c["entity_id"])
			"inventory_add":
				inventory.append(c["item"])
				if resolver != null:
					resolver.ingest_inventory_item(c["item"])
			"inventory_update":
				_update_inventory_item(c["item"])
				if resolver != null:
					resolver.ingest_inventory_item(c["item"])
			"inventory_remove":
				_remove_inventory_item(str(c["item_instance_id"]))
			"equipped_update":
				equipped[c["slot"]] = c.get("item_instance_id")
				if resolver != null:
					resolver.apply_equipped_update(c["slot"], c.get("item_instance_id"))
				if c.has("hotbar_capacity"):
					hotbar_capacity = int(c.get("hotbar_capacity", hotbar_capacity))
					if consumable_bar != null:
						consumable_bar.set_hotbar_state(hotbar_capacity, hotbar)
			"hotbar_update":
				_apply_hotbar_update(int(c.get("slot_index", -1)), c.get("item_instance_id"))
			"teleporter_discovery_update":
				var discovered_level := int(c.get("level", 0))
				var discovered := bool(c.get("discovered", false))
				discovered_teleporters[discovered_level] = discovered
				_refresh_waypoint_panel()
				if discovered and discovered_level == current_level:
					_show_waypoint_panel()
			"character_progression_update":
				character_progression = c.get("character_progression", {})
				_refresh_progression_ui()
			_:
				pass
	_refresh_inventory_ui()
	for ev in p.get("events", []):
		var eid := _event_subject_entity_id(ev)
		var event_type := str(ev.get("event_type", ""))
		if visual_replay_enabled and inventory_panel != null:
			var hint: Variant = INVENTORY_REPLAY_EVENT_HINTS.get(event_type, null)
			if hint != null:
				inventory_panel.show_gesture_hint(str(hint))
		if eid == player_id:
			if event_type == "player_healed":
				_show_damage_number(eid, Color(0.3, 1.0, 0.45), ev.get("heal", null), "+", 1.0)
				if _health_bar != null:
					_health_bar.update_hp(player_hp, player_max_hp, true)
				continue
			if event_type == "player_damaged":
				_show_combat_text_for_event(eid, ev, Color(1.0, 0.32, 0.2))
				_play_entity_reaction(eid, ev, "hit")
				if _health_bar != null:
					_health_bar.update_hp(player_hp, player_max_hp)
			if event_type == "player_killed":
				_play_entity_reaction(eid, ev, "death")
			if event_type == "attack_missed":
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			var player_clip = PLAYER_EVENT_CLIPS.get(event_type, null)
			if player_clip == null or player_anim == null:
				continue
			if player_clip == "death":
				player_anim.enter_terminal("death")
			else:
				player_anim.play_one_shot(player_clip)
			continue
		if PLAYER_EVENT_CLIPS.has(event_type) and entities.has(eid):
			if event_type == "player_damaged":
				_show_combat_text_for_event(eid, ev, Color(1.0, 0.32, 0.2))
				_play_entity_reaction(eid, ev, "hit")
			if event_type == "player_killed":
				var remote_dead: Dictionary = entities[eid]
				remote_dead["hp"] = 0
				_play_entity_reaction(eid, ev, "death")
			var remote_player_clip = PLAYER_EVENT_CLIPS.get(event_type, null)
			var remote_ctrl = entities[eid].get("controller", null)
			if remote_ctrl != null:
				if remote_player_clip == "death":
					remote_ctrl.enter_terminal("death")
				else:
					remote_ctrl.play_one_shot(remote_player_clip)
			continue
		if event_type == "interactable_activated" and entities.has(eid):
			_set_interactable_state(eid, entities[eid], "open")
			continue
		var clip = MONSTER_EVENT_CLIPS.get(event_type, null)
		if clip == null:
			if event_type == "attack_missed":
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			continue
		if event_type == "monster_damaged" or event_type == "monster_killed":
			_show_combat_text_for_event(eid, ev, Color(1.0, 0.92, 0.25))
		if event_type == "monster_damaged":
			_play_entity_reaction(eid, ev, "hit")
		if event_type == "monster_killed":
			_remove_monster_health_bar(eid)
			if entities.has(eid):
				var dead_rec: Dictionary = entities[eid]
				dead_rec["hp"] = 0
				_set_pickable(dead_rec["node"] as Node3D, false)
				_play_entity_reaction(eid, ev, "death")
		if autoplay_enabled and event_type == "monster_killed":
			autoplay_phase = "pickup"
			autoplay_timer = autoplay_step_delay
		if not entities.has(eid):
			continue
		var ctrl = entities[eid]["controller"]
		if ctrl == null:
			continue
		if clip == "death":
			ctrl.enter_terminal("death")
		else:
			ctrl.play_one_shot(clip)
	if bot_mode:
		for ev in p.get("events", []):
			_bot_pending_events.append(ev)
	_reconcile_player()


func _upsert_entity(e: Dictionary) -> void:
	var id := str(e["id"])
	var pos: Dictionary = e["position"]
	var server_pos := Vector3(pos["x"], 0.0, pos["y"])
	if e["type"] == "player" and (id == player_id or player_id == ""):
		# The player is the humanoid under PlayerAnchor, not an entity-dict node.
		player_id = id
		if e.has("hp"):
			player_hp = int(e["hp"])
			if e.has("max_hp"):
				player_max_hp = int(e["max_hp"])
			if _health_bar != null:
				_health_bar.update_hp(player_hp, player_max_hp)
			if player_hp <= 0 and player_anim != null:
				player_anim.enter_terminal("death")
			if player_hp <= 0 and player_reaction != null:
				player_reaction.enter_death()
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		player_anchor.position = server_pos
		return
	var rec: Dictionary
	var is_new := false
	if entities.has(id):
		rec = entities[id]
	else:
		is_new = true
		var node := _make_entity_node(e)
		entities_root.add_child(node)
		var controller: AnimationController = null
		if e["type"] == "monster" or e["type"] == "player":
			var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
			if ap != null:
				controller = AnimationControllerScript.new(ap)
			else:
				push_warning("[main] %s %s has no AnimationPlayer" % [str(e["type"]), id])
		var base_tint := _entity_base_tint(e)
		var reaction = null
		if e["type"] == "monster" or e["type"] == "player":
			reaction = ModelReactionControllerScript.new(node, base_tint)
		rec = {"node": node, "controller": controller, "reaction": reaction, "type": str(e["type"]), "base_tint": base_tint.to_html(false)}
		if e.has("item_def_id"):
			rec["item_def_id"] = str(e["item_def_id"])
		if e.has("monster_def_id"):
			rec["monster_def_id"] = str(e["monster_def_id"])
		for key in ["item_template_id", "display_name", "rarity", "rolled_stats", "requirements", "effect_ids", "character_id"]:
			if e.has(key):
				rec[key] = e[key]
		if e.has("interactable_def_id"):
			rec["interactable_def_id"] = str(e["interactable_def_id"])
		entities[id] = rec
		if e["type"] != "projectile" and e["type"] != "player":
			_attach_pick_collider(node, id, str(e["type"]))
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
		if e["type"] == "monster" and not monster_ids.has(id):
			monster_ids.append(id)
		if e["type"] == "interactable" and not interactable_ids.has(id):
			interactable_ids.append(id)
	if rec["type"] == "projectile":
		if is_new:
			(rec["node"] as Node3D).position = server_pos
			rec["last_server_pos"] = server_pos
			return
		_move_projectile_node(rec, server_pos)
	else:
		var node := rec["node"] as Node3D
		var prev_pos := node.position
		node.position = server_pos
		if rec["type"] == "monster" and rec["controller"] != null and not is_new:
			var hp_val := int(e.get("hp", rec.get("hp", 1)))
			var moved := prev_pos.distance_to(server_pos) > 0.001
			rec["controller"].set_locomotion(moved and hp_val > 0)
		if rec["type"] == "player":
			rec["hp"] = int(e.get("hp", rec.get("hp", PLAYER_START_HP)))
			rec["max_hp"] = int(e.get("max_hp", rec.get("max_hp", PLAYER_START_HP)))
			if int(rec["hp"]) <= 0:
				_enter_entity_terminal_death(id, rec)
	if rec["type"] == "interactable":
		var state := str(e.get("state", rec.get("state", "closed")))
		_set_interactable_state(id, rec, state)
	# Resume/snapshot consistency: a monster already dead in the snapshot enters
	# the terminal death pose without waiting for an event (spec §5.4).
	if rec["type"] == "monster" and rec["controller"] != null:
		var hp = e.get("hp", null)
		var max_hp = e.get("max_hp", null)
		if hp != null and max_hp != null:
			rec["hp"] = int(hp)
			_upsert_monster_health_bar(id, rec["node"] as Node3D, int(hp), int(max_hp))
		if hp != null and int(hp) <= 0:
			_set_pickable(rec["node"] as Node3D, false)
			_enter_entity_terminal_death(id, rec)


func _remove_entity(id: String) -> void:
	if str(pending_interactable_action.get("target_id", "")) == id:
		pending_interactable_action.clear()
	if id == player_id:
		return
	if entities.has(id):
		var rec: Dictionary = entities[id]
		if rec.has("move_tween"):
			var tween = rec["move_tween"]
			if is_instance_valid(tween):
				tween.kill()
		(entities[id]["node"] as Node3D).queue_free()
		entities.erase(id)
	_remove_monster_health_bar(id)
	loot_ids.erase(id)
	monster_ids.erase(id)
	interactable_ids.erase(id)


func _clear_level_entities() -> void:
	for id in entities.keys():
		(entities[id]["node"] as Node3D).queue_free()
	entities.clear()
	for id in monster_health_bars.keys():
		var bar = monster_health_bars[id]
		if is_instance_valid(bar):
			bar.queue_free()
	monster_health_bars.clear()
	loot_ids.clear()
	monster_ids.clear()
	interactable_ids.clear()


func _update_inventory_item(item: Dictionary) -> void:
	for i in range(inventory.size()):
		if inventory[i]["item_instance_id"] == item["item_instance_id"]:
			inventory[i] = item
			return
	inventory.append(item)


func _remove_inventory_item(item_instance_id: String) -> void:
	for i in range(inventory.size() - 1, -1, -1):
		if str(inventory[i].get("item_instance_id", "")) == item_instance_id:
			inventory.remove_at(i)


func _apply_hotbar_update(slot_index: int, item_instance_id) -> void:
	if slot_index < 0 or slot_index >= 10:
		return
	while hotbar.size() < 10:
		hotbar.append({"slot_index": hotbar.size(), "item_instance_id": null})
	hotbar[slot_index] = {"slot_index": slot_index, "item_instance_id": item_instance_id}
	if consumable_bar != null:
		consumable_bar.apply_hotbar_update(slot_index, item_instance_id)


func _refresh_inventory_ui() -> void:
	if inventory_panel != null:
		inventory_panel.set_inventory_state(inventory, equipped)
	if consumable_bar != null:
		consumable_bar.set_inventory_state(inventory)
		consumable_bar.set_hotbar_state(hotbar_capacity, hotbar)


func _refresh_inventory_panel() -> void:
	_refresh_inventory_ui()
	if visual_replay_enabled:
		_sync_inventory_replay_display()


func _reconcile_player() -> void:
	if player_anchor != null:
		player_anchor.position = predicted_pos
		_sync_camera_to_player()


func _sync_camera_to_player() -> void:
	if _camera == null or player_anchor == null:
		return
	var target := player_anchor.global_position
	_camera.global_position = target + CAMERA_FOLLOW_OFFSET
	_camera.look_at(target, Vector3.UP)


func _show_combat_text_for_event(entity_id: String, ev: Dictionary, default_color: Color) -> void:
	var outcome := str(ev.get("outcome", ""))
	var damage = ev.get("damage", null)
	if outcome == "miss":
		_show_damage_number(entity_id, Color(0.82, 0.86, 0.92), null, "", 0.0, "miss", "MISS")
		return
	if outcome == "block":
		_show_damage_number(entity_id, Color(0.35, 0.78, 1.0), null, "", 0.0, "block", "BLOCK")
		return
	if outcome == "crit" or bool(ev.get("critical", false)):
		var crit_damage := 0 if damage == null else int(damage)
		_show_damage_number(entity_id, Color(1.0, 0.58, 0.22), crit_damage, "", 0.0, "crit", "%d!" % crit_damage)
		return
	_show_damage_number(entity_id, default_color, damage)


func _play_entity_reaction(entity_id: String, ev: Dictionary, reaction_name: String) -> void:
	var reaction = _reaction_for_entity(entity_id)
	if reaction == null:
		return
	var source_pos := _source_position_for_event(ev)
	var fallback := _fallback_reaction_direction(entity_id)
	if reaction_name == "death":
		reaction.enter_death(source_pos, fallback)
	else:
		reaction.play_hit(source_pos, fallback)


func _reaction_for_entity(entity_id: String):
	if entity_id == player_id:
		return player_reaction
	if entities.has(entity_id):
		var rec: Dictionary = entities[entity_id]
		return rec.get("reaction", null)
	return null


func _source_position_for_event(ev: Dictionary) -> Vector3:
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "":
		return ModelReactionControllerScript.UNRESOLVED_SOURCE
	if source_id == player_id and player_anchor != null:
		return _node_world_or_local_position(player_anchor)
	if entities.has(source_id):
		var rec: Dictionary = entities[source_id]
		var node := rec.get("node", null) as Node3D
		if node != null:
			return _node_world_or_local_position(node)
	return ModelReactionControllerScript.UNRESOLVED_SOURCE


func _fallback_reaction_direction(entity_id: String) -> Vector3:
	var target := _entity_world_position(entity_id)
	if target != ModelReactionControllerScript.UNRESOLVED_SOURCE and player_anchor != null:
		var direction := target - _node_world_or_local_position(player_anchor)
		direction.y = 0.0
		if direction.length() > 0.001:
			return direction.normalized()
	return Vector3.BACK


func _entity_world_position(entity_id: String) -> Vector3:
	if entity_id == player_id and player_anchor != null:
		return _node_world_or_local_position(player_anchor)
	if entities.has(entity_id):
		var rec: Dictionary = entities[entity_id]
		var node := rec.get("node", null) as Node3D
		if node != null:
			return _node_world_or_local_position(node)
	return ModelReactionControllerScript.UNRESOLVED_SOURCE


func _node_world_or_local_position(node: Node3D) -> Vector3:
	if node.is_inside_tree():
		return node.global_position
	return node.position


func _enter_entity_terminal_death(entity_id: String, rec: Dictionary) -> void:
	var ctrl = rec.get("controller", null)
	if ctrl != null:
		ctrl.enter_terminal("death")
	var reaction = rec.get("reaction", null)
	if reaction != null:
		reaction.enter_death()


func _show_damage_number(entity_id: String, color: Color, event_damage = null, prefix: String = "", side_override: float = 0.0, variant: String = "normal", text_override: String = "") -> void:
	if damage_numbers_layer == null or _camera == null:
		return
	if client_settings != null and not client_settings.floating_combat_text:
		return
	if event_damage == null and text_override == "":
		return
	var amount := 0 if event_damage == null else int(event_damage)
	var target: Node3D = null
	var world_position := Vector3.ZERO
	if entity_id == player_id:
		target = player_anchor
		world_position = player_anchor.global_position
	elif entities.has(entity_id):
		target = entities[entity_id]["node"] as Node3D
		world_position = target.global_position
	else:
		return
	var pop := DamageNumberScript.new() as DamageNumber
	damage_numbers_layer.add_child(pop)
	var side := side_override if side_override != 0.0 else (-1.0 if entity_id == player_id else 1.0)
	pop.setup(_camera, target, world_position, amount, color, side, prefix, variant, text_override)


func _remove_monster_health_bar(entity_id: String) -> void:
	if not monster_health_bars.has(entity_id):
		return
	var bar = monster_health_bars[entity_id]
	if is_instance_valid(bar):
		bar.queue_free()
	monster_health_bars.erase(entity_id)


func _upsert_monster_health_bar(entity_id: String, target: Node3D, hp: int, max_hp: int) -> void:
	if hp <= 0:
		_remove_monster_health_bar(entity_id)
		return
	if health_bars_layer == null or _camera == null or target == null:
		return
	if monster_health_bars.has(entity_id):
		(monster_health_bars[entity_id] as MonsterHealthBar).update_hp(hp, max_hp)
		return
	var bar := MonsterHealthBarScript.new() as MonsterHealthBar
	health_bars_layer.add_child(bar)
	bar.setup(_camera, target, hp, max_hp)
	monster_health_bars[entity_id] = bar


# --- input + prediction -----------------------------------------------------

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT and not event.pressed:
		_sustained_click.clear()
	if event is InputEventKey and event.pressed and not event.echo and _is_escape_key(event):
		_handle_escape()
		get_viewport().set_input_as_handled()
		return
	if _input_locked():
		return
	if bot_mode and not (event is InputEventKey):
		return
	if event is InputEventKey and event.pressed and not event.echo:
		var hotbar_slot := _hotbar_slot_for_key(event)
		if hotbar_slot >= 0:
			if consumable_bar != null:
				consumable_bar.use_slot(hotbar_slot)
			get_viewport().set_input_as_handled()
			return
		if _is_inventory_key(event):
			if inventory_panel != null:
				inventory_panel.toggle()
			get_viewport().set_input_as_handled()
			return
		if _is_character_stats_key(event):
			if character_stats_panel != null:
				character_stats_panel.toggle()
				_refresh_progression_ui()
			get_viewport().set_input_as_handled()
			return
	if event is InputEventMouseButton and event.pressed:
		match event.button_index:
			MOUSE_BUTTON_LEFT:
				if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN and player_hp > 0:
					var pick := _resolve_click_at_mouse()
					_sustained_click.begin_from_pick(pick)
					_execute_click_pick(pick)
			MOUSE_BUTTON_WHEEL_UP:
				_adjust_camera_zoom(-CAMERA_ZOOM_STEP)
			MOUSE_BUTTON_WHEEL_DOWN:
				_adjust_camera_zoom(CAMERA_ZOOM_STEP)


func _handle_input(delta: float) -> void:
	if _user_input_blocked() or client.ready_state() != WebSocketPeer.STATE_OPEN:
		if _sustained_click.active:
			_sustained_click.clear()
		return

	_send_cooldown -= delta
	_attack_cooldown -= delta
	if player_hp <= 0:
		if _sustained_click.active:
			_sustained_click.clear()
		return

	var input := Vector2.ZERO
	if Input.is_key_pressed(KEY_W): input.y -= 1
	if Input.is_key_pressed(KEY_S): input.y += 1
	if Input.is_key_pressed(KEY_A): input.x -= 1
	if Input.is_key_pressed(KEY_D): input.x += 1

	if input != Vector2.ZERO and _send_cooldown <= 0.0:
		var dir := _camera_relative_flat_direction(input)
		# Local prediction: move immediately for responsive feel.
		predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
		_reconcile_player()
		client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
		_send_cooldown = SEND_INTERVAL

	if _hold_input_allowed():
		_tick_sustained_click()
	elif _sustained_click.active:
		_sustained_click.clear()


func _is_inventory_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_I or event.physical_keycode == KEY_I or event.unicode == 105 or event.unicode == 73


func _is_character_stats_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_C or event.physical_keycode == KEY_C or event.unicode == 99 or event.unicode == 67


func _is_escape_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_ESCAPE or event.physical_keycode == KEY_ESCAPE


func _handle_escape() -> void:
	if settings_panel != null and settings_panel.visible:
		_on_settings_back()
		return
	if character_panel != null and character_panel.visible:
		character_panel.hide_panel()
		if main_menu != null:
			main_menu.show_menu()
		return
	if pause_menu != null and pause_menu.visible:
		_resume_from_pause()
		return
	if gameplay_active:
		_show_pause_menu()


func _hotbar_slot_for_key(event: InputEventKey) -> int:
	var code := event.keycode if event.keycode != KEY_NONE else event.physical_keycode
	if code >= KEY_1 and code <= KEY_9:
		return int(code - KEY_1)
	if code == KEY_0:
		return 9
	return -1


func _input_locked() -> bool:
	return visual_replay_enabled or autoplay_enabled or _menu_blocks_gameplay_input()


func _user_input_blocked() -> bool:
	# Replay/autoplay fully lock input. Bot mode blocks real mouse/WASD but still
	# allows push_input() key events through _unhandled_input().
	return _input_locked() or bot_mode


func _menu_blocks_gameplay_input() -> bool:
	return (main_menu != null and main_menu.visible) \
		or (character_panel != null and character_panel.visible) \
		or (settings_panel != null and settings_panel.visible) \
		or (pause_menu != null and pause_menu.visible)


func _handle_autoplay(delta: float) -> void:
	if client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	autoplay_timer -= delta
	autoplay_attack_cooldown -= delta
	if autoplay_timer > 0.0:
		return

	match autoplay_phase:
		"move":
			if not autoplay_move_sent:
				var dir := Vector2(1, 0)
				predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
				_reconcile_player()
				client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
				autoplay_move_sent = true
				autoplay_timer = autoplay_step_delay
				return
			autoplay_phase = "attack"
		"attack":
			if monster_ids.is_empty():
				return
			var target_id := str(monster_ids[0])
			if not entities.has(target_id):
				return
			var rec: Dictionary = entities[target_id]
			var target_node := rec["node"] as Node3D
			if target_node == null:
				return
			var to_target := target_node.position - predicted_pos
			var aim := Vector2(to_target.x, to_target.z).normalized()
			if aim != Vector2.ZERO:
				_face_direction(aim)
			if player_anim != null:
				player_anim.play_one_shot("attack")
			if autoplay_attack_cooldown <= 0.0:
				client.send("action_intent", last_server_tick, {"target_id": target_id})
				autoplay_attack_cooldown = autoplay_step_delay
			autoplay_timer = autoplay_step_delay
		"pickup":
			if not autoplay_pickup_sent and loot_ids.size() > 0:
				client.send("action_intent", last_server_tick, {"target_id": loot_ids[0]})
				autoplay_pickup_sent = true
				autoplay_timer = autoplay_step_delay
				return
			if autoplay_pickup_sent and inventory.size() > 0:
				autoplay_phase = "equip"
			else:
				autoplay_timer = autoplay_step_delay
		"equip":
			if not autoplay_equip_sent and inventory.size() > 0:
				var item_id := str(inventory[0]["item_instance_id"])
				client.send("equip_intent", last_server_tick, {"item_instance_id": item_id, "slot": "main_hand"})
				autoplay_equip_sent = true
				autoplay_timer = autoplay_step_delay
				return
			var weapon_id = equipped.get("main_hand", null)
			if weapon_id != null:
				autoplay_phase = "done"
				_debug("visual bot complete: equipped weapon %s, player_hp=%d" % [str(weapon_id), player_hp])
		"done":
			return


func _hold_input_allowed() -> bool:
	if _input_locked() or bot_mode:
		return false

	if inventory_panel != null and inventory_panel.visible:
		return false
	if character_stats_panel != null and character_stats_panel.visible:
		return false

	return true


func _resolve_click_at_mouse() -> Dictionary:
	var target_id := _pick_entity_at_mouse()
	if target_id == "" or not entities.has(target_id):
		var ground := _mouse_ground_point()
		var loot_id := _nearest_loot_at_ground(ground)
		if loot_id == "":
			return {"kind": "floor", "ground": ground}
		return {"kind": "oneshot", "target_id": loot_id}

	var rec: Dictionary = entities[target_id]
	var typ := str(rec.get("type", ""))
	if typ == "monster" and not _is_dead_monster(target_id):
		return {"kind": "monster", "target_id": target_id}

	return {"kind": "oneshot", "target_id": target_id}


func _execute_click_pick(pick: Dictionary) -> void:
	if _attack_cooldown > 0.0 or player_hp <= 0:
		return

	var kind := str(pick.get("kind", ""))
	if kind == "floor":
		var ground: Vector3 = pick.get("ground", Vector3.ZERO)
		client.send("move_to_intent", last_server_tick, {"position": {"x": ground.x, "y": ground.z}})
		_attack_cooldown = SEND_INTERVAL
		if _sustained_click.mode == "move":
			_sustained_click.mark_move_sent(ground)
		return

	var target_id := str(pick.get("target_id", ""))
	if target_id == "" or not entities.has(target_id):
		return

	var rec: Dictionary = entities[target_id]
	var target_node := rec["node"] as Node3D
	var flat := Vector2(target_node.global_position.x - player_anchor.global_position.x, target_node.global_position.z - player_anchor.global_position.z)
	if flat.length_squared() > 0.0001:
		_face_direction(flat.normalized())

	var typ := str(rec.get("type", ""))
	var state := str(rec.get("state", ""))
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	if typ == "interactable" and interactable_def_id in ["stairs_down", "stairs_up"]:
		_activate_or_approach_interactable(target_id, rec)
		return
	if typ == "interactable" and interactable_def_id == "teleporter":
		_activate_or_approach_interactable(target_id, rec)
		return
	if player_anim != null and (typ == "monster" or (typ == "interactable" and state == "closed")):
		player_anim.play_one_shot("attack")

	client.send("action_intent", last_server_tick, {"target_id": target_id})
	_attack_cooldown = SEND_INTERVAL


func _tick_sustained_click() -> void:
	if not _sustained_click.active:
		return

	if _attack_cooldown > 0.0:
		return

	if _sustained_click.should_stop(player_hp, entities):
		_sustained_click.clear()
		return

	if _sustained_click.mode == "attack":
		_repeat_hold_attack()
	elif _sustained_click.mode == "move":
		_repeat_hold_move()


func _repeat_hold_attack() -> void:
	var target_id := _sustained_click.target_id
	if target_id == "" or not entities.has(target_id):
		_sustained_click.clear()
		return

	var rec: Dictionary = entities[target_id]
	var target_node := rec["node"] as Node3D
	if target_node == null:
		_sustained_click.clear()
		return

	var flat := Vector2(
		target_node.global_position.x - player_anchor.global_position.x,
		target_node.global_position.z - player_anchor.global_position.z
	)
	if flat.length_squared() > 0.0001:
		_face_direction(flat.normalized())

	if player_anim != null:
		player_anim.play_one_shot("attack")

	client.send("action_intent", last_server_tick, {"target_id": target_id})
	_attack_cooldown = SEND_INTERVAL


func _repeat_hold_move() -> void:
	var ground := _mouse_ground_point()
	if not _sustained_click.can_repeat_move(ground):
		return

	client.send("move_to_intent", last_server_tick, {"position": {"x": ground.x, "y": ground.z}})
	_sustained_click.mark_move_sent(ground)
	_attack_cooldown = SEND_INTERVAL


func _try_action_at_mouse() -> void:
	if _attack_cooldown > 0.0 or player_hp <= 0:
		return

	_execute_click_pick(_resolve_click_at_mouse())


func _activate_or_approach_interactable(target_id: String, rec: Dictionary) -> void:
	if _interactable_in_activation_range(rec):
		_activate_interactable_now(target_id, rec)
		return
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	pending_interactable_action = {
		"target_id": target_id,
		"interactable_def_id": interactable_def_id,
	}
	if interactable_def_id == "teleporter":
		client.send("action_intent", last_server_tick, {"target_id": target_id})
		_attack_cooldown = SEND_INTERVAL
		return
	var target_node := rec["node"] as Node3D
	if target_node == null:
		pending_interactable_action.clear()
		return
	client.send("move_to_intent", last_server_tick, {
		"position": {"x": target_node.global_position.x, "y": target_node.global_position.z},
	})
	_attack_cooldown = SEND_INTERVAL


func _try_complete_pending_interactable_action() -> void:
	if pending_interactable_action.is_empty() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	var target_id := str(pending_interactable_action.get("target_id", ""))
	if target_id == "" or not entities.has(target_id):
		pending_interactable_action.clear()
		return
	var rec: Dictionary = entities[target_id]
	if not _interactable_in_activation_range(rec):
		return
	pending_interactable_action.clear()
	_activate_interactable_now(target_id, rec)


func _interactable_in_activation_range(rec: Dictionary) -> bool:
	var target_node := rec.get("node") as Node3D
	if target_node == null or player_anchor == null:
		return false
	var flat := Vector2(
		target_node.global_position.x - player_anchor.global_position.x,
		target_node.global_position.z - player_anchor.global_position.z
	)
	return flat.length() <= INTERACTABLE_ACTIVATION_RANGE


func _activate_interactable_now(target_id: String, rec: Dictionary) -> void:
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	if interactable_def_id == "stairs_down":
		client.send("descend_intent", last_server_tick, {})
		_attack_cooldown = SEND_INTERVAL
		return
	if interactable_def_id == "stairs_up":
		client.send("ascend_intent", last_server_tick, {})
		_attack_cooldown = SEND_INTERVAL
		return
	if interactable_def_id == "teleporter":
		if bool(discovered_teleporters.get(current_level, false)):
			_show_waypoint_panel()
		else:
			client.send("action_intent", last_server_tick, {"target_id": target_id})
			_attack_cooldown = SEND_INTERVAL


func _update_facing_toward_mouse() -> void:
	var aim := _aim_direction_from_mouse()
	if aim != Vector2.ZERO:
		_face_direction(aim)


func _face_direction(flat_dir: Vector2) -> void:
	if character_visual == null or player_anchor == null:
		return

	var target := player_anchor.global_position + Vector3(flat_dir.x, 0.0, flat_dir.y)
	character_visual.look_at(target, Vector3.UP)


func _camera_relative_flat_direction(input: Vector2) -> Vector2:
	# WASD is screen-relative under the isometric camera, not world X/Z.
	if _camera == null or input == Vector2.ZERO:
		return Vector2.ZERO

	var forward := -_camera.global_transform.basis.z
	forward.y = 0.0
	if forward.length_squared() < 0.0001:
		return input.normalized()
	forward = forward.normalized()

	var right := _camera.global_transform.basis.x
	right.y = 0.0
	if right.length_squared() < 0.0001:
		return input.normalized()
	right = right.normalized()

	var world := right * input.x - forward * input.y
	return Vector2(world.x, world.z).normalized()


func _aim_direction_from_mouse() -> Vector2:
	if _camera == null or player_anchor == null:
		return Vector2.ZERO

	var ground := _mouse_ground_point()
	var flat := Vector2(ground.x - player_anchor.global_position.x, ground.z - player_anchor.global_position.z)
	if flat.length_squared() < 0.0001:
		return Vector2.ZERO

	return flat.normalized()


func _mouse_ground_point() -> Vector3:
	var mouse_pos := get_viewport().get_mouse_position()
	var origin := _camera.project_ray_origin(mouse_pos)
	var normal := _camera.project_ray_normal(mouse_pos)
	if abs(normal.y) < 0.0001:
		return player_anchor.global_position

	var t := -origin.y / normal.y
	if t < 0.0:
		return player_anchor.global_position

	return origin + normal * t


func _pick_entity_at_mouse() -> String:
	if _camera == null:
		return ""
	var mouse_pos := get_viewport().get_mouse_position()
	var origin := _camera.project_ray_origin(mouse_pos)
	var normal := _camera.project_ray_normal(mouse_pos)
	var query := PhysicsRayQueryParameters3D.create(origin, origin + normal * 200.0)
	query.collide_with_areas = true
	query.collide_with_bodies = true
	var hit := get_world_3d().direct_space_state.intersect_ray(query)
	if hit.is_empty():
		return ""
	var collider = hit.get("collider")
	if collider != null and collider.has_meta("entity_id"):
		var hit_entity_id := str(collider.get_meta("entity_id"))
		if _is_dead_monster(hit_entity_id):
			var loot_id := _nearest_loot_at_ground(hit.get("position", _mouse_ground_point()))
			if loot_id != "":
				return loot_id
		return hit_entity_id
	return ""


func _is_dead_monster(entity_id: String) -> bool:
	if not entities.has(entity_id):
		return false
	var rec: Dictionary = entities[entity_id]
	return str(rec.get("type", "")) == "monster" and int(rec.get("hp", 1)) <= 0


func _nearest_loot_at_ground(ground: Vector3) -> String:
	var best_id := ""
	var best_dist := 999999.0
	for loot_id in loot_ids:
		if not entities.has(loot_id):
			continue
		var node := entities[loot_id]["node"] as Node3D
		if node == null:
			continue
		var flat_dist := Vector2(node.global_position.x - ground.x, node.global_position.z - ground.z).length()
		if flat_dist < best_dist:
			best_dist = flat_dist
			best_id = str(loot_id)
	if best_id != "" and best_dist <= 0.9:
		return best_id
	return ""


func _set_pickable(node: Node3D, pickable: bool) -> void:
	if node == null:
		return
	var body := node.find_child("PickBody", true, false) as CollisionObject3D
	if body == null:
		return
	body.collision_layer = 1 if pickable else 0
	body.collision_mask = 1 if pickable else 0
	body.input_ray_pickable = pickable


func _adjust_camera_zoom(delta_size: float) -> void:
	if _camera == null:
		return

	_camera.size = clampf(_camera.size + delta_size, CAMERA_ZOOM_MIN, CAMERA_ZOOM_MAX)


# --- visual replay playlist -------------------------------------------------

func _load_visual_replay_manifest(path: String) -> bool:
	if not FileAccess.file_exists(path):
		push_error("visual replay manifest not found: %s" % path)
		return false
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_error("visual replay manifest is not a JSON object: %s" % path)
		return false
	visual_replay_scenarios = parsed.get("scenarios", [])
	return visual_replay_scenarios.size() > 0


func _start_next_visual_replay() -> void:
	visual_replay_index += 1
	visual_replay_envelopes = []
	visual_replay_envelope_index = 0
	visual_replay_timer = autoplay_step_delay
	if visual_replay_index >= visual_replay_scenarios.size():
		visual_replay_title = "complete"
		_debug("visual replay playlist complete")
		if visual_replay_exit_on_complete:
			visual_replay_exit_requested = true
			visual_replay_exit_timer = maxf(autoplay_step_delay, 0.25)
		return

	var scenario: Dictionary = visual_replay_scenarios[visual_replay_index]
	var session_id := str(scenario.get("session_id", ""))
	var world_id := str(scenario.get("world_id", "vertical_slice"))
	current_world_id = world_id
	visual_replay_title = str(scenario.get("title", scenario.get("id", session_id)))
	var visual_cfg: Dictionary = scenario.get("visual", {})
	visual_replay_show_inventory = bool(visual_cfg.get("inventory_panel", false)) \
		or world_id == "inventory_lab" \
		or str(scenario.get("id", "")) == "inventory_lab"
	if inventory_panel != null:
		inventory_panel.set_interactive(false)
		if not visual_replay_show_inventory:
			inventory_panel.hide_display()
	if consumable_bar != null:
		consumable_bar.set_interactive(false)
	if character_stats_panel != null:
		character_stats_panel.set_allocation_enabled(false)
	_render_world_walls(world_id)
	if session_id == "":
		_debug("visual replay entry missing session_id; skipping")
		_start_next_visual_replay()
		return
	var replay_email := str(scenario.get("replay_email", ""))
	if replay_email != "" and replay_email != _env("ARPG_EMAIL", "client@example.test"):
		if not client.login(replay_email, visual_replay_dev_token):
			_debug("visual replay login failed for %s; skipping %s" % [replay_email, visual_replay_title])
			_start_next_visual_replay()
			return
	var through_tick := int(scenario.get("final_tick", -1))
	var timeline := client.get_replay_timeline(visual_replay_debug_token, session_id, through_tick)
	visual_replay_envelopes = timeline.get("envelopes", [])
	_debug("visual replay %d/%d: %s (%d envelopes)" % [
		visual_replay_index + 1, visual_replay_scenarios.size(), visual_replay_title, visual_replay_envelopes.size()])
	if visual_replay_envelopes.is_empty():
		_start_next_visual_replay()


func _handle_visual_replay(delta: float) -> void:
	if visual_replay_exit_requested:
		visual_replay_exit_timer -= delta
		if visual_replay_exit_timer <= 0.0:
			_debug("visual replay exit requested")
			if client != null:
				client.close()
			get_tree().quit(0)
		return
	if visual_replay_index >= visual_replay_scenarios.size():
		return
	visual_replay_timer -= delta
	if visual_replay_timer > 0.0:
		return
	if visual_replay_envelope_index >= visual_replay_envelopes.size():
		visual_replay_timer = maxf(autoplay_step_delay * 4.0, 0.5)
		_start_next_visual_replay()
		return

	var env: Dictionary = visual_replay_envelopes[visual_replay_envelope_index]
	visual_replay_envelope_index += 1
	_handle_message(env)
	visual_replay_timer = _visual_replay_delay_for(env)


func _visual_replay_delay_for(env: Dictionary) -> float:
	if str(env.get("type", "")) != "state_delta":
		return autoplay_step_delay
	var payload: Dictionary = env.get("payload", {})
	var delay := autoplay_step_delay
	for change in payload.get("changes", []):
		var op := str(change.get("op", ""))
		if op in ["entity_spawn", "entity_update"]:
			var entity: Dictionary = change.get("entity", {})
			if str(entity.get("type", "")) == "projectile":
				return 0.05
		if op == "entity_remove":
			return 0.08
		if op in ["inventory_add", "inventory_update", "inventory_remove", "equipped_update"]:
			delay = maxf(delay, autoplay_step_delay * 1.35)
	return delay


func _sync_inventory_replay_display() -> void:
	if inventory_panel == null or not visual_replay_enabled:
		return
	var has_inventory := inventory.size() > 0 or equipped.get("main_hand") != null
	if visual_replay_show_inventory or has_inventory:
		inventory_panel.ensure_display_visible()
		inventory_panel.set_interactive(false)
	else:
		inventory_panel.hide_display()


func _entity_world_center(entity_id: String) -> Vector3:
	if not entities.has(entity_id):
		return Vector3.ZERO
	var node := entities[entity_id]["node"] as Node3D
	if node == null:
		return Vector3.ZERO
	return node.global_position


# --- scene construction (placeholder primitives) ----------------------------

func _build_scene() -> void:
	_camera = Camera3D.new()
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	_camera.size = CAMERA_ZOOM_DEFAULT
	_camera.position = CAMERA_FOLLOW_OFFSET
	add_child(_camera)
	# look_at requires the node to be inside the scene tree (Godot 4).
	_sync_camera_to_player()

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-50, -40, 0)
	add_child(light)

	var ui := CanvasLayer.new()
	add_child(ui)
	_debug_label = Label.new()
	_debug_label.position = Vector2(12, 12)
	ui.add_child(_debug_label)
	_level_label = Label.new()
	_level_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_level_label.position = Vector2(0, 12)
	_level_label.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	_level_label.offset_left = -260
	_level_label.offset_right = -12
	_level_label.offset_top = 12
	_level_label.offset_bottom = 44
	ui.add_child(_level_label)
	_update_level_hud()
	_setup_waypoint_panel(ui)
	inventory_panel = InventoryPanelScript.new()
	inventory_panel.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(inventory_panel)
	consumable_bar = ConsumableBarScript.new()
	consumable_bar.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(consumable_bar)
	character_stats_panel = CharacterStatsPanelScript.new()
	character_stats_panel.allocate_stat_requested.connect(_on_character_stat_requested)
	ui.add_child(character_stats_panel)
	_health_bar = PlayerHealthBarScript.new()
	ui.add_child(_health_bar)
	_setup_menu_layer()

	input_shadow = InputShadowOverlayScript.new()
	add_child(input_shadow)
	input_shadow.bind_camera(_camera)

	damage_numbers_layer = CanvasLayer.new()
	damage_numbers_layer.layer = 2
	add_child(damage_numbers_layer)

	health_bars_layer = CanvasLayer.new()
	health_bars_layer.layer = 1
	add_child(health_bars_layer)

	walls_root = Node3D.new()
	walls_root.name = "StaticWalls"
	add_child(walls_root)


func _setup_menu_layer() -> void:
	menu_layer = CanvasLayer.new()
	menu_layer.layer = 10
	add_child(menu_layer)

	main_menu = MainMenuScript.new()
	main_menu.continue_pressed.connect(_on_continue_pressed)
	main_menu.new_game_pressed.connect(_on_new_game_pressed)
	main_menu.settings_pressed.connect(_on_settings_from_main)
	main_menu.exit_pressed.connect(_exit_game)
	menu_layer.add_child(main_menu)

	character_panel = CharacterSelectPanelScript.new()
	character_panel.back_requested.connect(func() -> void:
		character_panel.hide_panel()
		main_menu.show_menu()
	)
	character_panel.start_requested.connect(_start_character_session)
	character_panel.create_requested.connect(_on_character_create_requested)
	character_panel.delete_requested.connect(_on_character_delete_requested)
	menu_layer.add_child(character_panel)

	settings_panel = SettingsPanelScript.new()
	settings_panel.back_requested.connect(_on_settings_back)
	settings_panel.size_selected.connect(_on_window_size_selected)
	settings_panel.floating_combat_text_toggled.connect(_on_floating_combat_text_toggled)
	menu_layer.add_child(settings_panel)

	pause_menu = PauseMenuScript.new()
	pause_menu.resume_pressed.connect(_resume_from_pause)
	pause_menu.settings_pressed.connect(_on_settings_from_pause)
	pause_menu.return_to_menu_pressed.connect(_return_to_main_menu)
	pause_menu.exit_pressed.connect(_exit_game)
	menu_layer.add_child(pause_menu)


func _on_inventory_intent_requested(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send(intent_type, last_server_tick, payload)


func _on_character_stat_requested(stat: String) -> void:
	if _stat_allocation_blocked():
		return
	client.send("allocate_stat_intent", last_server_tick, {"stat": stat, "points": 1})


func _stat_allocation_blocked() -> bool:
	return visual_replay_enabled \
		or autoplay_enabled \
		or _menu_blocks_gameplay_input() \
		or client == null \
		or client.ready_state() != WebSocketPeer.STATE_OPEN \
		or player_hp <= 0 \
		or int(character_progression.get("unspent_stat_points", 0)) <= 0


func _refresh_progression_ui() -> void:
	if character_stats_panel != null:
		character_stats_panel.set_progression(character_progression)
		character_stats_panel.set_allocation_enabled(not _stat_allocation_blocked())
	if consumable_bar != null:
		consumable_bar.set_character_progression(character_progression)


func _sync_progression_interactivity() -> void:
	if character_stats_panel != null:
		character_stats_panel.set_allocation_enabled(not _stat_allocation_blocked())


func _setup_waypoint_panel(ui: CanvasLayer) -> void:
	waypoint_panel = PanelContainer.new()
	waypoint_panel.visible = false
	waypoint_panel.position = Vector2(16, 96)
	waypoint_panel.custom_minimum_size = Vector2(WaypointPanelConfig.PANEL_MIN_WIDTH_PX, 0)
	var panel_box := VBoxContainer.new()
	panel_box.add_theme_constant_override("separation", 6)
	var title := Label.new()
	title.text = "Teleport"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_LEFT
	panel_box.add_child(title)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(
		WaypointPanelConfig.SCROLL_MIN_WIDTH_PX,
		WaypointPanelConfig.SCROLL_VIEWPORT_UNIT_PX * WaypointPanelConfig.SCROLL_MAX_VISIBLE_ROWS
	)
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	waypoint_rows = VBoxContainer.new()
	waypoint_rows.add_theme_constant_override("separation", 4)
	scroll.add_child(waypoint_rows)
	panel_box.add_child(scroll)
	waypoint_panel.add_child(panel_box)
	ui.add_child(waypoint_panel)


func _apply_teleporter_snapshot(rows: Array) -> void:
	discovered_teleporters.clear()
	for row in rows:
		if typeof(row) != TYPE_DICTIONARY:
			continue
		discovered_teleporters[int(row.get("level", 0))] = bool(row.get("discovered", false))


func _show_waypoint_panel() -> void:
	if waypoint_panel == null:
		return
	_refresh_waypoint_panel()
	waypoint_panel.visible = true


func _hide_waypoint_panel() -> void:
	if waypoint_panel != null:
		waypoint_panel.visible = false


func _sync_waypoint_panel_reach() -> void:
	if waypoint_panel == null or not waypoint_panel.visible:
		return
	var teleporter := _current_teleporter_record()
	if teleporter.is_empty() or not _interactable_in_activation_range(teleporter):
		_hide_waypoint_panel()


func _refresh_waypoint_panel() -> void:
	if waypoint_panel == null or waypoint_rows == null:
		return
	for child in waypoint_rows.get_children():
		child.queue_free()
	var levels := discovered_teleporter_levels()
	for level in levels:
		var row := Button.new()
		row.custom_minimum_size = Vector2(204, WaypointPanelConfig.ROW_HEIGHT_PX)
		row.text = _waypoint_row_text(level)
		row.disabled = not bool(discovered_teleporters.get(level, false))
		row.pressed.connect(_on_waypoint_level_pressed.bind(level))
		waypoint_rows.add_child(row)


func discovered_teleporter_levels() -> Array:
	var levels: Array = discovered_teleporters.keys()
	levels.sort()
	return levels


func _waypoint_row_text(level: int) -> String:
	var depth: int = abs(level)
	var state := "" if bool(discovered_teleporters.get(level, false)) else " (undiscovered)"
	return "Level %d - %s%s" % [depth, _dungeon_level_name(level), state]


func _on_waypoint_level_pressed(level: int) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if level == current_level:
		_hide_waypoint_panel()
		return
	var teleporter := _current_teleporter_record()
	if teleporter.is_empty() or _interactable_in_activation_range(teleporter):
		client.send("teleport_intent", last_server_tick, {"target_level": level})
	else:
		var target_node := teleporter["node"] as Node3D
		if target_node == null:
			return
		pending_waypoint_target_level = level
		pending_waypoint_travel = true
		client.send("move_to_intent", last_server_tick, {
			"position": {"x": target_node.global_position.x, "y": target_node.global_position.z},
		})
	_hide_waypoint_panel()


func _try_complete_pending_waypoint_travel() -> void:
	if not pending_waypoint_travel or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	var teleporter := _current_teleporter_record()
	if teleporter.is_empty():
		pending_waypoint_travel = false
		return
	if not _interactable_in_activation_range(teleporter):
		return
	var target_level := pending_waypoint_target_level
	pending_waypoint_travel = false
	client.send("teleport_intent", last_server_tick, {"target_level": target_level})


func _current_teleporter_record() -> Dictionary:
	for id in interactable_ids:
		var rec: Dictionary = entities.get(id, {})
		if str(rec.get("interactable_def_id", "")) == "teleporter":
			return rec
	return {}


func _render_world_walls(world_id: String) -> void:
	if walls_root == null:
		return
	for child in walls_root.get_children():
		child.queue_free()

	var rules_path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/worlds.v0.json")
	var parsed = _read_json(rules_path)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("[main] could not read world rules for walls: %s" % rules_path)
		return
	var worlds: Dictionary = parsed.get("worlds", {})
	var world: Dictionary = worlds.get(world_id, {})
	if str(world.get("mode", "")) == "multi_level":
		if current_level < 0:
			_render_dungeon_walls()
		return
	for entity in world.get("entities", []):
		if str(entity.get("type", "")) != "wall":
			continue
		var pos: Dictionary = entity.get("position", {})
		var size: Dictionary = entity.get("size", {})
		var node := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(float(size.get("x", 1.0)), 1.0, float(size.get("y", 1.0)))
		node.mesh = mesh
		node.position = Vector3(float(pos.get("x", 0.0)), 0.5, float(pos.get("y", 0.0)))
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color(0.32, 0.34, 0.36)
		node.material_override = mat
		walls_root.add_child(node)


func _render_dungeon_walls() -> void:
	if dungeon_generation.is_empty():
		return
	var floor_size: Dictionary = dungeon_generation.get("floor_size", {})
	var width := float(floor_size.get("width", 32.0))
	var height := float(floor_size.get("height", 20.0))
	var thickness := float(dungeon_generation.get("wall_thickness", 1.0))
	var half := thickness / 2.0
	var walls := [
		{"position": {"x": width / 2.0, "y": -half}, "size": {"x": width + thickness * 2.0, "y": thickness}},
		{"position": {"x": width / 2.0, "y": height + half}, "size": {"x": width + thickness * 2.0, "y": thickness}},
		{"position": {"x": -half, "y": height / 2.0}, "size": {"x": thickness, "y": height}},
		{"position": {"x": width + half, "y": height / 2.0}, "size": {"x": thickness, "y": height}},
	]
	for wall in walls:
		var pos: Dictionary = wall["position"]
		var size: Dictionary = wall["size"]
		var node := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(float(size["x"]), 1.0, float(size["y"]))
		node.mesh = mesh
		node.position = Vector3(float(pos["x"]), 0.5, float(pos["y"]))
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color(0.24, 0.25, 0.27)
		node.material_override = mat
		walls_root.add_child(node)


func _update_level_hud() -> void:
	if _level_label == null:
		return
	if current_level == 0:
		_level_label.visible = false
		_level_label.text = ""
		return
	_level_label.visible = true
	var depth: int = abs(current_level)
	_level_label.text = "Level %d - %s" % [depth, _dungeon_level_name(current_level)]


func _dungeon_level_name(level: int) -> String:
	var names: Dictionary = dungeon_generation.get("level_names", {})
	var key := str(level)
	if names.has(key):
		return str(names[key])
	var template := str(dungeon_generation.get("default_level_name_template", "Depth {n}"))
	return template.replace("{n}", str(abs(level)))


func _read_json(path: String):
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return null
	return JSON.parse_string(f.get_as_text())


func _make_entity_node(e: Dictionary) -> Node3D:
	var kind := str(e.get("type", ""))
	# Monster adopts the rigged dummy scene (spec §5.3); loot uses shared
	# presentation metadata while gameplay stays server-owned.
	if kind == "monster":
		var packed := MonsterDummyScene
		if packed != null:
			var monster := packed.instantiate()
			_apply_model_tint(monster, _monster_tint(str(e.get("rarity", "common"))))
			return monster
		# Fallback: red primitive so positioning/targeting still works.
		var fallback := MeshInstance3D.new()
		var fm := StandardMaterial3D.new()
		fm.albedo_color = _monster_tint(str(e.get("rarity", "common")))
		fallback.mesh = BoxMesh.new()
		fallback.material_override = fm
		return fallback
	if kind == "player":
		return _make_remote_player_node(e)
	if kind == "interactable":
		var def_id := str(e.get("interactable_def_id", ""))
		if def_id == "stairs_down" or def_id == "stairs_up":
			return _make_stair_node(def_id)
		if def_id == "teleporter":
			return _make_teleporter_node()
		return _make_door_node()
	if kind == "projectile":
		return _make_projectile_node()
	return _make_loot_node(str(e.get("item_def_id", "")))


func _make_remote_player_node(e: Dictionary) -> Node3D:
	var root = CharacterScene.instantiate() as Node3D
	root.name = "RemotePlayer_%s" % str(e.get("id", ""))
	_apply_model_tint(root, REMOTE_PLAYER_TINT)
	return root


func _monster_tint(rarity: String) -> Color:
	return MONSTER_RARITY_TINTS.get(rarity, MONSTER_RARITY_TINTS["common"])


func _entity_base_tint(e: Dictionary) -> Color:
	var kind := str(e.get("type", ""))
	if kind == "player":
		return REMOTE_PLAYER_TINT
	if kind == "monster":
		return _monster_tint(str(e.get("rarity", "common")))
	return Color.WHITE


func _apply_model_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		_apply_model_tint(child, color)


func _make_loot_node(item_def_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "Loot_%s" % item_def_id
	var ground: Dictionary = item_presentations.get(item_def_id, {}).get("ground", {})
	var shape := str(ground.get("shape", "box"))
	var color := Color(str(ground.get("color", "#" + _loot_color(item_def_id).to_html(false))))
	var accent := Color(str(ground.get("accent", "#f6e8b1")))
	var scale := float(ground.get("scale", 1.0))
	match shape:
		"blade":
			_add_loot_box(root, "Blade", Vector3(0.12, 0.08, 0.78) * scale, Vector3(0.0, 0.20, 0.0), color)
			_add_loot_box(root, "Grip", Vector3(0.34, 0.10, 0.10) * scale, Vector3(0.0, 0.16, 0.34 * scale), accent)
		"bow":
			_add_loot_box(root, "BowTop", Vector3(0.10, 0.08, 0.42) * scale, Vector3(0.14 * scale, 0.20, -0.18 * scale), color)
			_add_loot_box(root, "BowBottom", Vector3(0.10, 0.08, 0.42) * scale, Vector3(-0.14 * scale, 0.20, 0.18 * scale), color)
			_add_loot_box(root, "String", Vector3(0.04, 0.06, 0.75) * scale, Vector3(0.0, 0.18, 0.0), accent)
		"coin":
			_add_loot_cylinder(root, "Badge", 0.24 * scale, 0.08 * scale, Vector3(0.0, 0.16, 0.0), color)
			_add_loot_cylinder(root, "BadgeMark", 0.12 * scale, 0.10 * scale, Vector3(0.0, 0.21, 0.0), accent)
		"leaf":
			_add_loot_box(root, "Leaf", Vector3(0.42, 0.06, 0.24) * scale, Vector3(0.0, 0.16, 0.0), color)
			_add_loot_box(root, "Stem", Vector3(0.06, 0.08, 0.46) * scale, Vector3(0.0, 0.18, 0.0), accent)
		"potion":
			_add_loot_cylinder(root, "Bottle", 0.17 * scale, 0.32 * scale, Vector3(0.0, 0.26, 0.0), color)
			_add_loot_box(root, "Cork", Vector3(0.14, 0.10, 0.14) * scale, Vector3(0.0, 0.48 * scale, 0.0), accent)
		_:
			_add_loot_box(root, "Box", Vector3(0.5, 0.5, 0.5) * scale, Vector3(0.0, 0.25 * scale, 0.0), color)
	return root


func _add_loot_box(parent: Node3D, name: String, size: Vector3, position: Vector3, color: Color) -> void:
	var mesh := BoxMesh.new()
	mesh.size = size
	_add_loot_mesh(parent, name, mesh, position, color)


func _add_loot_cylinder(parent: Node3D, name: String, radius: float, height: float, position: Vector3, color: Color) -> void:
	var mesh := CylinderMesh.new()
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = 16
	_add_loot_mesh(parent, name, mesh, position, color)


func _add_loot_mesh(parent: Node3D, name: String, mesh: Mesh, position: Vector3, color: Color) -> void:
	var node := MeshInstance3D.new()
	node.name = name
	node.mesh = mesh
	node.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	node.material_override = mat
	parent.add_child(node)


func _loot_color(item_def_id: String) -> Color:
	var def: Dictionary = item_rules.get(item_def_id, {})
	var category := str(def.get("category", "equipment" if bool(def.get("equippable", false)) else "currency"))
	match category:
		"equipment":
			return Color(0.62, 0.62, 0.62)
		"quest":
			return Color(0.2, 0.85, 0.35)
		"consumable":
			return Color(0.95, 0.15, 0.12)
		_:
			return Color(1.0, 0.85, 0.2)


func _load_item_rules() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/items.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		item_rules = parsed.get("items", {})


func _load_item_presentations() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/item_presentations.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		item_presentations = parsed.get("items", {})


func _load_dungeon_generation() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/dungeon_generation.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		dungeon_generation = parsed


func _make_projectile_node() -> Node3D:
	var root := Node3D.new()
	root.name = "Projectile"
	var shaft := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.16, 0.16, 0.7)
	shaft.mesh = mesh
	shaft.position = Vector3(0.0, 0.35, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.65, 0.90, 1.0)
	mat.emission_enabled = true
	mat.emission = Color(0.25, 0.55, 0.9)
	shaft.material_override = mat
	root.add_child(shaft)
	return root


func _move_projectile_node(rec: Dictionary, target_pos: Vector3) -> void:
	var node := rec["node"] as Node3D
	if node == null:
		return
	var from := node.position
	var flat := Vector2(target_pos.x - from.x, target_pos.z - from.z)
	if flat.length_squared() > 0.0001:
		node.look_at(Vector3(target_pos.x, from.y, target_pos.z), Vector3.UP)
	if rec.has("move_tween"):
		var old_tween = rec["move_tween"]
		if is_instance_valid(old_tween):
			old_tween.kill()
	var duration := PROJECTILE_LERP_SECONDS
	if visual_replay_enabled:
		duration = clampf(autoplay_step_delay * 0.35, 0.06, 0.18)
	var tween := create_tween()
	rec["move_tween"] = tween
	tween.tween_property(node, "position", target_pos, duration).set_trans(Tween.TRANS_LINEAR)


func _make_door_node() -> Node3D:
	var root := Node3D.new()
	root.name = "InteractableDoor"
	var pivot := Node3D.new()
	pivot.name = "DoorPivot"
	pivot.position = Vector3(-0.5, 0.0, 0.0)
	root.add_child(pivot)
	var panel := MeshInstance3D.new()
	panel.name = "DoorPanel"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.0, 1.0, 0.25)
	panel.mesh = mesh
	panel.position = Vector3(0.5, 0.5, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.55, 0.32, 0.15)
	panel.material_override = mat
	pivot.add_child(panel)
	return root


func _make_stair_node(def_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "Stairs_%s" % def_id
	var base := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.1, 0.18, 1.1)
	base.mesh = mesh
	base.position = Vector3(0.0, 0.09, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.22, 0.24, 0.27) if def_id == "stairs_down" else Color(0.46, 0.48, 0.50)
	base.material_override = mat
	root.add_child(base)
	for i in range(3):
		var step := MeshInstance3D.new()
		var smesh := BoxMesh.new()
		smesh.size = Vector3(0.95 - float(i) * 0.18, 0.12, 0.22)
		step.mesh = smesh
		step.position = Vector3(0.0, 0.22 + float(i) * 0.12, -0.34 + float(i) * 0.27)
		var step_mat := StandardMaterial3D.new()
		step_mat.albedo_color = Color(0.58, 0.56, 0.50)
		step.material_override = step_mat
		root.add_child(step)
	return root


func _make_teleporter_node() -> Node3D:
	var root := Node3D.new()
	root.name = "Teleporter"
	var base := MeshInstance3D.new()
	var base_mesh := CylinderMesh.new()
	base_mesh.top_radius = 0.62
	base_mesh.bottom_radius = 0.72
	base_mesh.height = 0.16
	base.mesh = base_mesh
	base.position = Vector3(0.0, 0.08, 0.0)
	var base_mat := StandardMaterial3D.new()
	base_mat.albedo_color = Color(0.16, 0.19, 0.22)
	base.material_override = base_mat
	root.add_child(base)

	var core := MeshInstance3D.new()
	var core_mesh := CylinderMesh.new()
	core_mesh.top_radius = 0.34
	core_mesh.bottom_radius = 0.34
	core_mesh.height = 0.42
	core.mesh = core_mesh
	core.position = Vector3(0.0, 0.32, 0.0)
	var core_mat := StandardMaterial3D.new()
	core_mat.albedo_color = Color(0.15, 0.62, 0.70)
	core_mat.emission_enabled = true
	core_mat.emission = Color(0.05, 0.55, 0.68)
	core.material_override = core_mat
	root.add_child(core)
	return root


func _attach_pick_collider(node: Node3D, entity_id: String, kind: String) -> void:
	var body := StaticBody3D.new()
	body.name = "PickBody"
	body.set_meta("entity_id", entity_id)
	var shape := CollisionShape3D.new()
	var box := BoxShape3D.new()
	match kind:
		"monster":
			box.size = Vector3(1.0, 1.6, 1.0)
			shape.position = Vector3(0.0, 0.8, 0.0)
		"interactable":
			box.size = Vector3(1.2, 1.2, 0.45)
			shape.position = Vector3(0.0, 0.6, 0.0)
		_:
			box.size = Vector3(0.75, 0.75, 0.75)
			shape.position = Vector3(0.0, 0.375, 0.0)
	shape.shape = box
	body.add_child(shape)
	node.add_child(body)


func _set_interactable_state(_entity_id: String, rec: Dictionary, state: String) -> void:
	if rec.get("state", "") == state:
		return
	rec["state"] = state
	var node := rec["node"] as Node3D
	if node == null:
		return
	var pivot := node.find_child("DoorPivot", true, false) as Node3D
	if pivot == null:
		return
	var target_rot := deg_to_rad(90.0) if state == "open" else 0.0
	var tween := create_tween()
	tween.tween_property(pivot, "rotation:y", target_rot, 0.25)


# --- bot API (read-only state + intent dispatch) ----------------------------

func get_bot_state() -> Dictionary:
	# Exclude dead monsters (hp==0) from monster_ids so assert_entity_removed
	# treats a killed monster as "gone" even if the server hasn't sent entity_remove.
	var live_monster_ids: Array = []
	for mid in monster_ids:
		var rec: Dictionary = entities.get(mid, {})
		if int(rec.get("hp", 1)) > 0:
			live_monster_ids.append(mid)
	var out := {
		"ws_open": client != null and client.ready_state() == WebSocketPeer.STATE_OPEN,
		"local_player_id": player_id,
		"party": party.duplicate(true),
		"remote_player_ids": _remote_player_ids(),
		"player_hp": player_hp,
		"player_max_hp": player_max_hp,
		"player_pos": {"x": predicted_pos.x, "z": predicted_pos.z},
		"character_progression": character_progression.duplicate(true),
		"inventory": inventory.duplicate(true),
		"equipped": equipped.duplicate(true),
		"monster_ids": live_monster_ids,
		"entities_debug": _bot_entities_debug(live_monster_ids),
		"local_player_presentation": _bot_local_player_presentation(),
		"entities_presentation_debug": _bot_entities_presentation_debug(),
		"loot_ids": loot_ids.duplicate(),
		"loot": _bot_loot_debug(),
		"interactable_ids": interactable_ids.duplicate(),
		"loot_presentations": _bot_loot_presentations(),
		"inventory_panel_visible": inventory_panel != null and inventory_panel.visible,
		"character_stats_panel_visible": character_stats_panel != null and character_stats_panel.visible,
		"waypoint_panel_visible": waypoint_panel != null and waypoint_panel.visible,
		"inventory_panel": inventory_panel.get_debug_state() if inventory_panel != null else {},
		"character_stats_panel": character_stats_panel.get_debug_state() if character_stats_panel != null else {},
		"consumable_bar": consumable_bar.get_debug_state() if consumable_bar != null else {},
		"pending_events": _bot_pending_events.duplicate(true),
		"main_menu_visible": main_menu != null and main_menu.visible,
		"character_panel_visible": character_panel != null and character_panel.visible,
		"settings_panel_visible": settings_panel != null and settings_panel.visible,
		"pause_menu_visible": pause_menu != null and pause_menu.visible,
		"selected_window_size": ClientSettingsScript.size_label(client_settings.window_size) if client_settings != null else "",
		"floating_combat_text_enabled": client_settings != null and client_settings.floating_combat_text,
		"damage_numbers": _bot_damage_numbers(),
		"known_characters": character_panel.known_characters() if character_panel != null else [],
		"current_session_id": client.session_id if client != null else "",
		"gameplay_active": gameplay_active,
	}
	return out


func _remote_player_ids() -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) == "player":
			out.append(str(id))
	out.sort()
	return out


func _bot_entities_debug(live_monster_ids: Array) -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) == "monster" and not live_monster_ids.has(id):
			continue
		out.append({
			"id": str(id),
			"type": str(rec.get("type", "")),
			"monster_def_id": str(rec.get("monster_def_id", "")),
			"interactable_def_id": str(rec.get("interactable_def_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"state": str(rec.get("state", "")),
		})
	return out


func _bot_local_player_presentation() -> Dictionary:
	return {
		"id": player_id,
		"type": "player",
		"visual_model": "character",
		"base_tint": PLAYER_TINT.to_html(false),
		"reaction": player_reaction.get_debug_state() if player_reaction != null else {},
		"animation": player_anim.get_debug_state() if player_anim != null else {},
	}


func _bot_entities_presentation_debug() -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		var reaction = rec.get("reaction", null)
		var controller = rec.get("controller", null)
		out.append({
			"id": str(id),
			"type": str(rec.get("type", "")),
			"monster_def_id": str(rec.get("monster_def_id", "")),
			"character_id": str(rec.get("character_id", "")),
			"visual_model": _visual_model_name(rec, node),
			"base_tint": str(rec.get("base_tint", "")),
			"hp": int(rec.get("hp", 1)),
			"reaction": reaction.get_debug_state() if reaction != null else {},
			"animation": controller.get_debug_state() if controller != null else {},
		})
	return out


func _visual_model_name(rec: Dictionary, node: Node3D) -> String:
	if node != null and node.find_child("ModelRoot", true, false) != null:
		return "character"
	if str(rec.get("type", "")) == "player":
		return "primitive"
	return ""


func _bot_damage_numbers() -> Array:
	var out: Array = []
	if damage_numbers_layer == null:
		return out
	for child in damage_numbers_layer.get_children():
		if child is DamageNumber:
			var pop := child as DamageNumber
			out.append({
				"text": pop.combat_text,
				"variant": pop.combat_variant,
			})
	return out


func _bot_loot_presentations() -> Dictionary:
	var out := {}
	for loot_id in loot_ids:
		var rec: Dictionary = entities.get(loot_id, {})
		var item_def_id := str(rec.get("item_def_id", ""))
		if item_def_id != "":
			out[item_def_id] = item_presentations.has(item_def_id)
	return out


func _bot_loot_debug() -> Array:
	var out: Array = []
	for loot_id in loot_ids:
		var rec: Dictionary = entities.get(loot_id, {})
		out.append({
			"id": loot_id,
			"item_def_id": str(rec.get("item_def_id", "")),
		})
	return out


func bot_dispatch_action(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send(intent_type, last_server_tick, payload)
	_attack_cooldown = SEND_INTERVAL


func bot_click_entity_id(target_id: String) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if target_id == "" or not entities.has(target_id):
		return
	var rec: Dictionary = entities[target_id]
	var typ := str(rec.get("type", ""))
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	if typ == "interactable" and interactable_def_id in ["stairs_down", "stairs_up", "teleporter"]:
		_activate_or_approach_interactable(target_id, rec)
		return
	client.send("action_intent", last_server_tick, {"target_id": target_id})
	_attack_cooldown = SEND_INTERVAL


func bot_dispatch_inventory_intent(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send(intent_type, last_server_tick, payload)


func bot_assign_consumable_hotbar(slot_index: int, item_instance_id: String) -> void:
	if consumable_bar == null:
		return
	consumable_bar.assign_slot(slot_index, item_instance_id)


func bot_use_consumable_hotbar(slot_index: int) -> void:
	if consumable_bar == null:
		return
	consumable_bar.use_slot(slot_index)


func bot_click_stat_button(stat: String) -> void:
	if character_stats_panel == null:
		return
	character_stats_panel.bot_click_stat_button(stat)


func bot_click_menu_button(button: String) -> void:
	match button:
		"continue":
			_on_continue_pressed()
		"new_game":
			_on_new_game_pressed()
		"settings":
			if pause_menu != null and pause_menu.visible:
				_on_settings_from_pause()
			else:
				_on_settings_from_main()
		"back":
			if settings_panel != null and settings_panel.visible:
				_on_settings_back()
			elif character_panel != null and character_panel.visible:
				character_panel.hide_panel()
				main_menu.show_menu()
		"start":
			if character_panel != null:
				character_panel.submit_name()
		"resume":
			_resume_from_pause()
		"return_to_main_menu":
			_return_to_main_menu()
		"exit":
			_exit_game()


func bot_enter_character_name(name: String) -> void:
	if character_panel != null:
		character_panel.set_name_text(name)


func bot_select_character(index: int) -> void:
	if character_panel != null:
		character_panel.start_character_at_index(index)


func bot_select_window_size(size: String) -> void:
	_on_window_size_selected(size)


func bot_set_floating_combat_text(enabled: bool) -> void:
	_on_floating_combat_text_toggled(enabled)


func bot_consume_pending_event_at(index: int) -> void:
	if index < 0 or index >= _bot_pending_events.size():
		return
	_bot_pending_events.remove_at(index)


func bot_show_action_shadow(action: Dictionary, state: Dictionary) -> void:
	if not bot_mode or input_shadow == null or DisplayServer.get_name() == "headless":
		return

	var stype := str(action.get("_type", action.get("type", "")))
	match stype:
		"press_key":
			var key_name := str(action.get("keycode", "")).trim_prefix("KEY_")
			if key_name != "":
				input_shadow.show_keys(PackedStringArray([key_name]))
		"click_entity":
			var ids_key := "%s_ids" % str(action.get("entity_type", ""))
			var ids: Array = state.get(ids_key, [])
			if ids.is_empty():
				return
			var world := _entity_world_center(str(ids[0]))
			if world != Vector3.ZERO:
				input_shadow.pulse_world_target(world, PackedStringArray(["LMB"]))
		"click_floor":
			var wx := float(action.get("x", 0.0))
			var wz := float(action.get("z", 0.0))
			input_shadow.pulse_world_target(Vector3(wx, 0.0, wz), PackedStringArray(["LMB"]))
		"drag_bag_to_weapon_slot":
			var item_id := _bot_bag_item_id_for_def(str(action.get("item_def_id", "")), state)
			if item_id != "":
				_bot_shadow_inventory_equip(item_id)
		"drag_weapon_to_bag":
			_bot_shadow_inventory_unequip()
		"drag_bag_to_outside":
			var drop_id := _bot_bag_item_id_for_def(str(action.get("item_def_id", "")), state)
			if drop_id != "":
				_bot_shadow_inventory_drop(drop_id)


func _bot_bag_item_id_for_def(item_def_id: String, state: Dictionary) -> String:
	var inv: Array = state.get("inventory", [])
	var eq: Dictionary = state.get("equipped", {})
	var equipped_weapon = eq.get("main_hand", null)
	for item in inv:
		if str(item.get("item_def_id", "")) == item_def_id:
			var iid := str(item.get("item_instance_id", ""))
			if str(equipped_weapon) != iid:
				return iid
	return ""


func _bot_shadow_inventory_equip(item_instance_id: String) -> void:
	if inventory_panel == null:
		return
	inventory_panel.ensure_display_visible()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	var to_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag"]))


func _bot_shadow_inventory_unequip() -> void:
	if inventory_panel == null:
		return
	inventory_panel.ensure_display_visible()
	var from_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_bag_area_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "bag"]))


func _bot_shadow_inventory_drop(item_instance_id: String) -> void:
	if inventory_panel == null:
		return
	inventory_panel.ensure_display_visible()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	if from_pos == Vector2.ZERO:
		from_pos = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_drop_outside_screen_point()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "drop"]))


# --- debug ------------------------------------------------------------------

func _update_debug() -> void:
	var eq = equipped.get("main_hand", null)
	var ws_state := "?"
	if client != null:
		match client.ready_state():
			WebSocketPeer.STATE_CONNECTING: ws_state = "connecting"
			WebSocketPeer.STATE_OPEN: ws_state = "open"
			WebSocketPeer.STATE_CLOSING: ws_state = "closing"
			WebSocketPeer.STATE_CLOSED: ws_state = "closed"
	var weapon_vis := "none"
	if resolver != null:
		var w = resolver.get_debug_state()["equipped_visuals"]["weapon"]
		if w != null:
			weapon_vis = "%s(visible=%s)" % [w["asset_id"], w["visible"]]
	var mode := "visual-replay:%d/%d %s" % [
		min(visual_replay_index + 1, visual_replay_scenarios.size()),
		visual_replay_scenarios.size(),
		visual_replay_title,
	] if visual_replay_enabled else ("bot-client" if bot_mode else ("visual-bot:%s" % autoplay_phase if autoplay_enabled else "manual"))
	_debug_label.text = "ws=%s  tick=%d  mode=%s  recon_delta=%.2f\ninv=%d  entities=%d  equipped_weapon=%s\nweapon_visual=%s\nW/A/S/D move  LMB action  scroll zoom  I inventory" % [
		ws_state, last_server_tick, mode, reconciliation_delta, inventory.size(), entities.size(), str(eq), weapon_vis]


func _debug(msg: String) -> void:
	print("[client] ", msg)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback


func _truthy_env(key: String) -> bool:
	return _truthy_text(OS.get_environment(key))


func _truthy_text(value: String) -> bool:
	var v := value.to_lower()
	return v in ["1", "true", "yes", "on"]
