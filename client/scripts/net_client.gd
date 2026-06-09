# NetClient: the thin transport layer for the arpg client.
#
# Speaks the same auth + WebSocket protocol as the Python bot (ADR-0001 D4/D5):
# dev-login + create session over HTTP, then the v0 JSON envelope over a
# WebSocket. Used by both the interactive scene (main.gd) and the headless
# smoke runner (smoke.gd).
extends RefCounted
class_name NetClient

var base_url: String
var host: String
var port: int
var use_tls: bool

var token: String = ""
var account_id: String = ""
var session_id: String = ""
var seed: String = ""
var world_id: String = ""
var session_mode: String = ""
var session_listed: bool = false
var ws_url: String = ""

var _ws := WebSocketPeer.new()
var _msg_counter: int = 0
var _path_prefix: String = ""


func _init(p_base_url: String) -> void:
	base_url = p_base_url.strip_edges().trim_suffix("/")
	use_tls = base_url.begins_with("https://")
	var rest := base_url.replace("https://", "").replace("http://", "")
	var hostport := rest.split("/")[0]
	_path_prefix = ""
	if rest.find("/") >= 0:
		_path_prefix = "/" + rest.substr(rest.find("/") + 1).trim_suffix("/")
	var parts := hostport.split(":")
	host = parts[0]
	if parts.size() > 1:
		port = int(parts[1])
	else:
		port = 443 if use_tls else 80


func _request_path(path: String) -> String:
	# BASE_URL paths are accepted for launcher ergonomics, but the game API is
	# rooted at /v0 on the backend, so strip any path component.
	return path


# --- HTTP (blocking, dev-only) ---------------------------------------------

func _http(method: int, path: String, headers: Array, body: String) -> Dictionary:
	var client := HTTPClient.new()
	var tls = TLSOptions.client() if use_tls else null
	var err := client.connect_to_host(host, port, tls)
	if err != OK:
		return {"_error": "connect_to_host failed: %d" % err}
	while client.get_status() in [HTTPClient.STATUS_CONNECTING, HTTPClient.STATUS_RESOLVING]:
		client.poll()
		OS.delay_msec(5)
	if client.get_status() != HTTPClient.STATUS_CONNECTED:
		return {"_error": "not connected: %d" % client.get_status()}

	var all_headers := ["Content-Type: application/json"]
	all_headers.append_array(headers)
	err = client.request(method, _request_path(path), all_headers, body)
	if err != OK:
		return {"_error": "request failed: %d" % err}
	while client.get_status() == HTTPClient.STATUS_REQUESTING:
		client.poll()
		OS.delay_msec(5)

	var code := client.get_response_code()
	var buf := PackedByteArray()
	while client.get_status() == HTTPClient.STATUS_BODY:
		client.poll()
		var chunk := client.read_response_body_chunk()
		if chunk.size() > 0:
			buf.append_array(chunk)
		else:
			OS.delay_msec(5)
	var parsed = JSON.parse_string(buf.get_string_from_utf8())
	return {"_code": code, "body": parsed}


func login(email: String, dev_token: String) -> bool:
	var r := _http(HTTPClient.METHOD_POST, "/v0/auth/dev-login", [],
		JSON.stringify({"email": email, "dev_token": dev_token}))
	if r.get("_code", 0) == 200 and r.has("body"):
		token = r["body"]["access_token"]
		account_id = r["body"]["account_id"]
		return true
	push_error("login failed: %s" % r)
	return false


func list_characters() -> Array:
	var r := _http(HTTPClient.METHOD_GET, "/v0/characters",
		["Authorization: Bearer " + token], "")
	if r.get("_code", 0) == 200 and r.has("body"):
		return r["body"].get("characters", [])
	push_error("list_characters failed: %s" % r)
	return []


func create_character(name: String) -> Dictionary:
	var r := _http(HTTPClient.METHOD_POST, "/v0/characters",
		["Authorization: Bearer " + token], JSON.stringify({"name": name}))
	if r.get("_code", 0) == 201 and r.has("body"):
		return r["body"]
	push_error("create_character failed: %s" % r)
	return {}


func rename_character(character_id: String, name: String) -> Dictionary:
	if character_id == "":
		return {}
	var r := _http(HTTPClient.METHOD_PATCH, "/v0/characters/" + character_id,
		["Authorization: Bearer " + token], JSON.stringify({"name": name}))
	if r.get("_code", 0) == 200 and r.has("body"):
		return r["body"]
	push_error("rename_character failed: %s" % r)
	return {}


func delete_character(character_id: String) -> bool:
	if character_id == "":
		return false
	var r := _http(HTTPClient.METHOD_DELETE, "/v0/characters/" + character_id,
		["Authorization: Bearer " + token], "")
	return r.get("_code", 0) == 204


