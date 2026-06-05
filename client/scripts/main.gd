# Interactive client scene (ADR-0001 D3/D4): a thin renderer over the
# authoritative server. The client predicts the player's movement locally and
# reconciles to authoritative snapshots/deltas; the server owns all combat,
# loot, and inventory outcomes. Visuals are placeholder primitives (slice v1).
extends Node3D

const NetClientScript := preload("res://scripts/net_client.gd")
const EquipmentResolverScript := preload("res://scripts/equipment_visuals.gd")
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const DamageNumberScript := preload("res://scripts/damage_number.gd")
const MonsterHealthBarScript := preload("res://scripts/monster_health_bar.gd")
const MonsterDummyScene := preload("res://scenes/monster_dummy.tscn")
const MONSTER_EVENT_CLIPS := {
	"monster_damaged": "hit",
	"monster_killed": "death",
}
const PLAYER_EVENT_CLIPS := {
	"player_damaged": "hit",
	"player_killed": "death",
}
const PLAYER_START_HP := 10

var client: NetClient
var resolver: EquipmentVisualResolver
var player_anim: AnimationController
var entities: Dictionary = {}        # id (String) -> {node:Node3D, controller:AnimationController|null, type:String}
var player_id: String = ""
var player_hp: int = PLAYER_START_HP
var predicted_pos := Vector3.ZERO    # client-predicted player position
var reconciliation_delta: float = 0.0
var last_server_tick: int = 0
var inventory: Array = []
var equipped: Dictionary = {}
var loot_ids: Array = []
var monster_ids: Array = []
var interactable_ids: Array = []
var ready_sent: bool = false
var item_to_equip: String = ""
var autoplay_enabled: bool = false
var autoplay_phase: String = "idle"
var autoplay_timer: float = 0.0
var autoplay_attack_cooldown: float = 0.0
var autoplay_move_sent: bool = false
var autoplay_pickup_sent: bool = false
var autoplay_equip_sent: bool = false
var autoplay_step_delay: float = 0.35
var visual_replay_enabled: bool = false
var visual_replay_manifest_path: String = ""
var visual_replay_scenarios: Array = []
var visual_replay_index: int = -1
var visual_replay_envelopes: Array = []
var visual_replay_envelope_index: int = 0
var visual_replay_timer: float = 0.0
var visual_replay_title: String = ""
var visual_replay_debug_token: String = ""
var visual_replay_exit_on_complete: bool = false
var visual_replay_exit_requested: bool = false
var visual_replay_exit_timer: float = 0.0

# Slice v2 scene graph (spec §5.1): the local player is a humanoid under a
# PlayerAnchor that follows authoritative position; monsters/loot live under
# Entities. These are defined in main.tscn and cached on ready.
var player_anchor: Node3D
var character_visual: Node3D
var entities_root: Node3D
var damage_numbers_layer: CanvasLayer
var health_bars_layer: CanvasLayer
var monster_health_bars: Dictionary = {} # id (String) -> MonsterHealthBar
var walls_root: Node3D

var _send_cooldown: float = 0.0
var _attack_cooldown: float = 0.0
var _debug_label: Label
var _camera: Camera3D

const SEND_INTERVAL := 0.1
const PLAYER_SPEED := 4.0
const CAMERA_ZOOM_DEFAULT := 20.0
const CAMERA_ZOOM_STEP := 1.5
const CAMERA_ZOOM_MIN := 8.0
const CAMERA_ZOOM_MAX := 40.0
const PROJECTILE_LERP_SECONDS := 0.10


