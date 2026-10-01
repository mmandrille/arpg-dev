class_name PaperDollLayout
extends RefCounted
## Slot geometry for the inventory paper-doll (v516): a 3-column grid with no overlapping
## slots, in the 340x360 paper area. Pure data; the inventory panel and the showme capture
## both read it so they cannot drift apart.

const POSITIONS := {
	"head": Vector2(122, 8),
	"amulet": Vector2(236, 8),
	"main_hand": Vector2(8, 76),
	"chest": Vector2(122, 76),
	"off_hand": Vector2(236, 76),
	"ring_left": Vector2(8, 144),
	"belt": Vector2(122, 144),
	"ring_right": Vector2(236, 144),
	"gloves": Vector2(8, 212),
	"boots": Vector2(122, 212),
}


static func position_for(slot: String) -> Vector2:
	return POSITIONS.get(slot, Vector2.ZERO)


static func slot_rects(slot_size: Vector2) -> Dictionary:
	var out := {}
	for slot in POSITIONS:
		out[slot] = Rect2(POSITIONS[slot], slot_size)
	return out
