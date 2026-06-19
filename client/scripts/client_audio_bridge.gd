extends RefCounted
class_name ClientAudioBridge


static func apply_settings(controller: ClientAudioController, settings: ClientSettings) -> void:
	if controller == null or settings == null:
		return
	controller.apply_volumes(settings.master_volume, settings.music_volume, settings.sfx_volume)


static func show_settings(panel: SettingsPanel, settings: ClientSettings) -> void:
	if panel == null or settings == null:
		return
	panel.show_settings(
		ClientSettings.size_label(settings.window_size),
		settings.floating_combat_text,
		settings.status_text,
		settings.create_game_session_type,
		settings.language,
		settings.monster_health_bar_mode,
		settings.master_volume,
		settings.music_volume,
		settings.sfx_volume,
		settings.map_opacity
	)


static func sync_settings_panel(panel: SettingsPanel, settings: ClientSettings) -> void:
	if panel == null or settings == null:
		return
	panel.set_audio_volumes(settings.master_volume, settings.music_volume, settings.sfx_volume)
	panel.set_map_opacity(settings.map_opacity)


static func set_master_volume(controller: ClientAudioController, settings: ClientSettings, value: float) -> void:
	if settings == null:
		return
	settings.set_master_volume(value)
	apply_settings(controller, settings)


static func set_music_volume(controller: ClientAudioController, settings: ClientSettings, value: float) -> void:
	if settings == null:
		return
	settings.set_music_volume(value)
	apply_settings(controller, settings)


static func set_sfx_volume(controller: ClientAudioController, settings: ClientSettings, value: float) -> void:
	if settings == null:
		return
	settings.set_sfx_volume(value)
	apply_settings(controller, settings)


static func connect_volume_signals(panel: SettingsPanel, controller: ClientAudioController, settings: ClientSettings, sync_callable: Callable) -> void:
	panel.master_volume_changed.connect(func(value: float) -> void:
		set_master_volume(controller, settings, value)
	)
	panel.music_volume_changed.connect(func(value: float) -> void:
		set_music_volume(controller, settings, value)
	)
	panel.sfx_volume_changed.connect(func(value: float) -> void:
		set_sfx_volume(controller, settings, value)
	)


static func movement(controller: ClientAudioController) -> void:
	if controller != null:
		controller.play_movement()


static func attack(controller: ClientAudioController) -> void:
	if controller != null:
		controller.play_attack()


static func skill(controller: ClientAudioController, skill_id: String) -> void:
	if controller != null:
		controller.play_skill(skill_id)


static func heal(controller: ClientAudioController) -> void:
	if controller != null:
		controller.play_heal()


static func damage(controller: ClientAudioController, local_player: bool) -> void:
	if controller != null:
		controller.play_damage(local_player)


static func kill(controller: ClientAudioController, is_boss: bool) -> void:
	if controller != null:
		controller.play_kill(is_boss)


static func boss_phase(controller: ClientAudioController, ev: Dictionary = {}) -> void:
	if controller != null:
		controller.play_boss_phase(str(ev.get("pattern_id", "")), str(ev.get("phase_kind", "")))


static func stop_boss_music(controller: ClientAudioController) -> void:
	if controller != null:
		controller.stop_boss_music()


static func ambience_for_level(controller: ClientAudioController, level: int) -> void:
	if controller != null:
		controller.set_ambient_level(level)
