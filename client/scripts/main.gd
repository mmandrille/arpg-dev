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
const ConsumableHealEffectScript := preload("res://scripts/consumable_heal_effect.gd")
const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")
const EliteAuraPreviewSync := preload("res://scripts/elite_aura_preview_sync.gd")
const ProjectileVisualsScript := preload("res://scripts/projectile_visuals.gd")
const MonsterHealthBarScript := preload("res://scripts/monster_health_bar.gd")
const EnemyHealthBarVisibilityScript := preload("res://scripts/enemy_health_bar_visibility.gd")
const CorpseStatusBarScript := preload("res://scripts/corpse_status_bar.gd")
const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")
const BossHealthBarScript := preload("res://scripts/boss_health_bar.gd")
const BossVisualsContextScript := preload("res://scripts/boss_visuals_context.gd")
const BossVisualsControllerScript := preload("res://scripts/boss_visuals_controller.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const StashPanelScript := preload("res://scripts/stash_panel.gd")
const BishopPanelScript := preload("res://scripts/bishop_panel.gd")
const MercenaryPanelScript := preload("res://scripts/mercenary_panel.gd")
const MercenaryPanelBridgeScript := preload("res://scripts/mercenary_panel_bridge.gd")
const BotDebugProgressionSetupScript := preload("res://scripts/bot_debug_progression_setup.gd")
const MarketPanelScript := preload("res://scripts/market_panel.gd")
const MarketBoardBadgesScript := preload("res://scripts/market_board_badges.gd")
const BlacksmithPanelScript := preload("res://scripts/blacksmith_panel.gd")
const TownServiceBridgeScript := preload("res://scripts/town_service_bridge.gd")
const ConsumableBarScript := preload("res://scripts/consumable_bar.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const SkillsPanelScript := preload("res://scripts/skills_panel.gd")
const QuestJournalPanelScript := preload("res://scripts/quest_journal_panel.gd")
const QuestEliteObjectiveStateScript := preload("res://scripts/quest_elite_objective_state.gd")
const EliteObjectiveTrackerScript := preload("res://scripts/elite_objective_tracker.gd")
const DiscoveryMinimapScript := preload("res://scripts/discovery_minimap.gd")
const CharacterBarScript := preload("res://scripts/character_bar.gd")
const SkillBarScript := preload("res://scripts/skill_bar.gd")
const CompanionBarScript := preload("res://scripts/companion_bar.gd")
const StatusEffectsBarScript := preload("res://scripts/status_effects_bar.gd")
const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const InputShadowOverlayScript := preload("res://scripts/input_shadow_overlay.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const ClientAudioControllerScript := preload("res://scripts/client_audio_controller.gd")
const ClientAudioBridgeScript := preload("res://scripts/client_audio_bridge.gd")
const PerformanceStatusFormatterScript := preload("res://scripts/performance_status_formatter.gd")
const MainMenuScript := preload("res://scripts/main_menu.gd")
const CharacterSelectPanelScript := preload("res://scripts/character_select_panel.gd")
const MultiplayerSessionsPanelScript := preload("res://scripts/multiplayer_sessions_panel.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const PauseMenuScript := preload("res://scripts/pause_menu.gd")
const SustainedClickInputScript := preload("res://scripts/sustained_click_input.gd")
const DirectionalAttackInputScript := preload("res://scripts/directional_attack_input.gd")
const CombatInputBufferScript := preload("res://scripts/combat_input_buffer.gd")
const CombatReachScript := preload("res://scripts/combat_reach.gd")
const CombatStickyTargetScript := preload("res://scripts/combat_sticky_target.gd")
const CombatLocalAttackPresentationScript := preload("res://scripts/combat_local_attack_presentation.gd")
const MovementVisualSmoothingScript := preload("res://scripts/movement_visual_smoothing.gd")
const CommandRetargetGraceScript := preload("res://scripts/command_retarget_grace.gd")
const ChannelSkillInputScript := preload("res://scripts/channel_skill_input.gd")
const ChargeChannelVisualScript := preload("res://scripts/charge_channel_visual.gd")
const MonsterVisualsLoaderScript := preload("res://scripts/monster_visuals_loader.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")
const MonsterAttackAnimationEventsScript := preload("res://scripts/monster_attack_animation_events.gd")
const DamageTypeCombatTextScript := preload("res://scripts/damage_type_combat_text.gd")
const CharacterScene := preload("res://scenes/character.tscn")
const MonsterScenesByVisual := {
	"monster_dummy": preload("res://scenes/monster_dummy.tscn"),
	"monster_dark_purple": preload("res://scenes/monster_dark_purple.tscn"),
	"monster_crocodile_archer": preload("res://scenes/monster_crocodile_archer.tscn"),
	"monster_quadruped": preload("res://scenes/monster_quadruped.tscn"),
	"monster_wolf": preload("res://scenes/monster_wolf.tscn"),
	"monster_tiny_flyer": preload("res://scenes/monster_tiny_flyer.tscn"),
	"monster_skeleton": preload("res://scenes/monster_skeleton.tscn"),
}
var client: NetClient
var resolver: EquipmentVisualResolver
var player_anim: AnimationController
var player_reaction: ModelReactionController
var entities: Dictionary = {}        # id (String) -> {node:Node3D, controller:AnimationController|null, type:String}
var player_id: String = ""
var party: Array = []
var player_hp: int = ClientConstants.PLAYER_START_HP
var player_max_hp: int = ClientConstants.PLAYER_START_HP
var player_mana: int = 10
var player_max_mana: int = 10
var player_visual_scale: float = 1.0
var _local_player_class_asset_id: String = ""
var predicted_pos := Vector3.ZERO    # client-predicted player position
var reconciliation_delta: float = 0.0
var local_leap_visual_active: bool = false
var local_charge_visual_active: bool = false
var last_server_tick: int = 0
var inventory: Array = []
var equipped: Dictionary = {}
var active_weapon_set: int = 0
var weapon_sets: Array = []
var inventory_rows: int = 3
var inventory_capacity: int = 15
var gold: int = 0
var stash_items: Array = []
var stash_gold: int = 0
var stash_capacity: int = 50
var resource_wallet: Dictionary = {}
var pending_stash_equips: Dictionary = {}
var hotbar_capacity: int = 2
var hotbar: Array = []
var character_progression: Dictionary = {}
var skill_progression: Dictionary = {}
var skill_cooldowns: Array = []
var right_click_skill_id: String = ""
var skill_function_keys: Array = ["", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""]
var pending_skill_casts: Dictionary = {}
var _channel_skill_input := ChannelSkillInputScript.new()
var _charge_channel_visual := ChargeChannelVisualScript.new()
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
var _loot_filter := preload("res://scripts/loot_label_filter.gd").new()
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
var gameplay_debug_enabled: bool = false
var visual_replay_dev_token: String = ""
var visual_replay_exit_on_complete: bool = false
var visual_replay_exit_requested: bool = false
var waypoint_panel: PanelContainer
var waypoint_rows: VBoxContainer
var visual_replay_exit_timer: float = 0.0
var visual_replay_show_inventory: bool = false
var visual_replay_completion_hold_s: float = 0.5
var _perf_debug_sampler := preload("res://scripts/perf_debug_sampler.gd").new()
var client_settings: ClientSettings
var menu_layer: CanvasLayer
var main_menu: MainMenu
var character_panel: CharacterSelectPanel
var multiplayer_panel: MultiplayerSessionsPanel
var settings_panel: SettingsPanel
var audio_controller: ClientAudioController
var pause_menu: PauseMenu
var loss_popup: Control
var gameplay_active: bool = false
var settings_return_target: String = "main"
var character_flow: String = ClientConstants.CHARACTER_FLOW_CREATE_GAME
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
var gameplay_ui_layer: CanvasLayer
var health_bars_layer: CanvasLayer
var monster_health_bars: Dictionary = {} # id (String) -> MonsterHealthBar
var revive_corpse_status_bars: Dictionary = {} # id (String) -> CorpseStatusBar
var boss_health_bar: BossHealthBar
var walls_root: Node3D
var inventory_panel: InventoryPanel
var shop_panel: ShopPanel
var stash_panel: StashPanel
var bishop_panel: BishopPanel
var mercenary_panel: MercenaryPanel
var market_panel
var blacksmith_panel: BlacksmithPanel
var consumable_bar: ConsumableBar
var character_stats_panel: CharacterStatsPanel
var skills_panel: SkillsPanel
var quest_journal_panel: QuestJournalPanel
var elite_objective_tracker: EliteObjectiveTracker
var discovery_minimap: DiscoveryMinimap
var character_bar: Control
var skill_bar: SkillBar
var companion_bar: Control
var status_effects_bar: StatusEffectsBar
var character_info_panel: PanelContainer
var character_info_name_label: Label
var character_info_level_label: Label
var character_info_area_label: Label
var input_shadow: InputShadowOverlay
var fog_overlay: FogOfWarOverlay
var _health_bar: PlayerHealthBar
var _send_cooldown: float = 0.0
var _attack_cooldown: float = 0.0
var _sustained_click: SustainedClickInput = SustainedClickInputScript.new()
var _attack_buffer: CombatInputBuffer = CombatInputBufferScript.new()
var _sticky_attack: CombatStickyTarget = CombatStickyTargetScript.new()
var _local_attack_presentation: CombatLocalAttackPresentation = CombatLocalAttackPresentationScript.new()
var _movement_visual_smoothing: MovementVisualSmoothing = MovementVisualSmoothingScript.new()
var _command_retarget_grace: CommandRetargetGrace = CommandRetargetGraceScript.new()
var _movement_requires_fresh_input: bool = false
var _player_walk_linger: float = 0.0
var _last_facing_direction := Vector2(1.0, 0.0)
var _debug_label: Label
var _level_label: Label
var last_performance_status: Dictionary = {}
var _last_ping_ms: int = -1
var _camera: Camera3D
var ground_node: MeshInstance3D
var _ground_factory: GroundWallFactory = GroundWallFactory.new()
var _wall_renderer: WallRenderer
var _loot_factory: LootNodeFactory = LootNodeFactory.new()
var _boss_visuals_context: BossVisualsContext
var _boss_visuals: BossVisualsController

func _ready() -> void:
	player_anchor = $World/PlayerAnchor
	character_visual = $World/PlayerAnchor/CharacterVisual
	_movement_visual_smoothing.reset(player_anchor, character_visual)
	entities_root = $Entities
	# Mount-root is injected (spec §4.8): the resolver finds the named socket
	# within CharacterVisual, never via an absolute scene path.
	resolver = EquipmentResolverScript.new(character_visual)
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)
	_apply_model_tint(character_visual, ClientConstants.PLAYER_TINT)
	player_reaction = ModelReactionControllerScript.new(character_visual, ClientConstants.PLAYER_TINT)
	gameplay_debug_enabled = _truthy_env("ARPG_GAMEPLAY_DEBUG")
	_build_scene()
	client_settings = ClientSettingsScript.new()
	client_settings.load()
	client_settings.set_language(client_settings.language, false)
	_loot_filter.set_mode_label(client_settings.loot_filter_mode)
	client_settings.apply()
	ClientAudioBridgeScript.apply_settings(audio_controller, client_settings)
	if discovery_minimap != null: discovery_minimap.set_panel_opacity(client_settings.map_opacity)
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
	var requested_character_id := BotDebugProgressionSetupScript.prepare_character(client, _env("ARPG_DEBUG_TOKEN", "local-debug-token"), _env("ARPG_BOT_DEBUG_PROGRESSION", "") if bot_client_run else "", _env("ARPG_BOT_DEBUG_GOLD", "") if bot_client_run else "")
	if requested_world_id == "" and not bot_client_run:
		requested_world_id = "dungeon_levels"
	if bot_client_run or resume_session_id != "" or _truthy_env("ARPG_AUTOSTART"):
		if not _start_automation_session(resume_session_id, requested_world_id, requested_seed, bot_client_run, requested_character_id):
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
func _start_automation_session(resume_session_id: String, requested_world_id: String, requested_seed: String, bot_client_run: bool, requested_character_id: String = "") -> bool:
	if not client.create_session(resume_session_id, requested_world_id, requested_character_id, requested_seed):
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
	for panel in [inventory_panel, shop_panel, stash_panel, bishop_panel, market_panel, blacksmith_panel, character_stats_panel, skills_panel, quest_journal_panel, character_info_panel]:
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
	character_flow = ClientConstants.CHARACTER_FLOW_CREATE_GAME
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
	character_flow = ClientConstants.CHARACTER_FLOW_JOIN_GAME
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
	character_flow = ClientConstants.CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST
	pending_join_session_id = ""
	_show_character_picker_for_flow("Choose Character")

func _on_join_listed_session_requested(session_id: String) -> void:
	if session_id == "":
		return
	character_flow = ClientConstants.CHARACTER_FLOW_JOIN_GAME
	pending_join_session_id = session_id
	_show_character_picker_for_flow("Choose Character")

func _on_character_panel_back() -> void:
	if character_panel != null:
		character_panel.hide_panel()
	match character_flow:
		ClientConstants.CHARACTER_FLOW_JOIN_GAME, ClientConstants.CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN, ClientConstants.CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST:
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
		ClientConstants.CHARACTER_FLOW_CREATE_GAME:
			_start_create_game_session(character_id)
		ClientConstants.CHARACTER_FLOW_JOIN_GAME:
			_start_listed_join_session(character_id)
		ClientConstants.CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST:
			_start_listed_host_session(character_id)
		ClientConstants.CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN:
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
	character_flow = ClientConstants.CHARACTER_FLOW_CREATE_GAME
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
	character_flow = ClientConstants.CHARACTER_FLOW_CREATE_GAME
	pending_join_session_id = ""
	_begin_gameplay_connection(false)

func _on_settings_from_main() -> void:
	settings_return_target = "main"
	main_menu.visible = false
	ClientAudioBridgeScript.show_settings(settings_panel, client_settings)
func _on_settings_from_pause() -> void:
	settings_return_target = "pause"
	if pause_menu != null:
		pause_menu.hide_pause()
	ClientAudioBridgeScript.show_settings(settings_panel, client_settings)
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

func _on_monster_health_bar_mode_selected(mode: String) -> void:
	if client_settings != null:
		client_settings.set_monster_health_bar_mode(mode)
		_sync_settings_panel()
		_refresh_monster_health_bar_visibility()

func _on_master_volume_changed(value: float) -> void:
	ClientAudioBridgeScript.set_master_volume(audio_controller, client_settings, value)

func _on_music_volume_changed(value: float) -> void:
	ClientAudioBridgeScript.set_music_volume(audio_controller, client_settings, value)

func _on_sfx_volume_changed(value: float) -> void:
	ClientAudioBridgeScript.set_sfx_volume(audio_controller, client_settings, value)

func _on_map_opacity_changed(value: float) -> void:
	if client_settings == null: return
	client_settings.set_map_opacity(value)
	if discovery_minimap != null: discovery_minimap.set_panel_opacity(client_settings.map_opacity)
	_sync_settings_panel()

func _on_companion_stance_requested(stance: String) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	client.send("companion_command_intent", last_server_tick, {"stance": stance})

func _on_companion_bar_selected(_companion: Dictionary) -> void:
	if mercenary_panel == null:
		return
	_close_gameplay_panels("mercenary")
	_sync_companion_bar()
	mercenary_panel.show_roster(companion_bar.get_debug_state().get("companions", []) if companion_bar != null else [])
	_raise_gameplay_windows()

func _sync_settings_panel() -> void:
	if settings_panel != null and client_settings != null:
		settings_panel.set_selected_size_label(ClientSettingsScript.size_label(client_settings.window_size))
		settings_panel.set_floating_combat_text_enabled(client_settings.floating_combat_text)
		settings_panel.set_status_text_enabled(client_settings.status_text)
		settings_panel.set_create_game_session_type(client_settings.create_game_session_type)
		settings_panel.set_language(client_settings.language)
		settings_panel.set_monster_health_bar_mode(client_settings.monster_health_bar_mode)
		ClientAudioBridgeScript.sync_settings_panel(settings_panel, client_settings)

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
	player_hp = ClientConstants.PLAYER_START_HP
	player_max_hp = ClientConstants.PLAYER_START_HP
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
	skill_function_keys = ["", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""]
	if status_effects_bar != null:
		status_effects_bar.clear_effects()
	loot_ids.clear()
	monster_ids.clear()
	ClientAudioBridgeScript.stop_boss_music(audio_controller)
	interactable_ids.clear()
	current_wall_layout = []
	discovered_teleporters.clear()
	pending_interactable_action.clear()
	_clear_pending_attack_commands()
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
	if _boss_visuals != null: _boss_visuals.hide_boss_health_bar()
	if walls_root != null: _clear_wall_nodes()
	if discovery_minimap != null: discovery_minimap.reset_session()
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
		_movement_visual_smoothing.reset(player_anchor, character_visual)
	if clear_session and client != null:
		client.session_id = ""
		client.character_id = ""
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
		if _boss_visuals != null:
			_boss_visuals_context.last_server_tick = last_server_tick
			_boss_visuals.advance_boss_phase_display(delta)
		_tick_skill_cooldowns(delta)
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
	_movement_visual_smoothing.tick(delta, character_visual)
	if not _user_input_blocked():
		_update_facing_toward_mouse()
	_refresh_monster_health_bar_visibility()
	_update_debug()
	_perf_debug_sampler.sample(delta, state, last_server_tick, reconciliation_delta, entities, monster_ids)

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
			var accepted_message_id := str(payload.get("accepted_message_id", ""))
			_record_ping(accepted_message_id)
			pending_action_targets.erase(accepted_message_id)
		"intent_rejected":
			_record_ping(str(payload.get("rejected_message_id", "")))
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
	elif blacksmith_panel != null and blacksmith_panel.visible:
		blacksmith_panel.show_status(reason.replace("_", " "), true)

func _record_ping(message_id: String) -> void:
	if client == null:
		return
	var latency_ms := client.consume_latency_ms(message_id)
	if latency_ms >= 0:
		_last_ping_ms = latency_ms

func _apply_snapshot(p: Dictionary) -> void:
	current_level = int(p.get("current_level", 0))
	if discovery_minimap != null: discovery_minimap.sync_session(str(p.get("session_id", client.session_id if client != null else "")))
	ClientAudioBridgeScript.ambience_for_level(audio_controller, current_level)
	_update_ground_material()
	var local_id := _snapshot_local_player_id(p)
	if local_id != "": player_id = local_id
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
	if _boss_visuals != null:
		_boss_visuals_context.last_server_tick = last_server_tick
		_boss_visuals.sync_boss_health_bar()
	inventory = p.get("inventory", [])
	equipped = p.get("equipped", {})
	active_weapon_set = int(p.get("active_weapon_set", active_weapon_set))
	weapon_sets = p.get("weapon_sets", weapon_sets)
	inventory_rows = int(p.get("inventory_rows", inventory_rows))
	inventory_capacity = int(p.get("inventory_capacity", inventory_capacity))
	gold = int(p.get("gold", gold))
	stash_items = p.get("stash_items", [])
	stash_gold = int(p.get("stash_gold", stash_gold))
	stash_capacity = int(p.get("stash_capacity", stash_capacity))
	_apply_resource_wallet_snapshot(p.get("resource_wallet", []))
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
	_sync_quest_journal()
	_sync_elite_objective_tracker()
	_sync_discovery_minimap()
	_refresh_market_board_summary()
	_reconcile_player()
	if bot_mode and not _bot_logged_snapshot:
		_bot_logged_snapshot = true
		print("[bot-client] snapshot applied entities=%d monsters=%d loot=%d hp=%d" % [
			entities.size(), monster_ids.size(), loot_ids.size(), player_hp
		])

func _apply_delta(p: Dictionary) -> void:
	var perf_payload = p.get("performance", {})
	if perf_payload is Dictionary:
		last_performance_status = (perf_payload as Dictionary).duplicate(true)
	for ev in p.get("events", []):
		if str(ev.get("event_type", "")) == "level_changed":
			current_level = int(ev.get("to_level", current_level))
			ClientAudioBridgeScript.ambience_for_level(audio_controller, current_level)
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
			_hide_blacksmith_panel()
			if quest_journal_panel != null:
				quest_journal_panel.hide_display()
	var local_motion_skill_id := _local_player_motion_skill_id(p.get("events", []))
	var local_motion_landing := Vector3.ZERO
	var local_motion_has_landing := false
	var changes: Array = p.get("changes", [])
	for c in changes:
		match c.get("op", ""):
			"wall_layout_update":
				_render_wall_layout(c.get("walls", []))
			"entity_spawn", "entity_update":
				var entity: Dictionary = c.get("entity", {})
				if local_motion_skill_id != "" and _is_local_player_entity_update(entity):
					local_motion_landing = _entity_position(entity)
					local_motion_has_landing = true
					_upsert_entity(entity, false)
				else:
					_upsert_entity(entity)
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
			"weapon_set_update":
				active_weapon_set = int(c.get("active_weapon_set", active_weapon_set))
				weapon_sets = c.get("weapon_sets", weapon_sets)
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
			"resource_wallet_update":
				var resource_id := str(c.get("resource_id", ""))
				if resource_id != "":
					resource_wallet[resource_id] = max(0, int(c.get("amount", resource_wallet.get(resource_id, 0))))
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
	_sync_quest_journal()
	_sync_elite_objective_tracker()
	_sync_discovery_minimap()
	var heal_cast_rain_correlations := _heal_cast_rain_correlations(p.get("events", []))
	for ev in p.get("events", []):
		var eid := _event_subject_entity_id(ev)
		var event_type := str(ev.get("event_type", ""))
		if eid == player_id and str(ev.get("skill_id", "")) == "charge":
			if event_type == "skill_channel_started":
				_start_charge_channel_visual(ev)
				continue
			if event_type == "skill_channel_updated":
				_update_charge_channel_visual(ev)
				continue
			if event_type == "skill_channel_ended":
				_stop_charge_channel_visual()
				continue
		if event_type == "skill_cooldown_started" and eid == player_id:
			if skill_bar != null:
				skill_bar.start_skill_cooldown(
					str(ev.get("skill_id", "")),
					int(ev.get("remaining_ticks", 0)),
					int(ev.get("total_ticks", 0))
				)
		if event_type == "skill_cast" and eid == player_id:
			ClientAudioBridgeScript.skill(audio_controller, str(ev.get("skill_id", "")))
			if skill_bar != null:
				skill_bar.flash_cast()
			if player_anim != null:
				player_anim.play_one_shot("attack")
			if str(ev.get("skill_id", "")) == "earthbreaker":
				EarthbreakerJump.play(character_visual, self)
			if str(ev.get("skill_id", "")) == "leap":
				_play_leap_visual(ev, local_motion_landing if local_motion_has_landing else Vector3.INF)
			if str(ev.get("skill_id", "")) == "charge":
				_play_charge_visual(ev, local_motion_landing if local_motion_has_landing else Vector3.INF)
			if ev.has("projectile_def_id") and ev.has("position") and ev.has("direction"):
				_spawn_skill_projectile_visual(ev)
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
			ClientAudioBridgeScript.heal(audio_controller)
			_show_damage_number(eid, Color(0.3, 1.0, 0.45), ev.get("heal", null), "+", 1.0)
			if str(ev.get("skill_id", "")) == "heal":
				_spawn_heal_rain(eid)
			else:
				_spawn_consumable_heal_effect(eid)
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
		if str(ev.get("skill_id", "")) == "pinning_shot":
			if event_type == "skill_effect_started":
				_set_entity_pinning_root(eid, true)
				continue
			if event_type == "skill_effect_ended":
				_set_entity_pinning_root(eid, false)
				continue
		if PlayerStatusEffectMarkers.is_stun_skill_id(str(ev.get("skill_id", ""))):
			if event_type == "skill_effect_started":
				_set_entity_stun(eid, true)
				continue
			if event_type == "skill_effect_ended":
				_set_entity_stun(eid, false)
				continue
		if event_type == "skill_effect_started" and eid == player_id:
			if status_effects_bar != null:
				status_effects_bar.start_effect(ev)
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.RAGE_EFFECT_ID:
				_apply_local_player_visual_scale(1.0 + float(ev.get("amount", 0)) / 100.0)
				PlayerStatusEffectMarkers.sync_rage_effect(player_anchor, true)
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID:
				PlayerStatusEffectMarkers.sync_sanctuary_effect(player_anchor, [PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID], _sanctuary_radius())
			continue
		if event_type == "skill_effect_ended" and eid == player_id:
			if status_effects_bar != null:
				status_effects_bar.end_effect(str(ev.get("skill_id", "")))
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.RAGE_EFFECT_ID:
				_apply_local_player_visual_scale(1.0)
				PlayerStatusEffectMarkers.sync_rage_effect(player_anchor, false)
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID:
				PlayerStatusEffectMarkers.sync_holy_shield_effect(player_anchor, [])
			if str(ev.get("skill_id", "")) == PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID:
				PlayerStatusEffectMarkers.sync_sanctuary_effect(player_anchor, [])
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
				ClientAudioBridgeScript.damage(audio_controller, eid == player_id)
				MonsterAttackAnimationEventsScript.play_source_attack_for_event(ev, entities)
				_show_combat_text_for_event(eid, ev, Color(1.0, 0.32, 0.2))
				if str(ev.get("outcome", "")) != "immune":
					_play_entity_reaction(eid, ev, "hit")
				if _health_bar != null:
					_health_bar.update_hp(player_hp, player_max_hp)
			if event_type == "player_killed":
				MonsterAttackAnimationEventsScript.play_source_attack_for_event(ev, entities)
				_play_entity_reaction(eid, ev, "death")
				_show_loss_popup()
			if event_type == "attack_missed":
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			var player_clip = ClientConstants.PLAYER_EVENT_CLIPS.get(event_type, null)
			if player_clip == null or player_anim == null:
				continue
			if player_clip == "death":
				player_anim.enter_terminal("death")
			else:
				_play_local_player_reaction_animation(player_clip)
			continue
		if ClientConstants.PLAYER_EVENT_CLIPS.has(event_type) and entities.has(eid):
			if event_type == "player_damaged":
				ClientAudioBridgeScript.damage(audio_controller, eid == player_id)
				MonsterAttackAnimationEventsScript.play_source_attack_for_event(ev, entities)
				_show_combat_text_for_event(eid, ev, Color(1.0, 0.32, 0.2))
				if str(ev.get("outcome", "")) != "immune":
					_play_entity_reaction(eid, ev, "hit")
			if event_type == "player_killed":
				MonsterAttackAnimationEventsScript.play_source_attack_for_event(ev, entities)
				var remote_dead: Dictionary = entities[eid]
				remote_dead["hp"] = 0
				_play_entity_reaction(eid, ev, "death")
			var remote_player_clip = ClientConstants.PLAYER_EVENT_CLIPS.get(event_type, null)
			var remote_ctrl = entities[eid].get("controller", null)
			if remote_ctrl != null:
				if remote_player_clip == "death":
					remote_ctrl.enter_terminal("death")
				else:
					remote_ctrl.play_one_shot(remote_player_clip)
			continue
		if (event_type == "interactable_activated" or event_type == "interactable_state_changed") and entities.has(eid):
			_set_interactable_state(eid, entities[eid], "open" if event_type == "interactable_activated" else str(ev.get("state", "")))
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
		if event_type == "corpse_opened":
			_show_corpse_panel(ev)
			continue
		if event_type == "corpse_item_recovered":
			_show_corpse_panel(ev)
			if stash_panel != null and stash_panel.visible:
				stash_panel.show_status("Item recovered")
			continue
		if event_type == "unique_chest_opened":
			_show_unique_chest_panel(ev)
			continue
		if event_type == "unique_chest_item_taken":
			_show_unique_chest_panel(ev)
			if stash_panel != null and stash_panel.visible:
				stash_panel.show_status("Item taken")
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
		if MercenaryPanelBridgeScript.try_handle_event(self, mercenary_panel, ev, gold):
			continue
		if event_type == "market_service_opened":
			_show_market_panel(ev)
			continue
		if event_type == "blacksmith_service_opened":
			_show_blacksmith_panel(ev)
			continue
		if event_type == "bishop_respec" and bishop_panel != null and bishop_panel.visible:
			bishop_panel.set_gold(gold)
			bishop_panel.show_status("Respec complete")
			continue
		if event_type == "bishop_revive_all" and bishop_panel != null and bishop_panel.visible:
			bishop_panel.show_status("Account heroes revived")
			continue
		if event_type in ["bishop_debug_level_gained", "bishop_debug_skill_point_gained", "bishop_debug_stat_point_gained"] and bishop_panel != null and bishop_panel.visible:
			if event_type == "bishop_debug_level_gained":
				bishop_panel.show_status("Level gained")
			elif event_type == "bishop_debug_skill_point_gained":
				bishop_panel.show_status("Skill point gained")
			else:
				bishop_panel.show_status("Stat point gained")
			continue
		if event_type == "boss_killed":
			ClientAudioBridgeScript.kill(audio_controller, true)
			ClientAudioBridgeScript.stop_boss_music(audio_controller)
			_last_boss_reward_status = _boss_visuals.show_boss_reward_status(str(ev.get("boss_template_id", ""))) if _boss_visuals != null else "Boss defeated"
			continue
		if event_type == "boss_phase_started" and entities.has(eid):
			ClientAudioBridgeScript.boss_phase(audio_controller, ev)
			if _boss_visuals != null:
				_boss_visuals_context.last_server_tick = last_server_tick
				_boss_visuals.apply_boss_phase_started(eid, ev)
			continue
		if event_type == "boss_phase_ended" and entities.has(eid):
			if _boss_visuals != null:
				_boss_visuals.apply_boss_phase_ended(eid, ev)
			continue
		if event_type == "monster_aggro":
			_show_damage_number(eid, Color("#80ff8f"), null, "", 0.0, "threat", "AGGRO")
			continue
		var clip = ClientConstants.MONSTER_EVENT_CLIPS.get(event_type, null)
		if clip == null:
			if event_type in ["attack_missed", "attack_blocked"]:
				_face_event_source_toward_target(ev)
				CombatLocalAttackPresentationScript.present_result(_local_attack_presentation, ev, player_id, audio_controller, player_anim, CombatReachScript.local_player_attack_mode(inventory, equipped))
				_show_combat_text_for_event(eid, ev, Color(0.82, 0.86, 0.92))
			continue
		if event_type == "monster_damaged" or event_type == "monster_killed":
			_face_event_source_toward_target(ev)
			CombatLocalAttackPresentationScript.present_result(_local_attack_presentation, ev, player_id, audio_controller, player_anim, CombatReachScript.local_player_attack_mode(inventory, equipped))
			_show_combat_text_for_event(eid, ev, Color(1.0, 0.92, 0.25))
		if event_type == "monster_damaged":
			ClientAudioBridgeScript.damage(audio_controller, false)
			_play_entity_reaction(eid, ev, "hit")
		if event_type == "monster_killed":
			ClientAudioBridgeScript.kill(audio_controller, false)
			_remove_monster_health_bar(eid)
			if entities.has(eid):
				var dead_rec: Dictionary = entities[eid]
				dead_rec["hp"] = 0
				_set_pickable(dead_rec["node"] as Node3D, false)
				_clear_terminal_entity_status_markers(dead_rec)
				_clear_elite_command_for_pack_if_leader_died(dead_rec)
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
	if _boss_visuals != null:
		_boss_visuals_context.last_server_tick = last_server_tick
		_boss_visuals.sync_boss_health_bar()
	_reconcile_player()

func _local_player_motion_skill_id(events: Array) -> String:
	for raw in events:
		if not (raw is Dictionary):
			continue
		var ev := raw as Dictionary
		if str(ev.get("event_type", "")) == "skill_cast" \
				and _event_subject_entity_id(ev) == player_id:
			var skill_id := str(ev.get("skill_id", ""))
			if skill_id == "leap" or skill_id == "charge":
				return skill_id
	return ""

func _is_local_player_entity_update(e: Dictionary) -> bool:
	return str(e.get("type", "")) == "player" and (str(e.get("id", "")) == player_id or player_id == "")

func _entity_position(e: Dictionary) -> Vector3:
	var pos: Dictionary = e.get("position", {})
	return Vector3(float(pos.get("x", 0.0)), 0.0, float(pos.get("y", 0.0)))

func _upsert_entity(e: Dictionary, apply_local_player_position: bool = true) -> void:
	var id := str(e["id"])
	var server_pos := _entity_position(e)
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
			_refresh_skill_ui()
		if e.has("visual_scale"):
			_apply_local_player_visual_scale(float(e["visual_scale"]))
		if e.has("effect_ids"):
			PlayerStatusEffectMarkers.sync_holy_shield_effect(player_anchor, e.get("effect_ids", []))
			PlayerStatusEffectMarkers.sync_sanctuary_effect(player_anchor, e.get("effect_ids", []), _sanctuary_radius())
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		var prev_predicted_pos := predicted_pos
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		if apply_local_player_position and not local_leap_visual_active and not local_charge_visual_active:
			player_anchor.position = server_pos
			_movement_visual_smoothing.preserve_after_anchor_move(player_anchor, character_visual)
		if apply_local_player_position and prev_predicted_pos.distance_to(server_pos) > 0.001 and player_hp > 0:
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
		if _entity_type_uses_combat_presentation(str(e["type"])):
			var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
			if ap != null:
				controller = AnimationControllerScript.new(ap)
			else:
				push_warning("[main] %s %s has no AnimationPlayer" % [str(e["type"]), id])
		var base_tint := _entity_base_tint(e)
		var reaction = null
		if _entity_type_uses_combat_presentation(str(e["type"])):
			reaction = ModelReactionControllerScript.new(node, base_tint)
		rec = {"node": node, "controller": controller, "reaction": reaction, "type": str(e["type"]), "base_tint": base_tint.to_html(false)}
		if e.has("item_def_id"):
			rec["item_def_id"] = str(e["item_def_id"])
		if e.has("amount"):
			rec["amount"] = int(e["amount"])
		if e.has("monster_def_id"):
			rec["monster_def_id"] = str(e["monster_def_id"])
	for key in ["item_template_id", "display_name", "rarity", "rolled_stats", "requirements", "requirement_status", "requirements_met", "equip_preview", "effect_ids", "character_id", "boss_template_id", "visual_model", "visual_tint", "boss_phase", "elite_objective", "quest_reward", "owner_id", "target_id", "combat_stats", "remaining_ticks", "total_ticks", "companion_stance"]:
		if e.has(key):
			rec[key] = e[key]
	for key in ["corpse_character_id", "corpse_name", "corpse_level", "corpse_item_count"]:
		if e.has(key):
			rec[key] = e[key]
	if e.has("is_boss"):
		rec["is_boss"] = bool(e["is_boss"])
	if e.has("visual_scale"):
		rec["visual_scale"] = float(e["visual_scale"])
	if e.has("interactable_def_id"):
		rec["interactable_def_id"] = str(e["interactable_def_id"])
	if is_new:
		entities[id] = rec
		var new_node := rec["node"] as Node3D
		if e["type"] != "projectile" and e["type"] != "player":
			_attach_pick_collider(new_node, id, str(e["type"]), str(rec.get("interactable_def_id", "")))
		if str(rec.get("interactable_def_id", "")) == "hero_corpse":
			_set_loot_label_visible(id, loot_label_reveal_held or id == hovered_loot_id, id == hovered_loot_id)
	if e["type"] == "loot" and not loot_ids.has(id):
		loot_ids.append(id)
		_refresh_loot_label_visibility()
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
		if _entity_type_uses_combat_presentation(str(rec["type"])) and rec["controller"] != null and not is_new:
			var hp_val := int(e.get("hp", rec.get("hp", 1)))
			var moved := prev_pos.distance_to(server_pos) > 0.001
			if moved and hp_val > 0:
				rec["walk_linger"] = ClientConstants.WALK_ANIMATION_LINGER_SECONDS
				_face_entity_direction(node, Vector2(server_pos.x - prev_pos.x, server_pos.z - prev_pos.z))
			elif hp_val > 0 and str(rec.get("target_id", "")) != "":
				_face_node_toward_entity(node, str(rec.get("target_id", "")))
			rec["controller"].set_locomotion(float(rec.get("walk_linger", 0.0)) > 0.0 and hp_val > 0)
		if rec["type"] == "player":
			rec["hp"] = int(e.get("hp", rec.get("hp", ClientConstants.PLAYER_START_HP)))
			rec["max_hp"] = int(e.get("max_hp", rec.get("max_hp", ClientConstants.PLAYER_START_HP)))
			if int(rec["hp"]) <= 0:
				_enter_entity_terminal_death(id, rec)
	if rec["type"] == "interactable":
		var state := str(e.get("state", rec.get("state", "closed")))
		_set_interactable_state(id, rec, state)
	if _entity_type_uses_combat_presentation(str(rec["type"])):
		_apply_entity_visual_metadata(rec, e)
	# Resume/snapshot consistency: a monster already dead in the snapshot enters
	# the terminal death pose without waiting for an event (spec §5.4).
	if rec["type"] == "monster" or rec["type"] == "companion":
		var hp = e.get("hp", null)
		var max_hp = e.get("max_hp", null)
		if hp != null and max_hp != null:
			rec["hp"] = int(hp)
			rec["max_hp"] = int(max_hp)
			if rec["type"] == "monster":
				_upsert_monster_health_bar(id, rec["node"] as Node3D, int(hp), int(max_hp))
		if rec["type"] == "monster" and hp != null and int(hp) <= 0:
			_set_pickable(rec["node"] as Node3D, false)
			_clear_terminal_entity_status_markers(rec)
			_ensure_dead_monster_revive_label(id, rec)
			_enter_entity_terminal_death(id, rec)
	_sync_companion_bar()

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
	_remove_revive_corpse_status_bar(id)
	loot_ids.erase(id)
	monster_ids.erase(id)
	interactable_ids.erase(id)
	if _boss_visuals != null:
		_boss_visuals_context.last_server_tick = last_server_tick
		_boss_visuals.sync_boss_health_bar()
	_sync_companion_bar()

func _entity_type_uses_combat_presentation(entity_type: String) -> bool:
	return entity_type == "monster" or entity_type == "player" or entity_type == "companion"

func _clear_terminal_entity_status_markers(rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	PlayerStatusEffectMarkers.sync_holy_shield_effect(node, [])
	PlayerStatusEffectMarkers.sync_burning_effect(node, false)
	PlayerStatusEffectMarkers.sync_elite_command_effect(node, false)
	PlayerStatusEffectMarkers.sync_pinning_root_effect(node, false)
	PlayerStatusEffectMarkers.sync_stun_effect(node, false)
	PlayerStatusEffectMarkers.sync_elite_command_radius_preview(node, false, 0.0)

func _clear_elite_command_for_pack_if_leader_died(dead_rec: Dictionary) -> void:
	if not bool(dead_rec.get("monster_pack_leader", false)):
		return
	var pack_id := str(dead_rec.get("monster_pack_id", ""))
	if pack_id == "":
		return
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("monster_pack_id", "")) != pack_id:
			continue
		var node := rec.get("node", null) as Node3D
		if node != null:
			PlayerStatusEffectMarkers.sync_elite_command_effect(node, false)
			PlayerStatusEffectMarkers.sync_elite_command_radius_preview(node, false, 0.0)
	EliteAuraPreviewSync.sync(entities, dungeon_generation)

func _clear_level_entities() -> void:
	for id in entities.keys():
		(entities[id]["node"] as Node3D).queue_free()
	entities.clear()
	for id in monster_health_bars.keys():
		var bar = monster_health_bars[id]
		if is_instance_valid(bar):
			bar.queue_free()
	monster_health_bars.clear()
	if _boss_visuals != null:
		_boss_visuals.hide_boss_health_bar()
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

func _remove_inventory_items_by_def(item_def_id: String, count: int) -> void:
	if item_def_id == "" or count <= 0:
		return
	var removed := 0
	for i in range(inventory.size() - 1, -1, -1):
		if str(inventory[i].get("item_def_id", "")) == item_def_id:
			inventory.remove_at(i)
			removed += 1
			if removed >= count:
				return

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

func _apply_resource_wallet_snapshot(rows: Variant) -> void:
	resource_wallet.clear()
	if typeof(rows) != TYPE_ARRAY:
		return
	for value in rows:
		if typeof(value) != TYPE_DICTIONARY:
			continue
		var row := value as Dictionary
		var resource_id := str(row.get("resource_id", ""))
		if resource_id != "":
			resource_wallet[resource_id] = max(0, int(row.get("amount", 0)))
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
		inventory_panel.set_inventory_state(inventory, equipped, inventory_rows, inventory_capacity, gold, hotbar, hotbar_capacity, active_weapon_set, weapon_sets)
	if shop_panel != null and shop_panel.visible:
		shop_panel.set_inventory_state(inventory, equipped, gold)
	if stash_panel != null:
		stash_panel.set_stash_state(stash_items, stash_gold, stash_capacity)
		stash_panel.set_inventory_state(inventory, equipped, gold, hotbar)
	if bishop_panel != null and bishop_panel.visible:
		bishop_panel.set_gold(gold)
	if mercenary_panel != null and mercenary_panel.visible: mercenary_panel.set_gold(gold)
	if blacksmith_panel != null and blacksmith_panel.visible:
		blacksmith_panel.show_blacksmith(
			blacksmith_panel.blacksmith_entity_id,
			inventory,
			gold,
			stash_gold,
			_blacksmith_config(),
			blacksmith_panel.get_debug_state().get("status", ""),
			resource_wallet
		)
	if bishop_panel != null and bishop_panel.visible: bishop_panel.set_resource_wallet(resource_wallet)
	if character_bar != null: character_bar.set_resource_wallet(resource_wallet)
	if consumable_bar != null:
		consumable_bar.set_inventory_state(inventory)
		consumable_bar.set_hotbar_state(hotbar_capacity, hotbar)
func _refresh_inventory_panel() -> void:
	_refresh_inventory_ui()
	if visual_replay_enabled:
		_sync_inventory_replay_display()

func _reconcile_player() -> void:
	if player_anchor != null:
		if local_leap_visual_active:
			_sync_camera_to_player()
			return
		player_anchor.position = predicted_pos
		_movement_visual_smoothing.preserve_after_anchor_move(player_anchor, character_visual)
		_sync_camera_to_player()

func _sync_camera_to_player() -> void:
	if _camera == null or player_anchor == null:
		return
	var target := player_anchor.global_position
	_camera.global_position = target + ClientConstants.CAMERA_FOLLOW_OFFSET
	_camera.look_at(target, Vector3.UP)

func _show_combat_text_for_event(entity_id: String, ev: Dictionary, default_color: Color) -> void:
	var outcome := str(ev.get("outcome", ""))
	var damage = ev.get("damage", null)
	var special := DamageTypeCombatTextScript.special_outcome(outcome)
	if not special.is_empty():
		_show_damage_number(entity_id, special.get("color", Color.WHITE), null, "", 0.0, str(special.get("variant", outcome)), str(special.get("text", "")))
		return
	var presentation := DamageTypeCombatTextScript.number_for_event(ev, default_color)
	if not presentation.is_empty():
		_show_damage_number(entity_id, presentation.get("color", default_color), presentation.get("amount", damage), "", 0.0, str(presentation.get("variant", "normal")), str(presentation.get("text", "")), str(presentation.get("damage_type", "")))
		return
	_show_damage_number(entity_id, default_color, damage)

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
	return _skill_effect_radius(PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID, 5.0)

func _sanctuary_radius() -> float:
	return _skill_effect_radius(PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID, 5.0)

func _skill_effect_radius(skill_id: String, fallback: float) -> float:
	var def := SkillRulesLoader.skill_definition(skill_id)
	for effect in def.get("effects", []):
		var row: Dictionary = effect
		if str(row.get("effect_id", skill_id)) == skill_id and row.has("radius"):
			return maxf(float(row.get("radius", fallback)), 0.5)
	return fallback

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

func _show_damage_number(entity_id: String, color: Color, event_damage = null, prefix: String = "", side_override: float = 0.0, variant: String = "normal", text_override: String = "", damage_type: String = "") -> void:
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
	pop.setup(_camera, target, world_position, amount, color, side, prefix, variant, text_override, damage_type)

func _spawn_heal_rain(entity_id: String) -> void:
	var target := _node_for_entity_id(entity_id)
	if target == null:
		return
	_spawn_heal_rain_at_position(_node_world_or_local_position(target))

func _spawn_consumable_heal_effect(entity_id: String) -> void:
	var target := _node_for_entity_id(entity_id)
	if target == null:
		return
	var effect := ConsumableHealEffectScript.new() as Node3D
	effect.position = _node_world_or_local_position(target) + Vector3(0.0, 0.45, 0.0)
	add_child(effect)

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
	effect.setup(_skill_presentation_float("heal", "visual_radius"))
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
	wedge.name = "EarthbreakerRing" if str(ev.get("skill_id", "")) == "earthbreaker" else "CleaveCone"
	wedge.mesh = mesh
	wedge.material_override = mat
	wedge.position = Vector3(pos.x, 0.08, pos.y)
	add_child(wedge)
	var tween := wedge.create_tween()
	tween.tween_property(wedge, "scale", Vector3.ONE * 1.03, 0.16)
	tween.tween_callback(wedge.queue_free)

func _play_leap_visual(ev: Dictionary, landing_override: Vector3 = Vector3.INF) -> void:
	if player_anchor == null or not ev.has("position"):
		return
	var start_2d := _vec2_from_dict(ev.get("position", {}))
	var landing := landing_override if landing_override != Vector3.INF else predicted_pos
	var start := Vector3(start_2d.x, landing.y, start_2d.y)
	var visual_duration: float = _skill_presentation_float("leap", "visual_duration")
	var apex := (start + landing) * 0.5 + Vector3(0.0, _skill_presentation_float("leap", "visual_height"), 0.0)
	local_leap_visual_active = true
	player_anchor.position = start
	var tween := create_tween()
	tween.set_parallel(true)
	tween.tween_property(player_anchor, "position:x", landing.x, visual_duration).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	tween.tween_property(player_anchor, "position:z", landing.z, visual_duration).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_IN_OUT)
	tween.chain().tween_callback(_finish_leap_visual)
	var y_tween := create_tween()
	y_tween.tween_property(player_anchor, "position:y", apex.y, visual_duration * 0.48).set_trans(Tween.TRANS_SINE).set_ease(Tween.EASE_OUT)
	y_tween.tween_property(player_anchor, "position:y", landing.y, visual_duration * 0.52).set_trans(Tween.TRANS_BOUNCE).set_ease(Tween.EASE_OUT)

func _finish_leap_visual() -> void:
	local_leap_visual_active = false
	if player_anchor != null:
		player_anchor.position = predicted_pos

func _play_charge_visual(ev: Dictionary, landing_override: Vector3 = Vector3.INF) -> void:
	if player_anchor == null or not ev.has("position"):
		return
	var start_2d := _vec2_from_dict(ev.get("position", {}))
	var landing := landing_override if landing_override != Vector3.INF else predicted_pos
	var start := Vector3(start_2d.x, landing.y, start_2d.y)
	var distance_tiles: float = max(0.01, Vector2(start.x, start.z).distance_to(Vector2(landing.x, landing.z)))
	var speed_tiles_per_second: float = ClientConstants.PLAYER_SPEED * _skill_mobility_float("charge", "speed_multiplier")
	var duration: float = max(0.08, distance_tiles / speed_tiles_per_second)
	local_charge_visual_active = true
	player_anchor.position = start
	_face_direction(Vector2(landing.x - start.x, landing.z - start.z))
	_mark_local_player_walking()
	var tween := create_tween()
	tween.set_parallel(true)
	tween.tween_property(player_anchor, "position:x", landing.x, duration).set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_IN_OUT)
	tween.tween_property(player_anchor, "position:z", landing.z, duration).set_trans(Tween.TRANS_LINEAR).set_ease(Tween.EASE_IN_OUT)
	tween.chain().tween_callback(_finish_charge_visual)

func _finish_charge_visual() -> void:
	local_charge_visual_active = false
	if player_anchor != null:
		player_anchor.position = predicted_pos

func _start_charge_channel_visual(ev: Dictionary) -> void:
	var direction := _vec2_from_dict(ev.get("direction", {}))
	_charge_channel_visual.start(player_anchor, character_visual, direction)
	_mark_local_player_walking()
	_face_direction(direction)

func _update_charge_channel_visual(ev: Dictionary) -> void:
	var direction := _vec2_from_dict(ev.get("direction", {}))
	_charge_channel_visual.update_direction(direction)
	_mark_local_player_walking()
	_face_direction(direction)

func _stop_charge_channel_visual() -> void:
	_charge_channel_visual.stop()

func _skill_presentation_float(skill_id: String, field: String) -> float:
	var presentation: Dictionary = SkillRulesLoaderScript.skill_presentation(skill_id)
	var value := float(presentation.get(field, 0.0))
	if value > 0.0:
		return value
	push_warning("skill presentation %s.%s must be positive" % [skill_id, field])
	return 0.1

func _skill_mobility_float(skill_id: String, field: String) -> float:
	var def: Dictionary = SkillRulesLoaderScript.skill_definition(skill_id)
	var mobility: Dictionary = def.get("mobility", {})
	var value := float(mobility.get(field, 0.0))
	if value > 0.0:
		return value
	push_warning("skill mobility %s.%s must be positive" % [skill_id, field])
	return 0.1

func _spawn_skill_projectile_visual(ev: Dictionary) -> void:
	var projectile_def_id := str(ev.get("projectile_def_id", ""))
	if projectile_def_id == "":
		return
	var pos := _vec2_from_dict(ev.get("position", {}))
	var dir := _vec2_from_dict(ev.get("direction", {}))
	if dir.length_squared() <= 0.0001:
		return
	dir = dir.normalized()
	var max_range := maxf(float(ev.get("range", 0.0)), 1.0)
	var start := Vector3(pos.x + dir.x * 0.45, 0.0, pos.y + dir.y * 0.45)
	var distance := minf(max_range, 8.5)
	var target_id := str(ev.get("target_entity_id", ""))
	if target_id != "" and entities.has(target_id):
		var target_node := (entities[target_id] as Dictionary).get("node", null) as Node3D
		if target_node != null:
			var target_pos := _node_world_or_local_position(target_node)
			var target_flat := Vector2(target_pos.x - start.x, target_pos.z - start.z)
			if target_flat.length_squared() > 0.0001:
				distance = clampf(target_flat.length(), 1.0, max_range)
	var arrow_count := 1
	var spread_degrees := 0.0
	if str(ev.get("skill_id", "")) == "volley":
		arrow_count = 5
		spread_degrees = 32.0
	for i in range(arrow_count):
		var shot_dir := dir
		if arrow_count > 1:
			var t := 0.0 if arrow_count == 1 else (float(i) / float(arrow_count - 1) - 0.5)
			shot_dir = dir.rotated(deg_to_rad(spread_degrees * t)).normalized()
		var finish := Vector3(start.x + shot_dir.x * distance, 0.0, start.z + shot_dir.y * distance)
		_spawn_single_projectile_visual(projectile_def_id, start, finish, i)

func _spawn_single_projectile_visual(projectile_def_id: String, start: Vector3, finish: Vector3, index: int = 0) -> void:
	var node := ProjectileVisualsScript.make_node(projectile_def_id)
	node.name = "SkillProjectilePreview_%s" % projectile_def_id
	node.position = start
	entities_root.add_child(node)
	var flat := Vector2(finish.x - start.x, finish.z - start.z)
	if flat.length_squared() > 0.0001:
		node.look_at(Vector3(finish.x, start.y, finish.z), Vector3.UP)
	var duration := 0.42
	if visual_replay_enabled:
		duration = maxf(autoplay_step_delay * 0.85, 0.32)
	var tween := create_tween()
	if index > 0:
		tween.tween_interval(0.025 * float(index))
	tween.tween_property(node, "position", finish, duration).set_trans(Tween.TRANS_LINEAR)
	tween.tween_callback(node.queue_free)

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
	_last_inventory_feedback_text = ClientConstants.BAG_FULL_CANT_UNEQUIP_TEXT
	if inventory_panel != null:
		inventory_panel.show_gesture_hint(ClientConstants.BAG_FULL_CANT_UNEQUIP_TEXT)
	_show_damage_number(player_id, Color("#ffcf5a"), null, "", 0.0, "inventory", ClientConstants.BAG_FULL_CANT_UNEQUIP_TEXT)

func _send_action_intent(target_id: String) -> void:
	if client == null or target_id == "":
		return
	var message_id := client.send("action_intent", last_server_tick, {"target_id": target_id})
	pending_action_targets[message_id] = {"target_id": target_id}

func _basic_attack_cooldown_seconds() -> float:
	var ticks := ClientConstants.DEFAULT_ATTACK_INTERVAL_TICKS
	var derived = character_progression.get("derived_stats", {})
	if typeof(derived) == TYPE_DICTIONARY:
		ticks = int(derived.get("attack_interval_ticks", ClientConstants.DEFAULT_ATTACK_INTERVAL_TICKS))
	if ticks <= 0:
		ticks = ClientConstants.DEFAULT_ATTACK_INTERVAL_TICKS
	return maxf(ClientConstants.SEND_INTERVAL, float(ticks) / ClientConstants.SERVER_TICK_RATE)

func _start_basic_attack_recovery_ui(duration_seconds: float = -1.0) -> void:
	if character_bar == null:
		return
	var duration := duration_seconds
	if duration <= 0.0:
		duration = _basic_attack_cooldown_seconds()
	character_bar.start_attack_recovery(duration)

func _remove_monster_health_bar(entity_id: String) -> void:
	if not monster_health_bars.has(entity_id):
		return
	var bar = monster_health_bars[entity_id]
	if is_instance_valid(bar):
		bar.queue_free()
	monster_health_bars.erase(entity_id)

func _remove_revive_corpse_status_bar(entity_id: String) -> void:
	if not revive_corpse_status_bars.has(entity_id):
		return
	var bar = revive_corpse_status_bars[entity_id]
	if is_instance_valid(bar):
		bar.queue_free()
	revive_corpse_status_bars.erase(entity_id)

func _upsert_revive_corpse_status_bar(entity_id: String, target: Node3D, text: String) -> void:
	if gameplay_ui_layer == null or _camera == null or target == null:
		return
	if revive_corpse_status_bars.has(entity_id):
		revive_corpse_status_bars[entity_id].setup(_camera, target, text)
		return
	var bar = CorpseStatusBarScript.new()
	gameplay_ui_layer.add_child(bar)
	bar.setup(_camera, target, text)
	revive_corpse_status_bars[entity_id] = bar

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
	_refresh_monster_health_bar_visibility(entity_id)

func _refresh_monster_health_bar_visibility(entity_id: String = "") -> void:
	var ids := [entity_id] if entity_id != "" else monster_health_bars.keys()
	var hovered_id := _pick_entity_at_mouse()
	var mode := client_settings.monster_health_bar_mode if client_settings != null else ClientSettingsScript.DEFAULT_MONSTER_HEALTH_BAR_MODE
	for raw_id in ids:
		var id := str(raw_id)
		if monster_health_bars.has(id) and is_instance_valid(monster_health_bars[id]):
			monster_health_bars[id].visible = EnemyHealthBarVisibilityScript.should_show(mode, id, hovered_id, pending_action_targets, pending_skill_casts)

func _sync_companion_bar() -> void:
	if companion_bar == null:
		return
	var companions: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) != "companion":
			continue
		if player_id != "" and str(rec.get("owner_id", "")) != player_id:
			continue
		var hp := int(rec.get("hp", 0))
		if hp <= 0:
			continue
		companions.append({
			"id": str(id),
			"monster_def_id": str(rec.get("monster_def_id", "")),
			"hp": hp,
			"max_hp": int(rec.get("max_hp", hp)),
			"visual_tint": str(rec.get("visual_tint", "")),
			"visual_model": str(rec.get("visual_model", "")),
			"remaining_ticks": int(rec.get("remaining_ticks", 0)),
			"total_ticks": int(rec.get("total_ticks", 0)),
			"companion_stance": str(rec.get("companion_stance", "assist")), "combat_stats": (rec.get("combat_stats", {}) as Dictionary).duplicate(true),
		})
	companion_bar.set_companions(companions)
	if mercenary_panel != null: mercenary_panel.set_companions(companions)

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
		if _is_quest_journal_key(event):
			if quest_journal_panel != null:
				_close_gameplay_panels("quest_journal")
				_sync_quest_journal()
				quest_journal_panel.toggle()
				_raise_gameplay_windows()
			get_viewport().set_input_as_handled()
			return
		if event.keycode == KEY_TAB or event.physical_keycode == KEY_TAB:
			if discovery_minimap != null: discovery_minimap.cycle_display_mode(); _sync_discovery_minimap()
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
		if (event as InputEventKey).keycode == KEY_R:
			if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN and player_hp > 0:
				client.send("swap_weapon_set_intent", last_server_tick, {})
			get_viewport().set_input_as_handled()
			return
		if (event as InputEventKey).keycode == KEY_L:
			_loot_filter.cycle()
			if client_settings != null:
				client_settings.set_loot_filter_mode(_loot_filter.mode_label())
			_refresh_loot_label_visibility()
			_update_level_hud()
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
				_adjust_camera_zoom(-ClientConstants.CAMERA_ZOOM_STEP)
			MOUSE_BUTTON_WHEEL_DOWN:
				_adjust_camera_zoom(ClientConstants.CAMERA_ZOOM_STEP)
	if event is InputEventMouseButton and not event.pressed:
		if event.button_index == MOUSE_BUTTON_RIGHT:
			_stop_active_channel_skill()
			get_viewport().set_input_as_handled()
			return
