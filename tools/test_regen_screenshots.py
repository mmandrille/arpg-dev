"""Tests for data-driven showme screenshot regeneration."""
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

from tools.showme.screenshot_catalog import (
    DEFAULT_SUITES,
    class_ids,
    discover_jobs,
    item_asset_ids,
    item_def_ids_with_visuals,
    item_family_ids,
    skill_ids,
    suite_summary,
)

ROOT = Path(__file__).resolve().parents[1]


def test_default_suites_cover_requested_focus_areas() -> None:
    assert "skeleton" in DEFAULT_SUITES
    assert "gear" in DEFAULT_SUITES
    assert "skill-icon" in DEFAULT_SUITES
    assert "item-icon" in DEFAULT_SUITES
    assert "item-icons" in DEFAULT_SUITES
    assert "floor-item" in DEFAULT_SUITES
    assert "item-asset" in DEFAULT_SUITES


def test_discover_jobs_are_data_driven() -> None:
    jobs = discover_jobs(["skeleton", "gear", "skill-icon"])
    class_set = set(class_ids())
    skill_set = set(skill_ids())
    assert class_set
    assert skill_set
    skeleton_classes = {job.slug for job in jobs if job.suite == "skeleton"}
    gear_classes = {job.slug for job in jobs if job.suite == "gear"}
    skill_slugs = {job.slug for job in jobs if job.suite == "skill-icon"}
    assert skeleton_classes == class_set
    assert gear_classes == class_set
    assert skill_slugs == skill_set


def test_item_jobs_use_shared_catalog_ids() -> None:
    jobs = discover_jobs(["item-icon", "floor-item", "item-asset"])
    families = set(item_family_ids())
    item_defs = set(item_def_ids_with_visuals())
    assets = set(item_asset_ids())
    assert {job.slug for job in jobs if job.suite == "item-icon"} == families
    assert {job.slug for job in jobs if job.suite == "floor-item"} == item_defs
    assert {job.slug for job in jobs if job.suite == "item-asset"}.issubset(assets)


def test_suite_summary_matches_discovered_counts() -> None:
    for row in suite_summary():
        expected = len(discover_jobs([row["suite"]]))
        assert row["count"] == expected


def test_regen_screenshots_dry_run() -> None:
    result = subprocess.run(
        [
            sys.executable,
            "-m",
            "tools.showme.regen_screenshots",
            "--suite",
            "skeleton",
            "--suite",
            "gear",
            "--dry-run",
        ],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert "--focus skeleton" in result.stdout
    assert "--focus gear" in result.stdout


def test_makefile_exposes_regen_screenshots_target() -> None:
    makefile = (ROOT / "make" / "client.mk").read_text(encoding="utf-8")
    assert "regen-screenshots:" in makefile
    assert "tools.showme.regen_screenshots" in makefile
