from __future__ import annotations

from typing import Any

from tools.bot.runtime_economy_assertions import handle_runtime_economy_assertion


def run_assertions(
    assertions: list[Any],
    entities: list[dict],
    inventory: list[dict],
    equipped: dict,
    item_id: str | None,
    where: str,
    current_level: int | None = None,
    walls: list[dict] | None = None,
    discovered_teleporters: dict[int, bool] | None = None,
    character_progression: dict[str, Any] | None = None,
    hotbar_capacity: int | None = None,
    hotbar: list[dict] | None = None,
    inventory_rows: int | None = None,
    inventory_capacity: int | None = None,
    gold: int | None = None,
    stash_items: list[dict[str, Any]] | None = None,
    stash_gold: int | None = None,
    stash_capacity: int | None = None,
    resource_wallet: dict[str, int] | None = None,
    skill_progression: dict[str, Any] | None = None,
    skill_cooldowns: list[dict[str, Any]] | None = None,
    helpers: dict[str, Any] | None = None,
) -> None:
    if helpers is None:
        raise AssertionError("run_assertions requires helper bindings")

    assert_count_matches = helpers["assert_count_matches"]
    assert_equipped_sword = helpers["assert_equipped_sword"]
    assert_player_damaged = helpers["assert_player_damaged"]
    assert_monster_dead = helpers["assert_monster_dead"]
    assert_inventory_contains = helpers["assert_inventory_contains"]
    assert_inventory_requirement_status = helpers["assert_inventory_requirement_status"]
    assert_loot_requirement_status = helpers["assert_loot_requirement_status"]
    assert_stash_item_count = helpers["assert_stash_item_count"]
    assert_rolled_inventory_item = helpers["assert_rolled_inventory_item"]
    assert_rolled_inventory_any = helpers["assert_rolled_inventory_any"]
    assert_inventory_unique_effect_coverage = helpers["assert_inventory_unique_effect_coverage"]
    assert_entity_count = helpers["assert_entity_count"]
    assert_hotbar_slot = helpers["assert_hotbar_slot"]
    assert_character_progression = helpers["assert_character_progression"]
    assert_skill_progression = helpers["assert_skill_progression"]
    assert_skill_cooldown = helpers["assert_skill_cooldown"]
    assert_player_hp_equals = helpers["assert_player_hp_equals"]
    assert_player_max_hp_equals = helpers["assert_player_max_hp_equals"]
    for assertion in assertions:
        if isinstance(assertion, str):
            if assertion == "equipped_rusty_sword":
                assert_equipped_sword(inventory, equipped, item_id, where)
            elif assertion == "player_damaged":
                assert_player_damaged(entities, where)
            elif assertion == "monster_dead":
                assert_monster_dead(entities, where)
            else:
                raise AssertionError(f"{where}: unknown assertion {assertion}")
            continue

        if not isinstance(assertion, dict):
            raise AssertionError(f"{where}: assertion must be string or object: {assertion}")

        typ = assertion.get("type")
        if typ == "inventory_count":
            matches = inventory
            if assertion.get("item_def_id") is not None:
                matches = [item for item in matches if str(item.get("item_def_id", "")) == str(assertion["item_def_id"])]
            if assertion.get("item_template_id") is not None:
                matches = [item for item in matches if str(item.get("item_template_id", "")) == str(assertion["item_template_id"])]
            if assertion.get("display_name") is not None:
                matches = [item for item in matches if str(item.get("display_name", "")) == str(assertion["display_name"])]
            if assertion.get("equipped") is not None:
                matches = [item for item in matches if bool(item.get("equipped")) == bool(assertion["equipped"])]
            assert_count_matches(len(matches), assertion, f"{where}: inventory count", f": {matches}")
        elif typ == "inventory_contains":
            expected_equipped = assertion.get("equipped")
            assert_inventory_contains(inventory, str(assertion["item_def_id"]), expected_equipped, where)
        elif typ == "inventory_requirement_status":
            assert_inventory_requirement_status(inventory, assertion, where)
        elif typ == "loot_requirement_status":
            assert_loot_requirement_status(entities, assertion, where)
        elif typ == "gold":
            assert_count_matches(int(gold or 0), assertion, f"{where}: gold")
        elif typ == "stash_item_count":
            assert_stash_item_count(list(stash_items or []), assertion, where, assert_count_matches)
        elif typ == "stash_gold":
            assert_count_matches(int(stash_gold or 0), assertion, f"{where}: stash gold")
        elif typ == "stash_capacity":
            assert_count_matches(int(stash_capacity or 0), assertion, f"{where}: stash capacity")
        elif typ == "resource_wallet_count":
            resource_id = str(assertion.get("resource_id", assertion.get("item_def_id", "")))
            if not resource_id:
                raise AssertionError(f"{where}: resource_wallet_count requires resource_id")
            wallet = resource_wallet or {}
            assert_count_matches(int(wallet.get(resource_id, 0)), assertion, f"{where}: resource_wallet_count {resource_id}")
        elif typ == "wall_count":
            wall_rows = list(walls or [])
            if assertion.get("source") is not None:
                source = str(assertion["source"])
                wall_rows = [wall for wall in wall_rows if str(wall.get("source", "")) == source]
            assert_count_matches(len(wall_rows), assertion, f"{where}: wall count", f": {wall_rows}")
        elif typ == "non_perimeter_wall_exists":
            wall_rows = list(walls or [])
            if not any(str(wall.get("source", "")) != "perimeter" for wall in wall_rows):
                raise AssertionError(f"{where}: no non-perimeter wall in {wall_rows}")
        elif typ == "rolled_inventory_item":
            assert_rolled_inventory_item(inventory, assertion, where)
        elif typ == "rolled_inventory_any":
            expected_equipped = assertion.get("equipped")
            assert_rolled_inventory_any(inventory, expected_equipped, where)
        elif typ == "inventory_unique_effect_coverage":
            assert_inventory_unique_effect_coverage(inventory, assertion, where)
        elif typ == "entity_count":
            assert_entity_count(entities, assertion, where)
        elif typ == "equipped_weapon_def":
            expected_def = str(assertion["item_def_id"])
            slot = str(assertion.get("slot", "main_hand"))
            weapon_id = equipped.get(slot)
            if weapon_id is None:
                raise AssertionError(f"{where}: equipped {slot} is empty, want {expected_def}")
            item = next((i for i in inventory if str(i.get("item_instance_id")) == str(weapon_id)), None)
            if item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {weapon_id} row = {item}, want {expected_def}")
        elif typ == "equipped_slot_def":
            slot = str(assertion["slot"])
            expected_def = str(assertion["item_def_id"])
            item_id = equipped.get(slot)
            if item_id is None:
                raise AssertionError(f"{where}: equipped {slot} is empty, want {expected_def}")
            item = next((i for i in inventory if str(i.get("item_instance_id")) == str(item_id)), None)
            if item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {item_id} row = {item}, want {expected_def}")
        elif typ == "equipped_slot_empty":
            slot = str(assertion["slot"])
            if equipped.get(slot) is not None:
                raise AssertionError(f"{where}: equipped {slot}={equipped.get(slot)}, want empty")
        elif typ == "hotbar_capacity":
            assert_count_matches(int(hotbar_capacity or 0), assertion, f"{where}: hotbar_capacity")
        elif typ == "hotbar_slot":
            assert_hotbar_slot(hotbar or [], int(assertion["slot_index"]), assertion.get("item_def_id"), where, inventory)
        elif typ == "inventory_capacity":
            if "rows" in assertion:
                assert_count_matches(int(inventory_rows or 0), {"equals": int(assertion["rows"])}, f"{where}: inventory_rows")
            assert_count_matches(int(inventory_capacity or 0), assertion, f"{where}: inventory_capacity")
        elif typ == "monster_dead":
            assert_monster_dead(entities, where, str(assertion["monster_def_id"]))
        elif typ == "interactable_state":
            expected_id = str(assertion["interactable_def_id"])
            expected_state = str(assertion["state"])
            matches = [
                e for e in entities
                if e.get("type") == "interactable" and e.get("interactable_def_id") == expected_id
            ]
            if len(matches) != 1 or matches[0].get("state") != expected_state:
                raise AssertionError(f"{where}: interactable {expected_id} state = {matches}, want {expected_state}")
        elif typ == "monster_killed_in_attacks":
            continue
        elif typ in {
            "monster_moved",
            "entity_moved",
            "monster_within_player_distance",
            "monster_near_spawn",
            "event_seen",
            "event_count",
            "combat_event_seen",
            "player_never_in_melee_range_of",
            "monster_damage_at_least",
            "player_hp_decreased_from_recorded",
            "shop_offer_count",
            "shop_offer_details",
            "shop_sell_appraisal_count",
            "shop_sell_appraisal_details",
            "shop_event",
            "stash_event",
        }:
            continue
        elif typ == "player_hp_equals":
            assert_player_hp_equals(entities, int(assertion["equals"]), where)
        elif typ == "player_max_hp_equals":
            assert_player_max_hp_equals(entities, int(assertion["equals"]), where)
        elif typ == "character_progression":
            assert_character_progression(character_progression or {}, assertion, where)
        elif typ == "skill_progression":
            assert_skill_progression(skill_progression or {}, assertion, where)
        elif typ == "skill_cooldown":
            assert_skill_cooldown(skill_cooldowns or [], assertion, where)
        elif typ == "current_level":
            if current_level is None:
                raise AssertionError(f"{where}: current_level unavailable")
            want = int(assertion["equals"])
            if current_level != want:
                raise AssertionError(f"{where}: current_level {current_level} != {want}")
        elif typ == "teleporter_discovered":
            if discovered_teleporters is None:
                raise AssertionError(f"{where}: discovered_teleporters unavailable")
            level = int(assertion["level"])
            want = bool(assertion.get("discovered", True))
            got = bool(discovered_teleporters.get(level, False))
            if got != want:
                raise AssertionError(f"{where}: teleporter level {level} discovered={got}, want {want}")
        elif typ == "teleporter_list_contains":
            if discovered_teleporters is None:
                raise AssertionError(f"{where}: discovered_teleporters unavailable")
            level = int(assertion["level"])
            if level not in discovered_teleporters:
                raise AssertionError(f"{where}: teleporter level {level} missing; have {sorted(discovered_teleporters)}")
        elif typ == "visited_levels_contain":
            continue
        else:
            raise AssertionError(f"{where}: unknown assertion type {typ}")


