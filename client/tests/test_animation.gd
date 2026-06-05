extends SceneTree
# Headless tests for the v3 animation layer (spec §10). Server-independent.
# Run: godot --headless --path client --script res://tests/test_animation.gd
const ControllerScript := preload("res://scripts/animation_controller.gd")


var _failed: bool = false


func _initialize() -> void:
	_test_controller_locomotion()
	if _failed: quit(1); return
	_test_controller_one_shot_returns()
	if _failed: quit(1); return
	_test_controller_terminal_latches()
	if _failed: quit(1); return
	_test_controller_hit_ignored_after_death()
	if _failed: quit(1); return
	_test_controller_locomotion_change_during_one_shot()
	if _failed: quit(1); return
	print("[gdtest] PASS: animation controller")
	quit(0)


func _make_player(clips: Array) -> AnimationPlayer:
	var ap := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	for c in clips:
		var a := Animation.new()
		a.length = (0.5 if c != "idle" and c != "walk" else 1.0)
		lib.add_animation(c, a)
	ap.add_animation_library("", lib)
	get_root().add_child(ap)  # animations need the player in-tree
	return ap


func _test_controller_locomotion() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(false)
	_assert(c.current_clip() == "idle", "idle when not moving, got %s" % c.current_clip())
	c.set_locomotion(true)
	_assert(c.current_clip() == "walk", "walk when moving, got %s" % c.current_clip())
	ap.queue_free()


func _test_controller_one_shot_returns() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(true)
	c.play_one_shot("attack")
	_assert(c.current_clip() == "attack", "attack active, got %s" % c.current_clip())
	# Simulate the clip finishing.
	ap.emit_signal("animation_finished", "attack")
	_assert(c.current_clip() == "walk", "returns to locomotion (walk) after one-shot, got %s" % c.current_clip())
	ap.queue_free()


func _test_controller_terminal_latches() -> void:
	var ap := _make_player(["idle", "hit", "death"])
	var c = ControllerScript.new(ap)
	c.enter_terminal("death")
	_assert(c.current_clip() == "death", "death active, got %s" % c.current_clip())
	c.play_one_shot("hit")        # ignored
	c.set_locomotion(true)        # ignored
	_assert(c.current_clip() == "death", "terminal latched, got %s" % c.current_clip())
	_assert(c.get_debug_state()["terminal"] == true, "terminal flag set")
	ap.queue_free()


func _test_controller_hit_ignored_after_death() -> void:
	var ap := _make_player(["idle", "hit", "death"])
	var c = ControllerScript.new(ap)
	c.enter_terminal("death")
	c.play_one_shot("hit")
	_assert(c.current_clip() == "death", "hit ignored after terminal death, got %s" % c.current_clip())
	ap.queue_free()


func _test_controller_locomotion_change_during_one_shot() -> void:
	var ap := _make_player(["idle", "walk", "attack"])
	var c = ControllerScript.new(ap)
	c.set_locomotion(false)
	c.play_one_shot("attack")
	c.set_locomotion(true)  # must NOT switch to walk while one-shot is active
	_assert(c.current_clip() == "attack", "locomotion change ignored during one-shot, got %s" % c.current_clip())
	ap.emit_signal("animation_finished", "attack")
	_assert(c.current_clip() == "walk", "fallback honors latched _moving after one-shot, got %s" % c.current_clip())
	ap.queue_free()


func _assert(cond: bool, msg: String) -> void:
	if not cond:
		printerr("[gdtest] FAIL: ", msg)
		_failed = true
		quit(1)
