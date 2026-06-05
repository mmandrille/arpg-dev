# Headless client smoke test (ADR-0001 D8.5 layer 2).
#
# Run with: godot --headless --path client --script res://scripts/smoke.gd
# Drives the full slice through the real protocol and verifies client + server
# state via the debug /state API, then quits with exit code 0 (pass) or 1.
extends SceneTree

const NetClientScript := preload("res://scripts/net_client.gd")
const TIMEOUT_S := 20.0

var client: NetClient
var elapsed: float = 0.0
var attack_cd: float = 0.0
var last_tick: int = 0
var ready_sent: bool = false
var killed: bool = false
var picked: bool = false
var equip_sent: bool = false
var equipped: bool = false
var loot_id: String = ""
var item_id: String = ""
var debug_token: String = ""


func _initialize() -> void:
	var base := _env("BASE_URL", "http://localhost:8080")
	var dev := _env("DEV_TOKEN", "local-dev-token")
	debug_token = _env("DEBUG_TOKEN", "local-debug-token")

	client = NetClientScript.new(base)
	if not client.login("smoke@example.test", dev):
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
		_fail("timeout (killed=%s picked=%s equipped=%s)" % [killed, picked, equipped])
		return true

	# Pump the socket every frame (advances the handshake and drains packets).
	var msgs := client.poll()
	var st := client.ready_state()
	if st == WebSocketPeer.STATE_CLOSED:
		_fail("websocket closed")
		return true
	if st != WebSocketPeer.STATE_OPEN:
		return false

	if not ready_sent:
		client.send("client_ready", last_tick, {"client_version": "godot-smoke", "last_seen_tick": last_tick})
		client.send("attack_intent", last_tick, {"target_id": "1002"})
		attack_cd = 0.15
		ready_sent = true

	for env in msgs:
		_handle(env)

	attack_cd -= delta
	if not killed and attack_cd <= 0.0:
		client.send("attack_intent", last_tick, {"target_id": "1002"})
		attack_cd = 0.15
	if killed and not picked and loot_id != "":
		client.send("pick_up_intent", last_tick, {"entity_id": loot_id})
		picked = true
	if picked and item_id != "" and not equip_sent:
		client.send("equip_intent", last_tick, {"item_instance_id": item_id, "slot": "weapon"})
		equip_sent = true
	if equipped:
		return _verify_and_quit()
	return false


func _handle(env: Dictionary) -> void:
	last_tick = max(last_tick, int(env.get("tick", 0)))
	if env.get("type", "") != "state_delta":
		return
	var p: Dictionary = env["payload"]
	for ev in p.get("events", []):
		if ev.get("event_type", "") == "monster_killed":
			killed = true
	for c in p.get("changes", []):
		match c.get("op", ""):
			"entity_spawn":
				if c["entity"].get("type", "") == "loot":
					loot_id = str(c["entity"]["id"])
			"inventory_add":
				item_id = str(c["item"]["item_instance_id"])
			"equipped_update":
				if c.get("slot", "") == "weapon" and str(c.get("item_instance_id", "")) == item_id:
					equipped = true


func _verify_and_quit() -> bool:
	var state := client.get_state(debug_token)
	var inv: Array = state.get("inventory", [])
	var eq: Dictionary = state.get("equipped", {})
	var ok: bool = inv.size() == 1 \
		and inv[0].get("item_def_id", "") == "rusty_sword" \
		and inv[0].get("equipped", false) \
		and str(eq.get("weapon", "")) == item_id
	if ok:
		print("[smoke] PASS: slice complete and verified via /state")
		quit(0)
	else:
		printerr("[smoke] FAIL: unexpected state ", state)
		quit(1)
	return true


func _fail(msg: String) -> void:
	printerr("[smoke] FAIL: ", msg)
	quit(1)


func _env(key: String, fallback: String) -> String:
	var v := OS.get_environment(key)
	return v if v != "" else fallback
