# Interactive client scene (ADR-0001 D3/D4): a thin renderer over the
# authoritative server. The client predicts the player's movement locally and
# reconciles to authoritative snapshots/deltas; the server owns all combat,
# loot, and inventory outcomes. Visuals are placeholder primitives (slice v1).
extends Node3D

const NetClientScript := preload("res://scripts/net_client.gd")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")

var client: NetClient
var resolver: EquipmentVisualResolver
var entities: Dictionary = {}        # id (String) -> MeshInstance3D (monsters/loot only)
var player_id: String = ""
var predicted_pos := Vector3.ZERO    # client-predicted player position
var reconciliation_delta: float = 0.0
var last_server_tick: int = 0
var inventory: Array = []
var equipped: Dictionary = {}
var loot_ids: Array = []
var monster_ids: Array = []
var ready_sent: bool = false
var item_to_equip: String = ""

# Slice v2 scene graph (spec §5.1): the local player is a humanoid under a
# PlayerAnchor that follows authoritative position; monsters/loot live under
# Entities. These are defined in main.tscn and cached on ready.
var player_anchor: Node3D
var character_visual: Node3D
var entities_root: Node3D

var _send_cooldown: float = 0.0
var _attack_cooldown: float = 0.0
var _debug_label: Label
var _camera: Camera3D

const SEND_INTERVAL := 0.1
const PLAYER_SPEED := 4.0
const ATTACK_AIM_MIN_DOT := 0.0  # monster must be in the forward half-space
const CAMERA_ZOOM_DEFAULT := 20.0
const CAMERA_ZOOM_STEP := 1.5
const CAMERA_ZOOM_MIN := 8.0
const CAMERA_ZOOM_MAX := 40.0


func _ready() -> void:
	player_anchor = $World/PlayerAnchor
	character_visual = $World/PlayerAnchor/CharacterVisual
	entities_root = $Entities
	# Mount-root is injected (spec §4.8): the resolver finds the named socket
	# within CharacterVisual, never via an absolute scene path.
	resolver = EquipmentResolverScript.new(character_visual)
	_build_scene()
	var base_url := _env("ARPG_BASE_URL", "http://localhost:8080")
	var dev_token := _env("ARPG_DEV_TOKEN", "local-dev-token")

	client = NetClientScript.new(base_url)
	if not client.login(_env("ARPG_EMAIL", "client@example.test"), dev_token):
		_debug("login failed")
		return
	if not client.create_session():
		_debug("session failed")
		return
	predicted_pos = Vector3.ZERO
	client.connect_ws()
	_debug("connecting session %s" % client.session_id)


func _process(delta: float) -> void:
	if client == null:
		return

	var state := client.ready_state()
	if state == WebSocketPeer.STATE_OPEN and not ready_sent:
		client.send("client_ready", last_server_tick, {"client_version": "godot", "last_seen_tick": last_server_tick})
		ready_sent = true

	for env in client.poll():
		_handle_message(env)

	_handle_input(delta)
	_update_facing_toward_mouse()
	_update_debug()


# --- message handling -------------------------------------------------------

func _handle_message(env: Dictionary) -> void:
	last_server_tick = max(last_server_tick, int(env.get("tick", 0)))
	match env.get("type", ""):
		"session_snapshot":
			_apply_snapshot(env["payload"])
		"state_delta":
			_apply_delta(env["payload"])
		"intent_rejected":
			_debug("rejected: %s" % env["payload"].get("reason", "?"))
		"error":
			_debug("error: %s" % env["payload"].get("message", "?"))


func _apply_snapshot(p: Dictionary) -> void:
	for id in entities.keys():
		entities[id].queue_free()
	entities.clear()
	loot_ids.clear()
	monster_ids.clear()
	# (player is the PlayerAnchor/CharacterVisual, not a per-snapshot entity node)
	for e in p.get("entities", []):
		_upsert_entity(e)
	inventory = p.get("inventory", [])
	equipped = p.get("equipped", {})
	if resolver != null:
		resolver.apply_snapshot(p)
	_reconcile_player()


