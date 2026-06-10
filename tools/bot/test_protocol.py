import json

from tools.bot.client_join_preflight import build_metadata, write_metadata
from tools.bot.protocol import make_envelope, next_message_id, to_ws_url
from tools.bot.bot_types import RuntimeState
from tools.bot.run import (
    clean_bot_run_artifacts,
    create_listed_coop_session,
    default_manifest_path,
    directional_attack_direction,
    filtered_shop_offers,
    find_inventory_item,
    find_interactable,
    find_loot,
    find_monster,
    find_player,
    ingest_message,
    join_listed_session,
    list_active_sessions,
    load_scenarios,
    range_candidate_positions,
    run_assertions,
    run_runtime_assertions,
    select_shop_offer,
    select_scenarios,
    should_clean_bot_run_artifacts,
)


def test_to_ws_url_http():
    assert to_ws_url("http://localhost:8888", "/v0/ws?session_id=s") == "ws://localhost:8888/v0/ws?session_id=s"


def test_to_ws_url_https():
    assert to_ws_url("https://example.test", "/v0/ws?session_id=s") == "wss://example.test/v0/ws?session_id=s"


def test_make_envelope_required_fields():
    env = make_envelope("move_intent", "sess_x", 7, {"direction": {"x": 1, "y": 0}, "duration_ticks": 5})
    assert env["type"] == "move_intent"
    assert env["session_id"] == "sess_x"
    assert env["tick"] == 7
    assert env["message_id"]  # non-empty
    assert env["payload"]["duration_ticks"] == 5
    assert "correlation_id" not in env


def test_message_ids_unique():
    assert next_message_id() != next_message_id()


def test_clean_bot_run_artifacts_removes_only_json_files(tmp_path):
    old_manifest = tmp_path / "old.json"
    visual_manifest = tmp_path / "old-visual.json"
    keep_note = tmp_path / "notes.txt"
    nested = tmp_path / "nested"
    nested.mkdir()
    nested_manifest = nested / "nested.json"

    old_manifest.write_text("{}\n")
    visual_manifest.write_text("{}\n")
    keep_note.write_text("keep\n")
    nested_manifest.write_text("{}\n")

    assert clean_bot_run_artifacts(tmp_path) == 2
    assert not old_manifest.exists()
    assert not visual_manifest.exists()
    assert keep_note.exists()
    assert nested_manifest.exists()


def test_should_clean_bot_run_artifacts_only_default_dir(tmp_path):
    assert should_clean_bot_run_artifacts(default_manifest_path())
    assert not should_clean_bot_run_artifacts(tmp_path / "manifest.json")


def test_client_join_preflight_metadata_is_atomic_json(tmp_path):
    metadata_path = tmp_path / "join-preflight.json"
    metadata = build_metadata(
        session={
            "session_id": "sess_join",
            "world_id": "dungeon_levels",
            "mode": "coop",
            "listed": True,
        },
        host_email="host@example.test",
        host_account_id="acct_host",
        host_character_id="char_host",
        host_player_id="1001",
    )

    write_metadata(metadata_path, metadata)
    loaded = json.loads(metadata_path.read_text())

    assert loaded == {
        "ready": True,
        "session_id": "sess_join",
        "world_id": "dungeon_levels",
        "mode": "coop",
        "listed": True,
        "host_email": "host@example.test",
        "host_account_id": "acct_host",
        "host_character_id": "char_host",
        "host_player_id": "1001",
    }


def test_load_scenarios_discovers_vertical_slice():
    scenarios = load_scenarios()
    ids = {scenario.id for scenario in scenarios}

    assert "vertical_slice" in ids
    assert "town_vendor_gold_sink" in ids
    vertical = next(s for s in scenarios if s.id == "vertical_slice")
    assert vertical.world_id == "vertical_slice"


def test_load_scenarios_catalog_order():
    scenarios = load_scenarios()

    assert [s.path.name for s in scenarios[:2]] == ["01_vertical_slice.json", "02_gear_before_combat.json"]
    assert [s.id for s in scenarios[:2]] == ["vertical_slice", "gear_before_combat"]


def test_select_shop_offer_prefers_cheapest_affordable_generated():
    state = RuntimeState(gold=70)
    state.shop_offers = {
        "town_vendor": {
            "fixed:red_potion": {
                "offer_id": "fixed:red_potion",
                "kind": "fixed",
                "item_def_id": "red_potion",
                "buy_price": 20,
            },
            "generated:depth3:000": {
                "offer_id": "generated:depth3:000",
                "kind": "generated",
                "item_def_id": "cave_blade",
                "item_template_id": "cave_blade",
                "buy_price": 120,
            },
            "generated:depth3:001": {
                "offer_id": "generated:depth3:001",
                "kind": "generated",
                "item_def_id": "cave_boots",
                "item_template_id": "cave_boots",
                "buy_price": 40,
            },
        }
    }

    offer = select_shop_offer(state, {"action": "buy_shop_offer", "offer_kind": "generated", "affordable": True})

    assert offer["offer_id"] == "generated:depth3:001"
    assert [o["offer_id"] for o in filtered_shop_offers(state, {"shop_id": "town_vendor", "offer_kind": "generated"})] == [
        "generated:depth3:001",
        "generated:depth3:000",
    ]


