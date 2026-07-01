extends Control
class_name CodexPanel

const TextCatalogScript := preload("res://scripts/text_catalog.gd")
const CodexLoaderScript := preload("res://scripts/codex_loader.gd")

signal back_requested

var _title: Label
var _chapter_list: ItemList
var _page_list: ItemList
var _page_title: Label
var _page_subtitle: Label
var _page_body: RichTextLabel
var _back_button: Button
var _selected_chapter_id: String = ""
var _selected_page_id: String = ""


func _ready() -> void:
	_sync_viewport_size()
	get_viewport().size_changed.connect(_sync_viewport_size)
	mouse_filter = Control.MOUSE_FILTER_STOP
	_build()
	visible = false


func show_codex() -> void:
	ensure_built()
	_sync_viewport_size()
	CodexLoaderScript.ensure_loaded()
	if _selected_chapter_id == "":
		var chapter_ids := CodexLoaderScript.chapter_ids()
		if not chapter_ids.is_empty():
			_select_chapter(str(chapter_ids[0]))
	visible = true


func ensure_built() -> void:
	if _chapter_list == null:
		_build()


func hide_panel() -> void:
	visible = false


func _sync_viewport_size() -> void:
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


func _build() -> void:
	var bg := ColorRect.new()
	bg.color = Color(0.04, 0.045, 0.05, 0.96)
	bg.set_anchors_preset(Control.PRESET_FULL_RECT)
	add_child(bg)

	var root := MarginContainer.new()
	root.set_anchors_preset(Control.PRESET_FULL_RECT)
	root.add_theme_constant_override("margin_left", 24)
	root.add_theme_constant_override("margin_right", 24)
	root.add_theme_constant_override("margin_top", 20)
	root.add_theme_constant_override("margin_bottom", 20)
	add_child(root)

	var outer := VBoxContainer.new()
	outer.add_theme_constant_override("separation", 12)
	root.add_child(outer)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 12)
	outer.add_child(header)

	_title = Label.new()
	_title.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_title.add_theme_font_size_override("font_size", 34)
	_title.add_theme_color_override("font_color", Color("#f1efe4"))
	header.add_child(_title)

	_back_button = Button.new()
	_back_button.custom_minimum_size = Vector2(120, 40)
	_back_button.pressed.connect(back_requested.emit)
	header.add_child(_back_button)

	var body := HBoxContainer.new()
	body.size_flags_vertical = Control.SIZE_EXPAND_FILL
	body.add_theme_constant_override("separation", 16)
	outer.add_child(body)

	var nav := VBoxContainer.new()
	nav.custom_minimum_size = Vector2(220, 0)
	nav.add_theme_constant_override("separation", 8)
	body.add_child(nav)

	_chapter_list = ItemList.new()
	_chapter_list.custom_minimum_size = Vector2(220, 140)
	_chapter_list.item_selected.connect(_on_chapter_selected)
	nav.add_child(_chapter_list)

	_page_list = ItemList.new()
	_page_list.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_page_list.item_selected.connect(_on_page_selected)
	nav.add_child(_page_list)

	var page_panel := PanelContainer.new()
	page_panel.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	page_panel.size_flags_vertical = Control.SIZE_EXPAND_FILL
	body.add_child(page_panel)

	var page_margin := MarginContainer.new()
	page_margin.add_theme_constant_override("margin_left", 16)
	page_margin.add_theme_constant_override("margin_right", 16)
	page_margin.add_theme_constant_override("margin_top", 12)
	page_margin.add_theme_constant_override("margin_bottom", 12)
	page_panel.add_child(page_margin)

	var page_box := VBoxContainer.new()
	page_box.add_theme_constant_override("separation", 8)
	page_margin.add_child(page_box)

	_page_title = Label.new()
	_page_title.add_theme_font_size_override("font_size", 28)
	_page_title.add_theme_color_override("font_color", Color("#f7f3df"))
	page_box.add_child(_page_title)

	_page_subtitle = Label.new()
	_page_subtitle.add_theme_font_size_override("font_size", 14)
	_page_subtitle.add_theme_color_override("font_color", Color("#b9b4a7"))
	page_box.add_child(_page_subtitle)

	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	page_box.add_child(scroll)

	_page_body = RichTextLabel.new()
	_page_body.bbcode_enabled = false
	_page_body.fit_content = true
	_page_body.scroll_active = false
	_page_body.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	_page_body.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	_page_body.add_theme_font_size_override("normal_font_size", 16)
	_page_body.add_theme_color_override("default_color", Color("#ddd8cb"))
	scroll.add_child(_page_body)

	refresh_texts()
	_reload_chapters()


