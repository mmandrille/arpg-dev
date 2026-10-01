extends SceneTree

const SkillVisualCaptureScript := preload("res://scripts/skill_visual_capture.gd")
const SkillVisualCaptureRuntimeScript := preload("res://scripts/skill_visual_capture_runtime.gd")
const PlayerCameraControllerScript := preload("res://scripts/player_camera_controller.gd")

## v492 combat VFX: bursts built from vfx_presentation.v0.json, damage-type colours, reaction spawning
## (parent + direction), quality scaling, heal rain structure. Expectations derive from the catalogs.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_burst_matches_catalog()
	_test_damage_type_colors()
	_test_skill_hit_profiles_from_catalog()
	_test_unmapped_skill_fallback_and_death()
	_test_skill_visual_capture_selection()
	_test_visual_replay_camera_view()
	_test_reaction_spawns_on_parent_away_from_source()
	_test_quality_scales_amount()
	_test_heal_rain_has_particles_and_ring()
	_finish()


func _catalog() -> Dictionary:
	var path := ProjectSettings.globalize_path("res://").path_join("../shared/assets/vfx_presentation.v0.json")
	return JSON.parse_string(FileAccess.get_file_as_string(path))


func _test_burst_matches_catalog() -> void:
	CombatVfx.set_quality("balanced")
	for effect_id in ["hit_spark", "death_burst"]:
		var cfg: Dictionary = _catalog()["effects"][effect_id]
		var burst := CombatVfx.make_burst(effect_id)
		_assert_true("%s is a one-shot burst" % effect_id, burst != null and burst.one_shot and burst.emitting)
		if burst == null:
			continue
		var mat := burst.process_material as ParticleProcessMaterial
		_assert_true("%s amount from catalog" % effect_id, burst.amount == CombatVfx.scaled_amount(int(cfg["amount"])))
		_assert_true("%s lifetime from catalog" % effect_id, is_equal_approx(burst.lifetime, float(cfg["lifetime"])))
		_assert_true("%s gravity from catalog" % effect_id, is_equal_approx(mat.gravity.y, float(cfg["gravity"])))
		var shader_mat := (burst.draw_pass_1 as QuadMesh).material as ShaderMaterial
		_assert_true("%s draws with the glow shader" % effect_id, shader_mat != null and shader_mat.shader.resource_path.ends_with("vfx_soft_glow.gdshader"))
		_assert_true("%s frees itself when done" % effect_id, burst.finished.is_connected(burst.queue_free))
		burst.free()


func _test_damage_type_colors() -> void:
	var by_type: Dictionary = _catalog()["damage_type_colors"]
	for damage_type in by_type.keys():
		var colors := CombatVfx.colors_for("hit_spark", damage_type)
		_assert_true("%s spark colour from catalog" % damage_type, colors[0].is_equal_approx(Color(str(by_type[damage_type]["color"]))))
	var fallback := CombatVfx.colors_for("hit_spark", "not_a_type")
	_assert_true("unknown damage type uses the effect colour", fallback[0].is_equal_approx(Color(str(_catalog()["effects"]["hit_spark"]["color"]))))


