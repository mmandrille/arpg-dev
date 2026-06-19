from __future__ import annotations

from datetime import datetime, timezone
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
BOT_RUN_ARTIFACT_DIR = ROOT / ".artifacts" / "bot-runs"


def default_manifest_path() -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    return BOT_RUN_ARTIFACT_DIR / f"{stamp}.json"


def should_clean_bot_run_artifacts(manifest_path: Path) -> bool:
    return manifest_path.parent.resolve() == BOT_RUN_ARTIFACT_DIR.resolve()


def clean_bot_run_artifacts(artifact_dir: Path = BOT_RUN_ARTIFACT_DIR) -> int:
    if not artifact_dir.exists():
        return 0
    removed = 0
    for path in artifact_dir.glob("*.json"):
        if path.is_file():
            path.unlink()
            removed += 1
    return removed


def write_manifest(path: Path, base_url: str, results: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    body = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "scenarios": results,
    }
    path.write_text(json.dumps(body, indent=2) + "\n")
