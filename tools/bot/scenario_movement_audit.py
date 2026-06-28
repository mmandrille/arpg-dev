"""Movement-step audit for bot scenarios (v358 scenario-movement-decoupling)."""

from __future__ import annotations

import csv
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from tools.bot.ci_pack import all_pack_ids

ROOT = Path(__file__).resolve().parent.parent.parent
SCENARIOS_DIR = Path(__file__).resolve().parent / "scenarios"
AUDIT_TSV_PATH = ROOT / "docs" / "progress" / "scenario-movement-audit.tsv"

# v358 refactor snapshot: movement_steps_before for scenarios touched this slice.
MOVEMENT_STEPS_BEFORE_BASELINE: dict[str, int] = {
    "21_monster_rarity_loot_scaling.json": 2,
    "42_pack_aggro_and_dungeon_packs.json": 2,
    "77_elite_minion_pack_ai.json": 1,
    "14_dungeon_monsters.json": 2,
    "68_dungeon_elite_side_objective.json": 2,
    "17_treasure_classes_and_guarded_chests.json": 1,
    "55_survival_reactive_unique_effects.json": 1,
    "65_random_quest_reward_floor.json": 1,
    "72_upgrade_resource_drop.json": 1,
    "32_skill_points_and_magic_bolt.json": 1,
    "103_dungeon_combat_perf_probe.json": 2,
    "88_mercenary_hiring_board.json": 1,
    "89_companion_stance_command.json": 1,
    "96_mercenary_offer_variants.json": 1,
    "61_purple_town_unique_chest.json": 1,
    "client/39_blacksmith_upgrade_ui.json": 2,
    "client/55_blacksmith_recipe_selector.json": 2,
    "client/62_blacksmith_second_recipe.json": 2,
    "client/63_blacksmith_upgrade_history.json": 2,
    "client/61_material_wallet_window.json": 2,
    "client/54_material_wallet_details.json": 2,
    "client/70_blacksmith_armor_recipe.json": 4,
    "client/24_mystery_seller_core.json": 2,
    "client/29_mystery_seller_paid_reroll.json": 2,
    "client/53_mystery_seller_silhouettes.json": 2,
    "client/79_wall_floor_dungeon_rollout.json": 2,
}

MOVEMENT_CONTRACT_ALLOWLIST: frozenset[str] = frozenset(
    {
        "path_maze",
        "click_to_move",
        "town_floor_click_to_move",
        "chase_lab",
        "chase_maze",
        "leash_lab",
        "dungeon_levels",
        "teleporter_lab",
        "reachable_dungeon_obstacles",
        "collision_lab",
        "player_path_budget_lab",
        "town_teleporter_auto_approach",
        "attack_move_sticky_targeting",
        "movement_visual_smoothing",
        "entity_tick_smoothing",
        "mobility_skill_smoothing",
        "melee_lunge_micro_step",
        "torch_walk_visual",
        "flying_navigation_trait",
        "boss_floor_gate",
        "vertical_slice",
        "gear_before_combat",
        "line_of_sight_blockers",
        "fog_of_war_radius",
        "companion_ai_foundation",
        "resource_support_mobility_unique_effects",
    }
)

PROTOCOL_MOVEMENT_ACTIONS = frozenset(
    {
        "use_stair",
        "walk_to_loot",
        "walk_to_monster",
        "move_until_in_range",
        "move_until_player_position",
        "teleport_to_level",
    }
)

CLIENT_MOVEMENT_STEP_TYPES = frozenset({"click_floor", "wait_player_near"})

TSV_COLUMNS = [
    "scenario_path",
    "scenario_id",
    "runner",
    "ci_tier",
    "movement_class",
    "movement_steps_before",
    "movement_steps_after",
    "action",
    "merged_into_or_reason",
]


@dataclass(frozen=True)
class ScenarioAuditRow:
    scenario_path: str
    scenario_id: str
    runner: str
    ci_tier: str
    movement_class: str
    movement_steps_before: int
    movement_steps_after: int
    action: str
    merged_into_or_reason: str

    def as_dict(self) -> dict[str, str]:
        return {
            "scenario_path": self.scenario_path,
            "scenario_id": self.scenario_id,
            "runner": self.runner,
            "ci_tier": self.ci_tier,
            "movement_class": self.movement_class,
            "movement_steps_before": str(self.movement_steps_before),
            "movement_steps_after": str(self.movement_steps_after),
            "action": self.action,
            "merged_into_or_reason": self.merged_into_or_reason,
        }


def discover_scenario_paths() -> list[Path]:
    protocol = sorted(SCENARIOS_DIR.glob("*.json"))
    client = sorted((SCENARIOS_DIR / "client").glob("*.json"))
    return protocol + client


