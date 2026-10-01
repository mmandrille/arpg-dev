## v506: small client-only town residents driven by the shared town presentation catalog.
## Actors reuse the registered KayKit class bodies and the shared hero idle/walk clips. They have
## no collision, interaction, or server-facing nodes; a fixed patrol is presentation only.
class_name TownAmbientLife
extends Node3D

const CharacterScene := preload("res://scenes/character.tscn")
const ClassPresentationsScript := preload("res://scripts/class_presentations_loader.gd")
const ClassIdleStanceScript := preload("res://scripts/class_idle_stance.gd")
const KitHeroClipsScript := preload("res://scripts/kit_hero_clips.gd")
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")

const ROOT_NAME := "TownAmbientLife"
const RESIDENT_PREFIX := "AmbientResident_"

var _elapsed_s := 0.0
var _actors: Array[Dictionary] = []


static func build(dressing: Dictionary) -> Node3D:
	var ambient_script := load("res://scripts/town_ambient_life.gd") as Script
	var layer := ambient_script.new() as Node3D
	layer.name = ROOT_NAME
	var cfg := dressing.get("ambient_life", {}) as Dictionary
	if not bool(cfg.get("enabled", false)):
		layer.set_process(false)
		return layer
	var residents: Array = cfg.get("actors", [])
	var has_patrol := false
	for raw in residents:
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var actor_cfg := raw as Dictionary
		var resident: Dictionary = layer.call("_make_resident", actor_cfg)
		if resident.is_empty():
			continue
		layer.add_child(resident["node"] as Node3D)
		var actors: Array = layer.get("_actors")
		actors.append(resident)
		layer.set("_actors", actors)
		has_patrol = has_patrol or str(actor_cfg.get("mode", "idle")) == "patrol"
	if not has_patrol:
		layer.set_process(false)
	return layer


func _process(delta: float) -> void:
	_elapsed_s += maxf(delta, 0.0)
	for resident in _actors:
		var cfg: Dictionary = resident["config"]
		if str(cfg.get("mode", "idle")) != "patrol":
			continue
		var node := resident["node"] as Node3D
		var state := route_state(cfg, _elapsed_s)
		var position: Vector2 = state["position"]
		node.position = Vector3(position.x, 0.0, position.y)
		var direction: Vector2 = state["direction"]
		if direction.length_squared() > 0.0:
			# Imported KayKit heroes face local -Z. Convert town X/Z travel into that heading.
			node.rotation.y = atan2(-direction.x, -direction.y)
		var moving := bool(state["moving"])
		if bool(resident["moving"]) != moving:
			(resident["controller"] as AnimationController).set_locomotion(moving)
			resident["moving"] = moving


## Pure deterministic patrol sampler shared with the focused test. ``elapsed_s`` is the town-local
## presentation clock; ``phase_s`` offsets the actor without introducing randomness.
static func route_state(cfg: Dictionary, elapsed_s: float) -> Dictionary:
	if str(cfg.get("mode", "idle")) != "patrol":
		var fixed: Dictionary = cfg.get("position", {})
		return {
			"position": Vector2(float(fixed.get("x", 0.0)), float(fixed.get("y", 0.0))),
			"direction": Vector2.ZERO,
			"moving": false,
		}
	var path: Array = cfg.get("path", [])
	if path.size() < 2:
		return {"position": Vector2.ZERO, "direction": Vector2.ZERO, "moving": false}
	var speed := maxf(float(cfg.get("speed_mps", 0.0)), 0.001)
	var pause_s := maxf(float(cfg.get("pause_s", 0.0)), 0.0)
	var cycle_s := 0.0
	for i in path.size():
		var cycle_start := _point(path[i])
		var cycle_end := _point(path[(i + 1) % path.size()])
		cycle_s += cycle_start.distance_to(cycle_end) / speed + pause_s
	var phase := fposmod(elapsed_s + float(cfg.get("phase_s", 0.0)), maxf(cycle_s, 0.001))
	for i in path.size():
		var start_point := _point(path[i])
		var end_point := _point(path[(i + 1) % path.size()])
		var travel_s := start_point.distance_to(end_point) / speed
		if phase < travel_s:
			var direction := (end_point - start_point).normalized()
			return {"position": start_point.lerp(end_point, phase / travel_s), "direction": direction, "moving": true}
		phase -= travel_s
		if phase < pause_s:
			return {"position": end_point, "direction": Vector2.ZERO, "moving": false}
		phase -= pause_s
	var first := _point(path[0])
	return {"position": first, "direction": Vector2.ZERO, "moving": false}


func _make_resident(cfg: Dictionary) -> Dictionary:
	var class_id := str(cfg.get("class_id", ""))
	var resolved := ClassPresentationsScript.resolve(class_id)
	var packed := ClassPresentationsScript.packed_scene_for_class(class_id)
	var clips: AnimationLibrary = KitHeroClipsScript.library(str(resolved.get("clip_profile", "")))
	if packed == null or clips == null:
		push_warning("TownAmbientLife: cannot resolve resident %s" % str(cfg.get("id", "")))
		return {}
	var model := packed.instantiate() as Node3D
	if model == null:
		return {}
	model.name = "ModelRoot"
	model.scale = Vector3.ONE * float(resolved.get("scale", 1.0))
	model.position.y = float(resolved.get("height_offset", 0.0))
	ClassIdleStanceScript.apply_to_model(model, class_id)
	var visual := CharacterScene.instantiate() as Node3D
	if visual == null:
		model.free()
		return {}
	# Reuse the same model/AnimationPlayer hierarchy as a player, without attaching player gear
	# sockets or the CharacterVisual script's equipment hooks to ambient townsfolk.
	visual.set_script(null)
	visual.name = "CharacterVisual"
	var old_model := visual.get_node_or_null("ModelRoot")
	if old_model != null:
		visual.remove_child(old_model)
		old_model.free()
	visual.add_child(model)
	visual.move_child(model, 0)
	var player := visual.get_node_or_null("AnimationPlayer") as AnimationPlayer
	if player == null:
		visual.free()
		return {}
	player.root_node = NodePath("../ModelRoot")
	if player.has_animation_library(""):
		player.remove_animation_library("")
	player.add_animation_library("", clips)
	var controller := AnimationControllerScript.new(player) as AnimationController
	var resident := Node3D.new()
	resident.name = RESIDENT_PREFIX + str(cfg.get("id", "resident"))
	resident.add_child(visual)
	var state := route_state(cfg, 0.0)
	if str(cfg.get("mode", "idle")) == "idle":
		var point: Dictionary = cfg.get("position", {})
		resident.position = Vector3(float(point.get("x", 0.0)), 0.0, float(point.get("y", 0.0)))
		resident.rotation.y = deg_to_rad(float(cfg.get("yaw_degrees", 0.0)))
	else:
		var point: Vector2 = state["position"]
		resident.position = Vector3(point.x, 0.0, point.y)
		var direction: Vector2 = state["direction"]
		if direction.length_squared() > 0.0:
			resident.rotation.y = atan2(-direction.x, -direction.y)
		controller.set_locomotion(bool(state["moving"]))
	return {"node": resident, "config": cfg.duplicate(true), "controller": controller, "moving": bool(state["moving"])}


static func _point(raw) -> Vector2:
	if typeof(raw) != TYPE_DICTIONARY:
		return Vector2.ZERO
	var point := raw as Dictionary
	return Vector2(float(point.get("x", 0.0)), float(point.get("y", 0.0)))
