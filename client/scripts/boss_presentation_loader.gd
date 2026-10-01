## Static loader for per-boss presentation (v512). Data: shared/assets/boss_presentation.v0.json.
## Keyed by boss_template_id. Client-only: scale/tint here override the wire visual_scale/visual_tint;
## gameplay (hp, patterns, hit shapes) stays server-owned.
class_name BossPresentationLoader
extends RefCounted

const DEFAULT_PATH := "../shared/assets/boss_presentation.v0.json"
const MounterScript := preload("res://scripts/boss_presentation_mounter.gd")
## Nodes owned by boss presentation that must keep their own materials when a model tint is applied.
const UNTINTED_NODE_NAMES := ["BossLaneMarker", "BossArenaPresence", "BossAuraDisc", "BossHeadgear"]

static var _loaded: bool = false
static var _bosses: Dictionary = {}


static func ensure_loaded() -> void:
	if _loaded:
		return
	_loaded = true
	_bosses = {}
	var path := ProjectSettings.globalize_path("res://").path_join(DEFAULT_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("BossPresentationLoader: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY and typeof((parsed as Dictionary).get("bosses", null)) == TYPE_DICTIONARY:
		_bosses = (parsed as Dictionary)["bosses"]
	else:
		push_warning("BossPresentationLoader: malformed JSON: %s" % path)


## Test hook: replace the catalog without touching disk.
static func set_bosses_for_test(bosses: Dictionary) -> void:
	_loaded = true
	_bosses = bosses


static func reset_for_test() -> void:
	_loaded = false
	_bosses = {}


static func has(template_id: String) -> bool:
	ensure_loaded()
	return template_id != "" and _bosses.has(template_id)


## Returns the cached catalog entry itself (no copy: it is read every boss every delta). Callers
## must treat it as read-only; copy with `.duplicate(true)` before mutating.
static func entry(template_id: String) -> Dictionary:
	ensure_loaded()
	if template_id == "" or not _bosses.has(template_id):
		return {}
	return _bosses[template_id] as Dictionary


## Catalog scale when the boss has presentation, else the wire value (never <= 0).
static func effective_scale(template_id: String, wire_scale: float) -> float:
	var cfg := entry(template_id)
	var scale := float(cfg.get("scale", wire_scale))
	return scale if scale > 0.0 else 1.0


static func base_tint(template_id: String, fallback: Color) -> Color:
	var cfg := entry(template_id)
	return Color(str(cfg["tint"])) if cfg.has("tint") else fallback


static func is_untinted(node_name: String) -> bool:
	return UNTINTED_NODE_NAMES.has(node_name)


## Builds the entity root for a catalogued boss (kit model + attachments + headgear), or null so the
## caller falls back to the wire-driven monster path.
static func make_root(e: Dictionary, apply_tint: Callable, tint: Color) -> Node3D:
	var cfg := entry(str(e.get("boss_template_id", "")))
	if cfg.is_empty():
		return null
	var packed := load("res://scenes/%s.tscn" % str(cfg.get("visual_key", ""))) as PackedScene
	if packed == null:
		push_warning("BossPresentationLoader: no scene for %s" % str(cfg.get("visual_key", "")))
		return null
	var root := Node3D.new()
	root.name = "MonsterVisualRoot"
	var monster := packed.instantiate() as Node3D
	root.add_child(monster)
	root.scale = Vector3.ONE * effective_scale(str(e.get("boss_template_id", "")), float(e.get("visual_scale", 1.0)))
	MounterScript.mount(monster, cfg)
	if apply_tint.is_valid():
		apply_tint.call(root, tint)
	return root