def test_runtime_assertions_support_shop_offer_count_and_events():
    state = RuntimeState(gold=50)
    state.shop_offers = {
        "town_vendor": {
            "fixed:red_potion": {"offer_id": "fixed:red_potion", "kind": "fixed", "item_def_id": "red_potion", "buy_price": 20},
            "fixed:blue_potion": {"offer_id": "fixed:blue_potion", "kind": "fixed", "item_def_id": "blue_potion", "buy_price": 20},
            "generated:depth3:000": {"offer_id": "generated:depth3:000", "kind": "generated", "item_def_id": "cave_boots", "buy_price": 40, "source_depth": 3},
            "buyback:2001": {"offer_id": "buyback:2001", "kind": "buyback", "item_def_id": "cave_boots", "buy_price": 12},
        }
    }
    state.shop_events = [
        {"event_type": "shop_opened", "shop_id": "town_vendor"},
        {"event_type": "shop_purchase", "shop_id": "town_vendor"},
    ]

    run_runtime_assertions(
        [
            {"type": "shop_offer_count", "shop_id": "town_vendor", "equals": 4},
            {"type": "shop_offer_count", "shop_id": "town_vendor", "offer_kind": "fixed", "equals": 2},
            {"type": "shop_offer_count", "shop_id": "town_vendor", "offer_kind": "buyback", "equals": 1},
            {"type": "shop_offer_details", "shop_id": "town_vendor", "offer_kind": "generated", "equals": 1, "source_depth_min": 2, "source_depth_max": 3, "requires_summary": False},
            {"type": "shop_event", "shop_id": "town_vendor", "event_type": "shop_purchase", "equals": 1},
        ],
        state,
        "unit",
    )


def test_runtime_assertions_support_concealed_mystery_offers_and_reveal_event():
    state = RuntimeState(gold=250)
    state.shop_offers = {
        "town_mystery_seller": {
            "mystery:main_hand:000": {
                "offer_id": "mystery:main_hand:000",
                "kind": "mystery",
                "concealed": True,
                "mystery_label": "Unidentified main hand",
                "slot": "main_hand",
                "category": "weapon",
                "buy_price": 120,
                "source_depth_min": 1,
                "source_depth_max": 3,
            }
        }
    }
    state.shop_events = [
        {
            "event_type": "shop_purchase",
            "shop_id": "town_mystery_seller",
            "offer_id": "mystery:main_hand:000",
            "item": {
                "item_instance_id": "itm_1",
                "item_def_id": "cave_blade",
                "item_template_id": "cave_blade",
                "display_name": "Bright Cave Blade",
                "rarity": "magic",
            },
        }
    ]

    run_runtime_assertions(
        [
            {
                "type": "shop_offer_details",
                "shop_id": "town_mystery_seller",
                "offer_kind": "mystery",
                "concealed": True,
                "equals": 1,
                "source_depth_min": 1,
                "source_depth_max": 3,
                "requires_summary": False,
                "requires_slot": True,
                "requires_category": True,
                "requires_concealed": True,
                "requires_mystery_label": True,
                "forbids_item_identity": True,
            },
            {
                "type": "shop_event",
                "shop_id": "town_mystery_seller",
                "event_type": "shop_purchase",
                "offer_kind": "mystery",
                "requires_revealed_item": True,
                "rarity_in": ["magic", "rare"],
                "equals": 1,
            },
        ],
        state,
        "unit",
    )


def test_shop_purchase_event_refreshes_cached_offers():
    state = RuntimeState(gold=100)
    ingest_message(
        {
            "type": "state_delta",
            "tick": 1,
            "payload": {
                "server_tick": 1,
                "level": 0,
                "changes": [],
                "events": [
                    {
                        "event_type": "shop_opened",
                        "shop_id": "town_vendor",
                        "offers": [
                            {"offer_id": "fixed:red_potion", "kind": "fixed", "buy_price": 20},
                            {"offer_id": "generated:wp:none:000", "kind": "generated", "buy_price": 50, "source_depth": 1},
                        ],
                        "sell_appraisals": [],
                    }
                ],
            },
        },
        state,
    )
    assert [o["offer_id"] for o in filtered_shop_offers(state, {"shop_id": "town_vendor"})] == [
        "fixed:red_potion",
        "generated:wp:none:000",
    ]

    ingest_message(
        {
            "type": "state_delta",
            "tick": 2,
            "payload": {
                "server_tick": 2,
                "level": 0,
                "changes": [],
                "events": [
                    {
                        "event_type": "shop_purchase",
                        "shop_id": "town_vendor",
                        "offers": [
                            {"offer_id": "fixed:red_potion", "kind": "fixed", "buy_price": 20},
                        ],
                        "sell_appraisals": [],
                    }
                ],
            },
        },
        state,
    )

    assert [o["offer_id"] for o in filtered_shop_offers(state, {"shop_id": "town_vendor"})] == ["fixed:red_potion"]
    assert filtered_shop_offers(state, {"shop_id": "town_vendor", "offer_id": "generated:wp:none:000"}) == []


def test_load_scenarios_gear_asserts_outcome_not_timing():
    scenarios = load_scenarios()
    gear = next(s for s in scenarios if s.id == "gear_before_combat")

    assert {
        "type": "monster_dead",
        "monster_def_id": "training_dummy_reward",
    } in gear.assertions


