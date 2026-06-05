# Interactive client scene (ADR-0001 D3/D4): a thin renderer over the
# authoritative server. The client predicts the player's movement locally and
# reconciles to authoritative snapshots/deltas; the server owns all combat,
# loot, and inventory outcomes. Visuals are placeholder primitives (slice v1).
extends Node3D

const NetClientScript := preload("res://scripts/net_client.gd")

var client: NetClient
var entities: Dictionary = {}        # id (String) -> MeshInstance3D
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

var _send_cooldown: float = 0.0
var _debug_label: Label

const SEND_INTERVAL := 0.1
const PLAYER_SPEED := 4.0


func _ready() -> void:
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
	for e in p.get("entities", []):
		_upsert_entity(e)
	inventory = p.get("inventory", [])
	equipped = p.get("equipped", {})
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
			"inventory_update":
				_update_inventory_item(c["item"])
			"equipped_update":
				equipped[c["slot"]] = c.get("item_instance_id")
	_reconcile_player()


func _upsert_entity(e: Dictionary) -> void:
	var id := str(e["id"])
	var node: MeshInstance3D
	if entities.has(id):
		node = entities[id]
	else:
		node = _make_entity_node(e["type"])
		add_child(node)
		entities[id] = node
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
		if e["type"] == "monster" and not monster_ids.has(id):
			monster_ids.append(id)
	var pos: Dictionary = e["position"]
	var server_pos := Vector3(pos["x"], 0.0, pos["y"])
	if e["type"] == "player":
		player_id = id
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		node.position = server_pos
	else:
		node.position = server_pos


func _remove_entity(id: String) -> void:
	if entities.has(id):
		entities[id].queue_free()
		entities.erase(id)
	loot_ids.erase(id)


func _update_inventory_item(item: Dictionary) -> void:
	for i in range(inventory.size()):
		if inventory[i]["item_instance_id"] == item["item_instance_id"]:
			inventory[i] = item
			return
	inventory.append(item)


func _reconcile_player() -> void:
	if player_id != "" and entities.has(player_id):
		entities[player_id].position = predicted_pos


# --- input + prediction -----------------------------------------------------

func _handle_input(delta: float) -> void:
	if client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	_send_cooldown -= delta

	var dir := Vector2.ZERO
	if Input.is_key_pressed(KEY_W): dir.y -= 1
	if Input.is_key_pressed(KEY_S): dir.y += 1
	if Input.is_key_pressed(KEY_A): dir.x -= 1
	if Input.is_key_pressed(KEY_D): dir.x += 1

	if dir != Vector2.ZERO and _send_cooldown <= 0.0:
		dir = dir.normalized()
		# Local prediction: move immediately for responsive feel.
		predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
		_reconcile_player()
		client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
		_send_cooldown = SEND_INTERVAL

	if Input.is_key_pressed(KEY_SPACE) and monster_ids.size() > 0 and _send_cooldown <= 0.0:
		client.send("attack_intent", last_server_tick, {"target_id": monster_ids[0]})
		_send_cooldown = SEND_INTERVAL
	if Input.is_key_pressed(KEY_E) and loot_ids.size() > 0 and _send_cooldown <= 0.0:
		client.send("pick_up_intent", last_server_tick, {"entity_id": loot_ids[0]})
		_send_cooldown = SEND_INTERVAL
	if Input.is_key_pressed(KEY_Q) and inventory.size() > 0 and _send_cooldown <= 0.0:
		client.send("equip_intent", last_server_tick, {"item_instance_id": inventory[0]["item_instance_id"], "slot": "weapon"})
		_send_cooldown = SEND_INTERVAL


# --- scene construction (placeholder primitives) ----------------------------

func _build_scene() -> void:
	var cam := Camera3D.new()
	cam.projection = Camera3D.PROJECTION_ORTHOGONAL
	cam.size = 20.0
	cam.position = Vector3(20, 20, 20)
	cam.look_at(Vector3(11, 0, 5), Vector3.UP)
	add_child(cam)

	var light := DirectionalLight3D.new()
	light.rotation_degrees = Vector3(-50, -40, 0)
	add_child(light)

	var ui := CanvasLayer.new()
	add_child(ui)
	_debug_label = Label.new()
	_debug_label.position = Vector2(12, 12)
	ui.add_child(_debug_label)


func _make_entity_node(kind: String) -> MeshInstance3D:
	var node := MeshInstance3D.new()
	var mat := StandardMaterial3D.new()
	match kind:
		"player":
			node.mesh = CapsuleMesh.new()
			mat.albedo_color = Color(0.2, 0.6, 1.0)
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
	_debug_label.text = "tick=%d  recon_delta=%.2f\ninv=%d  equipped_weapon=%s\nW/A/S/D move  SPACE attack  E pickup  Q equip" % [
		last_server_tick, reconciliation_delta, inventory.size(), str(eq)]


func _debug(msg: String) -> void:
	print("[client] ", msg)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback
