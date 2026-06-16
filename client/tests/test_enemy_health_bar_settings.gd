# Unit tests for enemy health bar settings and visibility.
# Run via: godot --headless --path client --script res://tests/test_enemy_health_bar_settings.gd
extends SceneTree

const MainScript := preload("res://scripts/main.gd")
const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const SettingsPanelScript := preload("res://scripts/settings_panel.gd")
const TextCatalogScript := preload("res://scripts/text_catalog.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	_test_client_settings_monster_health_bar_mode_from_data()
	_test_client_settings_monster_health_bar_mode_save_shape()
	_test_settings_panel_enemy_health_bar_mode_sync()
	_test_monster_health_bars_follow_settings_and_targeting()

	print("[gdtest] PASS: test_enemy_health_bar_settings (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_client_settings_monster_health_bar_mode_from_data() -> void:
	_assert_eq("settings monster bars default contextual", ClientSettingsScript.monster_health_bar_mode_from_data({}), "contextual")
	_assert_eq("settings monster bars parses always", ClientSettingsScript.monster_health_bar_mode_from_data({"monster_health_bar_mode": "always"}), "always")
	_assert_eq("settings monster bars parses contextual", ClientSettingsScript.monster_health_bar_mode_from_data({"monster_health_bar_mode": "contextual"}), "contextual")
	_assert_eq("settings monster bars legacy default alias", ClientSettingsScript.monster_health_bar_mode_from_data({"monster_health_bar_mode": "default"}), "contextual")
	_assert_eq("settings monster bars rejects unknown", ClientSettingsScript.monster_health_bar_mode_from_data({"monster_health_bar_mode": "sometimes"}), "contextual")


func _test_client_settings_monster_health_bar_mode_save_shape() -> void:
	var path := "user://test_enemy_health_bar_settings.json"
	var absolute_path := ProjectSettings.globalize_path(path)
	if FileAccess.file_exists(path):
		DirAccess.remove_absolute(absolute_path)
	var settings := ClientSettingsScript.new(path)
	settings.set_monster_health_bar_mode("always", false)
	settings.save()
	var parsed = JSON.parse_string(FileAccess.get_file_as_string(path))
	_assert_eq("settings save includes monster health bar mode", str((parsed as Dictionary).get("monster_health_bar_mode", "")), "always")
	var reloaded := ClientSettingsScript.new(path)
	reloaded.load()
	_assert_eq("settings reload restores monster health bar mode", reloaded.monster_health_bar_mode, "always")
	DirAccess.remove_absolute(absolute_path)


func _test_settings_panel_enemy_health_bar_mode_sync() -> void:
	TextCatalogScript.reset_for_tests()
	TextCatalogScript.set_locale("es")
	var panel = SettingsPanelScript.new()
	get_root().add_child(panel)
	panel._build()
	panel.show_settings("1920x1080", true, true, "solo", "es", "always")
	_assert_true("always enemy health bars selected", (panel._monster_health_bar_buttons["always"] as Button).disabled)
	_assert_true("contextual enemy health bars available", not (panel._monster_health_bar_buttons["contextual"] as Button).disabled)
	_assert_eq("enemy health bar label text", panel._monster_health_bar_label.text, "Barras de vida enemigas")
	panel.set_monster_health_bar_mode("contextual")
	_assert_true("contextual enemy health bars selected", (panel._monster_health_bar_buttons["contextual"] as Button).disabled)
	TextCatalogScript.set_locale("en")
	panel.free()


func _test_monster_health_bars_follow_settings_and_targeting() -> void:
	var main = MainScript.new()
	main.client_settings = ClientSettingsScript.new()
	var bar := Control.new()
	main.monster_health_bars["2001"] = bar

	main._refresh_monster_health_bar_visibility()
	_assert_true("default enemy health bar hidden without hover or target", not bar.visible)

	main.pending_action_targets["cmsg-action"] = {"target_id": "2001"}
	main._refresh_monster_health_bar_visibility()
	_assert_true("action target enemy health bar visible", bar.visible)

	main.pending_action_targets.clear()
	main.pending_skill_casts["cmsg-skill"] = {"skill_id": "magic_bolt", "target_id": "2001"}
	main._refresh_monster_health_bar_visibility()
	_assert_true("skill target enemy health bar visible", bar.visible)

	main.pending_skill_casts.clear()
	main.client_settings.set_monster_health_bar_mode("always", false)
	main._refresh_monster_health_bar_visibility()
	_assert_true("always mode enemy health bar visible", bar.visible)

	bar.free()
	main.free()


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("%s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, condition: bool) -> void:
	if condition:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("%s: condition is false" % label)