func _handle_input(delta: float) -> void:
	_update_loot_hover_label()
	if _input_locked() or client.ready_state() != WebSocketPeer.STATE_OPEN:
		if _sustained_click.active:
			_sustained_click.clear()
		_clear_pending_attack_commands()
		_stop_active_channel_skill()
		return
	_send_cooldown -= delta
	_attack_cooldown -= delta
	if player_hp <= 0:
		if _sustained_click.active:
			_sustained_click.clear()
		_clear_pending_attack_commands()
		_stop_active_channel_skill()
		return
	_tick_active_channel_skill(delta)
	_tick_attack_buffer(delta)
	_tick_sticky_attack()
	if _command_retarget_grace.tick_and_dispatch(delta, _attack_cooldown, client, last_server_tick, Callable(self, "_close_gameplay_panels_for_movement"), Callable(self, "_mark_local_player_walking")):
		_attack_cooldown = maxf(_attack_cooldown, ClientConstants.SEND_INTERVAL)
	if bot_mode:
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
		predicted_pos += Vector3(dir.x, 0, dir.y) * ClientConstants.PLAYER_SPEED * ClientConstants.SEND_INTERVAL
		_reconcile_player()
		_mark_local_player_walking()
		client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
		_send_cooldown = ClientConstants.SEND_INTERVAL
	if _hold_input_allowed():
		_tick_sustained_click()
	elif _sustained_click.active:
		_sustained_click.clear()
		_clear_pending_attack_commands()