def test_load_scenarios_discovers_door_lab():
    scenarios = load_scenarios()
    door = next(s for s in scenarios if s.id == "door_lab")

    assert door.world_id == "door_lab"
    assert not any("expect_reject" in step for step in door.steps)
    assert {
        "type": "interactable_state",
        "interactable_def_id": "wooden_door",
        "state": "open",
    } in door.assertions


def test_load_scenarios_discovers_path_maze():
    scenarios = load_scenarios()
    maze = next(s for s in scenarios if s.id == "path_maze")

    assert maze.world_id == "path_maze"
    assert maze.steps == [{
        "action": "action_once_until_event",
        "monster_def_id": "training_dummy_path",
        "event_type": "monster_killed",
    }]


def test_load_scenarios_dungeon_levels_loots_coin_before_returning():
    scenarios = load_scenarios()
    dungeon = next(s for s in scenarios if s.id == "dungeon_levels")

    assert dungeon.world_id == "dungeon_levels"
    assert {"action": "pick_up_loot", "item_def_id": "gold"} in dungeon.steps
    assert dungeon.steps[-1] == {
        "action": "assert_player_at_used_stair",
        "direction": "down",
        "tolerance": 0.001,
    }
    assert {"type": "gold", "at_least": 8} in dungeon.assertions


def test_select_scenarios_all_returns_catalog_order():
    scenarios = load_scenarios()

    assert select_scenarios(scenarios, "all") == scenarios


def test_select_scenarios_accepts_file_name_and_stem():
    scenarios = load_scenarios()

    by_file = select_scenarios(scenarios, "06_ranged_lab.json")
    assert [s.id for s in by_file] == ["ranged_lab"]

    by_stem = select_scenarios(scenarios, "06_ranged_lab")
    assert [s.id for s in by_stem] == ["ranged_lab"]


def test_load_scenarios_discovers_inventory_lab():
    scenarios = load_scenarios()
    inventory = next(s for s in scenarios if s.id == "inventory_lab")
    raw = json.loads(inventory.path.read_text())

    assert inventory.world_id == "inventory_lab"
    assert {"action": "unequip_slot", "slot": "main_hand"} in inventory.steps
    assert {"action": "drop_inventory_item", "item_def_id": "rusty_sword"} in inventory.steps
    assert {
        "type": "equipped_weapon_def",
        "item_def_id": "rusty_sword",
    } in inventory.assertions
    assert raw.get("visual", {}).get("inventory_panel") is True


def test_load_scenarios_discovers_inventory_capacity_lab():
    scenarios = load_scenarios()
    capacity = next(s for s in scenarios if s.id == "inventory_capacity_and_paper_doll")

    assert capacity.world_id == "inventory_capacity_lab"
    assert {"type": "inventory_capacity", "rows": 4, "equals": 20} in capacity.assertions
    assert any(step.get("expect_reject") == "inventory_full" for step in capacity.steps)


def test_load_scenarios_discovers_equipment_requirements_and_preview():
    scenarios = load_scenarios()
    requirements = next(s for s in scenarios if s.id == "equipment_requirements_and_preview")

    assert requirements.world_id == "requirements_lab"
    assert any(step.get("expect_reject") == "requirements_not_met" for step in requirements.steps)
    assert any(assertion.get("type") == "inventory_requirement_status" for assertion in requirements.assertions)


def test_load_scenarios_discovers_combat_control_and_boss_ai_fixes():
    scenarios = load_scenarios()
    combat = next(s for s in scenarios if s.id == "combat_control_and_boss_ai_fixes")

    assert combat.world_id == "combat_control_lab"
    assert any(step.get("action") == "directional_attack" for step in combat.steps)
    assert {"type": "event_seen", "event_type": "monster_aggro"} in combat.assertions
    assert {
        "type": "entity_moved",
        "entity_type": "monster",
        "monster_def_id": "dungeon_mob",
        "min_distance": 0.5,
    } in combat.assertions


def test_load_scenarios_discovers_session_browser_uncapped_coop():
    scenarios = load_scenarios()
    coop = next(s for s in scenarios if s.id == "session_browser_uncapped_coop")

    assert coop.world_id == "dungeon_levels"
    assert coop.peer_count == 3
    assert coop.steps == []
    assert coop.assertions == []


def test_load_scenarios_discovers_coop_rewards_and_scaling():
    scenarios = load_scenarios()
    coop = next(s for s in scenarios if s.id == "coop_rewards_and_scaling")

    assert coop.world_id == "dungeon_levels"
    assert coop.peer_count == 3
    assert coop.steps == []
    assert coop.assertions == []