func _ready() -> void:
	player_anchor = $World/PlayerAnchor
	character_visual = $World/PlayerAnchor/CharacterVisual
	entities_root = $Entities
	# Mount-root is injected (spec §4.8): the resolver finds the named socket
	# within CharacterVisual, never via an absolute scene path.
	resolver = EquipmentResolverScript.new(character_visual)
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)
	_build_scene()
	var base_url := _env("ARPG_BASE_URL", "http://localhost:8080")
	var dev_token := _env("ARPG_DEV_TOKEN", "local-dev-token")

	client = NetClientScript.new(base_url)
	if not client.login(_env("ARPG_EMAIL", "client@example.test"), dev_token):
		_debug("login failed")
		return
	visual_replay_manifest_path = _env("ARPG_VISUAL_REPLAY_MANIFEST", "")
	visual_replay_enabled = visual_replay_manifest_path != ""
	if visual_replay_enabled:
		visual_replay_debug_token = _env("ARPG_DEBUG_TOKEN", "local-debug-token")
		visual_replay_exit_on_complete = _truthy_text(_env("ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE", "1"))
		autoplay_step_delay = maxf(0.05, float(_env("ARPG_AUTOPLAY_STEP_DELAY", "0.35")))
		if not _load_visual_replay_manifest(visual_replay_manifest_path):
			_debug("visual replay manifest failed: %s" % visual_replay_manifest_path)
			return
		_debug("visual replay playlist loaded: %d scenario(s)" % visual_replay_scenarios.size())
		_start_next_visual_replay()
		return
	var resume_session_id := _env("ARPG_SESSION_ID", "")
	var requested_world_id := _env("ARPG_WORLD_ID", "")
	if not client.create_session(resume_session_id, requested_world_id):
		_debug("session failed")
		return
	_render_world_walls(client.world_id)
	autoplay_enabled = _truthy_env("ARPG_AUTOPLAY")
	if autoplay_enabled:
		autoplay_phase = "move"
		autoplay_step_delay = maxf(0.05, float(_env("ARPG_AUTOPLAY_STEP_DELAY", "0.35")))
		_debug("visual bot enabled for session %s" % client.session_id)
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

	if visual_replay_enabled:
		_handle_visual_replay(delta)
	elif autoplay_enabled:
		_handle_autoplay(delta)
	else:
		_handle_input(delta)
	if player_anim != null:
		var moving := client.ready_state() == WebSocketPeer.STATE_OPEN \
			and player_hp > 0 \
			and not _input_locked() \
			and (Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_A) \
			or Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_D))
		player_anim.set_locomotion(moving)
	if not _input_locked():
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
		(entities[id]["node"] as Node3D).queue_free()
	entities.clear()
	for id in monster_health_bars.keys():
		var bar = monster_health_bars[id]
		if is_instance_valid(bar):
			bar.queue_free()
	monster_health_bars.clear()
	loot_ids.clear()
	monster_ids.clear()
	interactable_ids.clear()
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
	for ev in p.get("events", []):
		var eid := str(ev.get("entity_id", ""))
		var event_type := str(ev.get("event_type", ""))
		if eid == player_id:
			var player_clip = PLAYER_EVENT_CLIPS.get(event_type, null)
			if player_clip == null or player_anim == null:
				continue
			if player_clip == "death":
				player_anim.enter_terminal("death")
			else:
				player_anim.play_one_shot(player_clip)
			continue
		if event_type == "interactable_activated" and entities.has(eid):
			_set_interactable_state(eid, entities[eid], "open")
			continue
		var clip = MONSTER_EVENT_CLIPS.get(event_type, null)
		if clip == null:
			continue
		if event_type == "monster_damaged" or event_type == "monster_killed":
			_show_damage_number(eid, Color(1.0, 0.92, 0.25), ev.get("damage", null))
		if autoplay_enabled and event_type == "monster_killed":
			autoplay_phase = "pickup"
			autoplay_timer = autoplay_step_delay
		if not entities.has(eid):
			continue
		var ctrl = entities[eid]["controller"]
		if ctrl == null:
			continue
		if clip == "death":
			ctrl.enter_terminal("death")
		else:
			ctrl.play_one_shot(clip)
	_reconcile_player()