def _is_movement_protocol_step(step: dict[str, Any]) -> bool:
    action = str(step.get("action", "")).strip()
    if action in PROTOCOL_MOVEMENT_ACTIONS:
        return True

    return action.startswith("walk_to_")


def _count_protocol_movement_steps(raw: dict[str, Any]) -> int:
    steps = raw.get("steps", [])
    if not isinstance(steps, list):
        return 0

    count = 0
    for step in steps:
        if isinstance(step, dict) and _is_movement_protocol_step(step):
            count += 1

    for block in raw.get("fresh_session_checks", []) or []:
        if not isinstance(block, dict):
            continue
        block_steps = block.get("steps", [])
        if not isinstance(block_steps, list):
            continue
        for step in block_steps:
            if isinstance(step, dict) and _is_movement_protocol_step(step):
                count += 1

    return count


def _count_client_movement_steps(raw: dict[str, Any]) -> int:
    steps = raw.get("client_steps", [])
    if not isinstance(steps, list):
        return 0

    count = 0
    for step in steps:
        if not isinstance(step, dict):
            continue
        step_type = str(step.get("type", "")).strip()
        if step_type in CLIENT_MOVEMENT_STEP_TYPES:
            count += 1

    return count


def count_movement_steps(path: Path, raw: dict[str, Any] | None = None) -> int:
    data = raw if raw is not None else json.loads(path.read_text(encoding="utf-8"))
    runner = str(data.get("runner", "protocol")).strip() or "protocol"
    if runner == "godot_client":
        return _count_client_movement_steps(data)

    return _count_protocol_movement_steps(data)


def ci_tier_for(raw: dict[str, Any], scenario_id: str) -> str:
    if scenario_id in all_pack_ids():
        return "pack"
    tier = str(raw.get("ci_tier", "")).strip()
    return tier if tier else "pack"


def classify_movement(scenario_id: str, movement_steps: int, movement_class_override: str = "") -> str:
    if movement_class_override in {"deleted", "delete-candidate"}:
        return movement_class_override
    if scenario_id in MOVEMENT_CONTRACT_ALLOWLIST:
        return "contract"
    if movement_steps == 0:
        return "contract"
    return "setup-eliminable"


def scenario_relative_path(path: Path) -> str:
    return str(path.relative_to(SCENARIOS_DIR))


def audit_scenario_file(path: Path, overrides: dict[str, dict[str, str]] | None = None) -> ScenarioAuditRow:
    raw = json.loads(path.read_text(encoding="utf-8"))
    scenario_path = scenario_relative_path(path)
    scenario_id = str(raw.get("id", "")).strip()
    if not scenario_id:
        raise ValueError(f"{path}: missing scenario id")

    runner = str(raw.get("runner", "protocol")).strip() or "protocol"
    steps_now = count_movement_steps(path, raw)
    override = (overrides or {}).get(scenario_path, {})
    movement_class = classify_movement(
        scenario_id,
        steps_now,
        str(override.get("movement_class", "")).strip(),
    )
    steps_before = int(
        override.get(
            "movement_steps_before",
            MOVEMENT_STEPS_BEFORE_BASELINE.get(scenario_path, steps_now),
        )
        or steps_now
    )
    steps_after = int(override.get("movement_steps_after", steps_now) or steps_now)

    return ScenarioAuditRow(
        scenario_path=scenario_path,
        scenario_id=scenario_id,
        runner=runner,
        ci_tier=ci_tier_for(raw, scenario_id),
        movement_class=movement_class,
        movement_steps_before=steps_before,
        movement_steps_after=steps_after,
        action=str(override.get("action", "")).strip(),
        merged_into_or_reason=str(override.get("merged_into_or_reason", "")).strip(),
    )


def load_audit_tsv_overrides(path: Path = AUDIT_TSV_PATH) -> dict[str, dict[str, str]]:
    if not path.is_file():
        return {}

    overrides: dict[str, dict[str, str]] = {}
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        for row in reader:
            scenario_path = str(row.get("scenario_path", "")).strip()
            scenario_id = str(row.get("scenario_id", "")).strip()
            key = scenario_path or scenario_id
            if not key:
                raise ValueError(f"{path}: blank scenario_path/scenario_id row")
            if key in overrides:
                raise ValueError(f"{path}: duplicate audit row {key}")
            overrides[key] = {col: str(row.get(col, "")).strip() for col in TSV_COLUMNS}

    return overrides


