class_name BotReconnectProofActions
extends RefCounted

static var _enabled: bool = false


static func reset() -> void:
	_enabled = false


static func enable(_main) -> void:
	_enabled = true


static func is_enabled() -> bool:
	return _enabled


static func simulate_ws_drop(main) -> void:
	if main == null or main.client == null:
		return
	main.client.close()
