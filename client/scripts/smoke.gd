# Headless client smoke test (ADR-0001 D8.5 layer 2; spec equip-and-see-it §10.6).
#
# Run with: godot --headless --path client --script res://scripts/smoke.gd
# Drives the full slice through the real protocol and verifies, in one run:
#   1. kill -> pickup -> equip, then BOTH server /state AND the client visual
#      (EquipmentVisualResolver.get_debug_state(): weapon.visible == true with
#      matching ids + socket)                                  (acceptance #6/#7)
#   2. a move intent after equip leaves the weapon mounted     (acceptance #13)
#   3. disconnect + resume the SAME session and restore the mounted weapon from
#      the session_snapshot alone, with restored HP/death state (acceptance #8)
# Quits 0 (pass) or 1 (fail). The resolver mounts under an injected minimal
# mount-root (Node3D + right_hand_socket child), the §4.8 decoupled code path.
extends SceneTree

const NetClientScript := preload("res://scripts/net_client.gd")
const ResolverScript := preload("res://scripts/equipment_visuals.gd")
const AnimControllerScript := preload("res://scripts/animation_controller.gd")
const InventoryPanelScript := preload("res://scripts/inventory_panel.gd")
const CharacterScene := preload("res://scenes/character.tscn")
const MonsterScene := preload("res://scenes/monster_dummy.tscn")
const TIMEOUT_S := 40.0

var base: String = ""
var dev: String = ""
var debug_token: String = ""

var client: NetClient
var resolver: EquipmentVisualResolver

# A real monster controller driven through the same event/snapshot path as the
# player slice, to prove monster hit -> death and the resume death pose.
var monster_anim: AnimationController
var monster_saw_hit: bool = false
var player_anim: AnimationController
var player_saw_hit: bool = false
var player_hit_clip_seen: bool = false

# Resume phase uses a fresh client + resolver so the restored visual provably
# comes from the snapshot, not from leftover live state.
var client2: NetClient
var resolver_resume: EquipmentVisualResolver
var resume_ready_sent: bool = false
# Fresh monster controller for the resume phase: a dead monster in the snapshot
# (type=="monster", hp<=0) must enter terminal "death" from hp alone -- no event
# / delta replay -- proving resume-from-snapshot (acceptance #8 for monsters).
var monster_anim_resume: AnimationController

var phase: String = "play"   # play -> verify_equip -> moving -> resuming
var elapsed: float = 0.0
var attack_cd: float = 0.0
var move_wait: float = 0.0
var last_tick: int = 0
var ready_sent: bool = false
var killed: bool = false
var picked: bool = false
var equip_sent: bool = false
var equipped: bool = false
var verify_equip_wait: float = 0.0
var loot_id: String = ""
var item_id: String = ""


func _initialize() -> void:
	if not _verify_inventory_panel_model():
		return
	base = _env("BASE_URL", "http://localhost:8888")
	dev = _env("DEV_TOKEN", "local-dev-token")
	debug_token = _env("DEBUG_TOKEN", "local-debug-token")

	resolver = _make_resolver()
	var player := CharacterScene.instantiate()
	get_root().add_child(player)
	var player_ap := player.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if player_ap == null:
		_fail("character scene has no AnimationPlayer")
		return
	player_anim = AnimControllerScript.new(player_ap)
	# Real monster_dummy + controller, driven by the same authoritative event
	# stream as the slice. The scene has no script/_ready, so the AnimationPlayer
	# (static scene data) is available right after instantiate(); the controller's
	# _init plays "idle" synchronously even out-of-tree (same as Task 6).
	var mon := MonsterScene.instantiate()
	get_root().add_child(mon)
	var mon_ap := mon.find_child("AnimationPlayer", true, false) as AnimationPlayer
	monster_anim = AnimControllerScript.new(mon_ap)
	client = NetClientScript.new(base)
	var email := OS.get_environment("ARPG_EMAIL")
	if email == "":
		email = "smoke@example.test"
	if not client.login(email, dev):
		_fail("login failed")
		return
	if not client.create_session():
		_fail("create_session failed")
		return
	client.connect_ws()
	print("[smoke] session ", client.session_id)


