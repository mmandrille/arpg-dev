"""Focused tests for the live payload schema gate (tools/bot/payload_schema.py)."""
from __future__ import annotations

import copy
import io
import json
import subprocess
import sys
from pathlib import Path
from typing import Callable

import pytest
from jsonschema import Draft202012Validator

from tools.bot import payload_schema
from tools.bot.bot_context import StateIngestContext
from tools.bot.bot_types import RuntimeState
from tools.bot.payload_schema import (
    PayloadSchemaError,
    PayloadSchemaGate,
    compiled_schema,
    payload_violations,
)

ROOT = Path(__file__).resolve().parents[2]
EXAMPLES = ROOT / "shared" / "protocol" / "examples"
MESSAGE_TYPES = ("session_snapshot", "state_delta")


def example(message_type: str) -> dict:
    return json.loads((EXAMPLES / f"{message_type}.json").read_text(encoding="utf-8"))


def plain_validator(message_type: str) -> Draft202012Validator:
    path = payload_schema.PROTOCOL_DIR / payload_schema.SCHEMA_FILES[message_type]
    return Draft202012Validator(json.loads(path.read_text(encoding="utf-8")))


def first_index(items: list[dict], key: str, value: str) -> int:
    return next(i for i, item in enumerate(items) if item.get(key) == value)


def test_module_imports_without_run() -> None:
    code = "import sys, tools.bot.payload_schema; assert 'tools.bot.run' not in sys.modules"
    subprocess.run([sys.executable, "-c", code], cwd=ROOT, check=True)


@pytest.mark.parametrize("message_type", MESSAGE_TYPES)
def test_committed_examples_are_clean(message_type: str) -> None:
    assert payload_violations(message_type, example(message_type)) == []


def test_unknown_message_type_is_not_validated() -> None:
    assert payload_violations("intent_accepted", {"anything": object()}) == []


def test_compiled_schema_is_cached_and_dispatches_item_arrays() -> None:
    assert compiled_schema("state_delta") is compiled_schema("state_delta")
    delta = compiled_schema("state_delta")
    snapshot = compiled_schema("session_snapshot")
    assert delta is not None and snapshot is not None
    # Losing dispatch silently would bring back the ~5x slower oneOf/allOf fan-out.
    assert delta.dispatch["changes"].field == "op"
    assert delta.dispatch["events"].field == "event_type"
    assert snapshot.dispatch["recent_events"].field == "event_type"


def test_off_contract_entity_prop_reports_generic_key_and_concrete_path() -> None:
    delta = example("state_delta")
    index = first_index(delta["changes"], "op", "entity_update")
    delta["changes"][index]["entity"]["not_in_contract"] = 1

    violations = payload_violations("state_delta", delta)

    assert len(violations) == 1
    key, detail = violations[0]
    assert key.startswith("state_delta changes/*/entity [op=entity_update]: Additional properties")
    assert "'not_in_contract'" in key
    assert f"changes/{index}/entity" in detail


def mutations() -> list[tuple[str, str, Callable[[dict], None]]]:
    def unknown_op(p: dict) -> None:
        p["changes"][0]["op"] = "not_an_op"

    def entity_update_missing_entity(p: dict) -> None:
        del p["changes"][first_index(p["changes"], "op", "entity_update")]["entity"]

    def damage_wrong_type(p: dict) -> None:
        p["events"][first_index(p["events"], "event_type", "monster_damaged")]["damage"] = "three"

    def event_type_required_field_missing(p: dict) -> None:
        del p["events"][first_index(p["events"], "event_type", "monster_killed")]["target_entity_id"]

    def enum_branch_required_field_missing(p: dict) -> None:
        del p["events"][first_index(p["events"], "event_type", "player_damaged")]["damage"]

    def top_level_extra(p: dict) -> None:
        p["unexpected_top_level"] = True

    def snapshot_event_wrong_type(p: dict) -> None:
        p["recent_events"].append({"event_type": "monster_damaged", "damage": -1})

    return [
        ("state_delta", "unknown_op", unknown_op),
        ("state_delta", "entity_update_missing_entity", entity_update_missing_entity),
        ("state_delta", "damage_wrong_type", damage_wrong_type),
        ("state_delta", "event_type_required_field_missing", event_type_required_field_missing),
        ("state_delta", "enum_branch_required_field_missing", enum_branch_required_field_missing),
        ("state_delta", "top_level_extra", top_level_extra),
        ("session_snapshot", "top_level_extra", top_level_extra),
        ("session_snapshot", "snapshot_event_wrong_type", snapshot_event_wrong_type),
    ]


@pytest.mark.parametrize(("message_type", "name", "mutate"), mutations(), ids=lambda v: v if isinstance(v, str) else "")
def test_dispatch_agrees_with_plain_validator(message_type: str, name: str, mutate) -> None:
    payload = copy.deepcopy(example(message_type))
    mutate(payload)

    plain_errors = list(plain_validator(message_type).iter_errors(payload))
    dispatched = payload_violations(message_type, payload)

    # Every mutation is off-contract; dispatch must agree and land inside the items the
    # plain validator flagged.
    assert plain_errors and dispatched, name
    plain_roots = {tuple(error.absolute_path)[:2] for error in plain_errors}
    for _, detail in dispatched:
        location = detail.split(" at ", 1)[1].split(": ", 1)[0].split(" [", 1)[0]
        parts = tuple(int(p) if p.isdigit() else p for p in location.split("/")) if location != "<root>" else ()
        assert parts[:2] in plain_roots or parts[:1] in plain_roots or parts == (), (name, detail, plain_roots)


