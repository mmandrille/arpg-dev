## PlayerCameraContext — lightweight data bag passed from main.gd to the camera controller.
## All fields are plain references; none are owned or freed here.
class_name PlayerCameraContext
extends RefCounted

var player_anchor: Node3D           ## Node3D the player character anchor rides on.
var character_visual: Node3D        ## Used to resolve sockets (e.g. chest_socket).
var client_settings                  ## ClientSettings reference (avoid typed ref).
var menu_blocks_gameplay: Callable   ## () -> bool; true when menus block input.

## Convenience factory — populate all fields in one call.
static func make(anchor: Node3D, visual: Node3D, settings, blocks: Callable) -> RefCounted:
	var ctx: RefCounted = load("res://scripts/player_camera_context.gd").new()
	ctx.set("player_anchor", anchor)
	ctx.set("character_visual", visual)
	ctx.set("client_settings", settings)
	ctx.set("menu_blocks_gameplay", blocks)
	return ctx