func _apply_delta(p: Dictionary) -> void:
	for c in p.get("changes", []):
		match c.get("op", ""):
			"entity_spawn", "entity_update":
				_upsert_entity(c["entity"])
			"entity_remove":
				_remove_entity(c["entity_id"])
			"inventory_add":
				inventory.append(c["item"])
				if resolver != null:
					resolver.ingest_inventory_item(c["item"])
			"inventory_update":
				_update_inventory_item(c["item"])
				if resolver != null:
					resolver.ingest_inventory_item(c["item"])
			"equipped_update":
				equipped[c["slot"]] = c.get("item_instance_id")
				if resolver != null:
					resolver.apply_equipped_update(c["slot"], c.get("item_instance_id"))
	_reconcile_player()


func _upsert_entity(e: Dictionary) -> void:
	var id := str(e["id"])
	var pos: Dictionary = e["position"]
	var server_pos := Vector3(pos["x"], 0.0, pos["y"])
	if e["type"] == "player":
		# The player is the humanoid under PlayerAnchor, not an entity-dict node.
		player_id = id
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		player_anchor.position = server_pos
		return
	var node: MeshInstance3D
	if entities.has(id):
		node = entities[id]
	else:
		node = _make_entity_node(e["type"])
		entities_root.add_child(node)
		entities[id] = node
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
		if e["type"] == "monster" and not monster_ids.has(id):
			monster_ids.append(id)
	node.position = server_pos


func _remove_entity(id: String) -> void:
	if entities.has(id):
		entities[id].queue_free()
		entities.erase(id)
	loot_ids.erase(id)
	monster_ids.erase(id)


func _update_inventory_item(item: Dictionary) -> void:
	for i in range(inventory.size()):
		if inventory[i]["item_instance_id"] == item["item_instance_id"]:
			inventory[i] = item
			return
	inventory.append(item)


func _reconcile_player() -> void:
	if player_anchor != null:
		player_anchor.position = predicted_pos


# --- input + prediction -----------------------------------------------------

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.pressed:
		match event.button_index:
			MOUSE_BUTTON_LEFT:
				if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN:
					_try_attack_toward_mouse()
			MOUSE_BUTTON_WHEEL_UP:
				_adjust_camera_zoom(-CAMERA_ZOOM_STEP)
			MOUSE_BUTTON_WHEEL_DOWN:
				_adjust_camera_zoom(CAMERA_ZOOM_STEP)


func _handle_input(delta: float) -> void:
	if client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	_send_cooldown -= delta
	_attack_cooldown -= delta

	var input := Vector2.ZERO
	if Input.is_key_pressed(KEY_W): input.y -= 1
	if Input.is_key_pressed(KEY_S): input.y += 1
	if Input.is_key_pressed(KEY_A): input.x -= 1
	if Input.is_key_pressed(KEY_D): input.x += 1

	if input != Vector2.ZERO and _send_cooldown <= 0.0:
		var dir := _camera_relative_flat_direction(input)
		# Local prediction: move immediately for responsive feel.
		predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
		_reconcile_player()
		client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
		_send_cooldown = SEND_INTERVAL

	if Input.is_key_pressed(KEY_E) and loot_ids.size() > 0 and _send_cooldown <= 0.0:
		client.send("pick_up_intent", last_server_tick, {"entity_id": loot_ids[0]})
		_send_cooldown = SEND_INTERVAL
	if Input.is_key_pressed(KEY_Q) and inventory.size() > 0 and _send_cooldown <= 0.0:
		client.send("equip_intent", last_server_tick, {"item_instance_id": inventory[0]["item_instance_id"], "slot": "weapon"})
		_send_cooldown = SEND_INTERVAL


func _try_attack_toward_mouse() -> void:
	if _attack_cooldown > 0.0:
		return

	var aim := _aim_direction_from_mouse()
	if aim == Vector2.ZERO:
		return

	_face_direction(aim)
	if resolver != null:
		resolver.play_attack_swing()

	var target_id := _best_monster_in_direction(aim)
	if target_id == "":
		_attack_cooldown = SEND_INTERVAL
		return

	client.send("attack_intent", last_server_tick, {"target_id": target_id})
	_attack_cooldown = SEND_INTERVAL


func _update_facing_toward_mouse() -> void:
	var aim := _aim_direction_from_mouse()
	if aim != Vector2.ZERO:
		_face_direction(aim)


func _face_direction(flat_dir: Vector2) -> void:
	if character_visual == null or player_anchor == null:
		return

	var target := player_anchor.global_position + Vector3(flat_dir.x, 0.0, flat_dir.y)
	character_visual.look_at(target, Vector3.UP)


