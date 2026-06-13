from __future__ import annotations

import os
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "check-file-size-ratchet.sh"


def write_lines(path: Path, count: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("x = 1\n" * count)


def run_ratchet(tmp_path: Path, baseline_count: int, line_count: int) -> subprocess.CompletedProcess[str]:
    work = tmp_path / "repo"
    work.mkdir()
    subprocess.run(["git", "init"], cwd=work, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    fixture = work / "tools" / "fixture.py"
    write_lines(fixture, line_count)
    subprocess.run(["git", "add", "tools/fixture.py"], cwd=work, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    baseline = work / ".maintainability" / "file-size-baseline.tsv"
    baseline.parent.mkdir()
    baseline.write_text(f"tools/fixture.py\t{baseline_count}\n")
    env = {
        **os.environ,
        "ROOT": str(work),
        "BASELINE": str(baseline),
        "MAX_LINES": "600",
        "GROWTH_ALLOWANCE": "25",
    }
    return subprocess.run([str(SCRIPT)], cwd=work, env=env, text=True, capture_output=True)


def test_file_size_ratchet_fails_above_baseline_allowance(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=600, line_count=700)

    assert result.returncode != 0
    assert "exceeds grandfathered baseline" in result.stderr


def test_file_size_ratchet_fails_below_baseline_allowance(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=700, line_count=100)

    assert result.returncode != 0
    assert "drop the baseline entry" in result.stderr


def test_file_size_ratchet_passes_within_allowance(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=600, line_count=610)

    assert result.returncode == 0
    assert "file-size ratchet passed" in result.stdout
    assert "grandfathered: 1 files, 610 lines" in result.stdout
