class_name MercenaryPanelBridge
extends RefCounted


static func show_board(owner, panel: MercenaryPanel, ev: Dictionary, gold: int) -> void:
	if panel == null:
		return
	owner._close_gameplay_panels("mercenary")
	panel.show_board_from_event(ev, gold)
	owner._sync_companion_bar()
	owner._raise_gameplay_windows()


static func apply_hired(owner, panel: MercenaryPanel, ev: Dictionary) -> void:
	if panel == null:
		return
	panel.apply_hired_event(ev)
	owner._sync_companion_bar()
	owner._raise_gameplay_windows()


static func apply_stance_changed(owner, panel: MercenaryPanel, ev: Dictionary) -> void:
	if panel == null:
		return
	panel.apply_stance_changed(ev)
	owner._sync_companion_bar()
	owner._raise_gameplay_windows()


static func apply_lost(owner, panel: MercenaryPanel, ev: Dictionary) -> void:
	if panel == null:
		return
	panel.apply_lost_event(ev)
	var lost_id := str(ev.get("target_entity_id", ev.get("entity_id", "")))
	if lost_id != "" and owner.entities.has(lost_id):
		owner._remove_entity(lost_id)
	else:
		owner._sync_companion_bar()
	owner._raise_gameplay_windows()


static func try_handle_event(owner, panel: MercenaryPanel, ev: Dictionary, gold: int) -> bool:
	match str(ev.get("event_type", "")):
		"mercenary_board_opened":
			show_board(owner, panel, ev, gold)
			return true
		"mercenary_hired":
			apply_hired(owner, panel, ev)
			return true
		"companion_stance_changed":
			apply_stance_changed(owner, panel, ev)
			return true
		"mercenary_lost":
			apply_lost(owner, panel, ev)
			return true
	return false
