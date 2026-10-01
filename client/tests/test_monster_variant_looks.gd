extends SceneTree
## v511: monster variant looks (rarity x depth x family). Expectations are derived from the
## loaded catalog, never from hardcoded tuning values (Test Locking Policy).

const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const ResolverScript := preload("res://scripts/monster_variant_resolver.gd")
const LookScript := preload("res://scripts/monster_variant_look.gd")
const ReactionScript := preload("res://scripts/model_reaction_controller.gd")
const MainScript := preload("res://scripts/main.gd")

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_resolver_distinct_per_rarity()
	_test_resolver_depth_axis_and_fallbacks()
	_test_ownership()
	await _test_apply_and_clear()
	await _test_shared_materials_and_bound()
	await _test_reaction_does_not_wipe_variant()
	await _test_unowned_entities_are_untouched()
	_test_main_base_tint_rule()
	_finish()


func _families() -> Array:
	var keys: Array = (LoaderScript.variants().get("families", {}) as Dictionary).keys()
	keys.sort()
	return keys


func _rarities() -> Array:
	var keys: Array = (LoaderScript.variants().get("rarities", {}) as Dictionary).keys()
	keys.sort()
	return keys


func _palettes() -> Array:
	var keys: Array = (LoaderScript.variants().get("depth", {}) as Dictionary).keys()
	keys.sort()
	return keys


func _test_resolver_distinct_per_rarity() -> void:
	var palette := str(_palettes()[0])
	for family in _families():
		var seen := {}
		for rarity in _rarities():
			var look := ResolverScript.resolve(str(family), str(rarity), palette)
			if look.is_empty() or seen.has(look["key"]):
				_fail("%s: rarity %s must resolve to a distinct look" % [family, rarity])
				return
			seen[look["key"]] = rarity
			if ResolverScript.resolve(str(family), str(rarity), palette)["key"] != look["key"]:
				_fail("%s/%s resolution must be deterministic" % [family, rarity])
				return
	_pass("every family resolves a distinct, deterministic look per rarity")


func _test_resolver_depth_axis_and_fallbacks() -> void:
	var family := str(_families()[0])
	var keys := {}
	for palette in _palettes():
		keys[ResolverScript.resolve(family, "common", str(palette))["key"]] = true
	if keys.size() != _palettes().size():
		_fail("depth palettes must produce distinct common looks for %s" % family)
		return
	var common := ResolverScript.resolve(family, "common", "")
	var unknown_rarity := ResolverScript.resolve(family, "no_such_rarity", "")
	if common["key"] != unknown_rarity["key"] or float(common["tint"]["strength"]) != 0.0:
		_fail("unknown rarity/palette must resolve to the neutral look")
		return
	if not ResolverScript.resolve("monster_dummy", "rare", "").is_empty():
		_fail("uncatalogued family must resolve to no look")
		return
	_pass("depth axis differs per palette and unknown inputs fall back to neutral")


func _test_ownership() -> void:
	var family := str(_families()[0])
	var ok := ResolverScript.owns_look({"type": "monster", "rarity": "rare"}, family)
	var boss := ResolverScript.owns_look({"type": "monster", "is_boss": true}, family)
	var boss_t := ResolverScript.owns_look({"type": "monster", "boss_template_id": "cave_warden"}, family)
	var companion := ResolverScript.owns_look({"type": "companion"}, family)
	var override := ResolverScript.owns_look({"type": "monster", "visual_model": "monster_tiny_flyer"}, family)
	var unlisted := ResolverScript.owns_look({"type": "monster"}, "monster_dummy")
	if not ok or boss or boss_t or companion or override or unlisted:
		_fail("only ordinary catalogued monsters own a variant look")
		return
	_pass("bosses, companions, overrides and uncatalogued scenes are not owned")


func _scene_root(scene_key: String) -> Node3D:
	var packed := load("res://scenes/%s.tscn" % scene_key) as PackedScene
	var root := Node3D.new()
	root.name = "MonsterVisualRoot"
	root.add_child(packed.instantiate())
	return root


func _meshes(node: Node) -> Array:
	var out: Array = []
	if node is MeshInstance3D and node.name != LookScript.AURA_NAME:
		out.append(node)
	for child in node.get_children():
		out.append_array(_meshes(child))
	return out


