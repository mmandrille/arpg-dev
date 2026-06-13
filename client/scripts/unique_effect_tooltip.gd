class_name UniqueEffectTooltip
extends RefCounted

const EFFECT_TITLE_COLOR := Color("#ffb26b")
const EFFECT_TEXT_COLOR := Color("#e8dcc8")
const EFFECT_FONT_SIZE := 19


static func text_lines_for_item(item: Dictionary) -> Array:
	var out: Array = []
	for line in rich_lines_for_item(item):
		out.append(str((line as Dictionary).get("text", "")))
	return out


static func rich_lines_for_item(item: Dictionary) -> Array:
	var out: Array = []
	var effect_ids = item.get("effect_ids", [])
	if typeof(effect_ids) != TYPE_ARRAY:
		return out
	for raw_id in effect_ids:
		var effect_id := str(raw_id)
		if effect_id == "":
			continue
		var effect: Dictionary = ItemRulesLoader.unique_effect_definition(effect_id)
		var title := str(effect.get("display_name", _fallback_effect_name(effect_id)))
		var summary := str(effect.get("summary", ""))
		out.append({"text": "Unique effect: %s" % title, "color": EFFECT_TITLE_COLOR, "font_size": EFFECT_FONT_SIZE})
		if summary != "":
			out.append({"text": summary, "color": EFFECT_TEXT_COLOR, "font_size": EFFECT_FONT_SIZE})
	return out


static func _fallback_effect_name(effect_id: String) -> String:
	return effect_id.replace("_", " ").capitalize()
