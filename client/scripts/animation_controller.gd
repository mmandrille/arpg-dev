extends RefCounted
class_name AnimationController
# Per-entity animation state machine (spec §4.5). Injected with its
# AnimationPlayer (no absolute scene-path lookups), so main.gd and smoke.gd
# share one code path. It does NOT parse protocol events or know entity types.
#
# State priority (highest wins): terminal (death) > one-shot (attack/hit) >
# locomotion (idle/walk).

var _player: AnimationPlayer
var _moving: bool = false
var _one_shot: String = ""
var _terminal: bool = false
var _terminal_clip: String = ""
var _warnings: Array = []
var _motion_interruptible: String = ""
var _idle_variation = null

const MeleeLungePresentationScript := preload("res://scripts/melee_lunge_presentation.gd")
const MonsterAnimVariantsScript := preload("res://scripts/monster_anim_variants.gd")
const IdleVariationScript := preload("res://scripts/monster_idle_variation.gd")
const IDLE := "idle"
const WALK := "walk"


func _init(player: AnimationPlayer) -> void:
	_player = player
	if _player != null:
		_player.animation_finished.connect(_on_finished)
	_play(IDLE)
	if _player != null:
		# v513: rigged monsters carry a kit clip profile; heroes and clip-less visuals do not.
		_idle_variation = IdleVariationScript.attach(self, _player, str(_player.get_instance_id()))
		if _idle_variation != null:
			_idle_variation.on_locomotion_clip(IDLE)


func set_locomotion(is_moving: bool) -> void:
	_moving = is_moving
	if is_moving and _one_shot != "" and _one_shot == _motion_interruptible:
		cancel_one_shot(_one_shot)
	if _terminal or _one_shot != "":
		return
	_play(WALK if is_moving else IDLE)


func play_one_shot(name: String, attack_mode: String = "", speed_scale: float = 1.0) -> void:
	if _terminal:
		return
	_one_shot = name
	_motion_interruptible = ""
	if _player != null:
		_player.speed_scale = maxf(speed_scale, 0.01)
	if _play(name):
		MeleeLungePresentationScript.start(_player, name, attack_mode)


func enter_terminal(name: String) -> void:
	_terminal = true
	_terminal_clip = name
	_one_shot = ""
	_motion_interruptible = ""
	_play(name)


func reset_terminal() -> void:
	_terminal = false
	_terminal_clip = ""
	_one_shot = ""
	_play(WALK if _moving else IDLE)


## v513: drop the current one-shot (spawn-in, alt idle) and fall back to locomotion.
func cancel_one_shot(name: String) -> void:
	if _terminal or _one_shot == "" or _one_shot != name:
		return
	_one_shot = ""
	_motion_interruptible = ""
	if _player != null:
		_player.speed_scale = 1.0
	_play(WALK if _moving else IDLE)


## v513: the named one-shot is cancelled as soon as the monster starts moving.
func mark_motion_interruptible(name: String) -> void:
	_motion_interruptible = name


func current_one_shot() -> String:
	return _one_shot


func is_idle() -> bool:
	return not _terminal and _one_shot == "" and not _moving


func has_clip(name: String) -> bool:
	return _player != null and _player.has_animation(name)


func clip_length(name: String) -> float:
	return _player.get_animation(name).length if has_clip(name) else 0.0


func profile() -> Dictionary:
	return MonsterAnimVariantsScript.profile_of(_player)


func is_terminal() -> bool:
	return _terminal


func current_clip() -> String:
	if _player == null:
		return ""
	return str(_player.current_animation)


func get_debug_state() -> Dictionary:
	return {
		"current_clip": current_clip(),
		"clip_position_s": _player.current_animation_position if _player != null else 0.0,
		"terminal": _terminal,
		"terminal_clip": _terminal_clip,
		"is_moving": _moving,
		"one_shot": _one_shot,
		"speed_scale": _player.speed_scale if _player != null else 1.0,
		"melee_lunge": MeleeLungePresentationScript.get_debug_state(_player),
		"warnings": _warnings,
	}


func _on_finished(name: String) -> void:
	if _terminal:
		return
	if name == _one_shot:
		_one_shot = ""
		if _player != null:
			_player.speed_scale = 1.0
		_play(WALK if _moving else IDLE)


func _play(name: String) -> bool:
	if _player == null:
		return false
	if not _player.has_animation(name):
		_warn({"code": "unknown_clip", "clip": name})
		return false
	_player.play(name)
	if _idle_variation != null and (name == IDLE or name == WALK):
		_idle_variation.on_locomotion_clip(name)
	return true


func _warn(entry: Dictionary) -> void:
	push_warning("[anim] %s" % JSON.stringify(entry))
	_warnings.append(entry)
