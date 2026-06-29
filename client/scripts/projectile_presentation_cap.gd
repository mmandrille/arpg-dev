class_name ProjectilePresentationCap
extends RefCounted

const MainConfigLoaderScript := preload("res://scripts/main_config_loader.gd")


static func apply(entities: Dictionary, hero_pos: Vector3) -> void:
	var cap := MainConfigLoaderScript.projectile_visible_cap()
	var projectiles: Array = []
	for id in entities.keys():
		var rec: Dictionary = entities[id]
		if str(rec.get("type", "")) != "projectile":
			continue
		var node := rec.get("node", null) as Node3D
		if node == null:
			continue
		var dist_sq := Vector2(node.global_position.x - hero_pos.x, node.global_position.z - hero_pos.z).length_squared()
		projectiles.append({"id": str(id), "node": node, "dist_sq": dist_sq})
	if projectiles.size() <= cap:
		for entry in projectiles:
			(entry["node"] as Node3D).visible = true
		return
	projectiles.sort_custom(func(a: Dictionary, b: Dictionary) -> bool:
		return float(a.get("dist_sq", 0.0)) < float(b.get("dist_sq", 0.0))
	)
	for i in range(projectiles.size()):
		var node := projectiles[i]["node"] as Node3D
		node.visible = i < cap
