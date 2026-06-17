from __future__ import annotations

from typing import Any

from tools.bot.bot_context import StateIngestContext
from tools.bot.bot_types import RuntimeState
from tools.bot.runtime_queries import find_player


def _require_context(ctx: StateIngestContext | None) -> StateIngestContext:
    if ctx is None:
        raise AssertionError("state ingest requires runtime context")
    return ctx


def ingest_message(m: dict[str, Any], state: RuntimeState, ctx: StateIngestContext | None = None) -> None:
    ctx = _require_context(ctx)
    if m.get("type") == "session_snapshot":
        ingest_snapshot(m["payload"], state, ctx=ctx)
        return
    if m.get("type") == "intent_accepted":
        accepted_id = str(m.get("payload", {}).get("accepted_message_id", ""))
        state.accepted_message_ids.add(accepted_id)
        monster_def_id = state.pending_attack_monsters.pop(accepted_id, None)
        if monster_def_id:
            state.accepted_attack_counts[monster_def_id] = state.accepted_attack_counts.get(monster_def_id, 0) + 1
        return
    if m.get("type") == "intent_rejected":
        rejected_id = str(m.get("payload", {}).get("rejected_message_id", ""))
        state.rejected_message_reasons[rejected_id] = str(m.get("payload", {}).get("reason", "unknown"))
        monster_def_id = state.pending_attack_monsters.pop(rejected_id, None)
        if monster_def_id:
            reason = m.get("payload", {}).get("reason", "unknown")
            if reason == "invalid_target" and monster_def_id in state.killed_monster_def_ids:
                return
            if reason == "basic_attack_on_cooldown":
                return
            raise AssertionError(f"action_intent for {monster_def_id} was rejected: {reason}")
        return
    if m.get("type") != "state_delta":
        state.last_tick = max(state.last_tick, int(m.get("tick", 0)))
        return

    p = m["payload"]
    previous_tick = state.last_tick
    next_tick = max(state.last_tick, int(m.get("tick", 0)), int(p.get("server_tick", state.last_tick)))
    decay_skill_cooldowns(state, next_tick - previous_tick)
    state.last_tick = next_tick
    delta_level = int(p.get("level", state.current_level))
    state.last_delta_level = delta_level
    state.visited_levels.add(delta_level)
    for ev in (p.get("events") or []):
        event_type = ev["event_type"]
        state.seen_events.add(event_type)
        state.events.append(dict(ev))
        if event_type in {"shop_opened", "shop_purchase", "shop_sale", "shop_reroll"}:
            shop_event = dict(ev)
            state.shop_events.append(shop_event)
            state.last_shop_event = shop_event
            shop_id = str(ev.get("shop_id", ""))
            if shop_id and "offers" in ev:
                state.shop_offers[shop_id] = {
                    str(offer["offer_id"]): dict(offer)
                    for offer in ev.get("offers", [])
                }
            if shop_id and "sell_appraisals" in ev:
                state.shop_sell_appraisals[shop_id] = {
                    str(appraisal["item_instance_id"]): dict(appraisal)
                    for appraisal in ev.get("sell_appraisals", [])
                }
        if event_type in {"stash_opened", "stash_item_deposited", "stash_item_withdrawn", "stash_gold_deposited", "stash_gold_withdrawn", "unique_chest_opened", "unique_chest_item_taken"}:
            stash_event = dict(ev)
            state.stash_events.append(stash_event)
            state.last_stash_event = stash_event
            if "stash_items" in ev:
                state.stash_items = [dict(item) for item in ev.get("stash_items", [])]
            if "stash_gold" in ev:
                state.stash_gold = int(ev.get("stash_gold", state.stash_gold))
            if "stash_capacity" in ev:
                state.stash_capacity = int(ev.get("stash_capacity", state.stash_capacity))
            if "total_gold" in ev:
                state.gold = int(ev.get("total_gold", state.gold))
        if event_type in {"bishop_service_opened", "bishop_respec"} and "total_gold" in ev:
            state.gold = int(ev.get("total_gold", state.gold))
        if event_type in {"monster_damaged", "player_damaged", "player_killed", "attack_missed", "attack_blocked"}:
            combat_event = dict(ev)
            for prefix in ("source", "target"):
                entity = state.entities.get(str(ev.get(f"{prefix}_entity_id", "")))
                if entity is not None and entity.get("monster_def_id"):
                    combat_event[f"{prefix}_monster_def_id"] = str(entity["monster_def_id"])
            state.combat_events.append(combat_event)
        if event_type == "skill_effect_started" and str(ev.get("skill_id", "")) == "poison_stab":
            target_id = str(ev.get("target_entity_id") or ev.get("entity_id") or "")
            if target_id:
                state.remembered_entity_ids["poison_stab_target"] = target_id
        if event_type == "level_changed":
            state.current_level = int(ev["to_level"])
            state.pending_level_load = state.current_level
            clear_active_level_state(state)
            state.visited_levels.add(int(ev["from_level"]))
            state.visited_levels.add(int(ev["to_level"]))
        if event_type == "monster_killed":
            state.killed = True
            entity = state.entities.get(str(ev.get("entity_id")))
            if entity is not None and entity.get("monster_def_id"):
                state.killed_monster_def_ids.add(str(entity["monster_def_id"]))
            ctx.log("monster killed at tick", p.get("server_tick"))
        if event_type == "monster_damaged":
            entity = state.entities.get(str(ev.get("entity_id")))
            if entity is not None and entity.get("monster_def_id") and isinstance(ev.get("damage"), int):
                monster_def_id = str(entity["monster_def_id"])
                state.max_monster_damage_by_def[monster_def_id] = max(
                    state.max_monster_damage_by_def.get(monster_def_id, 0),
                    int(ev["damage"]),
                )
    apply_level_entities = delta_level == state.current_level
    for c in (p.get("changes") or []):
        if c["op"] == "wall_layout_update":
            if apply_level_entities:
                state.walls = [dict(wall) for wall in c.get("walls", [])]
        elif c["op"] in {"entity_spawn", "entity_update"}:
            if not apply_level_entities:
                continue
            entity = c["entity"]
            existing = state.entities.get(entity["id"], {})
            existing.update(entity)
            state.entities[entity["id"]] = existing
            track_initial_entity_position(state, existing)
            track_initial_monster_position(state, existing)
            if c["op"] == "entity_spawn" and entity["type"] == "loot":
                loot_id = entity["id"]
                if loot_id not in state.loot_ids:
                    state.loot_ids.append(loot_id)
        elif c["op"] == "entity_remove":
            if not apply_level_entities:
                continue
            entity_id = c["entity_id"]
            state.entities.pop(entity_id, None)
            if entity_id in state.loot_ids:
                state.loot_ids.remove(entity_id)
        elif c["op"] == "inventory_add":
            upsert_inventory(state, c["item"])
            state.item_id = c["item"]["item_instance_id"]
        elif c["op"] == "inventory_update":
            upsert_inventory(state, c["item"])
        elif c["op"] == "inventory_remove":
            remove_inventory_item(state, str(c["item_instance_id"]))
        elif c["op"] == "equipped_update":
            state.equipped_item_id = c.get("item_instance_id")
            state.equipped[c["slot"]] = c.get("item_instance_id")
            if "hotbar_capacity" in c:
                state.hotbar_capacity = int(c["hotbar_capacity"])
            if "inventory_rows" in c:
                state.inventory_rows = int(c["inventory_rows"])
            if "inventory_capacity" in c:
                state.inventory_capacity = int(c["inventory_capacity"])
        elif c["op"] == "weapon_set_update":
            state.active_weapon_set = int(c.get("active_weapon_set", state.active_weapon_set))
            state.weapon_sets = [dict(row) for row in c.get("weapon_sets", state.weapon_sets)]
        elif c["op"] == "hotbar_update":
            upsert_hotbar(state, int(c["slot_index"]), c.get("item_instance_id"), c.get("item"))
            if "inventory_rows" in c:
                state.inventory_rows = int(c["inventory_rows"])
            if "inventory_capacity" in c:
                state.inventory_capacity = int(c["inventory_capacity"])
        elif c["op"] == "gold_update":
            state.gold = int(c.get("gold", state.gold))
        elif c["op"] == "stash_item_add":
            upsert_stash_item(state, c["item"])
        elif c["op"] == "stash_item_remove":
            remove_stash_item(state, str(c["stash_item_id"]))
        elif c["op"] == "stash_gold_update":
            state.stash_gold = int(c.get("stash_gold", state.stash_gold))
        elif c["op"] == "resource_wallet_update":
            resource_id = str(c.get("resource_id", ""))
            if resource_id:
                state.resource_wallet[resource_id] = max(0, int(c.get("amount", state.resource_wallet.get(resource_id, 0))))
        elif c["op"] == "teleporter_discovery_update":
            state.discovered_teleporters[int(c["level"])] = bool(c["discovered"])
        elif c["op"] == "character_progression_update":
            state.character_progression = dict(c.get("character_progression") or {})
            if "gold" in state.character_progression:
                state.gold = int(state.character_progression["gold"])
        elif c["op"] == "skill_progression_update":
            state.skill_progression = dict(c.get("skill_progression") or {})
        elif c["op"] == "skill_cooldown_update":
            state.skill_cooldowns = [dict(row) for row in c.get("skill_cooldowns", [])]
        elif c["op"] == "skill_bindings_update":
            bindings = dict(c.get("skill_bindings") or {})
            state.skill_function_keys = [str(row) for row in bindings.get("function_keys", state.skill_function_keys)]
            state.right_click_skill_id = str(bindings.get("right_click_skill_id", state.right_click_skill_id))
    if state.pending_level_load is not None and delta_level == state.pending_level_load and find_player(state) is not None:
        state.pending_level_load = None
    update_runtime_distances(state)


