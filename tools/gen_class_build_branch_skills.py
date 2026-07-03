#!/usr/bin/env python3
"""One-shot generator for v420–v424 class build-branch skills. Data-only; reuses existing kinds."""
from __future__ import annotations

import json
import sys
from copy import deepcopy
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SKILLS_PATH = ROOT / "shared/rules/skills.v0.json"
I18N_PATH = ROOT / "shared/i18n/en.json"
PRES_PATH = ROOT / "shared/assets/skill_presentations.v0.json"

CLASSES = ("barbarian", "sorcerer", "paladin", "rogue", "ranger")


def load_json(path: Path) -> dict:
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


def save_json(path: Path, data: dict) -> None:
    with path.open("w", encoding="utf-8") as fh:
        json.dump(data, fh, indent=2, ensure_ascii=False)
        fh.write("\n")


def clone_skill(catalog: dict, source_id: str, new_id: str, **overrides) -> dict:
    base = deepcopy(catalog[source_id])
    if "class_id" in overrides:
        base["class"] = overrides.pop("class_id")
    base.update(overrides)
    if "requirements" in overrides:
        base["requirements"] = overrides["requirements"]
    if "tree" in overrides:
        base["tree"] = overrides["tree"]
    if "synergies" in overrides:
        base["synergies"] = overrides["synergies"]
    base["name_key"] = f"skill.{new_id}.name"
    return base