func _test_skill_hit_profiles_from_catalog() -> void:
	var catalog := _catalog()
	var mappings: Dictionary = catalog.get("skill_hit_effects", {})
	var skill_ids := mappings.keys()
	skill_ids.sort()
	_assert_true("only the selected projectile skills use dedicated hit VFX", skill_ids == ["ice_shard", "lightning", "magic_bolt"])
	var parent := Node3D.new()
	var profile_signatures := {}
	for skill_id in skill_ids:
		var effect_id := str(mappings[skill_id])
		var cfg: Dictionary = catalog["effects"].get(effect_id, {})
		var burst := CombatVfx.spawn_for_reaction(
			parent,
			Vector3(5, 0, 5),
			Vector3(0, 0, 5),
			{"skill_id": str(skill_id), "damage_type": "force"},
			"hit",
		)
		_assert_true("%s selects its catalog burst" % skill_id, burst != null and burst.name == CombatVfx.BURST_PREFIX + effect_id)
		_assert_true("%s burst is attached to the target parent" % skill_id, burst != null and burst.get_parent() == parent)
		if burst == null or cfg.is_empty():
			continue
		var process := burst.process_material as ParticleProcessMaterial
		_assert_true("%s amount comes from its catalog preset" % skill_id, burst.amount == CombatVfx.scaled_amount(int(cfg["amount"])))
		_assert_true("%s lifetime comes from its catalog preset" % skill_id, is_equal_approx(burst.lifetime, float(cfg["lifetime"])))
		_assert_true("%s spread comes from its catalog preset" % skill_id, is_equal_approx(process.spread, float(cfg["spread_degrees"])))
		_assert_true("%s gravity comes from its catalog preset" % skill_id, is_equal_approx(process.gravity.y, float(cfg["gravity"])))
		_assert_true("%s uses its own palette instead of force damage colors" % skill_id, _ramp_start(process).is_equal_approx(Color(str(cfg["color"]))))
		_assert_true("%s burst frees itself when done" % skill_id, burst.finished.is_connected(burst.queue_free))
		_assert_true("%s sprays away from its source" % skill_id, process.direction.x > 0.0)
		_assert_true("%s spawn height comes from its catalog preset" % skill_id, is_equal_approx(burst.position.y, float(cfg["spawn_height"])))
		var surface_offset := float(cfg.get("spawn_surface_offset", 0.0))
		_assert_true("%s impact burst offsets onto the target surface" % skill_id,
			burst.position.is_equal_approx(Vector3(5.0, float(cfg["spawn_height"]), 5.0) + Vector3.RIGHT * surface_offset))
		profile_signatures[JSON.stringify([cfg["amount"], cfg["lifetime"], cfg["spread_degrees"], cfg["velocity_max"], cfg["size_max"]])] = true
		burst.free()
	_assert_true("selected skills have distinct particle profiles", profile_signatures.size() == skill_ids.size())
	parent.free()


func _test_unmapped_skill_fallback_and_death() -> void:
	var parent := Node3D.new()
	var generic := CombatVfx.spawn_for_reaction(parent, Vector3.ZERO, Vector3.ZERO, {"skill_id": "cleave", "damage_type": "fire"}, "hit")
	_assert_true("unmapped skill keeps the generic hit spark", generic != null and generic.name == CombatVfx.BURST_PREFIX + "hit_spark")
	if generic != null:
		var fire: Dictionary = _catalog()["damage_type_colors"]["fire"]
		_assert_true("generic fallback keeps damage-type colors", _ramp_start(generic.process_material as ParticleProcessMaterial).is_equal_approx(Color(str(fire["color"]))))
		generic.free()
	var death := CombatVfx.spawn_for_reaction(parent, Vector3.ZERO, Vector3.ZERO, {"skill_id": "ice_shard", "damage_type": "cold"}, "death")
	_assert_true("death reaction ignores skill-specific hit mapping", death != null and death.name == CombatVfx.BURST_PREFIX + "death_burst")
	if death != null:
		death.free()
	parent.free()


