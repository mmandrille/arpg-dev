## Pure resolver for v511 monster variant looks: (scene key, rarity, depth palette id) -> look.
## Data: shared/assets/kit_monster_presentation.v0.json `variants`. No scene-tree access, so it
## is testable on its own. Unknown family -> {} (the model keeps its base look); unknown
## rarity/palette -> the neutral contribution for that axis.
class_name MonsterVariantResolver
extends RefCounted

const LoaderScript := preload("res://scripts/kit_monster_presentation_loader.gd")
const BOSS_VISUAL_MODEL := "current_humanoid_player"


## Looks are only applied to ordinary catalogued dungeon monsters. Bosses, companions and
## explicit `visual_model` overrides keep their existing explicit visual path.
static func owns_look(entity: Dictionary, scene_key: String, variants: Dictionary = {}) -> bool:
	if str(entity.get("type", "")) != "monster" or str(entity.get("character_class", "")) != "":
		return false
	if bool(entity.get("is_boss", false)) or str(entity.get("boss_template_id", "")) != "":
		return false
	var model := str(entity.get("visual_model", ""))
	if model != "" and model != scene_key:
		return false
	var cfg := variants if not variants.is_empty() else LoaderScript.variants()
	return (cfg.get("families", {}) as Dictionary).has(scene_key)


static func resolve(scene_key: String, rarity: String, palette_id: String, variants: Dictionary = {}) -> Dictionary:
	var cfg := variants if not variants.is_empty() else LoaderScript.variants()
	var family: Dictionary = (cfg.get("families", {}) as Dictionary).get(scene_key, {})
	if family.is_empty():
		return {}
	var rarities: Dictionary = cfg.get("rarities", {})
	var rarity_cfg: Dictionary = rarities.get(rarity, rarities.get("common", {}))
	var depth_cfg: Dictionary = (cfg.get("depth", {}) as Dictionary).get(palette_id, {})
	var tint := _blend_tints(depth_cfg.get("tint", {}), rarity_cfg.get("tint", {}))
	var look := {
		"scene_key": scene_key,
		"tint": tint,
		"scale_multiplier": float(rarity_cfg.get("scale_multiplier", 1.0)),
		"eye_mesh": str(family.get("eye_mesh", "")),
		"eye": {},
		"aura": {},
	}
	var eye: Dictionary = rarity_cfg.get("eye", {})
	if not eye.is_empty() and look["eye_mesh"] != "":
		look["eye"] = {"color": Color(str(eye.get("color", "#ffffff"))), "energy": float(eye.get("energy", 1.0))}
	var aura: Dictionary = rarity_cfg.get("aura", {})
	if not aura.is_empty():
		look["aura"] = {
			"color": Color(str(aura.get("color", "#ffffff"))),
			"alpha": float(aura.get("alpha", 0.2)),
			"radius": float(aura.get("radius", 0.6)) * float(family.get("aura_radius_scale", 1.0)),
			"energy": float(aura.get("energy", 0.5)),
		}
	look["key"] = signature(look)
	return look


## Stable identity of the visible result, used for material caches and equality checks.
static func signature(look: Dictionary) -> String:
	var tint: Dictionary = look.get("tint", {})
	var parts: Array[String] = [
		(tint.get("color", Color.WHITE) as Color).to_html(false),
		"%.3f" % float(tint.get("strength", 0.0)),
		"%.3f" % float(look.get("scale_multiplier", 1.0)),
	]
	var eye: Dictionary = look.get("eye", {})
	if not eye.is_empty():
		parts.append("e%s:%.2f" % [(eye["color"] as Color).to_html(false), float(eye["energy"])])
	var aura: Dictionary = look.get("aura", {})
	if not aura.is_empty():
		parts.append("a%s:%.2f:%.2f:%.2f" % [(aura["color"] as Color).to_html(false), float(aura["alpha"]), float(aura["radius"]), float(aura["energy"])])
	return "|".join(parts)


## Two overlays on the detail layer combine as independent coverages: the result strength is
## 1-(1-d)(1-r) and the colour leans toward whichever overlay is stronger.
static func _blend_tints(depth: Dictionary, rarity: Dictionary) -> Dictionary:
	var ds := clampf(float(depth.get("strength", 0.0)), 0.0, 1.0)
	var rs := clampf(float(rarity.get("strength", 0.0)), 0.0, 1.0)
	var dc := Color(str(depth.get("color", "#ffffff")))
	var rc := Color(str(rarity.get("color", "#ffffff")))
	if ds + rs <= 0.0:
		return {"color": Color.WHITE, "strength": 0.0}
	return {"color": dc.lerp(rc, rs / (ds + rs)), "strength": 1.0 - (1.0 - ds) * (1.0 - rs)}