def test_unknown_event_type_matches_plain_verdict() -> None:
    # An event_type no allOf rule names is only held to the base event properties,
    # exactly as the plain validator does (no if matches, so no then applies).
    delta = example("state_delta")
    delta["events"] = [{"event_type": "never_emitted"}]
    assert list(plain_validator("state_delta").iter_errors(delta)) == []
    assert payload_violations("state_delta", delta) == []


def bad_delta() -> dict:
    delta = example("state_delta")
    delta["changes"][first_index(delta["changes"], "op", "entity_update")]["entity"]["drift"] = 1
    return delta


def test_gate_off_ignores_violations() -> None:
    PayloadSchemaGate("off").check("state_delta", bad_delta())


def test_gate_strict_raises_assertion_error() -> None:
    with pytest.raises(PayloadSchemaError, match="schema violation: state_delta at changes/"):
        PayloadSchemaGate("strict").check("state_delta", bad_delta())
    assert issubclass(PayloadSchemaError, AssertionError)


def test_gate_report_logs_each_distinct_violation_once() -> None:
    stream = io.StringIO()
    gate = PayloadSchemaGate("report", stream=stream)

    gate.check("state_delta", bad_delta())
    gate.check("state_delta", bad_delta())

    lines = stream.getvalue().splitlines()
    assert len(lines) == 1 and lines[0].startswith("schema violation: state_delta at changes/")
    assert len(gate.seen) == 1


def test_gate_rejects_unknown_mode() -> None:
    with pytest.raises(ValueError, match="ARPG_BOT_SCHEMA_VALIDATION"):
        PayloadSchemaGate("loud")


def test_state_ingest_routes_deltas_and_snapshots_through_gate(monkeypatch: pytest.MonkeyPatch) -> None:
    from tools.bot.state_ingest import ingest_message, ingest_snapshot

    seen: list[str] = []

    class RecordingGate:
        def check(self, message_type: str, payload: object) -> None:
            seen.append(message_type)

    monkeypatch.setattr(payload_schema, "default_gate", RecordingGate)
    ctx = StateIngestContext(log=lambda *_a: None)

    ingest_snapshot(example("session_snapshot"), RuntimeState(), ctx=ctx)
    ingest_message({"type": "state_delta", "tick": 1, "payload": example("state_delta")}, RuntimeState(), ctx=ctx)

    assert seen == ["session_snapshot", "state_delta"]


def test_state_ingest_strict_gate_fails_off_contract_delta(monkeypatch: pytest.MonkeyPatch) -> None:
    from tools.bot.state_ingest import ingest_message

    monkeypatch.setattr(payload_schema, "default_gate", lambda: PayloadSchemaGate("strict"))
    with pytest.raises(PayloadSchemaError):
        ingest_message(
            {"type": "state_delta", "tick": 1, "payload": bad_delta()},
            RuntimeState(),
            ctx=StateIngestContext(log=lambda *_a: None),
        )


def test_value_echoing_rules_dedupe_by_rule_not_by_value() -> None:
    # A pattern/type/bounds message echoes the offending value; two different bad values
    # of the same drift must collapse to one report key.
    keys = set()
    for bad_level in ("one", "two"):
        delta = example("state_delta")
        delta["level"] = bad_level
        keys.update(key for key, _ in payload_violations("state_delta", delta))
    assert keys == {'state_delta level: type="integer"'}


def test_one_of_leaf_reports_from_the_branch_the_item_was_aiming_at() -> None:
    # A visible shop offer with one drifted field must report that field, not every
    # mystery-offer requirement from the other oneOf branch.
    delta = example("state_delta")
    offer = {"offer_id": "fixed:minor_health_potion", "kind": "fixed", "item_def_id": "minor_health_potion",
             "display_name": "Minor Health Potion", "buy_price": 5, "unexpected": True}
    delta["events"] = [{"event_type": "shop_opened", "entity_id": "1001", "shop_id": "town_vendor", "offers": [offer]}]
    details = [detail for _, detail in payload_violations("state_delta", delta)]
    assert len(details) == 1 and "'unexpected' was unexpected" in details[0], details


@pytest.mark.parametrize("message_type", MESSAGE_TYPES)
def test_every_local_ref_resolves(message_type: str) -> None:
    # v486 found state_delta events referencing #/$defs/equipped and #/$defs/hotbar_slot, which
    # only the snapshot schema defined; any payload carrying them crashed validation.
    import re

    path = payload_schema.PROTOCOL_DIR / payload_schema.SCHEMA_FILES[message_type]
    schema = json.loads(path.read_text(encoding="utf-8"))
    refs = set(re.findall(r'"\$ref": "#/\$defs/([a-z0-9_]+)"', json.dumps(schema)))
    assert refs - set(schema["$defs"]) == set()


def test_broken_schema_is_reported_not_raised(monkeypatch: pytest.MonkeyPatch) -> None:
    broken = payload_schema.CompiledPayloadSchema({
        "type": "object",
        "properties": {"thing": {"$ref": "#/$defs/missing"}},
        "$defs": {},
    })
    monkeypatch.setattr(payload_schema, "compiled_schema", lambda _t: broken)

    violations = payload_violations("state_delta", {"thing": 1})

    assert len(violations) == 1 and "schema error" in violations[0][1]
    with pytest.raises(PayloadSchemaError, match="schema error"):
        PayloadSchemaGate("strict").check("state_delta", {"thing": 1})


def test_suspended_pauses_and_restores_the_process_gate(monkeypatch: pytest.MonkeyPatch) -> None:
    gate = PayloadSchemaGate("strict")
    monkeypatch.setattr(payload_schema, "default_gate", lambda: gate)

    with payload_schema.suspended():
        payload_schema.check_payload("state_delta", bad_delta())  # perf window: not validated
    assert gate.mode == "strict"
    with pytest.raises(PayloadSchemaError):
        payload_schema.check_payload("state_delta", bad_delta())
