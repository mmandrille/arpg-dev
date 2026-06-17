class_name BotDebugProgressionSetup
extends RefCounted


static func prepare_character(client, debug_token: String, debug_progression_json: String, debug_gold: String) -> String:
	var progression := _progression_from_env(debug_progression_json, debug_gold)
	if client == null or progression.is_empty():
		return ""
	var character_id := _first_character_id(client)
	if character_id == "":
		var created: Dictionary = client.create_character("Client Bot")
		character_id = str(created.get("character_id", ""))
	if character_id == "":
		printerr("[bot-client] debug progression character setup failed")
		return ""
	var ok: bool = client.set_debug_progression(debug_token, character_id, progression)
	if not ok:
		printerr("[bot-client] debug progression seed failed character=%s" % character_id)
		return ""
	print("[bot-client] debug progression seeded character=%s keys=%s" % [character_id, str(progression.keys())])
	return character_id


static func _progression_from_env(debug_progression_json: String, debug_gold: String) -> Dictionary:
	if debug_progression_json != "":
		var parsed = JSON.parse_string(debug_progression_json)
		if typeof(parsed) == TYPE_DICTIONARY:
			return parsed as Dictionary
		printerr("[bot-client] debug progression JSON invalid")
	if debug_gold != "":
		return {"gold": int(debug_gold)}
	return {}


static func _first_character_id(client) -> String:
	for row in client.list_characters():
		if typeof(row) == TYPE_DICTIONARY:
			return str((row as Dictionary).get("character_id", ""))
	return ""