def test_listed_session_http_helpers():
    requests: list[tuple[str, str, dict | None]] = []

    def handler(request):
        payload = json.loads(request.content.decode() or "{}") if request.content else None
        requests.append((request.method, request.url.path, payload))
        if request.method == "POST" and request.url.path == "/v0/sessions":
            assert payload == {
                "mode": "coop",
                "listed": True,
                "world_id": "dungeon_levels",
                "character_id": "char_host",
                "seed": "seed-1",
            }
            return httpx.Response(201, json={
                "session_id": "sess_1",
                "character_id": "char_host",
                "seed": "seed-1",
                "world_id": "dungeon_levels",
                "mode": "coop",
                "listed": True,
                "join_code": "join_secret",
                "ws_url": "/v0/ws?session_id=sess_1",
            })
        if request.method == "GET" and request.url.path == "/v0/sessions/active":
            return httpx.Response(200, json={"sessions": [{
                "session_id": "sess_1",
                "world_id": "dungeon_levels",
                "mode": "coop",
                "listed": True,
                "host_character_id": "char_host",
                "host_display_name": "Host",
                "member_count": 1,
                "connected_count": 0,
                "created_at": "2026-06-08T00:00:00Z",
                "updated_at": "2026-06-08T00:00:00Z",
            }]})
        if request.method == "POST" and request.url.path == "/v0/sessions/sess_1/join":
            assert payload == {"character_id": "char_guest"}
            return httpx.Response(200, json={
                "session_id": "sess_1",
                "character_id": "char_guest",
                "seed": "seed-1",
                "world_id": "dungeon_levels",
                "mode": "coop",
                "listed": True,
                "ws_url": "/v0/ws?session_id=sess_1",
            })
        return httpx.Response(404)

    import httpx
    transport = httpx.MockTransport(handler)
    with httpx.Client(base_url="http://testserver", transport=transport) as client:
        created = create_listed_coop_session(client, "token", "dungeon_levels", "char_host", "seed-1")
        active = list_active_sessions(client, "token")
        joined = join_listed_session(client, "token", "sess_1", "char_guest")

    assert created["listed"] is True
    assert active[0]["session_id"] == "sess_1"
    assert joined["character_id"] == "char_guest"
    assert all(req[2] is None or "join_code" not in req[2] for req in requests)


def test_select_scenarios_rejects_unknown_id():
    scenarios = load_scenarios()

    try:
        select_scenarios(scenarios, "missing")
    except ValueError as exc:
        assert "unknown scenario" in str(exc)
    else:
        raise AssertionError("expected ValueError")


def test_load_scenarios_rejects_unknown_world(tmp_path):
    (tmp_path / "01_bad.json").write_text(json.dumps({
        "id": "bad",
        "world_id": "missing",
        "steps": [],
        "assertions": [],
    }))

    try:
        load_scenarios(tmp_path)
    except ValueError as exc:
        assert "unknown world_id" in str(exc)
    else:
        raise AssertionError("expected ValueError")


def test_runtime_state_selectors_from_snapshot_and_delta():
    state = RuntimeState()
    ingest_message({
        "type": "session_snapshot",
        "tick": 0,
        "payload": {
            "server_tick": 0,
            "entities": [
                {"id": "1001", "type": "player", "position": {"x": 0, "y": 5}, "hp": 10, "max_hp": 10},
                {"id": "1002", "type": "loot", "item_def_id": "rusty_sword", "position": {"x": 6, "y": 5}},
                {"id": "1003", "type": "monster", "monster_def_id": "training_dummy_reward", "rarity": "champion", "position": {"x": 12, "y": 5}, "hp": 3, "max_hp": 3},
                {"id": "1004", "type": "interactable", "interactable_def_id": "wooden_door", "state": "closed", "position": {"x": 4, "y": 5}},
            ],
            "inventory": [],
            "equipped": {"main_hand": None},
            "hotbar_capacity": 2,
            "hotbar": [{"slot_index": i, "item_instance_id": None} for i in range(10)],
            "inventory_rows": 3,
            "inventory_capacity": 15,
        },
    }, state)

    assert find_player(state)["id"] == "1001"
    assert find_loot(state, "rusty_sword")["id"] == "1002"
    assert find_monster(state, "training_dummy_reward")["id"] == "1003"
    assert find_monster(state, "training_dummy_reward", "champion")["id"] == "1003"
    assert find_monster(state, "training_dummy_reward", "rare") is None
    assert find_interactable(state, "wooden_door")["id"] == "1004"

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "changes": [
                {"op": "entity_remove", "entity_id": "1002"},
                {"op": "inventory_add", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": False}},
                {"op": "inventory_update", "item": {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": True}},
                {"op": "equipped_update", "slot": "main_hand", "item_instance_id": "1004", "inventory_rows": 4, "inventory_capacity": 20},
                {"op": "inventory_remove", "item_instance_id": "1004"},
            ],
            "events": [],
        },
    }, state)

    assert find_loot(state, "rusty_sword") is None
    assert find_inventory_item(state.inventory, "rusty_sword") is None
    assert state.equipped["main_hand"] == "1004"
    assert state.inventory_rows == 4
    assert state.inventory_capacity == 20


