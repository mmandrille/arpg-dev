## Fixed HUD fixture: player vitals, hotbar cluster, minimap and boss bar over a neutral backdrop.
## Dedicated script so the grandfathered visual_capture.gd does not grow (v517).
extends SceneTree

const PlayerHealthBarScript := preload("res://scripts/player_health_bar.gd")
const ConsumableBarScript := preload("res://scripts/consumable_bar.gd")
const SkillBarScript := preload("res://scripts/skill_bar.gd")
const CharacterBarScript := preload("res://scripts/character_bar.gd")
const DiscoveryMinimapScript := preload("res://scripts/discovery_minimap.gd")
const BossHealthBarScript := preload("res://scripts/boss_health_bar.gd")
const WARMUP_FRAMES := 8

# state -> [hp, max_hp, mana, max_mana, boss_hp, boss_max, populated_hotbar, reward]
const STATES := {
	"full": [120, 120, 60, 60, 1000, 1000, true, false],
	"half": [60, 120, 30, 60, 500, 1000, true, false],
	"low": [18, 120, 0, 60, 100, 1000, true, false],
	"empty": [120, 120, 60, 60, 0, 1000, false, true],
}

var _output := ""
var _state := "full"
var _width := 1280
var _height := 720


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var cfg: Array = STATES.get(_state, STATES["full"])
	_add_backdrop()
	await _add_hud(cfg)
	for _i in range(WARMUP_FRAMES):
		await process_frame
	var image := get_root().get_texture().get_image()
	if image == null or image.save_png(_output) != OK:
		push_error("hud capture failed: %s" % _output)
		quit(1)
		return
	print("[hud-capture] screenshot: %s state=%s %dx%d" % [_output, _state, _width, _height])
	quit(0)


func _add_backdrop() -> void:
	var back := ColorRect.new()
	back.color = Color("#262a27")
	back.set_anchors_preset(Control.PRESET_FULL_RECT)
	get_root().add_child(back)
	var band := ColorRect.new()
	band.color = Color("#4a463a")
	band.set_anchors_preset(Control.PRESET_FULL_RECT)
	band.anchor_top = 0.55
	band.anchor_bottom = 0.8
	get_root().add_child(band)


func _add_hud(cfg: Array) -> void:
	var vitals = PlayerHealthBarScript.new()
	var hotbar = ConsumableBarScript.new()
	var skill = SkillBarScript.new()
	var character = CharacterBarScript.new()
	var minimap = DiscoveryMinimapScript.new()
	var boss = BossHealthBarScript.new()
	for node in [vitals, hotbar, skill, character, minimap, boss]:
		get_root().add_child(node)
	await process_frame  # _ready (and _build) must run before state is applied
	vitals.set_identity("Astra", 12)
	vitals.update_hp(int(cfg[0]), int(cfg[1]))
	vitals.update_mana(int(cfg[2]), int(cfg[3]))

	if bool(cfg[6]):
		var inv := [
			{"item_instance_id": "101", "item_def_id": "red_potion", "quantity": 4},
			{"item_instance_id": "102", "item_def_id": "blue_potion", "quantity": 2},
		]
		hotbar.set_inventory_state(inv)
		hotbar.set_hotbar_state(10, [
			{"slot_index": 0, "item_instance_id": "101"},
			{"slot_index": 1, "item_instance_id": "102"},
		])
	hotbar.set_character_progression({"level": 12, "experience": 340, "experience_to_next_level": 400})

	skill.set_player_mana(int(cfg[2]), int(cfg[3]))

	minimap.set_display_mode("compact")
	var cells: Array = []
	for x in range(-6, 7):
		for y in range(-5, 6):
			cells.append({"x": x, "y": y})
	var walls: Array = [
		{"x": 0.0, "y": -6.0, "w": 14.0, "h": 1.0},
		{"x": 0.0, "y": 6.0, "w": 14.0, "h": 1.0},
		{"x": -7.0, "y": 0.0, "w": 1.0, "h": 12.0},
		{"x": 7.0, "y": 0.0, "w": 1.0, "h": 12.0},
	]
	minimap.set_state({"explored_cells": cells, "walls": walls, "player_x": 0.0, "player_y": 0.0})

	if bool(cfg[7]):
		boss.show_reward_status("boss_template", "Ashen Warden", "Reward ready", "Return to the reward chest")
	if int(cfg[4]) > 0:
		boss.show_boss("boss_1", "boss_template", "Ashen Warden", int(cfg[4]), int(cfg[5]))
		boss.set_phase_state({"phase_kind": "telegraph", "pattern_id": "charged_melee", "phase_index": 0, "duration_ticks": 30, "remaining_ticks": 18})


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i + 1 < args.size():
		var key := str(args[i])
		var value := str(args[i + 1])
		match key:
			"--output": _output = value
			"--hud-state": _state = value
			"--width": _width = int(value)
			"--height": _height = int(value)
		i += 2
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/hud.png")
