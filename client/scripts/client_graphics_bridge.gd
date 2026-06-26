extends RefCounted
class_name ClientGraphicsBridge

const ClientSettingsScript := preload("res://scripts/client_settings.gd")
const FogPresentationLoaderScript := preload("res://scripts/fog_presentation_loader.gd")


static func sync_fog_performance_throttle(fog_overlay: FogOfWarOverlay, settings: ClientSettings) -> void:
	if fog_overlay == null or settings == null:
		return
	var enabled := settings.graphics_quality == ClientSettingsScript.GRAPHICS_QUALITY_PERFORMANCE
	fog_overlay.set_performance_throttle(enabled)


static func apply_graphics_quality_selected(
	fog_overlay: FogOfWarOverlay,
	settings: ClientSettings,
	quality: String,
	sync_settings_panel: Callable = Callable(),
) -> void:
	if settings == null:
		return
	settings.set_graphics_quality(quality, true, true)
	sync_fog_performance_throttle(fog_overlay, settings)
	if sync_settings_panel.is_valid():
		sync_settings_panel.call()
