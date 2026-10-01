extends SceneTree
## v512: per-boss presentation catalog, model/scale override, attachments and headgear.

const LoaderScript := preload("res://scripts/boss_presentation_loader.gd")
const MounterScript := preload("res://scripts/boss_presentation_mounter.gd")

var failures := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	LoaderScript.reset_for_test()
	_check(LoaderScript.has("cave_warden") and LoaderScript.has("crypt_matron"), "catalog covers both boss templates")
	_check(not LoaderScript.has("") and not LoaderScript.has("unknown_boss"), "unknown template has no presentation")
	var warden_scale := LoaderScript.effective_scale("cave_warden", 2.0)
	var matron_scale := LoaderScript.effective_scale("crypt_matron", 2.2)
	_check(maxf(warden_scale, matron_scale) / minf(warden_scale, matron_scale) >= 1.15, "boss scales are visibly distinct")
	_check(is_equal_approx(LoaderScript.effective_scale("unknown_boss", 1.7), 1.7), "unknown template keeps the wire scale")
	_check(is_equal_approx(LoaderScript.effective_scale("", -3.0), 1.0), "non-positive wire scale falls back to 1")
	var fallback := Color("#123456")
	_check(LoaderScript.base_tint("unknown_boss", fallback) == fallback, "unknown template keeps the wire tint")
	_check(LoaderScript.base_tint("cave_warden", fallback) != fallback, "catalog tint overrides the wire tint")
	_check(LoaderScript.is_untinted("BossArenaPresence") and LoaderScript.is_untinted("BossHeadgear") and not LoaderScript.is_untinted("MonsterVisualRoot"), "tint skip list")

	# Configurable: a temp catalog proves behavior follows data, not constants.
	LoaderScript.set_bosses_for_test({"temp_boss": {"scale": 9.0, "tint": "#ff0000"}})
	_check(is_equal_approx(LoaderScript.effective_scale("temp_boss", 1.0), 9.0), "scale follows catalog data")
	LoaderScript.reset_for_test()

	for template_id in ["cave_warden", "crypt_matron"]:
		var cfg := LoaderScript.entry(template_id)
		var tinted := []
		var root := LoaderScript.make_root({"boss_template_id": template_id, "visual_scale": 2.0, "visual_model": "monster_tiny_flyer"}, func(r, c): tinted.append(c), Color.WHITE)
		_check(root != null, "%s builds a kit root regardless of wire visual_model" % template_id)
		if root == null:
			continue
		_check(is_equal_approx(root.scale.x, float(cfg["scale"])), "%s root uses catalog scale" % template_id)
		_check(tinted.size() == 1, "%s root is tinted once through the caller" % template_id)
		get_root().add_child(root)
		await process_frame
		var mounts := root.find_children("BossAttach_*", "BoneAttachment3D", true, false)
		_check(mounts.size() == (cfg["attachments"] as Array).size(), "%s mounts every catalog attachment" % template_id)
		var wants_headgear := str((cfg["headgear"] as Dictionary).get("kind", "none")) != "none"
		_check((root.find_child("BossHeadgear", true, false) != null) == wants_headgear, "%s headgear follows the catalog kind" % template_id)
		var ap := root.find_child("AnimationPlayer", true, false) as AnimationPlayer
		_check(ap != null and ap.has_animation("idle") and ap.has_animation("attack") and ap.has_animation("death"), "%s keeps aliased clips" % template_id)
		root.queue_free()
		await process_frame
	_check(LoaderScript.make_root({"boss_template_id": "unknown_boss"}, Callable(), Color.WHITE) == null, "unknown boss falls back to the wire path")

	var horns := MounterScript.build_primitive("horns", Color.WHITE, 1.0)
	var crown := MounterScript.build_primitive("crown", Color.WHITE, 2.0)
	_check(horns.get_child_count() == 2 and crown.get_child_count() == 6 and crown.scale.x == 2.0, "primitive headgear shapes follow data")
	horns.free()
	crown.free()
	print("[gdtest] PASS: test_boss_presentation_loader (%d failed)" % failures)
	quit(1 if failures > 0 else 0)


func _check(ok: bool, label: String) -> void:
	if not ok:
		failures += 1
		push_error("FAIL: %s" % label)
