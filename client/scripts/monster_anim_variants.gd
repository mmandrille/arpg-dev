## Pure selection / timing math for monster combat animation polish (v513).
## No scene access: importable and unit-testable on its own. Data comes from the
## clip profile in shared/assets/kit_monster_presentation.v0.json; every function
## degrades to "use the base clip" when the optional profile keys are absent.
class_name MonsterAnimVariants
extends RefCounted

const PROFILE_META := "kit_clip_profile"
const DEFAULT_CONTACT_FRACTION := 0.45
const DEFAULT_TICK_SECONDS := 0.1


static func profile_of(player: AnimationPlayer) -> Dictionary:
	if player == null or not player.has_meta(PROFILE_META):
		return {}
	return player.get_meta(PROFILE_META) as Dictionary


static func stable_key(text: String) -> int:
	return absi(text.hash())


## Variant clips for a logical group, always non-empty (falls back to [group]).
static func group_clips(profile: Dictionary, group: String) -> Array:
	var variants: Dictionary = profile.get("variants", {})
	var clips: Array = variants.get(group, [])
	return clips if not clips.is_empty() else [group]


## Deterministic pick: entity-keyed offset plus a per-entity counter, so one monster cycles
## through its variants and a crowd starts out of step.
static func pick(group: Array, entity_key: String, counter: int) -> String:
	if group.is_empty():
		return ""
	return str(group[(stable_key(entity_key) + maxi(counter, 0)) % group.size()])


## Playback speed so the clip's contact point lands when the server windup ends.
static func windup_speed_scale(clip_length_s: float, contact_fraction: float, windup_ticks: int, tick_seconds: float, speed_min: float, speed_max: float) -> float:
	if clip_length_s <= 0.0 or windup_ticks <= 0 or tick_seconds <= 0.0:
		return 1.0
	var contact := clampf(contact_fraction, 0.05, 1.0)
	var raw := clip_length_s * contact / (float(windup_ticks) * tick_seconds)
	return clampf(raw, minf(speed_min, speed_max), maxf(speed_min, speed_max))


## "left" or "right" of the monster's own facing. `forward` and `to_source` are flat world vectors.
static func hit_side(forward: Vector3, to_source: Vector3) -> String:
	return "left" if forward.cross(to_source).y >= 0.0 else "right"


static func idle_interval(profile: Dictionary, entity_key: String, counter: int) -> float:
	var cfg: Dictionary = profile.get("idle_variation", {})
	if cfg.is_empty():
		return 0.0
	var lo := float(cfg.get("min_interval_s", 0.0))
	var hi := maxf(float(cfg.get("max_interval_s", lo)), lo)
	var unit := float((stable_key(entity_key) + counter * 7919) % 1000) / 999.0
	return lerpf(lo, hi, unit)


## Fraction in [0, 1) of the idle loop a monster starts at.
static func phase_fraction(entity_key: String) -> float:
	return float(stable_key(entity_key + "#phase") % 1000) / 1000.0


## Resolves the hit clip for a profile: directional clip when configured, else variant rotation.
static func hit_clip(profile: Dictionary, entity_key: String, counter: int, forward: Vector3, to_source: Vector3, has_source: bool) -> String:
	var directional: Dictionary = profile.get("hit_directional", {})
	if not directional.is_empty() and has_source:
		return str(directional.get(hit_side(forward, to_source), "hit"))
	return pick(group_clips(profile, "hit"), entity_key, counter)
