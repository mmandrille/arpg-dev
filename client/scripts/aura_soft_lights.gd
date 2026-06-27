## Data-driven soft OmniLight3D auras for heroes and monsters.
class_name AuraSoftLights
extends RefCounted

const LoaderScript := preload("res://scripts/aura_light_presentation_loader.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")

const HOLY_SHIELD_EFFECT_ID := "holy_shield"
const SANCTUARY_EFFECT_ID := "sanctuary"
const RAGE_EFFECT_ID := "rage"
const ELITE_COMMAND_EFFECT_ID := "elite_command"
const ELITE_COMMAND_RADIUS_PREVIEW_ID := "elite_command_radius_preview"

const AURA_LIGHT_NAME := "AuraSoftLight"
const META_ACTIVE_AURA_ID := "active_aura_id"
const META_AURA_LIGHT_RANGE := "aura_light_range"
const CAST_PULSE_META := "holy_shield_cast_pulses"
const ALLY_PULSE_META := "holy_shield_target_pulses"


static func sync_aura(root: Node3D, state: Dictionary) -> void:
	if root == null:
		return
	LoaderScript.ensure_loaded()
	var aura_id := _resolve_active_aura(state)
	var existing := root.find_child(AURA_LIGHT_NAME, false, false) as OmniLight3D
	if aura_id == "":
		_clear_aura_meta(root)
		if existing != null:
			root.remove_child(existing)
			existing.queue_free()
		return
	var radius := _resolve_radius(aura_id, state)
	if existing == null:
		existing = OmniLight3D.new()
		existing.name = AURA_LIGHT_NAME
		root.add_child(existing)
	_apply_light_config(existing, aura_id, state, radius)
	root.set_meta(META_ACTIVE_AURA_ID, aura_id)
	root.set_meta(META_AURA_LIGHT_RANGE, radius)


static func pulse_holy_shield_cast(source_root: Node3D, affected_roots: Array, radius: float) -> void:
	if source_root == null:
		return
	LoaderScript.ensure_loaded()
	var pulse_cfg := LoaderScript.cast_pulse(HOLY_SHIELD_EFFECT_ID)
	var duration := maxf(float(pulse_cfg.get("duration_s", 0.3)), 0.05)
	var energy_peak := float(pulse_cfg.get("energy_peak_multiplier", 2.6))
	var range_peak := float(pulse_cfg.get("range_peak_multiplier", 1.12))
	var ally_bump := float(pulse_cfg.get("target_energy_bump", 1.35))
	var safe_radius := maxf(radius, 0.5)
	_increment_pulse_counter(source_root, CAST_PULSE_META)
	_pulse_light(
		source_root,
		safe_radius,
		HOLY_SHIELD_EFFECT_ID,
		"monster" if _is_monster_root(source_root) else "hero",
		duration,
		energy_peak,
		range_peak,
	)
	for affected in affected_roots:
		var affected_root := affected as Node3D
		if affected_root == null:
			continue
		_increment_pulse_counter(affected_root, ALLY_PULSE_META)
		_pulse_existing_light(affected_root, duration, ally_bump, 1.0)


static func active_aura_id(root: Node3D) -> String:
	if root == null or not root.has_meta(META_ACTIVE_AURA_ID):
		return ""
	return str(root.get_meta(META_ACTIVE_AURA_ID, ""))


static func aura_light_range(root: Node3D) -> float:
	if root == null:
		return 0.0
	if root.has_meta(META_AURA_LIGHT_RANGE):
		return float(root.get_meta(META_AURA_LIGHT_RANGE, 0.0))
	var light := root.find_child(AURA_LIGHT_NAME, false, false) as OmniLight3D
	if light == null:
		return 0.0
	return light.omni_range


static func aura_light_color(root: Node3D) -> String:
	var light := root.find_child(AURA_LIGHT_NAME, false, false) as OmniLight3D
	if light == null:
		return ""
	return light.light_color.to_html(false)


static func has_aura(root: Node3D, aura_id: String) -> bool:
	return active_aura_id(root) == aura_id


static func has_rage_effect(root: Node3D) -> bool:
	return has_aura(root, RAGE_EFFECT_ID)


static func has_holy_shield_effect(root: Node3D) -> bool:
	return has_aura(root, HOLY_SHIELD_EFFECT_ID)


static func has_sanctuary_effect(root: Node3D) -> bool:
	return has_aura(root, SANCTUARY_EFFECT_ID)


static func has_elite_command_effect(root: Node3D) -> bool:
	return has_aura(root, ELITE_COMMAND_EFFECT_ID)


static func has_elite_command_radius_preview(root: Node3D) -> bool:
	return has_aura(root, ELITE_COMMAND_RADIUS_PREVIEW_ID)