func _is_inventory_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_I or event.physical_keycode == KEY_I or event.unicode == 105 or event.unicode == 73
func _is_character_stats_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_C or event.physical_keycode == KEY_C or event.unicode == 99 or event.unicode == 67

func _is_character_info_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_P or event.physical_keycode == KEY_P or event.unicode == 112 or event.unicode == 80

func _is_skills_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_K or event.physical_keycode == KEY_K or event.unicode == 107 or event.unicode == 75

func _is_quest_journal_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_J or event.physical_keycode == KEY_J or event.unicode == 106 or event.unicode == 74

func _is_skill_slot_key(event: InputEventKey) -> bool:
	return event.keycode == KEY_Q or event.physical_keycode == KEY_Q or event.unicode == 113 or event.unicode == 81

func _close_gameplay_panels(except: String = "") -> void:
	if not (except in ["inventory", "stats", "shop_with_inventory", "stash_with_inventory", "bishop", "mercenary", "market", "blacksmith"]) and inventory_panel != null:
		inventory_panel.hide_display()
	if not (except in ["shop", "shop_with_inventory"]):
		_hide_shop_panel()
	if not (except in ["stash", "stash_with_inventory"]):
		_hide_stash_panel()
	if except != "bishop":
		_hide_bishop_panel()
	if except != "mercenary" and mercenary_panel != null:
		mercenary_panel.hide_display()
	if except != "market":
		_hide_market_panel()
	if except != "blacksmith":
		_hide_blacksmith_panel()
	if not (except in ["stats", "skills", "inventory"]) and character_stats_panel != null:
		character_stats_panel.hide_display()
	if not (except in ["skills", "stats"]) and skills_panel != null:
		skills_panel.hide_display()
	if except != "quest_journal" and quest_journal_panel != null:
		quest_journal_panel.hide_display()
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
	_player_walk_linger = ClientConstants.WALK_ANIMATION_LINGER_SECONDS

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
	_clear_pending_attack_commands()
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
		var offset := 8 if event.shift_pressed else 0
		return int(code - KEY_F1) + offset
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
				predicted_pos += Vector3(dir.x, 0, dir.y) * ClientConstants.PLAYER_SPEED * ClientConstants.SEND_INTERVAL
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
			if autoplay_attack_cooldown <= 0.0:
				CombatLocalAttackPresentationScript.present_local_start(_local_attack_presentation, target_id, audio_controller, player_anim, "main_hand", CombatReachScript.local_player_attack_mode(inventory, equipped))
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
		if str(pick.get("kind", "")) == "floor":
			_clear_pending_attack_commands()
			if _command_retarget_grace.dispatch_or_queue_floor(pick.get("ground", Vector3.ZERO), _attack_cooldown, client, last_server_tick, Callable(self, "_close_gameplay_panels_for_movement"), Callable(self, "_mark_local_player_walking")):
				_attack_cooldown = maxf(_attack_cooldown, ClientConstants.SEND_INTERVAL)
		elif str(pick.get("kind", "")) == "monster":
			_defer_monster_click(str(pick.get("target_id", "")))
		else:
			_clear_pending_attack_commands()
		return
	if _is_force_stand_held():
		_clear_pending_attack_commands()
		_start_directional_attack_hold()
		return
	var kind := str(pick.get("kind", ""))
	if kind == "floor":
		_clear_pending_attack_commands()
		var ground: Vector3 = pick.get("ground", Vector3.ZERO)
		if _command_retarget_grace.dispatch_or_queue_floor(ground, _attack_cooldown, client, last_server_tick, Callable(self, "_close_gameplay_panels_for_movement"), Callable(self, "_mark_local_player_walking")):
			_attack_cooldown = maxf(_attack_cooldown, ClientConstants.SEND_INTERVAL)
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
		_clear_pending_attack_commands()
		_activate_or_approach_interactable(target_id, rec)
		return
	if typ == "monster":
		_try_dispatch_monster_attack(target_id, true)
		return
	_clear_pending_attack_commands()
	if player_anim != null and typ == "interactable" and state == "closed" and _target_in_local_attack_range(target_id):
		player_anim.play_one_shot("attack")
	_send_action_intent(target_id)
	_attack_cooldown = ClientConstants.SEND_INTERVAL

