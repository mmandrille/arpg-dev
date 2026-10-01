from __future__ import annotations

import copy
import json
from pathlib import Path

from tools.validate_boss_presentation import validate_boss_presentation

ROOT = Path(__file__).resolve().parents[1]
SHARED = ROOT / "shared"
MANIFEST = ROOT / "assets" / "manifests" / "assets.v0.json"
SCENES = ROOT / "client" / "scenes"


class CapturingReport:
    def __init__(self) -> None:
        self.ok_labels: list[str] = []
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.ok_labels.append(label)

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def _load(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def _run(mutate=None) -> CapturingReport:
    """Validate with an optionally mutated copy of boss_presentation.v0.json."""
    pres_path = SHARED / "assets" / "boss_presentation.v0.json"
    cfg = copy.deepcopy(_load(pres_path))
    if mutate:
        mutate(cfg)

    def load(path: Path) -> dict:
        return cfg if path == pres_path else _load(path)

    report = CapturingReport()
    validate_boss_presentation(report, load, SHARED, MANIFEST, SCENES)
    return report


def test_committed_catalog_passes() -> None:
    report = _run()
    assert report.failures == []
    assert report.ok_labels


def test_missing_boss_entry_fails() -> None:
    report = _run(lambda cfg: cfg["bosses"].pop("crypt_matron"))
    assert any("crypt_matron has no presentation" in f for f in report.failures)


def test_unknown_attachment_asset_fails() -> None:
    def mutate(cfg):
        cfg["bosses"]["cave_warden"]["attachments"][0]["asset_id"] = "no_such_asset_v0"

    assert any("no_such_asset_v0" in f for f in _run(mutate).failures)


def test_unknown_bone_fails() -> None:
    def mutate(cfg):
        cfg["bosses"]["cave_warden"]["headgear"]["bone"] = "not_a_bone"

    assert any("not_a_bone" in f for f in _run(mutate).failures)


def test_non_positive_scale_fails() -> None:
    def mutate(cfg):
        cfg["bosses"]["cave_warden"]["scale"] = 0

    assert any("scale must be positive" in f for f in _run(mutate).failures)


def test_banner_duration_out_of_range_fails() -> None:
    def mutate(cfg):
        cfg["bosses"]["cave_warden"]["banner"]["duration_s"] = 9.0

    assert any("duration_s=9.0" in f for f in _run(mutate).failures)


def test_close_scales_fail() -> None:
    def mutate(cfg):
        cfg["bosses"]["crypt_matron"]["scale"] = cfg["bosses"]["cave_warden"]["scale"] * 1.05

    assert any("differ by less than" in f for f in _run(mutate).failures)


def test_unknown_visual_key_fails() -> None:
    def mutate(cfg):
        cfg["bosses"]["cave_warden"]["visual_key"] = "monster_kit_nope"

    assert any("monster_kit_nope" in f for f in _run(mutate).failures)
