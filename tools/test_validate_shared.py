from __future__ import annotations

import json
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

from tools.content_manifest import ManifestError, merge_catalog_files, skill_rule_entries


ROOT = Path(__file__).resolve().parent.parent
MANIFEST_SCHEMA = ROOT / "shared/content/content_libraries.v0.schema.json"


def load(path: Path):
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


def test_content_manifest_schema_rejects_unknown_top_level_group() -> None:
    manifest = {
        "version": 0,
        "rules": {"skills": [{"group": "mage", "path": "../rules/skills.v0.json"}]},
        "assets": {
            "skills": {
                "presentations": [{"group": "default", "path": "../assets/skill_presentations.v0.json"}]
            }
        },
        "runtime": {},
    }
    errors = list(Draft202012Validator(load(MANIFEST_SCHEMA)).iter_errors(manifest))
    assert any("Additional properties" in error.message and "runtime" in error.message for error in errors)


def test_content_manifest_merge_rejects_duplicate_skill_ids(tmp_path: Path) -> None:
    content_dir = tmp_path / "content"
    rules_dir = tmp_path / "rules"
    content_dir.mkdir()
    rules_dir.mkdir()
    (rules_dir / "a.v0.json").write_text('{"version": 0, "skills": {"magic_bolt": {}}}\n', encoding="utf-8")
    (rules_dir / "b.v0.json").write_text('{"version": 0, "skills": {"magic_bolt": {}}}\n', encoding="utf-8")
    manifest_path = content_dir / "content_libraries.v0.json"
    manifest_path.write_text(
        json.dumps(
            {
                "version": 0,
                "rules": {
                    "skills": [
                        {"group": "a", "path": "../rules/a.v0.json"},
                        {"group": "b", "path": "../rules/b.v0.json"},
                    ]
                },
                "assets": {"skills": {"presentations": [{"group": "default", "path": "../rules/a.v0.json"}]}},
            }
        )
        + "\n",
        encoding="utf-8",
    )

    with pytest.raises(ManifestError, match="duplicate skills id magic_bolt"):
        merge_catalog_files(manifest_path, skill_rule_entries(load(manifest_path)), "skills")
