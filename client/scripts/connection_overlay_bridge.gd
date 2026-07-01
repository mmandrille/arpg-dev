class_name ConnectionOverlayBridge
extends RefCounted

const ConnectionOverlayScript := preload("res://scripts/connection_overlay.gd")


static func install(parent: Node, on_cancel: Callable, on_return_to_menu: Callable) -> ConnectionOverlay:
	var overlay: ConnectionOverlay = ConnectionOverlayScript.new()
	overlay.cancel_pressed.connect(on_cancel)
	overlay.return_to_menu_pressed.connect(on_return_to_menu)
	var layer := CanvasLayer.new()
	layer.layer = 25
	layer.add_child(overlay)
	parent.add_child(layer)
	return overlay
