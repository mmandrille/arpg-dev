"""Render the current skill logo metadata into a labeled SVG contact sheet."""

from __future__ import annotations

import argparse
import html
import json
import math
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[2]
RULES_PATH = ROOT / "shared/rules/skills.v0.json"
PRESENTATIONS_PATH = ROOT / "shared/assets/skill_presentations.v0.json"
DEFAULT_OUT = ROOT / ".artifacts/skill-logo-sheet.svg"

TILE_W = 220
TILE_H = 190
ICON_SIZE = 88
MARGIN = 28
GAP = 18
TITLE_H = 56
COLS = 3


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--out", default=str(DEFAULT_OUT), help="SVG output path")
    args = parser.parse_args()

    rules = _read_json(RULES_PATH).get("skills", {})
    presentations = _read_json(PRESENTATIONS_PATH).get("skills", {})
    skill_ids = sorted(rules, key=lambda skill_id: _tree_sort_key(skill_id, rules[skill_id]))
    if not skill_ids:
        raise SystemExit("no skills found")

    rows = math.ceil(len(skill_ids) / COLS)
    width = MARGIN * 2 + COLS * TILE_W + (COLS - 1) * GAP
    height = MARGIN * 2 + TITLE_H + rows * TILE_H + (rows - 1) * GAP

    parts = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">',
        "<style>",
        "text{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}",
        ".title{font-size:22px;font-weight:700;fill:#eef2f6}",
        ".name{font-size:17px;font-weight:700;fill:#f6f0e3}",
        ".id{font-size:12px;fill:#9aa7b6}",
        "</style>",
        f'<rect width="{width}" height="{height}" fill="#111318"/>',
        f'<text class="title" x="{MARGIN}" y="38">Current skill logos - shared/assets/skill_presentations.v0.json</text>',
    ]

    for index, skill_id in enumerate(skill_ids):
        parts.append(_tile_svg(index, skill_id, rules[skill_id], presentations.get(skill_id, {})))

    parts.append("</svg>\n")
    out = Path(args.out).expanduser()
    if not out.is_absolute():
        out = ROOT / out
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text("\n".join(parts), encoding="utf-8")
    print(f"[skill-logo-sheet] wrote {out}")
    return 0


def _read_json(path: Path) -> dict[str, Any]:
    with path.open(encoding="utf-8") as handle:
        parsed = json.load(handle)
    if not isinstance(parsed, dict):
        raise SystemExit(f"{path} is not a JSON object")
    return parsed


def _tree_sort_key(skill_id: str, skill: dict[str, Any]) -> tuple[int, int, str]:
    tree = skill.get("tree", {})
    if not isinstance(tree, dict):
        tree = {}
    return (int(tree.get("tier", 999)), int(tree.get("column", 999)), skill_id)


def _tile_svg(index: int, skill_id: str, skill: dict[str, Any], presentation: dict[str, Any]) -> str:
    col = index % COLS
    row = index // COLS
    x = MARGIN + col * (TILE_W + GAP)
    y = MARGIN + TITLE_H + row * (TILE_H + GAP)
    icon = presentation.get("icon", {})
    if not isinstance(icon, dict):
        icon = {}
    shape = str(icon.get("shape", "bolt"))
    fill = str(icon.get("color", "#62b7ff"))
    accent = str(icon.get("accent", "#e8f7ff"))
    name = str(skill.get("name", skill_id))
    cx = x + TILE_W / 2
    cy = y + 62
    radius = ICON_SIZE * 0.42

    return "\n".join(
        [
            f'<g transform="translate({x},{y})">',
            f'<rect width="{TILE_W}" height="{TILE_H}" rx="8" fill="#1d222b" stroke="#445063"/>',
            "</g>",
            f'<circle cx="{cx:.2f}" cy="{cy:.2f}" r="{radius:.2f}" fill="#040403" opacity="0.92"/>',
            _shape_svg(shape, cx, cy, radius, fill, accent),
            f'<circle cx="{cx:.2f}" cy="{cy:.2f}" r="{radius:.2f}" fill="none" stroke="{html.escape(accent)}" stroke-width="2"/>',
            f'<text class="name" x="{cx:.2f}" y="{y + 135}" text-anchor="middle">{html.escape(name)}</text>',
            f'<text class="id" x="{cx:.2f}" y="{y + 171}" text-anchor="middle">{html.escape(skill_id)}</text>',
        ]
    )


def _shape_svg(shape: str, cx: float, cy: float, radius: float, fill: str, accent: str) -> str:
    if shape == "burst":
        points = []
        for i in range(12):
            r = radius * (0.95 if i % 2 == 0 else 0.48)
            angle = (-math.pi * 0.5) + (math.tau * i / 12.0)
            points.append((cx + math.cos(angle) * r, cy + math.sin(angle) * r))
        return "\n".join(
            [
                _polygon(points, fill, accent),
                f'<circle cx="{cx:.2f}" cy="{cy:.2f}" r="{radius * 0.34:.2f}" fill="{html.escape(accent)}"/>',
            ]
        )
    if shape == "heart":
        points = []
        for i in range(34):
            t = math.tau * i / 34.0
            x = 16.0 * math.sin(t) ** 3
            y = -(13.0 * math.cos(t) - 5.0 * math.cos(2.0 * t) - 2.0 * math.cos(3.0 * t) - math.cos(4.0 * t))
            points.append((cx + x * (radius / 18.0), cy + y * (radius / 18.0)))
        return _polygon(points, fill, accent)
    if shape == "slash":
        upper = (cx + radius * 0.58, cy - radius * 0.68)
        lower = (cx - radius * 0.58, cy + radius * 0.68)
        points = [
            (upper[0] + radius * 0.12, upper[1] + radius * 0.02),
            (cx + radius * 0.12, cy + radius * 0.12),
            (lower[0] - radius * 0.06, lower[1] + radius * 0.14),
            (cx - radius * 0.16, cy - radius * 0.06),
            (upper[0] + radius * 0.02, upper[1] - radius * 0.18),
        ]
        line = (
            f'<line x1="{cx - radius * 0.42:.2f}" y1="{cy - radius * 0.44:.2f}" '
            f'x2="{cx + radius * 0.44:.2f}" y2="{cy + radius * 0.42:.2f}" '
            f'stroke="{html.escape(accent)}" stroke-width="2" stroke-linecap="round"/>'
        )
        return _polygon(points, fill, accent) + "\n" + line
    points = [
        (cx - radius * 0.10, cy - radius * 0.82),
        (cx + radius * 0.34, cy - radius * 0.18),
        (cx + radius * 0.08, cy - radius * 0.18),
        (cx + radius * 0.22, cy + radius * 0.82),
        (cx - radius * 0.38, cy + radius * 0.04),
        (cx - radius * 0.10, cy + radius * 0.04),
    ]
    return _polygon(points, fill, accent)


def _polygon(points: list[tuple[float, float]], fill: str, accent: str) -> str:
    encoded = " ".join(f"{x:.2f},{y:.2f}" for x, y in points)
    return f'<polygon points="{encoded}" fill="{html.escape(fill)}" stroke="{html.escape(accent)}" stroke-width="2"/>'


if __name__ == "__main__":
    raise SystemExit(main())
