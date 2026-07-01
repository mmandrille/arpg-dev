## CodexLoader — static singleton for compiled codex pages.
class_name CodexLoader
extends RefCounted

const INDEX_REL := "../shared/content/codex_index.v0.json"

static var chapters: Array = []
static var _pages_by_id: Dictionary = {}
static var _loaded: bool = false


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	chapters = []
	_pages_by_id = {}
	var path := ProjectSettings.globalize_path("res://").path_join(INDEX_REL)
	var data := _read_json(path)
	if data.is_empty():
		return
	for raw_chapter in data.get("chapters", []):
		if typeof(raw_chapter) != TYPE_DICTIONARY:
			continue
		var chapter: Dictionary = raw_chapter
		var pages: Array = []
		for raw_page in chapter.get("pages", []):
			if typeof(raw_page) != TYPE_DICTIONARY:
				continue
			var page: Dictionary = raw_page
			pages.append(page)
			var page_id := str(page.get("id", ""))
			if page_id != "":
				_pages_by_id[page_id] = page
		chapters.append({
			"id": str(chapter.get("id", "")),
			"title": str(chapter.get("title", "")),
			"pages": pages,
		})


static func chapter_ids() -> Array:
	ensure_loaded()
	var ids: Array = []
	for chapter in chapters:
		if typeof(chapter) != TYPE_DICTIONARY:
			continue
		ids.append(str((chapter as Dictionary).get("id", "")))
	return ids


static func chapter_title(chapter_id: String) -> String:
	ensure_loaded()
	for chapter in chapters:
		if typeof(chapter) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = chapter
		if str(rec.get("id", "")) == chapter_id:
			return str(rec.get("title", chapter_id))
	return chapter_id


static func pages_for_chapter(chapter_id: String) -> Array:
	ensure_loaded()
	for chapter in chapters:
		if typeof(chapter) != TYPE_DICTIONARY:
			continue
		var rec: Dictionary = chapter
		if str(rec.get("id", "")) == chapter_id:
			return (rec.get("pages", []) as Array).duplicate(true)
	return []


static func page(page_id: String) -> Dictionary:
	ensure_loaded()
	return (_pages_by_id.get(page_id, {}) as Dictionary).duplicate(true)


static func first_page_id(chapter_id: String) -> String:
	var pages := pages_for_chapter(chapter_id)
	if pages.is_empty():
		return ""
	var first: Dictionary = pages[0]
	return str(first.get("id", ""))


static func reset_for_tests() -> void:
	chapters = []
	_pages_by_id = {}
	_loaded = false


static func _read_json(path: String) -> Dictionary:
	if not FileAccess.file_exists(path):
		push_warning("codex index missing: %s" % path)
		return {}
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		return {}
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("codex index malformed: %s" % path)
		return {}
	return parsed
