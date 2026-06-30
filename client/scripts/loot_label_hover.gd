## Screen-space hover pick for visible ground loot labels (Alt reveal mode).
class_name LootLabelHover
extends RefCounted

const LABEL_PIXEL_SIZE := 0.0018
const LABEL_BASE_Y := 0.58
const MIN_SCREEN_RADIUS_PX := 18.0
const CHAR_WIDTH_PX := 9.0
const LINE_HEIGHT_PX := 22.0


static func pick_loot_id(
	camera: Camera3D,
	viewport: Viewport,
	mouse_pos: Vector2,
	label_ids: Array,
	entities: Dictionary,
) -> String:
	if camera == null or viewport == null:
		return ""

	var best_id := ""
	var best_dist := MIN_SCREEN_RADIUS_PX
	for raw_id in label_ids:
		var loot_id := str(raw_id)
		if not entities.has(loot_id):
			continue
		var rec: Dictionary = entities[loot_id]
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var label := node.find_child("LootLabel", true, false) as Label3D
		if label == null or not label.visible:
			continue
		if camera.is_position_behind(label.global_position):
			continue
		var screen := camera.unproject_position(label.global_position)
		var text_len := maxi(label.text.length(), 1)
		var half_w := maxf(text_len * CHAR_WIDTH_PX * 0.5, MIN_SCREEN_RADIUS_PX * 0.65)
		var half_h := LINE_HEIGHT_PX * 0.55
		var local := mouse_pos - screen
		var dx := maxf(0.0, absf(local.x) - half_w)
		var dy := maxf(0.0, absf(local.y) - half_h)
		var dist := sqrt(dx * dx + dy * dy)
		if dist < best_dist:
			best_dist = dist
			best_id = loot_id

	return best_id


static func compute_hover_update(
	crosshair_locked_id: String,
	reveal_held: bool,
	prev_reveal_held: bool,
	prev_hover: String,
	camera: Camera3D,
	viewport: Viewport,
	mouse_pos: Vector2,
	label_ids: Array,
	entities: Dictionary,
) -> Dictionary:
	var reveal_changed := reveal_held != prev_reveal_held
	var next_hover := crosshair_locked_id
	if next_hover == "" and reveal_held:
		next_hover = pick_loot_id(camera, viewport, mouse_pos, label_ids, entities)
	var changed := next_hover != prev_hover or reveal_changed
	return {"hover_id": next_hover, "reveal_held": reveal_held, "changed": changed}
