from __future__ import annotations

import os
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "check-extraction-coupling-ratchet.py"
COUPLED_CALL = "helpers" + "=globals()"


def write_file(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body)


def run_ratchet(tmp_path: Path, baseline_count: int | None, body: str) -> subprocess.CompletedProcess[str]:
    work = tmp_path / "repo"
    work.mkdir()
    subprocess.run(["git", "init"], cwd=work, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    fixture = work / "tools" / "fixture.py"
    write_file(fixture, body)
    subprocess.run(["git", "add", "tools/fixture.py"], cwd=work, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    baseline = work / ".maintainability" / "extraction-coupling-baseline.tsv"
    baseline.parent.mkdir()
    if baseline_count is None:
        baseline.write_text("")
    else:
        baseline.write_text(f"tools/fixture.py\t{baseline_count}\n")
    env = {
        **os.environ,
        "ROOT": str(work),
        "EXTRACTION_COUPLING_BASELINE": str(baseline),
    }
    return subprocess.run(["python3", str(SCRIPT)], cwd=work, env=env, text=True, capture_output=True)


def test_extraction_coupling_ratchet_passes_at_baseline(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=1, body=f"run({COUPLED_CALL})\n")

    assert result.returncode == 0
    assert "extraction-coupling ratchet passed" in result.stdout


def test_extraction_coupling_ratchet_fails_when_count_grows(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=1, body=f"run({COUPLED_CALL})\nrun({COUPLED_CALL})\n")

    assert result.returncode != 0
    assert "exceeds baseline" in result.stderr


def test_extraction_coupling_ratchet_fails_when_baseline_stale(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=2, body=f"run({COUPLED_CALL})\n")

    assert result.returncode != 0
    assert "lower or remove the extraction coupling baseline" in result.stderr


def test_extraction_coupling_ratchet_blocks_unbaselined_coupling(tmp_path: Path) -> None:
    result = run_ratchet(tmp_path, baseline_count=None, body=f"run({COUPLED_CALL})\n")

    assert result.returncode != 0
    assert "unbaselined" in result.stderr
