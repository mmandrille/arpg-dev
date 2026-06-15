# Unit test for quest journal and elite objective state derivation.
extends SceneTree

const QuestEliteObjectiveStateScript := preload("res://scripts/quest_elite_objective_state.gd")

var _failures: int = 0


func _initialize() -> void:
	_test_quest_objectives()
	_test_elite_tracker_state()
	if _failures > 0:
		quit(1)
	print("[gdtest] PASS: test_quest_elite_objective_state")
	quit(0)


func _test_quest_objectives() -> void:
	_assert_eq("no reward", QuestEliteObjectiveStateScript.quest_journal_objectives({}).size(), 0)
	var active := QuestEliteObjectiveStateScript.quest_journal_objectives({
		"chest": {"quest_reward": true, "state": "closed"},
	})
	_assert_eq("active count", active.size(), 1)
	_assert_true("active incomplete", not bool((active[0] as Dictionary).get("complete", true)))
	var complete := QuestEliteObjectiveStateScript.quest_journal_objectives({
		"chest": {"quest_reward": true, "state": "open"},
	})
	_assert_true("complete reward", bool((complete[0] as Dictionary).get("complete", false)))


func _test_elite_tracker_state() -> void:
	var hidden := QuestEliteObjectiveStateScript.elite_tracker_state({})
	_assert_eq("hidden status", str(hidden.get("status", "")), "hidden")
	_assert_true("hidden invisible", not bool(hidden.get("visible", true)))
	var active := QuestEliteObjectiveStateScript.elite_tracker_state({
		"chest": {"elite_objective": true, "state": "closed"},
		"leader": {"monster_pack_leader": true, "hp": 3},
	})
	_assert_eq("active status", str(active.get("status", "")), "active")
	_assert_eq("active remaining", int(active.get("remaining_leaders", -1)), 1)
	var claim := QuestEliteObjectiveStateScript.elite_tracker_state({
		"chest": {"elite_objective": true, "state": "closed"},
		"leader": {"monster_pack_leader": true, "hp": 0},
	})
	_assert_eq("claim status", str(claim.get("status", "")), "claim")
	var complete := QuestEliteObjectiveStateScript.elite_tracker_state({
		"chest": {"elite_objective": true, "state": "open"},
		"leader": {"monster_pack_leader": true, "hp": 10},
	})
	_assert_eq("complete status", str(complete.get("status", "")), "complete")
	_assert_eq("complete remaining", int(complete.get("remaining_leaders", -1)), 0)


func _assert_true(label: String, value: bool) -> void:
	if not value:
		_failures += 1
		printerr("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got != want:
		_failures += 1
		printerr("[gdtest] FAIL %s got=%s want=%s" % [label, str(got), str(want)])
