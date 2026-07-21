## PlayerCameraController — owns the camera rig for all camera modes.
##
## Modes (defined in shared/assets/camera_presentations.v0.json):
##   isometric  — orthographic Camera3D at follow_offset above player.
##   chest_view — perspective Camera3D held at configured eye height above character_visual.
##
## Usage:
##   var ctx := PlayerCameraContext.new()
##   ... populate ctx fields ...
##   _camera_controller = PlayerCameraControllerScript.new()
##   _camera_controller.setup(ctx, self)
##   _camera = _camera_controller.get_gameplay_camera()
class_name PlayerCameraController
extends RefCounted

const CameraPresentationsLoaderScript := preload("res://scripts/camera_presentations_loader.gd")
const CameraImpactFeedbackScript := preload("res://scripts/camera_impact_feedback.gd")
const PlayerCameraContextScript := preload("res://scripts/player_camera_context.gd")

var _ctx  # PlayerCameraContext
var _scene_root: Node3D

var _camera: Camera3D
var _chest_socket_node: Node3D  # non-null when camera is parented to chest socket
var _chest_socket_fallback: Node3D  # non-null when fallback node was created in _setup_chest_view
var _head_socket_node: Node3D  # head bone socket; camera anchors here to track animations
var _eye_view_light: OmniLight3D
var _chest_view_mirror_applied := false
var _current_mode: String = ""
var _cfg: Dictionary = {}  # current mode data from CameraPresentationsLoader
var _iso_offset: Vector3 = Vector3(9.0, 20.0, 15.0)  # isometric follow offset, read from data
var _yaw: float = 0.0    # horizontal orbit angle (radians) for perspective modes
var _pitch: float = 0.0  # vertical orbit angle (radians) for perspective modes
var _follow_initialized := false


## Must be called once before any other method.
## scene_root is the Node3D to attach camera nodes to (typically `self` in main.gd).
func setup(ctx, scene_root: Node3D) -> void:  # ctx: PlayerCameraContext
	_ctx = ctx
	_scene_root = scene_root
	_camera = Camera3D.new()
	_camera.name = "PlayerCamera"
	_camera.current = true
	scene_root.add_child(_camera)
	var initial_mode: String = "isometric"
	if ctx.client_settings != null:
		initial_mode = ctx.client_settings.camera_mode
	apply_mode(initial_mode)


## Switch to a named mode; rebuilds the rig as needed.
func apply_mode(mode: String) -> void:
	if mode == _current_mode:
		return
	_teardown_rig()
	_current_mode = mode
	CameraPresentationsLoaderScript.ensure_loaded()
	_cfg = CameraPresentationsLoaderScript.mode(mode)
	_setup_rig()
	_follow_initialized = false
	sync_to_player()


## Position the camera relative to the player anchor for the current mode.
func sync_to_player() -> void:
	if _camera == null or _ctx == null or _ctx.player_anchor == null:
		return
	match _current_mode:
		"isometric":
			_sync_isometric(0.0, true)
		"chest_view":
			_sync_chest_view()


## Smooth isometric follow; call every frame during gameplay.
func tick_follow(delta: float) -> void:
	if _camera == null or _ctx == null or _ctx.player_anchor == null:
		return
	if _current_mode == "isometric":
		_sync_isometric(delta, false)
	elif _current_mode == "chest_view":
		_sync_chest_view()


## Returns the active camera mode name.
func get_current_mode() -> String:
	return _current_mode


## Zoom handler: isometric adjusts camera.size; chest_view adjusts FOV.
func adjust_zoom(delta_size: float) -> void:
	if _camera == null:
		return
	if _current_mode == "isometric":
		var zoom_min: float = _cfg.get("zoom_min", 8.0)
		var zoom_max: float = _cfg.get("zoom_max", 20.0)
		_camera.size = clampf(_camera.size + delta_size, zoom_min, zoom_max)
	elif _current_mode == "chest_view":
		var fov_min: float = _cfg.get("fov_min", 60.0)
		var fov_max: float = _cfg.get("fov_max", 120.0)
		_camera.fov = clampf(_camera.fov + delta_size, fov_min, fov_max)


