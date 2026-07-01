class_name BotPresentationDebug
extends RefCounted

const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")
const PlayerStatusEffectMarkers := preload("res://scripts/player_status_effect_markers.gd")
const AuraSoftLightsScript := preload("res://scripts/aura_soft_lights.gd")
const ClientConstantsScript := preload("res://scripts/client_constants.gd")


static func entities_debug(entities: Dictionary, live_monster_ids: Array) -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) == "monster" and not live_monster_ids.has(id):
			continue
		out.append({
			"id": str(id),
			"type": str(rec.get("type", "")),
			"monster_def_id": str(rec.get("monster_def_id", "")),
			"interactable_def_id": str(rec.get("interactable_def_id", "")), "elite_objective": bool(rec.get("elite_objective", false)), "quest_reward": bool(rec.get("quest_reward", false)), "is_boss": bool(rec.get("is_boss", false)), "boss_template_id": str(rec.get("boss_template_id", "")),
			"item_def_id": str(rec.get("item_def_id", "")),
			"item_template_id": str(rec.get("item_template_id", "")),
			"rarity": str(rec.get("rarity", "")),
			"state": str(rec.get("state", "")),
		})
	return out


static func local_player_effect_ids(player_anchor: Node3D) -> Array:
	var out := []
	if player_anchor == null:
		return out
	if PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor):
		out.append(PlayerStatusEffectMarkers.HOLY_SHIELD_EFFECT_ID)
	if PlayerStatusEffectMarkers.has_sanctuary_effect(player_anchor):
		out.append(PlayerStatusEffectMarkers.SANCTUARY_EFFECT_ID)
	return out


static func local_player_presentation(
	player_id: String,
	player_visual_scale: float,
	player_anchor: Node3D,
	charge_channel_visual,
	player_reaction,
	player_anim,
) -> Dictionary:
	return {
		"id": player_id, "type": "player", "visual_model": "character", "visual_scale": player_visual_scale,
		"effect_ids": local_player_effect_ids(player_anchor),
		"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(player_anchor),
		"has_sanctuary_effect": PlayerStatusEffectMarkers.has_sanctuary_effect(player_anchor),
		"active_aura_id": AuraSoftLightsScript.active_aura_id(player_anchor),
		"aura_light_range": AuraSoftLightsScript.aura_light_range(player_anchor),
		"aura_light_color": AuraSoftLightsScript.aura_light_color(player_anchor),
		"holy_shield_cast_pulses": AuraSoftLightsScript.active_holy_shield_cast_pulse_count(player_anchor),
		"holy_shield_aura_pulses": AuraSoftLightsScript.active_holy_shield_cast_pulse_count(player_anchor),
		"holy_shield_target_pulses": AuraSoftLightsScript.active_holy_shield_target_pulse_count(player_anchor),
		"has_rage_effect": PlayerStatusEffectMarkers.has_rage_effect(player_anchor), "base_tint": ClientConstantsScript.PLAYER_TINT.to_html(false),
		"charge_channel_visual": charge_channel_visual.get_debug_state(),
		"reaction": player_reaction.get_debug_state() if player_reaction != null else {},
		"animation": player_anim.get_debug_state() if player_anim != null else {},
	}


