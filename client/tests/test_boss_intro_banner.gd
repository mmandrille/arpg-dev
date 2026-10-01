extends SceneTree
## v512: intro banner plays once per boss at full health, re-arms on reset, stays silent otherwise.

const LoaderScript := preload("res://scripts/boss_presentation_loader.gd")
const BannerScript := preload("res://scripts/boss_intro_banner.gd")
const BossVisualsContextScript := preload("res://scripts/boss_visuals_context.gd")
const BossVisualsControllerScript := preload("res://scripts/boss_visuals_controller.gd")
const BossHealthBarScript := preload("res://scripts/boss_health_bar.gd")

var failures := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	LoaderScript.reset_for_test()
	var host := Control.new()
	get_root().add_child(host)
	var bar := BossHealthBarScript.new()
	host.add_child(bar)
	var context := BossVisualsContextScript.new()
	context.ui_host = host
	var boss := {"type": "monster", "is_boss": true, "hp": 100, "max_hp": 100, "boss_template_id": "crypt_matron"}
	context.entities = {"boss_1": boss}
	var controller := BossVisualsControllerScript.new(context, bar)
	controller.sync_at_tick(5)
	var state := controller.intro_banner_debug_state()
	_check(int(state["played_count"]) == 1 and bool(state["visible"]), "full-health boss plays the banner")
	_check(str(state["title"]) == "Crypt Matron" and str(state["template_id"]) == "crypt_matron", "banner shows the catalog name")
	var banner := controller.intro_banner() as Control
	_check(banner.mouse_filter == Control.MOUSE_FILTER_IGNORE and banner.get_parent() == host and host.get_child(0) == banner, "banner is mouse-transparent and below other UI")
	controller.sync_at_tick(6)
	_check(int(controller.intro_banner_debug_state()["played_count"]) == 1, "banner plays once per boss entity")
	controller.hide_boss_health_bar()
	controller.sync_at_tick(7)
	_check(int(controller.intro_banner_debug_state()["played_count"]) == 2, "reset re-arms the banner")

	# Fade timing comes from the catalog.
	var timing: Dictionary = LoaderScript.entry("crypt_matron")["banner"]
	await create_timer(float(timing["duration_s"]) + 0.4).timeout
	_check(not bool(controller.intro_banner_debug_state()["visible"]) and not banner.visible, "banner hides after the catalog duration")

	# Reconnect into a fight in progress stays silent.
	var hurt := {"type": "monster", "is_boss": true, "hp": 40, "max_hp": 100, "boss_template_id": "cave_warden"}
	context.entities = {"boss_2": hurt}
	controller.sync_at_tick(8)
	_check(int(controller.intro_banner_debug_state()["played_count"]) == 2, "damaged boss does not replay the banner")
	var phased := {"type": "monster", "is_boss": true, "hp": 100, "max_hp": 100, "boss_template_id": "cave_warden", "boss_phase": {"phase_kind": "active", "duration_ticks": 10, "remaining_ticks": 4}}
	context.entities = {"boss_3": phased}
	controller.sync_at_tick(9)
	_check(int(controller.intro_banner_debug_state()["played_count"]) == 2, "boss mid-attack does not replay the banner")
	var plain := {"type": "monster", "is_boss": true, "hp": 100, "max_hp": 100, "boss_template_id": "no_such_boss"}
	context.entities = {"boss_4": plain}
	controller.sync_at_tick(10)
	_check(int(controller.intro_banner_debug_state()["played_count"]) == 2, "uncatalogued boss has no banner")
	host.queue_free()
	print("[gdtest] PASS: test_boss_intro_banner (%d failed)" % failures)
	quit(1 if failures > 0 else 0)


func _check(ok: bool, label: String) -> void:
	if not ok:
		failures += 1
		push_error("FAIL: %s" % label)