func _test_skill_visual_capture_selection() -> void:
	var scenario_path := ProjectSettings.globalize_path("res://").path_join("../tools/bot/scenarios/44_skill_visual.json")
	var scenario = JSON.parse_string(FileAccess.get_file_as_string(scenario_path))
	_assert_true("skill visual scenario config loads", typeof(scenario) == TYPE_DICTIONARY)
	if typeof(scenario) != TYPE_DICTIONARY:
		return
	var visual_cfg: Dictionary = scenario.get("visual", {})
	var capture_cfg: Dictionary = visual_cfg.get("capture_frame", {})
	var configured_skills: Array = capture_cfg.get("skill_ids", [])
	_assert_true("visual replay capture is a headless-skippable capture_frame", \
		capture_cfg.get("type", "") == "capture_frame" and bool(capture_cfg.get("skip_if_headless", false)))
	_assert_true("skill impact capture allows three rendered frames for the burst", \
		SkillVisualCaptureScript.capture_delay_frames(capture_cfg) == 3)
	_assert_true("capture delay is bounded", \
		SkillVisualCaptureScript.capture_delay_frames({"delay_frames": 99}) == 3)
	_assert_true("capture allowlist matches the configured skill hit VFX", configured_skills == ["magic_bolt", "ice_shard", "lightning"])
	var mappings: Dictionary = _catalog().get("skill_hit_effects", {})
	var captured := {}
	for skill_id in configured_skills:
		var event := {"event_type": "monster_damaged", "outcome": "hit", "skill_id": str(skill_id), "monster_def_id": "combat_lab_soft_target", "target_entity_id": "target_%s" % skill_id}
		var capture := SkillVisualCaptureScript.resolve(capture_cfg, event, captured)
		_assert_true("%s resolves a deterministic capture name" % skill_id, capture.get("name", "") == "v502_skill_%s" % skill_id)
		_assert_true("%s records its mapped effect" % skill_id, capture.get("effect_id", "") == mappings.get(skill_id, ""))
		_assert_true("%s duplicate hit does not save twice" % skill_id, SkillVisualCaptureScript.resolve(capture_cfg, event, captured).is_empty())
	_assert_true("wrong target cannot trigger skill capture", SkillVisualCaptureScript.resolve(
		capture_cfg, {"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "monster_def_id": "skill_xp_level6_dummy", "target_entity_id": "other_123"}, {}).is_empty())
	for invalid_contact in [
		{"outcome": "block", "blocked": true},
		{"outcome": "miss", "blocked": false},
		{"outcome": "immune", "blocked": false},
		{"outcome": "hit", "blocked": true},
	]:
		var invalid_event := {"event_type": "monster_damaged", "skill_id": "magic_bolt", "monster_def_id": "combat_lab_soft_target", "target_entity_id": "target_123"}
		invalid_event.merge(invalid_contact, true)
		_assert_true("non-contact outcome cannot trigger a visual proof capture", SkillVisualCaptureScript.resolve(
			capture_cfg, invalid_event, {}).is_empty())
	_assert_true("missing outcome cannot trigger a visual proof capture", SkillVisualCaptureScript.resolve(
		capture_cfg, {"event_type": "monster_damaged", "skill_id": "magic_bolt", "monster_def_id": "combat_lab_soft_target", "target_entity_id": "target_123"}, {}).is_empty())
	_assert_true("missing target cannot trigger a visual proof capture", SkillVisualCaptureScript.resolve(
		capture_cfg, {"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "monster_def_id": "combat_lab_soft_target"}, {}).is_empty())
	_assert_true("unlisted skill hit cannot trigger capture", SkillVisualCaptureScript.resolve(
		capture_cfg, {"event_type": "monster_damaged", "outcome": "hit", "skill_id": "cleave", "target_entity_id": "target_123"}, captured).is_empty())
	_assert_true("non-damage event cannot trigger capture", SkillVisualCaptureScript.resolve(
		capture_cfg, {"event_type": "monster_killed", "outcome": "hit", "skill_id": "magic_bolt", "target_entity_id": "target_123"}, captured).is_empty())
	var disabled_cfg := capture_cfg.duplicate(true)
	disabled_cfg["type"] = "disabled"
	_assert_true("non-capture directives cannot start a frame write", SkillVisualCaptureScript.resolve(
		disabled_cfg, {"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "target_entity_id": "target_123"}, {}).is_empty())
	var unsafe_cfg := capture_cfg.duplicate(true)
	unsafe_cfg["name_prefix"] = "../escape"
	_assert_true("unsafe capture path is rejected", SkillVisualCaptureScript.resolve(
		unsafe_cfg, {"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "target_entity_id": "target_123"}, {}).is_empty())
	var camera_cfg: Dictionary = visual_cfg.get("camera", {})
	_assert_true("visual replay uses zoom without guessing a world focus",
		camera_cfg.has("zoom") and not camera_cfg.has("x") and not camera_cfg.has("y"))
	var capture_runtime = SkillVisualCaptureRuntimeScript.new()
	capture_runtime.configure(visual_cfg)
	var damage_env := {"type": "state_delta", "payload": {"events": [{
		"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "monster_def_id": "combat_lab_soft_target", "target_entity_id": "target_123",
	}]}}
	_assert_true("eligible skill hit supplies its target for replay framing",
		capture_runtime.capture_target_id(damage_env) == "target_123")
	var unrelated_env := {"type": "state_delta", "payload": {"events": [{
		"event_type": "monster_damaged", "outcome": "hit", "skill_id": "cleave", "monster_def_id": "combat_lab_soft_target", "target_entity_id": "other_123",
	}]}}
	_assert_true("unlisted hit does not alter replay framing", capture_runtime.capture_target_id(unrelated_env) == "")
	var wrong_target_env := {"type": "state_delta", "payload": {"events": [{
		"event_type": "monster_damaged", "outcome": "hit", "skill_id": "magic_bolt", "monster_def_id": "skill_xp_level6_dummy", "target_entity_id": "other_123",
	}]}}
	_assert_true("wrong target does not alter replay framing", capture_runtime.capture_target_id(wrong_target_env) == "")


func _test_visual_replay_camera_view() -> void:
	_assert_true("missing visual camera config leaves replay framing unchanged",
		PlayerCameraControllerScript.visual_replay_camera_config({}).is_empty())
	_assert_true("invalid visual camera zoom leaves replay framing unchanged",
		PlayerCameraControllerScript.visual_replay_camera_config({"camera": {"zoom": 0.0}}).is_empty())
	var scenario_path := ProjectSettings.globalize_path("res://").path_join("../tools/bot/scenarios/44_skill_visual.json")
	var scenario: Dictionary = JSON.parse_string(FileAccess.get_file_as_string(scenario_path))
	var visual_config: Dictionary = scenario.get("visual", {})
	var camera_config := PlayerCameraControllerScript.visual_replay_camera_config(visual_config)
	_assert_true("zoom-only replay camera config is supported", not camera_config.is_empty())
	_assert_true("capture camera uses the maximum zoom allowed by camera catalog bounds",
		is_equal_approx(float(camera_config.get("zoom", 0.0)), 1.5))
	var controller = PlayerCameraControllerScript.new()
	controller._camera = Camera3D.new()
	root.add_child(controller._camera)
	controller._current_mode = "isometric"
	controller._cfg = {"zoom_default": 12.0, "zoom_min": 8.0, "zoom_max": 20.0}
	controller._iso_offset = Vector3(9.0, 20.0, 15.0)
	var world_focus := Vector3(4.0, 0.0, 5.0)
	controller.apply_visual_replay_view(world_focus, float(camera_config["zoom"]))
	_assert_true("visual replay camera applies configured zoom through the isometric bounds",
		is_equal_approx(controller._camera.size, clampf(12.0 / float(camera_config["zoom"]), 8.0, 20.0)))
	_assert_true("visual replay camera frames its supplied player/target anchor",
		controller._camera.global_position.is_equal_approx(world_focus + Vector3(9.0, 20.0, 15.0)))
	controller._camera.free()


func _ramp_start(process: ParticleProcessMaterial) -> Color:
	if process == null:
		return Color.TRANSPARENT
	var ramp := process.color_ramp as GradientTexture1D
	if ramp == null or ramp.gradient == null:
		return Color.TRANSPARENT
	return ramp.gradient.get_color(0)


func _test_reaction_spawns_on_parent_away_from_source() -> void:
	var parent := Node3D.new()
	var hit := CombatVfx.spawn_for_reaction(parent, Vector3(5, 0, 5), Vector3(0, 0, 5), {"damage_type": "fire"}, "hit")
	_assert_true("hit spark added to the parent", hit != null and hit.get_parent() == parent and str(hit.name).ends_with("hit_spark"))
	if hit != null:
		var dir := (hit.process_material as ParticleProcessMaterial).direction
		_assert_true("hit spark sprays away from the source", dir.x > 0.0)
		_assert_true("hit spark at spawn height", is_equal_approx(hit.position.y, float(_catalog()["effects"]["hit_spark"]["spawn_height"])))
	var death := CombatVfx.spawn_for_reaction(parent, Vector3(1, 0, 1), Vector3.ZERO, {"skill_id": "ice_shard"}, "death")
	_assert_true("death reaction spawns the death burst", death != null and str(death.name).ends_with("death_burst"))
	parent.free()


func _test_quality_scales_amount() -> void:
	var cfg: Dictionary = _catalog()["effects"]["death_burst"]
	CombatVfx.set_quality("performance")
	var scale := float(RenderPresentationLoader.quality_tier("performance").get("particle_scale", 1.0))
	var burst := CombatVfx.make_burst("death_burst")
	_assert_true("performance tier scales the amount", burst.amount == maxi(1, int(round(float(cfg["amount"]) * scale))) and burst.amount < int(cfg["amount"]))
	burst.free()
	CombatVfx.set_quality("balanced")


func _test_heal_rain_has_particles_and_ring() -> void:
	var rain := HealRainEffect.new()
	rain.setup(3.0)
	get_root().add_child(rain)
	_assert_true("heal rain has falling particles", rain.find_child("Vfx_heal_rain", false, false) is GPUParticles3D)
	_assert_true("heal rain has a ground ring", rain.find_child("HealRainRing", false, false) != null)
	rain.free()


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_combat_vfx (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_combat_vfx (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
