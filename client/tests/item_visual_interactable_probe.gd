extends RefCounted

# Interactable presentation probes for test_item_visuals.gd.


const MainScript := preload("res://scripts/main.gd")


func verify_chest_models(tree: SceneTree, fail: Callable) -> bool:
	var main = MainScript.new()
	var stash := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_stash"})
	var chest := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest"})
	var objective := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest", "elite_objective": true})
	var quest := main._make_entity_node({"type": "interactable", "interactable_def_id": "treasure_chest", "quest_reward": true})
	if stash == null or stash.name != "TownStashChest" or stash.find_child("ChestStashCrest", true, false) == null:
		fail.call("town stash did not use fortified chest model")
		main.free()
		return false
	if chest == null or chest.name != "TreasureChest" or chest.find_child("ChestLockPlate", true, false) == null:
		fail.call("treasure chest did not use chest model")
		stash.free()
		main.free()
		return false
	var glow := chest.find_child("ChestInnerGlow", true, false) as MeshInstance3D
	var lid := chest.find_child("ChestLidPivot", true, false) as Node3D
	if glow == null or lid == null or glow.visible:
		fail.call("treasure chest missing closed lid/glow state")
		stash.free()
		chest.free()
		main.free()
		return false
	if objective == null or objective.find_child("EliteObjectiveMarker", true, false) == null:
		fail.call("objective treasure chest did not expose marker")
		stash.free()
		chest.free()
		quest.free()
		main.free()
		return false
	if quest == null or quest.find_child("QuestRewardMarker", true, false) == null:
		fail.call("quest reward treasure chest did not expose marker")
		stash.free()
		chest.free()
		objective.free()
		main.free()
		return false
	tree.get_root().add_child(chest)
	main._apply_interactable_state_tint({"node": chest, "interactable_def_id": "treasure_chest", "elite_objective": false, "quest_reward": false}, "open")
	if not glow.visible:
		fail.call("opened treasure chest did not reveal inner glow")
		stash.free()
		main.free()
		return false
	var objective_marker := objective.find_child("EliteObjectiveMarker", true, false) as MeshInstance3D
	tree.get_root().add_child(objective)
	main._apply_interactable_state_tint({"node": objective, "interactable_def_id": "treasure_chest", "elite_objective": true, "quest_reward": false, "state": "closed"}, "open")
	if objective_marker == null or not objective_marker.visible:
		fail.call("opened objective treasure chest did not keep marker visible")
		stash.free()
		main.free()
		return false
	var quest_marker := quest.find_child("QuestRewardMarker", true, false) as MeshInstance3D
	tree.get_root().add_child(quest)
	main._apply_interactable_state_tint({"node": quest, "interactable_def_id": "treasure_chest", "elite_objective": false, "quest_reward": true, "state": "closed"}, "open")
	if quest_marker == null or not quest_marker.visible:
		fail.call("opened quest reward treasure chest did not keep marker visible")
		stash.free()
		main.free()
		return false
	stash.free()
	main.free()

	return true


func verify_vendor_models(fail: Callable) -> bool:
	var main = MainScript.new()
	var vendor := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_vendor"})
	var mystery := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_mystery_seller"})
	if vendor == null or vendor.name != "ShopVendor" or vendor.find_child("VendorSign", true, false) == null:
		fail.call("town vendor did not use merchant model")
		main.free()
		return false
	if mystery == null or mystery.name != "MysterySeller" or mystery.find_child("CrystalOrb", true, false) == null:
		fail.call("mystery seller did not use dark-violet merchant model")
		vendor.free()
		main.free()
		return false
	var vendor_body := vendor.find_child("Body", true, false) as MeshInstance3D
	var mystery_body := mystery.find_child("Body", true, false) as MeshInstance3D
	if vendor_body == null or (vendor_body.material_override as StandardMaterial3D).albedo_color.to_html(false) != "e2b92e":
		fail.call("town vendor body is not yellow")
		vendor.free()
		mystery.free()
		main.free()
		return false
	if mystery_body == null or (mystery_body.material_override as StandardMaterial3D).albedo_color.to_html(false) != "2b124a":
		fail.call("mystery seller body is not dark violet")
		vendor.free()
		mystery.free()
		main.free()
		return false
	vendor.free()
	mystery.free()
	main.free()

	return true


func verify_market_board_model(fail: Callable) -> bool:
	var main = MainScript.new()
	var board := main._make_entity_node({"type": "interactable", "interactable_def_id": "town_market_board"})
	if board == null or board.name != "MarketBoard":
		fail.call("market board did not use board model")
		main.free()
		return false
	if board.find_child("IncomingBidCount", true, false) == null:
		fail.call("market board missing incoming bid counter")
		board.free()
		main.free()
		return false
	if board.find_child("PublishedListingCount", true, false) == null:
		fail.call("market board missing published listing counter")
		board.free()
		main.free()
		return false
	board.free()
	main.free()

	return true


func verify_stair_models(fail: Callable) -> bool:
	var main = MainScript.new()
	var up := main._make_entity_node({"type": "interactable", "interactable_def_id": "stairs_up"})
	var down := main._make_entity_node({"type": "interactable", "interactable_def_id": "stairs_down"})
	if up == null or up.find_child("UpHighLanding", true, false) == null or up.find_child("UpBackWall", true, false) == null:
		fail.call("stairs_up did not use raised stair model")
		main.free()
		return false
	if down == null or down.find_child("DownPitOpening", true, false) == null or down.find_child("DownBackWall", true, false) == null:
		fail.call("stairs_down did not use descending pit model")
		up.free()
		main.free()
		return false
	var first_down_step := down.find_child("DownStep0", true, false) as Node3D
	var last_down_step := down.find_child("DownStep4", true, false) as Node3D
	if first_down_step == null or last_down_step == null or first_down_step.position.y <= last_down_step.position.y:
		fail.call("stairs_down steps do not descend into the opening")
		up.free()
		down.free()
		main.free()
		return false
	var first_up_step := up.find_child("UpStep0", true, false) as Node3D
	var last_up_step := up.find_child("UpStep4", true, false) as Node3D
	if first_up_step == null or last_up_step == null or first_up_step.position.y >= last_up_step.position.y:
		fail.call("stairs_up steps do not rise to the landing")
		up.free()
		down.free()
		main.free()
		return false
	up.free()
	down.free()
	main.free()

	return true