func _upsert_entity(e: Dictionary) -> void:
	var id := str(e["id"])
	var pos: Dictionary = e["position"]
	var server_pos := Vector3(pos["x"], 0.0, pos["y"])
	if e["type"] == "player":
		# The player is the humanoid under PlayerAnchor, not an entity-dict node.
		player_id = id
		if e.has("hp"):
			player_hp = int(e["hp"])
			if player_hp <= 0 and player_anim != null:
				player_anim.enter_terminal("death")
		reconciliation_delta = predicted_pos.distance_to(server_pos)
		# Reconcile: snap prediction back toward authoritative truth.
		predicted_pos = server_pos
		player_anchor.position = server_pos
		return
	var rec: Dictionary
	var is_new := false
	if entities.has(id):
		rec = entities[id]
	else:
		is_new = true
		var node := _make_entity_node(e["type"])
		entities_root.add_child(node)
		var controller: AnimationController = null
		if e["type"] == "monster":
			var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
			if ap != null:
				controller = AnimationControllerScript.new(ap)
			else:
				push_warning("[main] monster %s has no AnimationPlayer" % id)
		rec = {"node": node, "controller": controller, "type": str(e["type"])}
		entities[id] = rec
		if e["type"] != "projectile":
			_attach_pick_collider(node, id, str(e["type"]))
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
		if e["type"] == "monster" and not monster_ids.has(id):
			monster_ids.append(id)
		if e["type"] == "interactable" and not interactable_ids.has(id):
			interactable_ids.append(id)
	if rec["type"] == "projectile":
		if is_new:
			(rec["node"] as Node3D).position = server_pos
			rec["last_server_pos"] = server_pos
			return
		_move_projectile_node(rec, server_pos)
	else:
		(rec["node"] as Node3D).position = server_pos
	if rec["type"] == "interactable":
		var state := str(e.get("state", rec.get("state", "closed")))
		_set_interactable_state(id, rec, state)
	# Resume/snapshot consistency: a monster already dead in the snapshot enters
	# the terminal death pose without waiting for an event (spec §5.4).
	if rec["type"] == "monster" and rec["controller"] != null:
		var hp = e.get("hp", null)
		var max_hp = e.get("max_hp", null)
		if hp != null and max_hp != null:
			_upsert_monster_health_bar(id, rec["node"] as Node3D, int(hp), int(max_hp))
		if hp != null and int(hp) <= 0:
			rec["controller"].enter_terminal("death")


func _remove_entity(id: String) -> void:
	if entities.has(id):
		var rec: Dictionary = entities[id]
		if rec.has("move_tween"):
			var tween = rec["move_tween"]
			if is_instance_valid(tween):
				tween.kill()
		(entities[id]["node"] as Node3D).queue_free()
		entities.erase(id)
	if monster_health_bars.has(id):
		var bar = monster_health_bars[id]
		if is_instance_valid(bar):
			bar.queue_free()
		monster_health_bars.erase(id)
	loot_ids.erase(id)
	monster_ids.erase(id)
	interactable_ids.erase(id)


func _update_inventory_item(item: Dictionary) -> void:
	for i in range(inventory.size()):
		if inventory[i]["item_instance_id"] == item["item_instance_id"]:
			inventory[i] = item
			return
	inventory.append(item)


func _reconcile_player() -> void:
	if player_anchor != null:
		player_anchor.position = predicted_pos


func _show_damage_number(entity_id: String, color: Color, event_damage = null) -> void:
	if damage_numbers_layer == null or _camera == null:
		return

	if event_damage == null:
		return
	var amount := int(event_damage)

	var target: Node3D = null
	var world_position := Vector3.ZERO
	if entity_id == player_id:
		target = player_anchor
		world_position = player_anchor.global_position
	elif entities.has(entity_id):
		target = entities[entity_id]["node"] as Node3D
		world_position = target.global_position
	else:
		return

	var pop := DamageNumberScript.new() as DamageNumber
	damage_numbers_layer.add_child(pop)
	var side := -1.0 if entity_id == player_id else 1.0
	pop.setup(_camera, target, world_position, amount, color, side)


