class_name BotIntentRejectAssertions
extends RefCounted


static func matches(step: Dictionary, state: Dictionary) -> bool:
	var want := str(step.get("reason", ""))
	var got := str(state.get("last_intent_reject_reason", ""))
	if want == "":
		return got != ""

	return got == want


static func assert_step(runner, step: Dictionary, state: Dictionary) -> bool:
	if matches(step, state):
		return true
	runner._fail("assert_intent_rejected failed: want_reason=%s got=%s step=%d scenario=%s" % [
		str(step.get("reason", "")),
		str(state.get("last_intent_reject_reason", "")),
		runner._step_index,
		str(runner.scenario.get("id", "?")),
	])

	return false
