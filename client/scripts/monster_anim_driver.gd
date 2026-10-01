## Event -> clip routing for monster combat animation polish (v513).
## Client-only presentation (ADR-0007): reacts to events the server already emits
## (monster_attack_windup, player_damaged, monster_damaged, entity_spawn). Every entry point
## is a no-op for monsters without a kit clip profile or with the optional keys absent.
class_name MonsterAnimDriver
extends RefCounted

const VariantsScript := preload("res://scripts/monster_anim_variants.gd")
const TICK_SECONDS := 0.1
## Structural performance guards (not gameplay tuning).
const MAX_CONCURRENT_SPAWN_CLIPS := 4
const WINDUP_FOLLOW_THROUGH_S := 0.5

static var _spawn_end_ms: Array = []


static func reset_for_tests() -> void:
	_spawn_end_ms.clear()


static func _monster_controller(rec: Dictionary):
	if str(rec.get("type", "")) != "monster":
		return null
	var ctrl = rec.get("controller", null)
	if ctrl == null or ctrl.is_terminal():
		return null
	return ctrl


static func _next_counter(rec: Dictionary, key: String) -> int:
	var counters: Dictionary = rec.get("anim_counters", {})
	var value := int(counters.get(key, 0))
	counters[key] = value + 1
	rec["anim_counters"] = counters
	return value


static func windup_active(rec: Dictionary) -> bool:
	return Time.get_ticks_msec() < int(rec.get("windup_pose_until_ms", 0))


## monster_attack_windup: start the attack clip, scaled so its contact point lands as the windup ends.
static func on_windup(ev: Dictionary, entities: Dictionary) -> bool:
	var rec: Dictionary = entities.get(str(ev.get("source_entity_id", "")), {})
	var ctrl = _monster_controller(rec)
	if ctrl == null or str(ev.get("attack_style", "melee")) != "melee":
		return false
	var profile: Dictionary = ctrl.profile()
	var contact: Dictionary = profile.get("attack_contact", {})
	var total_ticks := int(ev.get("total_ticks", 0))
	if contact.is_empty() or total_ticks <= 0:
		return false
	var key := str(ev.get("source_entity_id", ""))
	var clip := VariantsScript.pick(VariantsScript.group_clips(profile, "attack"), key, _next_counter(rec, "attack"))
	if not ctrl.has_clip(clip):
		return false
	var limits: Dictionary = profile.get("windup_speed", {})
	var scale := VariantsScript.windup_speed_scale(
		ctrl.clip_length(clip),
		float(contact.get(clip, VariantsScript.DEFAULT_CONTACT_FRACTION)),
		total_ticks, TICK_SECONDS,
		float(limits.get("min", 0.5)), float(limits.get("max", 1.6)))
	ctrl.play_one_shot(clip, "", scale)
	rec["windup_pose_until_ms"] = Time.get_ticks_msec() + int((float(total_ticks) * TICK_SECONDS + WINDUP_FOLLOW_THROUGH_S) * 1000.0)
	return true


## player_damaged/player_killed from a melee-style monster strike. Windup monsters already
## swing; monsters without a windup swing here. Ranged profiles opt out (strike_on_damage false):
## the damage event lands at projectile impact, too late to read as the shot.
static func on_strike(ev: Dictionary, entities: Dictionary) -> bool:
	var source_id := str(ev.get("source_entity_id", ""))
	var rec: Dictionary = entities.get(source_id, {})
	var ctrl = _monster_controller(rec)
	if ctrl == null:
		return false
	if windup_active(rec):
		rec["windup_pose_until_ms"] = 0
		return false
	var profile: Dictionary = ctrl.profile()
	if profile.is_empty() or not bool(profile.get("strike_on_damage", true)):
		return false
	var clip := VariantsScript.pick(VariantsScript.group_clips(profile, "attack"), source_id, _next_counter(rec, "attack"))
	if not ctrl.has_clip(clip):
		return false
	ctrl.play_one_shot(clip)
	return true


## Plays the controller clip for a monster event; replaces the bare `play_one_shot(clip)`.
## Hits pick a directional/alternate clip and never cut into a windup pose.
static func play_event_clip(ctrl, rec: Dictionary, ev: Dictionary, clip: String, entities: Dictionary) -> void:
	if clip != "hit":
		ctrl.play_one_shot(clip)
		return
	var profile: Dictionary = ctrl.profile()
	if profile.is_empty():
		ctrl.play_one_shot(clip)
		return
	if windup_active(rec):
		return
	var node := rec.get("node", null) as Node3D
	var forward := Vector3.BACK
	var to_source := Vector3.ZERO
	var has_source := false
	var source: Dictionary = entities.get(str(ev.get("source_entity_id", "")), {})
	var source_node := source.get("node", null) as Node3D
	if node != null and node.is_inside_tree() and source_node != null and source_node.is_inside_tree():
		forward = node.global_transform.basis * Vector3.BACK
		to_source = source_node.global_position - node.global_position
		forward.y = 0.0
		to_source.y = 0.0
		has_source = to_source.length() > 0.001 and forward.length() > 0.001
	var key := str(ev.get("target_entity_id", ev.get("entity_id", "")))
	var picked := VariantsScript.hit_clip(profile, key, _next_counter(rec, "hit"), forward, to_source, has_source)
	if not ctrl.has_clip(picked):
		picked = clip
	ctrl.play_one_shot(picked, "", float(profile.get("hit_speed_scale", 1.0)))


## First appearance through an entity_spawn delta (fog reveal, summons, boss adds). Skipped for
## level-arrival batches (they carry wall_layout_update), off-screen monsters, disabled profiles,
## and when too many spawn clips are already running. Cancelled by the first movement.
static func on_live_spawn(rec: Dictionary, changes: Array, camera: Camera3D) -> bool:
	for change in changes:
		if str((change as Dictionary).get("op", "")) == "wall_layout_update":
			return false
	var ctrl = _monster_controller(rec)
	if ctrl == null or int(rec.get("hp", 1)) <= 0:
		return false
	var profile: Dictionary = ctrl.profile()
	var cfg: Dictionary = profile.get("spawn", {})
	var clip := str(cfg.get("clip", ""))
	if not bool(cfg.get("enabled", false)) or clip == "" or not ctrl.has_clip(clip):
		return false
	var node := rec.get("node", null) as Node3D
	if node == null or not node.is_inside_tree() or camera == null or not camera.is_position_in_frustum(node.global_position):
		return false
	var now := Time.get_ticks_msec()
	_spawn_end_ms = _spawn_end_ms.filter(func(end_ms): return end_ms > now)
	if _spawn_end_ms.size() >= MAX_CONCURRENT_SPAWN_CLIPS:
		return false
	_spawn_end_ms.append(now + int(ctrl.clip_length(clip) * 1000.0))
	ctrl.play_one_shot(clip)
	ctrl.mark_motion_interruptible(clip)
	return true