func _process(delta: float) -> bool:
	if client == null:
		return true
	elapsed += delta
	if elapsed > TIMEOUT_S:
		_fail("timeout (phase=%s killed=%s picked=%s equipped=%s)" % [phase, killed, picked, equipped])
		return true
	if phase == "resuming":
		return _step_resume(delta)
	return _step_primary(delta)


# --- primary phases: play / verify_equip / moving ---------------------------

func _step_primary(delta: float) -> bool:
	var msgs := client.poll()
	var st := client.ready_state()
	if st == WebSocketPeer.STATE_CLOSED:
		_fail("primary websocket closed unexpectedly (phase=%s)" % phase)
		return true
	if st != WebSocketPeer.STATE_OPEN:
		return false

	if not ready_sent:
		client.send("client_ready", last_tick, {"client_version": "godot-smoke", "last_seen_tick": last_tick})
		client.send("move_intent", last_tick, {"direction": {"x": 1, "y": 0}, "duration_ticks": 1})
		attack_cd = 0.5
		ready_sent = true

	for env in msgs:
		_handle(env)

	attack_cd -= delta
	if not killed and attack_cd <= 0.0:
		client.send("action_intent", last_tick, {"target_id": "1002"})
		attack_cd = 0.15
	if killed and not picked and loot_id != "":
		client.send("action_intent", last_tick, {"target_id": loot_id})
		picked = true
	if picked and item_id != "" and not equip_sent:
		client.send("equip_intent", last_tick, {"item_instance_id": item_id, "slot": "main_hand"})
		equip_sent = true

	if phase == "play" and equipped:
		phase = "verify_equip"

	if phase == "verify_equip":
		verify_equip_wait += delta
		_sync_resolver_from_server_state()
		if _verify_equip():
			client.send("move_intent", last_tick, {"direction": {"x": 1, "y": 0}, "duration_ticks": 2})
			move_wait = 0.5
			phase = "moving"
			return false
		if verify_equip_wait > 2.0:
			_fail_equip_verification()
			return true
		return false

	if phase == "moving":
		move_wait -= delta
		if move_wait <= 0.0:
			if not _weapon_mounted(resolver):
				_fail("weapon visual not mounted after move intent (acceptance #13): %s" % resolver.get_debug_state())
				return true
			print("[smoke] move-after-equip OK: weapon still mounted")
			_begin_resume()
			phase = "resuming"
	return false


# --- resume phase: rejoin the same session, restore from snapshot -----------

func _begin_resume() -> void:
	var resumed_id := client.session_id
	client.close()
	client2 = NetClientScript.new(base)
	var email := OS.get_environment("ARPG_EMAIL")
	if email == "":
		email = "smoke@example.test"
	if not client2.login(email, dev):
		_fail("resume login failed")
		return
	if not client2.create_session(resumed_id):
		_fail("resume create_session(resume_session_id=%s) failed" % resumed_id)
		return
	if client2.session_id != resumed_id:
		_fail("resume returned different session id (%s != %s)" % [client2.session_id, resumed_id])
		return
	resolver_resume = _make_resolver()
	client2.connect_ws()
	print("[smoke] resuming session ", client2.session_id)


func _step_resume(delta: float) -> bool:
	if client2 == null:
		return true  # _begin_resume already failed + quit
	var msgs := client2.poll()
	var st := client2.ready_state()
	if st == WebSocketPeer.STATE_CLOSED:
		_fail("resume websocket closed unexpectedly")
		return true
	if st != WebSocketPeer.STATE_OPEN:
		return false

	if not resume_ready_sent:
		client2.send("client_ready", last_tick, {"client_version": "godot-smoke-resume", "last_seen_tick": last_tick})
		resume_ready_sent = true

	var got_snapshot := false
	var resume_snap: Dictionary = {}
	for env in msgs:
		if env.get("type", "") == "session_snapshot":
			var snap: Dictionary = env["payload"]
			resume_snap = snap
			resolver_resume.apply_snapshot(snap)
			# Acceptance #8 (monsters): a snapshot entity with type=="monster" and
			# hp<=0 must drive the controller to terminal "death" from hp ALONE --
			# no delta / recent_events replay (spec §5.4).
			_resume_monster_from_snapshot(snap)
			got_snapshot = true

	if got_snapshot:
		if _monster_hp_from_state(resume_snap) != 0:
			_fail("resume snapshot did not restore dead monster: %s" % resume_snap.get("entities", []))
			return true
		# The client put the real hp==0 snapshot monster into the terminal death pose.
		if monster_anim_resume == null:
			_fail("resume snapshot carried no monster entity to assert death pose (acceptance #8)")
			return true
		if monster_anim_resume.get_debug_state()["terminal"] != true:
			_fail("resumed monster did not enter terminal death pose from snapshot hp (acceptance #8): %s" % monster_anim_resume.get_debug_state())
			return true
		var w = resolver_resume.get_debug_state()["equipped_visuals"]["weapon"]
		var ok: bool = w != null and w["visible"] == true \
			and str(w["item_instance_id"]) == item_id \
			and w["item_def_id"] == "rusty_sword" \
			and w["asset_id"] == "weapon_rusty_sword_v0" \
			and w["mount_socket"] == "right_hand_socket"
		if ok:
			print("[smoke] PASS: equip visible, survives move, and restored monster death state on resume")
			quit(0)
		else:
			_fail("resume did not restore weapon visual from snapshot (acceptance #8): %s" % w)
		return true
	return false