func _camera_relative_flat_direction(input: Vector2) -> Vector2:
	# WASD is screen-relative under the isometric camera, not world X/Z.
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


func _aim_direction_from_mouse() -> Vector2:
	if _camera == null or player_anchor == null:
		return Vector2.ZERO

	var ground := _mouse_ground_point()
	var flat := Vector2(ground.x - player_anchor.global_position.x, ground.z - player_anchor.global_position.z)
	if flat.length_squared() < 0.0001:
		return Vector2.ZERO

	return flat.normalized()


func _mouse_ground_point() -> Vector3:
	var mouse_pos := get_viewport().get_mouse_position()
	var origin := _camera.project_ray_origin(mouse_pos)
	var normal := _camera.project_ray_normal(mouse_pos)
	if abs(normal.y) < 0.0001:
		return player_anchor.global_position

	var t := -origin.y / normal.y
	if t < 0.0:
		return player_anchor.global_position

	return origin + normal * t


func _best_monster_in_direction(aim: Vector2) -> String:
	var best_id := ""
	var best_dot := ATTACK_AIM_MIN_DOT

	for id in monster_ids:
		if not entities.has(id):
			continue

		var entity_node: MeshInstance3D = entities[id]
		var to_monster: Vector3 = entity_node.position - predicted_pos
		var flat := Vector2(to_monster.x, to_monster.z)
		if flat.length_squared() < 0.0001:
			return id

		var dot := flat.normalized().dot(aim)
		if dot > best_dot:
			best_dot = dot
			best_id = id

	return best_id


func _adjust_camera_zoom(delta_size: float) -> void:
	if _camera == null:
		return

	_camera.size = clampf(_camera.size + delta_size, CAMERA_ZOOM_MIN, CAMERA_ZOOM_MAX)


# --- scene construction (placeholder primitives) ----------------------------

func _build_scene() -> void:
	_camera = Camera3D.new()
	_camera.projection = Camera3D.PROJECTION_ORTHOGONAL
	_camera.size = CAMERA_ZOOM_DEFAULT
	_camera.position = Vector3(20, 20, 20)
	add_child(_camera)
	# look_at requires the node to be inside the scene tree (Godot 4).
	_camera.look_at(Vector3(11, 0, 5), Vector3.UP)

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-50, -40, 0)
	add_child(light)

	var ui := CanvasLayer.new()
	add_child(ui)
	_debug_label = Label.new()
	_debug_label.position = Vector2(12, 12)
	ui.add_child(_debug_label)


func _make_entity_node(kind: String) -> MeshInstance3D:
	# Monsters and loot stay v0 placeholder primitives (spec §5.1); only the
	# local player adopts the humanoid GLB pipeline in v2.
	var node := MeshInstance3D.new()
	var mat := StandardMaterial3D.new()
	match kind:
		"monster":
			node.mesh = BoxMesh.new()
			mat.albedo_color = Color(1.0, 0.3, 0.3)
		"loot":
			var box := BoxMesh.new()
			box.size = Vector3(0.5, 0.5, 0.5)
			node.mesh = box
			mat.albedo_color = Color(1.0, 0.85, 0.2)
	node.material_override = mat
	return node


# --- debug ------------------------------------------------------------------

func _update_debug() -> void:
	var eq = equipped.get("weapon", null)
	var ws_state := "?"
	if client != null:
		match client.ready_state():
			WebSocketPeer.STATE_CONNECTING: ws_state = "connecting"
			WebSocketPeer.STATE_OPEN: ws_state = "open"
			WebSocketPeer.STATE_CLOSING: ws_state = "closing"
			WebSocketPeer.STATE_CLOSED: ws_state = "closed"
	var weapon_vis := "none"
	if resolver != null:
		var w = resolver.get_debug_state()["equipped_visuals"]["weapon"]
		if w != null:
			weapon_vis = "%s(visible=%s)" % [w["asset_id"], w["visible"]]
	_debug_label.text = "ws=%s  tick=%d  recon_delta=%.2f\ninv=%d  entities=%d  equipped_weapon=%s\nweapon_visual=%s\nW/A/S/D move  LMB attack  scroll zoom  E pickup  Q equip" % [
		ws_state, last_server_tick, reconciliation_delta, inventory.size(), entities.size(), str(eq), weapon_vis]


func _debug(msg: String) -> void:
	print("[client] ", msg)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback
