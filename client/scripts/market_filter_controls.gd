class_name MarketFilterControls
extends HBoxContainer

signal filter_changed

const SORT_DEFAULT := "default"
const SORT_NAME := "name"
const SORT_PRICE_LOW := "price_low"
const SORT_PRICE_HIGH := "price_high"
const SORT_STATUS := "status"
const SORT_MODES := [SORT_DEFAULT, SORT_NAME, SORT_PRICE_LOW, SORT_PRICE_HIGH, SORT_STATUS]
const SORT_LABELS := ["Default", "Name", "Price up", "Price down", "Status"]

var _search_field: LineEdit
var _sort_option: OptionButton
var _query: String = ""
var _sort_mode: String = SORT_DEFAULT


func _init() -> void:
	add_theme_constant_override("separation", 8)
	_search_field = LineEdit.new()
	_search_field.placeholder_text = "Search market"
	_search_field.clear_button_enabled = true
	_search_field.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_search_field.custom_minimum_size = Vector2(260, 34)
	_search_field.text_changed.connect(func(text: String) -> void:
		_query = text.strip_edges()
		filter_changed.emit()
	)
	add_child(_search_field)

	_sort_option = OptionButton.new()
	_sort_option.custom_minimum_size = Vector2(134, 34)
	for i in range(SORT_MODES.size()):
		_sort_option.add_item(SORT_LABELS[i])
		_sort_option.set_item_metadata(i, SORT_MODES[i])
	_sort_option.item_selected.connect(func(index: int) -> void:
		_set_sort_mode(str(_sort_option.get_item_metadata(index)))
		filter_changed.emit()
	)
	add_child(_sort_option)


func query() -> String:
	return _query


func sort_mode() -> String:
	return _sort_mode


func sort_options() -> Array:
	return SORT_MODES.duplicate()


func bot_set_search(text: String) -> void:
	_query = text.strip_edges()
	if _search_field != null and _search_field.text != _query:
		_search_field.text = _query
	filter_changed.emit()


func bot_select_sort(mode: String) -> void:
	_set_sort_mode(mode)
	filter_changed.emit()


func debug_state() -> Dictionary:
	return {
		"market_filter_visible": visible,
		"market_search_text": _query,
		"market_sort_mode": _sort_mode,
		"market_sort_options": sort_options(),
	}


func _set_sort_mode(mode: String) -> void:
	_sort_mode = mode if SORT_MODES.has(mode) else SORT_DEFAULT
	if _sort_option == null:
		return
	for i in range(_sort_option.item_count):
		if str(_sort_option.get_item_metadata(i)) == _sort_mode:
			_sort_option.select(i)
			return
