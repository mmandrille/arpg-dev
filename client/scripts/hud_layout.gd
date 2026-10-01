class_name HudLayout
extends RefCounted
## Pure HUD rectangle math (v517). Every always-on HUD widget asks this helper where it lives so the
## placement rules are testable without a scene. No scene, no autoload, no main.gd dependency.

const EDGE_MARGIN := 12.0
const CLUSTER_GAP := 10.0
const SLOT_ROW_TOP := 78.0 # distance from the viewport bottom to the top of the slot row
const SLOT_SIZE := Vector2(64.0, 64.0)
const HOTBAR_SIZE := Vector2(580.0, 64.0)
const CHARACTER_SLOT_OFFSET := -366.0 # relative to the horizontal viewport center
const SKILL_SLOT_OFFSET := 302.0
const XP_BAR_HEIGHT := 8.0
const XP_BAR_BOTTOM := 12.0
const GLOBE_MIN := 56.0
const GLOBE_MAX := 140.0
const GLOBE_HEIGHT_FRACTION := 0.15
const IDENTITY_HEIGHT := 22.0
const MINIMAP_COMPACT_SIZE := Vector2(216.0, 216.0)
const MINIMAP_RIGHT_INSET := 18.0
const MINIMAP_TOP := 44.0


static func hotbar_rect(vp: Vector2) -> Rect2:
	return Rect2(Vector2((vp.x - HOTBAR_SIZE.x) * 0.5, vp.y - SLOT_ROW_TOP), HOTBAR_SIZE)


static func xp_bar_rect(vp: Vector2) -> Rect2:
	return Rect2(Vector2((vp.x - HOTBAR_SIZE.x) * 0.5, vp.y - XP_BAR_BOTTOM), Vector2(HOTBAR_SIZE.x, XP_BAR_HEIGHT))


static func character_slot_rect(vp: Vector2) -> Rect2:
	return Rect2(Vector2(vp.x * 0.5 + CHARACTER_SLOT_OFFSET, vp.y - SLOT_ROW_TOP), SLOT_SIZE)


static func skill_slot_rect(vp: Vector2) -> Rect2:
	return Rect2(Vector2(vp.x * 0.5 + SKILL_SLOT_OFFSET, vp.y - SLOT_ROW_TOP), SLOT_SIZE)


## Slot row cluster: character slot, hotbar and skill slot.
static func cluster_rect(vp: Vector2) -> Rect2:
	return character_slot_rect(vp).merge(hotbar_rect(vp)).merge(skill_slot_rect(vp))


## Diameter of both globes: scales with viewport height, never intrudes into the slot cluster.
static func globe_diameter(vp: Vector2) -> float:
	var wanted := clampf(vp.y * GLOBE_HEIGHT_FRACTION, GLOBE_MIN, GLOBE_MAX)
	var room := cluster_rect(vp).position.x - EDGE_MARGIN - CLUSTER_GAP
	return maxf(minf(wanted, room), 0.0)


static func health_globe_rect(vp: Vector2) -> Rect2:
	var d := globe_diameter(vp)
	return Rect2(Vector2(EDGE_MARGIN, vp.y - EDGE_MARGIN - d), Vector2(d, d))


static func mana_globe_rect(vp: Vector2) -> Rect2:
	var d := globe_diameter(vp)
	return Rect2(Vector2(vp.x - EDGE_MARGIN - d, vp.y - EDGE_MARGIN - d), Vector2(d, d))


## Name and level strip seated directly above the health globe.
static func identity_rect(vp: Vector2) -> Rect2:
	var globe := health_globe_rect(vp)
	var width := maxf(globe.size.x, 110.0)
	return Rect2(Vector2(globe.position.x, globe.position.y - IDENTITY_HEIGHT - 4.0), Vector2(width, IDENTITY_HEIGHT))


static func minimap_compact_rect(vp: Vector2) -> Rect2:
	return Rect2(Vector2(vp.x - MINIMAP_RIGHT_INSET - MINIMAP_COMPACT_SIZE.x, MINIMAP_TOP), MINIMAP_COMPACT_SIZE)


static func boss_bar_rect(vp: Vector2, panel_top: float, panel_width: float, panel_height: float) -> Rect2:
	return Rect2(Vector2(maxf(8.0, (vp.x - panel_width) * 0.5), panel_top), Vector2(panel_width, panel_height))
