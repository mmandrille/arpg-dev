# Unit tests for AuraSoftLights priority and radius wiring.
#
# Run via: godot --headless --path client --script res://tests/test_aura_soft_lights.gd
extends SceneTree

const AuraSoftLightsScript := preload("res://scripts/aura_soft_lights.gd")
const AuraLightPresentationLoaderScript := preload("res://scripts/aura_light_presentation_loader.gd")
const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	AuraLightPresentationLoaderScript.ensure_loaded()
	_test_sanctuary_beats_holy_shield_and_rage()
	_test_holy_shield_beats_rage()
	_test_rage_personal_radius()
	_test_elite_command_radius_preview_uses_dungeon_radius()
	_test_no_aura_removes_light()
	print("[gdtest] PASS: test_aura_soft_lights (%d passed, %d failed)" % [_pass_count, _fail_count])
	if _fail_count > 0:
		quit(1)
	else:
		quit(0)


func _test_sanctuary_beats_holy_shield_and_rage() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	var state := AuraSoftLightsScript.build_state(
		["sanctuary", "holy_shield"],
		"hero",
		{
			"rage_active": true,
			"sanctuary_radius": 5.0,
			"holy_shield_radius": 5.0,
		},
	)
	AuraSoftLightsScript.sync_aura(root, state)
	_assert_eq("sanctuary wins priority", AuraSoftLightsScript.active_aura_id(root), PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID)
	var light := root.find_child(AuraSoftLightsScript.AURA_LIGHT_NAME, false, false) as OmniLight3D
	_assert_true("sanctuary light exists", light != null)
	if light != null:
		_assert_float("sanctuary light range follows skill radius", light.omni_range, 5.0)
	root.queue_free()


func _test_holy_shield_beats_rage() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	AuraSoftLightsScript.sync_aura(root, AuraSoftLightsScript.build_state(
		["holy_shield"],
		"hero",
		{"rage_active": true, "holy_shield_radius": 5.0},
	))
	_assert_eq("holy shield wins over rage", AuraSoftLightsScript.active_aura_id(root), PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID)
	root.queue_free()


func _test_rage_personal_radius() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	AuraSoftLightsScript.sync_aura(root, AuraSoftLightsScript.build_state([], "hero", {"rage_active": true}))
	var expected := AuraLightPresentationLoaderScript.presentation_personal_radius()
	_assert_eq("rage aura id", AuraSoftLightsScript.active_aura_id(root), PlayerStatusEffectMarkers.RAGE_EFFECT_ID)
	_assert_float("rage personal radius", AuraSoftLightsScript.aura_light_range(root), expected)
	root.queue_free()


func _test_elite_command_radius_preview_uses_dungeon_radius() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	AuraSoftLightsScript.sync_aura(root, AuraSoftLightsScript.build_state([], "monster", {
		"monster_pack_leader": true,
		"elite_radius_preview_active": true,
		"elite_aura_radius": 4.0,
	}))
	_assert_eq("leader preview aura id", AuraSoftLightsScript.active_aura_id(root), "elite_command_radius_preview")
	_assert_float("leader preview radius", AuraSoftLightsScript.elite_command_radius_preview_value(root), 4.0)
	root.queue_free()


func _test_no_aura_removes_light() -> void:
	var root := Node3D.new()
	get_root().add_child(root)
	AuraSoftLightsScript.sync_aura(root, AuraSoftLightsScript.build_state(["elite_command"], "monster", {}))
	_assert_true("elite command light active", AuraSoftLightsScript.has_elite_command_effect(root))
	AuraSoftLightsScript.sync_aura(root, AuraSoftLightsScript.build_state([], "monster", {}))
	_assert_eq("aura cleared", AuraSoftLightsScript.active_aura_id(root), "")
	_assert_true("light removed", root.find_child(AuraSoftLightsScript.AURA_LIGHT_NAME, false, false) == null)
	root.queue_free()


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_float(label: String, got: float, want: float, epsilon: float = 0.001) -> void:
	if abs(got - want) <= epsilon:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)
