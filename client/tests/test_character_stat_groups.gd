# Unit test for CharacterStatGroups. Run via: godot --headless --path client --script res://tests/test_character_stat_groups.gd
extends SceneTree

const CharacterStatGroupsScript := preload("res://scripts/character_stat_groups.gd")
const CharacterStatsPanelScript := preload("res://scripts/character_stats_panel.gd")

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
	var derived_keys: Array = CharacterStatsPanelScript.DERIVED_LABELS.keys()
	var ordered: Array = CharacterStatGroupsScript.ordered_keys()
	_check("every derived key is grouped", derived_keys.all(func(k): return k in ordered))
	_check("no unknown grouped keys", ordered.all(func(k): return k in derived_keys))
	_check("each key appears exactly once", ordered.size() == derived_keys.size())
	var seen := {}
	for k in ordered:
		seen[k] = true
	_check("no duplicate keys", seen.size() == ordered.size())
	_check("group ids stable", str(CharacterStatGroupsScript.group_ids()) == str(["offense", "defense", "vitals", "utility"]))
	_check("damage is offense", CharacterStatGroupsScript.group_for("damage_min") == "offense")
	_check("armor is defense", CharacterStatGroupsScript.group_for("armor") == "defense")
	_check("unknown key has no group", CharacterStatGroupsScript.group_for("nope") == "")
	for group in CharacterStatGroupsScript.GROUPS:
		_check("title resolves for %s" % group["id"], CharacterStatGroupsScript.group_title(group) == str(group["title_fallback"]))
	print("[gdtest] PASS: test_character_stat_groups (%d passed, %d failed)" % [_pass, _fail])
	quit(1 if _fail > 0 else 0)
