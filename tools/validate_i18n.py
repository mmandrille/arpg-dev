"""Shared localization catalog validation helpers."""
from __future__ import annotations

from typing import Any


def validate_i18n_catalog(
    report: Any,
    english_text: dict[str, Any],
    skills: dict[str, Any],
    skill_presentations: dict[str, Any],
    monsters: dict[str, Any],
) -> None:
    english_strings = english_text.get("strings", {})
    if english_text.get("locale") == "en":
        report.ok("english text catalog locale is en")
    else:
        report.fail("english text catalog locale", f"expected en, got {english_text.get('locale')!r}")

    for key, value in sorted(english_strings.items()):
        if isinstance(key, str) and key.strip() == key and key and isinstance(value, str) and value.strip():
            report.ok(f"english text {key}")
        else:
            report.fail("english text catalog entry", f"invalid key/value {key!r}: {value!r}")

    for skill_id, skill in sorted(skills.get("skills", {}).items()):
        _require_key(report, english_strings, "skill name_key", f"skill {skill_id} name_key resolves", skill_id, skill)
    for skill_id, presentation in sorted(skill_presentations.get("skills", {}).items()):
        key = presentation.get("summary_key")
        if key in english_strings:
            report.ok(f"skill {skill_id} summary_key resolves")
        else:
            report.fail("skill summary_key", f"{skill_id} references missing English text key {key!r}")
    for monster_id, monster in sorted(monsters.get("monsters", {}).items()):
        _require_key(report, english_strings, "monster name_key", f"monster {monster_id} name_key resolves", monster_id, monster)


def validate_locale_catalog(report: Any, locale_text: dict[str, Any], english_text: dict[str, Any]) -> None:
    locale = str(locale_text.get("locale", ""))
    strings = locale_text.get("strings", {})
    english_strings = english_text.get("strings", {})
    if locale:
        report.ok(f"text catalog {locale} declares locale")
    else:
        report.fail("text catalog locale", "missing locale")
    for key, value in sorted(strings.items()):
        if key in english_strings and isinstance(value, str) and value.strip():
            report.ok(f"text catalog {locale} {key}")
        elif key not in english_strings:
            report.fail("text catalog key", f"{locale} has unknown key {key!r}")
        else:
            report.fail("text catalog entry", f"{locale} has empty value for {key!r}")


def _require_key(
    report: Any,
    english_strings: dict[str, Any],
    label: str,
    ok_label: str,
    content_id: str,
    content: dict[str, Any],
) -> None:
    key = content.get("name_key")
    if key in english_strings:
        report.ok(ok_label)
    else:
        report.fail(label, f"{content_id} references missing English text key {key!r}")
