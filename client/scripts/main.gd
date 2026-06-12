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
const HealRainEffectScript := preload("res://scripts/heal_rain_effect.gd")
const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")
const ProjectileVisualsScript := preload("res://scripts/projectile_visuals.gd")
const MonsterHealthBarScript := preload("res://scripts/monster_health_bar.gd")
const BossHealthBarScript := preload("res://scripts/boss_health_bar.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const StashPanelScript := preload("res://scripts/stash_panel.gd")
const BishopPanelScript := preload("res://scripts/bishop_panel.gd")
const MarketPanelScript := preload("res://scripts/market_panel.gd")
const ConsumableBarScript := preload("res://scripts/consumable_bar.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const CharacterBarScript := preload("res://scripts/character_bar.gd")
const SkillBarScript := preload("res://scripts/skill_bar.gd")
const StatusEffectsBarScript := preload("res://scripts/status_effects_bar.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const InputShadowOverlayScript := preload("res://scripts/input_shadow_overlay.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const MainMenuScript := preload("res://scripts/main_menu.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const MultiplayerSessionsPanelScript := preload("res://scripts/multiplayer_sessions_panel.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const PauseMenuScript := preload("res://scripts/pause_menu.gd")
const SustainedClickInputScript := preload("res://scripts/sustained_click_input.gd")
const DirectionalAttackInputScript := preload("res://scripts/directional_attack_input.gd")
const MonsterVisualsLoaderScript := preload("res://scripts/monster_visuals_loader.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const CharacterScene := preload("res://scenes/character.tscn")
const MonsterDummyScene := preload("res://scenes/monster_dummy.tscn")
const MonsterQuadrupedScene := preload("res://scenes/monster_quadruped.tscn")
const MonsterTinyFlyerScene := preload("res://scenes/monster_tiny_flyer.tscn")
const MonsterSkeletonScene := preload("res://scenes/monster_skeleton.tscn")
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
const SKILL_FUNCTION_KEY_COUNT := 8
const LOCAL_UNARMED_REACH := 1.0
const LOCAL_MONSTER_RADIUS := 0.45
const LOCAL_LOOT_RADIUS := 0.35
const LOCAL_INTERACTABLE_RADIUS := 0.50
const LOCAL_REACH_EPSILON := 0.000001
const PLAYER_TINT := Color("#8fe8a7")
const REMOTE_PLAYER_TINT := Color("#202934")
const POISON_TINT := Color("#38f06f")
const BAG_FULL_CANT_UNEQUIP_TEXT := "bag full, cant unequip"
const NO_MANA_TEXT := "NO MANA"
const HEAL_RAIN_RADIUS := 4.0
const MONSTER_RARITY_TINTS := {
	"common": Color("#f2f2ec"),
	"champion": Color("#9fc7ff"),
	"rare": Color("#ff9b9b"),
	"unique": Color("#ffd978"),
}
const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
}
const LOOT_LABEL_RARITY_COLORS := {
	"common": Color("#e8dcc8"),
	"magic": Color("#93c5fd"),
	"rare": Color("#f4d481"),
	"unique": Color("#ffb26b"),
}
const LOOT_LABEL_CATEGORY_COLORS := {
	"currency": Color("#ffd75e"),
	"quest": Color("#6ee68b"),
	"consumable": Color("#ff8f70"),
}
const LOOT_LABEL_REVEAL_DIM_FACTOR := 0.58
const GROUND_EQUIPMENT_MODEL_SCALE := 1.0
const BOSS_VISUAL_MODEL := "current_humanoid_player"
const BOSS_PHASE_TICK_RATE := 10.0
const BOSS_TELEGRAPH_MARKER_NAME := "BossTelegraphMarker"
const ARCHER_MONSTER_DEF_ID := "dungeon_archer"
const ARCHER_BOW_MARKER_NAME := "ArcherBowMarker"
const CHARACTER_FLOW_CREATE_GAME := "create_game"
const CHARACTER_FLOW_JOIN_GAME := "join_game"
const CHARACTER_FLOW_LEGACY_SOLO := "solo"
const CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST := "multiplayer_host"
const CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN := "multiplayer_join"

var client: NetClient
var resolver: EquipmentVisualResolver
var player_anim: AnimationController
var player_reaction: ModelReactionController
var entities: Dictionary = {}        # id (String) -> {node:Node3D, controller:AnimationController|null, type:String}
var player_id: String = ""
var party: Array = []
var player_hp: int = PLAYER_START_HP
var player_max_hp: int = PLAYER_START_HP
var player_mana: int = 10
var player_max_mana: int = 10
var player_visual_scale: float = 1.0
var _local_player_class_asset_id: String = ""
var predicted_pos := Vector3.ZERO    # client-predicted player position
var reconciliation_delta: float = 0.0
var last_server_tick: int = 0
var inventory: Array = []
var equipped: Dictionary = {}
var inventory_rows: int = 3
var inventory_capacity: int = 15
var gold: int = 0
var stash_items: Array = []
var stash_gold: int = 0
var stash_capacity: int = 50
var pending_stash_equips: Dictionary = {}
var hotbar_capacity: int = 2
var hotbar: Array = []
var character_progression: Dictionary = {}
var skill_progression: Dictionary = {}
var skill_cooldowns: Array = []
var right_click_skill_id: String = ""
var skill_function_keys: Array = ["", "", "", "", "", "", "", ""]
var pending_skill_casts: Dictionary = {}
var _last_holy_shield_aura_pulse_key: String = ""
var item_rules: Dictionary:
	get: return ItemRulesLoader.item_rules
	set(v): ItemRulesLoader.item_rules = v
var item_templates: Dictionary:
	get: return ItemRulesLoader.item_templates
	set(v): ItemRulesLoader.item_templates = v
var item_presentations: Dictionary:
	get: return ItemRulesLoader.item_presentations
	set(v): ItemRulesLoader.item_presentations = v
var asset_manifest: Dictionary = {}
var dungeon_generation: Dictionary = {}
var loot_ids: Array = []
var hovered_loot_id: String = ""
var loot_label_reveal_held: bool = false
var monster_ids: Array = []
var interactable_ids: Array = []
var current_world_id: String = "vertical_slice"
var current_level: int = 0
var current_wall_layout: Array = []
var discovered_teleporters: Dictionary = {}
var pending_interactable_action: Dictionary = {}
var pending_action_targets: Dictionary = {}
var pending_waypoint_target_level: int = 0
var pending_waypoint_travel: bool = false
var _last_inventory_feedback_text: String = ""
var _last_boss_reward_status: String = ""
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
var visual_replay_completion_hold_s: float = 0.5
var client_settings: ClientSettings
var menu_layer: CanvasLayer
var main_menu: MainMenu
var character_panel: CharacterSelectPanel
var multiplayer_panel: MultiplayerSessionsPanel
var settings_panel: SettingsPanel
var pause_menu: PauseMenu
var loss_popup: Control
var gameplay_active: bool = false
var settings_return_target: String = "main"
var character_flow: String = CHARACTER_FLOW_CREATE_GAME
var pending_join_session_id: String = ""

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
var boss_health_bar: BossHealthBar
var walls_root: Node3D
var inventory_panel: InventoryPanel
var shop_panel: ShopPanel
var stash_panel: StashPanel
var bishop_panel: BishopPanel
var market_panel
var consumable_bar: ConsumableBar
var character_stats_panel: CharacterStatsPanel
var skills_panel: SkillsPanel
var character_bar: Control
var skill_bar: SkillBar
var status_effects_bar: StatusEffectsBar
var character_info_panel: PanelContainer
var character_info_name_label: Label
var character_info_level_label: Label
var character_info_area_label: Label
var input_shadow: InputShadowOverlay
var _health_bar: PlayerHealthBar

var _send_cooldown: float = 0.0
var _attack_cooldown: float = 0.0
var _sustained_click: SustainedClickInput = SustainedClickInputScript.new()
var _movement_requires_fresh_input: bool = false
var _player_walk_linger: float = 0.0
var _last_facing_direction := Vector2(1.0, 0.0)
var _debug_label: Label
var _level_label: Label
var _camera: Camera3D
var ground_node: MeshInstance3D
var ground_textures: Dictionary = {}
var wall_textures: Dictionary = {}

const SEND_INTERVAL := 0.1
const SERVER_TICK_RATE := 10.0
const DEFAULT_ATTACK_INTERVAL_TICKS := 20
const PLAYER_SPEED := 2.8
const WALK_ANIMATION_LINGER_SECONDS := 0.28
const CAMERA_ZOOM_DEFAULT := 20.0
const CAMERA_ZOOM_STEP := 1.5
const CAMERA_ZOOM_MIN := 8.0
const CAMERA_ZOOM_MAX := 40.0
const CAMERA_FOLLOW_OFFSET := Vector3(9.0, 20.0, 15.0)
const PROJECTILE_LERP_SECONDS := 0.10
const GROUND_TEXTURE_TOWN := "town_grass"
const GROUND_TEXTURE_DUNGEON := "dungeon_rock"
const WALL_TEXTURE_CAVE := "cave_wall"


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
	client_settings.set_language(client_settings.language, false)
	client_settings.apply()
	_refresh_localized_texts()
	_sync_status_text_visibility()
	_sync_settings_panel()
	ItemRulesLoader.ensure_loaded()
	SkillRulesLoader.ensure_loaded()
	ClassPresentationsLoaderScript.ensure_loaded()
	_load_dungeon_generation()
	_load_ground_item_visual_data()
	var base_url := _env("ARPG_BASE_URL", "http://localhost:8888")
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
	if main_menu != null:
		main_menu.set_account_email(client.account_email)
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
	var file_name := scenario_path.get_file()
	return file_name.begins_with("08_main_menu_flow") \
		or file_name.begins_with("20_menu_create_join_flow") \
		or file_name.begins_with("21_join_game_listed_session") \
		or file_name.begins_with("27_character_select_summaries")


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
	current_wall_layout = []
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
		main_menu.set_account_email(_account_email())
		main_menu.show_menu()


func _raise_gameplay_windows() -> void:
	for panel in [inventory_panel, shop_panel, stash_panel, bishop_panel, market_panel, character_stats_panel, skills_panel, character_info_panel]:
		if panel != null and panel is CanvasItem:
			(panel as CanvasItem).move_to_front()


func _account_email() -> String:
	if client == null:
		return ""
	return client.account_email


func _account_title(title: String) -> String:
	var email := _account_email()
	if email == "":
		return title
	return "%s - %s" % [title, email]


func _hide_all_menus() -> void:
	if main_menu != null:
		main_menu.visible = false
	if character_panel != null:
		character_panel.hide_panel()
	if multiplayer_panel != null:
		multiplayer_panel.hide_panel()
	if settings_panel != null:
		settings_panel.hide_panel()
	if pause_menu != null:
		pause_menu.hide_pause()
	if loss_popup != null:
		loss_popup.visible = false
	if character_stats_panel != null:
		character_stats_panel.hide_display()
	if skills_panel != null:
		skills_panel.hide_display()
	_hide_character_info_panel()


func _on_create_game_pressed() -> void:
	character_flow = CHARACTER_FLOW_CREATE_GAME
	pending_join_session_id = ""
	var characters := client.list_characters()
	if character_panel != null:
		if main_menu != null:
			main_menu.visible = false
		if characters.is_empty():
			character_panel.show_forced_create(_account_title("Create Character"))
		else:
			character_panel.show_choose_or_create(characters, _account_title("Choose Character"))


func _on_join_game_pressed() -> void:
	character_flow = CHARACTER_FLOW_JOIN_GAME
	pending_join_session_id = ""
	_show_join_game_panel(true)


func _on_continue_pressed() -> void:
	_on_create_game_pressed()


func _on_new_game_pressed() -> void:
	_on_create_game_pressed()


func _on_multiplayer_pressed() -> void:
	_on_join_game_pressed()


func _show_join_game_panel(refresh: bool = false) -> void:
	_hide_all_menus()
	gameplay_active = false
	if multiplayer_panel == null:
		return
	multiplayer_panel.show_panel()
	if refresh:
		_refresh_multiplayer_sessions()


func _show_multiplayer_panel(refresh: bool = false) -> void:
	_show_join_game_panel(refresh)


func _refresh_multiplayer_sessions() -> void:
	if multiplayer_panel == null:
		return
	multiplayer_panel.set_sessions(client.list_active_sessions())


func _on_host_listed_session_requested() -> void:
	character_flow = CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST
	pending_join_session_id = ""
	_show_character_picker_for_flow("Choose Character")


func _on_join_listed_session_requested(session_id: String) -> void:
	if session_id == "":
		return
	character_flow = CHARACTER_FLOW_JOIN_GAME
	pending_join_session_id = session_id
	_show_character_picker_for_flow("Choose Character")


func _on_character_panel_back() -> void:
	if character_panel != null:
		character_panel.hide_panel()
	match character_flow:
		CHARACTER_FLOW_JOIN_GAME, CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN, CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST:
			_show_join_game_panel(false)
		_:
			pending_join_session_id = ""
			if main_menu != null:
				main_menu.show_menu()


func _show_character_picker_for_flow(title: String = "Choose Character") -> void:
	if character_panel == null:
		return
	var characters := client.list_characters()
	if multiplayer_panel != null:
		multiplayer_panel.hide_panel()
	if main_menu != null:
		main_menu.visible = false
	if characters.is_empty():
		character_panel.show_forced_create(_account_title("Create Character"))
	else:
		character_panel.show_choose_or_create(characters, _account_title(title))


func _on_character_create_requested(name: String, character_class: String = "barbarian") -> void:
	var character := client.create_character(name, character_class)
	if character.is_empty():
		character_panel.set_error("Could not create character")
		return
	_start_selected_character(str(character.get("character_id", "")))


func _on_character_delete_requested(character_id: String) -> void:
	if not client.delete_character(character_id):
		if character_panel != null:
			character_panel.set_error("Could not delete character")
		return
	if character_panel != null:
		_refresh_character_panel_for_current_flow()


func _on_character_rename_requested(character_id: String, name: String) -> void:
	if client.rename_character(character_id, name).is_empty():
		if character_panel != null:
			character_panel.set_error("Could not rename character")
		return
	if character_panel != null:
		_refresh_character_panel_for_current_flow()


func _refresh_character_panel_for_current_flow() -> void:
	if character_panel == null:
		return
	var characters := client.list_characters()
	if characters.is_empty():
		character_panel.show_forced_create(_account_title("Create Character"))
	else:
		character_panel.show_choose_or_create(characters, _account_title("Choose Character"))


func _start_selected_character(character_id: String) -> void:
	match character_flow:
		CHARACTER_FLOW_CREATE_GAME:
			_start_create_game_session(character_id)
		CHARACTER_FLOW_JOIN_GAME:
			_start_listed_join_session(character_id)
		CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST:
			_start_listed_host_session(character_id)
		CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN:
			_start_listed_join_session(character_id)
		_:
			_start_character_session(character_id)


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


func _start_create_game_session(character_id: String) -> void:
	if client_settings != null and client_settings.create_game_session_type == ClientSettingsScript.CREATE_GAME_SESSION_TYPE_SOLO:
		_start_character_session(character_id)
	else:
		_start_listed_host_session(character_id)


func _start_listed_host_session(character_id: String) -> void:
	if character_id == "":
		character_panel.set_error("Could not host session")
		return
	_teardown_gameplay_state(false)
	if not client.create_listed_coop_session(character_id):
		character_panel.set_error("Could not host listed session")
		return
	bot_mode = false
	character_flow = CHARACTER_FLOW_CREATE_GAME
	_begin_gameplay_connection(false)


func _start_listed_join_session(character_id: String) -> void:
	if character_id == "" or pending_join_session_id == "":
		character_panel.set_error("Could not join session")
		return
	_teardown_gameplay_state(false)
	if not client.join_listed_session(pending_join_session_id, character_id):
		character_panel.set_error("Could not join listed session")
		return
	bot_mode = false
	character_flow = CHARACTER_FLOW_CREATE_GAME
	pending_join_session_id = ""
	_begin_gameplay_connection(false)


func _on_settings_from_main() -> void:
	settings_return_target = "main"
	main_menu.visible = false
	if settings_panel != null:
		settings_panel.show_settings(
			ClientSettingsScript.size_label(client_settings.window_size),
			client_settings.floating_combat_text,
			client_settings.status_text,
			client_settings.create_game_session_type,
			client_settings.language
		)


func _on_settings_from_pause() -> void:
	settings_return_target = "pause"
	if pause_menu != null:
		pause_menu.hide_pause()
	if settings_panel != null:
		settings_panel.show_settings(
			ClientSettingsScript.size_label(client_settings.window_size),
			client_settings.floating_combat_text,
			client_settings.status_text,
			client_settings.create_game_session_type,
			client_settings.language
		)


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


func _on_status_text_toggled(enabled: bool) -> void:
	if client_settings == null:
		return
	client_settings.set_status_text(enabled)
	_sync_settings_panel()
	_sync_status_text_visibility()


func _on_create_game_session_type_selected(session_type: String) -> void:
	if client_settings == null:
		return
	client_settings.set_create_game_session_type(session_type)
	_sync_settings_panel()


func _on_language_selected(language: String) -> void:
	if client_settings == null:
		return
	client_settings.set_language(language)
	_refresh_localized_texts()
	_sync_settings_panel()


func _sync_settings_panel() -> void:
	if settings_panel != null and client_settings != null:
		settings_panel.set_selected_size_label(ClientSettingsScript.size_label(client_settings.window_size))
		settings_panel.set_floating_combat_text_enabled(client_settings.floating_combat_text)
		settings_panel.set_status_text_enabled(client_settings.status_text)
		settings_panel.set_create_game_session_type(client_settings.create_game_session_type)
		settings_panel.set_language(client_settings.language)


func _refresh_localized_texts() -> void:
	if main_menu != null:
		main_menu.refresh_texts()
	if pause_menu != null:
		pause_menu.refresh_texts()
	if settings_panel != null:
		settings_panel.refresh_texts()


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
	player_mana = 10
	player_max_mana = 10
	predicted_pos = Vector3.ZERO
	reconciliation_delta = 0.0
	last_server_tick = 0
	inventory = []
	equipped = {}
	gold = 0
	pending_stash_equips.clear()
	character_progression = {}
	skill_progression = {}
	skill_cooldowns = []
	right_click_skill_id = ""
	skill_function_keys = ["", "", "", "", "", "", "", ""]
	if status_effects_bar != null:
		status_effects_bar.clear_effects()
	loot_ids.clear()
	monster_ids.clear()
	interactable_ids.clear()
	current_wall_layout = []
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
	_hide_boss_health_bar()
	if walls_root != null:
		_clear_wall_nodes()
	if resolver != null:
		resolver.apply_snapshot({"inventory": [], "equipped": {}})
	_refresh_inventory_ui()
	if _health_bar != null:
		_health_bar.update_hp(player_hp, player_max_hp)
		_health_bar.update_mana(player_mana, player_max_mana)
		_refresh_player_hud_identity()
	if character_stats_panel != null:
		character_stats_panel.hide_display()
	if skills_panel != null:
		skills_panel.hide_display()
	_hide_character_info_panel()
	_refresh_progression_ui()
	_refresh_skill_ui()
	_hide_waypoint_panel()
	if player_anchor != null:
		player_anchor.position = Vector3.ZERO
	if clear_session and client != null:
		client.session_id = ""
		client.seed = ""
		client.world_id = ""
		client.session_mode = ""
		client.session_listed = false
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
		_advance_boss_phase_display(delta)
	_sync_progression_interactivity()
	_try_complete_pending_interactable_action()
	_try_complete_pending_waypoint_travel()
	_tick_movement_animation_linger(delta)

	if visual_replay_enabled:
		_handle_visual_replay(delta)
	elif autoplay_enabled:
		_handle_autoplay(delta)
	else:
		_handle_input(delta)
	_sync_waypoint_panel_reach()
	_sync_actionable_panel_reach()
	if player_anim != null:
		player_anim.set_locomotion(_local_player_is_walking())
	if not _user_input_blocked():
		_update_facing_toward_mouse()
	_update_debug()


# --- message handling -------------------------------------------------------

func _handle_message(env: Dictionary) -> void:
	last_server_tick = max(last_server_tick, int(env.get("tick", 0)))
	var payload := _envelope_payload(env)
	match env.get("type", ""):
		"session_snapshot":
			_apply_snapshot(payload)
		"state_delta":
			_apply_delta(payload)
		"intent_accepted":
			pending_action_targets.erase(str(payload.get("accepted_message_id", "")))
		"intent_rejected":
			pending_interactable_action.clear()
			pending_waypoint_travel = false
			_handle_intent_rejected(payload)
			_debug("rejected: %s" % payload.get("reason", "?"))
		"error":
			_debug("error: %s" % payload.get("message", "?"))
		_:
			push_warning("_handle_message: unknown server message type '%s'" % env.get("type", ""))


func _envelope_payload(env: Dictionary) -> Dictionary:
	var payload = env.get("payload", {})
	if payload is Dictionary:
		return payload
	return {}


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


func _handle_intent_rejected(payload: Dictionary) -> void:
	var message_id := str(payload.get("rejected_message_id", ""))
	var reason := str(payload.get("reason", ""))
	var pending: Dictionary = pending_action_targets.get(message_id, {})
	var was_skill_cast := pending_skill_casts.has(message_id)
	if message_id != "":
		pending_action_targets.erase(message_id)
		pending_skill_casts.erase(message_id)
	if reason == "no_path" or reason == "path_too_long":
		_sustained_click.clear()
		pending_interactable_action.clear()
	if reason == "inventory_full":
		var target_id := str(pending.get("target_id", ""))
		_show_inventory_full_text(target_id)
	elif reason == "capacity_would_overflow":
		_show_bag_full_cant_unequip_text()
	elif was_skill_cast or _is_skill_reject_reason(reason):
		_show_skill_rejected_feedback(reason)
	elif shop_panel != null and shop_panel.visible:
		shop_panel.show_status(reason.replace("_", " "), true)
	elif stash_panel != null and stash_panel.visible:
		stash_panel.show_status(reason.replace("_", " "), true)
	elif bishop_panel != null and bishop_panel.visible:
		bishop_panel.show_status(reason.replace("_", " "), true)


func _apply_snapshot(p: Dictionary) -> void:
	current_level = int(p.get("current_level", 0))
	_update_ground_material()
	var local_id := _snapshot_local_player_id(p)
	if local_id != "":
		player_id = local_id
	party = p.get("party", [])
	pending_interactable_action.clear()
	pending_waypoint_travel = false
	_apply_teleporter_snapshot(p.get("discovered_teleporters", []))
	_clear_level_entities()
	var snapshot_walls = p.get("walls", null)
	if typeof(snapshot_walls) == TYPE_ARRAY:
		_render_wall_layout(snapshot_walls as Array)
	else:
		_render_world_walls(current_world_id)
	_update_level_hud()
	_refresh_waypoint_panel()
	# (player is the PlayerAnchor/CharacterVisual, not a per-snapshot entity node)
	for e in p.get("entities", []):
		_upsert_entity(e)
	_sync_boss_health_bar()
	inventory = p.get("inventory", [])
	equipped = p.get("equipped", {})
	inventory_rows = int(p.get("inventory_rows", inventory_rows))
	inventory_capacity = int(p.get("inventory_capacity", inventory_capacity))
	gold = int(p.get("gold", gold))
	stash_items = p.get("stash_items", [])
	stash_gold = int(p.get("stash_gold", stash_gold))
	stash_capacity = int(p.get("stash_capacity", stash_capacity))
	hotbar_capacity = int(p.get("hotbar_capacity", 2))
	hotbar = p.get("hotbar", [])
	character_progression = p.get("character_progression", {})
	skill_progression = p.get("skill_progression", {})
	skill_cooldowns = p.get("skill_cooldowns", [])
	_apply_skill_bindings(p.get("skill_bindings", {}))
	_apply_local_player_class_model()
	_refresh_player_hud_identity()
	if resolver != null:
		resolver.apply_snapshot(p)
	_refresh_inventory_ui()
	_refresh_progression_ui()
	_refresh_skill_ui()
	_update_character_info_panel()
	_refresh_market_board_summary()
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
			_update_ground_material()
			pending_interactable_action.clear()
			pending_waypoint_travel = false
			_clear_level_entities()
			current_wall_layout = []
			_clear_wall_nodes()
			_update_level_hud()
			_update_character_info_panel()
			_hide_waypoint_panel()
			_hide_shop_panel()
			_hide_stash_panel()
			_hide_bishop_panel()
			_hide_market_panel()
	var changes: Array = p.get("changes", [])
	for c in changes:
		match c.get("op", ""):
			"wall_layout_update":
				_render_wall_layout(c.get("walls", []))
			"entity_spawn", "entity_update":
				_upsert_entity(c.get("entity", {}))
			"entity_remove":
				_remove_entity(str(c.get("entity_id", "")))
			"inventory_add":
				var inv_item: Dictionary = c.get("item", {})
				inventory.append(inv_item)
				if resolver != null:
					resolver.ingest_inventory_item(inv_item)
			"inventory_update":
				var upd_item: Dictionary = c.get("item", {})
				_update_inventory_item(upd_item)
				if resolver != null:
					resolver.ingest_inventory_item(upd_item)
			"inventory_remove":
				_remove_inventory_item(str(c.get("item_instance_id", "")))
			"equipped_update":
				equipped[c["slot"]] = c.get("item_instance_id")
				if resolver != null:
					resolver.apply_equipped_update(c["slot"], c.get("item_instance_id"))
				if c.has("inventory_rows"):
					inventory_rows = int(c.get("inventory_rows", inventory_rows))
				if c.has("inventory_capacity"):
					inventory_capacity = int(c.get("inventory_capacity", inventory_capacity))
				if c.has("hotbar_capacity"):
					hotbar_capacity = int(c.get("hotbar_capacity", hotbar_capacity))
					if consumable_bar != null:
						consumable_bar.set_hotbar_state(hotbar_capacity, hotbar)
			"hotbar_update":
				if c.has("inventory_rows"):
					inventory_rows = int(c.get("inventory_rows", inventory_rows))
				if c.has("inventory_capacity"):
					inventory_capacity = int(c.get("inventory_capacity", inventory_capacity))
				_apply_hotbar_update(int(c.get("slot_index", -1)), c.get("item_instance_id"), c.get("item", {}))
			"gold_update":
				gold = int(c.get("gold", gold))
			"stash_item_add":
				_upsert_stash_item(c.get("item", {}))
			"stash_item_remove":
				_remove_stash_item(str(c.get("stash_item_id", "")))
			"stash_gold_update":
				stash_gold = int(c.get("stash_gold", stash_gold))
			"teleporter_discovery_update":
				var discovered_level := int(c.get("level", 0))
				var discovered := bool(c.get("discovered", false))
				discovered_teleporters[discovered_level] = discovered
				_refresh_waypoint_panel()
				if discovered and discovered_level == current_level:
					_show_waypoint_panel()
			"character_progression_update":
				character_progression = c.get("character_progression", {})
				_apply_local_player_class_model()
				_refresh_progression_ui()
				_refresh_player_hud_identity()
				_update_character_info_panel()
			"skill_progression_update":
				skill_progression = c.get("skill_progression", {})
				_refresh_skill_ui()
			"skill_cooldown_update":
				skill_cooldowns = c.get("skill_cooldowns", [])
				_refresh_skill_ui()
			"skill_bindings_update":
				_apply_skill_bindings(c.get("skill_bindings", {}))
				_refresh_skill_ui()
			_:
				pass
	_refresh_inventory_ui()
	var heal_cast_rain_correlations := _heal_cast_rain_correlations(p.get("events", []))
	for ev in p.get("events", []):
		var eid := _event_subject_entity_id(ev)
		var event_type := str(ev.get("event_type", ""))
		if event_type == "skill_cast" and eid == player_id:
			if skill_bar != null:
				skill_bar.flash_cast()
			if player_anim != null:
				player_anim.play_one_shot("attack")
			if ev.has("angle_degrees") and ev.has("range") and ev.has("direction"):
				_spawn_skill_cone(ev)
			var correlation_id := str(ev.get("correlation_id", ""))
			var is_heal_cast := str(ev.get("skill_id", "")) == "heal"
			if is_heal_cast and (heal_cast_rain_correlations.has(correlation_id) or heal_cast_rain_correlations.has("__uncorrelated__")):
				var heal_target_id := str(ev.get("target_entity_id", ""))
				if heal_target_id == "":
					heal_target_id = eid
				_spawn_heal_rain(heal_target_id)
			continue
		if event_type == "skill_chain_hit":
			_spawn_ligthing_chain(ev)
			continue
		if event_type == "skill_cooldown_rejected" and eid == player_id:
			_show_skill_rejected_feedback(str(ev.get("reason", "")))
			continue
		if event_type == "player_healed":
			_show_damage_number(eid, Color(0.3, 1.0, 0.45), ev.get("heal", null), "+", 1.0)
			_spawn_heal_rain(eid)
			if eid == player_id and _health_bar != null:
				_health_bar.update_hp(player_hp, player_max_hp, true)
			continue
		if event_type == "skill_effect_started" and str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID:
			_pulse_holy_shield_aura(ev)
		if str(ev.get("skill_id", "")) == "poison_stab":
			if event_type == "skill_effect_started":
				_set_entity_poison_tint(eid, true)
				continue
			if event_type == "skill_effect_ended":
				_set_entity_poison_tint(eid, false)
				continue
		if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.BURNING_EFFECT_ID:
			if event_type == "skill_effect_started":
				_set_entity_burning(eid, true)
				continue
			if event_type == "skill_effect_ended":
				_set_entity_burning(eid, false)
				continue
		if event_type == "skill_effect_started" and eid == player_id:
			if status_effects_bar != null:
				status_effects_bar.start_effect(ev)
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.RAGE_EFFECT_ID:
				_apply_local_player_visual_scale(1.0 + float(ev.get("amount", 0)) / 100.0)
				PlayerStatusEffectMarkers.sync_rage_effect(player_anchor, true)
			continue
		if event_type == "skill_effect_ended" and eid == player_id:
			if status_effects_bar != null:
				status_effects_bar.end_effect(str(ev.get("skill_id", "")))
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.RAGE_EFFECT_ID:
				_apply_local_player_visual_scale(1.0)
				PlayerStatusEffectMarkers.sync_rage_effect(player_anchor, false)
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID:
				PlayerStatusEffectMarkers.sync_holy_shield_effect(player_anchor, [])
			continue
		if visual_replay_enabled and inventory_panel != null:
			var hint: Variant = INVENTORY_REPLAY_EVENT_HINTS.get(event_type, null)
			if hint != null:
				inventory_panel.show_gesture_hint(str(hint))
		if eid == player_id:
			if event_type == "player_mana_restored":
				_show_damage_number(eid, Color("#54c7f3"), ev.get("mana", null), "+", 1.0)
				if _health_bar != null:
					_health_bar.update_mana(player_mana, player_max_mana, true)
				continue
			if event_type == "player_damaged":
				_show_combat_text_for_event(eid, ev, Color(1.0, 0.32, 0.2))
				_play_entity_reaction(eid, ev, "hit")
				if _health_bar != null:
					_health_bar.update_hp(player_hp, player_max_hp)
			if event_type == "player_killed":
				_play_entity_reaction(eid, ev, "death")
				_show_loss_popup()
			if event_type == "attack_missed":
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			var player_clip = PLAYER_EVENT_CLIPS.get(event_type, null)
			if player_clip == null or player_anim == null:
				continue
			if player_clip == "death":
				player_anim.enter_terminal("death")
			else:
				_play_local_player_reaction_animation(player_clip)
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
		if event_type == "shop_opened":
			_show_shop_panel(ev)
			continue
		if event_type == "shop_purchase" and shop_panel != null and shop_panel.visible:
			_apply_shop_event_refresh(ev)
			shop_panel.show_status("Bought for %d" % int(ev.get("price", 0)))
			continue
		if event_type == "shop_sale" and shop_panel != null and shop_panel.visible:
			_apply_shop_event_refresh(ev)
			shop_panel.show_status("Sold for %d" % int(ev.get("price", 0)))
			continue
		if event_type == "shop_reroll" and shop_panel != null and shop_panel.visible:
			_apply_shop_event_refresh(ev)
			shop_panel.show_status("Rerolled for %d" % int(ev.get("price", 0)))
			continue
		if event_type == "stash_opened":
			_show_stash_panel(ev)
			continue
		if event_type == "stash_item_deposited" and stash_panel != null and stash_panel.visible:
			stash_panel.show_status("Item stored")
			continue
		if event_type == "stash_item_withdrawn":
			_handle_stash_item_withdrawn(ev)
			if stash_panel != null and stash_panel.visible:
				stash_panel.show_status("Item withdrawn")
			continue
		if event_type == "stash_gold_deposited" and stash_panel != null and stash_panel.visible:
			stash_panel.show_status("Stored %d gold" % int(ev.get("amount", 0)))
			continue
		if event_type == "stash_gold_withdrawn" and stash_panel != null and stash_panel.visible:
			stash_panel.show_status("Withdrew %d gold" % int(ev.get("amount", 0)))
			continue
		if event_type == "bishop_service_opened":
			_show_bishop_panel(ev)
			continue
		if event_type == "market_service_opened":
			_show_market_panel(ev)
			continue
		if event_type == "bishop_respec" and bishop_panel != null and bishop_panel.visible:
			bishop_panel.set_gold(gold)
			bishop_panel.show_status("Respec complete")
			continue
		if event_type == "boss_killed":
			_last_boss_reward_status = "%s defeated" % _boss_health_bar_title(str(ev.get("boss_template_id", "")))
			continue
		if event_type == "boss_phase_started" and entities.has(eid):
			_apply_boss_phase_started(eid, ev)
			continue
		if event_type == "boss_phase_ended" and entities.has(eid):
			_apply_boss_phase_ended(eid, ev)
			continue
		if event_type == "monster_aggro":
			_show_damage_number(eid, Color("#80ff8f"), null, "", 0.0, "threat", "AGGRO")
			continue
		var clip = MONSTER_EVENT_CLIPS.get(event_type, null)
		if clip == null:
			if event_type in ["attack_missed", "attack_blocked"]:
				if str(ev.get("source_entity_id", "")) == player_id:
					_play_local_attack_animation_for_event(ev)
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			continue
		if event_type == "monster_damaged" or event_type == "monster_killed":
			if str(ev.get("source_entity_id", "")) == player_id:
				_play_local_attack_animation_for_event(ev)
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
	_sync_boss_health_bar()
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
			if player_hp <= 0:
				if player_anim != null:
					player_anim.enter_terminal("death")
				if player_reaction != null:
					player_reaction.enter_death()
				_show_loss_popup()
			else:
				if player_anim != null and player_anim.is_terminal():
					player_anim.reset_terminal()
				if player_reaction != null and player_reaction.is_terminal():
					player_reaction.reset_terminal()
		if e.has("mana"):
			player_mana = int(e["mana"])
			if e.has("max_mana"):
				player_max_mana = int(e["max_mana"])
			if _health_bar != null:
				_health_bar.update_mana(player_mana, player_max_mana)
		if e.has("visual_scale"):
			_apply_local_player_visual_scale(float(e["visual_scale"]))
		if e.has("effect_ids"):
			PlayerStatusEffectMarkers.sync_holy_shield_effect(player_anchor, e.get("effect_ids", []))
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		var prev_predicted_pos := predicted_pos
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		player_anchor.position = server_pos
		if prev_predicted_pos.distance_to(server_pos) > 0.001 and player_hp > 0:
			_mark_local_player_walking()
			_face_direction(Vector2(server_pos.x - prev_predicted_pos.x, server_pos.z - prev_predicted_pos.z))
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
		if e.has("amount"):
			rec["amount"] = int(e["amount"])
		if e.has("monster_def_id"):
			rec["monster_def_id"] = str(e["monster_def_id"])
		for key in ["item_template_id", "display_name", "rarity", "rolled_stats", "requirements", "requirement_status", "requirements_met", "equip_preview", "effect_ids", "character_id", "boss_template_id", "visual_model", "visual_tint", "boss_phase"]:
			if e.has(key):
				rec[key] = e[key]
		if e.has("is_boss"):
			rec["is_boss"] = bool(e["is_boss"])
		if e.has("visual_scale"):
			rec["visual_scale"] = float(e["visual_scale"])
		if e.has("interactable_def_id"):
			rec["interactable_def_id"] = str(e["interactable_def_id"])
		entities[id] = rec
		if e["type"] != "projectile" and e["type"] != "player":
			_attach_pick_collider(node, id, str(e["type"]))
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
			_set_loot_label_visible(id, loot_label_reveal_held or id == hovered_loot_id, id == hovered_loot_id)
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
		if (rec["type"] == "monster" or rec["type"] == "player") and rec["controller"] != null and not is_new:
			var hp_val := int(e.get("hp", rec.get("hp", 1)))
			var moved := prev_pos.distance_to(server_pos) > 0.001
			if moved and hp_val > 0:
				rec["walk_linger"] = WALK_ANIMATION_LINGER_SECONDS
				_face_entity_direction(node, Vector2(server_pos.x - prev_pos.x, server_pos.z - prev_pos.z))
			rec["controller"].set_locomotion(float(rec.get("walk_linger", 0.0)) > 0.0 and hp_val > 0)
		if rec["type"] == "player":
			rec["hp"] = int(e.get("hp", rec.get("hp", PLAYER_START_HP)))
			rec["max_hp"] = int(e.get("max_hp", rec.get("max_hp", PLAYER_START_HP)))
			if int(rec["hp"]) <= 0:
				_enter_entity_terminal_death(id, rec)
	if rec["type"] == "interactable":
		var state := str(e.get("state", rec.get("state", "closed")))
		_set_interactable_state(id, rec, state)
	if rec["type"] == "monster" or rec["type"] == "player":
		_apply_entity_visual_metadata(rec, e)
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
	if hovered_loot_id == id:
		hovered_loot_id = ""
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
	_sync_boss_health_bar()


func _clear_level_entities() -> void:
	for id in entities.keys():
		(entities[id]["node"] as Node3D).queue_free()
	entities.clear()
	for id in monster_health_bars.keys():
		var bar = monster_health_bars[id]
		if is_instance_valid(bar):
			bar.queue_free()
	monster_health_bars.clear()
	_hide_boss_health_bar()
	loot_ids.clear()
	hovered_loot_id = ""
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


func _upsert_stash_item(item: Dictionary) -> void:
	var stash_item_id := str(item.get("stash_item_id", ""))
	if stash_item_id == "":
		return
	for i in range(stash_items.size()):
		if str(stash_items[i].get("stash_item_id", "")) == stash_item_id:
			var merged: Dictionary = stash_items[i].duplicate(true)
			merged.merge(item, true)
			stash_items[i] = merged
			return
	stash_items.append(item.duplicate(true))


func _remove_stash_item(stash_item_id: String) -> void:
	for i in range(stash_items.size() - 1, -1, -1):
		if str(stash_items[i].get("stash_item_id", "")) == stash_item_id:
			stash_items.remove_at(i)


func _apply_hotbar_update(slot_index: int, item_instance_id, item: Dictionary = {}) -> void:
	if slot_index < 0 or slot_index >= 10:
		return
	while hotbar.size() < 10:
		hotbar.append({"slot_index": hotbar.size(), "item_instance_id": null})
	var slot := {"slot_index": slot_index, "item_instance_id": item_instance_id}
	if not item.is_empty():
		slot["item"] = item.duplicate(true)
	hotbar[slot_index] = slot
	if consumable_bar != null:
		consumable_bar.apply_hotbar_update(slot_index, item_instance_id, item)


func _refresh_inventory_ui() -> void:
	if inventory_panel != null:
		inventory_panel.set_inventory_state(inventory, equipped, inventory_rows, inventory_capacity, gold, hotbar, hotbar_capacity)
	if shop_panel != null and shop_panel.visible:
		shop_panel.set_inventory_state(inventory, equipped, gold)
	if stash_panel != null and stash_panel.visible:
		stash_panel.set_stash_state(stash_items, stash_gold, stash_capacity)
		stash_panel.set_inventory_state(inventory, equipped, gold, hotbar)
	if bishop_panel != null and bishop_panel.visible:
		bishop_panel.set_gold(gold)
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
	var skill_id := str(ev.get("skill_id", ""))
	if outcome == "miss":
		_show_damage_number(entity_id, Color(0.82, 0.86, 0.92), null, "", 0.0, "miss", "MISS")
		return
	if outcome == "block":
		_show_damage_number(entity_id, Color(0.35, 0.78, 1.0), null, "", 0.0, "block", "BLOCK")
		return
	if skill_id == "poison_stab":
		_show_damage_number(entity_id, Color("#55e66f"), damage, "", 0.0, "poison")
		return
	if outcome == "crit" or bool(ev.get("critical", false)):
		var crit_damage := 0 if damage == null else int(damage)
		_show_damage_number(entity_id, Color(1.0, 0.58, 0.22), crit_damage, "", 0.0, "crit", "%d!" % crit_damage)
		return
	_show_damage_number(entity_id, default_color, damage)


func _play_local_attack_animation_for_event(ev: Dictionary) -> void:
	if player_anim == null:
		return
	var weapon_slot := str(ev.get("weapon_slot", "main_hand"))
	player_anim.play_one_shot("attack_off_hand" if weapon_slot == "off_hand" else "attack")


func _play_local_player_reaction_animation(clip: String) -> void:
	if player_anim == null:
		return
	if clip == "hit" and player_anim.current_clip() in ["attack", "attack_off_hand"]:
		return
	player_anim.play_one_shot(clip)


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


func _pulse_holy_shield_aura(ev: Dictionary) -> void:
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "":
		source_id = str(ev.get("entity_id", ""))
	var source_root := _entity_root_for_id(source_id)
	if source_root == null:
		return
	var pulse_key := str(ev.get("correlation_id", ""))
	if pulse_key == "":
		pulse_key = "%s:%s:%d" % [source_id, str(ev.get("skill_id", "")), last_server_tick]
	if pulse_key == _last_holy_shield_aura_pulse_key:
		return
	_last_holy_shield_aura_pulse_key = pulse_key
	var radius := _holy_shield_aura_radius()
	PlayerStatusEffectMarkers.pulse_holy_shield_aura(source_root, _entity_roots_in_radius(source_root, radius), radius)


func _entity_root_for_id(entity_id: String) -> Node3D:
	if entity_id == player_id:
		return player_anchor
	if entities.has(entity_id):
		var rec: Dictionary = entities[entity_id]
		return rec.get("node", null) as Node3D
	return null


func _entity_roots_in_radius(center_root: Node3D, radius: float) -> Array:
	var roots: Array = []
	var center := _node_world_or_local_position(center_root)
	if player_anchor != null and _flat_distance(center, _node_world_or_local_position(player_anchor)) <= radius:
		roots.append(player_anchor)
	for entity_id in entities.keys():
		var rec: Dictionary = entities[entity_id]
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		if _flat_distance(center, _node_world_or_local_position(node)) <= radius:
			roots.append(node)
	return roots


func _holy_shield_aura_radius() -> float:
	var def := SkillRulesLoader.skill_definition(PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID)
	for effect in def.get("effects", []):
		var row: Dictionary = effect
		if str(row.get("type", "")) == "area_stat_percent_buff":
			return maxf(float(row.get("radius", 5.0)), 0.5)
	return 5.0


func _on_status_effect_expired(skill_id: String) -> void:
	if skill_id == PlayerStatusEffectMarkers.RAGE_EFFECT_ID:
		_apply_local_player_visual_scale(1.0)
		PlayerStatusEffectMarkers.sync_rage_effect(player_anchor, false)
	if skill_id == PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID:
		PlayerStatusEffectMarkers.sync_holy_shield_effect(player_anchor, [])


func _flat_distance(a: Vector3, b: Vector3) -> float:
	return Vector2(a.x - b.x, a.z - b.z).length()


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
		if player_anchor == null:
			return
		target = player_anchor
		world_position = _node_world_or_local_position(player_anchor)
	elif entities.has(entity_id):
		target = entities[entity_id]["node"] as Node3D
		world_position = _node_world_or_local_position(target)
	else:
		return
	var pop := DamageNumberScript.new() as DamageNumber
	damage_numbers_layer.add_child(pop)
	var side := side_override if side_override != 0.0 else (-1.0 if entity_id == player_id else 1.0)
	pop.setup(_camera, target, world_position, amount, color, side, prefix, variant, text_override)


func _spawn_heal_rain(entity_id: String) -> void:
	var target := _node_for_entity_id(entity_id)
	if target == null:
		return
	_spawn_heal_rain_at_position(_node_world_or_local_position(target))


func _heal_cast_rain_correlations(events: Array) -> Dictionary:
	var heal_casts := {}
	var healed_casts := {}
	var has_uncorrelated_heal_cast := false
	var has_heal_result := false
	for raw in events:
		if not (raw is Dictionary):
			continue
		var ev := raw as Dictionary
		var correlation_id := str(ev.get("correlation_id", ""))
		var event_type := str(ev.get("event_type", ""))
		if event_type == "skill_cast" and str(ev.get("skill_id", "")) == "heal":
			if correlation_id == "":
				has_uncorrelated_heal_cast = true
			else:
				heal_casts[correlation_id] = true
		elif event_type == "player_healed" and str(ev.get("skill_id", "")) == "heal":
			has_heal_result = true
			if correlation_id != "":
				healed_casts[correlation_id] = true
	for correlation_id in healed_casts.keys():
		heal_casts.erase(correlation_id)
	if has_uncorrelated_heal_cast and not has_heal_result:
		heal_casts["__uncorrelated__"] = true
	return heal_casts


func _spawn_heal_rain_at_position(world_position: Vector3) -> void:
	var effect := HealRainEffectScript.new() as HealRainEffect
	effect.setup(HEAL_RAIN_RADIUS)
	effect.position = world_position
	add_child(effect)


func _spawn_skill_cone(ev: Dictionary) -> void:
	var pos := _vec2_from_dict(ev.get("position", {}))
	var dir := _vec2_from_dict(ev.get("direction", {}))
	if dir.length_squared() <= 0.0001:
		return
	dir = dir.normalized()
	var radius := maxf(float(ev.get("range", 0.0)), 0.1)
	var angle := deg_to_rad(clampf(float(ev.get("angle_degrees", 0.0)), 1.0, 360.0))
	var segments := 18
	var points := PackedVector3Array()
	points.append(Vector3.ZERO)
	var base_angle := atan2(dir.y, dir.x)
	for i in range(segments + 1):
		var t := -angle * 0.5 + angle * float(i) / float(segments)
		var a := base_angle + t
		points.append(Vector3(cos(a) * radius, 0.0, sin(a) * radius))
	var vertices := PackedVector3Array()
	for i in range(1, points.size() - 1):
		vertices.append(points[0])
		vertices.append(points[i])
		vertices.append(points[i + 1])
	var arrays := []
	arrays.resize(Mesh.ARRAY_MAX)
	arrays[Mesh.ARRAY_VERTEX] = vertices
	var mesh := ArrayMesh.new()
	mesh.add_surface_from_arrays(Mesh.PRIMITIVE_TRIANGLES, arrays)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.08, 0.03, 0.38)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.04, 0.02)
	mat.emission_energy_multiplier = 1.4
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	var wedge := MeshInstance3D.new()
	wedge.name = "CleaveCone"
	wedge.mesh = mesh
	wedge.material_override = mat
	wedge.position = Vector3(pos.x, 0.08, pos.y)
	add_child(wedge)
	var tween := wedge.create_tween()
	tween.tween_property(wedge, "scale", Vector3.ONE * 1.03, 0.16)
	tween.tween_callback(wedge.queue_free)


func _spawn_ligthing_chain(ev: Dictionary) -> void:
	var source := _node_for_entity_id(str(ev.get("source_entity_id", "")))
	var target := _node_for_entity_id(str(ev.get("target_entity_id", "")))
	if source == null or target == null:
		return
	var start := _node_world_or_local_position(source)
	var finish := _node_world_or_local_position(target)
	var delta := finish - start
	var length := delta.length()
	if length <= 0.05:
		return
	var root := Node3D.new()
	root.name = "LigthingChain"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.07, 0.07, length)
	var bolt := MeshInstance3D.new()
	bolt.mesh = mesh
	bolt.position = Vector3(0.0, 0.55, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.94, 0.28, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(1.0, 0.88, 0.2)
	mat.emission_energy_multiplier = 2.4
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	bolt.material_override = mat
	root.add_child(bolt)
	root.position = (start + finish) * 0.5
	root.position.y = 0.15
	add_child(root)
	root.look_at(Vector3(finish.x, root.position.y, finish.z), Vector3.UP)
	var tween := root.create_tween()
	tween.tween_property(root, "scale", Vector3(1.25, 1.25, 1.0), 0.5)
	tween.tween_callback(root.queue_free)


func _vec2_from_dict(value) -> Vector2:
	if value is Dictionary:
		return Vector2(float(value.get("x", 0.0)), float(value.get("y", 0.0)))
	return Vector2.ZERO


func _node_for_entity_id(entity_id: String) -> Node3D:
	if entity_id == player_id:
		return player_anchor
	if entities.has(entity_id):
		return entities[entity_id].get("node", null) as Node3D
	return null


func _show_inventory_full_text(target_id: String) -> void:
	if damage_numbers_layer == null or _camera == null:
		return
	if target_id == "" or not entities.has(target_id):
		return
	var rec: Dictionary = entities[target_id]
	if str(rec.get("type", "")) != "loot":
		return
	var target := rec.get("node", null) as Node3D
	if target == null:
		return
	var pop := DamageNumberScript.new() as DamageNumber
	damage_numbers_layer.add_child(pop)
	pop.setup(_camera, target, target.global_position, null, Color("#ffcf5a"), 0.0, "", "inventory", "BAG FULL")


func _show_bag_full_cant_unequip_text() -> void:
	_last_inventory_feedback_text = BAG_FULL_CANT_UNEQUIP_TEXT
	if inventory_panel != null:
		inventory_panel.show_gesture_hint(BAG_FULL_CANT_UNEQUIP_TEXT)
	_show_damage_number(player_id, Color("#ffcf5a"), null, "", 0.0, "inventory", BAG_FULL_CANT_UNEQUIP_TEXT)


func _send_action_intent(target_id: String) -> void:
	if client == null or target_id == "":
		return
	var message_id := client.send("action_intent", last_server_tick, {"target_id": target_id})
	pending_action_targets[message_id] = {"target_id": target_id}


func _basic_attack_cooldown_seconds() -> float:
	var ticks := DEFAULT_ATTACK_INTERVAL_TICKS
	var derived = character_progression.get("derived_stats", {})
	if typeof(derived) == TYPE_DICTIONARY:
		ticks = int(derived.get("attack_interval_ticks", DEFAULT_ATTACK_INTERVAL_TICKS))
	if ticks <= 0:
		ticks = DEFAULT_ATTACK_INTERVAL_TICKS
	return maxf(SEND_INTERVAL, float(ticks) / SERVER_TICK_RATE)


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


func _hide_boss_health_bar() -> void:
	if boss_health_bar != null:
		boss_health_bar.hide_boss()


func _sync_boss_health_bar() -> void:
	if boss_health_bar == null:
		return
	var boss_id := _active_boss_entity_id()
	if boss_id == "":
		boss_health_bar.hide_boss()
		return
	var rec: Dictionary = entities[boss_id]
	var hp := int(rec.get("hp", 0))
	var max_hp := int(rec.get("max_hp", hp))
	var template_id := str(rec.get("boss_template_id", ""))
	boss_health_bar.show_boss(boss_id, template_id, _boss_health_bar_title(template_id), hp, max_hp)
	var phase := _boss_phase_for_display(rec)
	if phase.is_empty():
		boss_health_bar.clear_phase_state()
	else:
		boss_health_bar.set_phase_state(phase)


func _advance_boss_phase_display(delta: float) -> void:
	var changed := false
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var phase := _boss_phase_for_display(rec)
		if phase.is_empty():
			continue
		var remaining := float(phase.get("remaining_ticks_float", float(phase.get("remaining_ticks", 0))))
		remaining = maxf(0.0, remaining - maxf(0.0, delta) * BOSS_PHASE_TICK_RATE)
		phase["remaining_ticks_float"] = remaining
		phase["remaining_ticks"] = int(ceil(remaining))
		rec["boss_phase"] = phase
		changed = true
	if changed:
		_sync_boss_health_bar()


func _boss_phase_for_display(rec: Dictionary) -> Dictionary:
	var raw = rec.get("boss_phase", {})
	if typeof(raw) != TYPE_DICTIONARY:
		return {}
	var phase: Dictionary = (raw as Dictionary).duplicate(true)
	var kind := str(phase.get("phase_kind", ""))
	if kind == "":
		return {}
	var duration := maxi(0, int(phase.get("duration_ticks", 0)))
	phase["duration_ticks"] = duration
	var remaining := int(phase.get("remaining_ticks", -1))
	if remaining < 0:
		var started_tick := int(phase.get("started_tick", last_server_tick))
		remaining = max(0, duration - max(0, last_server_tick - started_tick))
	phase["remaining_ticks"] = clampi(remaining, 0, duration)
	phase["remaining_ticks_float"] = clampf(float(phase.get("remaining_ticks_float", phase["remaining_ticks"])), 0.0, float(duration))
	rec["boss_phase"] = phase
	return phase


func _active_boss_entity_id() -> String:
	var candidates: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) != "monster":
			continue
		if not bool(rec.get("is_boss", false)):
			continue
		if int(rec.get("hp", 0)) <= 0:
			continue
		candidates.append(str(id))
	candidates.sort()
	return str(candidates[0]) if not candidates.is_empty() else ""


