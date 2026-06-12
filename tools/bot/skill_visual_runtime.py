"""Runtime expansion for the reusable skill visual scenario."""
from __future__ import annotations

import asyncio
from dataclasses import replace
import os
from typing import Any

import httpx

from tools.bot.bot_types import CoopPeer, RuntimeState, Scenario
from tools.bot.skill_demo import SkillDemoEntry, skill_demo_entry

SKILL_VISUAL_ENV = "ARPG_SKILL_VISUAL_SKILL_ID"


def selected_skill_visual_entry() -> SkillDemoEntry:
    skill_id = os.environ.get(SKILL_VISUAL_ENV, "").strip()
    if not skill_id:
        raise ValueError(f"skill_visual requires {SKILL_VISUAL_ENV}=<skill_id>")
    return skill_demo_entry(skill_id)


def runtime_scenario(scenario: Scenario, entry: SkillDemoEntry) -> Scenario:
    return replace(
        scenario,
        seed=f"skill_visual_{entry.skill_id}",
        character_class=entry.class_id,
        title=f"Skill Visual - {entry.name}",
        description=(
            f"Reusable skill visual replay for {entry.name}: learn the skill, cast it through "
            f"{entry.targeting}, and hold after the cast."
        ),
    )


def setup_monster_id(entry: SkillDemoEntry) -> str:
    if entry.kind == "self_buff":
        return "skill_xp_level7_dummy"
    return "skill_xp_level6_dummy"


def rank_assertion(entry: SkillDemoEntry) -> dict[str, Any]:
    return {
        "type": "skill_progression",
        "skill_id": entry.skill_id,
        "rank": 1,
        "max_rank": entry.max_rank,
        "can_spend": False,
    }


def build_steps(entry: SkillDemoEntry) -> list[dict[str, Any]]:
    steps: list[dict[str, Any]] = [
        {
            "action": "attack_until_event",
            "monster_def_id": setup_monster_id(entry),
            "event_type": "monster_killed",
            "timeout_s": 20,
        },
        {"action": "allocate_skill_point", "skill_id": entry.skill_id},
    ]
    if entry.kind == "projectile_attack":
        steps.extend([
            {"action": "move_until_player_position", "x": 4, "y": 5, "pathfind": True, "max_ticks": 220},
            {
                "action": "cast_skill",
                "skill_id": entry.skill_id,
                "monster_def_id": "combat_lab_soft_target",
                "event_type": "skill_cast",
            },
            {
                "action": "wait_until_assertion",
                "assertion": {"type": "combat_event_seen", "event_type": "monster_damaged", "min_damage": 1},
                "timeout_s": 8,
            },
        ])
    else:
        steps.append({
            "action": "cast_skill",
            "skill_id": entry.skill_id,
            "target_self": True,
            "event_type": "skill_cast",
        })
    steps.append({"action": "wait_ticks", "ticks": 30})
    return steps


def build_assertions(entry: SkillDemoEntry) -> list[dict[str, Any]]:
    assertions: list[dict[str, Any]] = [
        {"type": "event_seen", "event_type": "skill_rank_updated", "skill_id": entry.skill_id, "rank": 1},
        {"type": "event_seen", "event_type": "skill_cast", "skill_id": entry.skill_id, "rank": 1},
        rank_assertion(entry),
    ]
    if entry.kind == "projectile_attack":
        assertions.append({"type": "combat_event_seen", "event_type": "monster_damaged", "min_damage": 1})
    elif entry.kind in {"self_buff", "area_stat_buff"}:
        assertions.append({"type": "event_seen", "event_type": "skill_effect_started", "skill_id": entry.skill_id})
    elif entry.kind == "area_heal":
        assertions.append({"type": "event_seen", "event_type": "player_healed", "skill_id": entry.skill_id})
    return assertions


def targets_ally(entry: SkillDemoEntry) -> bool:
    return entry.kind in {"area_heal", "area_stat_buff"} and entry.targeting != "self"


async def wait_for_skill_rank(ctx: dict[str, Any], peer: CoopPeer, skill_id: str, rank: int, peers: list[CoopPeer]) -> None:
    await ctx["wait_coop_until"](
        peers,
        f"{peer.label} {skill_id} rank {rank}",
        lambda: any(
            str(row.get("skill_id")) == skill_id and int(row.get("rank", 0)) == rank
            for row in peer.state.skill_progression.get("skills", [])
        ),
    )


async def wait_for_event(
    ctx: dict[str, Any],
    peer: CoopPeer,
    event_type: str,
    peers: list[CoopPeer],
    *,
    skill_id: str = "",
) -> None:
    await ctx["wait_coop_until"](
        peers,
        f"{peer.label} event {event_type}",
        lambda: any(
            event.get("event_type") == event_type and (not skill_id or event.get("skill_id") == skill_id)
            for event in peer.state.events
        ),
    )


async def wait_ticks(ctx: dict[str, Any], peers: list[CoopPeer], driver: CoopPeer, ticks: int) -> None:
    target_tick = driver.state.last_tick + ticks
    while driver.state.last_tick < target_tick:
        message_id = await ctx["send_coop_intent"](
            driver,
            "move_intent",
            {"direction": {"x": 0, "y": 0}, "duration_ticks": 1},
        )
        await ctx["wait_coop_accept"](peers, driver, message_id)
        await ctx["pump_coop"](peers, timeout=0.05)


