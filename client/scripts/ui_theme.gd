class_name UiTheme
extends RefCounted
## Shared client UI tokens (colors, spacing, font roles, frame recipes) from
## shared/assets/ui_theme.v0.json. Static singleton: no autoload, no scene tree.
## Presentation only; panels build styles through frame()/item_slot_frame().

const CATALOG_PATH := "../shared/assets/ui_theme.v0.json"
const FALLBACK_COLOR := Color(1.0, 0.0, 1.0, 1.0)
const FRAME_BASE_STATE := ""

static var _loaded: bool = false
static var _colors: Dictionary = {}
static var _spacing: Dictionary = {}
static var _fonts: Dictionary = {}
static var _frames: Dictionary = {}
static var _rarity_backgrounds: Dictionary = {}
static var _modifiers: Dictionary = {}
static var _reported: Dictionary = {}


static func invalidate() -> void:
	_loaded = false
	_colors = {}
	_spacing = {}
	_fonts = {}
	_frames = {}
	_rarity_backgrounds = {}
	_modifiers = {}
	_reported = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	load_from_path(ProjectSettings.globalize_path("res://").path_join(CATALOG_PATH))


## Loads (or reloads) a catalog from an absolute path. A missing or corrupt file
## leaves the catalog empty so every lookup returns its documented fallback.
static func load_from_path(path: String) -> bool:
	invalidate()
	_loaded = true
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_error("UiTheme: cannot read UI theme catalog at %s" % path)
		return false
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) != TYPE_DICTIONARY:
		push_error("UiTheme: UI theme catalog is not a JSON object")
		return false
	for token in parsed.get("colors", {}):
		_colors[token] = _parse_color(parsed["colors"][token])
	for rarity in parsed.get("rarity_slot_backgrounds", {}):
		_rarity_backgrounds[rarity] = _parse_color(parsed["rarity_slot_backgrounds"][rarity])
	_spacing = parsed.get("spacing", {})
	_fonts = parsed.get("fonts", {})
	_frames = parsed.get("frames", {})
	_modifiers = parsed.get("state_modifiers", {})
	return true


static func has_token(kind: String, token: String) -> bool:
	ensure_loaded()
	match kind:
		"color":
			return _colors.has(token)
		"spacing":
			return _spacing.has(token)
		"font":
			return _fonts.has(token)
		"frame":
			return _frames.has(token)
		"rarity":
			return _rarity_backgrounds.has(token)
	return false


static func color(token: String) -> Color:
	ensure_loaded()
	if _colors.has(token):
		return _colors[token]
	_report_once("color:" + token)
	return FALLBACK_COLOR


static func spacing(token: String) -> int:
	ensure_loaded()
	if _spacing.has(token):
		return int(_spacing[token])
	_report_once("spacing:" + token)
	return 0


static func rarity_background(rarity: String) -> Color:
	ensure_loaded()
	var key := rarity.to_lower()
	if _rarity_backgrounds.has(key):
		return _rarity_backgrounds[key]
	if _rarity_backgrounds.has("common"):
		return _rarity_backgrounds["common"]
	_report_once("rarity:common")
	return FALLBACK_COLOR


static func font_size(role: String) -> int:
	var font := _font(role)
	return int(font.get("size", 0))


static func font_color(role: String) -> Color:
	var font := _font(role)
	return color(str(font.get("color", ""))) if not font.is_empty() else FALLBACK_COLOR


static func apply_font(control: Control, role: String) -> void:
	var font := _font(role)
	if font.is_empty():
		return
	control.add_theme_font_size_override("font_size", int(font.get("size", 0)))
	control.add_theme_color_override("font_color", color(str(font.get("color", ""))))
	if font.has("outline_size"):
		control.add_theme_constant_override("outline_size", int(font["outline_size"]))
	if font.has("outline_color"):
		control.add_theme_color_override("font_outline_color", color(str(font["outline_color"])))


## Fresh StyleBoxFlat for a named frame recipe; callers may mutate it.
static func frame(name: String, state: String = FRAME_BASE_STATE) -> StyleBoxFlat:
	ensure_loaded()
	var s := StyleBoxFlat.new()
	if not _frames.has(name):
		_report_once("frame:" + name)
		s.bg_color = FALLBACK_COLOR
		return s
	var recipe: Dictionary = _frames[name]
	var overrides: Dictionary = recipe.get("states", {}).get(state, {})
	s.bg_color = color(str(overrides.get("bg", recipe.get("bg", ""))))
	s.border_color = color(str(overrides.get("border", recipe.get("border", ""))))
	var border := _box(recipe.get("border_width", 0))
	s.border_width_left = border[0]
	s.border_width_top = border[1]
	s.border_width_right = border[2]
	s.border_width_bottom = border[3]
	if recipe.has("radius"):
		var radius := _box(recipe["radius"])
		s.corner_radius_top_left = radius[0]
		s.corner_radius_top_right = radius[1]
		s.corner_radius_bottom_right = radius[2]
		s.corner_radius_bottom_left = radius[3]
	if recipe.has("margin"):
		var margin := _box(recipe["margin"])
		s.content_margin_left = margin[0]
		s.content_margin_top = margin[1]
		s.content_margin_right = margin[2]
		s.content_margin_bottom = margin[3]
	return s


## Rarity-tinted item slot. dimmed marks an unusable item (invalid requirements).
static func item_slot_frame(rarity: String, hover: bool = false, dimmed: bool = false, base: String = "slot") -> StyleBoxFlat:
	var s := frame(base, "hover" if hover else FRAME_BASE_STATE)
	var base_color := rarity_background(rarity)
	if dimmed:
		base_color = base_color.darkened(float(_modifiers.get("invalid_darken", 0.0)))
	var hover_lighten := float(_modifiers.get("hover_lighten", 0.0))
	s.bg_color = base_color.lightened(hover_lighten) if hover else base_color
	if dimmed:
		s.border_color = color(str(_modifiers.get("invalid_border_hover" if hover else "invalid_border", "")))
	else:
		var amount := float(_modifiers.get("border_hover_lighten" if hover else "border_lighten", 0.0))
		s.border_color = base_color.lightened(amount)
	return s


static func catalog() -> Dictionary:
	ensure_loaded()
	return {
		"colors": _colors.duplicate(),
		"rarity_slot_backgrounds": _rarity_backgrounds.duplicate(),
		"spacing": _spacing.duplicate(),
		"fonts": _fonts.duplicate(true),
		"frames": _frames.duplicate(true),
		"state_modifiers": _modifiers.duplicate(),
	}


static func _font(role: String) -> Dictionary:
	ensure_loaded()
	if _fonts.has(role):
		return _fonts[role]
	_report_once("font:" + role)
	return {}


static func _parse_color(value) -> Color:
	if typeof(value) == TYPE_ARRAY and value.size() == 4:
		return Color(float(value[0]), float(value[1]), float(value[2]), float(value[3]))
	if typeof(value) == TYPE_STRING:
		return Color(value)
	return FALLBACK_COLOR


## Resolves an int-or-list of spacing tokens to four ints.
static func _box(ref) -> Array:
	if typeof(ref) == TYPE_ARRAY and ref.size() == 4:
		return [spacing(str(ref[0])), spacing(str(ref[1])), spacing(str(ref[2])), spacing(str(ref[3]))]
	var v := spacing(str(ref))
	return [v, v, v, v]


static func _report_once(key: String) -> void:
	if _reported.has(key):
		return
	_reported[key] = true
	push_error("UiTheme: unknown token '%s'" % key)