def barbarian_skills(catalog: dict) -> dict[str, dict]:
    return {
        "rampage": clone_skill(
            catalog,
            "rage",
            "rampage",
            name="Rampage",
            tree={"tier": 3, "column": 2},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"str": 12, "vit": 10},
                "stats_per_rank": {"str": 1, "vit": 1},
                "skills": [{"skill_id": "rage", "rank": 1}],
            },
            effects=[
                {
                    "type": "stat_percent_buff",
                    "stats": ["str", "dex"],
                    "percent_base": 15,
                    "percent_per_rank": 10,
                    "duration_ticks": 450,
                    "visual_scale": True,
                }
            ],
            synergies=[
                {"source_skill_id": "rage", "modifier": "buff_power_percent", "percent_per_source_rank": 8}
            ],
        ),
        "shatter_strike": clone_skill(
            catalog,
            "skullcrusher",
            "shatter_strike",
            name="Shatter Strike",
            tree={"tier": 3, "column": 1},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"str": 12, "vit": 10},
                "stats_per_rank": {"str": 1, "vit": 1},
                "skills": [{"skill_id": "earthbreaker", "rank": 1}],
            },
            cone={"range": 5, "angle_degrees": 35, "push_min": 2, "push_max": 5, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "earthbreaker", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "battle_roar": clone_skill(
            catalog,
            "war_cry",
            "battle_roar",
            name="Battle Roar",
            tree={"tier": 4, "column": 4},
            requirements={
                "level": 18,
                "level_per_rank": 1,
                "stats": {"str": 14, "vit": 12},
                "stats_per_rank": {"str": 1, "vit": 1},
                "skills": [{"skill_id": "war_cry", "rank": 1}],
            },
            effects=[
                {
                    "type": "area_stat_percent_buff",
                    "target": "allies",
                    "include_caster": True,
                    "range": 9,
                    "radius": 6,
                    "stats": ["str", "vit"],
                    "percent_base": 15,
                    "percent_per_rank": 8,
                    "duration_ticks": 350,
                    "effect_id": "battle_roar",
                }
            ],
            synergies=[
                {"source_skill_id": "leap", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
        "worldbreaker": clone_skill(
            catalog,
            "earthbreaker",
            "worldbreaker",
            name="Worldbreaker",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"str": 16, "vit": 14},
                "stats_per_rank": {"str": 2, "vit": 1},
                "skills": [{"skill_id": "shatter_strike", "rank": 1}],
            },
            cone={"range": 8, "angle_degrees": 360, "push_min": 4, "push_max": 10, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "earthbreaker", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "blood_frenzy": clone_skill(
            catalog,
            "rage",
            "blood_frenzy",
            name="Blood Frenzy",
            tree={"tier": 4, "column": 2},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"str": 16, "vit": 14},
                "stats_per_rank": {"str": 2, "vit": 1},
                "skills": [{"skill_id": "rampage", "rank": 1}],
            },
            effects=[
                {
                    "type": "stat_percent_buff",
                    "stats": ["str", "vit"],
                    "percent_base": 20,
                    "percent_per_rank": 12,
                    "duration_ticks": 550,
                    "visual_scale": True,
                }
            ],
            synergies=[
                {"source_skill_id": "rage", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
        "gore_strike": clone_skill(
            catalog,
            "rend",
            "gore_strike",
            name="Gore Strike",
            tree={"tier": 4, "column": 3},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"str": 16, "vit": 14},
                "stats_per_rank": {"str": 2, "vit": 1},
                "skills": [{"skill_id": "rend", "rank": 1}],
            },
            bleed={
                "effect_id": "gore_bleed",
                "damage_percent_max_hp": 8,
                "damage_percent_max_hp_per_rank": 2,
                "duration_ticks": 60,
                "duration_ticks_per_rank": 10,
                "interval_ticks": 10,
            },
            synergies=[
                {"source_skill_id": "ground_slam", "modifier": "bleed_duration_percent", "percent_per_source_rank": 8}
            ],
        ),
    }


def sorcerer_skills(catalog: dict) -> dict[str, dict]:
    return {
        "glacial_lance": clone_skill(
            catalog,
            "ice_shard",
            "glacial_lance",
            name="Glacial Lance",
            tree={"tier": 3, "column": 1},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"magic": 14},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "ice_shard", "rank": 3}],
            },
            damage={
                "type": "weapon_multiplier_range",
                "min_base": 200,
                "max_base": 175,
                "min_per_rank": 55,
                "max_per_rank": 30,
                "magic_scaling": {
                    "stat": "magic",
                    "percent_per_point": 1,
                    "max_bonus_percent": 25,
                    "use_requirement_baseline": True,
                },
            },
            slow={"effect_id": "ice_slow", "percent": 35, "duration_ticks": 80, "max_percent": 80},
            synergies=[
                {"source_skill_id": "magic_bolt", "modifier": "damage_percent", "percent_per_source_rank": 8}
            ],
        ),
        "chain_storm": clone_skill(
            catalog,
            "lightning",
            "chain_storm",
            name="Chain Storm",
            tree={"tier": 3, "column": 2},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"magic": 14},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "lightning", "rank": 1}],
            },
            chain={"range_multiplier": 0.9, "max_jumps": 10, "visual": "lightning_chain"},
            synergies=[
                {"source_skill_id": "magic_bolt", "modifier": "damage_percent", "percent_per_source_rank": 8}
            ],
        ),
        "renewing_light": clone_skill(
            catalog,
            "heal",
            "renewing_light",
            name="Renewing Light",
            class_id="sorcerer",
            tree={"tier": 3, "column": 2},
            kind="area_heal",
            max_rank=10,
            targeting="direction_or_target_area",
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"magic": 12},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "revive", "rank": 1}],
            },
            cost={"mana": {"base": 8, "per_rank": 1}},
            effects=[
                {
                    "type": "area_percent_heal",
                    "target": "allies",
                    "include_caster": True,
                    "range": 8,
                    "radius": 5,
                    "percent_base": 20,
                    "percent_per_rank": 10,
                    "duration_ticks": 30,
                    "magic_scaling": {
                        "stat": "magic",
                        "percent_per_point": 1,
                        "max_bonus_percent": 25,
                        "use_requirement_baseline": True,
                    },
                }
            ],
            cooldown={"type": "attack_interval_multiplier", "multiplier": 12},
            synergies=[
                {"source_skill_id": "revive", "modifier": "revive_power_percent", "percent_per_source_rank": 10}
            ],
        ),
        "inferno": clone_skill(
            catalog,
            "fireball",
            "inferno",
            name="Inferno",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"magic": 18},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "fireball", "rank": 1}],
            },
            damage={
                "type": "weapon_multiplier_range",
                "min_base": 320,
                "max_base": 260,
                "min_per_rank": 60,
                "max_per_rank": 35,
                "magic_scaling": {
                    "stat": "magic",
                    "percent_per_point": 1,
                    "max_bonus_percent": 25,
                    "use_requirement_baseline": True,
                },
            },
            synergies=[
                {"source_skill_id": "ice_shard", "modifier": "damage_percent", "percent_per_source_rank": 10}
            ],
        ),
        "arcane_overload": clone_skill(
            catalog,
            "energy_ward",
            "arcane_overload",
            name="Arcane Overload",
            tree={"tier": 4, "column": 3},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"magic": 16},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "arcane_barrage", "rank": 1}],
            },
            effects=[
                {
                    "type": "stat_percent_buff",
                    "stats": ["vit", "magic"],
                    "percent_base": 20,
                    "percent_per_rank": 10,
                    "duration_ticks": 500,
                    "visual_scale": True,
                }
            ],
            synergies=[
                {"source_skill_id": "lightning", "modifier": "buff_power_percent", "percent_per_source_rank": 8}
            ],
        ),
        "arcane_renewal": clone_skill(
            catalog,
            "revive",
            "arcane_renewal",
            name="Arcane Renewal",
            tree={"tier": 4, "column": 2},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"magic": 16},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "renewing_light", "rank": 1}],
            },
            cost={"mana": {"base": 10, "per_rank": 1}},
            synergies=[
                {"source_skill_id": "revive", "modifier": "revive_power_percent", "percent_per_source_rank": 12}
            ],
        ),
    }


