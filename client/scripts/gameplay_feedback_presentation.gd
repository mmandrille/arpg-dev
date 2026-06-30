## Local-player feedback: level-up burst, consumable use FX, and monster impact gating.
class_name GameplayFeedbackPresentation
extends RefCounted

const CombatFeelConfigScript := preload("res://scripts/combat_feel_config.gd")
const CombatEventPresentationScript := preload("res://scripts/combat_event_presentation.gd")
const ConsumableHealEffectScript := preload("res://scripts/consumable_heal_effect.gd")
const ConsumableUseEffectScript := preload("res://scripts/consumable_use_effect.gd")
const LevelUpBurstScript := preload("res://scripts/level_up_burst.gd")
const ClientAudioBridgeScript := preload("res://scripts/client_audio_bridge.gd")
const ModelReactionControllerScript := preload("res://scripts/model_reaction_controller.gd")


static func bind_session(main: Node, entities: Dictionary) -> void:
	CombatFeelConfigScript.ensure_loaded()
	CombatEventPresentationScript.configure_entity_impact_gate(
		func(entity_id: String) -> bool:
			return entity_combat_impacts_allowed(entities, entity_id)
	)


static func entity_combat_impacts_allowed(entities: Dictionary, entity_id: String) -> bool:
	if CombatFeelConfigScript.enemy_impact_feedback_enabled():
		return true
	if not entities.has(entity_id):
		return true
	return str((entities[entity_id] as Dictionary).get("type", "")) != "monster"


static func handle_player_healed(main: Node, ev: Dictionary, entity_id: String, audio_controller: Node) -> void:
	ClientAudioBridgeScript.heal(audio_controller)
	var heal_skill := str(ev.get("skill_id", ""))
	var level_up_restore := str(ev.get("reason", "")) == "level_up"
	if not level_up_restore and heal_skill != "heal":
		spawn_consumable_use_feedback(main, entity_id, true, false, Callable(main, "_node_for_entity_id"), Callable(main, "_node_world_or_local_position"))
	if not level_up_restore:
		main.call("_show_damage_number", entity_id, Color(0.3, 1.0, 0.45), ev.get("heal", null), "+", 1.0)
	if heal_skill == "heal":
		main.call("_spawn_heal_rain", entity_id)
	elif heal_skill == "" and not level_up_restore:
		spawn_consumable_heal_effect(main, entity_id, Callable(main, "_node_for_entity_id"), Callable(main, "_node_world_or_local_position"))


static func handle_player_mana_restored(main: Node, ev: Dictionary, entity_id: String) -> void:
	if str(ev.get("skill_id", "")) == "" and str(ev.get("reason", "")) != "level_up":
		spawn_consumable_use_feedback(main, entity_id, false, true, Callable(main, "_node_for_entity_id"), Callable(main, "_node_world_or_local_position"))
	if str(ev.get("reason", "")) != "level_up":
		main.call("_show_damage_number", entity_id, Color("#54c7f3"), ev.get("mana", null), "+", 1.0)


static func handle_character_leveled(main: Node, audio_controller: Node, player_anchor: Node3D, world_pos_fn: Callable) -> void:
	if player_anchor == null:
		return
	LevelUpBurstScript.spawn(main, world_pos_fn.call(player_anchor))
	ClientAudioBridgeScript.heal(audio_controller)


static func spawn_consumable_heal_effect(main: Node, entity_id: String, node_for_entity: Callable, world_pos_fn: Callable) -> void:
	var target: Node3D = node_for_entity.call(entity_id)
	if target == null:
		return
	var effect := ConsumableHealEffectScript.new() as Node3D
	effect.position = world_pos_fn.call(target) + Vector3(0.0, 0.45, 0.0)
	main.add_child(effect)


static func spawn_consumable_use_feedback(
	main: Node,
	entity_id: String,
	restores_hp: bool,
	restores_mana: bool,
	node_for_entity: Callable,
	world_pos_fn: Callable,
) -> void:
	var target: Node3D = node_for_entity.call(entity_id)
	if target == null:
		return
	ConsumableUseEffectScript.spawn(main, world_pos_fn.call(target), restores_hp, restores_mana)


static func play_entity_reaction(
	entities: Dictionary,
	player_id: String,
	player_anchor: Node3D,
	player_reaction,
	entity_id: String,
	ev: Dictionary,
	reaction_name: String,
	world_pos_fn: Callable,
) -> void:
	if not entity_combat_impacts_allowed(entities, entity_id):
		return
	var reaction = reaction_for_entity(entities, player_id, player_reaction, entity_id)
	if reaction == null:
		return
	var source_pos := source_position_for_event(entities, player_id, player_anchor, ev, world_pos_fn)
	var fallback := fallback_reaction_direction(entities, player_id, player_anchor, entity_id, world_pos_fn)
	if reaction_name == "death":
		reaction.enter_death(source_pos, fallback)
	else:
		reaction.play_hit(source_pos, fallback)


static func reaction_for_entity(entities: Dictionary, player_id: String, player_reaction, entity_id: String):
	if entity_id == player_id:
		return player_reaction
	if entities.has(entity_id):
		return (entities[entity_id] as Dictionary).get("reaction", null)
	return null


static func source_position_for_event(
	entities: Dictionary,
	player_id: String,
	player_anchor: Node3D,
	ev: Dictionary,
	world_pos_fn: Callable,
) -> Vector3:
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "":
		return ModelReactionControllerScript.UNRESOLVED_SOURCE
	if source_id == player_id and player_anchor != null:
		return world_pos_fn.call(player_anchor)
	if entities.has(source_id):
		var rec: Dictionary = entities[source_id]
		var node := rec.get("node", null) as Node3D
		if node != null:
			return world_pos_fn.call(node)
	return ModelReactionControllerScript.UNRESOLVED_SOURCE


static func fallback_reaction_direction(
	entities: Dictionary,
	player_id: String,
	player_anchor: Node3D,
	entity_id: String,
	world_pos_fn: Callable,
) -> Vector3:
	var target := entity_world_position(entities, player_id, player_anchor, entity_id, world_pos_fn)
	if target != ModelReactionControllerScript.UNRESOLVED_SOURCE and player_anchor != null:
		var direction: Vector3 = target - world_pos_fn.call(player_anchor)
		direction.y = 0.0
		if direction.length() > 0.001:
			return direction.normalized()
	return Vector3.BACK


static func entity_world_position(
	entities: Dictionary,
	player_id: String,
	player_anchor: Node3D,
	entity_id: String,
	world_pos_fn: Callable,
) -> Vector3:
	if entity_id == player_id and player_anchor != null:
		return world_pos_fn.call(player_anchor)
	if entities.has(entity_id):
		var rec: Dictionary = entities[entity_id]
		var node := rec.get("node", null) as Node3D
		if node != null:
			return world_pos_fn.call(node)
	return ModelReactionControllerScript.UNRESOLVED_SOURCE