func _boss_health_bar_title(template_id: String) -> String:
	if template_id == "":
		return "Boss"
	var pieces := template_id.replace("-", "_").split("_", false)
	var words := PackedStringArray()
	for piece in pieces:
		var word := str(piece)
		if word == "":
			continue
		words.append(word.substr(0, 1).to_upper() + word.substr(1).to_lower())
	return " ".join(words) if words.size() > 0 else template_id


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
	if event is InputEventKey and not event.echo and _is_force_stand_key(event):
		if event.pressed:
			_begin_force_stand()
		get_viewport().set_input_as_handled()
		return
	if event is InputEventKey and event.pressed and not event.echo:
		var skill_key_slot := _skill_function_key_slot(event)
		if skill_key_slot >= 0:
			if _handle_skill_function_key(skill_key_slot):
				get_viewport().set_input_as_handled()
				return
		var hotbar_slot := _hotbar_slot_for_key(event)
		if hotbar_slot >= 0:
			if consumable_bar != null:
				consumable_bar.use_slot(hotbar_slot)
			get_viewport().set_input_as_handled()
			return
		if _is_inventory_key(event):
			if inventory_panel != null:
				_close_gameplay_panels("inventory")
				inventory_panel.toggle()
				_raise_gameplay_windows()
			get_viewport().set_input_as_handled()
			return
		if _is_character_stats_key(event):
			if character_stats_panel != null:
				_close_gameplay_panels("stats")
				character_stats_panel.toggle()
				_refresh_progression_ui()
				_raise_gameplay_windows()
			get_viewport().set_input_as_handled()
			return
		if _is_skills_key(event):
			if skills_panel != null:
				_close_gameplay_panels("skills")
				skills_panel.toggle()
				_refresh_skill_ui()
				_raise_gameplay_windows()
			get_viewport().set_input_as_handled()
			return
		if _is_character_info_key(event):
			_toggle_character_info_panel()
			get_viewport().set_input_as_handled()
			return
		if _is_skill_slot_key(event):
			if skill_bar != null:
				skill_bar.use_slot()
			get_viewport().set_input_as_handled()
			return
	if event is InputEventMouseButton and event.pressed:
		match event.button_index:
			MOUSE_BUTTON_LEFT:
				if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN and player_hp > 0:
					if _is_force_stand_held():
						_start_directional_attack_hold()
						get_viewport().set_input_as_handled()
						return
					var pick := _resolve_click_at_mouse()
					_sustained_click.begin_from_pick(pick)
					_execute_click_pick(pick)
			MOUSE_BUTTON_RIGHT:
				if _try_use_right_click_skill():
					get_viewport().set_input_as_handled()
					return
			MOUSE_BUTTON_WHEEL_UP:
				_adjust_camera_zoom(-CAMERA_ZOOM_STEP)
			MOUSE_BUTTON_WHEEL_DOWN:
				_adjust_camera_zoom(CAMERA_ZOOM_STEP)