def test_runtime_state_waits_for_destination_level_delta():
    state = RuntimeState()
    ingest_message({
        "type": "session_snapshot",
        "tick": 0,
        "payload": {
            "server_tick": 0,
            "current_level": -1,
            "walls": [{"id": "wall_-1_0000", "position": {"x": 1, "y": 1}, "size": {"x": 2, "y": 1}, "source": "perimeter"}],
            "entities": [
                {"id": "1001", "type": "player", "position": {"x": 14, "y": 18}, "hp": 10, "max_hp": 10},
                {"id": "1002", "type": "interactable", "interactable_def_id": "stairs_down", "state": "ready", "position": {"x": 14, "y": 18}},
            ],
            "inventory": [],
            "equipped": {"main_hand": None},
            "hotbar_capacity": 2,
            "hotbar": [{"slot_index": i, "item_instance_id": None} for i in range(10)],
        },
    }, state)

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "level": -1,
            "changes": [{"op": "entity_remove", "entity_id": "1001"}],
            "events": [{"event_type": "level_changed", "from_level": -1, "to_level": -2}],
        },
    }, state)

    assert state.current_level == -2
    assert state.pending_level_load == -2
    assert state.walls == []
    assert find_interactable(state, "stairs_down") is None
    assert find_interactable(state, "stairs_up") is None

    ingest_message({
        "type": "state_delta",
        "tick": 1,
        "payload": {
            "server_tick": 1,
            "level": -2,
            "changes": [
                {
                    "op": "wall_layout_update",
                    "walls": [
                        {"id": "wall_-2_0000", "position": {"x": 2, "y": 2}, "size": {"x": 2, "y": 1}, "source": "perimeter"},
                        {"id": "wall_-2_0004", "position": {"x": 6, "y": 6}, "size": {"x": 4, "y": 1}, "source": "generated"},
                    ],
                },
                {"op": "entity_spawn", "entity": {"id": "1001", "type": "player", "position": {"x": 9, "y": 11}, "hp": 10, "max_hp": 10}},
                {"op": "entity_spawn", "entity": {"id": "1003", "type": "interactable", "interactable_def_id": "stairs_up", "state": "ready", "position": {"x": 9, "y": 11}}},
                {"op": "entity_spawn", "entity": {"id": "1004", "type": "interactable", "interactable_def_id": "stairs_down", "state": "ready", "position": {"x": 28, "y": 14}}},
            ],
            "events": [],
        },
    }, state)

    assert state.pending_level_load is None
    assert len(state.walls) == 2
    assert state.walls[1]["source"] == "generated"
    assert find_player(state)["id"] == "1001"
    assert find_interactable(state, "stairs_up")["id"] == "1003"


def test_intent_accepted_increments_pending_attack_count():
    state = RuntimeState()
    state.pending_attack_monsters["msg-attack"] = "training_dummy_reward"

    ingest_message({
        "type": "intent_accepted",
        "tick": 4,
        "payload": {"accepted_message_id": "msg-attack", "server_tick": 4},
    }, state)

    assert state.accepted_attack_counts["training_dummy_reward"] == 1
    assert "msg-attack" not in state.pending_attack_monsters


def test_runtime_state_records_combat_event_metadata():
    state = RuntimeState()

    ingest_message({
        "type": "state_delta",
        "tick": 7,
        "payload": {
            "server_tick": 7,
            "events": [{
                "event_type": "monster_damaged",
                "entity_id": "1003",
                "source_entity_id": "1001",
                "target_entity_id": "1003",
                "damage": 0,
                "outcome": "block",
                "raw_damage": 0,
                "mitigated_damage": 0,
                "blocked": True,
                "critical": False,
            }],
            "changes": [],
        },
    }, state)

    assert state.combat_events == [{
        "event_type": "monster_damaged",
        "entity_id": "1003",
        "source_entity_id": "1001",
        "target_entity_id": "1003",
        "damage": 0,
        "outcome": "block",
        "raw_damage": 0,
        "mitigated_damage": 0,
        "blocked": True,
        "critical": False,
    }]
    run_runtime_assertions([
        {"type": "combat_event_seen", "event_type": "monster_damaged", "outcome": "block", "blocked": True, "damage": 0}
    ], state, "test")


def test_runtime_assertion_matches_non_combat_event_payloads():
    state = RuntimeState()

    ingest_message({
        "type": "state_delta",
        "tick": 12,
        "payload": {
            "server_tick": 12,
            "events": [{
                "event_type": "boss_phase_started",
                "entity_id": "3001",
                "pattern_id": "ground_slam",
                "phase_index": 0,
                "phase_kind": "telegraph",
                "duration_ticks": 35,
            }],
            "changes": [],
        },
    }, state)

    assert state.events == [{
        "event_type": "boss_phase_started",
        "entity_id": "3001",
        "pattern_id": "ground_slam",
        "phase_index": 0,
        "phase_kind": "telegraph",
        "duration_ticks": 35,
    }]
    run_runtime_assertions([
        {
            "type": "event_seen",
            "event_type": "boss_phase_started",
            "pattern_id": "ground_slam",
            "phase_kind": "telegraph",
            "phase_index": 0,
        }
    ], state, "test")

    try:
        run_runtime_assertions([
            {"type": "event_seen", "event_type": "boss_phase_started", "pattern_id": "charged_melee"}
        ], state, "test")
    except AssertionError as exc:
        assert "pattern_id=charged_melee" in str(exc)
    else:
        raise AssertionError("expected event payload mismatch")


def test_runtime_assertion_matches_combat_event_entity_selectors():
    state = RuntimeState(local_player_id="1001")
    state.entities = {
        "1001": {"id": "1001", "type": "player", "hp": 9},
        "1002": {
            "id": "1002",
            "type": "monster",
            "monster_def_id": "dungeon_archer",
            "hp": 4,
        },
        "1003": {
            "id": "1003",
            "type": "monster",
            "monster_def_id": "dungeon_mob",
            "hp": 4,
        },
    }
    state.combat_events = [
        {
            "event_type": "player_damaged",
            "source_entity_id": "1002",
            "target_entity_id": "1001",
            "damage": 1,
        }
    ]

    run_runtime_assertions([
        {
            "type": "combat_event_seen",
            "event_type": "player_damaged",
            "source_monster_def_id": "dungeon_archer",
            "target_entity_type": "player",
            "min_damage": 1,
        }
    ], state, "test")

    try:
        run_runtime_assertions([
            {
                "type": "combat_event_seen",
                "event_type": "player_damaged",
                "source_monster_def_id": "dungeon_mob",
            }
        ], state, "test")
    except AssertionError as exc:
        assert "source_monster_def_id=dungeon_mob" in str(exc)
    else:
        raise AssertionError("expected selector mismatch")


