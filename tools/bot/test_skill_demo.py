from __future__ import annotations

import pytest

from tools.bot.skill_demo import all_skill_demo_entries, skill_demo_entry


def test_skill_demo_entries_cover_current_skill_kinds() -> None:
    entries = {entry.skill_id: entry for entry in all_skill_demo_entries()}

    assert entries["magic_bolt"].category == "attack"
    assert entries["magic_bolt"].class_id == "sorcerer"
    assert entries["magic_bolt"].icon_label == "M"
    assert entries["magic_bolt"].rank_targets == [1, 5]

    assert entries["ligthing"].category == "attack"
    assert entries["ligthing"].class_id == "sorcerer"
    assert entries["ligthing"].icon_label == "L"

    assert entries["rage"].category == "self_buff"
    assert entries["rage"].class_id == "barbarian"
    assert entries["rage"].icon_shape == "burst"

    assert entries["earthbreaker"].category == "attack"
    assert entries["earthbreaker"].class_id == "barbarian"
    assert entries["earthbreaker"].icon_label == "E"

    assert entries["heal"].category == "heal"
    assert entries["heal"].class_id == "paladin"
    assert entries["heal"].icon_shape == "heart"

    assert entries["holy_shield"].category == "stat_buff"
    assert entries["holy_shield"].class_id == "paladin"
    assert entries["holy_shield"].icon_shape == "shield"

    assert entries["shadow_flurry"].category == "attack"
    assert entries["shadow_flurry"].class_id == "rogue"
    assert entries["shadow_flurry"].icon_shape == "slash"

    assert entries["split_arrow"].category == "attack"
    assert entries["split_arrow"].class_id == "ranger"
    assert entries["split_arrow"].icon_label == "S"


def test_single_skill_demo_entry_has_display_metadata() -> None:
    entry = skill_demo_entry("holy_shield")

    assert entry.name == "Holy Shield"
    assert entry.kind == "area_stat_buff"
    assert entry.max_rank == 5
    assert entry.targeting == "self_or_ally_area"
    assert entry.icon_color == "#f0c23d"
    assert entry.icon_accent == "#fff7b0"
    assert entry.summary == "Area ally defense"


def test_unknown_skill_fails_clearly() -> None:
    with pytest.raises(KeyError, match="unknown skill_id 'meteor'"):
        skill_demo_entry("meteor")
