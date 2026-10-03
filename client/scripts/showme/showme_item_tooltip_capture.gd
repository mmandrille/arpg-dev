extends SceneTree
## Dedicated v532 capture driver: the shared item tooltip for one rarity, built from real tooltip content.
## Invoked by skills/showme/scripts/render_focus.py with --focus item-tooltip.

const ContentScript := preload("res://scripts/inventory_tooltip_content.gd")
const TooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")

const SETTLE_FRAMES := 8
const RARITIES := ["common", "magic", "rare", "unique", "set"]

var _output := ""
var _rarity := "rare"
var _width := 640
var _height := 520


func _initialize() -> void:
	_parse_args()
	call_deferred("_run")


func _run() -> void:
	ItemRulesLoader.ensure_loaded()
	DisplayServer.window_set_size(Vector2i(_width, _height))
	get_root().size = Vector2i(_width, _height)
	var backdrop := ColorRect.new()
	backdrop.color = Color("#14110d")
	backdrop.set_anchors_preset(Control.PRESET_FULL_RECT)
	get_root().add_child(backdrop)
	var item := {
		"item_instance_id": "tooltip_capture", "item_def_id": "long_sword", "display_name": "Captured Blade",
		"rarity": _rarity, "slot": "main_hand", "item_level": 12,
		"requirement_status": [{"stat": "str", "required": 14, "current": 12, "met": false}],
		"comparison": {"deltas": [{"stat": "armor", "delta": 2}, {"stat": "max_hp", "delta": -3}]},
	}
	var tooltip = TooltipPanelScript.new()
	tooltip.position = Vector2(24, 24)
	tooltip.setup(item, ItemRulesLoader.item_presentations, ContentScript.tooltip_lines(item, ContentScript.Context.new([])),
		ContentScript.requirement_lines(item), ContentScript.comparison_entries(item), 120, true)
	get_root().add_child(tooltip)
	for _i in range(SETTLE_FRAMES):
		await process_frame
	var image := get_root().get_texture().get_image()
	if image.save_png(_output) != OK:
		printerr("[showme] failed to save screenshot: %s" % _output)
		quit(1)
		return
	print("[showme] saved %s" % _output)
	quit(0)


func _parse_args() -> void:
	var args := OS.get_cmdline_user_args()
	var i := 0
	while i < args.size():
		match str(args[i]):
			"--output":
				i += 1
				_output = str(args[i])
			"--rarity":
				i += 1
				_rarity = str(args[i]).strip_edges()
			"--width":
				i += 1
				_width = int(args[i])
			"--height":
				i += 1
				_height = int(args[i])
		i += 1
	if not RARITIES.has(_rarity):
		_rarity = "rare"
	if _output == "":
		_output = ProjectSettings.globalize_path("res://").path_join("../.artifacts/showme/item-tooltip-%s.png" % _rarity)