def test_range_candidate_positions_prefers_player_side_offset():
    candidates = range_candidate_positions({"x": 4, "y": 5}, {"x": 10, "y": 5}, 6.5)

    assert candidates[0] == {"x": 4.475, "y": 5.0}
    assert len(candidates) == 8


def test_directional_attack_direction_supports_explicit_and_target_direction():
    state = RuntimeState(local_player_id="1001")
    state.entities = {
        "1001": {"id": "1001", "type": "player", "position": {"x": 2, "y": 5}, "hp": 10},
        "1002": {
            "id": "1002",
            "type": "monster",
            "monster_def_id": "dungeon_mob",
            "position": {"x": 5, "y": 9},
            "hp": 4,
        },
    }

    assert directional_attack_direction(state, {"direction": {"x": 0, "y": -1}}) == {"x": 0.0, "y": -1.0}

    direction = directional_attack_direction(state, {"monster_def_id": "dungeon_mob"})
    assert round(direction["x"], 3) == 0.6
    assert round(direction["y"], 3) == 0.8


def test_runtime_assertion_entity_moved_and_player_hp_drop():
    state = RuntimeState(local_player_id="1001", recorded_player_hp=10)
    state.entities = {
        "1001": {"id": "1001", "type": "player", "position": {"x": 2, "y": 5}, "hp": 6},
        "1002": {
            "id": "1002",
            "type": "monster",
            "monster_def_id": "dungeon_mob",
            "position": {"x": 12.3, "y": 5},
            "hp": 4,
        },
    }
    state.initial_entity_positions["1002"] = {"x": 13, "y": 5}

    run_runtime_assertions([
        {
            "type": "entity_moved",
            "entity_type": "monster",
            "monster_def_id": "dungeon_mob",
            "min_distance": 0.5,
        },
        {"type": "player_hp_decreased_from_recorded"},
    ], state, "test")


def test_runtime_assertion_monster_killed_in_attacks_passes():
    state = RuntimeState(
        accepted_attack_counts={"training_dummy_reward": 1},
        killed_monster_def_ids={"training_dummy_reward"},
        inventory_rows=4,
        inventory_capacity=20,
    )

    run_runtime_assertions([
        {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1},
        {"type": "inventory_capacity", "rows": 4, "equals": 20},
    ], state, "test")


def test_runtime_assertion_monster_killed_in_attacks_fails_over_max():
    state = RuntimeState(
        accepted_attack_counts={"training_dummy_reward": 2},
        killed_monster_def_ids={"training_dummy_reward"},
    )

    try:
        run_runtime_assertions([
            {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1}
        ], state, "test")
    except AssertionError as exc:
        assert "killed in 2 accepted attacks, max 1" in str(exc)
    else:
        raise AssertionError("expected AssertionError")


def test_structured_assertions():
    entities = [
        {"id": "1001", "type": "player", "hp": 9},
        {"id": "1003", "type": "monster", "monster_def_id": "training_dummy_reward", "rarity": "champion", "hp": 0},
        {"id": "1007", "type": "interactable", "interactable_def_id": "wooden_door", "state": "open"},
    ]
    inventory = [
        {"item_instance_id": "1004", "item_def_id": "rusty_sword", "slot": "main_hand", "equipped": True},
        {"item_instance_id": "1006", "item_def_id": "quest_leaf", "slot": "", "equipped": False},
    ]

    run_assertions([
        {"type": "inventory_count", "equals": 2},
        {"type": "inventory_contains", "item_def_id": "rusty_sword", "equipped": True},
        {"type": "inventory_contains", "item_def_id": "quest_leaf", "equipped": False},
        {"type": "monster_dead", "monster_def_id": "training_dummy_reward"},
        {"type": "entity_count", "entity_type": "monster", "monster_def_id": "training_dummy_reward", "rarity": "champion", "equals": 1},
        {"type": "monster_killed_in_attacks", "monster_def_id": "training_dummy_reward", "max_attacks": 1},
        {"type": "interactable_state", "interactable_def_id": "wooden_door", "state": "open"},
        {"type": "equipped_weapon_def", "item_def_id": "rusty_sword"},
        {"type": "inventory_capacity", "rows": 3, "equals": 15},
    ], entities, inventory, {"main_hand": "1004"}, None, "test", inventory_rows=3, inventory_capacity=15)


