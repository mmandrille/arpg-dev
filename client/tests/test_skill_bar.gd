# Unit test for the v44 skill hotbar slot.
# Run via: godot --headless --path client --script res://tests/test_skill_bar.gd
extends SceneTree

const SkillBarScript := preload("res://scripts/skill_bar.gd")

var _pass_count: int = 0
var _fail_count: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var bar := SkillBarScript.new()
	root.add_child(bar)
	await process_frame

	var emitted: Array = []
	bar.cast_skill_requested.connect(func(skill_id: String) -> void:
		emitted.append(skill_id)
	)
	var open_count := {"count": 0}
	bar.open_skills_requested.connect(func() -> void:
		open_count["count"] = int(open_count["count"]) + 1
	)

	bar.set_skill_id("magic_bolt")
	bar.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": _skill_rows(0),
	})
	bar.set_interactive(true)
	var state := bar.get_debug_state()
	_assert_eq("rank starts zero", int(state.get("rank", -1)), 0)
	_assert_false("skill upgrade badge hidden without points", bool(state.get("upgrade_badge_visible", true)))
	_assert_false("unranked skill disabled", bool(state.get("enabled", true)))
	_assert_eq("unranked slot text", str(state.get("slot_text", "")), "-")
	_assert_true("tooltip uses catalog name", str(state.get("tooltip_text", "")).contains("Magic Bolt"))
	bar._slot.pressed.emit()
	_assert_eq("slot click opens skills panel", int(open_count["count"]), 1)
	_assert_eq("slot click does not cast", emitted.size(), 0)

	bar.set_skill_progression({
		"unspent_skill_points": 1,
		"skills": _skill_rows(1),
	})
	state = bar.get_debug_state()
	_assert_true("skill upgrade badge visible with points", bool(state.get("upgrade_badge_visible", false)))
	_assert_eq("skill upgrade badge text", str(state.get("upgrade_badge_text", "")), "+")
	_assert_true("ranked skill enabled", bool(state.get("enabled", false)))
	_assert_eq("ranked slot uses icon control", str(state.get("slot_text", "")), "")
	_assert_eq("magic bolt icon shape", str(state.get("icon_shape", "")), "orb_projectile")
	bar.use_slot()
	_assert_eq("cast signal count", emitted.size(), 1)
	_assert_eq("cast signal skill", str(emitted[0]), "magic_bolt")
	bar.set_skill_id("heal")
	bar.set_skill_progression({
		"unspent_skill_points": 0,
		"skills": _skill_rows(1, 0, 1),
	})
	state = bar.get_debug_state()
	_assert_eq("selected heal skill id", str(state.get("skill_id", "")), "heal")
	_assert_eq("selected heal slot uses icon control", str(state.get("slot_text", "")), "")
	_assert_eq("selected heal icon shape", str(state.get("icon_shape", "")), "heart")
	_assert_true("selected heal tooltip", str(state.get("tooltip_text", "")).contains("Heal"))
	bar.use_slot()
	_assert_eq("heal cast signal count", emitted.size(), 2)
	_assert_eq("heal cast signal skill", str(emitted[1]), "heal")
	bar.set_skill_id("magic_bolt")

	bar.set_skill_cooldowns([{"skill_id": "magic_bolt", "remaining_ticks": 40, "total_ticks": 40}])
	state = bar.get_debug_state()
	_assert_false("cooldown disables slot", bool(state.get("enabled", true)))
	_assert_eq("cooldown remaining", int(state.get("remaining_ticks", -1)), 40)
	_assert_eq("cooldown total", int(state.get("total_ticks", -1)), 40)
	_assert_true("cooldown fraction full", float(state.get("cooldown_fraction", 0.0)) > 0.99)

	bar._process(0.10)
	state = bar.get_debug_state()
	_assert_true("cooldown locally decays", int(state.get("remaining_ticks", 40)) < 40)

	bar.set_skill_cooldowns([])
	state = bar.get_debug_state()
	_assert_true("slot re-enables without cooldown", bool(state.get("enabled", false)))

	bar.queue_free()
	print("[gdtest] PASS: test_skill_bar (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit(1 if _fail_count > 0 else 0)


func _assert_eq(label: String, got, expected) -> void:
	if got == expected:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s: expected=%s got=%s" % [label, str(expected), str(got)])


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
	else:
		_fail_count += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _skill_rows(magic_rank: int, rage_rank: int = 0, heal_rank: int = 0, cleave_rank: int = 0, ice_shard_rank: int = 0, holy_shield_rank: int = 0, ligthing_rank: int = 0) -> Array:
	return [
		{"skill_id": "cleave", "rank": cleave_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "magic_bolt", "rank": magic_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "ice_shard", "rank": ice_shard_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "ligthing", "rank": ligthing_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "rage", "rank": rage_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "heal", "rank": heal_rank, "max_rank": 5, "can_spend": false},
		{"skill_id": "holy_shield", "rank": holy_shield_rank, "max_rank": 5, "can_spend": false},
	]
