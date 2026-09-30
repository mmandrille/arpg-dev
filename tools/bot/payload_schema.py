"""Live wire-contract gate: validate received protocol payloads against the v8 schemas.

`tools/validate_shared.py` only checks the hand-written examples under
`shared/protocol/examples/`. This module checks what the server actually sends, so an
off-contract field (a prop the schema's `additionalProperties: false` rejects, a wrong
type, a missing required key) fails the protocol bot instead of shipping silently.

Mode comes from ``ARPG_BOT_SCHEMA_VALIDATION``:

* ``off`` (default) — no validation; interactive/benchmark runs pay nothing.
* ``report`` — log each distinct violation once and keep going (drift survey).
* ``strict`` — raise ``PayloadSchemaError`` on the first violating payload
  (``make ci`` protocol scenarios).

Cost: a plain ``Draft202012Validator`` tries every ``change`` against all ``oneOf``
branches and every ``event`` against every ``allOf`` ``if``, which costs ~20 ms for a
busy delta. ``CompiledPayloadSchema`` compiles once per message type and dispatches each
array item on its discriminator (``op`` / ``event_type``) to a validator that holds only
the branches that can apply. That is equivalent: a ``oneOf`` whose branches pin distinct
``op`` consts can only pass on the matching branch, and an ``if`` whose discriminator does
not match is vacuous. Unknown discriminators fall back to the full definition, so they
still fail exactly as the plain validator would.
"""
from __future__ import annotations

import json
import os
import sys
from contextlib import contextmanager
from functools import cache
from pathlib import Path
from typing import Any, Iterable, Iterator

from jsonschema import Draft202012Validator
from jsonschema.exceptions import ValidationError, best_match

MODE_ENV = "ARPG_BOT_SCHEMA_VALIDATION"
MODES = ("off", "report", "strict")

PROTOCOL_DIR = Path(__file__).resolve().parents[2] / "shared" / "protocol"
SCHEMA_FILES = {
    "session_snapshot": "session_snapshot.v8.schema.json",
    "state_delta": "state_delta.v8.schema.json",
}


class PayloadSchemaError(AssertionError):
    """A live server payload violated its protocol schema."""


class _ItemDispatch:
    """Validators for one ``$defs`` item type, keyed by its discriminator value."""

    def __init__(self, field: str, by_value: dict[str, Draft202012Validator], fallback: Draft202012Validator) -> None:
        self.field = field
        self.by_value = by_value
        self.fallback = fallback

    def validator_for(self, item: Any) -> Draft202012Validator:
        value = item.get(self.field) if isinstance(item, dict) else None
        return self.by_value.get(value, self.fallback) if isinstance(value, str) else self.fallback


def _discriminator_values(condition: Any) -> tuple[str, list[str]] | None:
    """Return ``(field, values)`` for an ``if`` that only pins one required property's const/enum.

    The field must be required: otherwise an item missing it passes the ``if`` vacuously and
    the ``then`` applies to it, which per-value dispatch would skip.
    """
    if not isinstance(condition, dict) or set(condition) - {"properties", "required"}:
        return None
    props = condition.get("properties") or {}
    if len(props) != 1:
        return None
    (field, spec), = props.items()
    if field not in condition.get("required", []):
        return None
    if "const" in spec:
        return field, [spec["const"]]
    if "enum" in spec:
        return field, list(spec["enum"])
    return None


def _one_of_dispatch(defs: dict[str, Any], definition: dict[str, Any]) -> _ItemDispatch | None:
    branches = definition.get("oneOf")
    if not branches or set(definition) != {"oneOf"}:
        return None
    by_value: dict[str, Draft202012Validator] = {}
    for branch in branches:
        op = (branch.get("properties") or {}).get("op") or {}
        if "const" not in op or "op" not in branch.get("required", []) or op["const"] in by_value:
            return None
        by_value[op["const"]] = _validator(defs, branch)
    return _ItemDispatch("op", by_value, _validator(defs, definition))


def _all_of_dispatch(defs: dict[str, Any], definition: dict[str, Any]) -> _ItemDispatch | None:
    rules = definition.get("allOf")
    if not rules:
        return None
    field = None
    by_value: dict[str, list[dict[str, Any]]] = {}
    for rule in rules:
        pinned = _discriminator_values(rule.get("if")) if set(rule) <= {"if", "then"} else None
        if pinned is None or (field is not None and pinned[0] != field):
            return None
        field = pinned[0]
        for value in pinned[1]:
            by_value.setdefault(value, []).append(rule)
    base = {key: value for key, value in definition.items() if key != "allOf"}
    return _ItemDispatch(
        str(field),
        {value: _validator(defs, {**base, "allOf": subset}) for value, subset in by_value.items()},
        _validator(defs, base),
    )


def _validator(defs: dict[str, Any], schema: dict[str, Any]) -> Draft202012Validator:
    return Draft202012Validator({**schema, "$defs": defs})