func _upsert_monster_health_bar(entity_id: String, target: Node3D, hp: int, max_hp: int) -> void:
	if health_bars_layer == null or _camera == null or target == null:
		return
	if monster_health_bars.has(entity_id):
		(monster_health_bars[entity_id] as MonsterHealthBar).update_hp(hp, max_hp)
		return
	var bar := MonsterHealthBarScript.new() as MonsterHealthBar
	health_bars_layer.add_child(bar)
	bar.setup(_camera, target, hp, max_hp)
	monster_health_bars[entity_id] = bar


# --- input + prediction -----------------------------------------------------

func _unhandled_input(event: InputEvent) -> void:
	if _input_locked():
		return
	if event is InputEventMouseButton and event.pressed:
		match event.button_index:
			MOUSE_BUTTON_LEFT:
				if client != null and client.ready_state() == WebSocketPeer.STATE_OPEN and player_hp > 0:
					_try_action_at_mouse()
			MOUSE_BUTTON_WHEEL_UP:
				_adjust_camera_zoom(-CAMERA_ZOOM_STEP)
			MOUSE_BUTTON_WHEEL_DOWN:
				_adjust_camera_zoom(CAMERA_ZOOM_STEP)


func _handle_input(delta: float) -> void:
	if _input_locked() or client.ready_state() != WebSocketPeer.STATE_OPEN:
		return
	_send_cooldown -= delta
	_attack_cooldown -= delta
	if player_hp <= 0:
		return

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

	if Input.is_key_pressed(KEY_Q) and inventory.size() > 0 and _send_cooldown <= 0.0:
		client.send("equip_intent", last_server_tick, {"item_instance_id": inventory[0]["item_instance_id"], "slot": "weapon"})
		_send_cooldown = SEND_INTERVAL


func _input_locked() -> bool:
	return visual_replay_enabled or autoplay_enabled


func _handle_autoplay(delta: float) -> void:
	if client.ready_state() != WebSocketPeer.STATE_OPEN or player_hp <= 0:
		return
	autoplay_timer -= delta
	autoplay_attack_cooldown -= delta
	if autoplay_timer > 0.0:
		return

	match autoplay_phase:
		"move":
			if not autoplay_move_sent:
				var dir := Vector2(1, 0)
				predicted_pos += Vector3(dir.x, 0, dir.y) * PLAYER_SPEED * SEND_INTERVAL
				_reconcile_player()
				client.send("move_intent", last_server_tick, {"direction": {"x": dir.x, "y": dir.y}, "duration_ticks": 2})
				autoplay_move_sent = true
				autoplay_timer = autoplay_step_delay
				return
			autoplay_phase = "attack"
		"attack":
			if monster_ids.is_empty():
				return
			var target_id := str(monster_ids[0])
			if not entities.has(target_id):
				return
			var rec: Dictionary = entities[target_id]
			var target_node := rec["node"] as Node3D
			if target_node == null:
				return
			var to_target := target_node.position - predicted_pos
			var aim := Vector2(to_target.x, to_target.z).normalized()
			if aim != Vector2.ZERO:
				_face_direction(aim)
			if player_anim != null:
				player_anim.play_one_shot("attack")
			if autoplay_attack_cooldown <= 0.0:
				client.send("action_intent", last_server_tick, {"target_id": target_id})
				autoplay_attack_cooldown = autoplay_step_delay
			autoplay_timer = autoplay_step_delay
		"pickup":
			if not autoplay_pickup_sent and loot_ids.size() > 0:
				client.send("action_intent", last_server_tick, {"target_id": loot_ids[0]})
				autoplay_pickup_sent = true
				autoplay_timer = autoplay_step_delay
				return
			if autoplay_pickup_sent and inventory.size() > 0:
				autoplay_phase = "equip"
			else:
				autoplay_timer = autoplay_step_delay
		"equip":
			if not autoplay_equip_sent and inventory.size() > 0:
				client.send("equip_intent", last_server_tick, {"item_instance_id": inventory[0]["item_instance_id"], "slot": "weapon"})
				autoplay_equip_sent = true
				autoplay_timer = autoplay_step_delay
				return
			var weapon_id = equipped.get("weapon", null)
			if weapon_id != null:
				autoplay_phase = "done"
				_debug("visual bot complete: equipped weapon %s, player_hp=%d" % [str(weapon_id), player_hp])
		"done":
			return


