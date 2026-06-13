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


def run(args: argparse.Namespace) -> int:
    scenario = Scenario(
        id="client_market_preflight",
        world_id=args.world_id,
        seed=args.seed,
        peer_count=1,
        title="Client market preflight",
        description="Create a priced seller listing for a Godot client-bot buyer.",
        character_class="",
        debug_progression={},
        steps=[
            {"action": "action_entity", "item_def_id": args.item_def_id},
            {"action": "open_stash"},
            {"action": "deposit_stash_item", "item_def_id": args.item_def_id},
        ],
        assertions=[],
        fresh_session_checks=[],
        path=Path(__file__),
    )
    metadata_path = Path(args.metadata_file)
    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        seller_account_id, token = dev_login(client, args.email, args.dev_token)
        seller_character_id = ensure_character(client, token, args.character_name)
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
        matches = [item for item in state.stash_items if str(item.get("item_def_id", "")) == args.item_def_id]
        if not matches:
            raise AssertionError(f"seller stash missing {args.item_def_id}: {state.stash_items}")
        stash_item_id = str(matches[0]["stash_item_id"])
        listing = create_listing(client, token, stash_item_id, args.price_gold)
        write_metadata(
            metadata_path,
            {
                "ready": True,
                "type": "market_listing",
                "seller_email": args.email,
                "seller_account_id": seller_account_id,
                "seller_character_id": seller_character_id,
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
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    return run(parse_args(argv))


if __name__ == "__main__":
    raise SystemExit(main())