func refresh_texts() -> void:
	if _title != null:
		_title.text = TextCatalogScript.get_text("menu.codex", "Codex")
	if _back_button != null:
		_back_button.text = TextCatalogScript.get_text("menu.back", "Back")


func select_page(page_id: String) -> void:
	CodexLoaderScript.ensure_loaded()
	for chapter in CodexLoaderScript.chapters:
		if typeof(chapter) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = chapter
		var chapter_id := str(rec.get("id", ""))
		for raw_page in rec.get("pages", []):
			if typeof(raw_page) != TYPE_DICTIONARY:
				continue
			if str((raw_page as Dictionary).get("id", "")) == page_id:
				_select_chapter(chapter_id)
				_select_page(page_id)
				return


func get_debug_state() -> Dictionary:
	return {
		"visible": visible,
		"chapter_id": _selected_chapter_id,
		"page_id": _selected_page_id,
		"page_title": _page_title.text if _page_title != null else "",
		"chapter_count": CodexLoaderScript.chapters.size(),
	}


func _reload_chapters() -> void:
	CodexLoaderScript.ensure_loaded()
	_chapter_list.clear()
	for chapter in CodexLoaderScript.chapters:
		if typeof(chapter) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = chapter
		var chapter_id := str(rec.get("id", ""))
		_chapter_list.add_item(str(rec.get("title", chapter_id)), null, false)
		_chapter_list.set_item_metadata(_chapter_list.item_count - 1, chapter_id)


func _select_chapter(chapter_id: String) -> void:
	_selected_chapter_id = chapter_id
	_page_list.clear()
	var pages := CodexLoaderScript.pages_for_chapter(chapter_id)
	for raw_page in pages:
		if typeof(raw_page) != TYPE_DICTIONARY:
			continue
		var page: Dictionary = raw_page
		_page_list.add_item(str(page.get("title", "")), null, false)
		_page_list.set_item_metadata(_page_list.item_count - 1, str(page.get("id", "")))
	for index in range(_chapter_list.item_count):
		if str(_chapter_list.get_item_metadata(index)) == chapter_id:
			_chapter_list.select(index)
			break
	if not pages.is_empty():
		var first: Dictionary = pages[0]
		_select_page(str(first.get("id", "")))


func _select_page(page_id: String) -> void:
	_selected_page_id = page_id
	var page := CodexLoaderScript.page(page_id)
	_page_title.text = str(page.get("title", ""))
	var subtitle := str(page.get("subtitle", ""))
	_page_subtitle.text = subtitle
	_page_subtitle.visible = subtitle != ""
	_page_body.text = _format_page(page)
	for index in range(_page_list.item_count):
		if str(_page_list.get_item_metadata(index)) == page_id:
			_page_list.select(index)
			break


func _format_page(page: Dictionary) -> String:
	var lines: PackedStringArray = []
	for raw_section in page.get("sections", []):
		if typeof(raw_section) != TYPE_DICTIONARY:
			continue
		var section: Dictionary = raw_section
		var heading := str(section.get("heading", ""))
		if heading != "":
			lines.append(heading)
			lines.append("")
		for line in section.get("lines", []):
			lines.append(str(line))
			lines.append("")
		for bullet in section.get("bullets", []):
			lines.append("• %s" % str(bullet))
		if not section.get("bullets", []).is_empty():
			lines.append("")
	return "\n".join(lines).strip_edges()


func _on_chapter_selected(index: int) -> void:
	if index < 0:
		return
	_select_chapter(str(_chapter_list.get_item_metadata(index)))


func _on_page_selected(index: int) -> void:
	if index < 0:
		return
	_select_page(str(_page_list.get_item_metadata(index)))
