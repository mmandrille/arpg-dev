extends Control
class_name SettingsPanel

const TextCatalogScript := preload("res://scripts/text_catalog.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")

signal back_requested
signal size_selected(label: String)
signal floating_combat_text_toggled(enabled: bool)
signal status_text_toggled(enabled: bool)
signal create_game_session_type_selected(session_type: String)
signal language_selected(language: String)
signal monster_health_bar_mode_selected(mode: String)
signal camera_mode_selected(mode: String)
signal graphics_quality_selected(mode: String)
signal master_volume_changed(value: float)
signal music_volume_changed(value: float)
signal sfx_volume_changed(value: float)
signal map_opacity_changed(value: float)

var _buttons: Dictionary = {}
var _session_type_buttons: Dictionary = {}
var _language_buttons: Dictionary = {}
var _monster_health_bar_buttons: Dictionary = {}
var _camera_mode_buttons: Dictionary = {}
var _graphics_quality_buttons: Dictionary = {}
var _selected_label: String = ""
var _selected_session_type: String = "coop"
var _selected_language: String = "en"
var _selected_monster_health_bar_mode: String = ClientSettingsScript.DEFAULT_MONSTER_HEALTH_BAR_MODE
var _selected_camera_mode: String = ClientSettingsScript.DEFAULT_CAMERA_MODE
var _selected_graphics_quality: String = ClientSettingsScript.DEFAULT_GRAPHICS_QUALITY
var _floating_toggle: CheckButton
var _status_text_toggle: CheckButton
var _title: Label
var _session_type_label: Label
var _language_label: Label
var _monster_health_bar_label: Label
var _camera_mode_label: Label
var _graphics_quality_label: Label
var _master_volume_label: Label
var _music_volume_label: Label
var _sfx_volume_label: Label
var _map_opacity_label: Label
var _master_volume_slider: HSlider
var _music_volume_slider: HSlider
var _sfx_volume_slider: HSlider
var _map_opacity_slider: HSlider
var _back_button: Button
var _panel: PanelContainer
var _scroll: ScrollContainer


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_settings(selected_label: String, floating_combat_text_enabled: bool = true, status_text_enabled: bool = true, create_game_session_type: String = "coop", language: String = "en", monster_health_bar_mode: String = ClientSettingsScript.DEFAULT_MONSTER_HEALTH_BAR_MODE, master_volume: float = ClientSettingsScript.DEFAULT_MASTER_VOLUME, music_volume: float = ClientSettingsScript.DEFAULT_MUSIC_VOLUME, sfx_volume: float = ClientSettingsScript.DEFAULT_SFX_VOLUME, map_opacity: float = ClientSettingsScript.DEFAULT_MAP_OPACITY, camera_mode: String = ClientSettingsScript.DEFAULT_CAMERA_MODE, graphics_quality: String = ClientSettingsScript.DEFAULT_GRAPHICS_QUALITY) -> void:
	_sync_viewport_size()
	visible = true
	set_selected_size_label(selected_label)
	set_floating_combat_text_enabled(floating_combat_text_enabled)
	set_status_text_enabled(status_text_enabled)
	set_create_game_session_type(create_game_session_type)
	set_language(language)
	set_monster_health_bar_mode(monster_health_bar_mode)
	set_audio_volumes(master_volume, music_volume, sfx_volume)
	set_map_opacity(map_opacity)
	set_camera_mode(camera_mode)
	set_graphics_quality(graphics_quality)


func hide_panel() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	if _panel == null:
		return
	var viewport_size := get_viewport_rect().size if is_inside_tree() else Vector2(1280, 720)
	var panel_width: float = minf(460.0, maxf(360.0, viewport_size.x - 32.0))
	var panel_height: float = minf(660.0, maxf(360.0, viewport_size.y - 32.0))
	_panel.offset_left = -panel_width * 0.5
	_panel.offset_right = panel_width * 0.5
	_panel.offset_top = -panel_height * 0.5
	_panel.offset_bottom = panel_height * 0.5


func set_selected_size_label(label: String) -> void:
	_selected_label = label
	for key in _buttons.keys():
		var button: Button = _buttons[key]
		button.disabled = str(key) == _selected_label


func set_floating_combat_text_enabled(enabled: bool) -> void:
	if _floating_toggle == null:
		return
	_floating_toggle.button_pressed = enabled


func set_status_text_enabled(enabled: bool) -> void:
	if _status_text_toggle == null:
		return
	_status_text_toggle.button_pressed = enabled


