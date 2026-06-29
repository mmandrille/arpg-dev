# Headless unit tests for ProjectilePresentationCap (v375).
extends SceneTree

const ProjectilePresentationCapScript := preload("res://scripts/projectile_presentation_cap.gd")
const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")

var _pass := 0
var _fail := 0


func _init() -> void:
	MainConfigLoaderScript.ensure_loaded()
	_test_all_visible_when_under_cap()
	_test_far_projectiles_hidden_when_over_cap()
	_finish()


func _test_all_visible_when_under_cap() -> void:
	var entities := {
		"p1": {"type": "projectile", "node": _node_at(0.0, 0.0)},
		"p2": {"type": "projectile", "node": _node_at(5.0, 0.0)},
	}
	ProjectilePresentationCapScript.apply(entities, Vector3.ZERO)
	_assert_true("near projectile visible", (entities["p1"]["node"] as Node3D).visible)
	_assert_true("mid projectile visible", (entities["p2"]["node"] as Node3D).visible)


func _test_far_projectiles_hidden_when_over_cap() -> void:
	var cap := MainConfigLoaderScript.projectile_visible_cap()
	var entities := {}
	for i in range(cap + 2):
		entities["p%d" % i] = {
			"type": "projectile",
			"node": _node_at(float(i), 0.0),
		}
	ProjectilePresentationCapScript.apply(entities, Vector3.ZERO)
	var visible := 0
	for key in entities.keys():
		if (entities[key]["node"] as Node3D).visible:
			visible += 1
	_assert_eq("only cap nearest projectiles visible", visible, cap)


func _node_at(x: float, z: float) -> Node3D:
	var n := Node3D.new()
	n.position = Vector3(x, 0.0, z)
	n.visible = true
	return n


func _assert_eq(label: String, got, want) -> void:
	if got == want:
		_pass += 1
	else:
		_fail += 1
		push_error("[gdtest] FAIL %s: got=%s want=%s" % [label, str(got), str(want)])


func _assert_true(label: String, value: bool) -> void:
	_assert_eq(label, value, true)


func _finish() -> void:
	if _fail == 0:
		print("[gdtest] PASS: test_projectile_presentation_cap (%d assertions)" % _pass)
		quit(0)
	else:
		print("[gdtest] FAIL: test_projectile_presentation_cap (%d failures)" % _fail)
		quit(1)
