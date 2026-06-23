## PlayerCameraController — owns the camera rig for all three camera modes.
##
## Modes (defined in shared/assets/camera_presentations.v0.json):
##   isometric   — orthographic Camera3D at follow_offset above player.
##   third_person — SpringArm3D + perspective Camera3D with shoulder offset.
##   chest_view  — perspective Camera3D parented to chest_socket on character_visual.
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

# Isometric follow offset (matches legacy ClientConstants.CAMERA_FOLLOW_OFFSET).
const _ISO_OFFSET := Vector3(9.0, 20.0, 15.0)

var _ctx  # PlayerCameraContext
var _scene_root: Node3D

var _camera: Camera3D
var _spring_arm: SpringArm3D  # used only in third_person
var _chest_socket_node: Node3D  # non-null when camera is parented to chest socket
var _current_mode: String = ""
var _cfg: Dictionary = {}  # current mode data from CameraPresentationsLoader


## Must be called once before any other method.
## scene_root is the Node3D to attach camera nodes to (typically `self` in main.gd).
func setup(ctx, scene_root: Node3D) -> void:  # ctx: PlayerCameraContext
	_ctx = ctx
	_scene_root = scene_root
	_camera = Camera3D.new()
	_camera.name = "PlayerCamera"
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
	sync_to_player()


## Position the camera relative to the player anchor for the current mode.
func sync_to_player() -> void:
	if _camera == null or _ctx == null or _ctx.player_anchor == null:
		return
	match _current_mode:
		"isometric":
			_sync_isometric()
		"third_person":
			_sync_third_person()
		"chest_view":
			pass  # camera is a child of chest_socket; positioning handled by the scene graph


## Zoom handler: isometric adjusts camera.size; third_person adjusts spring arm length.
func adjust_zoom(delta_size: float) -> void:
	if _camera == null:
		return
	match _current_mode:
		"isometric":
			var zoom_min: float = _cfg.get("zoom_min", 8.0)
			var zoom_max: float = _cfg.get("zoom_max", 20.0)
			_camera.size = clampf(_camera.size + delta_size, zoom_min, zoom_max)
		"third_person":
			if _spring_arm != null:
				var zoom_min: float = _cfg.get("zoom_min", 3.0)
				var zoom_max: float = _cfg.get("zoom_max", 10.0)
				_spring_arm.spring_length = clampf(_spring_arm.spring_length + delta_size, zoom_min, zoom_max)


## Returns the active Camera3D (callers bind this for raycasting / UI world-space).
func get_gameplay_camera() -> Camera3D:
	return _camera


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


## Convert 2D input (WASD) into a flat world-space direction relative to camera orientation.
func camera_relative_flat_direction(input: Vector2) -> Vector2:
	if _camera == null or input == Vector2.ZERO:
		return Vector2.ZERO
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
		"third_person":
			_setup_third_person()
		"chest_view":
			_setup_chest_view()
		_:
			_setup_isometric()


func _teardown_rig() -> void:
	# Remove spring arm (and its child camera) if present from previous mode.
	if _spring_arm != null:
		if is_instance_valid(_spring_arm) and _spring_arm.is_inside_tree():
			_spring_arm.get_parent().remove_child(_spring_arm)
			_spring_arm.queue_free()
		_spring_arm = null

	# Detach camera from chest socket if it was parented there.
	if _chest_socket_node != null:
		if is_instance_valid(_camera) and _camera.get_parent() == _chest_socket_node:
			_chest_socket_node.remove_child(_camera)
			_scene_root.add_child(_camera)
		_chest_socket_node = null

	# Re-parent camera back to scene root if it ended up elsewhere.
	if is_instance_valid(_camera) and _camera.get_parent() != _scene_root:
		var p := _camera.get_parent()
		if p != null:
			p.remove_child(_camera)
		_scene_root.add_child(_camera)


func _setup_isometric() -> void:
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	var zoom_default: float = _cfg.get("zoom_default", 12.0)
	_camera.size = zoom_default
	_camera.position = _ISO_OFFSET


func _setup_third_person() -> void:
	_camera.projection = Camera3D.PROJECTION_PERSPECTIVE
	var arm_length: float = _cfg.get("spring_arm_length", 6.0)
	_spring_arm = SpringArm3D.new()
	_spring_arm.name = "CameraSpringArm"
	_spring_arm.spring_length = arm_length
	# Collision against environment layer (layer 1 — dungeon walls).
	_spring_arm.collision_mask = 1
	_scene_root.add_child(_spring_arm)
	# Move camera under the spring arm.
	if _camera.get_parent() == _scene_root:
		_scene_root.remove_child(_camera)
	_spring_arm.add_child(_camera)
	# Apply shoulder offset.
	var raw_offset = _cfg.get("shoulder_offset", [0.0, 0.0, 0.0])
	if raw_offset is Array and raw_offset.size() >= 3:
		_camera.position = Vector3(raw_offset[0], raw_offset[1], raw_offset[2])
	else:
		_camera.position = Vector3.ZERO


func _setup_chest_view() -> void:
	_camera.projection = Camera3D.PROJECTION_PERSPECTIVE
	# Try to find the chest socket on character_visual.
	var socket: Node3D = null
	if _ctx != null and _ctx.character_visual != null:
		socket = _ctx.character_visual.find_child("chest_socket", true, false) as Node3D
	if socket == null:
		# Fallback: create a temporary node at the known fallback position.
		var fallback := Node3D.new()
		fallback.name = "chest_socket_fallback"
		fallback.position = Vector3(0.0, 1.08, 0.0)
		if _ctx != null and _ctx.character_visual != null:
			_ctx.character_visual.add_child(fallback)
		else:
			_scene_root.add_child(fallback)
		socket = fallback
	_chest_socket_node = socket
	# Reparent camera to the socket.
	if _camera.get_parent() == _scene_root:
		_scene_root.remove_child(_camera)
	socket.add_child(_camera)
	var fwd_offset: float = _cfg.get("chest_forward_offset", 0.3)
	_camera.position = Vector3(0.0, 0.0, -fwd_offset)
	_camera.rotation_degrees = Vector3.ZERO


func _sync_isometric() -> void:
	var anchor := _ctx.player_anchor as Node3D
	var target: Vector3 = anchor.global_position if anchor != null else Vector3.ZERO
	_camera.global_position = target + _ISO_OFFSET + CameraImpactFeedbackScript.get_offset()
	_camera.look_at(target, Vector3.UP)


func _sync_third_person() -> void:
	if _spring_arm == null:
		return
	var anchor := _ctx.player_anchor as Node3D
	var target: Vector3 = anchor.global_position if anchor != null else Vector3.ZERO
	# Position the spring arm above the player with shake applied.
	_spring_arm.global_position = target + Vector3(0.0, 1.5, 0.0) + CameraImpactFeedbackScript.get_offset()