func _handle_input(delta: float) -> void:
	_update_loot_hover_label()
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
	if _is_force_stand_held():
		_movement_requires_fresh_input = true
		input = Vector2.ZERO
	elif input == Vector2.ZERO:
		_movement_requires_fresh_input = false

	if input != Vector2.ZERO and not _movement_requires_fresh_input and _send_cooldown <= 0.0:
		var dir := _camera_relative_flat_direction(input)
		_close_gameplay_panels_for_movement()
		# Local prediction: move immediately for responsive feel.
		predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
		_reconcile_player()
		_mark_local_player_walking()
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


func _is_character_info_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_P or event.physical_keycode == KEY_P or event.unicode == 112 or event.unicode == 80


func _is_skills_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_K or event.physical_keycode == KEY_K or event.unicode == 107 or event.unicode == 75


func _is_skill_slot_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_Q or event.physical_keycode == KEY_Q or event.unicode == 113 or event.unicode == 81


func _close_gameplay_panels(except: String = "") -> void:
	if not (except in ["inventory", "stats", "shop_with_inventory", "stash_with_inventory", "bishop", "market"]) and inventory_panel != null:
		inventory_panel.hide_display()
	if not (except in ["shop", "shop_with_inventory"]):
		_hide_shop_panel()
	if not (except in ["stash", "stash_with_inventory"]):
		_hide_stash_panel()
	if except != "bishop":
		_hide_bishop_panel()
	if except != "market":
		_hide_market_panel()
	if not (except in ["stats", "skills", "inventory"]) and character_stats_panel != null:
		character_stats_panel.hide_display()
	if not (except in ["skills", "stats"]) and skills_panel != null:
		skills_panel.hide_display()
	if except != "character_info":
		_hide_character_info_panel()
	if except != "waypoint":
		_hide_waypoint_panel()


func _close_gameplay_panels_for_movement() -> void:
	_close_gameplay_panels()


func _movement_intent_starts_motion(intent_type: String, payload: Dictionary) -> bool:
	if intent_type == "move_to_intent":
		return true
	if intent_type != "move_intent":
		return false
	var direction = payload.get("direction", {})
	if typeof(direction) != TYPE_DICTIONARY:
		return false
	return absf(float(direction.get("x", 0.0))) > 0.0001 or absf(float(direction.get("y", 0.0))) > 0.0001


func _mark_local_player_walking() -> void:
	_player_walk_linger = WALK_ANIMATION_LINGER_SECONDS


func _local_player_is_walking() -> bool:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return false
	if player_hp <= 0 or _user_input_blocked() or _is_force_stand_held() or _movement_requires_fresh_input:
		return false
	if _player_walk_linger > 0.0:
		return true
	return Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_A) \
		or Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_D)


func _tick_movement_animation_linger(delta: float) -> void:
	_player_walk_linger = maxf(0.0, _player_walk_linger - delta)
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if not rec.has("walk_linger"):
			continue
		rec["walk_linger"] = maxf(0.0, float(rec.get("walk_linger", 0.0)) - delta)
		var ctrl = rec.get("controller", null)
		if ctrl == null:
			continue
		var hp := int(rec.get("hp", 1))
		ctrl.set_locomotion(float(rec.get("walk_linger", 0.0)) > 0.0 and hp > 0)


func _is_escape_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_ESCAPE or event.physical_keycode == KEY_ESCAPE


func _is_force_stand_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_SHIFT or event.physical_keycode == KEY_SHIFT


func _is_force_stand_held() -> bool:
	return Input.is_key_pressed(KEY_SHIFT)


func _begin_force_stand() -> void:
	if not _hold_input_allowed() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	_sustained_click.clear()
	_movement_requires_fresh_input = true
	if player_anchor != null:
		predicted_pos = player_anchor.global_position
		_reconcile_player()
	_send_stop_movement_intent()


func _send_stop_movement_intent() -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send("move_intent", last_server_tick, {"direction": {"x": 0, "y": 0}, "duration_ticks": 1})


func _handle_escape() -> void:
	if settings_panel != null and settings_panel.visible:
		_on_settings_back()
		return
	if character_panel != null and character_panel.visible:
		_on_character_panel_back()
		return
	if multiplayer_panel != null and multiplayer_panel.visible:
		multiplayer_panel.hide_panel()
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


func _skill_function_key_slot(event: InputEventKey) -> int:
	var code := event.keycode if event.keycode != KEY_NONE else event.physical_keycode
	if code >= KEY_F1 and code <= KEY_F8:
		return int(code - KEY_F1)
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
		or (multiplayer_panel != null and multiplayer_panel.visible) \
		or (settings_panel != null and settings_panel.visible) \
		or (pause_menu != null and pause_menu.visible) \
		or (loss_popup != null and loss_popup.visible)


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
				_mark_local_player_walking()
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
				_send_action_intent(target_id)
				autoplay_attack_cooldown = autoplay_step_delay
			autoplay_timer = autoplay_step_delay
		"pickup":
			if not autoplay_pickup_sent and loot_ids.size() > 0:
				_send_action_intent(str(loot_ids[0]))
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
	if shop_panel != null and shop_panel.visible:
		return false
	if character_stats_panel != null and character_stats_panel.visible:
		return false
	if skills_panel != null and skills_panel.visible:
		return false
	if character_info_panel != null and character_info_panel.visible:
		return false
	if waypoint_panel != null and waypoint_panel.visible:
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
	if _is_force_stand_held():
		_start_directional_attack_hold()
		return

	var kind := str(pick.get("kind", ""))
	if kind == "floor":
		var ground: Vector3 = pick.get("ground", Vector3.ZERO)
		_close_gameplay_panels_for_movement()
		_mark_local_player_walking()
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
	if typ == "interactable" and _interactable_should_approach_before_action(interactable_def_id):
		_activate_or_approach_interactable(target_id, rec)
		return
	if player_anim != null and (typ == "monster" or (typ == "interactable" and state == "closed")) and _target_in_local_attack_range(target_id):
		player_anim.play_one_shot("attack")

	_send_action_intent(target_id)
	_attack_cooldown = _basic_attack_cooldown_seconds() if typ == "monster" else SEND_INTERVAL


func _target_in_local_attack_range(target_id: String) -> bool:
	if player_anchor == null or target_id == "" or not entities.has(target_id):
		return false
	var rec: Dictionary = entities[target_id]
	var target_node := rec.get("node", null) as Node3D
	if target_node == null:
		return false
	var target_position := _node_world_or_local_position(target_node)
	var player_position := _node_world_or_local_position(player_anchor)
	var flat := Vector2(target_position.x - player_position.x, target_position.z - player_position.z)
	var reach := _local_player_attack_reach()
	return flat.length() <= reach + _local_target_interaction_radius(rec) + LOCAL_REACH_EPSILON


func _local_player_attack_reach() -> float:
	var item := _local_equipped_weapon_item()
	if item.is_empty():
		return LOCAL_UNARMED_REACH
	var def := _local_equipped_weapon_definition(item)
	var reach := float(def.get("reach", LOCAL_UNARMED_REACH))
	return reach if reach > 0.0 else LOCAL_UNARMED_REACH


func _local_equipped_weapon_item() -> Dictionary:
	var raw_weapon_id = equipped.get("main_hand", null)
	if raw_weapon_id == null:
		return {}
	var weapon_id := str(raw_weapon_id)
	if weapon_id == "":
		return {}
	for item in inventory:
		var row: Dictionary = item
		if str(row.get("item_instance_id", "")) == weapon_id:
			return row
	return {}


func _local_equipped_weapon_definition(item: Dictionary) -> Dictionary:
	ItemRulesLoader.ensure_loaded()
	var template_id := str(item.get("item_template_id", ""))
	if template_id != "":
		var template: Variant = ItemRulesLoader.item_templates.get(template_id, {})
		if typeof(template) == TYPE_DICTIONARY:
			return template
	var item_def_id := str(item.get("item_def_id", ""))
	if item_def_id != "":
		return ItemRulesLoader.item_definition(item_def_id)
	return {}


func _local_target_interaction_radius(rec: Dictionary) -> float:
	match str(rec.get("type", "")):
		"monster":
			return LOCAL_MONSTER_RADIUS
		"loot":
			return LOCAL_LOOT_RADIUS
		"interactable":
			return LOCAL_INTERACTABLE_RADIUS
		_:
			return 0.0


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
	elif _sustained_click.mode == "directional_attack":
		_repeat_directional_attack()
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

	if not _target_in_local_attack_range(target_id):
		return

	if player_anim != null:
		player_anim.play_one_shot("attack")

	_send_action_intent(target_id)
	_attack_cooldown = _basic_attack_cooldown_seconds()


func _repeat_hold_move() -> void:
	if _is_force_stand_held():
		_sustained_click.clear()
		return
	var ground := _mouse_ground_point()
	if not _sustained_click.can_repeat_move(ground):
		return

	_close_gameplay_panels_for_movement()
	_mark_local_player_walking()
	client.send("move_to_intent", last_server_tick, {"position": {"x": ground.x, "y": ground.z}})
	_sustained_click.mark_move_sent(ground)
	_attack_cooldown = SEND_INTERVAL


func _try_action_at_mouse() -> void:
	if _attack_cooldown > 0.0 or player_hp <= 0:
		return
	if _is_force_stand_held():
		_start_directional_attack_hold()
		return

	_execute_click_pick(_resolve_click_at_mouse())