# --- message handling (primary) ---------------------------------------------

func _handle(env: Dictionary) -> void:
	last_tick = max(last_tick, int(env.get("tick", 0)))
	var t := str(env.get("type", ""))
	if t == "session_snapshot":
		resolver.apply_snapshot(env["payload"])
		return
	if t != "state_delta":
		return
	var p: Dictionary = env["payload"]
	for c in p.get("changes", []):
		match c.get("op", ""):
			"entity_spawn":
				if c["entity"].get("type", "") == "loot":
					loot_id = str(c["entity"]["id"])
			"inventory_add":
				resolver.ingest_inventory_item(c["item"])
				# Only track loot picked up after the kill; ignore persisted bag items.
				if killed and picked:
					item_id = str(c["item"]["item_instance_id"])
			"inventory_update":
				resolver.ingest_inventory_item(c["item"])
			"equipped_update":
				resolver.apply_equipped_update(c.get("slot", ""), c.get("item_instance_id"))
				if c.get("slot", "") == "main_hand" and c.get("item_instance_id") != null:
					item_id = str(c.get("item_instance_id"))
					equipped = true
			"weapon_set_update":
				resolver.apply_weapon_set_update(
					int(c.get("active_weapon_set", 0)),
					c.get("weapon_sets", [])
				)
	for ev in p.get("events", []):
		var et := str(ev.get("event_type", ""))
		if et == "monster_damaged" and monster_anim != null:
			monster_anim.play_one_shot("hit")
			monster_saw_hit = true
		if et == "monster_killed":
			killed = true
			if monster_anim != null:
				monster_anim.enter_terminal("death")
		if et == "player_damaged" and player_anim != null:
			player_anim.play_one_shot("hit")
			player_saw_hit = true
			player_hit_clip_seen = player_hit_clip_seen or player_anim.current_clip() == "hit"
		if et == "player_killed" and player_anim != null:
			player_anim.enter_terminal("death")


# --- verification helpers ----------------------------------------------------

func _sync_resolver_from_server_state() -> void:
	var state := client.get_state(debug_token)
	for item in state.get("inventory", []):
		if item is Dictionary:
			resolver.ingest_inventory_item(item)
	resolver.apply_weapon_set_update(
		int(state.get("active_weapon_set", 0)),
		state.get("weapon_sets", [])
	)
	var eq: Dictionary = state.get("equipped", {})
	for slot in ["head", "amulet", "chest", "gloves", "belt", "boots", "ring_left", "ring_right", "main_hand", "off_hand"]:
		if eq.has(slot):
			resolver.apply_equipped_update(slot, eq.get(slot))
	var main_id := _main_hand_item_id(state)
	if main_id != "":
		item_id = main_id


func _main_hand_item_id(state: Dictionary) -> String:
	var raw = state.get("equipped", {}).get("main_hand", null)
	if raw == null:
		return ""
	return str(raw)


func _inventory_item_for(state: Dictionary, instance_id: String) -> Dictionary:
	for item in state.get("inventory", []):
		if str(item.get("item_instance_id", "")) == instance_id:
			return item
	return {}


