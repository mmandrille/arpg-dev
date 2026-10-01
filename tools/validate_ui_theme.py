"""Semantic guard for the shared client UI theme catalog (token references, rarity parity)."""
from __future__ import annotations

from typing import Any


def _as_list(value: Any) -> list:
    return list(value) if isinstance(value, list) else [value]


def validate_ui_theme(report: Any, theme: dict, item_templates: dict) -> None:
    colors = set(theme["colors"])
    spacing = set(theme["spacing"])
    problems: list[str] = []

    for frame_name, frame in theme["frames"].items():
        for key in ("bg", "border"):
            if frame[key] not in colors:
                problems.append(f"frame {frame_name}.{key} -> unknown color {frame[key]}")
        for key in ("border_width", "radius", "margin"):
            for ref in _as_list(frame.get(key, [])):
                if ref not in spacing:
                    problems.append(f"frame {frame_name}.{key} -> unknown spacing {ref}")
        for state_name, state in frame.get("states", {}).items():
            for key, ref in state.items():
                if ref not in colors:
                    problems.append(f"frame {frame_name}.states.{state_name}.{key} -> unknown color {ref}")
    for role, font in theme["fonts"].items():
        for key in ("color", "outline_color"):
            if key in font and font[key] not in colors:
                problems.append(f"font {role}.{key} -> unknown color {font[key]}")
    for key in ("invalid_border", "invalid_border_hover"):
        if theme["state_modifiers"][key] not in colors:
            problems.append(f"state_modifiers.{key} -> unknown color {theme['state_modifiers'][key]}")
    if problems:
        report.fail("ui theme references", "; ".join(problems))
    else:
        report.ok("ui theme token references resolve")

    expected = set(item_templates["rarities"])
    actual = set(theme["rarity_slot_backgrounds"])
    if actual != expected:
        report.fail("ui theme rarity coverage", f"missing={sorted(expected - actual)}, extra={sorted(actual - expected)}")
    else:
        report.ok("ui theme rarity backgrounds cover exactly the gameplay rarity keys")
    values = [str(v).lower() for v in theme["rarity_slot_backgrounds"].values()]
    if len(values) != len(set(values)):
        report.fail("ui theme rarity distinct", "duplicate rarity slot background")
    else:
        report.ok("ui theme rarity backgrounds are distinct")

    missing = []
    for rarity in sorted(expected):
        if f"inventory_rarity_{rarity}" not in colors:
            missing.append(f"colors.inventory_rarity_{rarity}")
        if f"inventory_border_{rarity}" not in spacing:
            missing.append(f"spacing.inventory_border_{rarity}")
    if "inventory_border_common" not in colors:
        missing.append("colors.inventory_border_common")
    if missing:
        report.fail("ui theme inventory rarity tokens", "missing " + ", ".join(missing))
    else:
        report.ok("ui theme defines inventory rarity color and border tokens for every rarity")
