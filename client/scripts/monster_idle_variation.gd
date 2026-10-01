## Idle variation for rigged monsters (v513): entity-keyed start phase so crowds do not
## breathe in unison, plus an occasional alternate idle clip on a data-driven interval.
## Attached by AnimationController when its AnimationPlayer carries a kit clip profile.
class_name MonsterIdleVariation
extends RefCounted

const VariantsScript := preload("res://scripts/monster_anim_variants.gd")
## Structural performance guard (not gameplay tuning): at most this many alt idles at once.
const MAX_CONCURRENT_ALTS := 6

static var _alt_end_ms: Array = []

var _controller
var _player: AnimationPlayer
var _profile: Dictionary
var _key: String
var _counter: int = 0
var _armed: bool = false


static func attach(controller, player: AnimationPlayer, entity_key: String) -> MonsterIdleVariation:
	var profile := VariantsScript.profile_of(player)
	if profile.is_empty() or (profile.get("idle_variation", {}) as Dictionary).is_empty():
		return null
	var variation := MonsterIdleVariation.new()
	variation._controller = controller
	variation._player = player
	variation._profile = profile
	variation._key = entity_key
	variation._apply_phase()
	return variation


static func reset_for_tests() -> void:
	_alt_end_ms.clear()


## Called by AnimationController whenever it enters a locomotion clip.
func on_locomotion_clip(clip: String) -> void:
	if clip == "idle":
		_arm()


func _apply_phase() -> void:
	if _player == null or not _player.has_animation("idle"):
		return
	var length := _player.get_animation("idle").length
	_player.seek(length * VariantsScript.phase_fraction(_key), true)


func _arm() -> void:
	if _armed or _player == null or not _player.is_inside_tree():
		return
	var delay := VariantsScript.idle_interval(_profile, _key, _counter)
	if delay <= 0.0:
		return
	_armed = true
	_player.get_tree().create_timer(delay).timeout.connect(_on_timer)


func _on_timer() -> void:
	_armed = false
	if _player == null or not is_instance_valid(_player) or not _player.is_inside_tree():
		return
	if _controller.is_terminal():
		return
	_counter += 1
	var now := Time.get_ticks_msec()
	_alt_end_ms = _alt_end_ms.filter(func(end_ms): return end_ms > now)
	var clip := VariantsScript.pick(VariantsScript.group_clips(_profile, "idle_alt"), _key, _counter)
	if not _controller.is_idle() or _alt_end_ms.size() >= MAX_CONCURRENT_ALTS or not _player.has_animation(clip):
		_arm()
		return
	_alt_end_ms.append(now + int(_player.get_animation(clip).length * 1000.0))
	_controller.play_one_shot(clip)
	_controller.mark_motion_interruptible(clip)
	_arm()