func _try_dispatch_monster_attack(target_id: String, allow_buffer: bool) -> bool:
	if not _living_monster_target(target_id):
		if allow_buffer:
			_clear_pending_attack_commands()
		return false
	var rec: Dictionary = entities[target_id]
	if not _target_in_local_attack_range(target_id):
		if allow_buffer:
			_start_attack_move(target_id)
		return false
	_dispatch_monster_attack_now(target_id, rec)
	return true

func _dispatch_monster_attack_now(target_id: String, rec: Dictionary) -> void:
	_clear_pending_attack_commands()
	var target_node := rec.get("node", null) as Node3D
	if target_node != null:
		var flat := Vector2(target_node.global_position.x - player_anchor.global_position.x, target_node.global_position.z - player_anchor.global_position.z)
		if flat.length_squared() > 0.0001:
			_face_direction(flat.normalized())
	CombatLocalAttackPresentationScript.present_local_start(_local_attack_presentation, target_id, audio_controller, player_anim, "main_hand", CombatReachScript.local_player_attack_mode(inventory, equipped))
	_send_action_intent(target_id)
	_attack_cooldown = _basic_attack_cooldown_seconds()
	_start_basic_attack_recovery_ui(_attack_cooldown)
func _queue_attack_buffer(target_id: String) -> void:
	if not _living_monster_target(target_id):
		_attack_buffer.clear()
		return
	_attack_buffer.queue_attack(target_id)

