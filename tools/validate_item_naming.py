"""Shared item display-name grammar helpers for validate_shared cross-checks."""

from __future__ import annotations

AFFIX_WORDS: dict[str, tuple[str, int]] = {
    "bonus_cold_damage": ("Freezing", 95),
    "bonus_fire_damage": ("Burning", 95),
    "bonus_lightning_damage": ("Shocking", 95),
    "bonus_poison_damage": ("Venomous", 95),
    "all_skills": ("Arcane", 90),
    "skill_damage_percent": ("Arcane", 90),
    "skill_cooldown_reduction_percent": ("Focused", 85),
    "skill_mana_cost_reduction": ("Focused", 85),
    "crit_chance": ("Keen", 80),
    "hit_chance": ("Keen", 80),
    "attack_speed_percent": ("Keen", 80),
    "damage_min": ("Savage", 70),
    "damage_max": ("Savage", 70),
    "evade_chance": ("Stalwart", 65),
    "block_percent": ("Stalwart", 65),
    "armor": ("Stalwart", 65),
    "max_hp": ("Vigorous", 60),
    "health_regen_per_10_seconds": ("Vigorous", 60),
    "vit": ("Vigorous", 60),
    "max_mana": ("Mystic", 55),
    "mana_regen_per_10_seconds": ("Mystic", 55),
    "magic": ("Mystic", 55),
    "str": ("Mighty", 50),
    "dex": ("Nimble", 50),
    "magic_find_percent": ("Fortunate", 48),
    "inventory_rows": ("Traveler's", 45),
    "hotbar_slots": ("Traveler's", 45),
}

RARITY_RANK = {"common": 0, "magic": 1, "rare": 2, "unique": 3, "set": 3}


def item_rarity_rank(rarity_id: str) -> int:
    return RARITY_RANK.get(rarity_id, -1)


def best_affix_word(template: dict, stats: dict) -> str:
    base = template.get("base_stats", {})
    best_word = ""
    best_priority = -1
    best_gain = 0
    best_stat = ""
    for stat, total in stats.items():
        gain = int(total) - int(base.get(stat, 0))
        if gain <= 0:
            continue
        word, priority = AFFIX_WORDS.get(stat, ("", 0))
        if not word:
            continue
        if priority > best_priority or (
            priority == best_priority and (gain > best_gain or (gain == best_gain and stat < best_stat))
        ):
            best_word = word
            best_priority = priority
            best_gain = gain
            best_stat = stat

    return best_word


def rolled_equipment_display_name(template: dict, rarity_id: str, stats: dict, suffix: str = "") -> str:
    archetype = template["name"]
    name = archetype
    if item_rarity_rank(rarity_id) >= item_rarity_rank("magic"):
        affix = best_affix_word(template, stats)
        if affix:
            name = f"{affix} {archetype}"
    if suffix:
        name = f"{name} {suffix}"

    return name


def validate_shop_offers_golden(report, shop_offers_golden: dict, item_templates: dict, rarities: dict, generated_buy_price) -> None:
    failed_offers = False
    if shop_offers_golden["shop_id"] != "town_vendor":
        report.fail("shop_offers golden", "shop_id must be town_vendor")
        failed_offers = True
    for case in shop_offers_golden["cases"]:
        if failed_offers:
            break
        if len(case["expected"]) != int(case["expected_offer_count"]):
            report.fail("shop_offers golden", f"{case['name']}: expected_offer_count mismatch")
            failed_offers = True
            break
        for offer in case["expected"]:
            template_id = offer["item_template_id"]
            template = item_templates["templates"].get(template_id)
            rarity = offer["rarity"]
            if template is None:
                report.fail("shop_offers golden", f"{case['name']}: unknown template {template_id}")
                failed_offers = True
                break
            if rarity not in rarities:
                report.fail("shop_offers golden", f"{case['name']}: unknown rarity {rarity}")
                failed_offers = True
                break
            expected_name = rolled_equipment_display_name(template, rarity, offer["rolled_stats"])
            if offer["display_name"] != expected_name:
                report.fail(
                    "shop_offers golden",
                    f"{case['name']}: display_name {offer['display_name']!r} != {expected_name!r}",
                )
                failed_offers = True
                break
            if generated_buy_price(template_id, rarity, offer["rolled_stats"]) != int(offer["buy_price"]):
                report.fail("shop_offers golden", f"{case['name']}: buy_price mismatch for {template_id}")
                failed_offers = True
                break
        if failed_offers:
            break
    if not failed_offers:
        report.ok("shop_offers golden matches deterministic catalog")
