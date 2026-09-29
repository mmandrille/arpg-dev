## Builds (and caches) the AnimationLibrary for a KayKit hero clip profile (ADR-0018 P3a).
## Adventurers 2.0 heroes embed no clips; Character Animations 1.1 ships them in separate
## Rig_Medium_*.glb files whose tracks target `Rig_Medium/Skeleton3D:<bone>` — the same path the
## hero GLBs import with — so the clips bind without retargeting. Each logical clip is a copy of
## the kit clip so loop flags never touch the source library.
class_name KitHeroClips
extends RefCounted

const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const CATALOG_PATH := "../shared/assets/kit_hero_presentation.v0.json"

static var _catalog: Dictionary = {}
static var _catalog_loaded := false
static var _libraries: Dictionary = {}


static func profile(profile_id: String) -> Dictionary:
	_ensure_catalog()
	return ((_catalog.get("clip_profiles", {}) as Dictionary).get(profile_id, {}) as Dictionary).duplicate(true)


static func library(profile_id: String) -> AnimationLibrary:
	if _libraries.has(profile_id):
		return _libraries[profile_id]
	var cfg := profile(profile_id)
	var built: AnimationLibrary = null
	if not cfg.is_empty():
		built = _build(cfg, profile_id)
	_libraries[profile_id] = built
	return built


static func _build(cfg: Dictionary, profile_id: String) -> AnimationLibrary:
	var sources: Array = []
	var holders: Array = []
	for asset_id in cfg.get("libraries", []):
		var holder := LibraryScript.instantiate(str(asset_id))
		if holder == null:
			continue
		holders.append(holder)
		var player := holder.find_child("AnimationPlayer", true, false) as AnimationPlayer
		if player != null and player.has_animation_library(""):
			sources.append(player.get_animation_library(""))
	var out := AnimationLibrary.new()
	var loops: Array = cfg.get("loop", [])
	var clips: Dictionary = cfg.get("clips", {})
	for logical in clips:
		var kit_clip := str(clips[logical])
		var found: Animation = null
		for source in sources:
			if (source as AnimationLibrary).has_animation(kit_clip):
				found = (source as AnimationLibrary).get_animation(kit_clip)
				break
		if found == null:
			push_warning("KitHeroClips: profile %s has no kit clip %s for %s" % [profile_id, kit_clip, logical])
			continue
		var anim := found.duplicate(true) as Animation
		anim.loop_mode = Animation.LOOP_LINEAR if loops.has(logical) else Animation.LOOP_NONE
		out.add_animation(StringName(logical), anim)
	for holder in holders:
		(holder as Node).free()
	return out


static func _ensure_catalog() -> void:
	if _catalog_loaded:
		return
	_catalog_loaded = true
	var path := ProjectSettings.globalize_path("res://").path_join(CATALOG_PATH)
	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		push_warning("KitHeroClips: cannot open %s" % path)
		return
	var parsed = JSON.parse_string(file.get_as_text())
	if typeof(parsed) == TYPE_DICTIONARY:
		_catalog = parsed as Dictionary