func _defer_monster_click(target_id: String) -> void:
	if _target_in_local_attack_range(target_id):
		_sticky_attack.clear()
		_queue_attack_buffer(target_id)
	else:
		_start_attack_move(target_id)
func _start_attack_move(target_id: String) -> void:
	if not _living_monster_target(target_id):
		_clear_pending_attack_commands()
		return
	_attack_buffer.clear()
	_sticky_attack.set_target(target_id)
	var goal := CombatReachScript.attack_approach_point(player_anchor, entities, inventory, equipped, target_id, _last_facing_direction)
	_close_gameplay_panels_for_movement()
	_mark_local_player_walking()
	client.send("move_to_intent", last_server_tick, {"position": {"x": goal.x, "y": goal.z}})
	if _attack_cooldown <= 0.0:
		_attack_cooldown = ClientConstants.SEND_INTERVAL

func _clear_pending_attack_commands() -> void:
	_attack_buffer.clear()
	_sticky_attack.clear()
func _living_monster_target(target_id: String) -> bool:
	return target_id != "" and entities.has(target_id) and str(entities[target_id].get("type", "")) == "monster" and int(entities[target_id].get("hp", 1)) > 0
func _target_in_local_attack_range(target_id: String) -> bool:
	return CombatReachScript.target_in_local_attack_range(player_anchor, entities, inventory, equipped, target_id)

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

func _tick_attack_buffer(delta: float) -> void:
	_attack_buffer.tick(delta)
	if not _attack_buffer.active():
		return
	if _attack_buffer.should_clear(player_hp, entities):
		_attack_buffer.clear()
		return
	if _attack_cooldown > 0.0:
		return
	_try_dispatch_monster_attack(_attack_buffer.target_id, false)

func _tick_sticky_attack() -> void:
	if not _sticky_attack.active():
		return
	if _sticky_attack.should_clear(player_hp, entities):
		_sticky_attack.clear()
		return
	if _attack_cooldown <= 0.0:
		_try_dispatch_monster_attack(_sticky_attack.target_id, false)

func _repeat_hold_attack() -> void:
	var target_id := _sustained_click.target_id
	if target_id == "" or not entities.has(target_id):
		_sustained_click.clear()
		return
	_try_dispatch_monster_attack(target_id, false)

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
	_attack_cooldown = ClientConstants.SEND_INTERVAL

func _try_action_at_mouse() -> void:
	if _attack_cooldown > 0.0 or player_hp <= 0:
		return
	if _is_force_stand_held():
		_start_directional_attack_hold()
		return

	_execute_click_pick(_resolve_click_at_mouse())

func _interactable_should_approach_before_action(interactable_def_id: String) -> bool:
	return interactable_def_id in [
		"hero_corpse",
		"stairs_down",
		"stairs_up",
		"teleporter",
		"town_vendor",
		"town_mystery_seller",
		"town_stash",
		"town_bishop",
		"town_market_board",
		"town_blacksmith",
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
		_attack_cooldown = ClientConstants.SEND_INTERVAL
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
	_attack_cooldown = ClientConstants.SEND_INTERVAL

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
	return flat.length() <= ClientConstants.INTERACTABLE_ACTIVATION_RANGE

func _activate_interactable_now(target_id: String, rec: Dictionary) -> void:
	var interactable_def_id := str(rec.get("interactable_def_id", ""))
	if interactable_def_id == "stairs_down":
		client.send("descend_intent", last_server_tick, {})
		_attack_cooldown = ClientConstants.SEND_INTERVAL
		return
	if interactable_def_id == "stairs_up":
		client.send("ascend_intent", last_server_tick, {})
		_attack_cooldown = ClientConstants.SEND_INTERVAL
		return
	if interactable_def_id == "teleporter":
		if bool(discovered_teleporters.get(current_level, false)):
			_show_waypoint_panel()
		else:
			_send_action_intent(target_id)
			_attack_cooldown = ClientConstants.SEND_INTERVAL
		return
	_send_action_intent(target_id)
	_attack_cooldown = ClientConstants.SEND_INTERVAL

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

func _face_event_source_toward_target(ev: Dictionary) -> void:
	var source_id := str(ev.get("source_entity_id", ""))
	var target_id := str(ev.get("target_entity_id", ""))
	if source_id == "" or target_id == "" or source_id == target_id:
		return
	var source_node := _node_for_entity_id(source_id)
	if source_node == null:
		return
	_face_node_toward_entity(source_node, target_id)

func _face_node_toward_entity(source_node: Node3D, target_id: String) -> void:
	var target_node := _node_for_entity_id(target_id)
	if source_node == null or target_node == null:
		return
	var source_pos := _node_world_or_local_position(source_node)
	var target_pos := _node_world_or_local_position(target_node)
	_face_entity_direction(source_node, Vector2(target_pos.x - source_pos.x, target_pos.z - source_pos.z))

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
		player_anim.play_one_shot("attack", CombatReachScript.local_player_attack_mode(inventory, equipped))
	client.send("directional_attack_intent", last_server_tick, DirectionalAttackInputScript.payload(direction))
	_attack_cooldown = _basic_attack_cooldown_seconds()
	_start_basic_attack_recovery_ui(_attack_cooldown)

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
		if _is_dead_monster(hit_entity_id) and not _revive_hover_enabled():
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
		if target_id != "" and _entity_uses_loot_label(target_id):
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
	for label_id in _loot_label_entity_ids():
		var id := str(label_id)
		var highlighted := id == hovered_loot_id
		var rarity := str(entities.get(id, {}).get("rarity", "common"))
		var allowed := _loot_filter.allows(rarity)
		var revealed := loot_label_reveal_held and allowed
		if str(entities.get(id, {}).get("type", "")) == "loot":
			var node := entities[id].get("node", null) as Node3D
			if node != null:
				node.visible = allowed or highlighted
				_set_pickable(node, allowed or highlighted)
		elif _is_dead_monster(id):
			var node := entities[id].get("node", null) as Node3D
			if node != null:
				_set_pickable(node, _revive_hover_enabled())
		_set_loot_label_visible(id, revealed or highlighted, highlighted)

func _loot_label_entity_ids() -> Array:
	var out: Array = []
	for loot_id in loot_ids:
		out.append(str(loot_id))
	for interactable_id in interactable_ids:
		var id := str(interactable_id)
		if _entity_uses_loot_label(id) and not out.has(id):
			out.append(id)
	for id in entities.keys():
		var entity_id := str(id)
		if _entity_uses_loot_label(entity_id) and not out.has(entity_id):
			out.append(entity_id)
	return out

func _entity_uses_loot_label(entity_id: String) -> bool:
	if entity_id == "" or not entities.has(entity_id):
		return false
	var rec: Dictionary = entities[entity_id]
	if str(rec.get("type", "")) == "loot" or str(rec.get("interactable_def_id", "")) == "hero_corpse":
		return true
	return _revive_hover_enabled() and _is_dead_monster(entity_id) and _loot_label_node(entity_id) != null

func _revive_hover_enabled() -> bool:
	return right_click_skill_id == "revive" and _skill_rank("revive") > 0

func _set_loot_label_visible(loot_id: String, shown: bool, highlighted: bool = false) -> void:
	if loot_id == "" or not entities.has(loot_id):
		return
	var label := _loot_label_node(loot_id)
	if label != null:
		label.visible = shown
		var rec: Dictionary = entities.get(loot_id, {})
		label.modulate = _loot_filter.display_color(_loot_label_color(rec), highlighted)
	var ring := _revive_hover_ring_node(loot_id)
	if ring != null:
		ring.visible = highlighted and _revive_hover_enabled()
	if _is_dead_monster(loot_id):
		if shown and highlighted and _revive_hover_enabled():
			var rec: Dictionary = entities.get(loot_id, {})
			_upsert_revive_corpse_status_bar(loot_id, rec.get("node", null) as Node3D, _monster_corpse_label_text(rec))
		else:
			_remove_revive_corpse_status_bar(loot_id)

func _loot_label_node(loot_id: String) -> Label3D:
	if loot_id == "" or not entities.has(loot_id):
		return null
	var node := entities[loot_id].get("node", null) as Node3D
	if node == null:
		return null
	return node.find_child("LootLabel", true, false) as Label3D

func _revive_hover_ring_node(entity_id: String) -> Node3D:
	if entity_id == "" or not entities.has(entity_id):
		return null
	var node := entities[entity_id].get("node", null) as Node3D
	if node == null:
		return null
	return node.find_child("ReviveHoverRing", true, false) as Node3D

func _ensure_dead_monster_revive_label(entity_id: String, rec: Dictionary) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	if node.find_child("LootLabel", true, false) == null:
		var marker := Label3D.new()
		marker.name = "LootLabel"
		marker.text = _monster_corpse_label_text(rec)
		marker.visible = false
		marker.font_size = 72
		marker.modulate = Color("#c6e0a3")
		marker.outline_size = 12
		marker.outline_modulate = Color(0.03, 0.06, 0.02, 0.95)
		marker.position = Vector3(0.0, 2.25, 0.0)
		marker.billboard = BaseMaterial3D.BILLBOARD_ENABLED
		marker.no_depth_test = true
		marker.fixed_size = true
		marker.pixel_size = 0.0024
		node.add_child(marker)
	if node.find_child("ReviveHoverRing", true, false) == null:
		var ring := MeshInstance3D.new()
		ring.name = "ReviveHoverRing"
		var mesh := TorusMesh.new()
		mesh.inner_radius = 0.56
		mesh.outer_radius = 0.62
		ring.mesh = mesh
		ring.rotation_degrees.x = 90.0
		ring.position = Vector3(0.0, 0.055, 0.0)
		ring.visible = false
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color("#96f06f")
		mat.emission_enabled = true
		mat.emission = Color("#4fcf4a")
		ring.material_override = mat
		node.add_child(ring)

func _monster_corpse_label_text(rec: Dictionary) -> String:
	var monster_def_id := str(rec.get("monster_def_id", ""))
	if monster_def_id == "":
		return "Corpse"
	return "%s Corpse" % monster_def_id.replace("_", " ").capitalize()

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

	_camera.size = clampf(_camera.size + delta_size, ClientConstants.CAMERA_ZOOM_MIN, ClientConstants.CAMERA_ZOOM_MAX)

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
	ground_node = _ground_factory.make_ground_node(current_level)
	add_child(ground_node)

	_camera = Camera3D.new()
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	_camera.size = ClientConstants.CAMERA_ZOOM_DEFAULT
	_camera.position = ClientConstants.CAMERA_FOLLOW_OFFSET
	add_child(_camera)
	# look_at requires the node to be inside the scene tree (Godot 4).
	_sync_camera_to_player()

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-50, -40, 0)
	add_child(light)

	audio_controller = ClientAudioControllerScript.new()
	add_child(audio_controller)

	var ui := CanvasLayer.new()
	ui.layer = 5
	add_child(ui)
	gameplay_ui_layer = ui
	_debug_label = Label.new()
	_debug_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_LEFT
	_debug_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	_debug_label.set_anchors_preset(Control.PRESET_TOP_LEFT)
	_debug_label.offset_left = 12
	_debug_label.offset_right = 560
	_debug_label.offset_top = 12
	_debug_label.offset_bottom = 260
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
	_boss_visuals_context = BossVisualsContextScript.new()
	_boss_visuals_context.entities = entities
	_boss_visuals_context.apply_model_tint = Callable(self, "_apply_model_tint")
	_boss_visuals_context.apply_entity_status_tint = Callable(self, "_apply_entity_status_tint")
	_boss_visuals = BossVisualsControllerScript.new(_boss_visuals_context, boss_health_bar)
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
	bishop_panel.revive_all_requested.connect(_on_bishop_revive_all_requested)
	bishop_panel.debug_requested.connect(_on_bishop_debug_requested)
	bishop_panel.set_debug_enabled(gameplay_debug_enabled)
	ui.add_child(bishop_panel)
	mercenary_panel = MercenaryPanelScript.new()
	mercenary_panel.stance_requested.connect(_on_companion_stance_requested)
	ui.add_child(mercenary_panel)
	market_panel = MarketPanelScript.new()
	market_panel.market_action_requested.connect(_on_market_action_requested)
	market_panel.inventory_context_requested.connect(_on_market_inventory_context_requested)
	market_panel.staged_offer_items_changed.connect(_on_market_staged_offer_items_changed)
	ui.add_child(market_panel)
	blacksmith_panel = BlacksmithPanelScript.new()
	blacksmith_panel.upgrade_requested.connect(_on_blacksmith_upgrade_requested)
	blacksmith_panel.upgrade_inventory_requested.connect(_on_blacksmith_inventory_upgrade_requested)
	ui.add_child(blacksmith_panel)
	consumable_bar = ConsumableBarScript.new()
	consumable_bar.intent_requested.connect(_on_inventory_intent_requested)
	ui.add_child(consumable_bar)
	character_stats_panel = CharacterStatsPanelScript.new()
	character_stats_panel.allocate_stat_requested.connect(_on_character_stat_requested)
	ui.add_child(character_stats_panel)
	skills_panel = SkillsPanelScript.new()
	skills_panel.allocate_skill_point_requested.connect(_on_skill_point_requested)
	ui.add_child(skills_panel)
	quest_journal_panel = QuestJournalPanelScript.new()
	ui.add_child(quest_journal_panel)
	elite_objective_tracker = EliteObjectiveTrackerScript.new()
	ui.add_child(elite_objective_tracker)
	discovery_minimap = DiscoveryMinimapScript.new(); ui.add_child(discovery_minimap)
	character_bar = CharacterBarScript.new()
	character_bar.open_character_requested.connect(_open_character_panel_from_bar)
	ui.add_child(character_bar)
	skill_bar = SkillBarScript.new()
	skill_bar.cast_skill_requested.connect(_on_skill_cast_requested)
	skill_bar.open_skills_requested.connect(_open_skills_panel_from_bar)
	ui.add_child(skill_bar)
	companion_bar = CompanionBarScript.new()
	companion_bar.companion_selected.connect(_on_companion_bar_selected)
	ui.add_child(companion_bar)
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
	fog_overlay = FogOfWarOverlay.new()
	add_child(fog_overlay)
	fog_overlay.bind(_camera, player_anchor)

	damage_numbers_layer = CanvasLayer.new()
	damage_numbers_layer.layer = 2
	add_child(damage_numbers_layer)

	health_bars_layer = CanvasLayer.new()
	health_bars_layer.layer = 1
	add_child(health_bars_layer)

	walls_root = Node3D.new()
	walls_root.name = "StaticWalls"
	add_child(walls_root)
	_wall_renderer = WallRenderer.new(walls_root, _ground_factory)