func _try_action_at_mouse() -> void:
	if _attack_cooldown > 0.0 or player_hp <= 0:
		return

	var target_id := _pick_entity_at_mouse()
	if target_id == "" or not entities.has(target_id):
		var ground := _mouse_ground_point()
		client.send("move_to_intent", last_server_tick, {"position": {"x": ground.x, "y": ground.z}})
		_attack_cooldown = SEND_INTERVAL
		return

	var rec: Dictionary = entities[target_id]
	var target_node := rec["node"] as Node3D
	var flat := Vector2(target_node.global_position.x - player_anchor.global_position.x, target_node.global_position.z - player_anchor.global_position.z)
	if flat.length_squared() > 0.0001:
		_face_direction(flat.normalized())

	var typ := str(rec.get("type", ""))
	var state := str(rec.get("state", ""))
	if player_anim != null and (typ == "monster" or (typ == "interactable" and state == "closed")):
		player_anim.play_one_shot("attack")

	client.send("action_intent", last_server_tick, {"target_id": target_id})
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


func _pick_entity_at_mouse() -> String:
	if _camera == null:
		return ""
	var mouse_pos := get_viewport().get_mouse_position()
	var origin := _camera.project_ray_origin(mouse_pos)
	var normal := _camera.project_ray_normal(mouse_pos)
	var query := PhysicsRayQueryParameters3D.create(origin, origin + normal * 200.0)
	query.collide_with_areas = true
	query.collide_with_bodies = true
	var hit := get_world_3d().direct_space_state.intersect_ray(query)
	if hit.is_empty():
		return ""
	var collider = hit.get("collider")
	if collider != null and collider.has_meta("entity_id"):
		return str(collider.get_meta("entity_id"))
	return ""


func _adjust_camera_zoom(delta_size: float) -> void:
	if _camera == null:
		return

	_camera.size = clampf(_camera.size + delta_size, CAMERA_ZOOM_MIN, CAMERA_ZOOM_MAX)


# --- visual replay playlist -------------------------------------------------

func _load_visual_replay_manifest(path: String) -> bool:
	if not FileAccess.file_exists(path):
		push_error("visual replay manifest not found: %s" % path)
		return false
	var text := FileAccess.get_file_as_string(path)
	var parsed = JSON.parse_string(text)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_error("visual replay manifest is not a JSON object: %s" % path)
		return false
	visual_replay_scenarios = parsed.get("scenarios", [])
	return visual_replay_scenarios.size() > 0


func _start_next_visual_replay() -> void:
	visual_replay_index += 1
	visual_replay_envelopes = []
	visual_replay_envelope_index = 0
	visual_replay_timer = autoplay_step_delay
	if visual_replay_index >= visual_replay_scenarios.size():
		visual_replay_title = "complete"
		_debug("visual replay playlist complete")
		if visual_replay_exit_on_complete:
			visual_replay_exit_requested = true
			visual_replay_exit_timer = maxf(autoplay_step_delay, 0.25)
		return

	var scenario: Dictionary = visual_replay_scenarios[visual_replay_index]
	var session_id := str(scenario.get("session_id", ""))
	var world_id := str(scenario.get("world_id", "vertical_slice"))
	visual_replay_title = str(scenario.get("title", scenario.get("id", session_id)))
	_render_world_walls(world_id)
	if session_id == "":
		_debug("visual replay entry missing session_id; skipping")
		_start_next_visual_replay()
		return
	var timeline := client.get_replay_timeline(visual_replay_debug_token, session_id)
	visual_replay_envelopes = timeline.get("envelopes", [])
	_debug("visual replay %d/%d: %s (%d envelopes)" % [
		visual_replay_index + 1, visual_replay_scenarios.size(), visual_replay_title, visual_replay_envelopes.size()])
	if visual_replay_envelopes.is_empty():
		_start_next_visual_replay()


