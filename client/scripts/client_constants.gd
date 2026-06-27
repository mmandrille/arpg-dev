class_name ClientConstants
extends RefCounted

const MONSTER_EVENT_CLIPS := {
	"monster_damaged": "hit",
	"monster_killed": "death",
}
const PLAYER_EVENT_CLIPS := {
	"player_damaged": "hit",
	"player_killed": "death",
}
const PLAYER_START_HP := 10
const INTERACTABLE_ACTIVATION_RANGE := 1.5
const SKILL_FUNCTION_KEY_COUNT := 16
const LOCAL_UNARMED_REACH := 1.0
const LOCAL_MONSTER_RADIUS := 0.45
const LOCAL_LOOT_RADIUS := 0.35
const LOCAL_INTERACTABLE_RADIUS := 0.50
const LOCAL_REACH_EPSILON := 0.000001
const PLAYER_TINT := Color("#8fe8a7")
const REMOTE_PLAYER_TINT := Color("#202934")
const POISON_TINT := Color("#38f06f")
const BAG_FULL_CANT_UNEQUIP_TEXT := "bag full, cant unequip"
const NO_MANA_TEXT := "NO MANA"
const MONSTER_RARITY_TINTS := {
	"common": Color("#f2f2ec"),
	"champion": Color("#9fc7ff"),
	"rare": Color("#ff9b9b"),
	"unique": Color("#ffd978"),
}
const ITEM_RARITY_BACKGROUNDS := {
	"common": Color("#343432"),
	"magic": Color("#1b3458"),
	"rare": Color("#5a4520"),
	"unique": Color("#5a2f17"),
	"set": Color("#173f28"),
}
const LOOT_LABEL_RARITY_COLORS := {
	"common": Color("#e8dcc8"),
	"magic": Color("#93c5fd"),
	"rare": Color("#f4d481"),
	"unique": Color("#ffb26b"),
	"set": Color("#55e66f"),
}
const LOOT_LABEL_CATEGORY_COLORS := {
	"currency": Color("#ffd75e"),
	"quest": Color("#6ee68b"),
	"consumable": Color("#ff8f70"),
}
const GROUND_EQUIPMENT_MODEL_SCALE := 1.0
const BOSS_VISUAL_MODEL := "current_humanoid_player"
const BOSS_PHASE_TICK_RATE := 10.0
const BOSS_TELEGRAPH_MARKER_NAME := "BossTelegraphMarker"
const ARCHER_MONSTER_DEF_ID := "dungeon_archer"
const ARCHER_BOW_MARKER_NAME := "ArcherBowMarker"
const CHARACTER_FLOW_CREATE_GAME := "create_game"
const CHARACTER_FLOW_JOIN_GAME := "join_game"
const CHARACTER_FLOW_LEGACY_SOLO := "solo"
const CHARACTER_FLOW_LEGACY_MULTIPLAYER_HOST := "multiplayer_host"
const CHARACTER_FLOW_LEGACY_MULTIPLAYER_JOIN := "multiplayer_join"

const SEND_INTERVAL := 0.1
const SERVER_TICK_RATE := 10.0
const DEFAULT_ATTACK_INTERVAL_TICKS := 20
const PLAYER_SPEED := 2.8
const WALK_ANIMATION_LINGER_SECONDS := 0.28
const CAMERA_ZOOM_DEFAULT := 20.0
const CAMERA_ZOOM_STEP := 1.5
const CAMERA_ZOOM_MIN := 8.0
const CAMERA_ZOOM_MAX := 40.0
const CAMERA_FOLLOW_OFFSET := Vector3(9.0, 20.0, 15.0)
const GROUND_TEXTURE_TOWN := "town_grass"
const GROUND_TEXTURE_DUNGEON := "dungeon_rock"
const WALL_TEXTURE_CAVE := "cave_wall"

