class_name MarketBoardBadges
extends RefCounted

const INACTIVE_COLOR := Color("#776d5e")
const INCOMING_ACTIVE_COLOR := Color("#ffcf5a")
const PUBLISHED_ACTIVE_COLOR := Color("#9fd7ff")


static func apply_to_board(board: Node3D, incoming_bids: int, published_listings: int) -> void:
	if board == null:
		return
	_apply_badge(board, "IncomingBidBadge", "IncomingBidCount", incoming_bids, INCOMING_ACTIVE_COLOR)
	_apply_badge(board, "PublishedListingBadge", "PublishedListingCount", published_listings, PUBLISHED_ACTIVE_COLOR)


static func debug_state(board: Node3D) -> Dictionary:
	if board == null:
		return empty_state()
	var incoming := _label(board, "IncomingBidCount")
	var published := _label(board, "PublishedListingCount")
	var incoming_badge := _badge(board, "IncomingBidBadge")
	var published_badge := _badge(board, "PublishedListingBadge")
	return {
		"exists": true,
		"incoming_bids": _label_count(incoming),
		"published_listings": _label_count(published),
		"incoming_text": incoming.text if incoming != null else "",
		"published_text": published.text if published != null else "",
		"incoming_visible": incoming_badge != null and incoming_badge.visible,
		"published_visible": published_badge != null and published_badge.visible,
		"incoming_color": _label_color(incoming),
		"published_color": _label_color(published),
	}


static func empty_state() -> Dictionary:
	return {
		"exists": false,
		"incoming_bids": 0,
		"published_listings": 0,
		"incoming_text": "",
		"published_text": "",
		"incoming_visible": false,
		"published_visible": false,
		"incoming_color": "",
		"published_color": "",
	}


static func _apply_badge(board: Node3D, badge_name: String, label_name: String, count: int, active_color: Color) -> void:
	var safe_count: int = count
	if safe_count < 0:
		safe_count = 0
	var badge := _badge(board, badge_name)
	var label := _label(board, label_name)
	if badge != null:
		badge.visible = safe_count > 0
	if label != null:
		label.text = _count_text(safe_count)
		label.modulate = active_color if safe_count > 0 else INACTIVE_COLOR
		label.set_meta("market_count", safe_count)


static func _count_text(count: int) -> String:
	return "99+" if count > 99 else str(count)


static func _badge(board: Node3D, badge_name: String) -> Node3D:
	return board.find_child(badge_name, true, false) as Node3D


static func _label(board: Node3D, label_name: String) -> Label3D:
	return board.find_child(label_name, true, false) as Label3D


static func _label_count(label: Label3D) -> int:
	if label == null:
		return 0
	if label.has_meta("market_count"):
		return int(label.get_meta("market_count"))
	return int(label.text.replace("+", ""))


static func _label_color(label: Label3D) -> String:
	return label.modulate.to_html(false) if label != null else ""
