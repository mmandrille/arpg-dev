from __future__ import annotations

import copy

from tools.validate_unique_items import validate_unique_items_catalog


class CapturingReport:
    def __init__(self) -> None:
        self.ok_labels: list[str] = []
        self.failures: list[str] = []

    def ok(self, label: str) -> None:
        self.ok_labels.append(label)

    def fail(self, label: str, detail: str) -> None:
        self.failures.append(f"{label}: {detail}")


def base_catalogs() -> tuple[dict, dict, dict]:
    unique_items = {
        "uniques": {
            "embercall_blade": {
                "id": "embercall_blade",
                "enabled": True,
                "base_template_id": "cave_blade",
                "display_name": "Embercall Blade",
                "minimum_level": 1,
                "fixed_stats": {"damage_min": 4},
                "fixed_effect_ids": ["everburning_wound"],
                "behavior_hook": "Applies live effect behavior.",
                "status": "ready",
            },
            "stormstring_bow": {
                "id": "stormstring_bow",
                "enabled": True,
                "base_template_id": "cave_bow",
                "display_name": "Stormstring Bow",
                "minimum_level": 1,
                "fixed_stats": {"damage_max": 6},
                "fixed_effect_ids": ["stormbound_echo"],
                "behavior_hook": "Applies live stormbound_echo behavior.",
                "status": "ready",
            }
        }
    }
    item_templates = {"templates": {"cave_blade": {"item_type": "sword"}, "cave_bow": {"item_type": "bow"}}}
    unique_effects = {
        "effects": {
            "everburning_wound": {
                "id": "everburning_wound",
                "enabled": True,
                "status": "ready",
                "compatible_item_types": ["sword"],
            },
            "stormbound_echo": {
                "id": "stormbound_echo",
                "enabled": True,
                "status": "ready",
                "compatible_item_types": ["bow"],
            },
            "shield_only": {
                "id": "shield_only",
                "enabled": True,
                "status": "ready",
                "compatible_item_types": ["shield"],
            },
            "disabled_effect": {
                "id": "disabled_effect",
                "enabled": False,
                "status": "disabled_seed",
                "compatible_item_types": ["sword"],
            },
        }
    }
    return unique_items, item_templates, unique_effects


def run_validation(unique_items: dict, item_templates: dict | None = None, unique_effects: dict | None = None) -> CapturingReport:
    base_unique_items, base_templates, base_effects = base_catalogs()
    report = CapturingReport()
    validate_unique_items_catalog(
        report,
        unique_items or base_unique_items,
        item_templates or base_templates,
        unique_effects or base_effects,
    )
    return report


def test_unique_items_valid_catalog_reports_ok() -> None:
    report = run_validation(base_catalogs()[0])

    assert report.failures == []
    assert report.ok_labels == ["unique_items entries reference valid templates and behavior hooks"]


def test_unique_items_rejects_empty_catalog() -> None:
    report = run_validation({"uniques": {}})

    assert report.failures == ["unique_items catalog: must define at least one unique seed"]


def test_unique_items_rejects_bad_identity_and_template() -> None:
    unique_items, _, _ = base_catalogs()
    bad = copy.deepcopy(unique_items)
    unique = bad["uniques"]["embercall_blade"]
    unique["id"] = "wrong"
    unique["base_template_id"] = "missing_template"

    report = run_validation(bad)

    assert any("unique_items id" in failure and "must match key" in failure for failure in report.failures)
    assert any("unique_items base template" in failure and "unknown template" in failure for failure in report.failures)


def test_unique_items_rejects_enabled_and_disabled_status_mismatches() -> None:
    unique_items, _, _ = base_catalogs()
    enabled_bad = copy.deepcopy(unique_items)
    enabled_bad["uniques"]["embercall_blade"]["status"] = "disabled_seed"

    disabled_bad = copy.deepcopy(unique_items)
    unique = disabled_bad["uniques"]["embercall_blade"]
    unique["enabled"] = False
    unique["status"] = "ready"

    assert any("enabled entries must be ready" in failure for failure in run_validation(enabled_bad).failures)
    assert any("disabled entries must remain disabled_seed" in failure for failure in run_validation(disabled_bad).failures)


def test_unique_items_rejects_weak_hook_and_missing_fixed_stats() -> None:
    unique_items, _, _ = base_catalogs()
    bad = copy.deepcopy(unique_items)
    unique = bad["uniques"]["embercall_blade"]
    unique["behavior_hook"] = "Only grants stats."
    unique["fixed_stats"] = {}

    report = run_validation(bad)

    assert any("behavior hook" in failure and "must describe effect" in failure for failure in report.failures)
    assert any("fixed_stats" in failure and "enabled entries must define fixed stats" in failure for failure in report.failures)


def test_unique_items_rejects_duplicate_unknown_inactive_and_incompatible_effects() -> None:
    cases = [
        (["everburning_wound", "everburning_wound"], "duplicate effect ids"),
        (["missing_effect"], "unknown effect missing_effect"),
        (["disabled_effect"], "inactive effect disabled_effect"),
        (["shield_only"], "effect shield_only incompatible with cave_blade type sword"),
    ]
    for effect_ids, message in cases:
        unique_items, _, _ = base_catalogs()
        bad = copy.deepcopy(unique_items)
        bad["uniques"]["embercall_blade"]["fixed_effect_ids"] = effect_ids

        report = run_validation(bad)

        assert any(message in failure for failure in report.failures), (effect_ids, report.failures)


def test_unique_items_rejects_missing_effects_for_enabled_unique() -> None:
    unique_items, _, _ = base_catalogs()
    bad = copy.deepcopy(unique_items)
    bad["uniques"]["embercall_blade"]["fixed_effect_ids"] = []

    report = run_validation(bad)

    assert any("enabled entries must define at least one effect" in failure for failure in report.failures)
