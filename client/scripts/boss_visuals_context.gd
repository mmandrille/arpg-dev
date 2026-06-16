class_name BossVisualsContext
extends RefCounted

var entities: Dictionary = {}
var item_rules: Dictionary = {}
var last_server_tick: int = 0
var apply_model_tint: Callable
var apply_entity_status_tint: Callable
