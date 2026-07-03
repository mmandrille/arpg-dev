class_name TrainingDamageLogBridge
extends RefCounted

const TrainingDamageLogPanelScript := preload("res://scripts/training_damage_log_panel.gd")
const TrainingDollVisualScript := preload("res://scripts/training_doll_visual.gd")


static func attach_panel(ui: CanvasLayer) -> TrainingDamageLogPanel:
	var panel := TrainingDamageLogPanelScript.new()
	ui.add_child(panel)
	return panel


static func make_silhouette_root() -> Node3D:
	var silhouette_root := Node3D.new()
	silhouette_root.name = "MonsterVisualRoot"
	silhouette_root.add_child(TrainingDollVisualScript.make_node())
	return silhouette_root


static func handle_training_doll_revived(owner, entity_id: String, ev: Dictionary) -> void:
	if not owner.entities.has(entity_id):
		return
	var revived_rec: Dictionary = owner.entities[entity_id]
	revived_rec["hp"] = int(ev.get("damage", revived_rec.get("max_hp", 1)))
	owner._set_pickable(revived_rec["node"] as Node3D, true)
	var revived_reaction = revived_rec.get("reaction")
	if revived_reaction != null and revived_reaction.has_method("reset_terminal"):
		revived_reaction.reset_terminal()
	var revived_ctrl = revived_rec.get("controller")
	if revived_ctrl != null and revived_ctrl.has_method("reset_terminal"):
		revived_ctrl.reset_terminal()


static func notify_combat_event(owner, panel: TrainingDamageLogPanel, entity_id: String, ev: Dictionary) -> void:
	if panel == null or not owner.entities.has(entity_id):
		return
	var rec: Dictionary = owner.entities[entity_id]
	var monster_def_id := str(rec.get("monster_def_id", ev.get("monster_def_id", "")))
	panel.on_training_doll_combat_event(ev, monster_def_id)


static func bot_debug_state(panel: TrainingDamageLogPanel) -> Dictionary:
	return {
		"training_damage_log_panel_visible": panel != null and panel.is_open(),
		"training_damage_log_panel": panel.get_debug_state() if panel != null else {},
	}


static func bot_inject_event(panel: TrainingDamageLogPanel, event: Dictionary) -> void:
	if panel == null:
		return
	panel.on_training_doll_combat_event(event, "town_training_doll")


static func bot_click_close(panel: TrainingDamageLogPanel) -> void:
	if panel != null:
		panel.bot_click_close()