func _eye_mesh(root: Node, suffix: String) -> MeshInstance3D:
	for mesh in _meshes(root):
		if str(mesh.name).ends_with(suffix):
			return mesh as MeshInstance3D
	return null


func _test_apply_and_clear() -> void:
	var palette := str(_palettes()[0])
	for family in _families():
		var root := _scene_root(str(family))
		get_root().add_child(root)
		await process_frame
		var model := root.get_child(0) as Node3D
		var base_scale := model.scale
		var textured := {}
		for mesh in _meshes(root):
			var mat = (mesh as MeshInstance3D).get_active_material(0)
			if mat is StandardMaterial3D and (mat as StandardMaterial3D).albedo_texture != null:
				textured[mesh] = (mat as StandardMaterial3D).albedo_texture
		for rarity in _rarities():
			var look := ResolverScript.resolve(str(family), str(rarity), palette)
			LookScript.apply(root, look)
			LookScript.apply(root, look)  # idempotent
			var auras := root.find_children(LookScript.AURA_NAME, "MeshInstance3D", false, false)
			if auras.size() != (1 if not (look["aura"] as Dictionary).is_empty() else 0):
				_fail("%s/%s aura count must follow the look (and stay single when re-applied)" % [family, rarity])
				return
			if not model.scale.is_equal_approx(base_scale * float(look["scale_multiplier"])):
				_fail("%s/%s model scale must be base x multiplier, not compounded" % [family, rarity])
				return
			var eye_name := str(look["eye_mesh"])
			for mesh in textured:
				if eye_name != "" and str(mesh.name).ends_with(eye_name):
					continue  # the eye mesh is recolored, not detail-tinted
				var applied := (mesh as MeshInstance3D).material_override as StandardMaterial3D
				if applied == null or applied.albedo_texture != textured[mesh]:
					_fail("%s/%s tint must keep the albedo texture" % [family, rarity])
					return
				if (float(look["tint"]["strength"]) > 0.0) != applied.detail_enabled:
					_fail("%s/%s detail layer must be enabled exactly when the tint is non-zero" % [family, rarity])
					return
			if eye_name != "" and not (look["eye"] as Dictionary).is_empty():
				var eyes := _eye_mesh(root, eye_name)
				var emat := eyes.material_override as StandardMaterial3D if eyes != null else null
				if emat == null or not emat.emission_enabled or not emat.emission.is_equal_approx(look["eye"]["color"]):
					_fail("%s/%s eye glow must use the look color" % [family, rarity])
					return
			if not root.find_children("*", "Light3D", true, false).is_empty():
				_fail("variant looks must not add lights")
				return
		LookScript.clear(root)
		if not root.find_children(LookScript.AURA_NAME, "MeshInstance3D", false, false).is_empty() \
				or not model.scale.is_equal_approx(base_scale):
			_fail("%s clear must remove the aura and restore the scale" % family)
			return
		for mesh in textured:
			if (mesh as MeshInstance3D).material_override != (mesh as MeshInstance3D).get_meta("variant_source"):
				_fail("%s clear must restore the source materials" % family)
				return
		root.queue_free()
	_pass("looks apply idempotently, keep textures, add no lights, and clear back to base")


func _test_shared_materials_and_bound() -> void:
	var palettes := _palettes()
	var distinct := {}
	for family in _families():
		var a := _scene_root(str(family))
		var b := _scene_root(str(family))
		for rarity in _rarities():
			for palette in palettes:
				var look := ResolverScript.resolve(str(family), str(rarity), str(palette))
				LookScript.apply(a, look)
				LookScript.apply(b, look)
				for mesh in _meshes(a):
					var other := _eye_mesh(b, str(mesh.name))
					if other != null and other.material_override != (mesh as MeshInstance3D).material_override:
						_fail("%s/%s/%s: two monsters with one look must share materials" % [family, rarity, palette])
						return
					if (mesh as MeshInstance3D).material_override != null:
						distinct[(mesh as MeshInstance3D).material_override] = true
		a.free()
		b.free()
	if distinct.size() > LookScript.MAX_CACHED * 2:
		_fail("variant material count must stay bounded (%d)" % distinct.size())
		return
	_pass("same-look monsters share materials; the full sweep stays bounded (%d materials)" % distinct.size())


