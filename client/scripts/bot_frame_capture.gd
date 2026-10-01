## Saves a real client-bot viewport frame under this checkout's ignored artifacts.
class_name BotFrameCapture
extends RefCounted

class CaptureJob:
	extends RefCounted
	var done := false
	var error := ""


static func valid_name(name: String) -> bool:
	if name.is_empty() or name.length() > 64:
		return false
	for i in name.length():
		var code := name.unicode_at(i)
		if not ((code >= 48 and code <= 57) or (code >= 65 and code <= 90) \
				or (code >= 97 and code <= 122) or code == 45 or code == 95):
			return false
	return true


static func output_path(name: String) -> String:
	if not valid_name(name):
		return ""
	var client_dir := ProjectSettings.globalize_path("res://").trim_suffix("/")
	return client_dir.get_base_dir().path_join(".artifacts/bot-captures").path_join(name + ".png")


static func start(viewport: Viewport, name: String, quality: String, fixture: Dictionary = {}) -> CaptureJob:
	var job := CaptureJob.new()
	_capture(viewport, name, quality, fixture, job)
	return job


static func _capture(viewport: Viewport, name: String, quality: String, fixture: Dictionary, job: CaptureJob) -> void:
	var path := output_path(name)
	if viewport == null or path.is_empty():
		job.error = "invalid viewport or capture name"
		job.done = true
		return
	if DisplayServer.get_name() == "headless":
		job.error = "capture_frame requires a windowed renderer (HEADLESS=0)"
		job.done = true
		return
	var dir_error := DirAccess.make_dir_recursive_absolute(path.get_base_dir())
	if dir_error != OK:
		job.error = "could not create capture directory: %d" % dir_error
		job.done = true
		return
	await RenderingServer.frame_post_draw
	var save_error := viewport.get_texture().get_image().save_png(path)
	if save_error != OK:
		job.error = "could not save frame: %d" % save_error
		job.done = true
		return
	var manifest := {
		"png": path,
		"renderer": RenderingServer.get_current_rendering_method(),
		"driver": RenderingServer.get_current_rendering_driver_name(),
		"quality": quality,
		"viewport_size": {"width": viewport.size.x, "height": viewport.size.y},
		"captured_at_local": Time.get_datetime_string_from_system(),
		"fixture": fixture,
	}
	var manifest_file := FileAccess.open(path.trim_suffix(".png") + ".json", FileAccess.WRITE)
	if manifest_file == null:
		job.error = "could not write capture manifest: %d" % FileAccess.get_open_error()
		job.done = true
		return
	manifest_file.store_string(JSON.stringify(manifest, "\t") + "\n")
	manifest_file.close()
	print("[bot-capture] saved=%s renderer=%s driver=%s quality=%s" % [
		path, manifest["renderer"], manifest["driver"], quality,
	])
	job.done = true
