"""Semantic guard for the shared client UI theme catalog (token references, rarity parity)."""
from __future__ import annotations

import re
from pathlib import Path
from typing import Any

# Literal-token call sites: UiTheme.<kind>("token"); dynamic tokens (concatenation) are not scanned.
_TOKEN_CALL = re.compile(r'\bUiTheme\.(color|spacing|frame|font_size|font_color|rarity_background)\(\s*"([A-Za-z0-9_]+)"\s*[,)]')
_FONT_APPLY = re.compile(r'\bUiTheme\.apply_font\([^,()]+,\s*"([A-Za-z0-9_]+)"\s*\)')

_REQUIRED_SKILL_THEME_TOKENS = {
    "colors": {
        "skill_node_available_text", "skill_node_learned_text", "skill_node_locked_text",
        "skill_node_hover_bg", "skill_node_selected_bg", "skill_node_hover_border", "skill_node_selected_border", "skill_node_focus_border",
        "skill_connector_met", "skill_connector_unmet", "skill_points_available", "skill_points_empty",
        "skill_rank_available", "skill_rank_learned", "skill_rank_locked",
        "skill_icon_locked_modulate", "skill_icon_locked_selected_modulate",
    },
    "spacing": {
        "skill_tree_node_available_border", "skill_tree_node_learned_border", "skill_tree_node_locked_border",
        "skill_tree_node_selected_extra_border", "skill_tree_node_hover_extra_border", "skill_tree_connector_width",
    },
    "fonts": {"skill_points", "skill_node_status", "skill_rank", "skill_tooltip_title", "skill_tooltip_rank", "skill_tooltip_body"},
    "frames": {
        "skill_panel", "skill_tree_surface", "skill_node_available", "skill_node_learned", "skill_node_locked",
        "skill_status_available", "skill_status_learned", "skill_status_locked", "skill_tooltip", "skill_rank_badge",
    },
}


def _catalog_for(theme: dict, kind: str) -> set:
    key = {"color": "colors", "spacing": "spacing", "frame": "frames", "font_size": "fonts",
           "font_color": "fonts", "font": "fonts", "rarity_background": "rarity_slot_backgrounds"}[kind]
    return set(theme[key])


def unresolved_client_tokens(theme: dict, scripts_dir: Path) -> list[str]:
    """Client call sites that reference a UiTheme token missing from the catalog."""
    problems: list[str] = []
    for path in sorted(scripts_dir.glob("*.gd")):
        if path.name == "ui_theme.gd":
            continue
        for number, line in enumerate(path.read_text().splitlines(), 1):
            if line.lstrip().startswith("#"):
                continue
            refs = [(m.group(1), m.group(2)) for m in _TOKEN_CALL.finditer(line)]
            refs += [("font", m.group(1)) for m in _FONT_APPLY.finditer(line)]
            for kind, token in refs:
                if token not in _catalog_for(theme, kind):
                    problems.append(f"{path.name}:{number} UiTheme.{kind} -> unknown token {token}")
    return problems


def _as_list(value: Any) -> list:
    return list(value) if isinstance(value, list) else [value]


def validate_ui_theme(report: Any, theme: dict, item_templates: dict, scripts_dir: Path | None = None) -> None:
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

    missing_skill = [
        f"{section}.{token}"
        for section, tokens in _REQUIRED_SKILL_THEME_TOKENS.items()
        for token in sorted(tokens - set(theme[section]))
    ]
    if missing_skill:
        report.fail("ui theme skill tree tokens", "missing " + ", ".join(missing_skill))
    else:
        report.ok("ui theme defines required skill tree colors, spacing, fonts, and frames")

    if scripts_dir is not None:
        unresolved = unresolved_client_tokens(theme, scripts_dir)
        if unresolved:
            report.fail("ui theme client token references", "; ".join(unresolved))
        else:
            report.ok("client UiTheme literal token references resolve in the catalog")