async def damage_ally_for_heal(ctx: dict[str, Any], peers: list[CoopPeer], ally: CoopPeer) -> None:
    start_indexes = {peer.label: len(peer.state.combat_events) for peer in peers}
    loop = asyncio.get_event_loop()
    deadline = loop.time() + ctx["SLICE_TIMEOUT_S"]
    last_action = 0.0
    while loop.time() <= deadline:
        for peer in peers:
            for event in peer.state.combat_events[start_indexes[peer.label]:]:
                if (
                    event.get("event_type") == "player_damaged"
                    and str(event.get("target_entity_id", "")) == ally.state.local_player_id
                ):
                    return
        monster = ctx["find_monster"](ally.state, "training_dummy_heal")
        if monster is None:
            await ctx["pump_coop"](peers, timeout=0.1)
            continue
        if loop.time() - last_action >= 0.12:
            message_id = await ctx["send_coop_intent"](ally, "action_intent", {"target_id": str(monster["id"])})
            await ctx["wait_coop_until"](
                peers,
                f"{ally.label} takes heal setup damage {message_id}",
                lambda: message_id in ally.state.accepted_message_ids
                or message_id in ally.state.rejected_message_reasons,
            )
            reason = ally.state.rejected_message_reasons.pop(message_id, None)
            if reason is not None and reason not in {"basic_attack_on_cooldown", "no_path", "path_too_long"}:
                raise AssertionError(f"{ally.label} heal setup action rejected: {reason}")
            last_action = loop.time()
        await ctx["pump_coop"](peers, timeout=0.1)
    raise TimeoutError("heal setup did not damage ally")


async def run_ally_cast(
    ctx: dict[str, Any],
    *,
    client: httpx.Client,
    base_url: str,
    host_token: str,
    ally_token: str,
    scenario: Scenario,
    host_character_id: str,
    ally_character_id: str,
    entry: SkillDemoEntry,
) -> tuple[dict[str, Any], RuntimeState]:
    sess = ctx["create_coop_session"](client, host_token, scenario.world_id, host_character_id, scenario.seed)
    session_id = str(sess["session_id"])
    host = await ctx["connect_coop_peer"](base_url, host_token, sess, "host", scenario.world_id)
    joined = ctx["join_coop_session"](client, ally_token, session_id, str(sess["join_code"]), ally_character_id)
    ally = await ctx["connect_coop_peer"](base_url, ally_token, joined, "ally", scenario.world_id)
    peers = [host, ally]
    try:
        await ctx["wait_coop_until"](
            peers,
            "host and ally visible",
            lambda: ctx["player_entity_ids"](host.state) >= {host.state.local_player_id, ally.state.local_player_id}
            and ctx["player_entity_ids"](ally.state) >= {host.state.local_player_id, ally.state.local_player_id},
        )
        await ctx["coop_attack_until_kill"](peers, host, setup_monster_id(entry), companions=[ally])

        rank_message = await ctx["send_coop_intent"](host, "allocate_skill_point_intent", {"skill_id": entry.skill_id})
        await ctx["wait_coop_accept"](peers, host, rank_message)
        await wait_for_skill_rank(ctx, host, entry.skill_id, 1, peers)

        host_stage = {"x": 4.0, "y": 5.0}
        ally_stage = {"x": 4.0, "y": 8.0}
        await ctx["move_coop_peer_to"](peers, host, host_stage, stop_distance=0.35, max_ticks=320)
        await ctx["move_coop_peer_to"](peers, ally, ally_stage, stop_distance=0.35, max_ticks=320)

        if entry.kind == "area_heal":
            await damage_ally_for_heal(ctx, peers, ally)
            await ctx["move_coop_peer_to"](peers, ally, ally_stage, stop_distance=0.35, max_ticks=320)

        cast_message = await ctx["send_coop_intent"](
            host,
            "cast_skill_intent",
            {"skill_id": entry.skill_id, "target_id": ally.state.local_player_id},
        )
        await ctx["wait_coop_accept"](peers, host, cast_message)
        await wait_for_event(ctx, host, "skill_cast", peers, skill_id=entry.skill_id)
        if entry.kind == "area_heal":
            await wait_for_event(ctx, host, "player_healed", peers, skill_id=entry.skill_id)
        elif entry.kind == "area_stat_buff":
            await wait_for_event(ctx, host, "skill_effect_started", peers, skill_id=entry.skill_id)
        await wait_ticks(ctx, peers, host, 30)
        return sess, host.state
    finally:
        for peer in peers:
            try:
                await ctx["close_coop_peer"](peer)
            except Exception:
                pass


def run_selected(ctx: dict[str, Any], args: Any, client: httpx.Client, scenario: Scenario) -> tuple[Scenario, str, str, dict[str, Any], RuntimeState]:
    entry = selected_skill_visual_entry()
    scenario = runtime_scenario(scenario, entry)
    replay_email = ctx["scenario_email"](args.email, f"{scenario.id}-{entry.skill_id}-host")
    _, token = ctx["dev_login"](client, replay_email, args.dev_token)
    host_character_id = ctx["ensure_character"](client, token, f"{entry.name} Visual Host", entry.class_id)
    if targets_ally(entry):
        ally_email = ctx["scenario_email"](args.email, f"{scenario.id}-{entry.skill_id}-ally")
        _, ally_token = ctx["dev_login"](client, ally_email, args.dev_token)
        ally_character_id = ctx["ensure_character"](client, ally_token, f"{entry.name} Visual Ally")
        sess, observed = asyncio.run(run_ally_cast(
            ctx,
            client=client,
            base_url=args.base_url,
            host_token=token,
            ally_token=ally_token,
            scenario=scenario,
            host_character_id=host_character_id,
            ally_character_id=ally_character_id,
            entry=entry,
        ))
    else:
        sess, observed = ctx["run_verified_session"](
            client=client,
            base_url=args.base_url,
            token=token,
            debug_token=args.debug_token,
            scenario=scenario,
            world_id=scenario.world_id,
            steps=build_steps(entry),
            assertions=build_assertions(entry),
            seed=scenario.seed,
        )
    return scenario, replay_email, token, sess, observed