## Returns the active Camera3D (callers bind this for raycasting / UI world-space).
func get_gameplay_camera() -> Camera3D:
	return _camera

func get_debug_state() -> Dictionary:
	return {
		"mode": _current_mode,
		"projection": "perspective" if _camera != null and _camera.projection == Camera3D.PROJECTION_PERSPECTIVE else "orthogonal",
		"eye_height": float(_cfg.get("eye_height", 0.0)),
		"near_clip": float(_cfg.get("near_clip", 0.05)),
		"camera_position_y": _camera.global_position.y if _camera != null else 0.0,
		"eye_light_energy": _eye_view_light.light_energy if _eye_view_light != null and is_instance_valid(_eye_view_light) else 0.0,
		"eye_light_range": _eye_view_light.omni_range if _eye_view_light != null and is_instance_valid(_eye_view_light) else 0.0,
	}


## Return [origin, direction] for a ray from the camera through the given screen position.
## Used by callers that need a ray for physics casting (e.g. entity picking).
func mouse_ray(viewport: Viewport) -> Array:
	if _camera == null:
		return [Vector3.ZERO, Vector3.FORWARD]
	var pos := viewport.get_mouse_position()
	return [_camera.project_ray_origin(pos), _camera.project_ray_normal(pos)]


## Project a screen-space mouse position to the ground plane (y=0).
## Falls back to player_anchor position when the ray is nearly horizontal.
func screen_to_ground_point(mouse_pos: Vector2, viewport: Viewport) -> Vector3:
	if _camera == null or _ctx == null or _ctx.player_anchor == null:
		return Vector3.ZERO
	var origin := _camera.project_ray_origin(mouse_pos)
	var normal := _camera.project_ray_normal(mouse_pos)
	if abs(normal.y) < 0.0001:
		return _ctx.player_anchor.global_position
	var t := -origin.y / normal.y
	if t < 0.0:
		return _ctx.player_anchor.global_position
	return origin + normal * t


func center_ground_point(viewport: Viewport) -> Vector3:
	if _camera == null or viewport == null or _ctx == null or _ctx.player_anchor == null:
		return _ctx.player_anchor.global_position if _ctx != null and _ctx.player_anchor != null else Vector3.ZERO
	return screen_to_ground_point(viewport.get_visible_rect().get_center(), viewport)


## Flat aim direction from mouse cursor relative to the player anchor (isometric combat).
## Returns Vector2.ZERO when camera is null or cursor is directly on the anchor.
func aim_direction_from_mouse(viewport: Viewport) -> Vector2:
	if _camera == null or _ctx == null or _ctx.player_anchor == null:
		return Vector2.ZERO
	var ground := screen_to_ground_point(viewport.get_mouse_position(), viewport)
	var flat := Vector2(ground.x - _ctx.player_anchor.global_position.x, ground.z - _ctx.player_anchor.global_position.z)
	if flat.length_squared() < 0.0001:
		return Vector2.ZERO
	return flat.normalized()


## Apply mouse motion delta to orbit yaw/pitch for perspective camera modes.
## delta is the raw InputEventMouseMotion.relative in pixels.
func apply_mouse_motion(delta: Vector2) -> void:
	if _current_mode == "isometric":
		return
	var sensitivity: float = _cfg.get("mouse_sensitivity", 0.003)
	_yaw -= delta.x * sensitivity
	_pitch -= delta.y * sensitivity
	var pitch_min: float = deg_to_rad(_cfg.get("pitch_min_degrees", -60.0))
	var pitch_max: float = deg_to_rad(_cfg.get("pitch_max_degrees", -5.0))
	_pitch = clampf(_pitch, pitch_min, pitch_max)
	_apply_perspective_rotation()


