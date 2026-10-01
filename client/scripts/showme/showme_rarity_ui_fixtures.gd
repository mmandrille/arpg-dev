## Focused economy UI fixtures kept outside the grandfathered visual capture driver.
extends RefCounted

const ShopPanelScript := preload("res://scripts/shop_panel.gd")
const BlacksmithPanelScript := preload("res://scripts/blacksmith_panel.gd")


static func setup(tree: SceneTree, focus: String) -> void:
	if focus == "mystery-shop":
		var shop = ShopPanelScript.new()
		tree.root.add_child(shop)
		await tree.process_frame
		var offer := {"offer_id": "mystery:preview:ring", "kind": "mystery", "concealed": true, "mystery_label": "Unidentified ring", "slot": "ring", "category": "equipment", "buy_price": 75}
		shop.show_shop("preview_mystery", "town_mystery_seller", [offer], 100, [], {}, "Mystery Seller", [])
	elif focus == "blacksmith":
		var blacksmith = BlacksmithPanelScript.new()
		tree.root.add_child(blacksmith)
		await tree.process_frame
		var item := {"item_instance_id": "preview_rare_sword", "item_def_id": "long_sword", "display_name": "Long Sword", "rarity": "rare", "summary_lines": ["Damage 4-8"]}
		blacksmith.show_blacksmith("preview_blacksmith", [item], 550, 0, {})
		blacksmith.stage_inventory_item(item)