func _interactable_should_approach_before_action(interactable_def_id: String) -> bool:
	return interactable_def_id in [
		"stairs_down",
		"stairs_up",
		"teleporter",
		"town_vendor",
		"town_mystery_seller",
		"town_stash",
		"town_bishop",
		"town_market_board",
	]


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
		_send_action_intent(target_id)
		_attack_cooldown = SEND_INTERVAL
		return
	var target_node := rec["node"] as Node3D
	if target_node == null:
		pending_interactable_action.clear()
		return
	_close_gameplay_panels_for_movement()
	_mark_local_player_walking()
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
	var target_pos := target_node.global_position if target_node.is_inside_tree() else target_node.position
	var player_pos := player_anchor.global_position if player_anchor.is_inside_tree() else player_anchor.position
	var flat := Vector2(
		target_pos.x - player_pos.x,
		target_pos.z - player_pos.z
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
			_send_action_intent(target_id)
			_attack_cooldown = SEND_INTERVAL
		return
	_send_action_intent(target_id)
	_attack_cooldown = SEND_INTERVAL


func _update_facing_toward_mouse() -> void:
	var aim := _aim_direction_from_mouse()
	if aim != Vector2.ZERO:
		_face_direction(aim)


func _face_direction(flat_dir: Vector2) -> void:
	if character_visual == null or player_anchor == null:
		return
	if flat_dir.length_squared() <= 0.0001:
		return
	var facing := flat_dir.normalized()
	_last_facing_direction = facing

	character_visual.rotation.y = atan2(facing.x, facing.y)


func _face_entity_direction(node: Node3D, flat_dir: Vector2) -> void:
	if node == null or flat_dir.length_squared() <= 0.0001:
		return
	var facing := flat_dir.normalized()
	node.rotation.y = atan2(facing.x, facing.y)


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


func _start_directional_attack_hold() -> void:
	_sustained_click.begin_directional_attack()
	_try_send_directional_attack()


func _repeat_directional_attack() -> void:
	if not DirectionalAttackInputScript.can_repeat(_is_force_stand_held(), Input.is_mouse_button_pressed(MOUSE_BUTTON_LEFT), _hold_input_allowed(), player_hp):
		_sustained_click.clear()
		return
	_try_send_directional_attack()


func _try_send_directional_attack() -> void:
	if _attack_cooldown > 0.0 or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	var direction := DirectionalAttackInputScript.direction_or_fallback(_aim_direction_from_mouse(), _last_facing_direction)
	_face_direction(direction)
	if player_anim != null:
		player_anim.play_one_shot("attack")
	client.send("directional_attack_intent", last_server_tick, DirectionalAttackInputScript.payload(direction))
	_attack_cooldown = _basic_attack_cooldown_seconds()


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


func _update_loot_hover_label() -> void:
	var reveal_held := _is_loot_label_reveal_held()
	var reveal_changed := reveal_held != loot_label_reveal_held
	loot_label_reveal_held = reveal_held

	var next_hover := ""
	if not _input_locked() and _camera != null:
		var target_id := _pick_entity_at_mouse()
		if target_id != "" and entities.has(target_id) and str(entities[target_id].get("type", "")) == "loot":
			next_hover = target_id
		else:
			next_hover = _nearest_loot_at_ground(_mouse_ground_point())
	if next_hover == hovered_loot_id and not reveal_changed:
		return
	hovered_loot_id = next_hover
	_refresh_loot_label_visibility()


func _is_loot_label_reveal_held() -> bool:
	return Input.is_key_pressed(KEY_ALT)


func _refresh_loot_label_visibility() -> void:
	for loot_id in loot_ids:
		var id := str(loot_id)
		var highlighted := id == hovered_loot_id
		_set_loot_label_visible(id, loot_label_reveal_held or highlighted, highlighted)


func _set_loot_label_visible(loot_id: String, shown: bool, highlighted: bool = false) -> void:
	if loot_id == "" or not entities.has(loot_id):
		return
	var label := _loot_label_node(loot_id)
	if label != null:
		label.visible = shown
		label.modulate = _loot_label_display_color(loot_id, highlighted)


func _loot_label_display_color(loot_id: String, highlighted: bool) -> Color:
	var rec: Dictionary = entities.get(loot_id, {})
	var base := _loot_label_color(rec)
	if highlighted:
		return base
	return Color(
		base.r * LOOT_LABEL_REVEAL_DIM_FACTOR,
		base.g * LOOT_LABEL_REVEAL_DIM_FACTOR,
		base.b * LOOT_LABEL_REVEAL_DIM_FACTOR,
		base.a
	)


func _loot_label_node(loot_id: String) -> Label3D:
	if loot_id == "" or not entities.has(loot_id):
		return null
	var node := entities[loot_id].get("node", null) as Node3D
	if node == null:
		return null
	return node.find_child("LootLabel", true, false) as Label3D


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
	visual_replay_completion_hold_s = maxf(float(visual_cfg.get("post_complete_hold_s", 0.5)), 0.0)
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
	if character_bar != null:
		character_bar.set_interactive(false)
	if skills_panel != null:
		skills_panel.set_interactive(false)
	if skill_bar != null:
		skill_bar.set_interactive(false)
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
		if visual_replay_completion_hold_s > 0.0:
			visual_replay_timer = visual_replay_completion_hold_s
			visual_replay_completion_hold_s = 0.0
			return
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
	for ev in payload.get("events", []):
		var event_type := str(ev.get("event_type", ""))
		var skill_id := str(ev.get("skill_id", ""))
		if skill_id == "poison_stab" and event_type in ["skill_effect_started", "skill_effect_ended", "monster_damaged"]:
			delay = maxf(delay, autoplay_step_delay * 2.4)
		if str(ev.get("weapon_slot", "")) == "off_hand":
			delay = maxf(delay, autoplay_step_delay * 2.0)
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
		_raise_gameplay_windows()
	else:
		inventory_panel.hide_display()


func _apply_shop_event_refresh(ev: Dictionary) -> void:
	if shop_panel == null or not shop_panel.visible:
		return
	if not ev.has("offers") and not ev.has("sell_appraisals"):
		return
	shop_panel.apply_shop_refresh(ev.get("offers", []), ev.get("sell_appraisals", []))


func _entity_world_center(entity_id: String) -> Vector3:
	if not entities.has(entity_id):
		return Vector3.ZERO
	var node := entities[entity_id]["node"] as Node3D
	if node == null:
		return Vector3.ZERO
	return node.global_position


# --- scene construction (placeholder primitives) ----------------------------

func _build_scene() -> void:
	ground_node = _make_ground_node()
	add_child(ground_node)

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
	ui.layer = 5
	add_child(ui)
	_debug_label = Label.new()
	_debug_label.position = Vector2(12, 12)
	ui.add_child(_debug_label)
	_level_label = Label.new()
	_level_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
	_level_label.position = Vector2(0, 12)
	_level_label.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	_level_label.offset_left = -460
	_level_label.offset_right = -12
	_level_label.offset_top = 12
	_level_label.offset_bottom = 64
	ui.add_child(_level_label)
	_update_level_hud()
	boss_health_bar = BossHealthBarScript.new()
	ui.add_child(boss_health_bar)
	_setup_waypoint_panel(ui)
	inventory_panel = InventoryPanelScript.new()
	inventory_panel.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(inventory_panel)
	shop_panel = ShopPanelScript.new()
	shop_panel.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(shop_panel)
	stash_panel = StashPanelScript.new()
	stash_panel.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(stash_panel)
	bishop_panel = BishopPanelScript.new()
	bishop_panel.respec_requested.connect(_on_bishop_respec_requested)
	ui.add_child(bishop_panel)
	market_panel = MarketPanelScript.new()
	market_panel.market_action_requested.connect(_on_market_action_requested)
	ui.add_child(market_panel)
	consumable_bar = ConsumableBarScript.new()
	consumable_bar.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(consumable_bar)
	character_stats_panel = CharacterStatsPanelScript.new()
	character_stats_panel.allocate_stat_requested.connect(_on_character_stat_requested)
	ui.add_child(character_stats_panel)
	skills_panel = SkillsPanelScript.new()
	skills_panel.allocate_skill_point_requested.connect(_on_skill_point_requested)
	ui.add_child(skills_panel)
	character_bar = CharacterBarScript.new()
	character_bar.open_character_requested.connect(_open_character_panel_from_bar)
	ui.add_child(character_bar)
	skill_bar = SkillBarScript.new()
	skill_bar.cast_skill_requested.connect(_on_skill_cast_requested)
	skill_bar.open_skills_requested.connect(_open_skills_panel_from_bar)
	ui.add_child(skill_bar)
	status_effects_bar = StatusEffectsBarScript.new()
	status_effects_bar.effect_expired.connect(_on_status_effect_expired)
	ui.add_child(status_effects_bar)
	_setup_character_info_panel(ui)
	_health_bar = PlayerHealthBarScript.new()
	_refresh_player_hud_identity()
	ui.add_child(_health_bar)
	_raise_gameplay_windows()
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


func _make_ground_node() -> MeshInstance3D:
	var node := MeshInstance3D.new()
	node.name = "Ground"
	var mesh := PlaneMesh.new()
	mesh.size = Vector2(140.0, 90.0)
	mesh.subdivide_width = 32
	mesh.subdivide_depth = 20
	node.mesh = mesh
	node.position = Vector3(50.0, -0.02, 25.0)
	node.material_override = _ground_material_for_level(current_level)
	return node


func _update_ground_material() -> void:
	if ground_node == null:
		return
	ground_node.material_override = _ground_material_for_level(current_level)


func _ground_texture_id_for_level(level: int) -> String:
	return GROUND_TEXTURE_TOWN if level == 0 else GROUND_TEXTURE_DUNGEON


func _ground_material_for_level(level: int) -> StandardMaterial3D:
	var texture_id := _ground_texture_id_for_level(level)
	var mat := StandardMaterial3D.new()
	mat.albedo_texture = _make_ground_texture(texture_id)
	mat.albedo_color = Color.WHITE
	mat.roughness = 0.92
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.uv1_scale = Vector3(28.0, 18.0, 1.0)
	return mat


func _make_ground_texture(texture_id: String) -> ImageTexture:
	if ground_textures.has(texture_id):
		return ground_textures[texture_id] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, _ground_texel(texture_id, x, y))
	var texture := ImageTexture.create_from_image(image)
	ground_textures[texture_id] = texture
	return texture


func _ground_texel(texture_id: String, x: int, y: int) -> Color:
	var n := int((x * 37 + y * 19 + ((x / 8) * 11) + ((y / 8) * 23)) % 17)
	if texture_id == GROUND_TEXTURE_TOWN:
		var base := Color("#2f6136").lerp(Color("#79aa58"), float(n) / 16.0)
		var dirt_patch := int((x * 9 + y * 5 + ((x / 8) * 17) + ((y / 8) * 29)) % 41)
		if dirt_patch < 8:
			base = base.lerp(Color("#8f7447"), 0.50)
		elif dirt_patch < 13:
			base = base.lerp(Color("#6f5f39"), 0.30)
		if ((x * 5 + y * 3) % 23) == 0:
			base = base.lerp(Color("#b7b56c"), 0.34)
		if ((x * 11 + y * 7) % 31) == 0:
			base = base.lerp(Color("#274c2b"), 0.36)
		if ((x / 8 + y / 8) % 5) == 0 and ((x * 13 + y * 17) % 19) < 3:
			base = base.lerp(Color("#456f35"), 0.20)
		return base
	var rock := Color("#3c3f43").lerp(Color("#73706b"), float(n) / 16.0)
	if abs((x % 16) - (y % 16)) <= 1:
		rock = rock.lerp(Color("#25282c"), 0.35)
	if ((x * 3 + y * 5) % 29) == 0:
		rock = rock.lerp(Color("#a09a8e"), 0.28)
	return rock


func _make_wall_texture(texture_id: String) -> ImageTexture:
	if wall_textures.has(texture_id):
		return wall_textures[texture_id] as ImageTexture
	var image := Image.create(64, 64, false, Image.FORMAT_RGB8)
	for y in range(64):
		for x in range(64):
			image.set_pixel(x, y, _wall_texel(texture_id, x, y))
	var texture := ImageTexture.create_from_image(image)
	wall_textures[texture_id] = texture
	return texture


func _wall_texel(_texture_id: String, x: int, y: int) -> Color:
	var brick_w := 16
	var brick_h := 12
	var row := int(y / brick_h)
	var offset := 8 if row % 2 == 1 else 0
	var local_x := int((x + offset) % brick_w)
	var local_y := int(y % brick_h)
	var noise := int((x * 29 + y * 43 + row * 17 + int((x + offset) / brick_w) * 13) % 23)
	var stone := Color("#34363a").lerp(Color("#6b6255"), float(noise) / 22.0)
	if local_x <= 1 or local_y <= 1:
		return stone.lerp(Color("#17191c"), 0.62)
	if local_x >= brick_w - 2 or local_y >= brick_h - 2:
		stone = stone.lerp(Color("#202226"), 0.34)
	if ((x * 5 + y * 7) % 31) == 0:
		stone = stone.lerp(Color("#9b9386"), 0.32)
	if ((x - y) % 19) == 0:
		stone = stone.lerp(Color("#22252a"), 0.30)
	return stone


func _setup_menu_layer() -> void:
	menu_layer = CanvasLayer.new()
	menu_layer.layer = 10
	add_child(menu_layer)

	main_menu = MainMenuScript.new()
	main_menu.create_game_pressed.connect(_on_create_game_pressed)
	main_menu.join_game_pressed.connect(_on_join_game_pressed)
	main_menu.settings_pressed.connect(_on_settings_from_main)
	main_menu.exit_pressed.connect(_exit_game)
	menu_layer.add_child(main_menu)

	character_panel = CharacterSelectPanelScript.new()
	character_panel.back_requested.connect(_on_character_panel_back)
	character_panel.start_requested.connect(_start_selected_character)
	character_panel.create_requested.connect(_on_character_create_requested)
	character_panel.delete_requested.connect(_on_character_delete_requested)
	character_panel.rename_requested.connect(_on_character_rename_requested)
	menu_layer.add_child(character_panel)

	multiplayer_panel = MultiplayerSessionsPanelScript.new()
	multiplayer_panel.refresh_requested.connect(_refresh_multiplayer_sessions)
	multiplayer_panel.join_requested.connect(_on_join_listed_session_requested)
	multiplayer_panel.back_requested.connect(func() -> void:
		multiplayer_panel.hide_panel()
		main_menu.show_menu()
	)
	menu_layer.add_child(multiplayer_panel)

	settings_panel = SettingsPanelScript.new()
	settings_panel.back_requested.connect(_on_settings_back)
	settings_panel.size_selected.connect(_on_window_size_selected)
	settings_panel.floating_combat_text_toggled.connect(_on_floating_combat_text_toggled)
	settings_panel.status_text_toggled.connect(_on_status_text_toggled)
	settings_panel.create_game_session_type_selected.connect(_on_create_game_session_type_selected)
	settings_panel.language_selected.connect(_on_language_selected)
	menu_layer.add_child(settings_panel)

	pause_menu = PauseMenuScript.new()
	pause_menu.resume_pressed.connect(_resume_from_pause)
	pause_menu.settings_pressed.connect(_on_settings_from_pause)
	pause_menu.return_to_menu_pressed.connect(_return_to_main_menu)
	pause_menu.exit_pressed.connect(_exit_game)
	menu_layer.add_child(pause_menu)

	loss_popup = _build_loss_popup()
	menu_layer.add_child(loss_popup)


func _build_loss_popup() -> Control:
	var root := Control.new()
	root.visible = false
	root.mouse_filter = Control.MOUSE_FILTER_STOP
	root.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)

	var shade := ColorRect.new()
	shade.color = Color(0.02, 0.02, 0.025, 0.78)
	shade.set_anchors_preset(Control.PRESET_FULL_RECT)
	root.add_child(shade)

	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(340, 160)
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -170
	panel.offset_right = 170
	panel.offset_top = -80
	panel.offset_bottom = 80
	root.add_child(panel)

	var box := VBoxContainer.new()
	box.alignment = BoxContainer.ALIGNMENT_CENTER
	box.add_theme_constant_override("separation", 14)
	panel.add_child(box)

	var title := Label.new()
	title.text = "You lost"
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 34)
	title.add_theme_color_override("font_color", Color("#f1efe4"))
	box.add_child(title)

	var button := Button.new()
	button.text = "Return to Main Menu"
	button.custom_minimum_size = Vector2(240, 42)
	button.pressed.connect(_return_to_main_menu)
	box.add_child(button)
	return root


func _show_loss_popup() -> void:
	if loss_popup == null or loss_popup.visible:
		return
	if pause_menu != null:
		pause_menu.hide_pause()
	_hide_waypoint_panel()
	_hide_shop_panel()
	_hide_stash_panel()
	_hide_market_panel()
	if skills_panel != null:
		skills_panel.hide_display()
	loss_popup.visible = true


func _on_inventory_intent_requested(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if intent_type == "stash_equip_item_intent":
		_send_stash_equip_item_intent(payload)
		return
	client.send(intent_type, last_server_tick, payload)


func _send_stash_equip_item_intent(payload: Dictionary) -> void:
	var stash_item_id := str(payload.get("stash_item_id", ""))
	var stash_entity_id := str(payload.get("stash_entity_id", ""))
	var slot := str(payload.get("slot", ""))
	if stash_item_id == "" or stash_entity_id == "" or not _is_equipment_slot(slot):
		return
	pending_stash_equips[stash_item_id] = slot
	client.send("stash_withdraw_item_intent", last_server_tick, {
		"stash_entity_id": stash_entity_id,
		"stash_item_id": stash_item_id,
	})


func _handle_stash_item_withdrawn(ev: Dictionary) -> void:
	var stash_item_id := str(ev.get("stash_item_id", ""))
	if stash_item_id == "" or not pending_stash_equips.has(stash_item_id):
		return
	var slot := str(pending_stash_equips.get(stash_item_id, ""))
	pending_stash_equips.erase(stash_item_id)
	var item_instance_id := str(ev.get("item_instance_id", ""))
	if item_instance_id == "" or not _is_equipment_slot(slot) or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	client.send("equip_intent", last_server_tick, {
		"item_instance_id": item_instance_id,
		"slot": slot,
	})


func _is_equipment_slot(slot: String) -> bool:
	return slot in ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]


func _on_character_stat_requested(stat: String) -> void:
	if _stat_allocation_blocked():
		return
	client.send("allocate_stat_intent", last_server_tick, {"stat": stat, "points": 1})


func _on_skill_point_requested(skill_id: String) -> void:
	if _skill_allocation_blocked():
		return
	client.send("allocate_skill_point_intent", last_server_tick, {"skill_id": skill_id})


func _on_skill_cast_requested(skill_id: String) -> void:
	_send_skill_cast_intent(skill_id)


func _open_skills_panel_from_bar() -> void:
	if skills_panel == null:
		return
	_close_gameplay_panels("skills")
	skills_panel.ensure_display_visible()
	_refresh_skill_ui()
	_raise_gameplay_windows()


func _open_character_panel_from_bar() -> void:
	if character_stats_panel == null or _input_locked():
		return
	_close_gameplay_panels("stats")
	character_stats_panel.ensure_display_visible()
	_refresh_progression_ui()
	_raise_gameplay_windows()


func _stat_allocation_blocked() -> bool:
	return visual_replay_enabled \
		or autoplay_enabled \
		or _menu_blocks_gameplay_input() \
		or client == null \
		or client.ready_state() != WebSocketPeer.STATE_OPEN \
		or player_hp <= 0 \
		or int(character_progression.get("unspent_stat_points", 0)) <= 0


func _skill_allocation_blocked() -> bool:
	return visual_replay_enabled \
		or autoplay_enabled \
		or _menu_blocks_gameplay_input() \
		or client == null \
		or client.ready_state() != WebSocketPeer.STATE_OPEN \
		or player_hp <= 0 \
		or int(skill_progression.get("unspent_skill_points", 0)) <= 0


func _skill_cast_blocked(skill_id: String = "") -> bool:
	return _skill_cast_block_reason(skill_id) != ""


func _skill_cast_block_reason(skill_id: String = "") -> String:
	var resolved_skill_id := skill_id
	if resolved_skill_id == "":
		resolved_skill_id = right_click_skill_id
	if resolved_skill_id == "":
		resolved_skill_id = _first_learned_skill_id()
	if resolved_skill_id == "":
		resolved_skill_id = SkillRulesLoader.first_skill_id()
	if visual_replay_enabled:
		return "visual_replay"
	if autoplay_enabled:
		return "autoplay"
	if _menu_blocks_gameplay_input():
		return "menu_open"
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return "not_connected"
	if player_hp <= 0:
		return "dead"
	if _skill_rank(resolved_skill_id) <= 0:
		return "skill_not_learned"
	return ""


func _assign_right_click_skill(skill_id: String) -> bool:
	if _skill_rank(skill_id) <= 0:
		return false
	right_click_skill_id = skill_id
	_sync_skill_bindings_ui()
	_sync_skill_bar_selection()
	_send_skill_bindings_intent()
	return true


func _assign_skill_function_key(slot_index: int, skill_id: String) -> bool:
	if slot_index < 0 or slot_index >= SKILL_FUNCTION_KEY_COUNT or skill_id == "":
		return false
	_ensure_skill_function_key_slots()
	skill_function_keys[slot_index] = skill_id
	if _skill_rank(skill_id) > 0:
		right_click_skill_id = skill_id
	_sync_skill_bindings_ui()
	_sync_skill_bar_selection()
	_send_skill_bindings_intent()
	return true


func _select_right_click_skill_from_function_key(slot_index: int) -> bool:
	if slot_index < 0 or slot_index >= SKILL_FUNCTION_KEY_COUNT:
		return false
	_ensure_skill_function_key_slots()
	var skill_id := str(skill_function_keys[slot_index])
	if skill_id == "":
		return false
	return _assign_right_click_skill(skill_id)


func _handle_skill_function_key(slot_index: int) -> bool:
	if slot_index < 0 or slot_index >= SKILL_FUNCTION_KEY_COUNT:
		return false
	if skills_panel != null and skills_panel.visible:
		var hovered_skill_id := skills_panel.hovered_skill_id()
		if hovered_skill_id != "":
			return _assign_skill_function_key(slot_index, hovered_skill_id)
	return _select_right_click_skill_from_function_key(slot_index)


func _ensure_skill_function_key_slots() -> void:
	while skill_function_keys.size() < SKILL_FUNCTION_KEY_COUNT:
		skill_function_keys.append("")
	if skill_function_keys.size() > SKILL_FUNCTION_KEY_COUNT:
		skill_function_keys.resize(SKILL_FUNCTION_KEY_COUNT)


func _apply_skill_bindings(bindings: Dictionary) -> void:
	var keys: Array = bindings.get("function_keys", [])
	skill_function_keys = keys.duplicate(true)
	_ensure_skill_function_key_slots()
	right_click_skill_id = str(bindings.get("right_click_skill_id", right_click_skill_id))
	_sync_skill_bindings_ui()
	_sync_skill_bar_selection()


