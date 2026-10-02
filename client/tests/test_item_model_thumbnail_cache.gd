extends SceneTree

const ThumbnailCache := preload("res://scripts/item_model_thumbnail_cache.gd")
const ItemVisuals := preload("res://scripts/item_visuals_loader.gd")
const EquipmentDisplay := preload("res://scripts/equipment_display_loader.gd")

var _passed := 0
var _failed := 0


func _initialize() -> void:
	ItemVisuals.ensure_loaded()
	ThumbnailCache.reset_for_tests()
	_assert("hand model uses item visual asset mapping", ThumbnailCache.asset_id_for_item({"item_def_id": "long_sword"}) == ItemVisuals.hand_asset_id("long_sword"))
	_assert("items sharing a model resolve to one reusable cache key", ThumbnailCache.asset_id_for_item({"item_def_id": "long_sword"}) == ThumbnailCache.asset_id_for_item({"item_def_id": "rusty_sword"}))
	_assert("armor model resolves through the shared item presentation", ThumbnailCache.asset_id_for_item({"item_def_id": "helm"}) == "fallback_equipment_head_v0")
	_assert("unknown items retain the 2D fallback", ThumbnailCache.asset_id_for_item({"item_def_id": "missing_item"}) == "")
	_assert("non-equipment items retain the 2D fallback", ThumbnailCache.asset_id_for_item({"item_def_id": "red_potion"}) == "")
	_assert("equipped and ground tint use stronger shared setting", EquipmentDisplay.rig_native_tint_strength() > 0.25)
	_assert("resolving thumbnails creates no rendering surfaces", int(ThumbnailCache.debug_cache_state().get("live_viewports", -1)) == 0)
	_test_cache_is_bounded()
	print("[gdtest] %s: test_item_model_thumbnail_cache (%d passed, %d failed)" % ["PASS" if _failed == 0 else "FAIL", _passed, _failed])
	quit(1 if _failed else 0)


func _test_cache_is_bounded() -> void:
	for index in range(ThumbnailCache.MAX_CACHED_THUMBNAILS + 1):
		var image := Image.create(1, 1, false, Image.FORMAT_RGBA8)
		image.fill(Color.WHITE)
		ThumbnailCache._cache_texture("test_asset_%02d" % index, ImageTexture.create_from_image(image))
	var state := ThumbnailCache.debug_cache_state()
	var cached: Array = state.get("cached_asset_ids", [])
	_assert("thumbnail texture cache has a fixed maximum", cached.size() == ThumbnailCache.MAX_CACHED_THUMBNAILS)
	_assert("oldest cached model is evicted first", not cached.has("test_asset_00") and cached.has("test_asset_32"))


func _assert(label: String, condition: bool) -> void:
	if condition:
		_passed += 1
	else:
		_failed += 1
		push_error("[gdtest] FAIL: " + label)
