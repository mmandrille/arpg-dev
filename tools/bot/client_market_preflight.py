#!/usr/bin/env python3
"""Prepare a seller market listing for Godot client-bot market purchase scenarios."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
import sys
import tempfile
from typing import Any

ROOT = Path(__file__).resolve().parent.parent.parent
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

import httpx

from tools.bot.bot_types import Scenario
from tools.bot.run import auth, dev_login, ensure_character, log, run_verified_session


def write_metadata(path: Path, metadata: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", dir=path.parent, delete=False) as tmp:
        json.dump(metadata, tmp, sort_keys=True)
        tmp.write("\n")
        tmp_path = Path(tmp.name)
    tmp_path.replace(path)


def create_listing(client: httpx.Client, token: str, stash_item_id: str, price_gold: int) -> dict[str, Any]:
    resp = client.post(
        "/v0/market/listings",
        headers=auth(token),
        json={"stash_item_id": stash_item_id, "price_gold": price_gold},
    )
    resp.raise_for_status()
    return resp.json()


def create_offer(client: httpx.Client, token: str, listing_id: str, stash_item_ids: list[str]) -> dict[str, Any]:
    resp = client.post(
        f"/v0/market/listings/{listing_id}/offers",
        headers=auth(token),
        json={"stash_item_ids": stash_item_ids},
    )
    resp.raise_for_status()
    return resp.json()


def stash_item_for(
    client: httpx.Client,
    args: argparse.Namespace,
    email: str,
    character_name: str,
    item_def_id: str,
) -> tuple[str, str, str]:
    account_id, token = dev_login(client, email, args.dev_token)
    character_id = ensure_character(client, token, character_name)
    scenario = Scenario(
        id="client_market_preflight",
        world_id=args.world_id,
        seed=args.seed,
        peer_count=1,
        title="Client market preflight",
        description="Create a market stash item for a Godot client-bot market scenario.",
        character_class="",
        debug_progression={},
        steps=[
            {"action": "action_entity", "item_def_id": item_def_id},
            {"action": "open_stash"},
            {"action": "deposit_stash_item", "item_def_id": item_def_id},
        ],
        assertions=[],
        fresh_session_checks=[],
        path=Path(__file__),
    )
    _session, state = run_verified_session(
        client=client,
        base_url=args.base_url,
        token=token,
        debug_token=args.debug_token,
        scenario=scenario,
        world_id=args.world_id,
        steps=scenario.steps,
        assertions=[],
        seed=args.seed,
        debug_progression={},
    )
    matches = [item for item in state.stash_items if str(item.get("item_def_id", "")) == item_def_id]
    if not matches:
        raise AssertionError(f"stash missing {item_def_id}: {state.stash_items}")
    return account_id, character_id, str(matches[0]["stash_item_id"])


def run(args: argparse.Namespace) -> int:
    metadata_path = Path(args.metadata_file)
    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        seller_account_id, seller_character_id, stash_item_id = stash_item_for(
            client, args, args.email, args.character_name, args.item_def_id
        )
        _seller_account_id, seller_token = dev_login(client, args.email, args.dev_token)
        listing = create_listing(client, seller_token, stash_item_id, args.price_gold)
        offer: dict[str, Any] = {}
        bidder_account_id = ""
        if args.offer_email and args.offer_item_def_id:
            bidder_account_id, _bidder_character_id, offer_stash_item_id = stash_item_for(
                client, args, args.offer_email, args.offer_character_name, args.offer_item_def_id
            )
            _bidder_account_id, bidder_token = dev_login(client, args.offer_email, args.dev_token)
            offer = create_offer(client, bidder_token, str(listing.get("listing_id", "")), [offer_stash_item_id])
        write_metadata(
            metadata_path,
            {
                "ready": True,
                "type": "market_listing",
                "seller_email": args.email,
                "seller_account_id": seller_account_id,
                "seller_character_id": seller_character_id,
                "host_email": args.email,
                "bidder_email": args.offer_email,
                "bidder_account_id": bidder_account_id,
                "offer_id": str(offer.get("offer_id", "")),
                "listing_id": str(listing.get("listing_id", "")),
                "item_def_id": str(listing.get("item_def_id", args.item_def_id)),
                "price_gold": int(listing.get("price_gold", args.price_gold)),
            },
        )
        log("client market preflight ready", listing.get("listing_id"), args.item_def_id, args.price_gold)
    return 0


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="prepare a seller listing for a Godot client market purchase scenario")
    parser.add_argument("--base-url", default="http://localhost:8888")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--debug-token", default="local-debug-token")
    parser.add_argument("--world-id", default="vendor_lab")
    parser.add_argument("--seed", default="")
    parser.add_argument("--email", required=True)
    parser.add_argument("--character-name", default="Market Seller")
    parser.add_argument("--metadata-file", required=True)
    parser.add_argument("--item-def-id", default="cave_mail")
    parser.add_argument("--price-gold", type=int, default=37)
    parser.add_argument("--offer-email", default="")
    parser.add_argument("--offer-character-name", default="Market Bidder")
    parser.add_argument("--offer-item-def-id", default="cave_blade")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    return run(parse_args(argv))


if __name__ == "__main__":
    raise SystemExit(main())