func _update_ground_material() -> void:
	_ground_factory.update_ground_material(ground_node, current_level)
	if _wall_renderer != null: _wall_renderer.set_level(current_level)

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
	settings_panel.monster_health_bar_mode_selected.connect(_on_monster_health_bar_mode_selected)
	settings_panel.master_volume_changed.connect(_on_master_volume_changed)
	settings_panel.music_volume_changed.connect(_on_music_volume_changed)
	settings_panel.sfx_volume_changed.connect(_on_sfx_volume_changed)
	settings_panel.map_opacity_changed.connect(_on_map_opacity_changed)
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
	if quest_journal_panel != null:
		quest_journal_panel.hide_display()
	if elite_objective_tracker != null:
		elite_objective_tracker.visible = false
	loss_popup.visible = true

func _on_inventory_intent_requested(intent_type: String, payload: Dictionary) -> void:
	if _input_locked() or client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	if intent_type == "stash_equip_item_intent":
		_send_stash_equip_item_intent(payload)
		return
	if TownServiceBridgeScript.route_inventory_stage_intent(intent_type, payload, market_panel, blacksmith_panel):
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
	if player_mana < _skill_mana_cost(resolved_skill_id):
		return "not_enough_mana"
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
	if slot_index < 0 or slot_index >= ClientConstants.SKILL_FUNCTION_KEY_COUNT or skill_id == "":
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
	if slot_index < 0 or slot_index >= ClientConstants.SKILL_FUNCTION_KEY_COUNT:
		return false
	_ensure_skill_function_key_slots()
	var skill_id := str(skill_function_keys[slot_index])
	if skill_id == "":
		return false
	return _assign_right_click_skill(skill_id)

func _handle_skill_function_key(slot_index: int) -> bool:
	if slot_index < 0 or slot_index >= ClientConstants.SKILL_FUNCTION_KEY_COUNT:
		return false
	if skills_panel != null and skills_panel.visible:
		var hovered_skill_id := skills_panel.hovered_skill_id()
		if hovered_skill_id != "":
			return _assign_skill_function_key(slot_index, hovered_skill_id)
	return _select_right_click_skill_from_function_key(slot_index)

func _ensure_skill_function_key_slots() -> void:
	while skill_function_keys.size() < ClientConstants.SKILL_FUNCTION_KEY_COUNT:
		skill_function_keys.append("")
	if skill_function_keys.size() > ClientConstants.SKILL_FUNCTION_KEY_COUNT:
		skill_function_keys.resize(ClientConstants.SKILL_FUNCTION_KEY_COUNT)

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
	skill_bar.set_character_progression(character_progression)
	skill_bar.set_skill_progression(skill_progression)
	skill_bar.set_skill_cooldowns(skill_cooldowns)
	skill_bar.set_player_mana(player_mana, player_max_mana)
	skill_bar.set_interactive(not _skill_cast_blocked(selected_skill_id))
	_refresh_loot_label_visibility()

func _tick_skill_cooldowns(delta: float) -> void:
	if skill_cooldowns.is_empty():
		return
	var elapsed_ticks := maxf(0.0, delta) * ClientConstants.SERVER_TICK_RATE
	var next: Array = []
	var changed := false
	for row in skill_cooldowns:
		if typeof(row) != TYPE_DICTIONARY:
			changed = true
			continue
		var rec := (row as Dictionary).duplicate(true)
		var remaining := float(rec.get("remaining_ticks", 0.0)) - elapsed_ticks
		if remaining <= 0.0:
			changed = true
			continue
		var next_remaining := int(ceil(remaining))
		if next_remaining != int(rec.get("remaining_ticks", 0)):
			changed = true
		rec["remaining_ticks"] = next_remaining
		next.append(rec)
	if changed:
		skill_cooldowns = next
		if skill_bar != null:
			skill_bar.set_skill_cooldowns(skill_cooldowns)

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
	if inventory_panel != null:
		inventory_panel.set_character_progression(character_progression)
	if character_stats_panel != null:
		character_stats_panel.set_hero_name(_local_character_display_name())
		character_stats_panel.set_progression(character_progression)
		character_stats_panel.set_allocation_enabled(not _stat_allocation_blocked())
	if character_bar != null:
		character_bar.set_progression(character_progression)
	if consumable_bar != null:
		consumable_bar.set_character_progression(character_progression)
	if fog_overlay != null: fog_overlay.set_progression(character_progression); _sync_discovery_minimap()

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
		skill_bar.set_player_mana(player_mana, player_max_mana)
		skill_bar.set_interactive(not _skill_cast_blocked(str(skill_bar.get_debug_state().get("skill_id", ""))))

func _try_use_right_click_skill() -> bool:
	if right_click_skill_id == "":
		return false
	if ChannelSkillInputScript.is_channel_skill(right_click_skill_id):
		return _channel_skill_input.try_start(right_click_skill_id, _aim_direction_from_mouse(), _last_facing_direction, _skill_cast_block_reason(right_click_skill_id), Callable(self, "_send_channel_skill_payload"), Callable(self, "_face_direction"), Callable(self, "_show_skill_rejected_feedback"), ClientConstants.SEND_INTERVAL)
	var pick := _resolve_click_at_mouse()
	var target_id := ""
	var direction := Vector2.ZERO
	if str(pick.get("kind", "")) == "monster":
		target_id = str(pick.get("target_id", ""))
	elif right_click_skill_id == "revive" and _is_dead_monster(str(pick.get("target_id", ""))):
		target_id = str(pick.get("target_id", ""))
	else:
		direction = _aim_direction_from_mouse()
	var sent := _send_skill_cast_intent(right_click_skill_id, target_id, direction, false)
	return sent

func _tick_active_channel_skill(delta: float) -> void:
	_channel_skill_input.tick_and_send(delta, Input.is_mouse_button_pressed(MOUSE_BUTTON_RIGHT), _aim_direction_from_mouse(), _last_facing_direction, Callable(self, "_send_channel_skill_payload"), Callable(self, "_face_direction"), ClientConstants.SEND_INTERVAL)

func _stop_active_channel_skill() -> void:
	_channel_skill_input.stop_and_send(Callable(self, "_send_channel_skill_payload"))

func _send_channel_skill_payload(payload: Dictionary) -> bool:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or payload.is_empty():
		return false
	client.send("channel_skill_intent", last_server_tick, payload)
	return true

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
	var pending_skill := {"skill_id": skill_id}
	if payload.has("target_id"):
		pending_skill["target_id"] = str(payload.get("target_id", ""))
	pending_skill_casts[message_id] = pending_skill
	_attack_cooldown = ClientConstants.SEND_INTERVAL
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
	if skill_id == "revive" and chosen_target == "" and hovered_loot_id != "" and _is_dead_monster(hovered_loot_id):
		chosen_target = hovered_loot_id
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
			return ClientConstants.NO_MANA_TEXT
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

func _skill_mana_cost(skill_id: String) -> int:
	var rank := _skill_rank(skill_id)
	if rank <= 0:
		return 0
	var def := SkillRulesLoader.skill_definition(skill_id)
	var cost: Dictionary = def.get("cost", {})
	var mana: Dictionary = cost.get("mana", {})
	return maxi(0, int(mana.get("base", 0)) + int(mana.get("per_rank", 0)) * maxi(0, rank - 1))

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

func _show_corpse_panel(ev: Dictionary) -> void:
	if stash_panel == null:
		return
	_close_gameplay_panels("stash_with_inventory")
	var next_entity_id := str(ev.get("entity_id", ""))
	var corpse_name := str(ev.get("corpse_name", "Hero"))
	var corpse_items: Array = ev.get("corpse_items", [])
	if inventory_panel != null:
		inventory_panel.ensure_display_visible()
	stash_panel.show_corpse(
		next_entity_id,
		corpse_name,
		corpse_items,
		inventory,
		equipped,
		gold,
		hotbar
	)
	_raise_gameplay_windows()

func _show_unique_chest_panel(ev: Dictionary) -> void:
	if stash_panel == null:
		return
	_close_gameplay_panels("stash_with_inventory")
	var next_entity_id := str(ev.get("entity_id", ""))
	var chest_items: Array = ev.get("stash_items", [])
	if inventory_panel != null:
		inventory_panel.ensure_display_visible()
	stash_panel.show_unique_chest(
		next_entity_id,
		chest_items,
		inventory,
		equipped,
		gold,
		hotbar
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
		int(ev.get("price", 0)),
		bool(ev.get("affordable", gold >= int(ev.get("price", 0)))),
		gold, resource_wallet
	)
	_raise_gameplay_windows()

func _hide_bishop_panel() -> void:
	if bishop_panel != null:
		bishop_panel.hide_display()

func _show_market_panel(ev: Dictionary) -> void:
	if market_panel == null:
		return
	_close_gameplay_panels("market")
	TownServiceBridgeScript.open_market_inventory_context(inventory_panel)
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
	market_panel.show_market(next_entity_id, listings, inventory, client.account_id if client != null else "", status, equipped)
	_raise_gameplay_windows()

func _on_market_action_requested(action: String, payload: Dictionary) -> void:
	if client == null:
		return
	var result := {}
	if action == "publish":
		result = client.create_market_listing(str(payload.get("stash_item_id", "")), int(payload.get("price_gold", 0)))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not publish item", true)
			return
		_remove_market_stash_item(str(payload.get("stash_item_id", "")))
		_refresh_inventory_ui()
		if market_panel != null: market_panel.show_status("Item published")
	elif action == "publish_inventory":
		result = client.create_market_listing_from_inventory(str(payload.get("item_instance_id", "")), client.character_id, int(payload.get("price_gold", 0)))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not publish item", true)
			return
		_remove_inventory_item(str(payload.get("item_instance_id", "")))
		_refresh_inventory_ui()
		if market_panel != null: market_panel.show_status("Item published")
	elif action == "cancel_listing":
		result = client.cancel_market_listing(str(payload.get("listing_id", "")))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not cancel listing", true)
			return
		_upsert_stash_item(result)
		_refresh_inventory_ui()
		if market_panel != null: market_panel.show_status("Listing canceled")
	elif action == "offer":
		result = client.create_market_offer(str(payload.get("listing_id", "")), payload.get("stash_item_ids", []))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not make offer", true)
			return
		for stash_item_id in payload.get("stash_item_ids", []):
			_remove_market_stash_item(str(stash_item_id))
		_refresh_inventory_ui()
		if market_panel != null:
			market_panel.show_status("Offer sent")
			market_panel.return_to_browse_after_offer()
	elif action == "offer_inventory":
		result = client.create_market_offer_from_inventory(str(payload.get("listing_id", "")), payload.get("item_instance_ids", []), client.character_id)
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not make offer", true)
			return
		for item_instance_id in payload.get("item_instance_ids", []):
			_remove_inventory_item(str(item_instance_id))
		_refresh_inventory_ui()
		if market_panel != null:
			market_panel.show_status("Offer sent")
			market_panel.return_to_browse_after_offer()
	elif action == "purchase":
		result = client.purchase_market_listing(str(payload.get("listing_id", "")))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not purchase listing", true)
			return
		var delivered_item_raw = result.get("delivered_item", {})
		var delivered_item: Dictionary = delivered_item_raw if typeof(delivered_item_raw) == TYPE_DICTIONARY else {}
		if not delivered_item.is_empty():
			_upsert_stash_item(delivered_item)
			_refresh_inventory_ui()
		if market_panel != null:
			market_panel.show_status("Listing purchased")
	elif action == "list_offers" or action == "list_my_offers" or action == "list_market_receipts":
		var receipts := action == "list_market_receipts"
		var mine := action == "list_my_offers"
		result = client.list_market_receipts() if receipts else (client.list_my_market_offers() if mine else client.list_market_offers(str(payload.get("listing_id", ""))))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not load receipts" if receipts else ("Could not load my offers" if mine else "Could not load offers"), true)
			return
		if market_panel != null and receipts: market_panel.show_receipts(result.get("receipts", []), "Receipts")
		elif market_panel != null and mine: market_panel.show_my_offers(result.get("offers", []), "My offers")
		elif market_panel != null: market_panel.show_offers(str(payload.get("listing_id", "")), result.get("offers", []), "Active offers")
		return
	elif action == "accept_offer" or action == "cancel_offer":
		var cancel := action == "cancel_offer"
		result = client.cancel_market_offer(str(payload.get("listing_id", "")), str(payload.get("offer_id", ""))) if cancel else client.accept_market_offer(str(payload.get("listing_id", "")), str(payload.get("offer_id", "")))
		if result.has("_error"):
			if market_panel != null: market_panel.show_status("Could not cancel offer" if cancel else "Could not accept offer", true)
			return
		for item in result.get("items", []):
			if typeof(item) == TYPE_DICTIONARY:
				_upsert_stash_item(item as Dictionary)
		_refresh_inventory_ui()
		if cancel:
			var offers := client.list_my_market_offers()
			if market_panel != null: market_panel.show_my_offers(offers.get("offers", []), "Offer canceled" if not offers.has("_error") else "Offer canceled; refresh failed")
			return
		if market_panel != null: market_panel.show_status("Offer accepted")
	else:
		return
	_refresh_market_panel_data()

func _on_market_inventory_context_requested(context: String) -> void:
	TownServiceBridgeScript.set_market_inventory_context(inventory_panel, context)
	if context != "" and market_panel != null:
		_on_market_staged_offer_items_changed(market_panel.get_debug_state().get("staged_offer_item_ids", []))
	else:
		_on_market_staged_offer_items_changed([])

func _on_market_staged_offer_items_changed(item_instance_ids: Array) -> void:
	if inventory_panel != null:
		inventory_panel.set_market_hidden_item_ids(item_instance_ids)

func _refresh_market_panel_data() -> void:
	_refresh_market_board_summary()
	if market_panel == null or client == null:
		return
	var body := client.list_market_listings()
	var listings: Array = body.get("listings", [])
	market_panel.show_market(market_panel.market_entity_id, listings, inventory, client.account_id, market_panel.get_debug_state().get("status", ""), equipped)

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
		MarketBoardBadgesScript.apply_to_board(node, incoming_bids, published_listings)

func _market_board_badge_debug_state() -> Dictionary:
	for id in interactable_ids:
		var key := str(id)
		if not entities.has(key): continue
		var rec: Dictionary = entities[key]
		if str(rec.get("interactable_def_id", "")) != "town_market_board": continue
		var node := rec.get("node", null) as Node3D
		if node != null:
			return MarketBoardBadgesScript.debug_state(node)
	return MarketBoardBadgesScript.empty_state()

func _hide_market_panel() -> void:
	if market_panel != null:
		market_panel.hide_display()
	TownServiceBridgeScript.close_market_inventory_context(inventory_panel)
	_on_market_staged_offer_items_changed([])

func _show_blacksmith_panel(ev: Dictionary) -> void:
	if blacksmith_panel == null:
		return
	_close_gameplay_panels("blacksmith")
	TownServiceBridgeScript.open_blacksmith_inventory_context(inventory_panel)
	var next_entity_id := str(ev.get("entity_id", ""))
	stash_items = ev.get("stash_items", stash_items)
	stash_gold = int(ev.get("stash_gold", stash_gold))
	stash_capacity = int(ev.get("stash_capacity", stash_capacity))
	blacksmith_panel.show_blacksmith(next_entity_id, inventory, gold, stash_gold, _blacksmith_config(), "Choose an inventory item to upgrade", resource_wallet)
	_raise_gameplay_windows()

func _hide_blacksmith_panel() -> void:
	if blacksmith_panel != null:
		blacksmith_panel.hide_display()
	TownServiceBridgeScript.close_blacksmith_inventory_context(inventory_panel)

