extends RefCounted

static func play_source_attack_for_event(ev: Dictionary, entities: Dictionary) -> void:
	var attack_style := str(ev.get("attack_style", ""))
	if attack_style != "dive" and attack_style != "pounce":
		return
	var source_id := str(ev.get("source_entity_id", ""))
	if source_id == "" or not entities.has(source_id):
		return
	var source_rec: Dictionary = entities[source_id]
	if str(source_rec.get("type", "")) != "monster":
		return
	var ctrl = source_rec.get("controller", null)
	if ctrl != null:
		ctrl.play_one_shot(attack_style)