func _test_reaction_does_not_wipe_variant() -> void:
	var family := "monster_kit_skeleton_warrior"
	var root := _scene_root(family)
	get_root().add_child(root)
	await process_frame
	var look := ResolverScript.resolve(family, "unique", str(_palettes()[0]))
	LookScript.apply(root, look)
	var reaction := ReactionScript.new(root, Color.WHITE)
	var eyes := _eye_mesh(root, str(look["eye_mesh"]))
	var ring := root.find_child(LookScript.AURA_NAME, false, false) as MeshInstance3D
	reaction.set_highlight(true, Color.YELLOW)
	reaction.set_highlight(false)
	reaction.set_base_tint(Color(1.0, 0.4, 0.2))
	var emat := eyes.material_override as StandardMaterial3D
	var rmat := ring.material_override as StandardMaterial3D
	if not emat.emission_enabled or not is_equal_approx(emat.emission_energy_multiplier, float(look["eye"]["energy"])):
		_fail("eye glow must survive highlight and status tint")
		return
	if not is_equal_approx(rmat.albedo_color.a, float(look["aura"]["alpha"])):
		_fail("aura ring alpha must survive status tint")
		return
	var body: MeshInstance3D = null
	for mesh in _meshes(root):
		if mesh != eyes and (mesh as MeshInstance3D).material_override is StandardMaterial3D:
			body = mesh as MeshInstance3D
			break
	var bmat := body.material_override as StandardMaterial3D
	if not bmat.detail_enabled or not bmat.albedo_color.is_equal_approx(Color(1.0, 0.4, 0.2)):
		_fail("body keeps the variant detail tint while the status tint owns albedo_color")
		return
	reaction.dispose()
	root.queue_free()
	_pass("variant eye glow/aura/detail tint compose with highlight and status tint")


func _test_unowned_entities_are_untouched() -> void:
	var cases := [
		{"type": "monster", "monster_def_id": "dungeon_mob", "rarity": "unique", "is_boss": true, "boss_template_id": "cave_warden", "visual_model": "monster_tiny_flyer"},
		{"type": "companion", "monster_def_id": "companion_black_wolf", "rarity": "common"},
		{"type": "monster", "monster_def_id": "training_dummy", "rarity": "common"},
	]
	for entity in cases:
		var root := _scene_root("monster_dummy")
		var before_children := root.get_child_count()
		var before_mats := []
		for mesh in _meshes(root):
			before_mats.append((mesh as MeshInstance3D).material_override)
		LookScript.apply_for_entity(root, entity, str(_palettes()[0]))
		var after_mats := []
		for mesh in _meshes(root):
			after_mats.append((mesh as MeshInstance3D).material_override)
		if root.get_child_count() != before_children or before_mats != after_mats or LookScript.owns_entity(entity):
			_fail("unowned entity %s must render exactly as before" % entity.get("monster_def_id"))
			return
		root.free()
	_pass("bosses, companions and uncatalogued monsters are untouched")


func _test_main_base_tint_rule() -> void:
	var main = MainScript.new()
	var common := main._entity_base_tint({"type": "monster", "monster_def_id": "dungeon_mob", "rarity": "champion", "visual_tint": "#9fc7ff"}) as Color
	var boss := main._entity_base_tint({"type": "monster", "monster_def_id": "dungeon_mob", "rarity": "unique", "is_boss": true, "visual_tint": "#b77cff"}) as Color
	var dummy := main._entity_base_tint({"type": "monster", "monster_def_id": "training_dummy", "rarity": "rare"}) as Color
	main.free()
	if not common.is_equal_approx(Color.WHITE) or boss.is_equal_approx(Color.WHITE) or dummy.is_equal_approx(Color.WHITE):
		_fail("catalogued monsters take a neutral albedo; bosses and dummies keep their tint")
		return
	_pass("neutral albedo base only for catalogued rarity monsters")


func _pass(msg: String) -> void:
	_pass_count += 1
	print("  ok   %s" % msg)


func _fail(msg: String) -> void:
	_fail_count += 1
	print("  FAIL %s" % msg)


func _finish() -> void:
	if _fail_count > 0:
		print("[gdtest] FAIL: test_monster_variant_looks (%d failed)" % _fail_count)
		quit(1)
		return
	print("[gdtest] PASS: test_monster_variant_looks (%d checks)" % _pass_count)
	quit(0)