def ingest_snapshot(payload: dict[str, Any], state: RuntimeState, ctx: StateIngestContext | None = None) -> None:
    _require_context(ctx)
    state.last_tick = max(state.last_tick, int(payload.get("server_tick", 0)))
    state.current_level = int(payload.get("current_level", 0))
    state.local_player_id = str(payload.get("local_player_id", state.local_player_id))
    state.party = [dict(row) for row in payload.get("party", [])]
    state.visited_levels.add(state.current_level)
    state.last_delta_level = state.current_level
    state.pending_level_load = None
    state.walls = [dict(wall) for wall in payload.get("walls", [])]
    state.entities = {str(e["id"]): dict(e) for e in payload.get("entities", [])}
    state.inventory = [dict(i) for i in payload.get("inventory", [])]
    state.equipped = dict(payload.get("equipped", {}))
    state.active_weapon_set = int(payload.get("active_weapon_set", state.active_weapon_set))
    state.weapon_sets = [dict(row) for row in payload.get("weapon_sets", state.weapon_sets)]
    state.hotbar_capacity = int(payload.get("hotbar_capacity", 2))
    state.hotbar = [dict(slot) for slot in payload.get("hotbar", [])]
    state.inventory_rows = int(payload.get("inventory_rows", 3))
    state.inventory_capacity = int(payload.get("inventory_capacity", state.inventory_rows * 5))
    state.character_progression = dict(payload.get("character_progression", {}))
    state.skill_progression = dict(payload.get("skill_progression", {}))
    state.skill_cooldowns = [dict(row) for row in payload.get("skill_cooldowns", [])]
    bindings = dict(payload.get("skill_bindings", {}))
    state.skill_function_keys = [str(row) for row in bindings.get("function_keys", state.skill_function_keys)]
    state.right_click_skill_id = str(bindings.get("right_click_skill_id", state.right_click_skill_id))
    state.gold = int(payload.get("gold", state.character_progression.get("gold", 0)))
    state.stash_items = [dict(item) for item in payload.get("stash_items", [])]
    state.stash_gold = int(payload.get("stash_gold", 0))
    state.stash_capacity = int(payload.get("stash_capacity", 50))
    state.resource_wallet = {
        str(row.get("resource_id", "")): max(0, int(row.get("amount", 0)))
        for row in payload.get("resource_wallet", [])
        if str(row.get("resource_id", ""))
    }
    state.discovered_teleporters = parse_discovered_teleporters(payload)
    state.loot_ids = [
        entity_id
        for entity_id, entity in state.entities.items()
        if entity.get("type") == "loot"
    ]
    for item in state.inventory:
        if item.get("equipped"):
            state.equipped_item_id = item.get("item_instance_id")
    for entity in state.entities.values():
        track_initial_entity_position(state, entity)
        track_initial_monster_position(state, entity)
    update_runtime_distances(state)