def paladin_skills(catalog: dict) -> dict[str, dict]:
    return {
        "blessed_recovery": clone_skill(
            catalog,
            "heal",
            "blessed_recovery",
            name="Blessed Recovery",
            tree={"tier": 3, "column": 3},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"magic": 12, "vit": 10},
                "stats_per_rank": {"magic": 2, "vit": 1},
                "skills": [{"skill_id": "heal", "rank": 1}],
            },
            effects=[
                {
                    "type": "area_percent_heal",
                    "target": "allies",
                    "include_caster": True,
                    "range": 10,
                    "radius": 5,
                    "percent_base": 30,
                    "percent_per_rank": 10,
                    "duration_ticks": 30,
                    "magic_scaling": {
                        "stat": "magic",
                        "percent_per_point": 1,
                        "max_bonus_percent": 25,
                        "use_requirement_baseline": True,
                    },
                }
            ],
            synergies=[
                {"source_skill_id": "heal", "modifier": "area_radius_percent", "percent_per_source_rank": 5}
            ],
        ),
        "avenging_light": clone_skill(
            catalog,
            "consecrated_smite",
            "avenging_light",
            name="Avenging Light",
            tree={"tier": 3, "column": 1},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"str": 12, "magic": 10},
                "stats_per_rank": {"str": 1, "magic": 1},
                "skills": [{"skill_id": "consecrated_smite", "rank": 1}],
            },
            cone={"range": 7, "angle_degrees": 55, "push_min": 2, "push_max": 4, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "radiant_bolt", "modifier": "damage_percent", "percent_per_source_rank": 10}
            ],
        ),
        "bulwark_aura": clone_skill(
            catalog,
            "holy_shield",
            "bulwark_aura",
            name="Bulwark Aura",
            tree={"tier": 3, "column": 4},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"vit": 12, "magic": 10},
                "stats_per_rank": {"vit": 2, "magic": 1},
                "skills": [{"skill_id": "holy_shield", "rank": 1}],
            },
            effects=[
                {
                    "type": "area_stat_percent_buff",
                    "target": "allies",
                    "include_caster": True,
                    "range": 9,
                    "radius": 5,
                    "stats": ["armor", "block_percent"],
                    "percent_base": 25,
                    "percent_per_rank": 8,
                    "duration_ticks": 350,
                    "effect_id": "bulwark_aura",
                    "magic_scaling": {
                        "stat": "magic",
                        "percent_per_point": 1,
                        "max_bonus_percent": 20,
                        "use_requirement_baseline": True,
                    },
                }
            ],
            synergies=[
                {"source_skill_id": "holy_shield", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
        "divine_hammer": clone_skill(
            catalog,
            "hammer_of_light",
            "divine_hammer",
            name="Divine Hammer",
            tree={"tier": 4, "column": 2},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"str": 16, "magic": 12},
                "stats_per_rank": {"str": 2, "magic": 1},
                "skills": [{"skill_id": "hammer_of_light", "rank": 1}],
            },
            cone={"range": 8, "angle_degrees": 70, "push_min": 4, "push_max": 7, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "charge", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "sacred_ground": clone_skill(
            catalog,
            "sanctuary",
            "sacred_ground",
            name="Sacred Ground",
            tree={"tier": 4, "column": 4},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"vit": 16, "magic": 14},
                "stats_per_rank": {"vit": 2, "magic": 1},
                "skills": [{"skill_id": "sanctuary", "rank": 1}],
            },
            effects=[
                {
                    "type": "area_immunity_buff",
                    "target": "allies",
                    "include_caster": True,
                    "range": 6,
                    "radius": 5,
                    "duration_ticks": 40,
                    "effect_id": "sacred_ground",
                    "magic_scaling": {
                        "stat": "magic",
                        "percent_per_point": 1,
                        "max_bonus_percent": 15,
                        "use_requirement_baseline": True,
                    },
                }
            ],
            synergies=[
                {"source_skill_id": "holy_shield", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
        "righteous_fury": clone_skill(
            catalog,
            "retribution",
            "righteous_fury",
            name="Righteous Fury",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"vit": 16, "magic": 12},
                "stats_per_rank": {"vit": 2, "magic": 1},
                "skills": [{"skill_id": "retribution", "rank": 1}],
            },
            effects=[
                {
                    "type": "reflect_on_block_buff",
                    "percent_base": 50,
                    "percent_per_rank": 10,
                    "duration_ticks": 350,
                    "effect_id": "righteous_fury",
                }
            ],
            synergies=[
                {"source_skill_id": "hammer_of_light", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
    }


def rogue_skills(catalog: dict) -> dict[str, dict]:
    return {
        "blade_dance": clone_skill(
            catalog,
            "shadow_flurry",
            "blade_dance",
            name="Blade Dance",
            tree={"tier": 3, "column": 1},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"dex": 14},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "shadow_flurry", "rank": 1}],
            },
            cone={"range": 4, "angle_degrees": 100, "push_min": 0, "push_max": 0, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "dash", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "venom_spray": clone_skill(
            catalog,
            "fan_of_blades",
            "venom_spray",
            name="Venom Spray",
            tree={"tier": 3, "column": 3},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"dex": 14},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "fan_of_blades", "rank": 1}],
            },
            poison={
                "damage_percent_base": 20,
                "damage_percent_per_rank": 10,
                "duration_ticks": 50,
                "magic_duration_ticks_per_point": 2,
                "mark_damage_bonus_percent": 20,
                "mark_duration_ticks": 50,
                "mark_effect_id": "rogue_mark",
            },
            synergies=[
                {"source_skill_id": "poison_stab", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "shadow_veil": clone_skill(
            catalog,
            "smoke_screen",
            "shadow_veil",
            name="Shadow Veil",
            tree={"tier": 3, "column": 4},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"dex": 14},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "smoke_screen", "rank": 1}],
            },
            effects=[
                {
                    "type": "area_stat_percent_buff",
                    "target": "allies",
                    "include_caster": True,
                    "range": 7,
                    "radius": 5,
                    "stats": ["evade_chance"],
                    "percent_base": 35,
                    "percent_per_rank": 5,
                    "duration_ticks": 120,
                    "effect_id": "shadow_veil",
                }
            ],
            synergies=[
                {"source_skill_id": "shadowstep", "modifier": "buff_duration_percent", "percent_per_source_rank": 5}
            ],
        ),
        "death_blossom": clone_skill(
            catalog,
            "shadow_flurry",
            "death_blossom",
            name="Death Blossom",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"dex": 18},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "blade_dance", "rank": 1}],
            },
            cone={"range": 3.5, "angle_degrees": 360, "push_min": 0, "push_max": 0, "damage_source": "weapon"},
            synergies=[
                {"source_skill_id": "shadow_flurry", "modifier": "cone_size_percent", "percent_per_source_rank": 3}
            ],
        ),
        "assassinate": clone_skill(
            catalog,
            "eviscerate",
            "assassinate",
            name="Assassinate",
            tree={"tier": 4, "column": 2},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"dex": 18},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "eviscerate", "rank": 1}],
            },
            cone={"range": 4, "angle_degrees": 30, "push_min": 0, "push_max": 0, "damage_source": "weapon"},
            poison={
                "damage_percent_base": 25,
                "damage_percent_per_rank": 12,
                "duration_ticks": 60,
                "magic_duration_ticks_per_point": 2,
                "mark_damage_bonus_percent": 25,
                "mark_duration_ticks": 60,
                "mark_effect_id": "rogue_mark",
            },
            synergies=[
                {"source_skill_id": "poison_stab", "modifier": "mark_duration_percent", "percent_per_source_rank": 8}
            ],
        ),
        "killing_mark": clone_skill(
            catalog,
            "predators_mark",
            "killing_mark",
            name="Killing Mark",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"dex": 18},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "predators_mark", "rank": 1}],
            },
            mark={
                "damage_bonus_percent": 30,
                "damage_bonus_percent_per_rank": 5,
                "duration_ticks": 100,
                "effect_id": "rogue_mark",
            },
            synergies=[
                {"source_skill_id": "eviscerate", "modifier": "mark_duration_percent", "percent_per_source_rank": 8}
            ],
        ),
    }


