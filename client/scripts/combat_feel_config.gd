## Client-only combat/movement "feel" tuning constants.
##
## Values load from `shared/assets/combat_feel_presentation.v0.json` via
## `CombatFeelPresentationLoader`. Presentation-only — never crosses the wire (ADR-0007).
class_name CombatFeelConfig
extends RefCounted

const CombatFeelPresentationLoaderScript := preload("res://scripts/combat_feel_presentation_loader.gd")

static var _initialized := false
static var ATTACK_BUFFER_SECONDS := 0.45
static var COMMAND_RETARGET_GRACE_SECONDS := 0.22
static var MOVEMENT_SMOOTHING_CATCH_UP_SPEED := 18.0
static var MOVEMENT_SMOOTHING_MAX_OFFSET := 0.70
static var MOVEMENT_SMOOTHING_RESET_DISTANCE := 1.50
static var MOVEMENT_SMOOTHING_SETTLE_EPSILON := 0.01
static var MELEE_LUNGE_DISTANCE := 0.16
static var MELEE_LUNGE_RECOVERY_SECONDS := 0.14
static var MELEE_LUNGE_SETTLE_EPSILON := 0.01
static var LEVEL_LOADING_MIN_DISPLAY_SECONDS := 0.55


static func ensure_loaded() -> void:
	if _initialized:
		return
	_initialized = true
	CombatFeelPresentationLoaderScript.ensure_loaded()
	ATTACK_BUFFER_SECONDS = CombatFeelPresentationLoaderScript.input_buffer_seconds()
	COMMAND_RETARGET_GRACE_SECONDS = CombatFeelPresentationLoaderScript.command_retarget_grace_seconds()
	var movement := CombatFeelPresentationLoaderScript.movement_smoothing()
	MOVEMENT_SMOOTHING_CATCH_UP_SPEED = float(movement.get("catch_up_speed", MOVEMENT_SMOOTHING_CATCH_UP_SPEED))
	MOVEMENT_SMOOTHING_MAX_OFFSET = float(movement.get("max_offset", MOVEMENT_SMOOTHING_MAX_OFFSET))
	MOVEMENT_SMOOTHING_RESET_DISTANCE = float(movement.get("reset_distance", MOVEMENT_SMOOTHING_RESET_DISTANCE))
	MOVEMENT_SMOOTHING_SETTLE_EPSILON = float(movement.get("settle_epsilon", MOVEMENT_SMOOTHING_SETTLE_EPSILON))
	var lunge := CombatFeelPresentationLoaderScript.melee_lunge()
	MELEE_LUNGE_DISTANCE = float(lunge.get("distance", MELEE_LUNGE_DISTANCE))
	MELEE_LUNGE_RECOVERY_SECONDS = float(lunge.get("recovery_seconds", MELEE_LUNGE_RECOVERY_SECONDS))
	MELEE_LUNGE_SETTLE_EPSILON = float(lunge.get("settle_epsilon", MELEE_LUNGE_SETTLE_EPSILON))
	LEVEL_LOADING_MIN_DISPLAY_SECONDS = CombatFeelPresentationLoaderScript.level_loading_min_display_seconds()


static func attack_buffer_seconds() -> float:
	ensure_loaded()
	return ATTACK_BUFFER_SECONDS


static func command_retarget_grace_seconds() -> float:
	ensure_loaded()
	return COMMAND_RETARGET_GRACE_SECONDS


static func movement_smoothing_catch_up_speed() -> float:
	ensure_loaded()
	return MOVEMENT_SMOOTHING_CATCH_UP_SPEED


static func movement_smoothing_max_offset() -> float:
	ensure_loaded()
	return MOVEMENT_SMOOTHING_MAX_OFFSET


static func movement_smoothing_reset_distance() -> float:
	ensure_loaded()
	return MOVEMENT_SMOOTHING_RESET_DISTANCE


static func movement_smoothing_settle_epsilon() -> float:
	ensure_loaded()
	return MOVEMENT_SMOOTHING_SETTLE_EPSILON


static func melee_lunge_distance() -> float:
	ensure_loaded()
	return MELEE_LUNGE_DISTANCE


static func melee_lunge_recovery_seconds() -> float:
	ensure_loaded()
	return MELEE_LUNGE_RECOVERY_SECONDS


static func melee_lunge_settle_epsilon() -> float:
	ensure_loaded()
	return MELEE_LUNGE_SETTLE_EPSILON


static func level_loading_min_display_seconds() -> float:
	ensure_loaded()
	return LEVEL_LOADING_MIN_DISPLAY_SECONDS


static func reset_for_tests() -> void:
	_initialized = false
	CombatFeelPresentationLoaderScript.reset_for_tests()
	ATTACK_BUFFER_SECONDS = 0.45
	COMMAND_RETARGET_GRACE_SECONDS = 0.22
	MOVEMENT_SMOOTHING_CATCH_UP_SPEED = 18.0
	MOVEMENT_SMOOTHING_MAX_OFFSET = 0.70
	MOVEMENT_SMOOTHING_RESET_DISTANCE = 1.50
	MOVEMENT_SMOOTHING_SETTLE_EPSILON = 0.01
	MELEE_LUNGE_DISTANCE = 0.16
	MELEE_LUNGE_RECOVERY_SECONDS = 0.14
	MELEE_LUNGE_SETTLE_EPSILON = 0.01
	LEVEL_LOADING_MIN_DISPLAY_SECONDS = 0.55
