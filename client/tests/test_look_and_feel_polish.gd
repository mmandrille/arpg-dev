extends SceneTree

const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")
const ClassPresentationsLoaderScript := preload("res://scripts/class_presentations_loader.gd")
const MonsterFamilyAccentScript := preload("res://scripts/monster_family_accent.gd")
const TownAmbientLifeScript := preload("res://scripts/town_ambient_life.gd")
const CameraImpactFeedbackScript := preload("res://scripts/camera_impact_feedback.gd")
const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")
const CombatEventPresentationScript := preload("res://scripts/combat_event_presentation.gd")
const PlayerDamageVignetteScript := preload("res://scripts/player_damage_vignette.gd")
const MonsterMeleeWindupMarkerScript := preload("res://scripts/monster_melee_windup_marker.gd")
const ItemTooltipPanelScript := preload("res://scripts/item_tooltip_panel.gd")
const DamageNumberScript := preload("res://scripts/damage_number.gd")
const LevelUpBurstScript := preload("res://scripts/level_up_burst.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	ClassPresentationsLoaderScript.ensure_loaded()
	_test_class_idle_stance()
	_test_monster_family_accent()
	_test_town_ambient_life()
	_test_camera_impact_feedback()
	_test_chest_open_burst()
	_test_tooltip_rarity_border()
	_test_combat_camera_binding()
	_test_player_damage_vignette()
	_test_monster_melee_windup_marker()
	_test_level_up_floating_text()
	_test_level_up_burst_and_feedback()
	_finish()


func _test_class_idle_stance() -> void:
	var model := Node3D.new()
	ClassIdleStanceScript.apply_to_model(model, "warrior")
	var stance: Dictionary = ClassPresentationsLoaderScript.idle_stance_for_class("warrior")
	if not is_equal_approx(model.rotation_degrees.z, float(stance.get("lean_degrees", 0.0))):
		_fail("class idle stance should apply lean")
		model.free()
		return
	model.free()
	_pass("class idle stance")


func _test_monster_family_accent() -> void:
	var node := Node3D.new()
	MonsterFamilyAccentScript.sync_for_monster(node, "dungeon_wolf")
	var marker := node.find_child("MonsterFamilyAccent", false, false)
	if marker == null:
		_fail("dungeon wolf should get family accent marker")
		node.free()
		return
	MonsterFamilyAccentScript.sync_for_monster(node, "dungeon_mob")
	if node.find_child("MonsterFamilyAccent", false, false) == null:
		_fail("dungeon mob should get family accent marker")
		node.free()
		return
	node.free()
	_pass("monster family accent")


func _test_town_ambient_life() -> void:
	var root := Node3D.new()
	TownAmbientLifeScript.attach_to_town(root)
	var props := root.find_child("TownAmbientLife", false, false)
	if props == null or props.get_child_count() < 3:
		_fail("town ambient life should add silhouettes")
		root.free()
		return
	TownAmbientLifeScript.attach_to_town(root)
	if props.get_child_count() != 3:
		_fail("town ambient life should attach once")
		root.free()
		return
	root.free()
	_pass("town ambient life")


func _test_camera_impact_feedback() -> void:
	CameraImpactFeedbackScript.apply_from_damage(12, 40)
	if CameraImpactFeedbackScript.get_offset() == Vector3.ZERO:
		_fail("camera impact should produce a shake offset")
		return
	CameraImpactFeedbackScript.decay(1.0)
	if CameraImpactFeedbackScript.is_active():
		_fail("camera impact should decay to rest")
		return
	_pass("camera impact feedback")


func _test_chest_open_burst() -> void:
	var chest := Node3D.new()
	ChestPresentationScript.sync_open_burst(chest, true)
	if chest.find_child("ChestOpenBurst", true, false) == null:
		_fail("chest open should spawn burst marker")
		chest.free()
		return
	chest.free()
	_pass("chest open burst")


func _test_tooltip_rarity_border() -> void:
	if ItemTooltipPanelScript.border_width_for_rarity("common") != 1:
		_fail("common tooltip should use thin border")
		return
	if ItemTooltipPanelScript.border_width_for_rarity("rare") < 2:
		_fail("rare tooltip should use thicker rarity border")
		return
	_pass("tooltip rarity border")


func _test_combat_camera_binding() -> void:
	var camera := Camera3D.new()
	CombatEventPresentationScript.bind_camera(camera, 50, 0.0)
	CombatEventPresentationScript.show_combat_text_for_event(
		"1001",
		{"event_type": "player_damaged", "damage": 8},
		Color.WHITE,
		Callable(self, "_noop_damage"),
		Callable(self, "_noop_node"),
	)
	if CameraImpactFeedbackScript.get_offset() == Vector3.ZERO:
		_fail("combat presentation should route player damage to camera shake")
		camera.free()
		return
	camera.free()
	_pass("combat camera binding")


func _test_player_damage_vignette() -> void:
	var root := Node.new()
	PlayerDamageVignetteScript.attach(root)
	PlayerDamageVignetteScript.pulse(10, 50)
	if PlayerDamageVignetteScript.debug_strength() <= 0.0:
		_fail("player damage vignette should pulse on local damage")
		root.free()
		return
	CombatEventPresentationScript.bind_camera(Camera3D.new(), 50, 0.0, "p1")
	CombatEventPresentationScript.show_combat_text_for_event(
		"p2",
		{"event_type": "player_damaged", "damage": 12},
		Color.WHITE,
		Callable(self, "_noop_damage"),
		Callable(self, "_noop_node"),
	)
	var before := PlayerDamageVignetteScript.debug_strength()
	CombatEventPresentationScript.show_combat_text_for_event(
		"p1",
		{"event_type": "player_damaged", "damage": 12},
		Color.WHITE,
		Callable(self, "_noop_damage"),
		Callable(self, "_noop_node"),
	)
	if PlayerDamageVignetteScript.debug_strength() <= before:
		_fail("player damage vignette should only pulse for local player")
		root.free()
		return
	PlayerDamageVignetteScript.reset_session()
	if PlayerDamageVignetteScript.debug_strength() != 0.0:
		_fail("player damage vignette should reset on session clear")
		root.free()
		return
	root.free()
	_pass("player damage vignette")


func _test_monster_melee_windup_marker() -> void:
	var monster := Node3D.new()
	var rec := {"node": monster, "type": "monster"}
	var entities := {"9002": rec}
	MonsterMeleeWindupMarkerScript.sync_from_event(
		{"source_entity_id": "9002", "total_ticks": 10, "attack_style": "pounce"},
		entities,
	)
	if not bool(rec.get("has_melee_windup_marker", false)):
		_fail("melee windup should attach marker")
		monster.free()
		return
	MonsterMeleeWindupMarkerScript.clear_for_record(rec)
	if bool(rec.get("has_melee_windup_marker", true)):
		_fail("melee windup marker should clear")
		monster.free()
		return
	monster.free()
	_pass("monster melee windup marker")


func _test_level_up_floating_text() -> void:
	var camera := Camera3D.new()
	var anchor := Node3D.new()
	var root := Node.new()
	root.add_child(camera)
	root.add_child(anchor)
	var pop := DamageNumberScript.new() as DamageNumber
	root.add_child(pop)
	pop.setup(camera, anchor, Vector3.ZERO, null, Color("#ffe08a"), 0.0, "", "level_up", "Level Up!")
	if pop.combat_text != "Level Up!":
		_fail("level up floating text should read Level Up!")
		root.free()
		return
	if pop.combat_variant != "level_up":
		_fail("level up floating text should use level_up variant")
		root.free()
		return
	var settings := pop.label_settings
	if settings == null or settings.font_color != Color("#ffe08a"):
		_fail("level up floating text should use yellow color")
		root.free()
		return
	root.free()
	_pass("level up floating text")


func _test_level_up_burst_and_feedback() -> void:
	var root := Node3D.new()
	var anchor := Node3D.new()
	root.add_child(anchor)
	LevelUpBurstScript.spawn(root, anchor.position)
	if root.find_child("LevelUpBurst", true, false) == null:
		_fail("level up should spawn energy burst")
		root.free()
		return
	root.free()
	_pass("level up energy burst")


func _noop_damage(_entity_id: String, _color: Color, _amount = null, _prefix: String = "", _scale: float = 0.0, _variant: String = "", _text: String = "", _damage_type: String = "") -> void:
	pass


func _noop_node(_entity_id: String) -> Node3D:
	return null


func _pass(_label: String) -> void:
	_pass_count += 1


func _fail(label: String) -> void:
	_fail_count += 1
	push_error(label)


func _finish() -> void:
	if _fail_count == 0:
		print("[gdtest] PASS: test_look_and_feel_polish (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(0)
		return
	print("[gdtest] FAIL: test_look_and_feel_polish (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1)
