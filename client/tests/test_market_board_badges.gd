# Headless tests for market-board notification badges.
# Run via: godot --headless --path client --script res://tests/test_market_board_badges.gd
extends SceneTree

const MarketBoardBadgesScript := preload("res://scripts/market_board_badges.gd")
const TownNodeFactoryScript := preload("res://scripts/town_node_factory.gd")

var _pass: int = 0
var _fail: int = 0


func _initialize() -> void:
	call_deferred("_run")


func _run() -> void:
	var board := TownNodeFactoryScript.make_market_board_node()
	root.add_child(board)
	var state := MarketBoardBadgesScript.debug_state(board)
	_assert_true("board badge state exists", bool(state.get("exists", false)))
	_assert_eq("incoming starts at zero", int(state.get("incoming_bids", -1)), 0)
	_assert_eq("published starts at zero", int(state.get("published_listings", -1)), 0)
	_assert_false("incoming zero hidden", bool(state.get("incoming_visible", true)))
	_assert_false("published zero hidden", bool(state.get("published_visible", true)))

	MarketBoardBadgesScript.apply_to_board(board, 3, 1)
	state = MarketBoardBadgesScript.debug_state(board)
	_assert_eq("incoming active count", int(state.get("incoming_bids", 0)), 3)
	_assert_eq("published active count", int(state.get("published_listings", 0)), 1)
	_assert_eq("incoming active text", str(state.get("incoming_text", "")), "3")
	_assert_eq("published active text", str(state.get("published_text", "")), "1")
	_assert_true("incoming active visible", bool(state.get("incoming_visible", false)))
	_assert_true("published active visible", bool(state.get("published_visible", false)))
	_assert_eq("incoming active color", str(state.get("incoming_color", "")), "ffcf5a")
	_assert_eq("published active color", str(state.get("published_color", "")), "9fd7ff")

	MarketBoardBadgesScript.apply_to_board(board, 0, 0)
	state = MarketBoardBadgesScript.debug_state(board)
	_assert_eq("incoming reset count", int(state.get("incoming_bids", -1)), 0)
	_assert_eq("published reset count", int(state.get("published_listings", -1)), 0)
	_assert_false("incoming reset hidden", bool(state.get("incoming_visible", true)))
	_assert_false("published reset hidden", bool(state.get("published_visible", true)))

	board.queue_free()
	if _fail == 0:
		print("[gdtest] PASS: test_market_board_badges (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_market_board_badges (%d failures, %d assertions)" % [_fail, _pass])
		quit(1)


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass += 1
	else:
		_fail += 1
		push_error("assertion failed: " + label)


func _assert_false(label: String, value: bool) -> void:
	_assert_true(label, not value)


func _assert_eq(label: String, got, want) -> void:
	_assert_true("%s got=%s want=%s" % [label, str(got), str(want)], got == want)
