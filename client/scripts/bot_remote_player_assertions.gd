class_name BotRemotePlayerAssertions
extends RefCounted

## Remote co-op player expectations for wait/assert_remote_player_count.
## `rendered_classes` (v485) requires each listed class id to be the hero model
## applied to some remote player; `local_rendered_class` checks the local hero.


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var remote_ids: Array = state.get("remote_player_ids", [])
	if step.has("equals") and remote_ids.size() != int(step.get("equals", 0)):
		return false
	if step.has("at_least") and remote_ids.size() < int(step.get("at_least", 0)):
		return false
	var rendered: Dictionary = state.get("remote_player_classes", {})
	for class_id in step.get("rendered_classes", []):
		if not rendered.values().has(str(class_id)):
			return false
	if step.has("local_rendered_class") and str(state.get("local_player_rendered_class", "")) != str(step.get("local_rendered_class", "")):
		return false
	return step.has("equals") or step.has("at_least")
