# Unit test for CharacterStatsHeader. Run via: godot --headless --path client --script res://tests/test_character_stats_header.gd
extends SceneTree

const HeaderScript := preload("res://scripts/character_stats_header.gd")
const CharacterPanelStyles := preload("res://scripts/character_panel_styles.gd")

var _pass := 0
var _fail := 0


func _initialize() -> void:
	call_deferred("_run")


func _check(label: String, ok: bool) -> void:
	if ok:
		_pass += 1
	else:
		_fail += 1
		printerr("[gdtest] FAIL: %s" % label)


func _run() -> void:
	_check("xp fraction mid", absf(HeaderScript.xp_fraction({"experience": 30, "experience_to_next_level": 70}) - 0.3) < 0.0001)
	_check("xp fraction zero xp", HeaderScript.xp_fraction({"experience": 0, "experience_to_next_level": 100}) == 0.0)
	_check("xp fraction max level", HeaderScript.xp_fraction({"experience": 500, "experience_to_next_level": null}) == 1.0)
	_check("xp fraction missing field is max level", HeaderScript.xp_fraction({"experience": 5}) == 1.0)
	_check("xp fraction empty total", HeaderScript.xp_fraction({"experience": 0, "experience_to_next_level": 0}) == 0.0)
	_check("xp text pending", HeaderScript.xp_text({"experience": 30, "experience_to_next_level": 70}) == "XP 30 (+70)")
	_check("xp text max", HeaderScript.xp_text({"experience": 30, "experience_to_next_level": null}).contains("Max level"))
	_check("empty class uses neutral accent", HeaderScript.class_accent("") == CharacterPanelStyles.neutral_accent())
	var seen := {}
	for class_id in ["barbarian", "sorcerer", "paladin", "rogue", "ranger"]:
		seen[HeaderScript.class_accent(class_id).to_html(false)] = true
	_check("each class has a distinct accent", seen.size() == 5)

	var header = HeaderScript.new()
	root.add_child(header)
	await process_frame
	header.configure("Ayla", {"character_class": "ranger", "level": 7, "experience": 25, "experience_to_next_level": 75, "unspent_stat_points": 3})
	var state: Dictionary = header.get_header_state()
	_check("header class id", state.get("class_id") == "ranger")
	_check("header class name", state.get("class_name") == "Ranger")
	_check("header level", int(state.get("level", 0)) == 7)
	_check("header xp fraction", absf(float(state.get("xp_fraction", -1.0)) - 0.25) < 0.0001)
	_check("points badge visible with points", bool(state.get("points_visible", false)))
	_check("header accent equals class accent", state.get("accent") == HeaderScript.class_accent("ranger").to_html(false))
	header.configure("Ayla", {"character_class": "ranger", "level": 7, "experience": 25, "experience_to_next_level": 75, "unspent_stat_points": 0})
	_check("points badge hidden at zero points", not bool(header.get_header_state().get("points_visible", true)))
	header.configure("Hero", {})
	state = header.get_header_state()
	_check("empty class has no class name", state.get("class_name") == "")
	_check("empty class neutral accent", state.get("accent") == CharacterPanelStyles.neutral_accent().to_html(false))
	header.configure("Hero", {"character_class": "mystery_class", "level": 1})
	_check("unknown class does not crash and keeps a capitalized name", header.get_header_state().get("class_name") == "Mystery Class")
	header.free()
	print("[gdtest] PASS: test_character_stats_header (%d passed, %d failed)" % [_pass, _fail])
	quit(1 if _fail > 0 else 0)