func _verify_equip() -> bool:
	# Server authority (existing v0 check) AND client visual (new in v2).
	var state := client.get_state(debug_token)
	var main_id := _main_hand_item_id(state)
	var inv_item := _inventory_item_for(state, main_id)
	var hp := _player_hp_from_state(state)
	var server_ok: bool = main_id != "" \
		and inv_item.get("item_def_id", "") == "rusty_sword" \
		and bool(inv_item.get("equipped", false))

	var w = resolver.get_debug_state()["equipped_visuals"]["weapon"]
	var visual_ok: bool = w != null and w["visible"] == true \
		and str(w["item_instance_id"]) == main_id \
		and w["item_def_id"] == "rusty_sword" \
		and w["asset_id"] == "weapon_rusty_sword_v0" \
		and w["mount_socket"] == "right_hand_socket"

	if server_ok and visual_ok:
		if monster_anim != null and monster_anim.get_debug_state()["terminal"] != true:
			_fail("monster did not reach terminal death pose after kill: %s" % monster_anim.get_debug_state())
			return false
		print("[smoke] equip verified + monster death pose terminal hp=%d" % hp)
		return true
	return false


func _fail_equip_verification() -> void:
	var state := client.get_state(debug_token)
	var main_id := _main_hand_item_id(state)
	var inv_item := _inventory_item_for(state, main_id)
	var hp := _player_hp_from_state(state)
	var server_ok: bool = main_id != "" \
		and inv_item.get("item_def_id", "") == "rusty_sword" \
		and bool(inv_item.get("equipped", false))
	var w = resolver.get_debug_state()["equipped_visuals"]["weapon"]
	var visual_ok: bool = w != null and w["visible"] == true \
		and str(w["item_instance_id"]) == main_id \
		and w["item_def_id"] == "rusty_sword" \
		and w["asset_id"] == "weapon_rusty_sword_v0" \
		and w["mount_socket"] == "right_hand_socket"
	_fail("equip verification failed (server_ok=%s visual_ok=%s hp=%d main_hand=%s) state=%s visual=%s warnings=%s" % [server_ok, visual_ok, hp, main_id, state, w, resolver.get_debug_state().get("warnings", [])])


func _weapon_mounted(res: EquipmentVisualResolver) -> bool:
	var w = res.get_debug_state()["equipped_visuals"]["weapon"]
	return w != null and w.get("visible", false) == true


func _player_hp_from_state(state: Dictionary) -> int:
	for e in state.get("entities", []):
		if str(e.get("type", "")) == "player":
			return int(e.get("hp", -1))
	return -1


func _monster_hp_from_state(state: Dictionary) -> int:
	for e in state.get("entities", []):
		if str(e.get("id", "")) == "1002" and str(e.get("type", "")) == "monster":
			return int(e.get("hp", -1))
	return -1


func _resume_monster_from_snapshot(snap: Dictionary) -> void:
	# Mirror main.gd::_upsert_entity for monsters: instance a real monster +
	# fresh controller (not the primary monster_anim, so the pose provably comes
	# from this snapshot and not leftover live state) and enter terminal "death"
	# when the snapshot entity reports hp<=0.
	for e in snap.get("entities", []):
		if str(e.get("type", "")) != "monster":
			continue
		if int(e.get("hp", -1)) > 0:
			continue
		if monster_anim_resume == null:
			var mon := MonsterScene.instantiate()
			get_root().add_child(mon)
			var mon_ap := mon.find_child("AnimationPlayer", true, false) as AnimationPlayer
			monster_anim_resume = AnimControllerScript.new(mon_ap)
		monster_anim_resume.enter_terminal("death")


func _make_resolver() -> EquipmentVisualResolver:
	# Minimal injected mount-root: Node3D with a right_hand_socket child, added to
	# the tree so mounted nodes are inside_tree (spec §4.8 decoupled path).
	var mount := Node3D.new()
	mount.name = "CharacterVisual"
	var socket := Node3D.new()
	socket.name = "right_hand_socket"
	mount.add_child(socket)
	get_root().add_child(mount)
	return ResolverScript.new(mount)