class CompiledPayloadSchema:
    """One payload schema compiled into a shell validator plus per-item dispatch."""

    def __init__(self, schema: dict[str, Any]) -> None:
        defs = schema.get("$defs", {})
        shell = dict(schema)
        shell["properties"] = dict(schema.get("properties", {}))
        self.dispatch: dict[str, _ItemDispatch] = {}
        for prop, spec in schema.get("properties", {}).items():
            ref = (spec.get("items") or {}).get("$ref", "") if spec.get("type") == "array" else ""
            definition = defs.get(ref.removeprefix("#/$defs/")) if ref.startswith("#/$defs/") else None
            if definition is None:
                continue
            table = _one_of_dispatch(defs, definition) or _all_of_dispatch(defs, definition)
            if table is not None:
                shell["properties"][prop] = {**spec, "items": True}
                self.dispatch[prop] = table
        self.shell = Draft202012Validator(shell)

    def iter_errors(self, payload: Any) -> Iterator[tuple[list[Any], ValidationError, str]]:
        """Yield ``(absolute_path, leaf_error, item_tag)`` for every violation in ``payload``.

        ``item_tag`` names the dispatched item's discriminator (``event_type=boss_killed``)
        so a violation inside ``events/37`` says which event it was.
        """
        for error in self.shell.iter_errors(payload):
            leaf = _leaf_error(error)
            yield list(leaf.absolute_path), leaf, ""
        if not isinstance(payload, dict):
            return
        for prop, table in self.dispatch.items():
            items = payload.get(prop)
            if not isinstance(items, list):
                continue
            for index, item in enumerate(items):
                tag = f"{table.field}={item.get(table.field)}" if isinstance(item, dict) else ""
                for error in table.validator_for(item).iter_errors(item):
                    leaf = _leaf_error(error)
                    yield [prop, index, *leaf.absolute_path], leaf, tag


@cache
def compiled_schema(message_type: str) -> CompiledPayloadSchema | None:
    file_name = SCHEMA_FILES.get(message_type)
    if file_name is None:
        return None
    return CompiledPayloadSchema(json.loads((PROTOCOL_DIR / file_name).read_text(encoding="utf-8")))


def _leaf_error(error: ValidationError) -> ValidationError:
    """Drill through oneOf/anyOf wrappers to the most specific failing branch.

    The branch with the fewest errors is the one the payload was aiming at (a visible
    shop offer fails the mystery branch on every mystery-only field but the visible
    branch on just the real drift), so report from that branch.
    """
    while error.context:
        branches: dict[Any, list[ValidationError]] = {}
        for sub in error.context:
            branches.setdefault(sub.relative_schema_path[0] if sub.relative_schema_path else None, []).append(sub)
        error = best_match(min(branches.values(), key=len))
    return error


def _location(path: Iterable[Any], *, generic: bool) -> str:
    parts = ["*" if generic and isinstance(part, int) else str(part) for part in path]
    return "/".join(parts) or "<root>"


def payload_violations(message_type: str, payload: Any) -> list[tuple[str, str]]:
    """Return ``(dedupe_key, detail)`` for every schema violation in ``payload``.

    The key collapses array indices (``changes/*/entity``) so the same drift seen on
    every tick reports once; the detail keeps the concrete path for debugging.
    """
    compiled = compiled_schema(message_type)
    if compiled is None:
        return []
    violations: list[tuple[str, str]] = []
    try:
        for path, leaf, tag in compiled.iter_errors(payload):
            where = f" [{tag}]" if tag else ""
            key = f"{message_type} {_location(path, generic=True)}{where}: {_rule(leaf)}"
            detail = f"{message_type} at {_location(path, generic=False)}{where}: {_clip(leaf.message)}"
            violations.append((key, detail))
    except Exception as exc:  # noqa: BLE001 - a broken schema (dangling $ref) is a gate failure, not an ingest crash
        reason = f"schema error {type(exc).__name__}: {_clip(str(exc).splitlines()[0])}"
        violations.append((f"{message_type} {reason}", f"{message_type} {reason}"))
    return violations


# Keywords whose message echoes the offending instance (a whole party list, a seed
# string); key those by the rule instead so one drift reports once, not per value.
_INSTANCE_ECHO_KEYWORDS = frozenset({
    "type", "const", "pattern", "minLength", "maxLength", "minItems", "maxItems",
    "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "enum",
})


def _rule(error: ValidationError) -> str:
    if error.validator in _INSTANCE_ECHO_KEYWORDS:
        return f"{error.validator}={_clip(json.dumps(error.validator_value))}"
    return _clip(error.message)


def _clip(text: str, limit: int = 240) -> str:
    return text if len(text) <= limit else text[: limit - 3] + "..."


class PayloadSchemaGate:
    """Applies the configured mode to received payloads and remembers what it reported."""

    def __init__(self, mode: str = "off", *, stream: Any = None) -> None:
        if mode not in MODES:
            raise ValueError(f"{MODE_ENV} must be one of {MODES}, got {mode!r}")
        self.mode = mode
        self.stream = stream
        self.seen: dict[str, str] = {}

    def check(self, message_type: str, payload: Any) -> None:
        if self.mode == "off":
            return
        for key, detail in payload_violations(message_type, payload):
            if self.mode == "strict":
                raise PayloadSchemaError(f"schema violation: {detail}")
            if key not in self.seen:
                self.seen[key] = detail
                print(f"schema violation: {detail}", file=self.stream or sys.stderr, flush=True)


@cache
def default_gate() -> PayloadSchemaGate:
    gate = PayloadSchemaGate(os.environ.get(MODE_ENV, "off").strip().lower() or "off")
    if gate.mode != "off":
        print(f"payload schema gate: {gate.mode} ({', '.join(SCHEMA_FILES.values())})", file=sys.stderr, flush=True)
    return gate


@contextmanager
def suspended() -> Iterator[None]:
    """Pause the process-wide gate for a perf measurement window.

    Soak probes assert server tick overrun under a fixed client cadence. Validating inside
    that window slows the bot's reads; the server then coalesces backed-up deltas on its
    tick path and the probe measures the bot instead of the server (v486 A/B: overrun p95
    0 -> 54.7 ms). Setup and the initial snapshots stay validated.
    """
    gate = default_gate()
    previous = gate.mode
    gate.mode = "off"
    try:
        yield
    finally:
        gate.mode = previous


def check_payload(message_type: str, payload: Any) -> None:
    """Validate one received payload under the process-wide mode (no-op when off)."""
    default_gate().check(message_type, payload)
