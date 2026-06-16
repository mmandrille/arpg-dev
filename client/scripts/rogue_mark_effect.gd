extends Node3D


func _init() -> void:
	name = "RogueMarkSkullEffect"
	position.y = 3.05

	var skull := Label3D.new()
	skull.name = "RogueMarkSkull"
	skull.text = "☠"
	skull.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	skull.font_size = 192
	skull.pixel_size = 0.006
	skull.modulate = Color(1.0, 0.08, 0.05, 1.0)
	skull.outline_size = 10
	skull.outline_modulate = Color(0.08, 0.0, 0.0, 0.95)
	add_child(skull)

	var pulse := create_tween()
	pulse.set_loops()
	pulse.tween_property(self, "scale", Vector3.ONE * 1.08, 0.42).from(Vector3.ONE * 0.98)
	pulse.tween_property(self, "scale", Vector3.ONE * 0.98, 0.42)
