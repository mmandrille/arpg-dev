## Client-only combat/movement "feel" tuning constants.
##
## These are presentation-feel values (input buffering, retarget grace, movement
## smoothing, melee lunge) that never cross the wire — animation/feel is client-only
## per ADR-0007. The Data-Driven Configuration Policy puts balance-sensitive tuning in
## shared/rules/*.json by default; these are intentionally code-owned because they are
## local presentation timing the server never reads. Keep server-authoritative gameplay
## tuning (damage, speed, costs) out of this file.
class_name CombatFeelConfig
extends RefCounted

const ATTACK_BUFFER_SECONDS := 0.45
const COMMAND_RETARGET_GRACE_SECONDS := 0.22

const MOVEMENT_SMOOTHING_CATCH_UP_SPEED := 18.0
const MOVEMENT_SMOOTHING_MAX_OFFSET := 0.70
const MOVEMENT_SMOOTHING_RESET_DISTANCE := 1.50
const MOVEMENT_SMOOTHING_SETTLE_EPSILON := 0.01

const MELEE_LUNGE_DISTANCE := 0.16
const MELEE_LUNGE_RECOVERY_SECONDS := 0.14
const MELEE_LUNGE_SETTLE_EPSILON := 0.01
const LEVEL_LOADING_MIN_DISPLAY_SECONDS := 0.55
