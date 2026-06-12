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

const IDLE := "idle"
const WALK := "walk"


func _init(player: AnimationPlayer) -> void:
	_player = player
	if _player != null:
		_player.animation_finished.connect(_on_finished)
	_play(IDLE)


func set_locomotion(is_moving: bool) -> void:
	_moving = is_moving
	if _terminal or _one_shot != "":
		return
	_play(WALK if is_moving else IDLE)


func play_one_shot(name: String) -> void:
	if _terminal:
		return
	_one_shot = name
	_play(name)


func enter_terminal(name: String) -> void:
	_terminal = true
	_terminal_clip = name
	_one_shot = ""
	_play(name)


func reset_terminal() -> void:
	_terminal = false
	_terminal_clip = ""
	_one_shot = ""
	_play(WALK if _moving else IDLE)


func is_terminal() -> bool:
	return _terminal


func current_clip() -> String:
	if _player == null:
		return ""
	return str(_player.current_animation)


func get_debug_state() -> Dictionary:
	return {
		"current_clip": current_clip(),
		"terminal": _terminal,
		"terminal_clip": _terminal_clip,
		"is_moving": _moving,
		"warnings": _warnings,
	}


func _on_finished(name: String) -> void:
	if _terminal:
		return
	if name == _one_shot:
		_one_shot = ""
		_play(WALK if _moving else IDLE)


func _play(name: String) -> void:
	if _player == null:
		return
	if not _player.has_animation(name):
		_warn({"code": "unknown_clip", "clip": name})
		return
	_player.play(name)


func _warn(entry: Dictionary) -> void:
	push_warning("[anim] %s" % JSON.stringify(entry))
	_warnings.append(entry)
