from __future__ import annotations

from pathlib import Path

from tools.validate_fog_presentation import (
    validate_camera_fog_mode_alignment,
    validate_fog_presentation_ranges,
)


class CapturingReport:
    def __init__(self) -> None:
        self.ok_labels: list[str] = []
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.ok_labels.append(label)

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def _assets_dir() -> Path:
    return Path(__file__).resolve().parents[1] / "shared" / "assets"


def _load_json(path: Path) -> dict:
    import json

    return json.loads(path.read_text(encoding="utf-8"))


def test_fog_presentation_ranges_pass_on_committed_assets() -> None:
    report = CapturingReport()
    validate_fog_presentation_ranges(report, _load_json, _assets_dir())
    assert report.failures == []
    assert any("falloff_power" in label for label in report.ok_labels)


def test_camera_fog_mode_alignment_passes_on_committed_assets() -> None:
    report = CapturingReport()
    validate_camera_fog_mode_alignment(report, _load_json, _assets_dir())
    assert report.failures == []
    assert any("isometric" in label for label in report.ok_labels)


def test_camera_fog_mode_alignment_fails_when_isometric_mode_missing() -> None:
    report = CapturingReport()
    assets = _assets_dir()

    def load_json(path: Path) -> dict:
        if path.name == "camera_presentations.v0.json":
            return {"modes": {}}
        return _load_json(path)

    validate_camera_fog_mode_alignment(report, load_json, assets)
    assert any("isometric" in failure for failure in report.failures)
