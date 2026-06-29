class_name DeltaFrameCoalesce
extends RefCounted


static func merge_pending(payloads: Array) -> Dictionary:
	var merged_events: Array = []
	var merged_changes: Array = []
	var merged_perf: Dictionary = {}
	for payload in payloads:
		if payload is Dictionary:
			var p: Dictionary = payload
			merged_events.append_array(p.get("events", []))
			merged_changes.append_array(p.get("changes", []))
			if p.has("performance") and p.get("performance") is Dictionary:
				merged_perf = (p.get("performance") as Dictionary).duplicate(true)
	return {
		"events": merged_events,
		"changes": merged_changes,
		"performance": merged_perf,
	}
