from __future__ import annotations

from typing import Any


def handle_runtime_economy_assertion(state: Any, assertion: dict[str, Any], where: str, helpers: dict[str, Any]) -> bool:
    typ = assertion.get("type")
    assert_count_matches = helpers["assert_count_matches"]

    if typ == "stash_item_count":
        helpers["assert_stash_item_count"](state.stash_items, assertion, where, assert_count_matches)
        return True
    if typ == "stash_gold":
        helpers["assert_stash_gold"](state.stash_gold, assertion, where, assert_count_matches)
        return True
    if typ == "stash_capacity":
        helpers["assert_stash_capacity"](state.stash_capacity, assertion, where, assert_count_matches)
        return True
    if typ == "shop_offer_count":
        offers = helpers["filtered_shop_offers"](state, assertion)
        assert_count_matches(len(offers), assertion, f"{where}: shop_offer_count", f": {offers}")
        return True
    if typ == "shop_offer_details":
        offers = helpers["filtered_shop_offers"](state, assertion)
        helpers["assert_shop_detail_rows"](offers, assertion, f"{where}: shop_offer_details")
        return True
    if typ == "shop_sell_appraisal_count":
        rows = helpers["filtered_shop_sell_appraisals"](state, assertion)
        assert_count_matches(len(rows), assertion, f"{where}: shop_sell_appraisal_count", f": {rows}")
        return True
    if typ == "shop_sell_appraisal_details":
        rows = helpers["filtered_shop_sell_appraisals"](state, assertion)
        detail_assertion = dict(assertion)
        detail_assertion.setdefault("price_key", "sell_price")
        helpers["assert_shop_detail_rows"](rows, detail_assertion, f"{where}: shop_sell_appraisal_details")
        return True
    if typ == "shop_event":
        matches = helpers["filtered_shop_events"](state, assertion)
        assert_count_matches(len(matches), assertion, f"{where}: shop_event", f": {matches}")
        helpers["assert_shop_event_details"](matches, assertion, f"{where}: shop_event")
        return True
    if typ == "stash_event":
        helpers["assert_stash_event"](state.stash_events, assertion, where, assert_count_matches)
        return True
    return False