static func elite_command_radius_preview_value(root: Node3D) -> float:
	if not has_elite_command_radius_preview(root):
		return 0.0
	return aura_light_range(root)


static func active_holy_shield_cast_pulse_count(root: Node3D) -> int:
	return _pulse_counter(root, CAST_PULSE_META)


static func active_holy_shield_target_pulse_count(root: Node3D) -> int:
	return _pulse_counter(root, ALLY_PULSE_META)


static func sync_monster_aura_from_record(
	rec: Dictionary,
	sanctuary_radius: float,
	holy_shield_radius: float,
) -> void:
	var node := rec.get("node", null) as Node3D
	if node == null:
		return
	var alive := int(rec.get("hp", 1)) > 0
	sync_aura(node, build_state(
		rec.get("effect_ids", []) if alive else [],
		"monster",
		{
			"monster_pack_leader": bool(rec.get("monster_pack_leader", false)),
			"sanctuary_radius": sanctuary_radius,
			"holy_shield_radius": holy_shield_radius,
		},
	))


static func sync_hero_aura(
	root: Node3D,
	effect_ids_value,
	rage_active: bool,
	sanctuary_radius: float,
	holy_shield_radius: float,
) -> void:
	sync_aura(root, build_state(effect_ids_value, "hero", {
		"rage_active": rage_active,
		"sanctuary_radius": sanctuary_radius,
		"holy_shield_radius": holy_shield_radius,
	}))


static func strip_elite_command_effect_ids(effect_ids_value) -> Array:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	var filtered: Array = []
	for effect_id in effect_ids:
		if str(effect_id) != ELITE_COMMAND_EFFECT_ID:
			filtered.append(effect_id)
	return filtered


static func local_player_effect_ids(entities: Dictionary, player_id: String) -> Array:
	if not entities.has(player_id):
		return []
	var effect_ids = entities[player_id].get("effect_ids", [])
	if effect_ids is Array:
		return effect_ids.duplicate()
	return []


static func build_state(
	effect_ids_value,
	entity_kind: String,
	opts: Dictionary = {},
) -> Dictionary:
	var effect_ids: Array = effect_ids_value if effect_ids_value is Array else []
	return {
		"effect_ids": effect_ids,
		"entity_kind": entity_kind,
		"rage_active": bool(opts.get("rage_active", false)),
		"monster_pack_leader": bool(opts.get("monster_pack_leader", false)),
		"elite_radius_preview_active": bool(opts.get("elite_radius_preview_active", false)),
		"elite_aura_radius": float(opts.get("elite_aura_radius", 0.0)),
		"sanctuary_radius": float(opts.get("sanctuary_radius", 0.0)),
		"holy_shield_radius": float(opts.get("holy_shield_radius", 0.0)),
	}


static func _resolve_active_aura(state: Dictionary) -> String:
	var effect_ids: Array = state.get("effect_ids", [])
	var rage_active := bool(state.get("rage_active", false))
	var preview_active := bool(state.get("elite_radius_preview_active", false))
	var is_leader := bool(state.get("monster_pack_leader", false))
	for aura_id in LoaderScript.priority_list():
		match str(aura_id):
			SANCTUARY_EFFECT_ID:
				if effect_ids.has(SANCTUARY_EFFECT_ID):
					return SANCTUARY_EFFECT_ID
			HOLY_SHIELD_EFFECT_ID:
				if effect_ids.has(HOLY_SHIELD_EFFECT_ID):
					return HOLY_SHIELD_EFFECT_ID
			RAGE_EFFECT_ID:
				if rage_active or effect_ids.has(RAGE_EFFECT_ID):
					return RAGE_EFFECT_ID
			"elite_command_radius_preview":
				if preview_active and is_leader:
					return ELITE_COMMAND_RADIUS_PREVIEW_ID
			ELITE_COMMAND_EFFECT_ID:
				if effect_ids.has(ELITE_COMMAND_EFFECT_ID):
					return ELITE_COMMAND_EFFECT_ID
	return ""


static func _resolve_radius(aura_id: String, state: Dictionary) -> float:
	var entry := LoaderScript.aura_entry(aura_id)
	var source: Dictionary = entry.get("radius_source", {})
	match str(source.get("type", "")):
		"skill_effect_radius":
			var skill_id := str(source.get("skill_id", aura_id))
			if skill_id == SANCTUARY_EFFECT_ID and float(state.get("sanctuary_radius", 0.0)) > 0.0:
				return maxf(float(state.get("sanctuary_radius", 0.0)), 0.5)
			if skill_id == HOLY_SHIELD_EFFECT_ID and float(state.get("holy_shield_radius", 0.0)) > 0.0:
				return maxf(float(state.get("holy_shield_radius", 0.0)), 0.5)
			return _skill_effect_radius(skill_id, 5.0)
		"presentation_personal_radius":
			return maxf(LoaderScript.presentation_personal_radius(), 0.5)
		"dungeon_elite_aura_radius":
			return maxf(float(state.get("elite_aura_radius", 0.0)), 0.5)
	return 0.5