def ranger_skills(catalog: dict) -> dict[str, dict]:
    return {
        "alpha_call": clone_skill(
            catalog,
            "black_wolf_companion",
            "alpha_call",
            name="Alpha Call",
            tree={"tier": 3, "column": 3},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"magic": 12},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "black_wolf_companion", "rank": 1}],
            },
            companion={
                "monster_def_id": "companion_black_wolf",
                "visual_model": "monster_wolf",
                "visual_tint": "101014",
                "visual_scale": 1.1,
                "hero_stat_percent_base": 85,
                "hero_stat_percent_per_rank": 15,
                "limit": {"base": 1, "per_rank_step": 0, "ranks_per_step": 1},
            },
            synergies=[
                {"source_skill_id": "black_wolf_companion", "modifier": "buff_power_percent", "percent_per_source_rank": 10}
            ],
        ),
        "pinning_volley": clone_skill(
            catalog,
            "volley",
            "pinning_volley",
            name="Pinning Volley",
            tree={"tier": 3, "column": 2},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"dex": 14},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "pinning_shot", "rank": 1}],
            },
            volley={"arrow_count": 5, "spread_degrees": 40},
            synergies=[
                {"source_skill_id": "piercing_shot", "modifier": "root_duration_percent", "percent_per_source_rank": 8}
            ],
        ),
        "hunters_volley": clone_skill(
            catalog,
            "volley",
            "hunters_volley",
            name="Hunter's Volley",
            tree={"tier": 3, "column": 1},
            requirements={
                "level": 12,
                "level_per_rank": 1,
                "stats": {"dex": 14},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "volley", "rank": 3}],
            },
            volley={"arrow_count": 6, "spread_degrees": 55},
            synergies=[
                {"source_skill_id": "piercing_shot", "modifier": "volley_spread_percent", "percent_per_source_rank": 5}
            ],
        ),
        "pack_master": clone_skill(
            catalog,
            "black_wolf_companion",
            "pack_master",
            name="Pack Master",
            tree={"tier": 4, "column": 3},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"magic": 16},
                "stats_per_rank": {"magic": 2},
                "skills": [{"skill_id": "alpha_call", "rank": 1}],
            },
            companion={
                "monster_def_id": "companion_black_wolf",
                "visual_model": "monster_wolf",
                "visual_tint": "202028",
                "visual_scale": 1.15,
                "hero_stat_percent_base": 100,
                "hero_stat_percent_per_rank": 20,
                "limit": {"base": 1, "per_rank_step": 0, "ranks_per_step": 1},
            },
            cooldown={"type": "attack_interval_multiplier", "multiplier": 1, "fixed_ticks": 900, "magic_reduction_ticks_per_point": 10},
            synergies=[
                {"source_skill_id": "black_wolf_companion", "modifier": "buff_power_percent", "percent_per_source_rank": 12}
            ],
        ),
        "meteor_shot": clone_skill(
            catalog,
            "explosive_shot",
            "meteor_shot",
            name="Meteor Shot",
            tree={"tier": 4, "column": 4},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"dex": 18},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "explosive_shot", "rank": 1}],
            },
            damage={
                "type": "weapon_multiplier_range",
                "min_base": 420,
                "max_base": 360,
                "min_per_rank": 75,
                "max_per_rank": 40,
            },
            projectile={"range": 22, "speed": 30, "visual": "piercing_shot_projectile"},
            synergies=[
                {"source_skill_id": "snipe", "modifier": "damage_percent", "percent_per_source_rank": 10}
            ],
        ),
        "arrow_storm": clone_skill(
            catalog,
            "rain_of_arrows",
            "arrow_storm",
            name="Arrow Storm",
            tree={"tier": 4, "column": 1},
            requirements={
                "level": 22,
                "level_per_rank": 1,
                "stats": {"dex": 18},
                "stats_per_rank": {"dex": 2},
                "skills": [{"skill_id": "rain_of_arrows", "rank": 1}],
            },
            volley={"arrow_count": 9, "spread_degrees": 60},
            damage={
                "type": "weapon_multiplier_range",
                "min_base": 100,
                "max_base": 100,
                "min_per_rank": 40,
                "max_per_rank": 20,
            },
            synergies=[
                {"source_skill_id": "volley", "modifier": "volley_spread_percent", "percent_per_source_rank": 5}
            ],
        ),
    }


