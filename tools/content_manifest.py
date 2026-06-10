"""Helpers for the shared content-library manifest."""
from __future__ import annotations

from dataclasses import dataclass
import json
from pathlib import Path
from typing import Any


class ManifestError(ValueError):
    """Raised when manifest-listed content cannot be merged safely."""


@dataclass(frozen=True)
class ManifestEntry:
    group: str
    path: str


def load_json(path: Path) -> Any:
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


def skill_rule_entries(manifest: dict[str, Any]) -> list[ManifestEntry]:
    return _entries(manifest.get("rules", {}).get("skills", []))


def skill_presentation_entries(manifest: dict[str, Any]) -> list[ManifestEntry]:
    return _entries(manifest.get("assets", {}).get("skills", {}).get("presentations", []))


def merge_catalog_files(
    manifest_path: Path,
    entries: list[ManifestEntry],
    collection_key: str,
) -> dict[str, Any]:
    merged: dict[str, Any] = {}
    for entry in entries:
        path = resolve_manifest_path(manifest_path, entry)
        data = load_json(path)
        collection = data.get(collection_key)
        if not isinstance(collection, dict):
            raise ManifestError(f"{entry.path}: missing object '{collection_key}'")
        for content_id, definition in collection.items():
            if content_id in merged:
                raise ManifestError(f"duplicate {collection_key} id {content_id} in {entry.path}")
            merged[content_id] = definition
    return merged


def resolve_manifest_path(manifest_path: Path, entry: ManifestEntry) -> Path:
    raw = Path(entry.path)
    if raw.is_absolute():
        raise ManifestError(f"{entry.path}: manifest paths must be relative")
    path = (manifest_path.parent / raw).resolve()
    if not path.exists():
        raise ManifestError(f"{entry.path}: listed file does not exist")
    return path


def _entries(raw_entries: Any) -> list[ManifestEntry]:
    if not isinstance(raw_entries, list):
        raise ManifestError("manifest entry list must be an array")
    entries: list[ManifestEntry] = []
    for raw in raw_entries:
        if not isinstance(raw, dict):
            raise ManifestError("manifest entry must be an object")
        group = raw.get("group")
        path = raw.get("path")
        if not isinstance(group, str) or not group:
            raise ManifestError("manifest entry group must be a non-empty string")
        if not isinstance(path, str) or not path:
            raise ManifestError("manifest entry path must be a non-empty string")
        entries.append(ManifestEntry(group=group, path=path))
    return entries
