from __future__ import annotations

from tools.validate_dungeon_goldens import validate_dungeon_obstacle_goldens


class CapturingReport:
    def __init__(self) -> None:
        self.ok_labels: list[str] = []
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.ok_labels.append(label)

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def _generation(*, water_max=2, hole_max=1, water_enabled=True, hole_enabled=True, weights=None) -> dict:
    return {
        "floor_size": {"width": 100, "height": 50},
        "obstacle_generation": {
            "solid_kind_weights": weights if weights is not None else {"wall": 5, "rock": 2, "column": 2, "rubble": 2},
            "water": {"enabled": water_enabled, "target_count": {"min": 1, "max": water_max}},
            "holes": {"enabled": hole_enabled, "target_count": {"min": 1, "max": hole_max}},
        },
    }


def _golden(*, min_water=2, min_hole=1, solid_kinds=("column", "rock", "rubble"), level=-1) -> dict:
    return {
        "level": level,
        "expected": {
            "floor_size": {"width": 100, "height": 50},
            "shape_families": ["block", "l", "line"],
            "minimum_generated_wall_count": 12,
            "minimum_water_count": min_water,
            "minimum_hole_count": min_hole,
            "solid_kinds": list(solid_kinds),
            "walls": [{"source": "generated", "shape_family": "block"}],
        },
    }


def test_consistent_golden_passes() -> None:
    report = CapturingReport()
    validate_dungeon_obstacle_goldens(report, _generation(), _golden())
    assert report.failures == []
    assert any("achievable" in label for label in report.ok_labels)
    assert any("wall-layout contract" in label for label in report.ok_labels)


def test_water_floor_exceeding_generation_cap_fails() -> None:
    report = CapturingReport()
    validate_dungeon_obstacle_goldens(report, _generation(water_max=1), _golden(min_water=2))
    assert any("minimum_water_count" in f and "exceeds" in f for f in report.failures)


def test_floor_demanding_disabled_feature_fails() -> None:
    report = CapturingReport()
    validate_dungeon_obstacle_goldens(report, _generation(hole_enabled=False), _golden(min_hole=1))
    assert any("enabled is false" in f for f in report.failures)


def test_unknown_solid_kind_fails() -> None:
    report = CapturingReport()
    validate_dungeon_obstacle_goldens(report, _generation(weights={"wall": 5, "rock": 2}), _golden(solid_kinds=("column",)))
    assert any("solid_kinds" in f and "column" in f for f in report.failures)


def test_non_generated_level_fails() -> None:
    report = CapturingReport()
    validate_dungeon_obstacle_goldens(report, _generation(), _golden(level=3))
    assert any("generated dungeon floor" in f for f in report.failures)
