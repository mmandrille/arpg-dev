extends SceneTree
## v513 monster combat animation polish: pure selection/timing math, back-compat, and
## driver routing against the real kit rigs. Headless; no main.gd scene paths (CLAUDE.md rule 9).
## Run: godot --headless --path client --script res://tests/test_monster_anim_variants.gd

const VariantsScript := preload("res://scripts/monster_anim_variants.gd")
const DriverScript := preload("res://scripts/monster_anim_driver.gd")
const IdleScript := preload("res://scripts/monster_idle_variation.gd")
const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const ControllerScript := preload("res://scripts/animation_controller.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_pick_is_deterministic_and_spreads()
	_test_group_fallback_without_variants()
	_test_windup_speed_scale()
	_test_hit_side_and_directional_clip()
	_test_idle_interval_and_phase()
	_test_catalog_clips_exist_in_rigs_and_profile_refs_resolve()
	await _test_skeleton_windup_strike_and_hit()
	await _test_wolf_directional_hit()
	await _test_spawn_guard_and_cancel()
	await _test_terminal_suppresses_one_shots()
	await _test_idle_variation()
	_finish()


func _ok(cond: bool, msg: String) -> void:
	if cond:
		_pass_count += 1
		print("  ok   %s" % msg)
	else:
		_fail_count += 1
		push_error("FAIL: %s" % msg)
		print("  FAIL %s" % msg)


func _test_pick_is_deterministic_and_spreads() -> void:
	var group := ["attack", "attack_b", "attack_c"]
	_ok(VariantsScript.pick(group, "7", 0) == VariantsScript.pick(group, "7", 0), "pick is deterministic for the same key and counter")
	var seen := {}
	for i in range(3):
		seen[VariantsScript.pick(group, "7", i)] = true
	_ok(seen.size() == 3, "consecutive counters cycle through every variant")
	var starts := {}
	for id in range(40):
		starts[VariantsScript.pick(group, str(id), 0)] = true
	_ok(starts.size() == 3, "different entities start on different variants")
	_ok(VariantsScript.pick([], "1", 0) == "", "empty group picks nothing")


func _test_group_fallback_without_variants() -> void:
	_ok(VariantsScript.group_clips({}, "attack") == ["attack"], "missing variants fall back to the base logical clip")
	_ok(VariantsScript.hit_clip({}, "1", 0, Vector3.BACK, Vector3.RIGHT, true) == "hit", "profile without hit keys keeps the base hit")


func _test_windup_speed_scale() -> void:
	var s := VariantsScript.windup_speed_scale(1.0, 0.5, 10, 0.1, 0.5, 1.6)
	_ok(is_equal_approx(s, 0.5), "contact at 0.5 s of a 1.0 s clip over a 1.0 s windup is 0.5x")
	_ok(is_equal_approx(VariantsScript.windup_speed_scale(1.0, 0.5, 2, 0.1, 0.5, 1.6), 1.6), "short windups clamp to the max speed")
	_ok(is_equal_approx(VariantsScript.windup_speed_scale(1.0, 0.1, 30, 0.1, 0.5, 1.6), 0.5), "long windups clamp to the min speed")
	_ok(VariantsScript.windup_speed_scale(1.0, 0.5, 0, 0.1, 0.5, 1.6) == 1.0, "zero windup leaves speed at 1.0")
	_ok(VariantsScript.windup_speed_scale(0.0, 0.5, 8, 0.1, 0.5, 1.6) == 1.0, "zero-length clip leaves speed at 1.0")


func _test_hit_side_and_directional_clip() -> void:
	var forward := Vector3(0, 0, 1)
	_ok(VariantsScript.hit_side(forward, Vector3(1, 0, 0)) != VariantsScript.hit_side(forward, Vector3(-1, 0, 0)), "opposite sides give opposite hit sides")
	var profile := {"hit_directional": {"left": "hit_left", "right": "hit_right"}}
	var l := VariantsScript.hit_clip(profile, "1", 0, forward, Vector3(1, 0, 0), true)
	var r := VariantsScript.hit_clip(profile, "1", 0, forward, Vector3(-1, 0, 0), true)
	_ok(l != r and l in ["hit_left", "hit_right"], "directional profile picks different clips per side")
	_ok(VariantsScript.hit_clip(profile, "1", 0, forward, Vector3.ZERO, false) == "hit", "no resolvable source keeps the base hit")


func _test_idle_interval_and_phase() -> void:
	var profile := {"idle_variation": {"min_interval_s": 4.0, "max_interval_s": 9.0}}
	for i in range(30):
		var v := VariantsScript.idle_interval(profile, str(i), i)
		if v < 4.0 or v > 9.0:
			_ok(false, "idle interval %f outside configured bounds" % v)
			return
	_ok(true, "idle interval stays inside the configured bounds")
	_ok(VariantsScript.idle_interval({}, "1", 0) == 0.0, "no idle_variation key disables the alt idle")
	var p := VariantsScript.phase_fraction("42")
	_ok(p >= 0.0 and p < 1.0 and p == VariantsScript.phase_fraction("42"), "phase offset is deterministic in [0, 1)")


func _scene_for(visual_key: String) -> Node3D:
	var packed := load("res://scenes/%s.tscn" % visual_key) as PackedScene
	return packed.instantiate() as Node3D if packed != null else null


func _test_catalog_clips_exist_in_rigs_and_profile_refs_resolve() -> void:
	for key in LoaderScript.visual_keys():
		var cfg := LoaderScript.monster(str(key))
		var profile := LoaderScript.clip_profile(str(cfg["clip_profile"]))
		var logicals: Dictionary = profile.get("clips", {})
		var bad := ""
		for group in (profile.get("variants", {}) as Dictionary):
			for member in profile["variants"][group]:
				if not logicals.has(member):
					bad = "%s variants.%s -> %s" % [key, group, member]
		for ref in (profile.get("hit_directional", {}) as Dictionary).values():
			if not logicals.has(ref):
				bad = "%s hit_directional -> %s" % [key, ref]
		for ref in (profile.get("attack_contact", {}) as Dictionary).keys():
			if not logicals.has(ref):
				bad = "%s attack_contact -> %s" % [key, ref]
		_ok(bad == "", "profile references resolve for %s %s" % [key, bad])


func _make_rig(visual_key: String):
	var root := Node3D.new()
	var model := _scene_for(visual_key)
	root.add_child(model)
	get_root().add_child(root)
	await process_frame
	var ap := model.find_child("AnimationPlayer", true, false) as AnimationPlayer
	var ctrl = ControllerScript.new(ap)
	var rec := {"node": root, "controller": ctrl, "type": "monster"}
	return rec


func _test_skeleton_windup_strike_and_hit() -> void:
	DriverScript.reset_for_tests()
	var rec: Dictionary = await _make_rig("monster_kit_skeleton_warrior")
	var entities := {"10": rec}
	var ctrl = rec["controller"]
	var windup := {"event_type": "monster_attack_windup", "source_entity_id": "10", "total_ticks": 8, "attack_style": "melee"}
	_ok(DriverScript.on_windup(windup, entities), "melee windup starts an attack clip")
	_ok(ctrl.current_clip().begins_with("attack"), "windup plays an attack variant")
	var ap := ctrl._player as AnimationPlayer
	_ok(ap.speed_scale < 1.0, "windup slows the clip so contact lands at windup end")
	var hit := {"event_type": "monster_damaged", "target_entity_id": "10", "entity_id": "10"}
	var before: String = ctrl.current_clip()
	DriverScript.play_event_clip(ctrl, rec, hit, "hit", entities)
	_ok(ctrl.current_clip() == before, "a hit does not cut into a windup pose")
	var strike := {"event_type": "player_damaged", "source_entity_id": "10"}
	_ok(not DriverScript.on_strike(strike, entities), "the strike does not replay the swing the windup already started")
	_ok(DriverScript.on_strike(strike, entities), "a later strike without a windup swings a variant")
	var seen := {}
	for i in range(6):
		rec["controller"].cancel_one_shot(ctrl.current_one_shot())
		DriverScript.on_strike(strike, entities)
		seen[ctrl.current_clip()] = true
	_ok(seen.size() > 1, "consecutive swings rotate through attack variants")
	rec["windup_pose_until_ms"] = 0
	DriverScript.play_event_clip(ctrl, rec, hit, "hit", entities)
	_ok(ctrl.current_clip() in ["hit", "hit_b"], "outside a windup the hit plays a hit variant")


func _test_wolf_directional_hit() -> void:
	DriverScript.reset_for_tests()
	var rec: Dictionary = await _make_rig("monster_wolf")
	var src := Node3D.new()
	get_root().add_child(src)
	var entities := {"20": rec, "21": {"node": src, "type": "monster"}}
	var ctrl = rec["controller"]
	var clips := {}
	for pos in [Vector3(3, 0, 0), Vector3(-3, 0, 0)]:
		src.global_position = pos
		ctrl.cancel_one_shot(ctrl.current_one_shot())
		DriverScript.play_event_clip(ctrl, rec, {"source_entity_id": "21", "target_entity_id": "20"}, "hit", entities)
		clips[ctrl.current_clip()] = true
	_ok(clips.size() == 2 and clips.has("hit_left") and clips.has("hit_right"), "wolf picks left and right hit clips by damage side")


func _test_spawn_guard_and_cancel() -> void:
	DriverScript.reset_for_tests()
	var rec: Dictionary = await _make_rig("monster_kit_skeleton_warrior")
	var ctrl = rec["controller"]
	var camera := Camera3D.new()
	get_root().add_child(camera)
	camera.global_position = Vector3(0, 5, 8)
	camera.look_at(Vector3.ZERO)
	rec["hp"] = 10
	_ok(not DriverScript.on_live_spawn(rec, [{"op": "wall_layout_update"}], camera), "level-arrival batches never play spawn-in")
	_ok(DriverScript.on_live_spawn(rec, [{"op": "entity_spawn"}], camera), "a live spawn in view plays the spawn clip")
	_ok(ctrl.current_clip() == "spawn", "spawn clip is the active clip")
	ctrl.set_locomotion(true)
	_ok(ctrl.current_clip() == "walk", "first movement cancels the spawn clip")
	var off := {"node": rec["node"], "controller": rec["controller"], "type": "monster", "hp": 10}
	camera.look_at(Vector3(0, 5, 20))
	_ok(not DriverScript.on_live_spawn(off, [], camera), "off-screen monsters skip spawn-in")
	var wolf: Dictionary = await _make_rig("monster_wolf")
	wolf["hp"] = 5
	camera.look_at(Vector3.ZERO)
	_ok(not DriverScript.on_live_spawn(wolf, [], camera), "profiles with spawn disabled skip spawn-in")


func _test_terminal_suppresses_one_shots() -> void:
	DriverScript.reset_for_tests()
	var rec: Dictionary = await _make_rig("monster_kit_skeleton_warrior")
	var ctrl = rec["controller"]
	var entities := {"30": rec}
	ctrl.enter_terminal("death")
	_ok(not DriverScript.on_windup({"source_entity_id": "30", "total_ticks": 8, "attack_style": "melee"}, entities), "terminal monsters ignore windup")
	_ok(not DriverScript.on_strike({"source_entity_id": "30"}, entities), "terminal monsters ignore strikes")
	_ok(ctrl.current_clip() == "death", "death clip stays latched")


func _test_idle_variation() -> void:
	IdleScript.reset_for_tests()
	var rec: Dictionary = await _make_rig("monster_kit_skeleton_warrior")
	var ctrl = rec["controller"]
	_ok(ctrl._idle_variation != null, "kit monsters get idle variation attached")
	var ap := ctrl._player as AnimationPlayer
	_ok(ap.current_animation == "idle", "idle loop is playing before the alt fires")
	ctrl._idle_variation._on_timer()
	_ok(ctrl.current_clip() in ["idle_alt", "idle_alt_b"], "alt idle plays when the monster is idle")
	ctrl.set_locomotion(true)
	_ok(ctrl.current_clip() == "walk", "movement cancels the alt idle")
	ctrl.set_locomotion(false)
	ctrl.enter_terminal("death")
	ctrl._idle_variation._on_timer()
	_ok(ctrl.current_clip() == "death", "terminal monsters never play alt idles")
	var hero := ControllerScript.new(null)
	_ok(hero._idle_variation == null, "controllers without a kit profile get no idle variation")


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_monster_anim_variants (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_monster_anim_variants (%d checks)" % _pass_count)
	quit(0)