func _on_blacksmith_upgrade_requested(stash_item_id: String) -> void:
	if client == null or stash_item_id == "":
		return
	var result := client.upgrade_account_stash_item(stash_item_id, blacksmith_panel.selected_recipe_id() if blacksmith_panel != null and blacksmith_panel.has_method("selected_recipe_id") else "item_upgrade")
	if result.has("_error"):
		if blacksmith_panel != null:
			blacksmith_panel.show_status("Could not upgrade item", true)
		return
	var item: Dictionary = result.get("item", {})
	gold = int(result.get("gold", gold))
	stash_gold = int(result.get("stash_gold", stash_gold))
	_upsert_stash_item(item)
	if blacksmith_panel != null:
		blacksmith_panel.update_after_upgrade(item, gold, stash_gold, int(result.get("cost_gold", 0)), bool(result.get("success", true)), resource_wallet)
	if stash_panel != null and stash_panel.visible:
		stash_panel.set_stash_state(stash_items, stash_gold, stash_capacity)

func _on_blacksmith_inventory_upgrade_requested(item_instance_id: String) -> void:
	if client == null or item_instance_id == "":
		return
	var result := client.upgrade_inventory_item(item_instance_id, client.character_id, blacksmith_panel.selected_recipe_id() if blacksmith_panel != null and blacksmith_panel.has_method("selected_recipe_id") else "item_upgrade")
	if result.has("_error"):
		if blacksmith_panel != null:
			blacksmith_panel.show_status("Could not upgrade item", true)
		return
	var item: Dictionary = result.get("item", {})
	gold = int(result.get("gold", gold))
	stash_gold = int(result.get("stash_gold", stash_gold))
	_apply_upgrade_resource_wallet_response(result)
	_update_inventory_item(item)
	if blacksmith_panel != null:
		blacksmith_panel.update_after_upgrade(item, gold, stash_gold, int(result.get("cost_gold", 0)), bool(result.get("success", true)), resource_wallet)
	_refresh_inventory_ui()

func _apply_upgrade_resource_wallet_response(result: Dictionary) -> void:
	var resource_id := str(result.get("resource_item_def_id", ""))
	if resource_id == "" or not result.has("resource_wallet"):
		return
	resource_wallet[resource_id] = max(0, int(result.get("resource_wallet", 0)))

func _blacksmith_config() -> Dictionary:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/main_config.v0.json")
	var parsed = _read_json(path)
	if typeof(parsed) == TYPE_DICTIONARY:
		return (parsed as Dictionary).get("gameplay", {})
	return {}

