"""Runtime expansion for the reusable skill visual scenario."""
from __future__ import annotations

import asyncio
from dataclasses import replace
import math
import os
from pathlib import Path
from typing import Any

import httpx

from tools.bot.bot_types import CoopPeer, RuntimeState, Scenario
from tools.bot.skill_demo import ROOT, SkillDemoEntry, load_json, load_skill_rules, skill_demo_entry

SKILL_VISUAL_ENV = "ARPG_SKILL_VISUAL_SKILL_ID"
SKILL_VISUAL_RANK_ENV = "ARPG_SKILL_VISUAL_RANK"
SKILL_VISUAL_LEVEL_ENV = "ARPG_SKILL_VISUAL_LEVEL"
POST_CAST_HOLD_TICKS = 40
FAST_POST_CAST_HOLD_TICKS = 20


def selected_skill_visual_entry() -> SkillDemoEntry:
    skill_id = os.environ.get(SKILL_VISUAL_ENV, "").strip()
    if not skill_id:
        raise ValueError(f"skill_visual requires {SKILL_VISUAL_ENV}=<skill_id>")
    return skill_demo_entry(skill_id)


def selected_skill_visual_rank(entry: SkillDemoEntry) -> int:
    raw = os.environ.get(SKILL_VISUAL_RANK_ENV, "1").strip() or "1"
    rank = int(raw)
    if rank < 1 or rank > entry.max_rank:
        raise ValueError(f"skill_visual rank for {entry.skill_id} must be between 1 and {entry.max_rank}")
    return rank


def selected_skill_visual_level(entry: SkillDemoEntry, rank: int) -> int:
    raw = os.environ.get(SKILL_VISUAL_LEVEL_ENV, "").strip()
    minimum = skill_required_level(entry.skill_id, rank)
    if not raw:
        return minimum
    level = int(raw)
    if level < minimum:
        raise ValueError(f"skill_visual level for {entry.skill_id} rank {rank} must be >= {minimum}")
    return level


def runtime_scenario(scenario: Scenario, entry: SkillDemoEntry) -> Scenario:
    return replace(
        scenario,
        seed=f"skill_visual_{entry.skill_id}",
        character_class=entry.class_id,
        title=f"Skill Visual - {entry.name}",
        description=(
            f"Reusable skill visual replay for {entry.name}: seed the requested rank, cast it through "
            f"{entry.targeting}, and hold after the cast."
        ),
    )


def skill_rule(skill_id: str) -> dict[str, Any]:
    return dict(load_skill_rules().get(skill_id, {}))


def skill_required_level(skill_id: str, rank: int) -> int:
    req = dict(skill_rule(skill_id).get("requirements", {}))
    return int(req.get("level", 1)) + max(0, rank - 1) * int(req.get("level_per_rank", 0))


def skill_required_stats(skill_id: str, rank: int) -> dict[str, int]:
    req = dict(skill_rule(skill_id).get("requirements", {}))
    stats = {str(k): int(v) for k, v in dict(req.get("stats", {})).items()}
    per_rank = {str(k): int(v) for k, v in dict(req.get("stats_per_rank", {})).items()}
    for stat, value in per_rank.items():
        stats[stat] = stats.get(stat, 0) + max(0, rank - 1) * value
    return stats


def skill_required_skill_ranks(skill_id: str) -> dict[str, int]:
    req = dict(skill_rule(skill_id).get("requirements", {}))
    out: dict[str, int] = {}
    for raw in list(req.get("skills", [])):
        if not isinstance(raw, dict):
            continue
        prereq_id = str(raw.get("skill_id", ""))
        prereq_rank = int(raw.get("rank", 0))
        if prereq_id and prereq_rank > 0:
            out[prereq_id] = max(out.get(prereq_id, 0), prereq_rank)
    return out


def skill_mana_cost(skill_id: str, rank: int) -> int:
    mana = dict(dict(skill_rule(skill_id).get("cost", {})).get("mana", {}))
    base = int(mana.get("base", 0))
    per_rank = int(mana.get("per_rank", 0))
    return base + max(0, rank - 1) * per_rank


def magic_required_for_mana(mana_cost: int, root: Path = ROOT) -> int:
    if mana_cost <= 0:
        return 0
    progression = load_json(root / "shared" / "rules" / "character_progression.v0.json")
    max_mana = dict(dict(progression.get("derived_stats", {})).get("max_mana", {}))
    base = float(max_mana.get("base", 0.0))
    per_magic = float(max_mana.get("per_magic", 0.0))
    minimum = float(max_mana.get("min", 0.0))
    if per_magic <= 0:
        return 0
    required = max(float(mana_cost), minimum)
    return max(0, int(math.ceil((required - base) / per_magic)))


def base_stats_for_class(class_id: str, root: Path = ROOT) -> dict[str, int]:
    progression = load_json(root / "shared" / "rules" / "character_progression.v0.json")
    class_stats = dict(progression.get("classes", {}).get(class_id, {}).get("base_stats", {}))
    if not class_stats:
        class_stats = dict(progression.get("base_stats", {}))
    return {str(k): int(v) for k, v in class_stats.items()}


