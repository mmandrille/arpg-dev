#!/usr/bin/env python3
"""ADR-0018 D8 asset budget checks for the runtime asset manifest.

Budgets live in ``assets/manifests/asset_budgets.v0.json`` (data, not code):
per asset ``type`` a maximum triangle count and maximum texture edge in pixels.
Over-budget assets fail unless they carry a named exemption. An exemption
that no longer exceeds its budget (or points at a missing asset) also fails,
so the exemption list can only shrink.
"""
from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Protocol

from jsonschema import Draft202012Validator

from tools.assets import glb_reader

BUDGETS_REL = "assets/manifests/asset_budgets.v0.json"
BUDGETS_SCHEMA_REL = "assets/manifests/asset_budgets.v0.schema.json"


class Reporter(Protocol):
    def ok(self, label: str) -> None: ...
    def fail(self, label: str, detail: str) -> None: ...


@dataclass(frozen=True)
class AssetCost:
    triangles: int
    max_texture_px: int


def measure(path: Path) -> AssetCost:
    gltf = glb_reader.load_gltf(path)
    triangles = sum(glb_reader.mesh_triangle_count(gltf, i) for i in range(len(gltf.list("meshes"))))
    edges = [max(img["width"] or 0, img["height"] or 0) for img in glb_reader.image_summaries(gltf)]
    return AssetCost(triangles=triangles, max_texture_px=max(edges, default=0))


def over_budget(cost: AssetCost, budget: dict[str, Any]) -> list[str]:
    problems: list[str] = []
    if cost.triangles > int(budget["max_triangles"]):
        problems.append(f"{cost.triangles} tris > {budget['max_triangles']}")
    if cost.max_texture_px > int(budget["max_texture_px"]):
        problems.append(f"{cost.max_texture_px}px texture > {budget['max_texture_px']}")
    return problems


def check_budgets(root: Path, assets: dict[str, dict[str, Any]], report: Reporter) -> None:
    budgets_path = root / BUDGETS_REL
    if not budgets_path.is_file():
        report.fail("asset budgets", f"missing {BUDGETS_REL}")
        return
    data = json.loads(budgets_path.read_text(encoding="utf-8"))
    schema = json.loads((root / BUDGETS_SCHEMA_REL).read_text(encoding="utf-8"))
    errors = sorted(Draft202012Validator(schema).iter_errors(data), key=lambda e: list(e.path))
    if errors:
        for err in errors:
            report.fail("asset budgets schema", f"{'/'.join(map(str, err.path))}: {err.message}")
        return
    budgets: dict[str, Any] = data["budgets"]
    exemptions: dict[str, str] = data["exemptions"]
    for asset_id in sorted(exemptions):
        if asset_id not in assets:
            report.fail("stale budget exemption", f"{asset_id}: not in manifest")
    for asset_id, entry in sorted(assets.items()):
        budget = budgets.get(entry["type"])
        if budget is None:
            report.fail("asset budgets", f"{asset_id}: no budget for type {entry['type']}")
            continue
        path = root / entry["runtime_path"]
        if not path.is_file():
            continue  # missing files already fail the runtime_path check
        try:
            problems = over_budget(measure(path), budget)
        except (OSError, ValueError) as exc:
            report.fail("asset budgets", f"{asset_id}: cannot measure {exc}")
            continue
        exempt = asset_id in exemptions
        if problems and not exempt:
            report.fail("asset budget", f"{asset_id}: {'; '.join(problems)}")
        elif not problems and exempt:
            report.fail("stale budget exemption", f"{asset_id}: now within budget; remove the exemption")
        elif problems:
            report.ok(f"{asset_id} over budget but exempt ({exemptions[asset_id]})")
        else:
            report.ok(f"{asset_id} within {entry['type']} budget")
