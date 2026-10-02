class_name DeltaFrameCoalesce
extends RefCounted


static func merge_pending(payloads: Array) -> Dictionary:
	var merged_events: Array = []
	var merged_changes: Array = []
	var merged_perf: Dictionary = {}
	for payload in payloads:
		if payload is Dictionary:
			var p: Dictionary = payload
			var source_tick := int(p.get("_coalesce_source_tick", 0))
			for event in p.get("events", []):
				if event is Dictionary:
					var tagged_event := (event as Dictionary).duplicate(true)
					tagged_event["_coalesce_source_tick"] = source_tick
					merged_events.append(tagged_event)
				else:
					merged_events.append(event)
			merged_changes.append_array(p.get("changes", []))
			if p.has("performance") and p.get("performance") is Dictionary:
				merged_perf = (p.get("performance") as Dictionary).duplicate(true)
	return {
		"events": merged_events,
		"changes": merged_changes,
		"performance": merged_perf,
	}
