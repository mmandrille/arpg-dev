class_name CombatBreakdownFormat
extends RefCounted


static func attack_title(event: Dictionary) -> String:
	var skill_id := str(event.get("skill_id", ""))
	if skill_id != "":
		return skill_id.replace("_", " ").capitalize()
	var weapon_slot := str(event.get("weapon_slot", ""))
	if weapon_slot != "":
		return "Basic attack (%s)" % weapon_slot
	return "Basic attack"


static func outcome_label(event: Dictionary) -> String:
	var outcome := str(event.get("outcome", "hit"))
	var damage := int(event.get("damage", 0))
	if outcome == "crit":
		return "Critical hit — %d" % damage
	if outcome == "miss":
		return "Miss"
	if outcome == "block":
		return "Blocked"
	return "Hit — %d" % damage


static func lines_text(event: Dictionary) -> String:
	var rows: PackedStringArray = PackedStringArray()
	for row in event.get("damage_breakdown", []):
		if typeof(row) != TYPE_DICTIONARY:
			continue
		var rec := row as Dictionary
		rows.append("%s: %s" % [str(rec.get("label", "")), str(rec.get("value", ""))])
	return "\n".join(rows)
