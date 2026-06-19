class_name DamageTypeCombatText
extends RefCounted

const TYPES := {
	"physical": {"label": "PHYS", "color": Color("#f4d481"), "variant": "physical"},
	"fire": {"label": "FIRE", "color": Color("#ff7a2a"), "variant": "fire"},
	"cold": {"label": "COLD", "color": Color("#7dd6ff"), "variant": "cold"},
	"lightning": {"label": "LIGHT", "color": Color("#ffe66d"), "variant": "lightning"},
	"poison": {"label": "POIS", "color": Color("#55e66f"), "variant": "poison"},
	"force": {"label": "FORCE", "color": Color("#c7a9ff"), "variant": "force"},
}
const SPECIAL_OUTCOMES := {
	"miss": {"color": Color(0.82, 0.86, 0.92), "variant": "miss", "text": "MISS"},
	"block": {"color": Color(0.35, 0.78, 1.0), "variant": "block", "text": "BLOCK"},
	"immune": {"color": Color(1.0, 0.86, 0.28), "variant": "immune", "text": "IMMUNE"},
}


static func special_outcome(outcome: String) -> Dictionary:
	return SPECIAL_OUTCOMES.get(outcome, {})


static func number_for_event(ev: Dictionary, default_color: Color) -> Dictionary:
	var damage = ev.get("damage", null)
	var amount := 0 if damage == null else int(damage)
	var damage_type := _event_damage_type(ev)
	var typed: Dictionary = TYPES.get(damage_type, {})
	var critical := str(ev.get("outcome", "")) == "crit" or bool(ev.get("critical", false))
	if critical:
		var crit_color = Color(1.0, 0.58, 0.22)
		var crit_text := "%d!" % amount
		if not typed.is_empty():
			crit_color = typed.get("color", crit_color)
			crit_text = _format_text(typed, amount, true)
		return {
			"amount": amount,
			"color": crit_color,
			"variant": "crit",
			"text": crit_text,
			"damage_type": damage_type,
		}
	if typed.is_empty():
		return {}
	return {
		"amount": amount,
		"color": typed.get("color", default_color),
		"variant": str(typed.get("variant", "normal")),
		"text": _format_text(typed, amount, false),
		"damage_type": damage_type,
	}


static func _event_damage_type(ev: Dictionary) -> String:
	var value := str(ev.get("damage_type", "")).strip_edges().to_lower()
	return value if TYPES.has(value) else ""


static func _format_text(presentation: Dictionary, amount: int, critical: bool) -> String:
	var suffix := "!" if critical else ""
	return "%s %d%s" % [str(presentation.get("label", "")).to_upper(), amount, suffix]