def test_structured_requirement_and_shop_preview_assertions():
    requirement_payload = {
        "item_instance_id": "2004",
        "item_def_id": "cave_war_sword",
        "slot": "main_hand",
        "equipped": False,
        "requirements_met": False,
        "requirement_status": [
            {"stat": "level", "required": 2, "current": 1, "met": False},
            {"stat": "str", "required": 6, "current": 5, "met": False},
        ],
        "equip_preview": {
            "slot": "main_hand",
            "requirements_met": False,
            "deltas": [{"stat": "damage_max", "current": 4, "preview": 8, "delta": 4}],
        },
    }

    run_assertions([
        {
            "type": "inventory_requirement_status",
            "item_def_id": "cave_war_sword",
            "equipped": False,
            "requirements_met": False,
            "preview_slot": "main_hand",
            "preview_requirements_met": False,
            "preview_delta_stats": ["damage_max"],
            "status": [
                {"stat": "level", "required": 2, "current": 1, "met": False},
                {"stat": "str", "required": 6, "current": 5, "met": False},
            ],
        }
    ], [], [requirement_payload], {}, None, "test")

    state = RuntimeState(gold=100)
    state.inventory = [requirement_payload]
    state.shop_offers = {"town_vendor": {"generated:depth1:000": dict(requirement_payload, offer_id="generated:depth1:000", kind="generated", buy_price=50, summary_lines=["Requires STR 6"], category="equipment")}}
    state.shop_sell_appraisals = {"town_vendor": {"2004": dict(requirement_payload, sell_price=12, summary_lines=["Requires STR 6"], category="equipment")}}

    run_runtime_assertions([
        {
            "type": "inventory_requirement_status",
            "item_def_id": "cave_war_sword",
            "requirements_met": False,
            "requires_equip_preview": True,
        },
        {
            "type": "shop_offer_details",
            "shop_id": "town_vendor",
            "equals": 1,
            "requires_requirement_status": True,
            "requires_equip_preview": True,
        },
        {
            "type": "shop_sell_appraisal_details",
            "shop_id": "town_vendor",
            "equals": 1,
            "requires_requirement_status": True,
            "requires_equip_preview": True,
        },
    ], state, "test")


def test_stash_snapshot_delta_and_assertions():
    state = RuntimeState()
    ingest_message(
        {
            "type": "session_snapshot",
            "payload": {
                "server_tick": 1,
                "current_level": 0,
                "local_player_id": "1001",
                "party": [],
                "walls": [],
                "entities": [],
                "inventory": [],
                "equipped": {},
                "hotbar": [],
                "stash_items": [
                    {
                        "stash_item_id": "9001",
                        "item_def_id": "cave_bow",
                        "item_template_id": "cave_bow",
                        "display_name": "Common Cave Bow",
                    }
                ],
                "stash_gold": 3,
                "stash_capacity": 50,
                "discovered_teleporters": [],
                "character_progression": {"gold": 11},
                "skill_progression": {},
                "skill_cooldowns": [],
            },
        },
        state,
    )

    run_runtime_assertions([
        {"type": "stash_item_count", "item_template_id": "cave_bow", "equals": 1},
        {"type": "stash_gold", "equals": 3},
        {"type": "stash_capacity", "equals": 50},
    ], state, "unit runtime")
    run_assertions([
        {"type": "stash_item_count", "item_template_id": "cave_bow", "equals": 1},
        {"type": "stash_gold", "equals": 3},
        {"type": "stash_capacity", "equals": 50},
    ], [], [], {}, None, "unit state", stash_items=list(state.stash_items), stash_gold=state.stash_gold, stash_capacity=state.stash_capacity)

    ingest_message(
        {
            "type": "state_delta",
            "tick": 2,
            "payload": {
                "server_tick": 2,
                "level": 0,
                "events": [
                    {
                        "event_type": "stash_gold_deposited",
                        "entity_id": "1002",
                        "stash_id": "account_stash",
                        "amount": 4,
                        "total_gold": 7,
                        "stash_gold": 7,
                    }
                ],
                "changes": [
                    {
                        "op": "stash_item_add",
                        "item": {
                            "stash_item_id": "9002",
                            "item_def_id": "cave_blade",
                            "item_template_id": "cave_blade",
                        },
                    },
                    {"op": "stash_item_remove", "stash_item_id": "9001"},
                    {"op": "stash_gold_update", "stash_gold": 7},
                ],
            },
        },
        state,
    )

    run_runtime_assertions([
        {"type": "stash_item_count", "item_template_id": "cave_bow", "equals": 0},
        {"type": "stash_item_count", "item_template_id": "cave_blade", "equals": 1},
        {"type": "stash_gold", "equals": 7},
        {"type": "gold", "equals": 7},
        {"type": "stash_event", "event_type": "stash_gold_deposited", "equals": 1},
    ], state, "unit runtime")


