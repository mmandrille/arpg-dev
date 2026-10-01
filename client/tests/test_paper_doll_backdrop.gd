# Unit test for PaperDollBackdrop. Run via: godot --headless --path client --script res://tests/test_paper_doll_backdrop.gd
extends SceneTree

const BackdropScript := preload("res://scripts/paper_doll_backdrop.gd")
const LayoutScript := preload("res://scripts/paper_doll_layout.gd")
const CharacterStatsHeaderScript := preload("res://scripts/character_stats_header.gd")

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
	var backdrop = BackdropScript.new()
	root.add_child(backdrop)
	await process_frame
	_check("keeps the paper-doll node name", backdrop.name == "character_paper_doll")
	_check("ignores mouse so slot buttons stay interactive", backdrop.mouse_filter == Control.MOUSE_FILTER_IGNORE)
	var neutral: String = backdrop.get_backdrop_state().get("accent", "")
	_check("no class is neutral with no emblem", not bool(backdrop.get_backdrop_state().get("emblem_visible", true)))
	var accents := {}
	for class_id in ["barbarian", "sorcerer", "paladin", "rogue", "ranger"]:
		backdrop.configure(class_id)
		var state: Dictionary = backdrop.get_backdrop_state()
		accents[state.get("accent", "")] = true
		_check("%s shows emblem" % class_id, bool(state.get("emblem_visible", false)))
		_check("%s accent matches the stats header accent" % class_id, state.get("accent") == CharacterStatsHeaderScript.class_accent(class_id).to_html(false))
	_check("each class tints distinctly", accents.size() == 5)
	_check("tinted differs from neutral", not accents.has(neutral))
	backdrop.configure("not_a_class")
	_check("unknown class does not crash", backdrop.get_backdrop_state().get("class_id") == "not_a_class")
	var slots := {
		"head": Rect2(122, 10, 96, 58),
		"main_hand": Rect2(20, 116, 96, 58),
		"chest": Rect2(122, 112, 96, 58),
	}
	backdrop.configure("rogue", slots)
	_check("connector count follows slots", backdrop.connector_count() == 3)
	await process_frame
	backdrop.configure("rogue", {})
	_check("slots can be cleared", backdrop.connector_count() == 0)
	backdrop.free()
	var rects: Dictionary = LayoutScript.slot_rects(Vector2(96, 58))
	_check("layout covers the ten equipment slots", rects.size() == 10)
	var overlaps := 0
	var ids: Array = rects.keys()
	for i in range(ids.size()):
		for j in range(i + 1, ids.size()):
			if (rects[ids[i]] as Rect2).intersects(rects[ids[j]]):
				overlaps += 1
	_check("no two paper-doll slots overlap", overlaps == 0)
	_check("slots fit the 340x360 paper area", rects.values().all(func(r): return Rect2(0, 0, 340, 360).encloses(r)))
	print("[gdtest] PASS: test_paper_doll_backdrop (%d passed, %d failed)" % [_pass, _fail])
	quit(1 if _fail > 0 else 0)
