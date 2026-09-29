## KayKit dungeon props that stand in for gameplay entities (ADR-0018 P2b):
## wall-mounted torches and treasure chests. Presentation only — the server never sees these.
class_name DungeonKitProps
extends RefCounted

const LoaderScript := preload("res://scripts/dungeon_kit_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const ChestPresentationScript := preload("res://scripts/chest_presentation.gd")

## Node names shared with the procedural chest so InteractableStatePresentation (open tween,
## inner glow) and bot debug keep working unchanged.
const LID_PIVOT_NAME := "ChestLidPivot"
const INNER_GLOW_NAME := "ChestInnerGlow"


## Yaw that turns the kit torch's local +Z (bracket -> flame) toward `facing` (x, z).
static func torch_yaw(facing: Vector2) -> float:
	if facing.length_squared() <= 0.0:
		return 0.0
	return atan2(facing.x, facing.y)


## Kit torch body (no light): the caller keeps owning the OmniLight and emissive flame so the
## fog-of-war torch holes and glow stay data-driven by dungeon_torch_presentation.
static func make_torch_body(facing: Vector2) -> Node3D:
	var cfg := LoaderScript.torch_config()
	var piece := LibraryScript.instantiate(str(cfg.get("asset_id", "")))
	if piece == null:
		return null
	piece.name = "KitTorch"
	piece.rotation.y = torch_yaw(facing)
	return piece


## Flame anchor in torch-local space (rotated with the torch body).
static func flame_offset(facing: Vector2) -> Vector3:
	var cfg := LoaderScript.torch_config()
	var raw: Dictionary = cfg.get("flame_offset", {})
	var local := Vector3(float(raw.get("x", 0.0)), float(raw.get("y", 0.0)), float(raw.get("z", 0.0)))
	return Basis(Vector3.UP, torch_yaw(facing)) * local


static func flame_scale() -> float:
	return float(LoaderScript.torch_config().get("flame_scale", 1.0))


static func make_chest_node(def_id: String, elite_objective: bool = false, quest_reward: bool = false) -> Node3D:
	var cfg := LoaderScript.chest_config()
	var asset_key := "elite_objective_asset_id" if elite_objective else "asset_id"
	var piece := LibraryScript.instantiate(str(cfg.get(asset_key, cfg.get("asset_id", ""))))
	if piece == null:
		return null
	var root := Node3D.new()
	root.name = "TreasureChest"
	root.set_meta("kit_chest", true)
	var scale := float(cfg.get("scale", 1.0))
	piece.scale = Vector3.ONE * scale
	root.add_child(piece)
	# The kit lid node already sits on the hinge (back-top edge) with the lid extending toward +Z,
	# so rotating it about X opens the chest exactly like the procedural ChestLidPivot.
	var lid_node := str(cfg.get("elite_objective_lid_node" if elite_objective else "lid_node", ""))
	var lid := piece.find_child(lid_node, true, false)
	if lid != null:
		lid.name = LID_PIVOT_NAME
	else:
		push_warning("DungeonKitProps: chest lid node %s missing for %s" % [lid_node, def_id])
	var box := LibraryScript.bounds(str(cfg.get(asset_key, "")))
	var glow := ChestPresentationScript.add_part(
		root, INNER_GLOW_NAME, Vector3(box.size.x * 0.7, 0.04, box.size.z * 0.5) * scale,
		Vector3(0.0, box.end.y * scale * 0.62, 0.0), Color("#f5b449")
	)
	var glow_mat := glow.material_override as StandardMaterial3D
	glow_mat.emission_enabled = true
	glow_mat.emission = Color("#f5b449")
	glow.visible = false
	ChestPresentationScript.sync_objective_marker(root, elite_objective, false)
	ChestPresentationScript.sync_quest_marker(root, quest_reward, false)
	return root