func create_session(resume_session_id: String = "", requested_world_id: String = "", character_id: String = "", requested_seed: String = "") -> bool:
	# resume_session_id rejoins an existing session: the server rehydrates
	# inventory AND equipped state before the initial session_snapshot (no
	# protocol change — see spec §4.5). Empty string mints a fresh session.
	var body := {"mode": "solo"}
	if resume_session_id != "":
		body["resume_session_id"] = resume_session_id
	else:
		if requested_world_id != "":
			body["world_id"] = requested_world_id
		if character_id != "":
			body["character_id"] = character_id
		if requested_seed != "":
			body["seed"] = requested_seed
	var r := _http(HTTPClient.METHOD_POST, "/v0/sessions",
		["Authorization: Bearer " + token], JSON.stringify(body))
	if r.get("_code", 0) in [200, 201] and r.has("body"):
		session_id = r["body"]["session_id"]
		seed = r["body"]["seed"]
		world_id = str(r["body"].get("world_id", "vertical_slice"))
		session_mode = str(r["body"].get("mode", "solo"))
		session_listed = bool(r["body"].get("listed", false))
		ws_url = r["body"]["ws_url"]
		return true
	push_error("create_session failed: %s" % r)
	return false


func create_listed_coop_session(character_id: String) -> bool:
	var body := {
		"mode": "coop",
		"listed": true,
		"world_id": "dungeon_levels",
		"character_id": character_id,
	}
	var r := _http(HTTPClient.METHOD_POST, "/v0/sessions",
		["Authorization: Bearer " + token], JSON.stringify(body))
	if r.get("_code", 0) == 201 and r.has("body"):
		session_id = r["body"]["session_id"]
		seed = r["body"]["seed"]
		world_id = str(r["body"].get("world_id", "dungeon_levels"))
		session_mode = str(r["body"].get("mode", "coop"))
		session_listed = bool(r["body"].get("listed", false))
		ws_url = r["body"]["ws_url"]
		return bool(r["body"].get("listed", false))
	push_error("create_listed_coop_session failed: %s" % r)
	return false


func list_active_sessions() -> Array:
	var r := _http(HTTPClient.METHOD_GET, "/v0/sessions/active",
		["Authorization: Bearer " + token], "")
	if r.get("_code", 0) == 200 and r.has("body"):
		return r["body"].get("sessions", [])
	push_error("list_active_sessions failed: %s" % r)
	return []


func join_listed_session(listed_session_id: String, character_id: String) -> bool:
	var r := _http(HTTPClient.METHOD_POST, "/v0/sessions/%s/join" % listed_session_id,
		["Authorization: Bearer " + token], JSON.stringify({"character_id": character_id}))
	if r.get("_code", 0) == 200 and r.has("body"):
		session_id = r["body"]["session_id"]
		seed = r["body"]["seed"]
		world_id = str(r["body"].get("world_id", "dungeon_levels"))
		session_mode = str(r["body"].get("mode", "coop"))
		session_listed = bool(r["body"].get("listed", false))
		ws_url = r["body"]["ws_url"]
		return bool(r["body"].get("listed", false))
	push_error("join_listed_session failed: %s" % r)
	return false


func end_session() -> bool:
	if session_id == "":
		return true
	var r := _http(HTTPClient.METHOD_POST, "/v0/sessions/%s/end" % session_id,
		["Authorization: Bearer " + token], "{}")
	if r.get("_code", 0) == 200:
		return true
	push_warning("end_session failed: %s" % r)
	return false


func get_state(debug_token: String) -> Dictionary:
	var r := _http(HTTPClient.METHOD_GET, "/v0/sessions/%s/state" % session_id,
		["Authorization: Bearer " + token, "X-Debug-Token: " + debug_token], "")
	if r.get("_code", 0) == 200 and r.has("body"):
		return r["body"]
	return {}


func get_replay_timeline(debug_token: String, replay_session_id: String, through_tick: int = -1) -> Dictionary:
	var path := "/v0/sessions/%s/replay/timeline" % replay_session_id
	if through_tick >= 0:
		path += "?through_tick=%d" % through_tick
	var r := _http(HTTPClient.METHOD_GET, path,
		["Authorization: Bearer " + token, "X-Debug-Token: " + debug_token], "")
	if r.get("_code", 0) == 200 and r.has("body"):
		return r["body"]
	push_error("get_replay_timeline failed: %s" % r)
	return {}


# --- WebSocket --------------------------------------------------------------

func connect_ws() -> void:
	var url := websocket_url()
	_ws = WebSocketPeer.new()
	_ws.connect_to_url(url)


func websocket_url() -> String:
	var scheme := "wss" if use_tls else "ws"
	# Token via query param: WebSocketPeer cannot set the Authorization header.
	return "%s://%s:%d%s&access_token=%s" % [scheme, host, port, ws_url, token]


func ready_state() -> int:
	return _ws.get_ready_state()


# poll returns any envelopes received this frame as an Array of Dictionaries.
func poll() -> Array:
	_ws.poll()
	var out: Array = []
	while _ws.get_ready_state() == WebSocketPeer.STATE_OPEN and _ws.get_available_packet_count() > 0:
		var text := _ws.get_packet().get_string_from_utf8()
		var env = JSON.parse_string(text)
		if env != null:
			out.append(env)
	return out


func next_message_id() -> String:
	_msg_counter += 1
	return "cmsg-%d" % _msg_counter


func send(msg_type: String, tick: int, payload: Dictionary) -> String:
	var message_id := next_message_id()
	var env := {
		"type": msg_type,
		"message_id": message_id,
		"session_id": session_id,
		"tick": tick,
		"payload": payload,
	}
	_ws.send_text(JSON.stringify(env))
	return message_id


func close() -> void:
	_ws.close()