## Convert 2D input (WASD) into a flat world-space direction relative to camera orientation.
func camera_relative_flat_direction(input: Vector2) -> Vector2:
	if _camera == null or input == Vector2.ZERO:
		return Vector2.ZERO
	if _current_mode == "chest_view":
		# character_visual.rotation.y may lag behind _yaw by one frame due to server
		# facing overrides; derive forward/right from _yaw directly to stay frame-accurate.
		var fwd := Vector2(sin(_yaw), cos(_yaw))
		var right := Vector2(-cos(_yaw), sin(_yaw))
		var world := right * input.x - fwd * input.y
		return world.normalized() if world.length_squared() > 0.0001 else Vector2.ZERO
	var forward := -_camera.global_transform.basis.z
	forward.y = 0.0
	if forward.length_squared() < 0.0001:
		return input.normalized()
	forward = forward.normalized()
	var right := _camera.global_transform.basis.x
	right.y = 0.0
	if right.length_squared() < 0.0001:
		return input.normalized()
	right = right.normalized()
	var world := right * input.x - forward * input.y
	return Vector2(world.x, world.z).normalized()


# ---------------------------------------------------------------------------
# Private helpers
# ---------------------------------------------------------------------------

func _setup_rig() -> void:
	match _current_mode:
		"isometric":
			_setup_isometric()
		"chest_view":
			_setup_chest_view()
		_:
			_setup_isometric()


func _reset_perspective_angles() -> void:
	if _current_mode == "chest_view" and _ctx != null and _ctx.character_visual != null:
		_yaw = _ctx.character_visual.rotation.y
	else:
		_yaw = 0.0
	_pitch = deg_to_rad(_cfg.get("pitch_min_degrees", -60.0) * 0.5 + _cfg.get("pitch_max_degrees", -5.0) * 0.5)


func _teardown_rig() -> void:
	if _current_mode == "chest_view":
		_clear_chest_view_mirror()

	# Free chest socket fallback node if one was created.
	if _chest_socket_fallback != null:
		if is_instance_valid(_chest_socket_fallback):
			_chest_socket_fallback.queue_free()
		_chest_socket_fallback = null

	# Restore default near clip when leaving a mode that overrides it.
	if _camera != null and is_instance_valid(_camera):
		_camera.near = 0.05

	# Detach camera from chest socket if it was parented there.
	if _chest_socket_node != null:
		if is_instance_valid(_camera) and _camera.get_parent() == _chest_socket_node:
			_chest_socket_node.remove_child(_camera)
			_scene_root.add_child(_camera)
		_chest_socket_node = null
	if _eye_view_light != null:
		if is_instance_valid(_eye_view_light):
			_eye_view_light.queue_free()
		_eye_view_light = null


func _setup_isometric() -> void:
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	var zoom_default: float = _cfg.get("zoom_default", 12.0)
	_camera.size = zoom_default
	var raw_offset = _cfg.get("follow_offset", [])
	if raw_offset is Array and raw_offset.size() >= 3 and (float(raw_offset[0]) != 0.0 or float(raw_offset[1]) != 0.0 or float(raw_offset[2]) != 0.0):
		_iso_offset = Vector3(float(raw_offset[0]), float(raw_offset[1]), float(raw_offset[2]))
	_camera.position = _iso_offset


func _setup_chest_view() -> void:
	_reset_perspective_angles()
	_camera.projection = Camera3D.PROJECTION_PERSPECTIVE
	_camera.near = float(_cfg.get("near_clip", 0.25))
	_chest_socket_node = _ctx.character_visual if _ctx != null else null
	_head_socket_node = null
	if _ctx != null and _ctx.character_visual != null:
		_head_socket_node = _ctx.character_visual.find_child("head_socket", true, false) as Node3D
	if _camera.get_parent() != _scene_root:
		_camera.get_parent().remove_child(_camera)
		_scene_root.add_child(_camera)
	_setup_eye_view_light()
	_apply_chest_view_mirror()
	_sync_chest_view()


