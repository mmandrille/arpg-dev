class_name LootLabelFilter
extends RefCounted

## Client-side, display-only rarity threshold filter for ground loot labels (v153).
##
## Owns no authoritative state: it only decides which already-revealed loot
## labels this client draws when the reveal key is held. The server still owns
## every loot roll, item, and pickup. See
## docs/specs/v153_spec-loot-label-filter-core.md.
##
## Threshold cycles All -> Magic+ -> Rare+ -> Unique. Rarities not on the ladder
## (currency, quest, consumable, unknown) are never hidden.

const ClientConstantsScript := preload("res://scripts/client_constants.gd")

const RARITY_ORDER: Array[String] = ["common", "magic", "rare", "unique"]
const MODE_LABELS: Array[String] = ["All", "Magic+", "Rare+", "Unique"]
const REVEAL_DIM_FACTOR: float = 0.58

var _threshold: int = 0


## Dim a label's base color when it is revealed-but-not-hovered, so the hovered
## label stands out. Currency labels use white while highlighted so gold text
## does not read like unique-rarity loot.
func display_color(base: Color, highlighted: bool, category: String = "", rarity: String = "common") -> Color:
	if highlighted and category.to_lower() == "currency":
		return Color.WHITE
	if highlighted:
		return ClientConstantsScript.target_highlight_color("loot", rarity)
	return Color(
		base.r * REVEAL_DIM_FACTOR,
		base.g * REVEAL_DIM_FACTOR,
		base.b * REVEAL_DIM_FACTOR,
		base.a,
	)


func allows(rarity: String) -> bool:
	var rank := RARITY_ORDER.find(rarity.to_lower())
	if rank < 0:
		# Off-ladder loot (currency / quest / consumable / unknown) is never hidden.
		return true
	return rank >= _threshold


func cycle() -> void:
	_threshold = (_threshold + 1) % RARITY_ORDER.size()


func threshold_rarity() -> String:
	return RARITY_ORDER[_threshold]


func mode_label() -> String:
	return MODE_LABELS[_threshold]


func set_mode_label(label: String) -> void:
	var normalized := label.strip_edges().to_lower()
	for i in range(MODE_LABELS.size()):
		if MODE_LABELS[i].to_lower() == normalized:
			_threshold = i
			return
	_threshold = 0


func is_active() -> bool:
	return _threshold > 0