func set_create_game_session_type(session_type: String) -> void:
	_selected_session_type = _normalize_session_type(session_type)
	for key in _session_type_buttons.keys():
		var button: Button = _session_type_buttons[key]
		button.disabled = str(key) == _selected_session_type


func set_language(language: String) -> void:
	_selected_language = _normalize_language(language)
	for key in _language_buttons.keys():
		var button: Button = _language_buttons[key]
		button.disabled = str(key) == _selected_language
	refresh_texts()


func set_monster_health_bar_mode(mode: String) -> void:
	_selected_monster_health_bar_mode = ClientSettingsScript.normalize_monster_health_bar_mode(mode)
	for key in _monster_health_bar_buttons.keys():
		var button: Button = _monster_health_bar_buttons[key]
		button.disabled = str(key) == _selected_monster_health_bar_mode


func set_camera_mode(mode: String) -> void:
	_selected_camera_mode = ClientSettingsScript.normalize_camera_mode(mode)
	for key in _camera_mode_buttons.keys():
		var button: Button = _camera_mode_buttons[key]
		button.disabled = str(key) == _selected_camera_mode


func set_graphics_quality(quality: String) -> void:
	_selected_graphics_quality = ClientSettingsScript.normalize_graphics_quality(quality)
	for key in _graphics_quality_buttons.keys():
		var button: Button = _graphics_quality_buttons[key]
		button.disabled = str(key) == _selected_graphics_quality


func set_audio_volumes(master: float, music: float, sfx: float) -> void:
	_set_slider_value(_master_volume_slider, master)
	_set_slider_value(_music_volume_slider, music)
	_set_slider_value(_sfx_volume_slider, sfx)


