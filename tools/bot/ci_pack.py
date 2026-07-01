"""CI scenario pack: fast integration subset for ``make ci``."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

CI_PACK_PATH = Path(__file__).resolve().parent / "ci_pack.json"
PROTOCOL_SCENARIOS_DIR = Path(__file__).resolve().parent / "scenarios"
CLIENT_SCENARIOS_DIR = PROTOCOL_SCENARIOS_DIR / "client"
CI_SELECTOR = "ci"
EXTENDED_TIER = "extended"
BENCHMARK_TIER = "benchmark"


def load_ci_pack() -> dict[str, Any]:
    raw = json.loads(CI_PACK_PATH.read_text())
    protocol = [str(item) for item in raw.get("protocol", [])]
    client = [str(item) for item in raw.get("client", [])]
    if not protocol or not client:
        raise ValueError(f"{CI_PACK_PATH}: protocol and client lists must be non-empty")
    return {"protocol": protocol, "client": client}


def protocol_pack_ids() -> list[str]:
    return list(load_ci_pack()["protocol"])


def client_pack_ids() -> list[str]:
    return list(load_ci_pack()["client"])


def all_pack_ids() -> set[str]:
    pack = load_ci_pack()
    return set(pack["protocol"]) | set(pack["client"])


def _scenario_id(path: Path) -> str:
    raw = json.loads(path.read_text())
    scenario_id = str(raw.get("id", "")).strip()
    if not scenario_id:
        raise ValueError(f"{path}: missing scenario id")
    return scenario_id


def _scenario_raw(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text())


def list_protocol_scenario_paths() -> list[Path]:
    return sorted(PROTOCOL_SCENARIOS_DIR.glob("*.json"))


def list_client_scenario_paths() -> list[Path]:
    return sorted(CLIENT_SCENARIOS_DIR.glob("*.json"))


def resolve_client_pack_paths() -> list[Path]:
    wanted = client_pack_ids()
    by_id = {_scenario_id(path): path for path in list_client_scenario_paths()}
    missing = [scenario_id for scenario_id in wanted if scenario_id not in by_id]
    if missing:
        raise ValueError(f"ci client pack missing scenario file(s): {', '.join(missing)}")
    return [by_id[scenario_id] for scenario_id in wanted]


def select_pack_scenarios(scenarios: list[Any], kind: str) -> list[Any]:
    if kind not in {"protocol", "client"}:
        raise ValueError(f"unknown ci pack kind: {kind}")
    wanted = load_ci_pack()[kind]
    by_id = {scenario.id: scenario for scenario in scenarios}
    missing = [scenario_id for scenario_id in wanted if scenario_id not in by_id]
    if missing:
        raise ValueError(f"ci {kind} pack missing loaded scenario(s): {', '.join(missing)}")
    return [by_id[scenario_id] for scenario_id in wanted]


# Intentional protocol/client pairs that share the same scenario id string.
CROSS_TREE_SCENARIO_PAIRS: dict[str, str] = {
    "mystery_seller_core": "client/24_mystery_seller_core.json",
    "mystery_seller_paid_reroll": "client/29_mystery_seller_paid_reroll.json",
    "quest_town_turn_in": "client/75_quest_town_turn_in.json",
    "shop_stock_lifecycle": "client/22_shop_stock_lifecycle.json",
    "unique_burn_effect_live": "client/33_unique_burn_effect_live.json",
}


def validate_cross_tree_scenario_pairs() -> None:
    protocol_by_id = {_scenario_id(path): path for path in list_protocol_scenario_paths()}
    client_by_id = {_scenario_id(path): path for path in list_client_scenario_paths()}
    for scenario_id, client_suffix in CROSS_TREE_SCENARIO_PAIRS.items():
        if scenario_id not in protocol_by_id:
            raise ValueError(f"cross-tree pair missing protocol scenario id: {scenario_id}")
        client_path = CLIENT_SCENARIOS_DIR / Path(client_suffix).name
        if not client_path.exists():
            raise ValueError(f"cross-tree pair missing client scenario file: {client_path}")
        if _scenario_id(client_path) != scenario_id:
            raise ValueError(
                f"cross-tree pair id mismatch for {scenario_id}: client file {client_path} "
                f"has id {_scenario_id(client_path)!r}"
            )
    unexpected = (set(protocol_by_id) & set(client_by_id)) - set(CROSS_TREE_SCENARIO_PAIRS)
    if unexpected:
        raise ValueError(
            "unexpected duplicate scenario ids across protocol/client trees "
            f"(register in CROSS_TREE_SCENARIO_PAIRS or rename): {sorted(unexpected)}"
        )


def validate_ci_pack() -> None:
    validate_cross_tree_scenario_pairs()
    pack = load_ci_pack()
    protocol_paths = list_protocol_scenario_paths()
    client_paths = list_client_scenario_paths()
    protocol_ids = {_scenario_id(path) for path in protocol_paths}
    client_ids = {_scenario_id(path) for path in client_paths}
    overlap = set(pack["protocol"]) & set(pack["client"])
    if overlap:
        raise ValueError(f"ci pack ids appear in both protocol and client lists: {sorted(overlap)}")
    missing_protocol = [scenario_id for scenario_id in pack["protocol"] if scenario_id not in protocol_ids]
    if missing_protocol:
        raise ValueError(f"ci protocol pack unknown id(s): {', '.join(missing_protocol)}")
    missing_client = [scenario_id for scenario_id in pack["client"] if scenario_id not in client_ids]
    if missing_client:
        raise ValueError(f"ci client pack unknown id(s): {', '.join(missing_client)}")
    pack_ids = all_pack_ids()
    unassigned: list[str] = []
    for path in protocol_paths + client_paths:
        raw = _scenario_raw(path)
        scenario_id = str(raw.get("id", "")).strip()
        if scenario_id in pack_ids:
            if raw.get("ci_tier") == EXTENDED_TIER:
                raise ValueError(f"{path}: ci pack member must not set ci_tier=extended")
            continue
        if raw.get("ci_tier") not in {EXTENDED_TIER, BENCHMARK_TIER}:
            unassigned.append(scenario_id)
    if unassigned:
        raise ValueError(
            "scenario(s) must be listed in tools/bot/ci_pack.json or set "
            f'"ci_tier": "{EXTENDED_TIER}" or "{BENCHMARK_TIER}": {", ".join(sorted(unassigned))}'
        )
