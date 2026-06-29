class_name AttackMoveInputCoordinator
extends RefCounted

const ClientConstantsScript := preload("res://scripts/client_constants.gd")
const CombatReachScript := preload("res://scripts/combat_reach.gd")


static func living_monster_target(entities: Dictionary, target_id: String) -> bool:
	return target_id != "" and entities.has(target_id) \
		and str(entities[target_id].get("type", "")) == "monster" \
		and int(entities[target_id].get("hp", 1)) > 0


static func target_in_local_attack_range(host, target_id: String) -> bool:
	return CombatReachScript.target_in_local_attack_range(
		host.player_anchor,
		host.entities,
		host.inventory,
		host.equipped,
		target_id,
	)


static func defer_monster_click(host, target_id: String) -> void:
	if target_in_local_attack_range(host, target_id):
		host._sticky_attack.clear()
		queue_attack_buffer(host, target_id)
	else:
		start_attack_move(host, target_id)


static func start_attack_move(host, target_id: String) -> void:
	if not living_monster_target(host.entities, target_id):
		clear_pending_attack_commands(host)
		return
	if host._path_reject_backoff.blocks_target(target_id, Time.get_ticks_msec()):
		return
	host._attack_buffer.clear()
	host._sticky_attack.set_target(target_id)
	var goal := CombatReachScript.attack_approach_point(
		host.player_anchor,
		host.entities,
		host.inventory,
		host.equipped,
		target_id,
		host._last_facing_direction,
	)
	if host._path_reject_backoff.blocks_goal(Vector2(goal.x, goal.z), Time.get_ticks_msec()):
		return
	host._close_gameplay_panels_for_movement()
	host._mark_local_player_walking()
	host.client.send("move_to_intent", host.last_server_tick, {"position": {"x": goal.x, "y": goal.z}})
	if host._attack_cooldown <= 0.0:
		host._attack_cooldown = ClientConstantsScript.SEND_INTERVAL


static func clear_pending_attack_commands(host) -> void:
	host._attack_buffer.clear()
	host._sticky_attack.clear()


static func queue_attack_buffer(host, target_id: String) -> void:
	if not living_monster_target(host.entities, target_id):
		host._attack_buffer.clear()
		return
	host._attack_buffer.queue_attack(target_id)


static func tick_sustained_click(host) -> void:
	if not host._sustained_click.active:
		return
	if host._attack_cooldown > 0.0:
		return
	if host._sustained_click.should_stop(host.player_hp, host.entities):
		host._sustained_click.clear()
		return
	match host._sustained_click.mode:
		"attack":
			repeat_hold_attack(host)
		"directional_attack":
			host._repeat_directional_attack()
		"move":
			repeat_hold_move(host)


static func tick_attack_buffer(host, delta: float) -> void:
	host._attack_buffer.tick(delta)
	if not host._attack_buffer.active():
		return
	if host._attack_buffer.should_clear(host.player_hp, host.entities):
		host._attack_buffer.clear()
		return
	if host._attack_cooldown > 0.0:
		return
	host._try_dispatch_monster_attack(host._attack_buffer.target_id, false)


static func tick_sticky_attack(host) -> void:
	if not host._sticky_attack.active():
		return
	if host._sticky_attack.should_clear(host.player_hp, host.entities):
		host._sticky_attack.clear()
		return
	if host._attack_cooldown <= 0.0:
		host._try_dispatch_monster_attack(host._sticky_attack.target_id, false)


static func repeat_hold_attack(host) -> void:
	var target_id: String = str(host._sustained_click.target_id)
	if target_id == "" or not host.entities.has(target_id):
		host._sustained_click.clear()
		return
	if not target_in_local_attack_range(host, target_id):
		start_attack_move(host, target_id)
		return
	host._try_dispatch_monster_attack(target_id, false)


static func repeat_hold_move(host) -> void:
	if host._is_force_stand_held():
		host._sustained_click.clear()
		return
	var ground: Vector3 = host._mouse_ground_point()
	if not host._sustained_click.can_repeat_move(ground):
		return
	host._close_gameplay_panels_for_movement()
	host._mark_local_player_walking()
	host.client.send("move_to_intent", host.last_server_tick, {"position": {"x": ground.x, "y": ground.z}})
	host._sustained_click.mark_move_sent(ground)
	host._attack_cooldown = ClientConstantsScript.SEND_INTERVAL
