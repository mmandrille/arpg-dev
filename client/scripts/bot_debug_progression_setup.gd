class_name BotDebugProgressionSetup
extends RefCounted


static func prepare_character(client, debug_token: String, debug_gold: String) -> String:
	if client == null or debug_gold == "":
		return ""
	var character_id := _first_character_id(client)
	if character_id == "":
		var created: Dictionary = client.create_character("Client Bot")
		character_id = str(created.get("character_id", ""))
	if character_id == "":
		printerr("[bot-client] debug progression character setup failed")
		return ""
	var ok: bool = client.set_debug_progression(debug_token, character_id, {"gold": int(debug_gold)})
	if not ok:
		printerr("[bot-client] debug progression seed failed character=%s" % character_id)
		return ""
	print("[bot-client] debug progression seeded character=%s gold=%s" % [character_id, debug_gold])
	return character_id


static func _first_character_id(client) -> String:
	for row in client.list_characters():
		if typeof(row) == TYPE_DICTIONARY:
			return str((row as Dictionary).get("character_id", ""))
	return ""