def run_runtime_assertions(assertions: list[Any], state: Any, where: str, helpers: dict[str, Any]) -> None:

    assert_count_matches = helpers["assert_count_matches"]
    event_matches = helpers["event_matches"]
    event_summary = helpers["event_summary"]
    combat_event_matches = helpers["combat_event_matches"]
    combat_event_summary = helpers["combat_event_summary"]
    find_monster = helpers["find_monster"]
    find_player = helpers["find_player"]
    entity_matches_selector = helpers["entity_matches_selector"]
    assert_inventory_contains = helpers["assert_inventory_contains"]
    assert_inventory_requirement_status = helpers["assert_inventory_requirement_status"]
    assert_loot_requirement_status = helpers["assert_loot_requirement_status"]
    assert_rolled_inventory_item = helpers["assert_rolled_inventory_item"]
    assert_rolled_inventory_any = helpers["assert_rolled_inventory_any"]
    assert_inventory_unique_effect_coverage = helpers["assert_inventory_unique_effect_coverage"]
    assert_hotbar_slot = helpers["assert_hotbar_slot"]
    assert_entity_count = helpers["assert_entity_count"]
    assert_character_progression = helpers["assert_character_progression"]
    assert_skill_progression = helpers["assert_skill_progression"]
    assert_skill_cooldown = helpers["assert_skill_cooldown"]
    for assertion in assertions:
        if not isinstance(assertion, dict):
            continue
        typ = assertion.get("type")
        if typ == "player_never_in_melee_range_of":
            monster_def_id = str(assertion["monster_def_id"])
            min_distance = state.min_player_monster_distance.get(monster_def_id)
            if min_distance is None:
                raise AssertionError(f"{where}: no distance samples for monster {monster_def_id}")
            reach = float(assertion.get("reach", 1.5))
            monster_radius = float(assertion.get("monster_radius", 0.45))
            threshold = reach + monster_radius + 0.000001
            if min_distance <= threshold:
                raise AssertionError(
                    f"{where}: player entered melee range of {monster_def_id}: min_distance={min_distance:.3f}, threshold={threshold:.3f}"
                )
            continue
        if typ == "monster_killed_in_attacks":
            monster_def_id = str(assertion["monster_def_id"])
            max_attacks = int(assertion["max_attacks"])
            count = state.accepted_attack_counts.get(monster_def_id, 0)
            if monster_def_id not in state.killed_monster_def_ids:
                raise AssertionError(f"{where}: monster {monster_def_id} was not observed killed at runtime")
            if count < 1:
                raise AssertionError(f"{where}: no accepted action_intent observed for {monster_def_id}")
            if count > max_attacks:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} killed in {count} accepted attacks, max {max_attacks}"
                )
            continue
        if typ == "event_seen":
            if not any(event_matches(event, assertion) for event in state.events):
                raise AssertionError(
                    f"{where}: event {event_summary(assertion)} not seen; "
                    f"have events={state.events[-10:]} seen_types={sorted(state.seen_events)}"
                )
            continue
        if typ == "event_count":
            matches = sum(1 for event in state.events if event_matches(event, assertion))
            assert_count_matches(matches, assertion, f"{where}: event count {event_summary(assertion)}")
            continue
        if typ == "combat_event_seen":
            if not any(combat_event_matches(event, assertion, state) for event in state.combat_events):
                raise AssertionError(
                    f"{where}: combat event {combat_event_summary(assertion)} not seen; have {state.combat_events}"
                )
            continue
        if typ == "monster_moved":
            monster_def_id = str(assertion["monster_def_id"])
            min_distance = float(assertion["min_distance"])
            initial = state.initial_monster_positions.get(monster_def_id)
            monster = find_monster(state, monster_def_id)
            if initial is None or monster is None:
                raise AssertionError(f"{where}: missing initial/current monster {monster_def_id}")
            pos = monster.get("position", {})
            dx = float(pos.get("x", 0.0)) - initial["x"]
            dy = float(pos.get("y", 0.0)) - initial["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist < min_distance:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} moved {dist:.3f} < min_distance {min_distance}"
                )
            continue
        if typ == "entity_moved":
            min_distance = float(assertion["min_distance"])
            matches = [entity for entity in state.entities.values() if entity_matches_selector(entity, assertion)]
            if not matches:
                raise AssertionError(f"{where}: no entity matched {assertion}")
            matches.sort(key=lambda entity: str(entity.get("id", "")))
            entity = matches[0]
            entity_id = str(entity.get("id", ""))
            initial = state.initial_entity_positions.get(entity_id)
            if initial is None:
                raise AssertionError(f"{where}: missing initial position for entity {entity_id}")
            pos = entity.get("position", {})
            dx = float(pos.get("x", 0.0)) - initial["x"]
            dy = float(pos.get("y", 0.0)) - initial["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist < min_distance:
                raise AssertionError(
                    f"{where}: entity {entity_id} moved {dist:.3f} < min_distance {min_distance}"
                )
            continue
        if typ == "monster_within_player_distance":
            monster_def_id = str(assertion["monster_def_id"])
            max_distance = float(assertion["max_distance"])
            player = find_player(state)
            monster = find_monster(state, monster_def_id)
            if player is None or monster is None:
                raise AssertionError(f"{where}: missing player or monster {monster_def_id}")
            ppos = player.get("position", {})
            mpos = monster.get("position", {})
            dx = float(ppos.get("x", 0.0)) - float(mpos.get("x", 0.0))
            dy = float(ppos.get("y", 0.0)) - float(mpos.get("y", 0.0))
            dist = (dx * dx + dy * dy) ** 0.5
            if dist > max_distance:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} distance {dist:.3f} > max_distance {max_distance}"
                )
            continue
        if typ == "monster_near_spawn":
            monster_def_id = str(assertion["monster_def_id"])
            max_distance_from_spawn = float(assertion["max_distance_from_spawn"])
            spawn = state.initial_monster_positions.get(monster_def_id)
            monster = find_monster(state, monster_def_id)
            if spawn is None or monster is None:
                raise AssertionError(f"{where}: missing spawn/current monster {monster_def_id}")
            mpos = monster.get("position", {})
            dx = float(mpos.get("x", 0.0)) - spawn["x"]
            dy = float(mpos.get("y", 0.0)) - spawn["y"]
            dist = (dx * dx + dy * dy) ** 0.5
            if dist > max_distance_from_spawn:
                raise AssertionError(
                    f"{where}: monster {monster_def_id} spawn distance {dist:.3f} > {max_distance_from_spawn}"
                )
            continue
        if typ == "player_hp_decreased_from_recorded":
            if state.recorded_player_hp is None:
                raise AssertionError(f"{where}: no recorded player hp")
            player = find_player(state)
            if player is None or not isinstance(player.get("hp"), int):
                raise AssertionError(f"{where}: missing current player hp: {player}")
            current_hp = int(player["hp"])
            if current_hp >= state.recorded_player_hp:
                raise AssertionError(
                    f"{where}: player hp {current_hp} did not decrease below recorded {state.recorded_player_hp}"
                )
            continue
        if typ == "current_level":
            want = int(assertion["equals"])
            if state.current_level != want:
                raise AssertionError(f"{where}: current_level {state.current_level} != {want}")
            continue
        if typ == "wall_count":
            walls = state.walls
            if assertion.get("source") is not None:
                source = str(assertion["source"])
                walls = [wall for wall in walls if str(wall.get("source", "")) == source]
            assert_count_matches(len(walls), assertion, f"{where}: wall count", f": {walls}")
            continue
        if typ == "non_perimeter_wall_exists":
            if not any(str(wall.get("source", "")) != "perimeter" for wall in state.walls):
                raise AssertionError(f"{where}: no non-perimeter wall in {state.walls}")
            continue
        if typ == "visited_levels_contain":
            want = int(assertion["level"])
            if want not in state.visited_levels:
                raise AssertionError(f"{where}: level {want} not visited; have {sorted(state.visited_levels)}")
            continue
        if typ == "teleporter_discovered":
            level = int(assertion["level"])
            want = bool(assertion.get("discovered", True))
            got = bool(state.discovered_teleporters.get(level, False))
            if got != want:
                raise AssertionError(f"{where}: teleporter level {level} discovered={got}, want {want}")
            continue
        if typ == "teleporter_list_contains":
            level = int(assertion["level"])
            if level not in state.discovered_teleporters:
                raise AssertionError(
                    f"{where}: teleporter level {level} missing; have {sorted(state.discovered_teleporters)}"
                )
            continue
        if typ == "inventory_contains":
            expected_equipped = assertion.get("equipped")
            assert_inventory_contains(state.inventory, str(assertion["item_def_id"]), expected_equipped, where)
            continue
        if typ == "inventory_requirement_status":
            assert_inventory_requirement_status(state.inventory, assertion, where)
            continue
        if typ == "loot_requirement_status":
            assert_loot_requirement_status(list(state.entities.values()), assertion, where)
            continue
        if typ == "inventory_count":
            matches = state.inventory
            if assertion.get("item_def_id") is not None:
                matches = [item for item in matches if str(item.get("item_def_id", "")) == str(assertion["item_def_id"])]
            if assertion.get("item_template_id") is not None:
                matches = [item for item in matches if str(item.get("item_template_id", "")) == str(assertion["item_template_id"])]
            if assertion.get("equipped") is not None:
                matches = [item for item in matches if bool(item.get("equipped")) == bool(assertion["equipped"])]
            assert_count_matches(len(matches), assertion, f"{where}: inventory count", f": {matches}")
            continue
        if typ == "gold":
            assert_count_matches(state.gold, assertion, f"{where}: gold")
            continue
        if handle_runtime_economy_assertion(state, assertion, where, helpers):
            continue
        if typ == "rolled_inventory_item":
            assert_rolled_inventory_item(state.inventory, assertion, where)
        if typ == "rolled_inventory_any":
            expected_equipped = assertion.get("equipped")
            assert_rolled_inventory_any(state.inventory, expected_equipped, where)
            continue
        if typ == "inventory_unique_effect_coverage":
            assert_inventory_unique_effect_coverage(state.inventory, assertion, where)
            continue
        if typ == "equipped_slot_def":
            slot = str(assertion["slot"])
            expected_def = str(assertion["item_def_id"])
            item_id = state.equipped.get(slot)
            item = next((i for i in state.inventory if str(i.get("item_instance_id")) == str(item_id)), None)
            if item_id is None or item is None or item.get("item_def_id") != expected_def:
                raise AssertionError(f"{where}: equipped {slot} {item_id} row={item}, want {expected_def}")
            continue
        if typ == "equipped_slot_empty":
            slot = str(assertion["slot"])
            if state.equipped.get(slot) is not None:
                raise AssertionError(f"{where}: equipped {slot}={state.equipped.get(slot)}, want empty")
            continue
        if typ == "hotbar_capacity":
            assert_count_matches(state.hotbar_capacity, assertion, f"{where}: hotbar_capacity")
            continue
        if typ == "hotbar_slot":
            assert_hotbar_slot(state.hotbar, int(assertion["slot_index"]), assertion.get("item_def_id"), where, state.inventory)
            continue
        if typ == "inventory_capacity":
            if "rows" in assertion:
                assert_count_matches(state.inventory_rows, {"equals": int(assertion["rows"])}, f"{where}: inventory_rows")
            assert_count_matches(state.inventory_capacity, assertion, f"{where}: inventory_capacity")
            continue
        if typ == "entity_count":
            assert_entity_count(list(state.entities.values()), assertion, where)
            continue
        if typ == "monster_damage_at_least":
            monster_def_id = str(assertion["monster_def_id"])
            got = state.max_monster_damage_by_def.get(monster_def_id, 0)
            want = int(assertion["damage"])
            if got < want:
                raise AssertionError(f"{where}: max damage to {monster_def_id} = {got}, want at least {want}")
            continue
        if typ == "character_progression":
            assert_character_progression(state.character_progression, assertion, where)
            continue
        if typ == "skill_progression":
            assert_skill_progression(state.skill_progression, assertion, where)
            continue
        if typ == "skill_cooldown":
            assert_skill_cooldown(state.skill_cooldowns, assertion, where)
            continue
        if typ == "player_max_hp_equals":
            player = find_player(state)
            if player is None:
                raise AssertionError(f"{where}: player not found")
            got = int(player.get("max_hp", -1))
            want = int(assertion["equals"])
            if got != want:
                raise AssertionError(f"{where}: player max_hp {got} != {want}")
            continue
