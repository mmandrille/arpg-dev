extends SceneTree
## Dedicated v516 capture driver: character stats panel (and paper-doll backdrop) per class.
## Invoked by skills/showme/scripts/render_focus.py with --focus character-screen.

const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")

var _output := ""
var _class_id := "paladin"
var _variant := "points"
var _width := 960
var _height := 760


func _initialize() -> void:
	_parse_args()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var backdrop := ColorRect.new()
	backdrop.color = Color("#14110d")
	backdrop.set_anchors_preset(Control.PRESET_FULL_RECT)
	get_root().add_child(backdrop)
	if _variant == "paper-doll":
		await _setup_paper_doll()
		await _finish()
		return
	var panel = CharacterStatsPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.set_hero_name("Hero")
	panel.set_allocation_enabled(true)
	panel.set_progression(_progression())
	panel.ensure_display_visible()
	await _finish()


## Real inventory panel with the active class, so the paper-doll backdrop and slot grid are the shipped ones.
func _setup_paper_doll() -> void:
	var panel = InventoryPanelScript.new()
	get_root().add_child(panel)
	await process_frame
	panel.set_character_progression({"character_class": _class_id})
	panel.set_inventory_state([], {}, 4, 20, 145)
	panel.ensure_display_visible()
	await process_frame


func _finish() -> void:
	for _i in range(8):
		await process_frame
	var image := get_root().get_texture().get_image()
	if image.save_png(_output) != OK:
		printerr("[showme] failed to save screenshot: %s" % _output)
		quit(1)
		return
	print("[showme] saved %s" % _output)
	quit(0)


func _progression() -> Dictionary:
	var points := 0 if _variant.begins_with("nopoints") else 3
	var derived := {
		"damage_min": 13.4, "damage_max": 21.0, "ranged_damage_bonus_percent": 6.0,
		"armor": 34.0, "attack_speed": 1.2, "attack_interval_ticks": 8.0,
		"hit_chance": 0.82, "crit_chance": 0.06, "crit_damage": 1.5,
		"evade_chance": 0.12, "block_percent": 20.0, "movement_speed": 4.4,
		"max_hp": 148.0, "max_mana": 62.0, "health_regen_per_second": 1.2,
		"mana_regen_per_second": 0.8, "light_radius": 6.0,
	}
	if _variant.contains("dual"):
		derived["weapon_damage_by_slot"] = {
			"main_hand": {"min": 13.4, "max": 21.0},
			"off_hand": {"min": 9.0, "max": 14.0},
		}
	return {
		"character_class": _class_id, "level": 12, "experience": 4210,
		"experience_to_next_level": 1830, "unspent_stat_points": points,
		"base_stats": {"str": 25, "dex": 17, "vit": 15, "magic": 18},
		"effective_base_stats": {"str": 42, "dex": 12, "vit": 15, "magic": 18},
		"derived_stats": derived,
	}


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i < args.size():
		match str(args[i]):
			"--output":
				i += 1
				_output = str(args[i])
			"--class-id":
				i += 1
				_class_id = str(args[i]).strip_edges()
			"--variant":
				i += 1
				_variant = str(args[i]).strip_edges()
			"--width":
				i += 1
				_width = int(args[i])
			"--height":
				i += 1
				_height = int(args[i])
		i += 1
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/character-screen.png")