BUILDERS = {
    "barbarian": barbarian_skills,
    "sorcerer": sorcerer_skills,
    "paladin": paladin_skills,
    "rogue": rogue_skills,
    "ranger": ranger_skills,
}

PRESENTATION_TEMPLATES = {
    "rampage": "rage",
    "shatter_strike": "skullcrusher",
    "battle_roar": "war_cry",
    "worldbreaker": "earthbreaker",
    "blood_frenzy": "rage",
    "gore_strike": "rend",
    "glacial_lance": "ice_shard",
    "chain_storm": "lightning",
    "renewing_light": "heal",
    "inferno": "fireball",
    "arcane_overload": "energy_ward",
    "arcane_renewal": "revive",
    "blessed_recovery": "heal",
    "avenging_light": "consecrated_smite",
    "bulwark_aura": "holy_shield",
    "divine_hammer": "hammer_of_light",
    "sacred_ground": "sanctuary",
    "righteous_fury": "retribution",
    "blade_dance": "shadow_flurry",
    "venom_spray": "fan_of_blades",
    "shadow_veil": "smoke_screen",
    "death_blossom": "shadow_flurry",
    "assassinate": "eviscerate",
    "killing_mark": "predators_mark",
    "alpha_call": "black_wolf_companion",
    "pinning_volley": "volley",
    "hunters_volley": "volley",
    "pack_master": "black_wolf_companion",
    "meteor_shot": "explosive_shot",
    "arrow_storm": "rain_of_arrows",
}


