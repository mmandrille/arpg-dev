## Hero corpse interactable visual (ADR-0018 P3c, v484). The corpse is the fallback kit hero
## (class_presentations fallback_class; corpse interactables carry no class) frozen at the end of
## its authored `death` clip, gold-tinted, with a ground shadow and the name label main shows on
## hover. It replaced the legacy base_human mannequin that was rotated -88° onto its side.
class_name HeroCorpseVisual
extends RefCounted

const CORPSE_TINT := Color("#d4af37")
const DEATH_CLIP := "death"
const CENTER_BONE := "spine"


## `character_scene` is the hero scene (character.tscn); `apply_tint` is main's model-tint callable.
static func make(e: Dictionary, character_scene: PackedScene, apply_tint: Callable) -> Node3D:
	var root := Node3D.new()
	root.name = "HeroCorpse_%s" % str(e.get("corpse_character_id", e.get("id", "")))
	var body := character_scene.instantiate() as Node3D
	body.name = "FallenHeroBody"
	# CharacterVisual installs the kit clips in _ready; `ready` is emitted after it.
	body.ready.connect(pose_dead.bind(body), CONNECT_ONE_SHOT)
	apply_tint.call(body, CORPSE_TINT)
	root.add_child(body)
	root.add_child(_shadow())
	root.add_child(_label(e))
	return root


## Freeze the body on the last frame of its death clip. Returns false when the clip is missing.
static func pose_dead(body: Node) -> bool:
	var player := body.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if player == null or not player.has_animation(DEATH_CLIP):
		push_warning("HeroCorpseVisual: %s has no %s clip" % [body.name, DEATH_CLIP])
		return false
	player.play(DEATH_CLIP)
	player.seek(player.get_animation(DEATH_CLIP).length, true)
	player.pause()
	_center_on_torso(body)
	return true


## The death clip falls the body about a body-length away from where it stood; slide it back so the
## torso lies over the corpse origin (shadow disc and pick box).
static func _center_on_torso(body: Node) -> void:
	var skel := body.find_child("Skeleton3D", true, false) as Skeleton3D
	var body_3d := body as Node3D
	if skel == null or body_3d == null or not skel.is_inside_tree():
		return
	var bone := skel.find_bone(CENTER_BONE)
	if bone < 0:
		return
	var torso := (skel.global_transform * skel.get_bone_global_pose(bone)).origin
	var origin := (body_3d.get_parent() as Node3D).global_position if body_3d.get_parent() is Node3D else Vector3.ZERO
	body_3d.global_position -= Vector3(torso.x - origin.x, 0.0, torso.z - origin.z)


static func _shadow() -> MeshInstance3D:
	var shadow := MeshInstance3D.new()
	shadow.name = "CorpseShadow"
	var shadow_mesh := CylinderMesh.new()
	shadow_mesh.top_radius = 0.75
	shadow_mesh.bottom_radius = 0.75
	shadow_mesh.height = 0.025
	shadow.mesh = shadow_mesh
	shadow.scale.z = 0.48
	shadow.position = Vector3(0.0, 0.015, 0.0)
	var shadow_mat := StandardMaterial3D.new()
	shadow_mat.albedo_color = Color("#171412")
	shadow.material_override = shadow_mat
	return shadow


static func _label(e: Dictionary) -> Label3D:
	var marker := Label3D.new()
	marker.name = "LootLabel"
	var corpse_name := str(e.get("corpse_name", "Hero"))
	var corpse_level := int(e.get("corpse_level", 0))
	marker.text = "%s Lv %d" % [corpse_name, corpse_level] if corpse_level > 0 else corpse_name
	marker.visible = false
	marker.font_size = 60
	marker.modulate = Color("#e8dcc8")
	marker.outline_size = 10
	marker.position = Vector3(0.0, 1.15, 0.0)
	marker.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	return marker