func _sync_isometric(delta: float, snap: bool) -> void:
	var target := _isometric_follow_target()
	var desired: Vector3 = target + _iso_offset + CameraImpactFeedbackScript.get_offset()
	var damping: float = float(_cfg.get("follow_damping_seconds", 0.0))
	if snap or damping <= 0.0 or not _follow_initialized:
		_camera.global_position = desired
		_follow_initialized = true
	else:
		var alpha := 1.0 - exp(-maxf(delta, 0.0) / damping)
		_camera.global_position = _camera.global_position.lerp(desired, alpha)
	_camera.look_at(_camera.global_position - _iso_offset, Vector3.UP)


func _isometric_follow_target() -> Vector3:
	var anchor := _ctx.player_anchor as Node3D
	if anchor == null:
		return Vector3.ZERO
	var target := anchor.global_position
	var visual := _ctx.character_visual as Node3D
	# Follow the rendered hero for routine isometric locomotion so camera damping does not
	# exaggerate the difference between the authoritative anchor and the smoothed visual child.
	if visual != null and is_instance_valid(visual):
		target = visual.global_position
	return target


func _sync_chest_view() -> void:
	if _ctx == null or _ctx.character_visual == null:
		return
	_ctx.character_visual.rotation.y = _yaw
	var fwd_offset: float = _cfg.get("chest_forward_offset", 0.12)
	var yaw_basis := Basis(Vector3.UP, _yaw)
	if _head_socket_node != null and is_instance_valid(_head_socket_node):
		var eye_offset: float = _cfg.get("eye_height_from_socket", 0.05)
		_camera.global_position = _head_socket_node.global_position + yaw_basis * Vector3(0.0, eye_offset, -fwd_offset)
	else:
		var eye_height: float = _cfg.get("eye_height", 0.0)
		_camera.global_position = _ctx.character_visual.global_position + yaw_basis * Vector3(0.0, eye_height, -fwd_offset)
	_camera.global_rotation = Vector3(_pitch, _yaw + PI, 0.0)


func _setup_eye_view_light() -> void:
	if _eye_view_light == null or not is_instance_valid(_eye_view_light):
		_eye_view_light = OmniLight3D.new()
		_eye_view_light.name = "EyeViewLight"
		_camera.add_child(_eye_view_light)
	_eye_view_light.position = Vector3.ZERO
	_eye_view_light.light_energy = float(_cfg.get("eye_light_energy", 0.9))
	_eye_view_light.omni_range = float(_cfg.get("eye_light_range", 4.0))
	var raw_color = _cfg.get("eye_light_color", [1.0, 0.92, 0.78])
	if raw_color is Array and raw_color.size() >= 3:
		_eye_view_light.light_color = Color(float(raw_color[0]), float(raw_color[1]), float(raw_color[2]))


func _apply_perspective_rotation() -> void:
	if _current_mode == "chest_view":
		if _camera != null:
			_camera.rotation = Vector3(_pitch, PI, 0.0)
		if _ctx != null and _ctx.character_visual != null:
			_ctx.character_visual.rotation.y = _yaw


func _apply_chest_view_mirror() -> void:
	if _ctx == null or _ctx.character_visual == null:
		return
	var visual := _ctx.character_visual as Node3D
	var mag := _positive_uniform_scale(visual.scale)
	visual.scale = Vector3(-mag, mag, mag)
	_chest_view_mirror_applied = true


func _clear_chest_view_mirror() -> void:
	if not _chest_view_mirror_applied or _ctx == null or _ctx.character_visual == null:
		_chest_view_mirror_applied = false
		return
	var visual := _ctx.character_visual as Node3D
	var mag := _positive_uniform_scale(visual.scale)
	visual.scale = Vector3(mag, mag, mag)
	_chest_view_mirror_applied = false


func _positive_uniform_scale(scale: Vector3) -> float:
	return maxf(absf(scale.x), maxf(absf(scale.y), absf(scale.z)))