def parse_discovered_teleporters(payload: dict[str, Any]) -> dict[int, bool]:
    return {
        int(row["level"]): bool(row["discovered"])
        for row in payload.get("discovered_teleporters", [])
    }


def upsert_hotbar(state: RuntimeState, slot_index: int, item_instance_id: Any, item: dict[str, Any] | None = None) -> None:
    while len(state.hotbar) <= slot_index:
        state.hotbar.append({"slot_index": len(state.hotbar), "item_instance_id": None})
    slot = {"slot_index": slot_index, "item_instance_id": item_instance_id}
    if item:
        slot["item"] = dict(item)
    state.hotbar[slot_index] = slot


def decay_skill_cooldowns(state: RuntimeState, ticks: int) -> None:
    if ticks <= 0 or not state.skill_cooldowns:
        return
    next_rows: list[dict[str, Any]] = []
    for row in state.skill_cooldowns:
        remaining = max(0, int(row.get("remaining_ticks", 0)) - ticks)
        if remaining <= 0:
            continue
        updated = dict(row)
        updated["remaining_ticks"] = remaining
        next_rows.append(updated)
    state.skill_cooldowns = next_rows


def clear_active_level_state(state: RuntimeState) -> None:
    state.walls.clear()
    state.entities.clear()
    state.loot_ids.clear()
    state.min_player_monster_distance.clear()
    state.initial_monster_positions.clear()
    state.initial_entity_positions.clear()


