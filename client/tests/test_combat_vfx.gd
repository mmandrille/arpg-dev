extends SceneTree

## v492 combat VFX: bursts built from vfx_presentation.v0.json, damage-type colours, reaction spawning
## (parent + direction), quality scaling, heal rain structure. Expectations derive from the catalogs.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	_test_burst_matches_catalog()
	_test_damage_type_colors()
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


func _test_reaction_spawns_on_parent_away_from_source() -> void:
	var parent := Node3D.new()
	var hit := CombatVfx.spawn_for_reaction(parent, Vector3(5, 0, 5), Vector3(0, 0, 5), {"damage_type": "fire"}, "hit")
	_assert_true("hit spark added to the parent", hit != null and hit.get_parent() == parent and str(hit.name).ends_with("hit_spark"))
	if hit != null:
		var dir := (hit.process_material as ParticleProcessMaterial).direction
		_assert_true("hit spark sprays away from the source", dir.x > 0.0)
		_assert_true("hit spark at spawn height", is_equal_approx(hit.position.y, float(_catalog()["effects"]["hit_spark"]["spawn_height"])))
	var death := CombatVfx.spawn_for_reaction(parent, Vector3(1, 0, 1), Vector3.ZERO, {}, "death")
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
