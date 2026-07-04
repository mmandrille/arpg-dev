class_name ShowmeSkillIconCapture
extends RefCounted

const SkillIconScript := preload("res://scripts/skill_icon.gd")
const SkillRulesLoaderScript := preload("res://scripts/skill_rules_loader.gd")


static func setup(capture: SceneTree, skill_id: String) -> void:
	var root := Control.new()
	root.name = "VisualFeedbackSkillIcon"
	root.set_anchors_preset(Control.PRESET_FULL_RECT)
	capture.get_root().add_child(root)

	var panel := Panel.new()
	panel.custom_minimum_size = Vector2(320, 360)
	panel.position = Vector2(160, 60)
	panel.add_theme_stylebox_override("panel", _panel_style())
	root.add_child(panel)

	var layout := VBoxContainer.new()
	layout.position = Vector2(24, 24)
	layout.add_theme_constant_override("separation", 12)
	panel.add_child(layout)

	SkillRulesLoaderScript.ensure_loaded()
	var effective_id := skill_id if skill_id != "" else SkillRulesLoaderScript.first_skill_id()
	var presentation := SkillRulesLoaderScript.skill_presentation(effective_id)
	var display_name := SkillRulesLoaderScript.skill_display_name(effective_id)

	var icon := SkillIconScript.new()
	icon.custom_minimum_size = Vector2(128, 128)
	icon.size = Vector2(128, 128)
	icon.configure(effective_id, presentation, 1)
	layout.add_child(icon)

	var title := Label.new()
	title.text = display_name
	title.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	title.add_theme_font_size_override("font_size", 22)
	title.add_theme_color_override("font_color", Color("#f0dfbb"))
	layout.add_child(title)

	var meta := Label.new()
	var shape := str(presentation.get("icon", {}).get("shape", ""))
	var label := str(presentation.get("icon", {}).get("label", ""))
	meta.text = "%s\n%s · %s" % [effective_id, shape, label]
	meta.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	meta.add_theme_font_size_override("font_size", 14)
	meta.add_theme_color_override("font_color", Color("#9a9080"))
	layout.add_child(meta)

	await capture.process_frame


static func _panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color("#1a1b1d")
	style.border_color = Color("#4a4034")
	style.set_border_width_all(2)
	style.set_corner_radius_all(4)
	return style
