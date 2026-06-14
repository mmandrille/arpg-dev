class_name EarthbreakerJump

static var active_tween: Tween

static func play(character_visual: Node3D, owner: Node) -> void:
	if character_visual == null or owner == null:
		return
	if active_tween != null:
		active_tween.kill()
	character_visual.position.y = 0.0
	active_tween = owner.create_tween()
	active_tween.set_trans(Tween.TRANS_QUAD)
	active_tween.set_ease(Tween.EASE_OUT)
	active_tween.tween_property(character_visual, "position:y", 0.34, 0.10)
	active_tween.set_ease(Tween.EASE_IN)
	active_tween.tween_property(character_visual, "position:y", 0.0, 0.13)