def test_structured_assertions_support_range_comparators_and_filters():
    entities = [
        {"id": "1001", "type": "player", "hp": 9},
        {"id": "1003", "type": "monster", "monster_def_id": "dungeon_mob", "rarity": "champion", "hp": 4, "level": -1},
        {"id": "1004", "type": "monster", "monster_def_id": "dungeon_mob", "rarity": "common", "hp": 3, "level": -1},
        {"id": "1005", "type": "monster", "monster_def_id": "dungeon_mob", "rarity": "common", "hp": 0, "level": -1},
        {
            "id": "1008",
            "type": "monster",
            "monster_def_id": "dungeon_mob",
            "rarity": "unique",
            "hp": 32,
            "is_boss": True,
            "boss_template_id": "cave_warden",
            "visual_model": "current_humanoid_player",
            "visual_scale": 2.0,
        },
        {"id": "1006", "type": "interactable", "interactable_def_id": "treasure_chest", "state": "open"},
    ]
    inventory = [
        {"item_instance_id": "2001", "item_def_id": "red_potion", "equipped": False},
        {"item_instance_id": "2002", "item_def_id": "red_potion", "equipped": False},
        {"item_instance_id": "2003", "item_def_id": "cave_blade", "item_template_id": "sword_t1", "equipped": True},
    ]

    run_assertions([
        {"type": "entity_count", "entity_type": "monster", "monster_def_id": "dungeon_mob", "between": [2, 4]},
        {"type": "entity_count", "entity_type": "monster", "rarity": "champion", "level": -1, "equals": 1},
        {"type": "entity_count", "entity_type": "monster", "is_boss": True, "boss_template_id": "cave_warden", "visual_model": "current_humanoid_player", "visual_scale": 2.0, "equals": 1},
        {"type": "entity_count", "entity_type": "monster", "alive": True, "at_most": 3},
        {"type": "entity_count", "entity_type": "interactable", "interactable_def_id": "treasure_chest", "state": "open", "equals": 1},
        {"type": "inventory_count", "item_def_id": "red_potion", "between": [1, 3]},
        {"type": "inventory_count", "equipped": True, "equals": 1},
    ], entities, inventory, {}, None, "test")


def test_structured_assertions_reject_range_mismatch():
    try:
        run_assertions([
            {"type": "entity_count", "entity_type": "monster", "between": [1, 2]},
        ], [
            {"id": "1003", "type": "monster"},
            {"id": "1004", "type": "monster"},
            {"id": "1005", "type": "monster"},
        ], [], {}, None, "test")
    except AssertionError as exc:
        assert "not between 1 and 2" in str(exc)
    else:
        raise AssertionError("expected AssertionError")


def test_structured_assertions_reject_unknown_type():
    try:
        run_assertions([{"type": "nope"}], [], [], {}, None, "test")
    except AssertionError as exc:
        assert "unknown assertion type" in str(exc)
    else:
        raise AssertionError("expected AssertionError")


def test_structured_character_progression_stat_breakdowns():
    progression = {
        "level": 1,
        "experience": 0,
        "unspent_stat_points": 0,
        "base_stats": {"str": 1, "dex": 1, "vit": 1, "magic": 1},
        "derived_stats": {"armor": 4, "block_percent": 8},
        "stat_breakdowns": [
            {
                "key": "armor",
                "value": 4,
                "uncapped_value": 4,
                "cap": None,
                "sources": [
                    {"kind": "character_formula", "label": "Dexterity", "value": 1},
                    {"kind": "equipment_base", "label": "Shield", "value": 2},
                    {"kind": "equipment_roll", "label": "Rolled armor", "value": 1},
                ],
            },
            {
                "key": "block_percent",
                "value": 8,
                "uncapped_value": 8,
                "cap": 75,
                "sources": [{"kind": "equipment_base", "label": "Shield", "value": 8}],
            },
        ],
    }

    run_assertions([
        {
            "type": "character_progression",
            "level": 1,
            "stat_breakdowns": [
                {"key": "armor", "min_value": 4, "source_kinds": ["character_formula", "equipment_base", "equipment_roll"]},
                {"key": "block_percent", "min_uncapped_value": 8, "cap": 75, "source_kinds": ["equipment_base"]},
            ],
        }
    ], [], [], {}, None, "test", character_progression=progression)


def test_skill_progression_assertions_for_runtime_and_state_payloads():
    progression = {
        "unspent_skill_points": 0,
        "skills": [
            {"skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": False},
        ],
    }
    assertion = {"type": "skill_progression", "unspent_skill_points": 0, "skill_id": "magic_bolt", "rank": 1, "max_rank": 5, "can_spend": False}

    state = RuntimeState(skill_progression=progression)
    run_runtime_assertions([assertion], state, "unit runtime")
    run_assertions([assertion], [], [], {}, None, "unit state", skill_progression=progression)


def test_skill_cooldown_assertions_support_ranges_and_absence():
    cooldowns = [{"skill_id": "magic_bolt", "remaining_ticks": 18, "total_ticks": 40}]
    assertion = {"type": "skill_cooldown", "skill_id": "magic_bolt", "remaining_ticks": {"between": [1, 40]}, "total_ticks": 40}

    state = RuntimeState(skill_cooldowns=list(cooldowns))
    run_runtime_assertions([assertion], state, "unit runtime")
    run_assertions([assertion], [], [], {}, None, "unit state", skill_cooldowns=list(cooldowns))
    run_assertions([{"type": "skill_cooldown", "skill_id": "magic_bolt", "absent": True}], [], [], {}, None, "unit state", skill_cooldowns=[])


def test_state_delta_ingest_decays_skill_cooldowns_by_tick_delta():
    state = RuntimeState(last_tick=10, current_level=-1, skill_cooldowns=[{"skill_id": "magic_bolt", "remaining_ticks": 5, "total_ticks": 40}])

    ingest_message(
        {
            "type": "state_delta",
            "tick": 12,
            "payload": {"server_tick": 12, "level": -1, "changes": [], "events": []},
        },
        state,
    )

    assert state.skill_cooldowns == [{"skill_id": "magic_bolt", "remaining_ticks": 3, "total_ticks": 40}]