func _send_skill_bindings_intent() -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	_ensure_skill_function_key_slots()
	client.send("set_skill_bindings_intent", last_server_tick, {
		"function_keys": skill_function_keys.duplicate(true),
		"right_click_skill_id": right_click_skill_id,
	})


func _sync_skill_bindings_ui() -> void:
	if skills_panel != null:
		skills_panel.set_skill_bindings(skill_function_keys, right_click_skill_id)


func _sync_skill_bar_selection() -> void:
	if skill_bar == null:
		return
	var selected_skill_id := right_click_skill_id
	if selected_skill_id == "":
		selected_skill_id = _first_learned_skill_id()
	if selected_skill_id == "":
		selected_skill_id = SkillRulesLoader.first_skill_id()
	skill_bar.set_skill_id(selected_skill_id)
	skill_bar.set_skill_progression(skill_progression)
	skill_bar.set_skill_cooldowns(skill_cooldowns)
	skill_bar.set_interactive(not _skill_cast_blocked(selected_skill_id))


func _auto_select_right_click_skill() -> void:
	if right_click_skill_id != "" and _skill_rank(right_click_skill_id) > 0:
		return
	right_click_skill_id = ""
	for bound_skill in skill_function_keys:
		var skill_id := str(bound_skill)
		if skill_id != "" and _skill_rank(skill_id) > 0:
			right_click_skill_id = skill_id
			return
	right_click_skill_id = _first_learned_skill_id()


func _first_learned_skill_id() -> String:
	for id in SkillRulesLoader.skill_ids_by_tree():
		var skill_id := str(id)
		if _skill_rank(skill_id) > 0:
			return skill_id
	return ""


func _refresh_progression_ui() -> void:
	if character_stats_panel != null:
		character_stats_panel.set_progression(character_progression)
		character_stats_panel.set_allocation_enabled(not _stat_allocation_blocked())
	if character_bar != null:
		character_bar.set_progression(character_progression)
	if consumable_bar != null:
		consumable_bar.set_character_progression(character_progression)


func _refresh_skill_ui() -> void:
	_auto_select_right_click_skill()
	if skills_panel != null:
		skills_panel.set_character_progression(character_progression)
		skills_panel.set_skill_progression(skill_progression)
		skills_panel.set_skill_bindings(skill_function_keys, right_click_skill_id)
		skills_panel.set_interactive(not _skill_allocation_blocked())
	_sync_skill_bar_selection()


func _sync_progression_interactivity() -> void:
	if character_stats_panel != null:
		character_stats_panel.set_allocation_enabled(not _stat_allocation_blocked())
	if character_bar != null:
		character_bar.set_interactive(not _input_locked())
	if skills_panel != null:
		skills_panel.set_interactive(not _skill_allocation_blocked())
	if skill_bar != null:
		skill_bar.set_interactive(not _skill_cast_blocked(str(skill_bar.get_debug_state().get("skill_id", ""))))


func _try_use_right_click_skill() -> bool:
	if right_click_skill_id == "":
		return false
	var pick := _resolve_click_at_mouse()
	var target_id := ""
	var direction := Vector2.ZERO
	if str(pick.get("kind", "")) == "monster":
		target_id = str(pick.get("target_id", ""))
	else:
		direction = _aim_direction_from_mouse()
	var sent := _send_skill_cast_intent(right_click_skill_id, target_id, direction, false)
	return sent


func _send_skill_cast_intent(skill_id: String, target_id: String = "", direction: Vector2 = Vector2.ZERO, use_nearest_fallback: bool = true) -> bool:
	var blocked_reason := _skill_cast_block_reason(skill_id)
	if blocked_reason != "":
		_show_skill_rejected_feedback(blocked_reason)
		return false
	var payload := _skill_cast_payload(skill_id, target_id, direction, use_nearest_fallback)
	if payload.is_empty():
		_show_skill_rejected_feedback("invalid_target")
		return false
	var message_id := client.send("cast_skill_intent", last_server_tick, payload)
	pending_skill_casts[message_id] = {"skill_id": skill_id}
	_attack_cooldown = SEND_INTERVAL
	if player_anim != null:
		player_anim.play_one_shot("attack")
	return true


func _skill_cast_payload(skill_id: String, target_id: String = "", direction: Vector2 = Vector2.ZERO, use_nearest_fallback: bool = true) -> Dictionary:
	var payload := {"skill_id": skill_id}
	var targeting := _skill_targeting(skill_id)
	if targeting == "self":
		if player_id != "":
			payload["target_id"] = player_id
		else:
			payload["direction"] = {"x": _last_facing_direction.x, "y": _last_facing_direction.y}
		return payload
	if targeting == "self_or_ally_area":
		if player_id != "":
			payload["target_id"] = player_id
		else:
			payload["direction"] = {"x": _last_facing_direction.x, "y": _last_facing_direction.y}
		return payload
	if targeting == "direction_or_target_area" and target_id == "" and direction.length_squared() <= 0.0001 and use_nearest_fallback:
		if player_id != "":
			payload["target_id"] = player_id
			return payload
	var chosen_target := target_id
	if chosen_target == "" and use_nearest_fallback:
		chosen_target = _nearest_live_monster_id()
	if chosen_target != "":
		payload["target_id"] = chosen_target
		_face_toward_entity(chosen_target)
		return payload
	var dir := DirectionalAttackInputScript.direction_or_fallback(direction, _last_facing_direction)
	_face_direction(dir)
	payload["direction"] = {"x": dir.x, "y": dir.y}
	return payload


func _skill_targeting(skill_id: String) -> String:
	var def := SkillRulesLoader.skill_definition(skill_id)
	return str(def.get("targeting", "direction_or_target"))


func _is_skill_reject_reason(reason: String) -> bool:
	return reason.begins_with("skill_") \
		or reason == "unknown_skill" \
		or reason == "not_enough_mana" \
		or reason == "invalid_direction" \
		or reason == "target_out_of_range" \
		or reason == "invalid_target" \
		or reason == "invalid_payload" \
		or reason == "player_dead" \
		or reason == "projectile_busy" \
		or reason == "unsupported_skill_kind"


func _show_skill_rejected_feedback(reason: String = "") -> void:
	if skill_bar != null:
		skill_bar.flash_rejected()
	var message := _skill_reject_message(reason)
	var color := Color("#ffcf5a")
	var variant := "skill_reject"
	if reason == "not_enough_mana":
		color = Color("#54c7f3")
		variant = "mana"
	_show_damage_number(player_id, color, null, "", 0.0, variant, message)


func _skill_reject_message(reason: String) -> String:
	match reason:
		"not_enough_mana":
			return NO_MANA_TEXT
		"skill_not_learned":
			return "SKILL NOT LEARNED"
		"target_out_of_range":
			return "OUT OF RANGE"
		"invalid_direction", "invalid_target":
			return "INVALID TARGET"
		"unknown_skill":
			return "UNKNOWN SKILL"
		"menu_open":
			return "MENU OPEN"
		"not_connected":
			return "NOT CONNECTED"
		"dead":
			return "DEAD"
		"player_dead":
			return "DEAD"
		"skill_on_cooldown":
			return "ON COOLDOWN"
		"skill_class_not_allowed":
			return "WRONG CLASS"
		"skill_requirements_not_met":
			return "REQUIREMENTS NOT MET"
		"invalid_payload":
			return "INVALID CAST"
		"projectile_busy":
			return "PROJECTILE BUSY"
		"unsupported_skill_kind":
			return "UNSUPPORTED SKILL"
		"visual_replay":
			return "REPLAY MODE"
		"autoplay":
			return "AUTOPLAY"
		"":
			return "CANT CAST"
		_:
			return reason.replace("_", " ").to_upper()


func _skill_rank(skill_id: String) -> int:
	var row := _skill_progression_row(skill_id)
	return int(row.get("rank", 0))


func _skill_progression_row(skill_id: String) -> Dictionary:
	var rows: Array = skill_progression.get("skills", [])
	for row in rows:
		if typeof(row) == TYPE_DICTIONARY and str((row as Dictionary).get("skill_id", "")) == skill_id:
			return row as Dictionary
	return {}


func _nearest_live_monster_id() -> String:
	var best_id := ""
	var best_dist := INF
	for mid in monster_ids:
		var id := str(mid)
		if not entities.has(id):
			continue
		var rec: Dictionary = entities[id]
		if int(rec.get("hp", 1)) <= 0:
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var pos := _node_world_or_local_position(node)
		var dist := Vector2(pos.x - predicted_pos.x, pos.z - predicted_pos.z).length()
		if dist < best_dist:
			best_dist = dist
			best_id = id
	return best_id


func _face_toward_entity(target_id: String) -> void:
	if target_id == "" or not entities.has(target_id):
		return
	var rec: Dictionary = entities[target_id]
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var pos := _node_world_or_local_position(node)
	var flat := Vector2(pos.x - predicted_pos.x, pos.z - predicted_pos.z)
	if flat.length_squared() > 0.0001:
		_face_direction(flat.normalized())


func _setup_character_info_panel(ui: CanvasLayer) -> void:
	character_info_panel = PanelContainer.new()
	character_info_panel.visible = false
	character_info_panel.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	character_info_panel.offset_left = -292
	character_info_panel.offset_right = -12
	character_info_panel.offset_top = 76
	character_info_panel.offset_bottom = 190
	character_info_panel.custom_minimum_size = Vector2(280, 114)
	character_info_panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	character_info_panel.add_theme_stylebox_override("panel", _character_info_panel_style())
	ui.add_child(character_info_panel)

	var root := VBoxContainer.new()
	root.add_theme_constant_override("separation", 6)
	character_info_panel.add_child(root)

	var title := Label.new()
	title.text = "Character"
	title.add_theme_font_size_override("font_size", 24)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	root.add_child(title)

	character_info_name_label = _character_info_label()
	character_info_level_label = _character_info_label()
	character_info_area_label = _character_info_label()
	character_info_area_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	root.add_child(character_info_name_label)
	root.add_child(character_info_level_label)
	root.add_child(character_info_area_label)
	_update_character_info_panel()


func _toggle_character_info_panel() -> void:
	if character_info_panel == null or not gameplay_active:
		return
	_close_gameplay_panels("character_info")
	character_info_panel.visible = not character_info_panel.visible
	if character_info_panel.visible:
		_update_character_info_panel()
		_raise_gameplay_windows()


func _hide_character_info_panel() -> void:
	if character_info_panel != null:
		character_info_panel.visible = false


func _update_character_info_panel() -> void:
	if character_info_panel == null:
		return
	if character_info_name_label != null:
		character_info_name_label.text = "Name  %s" % _local_character_display_name()
	if character_info_level_label != null:
		character_info_level_label.text = "Level  %d" % int(character_progression.get("level", 1))
	if character_info_area_label != null:
		character_info_area_label.text = "Area  %s" % _current_area_label()


func _refresh_player_hud_identity() -> void:
	if _health_bar != null:
		_health_bar.set_identity(_local_character_display_name(), int(character_progression.get("level", 1)))


func _local_character_display_name() -> String:
	for member in party:
		if typeof(member) != TYPE_DICTIONARY:
			continue
		var rec := member as Dictionary
		if str(rec.get("player_id", "")) == player_id:
			var display_name := str(rec.get("display_name", "")).strip_edges()
			if display_name != "":
				return display_name
	return "Hero"


func _current_area_label() -> String:
	if current_level == 0:
		return "Town"
	var depth: int = abs(current_level)
	return "Dungeon lvl%d - %s" % [depth, _dungeon_level_name(current_level)]


func _character_info_label() -> Label:
	var label := Label.new()
	label.add_theme_color_override("font_color", Color("#d8c7a6"))
	label.add_theme_font_size_override("font_size", 21)
	label.clip_text = true
	return label


func _character_info_panel_style() -> StyleBoxFlat:
	var s := StyleBoxFlat.new()
	s.bg_color = Color(0.055, 0.055, 0.06, 0.9)
	s.border_color = Color("#53636f")
	s.border_width_left = 1
	s.border_width_top = 1
	s.border_width_right = 1
	s.border_width_bottom = 1
	s.corner_radius_top_left = 6
	s.corner_radius_top_right = 6
	s.corner_radius_bottom_left = 6
	s.corner_radius_bottom_right = 6
	s.content_margin_left = 12
	s.content_margin_right = 12
	s.content_margin_top = 10
	s.content_margin_bottom = 10
	return s


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
	_close_gameplay_panels("waypoint")
	_refresh_waypoint_panel()
	waypoint_panel.visible = true


func _hide_waypoint_panel() -> void:
	if waypoint_panel != null:
		waypoint_panel.visible = false


func _show_shop_panel(ev: Dictionary) -> void:
	if shop_panel == null:
		return
	_close_gameplay_panels("shop_with_inventory")
	var next_shop_id := str(ev.get("shop_id", "town_vendor"))
	var next_entity_id := str(ev.get("entity_id", ""))
	if inventory_panel != null:
		inventory_panel.ensure_display_visible()
		inventory_panel.set_shop_sell_context(next_entity_id)
	shop_panel.show_shop(
		next_entity_id,
		next_shop_id,
		ev.get("offers", []),
		gold,
		inventory,
		equipped,
		_shop_title(next_shop_id),
		ev.get("sell_appraisals", [])
	)
	_raise_gameplay_windows()


func _hide_shop_panel() -> void:
	if shop_panel != null:
		shop_panel.hide_display()
	if inventory_panel != null:
		inventory_panel.clear_shop_sell_context()


func _show_stash_panel(ev: Dictionary) -> void:
	if stash_panel == null:
		return
	_close_gameplay_panels("stash_with_inventory")
	var next_stash_id := str(ev.get("stash_id", "account_stash"))
	var next_entity_id := str(ev.get("entity_id", ""))
	stash_items = ev.get("stash_items", stash_items)
	stash_gold = int(ev.get("stash_gold", stash_gold))
	stash_capacity = int(ev.get("stash_capacity", stash_capacity))
	if inventory_panel != null:
		inventory_panel.ensure_display_visible()
	stash_panel.show_stash(
		next_entity_id,
		next_stash_id,
		stash_items,
		stash_gold,
		stash_capacity,
		inventory,
		equipped,
		gold,
		hotbar,
		_stash_title(next_stash_id)
	)
	_raise_gameplay_windows()


func _hide_stash_panel() -> void:
	if stash_panel != null:
		stash_panel.hide_display()


func _show_bishop_panel(ev: Dictionary) -> void:
	if bishop_panel == null:
		return
	_close_gameplay_panels("bishop")
	var next_entity_id := str(ev.get("entity_id", ""))
	bishop_panel.show_bishop(
		next_entity_id,
		str(ev.get("service", "bishop")),
		int(ev.get("price", 250)),
		bool(ev.get("affordable", gold >= int(ev.get("price", 250)))),
		gold
	)
	_raise_gameplay_windows()


func _hide_bishop_panel() -> void:
	if bishop_panel != null:
		bishop_panel.hide_display()


func _show_market_panel(ev: Dictionary) -> void:
	if market_panel == null:
		return
	_close_gameplay_panels("market")
	var next_entity_id := str(ev.get("entity_id", ""))
	var listings: Array = []
	var status := "Active listings"
	if client != null:
		_refresh_market_board_summary()
		var body := client.list_market_listings()
		listings = body.get("listings", [])
		if body.has("_error"):
			status = "Could not load market listings"
		elif listings.is_empty():
			status = "No active listings"
	market_panel.show_market(next_entity_id, listings, stash_items, client.account_id if client != null else "", status)
	_raise_gameplay_windows()


func _on_market_action_requested(action: String, payload: Dictionary) -> void:
	if client == null:
		return
	var result := {}
	if action == "publish":
		result = client.create_market_listing(str(payload.get("stash_item_id", "")))
		if result.has("_error"):
			if market_panel != null:
				market_panel.show_status("Could not publish item", true)
			return
		_remove_market_stash_item(str(payload.get("stash_item_id", "")))
		if market_panel != null:
			market_panel.show_status("Item published")
	elif action == "offer":
		result = client.create_market_offer(str(payload.get("listing_id", "")), payload.get("stash_item_ids", []))
		if result.has("_error"):
			if market_panel != null:
				market_panel.show_status("Could not make offer", true)
			return
		for stash_item_id in payload.get("stash_item_ids", []):
			_remove_market_stash_item(str(stash_item_id))
		if market_panel != null:
			market_panel.show_status("Offer sent")
	else:
		return
	_refresh_market_panel_data()


func _refresh_market_panel_data() -> void:
	_refresh_market_board_summary()
	if market_panel == null or client == null:
		return
	var body := client.list_market_listings()
	var listings: Array = body.get("listings", [])
	market_panel.show_market(market_panel.market_entity_id, listings, stash_items, client.account_id, market_panel.get_debug_state().get("status", ""))


func _remove_market_stash_item(stash_item_id: String) -> void:
	for i in range(stash_items.size() - 1, -1, -1):
		if str((stash_items[i] as Dictionary).get("stash_item_id", "")) == stash_item_id:
			stash_items.remove_at(i)
			return


func _refresh_market_board_summary() -> void:
	if client == null or interactable_ids.is_empty():
		return
	var has_market_board := false
	for id in interactable_ids:
		var key := str(id)
		if entities.has(key) and str((entities[key] as Dictionary).get("interactable_def_id", "")) == "town_market_board":
			has_market_board = true
			break
	if not has_market_board:
		return
	var summary := client.market_summary()
	_update_market_board_badges(
		int(summary.get("incoming_bids", 0)),
		int(summary.get("published_listings", 0))
	)


func _update_market_board_badges(incoming_bids: int, published_listings: int) -> void:
	for id in interactable_ids:
		var key := str(id)
		if not entities.has(key):
			continue
		var rec: Dictionary = entities[key]
		if str(rec.get("interactable_def_id", "")) != "town_market_board":
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var left := node.find_child("IncomingBidCount", true, false) as Label3D
		var right := node.find_child("PublishedListingCount", true, false) as Label3D
		if left != null:
			left.text = str(incoming_bids)
			left.modulate = Color("#ffcf5a") if incoming_bids > 0 else Color("#776d5e")
		if right != null:
			right.text = str(published_listings)
			right.modulate = Color("#9fd7ff") if published_listings > 0 else Color("#776d5e")


func _hide_market_panel() -> void:
	if market_panel != null:
		market_panel.hide_display()


