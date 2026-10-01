extends SceneTree

const DelayScript := preload("res://scripts/bot_transport_delay.gd")

var _failed: int = 0


func _initialize() -> void:
	var delay = DelayScript.new()
	_check("local profile accepted", delay.configure("local"))
	_check("local disabled", not delay.enabled())
	_check("unknown profile rejected", not delay.configure("unbounded"))
	_check("unknown leaves prior profile", not delay.enabled())
	_check("bounded profile accepted", delay.configure("bounded_80_20"))
	_check("bounded enabled", delay.enabled())
	for i in range(10):
		delay.queue_outbound(i, 1000)
		_check("no early release %d" % i, delay.release_outbound(1059).is_empty())
	_check("first release at lower bound", delay.release_outbound(1060) == [0])
	_check("ordered release by upper bound", delay.release_outbound(1120) == [1, 2, 3, 4, 5, 6, 7, 8, 9])
	delay.queue_inbound("a", 2000)
	delay.queue_inbound("b", 2000)
	_check("inbound lower bound", delay.release_inbound(2059).is_empty())
	_check("inbound first", delay.release_inbound(2060) == ["a"])
	_check("inbound second", delay.release_inbound(2080) == ["b"])
	delay.queue_outbound("old", 3000)
	delay.clear()
	_check("clear drops queued envelopes", delay.release_outbound(4000).is_empty())
	delay.queue_outbound("new", 3000)
	_check("schedule resets deterministically", delay.release_outbound(3060) == ["new"])
	if _failed == 0:
		print("[gdtest] PASS: test_bot_transport_delay")
	quit(1 if _failed > 0 else 0)


func _check(label: String, passed: bool) -> void:
	if not passed:
		_failed += 1
		push_error("[gdtest] FAIL: %s" % label)