func _handle_visual_replay(delta: float) -> void:
	if visual_replay_exit_requested:
		visual_replay_exit_timer -= delta
		if visual_replay_exit_timer <= 0.0:
			_debug("visual replay exit requested")
			if client != null:
				client.close()
			get_tree().quit(0)
		return
	if visual_replay_index >= visual_replay_scenarios.size():
		return
	visual_replay_timer -= delta
	if visual_replay_timer > 0.0:
		return
	if visual_replay_envelope_index >= visual_replay_envelopes.size():
		visual_replay_timer = maxf(autoplay_step_delay * 4.0, 0.5)
		_start_next_visual_replay()
		return

	var env: Dictionary = visual_replay_envelopes[visual_replay_envelope_index]
	visual_replay_envelope_index += 1
	_handle_message(env)
	visual_replay_timer = _visual_replay_delay_for(env)


func _visual_replay_delay_for(env: Dictionary) -> float:
	if str(env.get("type", "")) != "state_delta":
		return autoplay_step_delay
	var payload: Dictionary = env.get("payload", {})
	for change in payload.get("changes", []):
		if change.get("op", "") in ["entity_spawn", "entity_update"]:
			var entity: Dictionary = change.get("entity", {})
			if str(entity.get("type", "")) == "projectile":
				return 0.05
		if change.get("op", "") == "entity_remove":
			return 0.08
	return autoplay_step_delay


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

	damage_numbers_layer = CanvasLayer.new()
	damage_numbers_layer.layer = 2
	add_child(damage_numbers_layer)

	health_bars_layer = CanvasLayer.new()
	health_bars_layer.layer = 1
	add_child(health_bars_layer)

	walls_root = Node3D.new()
	walls_root.name = "StaticWalls"
	add_child(walls_root)


func _render_world_walls(world_id: String) -> void:
	if walls_root == null:
		return
	for child in walls_root.get_children():
		child.queue_free()

	var rules_path := ProjectSettings.globalize_path("res://").path_join("../shared/rules/worlds.v0.json")
	var parsed = _read_json(rules_path)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("[main] could not read world rules for walls: %s" % rules_path)
		return
	var worlds: Dictionary = parsed.get("worlds", {})
	var world: Dictionary = worlds.get(world_id, {})
	for entity in world.get("entities", []):
		if str(entity.get("type", "")) != "wall":
			continue
		var pos: Dictionary = entity.get("position", {})
		var size: Dictionary = entity.get("size", {})
		var node := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(float(size.get("x", 1.0)), 1.0, float(size.get("y", 1.0)))
		node.mesh = mesh
		node.position = Vector3(float(pos.get("x", 0.0)), 0.5, float(pos.get("y", 0.0)))
		var mat := StandardMaterial3D.new()
		mat.albedo_color = Color(0.32, 0.34, 0.36)
		node.material_override = mat
		walls_root.add_child(node)


func _read_json(path: String):
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		return null
	return JSON.parse_string(f.get_as_text())


func _make_entity_node(kind: String) -> Node3D:
	# Monster adopts the rigged dummy scene (spec §5.3); loot stays a primitive.
	if kind == "monster":
		var packed := MonsterDummyScene
		if packed != null:
			return packed.instantiate()
		# Fallback: red primitive so positioning/targeting still works.
		var fallback := MeshInstance3D.new()
		var fm := StandardMaterial3D.new()
		fm.albedo_color = Color(1.0, 0.3, 0.3)
		fallback.mesh = BoxMesh.new()
		fallback.material_override = fm
		return fallback
	if kind == "interactable":
		return _make_door_node()
	if kind == "projectile":
		return _make_projectile_node()
	var node := MeshInstance3D.new()  # loot
	var box := BoxMesh.new()
	box.size = Vector3(0.5, 0.5, 0.5)
	node.mesh = box
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.85, 0.2)
	node.material_override = mat
	return node


func _make_projectile_node() -> Node3D:
	var root := Node3D.new()
	root.name = "Projectile"
	var shaft := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.16, 0.16, 0.7)
	shaft.mesh = mesh
	shaft.position = Vector3(0.0, 0.35, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.65, 0.90, 1.0)
	mat.emission_enabled = true
	mat.emission = Color(0.25, 0.55, 0.9)
	shaft.material_override = mat
	root.add_child(shaft)
	return root


