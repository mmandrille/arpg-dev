extends SceneTree
## v512: per-template ground aura, telegraph precedence, quality gating, cleanup.

const Presence := preload("res://scripts/boss_arena_presence.gd")
const LoaderScript := preload("res://scripts/boss_presentation_loader.gd")
const CombatVfxScript := preload("res://scripts/combat_vfx.gd")

var failures := 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	LoaderScript.reset_for_test()
	var results := {}
	for template_id in ["cave_warden", "crypt_matron"]:
		CombatVfxScript.set_quality("balanced")
		var node := Node3D.new()
		get_root().add_child(node)
		var rec := {"node": node, "type": "monster", "is_boss": true, "hp": 10, "boss_template_id": template_id, "visual_scale": 2.0}
		Presence.sync_for_record(rec)
		var aura: Dictionary = LoaderScript.entry(template_id)["aura"]
		var marker := node.find_child("BossArenaPresence", false, false) as MeshInstance3D
		_check(marker != null and bool(rec["has_boss_arena_presence"]), "%s has ground presence" % template_id)
		_check(is_equal_approx(float(rec["boss_aura_radius"]), float(aura["radius"])), "%s aura radius from catalog" % template_id)
		_check(Color(str(aura["color"])).to_html(false) == Color(str(rec["boss_arena_color"])).to_html(false), "%s aura color from catalog" % template_id)
		_check(marker.find_child("BossAuraDisc", false, false) != null, "%s has inner disc" % template_id)
		_check(marker.find_child("BossAuraLight", false, false) != null, "%s has soft light on Balanced" % template_id)
		_check(float(marker.get_meta("boss_aura_pulse_hz", 0.0)) > 0.0, "%s pulses on Balanced" % template_id)
		results[template_id] = str(rec["boss_arena_color"])

		var mesh_before := marker.mesh
		Presence.sync_for_record(rec)
		_check(marker.mesh == mesh_before, "unchanged state does not rebuild meshes every tick")

		rec["boss_telegraph_active"] = true
		rec["telegraph_tint"] = "ff0000"
		Presence.sync_for_record(rec)
		_check(str(rec["boss_arena_color"]).begins_with("ff0000"), "%s telegraph color still wins" % template_id)
		rec["boss_telegraph_active"] = false

		CombatVfxScript.set_quality("performance")
		Presence.sync_for_record(rec)
		await process_frame
		_check(marker.find_child("BossAuraLight", false, false) == null and float(marker.get_meta("boss_aura_pulse_hz", 0.0)) == 0.0, "%s Performance tier drops light and pulse" % template_id)
		_check(marker.mesh != null, "%s Performance tier keeps the ring" % template_id)

		rec["hp"] = 0
		Presence.sync_for_record(rec)
		await process_frame
		_check(node.find_child("BossArenaPresence", false, false) == null and not bool(rec["has_boss_arena_presence"]), "%s aura removed on death" % template_id)
		node.queue_free()
	_check(results["cave_warden"] != results["crypt_matron"], "bosses have distinct aura colors")
	CombatVfxScript.set_quality("")
	print("[gdtest] PASS: test_boss_arena_presence (%d failed)" % failures)
	quit(1 if failures > 0 else 0)


func _check(ok: bool, label: String) -> void:
	if not ok:
		failures += 1
		push_error("FAIL: %s" % label)