func set_map_opacity(value: float) -> void:
	_set_slider_value(_map_opacity_slider, 1.0 - value)


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.03, 0.035, 0.04, 0.88)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	_panel = PanelContainer.new()
	_panel.custom_minimum_size = Vector2(360, 360)
	_panel.set_anchors_preset(Control.PRESET_CENTER)
	add_child(_panel)

	_scroll = ScrollContainer.new()
	_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	_scroll.vertical_scroll_mode = ScrollContainer.SCROLL_MODE_AUTO
	_scroll.mouse_filter = Control.MOUSE_FILTER_STOP
	_panel.add_child(_scroll)

	var box := VBoxContainer.new()
	box.add_theme_constant_override("separation", 6)
	box.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_scroll.add_child(box)

	_title = Label.new()
	_title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_title.add_theme_font_size_override("font_size", 34)
	box.add_child(_title)

	for label in ClientSettingsScript.supported_size_labels():
		var button := Button.new()
		button.text = label
		button.custom_minimum_size = Vector2(320, 36)
		button.add_theme_font_size_override("font_size", 28)
		button.pressed.connect(func() -> void:
			size_selected.emit(label)
		)
		_buttons[label] = button
		box.add_child(button)

	_session_type_label = Label.new()
	_session_type_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_session_type_label.add_theme_font_size_override("font_size", 23)
	box.add_child(_session_type_label)

	var session_type_row := HBoxContainer.new()
	session_type_row.add_theme_constant_override("separation", 8)
	box.add_child(session_type_row)
	_add_session_type_button(session_type_row, TextCatalogScript.get_text("settings.session_type.coop", "Co-op"), "coop")
	_add_session_type_button(session_type_row, TextCatalogScript.get_text("settings.session_type.solo", "Solo"), "solo")

	_floating_toggle = CheckButton.new()
	_floating_toggle.button_pressed = true
	_floating_toggle.custom_minimum_size = Vector2(320, 36)
	_floating_toggle.add_theme_font_size_override("font_size", 28)
	_floating_toggle.toggled.connect(func(enabled: bool) -> void:
		floating_combat_text_toggled.emit(enabled)
	)
	box.add_child(_floating_toggle)

	_status_text_toggle = CheckButton.new()
	_status_text_toggle.button_pressed = true
	_status_text_toggle.custom_minimum_size = Vector2(320, 36)
	_status_text_toggle.add_theme_font_size_override("font_size", 28)
	_status_text_toggle.toggled.connect(func(enabled: bool) -> void:
		status_text_toggled.emit(enabled)
	)
	box.add_child(_status_text_toggle)

	_monster_health_bar_label = Label.new()
	_monster_health_bar_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_monster_health_bar_label.add_theme_font_size_override("font_size", 23)
	box.add_child(_monster_health_bar_label)

	var monster_health_bar_row := HBoxContainer.new()
	monster_health_bar_row.add_theme_constant_override("separation", 8)
	box.add_child(monster_health_bar_row)
	_add_monster_health_bar_button(monster_health_bar_row, ClientSettingsScript.MONSTER_HEALTH_BAR_CONTEXTUAL)
	_add_monster_health_bar_button(monster_health_bar_row, ClientSettingsScript.MONSTER_HEALTH_BAR_ALWAYS)

	_camera_mode_label = Label.new()
	_camera_mode_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_camera_mode_label.add_theme_font_size_override("font_size", 23)
	box.add_child(_camera_mode_label)

	var camera_mode_row := HBoxContainer.new()
	camera_mode_row.add_theme_constant_override("separation", 8)
	box.add_child(camera_mode_row)
	_add_camera_mode_button(camera_mode_row, ClientSettingsScript.CAMERA_MODE_ISOMETRIC)
	_add_camera_mode_button(camera_mode_row, ClientSettingsScript.CAMERA_MODE_CHEST_VIEW)

	_graphics_quality_label = Label.new()
	_graphics_quality_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_graphics_quality_label.add_theme_font_size_override("font_size", 23)
	box.add_child(_graphics_quality_label)

	var graphics_quality_row := HBoxContainer.new()
	graphics_quality_row.add_theme_constant_override("separation", 8)
	box.add_child(graphics_quality_row)
	_add_graphics_quality_button(graphics_quality_row, ClientSettingsScript.GRAPHICS_QUALITY_BALANCED)
	_add_graphics_quality_button(graphics_quality_row, ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE)

	_master_volume_label = _add_slider_label(box)
	_master_volume_slider = _add_volume_slider(box, func(value: float) -> void:
		master_volume_changed.emit(value)
	)
	_music_volume_label = _add_slider_label(box)
	_music_volume_slider = _add_volume_slider(box, func(value: float) -> void:
		music_volume_changed.emit(value)
	)
	_sfx_volume_label = _add_slider_label(box)
	_sfx_volume_slider = _add_volume_slider(box, func(value: float) -> void:
		sfx_volume_changed.emit(value)
	)
	_map_opacity_label = _add_slider_label(box)
	_map_opacity_slider = _add_volume_slider(box, func(value: float) -> void:
		map_opacity_changed.emit(1.0 - value)
	)

	_language_label = Label.new()
	_language_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	_language_label.add_theme_font_size_override("font_size", 23)
	box.add_child(_language_label)

	var language_row := HBoxContainer.new()
	language_row.add_theme_constant_override("separation", 8)
	box.add_child(language_row)
	_add_language_button(language_row, "en")
	_add_language_button(language_row, "es")

	_back_button = Button.new()
	_back_button.custom_minimum_size = Vector2(320, 36)
	_back_button.add_theme_font_size_override("font_size", 28)
	_back_button.pressed.connect(back_requested.emit)
	box.add_child(_back_button)
	refresh_texts()
	_sync_viewport_size()


func _add_session_type_button(parent: Control, label: String, session_type: String) -> void:
	var button := Button.new()
	button.text = label
	button.custom_minimum_size = Vector2(154, 36)
	button.add_theme_font_size_override("font_size", 26)
	button.pressed.connect(func() -> void:
		create_game_session_type_selected.emit(session_type)
	)
	_session_type_buttons[session_type] = button
	parent.add_child(button)


func _add_language_button(parent: Control, language: String) -> void:
	var button := Button.new()
	button.custom_minimum_size = Vector2(154, 36)
	button.add_theme_font_size_override("font_size", 26)
	button.pressed.connect(func() -> void:
		language_selected.emit(language)
	)
	_language_buttons[language] = button
	parent.add_child(button)


func _add_monster_health_bar_button(parent: Control, mode: String) -> void:
	var button := Button.new()
	button.custom_minimum_size = Vector2(154, 36)
	button.add_theme_font_size_override("font_size", 26)
	button.pressed.connect(func() -> void:
		monster_health_bar_mode_selected.emit(mode)
	)
	_monster_health_bar_buttons[mode] = button
	parent.add_child(button)


func _add_camera_mode_button(parent: Control, mode: String) -> void:
	var button := Button.new()
	button.custom_minimum_size = Vector2(100, 36)
	button.add_theme_font_size_override("font_size", 22)
	button.pressed.connect(func() -> void:
		camera_mode_selected.emit(mode)
	)
	_camera_mode_buttons[mode] = button
	parent.add_child(button)


