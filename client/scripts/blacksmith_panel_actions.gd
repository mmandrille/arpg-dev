class_name BlacksmithPanelActions
extends RefCounted

const BlacksmithRecipesScript := preload("res://scripts/blacksmith_recipes.gd")
const BlacksmithShardInventoryScript := preload("res://scripts/blacksmith_shard_inventory.gd")
const BlacksmithUpgradePreviewScript := preload("res://scripts/blacksmith_upgrade_preview.gd")


static func action_enabled(ctx: Dictionary, item: Dictionary) -> bool:
	var recipe_id := str(ctx.get("selected_recipe_id", BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE))
	if recipe_id == BlacksmithRecipesScript.RECIPE_ITEM_RENEW:
		return is_upgrade_candidate(ctx, item) \
			and recipe_accepts_item(recipe_id, item) \
			and wallet_gold(ctx) >= next_cost(ctx, item) \
			and has_action_resource(ctx, item)
	return upgrade_enabled(ctx, item)


static func upgrade_enabled(ctx: Dictionary, item: Dictionary) -> bool:
	var level := item_level(item)
	var effective_max := effective_max_level(ctx)
	var recipe_id := str(ctx.get("selected_recipe_id", BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE))
	return is_upgrade_candidate(ctx, item) \
		and recipe_accepts_item(recipe_id, item) \
		and level < effective_max \
		and wallet_gold(ctx) >= next_cost(ctx, item) \
		and has_action_resource(ctx, item)


static func emit_renew(ctx: Dictionary, item: Dictionary) -> Dictionary:
	var recipe_id := str(ctx.get("selected_recipe_id", ""))
	var cost := next_cost(ctx, item)
	if not recipe_accepts_item(recipe_id, item):
		return {"ok": false, "message": BlacksmithRecipesScript.rejection_message(recipe_id)}
	if wallet_gold(ctx) < cost:
		return {"ok": false, "message": "Need %d gold" % cost}
	var resource_count := int(ctx.get("resource_count", 0))
	if not has_action_resource(ctx, item):
		var resource_id := BlacksmithRecipesScript.resource_item_def_id(recipe_id)
		return {"ok": false, "message": "Need %d %s" % [resource_count, BlacksmithShardInventoryScript.resource_display_name(resource_id)]}
	var item_instance_id := str(item.get("item_instance_id", ""))
	if item_instance_id == "":
		return {"ok": false, "message": "Stage a bag item to renew"}
	return {"ok": true, "item_instance_id": item_instance_id}


static func emit_upgrade(ctx: Dictionary, item: Dictionary) -> Dictionary:
	var recipe_id := str(ctx.get("selected_recipe_id", ""))
	var level := item_level(item)
	var cost := next_cost(ctx, item)
	if not recipe_accepts_item(recipe_id, item):
		return {"ok": false, "message": BlacksmithRecipesScript.rejection_message(recipe_id)}
	var effective_max := effective_max_level(ctx)
	if level >= effective_max:
		return {"ok": false, "message": "Item is already at max level"}
	if wallet_gold(ctx) < cost:
		return {"ok": false, "message": "Need %d gold" % cost}
	var resource_count := int(ctx.get("resource_count", 0))
	if not has_action_resource(ctx, item):
		var resource_id := BlacksmithRecipesScript.resource_item_def_id(recipe_id)
		return {"ok": false, "message": "Need %d %s" % [resource_count, BlacksmithShardInventoryScript.resource_display_name(resource_id)]}
	var item_instance_id := str(item.get("item_instance_id", ""))
	if item_instance_id == "":
		return {"ok": false, "message": "Stage a bag item to upgrade"}
	return {"ok": true, "item_instance_id": item_instance_id}


static func recipe_accepts_item(recipe_id: String, item: Dictionary) -> bool:
	return recipe_id == BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE \
		or recipe_id == BlacksmithRecipesScript.RECIPE_ITEM_RENEW


static func is_upgrade_candidate(ctx: Dictionary, item: Dictionary) -> bool:
	if item.is_empty():
		return false
	var resource_item_def_id := str(ctx.get("resource_item_def_id", ""))
	if resource_item_def_id != "" and str(item.get("item_def_id", "")) == resource_item_def_id:
		return false
	return str(item.get("item_template_id", "")) != "" \
		or str(item.get("slot", "")) != "" \
		or str(item.get("category", "")) == "equipment"


static func has_action_resource(ctx: Dictionary, item: Dictionary) -> bool:
	var resource_count := int(ctx.get("resource_count", 0))
	if resource_count <= 0:
		return true
	var recipe_id := str(ctx.get("selected_recipe_id", BlacksmithRecipesScript.RECIPE_ITEM_UPGRADE))
	var resource_id := BlacksmithRecipesScript.resource_item_def_id(recipe_id)
	var required_level := BlacksmithShardInventoryScript.required_resource_level(recipe_id, item)
	var staged_resource: Dictionary = ctx.get("staged_resource", {})
	if typeof(staged_resource) == TYPE_DICTIONARY and not staged_resource.is_empty():
		if str(staged_resource.get("item_def_id", "")) != resource_id:
			return false
		return BlacksmithUpgradePreviewScript.shard_level(staged_resource) >= required_level
	var inventory_items: Array = ctx.get("inventory_items", [])
	return BlacksmithShardInventoryScript.resource_inventory_count(inventory_items, resource_id, required_level) >= resource_count


static func max_item_level_for_deepest_depth(ctx: Dictionary) -> int:
	var depth: int = maxi(0, int(ctx.get("deepest_dungeon_depth", 0)))
	if depth < 1:
		return 1
	var levels_per_tier: int = maxi(1, int(ctx.get("item_level_levels_per_tier", 10)))
	return 1 + int((depth - 1) / levels_per_tier)


static func effective_max_level(ctx: Dictionary) -> int:
	var effective_max: int = int(ctx.get("max_level", 1))
	var depth_cap := max_item_level_for_deepest_depth(ctx)
	if depth_cap > 0:
		effective_max = mini(effective_max, depth_cap)
	return effective_max


static func wallet_gold(ctx: Dictionary) -> int:
	return int(ctx.get("gold", 0)) + int(ctx.get("stash_gold", 0))


static func item_level(item: Dictionary) -> int:
	return BlacksmithUpgradePreviewScript.item_level(item)


static func next_cost(ctx: Dictionary, item: Dictionary) -> int:
	return BlacksmithUpgradePreviewScript.next_cost(
		item,
		int(ctx.get("base_cost", 0)),
		int(ctx.get("growth_cost", 0)),
	)
