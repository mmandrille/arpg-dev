## Chooses one safe renderer capture for configured skill-hit events in a visual replay.
class_name SkillVisualCapture
extends RefCounted

const BotFrameCaptureScript := preload("res://scripts/bot_frame_capture.gd")


static func resolve(config: Dictionary, event: Dictionary, captured_skill_ids: Dictionary) -> Dictionary:
	if str(config.get("type", "")) != "capture_frame":
		return {}
	if str(event.get("event_type", "")) != str(config.get("event_type", "")):
		return {}
	if bool(event.get("blocked", false)) or str(event.get("outcome", "")) != "hit":
		return {}
	var skill_id := str(event.get("skill_id", ""))
	if str(event.get("target_entity_id", "")) == "":
		return {}
	var configured_skills = config.get("skill_ids", [])
	if typeof(configured_skills) != TYPE_ARRAY or skill_id == "" or skill_id not in configured_skills:
		return {}
	var configured_monster := str(config.get("monster_def_id", ""))
	if configured_monster != "" and str(event.get("monster_def_id", "")) != configured_monster:
		return {}
	if captured_skill_ids.has(skill_id):
		return {}
	var effect_id := CombatVfx.reaction_effect_id(event, "hit")
	if effect_id == "hit_spark":
		return {}
	var name := "%s_%s" % [str(config.get("name_prefix", "")), skill_id]
	if not BotFrameCaptureScript.valid_name(name):
		return {}
	captured_skill_ids[skill_id] = true
	return {"name": name, "skill_id": skill_id, "effect_id": effect_id}


static func capture_delay_frames(config: Dictionary) -> int:
	return clampi(int(config.get("delay_frames", 0)), 0, 3)