func _on_bishop_respec_requested(bishop_entity_id: String) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or bishop_entity_id == "":
		return
	client.send("bishop_respec_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})


func _shop_title(next_shop_id: String) -> String:
	match next_shop_id:
		"town_vendor":
			return "Town Vendor"
		"town_mystery_seller":
			return "Mystery Seller"
		_:
			return next_shop_id.replace("_", " ").capitalize()


func _stash_title(next_stash_id: String) -> String:
	match next_stash_id:
		"account_stash":
			return "Account Stash"
		_:
			return next_stash_id.replace("_", " ").capitalize()


func _sync_waypoint_panel_reach() -> void:
	if waypoint_panel == null or not waypoint_panel.visible:
		return
	var teleporter := _current_teleporter_record()
	if teleporter.is_empty() or not _interactable_in_activation_range(teleporter):
		_hide_waypoint_panel()


func _sync_actionable_panel_reach() -> void:
	var closed_actionable := false
	if shop_panel != null and shop_panel.visible:
		if not _panel_source_in_activation_range(shop_panel.shop_entity_id):
			_hide_shop_panel()
			closed_actionable = true
	if stash_panel != null and stash_panel.visible:
		if not _panel_source_in_activation_range(stash_panel.stash_entity_id):
			_hide_stash_panel()
			closed_actionable = true
	if bishop_panel != null and bishop_panel.visible:
		if not _panel_source_in_activation_range(bishop_panel.bishop_entity_id):
			_hide_bishop_panel()
			closed_actionable = true
	if market_panel != null and market_panel.visible:
		if not _panel_source_in_activation_range(market_panel.market_entity_id):
			_hide_market_panel()
			closed_actionable = true
	if closed_actionable:
		_hide_inventory_if_no_actionable_panel()


func _panel_source_in_activation_range(entity_id: String) -> bool:
	if entity_id == "" or not entities.has(entity_id):
		return false
	var rec: Dictionary = entities.get(entity_id, {})
	return _interactable_in_activation_range(rec)


func _hide_inventory_if_no_actionable_panel() -> void:
	if inventory_panel == null:
		return
	var shop_visible := shop_panel != null and shop_panel.visible
	var stash_visible := stash_panel != null and stash_panel.visible
	var bishop_visible := bishop_panel != null and bishop_panel.visible
	var market_visible: bool = market_panel != null and market_panel.visible
	if not shop_visible and not stash_visible and not bishop_visible and not market_visible:
		inventory_panel.hide_display()


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
		_close_gameplay_panels_for_movement()
		_mark_local_player_walking()
		client.send("move_to_intent", last_server_tick, {
			"position": {"x": target_node.global_position.x, "y": target_node.global_position.z},
		})
	_hide_waypoint_panel()


func bot_click_waypoint_level(level: int) -> void:
	if waypoint_panel == null or not waypoint_panel.visible:
		return
	if not bool(discovered_teleporters.get(level, false)):
		return
	_on_waypoint_level_pressed(level)


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
	current_wall_layout = []
	_clear_wall_nodes()

	var rules_path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/worlds.v0.json")
	var parsed = _read_json(rules_path)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("[main] could not read world rules for walls: %s" % rules_path)
		return
	var worlds: Dictionary = parsed.get("worlds", {})
	var world: Dictionary = worlds.get(world_id, {})
	if str(world.get("mode", "")) == "multi_level":
		return
	var local_walls: Array = []
	var local_index := 0
	for entity in world.get("entities", []):
		if str(entity.get("type", "")) != "wall":
			continue
		var pos: Dictionary = entity.get("position", {})
		var size: Dictionary = entity.get("size", {})
		local_walls.append({
			"id": "%s_wall_%03d" % [world_id, local_index],
			"position": {"x": float(pos.get("x", 0.0)), "y": float(pos.get("y", 0.0))},
			"size": {"x": float(size.get("x", 1.0)), "y": float(size.get("y", 1.0))},
			"source": "preset",
		})
		local_index += 1
	_render_wall_layout(local_walls)


func _render_wall_layout(walls: Array) -> void:
	if walls_root == null:
		return
	current_wall_layout = []
	_clear_wall_nodes()
	for wall in walls:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		var normalized := _normalized_wall_view(wall as Dictionary, current_wall_layout.size())
		current_wall_layout.append(normalized)
		walls_root.add_child(_make_wall_node(normalized))


func _clear_wall_nodes() -> void:
	if walls_root == null:
		return
	for child in walls_root.get_children():
		walls_root.remove_child(child)
		child.queue_free()


func _normalized_wall_view(wall: Dictionary, index: int) -> Dictionary:
	var pos: Dictionary = {}
	var size: Dictionary = {}
	if typeof(wall.get("position", {})) == TYPE_DICTIONARY:
		pos = wall.get("position", {})
	if typeof(wall.get("size", {})) == TYPE_DICTIONARY:
		size = wall.get("size", {})
	var out := {
		"id": str(wall.get("id", "wall_%03d" % index)),
		"position": {"x": float(pos.get("x", 0.0)), "y": float(pos.get("y", 0.0))},
		"size": {"x": float(size.get("x", 1.0)), "y": float(size.get("y", 1.0))},
	}
	if wall.has("source"):
		out["source"] = str(wall.get("source", ""))
	return out


func _make_wall_node(wall: Dictionary) -> MeshInstance3D:
	var pos: Dictionary = wall.get("position", {})
	var size: Dictionary = wall.get("size", {})
	var node := MeshInstance3D.new()
	node.name = "Wall_%s" % str(wall.get("id", ""))
	node.set_meta("wall_id", str(wall.get("id", "")))
	node.set_meta("source", str(wall.get("source", "")))
	var mesh := BoxMesh.new()
	mesh.size = Vector3(float(size.get("x", 1.0)), 1.0, float(size.get("y", 1.0)))
	node.mesh = mesh
	node.position = Vector3(float(pos.get("x", 0.0)), 0.5, float(pos.get("y", 0.0)))
	var mat := StandardMaterial3D.new()
	mat.albedo_texture = _make_wall_texture(WALL_TEXTURE_CAVE)
	mat.texture_filter = BaseMaterial3D.TEXTURE_FILTER_NEAREST
	mat.roughness = 0.96
	mat.uv1_scale = Vector3(max(1.0, float(size.get("x", 1.0)) / 2.0), max(1.0, float(size.get("y", 1.0)) / 2.0), 1.0)
	match str(wall.get("source", "")):
		"generated":
			mat.albedo_color = Color(0.92, 0.86, 0.76)
		"perimeter":
			mat.albedo_color = Color(0.62, 0.64, 0.68)
		_:
			mat.albedo_color = Color(0.78, 0.80, 0.82)
	node.material_override = mat
	return node


func _update_level_hud() -> void:
	if _level_label == null:
		return
	var lines: Array[String] = []
	if current_level == 0:
		lines.append("Town")
	else:
		var depth: int = abs(current_level)
		lines.append("Level %d - %s" % [depth, _dungeon_level_name(current_level)])
	if client != null and client.session_id != "" and (client.session_mode == "coop" or client.session_listed):
		lines.append("Session %s" % client.session_id)
	_level_label.visible = lines.size() > 0
	_level_label.text = "\n".join(lines)


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
	# Monsters resolve presentation through shared visual metadata while
	# gameplay stays server-owned.
	if kind == "monster":
		var visual := MonsterVisualsLoaderScript.resolve(str(e.get("monster_def_id", "")), str(e.get("visual_model", "")))
		var packed := _monster_scene_for_visual(str(visual.get("scene", "monster_dummy")))
		if str(e.get("visual_model", "")) == BOSS_VISUAL_MODEL:
			packed = CharacterScene
		if packed != null:
			var root := Node3D.new()
			root.name = "MonsterVisualRoot"
			var monster := packed.instantiate() as Node3D
			monster.position.y = float(visual.get("height_offset", 0.0))
			monster.scale = Vector3.ONE * float(visual.get("scale", 1.0))
			root.add_child(monster)
			root.scale = Vector3.ONE * _entity_visual_scale(e)
			_apply_model_tint(root, _entity_base_tint(e))
			_sync_archer_bow_marker(root, str(e.get("monster_def_id", "")))
			return root
		# Fallback: red primitive so positioning/targeting still works.
		var fallback := MeshInstance3D.new()
		var fm := StandardMaterial3D.new()
		fm.albedo_color = _entity_base_tint(e)
		fallback.mesh = BoxMesh.new()
		fallback.material_override = fm
		fallback.scale = Vector3.ONE * _entity_visual_scale(e)
		_sync_archer_bow_marker(fallback, str(e.get("monster_def_id", "")))
		return fallback
	if kind == "player":
		return _make_remote_player_node(e)
	if kind == "interactable":
		var def_id := str(e.get("interactable_def_id", ""))
		if def_id == "stairs_down" or def_id == "stairs_up":
			return _make_stair_node(def_id)
		if def_id == "teleporter":
			return _make_teleporter_node()
		if def_id == "treasure_chest" or def_id == "town_stash":
			return _make_chest_node(def_id)
		if def_id == "town_vendor" or def_id == "town_mystery_seller":
			return _make_merchant_node(def_id)
		if def_id == "town_bishop":
			return _make_bishop_node()
		if def_id == "town_market_board":
			return _make_market_board_node()
		return _make_door_node()
	if kind == "projectile":
		return ProjectileVisualsScript.make_node(str(e.get("projectile_def_id", "")))
	return _make_loot_node(e)


func _monster_scene_for_visual(scene_key: String) -> PackedScene:
	match scene_key:
		"monster_quadruped":
			return MonsterQuadrupedScene
		"monster_tiny_flyer":
			return MonsterTinyFlyerScene
		"monster_skeleton":
			return MonsterSkeletonScene
		_:
			return MonsterDummyScene


func _make_remote_player_node(e: Dictionary) -> Node3D:
	var root = CharacterScene.instantiate() as Node3D
	root.name = "RemotePlayer_%s" % str(e.get("id", ""))
	_apply_character_class_model(root, str(e.get("character_class", "")))
	root.scale = Vector3.ONE * _entity_visual_scale(e)
	_apply_model_tint(root, REMOTE_PLAYER_TINT)
	return root


func _monster_tint(rarity: String) -> Color:
	return MONSTER_RARITY_TINTS.get(rarity, MONSTER_RARITY_TINTS["common"])


func _entity_base_tint(e: Dictionary) -> Color:
	var kind := str(e.get("type", ""))
	if kind == "player":
		return REMOTE_PLAYER_TINT
	if kind == "monster":
		if e.has("visual_tint"):
			return Color(str(e.get("visual_tint", "#ffffff")))
		return _monster_tint(str(e.get("rarity", "common")))
	return Color.WHITE


func _entity_visual_scale(e: Dictionary) -> float:
	var scale := float(e.get("visual_scale", 1.0))
	if scale <= 0.0:
		return 1.0
	return scale


func _apply_local_player_visual_scale(scale: float) -> void:
	player_visual_scale = scale if scale > 0.0 else 1.0
	if character_visual != null:
		character_visual.scale = Vector3.ONE * player_visual_scale


func _apply_local_player_class_model() -> void:
	if character_visual == null:
		return
	var class_id := str(character_progression.get("character_class", ""))
	var resolved := ClassPresentationsLoaderScript.resolve(class_id)
	var asset_id := str(resolved.get("asset_id", ""))
	if asset_id == "" or asset_id == _local_player_class_asset_id:
		return
	_local_player_class_asset_id = asset_id
	_apply_character_class_model(character_visual, class_id)
	_apply_local_player_visual_scale(player_visual_scale)
	_apply_model_tint(character_visual, PLAYER_TINT)
	player_reaction = ModelReactionControllerScript.new(character_visual, PLAYER_TINT)
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)


func _apply_character_class_model(root: Node3D, class_id: String) -> void:
	var packed := ClassPresentationsLoaderScript.packed_scene_for_class(class_id)
	if packed == null:
		return
	var old_model := root.find_child("ModelRoot", false, false) as Node
	if old_model != null:
		root.remove_child(old_model)
		old_model.free()
	var model := packed.instantiate() as Node3D
	if model == null:
		return
	model.name = "ModelRoot"
	root.add_child(model)
	root.move_child(model, 0)
	var ap := root.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	if root.has_method("_ensure_weapon_socket"):
		root.call("_ensure_weapon_socket")
	if root.has_method("_ensure_fallback_sockets"):
		root.call("_ensure_fallback_sockets")


func _apply_entity_visual_metadata(rec: Dictionary, e: Dictionary) -> void:
	if e.has("monster_def_id"):
		rec["monster_def_id"] = str(e["monster_def_id"])
		var visual := MonsterVisualsLoaderScript.resolve(str(e["monster_def_id"]), str(e.get("visual_model", "")))
		if not e.has("visual_model"):
			rec["visual_model"] = str(visual.get("visual_model", visual.get("scene", "")))
	for key in ["boss_template_id", "visual_model", "visual_tint", "boss_phase", "effect_ids"]:
		if e.has(key):
			rec[key] = e[key]
	if e.has("is_boss"):
		rec["is_boss"] = bool(e["is_boss"])
	if e.has("visual_scale"):
		rec["visual_scale"] = float(e["visual_scale"])
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	node.scale = Vector3.ONE * float(rec.get("visual_scale", 1.0))
	var base_tint := _entity_base_tint(e)
	rec["base_tint"] = base_tint.to_html(false)
	_apply_entity_status_tint(rec)
	_sync_archer_bow_marker(node, str(rec.get("monster_def_id", "")))
	rec["has_bow_marker"] = _has_archer_bow_marker(node)
	PlayerStatusEffectMarkers.sync_holy_shield_effect(node, rec.get("effect_ids", []))
	PlayerStatusEffectMarkers.sync_burning_effect(node, PlayerStatusEffectMarkers.has_burning_effect_id(rec.get("effect_ids", [])))
	_normalize_boss_phase_metadata(rec)
	_sync_boss_telegraph_marker_from_record(rec)


func _sync_archer_bow_marker(root: Node3D, monster_def_id: String) -> void:
	if root == null:
		return
	var existing := root.find_child(ARCHER_BOW_MARKER_NAME, true, false) as Node3D
	if monster_def_id != ARCHER_MONSTER_DEF_ID:
		if existing != null:
			existing.queue_free()
		return
	if existing == null:
		existing = _make_archer_bow_marker()
		root.add_child(existing)
	_apply_archer_bow_material(existing)


func _has_archer_bow_marker(root: Node3D) -> bool:
	return root != null and root.find_child(ARCHER_BOW_MARKER_NAME, true, false) != null


func _make_archer_bow_marker() -> Node3D:
	var marker := Node3D.new()
	marker.name = ARCHER_BOW_MARKER_NAME
	marker.position = Vector3(0.42, 0.88, -0.28)
	marker.rotation_degrees = Vector3(0.0, 0.0, -8.0)
	marker.add_child(_make_archer_bow_part("BowGrip", Vector3(0.055, 0.46, 0.045), Vector3(0.0, 0.0, 0.0), 0.0, Color(0.39, 0.21, 0.08)))
	marker.add_child(_make_archer_bow_part("BowUpperLimb", Vector3(0.045, 0.40, 0.04), Vector3(0.05, 0.34, 0.0), -18.0, Color(0.52, 0.31, 0.12)))
	marker.add_child(_make_archer_bow_part("BowLowerLimb", Vector3(0.045, 0.40, 0.04), Vector3(0.05, -0.34, 0.0), 18.0, Color(0.52, 0.31, 0.12)))
	marker.add_child(_make_archer_bow_part("BowString", Vector3(0.018, 0.90, 0.018), Vector3(0.18, 0.0, 0.0), 0.0, Color(0.86, 0.82, 0.68)))
	return marker