def rank_assertion(entry: SkillDemoEntry, rank: int) -> dict[str, Any]:
    return {
        "type": "skill_progression",
        "skill_id": entry.skill_id,
        "rank": rank,
        "max_rank": entry.max_rank,
        "can_spend": False,
    }


def build_steps(entry: SkillDemoEntry) -> list[dict[str, Any]]:
    steps: list[dict[str, Any]] = []
    if entry.kind in {"cone_attack", "mobility"}:
        steps.extend([
            {"action": "move_until_player_position", "x": 7, "y": 5, "pathfind": True, "max_ticks": 220},
            {
                "action": "cast_skill",
                "skill_id": entry.skill_id,
                "direction": {"x": 1, "y": 0},
                "event_type": "skill_cast",
            },
        ])
        if entry.kind == "cone_attack":
            steps.append({
                "action": "wait_until_assertion",
                "assertion": {"type": "combat_event_seen", "event_type": "monster_damaged", "min_damage": 1},
                "timeout_s": 8,
            })
    elif entry.kind in {"projectile_attack", "cold_projectile_attack", "chain_projectile_attack"}:
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
    steps.append({"action": "wait_ticks", "ticks": post_cast_hold_ticks(entry)})
    return steps


def post_cast_hold_ticks(entry: SkillDemoEntry) -> int:
    if entry.kind == "area_heal":
        return FAST_POST_CAST_HOLD_TICKS
    return POST_CAST_HOLD_TICKS


def build_assertions(entry: SkillDemoEntry, rank: int = 1) -> list[dict[str, Any]]:
    assertions: list[dict[str, Any]] = [
        {"type": "event_seen", "event_type": "skill_cast", "skill_id": entry.skill_id, "rank": rank},
        rank_assertion(entry, rank),
    ]
    if entry.kind in {"projectile_attack", "cold_projectile_attack", "chain_projectile_attack", "cone_attack"}:
        assertions.append({"type": "combat_event_seen", "event_type": "monster_damaged", "min_damage": 1})
    elif entry.kind in {"self_buff", "area_stat_buff"}:
        assertions.append({"type": "event_seen", "event_type": "skill_effect_started", "skill_id": entry.skill_id})
    return assertions


def targets_ally(entry: SkillDemoEntry) -> bool:
    return entry.kind == "area_stat_buff" and entry.targeting != "self"


def seed_skill_visual_character(
    client: httpx.Client,
    token: str,
    debug_token: str,
    character_id: str,
    entry: SkillDemoEntry,
    rank: int,
    level: int,
) -> None:
    payload = skill_visual_debug_progression(entry, rank, level)
    resp = client.put(
        f"/v0/debug/characters/{character_id}/progression",
        headers={**ctx_auth(token), "X-Debug-Token": debug_token},
        json=payload,
    )
    resp.raise_for_status()


def skill_visual_debug_progression(entry: SkillDemoEntry, rank: int, level: int) -> dict[str, Any]:
    stats = base_stats_for_class(entry.class_id)
    for stat, required in skill_required_stats(entry.skill_id, rank).items():
        stats[stat] = max(stats.get(stat, 0), required)
    stats["magic"] = max(stats.get("magic", 0), magic_required_for_mana(skill_mana_cost(entry.skill_id, rank)))
    skill_ranks = skill_required_skill_ranks(entry.skill_id)
    skill_ranks[entry.skill_id] = rank
    return {
        "level": level,
        "experience": 0,
        "unspent_stat_points": 0,
        "unspent_skill_points": 0,
        "stats": {
            "str": stats.get("str", 0),
            "dex": stats.get("dex", 0),
            "vit": stats.get("vit", 0),
            "magic": stats.get("magic", 0),
        },
        "skill_ranks": skill_ranks,
    }


def ctx_auth(token: str) -> dict[str, str]:
    return {"Authorization": f"Bearer {token}"}


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
    rank: int,
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
        await wait_for_skill_rank(ctx, host, entry.skill_id, rank, peers)

        host_stage = {"x": 4.0, "y": 5.0}
        ally_stage = {"x": 4.0, "y": 6.0}
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
        await wait_ticks(ctx, peers, host, post_cast_hold_ticks(entry))
        return sess, host.state
    finally:
        for peer in peers:
            try:
                await ctx["close_coop_peer"](peer)
            except Exception:
                pass


def run_selected(ctx: dict[str, Any], args: Any, client: httpx.Client, scenario: Scenario) -> tuple[Scenario, str, str, dict[str, Any], RuntimeState]:
    entry = selected_skill_visual_entry()
    rank = selected_skill_visual_rank(entry)
    level = selected_skill_visual_level(entry, rank)
    scenario = runtime_scenario(scenario, entry)
    replay_email = ctx["scenario_email"](args.email, f"{scenario.id}-{entry.skill_id}-host")
    _, token = ctx["dev_login"](client, replay_email, args.dev_token)
    if targets_ally(entry):
        host_character_id = ctx["ensure_character"](client, token, f"{entry.name} Visual Host", entry.class_id)
        seed_skill_visual_character(client, token, args.debug_token, host_character_id, entry, rank, level)
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
            rank=rank,
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
            assertions=build_assertions(entry, rank),
            seed=scenario.seed,
            debug_progression=skill_visual_debug_progression(entry, rank, level),
        )
    return scenario, replay_email, token, sess, observed
