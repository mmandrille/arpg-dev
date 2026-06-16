class_name MercenaryPanelBridge
extends RefCounted


static func show_board(owner, panel: MercenaryPanel, ev: Dictionary, gold: int) -> void:
	if panel == null:
		return
	owner._close_gameplay_panels("mercenary")
	panel.show_board(
		str(ev.get("entity_id", "")),
		str(ev.get("service", "mercenary")),
		str(ev.get("offer_id", "fixed:mercenary_guard")),
		str(ev.get("monster_def_id", "mercenary_guard")),
		int(ev.get("price", 0)),
		bool(ev.get("affordable", gold >= int(ev.get("price", 0)))),
		int(ev.get("total_gold", gold))
	)
	owner._sync_companion_bar()
	owner._raise_gameplay_windows()


static func apply_hired(owner, panel: MercenaryPanel, ev: Dictionary) -> void:
	if panel == null:
		return
	panel.apply_hired_event(ev)
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
	return false