def presentation_entry(skill_id: str, skill: dict, template: dict) -> dict:
    entry = deepcopy(template)
    entry["summary_key"] = f"skill.{skill_id}.summary"
    name = skill.get("name", skill_id)
    if "icon" in entry:
        entry["icon"]["label"] = name[0].upper()
    entry["summary"] = f"{name} skill"
    return entry


def merge_class(class_id: str) -> list[str]:
    skills_doc = load_json(SKILLS_PATH)
    i18n = load_json(I18N_PATH)
    pres_doc = load_json(PRES_PATH)
    catalog = skills_doc["skills"]
    new_skills = BUILDERS[class_id](catalog)
    added: list[str] = []
    for skill_id, skill in new_skills.items():
        if skill_id in catalog:
            print(f"skip existing {skill_id}")
            continue
        catalog[skill_id] = skill
        added.append(skill_id)
        i18n["strings"][f"skill.{skill_id}.name"] = skill["name"]
        summary = presentation_entry(
            skill_id, skill, pres_doc["skills"][PRESENTATION_TEMPLATES[skill_id]]
        ).get("summary", skill["name"])
        i18n["strings"][f"skill.{skill_id}.summary"] = summary
        pres_doc["skills"][skill_id] = presentation_entry(
            skill_id, skill, pres_doc["skills"][PRESENTATION_TEMPLATES[skill_id]]
        )
    save_json(SKILLS_PATH, skills_doc)
    save_json(I18N_PATH, i18n)
    save_json(PRES_PATH, pres_doc)
    return added


def main() -> int:
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <class_id|all>")
        return 1
    target = sys.argv[1]
    if target == "all":
        classes = CLASSES
    elif target in CLASSES:
        classes = (target,)
    else:
        print(f"unknown class {target}")
        return 1
    for class_id in classes:
        added = merge_class(class_id)
        print(f"{class_id}: added {len(added)} skills -> {added}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
