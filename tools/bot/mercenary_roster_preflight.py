#!/usr/bin/env python3
"""Seed alt characters for mercenary hire client scenarios."""
from __future__ import annotations

import argparse
import json
from pathlib import Path

import httpx

from tools.bot.run import dev_login, ensure_character, seed_debug_progression


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--dev-token", required=True)
    parser.add_argument("--debug-token", required=True)
    parser.add_argument("--email", required=True)
    parser.add_argument("--scenario-path", required=True)
    args = parser.parse_args()

    scenario = json.loads(Path(args.scenario_path).read_text())
    roster = scenario.get("roster_characters", [])
    if not roster:
        print("[mercenary-roster-preflight] no roster_characters; skipping")
        return 0

    with httpx.Client(base_url=args.base_url, timeout=30.0) as client:
        _, token = dev_login(client, args.email, args.dev_token)
        for entry in roster:
            if not isinstance(entry, dict):
                continue
            name = str(entry.get("name", ""))
            if not name:
                continue
            character_id = ensure_character(client, token, name, str(entry.get("character_class", "")))
            progression = entry.get("debug_progression")
            if isinstance(progression, dict) and progression:
                seed_debug_progression(client, token, args.debug_token, character_id, progression)
            print("[mercenary-roster-preflight] ready", name, character_id)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
