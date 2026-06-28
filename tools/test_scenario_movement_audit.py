import json
from pathlib import Path

from tools.bot.scenario_movement_audit import (
    MOVEMENT_CONTRACT_ALLOWLIST,
    AUDIT_TSV_PATH,
    build_audit_rows,
    count_movement_steps,
    discover_scenario_paths,
    validate_audit_tsv,
)


def test_every_scenario_file_has_audit_row():
    validate_audit_tsv()


def test_allowlist_scenarios_classified_contract():
    rows = {row.scenario_id: row for row in build_audit_rows()}
    for scenario_id in MOVEMENT_CONTRACT_ALLOWLIST:
        path_protocol = Path("tools/bot/scenarios")
        path_client = path_protocol / "client"
        exists = any(
            json.loads(p.read_text()).get("id") == scenario_id
            for p in list(path_protocol.glob("*.json")) + list(path_client.glob("*.json"))
        )
        if exists:
            assert rows[scenario_id].movement_class == "contract"


def test_audit_tsv_columns_present():
    header = AUDIT_TSV_PATH.read_text(encoding="utf-8").splitlines()[0]
    assert "movement_steps_before" in header
    assert "movement_steps_after" in header
    assert "movement_class" in header


def test_movement_count_matches_json_for_sample():
    path = Path("tools/bot/scenarios/12_dungeon_levels.json")
    raw = json.loads(path.read_text(encoding="utf-8"))
    assert count_movement_steps(path, raw) == 3


def test_discover_includes_protocol_and_client():
    paths = discover_scenario_paths()
    assert any("/client/" in str(p) for p in paths)
    assert any(p.name.endswith(".json") and "/client/" not in str(p) for p in paths)
