class_name LocalPlayerAuthoritativeSync
extends RefCounted


static func effect_ids_changed(incoming: Array, cached: Array) -> bool:
	if incoming.size() != cached.size():
		return true
	for i in incoming.size():
		if str(incoming[i]) != str(cached[i]):
			return true

	return false


static func hp_changed(e: Dictionary, hp: int, max_hp: int) -> bool:
	if not e.has("hp"):
		return false

	return int(e["hp"]) != hp or int(e.get("max_hp", max_hp)) != max_hp


static func mana_changed(e: Dictionary, mana: int, max_mana: int) -> bool:
	if not e.has("mana"):
		return false

	return int(e["mana"]) != mana or int(e.get("max_mana", max_mana)) != max_mana


static func visual_scale_changed(e: Dictionary, scale: float) -> bool:
	if not e.has("visual_scale"):
		return false

	return absf(float(e["visual_scale"]) - scale) > 0.0001
