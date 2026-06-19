extends SceneTree

const ViewerScene := preload("res://scenes/model_viewer.tscn")
const ViewerScript := preload("res://scripts/model_viewer.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	await _test_catalog_resolves_expected_assets()
	await _test_viewer_loads_asset("character_paladin_v0")
	await _test_viewer_loads_asset("monster_tiny_flyer_v0")
	print("[gdtest] PASS: test_model_viewer (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _test_catalog_resolves_expected_assets() -> void:
	var paladin: Dictionary = ViewerScript.resolve("character_paladin_v0")
	_assert(str(paladin.get("type", "")) == "character", "paladin resolves as character")
	_assert(str(paladin.get("runtime_path", "")) == "client/assets/characters/paladin/paladin.glb", "paladin runtime path")
	_assert((paladin.get("used_by", []) as Array).has("paladin"), "paladin used_by label")
	var bat: Dictionary = ViewerScript.resolve("monster_tiny_flyer_v0")
	_assert(str(bat.get("type", "")) == "monster", "bat resolves as monster")
	_assert(str(bat.get("scene", "")) == "monster_tiny_flyer", "bat scene label")
	_assert((bat.get("used_by", []) as Array).has("dungeon_bat"), "bat used_by label")
	var sword: Dictionary = ViewerScript.resolve("weapon_rusty_sword_v0")
	_assert(sword.is_empty(), "equipment assets are not previewable in v284")


func _test_viewer_loads_asset(asset_id: String) -> void:
	ProjectSettings.set_setting("arpg/model_viewer/asset_id", asset_id)
	var viewer = ViewerScene.instantiate()
	viewer.auto_cycle = false
	root.add_child(viewer)
	await process_frame
	await process_frame
	_assert(viewer.current_instance != null, "%s creates current instance" % asset_id)
	_assert(viewer.current_animation_player != null, "%s has AnimationPlayer" % asset_id)
	_assert(not viewer.current_clips.is_empty(), "%s exposes preview clips" % asset_id)
	_assert(viewer.current_clips.has("idle"), "%s exposes idle clip" % asset_id)
	_assert(viewer.current_clips.has("walk"), "%s exposes walk clip" % asset_id)
	viewer.free()
	await process_frame


func _assert(condition: bool, message: String) -> void:
	if condition:
		_pass_count += 1
	else:
		_fail_count += 1
		printerr("[gdtest] FAIL: %s" % message)
