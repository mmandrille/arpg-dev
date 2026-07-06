extends RefCounted

# Sync scale probe helpers for test_item_visuals.gd. Frame advancement must happen
# on the SceneTree test runner (await process_frame); awaiting inside RefCounted
# coroutines from _initialize can deadlock headless --script runs.


func prepare(
	tree: SceneTree,
	main_script: Script,
	character_scene: PackedScene,
	resolver_script: Script,
	fail: Callable,
) -> Dictionary:
	var main = main_script.new()
	var character = character_scene.instantiate() as Node3D
	tree.get_root().add_child(character)
	main.character_visual = character
	main.inventory = [
		{"item_instance_id": "5001", "item_def_id": "starter_paladin_sword", "slot": "main_hand", "equipped": true, "rarity": "common"},
		{"item_instance_id": "5002", "item_def_id": "starter_paladin_shield", "slot": "off_hand", "equipped": true, "rarity": "common"},
	]
	main.equipped = {"main_hand": "5001", "off_hand": "5002"}
	main.resolver = resolver_script.new(character)
	main.character_progression = {"character_class": "paladin"}
	main.call("_apply_local_player_class_model")
	return {"main": main}


func verify_transforms(ctx: Dictionary, fail: Callable) -> bool:
	if ctx.is_empty():
		return false
	var main = ctx.get("main")
	var character: Node3D = main.character_visual as Node3D
	if character == null:
		fail.call("paladin character visual missing after class model apply")
		main.queue_free()
		return false
	var sword := character.find_child("weapon_long_sword_v0", true, false) as Node3D
	var shield := character.find_child("equipment_shield_kite_v0", true, false) as Node3D
	if sword == null or shield == null:
		fail.call("paladin mounted equipment missing after class model apply: sword=%s shield=%s" % [str(sword), str(shield)])
		character.queue_free()
		main.queue_free()
		return false
	var sword_scale := sword.global_transform.basis.get_scale()
	var shield_scale := shield.global_transform.basis.get_scale()
	if sword.global_position.y < 0.2:
		fail.call("paladin sword mounted below hand/floor after class scale compensation: %s local=%s" % [str(sword.global_position), str(sword.position)])
		character.queue_free()
		main.queue_free()
		return false
	if sword_scale.x > 1.2 or sword_scale.y > 1.2 or sword_scale.z > 1.2:
		fail.call("paladin sword inherited model scale: %s" % str(sword_scale))
		character.queue_free()
		main.queue_free()
		return false
	if shield_scale.x > 1.2 or shield_scale.y > 1.2 or shield_scale.z > 1.2:
		fail.call("paladin shield inherited model scale: %s" % str(shield_scale))
		character.queue_free()
		main.queue_free()
		return false
	character.queue_free()
	main.queue_free()
	return true
