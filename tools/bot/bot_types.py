"""Shared dataclasses and constants for the protocol bot.

Extracted from run.py so scenario scripts and test helpers can import just
the types without pulling in the full WS driver or assertion library.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

DEFAULT_WORLD_ID = "vertical_slice"


@dataclass(frozen=True)
class Scenario:
    id: str
    world_id: str
    seed: str
    peer_count: int
    title: str
    description: str
    character_class: str
    debug_progression: dict[str, Any]
    steps: list[dict[str, Any]]
    assertions: list[Any]
    fresh_session_checks: list[dict[str, Any]]
    max_elapsed_s: float
    path: Path


@dataclass
class RuntimeState:
    world_id: str = DEFAULT_WORLD_ID
    local_player_id: str = ""
    party: list[dict[str, Any]] = field(default_factory=list)
    last_tick: int = 0
    killed: bool = False
    walls: list[dict[str, Any]] = field(default_factory=list)
    entities: dict[str, dict[str, Any]] = field(default_factory=dict)
    inventory: list[dict[str, Any]] = field(default_factory=list)
    equipped: dict[str, Any] = field(default_factory=dict)
    active_weapon_set: int = 0
    weapon_sets: list[dict[str, Any]] = field(default_factory=list)
    hotbar_capacity: int = 2
    hotbar: list[dict[str, Any]] = field(default_factory=list)
    inventory_rows: int = 3
    inventory_capacity: int = 15
    gold: int = 0
    loot_ids: list[str] = field(default_factory=list)
    item_id: str | None = None
    equipped_item_id: str | None = None
    seen_events: set[str] = field(default_factory=set)
    events: list[dict[str, Any]] = field(default_factory=list)
    combat_events: list[dict[str, Any]] = field(default_factory=list)
    pending_attack_monsters: dict[str, str] = field(default_factory=dict)
    accepted_attack_counts: dict[str, int] = field(default_factory=dict)
    killed_monster_def_ids: set[str] = field(default_factory=set)
    max_monster_damage_by_def: dict[str, int] = field(default_factory=dict)
    accepted_message_ids: set[str] = field(default_factory=set)
    rejected_message_reasons: dict[str, str] = field(default_factory=dict)
    min_player_monster_distance: dict[str, float] = field(default_factory=dict)
    initial_monster_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    initial_entity_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    recorded_player_hp: int | None = None
    current_level: int = 0
    visited_levels: set[int] = field(default_factory=lambda: {0})
    last_delta_level: int | None = None
    pending_level_load: int | None = None
    used_stair_positions: dict[str, dict[str, float]] = field(default_factory=dict)
    discovered_teleporters: dict[int, bool] = field(default_factory=dict)
    used_teleporter_positions: dict[int, dict[str, float]] = field(default_factory=dict)
    character_progression: dict[str, Any] = field(default_factory=dict)
    skill_progression: dict[str, Any] = field(default_factory=dict)
    skill_cooldowns: list[dict[str, Any]] = field(default_factory=list)
    skill_function_keys: list[str] = field(default_factory=list)
    right_click_skill_id: str = ""
    shop_offers: dict[str, dict[str, dict[str, Any]]] = field(default_factory=dict)
    shop_sell_appraisals: dict[str, dict[str, dict[str, Any]]] = field(default_factory=dict)
    shop_events: list[dict[str, Any]] = field(default_factory=list)
    last_shop_event: dict[str, Any] | None = None
    stash_items: list[dict[str, Any]] = field(default_factory=list)
    stash_gold: int = 0
    stash_capacity: int = 50
    resource_wallet: dict[str, int] = field(default_factory=dict)
    stash_events: list[dict[str, Any]] = field(default_factory=list)
    last_stash_event: dict[str, Any] | None = None
    last_gold_before_action: int | None = None
    last_gold_after_action: int | None = None
    remembered_entity_ids: dict[str, str] = field(default_factory=dict)


@dataclass
class CoopPeer:
    label: str
    token: str
    session: dict[str, Any]
    state: RuntimeState
    ws: Any