static func _apply_light_config(light: OmniLight3D, aura_id: String, state: Dictionary, radius: float) -> void:
	var entry := LoaderScript.aura_entry(aura_id)
	var entity_kind := str(state.get("entity_kind", "hero"))
	var multipliers: Dictionary = entry.get(entity_kind, entry.get("hero", {}))
	var energy_multiplier := float(multipliers.get("energy_multiplier", 1.0))
	var range_multiplier := float(multipliers.get("range_multiplier", 1.0))
	light.light_color = Color(str(entry.get("light_color", "#ffffff")))
	light.light_energy = float(entry.get("omni_energy", 1.0)) * energy_multiplier
	light.light_specular = float(entry.get("light_specular", 0.0))
	light.omni_range = radius * range_multiplier
	light.omni_attenuation = float(entry.get("omni_attenuation", 1.5))
	light.position.y = float(entry.get("height_offset", 0.4))
	light.shadow_enabled = bool(entry.get("shadow_enabled", false))


static func _pulse_light(
	root: Node3D,
	radius: float,
	aura_id: String,
	entity_kind: String,
	duration: float,
	energy_peak: float,
	range_peak: float,
) -> void:
	var light := root.find_child(AURA_LIGHT_NAME, false, false) as OmniLight3D
	var created := false
	if light == null:
		light = OmniLight3D.new()
		light.name = AURA_LIGHT_NAME
		root.add_child(light)
		_apply_light_config(light, aura_id, {"entity_kind": entity_kind}, radius)
		created = true
	var base_energy := light.light_energy
	var base_range := light.omni_range
	var tween := root.create_tween()
	tween.set_parallel(true)
	tween.tween_property(light, "light_energy", base_energy * energy_peak, duration * 0.45)
	tween.tween_property(light, "omni_range", base_range * range_peak, duration * 0.45)
	tween.chain().set_parallel(true)
	tween.tween_property(light, "light_energy", base_energy, duration * 0.55)
	tween.tween_property(light, "omni_range", base_range, duration * 0.55)
	if created:
		tween.chain().tween_callback(light.queue_free)


static func _pulse_existing_light(root: Node3D, duration: float, energy_peak: float, range_peak: float) -> void:
	var light := root.find_child(AURA_LIGHT_NAME, false, false) as OmniLight3D
	if light == null:
		return
	var base_energy := light.light_energy
	var base_range := light.omni_range
	var tween := root.create_tween()
	tween.set_parallel(true)
	tween.tween_property(light, "light_energy", base_energy * energy_peak, duration * 0.45)
	tween.tween_property(light, "omni_range", base_range * range_peak, duration * 0.45)
	tween.chain().set_parallel(true)
	tween.tween_property(light, "light_energy", base_energy, duration * 0.55)
	tween.tween_property(light, "omni_range", base_range, duration * 0.55)


static func _skill_effect_radius(skill_id: String, fallback: float) -> float:
	var def := SkillRulesLoaderScript.skill_definition(skill_id)
	for effect in def.get("effects", []):
		var row: Dictionary = effect
		if str(row.get("effect_id", skill_id)) == skill_id and row.has("radius"):
			return maxf(float(row.get("radius", fallback)), 0.5)
		if row.has("radius") and skill_id in [SANCTUARY_EFFECT_ID, HOLY_SHIELD_EFFECT_ID]:
			return maxf(float(row.get("radius", fallback)), 0.5)
	return maxf(fallback, 0.5)


static func _clear_aura_meta(root: Node3D) -> void:
	if root.has_meta(META_ACTIVE_AURA_ID):
		root.remove_meta(META_ACTIVE_AURA_ID)
	if root.has_meta(META_AURA_LIGHT_RANGE):
		root.remove_meta(META_AURA_LIGHT_RANGE)


static func _increment_pulse_counter(root: Node3D, key: String) -> void:
	root.set_meta(key, _pulse_counter(root, key) + 1)


static func _pulse_counter(root: Node3D, key: String) -> int:
	if root == null or not root.has_meta(key):
		return 0
	return int(root.get_meta(key, 0))


static func _is_monster_root(root: Node3D) -> bool:
	return str(root.get_meta("entity_kind", "")) == "monster"
