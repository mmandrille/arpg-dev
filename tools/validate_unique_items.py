"""Validation helpers for named unique item catalog rules."""

from __future__ import annotations

from typing import Any, Protocol


class ValidationReport(Protocol):
    def ok(self, label: str) -> None: ...

    def fail(self, label: str, detail: str) -> None: ...


def validate_unique_items_catalog(
    report: ValidationReport,
    unique_items: dict[str, Any],
    item_templates: dict[str, Any],
    unique_effects: dict[str, Any],
) -> None:
    unique_defs = unique_items.get("uniques", {})
    if not unique_defs:
        report.fail("unique_items catalog", "must define at least one unique seed")
        return

    failed_uniques = False
    templates = item_templates.get("templates", {})
    effects = unique_effects.get("effects", {})
    for unique_id, unique in unique_defs.items():
        if unique.get("id") != unique_id:
            report.fail("unique_items id", f"{unique_id}: id field must match key")
            failed_uniques = True
        template_id = unique.get("base_template_id")
        template = templates.get(template_id)
        if template is None:
            report.fail("unique_items base template", f"{unique_id}: unknown template {template_id}")
            failed_uniques = True
        enabled = unique.get("enabled")
        status = unique.get("status")
        if enabled is True and status != "ready":
            report.fail("unique_items status", f"{unique_id}: enabled entries must be ready")
            failed_uniques = True
        elif enabled is False and status != "disabled_seed":
            report.fail("unique_items status", f"{unique_id}: disabled entries must remain disabled_seed")
            failed_uniques = True
        hook = str(unique.get("behavior_hook", "")).lower()
        if "effect" not in hook and "behavior" not in hook:
            report.fail("unique_items behavior hook", f"{unique_id}: must describe effect or behavior ownership, not only stats")
            failed_uniques = True
        fixed_stats = unique.get("fixed_stats", {})
        if not isinstance(fixed_stats, dict):
            report.fail("unique_items fixed_stats", f"{unique_id}: fixed_stats must be an object")
            failed_uniques = True
        elif enabled is True and not fixed_stats:
            report.fail("unique_items fixed_stats", f"{unique_id}: enabled entries must define fixed stats")
            failed_uniques = True
        fixed_effect_ids = unique.get("fixed_effect_ids", [])
        if not isinstance(fixed_effect_ids, list):
            report.fail("unique_items fixed_effect_ids", f"{unique_id}: fixed_effect_ids must be a list")
            failed_uniques = True
        elif len(fixed_effect_ids) != len(set(fixed_effect_ids)):
            report.fail("unique_items fixed_effect_ids", f"{unique_id}: duplicate effect ids")
            failed_uniques = True
        elif enabled is True and not fixed_effect_ids:
            report.fail("unique_items fixed_effect_ids", f"{unique_id}: enabled entries must define at least one effect")
            failed_uniques = True
        elif template is not None:
            item_type = template.get("item_type")
            for effect_id in fixed_effect_ids:
                effect = effects.get(effect_id)
                if effect is None:
                    report.fail("unique_items fixed_effect_ids", f"{unique_id}: unknown effect {effect_id}")
                    failed_uniques = True
                    continue
                if not effect.get("enabled") or effect.get("status") != "ready":
                    report.fail("unique_items fixed_effect_ids", f"{unique_id}: inactive effect {effect_id}")
                    failed_uniques = True
                if item_type not in effect.get("compatible_item_types", []):
                    report.fail(
                        "unique_items fixed_effect_ids",
                        f"{unique_id}: effect {effect_id} incompatible with {template_id} type {item_type}",
                    )
                    failed_uniques = True
    if not failed_uniques:
        report.ok("unique_items entries reference valid templates and behavior hooks")