func _add_graphics_quality_button(parent: Control, quality: String) -> void:
	var button := Button.new()
	button.custom_minimum_size = Vector2(154, 36)
	button.add_theme_font_size_override("font_size", 22)
	button.pressed.connect(func() -> void:
		graphics_quality_selected.emit(quality)
	)
	_graphics_quality_buttons[quality] = button
	parent.add_child(button)


func bot_select_camera_mode(mode: String) -> void:
	camera_mode_selected.emit(mode)


func _add_slider_label(parent: Control) -> Label:
	var label := Label.new()
	label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	label.add_theme_font_size_override("font_size", 21)
	parent.add_child(label)
	return label


func _add_volume_slider(parent: Control, handler: Callable) -> HSlider:
	var slider := HSlider.new()
	slider.min_value = 0.0
	slider.max_value = 1.0
	slider.step = 0.05
	slider.custom_minimum_size = Vector2(320, 30)
	slider.mouse_filter = Control.MOUSE_FILTER_STOP
	slider.value_changed.connect(handler)
	parent.add_child(slider)
	return slider


func _set_slider_value(slider: HSlider, value: float) -> void:
	if slider == null:
		return
	slider.set_value_no_signal(clampf(value, 0.0, 1.0))


func _normalize_session_type(session_type: String) -> String:
	var normalized := session_type.strip_edges().to_lower()
	if normalized == "solo":
		return "solo"
	return "coop"


func _normalize_language(language: String) -> String:
	var normalized := language.strip_edges().to_lower()
	if normalized == "es":
		return "es"
	return "en"


func refresh_texts() -> void:
	if _title != null:
		_title.text = TextCatalogScript.get_text("settings.title", "Settings")
	if _session_type_label != null:
		_session_type_label.text = TextCatalogScript.get_text("settings.create_game_type", "Create Game Type")
	if _floating_toggle != null:
		_floating_toggle.text = TextCatalogScript.get_text("settings.floating_combat_text", "Floating combat text")
	if _status_text_toggle != null:
		_status_text_toggle.text = TextCatalogScript.get_text("settings.status_text", "Performance status")
	if _monster_health_bar_label != null:
		_monster_health_bar_label.text = TextCatalogScript.get_text("settings.enemy_health_bars", "Enemy health bars")
	if _master_volume_label != null:
		_master_volume_label.text = TextCatalogScript.get_text("settings.master_volume", "Master volume")
	if _music_volume_label != null:
		_music_volume_label.text = TextCatalogScript.get_text("settings.music_volume", "Music volume")
	if _sfx_volume_label != null:
		_sfx_volume_label.text = TextCatalogScript.get_text("settings.sfx_volume", "SFX volume")
	if _map_opacity_label != null:
		_map_opacity_label.text = TextCatalogScript.get_text("settings.map_transparency", "Map transparency")
	if _language_label != null:
		_language_label.text = TextCatalogScript.get_text("settings.language", "Language")
	if _back_button != null:
		_back_button.text = TextCatalogScript.get_text("settings.back", "Back")
	for key in _session_type_buttons.keys():
		var session_button: Button = _session_type_buttons[key]
		session_button.text = TextCatalogScript.get_text("settings.session_type.%s" % str(key), str(key))
	for key in _language_buttons.keys():
		var language_button: Button = _language_buttons[key]
		language_button.text = TextCatalogScript.get_text("settings.language.%s" % str(key), str(key))
	for key in _monster_health_bar_buttons.keys():
		var mode_button: Button = _monster_health_bar_buttons[key]
		mode_button.text = TextCatalogScript.get_text("settings.enemy_health_bars.%s" % str(key), str(key))
	if _camera_mode_label != null:
		_camera_mode_label.text = TextCatalogScript.get_text("settings.camera_mode", "Camera Mode")
	for key in _camera_mode_buttons.keys():
		var cam_button: Button = _camera_mode_buttons[key]
		cam_button.text = TextCatalogScript.get_text("settings.camera_mode.%s" % str(key), str(key))
	if _graphics_quality_label != null:
		_graphics_quality_label.text = TextCatalogScript.get_text("settings.graphics_quality", "Graphics Quality")
	for key in _graphics_quality_buttons.keys():
		var quality_button: Button = _graphics_quality_buttons[key]
		quality_button.text = ClientSettingsScript.graphics_quality_label(str(key))