func _make_archer_bow_part(part_name: String, size: Vector3, position: Vector3, z_rotation_degrees: float, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	part.rotation_degrees.z = z_rotation_degrees
	part.set_meta("archer_bow_color", color.to_html(false))
	return part


func _apply_archer_bow_material(root: Node) -> void:
	if root is MeshInstance3D and root.has_meta("archer_bow_color"):
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color("#" + str(root.get_meta("archer_bow_color")))
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		_apply_archer_bow_material(child)


func _apply_boss_phase_started(entity_id: String, ev: Dictionary) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	var duration := maxi(0, int(ev.get("duration_ticks", 0)))
	rec["boss_phase"] = {
		"pattern_id": str(ev.get("pattern_id", "")),
		"phase_index": int(ev.get("phase_index", -1)),
		"phase_kind": str(ev.get("phase_kind", "")),
		"duration_ticks": duration,
		"remaining_ticks": duration,
		"remaining_ticks_float": float(duration),
		"telegraph": ev.get("telegraph", {}),
		"hit_shape": ev.get("hit_shape", {}),
	}
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var phase_kind := str(ev.get("phase_kind", ""))
	if phase_kind == "telegraph":
		var telegraph: Dictionary = ev.get("telegraph", {})
		var tint := Color(str(telegraph.get("to_color", "#ff0000")))
		rec["boss_telegraph_active"] = true
		rec["telegraph_tint"] = tint.to_html(false)
		_apply_model_tint(node, tint)
		_sync_boss_telegraph_marker(rec, telegraph)
	else:
		rec["boss_telegraph_active"] = false
		_remove_boss_telegraph_marker(rec)
		_apply_entity_status_tint(rec)
	_sync_boss_health_bar()


func _apply_boss_phase_ended(entity_id: String, _ev: Dictionary) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	rec["boss_telegraph_active"] = false
	rec.erase("boss_phase")
	_remove_boss_telegraph_marker(rec)
	_apply_entity_status_tint(rec)
	_sync_boss_health_bar()


func _normalize_boss_phase_metadata(rec: Dictionary) -> void:
	var phase := _boss_phase_for_display(rec)
	if phase.is_empty():
		return
	if str(phase.get("phase_kind", "")) == "telegraph":
		var telegraph: Dictionary = phase.get("telegraph", {})
		if not telegraph.is_empty():
			rec["boss_telegraph_active"] = true
			rec["telegraph_tint"] = Color(str(telegraph.get("to_color", "#ff0000"))).to_html(false)


func _sync_boss_telegraph_marker_from_record(rec: Dictionary) -> void:
	var phase := _boss_phase_for_display(rec)
	if phase.is_empty() or str(phase.get("phase_kind", "")) != "telegraph":
		_remove_boss_telegraph_marker(rec)
		return
	var telegraph: Dictionary = phase.get("telegraph", {})
	if telegraph.is_empty():
		_remove_boss_telegraph_marker(rec)
		return
	_sync_boss_telegraph_marker(rec, telegraph)


func _sync_boss_telegraph_marker(rec: Dictionary, telegraph: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var marker := node.find_child(BOSS_TELEGRAPH_MARKER_NAME, false, false) as MeshInstance3D
	if marker == null:
		marker = MeshInstance3D.new()
		marker.name = BOSS_TELEGRAPH_MARKER_NAME
		marker.position = Vector3(0.0, 0.035, 0.0)
		node.add_child(marker)
	var radius := maxf(0.1, float(telegraph.get("radius", 1.0)))
	var visual_scale := maxf(0.1, float(rec.get("visual_scale", 1.0)))
	var local_radius := radius / visual_scale
	var mesh := CylinderMesh.new()
	mesh.top_radius = local_radius
	mesh.bottom_radius = local_radius
	mesh.height = 0.035
	mesh.radial_segments = 48
	marker.mesh = mesh
	var color := Color(str(telegraph.get("to_color", "#ff4a32")))
	color.a = 0.34
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	marker.material_override = mat
	rec["boss_telegraph_active"] = true
	rec["telegraph_tint"] = color.to_html(false)
	rec["has_boss_telegraph_marker"] = true
	rec["telegraph_radius"] = radius
	rec["telegraph_marker_color"] = color.to_html(false)


func _remove_boss_telegraph_marker(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node != null:
		var marker := node.find_child(BOSS_TELEGRAPH_MARKER_NAME, false, false)
		if marker != null:
			marker.queue_free()
	rec["has_boss_telegraph_marker"] = false
	rec["telegraph_radius"] = 0.0
	rec["telegraph_marker_color"] = ""


func _set_entity_poison_tint(entity_id: String, active: bool) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	rec["poisoned"] = active
	_apply_entity_status_tint(rec)


func _set_entity_burning(entity_id: String, active: bool) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	rec["burning"] = active
	var node := rec.get("node", null) as Node3D
	PlayerStatusEffectMarkers.sync_burning_effect(node, active)
	_apply_entity_status_tint(rec)


func _apply_entity_status_tint(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null or bool(rec.get("boss_telegraph_active", false)):
		return
	var tint := Color("#" + str(rec.get("base_tint", "ffffff")))
	if PlayerStatusEffectMarkers.has_ice_slow_effect(rec.get("effect_ids", [])):
		tint = Color(0.62, 0.86, 1.0)
	if bool(rec.get("burning", false)) or PlayerStatusEffectMarkers.has_burning_effect_id(rec.get("effect_ids", [])):
		tint = Color(1.0, 0.38, 0.12)
	if bool(rec.get("poisoned", false)):
		tint = POISON_TINT
	var reaction = rec.get("reaction", null)
	if reaction != null and reaction.has_method("set_base_tint"):
		reaction.set_base_tint(tint)
	else:
		_apply_model_tint(node, tint)


func _apply_model_tint(root: Node, color: Color) -> void:
	if root is MeshInstance3D:
		var mat := StandardMaterial3D.new()
		mat.albedo_color = color
		(root as MeshInstance3D).material_override = mat
	for child in root.get_children():
		_apply_model_tint(child, color)


func _make_loot_node(e: Dictionary) -> Node3D:
	var item_def_id := str(e.get("item_def_id", ""))
	var root := Node3D.new()
	root.name = "Loot_%s" % item_def_id
	var ground: Dictionary = item_presentations.get(item_def_id, {}).get("ground", {})
	var shape := str(ground.get("shape", "box"))
	var color := Color(str(ground.get("color", "#" + _loot_color(item_def_id).to_html(false))))
	var accent := Color(str(ground.get("accent", "#f6e8b1")))
	var scale := float(ground.get("scale", 1.0))
	var model := _make_ground_equipment_model(item_def_id, str(e.get("rarity", "common")))
	if model != null:
		root.add_child(model)
	else:
		_add_loot_rarity_background(root, _item_rarity_background(str(e.get("rarity", "common"))), scale)
		_add_loot_primitive(root, shape, color, accent, scale)
	_add_loot_label(root, _loot_label_text(e), scale, _loot_label_color(e))
	return root


func _add_loot_primitive(root: Node3D, shape: String, color: Color, accent: Color, scale: float) -> void:
	match shape:
		"blade":
			_add_loot_box(root, "Blade", Vector3(0.12, 0.08, 0.78) * scale, Vector3(0.0, 0.20, 0.0), color)
			_add_loot_box(root, "Grip", Vector3(0.34, 0.10, 0.10) * scale, Vector3(0.0, 0.16, 0.34 * scale), accent)
		"bow":
			_add_loot_box(root, "BowTop", Vector3(0.10, 0.08, 0.42) * scale, Vector3(0.14 * scale, 0.20, -0.18 * scale), color)
			_add_loot_box(root, "BowBottom", Vector3(0.10, 0.08, 0.42) * scale, Vector3(-0.14 * scale, 0.20, 0.18 * scale), color)
			_add_loot_box(root, "String", Vector3(0.04, 0.06, 0.75) * scale, Vector3(0.0, 0.18, 0.0), accent)
		"badge", "coin":
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


func _make_ground_equipment_model(item_def_id: String, rarity: String) -> Node3D:
	var presentation: Dictionary = item_presentations.get(item_def_id, {})
	var asset_id := str(presentation.get("3d_model", ""))
	if asset_id == "":
		return null
	var entry = asset_manifest.get(asset_id, null)
	if typeof(entry) != TYPE_DICTIONARY:
		return null
	var runtime_path := str((entry as Dictionary).get("runtime_path", ""))
	var packed = load(_res_path(runtime_path))
	if packed == null or not (packed is PackedScene):
		return null
	var inst := (packed as PackedScene).instantiate() as Node3D
	if inst == null:
		return null
	inst.name = "GroundModel_%s" % asset_id
	inst.scale = Vector3.ONE * GROUND_EQUIPMENT_MODEL_SCALE
	inst.position = Vector3(0.0, 0.12, 0.0)
	inst.rotation_degrees = Vector3(90.0, 35.0, 0.0)
	_apply_model_tint(inst, _ground_item_tint(rarity))
	return inst


func _ground_item_tint(rarity: String) -> Color:
	match rarity.to_lower():
		"magic":
			return Color("#5aa7ff")
		"rare":
			return Color("#ffd75e")
		"unique":
			return Color("#ff9f52")
		_:
			return Color("#d8d0bd")


func _add_loot_rarity_background(parent: Node3D, color: Color, scale: float) -> void:
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.82, 0.04, 0.82) * maxf(scale, 0.85)
	_add_loot_mesh(parent, "RarityBackground", mesh, Vector3(0.0, 0.045, 0.0), color)


func _add_loot_label(parent: Node3D, text: String, scale: float, color: Color = Color("#f4ead8")) -> void:
	if text == "":
		return
	var label := Label3D.new()
	label.name = "LootLabel"
	label.text = text
	label.visible = false
	label.position = Vector3(0.0, 0.58 * maxf(scale, 0.8), 0.0)
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.no_depth_test = true
	label.fixed_size = true
	label.pixel_size = 0.0018
	label.modulate = color
	label.outline_modulate = Color(0.06, 0.045, 0.035, 0.92)
	label.outline_size = 4
	parent.add_child(label)


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


func _loot_label_color(e: Dictionary) -> Color:
	var item_def_id := str(e.get("item_def_id", ""))
	var def := _item_definition(item_def_id)
	var category := str(def.get("category", "")).to_lower()
	if item_def_id == "gold" or category == "currency":
		return LOOT_LABEL_CATEGORY_COLORS["currency"]
	if LOOT_LABEL_CATEGORY_COLORS.has(category):
		return LOOT_LABEL_CATEGORY_COLORS[category]
	var rarity := str(e.get("rarity", "common")).to_lower()
	return LOOT_LABEL_RARITY_COLORS.get(rarity, LOOT_LABEL_RARITY_COLORS["common"])


func _loot_label_text(e: Dictionary) -> String:
	var item_def_id := str(e.get("item_def_id", ""))
	var def := _item_definition(item_def_id)
	var category := str(def.get("category", "")).to_lower()
	if item_def_id == "gold" or category == "currency":
		var amount := int(e.get("amount", 0))
		if amount > 0:
			return "%d gold" % amount
		return "gold"
	return _generic_loot_name(item_def_id)


func _item_rarity_background(rarity: String) -> Color:
	var key := rarity.to_lower()
	return ITEM_RARITY_BACKGROUNDS.get(key, ITEM_RARITY_BACKGROUNDS["common"])


func _item_definition(item_def_id: String) -> Dictionary:
	return ItemRulesLoader.item_definition(item_def_id)


func _generic_loot_name(item_def_id: String) -> String:
	var def := _item_definition(item_def_id)
	var slot := str(def.get("slot", ""))
	match slot:
		"main_hand":
			var attack_mode := str(def.get("attack_mode", "melee"))
			if attack_mode == "ranged":
				return "Bow"
			return "Sword"
		"off_hand":
			return "Shield"
		"head":
			return "Helm"
		"chest":
			return "Armor"
		"gloves":
			return "Gloves"
		"belt":
			return "Belt"
		"boots":
			return "Boots"
		"amulet":
			return "Amulet"
		"ring":
			return "Ring"
	var category := str(def.get("category", ""))
	match category:
		"consumable":
			return "Potion"
		"currency":
			return "Badge"
		"quest":
			return "Item"
	return "Item"


func _load_dungeon_generation() -> void:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/dungeon_generation.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		dungeon_generation = parsed


func _load_ground_item_visual_data() -> void:
	var base := ProjectSettings.globalize_path("res://")
	var manifest = _read_json(base.path_join("../assets/manifests/assets.v0.json"))
	if typeof(manifest) == TYPE_DICTIONARY:
		asset_manifest = manifest.get("assets", {})


func _res_path(runtime_path: String) -> String:
	var p := runtime_path
	if p.begins_with("client/"):
		p = p.substr("client/".length())
	return "res://" + p


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


func _make_chest_node(def_id: String) -> Node3D:
	var is_stash := def_id == "town_stash"
	var root := Node3D.new()
	root.name = "TownStashChest" if is_stash else "TreasureChest"
	var scale := 1.12 if is_stash else 1.0
	var wood := Color("#6f3b18") if is_stash else Color("#744018")
	var dark_wood := Color("#3c2111") if is_stash else Color("#4a2711")
	var metal := Color("#d1b15d") if is_stash else Color("#8d8f8f")
	var metal_dark := Color("#6f5b2e") if is_stash else Color("#3d4143")
	var cloth := Color("#244e66") if is_stash else Color("#5a2017")
	var glow := Color("#f3d36b") if is_stash else Color("#f5b449")

	_add_chest_part(root, "ChestShadow", Vector3(1.28, 0.035, 0.82) * scale, Vector3(0.0, 0.018, 0.0), Color("#181715"))
	_add_chest_part(root, "ChestBody", Vector3(1.08, 0.48, 0.70) * scale, Vector3(0.0, 0.29 * scale, 0.0), wood)
	_add_chest_part(root, "ChestFrontPanel", Vector3(0.92, 0.30, 0.045) * scale, Vector3(0.0, 0.30 * scale, 0.374 * scale), dark_wood)
	_add_chest_part(root, "ChestBackPanel", Vector3(0.92, 0.30, 0.045) * scale, Vector3(0.0, 0.30 * scale, -0.374 * scale), dark_wood)
	_add_chest_part(root, "ChestLeftPanel", Vector3(0.045, 0.30, 0.56) * scale, Vector3(-0.564 * scale, 0.30 * scale, 0.0), dark_wood)
	_add_chest_part(root, "ChestRightPanel", Vector3(0.045, 0.30, 0.56) * scale, Vector3(0.564 * scale, 0.30 * scale, 0.0), dark_wood)
	_add_chest_part(root, "ChestFrontBand", Vector3(1.18, 0.08, 0.055) * scale, Vector3(0.0, 0.48 * scale, 0.405 * scale), metal)
	_add_chest_part(root, "ChestBottomBand", Vector3(1.18, 0.075, 0.055) * scale, Vector3(0.0, 0.13 * scale, 0.405 * scale), metal_dark)
	for x in [-0.42, 0.42]:
		_add_chest_part(root, "ChestVerticalBand", Vector3(0.085, 0.54, 0.085) * scale, Vector3(x * scale, 0.33 * scale, 0.405 * scale), metal)
	for x in [-0.43, 0.43]:
		for z in [-0.28, 0.28]:
			_add_chest_part(root, "ChestFoot", Vector3(0.22, 0.10, 0.16) * scale, Vector3(x * scale, 0.055 * scale, z * scale), metal_dark)

	var lid_pivot := Node3D.new()
	lid_pivot.name = "ChestLidPivot"
	lid_pivot.position = Vector3(0.0, 0.56 * scale, -0.36 * scale)
	root.add_child(lid_pivot)
	_add_chest_part(lid_pivot, "ChestLid", Vector3(1.16, 0.30, 0.72) * scale, Vector3(0.0, 0.15 * scale, 0.36 * scale), wood)
	_add_chest_part(lid_pivot, "ChestLidCrown", Vector3(0.92, 0.12, 0.58) * scale, Vector3(0.0, 0.33 * scale, 0.36 * scale), Color("#8a4f20") if is_stash else Color("#8b511f"))
	_add_chest_part(lid_pivot, "ChestLidFrontBand", Vector3(1.22, 0.08, 0.075) * scale, Vector3(0.0, 0.13 * scale, 0.74 * scale), metal)
	for x in [-0.42, 0.42]:
		_add_chest_part(lid_pivot, "ChestLidStrap", Vector3(0.075, 0.36, 0.80) * scale, Vector3(x * scale, 0.20 * scale, 0.36 * scale), metal)

	_add_chest_part(root, "ChestLockPlate", Vector3(0.22, 0.24, 0.075) * scale, Vector3(0.0, 0.42 * scale, 0.452 * scale), metal)
	_add_chest_part(root, "ChestLockSlot", Vector3(0.075, 0.11, 0.085) * scale, Vector3(0.0, 0.40 * scale, 0.502 * scale), metal_dark)
	_add_chest_part(root, "ChestLeftHandle", Vector3(0.075, 0.22, 0.30) * scale, Vector3(-0.64 * scale, 0.36 * scale, 0.0), metal)
	_add_chest_part(root, "ChestRightHandle", Vector3(0.075, 0.22, 0.30) * scale, Vector3(0.64 * scale, 0.36 * scale, 0.0), metal)
	if is_stash:
		_add_chest_part(root, "ChestStashCrest", Vector3(0.36, 0.12, 0.082) * scale, Vector3(0.0, 0.61 * scale, 0.456 * scale), cloth)

	var inner := _add_chest_part(root, "ChestInnerGlow", Vector3(0.84, 0.045, 0.46) * scale, Vector3(0.0, 0.57 * scale, 0.02 * scale), glow)
	var glow_mat := inner.material_override as StandardMaterial3D
	glow_mat.emission_enabled = true
	glow_mat.emission = glow
	inner.visible = false
	return root


func _add_chest_part(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	part.material_override = mat
	parent.add_child(part)
	return part


func _make_merchant_node(def_id: String) -> Node3D:
	var is_mystery := def_id == "town_mystery_seller"
	var root := Node3D.new()
	root.name = "MysterySeller" if is_mystery else "ShopVendor"
	var cloth := Color("#2b124a") if is_mystery else Color("#e2b92e")
	var cloth_dark := Color("#150824") if is_mystery else Color("#8c6418")
	var accent := Color("#7c4dff") if is_mystery else Color("#ffe37a")
	var trim := Color("#c7a6ff") if is_mystery else Color("#f6d85c")
	var wood := Color("#4b2a16")
	var skin := Color("#c1845a") if is_mystery else Color("#c99666")
	var glow := Color("#6f40ff") if is_mystery else Color("#ffd85a")

	_add_merchant_box(root, "MerchantShadow", Vector3(1.35, 0.035, 0.92), Vector3(0.0, 0.018, 0.0), Color("#181715"))
	_add_merchant_box(root, "CounterTop", Vector3(1.12, 0.18, 0.48), Vector3(0.0, 0.39, 0.31), wood)
	_add_merchant_box(root, "CounterFront", Vector3(1.18, 0.36, 0.09), Vector3(0.0, 0.22, 0.57), cloth_dark)
	_add_merchant_box(root, "CounterTrim", Vector3(1.24, 0.08, 0.10), Vector3(0.0, 0.44, 0.60), trim)
	_add_merchant_box(root, "CounterLeftLeg", Vector3(0.12, 0.34, 0.12), Vector3(-0.50, 0.17, 0.36), wood)
	_add_merchant_box(root, "CounterRightLeg", Vector3(0.12, 0.34, 0.12), Vector3(0.50, 0.17, 0.36), wood)

	_add_merchant_box(root, "Body", Vector3(0.42, 0.66, 0.30), Vector3(0.0, 0.78, -0.05), cloth)
	_add_merchant_box(root, "Belt", Vector3(0.48, 0.09, 0.34), Vector3(0.0, 0.62, -0.04), Color("#2f2117"))
	_add_merchant_box(root, "Head", Vector3(0.32, 0.30, 0.30), Vector3(0.0, 1.28, -0.03), skin)
	_add_merchant_box(root, "Nose", Vector3(0.08, 0.08, 0.08), Vector3(0.0, 1.27, 0.16), skin.lerp(Color.WHITE, 0.12))
	_add_merchant_box(root, "LeftArm", Vector3(0.12, 0.44, 0.16), Vector3(-0.32, 0.78, 0.02), cloth_dark)
	_add_merchant_box(root, "RightArm", Vector3(0.12, 0.44, 0.16), Vector3(0.32, 0.78, 0.02), cloth_dark)
	_add_merchant_box(root, "LeftHand", Vector3(0.13, 0.10, 0.13), Vector3(-0.32, 0.53, 0.08), skin)
	_add_merchant_box(root, "RightHand", Vector3(0.13, 0.10, 0.13), Vector3(0.32, 0.53, 0.08), skin)

	if is_mystery:
		_add_merchant_box(root, "Hood", Vector3(0.44, 0.22, 0.36), Vector3(0.0, 1.39, -0.02), cloth_dark)
		_add_merchant_box(root, "HoodRim", Vector3(0.40, 0.08, 0.08), Vector3(0.0, 1.30, 0.18), accent)
		_add_merchant_cylinder(root, "CrystalOrb", 0.16, 0.22, Vector3(0.34, 0.62, 0.43), glow, true)
		_add_merchant_box(root, "MysterySign", Vector3(0.34, 0.24, 0.07), Vector3(-0.36, 0.67, 0.62), accent)
	else:
		_add_merchant_box(root, "HatBrim", Vector3(0.54, 0.08, 0.42), Vector3(0.0, 1.43, -0.02), accent)
		_add_merchant_box(root, "HatCrown", Vector3(0.34, 0.22, 0.30), Vector3(0.0, 1.56, -0.02), cloth)
		_add_merchant_box(root, "CoinStackA", Vector3(0.16, 0.08, 0.16), Vector3(-0.30, 0.54, 0.43), glow)
		_add_merchant_box(root, "CoinStackB", Vector3(0.14, 0.14, 0.14), Vector3(-0.12, 0.57, 0.42), glow)
		_add_merchant_box(root, "VendorSign", Vector3(0.34, 0.24, 0.07), Vector3(0.36, 0.67, 0.62), accent)
	return root


func _make_bishop_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TownBishop"
	var robe := Color("#b92d2d")
	var robe_dark := Color("#621717")
	var trim := Color("#f3d7a8")
	var skin := Color("#c99666")
	var gold_trim := Color("#d8a342")

	_add_merchant_box(root, "BishopShadow", Vector3(0.92, 0.035, 0.78), Vector3(0.0, 0.018, 0.0), Color("#171313"))
	_add_merchant_box(root, "RobeLower", Vector3(0.54, 0.62, 0.36), Vector3(0.0, 0.46, 0.0), robe_dark)
	_add_merchant_box(root, "RobeUpper", Vector3(0.46, 0.58, 0.30), Vector3(0.0, 0.92, 0.0), robe)
	_add_merchant_box(root, "Sash", Vector3(0.12, 0.68, 0.34), Vector3(0.0, 0.78, 0.02), trim)
	_add_merchant_box(root, "Shoulders", Vector3(0.62, 0.12, 0.34), Vector3(0.0, 1.16, 0.0), robe_dark)
	_add_merchant_box(root, "Head", Vector3(0.30, 0.30, 0.28), Vector3(0.0, 1.42, 0.0), skin)
	_add_merchant_box(root, "MitreBase", Vector3(0.38, 0.16, 0.28), Vector3(0.0, 1.64, 0.0), robe)
	_add_merchant_box(root, "MitrePeak", Vector3(0.24, 0.24, 0.20), Vector3(0.0, 1.84, 0.0), robe)
	_add_merchant_box(root, "MitreTrim", Vector3(0.42, 0.06, 0.30), Vector3(0.0, 1.56, 0.0), trim)
	_add_merchant_box(root, "LeftArm", Vector3(0.12, 0.50, 0.14), Vector3(-0.36, 0.86, 0.02), robe_dark)
	_add_merchant_box(root, "RightArm", Vector3(0.12, 0.50, 0.14), Vector3(0.36, 0.86, 0.02), robe_dark)
	_add_merchant_box(root, "LeftHand", Vector3(0.12, 0.10, 0.12), Vector3(-0.36, 0.56, 0.06), skin)
	_add_merchant_box(root, "RightHand", Vector3(0.12, 0.10, 0.12), Vector3(0.36, 0.56, 0.06), skin)
	_add_merchant_cylinder(root, "Staff", 0.035, 1.18, Vector3(0.55, 0.82, 0.02), Color("#5a351c"))
	_add_merchant_cylinder(root, "StaffCrown", 0.11, 0.09, Vector3(0.55, 1.44, 0.02), gold_trim, true)
	_add_merchant_box(root, "ServiceBook", Vector3(0.30, 0.08, 0.22), Vector3(-0.18, 0.58, 0.24), Color("#efe0bc"))
	return root


func _make_market_board_node() -> Node3D:
	var root := Node3D.new()
	root.name = "MarketBoard"
	_add_merchant_box(root, "MarketBoardShadow", Vector3(1.70, 0.035, 0.52), Vector3(0.0, 0.018, 0.03), Color("#171513"))
	_add_merchant_box(root, "MarketBoardLeftPost", Vector3(0.13, 1.24, 0.13), Vector3(-0.68, 0.62, 0.0), Color("#4a2b17"))
	_add_merchant_box(root, "MarketBoardRightPost", Vector3(0.13, 1.24, 0.13), Vector3(0.68, 0.62, 0.0), Color("#4a2b17"))
	_add_merchant_box(root, "MarketBoardPanel", Vector3(1.34, 0.82, 0.12), Vector3(0.0, 0.88, 0.02), Color("#6f431f"))
	_add_merchant_box(root, "MarketBoardInset", Vector3(1.12, 0.58, 0.135), Vector3(0.0, 0.88, 0.09), Color("#2c2118"))
	_add_merchant_box(root, "MarketBoardHeader", Vector3(1.44, 0.16, 0.14), Vector3(0.0, 1.38, 0.04), Color("#c99d4a"))
	_add_merchant_box(root, "MarketBoardPaperA", Vector3(0.30, 0.36, 0.145), Vector3(-0.24, 0.83, 0.16), Color("#efe0bc"))
	_add_merchant_box(root, "MarketBoardPaperB", Vector3(0.27, 0.30, 0.145), Vector3(0.22, 0.96, 0.16), Color("#d7c99e"))
	_add_merchant_box(root, "MarketBoardSeal", Vector3(0.13, 0.13, 0.16), Vector3(0.48, 0.70, 0.17), Color("#b93131"))

	root.add_child(_make_market_badge("IncomingBidBadge", "IncomingBidCount", Vector3(-0.58, 1.42, 0.20), Color("#4f2b12"), Color("#776d5e")))
	root.add_child(_make_market_badge("PublishedListingBadge", "PublishedListingCount", Vector3(0.58, 1.42, 0.20), Color("#14324f"), Color("#776d5e")))
	return root


func make_town_preview_scene() -> Node3D:
	var root := Node3D.new()
	root.name = "TownPreview"

	var ground := _make_ground_node()
	ground.name = "TownPreviewGround"
	ground.position = Vector3(11.5, -0.02, 11.5)
	var ground_mesh := ground.mesh as PlaneMesh
	if ground_mesh != null:
		ground_mesh.size = Vector2(28.0, 22.0)
	root.add_child(ground)

	var service_entries := [
		{"def_id": "stairs_down", "position": Vector3(11.0, 0.0, 8.0)},
		{"def_id": "teleporter", "position": Vector3(2.0, 0.0, 12.0)},
		{"def_id": "town_vendor", "position": Vector3(17.0, 0.0, 10.0)},
		{"def_id": "town_mystery_seller", "position": Vector3(18.0, 0.0, 15.0)},
		{"def_id": "town_stash", "position": Vector3(7.0, 0.0, 14.0)},
		{"def_id": "town_bishop", "position": Vector3(15.0, 0.0, 6.0)},
		{"def_id": "town_market_board", "position": Vector3(10.0, 0.0, 18.0)},
	]
	for entry in service_entries:
		var service := _make_entity_node({"type": "interactable", "interactable_def_id": str(entry["def_id"])})
		service.name = "TownService_%s" % str(entry["def_id"])
		service.position = entry["position"] as Vector3
		root.add_child(service)

	var cabin_a := _make_town_cabin_node("west")
	cabin_a.name = "TownCabinWest"
	cabin_a.position = Vector3(5.0, 0.0, 7.0)
	cabin_a.rotation_degrees.y = -22.0
	root.add_child(cabin_a)
	var cabin_b := _make_town_cabin_node("east")
	cabin_b.name = "TownCabinEast"
	cabin_b.position = Vector3(21.0, 0.0, 12.5)
	cabin_b.rotation_degrees.y = 18.0
	root.add_child(cabin_b)

	var fire := _make_town_campfire_node()
	fire.position = Vector3(12.0, 0.0, 13.0)
	root.add_child(fire)
	return root


func _make_town_cabin_node(variant: String = "plain") -> Node3D:
	var root := Node3D.new()
	root.name = "TownCabin"
	var wood := Color("#6d3f1f")
	var dark_wood := Color("#3b2113")
	var roof := Color("#5b2b1d") if variant == "west" else Color("#69401f")
	var straw := Color("#b18a4a")
	_add_merchant_box(root, "CabinShadow", Vector3(2.45, 0.035, 1.85), Vector3(0.0, 0.018, 0.0), Color("#17130f"))
	_add_merchant_box(root, "CabinBody", Vector3(1.78, 1.02, 1.28), Vector3(0.0, 0.54, 0.0), wood)
	_add_merchant_box(root, "CabinFront", Vector3(1.86, 0.74, 0.08), Vector3(0.0, 0.46, 0.68), dark_wood)
	_add_merchant_box(root, "CabinDoor", Vector3(0.42, 0.62, 0.10), Vector3(-0.36, 0.35, 0.74), Color("#2b1710"))
	_add_merchant_box(root, "CabinWindow", Vector3(0.34, 0.30, 0.11), Vector3(0.38, 0.58, 0.75), Color("#d7ad58"))
	_add_merchant_box(root, "CabinRoofA", Vector3(2.15, 0.34, 1.58), Vector3(0.0, 1.15, -0.14), roof)
	_add_merchant_box(root, "CabinRoofRidge", Vector3(2.28, 0.18, 0.22), Vector3(0.0, 1.42, 0.0), straw)
	for x in [-0.76, 0.0, 0.76]:
		_add_merchant_box(root, "CabinWallLog", Vector3(0.08, 1.04, 1.36), Vector3(x, 0.56, 0.0), Color("#4f2d18"))
	return root


func _make_town_campfire_node() -> Node3D:
	var root := Node3D.new()
	root.name = "TownCampfire"
	_add_merchant_cylinder(root, "FireStoneRing", 0.58, 0.08, Vector3(0.0, 0.04, 0.0), Color("#4e4d48"))
	for i in range(6):
		var angle := TAU * float(i) / 6.0
		var stone := _add_merchant_box(root, "FireStone%d" % i, Vector3(0.20, 0.11, 0.16), Vector3(cos(angle) * 0.47, 0.10, sin(angle) * 0.47), Color("#777067"))
		stone.rotation_degrees.y = rad_to_deg(angle)
	for i in range(3):
		var log := _add_merchant_box(root, "FireLog%d" % i, Vector3(0.72, 0.11, 0.16), Vector3(0.0, 0.17 + float(i) * 0.025, 0.0), Color("#4b2815"))
		log.rotation_degrees.y = 60.0 * float(i)
	var flame_outer := _add_merchant_cylinder(root, "FireFlameOuter", 0.24, 0.62, Vector3(0.0, 0.52, 0.0), Color("#ff7a1a"), true)
	flame_outer.scale.x = 0.58
	flame_outer.scale.z = 0.58
	var flame_inner := _add_merchant_cylinder(root, "FireFlameInner", 0.14, 0.44, Vector3(0.0, 0.58, 0.0), Color("#ffd45a"), true)
	flame_inner.scale.x = 0.52
	flame_inner.scale.z = 0.52
	var light := OmniLight3D.new()
	light.name = "CampfireLight"
	light.light_color = Color("#ff9b3d")
	light.light_energy = 1.8
	light.omni_range = 4.0
	light.position = Vector3(0.0, 0.78, 0.0)
	root.add_child(light)
	return root


func _make_market_badge(badge_name: String, count_name: String, position: Vector3, bg_color: Color, text_color: Color) -> Node3D:
	var badge := Node3D.new()
	badge.name = badge_name
	badge.position = position
	_add_merchant_box(badge, "BadgeBack", Vector3(0.34, 0.24, 0.055), Vector3.ZERO, bg_color)
	var label := Label3D.new()
	label.name = count_name
	label.text = "0"
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 42
	label.modulate = text_color
	label.outline_size = 8
	label.outline_modulate = Color("#14110d")
	label.position = Vector3(0.0, -0.055, 0.04)
	label.pixel_size = 0.006
	badge.add_child(label)
	return badge


func _add_merchant_box(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	part.material_override = _merchant_material(color)
	parent.add_child(part)
	return part


func _add_merchant_cylinder(parent: Node3D, part_name: String, radius: float, height: float, position: Vector3, color: Color, emit: bool = false) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := CylinderMesh.new()
	mesh.top_radius = radius
	mesh.bottom_radius = radius
	mesh.height = height
	mesh.radial_segments = 24
	part.mesh = mesh
	part.position = position
	part.material_override = _merchant_material(color, emit)
	parent.add_child(part)
	return part


func _merchant_material(color: Color, emit: bool = false) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	if emit:
		mat.emission_enabled = true
		mat.emission = color
	return mat


func _make_stair_node(def_id: String) -> Node3D:
	var root := Node3D.new()
	root.name = "Stairs_%s" % def_id
	var is_down := def_id == "stairs_down"
	if is_down:
		_add_stair_box(root, "DownStoneFrame", Vector3(1.46, 0.16, 1.20), Vector3(0.0, 0.08, 0.0), _stair_base_color(def_id, "ready"))
		_add_stair_box(root, "DownPitOpening", Vector3(1.02, 0.06, 0.76), Vector3(0.0, 0.185, -0.04), Color("#030507"))
		_add_stair_box(root, "DownBackWall", Vector3(1.02, 0.38, 0.12), Vector3(0.0, 0.34, -0.46), Color("#10151c"))
		_add_stair_box(root, "DownLeftWall", Vector3(0.12, 0.30, 0.76), Vector3(-0.57, 0.30, -0.04), Color("#18202a"))
		for i in range(5):
			var t := float(i) / 4.0
			_add_stair_box(
				root,
				"DownStep%d" % i,
				Vector3(0.84 - t * 0.12, 0.09, 0.16),
				Vector3(0.0, 0.34 - t * 0.22, 0.28 - t * 0.54),
				Color("#66707a").lerp(Color("#10151d"), t)
			)
		_add_stair_box(root, "DownThreshold", Vector3(1.20, 0.10, 0.14), Vector3(0.0, 0.24, 0.42), Color("#89909a"))
	else:
		_add_stair_box(root, "UpGroundPad", Vector3(1.42, 0.14, 1.14), Vector3(0.0, 0.07, 0.0), _stair_base_color(def_id, "ready"))
		_add_stair_box(root, "UpHighLanding", Vector3(0.64, 0.16, 0.46), Vector3(0.0, 0.72, -0.45), Color("#d6cfaa"))
		_add_stair_box(root, "UpBackWall", Vector3(0.78, 0.42, 0.12), Vector3(0.0, 0.54, -0.72), Color("#aaa78c"))
		for i in range(5):
			var t := float(i) / 4.0
			_add_stair_box(
				root,
				"UpStep%d" % i,
				Vector3(1.12 - t * 0.42, 0.12, 0.20),
				Vector3(0.0, 0.19 + t * 0.45, 0.36 - t * 0.66),
				Color("#7c817a").lerp(Color("#d3cda9"), t)
			)
	return root


func _add_stair_box(parent: Node3D, part_name: String, size: Vector3, position: Vector3, color: Color) -> MeshInstance3D:
	var part := MeshInstance3D.new()
	part.name = part_name
	var mesh := BoxMesh.new()
	mesh.size = size
	part.mesh = mesh
	part.position = position
	part.material_override = _stair_material(color)
	parent.add_child(part)
	return part


func _stair_base_color(def_id: String, state: String) -> Color:
	if state == "locked" or state == "disabled":
		return Color("#6b2e2e")
	if def_id == "stairs_down":
		return Color("#111821")
	return Color("#666d68")


func _stair_material(color: Color, glow: bool = false) -> StandardMaterial3D:
	var mat := StandardMaterial3D.new()
	mat.albedo_color = color
	if glow:
		mat.emission_enabled = true
		mat.emission = color
	return mat


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
	_apply_interactable_state_tint(rec, state)
	var chest_pivot := node.find_child("ChestLidPivot", true, false) as Node3D
	if chest_pivot != null:
		var chest_rot := deg_to_rad(-68.0) if state == "open" else 0.0
		var chest_tween := create_tween()
		chest_tween.tween_property(chest_pivot, "rotation:x", chest_rot, 0.22)
		return
	var pivot := node.find_child("DoorPivot", true, false) as Node3D
	if pivot == null:
		return
	var target_rot := deg_to_rad(90.0) if state == "open" else 0.0
	var tween := create_tween()
	tween.tween_property(pivot, "rotation:y", target_rot, 0.25)


func _apply_interactable_state_tint(rec: Dictionary, state: String) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var def_id := str(rec.get("interactable_def_id", ""))
	if def_id == "treasure_chest" or def_id == "town_stash":
		var glow := node.find_child("ChestInnerGlow", true, false) as MeshInstance3D
		if glow != null:
			glow.visible = state == "open"
		var lock := node.find_child("ChestLockPlate", true, false) as MeshInstance3D
		if lock != null:
			var lock_mat := StandardMaterial3D.new()
			if state == "locked" or state == "disabled":
				lock_mat.albedo_color = Color("#7a2f2d")
			elif state == "open":
				lock_mat.albedo_color = Color("#f0cf72")
				lock_mat.emission_enabled = true
				lock_mat.emission = Color("#8a6122")
			else:
				lock_mat.albedo_color = Color("#d1b15d") if def_id == "town_stash" else Color("#8d8f8f")
			lock.material_override = lock_mat
		return
	if def_id == "teleporter":
		var core := node.get_child(1) as MeshInstance3D if node.get_child_count() > 1 else null
		if core == null:
			return
		var mat := StandardMaterial3D.new()
		if state == "disabled" or state == "locked":
			mat.albedo_color = Color(0.30, 0.16, 0.18)
			mat.emission_enabled = false
		else:
			mat.albedo_color = Color(0.15, 0.62, 0.70)
			mat.emission_enabled = true
			mat.emission = Color(0.05, 0.55, 0.68)
		core.material_override = mat
		return
	if def_id == "stairs_down" or def_id == "stairs_up":
		var base := node.get_child(0) as MeshInstance3D if node.get_child_count() > 0 else null
		if base == null:
			return
		var mat := StandardMaterial3D.new()
		if state == "locked" or state == "disabled":
			mat.albedo_color = _stair_base_color(def_id, state)
		else:
			mat.albedo_color = _stair_base_color(def_id, state)
		base.material_override = mat


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
		"last_tick": last_server_tick,
		"local_player_id": player_id,
		"party": party.duplicate(true),
		"remote_player_ids": _remote_player_ids(),
		"player_hp": player_hp,
		"player_max_hp": player_max_hp,
		"player_mana": player_mana,
		"player_max_mana": player_max_mana,
		"gold": gold,
		"player_pos": {"x": predicted_pos.x, "z": predicted_pos.z},
		"current_level": current_level,
		"walls": current_wall_layout.duplicate(true),
		"wall_count": current_wall_layout.size(),
		"generated_wall_count": _wall_count_by_source("generated"),
		"non_perimeter_wall_count": _non_perimeter_wall_count(),
		"character_progression": character_progression.duplicate(true),
		"skill_progression": skill_progression.duplicate(true),
		"skill_cooldowns": skill_cooldowns.duplicate(true),
		"right_click_skill_id": right_click_skill_id,
		"skill_function_keys": skill_function_keys.duplicate(true),
		"inventory": inventory.duplicate(true),
		"equipped": equipped.duplicate(true),
		"inventory_rows": inventory_rows,
		"inventory_capacity": inventory_capacity,
		"stash_items": stash_items.duplicate(true),
		"stash_gold": stash_gold,
		"stash_capacity": stash_capacity,
		"monster_ids": live_monster_ids,
		"entities_debug": _bot_entities_debug(live_monster_ids),
		"local_player_presentation": _bot_local_player_presentation(),
		"entities_presentation_debug": _bot_entities_presentation_debug(),
		"loot_ids": loot_ids.duplicate(),
		"loot": _bot_loot_debug(),
		"loot_labels": _bot_loot_label_debug(),
		"interactable_ids": interactable_ids.duplicate(),
		"loot_presentations": _bot_loot_presentations(),
		"inventory_panel_visible": inventory_panel != null and inventory_panel.visible,
		"shop_panel_visible": shop_panel != null and shop_panel.visible,
		"stash_panel_visible": stash_panel != null and stash_panel.visible,
		"bishop_panel_visible": bishop_panel != null and bishop_panel.visible,
		"market_panel_visible": market_panel != null and market_panel.visible,
		"character_stats_panel_visible": character_stats_panel != null and character_stats_panel.visible,
		"skills_panel_visible": skills_panel != null and skills_panel.visible,
		"character_info_panel_visible": character_info_panel != null and character_info_panel.visible,
		"waypoint_panel_visible": waypoint_panel != null and waypoint_panel.visible,
		"inventory_panel": inventory_panel.get_debug_state() if inventory_panel != null else {},
		"shop_panel": shop_panel.get_debug_state() if shop_panel != null else {},
		"stash_panel": stash_panel.get_debug_state() if stash_panel != null else {},
		"bishop_panel": bishop_panel.get_debug_state() if bishop_panel != null else {},
		"market_panel": market_panel.get_debug_state() if market_panel != null else {},
		"character_stats_panel": character_stats_panel.get_debug_state() if character_stats_panel != null else {},
		"skills_panel": skills_panel.get_debug_state() if skills_panel != null else {},
		"character_bar": character_bar.get_debug_state() if character_bar != null else {},
		"skill_bar": skill_bar.get_debug_state() if skill_bar != null else {},
		"status_effects_bar": status_effects_bar.get_debug_state() if status_effects_bar != null else {"effects": [], "visible": false},
		"boss_health_bar": boss_health_bar.get_debug_state() if boss_health_bar != null else {"visible": false},
		"character_info_panel": _character_info_debug_state(),
		"consumable_bar": consumable_bar.get_debug_state() if consumable_bar != null else {},
		"pending_events": _bot_pending_events.duplicate(true),
		"account_email": _account_email(),
		"main_menu_visible": main_menu != null and main_menu.visible,
		"main_menu_button_labels": main_menu.button_labels() if main_menu != null else [],
		"main_menu_actions": main_menu.available_actions() if main_menu != null else [],
		"character_panel_visible": character_panel != null and character_panel.visible,
		"character_panel": character_panel.get_debug_state() if character_panel != null else {},
		"character_panel_mode": character_panel.mode() if character_panel != null else "",
		"character_panel_title": character_panel.title_text() if character_panel != null else "",
		"multiplayer_panel_visible": multiplayer_panel != null and multiplayer_panel.visible,
		"settings_panel_visible": settings_panel != null and settings_panel.visible,
		"pause_menu_visible": pause_menu != null and pause_menu.visible,
		"loss_popup_visible": loss_popup != null and loss_popup.visible,
		"selected_window_size": ClientSettingsScript.size_label(client_settings.window_size) if client_settings != null else "",
		"floating_combat_text_enabled": client_settings != null and client_settings.floating_combat_text,
		"status_text_enabled": client_settings != null and client_settings.status_text,
		"language": client_settings.language if client_settings != null else ClientSettingsScript.DEFAULT_LANGUAGE,
		"boss_reward_status": _last_boss_reward_status,
		"create_game_session_type": client_settings.create_game_session_type if client_settings != null else ClientSettingsScript.DEFAULT_CREATE_GAME_SESSION_TYPE,
		"damage_numbers": _bot_damage_numbers(),
		"known_characters": character_panel.known_characters() if character_panel != null else [],
		"multiplayer_panel": multiplayer_panel.get_debug_state() if multiplayer_panel != null else {},
		"join_game_selected_session_id": pending_join_session_id if pending_join_session_id != "" else (str(multiplayer_panel.get_debug_state().get("selected_session_id", "")) if multiplayer_panel != null else ""),
		"current_session_id": client.session_id if client != null else "",
		"current_session_mode": client.session_mode if client != null else "",
		"current_session_listed": client.session_listed if client != null else false,
		"gameplay_active": gameplay_active,
	}
	return out


func _wall_count_by_source(source: String) -> int:
	var count := 0
	for wall in current_wall_layout:
		if typeof(wall) == TYPE_DICTIONARY and str((wall as Dictionary).get("source", "")) == source:
			count += 1
	return count


func _non_perimeter_wall_count() -> int:
	var count := 0
	for wall in current_wall_layout:
		if typeof(wall) != TYPE_DICTIONARY:
			continue
		if str((wall as Dictionary).get("source", "")) != "perimeter":
			count += 1
	return count


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
			"item_template_id": str(rec.get("item_template_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"state": str(rec.get("state", "")),
		})
	return out


func _bot_local_player_presentation() -> Dictionary:
	return {
		"id": player_id,
		"type": "player",
		"visual_model": "character",
		"visual_scale": player_visual_scale,
		"effect_ids": _local_player_effect_ids(),
		"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor),
		"holy_shield_aura_pulses": PlayerStatusEffectMarkers.active_holy_shield_aura_pulse_count(player_anchor),
		"holy_shield_target_pulses": PlayerStatusEffectMarkers.active_holy_shield_target_pulse_count(player_anchor),
		"has_rage_effect": PlayerStatusEffectMarkers.has_rage_effect(player_anchor),
		"base_tint": PLAYER_TINT.to_html(false),
		"reaction": player_reaction.get_debug_state() if player_reaction != null else {},
		"animation": player_anim.get_debug_state() if player_anim != null else {},
	}


func _local_player_effect_ids() -> Array:
	if player_anchor == null or not PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor):
		return []
	return [PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID]


func _character_info_debug_state() -> Dictionary:
	return {
		"visible": character_info_panel != null and character_info_panel.visible,
		"name": _local_character_display_name(),
		"level": int(character_progression.get("level", 1)),
		"area": _current_area_label(),
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
			"visual_scale": float(rec.get("visual_scale", 1.0)),
			"is_boss": bool(rec.get("is_boss", false)),
			"boss_template_id": str(rec.get("boss_template_id", "")),
			"boss_phase": rec.get("boss_phase", {}),
			"boss_telegraph_active": bool(rec.get("boss_telegraph_active", false)),
			"telegraph_tint": str(rec.get("telegraph_tint", "")),
			"has_boss_telegraph_marker": bool(rec.get("has_boss_telegraph_marker", false)),
			"telegraph_radius": float(rec.get("telegraph_radius", 0.0)),
			"telegraph_marker_color": str(rec.get("telegraph_marker_color", "")),
			"base_tint": str(rec.get("base_tint", "")),
			"has_bow_marker": bool(rec.get("has_bow_marker", false)),
			"effect_ids": rec.get("effect_ids", []),
			"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(node),
			"has_burning_effect": PlayerStatusEffectMarkers.has_burning_effect(node),
			"holy_shield_target_pulses": PlayerStatusEffectMarkers.active_holy_shield_target_pulse_count(node),
			"hp": int(rec.get("hp", 1)),
			"reaction": reaction.get_debug_state() if reaction != null else {},
			"animation": controller.get_debug_state() if controller != null else {},
		})
	return out


func _visual_model_name(rec: Dictionary, node: Node3D) -> String:
	if str(rec.get("visual_model", "")) != "":
		return str(rec.get("visual_model", ""))
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
				"color": pop.label_settings.font_color.to_html(false) if pop.label_settings != null else "",
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
			"item_template_id": str(rec.get("item_template_id", "")),
			"rarity": str(rec.get("rarity", "")),
		})
	return out


