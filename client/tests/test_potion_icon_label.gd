extends SceneTree

const PotionIconLabelScript := preload("res://scripts/potion_icon_label.gd")


func _init() -> void:
	call_deferred("_run")


func _run() -> void:
	if not PotionIconLabelScript.is_leveled_potion("red_potion"):
		push_error("red_potion should be a leveled consumable")
		quit(1)
		return
	if PotionIconLabelScript.is_leveled_potion("long_sword"):
		push_error("long_sword should not be a leveled consumable")
		quit(1)
		return
	var label := PotionIconLabelScript.icon_label({"item_def_id": "red_potion", "item_level": 5}, "HP")
	if label != "5":
		push_error("expected level label 5, got %s" % label)
		quit(1)
		return
	var rolled_label := PotionIconLabelScript.icon_label(
		{"item_def_id": "red_potion", "rolled_stats": {"item_level": 3}},
		"HP",
	)
	if rolled_label != "3":
		push_error("expected rolled_stats level label 3, got %s" % rolled_label)
		quit(1)
		return
	print("[gdtest] PASS: potion_icon_label")
	quit(0)