func _on_bishop_respec_requested(bishop_entity_id: String) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or bishop_entity_id == "":
		return
	client.send("bishop_respec_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})


func _on_bishop_revive_all_requested(bishop_entity_id: String) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or bishop_entity_id == "":
		return
	client.send("bishop_revive_all_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})


func _on_bishop_debug_requested(action: String, bishop_entity_id: String) -> void:
	if client == null or client.ready_state() != WebSocketPeer.STATE_OPEN or bishop_entity_id == "" or not gameplay_debug_enabled:
		return
	match action:
		"level":
			client.send("bishop_debug_level_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})
		"skill_point":
			client.send("bishop_debug_skill_point_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})
		"stat_point":
			client.send("bishop_debug_stat_point_intent", last_server_tick, {"bishop_entity_id": bishop_entity_id})

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

func _make_hero_corpse_node(e: Dictionary) -> Node3D:
	var root := Node3D.new()
	root.name = "HeroCorpse_%s" % str(e.get("corpse_character_id", e.get("id", "")))
	var body := CharacterScene.instantiate() as Node3D
	body.name = "FallenHeroBody"
	body.rotation_degrees = Vector3(0.0, 0.0, -88.0)
	body.position = Vector3(0.05, 0.18, 0.0)
	body.scale = Vector3.ONE * 0.82
	_apply_model_tint(body, Color("#d4af37"))
	root.add_child(body)

	var shadow := MeshInstance3D.new()
	shadow.name = "CorpseShadow"
	var shadow_mesh := CylinderMesh.new()
	shadow_mesh.top_radius = 0.75
	shadow_mesh.bottom_radius = 0.75
	shadow_mesh.height = 0.025
	shadow.mesh = shadow_mesh
	shadow.scale.z = 0.48
	shadow.position = Vector3(0.0, 0.015, 0.0)
	var shadow_mat := StandardMaterial3D.new()
	shadow_mat.albedo_color = Color("#171412")
	shadow.material_override = shadow_mat
	root.add_child(shadow)

	var marker := Label3D.new()
	marker.name = "LootLabel"
	var corpse_name := str(e.get("corpse_name", "Hero"))
	var corpse_level := int(e.get("corpse_level", 0))
	marker.text = "%s Lv %d" % [corpse_name, corpse_level] if corpse_level > 0 else corpse_name
	marker.visible = false
	marker.font_size = 60
	marker.modulate = Color("#e8dcc8")
	marker.outline_size = 10
	marker.position = Vector3(0.0, 1.15, 0.0)
	marker.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	root.add_child(marker)
	return root

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
	if blacksmith_panel != null and blacksmith_panel.visible:
		if not _panel_source_in_activation_range(blacksmith_panel.blacksmith_entity_id):
			_hide_blacksmith_panel()
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
	var blacksmith_visible := blacksmith_panel != null and blacksmith_panel.visible
	if not shop_visible and not stash_visible and not bishop_visible and not market_visible and not blacksmith_visible:
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
	_ensure_wall_renderer()
	current_wall_layout = _wall_renderer.render_world_walls(world_id) if _wall_renderer != null else []
	_sync_fog_wall_layout()

func _render_wall_layout(walls: Array) -> void:
	_ensure_wall_renderer()
	current_wall_layout = _wall_renderer.render_wall_layout(walls) if _wall_renderer != null else []
	_sync_fog_wall_layout()
func _sync_fog_wall_layout() -> void:
	if fog_overlay != null: InteractableRulesLoader.sync_fog_overlay(fog_overlay, current_wall_layout, interactable_ids, entities)
	_sync_discovery_minimap()
func _clear_wall_nodes() -> void:
	_ensure_wall_renderer()
	if _wall_renderer != null:
		_wall_renderer.clear_wall_nodes()

func _ensure_wall_renderer() -> void:
	if _wall_renderer == null and walls_root != null:
		_wall_renderer = WallRenderer.new(walls_root, _ground_factory)

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
	if _loot_filter.is_active():
		lines.append("Loot: " + _loot_filter.mode_label())
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
	if kind == "monster" or kind == "companion":
		var visual := MonsterVisualsLoaderScript.resolve(str(e.get("monster_def_id", "")), str(e.get("visual_model", "")))
		var packed := _monster_scene_for_visual(str(visual.get("scene", "monster_dummy")))
		if str(e.get("visual_model", "")) == ClientConstants.BOSS_VISUAL_MODEL:
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
		if def_id == "hero_corpse":
			return _make_hero_corpse_node(e)
		return TownNodeFactoryScript.make_interactable_node(def_id, bool(e.get("elite_objective", false)), bool(e.get("quest_reward", false)))
	if kind == "projectile":
		return ProjectileVisualsScript.make_node(str(e.get("projectile_def_id", "")))
	return _loot_factory.make_loot_node(e)

func _monster_scene_for_visual(scene_key: String) -> PackedScene:
	return MonsterScenesByVisual.get(scene_key, MonsterScenesByVisual["monster_dummy"]) as PackedScene

func _make_remote_player_node(e: Dictionary) -> Node3D:
	var root = CharacterScene.instantiate() as Node3D
	root.name = "RemotePlayer_%s" % str(e.get("id", ""))
	_apply_character_class_model(root, str(e.get("character_class", "")))
	root.scale = Vector3.ONE * _entity_visual_scale(e)
	_apply_model_tint(root, ClientConstants.REMOTE_PLAYER_TINT)
	return root

func _monster_tint(rarity: String) -> Color:
	return ClientConstants.MONSTER_RARITY_TINTS.get(rarity, ClientConstants.MONSTER_RARITY_TINTS["common"])

func _entity_base_tint(e: Dictionary) -> Color:
	var kind := str(e.get("type", ""))
	if kind == "player":
		return ClientConstants.REMOTE_PLAYER_TINT
	if kind == "monster" or kind == "companion":
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
	if player_reaction != null:
		player_reaction.dispose()
	_apply_character_class_model(character_visual, class_id)
	_apply_local_player_visual_scale(player_visual_scale)
	_apply_model_tint(character_visual, ClientConstants.PLAYER_TINT)
	_remount_local_equipment_visuals()
	player_reaction = ModelReactionControllerScript.new(character_visual, ClientConstants.PLAYER_TINT)
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)

func _apply_character_class_model(root: Node3D, class_id: String) -> void:
	var resolved := ClassPresentationsLoaderScript.resolve(class_id)
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
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	root.add_child(model)
	root.move_child(model, 0)
	var ap := root.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		ap.root_node = NodePath("../ModelRoot")
	if root.has_method("_ensure_weapon_socket"):
		root.call("_ensure_weapon_socket")
	if root.has_method("_ensure_fallback_sockets"):
		root.call("_ensure_fallback_sockets")

func _remount_local_equipment_visuals() -> void:
	if resolver == null:
		return
	resolver.apply_snapshot({"inventory": inventory, "equipped": equipped})

func _apply_entity_visual_metadata(rec: Dictionary, e: Dictionary) -> void:
	if e.has("monster_def_id"):
		rec["monster_def_id"] = str(e["monster_def_id"])
		var visual := MonsterVisualsLoaderScript.resolve(str(e["monster_def_id"]), str(e.get("visual_model", "")))
		if not e.has("visual_model"):
			rec["visual_model"] = str(visual.get("visual_model", visual.get("scene", "")))
	for key in ["boss_template_id", "visual_model", "visual_tint", "boss_phase", "effect_ids", "monster_pack_id", "monster_pack_leader"]:
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
	var alive := int(rec.get("hp", 1)) > 0
	PlayerStatusEffectMarkers.sync_holy_shield_effect(node, rec.get("effect_ids", []) if alive else [])
	PlayerStatusEffectMarkers.sync_sanctuary_effect(node, rec.get("effect_ids", []) if alive else [], _sanctuary_radius())
	PlayerStatusEffectMarkers.sync_burning_effect(node, alive and PlayerStatusEffectMarkers.has_burning_effect_id(rec.get("effect_ids", [])))
	PlayerStatusEffectMarkers.sync_elite_command_effect(node, alive and PlayerStatusEffectMarkers.has_elite_command_effect_id(rec.get("effect_ids", [])))
	PlayerStatusEffectMarkers.sync_pinning_root_effect(node, alive and PlayerStatusEffectMarkers.has_pinning_root_effect_id(rec.get("effect_ids", [])))
	PlayerStatusEffectMarkers.sync_stun_effect(node, alive and PlayerStatusEffectMarkers.has_stun_effect_id(rec.get("effect_ids", [])))
	PlayerStatusEffectMarkers.sync_rogue_mark_effect(node, alive and PlayerStatusEffectMarkers.has_rogue_mark_effect_id(rec.get("effect_ids", [])))
	if _boss_visuals != null:
		_boss_visuals_context.last_server_tick = last_server_tick
		_boss_visuals.normalize_boss_phase_metadata(rec)
		_boss_visuals.sync_boss_telegraph_marker_from_record(rec)
	EliteAuraPreviewSync.sync(entities, dungeon_generation)

func _sync_archer_bow_marker(root: Node3D, monster_def_id: String) -> void:
	if root == null:
		return
	var existing := root.find_child(ClientConstants.ARCHER_BOW_MARKER_NAME, true, false) as Node3D
	if monster_def_id != ClientConstants.ARCHER_MONSTER_DEF_ID:
		if existing != null:
			existing.queue_free()
		return
	if existing == null:
		existing = _make_archer_bow_marker()
		root.add_child(existing)
	_apply_archer_bow_material(existing)

func _has_archer_bow_marker(root: Node3D) -> bool:
	return root != null and root.find_child(ClientConstants.ARCHER_BOW_MARKER_NAME, true, false) != null

func _make_archer_bow_marker() -> Node3D:
	var marker := Node3D.new()
	marker.name = ClientConstants.ARCHER_BOW_MARKER_NAME
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

func _set_entity_pinning_root(entity_id: String, active: bool) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	var effect_ids: Array = rec.get("effect_ids", []) if rec.get("effect_ids", []) is Array else []
	if active:
		if not effect_ids.has(PlayerStatusEffectMarkers.PINNING_ROOT_EFFECT_ID):
			effect_ids.append(PlayerStatusEffectMarkers.PINNING_ROOT_EFFECT_ID)
	else:
		effect_ids.erase(PlayerStatusEffectMarkers.PINNING_ROOT_EFFECT_ID)
	rec["effect_ids"] = effect_ids
	var node := rec.get("node", null) as Node3D
	PlayerStatusEffectMarkers.sync_pinning_root_effect(node, active)

func _set_entity_stun(entity_id: String, active: bool) -> void:
	var rec: Dictionary = entities.get(entity_id, {})
	if rec.is_empty():
		return
	var effect_ids: Array = rec.get("effect_ids", []) if rec.get("effect_ids", []) is Array else []
	if active:
		if not effect_ids.has(PlayerStatusEffectMarkers.STUN_EFFECT_ID):
			effect_ids.append(PlayerStatusEffectMarkers.STUN_EFFECT_ID)
	else:
		effect_ids.erase(PlayerStatusEffectMarkers.STUN_EFFECT_ID)
		effect_ids.erase("leap_stun")
		effect_ids.erase("charge_stun")
		effect_ids.erase(PlayerStatusEffectMarkers.DASH_STUN_EFFECT_ID)
	rec["effect_ids"] = effect_ids
	var node := rec.get("node", null) as Node3D
	PlayerStatusEffectMarkers.sync_stun_effect(node, active)

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
		tint = ClientConstants.POISON_TINT
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

func _loot_label_color(e: Dictionary) -> Color:
	return _loot_factory.loot_label_color(e)

func _item_definition(item_def_id: String) -> Dictionary:
	return _loot_factory.item_definition(item_def_id)

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
	_loot_factory.configure(asset_manifest, item_presentations)

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
	var duration := ClientConstants.PROJECTILE_LERP_SECONDS
	if visual_replay_enabled:
		duration = clampf(autoplay_step_delay * 0.35, 0.06, 0.18)
	var tween := create_tween()
	rec["move_tween"] = tween
	tween.tween_property(node, "position", target_pos, duration).set_trans(Tween.TRANS_LINEAR)

func _attach_pick_collider(node: Node3D, entity_id: String, kind: String, interactable_def_id: String = "") -> void:
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
			if interactable_def_id == "hero_corpse":
				box.size = Vector3(1.8, 0.75, 1.35)
				shape.position = Vector3(0.0, 0.35, 0.0)
			else:
				box.size = InteractableRulesLoader.pick_collider_size(interactable_def_id)
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
	if InteractableRulesLoader.has_barrier_when_closed(str(rec.get("interactable_def_id", ""))): _sync_fog_wall_layout()
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
	if def_id == "treasure_chest" or def_id == "town_stash" or def_id == "town_unique_chest":
		ChestPresentationScript.sync_objective_marker(node, bool(rec.get("elite_objective", false)), state == "open")
		ChestPresentationScript.sync_quest_marker(node, bool(rec.get("quest_reward", false)), state == "open")
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
				lock_mat.albedo_color = Color("#c77dff") if def_id == "town_unique_chest" else (Color("#d1b15d") if def_id == "town_stash" else Color("#8d8f8f"))
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
			mat.albedo_color = TownNodeFactoryScript.stair_base_color(def_id, state)
		else:
			mat.albedo_color = TownNodeFactoryScript.stair_base_color(def_id, state)
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
		"movement_visual_smoothing": _movement_visual_smoothing.get_debug_state(character_visual),
		"command_retarget_grace": _command_retarget_grace.get_debug_state(),
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
		"resource_wallet": resource_wallet.duplicate(true),
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
		"mercenary_panel_visible": mercenary_panel != null and mercenary_panel.visible,
		"market_panel_visible": market_panel != null and market_panel.visible,
		"market_board_badges": _market_board_badge_debug_state(),
		"blacksmith_panel_visible": blacksmith_panel != null and blacksmith_panel.visible,
		"character_stats_panel_visible": character_stats_panel != null and character_stats_panel.visible,
		"skills_panel_visible": skills_panel != null and skills_panel.visible,
		"quest_journal_panel_visible": quest_journal_panel != null and quest_journal_panel.visible,
		"elite_objective_tracker_visible": elite_objective_tracker != null and elite_objective_tracker.visible,
		"character_info_panel_visible": character_info_panel != null and character_info_panel.visible,
		"waypoint_panel_visible": waypoint_panel != null and waypoint_panel.visible,
		"inventory_panel": inventory_panel.get_debug_state() if inventory_panel != null else {},
		"shop_panel": shop_panel.get_debug_state() if shop_panel != null else {},
		"stash_panel": stash_panel.get_debug_state() if stash_panel != null else {},
		"bishop_panel": bishop_panel.get_debug_state() if bishop_panel != null else {},
		"mercenary_panel": mercenary_panel.get_debug_state() if mercenary_panel != null else {},
		"market_panel": market_panel.get_debug_state() if market_panel != null else {},
		"blacksmith_panel": blacksmith_panel.get_debug_state() if blacksmith_panel != null else {},
		"character_stats_panel": character_stats_panel.get_debug_state() if character_stats_panel != null else {},
		"skills_panel": skills_panel.get_debug_state() if skills_panel != null else {},
		"quest_journal_panel": quest_journal_panel.get_debug_state() if quest_journal_panel != null else {},
		"elite_objective_tracker": elite_objective_tracker.get_debug_state() if elite_objective_tracker != null else {},
		"discovery_minimap": discovery_minimap.get_debug_state() if discovery_minimap != null else {},
		"character_bar": character_bar.get_debug_state() if character_bar != null else {},
		"skill_bar": skill_bar.get_debug_state() if skill_bar != null else {},
		"companion_bar": companion_bar.get_debug_state() if companion_bar != null else {"visible": false, "count": 0, "companions": []},
		"status_effects_bar": status_effects_bar.get_debug_state() if status_effects_bar != null else {"effects": [], "visible": false},
		"boss_health_bar": boss_health_bar.get_debug_state() if boss_health_bar != null else {"visible": false},
		"fog_of_war": fog_overlay.get_debug_state() if fog_overlay != null else {},
		"audio": audio_controller.get_debug_state() if audio_controller != null else {},
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
		"map_opacity": client_settings.map_opacity if client_settings != null else ClientSettingsScript.DEFAULT_MAP_OPACITY,
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
			"interactable_def_id": str(rec.get("interactable_def_id", "")), "elite_objective": bool(rec.get("elite_objective", false)), "quest_reward": bool(rec.get("quest_reward", false)), "is_boss": bool(rec.get("is_boss", false)), "boss_template_id": str(rec.get("boss_template_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"state": str(rec.get("state", "")),
		})
	return out

func _bot_local_player_presentation() -> Dictionary:
	return {
		"id": player_id, "type": "player", "visual_model": "character", "visual_scale": player_visual_scale,
		"effect_ids": _local_player_effect_ids(),
		"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor),
		"has_sanctuary_effect": PlayerStatusEffectMarkers.has_sanctuary_effect(player_anchor),
		"holy_shield_aura_pulses": PlayerStatusEffectMarkers.active_holy_shield_aura_pulse_count(player_anchor),
		"holy_shield_target_pulses": PlayerStatusEffectMarkers.active_holy_shield_target_pulse_count(player_anchor),
		"has_rage_effect": PlayerStatusEffectMarkers.has_rage_effect(player_anchor), "base_tint": ClientConstants.PLAYER_TINT.to_html(false),
		"charge_channel_visual": _charge_channel_visual.get_debug_state(),
		"reaction": player_reaction.get_debug_state() if player_reaction != null else {},
		"animation": player_anim.get_debug_state() if player_anim != null else {},
	}

func _local_player_effect_ids() -> Array:
	var out := []
	if player_anchor == null:
		return out
	if PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor):
		out.append(PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID)
	if PlayerStatusEffectMarkers.has_sanctuary_effect(player_anchor):
		out.append(PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID)
	return out

func _character_info_debug_state() -> Dictionary:
	return {
		"visible": character_info_panel != null and character_info_panel.visible,
		"name": _local_character_display_name(),
		"level": int(character_progression.get("level", 1)),
		"area": _current_area_label(),
	}

func _sync_quest_journal() -> void:
	if quest_journal_panel != null:
		quest_journal_panel.set_objectives(QuestEliteObjectiveStateScript.quest_journal_objectives(entities))

func _sync_elite_objective_tracker() -> void:
	if elite_objective_tracker != null:
		elite_objective_tracker.set_state(QuestEliteObjectiveStateScript.elite_tracker_state(entities))

func _sync_discovery_minimap() -> void:
	if discovery_minimap == null: return
	discovery_minimap.sync(current_level, player_anchor.position if player_anchor != null else Vector3.ZERO, float((character_progression.get("derived_stats", {}) as Dictionary).get("light_radius", 0.0)), current_wall_layout, entities)

func _bot_entities_presentation_debug() -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		var node_pos := node.position if node != null else Vector3.ZERO
		var reaction = rec.get("reaction", null)
		var controller = rec.get("controller", null)
		out.append({
			"id": str(id), "type": str(rec.get("type", "")), "monster_def_id": str(rec.get("monster_def_id", "")),
			"character_id": str(rec.get("character_id", "")), "visual_model": _visual_model_name(rec, node),
			"position": {"x": node_pos.x, "z": node_pos.z},
			"visual_scale": float(rec.get("visual_scale", 1.0)),
			"is_boss": bool(rec.get("is_boss", false)), "boss_template_id": str(rec.get("boss_template_id", "")),
			"boss_phase": rec.get("boss_phase", {}), "boss_telegraph_active": bool(rec.get("boss_telegraph_active", false)),
			"telegraph_tint": str(rec.get("telegraph_tint", "")), "has_boss_telegraph_marker": bool(rec.get("has_boss_telegraph_marker", false)),
			"telegraph_radius": float(rec.get("telegraph_radius", 0.0)), "telegraph_marker_color": str(rec.get("telegraph_marker_color", "")), "telegraph_marker_shape": str(rec.get("telegraph_marker_shape", "")),
			"base_tint": str(rec.get("base_tint", "")), "has_bow_marker": bool(rec.get("has_bow_marker", false)), "effect_ids": rec.get("effect_ids", []),
			"monster_pack_id": str(rec.get("monster_pack_id", "")), "monster_pack_leader": bool(rec.get("monster_pack_leader", false)),
			"interactable_def_id": str(rec.get("interactable_def_id", "")), "elite_objective": bool(rec.get("elite_objective", false)),
			"quest_reward": bool(rec.get("quest_reward", false)), "has_objective_marker": ChestPresentationScript.has_objective_marker(node), "has_quest_marker": ChestPresentationScript.has_quest_marker(node),
			"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(node), "has_sanctuary_effect": PlayerStatusEffectMarkers.has_sanctuary_effect(node),
			"has_burning_effect": PlayerStatusEffectMarkers.has_burning_effect(node), "has_elite_command_effect": PlayerStatusEffectMarkers.has_elite_command_effect(node),
			"has_pinning_root_effect": PlayerStatusEffectMarkers.has_pinning_root_effect(node), "has_stun_effect": PlayerStatusEffectMarkers.has_stun_effect(node),
			"has_rogue_mark_effect": PlayerStatusEffectMarkers.has_rogue_mark_effect(node),
			"has_elite_command_radius_preview": PlayerStatusEffectMarkers.has_elite_command_radius_preview(node), "elite_command_radius_preview": PlayerStatusEffectMarkers.elite_command_radius_preview_value(node),
			"holy_shield_target_pulses": PlayerStatusEffectMarkers.active_holy_shield_target_pulse_count(node), "hp": int(rec.get("hp", 1)),
			"reaction": reaction.get_debug_state() if reaction != null else {}, "animation": controller.get_debug_state() if controller != null else {},
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
				"damage_type": pop.combat_damage_type,
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
	for label_id in _loot_label_entity_ids():
		var id := str(label_id)
		var rec: Dictionary = entities.get(id, {})
		var label := _loot_label_node(id)
		out.append({
			"id": id,
			"item_def_id": str(rec.get("item_def_id", "")),
			"interactable_def_id": str(rec.get("interactable_def_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"text": label.text if label != null else "",
			"visible": label.visible if label != null else false,
			"color": label.modulate.to_html(false) if label != null else "",
			"font_size": label.font_size if label != null else 0,
		})
	return out

func bot_dispatch_action(intent_type: String, payload: Dictionary) -> void:
	BotFacade.dispatch_action(self, intent_type, payload)

func bot_click_entity_id(target_id: String) -> void:
	BotFacade.click_entity_id(self, target_id, false)

func bot_click_entity_buffered_id(target_id: String) -> void:
	BotFacade.click_entity_id(self, target_id, true)

func bot_dispatch_inventory_intent(intent_type: String, payload: Dictionary) -> void:
	BotFacade.dispatch_inventory_intent(self, intent_type, payload)

func bot_click_shop_buy_offer(offer_id: String = "", offer_kind: String = "", offer_index: int = 0) -> void:
	BotFacade.click_shop_buy_offer(self, offer_id, offer_kind, offer_index)
func bot_click_shop_sell_item(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	BotFacade.click_shop_sell_item(self, item_def_id, rolled, bag_index)
func bot_click_shop_reroll() -> void:
	BotFacade.click_shop_reroll(self)
func bot_drag_bag_to_stash(item_def_id: String = "", rolled: Variant = null, bag_index: int = 0) -> void:
	BotFacade.drag_bag_to_stash(self, item_def_id, rolled, bag_index)
func bot_drag_stash_to_bag(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	BotFacade.drag_stash_to_bag(self, stash_item_id, item_def_id, rolled, stash_index)
func bot_click_stash_deposit_gold(amount: int = 1) -> void:
	BotFacade.click_stash_deposit_gold(self, amount)
func bot_click_stash_withdraw_gold(amount: int = 1) -> void:
	BotFacade.click_stash_withdraw_gold(self, amount)
func bot_click_bishop_respec() -> void:
	BotFacade.click_bishop_respec(self)
func bot_click_blacksmith_upgrade(stash_item_id: String = "", item_def_id: String = "", stash_index: int = 0) -> void:
	BotFacade.click_blacksmith_upgrade(self, stash_item_id, item_def_id, stash_index)
func bot_click_mercenary_stance(stance: String = "assist") -> void:
	BotFacade.click_mercenary_stance(self, stance)
func bot_set_stash_search(text: String) -> void:
	BotFacade.set_stash_search(self, text)
func bot_select_stash_sort(mode: String) -> void:
	BotFacade.select_stash_sort(self, mode)
func bot_set_multiplayer_search(text: String) -> void:
	if multiplayer_panel != null: multiplayer_panel.bot_set_search(text)
func bot_select_multiplayer_sort(mode: String) -> void:
	if multiplayer_panel != null: multiplayer_panel.bot_select_sort(mode)
func bot_set_market_publish_price(price_gold: int) -> void:
	BotFacade.set_market_publish_price(self, price_gold)
func bot_click_market_publish_item(stash_item_id: String = "", item_def_id: String = "", rolled: Variant = null, stash_index: int = 0) -> void:
	BotFacade.click_market_publish_item(self, stash_item_id, item_def_id, rolled, stash_index)
func bot_click_market_purchase_listing(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	BotFacade.click_market_purchase_listing(self, listing_id, item_def_id, price_gold, listing_index)
func bot_click_market_view_offers(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	BotFacade.click_market_view_offers(self, listing_id, item_def_id, price_gold, listing_index)
func bot_click_market_cancel_listing(listing_id: String = "", item_def_id: String = "", price_gold: int = -1, listing_index: int = 0) -> void:
	BotFacade.click_market_cancel_listing(self, listing_id, item_def_id, price_gold, listing_index)
func bot_click_market_offer_action(action: String, offer_id: String = "", offer_index: int = 0) -> void:
	BotFacade.click_market_offer_action(self, action, offer_id, offer_index)
func bot_assign_consumable_hotbar(slot_index: int, item_instance_id: String) -> void:
	BotFacade.assign_consumable_hotbar(self, slot_index, item_instance_id)
func bot_use_consumable_hotbar(slot_index: int) -> void:
	BotFacade.use_consumable_hotbar(self, slot_index)
func bot_click_stat_button(stat: String) -> void:
	BotFacade.click_stat_button(self, stat)
func bot_click_skill_button(skill_id: String = "") -> void:
	BotFacade.click_skill_button(self, skill_id)
func bot_use_skill_bar(skill_id: String = "", target_id: String = "", force_direct: bool = false) -> void:
	BotFacade.use_skill_bar(self, skill_id, target_id, force_direct)
func bot_cast_skill_direction(skill_id: String = "", direction: Dictionary = {}) -> void:
	BotFacade.cast_skill_direction(self, skill_id, direction)
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
	if character_panel != null: character_panel.set_name_text(name)
func bot_select_character(index: int) -> void:
	if character_panel != null: character_panel.start_character_at_index(index)
func bot_select_character_class(class_id: String) -> void:
	if character_panel != null: character_panel.select_class(class_id)
func bot_select_window_size(size: String) -> void:
	_on_window_size_selected(size)
func bot_set_floating_combat_text(enabled: bool) -> void:
	_on_floating_combat_text_toggled(enabled)
func bot_select_create_game_type(session_type: String) -> void:
	_on_create_game_session_type_selected(session_type)
func bot_set_map_opacity(value: float) -> void:
	_on_map_opacity_changed(value)
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
	if inventory_panel == null: return
	inventory_panel.ensure_display_visible()
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	var to_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO: return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag"]))
func _bot_shadow_inventory_unequip() -> void:
	if inventory_panel == null: return
	inventory_panel.ensure_display_visible()
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_bag_area_screen_center()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO: return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "bag"]))
func _bot_shadow_inventory_drop(item_instance_id: String) -> void:
	if inventory_panel == null: return
	inventory_panel.ensure_display_visible()
	_raise_gameplay_windows()
	var from_pos: Vector2 = inventory_panel.get_bag_item_screen_center(item_instance_id)
	if from_pos == Vector2.ZERO:
		from_pos = inventory_panel.get_weapon_slot_screen_center()
	var to_pos: Vector2 = inventory_panel.get_drop_outside_screen_point()
	if from_pos == Vector2.ZERO or to_pos == Vector2.ZERO: return
	input_shadow.show_drag(from_pos, to_pos, PackedStringArray(["drag", "drop"]))

# --- debug ------------------------------------------------------------------
func _update_debug() -> void:
	if _debug_label == null: return
	_sync_status_text_visibility()
	if not _debug_label.visible: return
	var ws_state := "?"
	if client != null:
		match client.ready_state():
			WebSocketPeer.STATE_CONNECTING: ws_state = "connecting"
			WebSocketPeer.STATE_OPEN: ws_state = "open"
			WebSocketPeer.STATE_CLOSING: ws_state = "closing"
			WebSocketPeer.STATE_CLOSED: ws_state = "closed"
	var fps := int(round(Engine.get_frames_per_second()))
	_debug_label.text = PerformanceStatusFormatterScript.format_status(fps, _last_ping_ms, ws_state, last_server_tick, current_level, last_performance_status)
func _sync_status_text_visibility() -> void:
	if _debug_label == null: return
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