func _bot_loot_label_debug() -> Array:
	var out: Array = []
	for loot_id in loot_ids:
		var id := str(loot_id)
		var rec: Dictionary = entities.get(id, {})
		var label := _loot_label_node(id)
		out.append({
			"id": id,
			"item_def_id": str(rec.get("item_def_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"text": label.text if label != null else "",
			"visible": label.visible if label != null else false,
			"color": label.modulate.to_html(false) if label != null else "",
		})
	return out


func bot_dispatch_action(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if _movement_intent_starts_motion(intent_type, payload):
		_close_gameplay_panels_for_movement()
		_mark_local_player_walking()
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
	if typ == "interactable" and _interactable_should_approach_before_action(interactable_def_id):
		_activate_or_approach_interactable(target_id, rec)
		return
	_send_action_intent(target_id)
	_attack_cooldown = _basic_attack_cooldown_seconds() if typ == "monster" else SEND_INTERVAL


func bot_dispatch_inventory_intent(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send(intent_type, last_server_tick, payload)


func bot_click_shop_buy_offer(offer_id: String = "", offer_kind: String = "", offer_index: int = 0) -> void:
	if shop_panel == null:
		return
	shop_panel.bot_click_buy_offer(offer_id, offer_kind, offer_index)


func bot_click_shop_sell_item(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	if shop_panel == null:
		return
	shop_panel.bot_click_sell_item(item_def_id, rolled, bag_index)


func bot_click_shop_reroll() -> void:
	if shop_panel == null:
		return
	shop_panel.bot_click_reroll()


func bot_drag_bag_to_stash(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_drag_bag_to_stash(item_def_id, rolled, bag_index)


func bot_drag_stash_to_bag(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_drag_stash_to_bag(stash_item_id, item_def_id, rolled, stash_index)


func bot_click_stash_deposit_gold(amount: int = 1) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_click_deposit_gold(amount)


func bot_click_stash_withdraw_gold(amount: int = 1) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_click_withdraw_gold(amount)


func bot_click_bishop_respec() -> void:
	if bishop_panel == null:
		return
	bishop_panel.bot_click_respec()


func bot_set_stash_search(text: String) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_set_search_text(text)


func bot_select_stash_sort(mode: String) -> void:
	if stash_panel == null:
		return
	stash_panel.bot_select_sort_mode(mode)


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


func bot_click_skill_button(skill_id: String = "") -> void:
	if skills_panel == null:
		return
	skills_panel.bot_click_skill_button(skill_id)


func bot_use_skill_bar(skill_id: String = "", target_id: String = "", force_direct: bool = false) -> void:
	if skill_id == "":
		skill_id = SkillRulesLoader.first_skill_id()
	if force_direct or target_id != "":
		_send_skill_cast_intent(skill_id, target_id)
		return
	if skill_id != "" and _skill_rank(skill_id) > 0:
		right_click_skill_id = skill_id
		if skill_bar != null:
			skill_bar.set_skill_id(skill_id)
			skill_bar.set_skill_progression(skill_progression)
			skill_bar.set_skill_cooldowns(skill_cooldowns)
		_sync_skill_bindings_ui()
	if skill_bar != null:
		skill_bar.use_slot()


func bot_cast_skill_direction(skill_id: String = "", direction: Dictionary = {}) -> void:
	if skill_id == "":
		skill_id = SkillRulesLoader.first_skill_id()
	if _skill_cast_blocked(skill_id):
		return
	var dir := Vector2(float(direction.get("x", _last_facing_direction.x)), float(direction.get("y", _last_facing_direction.y)))
	if dir.length_squared() <= 0.0001:
		dir = Vector2(1.0, 0.0)
	dir = dir.normalized()
	_face_direction(dir)
	client.send("cast_skill_intent", last_server_tick, {"skill_id": skill_id, "direction": {"x": dir.x, "y": dir.y}})
	_attack_cooldown = SEND_INTERVAL
	if player_anim != null:
		player_anim.play_one_shot("attack")


func bot_click_menu_button(button: String) -> void:
	match button:
		"create_game":
			_on_create_game_pressed()
		"join_game":
			_on_join_game_pressed()
		"continue":
			_on_create_game_pressed()
		"new_game":
			_on_create_game_pressed()
		"multiplayer":
			_on_join_game_pressed()
		"refresh_sessions":
			_refresh_multiplayer_sessions()
		"host_listed_session":
			_on_host_listed_session_requested()
		"join_first_listed_session":
			if multiplayer_panel != null:
				multiplayer_panel.join_first_session()
		"select_expected_join_session":
			if multiplayer_panel != null:
				multiplayer_panel.select_session(OS.get_environment("ARPG_EXPECTED_JOIN_SESSION_ID"))
		"join_expected_session":
			if multiplayer_panel != null:
				multiplayer_panel.join_session(OS.get_environment("ARPG_EXPECTED_JOIN_SESSION_ID"))
		"settings":
			if pause_menu != null and pause_menu.visible:
				_on_settings_from_pause()
			else:
				_on_settings_from_main()
		"back":
			if settings_panel != null and settings_panel.visible:
				_on_settings_back()
			elif character_panel != null and character_panel.visible:
				_on_character_panel_back()
			elif multiplayer_panel != null and multiplayer_panel.visible:
				multiplayer_panel.hide_panel()
				main_menu.show_menu()
		"create_character", "confirm_character_create", "start":
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


func bot_select_character_class(class_id: String) -> void:
	if character_panel != null:
		character_panel.select_class(class_id)


func bot_select_window_size(size: String) -> void:
	_on_window_size_selected(size)


func bot_set_floating_combat_text(enabled: bool) -> void:
	_on_floating_combat_text_toggled(enabled)


func bot_select_create_game_type(session_type: String) -> void:
	_on_create_game_session_type_selected(session_type)


func bot_select_language(language: String) -> void:
	_on_language_selected(language)


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
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	var to_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag"]))


func _bot_shadow_inventory_unequip() -> void:
	if inventory_panel == null:
		return
	inventory_panel.ensure_display_visible()
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_bag_area_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "bag"]))


func _bot_shadow_inventory_drop(item_instance_id: String) -> void:
	if inventory_panel == null:
		return
	inventory_panel.ensure_display_visible()
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	if from_pos == Vector2.ZERO:
		from_pos = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_drop_outside_screen_point()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO:
		return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "drop"]))


# --- debug ------------------------------------------------------------------

func _update_debug() -> void:
	if _debug_label == null:
		return
	_sync_status_text_visibility()
	if not _debug_label.visible:
		return
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
	var boss_status := "\nboss=%s" % _last_boss_reward_status if _last_boss_reward_status != "" else ""
	_debug_label.text = "ws=%s  tick=%d  mode=%s  recon_delta=%.2f\ninv=%d  entities=%d  equipped_weapon=%s\nweapon_visual=%s%s\nW/A/S/D move  LMB action  scroll zoom  I inventory" % [
		ws_state, last_server_tick, mode, reconciliation_delta, inventory.size(), entities.size(), str(eq), weapon_vis, boss_status]


func _sync_status_text_visibility() -> void:
	if _debug_label == null:
		return
	_debug_label.visible = client_settings == null or client_settings.status_text


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