def track_initial_entity_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    entity_id = str(entity.get("id", ""))
    if not entity_id or entity_id in state.initial_entity_positions:
        return
    pos = entity.get("position")
    if not isinstance(pos, dict):
        return
    state.initial_entity_positions[entity_id] = {
        "x": float(pos.get("x", 0.0)),
        "y": float(pos.get("y", 0.0)),
    }


def track_initial_monster_position(state: RuntimeState, entity: dict[str, Any]) -> None:
    if entity.get("type") != "monster":
        return
    monster_def_id = str(entity.get("monster_def_id", ""))
    if not monster_def_id or monster_def_id in state.initial_monster_positions:
        return
    pos = entity.get("position", {})
    state.initial_monster_positions[monster_def_id] = {
        "x": float(pos.get("x", 0.0)),
        "y": float(pos.get("y", 0.0)),
    }


def update_runtime_distances(state: RuntimeState) -> None:
    player = find_player(state)
    if player is None:
        return
    ppos = player.get("position", {})
    px = float(ppos.get("x", 0.0))
    py = float(ppos.get("y", 0.0))
    for entity in state.entities.values():
        if entity.get("type") != "monster":
            continue
        if int(entity.get("hp", 0)) <= 0:
            continue
        monster_def_id = str(entity.get("monster_def_id", ""))
        if not monster_def_id:
            continue
        mpos = entity.get("position", {})
        dx = px - float(mpos.get("x", 0.0))
        dy = py - float(mpos.get("y", 0.0))
        dist = (dx * dx + dy * dy) ** 0.5
        current = state.min_player_monster_distance.get(monster_def_id)
        if current is None or dist < current:
            state.min_player_monster_distance[monster_def_id] = dist


def upsert_inventory(state: RuntimeState, item: dict[str, Any]) -> None:
    item_id = item["item_instance_id"]
    for i, current in enumerate(state.inventory):
        if current.get("item_instance_id") == item_id:
            merged = dict(current)
            merged.update(item)
            state.inventory[i] = merged
            return
    state.inventory.append(dict(item))


def remove_inventory_item(state: RuntimeState, item_instance_id: str) -> None:
    state.inventory = [
        item for item in state.inventory
        if str(item.get("item_instance_id")) != item_instance_id
    ]
    if state.item_id == item_instance_id:
        state.item_id = None
    if state.equipped_item_id == item_instance_id:
        state.equipped_item_id = None


def upsert_stash_item(state: RuntimeState, item: dict[str, Any]) -> None:
    stash_item_id = str(item["stash_item_id"])
    for i, current in enumerate(state.stash_items):
        if str(current.get("stash_item_id")) == stash_item_id:
            merged = dict(current)
            merged.update(item)
            state.stash_items[i] = merged
            return
    state.stash_items.append(dict(item))


def remove_stash_item(state: RuntimeState, stash_item_id: str) -> None:
    state.stash_items = [
        item for item in state.stash_items
        if str(item.get("stash_item_id")) != stash_item_id
    ]