func _verify_inventory_panel_model() -> bool:
	var panel = InventoryPanelScript.new()
	get_root().add_child(panel)
	panel.set_inventory_state([
		{"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": false},
		{"item_instance_id": "1005", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
	], {"main_hand": null})
	var state: Dictionary = panel.get_debug_state()
	if int(state["bag_count"]) != 2 or state["equipped_main_hand"] != null:
		_fail("inventory panel initial state mismatch: %s" % state)
		return false
	var window: Dictionary = state.get("window", {})
	if str(window.get("title", "")) != "Inventory" or not bool(window.get("close_visible", false)) or not bool(window.get("draggable", false)):
		_fail("inventory panel window chrome mismatch: %s" % state)
		return false
	panel.bot_drag_window_by(Vector2(30, 20))
	state = panel.get_debug_state()
	var moved_position: Dictionary = (state.get("window", {}) as Dictionary).get("position", {})
	if int(moved_position.get("x", 0)) <= 0 or int(moved_position.get("y", 0)) <= 0:
		_fail("inventory panel drag mismatch: %s" % state)
		return false
	panel.bot_click_close()
	if panel.visible:
		_fail("inventory panel close mismatch: %s" % state)
		return false
	var paper_slots: Dictionary = state["paper_doll_slots"]
	if float(paper_slots["ring_left"]["position"]["x"]) != float(paper_slots["main_hand"]["position"]["x"]) \
			or float(paper_slots["ring_left"]["position"]["y"]) <= float(paper_slots["main_hand"]["position"]["y"]) \
			or float(paper_slots["ring_right"]["position"]["x"]) != float(paper_slots["off_hand"]["position"]["x"]) \
			or float(paper_slots["ring_right"]["position"]["y"]) <= float(paper_slots["off_hand"]["position"]["y"]):
		_fail("inventory panel ring slot placement mismatch: %s" % state)
		return false
	panel.set_inventory_state([
		{"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": true},
		{"item_instance_id": "1005", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
	], {"main_hand": "1004"})
	state = panel.get_debug_state()
	if str(state["equipped_main_hand"]) != "1004":
		_fail("inventory panel equipped state mismatch: %s" % state)
		return false
	panel.set_inventory_state([
		{"item_instance_id": "1009", "item_def_id": "cave_bow", "item_template_id": "cave_bow", "slot": "main_hand", "equipped": true, "rarity": "common"},
	], {"main_hand": "1009"})
	state = panel.get_debug_state()
	var bow_slots: Dictionary = state["paper_doll_slots"]
	if not bool((bow_slots["off_hand"] as Dictionary).get("blocked_by_two_handed", false)) \
			or str((bow_slots["off_hand"] as Dictionary).get("item_def_id", "")) != "cave_bow":
		_fail("inventory panel two-handed off-hand blocker mismatch: %s" % state)
		return false
	panel.set_inventory_state([
		{"item_instance_id": "1005", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
	], {"main_hand": null})
	state = panel.get_debug_state()
	if int(state["bag_count"]) != 1 or state["weapon_item"] != {}:
		_fail("inventory panel remove state mismatch: %s" % state)
		return false
	panel.set_inventory_state([
		{"item_instance_id": "1005", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
	], {"main_hand": null}, 3, 15, 0, [{"slot_index": 0, "item_instance_id": "1006", "item": {"item_instance_id": "1006", "item_def_id": "red_potion", "slot": "", "equipped": false}}])
	state = panel.get_debug_state()
	var hotbar_assigned_ids: Array = state["hotbar_assigned_item_ids"]
	if int(state["bag_count"]) != 1 \
			or int(state["hotbar_assigned_inventory_count"]) != 0 \
			or hotbar_assigned_ids.has("1006"):
		_fail("inventory panel hotbar assignment state mismatch: %s" % state)
		return false
	panel.set_inventory_state([
		{"item_instance_id": "1005", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
		{"item_instance_id": "1006", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
		{"item_instance_id": "1007", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
		{"item_instance_id": "1008", "item_def_id": "quest_leaf", "slot": "", "equipped": false},
	], {"main_hand": null}, 0, 2)
	state = panel.get_debug_state()
	if int(state["available_slot_count"]) != 2 \
			or int(state["rendered_slot_count"]) < int(state["bag_count"]) \
			or int(state["rendered_slot_count"]) % int(state["bag_columns"]) != 0:
		_fail("inventory panel overflow render mismatch: %s" % state)
		return false
	panel.queue_free()
	return true


func _fail(msg: String) -> void:
	printerr("[smoke] FAIL: ", msg)
	quit(1)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback
