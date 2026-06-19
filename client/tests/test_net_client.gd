# Headless tests for NetClient helper logic.
# Run via: godot --headless --path client --script res://tests/test_net_client.gd
extends SceneTree

const BotDebugProgressionSetupScript := preload("res://scripts/bot_debug_progression_setup.gd")
const NetClientScript := preload("res://scripts/net_client.gd")
const PerformanceStatusFormatterScript := preload("res://scripts/performance_status_formatter.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	_test_debug_progression_json_parsing()
	_test_debug_progression_body_normalizes_nested_maps()
	_test_message_latency_is_consumed_once()
	_test_performance_status_formatter_replaces_old_debug_text()
	if _fail == 0:
		print("[gdtest] PASS: test_net_client (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_net_client (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _test_debug_progression_json_parsing() -> void:
	var parsed := BotDebugProgressionSetupScript._progression_from_env('{"gold":125,"stats":{"vit":9},"skill_ranks":{"magic_bolt":3}}', "50")
	_assert_eq("json gold wins over fallback", int(parsed.get("gold", 0)), 125)
	var parsed_stats: Dictionary = parsed.get("stats", {})
	var parsed_skills: Dictionary = parsed.get("skill_ranks", {})
	_assert_eq("json stat value", int(parsed_stats.get("vit", 0)), 9)
	_assert_eq("json skill rank value", int(parsed_skills.get("magic_bolt", 0)), 3)
	var fallback := BotDebugProgressionSetupScript._progression_from_env("", "50")
	_assert_eq("gold fallback", int(fallback.get("gold", 0)), 50)


func _test_debug_progression_body_normalizes_nested_maps() -> void:
	var client = NetClientScript.new("http://localhost:18080")
	var body: Dictionary = client._debug_progression_body({
		"level": 12.0,
		"experience": 345.0,
		"unspent_stat_points": 4.0,
		"unspent_skill_points": 2.0,
		"stats": {"vit": 9.0, &"magic": "11"},
		"gold": 125.0,
		"skill_ranks": {"magic_bolt": 3.0, &"heal": "2"},
	})
	_assert_eq("body level", int(body.get("level", 0)), 12)
	_assert_eq("body experience", int(body.get("experience", 0)), 345)
	_assert_eq("body stat points", int(body.get("unspent_stat_points", 0)), 4)
	_assert_eq("body skill points", int(body.get("unspent_skill_points", 0)), 2)
	_assert_eq("body gold", int(body.get("gold", 0)), 125)
	var stats: Dictionary = body.get("stats", {})
	var skill_ranks: Dictionary = body.get("skill_ranks", {})
	_assert_eq("stats vit int", int(stats.get("vit", 0)), 9)
	_assert_eq("stats magic int", int(stats.get("magic", 0)), 11)
	_assert_eq("skill magic bolt int", int(skill_ranks.get("magic_bolt", 0)), 3)
	_assert_eq("skill heal int", int(skill_ranks.get("heal", 0)), 2)
	_assert_eq("non-dictionary map normalizes empty", client._int_value_map([]).size(), 0)


func _test_message_latency_is_consumed_once() -> void:
	var client = NetClientScript.new("http://localhost:18080")
	client._sent_message_msec["cmsg-1"] = Time.get_ticks_msec() - 37
	var latency_ms := client.consume_latency_ms("cmsg-1")
	_assert_true("latency is non-negative", latency_ms >= 0)
	_assert_eq("latency message consumed", client.consume_latency_ms("cmsg-1"), -1)
	_assert_eq("missing latency message", client.consume_latency_ms("missing"), -1)


func _test_performance_status_formatter_replaces_old_debug_text() -> void:
	var text: String = PerformanceStatusFormatterScript.format_status(61, 42, "open", 88, -3, {
		"total_ms": 18.4,
		"sim_ms": 11.2,
		"ai_ms": 2.1,
		"pathfind_ms": 4.7,
		"combat_ms": 3.3,
		"persist_ms": 1.6,
		"broadcast_ms": 0.8,
		"tick_budget_ms": 100.0,
		"tick_over_budget": true,
		"tick_overrun_ms": 12.5,
		"path_requests": 9,
		"path_cache_hits": 5,
		"path_nodes_visited": 142,
		"monsters_moved": 12,
		"live_monsters": 34,
		"monsters": 36,
		"walls": 9,
		"entities": 41,
		"inputs": 2,
		"results": 1,
		"changes": 24,
		"events": 29,
	})
	_assert_true("performance status title", text.contains("Performance Status"))
	_assert_true("performance status fps", text.contains("FPS 61"))
	_assert_true("performance status ping", text.contains("Ping 42ms"))
	_assert_true("performance status backend timing", text.contains("Backend total 18.4ms"))
	_assert_true("performance status path counters", text.contains("Path req 9"))
	_assert_true("performance status budget overrun", text.contains("over +12.5ms"))
	_assert_true("old controls removed", not text.contains("W/A/S/D"))
	_assert_true("old weapon visual removed", not text.contains("weapon_visual"))
	_assert_true("old inventory debug removed", not text.contains("inv="))


func _assert_true(label: String, condition: bool) -> void:
	if condition:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s" % label)


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])
