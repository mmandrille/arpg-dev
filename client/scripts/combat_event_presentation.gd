## Client-only combat text + impact/outcome presentation for authoritative events.
class_name CombatEventPresentation
extends RefCounted

const DamageTypeCombatTextScript := preload("res://scripts/damage_type_combat_text.gd")
const ImpactSparksScript := preload("res://scripts/impact_sparks.gd")
const CombatOutcomePunchScript := preload("res://scripts/combat_outcome_punch.gd")
const CameraImpactFeedbackScript := preload("res://scripts/camera_impact_feedback.gd")
const PlayerDamageVignetteScript := preload("res://scripts/player_damage_vignette.gd")

static var _camera: Camera3D = null
static var _max_hp: int = 1
static var _local_player_id: String = ""


static func bind_camera(camera: Camera3D, max_hp: int, delta: float, local_player_id: String = "") -> void:
	_camera = camera
	_max_hp = max_hp
	if local_player_id != "":
		_local_player_id = local_player_id
	decay_camera(camera, delta)


static func clear_session() -> void:
	_camera = null
	_max_hp = 1
	_local_player_id = ""
	CameraImpactFeedbackScript.reset_session()
	PlayerDamageVignetteScript.reset_session()


static func show_combat_text_for_event(
	entity_id: String,
	ev: Dictionary,
	default_color: Color,
	show_damage_number: Callable,
	node_for_entity_id: Callable,
) -> void:
	if str(ev.get("event_type", "")) == "player_damaged":
		CameraImpactFeedbackScript.apply_from_damage(int(ev.get("damage", 0)), _max_hp)
		if entity_id == _local_player_id:
			PlayerDamageVignetteScript.pulse(int(ev.get("damage", 0)), _max_hp)
	var outcome := str(ev.get("outcome", ""))
	var damage = ev.get("damage", null)
	var special := DamageTypeCombatTextScript.special_outcome(outcome)
	if not special.is_empty():
		show_damage_number.call(
			entity_id,
			special.get("color", Color.WHITE),
			null,
			"",
			0.0,
			str(special.get("variant", outcome)),
			str(special.get("text", "")),
		)
		spawn_outcome_punch(entity_id, ev, node_for_entity_id)
		return

	var presentation := DamageTypeCombatTextScript.number_for_event(ev, default_color)
	if not presentation.is_empty():
		show_damage_number.call(
			entity_id,
			presentation.get("color", default_color),
			presentation.get("amount", damage),
			"",
			0.0,
			str(presentation.get("variant", "normal")),
			str(presentation.get("text", "")),
			str(presentation.get("damage_type", "")),
		)
		spawn_impact_sparks(entity_id, ev, presentation.get("color", default_color), node_for_entity_id)
		spawn_outcome_punch(entity_id, ev, node_for_entity_id)
		return

	show_damage_number.call(entity_id, default_color, damage)
	spawn_impact_sparks(entity_id, ev, default_color, node_for_entity_id)
	spawn_outcome_punch(entity_id, ev, node_for_entity_id)


static func spawn_outcome_punch(entity_id: String, ev: Dictionary, node_for_entity_id: Callable) -> void:
	if not CombatOutcomePunchScript.should_spawn(ev):
		return
	var target: Node3D = node_for_entity_id.call(entity_id)
	if target != null:
		target.add_child(CombatOutcomePunchScript.make_node(ev))


static func spawn_impact_sparks(
	entity_id: String,
	ev: Dictionary,
	fallback_color: Color,
	node_for_entity_id: Callable,
) -> void:
	if not ImpactSparksScript.should_spawn(ev):
		return
	var target: Node3D = node_for_entity_id.call(entity_id)
	if target != null:
		target.add_child(ImpactSparksScript.make_node(ev, fallback_color))


static func decay_camera(_camera: Camera3D, delta: float) -> void:
	CameraImpactFeedbackScript.decay(delta)
