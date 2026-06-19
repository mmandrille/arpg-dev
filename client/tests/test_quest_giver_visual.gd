extends SceneTree

const MainScript := preload("res://scripts/main.gd")

func _initialize() -> void:
	var main = MainScript.new()
	var quest_giver := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_quest_giver"})
	if quest_giver == null or quest_giver.name != "QuestSteward":
		_fail("town quest giver did not use quest steward model")
		main.free()
		return
	if quest_giver.find_child("QuestScroll", true, false) == null:
		_fail("quest steward missing scroll")
		quest_giver.free()
		main.free()
		return
	if quest_giver.find_child("QuestMarker", true, false) == null:
		_fail("quest steward missing quest marker")
		quest_giver.free()
		main.free()
		return
	quest_giver.free()
	main.free()
	print("[gdtest] PASS: quest giver visual")
	quit(0)

func _fail(message: String) -> void:
	push_error(message)
	quit(1)