func _move_projectile_node(rec: Dictionary, target_pos: Vector3) -> void:
	var node := rec["node"] as Node3D
	if node == null:
		return
	var from := node.position
	var flat := Vector2(target_pos.x - from.x, target_pos.z - from.z)
	if flat.length_squared() > 0.0001:
		node.look_at(Vector3(target_pos.x, from.y, target_pos.z), Vector3.UP)
	if rec.has("move_tween"):
		var old_tween = rec["move_tween"]
		if is_instance_valid(old_tween):
			old_tween.kill()
	var duration := PROJECTILE_LERP_SECONDS
	if visual_replay_enabled:
		duration = clampf(autoplay_step_delay * 0.35, 0.06, 0.18)
	var tween := create_tween()
	rec["move_tween"] = tween
	tween.tween_property(node, "position", target_pos, duration).set_trans(Tween.TRANS_LINEAR)


func _make_door_node() -> Node3D:
	var root := Node3D.new()
	root.name = "InteractableDoor"
	var pivot := Node3D.new()
	pivot.name = "DoorPivot"
	pivot.position = Vector3(-0.5, 0.0, 0.0)
	root.add_child(pivot)
	var panel := MeshInstance3D.new()
	panel.name = "DoorPanel"
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.0, 1.0, 0.25)
	panel.mesh = mesh
	panel.position = Vector3(0.5, 0.5, 0.0)
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.55, 0.32, 0.15)
	panel.material_override = mat
	pivot.add_child(panel)
	return root


func _attach_pick_collider(node: Node3D, entity_id: String, kind: String) -> void:
	var body := StaticBody3D.new()
	body.name = "PickBody"
	body.set_meta("entity_id", entity_id)
	var shape := CollisionShape3D.new()
	var box := BoxShape3D.new()
	match kind:
		"monster":
			box.size = Vector3(1.0, 1.6, 1.0)
			shape.position = Vector3(0.0, 0.8, 0.0)
		"interactable":
			box.size = Vector3(1.2, 1.2, 0.45)
			shape.position = Vector3(0.0, 0.6, 0.0)
		_:
			box.size = Vector3(0.75, 0.75, 0.75)
			shape.position = Vector3(0.0, 0.375, 0.0)
	shape.shape = box
	body.add_child(shape)
	node.add_child(body)


func _set_interactable_state(_entity_id: String, rec: Dictionary, state: String) -> void:
	if rec.get("state", "") == state:
		return
	rec["state"] = state
	var node := rec["node"] as Node3D
	if node == null:
		return
	var pivot := node.find_child("DoorPivot", true, false) as Node3D
	if pivot == null:
		return
	var target_rot := deg_to_rad(90.0) if state == "open" else 0.0
	var tween := create_tween()
	tween.tween_property(pivot, "rotation:y", target_rot, 0.25)


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
	var mode := "visual-replay:%d/%d %s" % [
		min(visual_replay_index + 1, visual_replay_scenarios.size()),
		visual_replay_scenarios.size(),
		visual_replay_title,
	] if visual_replay_enabled else ("visual-bot:%s" % autoplay_phase if autoplay_enabled else "manual")
	_debug_label.text = "ws=%s  tick=%d  mode=%s  recon_delta=%.2f\ninv=%d  entities=%d  equipped_weapon=%s\nweapon_visual=%s\nW/A/S/D move  LMB action  scroll zoom  Q equip" % [
		ws_state, last_server_tick, mode, reconciliation_delta, inventory.size(), entities.size(), str(eq), weapon_vis]


func _debug(msg: String) -> void:
	print("[client] ", msg)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback


func _truthy_env(key: String) -> bool:
	return _truthy_text(OS.get_environment(key))


func _truthy_text(value: String) -> bool:
	var v := value.to_lower()
	return v in ["1", "true", "yes", "on"]
