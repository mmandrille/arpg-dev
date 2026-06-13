from __future__ import annotations

from typing import Any, Callable

from tools.bot.bot_types import RuntimeState


CountMatcher = Callable[[int, dict[str, Any], str, str], None]


def filtered_stash_items(stash_items: list[dict[str, Any]], step: dict[str, Any]) -> list[dict[str, Any]]:
    items = list(stash_items)
    if step.get("stash_item_id") is not None:
        items = [item for item in items if str(item.get("stash_item_id", "")) == str(step["stash_item_id"])]
    if step.get("item_def_id") is not None:
        items = [item for item in items if str(item.get("item_def_id", "")) == str(step["item_def_id"])]
    if step.get("item_template_id") is not None:
        items = [item for item in items if str(item.get("item_template_id", "")) == str(step["item_template_id"])]
    if step.get("display_name") is not None:
        items = [item for item in items if str(item.get("display_name", "")) == str(step["display_name"])]
    if step.get("rolled") is not None:
        want_rolled = bool(step["rolled"])
        items = [item for item in items if bool(item.get("item_template_id")) == want_rolled]
    items.sort(key=lambda item: str(item.get("stash_item_id", "")))
    return items


def select_stash_item(state: RuntimeState, step: dict[str, Any]) -> dict[str, Any]:
    items = filtered_stash_items(state.stash_items, step)
    if not items:
        raise AssertionError(f"{step.get('action')}: no matching stash item for {step}; stash={state.stash_items}")
    index = int(step.get("stash_index", 0))
    if index < 0 or index >= len(items):
        raise AssertionError(f"{step.get('action')}: stash_index {index} out of range for {items}")
    return items[index]


def find_stash_item_by_id(stash_items: list[dict[str, Any]], stash_item_id: str | None) -> dict[str, Any] | None:
    if stash_item_id is None:
        return None
    return next((item for item in stash_items if str(item.get("stash_item_id")) == str(stash_item_id)), None)


def assert_stash_item_count(
    stash_items: list[dict[str, Any]],
    assertion: dict[str, Any],
    where: str,
    count_matches: CountMatcher,
) -> None:
    rows = filtered_stash_items(stash_items, assertion)
    count_matches(len(rows), assertion, f"{where}: stash_item_count", f": {rows}")


def assert_stash_gold(stash_gold: int, assertion: dict[str, Any], where: str, count_matches: CountMatcher) -> None:
    count_matches(stash_gold, assertion, f"{where}: stash_gold", "")


def assert_stash_capacity(stash_capacity: int, assertion: dict[str, Any], where: str, count_matches: CountMatcher) -> None:
    count_matches(stash_capacity, assertion, f"{where}: stash_capacity", "")


def filtered_stash_events(events: list[dict[str, Any]], assertion: dict[str, Any]) -> list[dict[str, Any]]:
    stash_id = str(assertion.get("stash_id", "account_stash"))
    event_type = str(assertion["event_type"])
    matches = [
        event for event in events
        if event.get("event_type") == event_type and str(event.get("stash_id", "")) == stash_id
    ]
    if assertion.get("stash_item_id") is not None:
        matches = [event for event in matches if str(event.get("stash_item_id", "")) == str(assertion["stash_item_id"])]
    return matches


def assert_stash_event(
    events: list[dict[str, Any]],
    assertion: dict[str, Any],
    where: str,
    count_matches: CountMatcher,
) -> None:
    matches = filtered_stash_events(events, assertion)
    count_matches(len(matches), assertion, f"{where}: stash_event", f": {matches}")