static func entities_presentation_debug(entities: Dictionary) -> Array:
	var out: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		var node := rec.get("node", null) as Node3D
		var node_pos := node.position if node != null else Vector3.ZERO
		var reaction = rec.get("reaction", null)
		var controller = rec.get("controller", null)
		out.append({
			"id": str(id), "type": str(rec.get("type", "")), "monster_def_id": str(rec.get("monster_def_id", "")),
			"character_id": str(rec.get("character_id", "")), "visual_model": visual_model_name(rec, node),
			"position": {"x": node_pos.x, "z": node_pos.z},
			"visual_scale": float(rec.get("visual_scale", 1.0)),
			"is_boss": bool(rec.get("is_boss", false)), "boss_template_id": str(rec.get("boss_template_id", "")),
			"boss_phase": rec.get("boss_phase", {}), "boss_telegraph_active": bool(rec.get("boss_telegraph_active", false)),
			"telegraph_tint": str(rec.get("telegraph_tint", "")), "has_boss_telegraph_marker": bool(rec.get("has_boss_telegraph_marker", false)),
			"telegraph_radius": float(rec.get("telegraph_radius", 0.0)), "telegraph_marker_color": str(rec.get("telegraph_marker_color", "")), "telegraph_marker_shape": str(rec.get("telegraph_marker_shape", "")),
			"base_tint": str(rec.get("base_tint", "")), "has_bow_marker": bool(rec.get("has_bow_marker", false)), "effect_ids": rec.get("effect_ids", []),
			"monster_pack_id": str(rec.get("monster_pack_id", "")), "monster_pack_leader": bool(rec.get("monster_pack_leader", false)),
			"interactable_def_id": str(rec.get("interactable_def_id", "")), "elite_objective": bool(rec.get("elite_objective", false)),
			"quest_reward": bool(rec.get("quest_reward", false)), "has_objective_marker": ChestPresentationScript.has_objective_marker(node), "has_quest_marker": ChestPresentationScript.has_quest_marker(node),
			"has_holy_shield_effect": PlayerStatusEffectMarkers.has_holy_shield_effect(node), "has_sanctuary_effect": PlayerStatusEffectMarkers.has_sanctuary_effect(node),
			"active_aura_id": AuraSoftLightsScript.active_aura_id(node),
			"aura_light_range": AuraSoftLightsScript.aura_light_range(node),
			"aura_light_color": AuraSoftLightsScript.aura_light_color(node),
			"has_burning_effect": PlayerStatusEffectMarkers.has_burning_effect(node), "has_bleed_effect": PlayerStatusEffectMarkers.has_bleed_effect(node), "has_elite_command_effect": PlayerStatusEffectMarkers.has_elite_command_effect(node),
			"has_pinning_root_effect": PlayerStatusEffectMarkers.has_pinning_root_effect(node), "has_stun_effect": PlayerStatusEffectMarkers.has_stun_effect(node),
			"has_rogue_mark_effect": PlayerStatusEffectMarkers.has_rogue_mark_effect(node),
			"has_elite_command_radius_preview": PlayerStatusEffectMarkers.has_elite_command_radius_preview(node), "elite_command_radius_preview": PlayerStatusEffectMarkers.elite_command_radius_preview_value(node),
			"holy_shield_target_pulses": AuraSoftLightsScript.active_holy_shield_target_pulse_count(node), "hp": int(rec.get("hp", 1)),
			"reaction": reaction.get_debug_state() if reaction != null else {}, "animation": controller.get_debug_state() if controller != null else {},
		})
	return out


static func visual_model_name(rec: Dictionary, node: Node3D) -> String:
	if str(rec.get("visual_model", "")) != "":
		return str(rec.get("visual_model", ""))
	if node != null and node.find_child("ModelRoot", true, false) != null:
		return "character"
	if str(rec.get("type", "")) == "player":
		return "primitive"
	return ""


static func active_damage_numbers(damage_numbers_layer: Node) -> Array:
	var out: Array = []
	if damage_numbers_layer == null:
		return out
	for child in damage_numbers_layer.get_children():
		if child is DamageNumber:
			out.append(damage_number_entry(child as DamageNumber))
	return out


static func damage_number_entry(pop: DamageNumber) -> Dictionary:
	return {
		"text": pop.combat_text,
		"variant": pop.combat_variant,
		"damage_type": pop.combat_damage_type,
		"color": pop.label_settings.font_color.to_html(false) if pop.label_settings != null else "",
	}


static func damage_number_key(entry: Dictionary) -> String:
	return "%s|%s|%s" % [str(entry.get("variant", "")), str(entry.get("text", "")), str(entry.get("damage_type", ""))]


static func recent_damage_numbers(
	damage_numbers_layer: Node,
	history: Array,
	prune_history: Callable,
) -> Array:
	var out := active_damage_numbers(damage_numbers_layer)
	prune_history.call()
	var seen: Dictionary = {}
	for entry in out:
		seen[damage_number_key(entry)] = true
	for entry in history:
		var key := damage_number_key(entry)
		if seen.has(key):
			continue
		seen[key] = true
		out.append((entry as Dictionary).duplicate(true))
	return out
