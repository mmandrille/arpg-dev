"""Shared dungeon density helpers for validators."""

from __future__ import annotations

import math
from typing import Protocol


class DensityReport(Protocol):
    def fail(self, label: str, detail: str) -> None: ...


def area_density_count(formula: dict, floor_size: dict) -> int:
    area_per_unit = float(formula.get("area_per_unit", 0))
    if area_per_unit <= 0:
        return 0
    area = float(floor_size.get("width", 0)) * float(floor_size.get("height", 0))
    raw = int(math.floor(area / area_per_unit + 0.5))
    return max(int(formula.get("min", 0)), min(int(formula.get("max", 0)), raw))


def validate_area_count_formula(report: DensityReport, label: str, formula: dict) -> bool:
    ok = True
    if float(formula.get("area_per_unit", 0)) <= 0:
        report.fail(label, "area_per_unit must be positive")
        ok = False
    if int(formula.get("min", -1)) < 0 or int(formula.get("max", -1)) < int(formula.get("min", 0)):
        report.fail(label, "min/max range is invalid")
        ok = False
    return ok


def validate_area_range_formula(report: DensityReport, label: str, formula: dict) -> bool:
    ok = validate_area_count_formula(report, label, formula)
    if int(formula.get("spread", -1)) < 0:
        report.fail(label, "spread must be non-negative")
        ok = False
    return ok