def build_audit_rows(overrides: dict[str, dict[str, str]] | None = None) -> list[ScenarioAuditRow]:
    merged_overrides = load_audit_tsv_overrides()
    if overrides:
        merged_overrides.update(overrides)

    rows: list[ScenarioAuditRow] = []
    for path in discover_scenario_paths():
        rows.append(audit_scenario_file(path, merged_overrides))

    deleted_rows = [
        ScenarioAuditRow(
            scenario_path=data.get("scenario_path", scenario_id),
            scenario_id=scenario_id,
            runner=data.get("runner", ""),
            ci_tier=data.get("ci_tier", ""),
            movement_class="deleted",
            movement_steps_before=int(data.get("movement_steps_before", 0) or 0),
            movement_steps_after=0,
            action=data.get("action", "deleted"),
            merged_into_or_reason=data.get("merged_into_or_reason", ""),
        )
        for scenario_id, data in merged_overrides.items()
        if data.get("movement_class") == "deleted"
        and scenario_id not in {row.scenario_id for row in rows}
        and data.get("scenario_path", scenario_id) not in {row.scenario_path for row in rows}
    ]
    rows.extend(deleted_rows)
    rows.sort(key=lambda row: row.scenario_path)
    return rows


def write_audit_tsv(path: Path = AUDIT_TSV_PATH, preserve_before: bool = True) -> list[ScenarioAuditRow]:
    existing = load_audit_tsv_overrides() if preserve_before else {}
    rows = build_audit_rows()
    finalized: list[ScenarioAuditRow] = []

    for row in rows:
        prev = existing.get(row.scenario_path, {})
        if preserve_before and prev.get("movement_steps_before"):
            finalized.append(
                ScenarioAuditRow(
                    scenario_path=row.scenario_path,
                    scenario_id=row.scenario_id,
                    runner=row.runner,
                    ci_tier=row.ci_tier,
                    movement_class=row.movement_class,
                    movement_steps_before=int(prev["movement_steps_before"]),
                    movement_steps_after=row.movement_steps_after,
                    action=prev.get("action", row.action) or row.action,
                    merged_into_or_reason=prev.get("merged_into_or_reason", row.merged_into_or_reason),
                )
            )
        elif preserve_before:
            finalized.append(
                ScenarioAuditRow(
                    scenario_path=row.scenario_path,
                    scenario_id=row.scenario_id,
                    runner=row.runner,
                    ci_tier=row.ci_tier,
                    movement_class=row.movement_class,
                    movement_steps_before=row.movement_steps_after,
                    movement_steps_after=row.movement_steps_after,
                    action=row.action,
                    merged_into_or_reason=row.merged_into_or_reason,
                )
            )
        else:
            finalized.append(row)

    rows = finalized

    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=TSV_COLUMNS, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        for row in rows:
            writer.writerow(row.as_dict())

    return rows


def validate_audit_tsv(path: Path = AUDIT_TSV_PATH) -> None:
    if not path.is_file():
        raise FileNotFoundError(f"missing audit TSV: {path}")

    rows_by_path: dict[str, dict[str, str]] = {}
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        for row in reader:
            scenario_path = str(row.get("scenario_path", "")).strip()
            scenario_id = str(row.get("scenario_id", "")).strip()
            key = scenario_path or scenario_id
            if not key:
                raise ValueError(f"{path}: blank scenario_path/scenario_id row")
            if key in rows_by_path:
                raise ValueError(f"{path}: duplicate audit row {key}")
            rows_by_path[key] = row

    discovered_paths = {scenario_relative_path(p) for p in discover_scenario_paths()}
    tsv_paths = {
        str(row.get("scenario_path", "")).strip()
        for row in rows_by_path.values()
        if row.get("movement_class") != "deleted"
    }
    tsv_paths.discard("")
    missing = sorted(discovered_paths - tsv_paths)
    if missing:
        raise ValueError(f"audit TSV missing scenario row(s): {', '.join(missing)}")

    extra = sorted(tsv_paths - discovered_paths)
    if extra:
        raise ValueError(f"audit TSV orphan scenario row(s): {', '.join(extra)}")

    for scenario_id in MOVEMENT_CONTRACT_ALLOWLIST:
        matching = [
            row
            for row in rows_by_path.values()
            if row.get("scenario_id") == scenario_id and row.get("movement_class") != "deleted"
        ]
        for row in matching:
            if row.get("movement_class") != "contract":
                raise ValueError(f"allowlist scenario {scenario_id} must be movement_class=contract")


def main() -> None:
    rows = write_audit_tsv()
    print(f"wrote {len(rows)} rows to {AUDIT_TSV_PATH}")


if __name__ == "__main__":
    main()
