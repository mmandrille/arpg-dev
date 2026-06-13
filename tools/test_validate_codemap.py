from __future__ import annotations

import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VALIDATOR = ROOT / "tools" / "validate_codemap.py"


def test_validate_codemap_passes_real_codemap() -> None:
    result = subprocess.run(
        [sys.executable, str(VALIDATOR)],
        cwd=ROOT,
        text=True,
        capture_output=True,
    )

    assert result.returncode == 0, result.stderr
    assert "codemap ok" in result.stdout


def test_validate_codemap_reports_missing_path(tmp_path: Path) -> None:
    codemap = tmp_path / "CODEMAP.md"
    codemap.write_text(
        "| Domain | Server | Client | Shared rules / protocol | Bot | Tests / Migrations |\n"
        "|--------|--------|--------|-------------------------|-----|--------------------|\n"
        "| Missing | `server/internal/nope.go` |  |  |  |  |\n"
    )

    result = subprocess.run(
        [sys.executable, str(VALIDATOR), str(codemap)],
        cwd=ROOT,
        text=True,
        capture_output=True,
    )

    assert result.returncode != 0
    assert "missing path: server/internal/nope.go" in result.stderr
