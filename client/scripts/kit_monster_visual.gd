## Scene root for KayKit monsters (ADR-0018 P4a). On _ready — which runs when main adds the node
## to the tree, before it builds the AnimationController — it:
##   1. aliases the logical clips AnimationController plays (idle/walk/attack/hit/death/...)
##      onto the kit's embedded clips, per the catalog clip profile, and
##   2. mounts catalog weapon attachments on rig bones with BoneAttachment3D.
class_name KitMonsterVisual
extends Node3D

const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")

@export var visual_key := ""

var _applied := false


func _ready() -> void:
	apply_presentation()


func apply_presentation() -> void:
	if _applied:
		return
	_applied = true
	var cfg := LoaderScript.monster(visual_key)
	if cfg.is_empty():
		push_warning("KitMonsterVisual: no catalog entry for %s" % visual_key)
		return
	var profile := LoaderScript.clip_profile(str(cfg.get("clip_profile", "")))
	_alias_clips(profile)
	# v513: AnimationController reads the profile from the player (variants, windup, spawn, idle).
	var player := animation_player()
	if player != null:
		player.set_meta("kit_clip_profile", profile)
	_mount_attachments(cfg.get("attachments", []))


func animation_player() -> AnimationPlayer:
	return find_child("AnimationPlayer", true, false) as AnimationPlayer


func skeleton() -> Skeleton3D:
	return find_child("Skeleton3D", true, false) as Skeleton3D


func _alias_clips(profile: Dictionary) -> void:
	var player := animation_player()
	if player == null:
		push_warning("KitMonsterVisual: %s has no AnimationPlayer" % visual_key)
		return
	var library := player.get_animation_library("")
	if library == null:
		return
	var loops: Array = profile.get("loop", [])
	var clips: Dictionary = profile.get("clips", {})
	for logical in clips:
		var kit_clip := str(clips[logical])
		if library.has_animation(logical):
			continue
		if not library.has_animation(kit_clip):
			push_warning("KitMonsterVisual: %s missing kit clip %s for %s" % [visual_key, kit_clip, logical])
			continue
		# A copy per alias so loop flags never leak into the kit's own clip.
		var anim := library.get_animation(kit_clip).duplicate() as Animation
		anim.loop_mode = Animation.LOOP_LINEAR if loops.has(logical) else Animation.LOOP_NONE
		library.add_animation(StringName(logical), anim)


func _mount_attachments(attachments: Array) -> void:
	var skel := skeleton()
	if skel == null:
		return
	for raw in attachments:
		var entry := raw as Dictionary
		var bone := str(entry.get("bone", ""))
		if skel.find_bone(bone) < 0:
			push_warning("KitMonsterVisual: %s has no bone %s" % [visual_key, bone])
			continue
		var piece := LibraryScript.instantiate(str(entry.get("asset_id", "")))
		if piece == null:
			continue
		var mount := BoneAttachment3D.new()
		mount.name = "Attach_%s" % bone.replace(".", "_")
		mount.bone_name = bone
		skel.add_child(mount)
		mount.add_child(piece)
