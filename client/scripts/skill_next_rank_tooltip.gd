class_name SkillNextRankTooltip
extends RefCounted

const PREFIX := "next rank: "
const RICH_COLOR := "#6fd66f"


static func percent_delta(subject: String, now: int, next: int) -> String:
	var delta := next - now
	if delta == 0:
		return ""
	return "%s%s%% %s" % [PREFIX, _signed(delta), subject.strip_edges().to_lower()]


static func value_delta(subject: String, now: int, next: int) -> String:
	var delta := next - now
	if delta == 0:
		return ""
	return "%s%s %s" % [PREFIX, _signed(delta), subject.strip_edges().to_lower()]


static func is_next_rank_line(text: String) -> bool:
	return text.begins_with(PREFIX)


static func rich_line(text: String, escape: Callable) -> String:
	if is_next_rank_line(text):
		return "[color=%s]%s[/color]" % [RICH_COLOR, escape.call(text)]
	return escape.call(text)


static func _signed(delta: int) -> String:
	return "%+d" % delta
