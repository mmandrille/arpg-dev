#!/usr/bin/env python3
"""Hold a connected listed co-op host for Godot Join Game client-bot scenarios."""

from __future__ import annotations

import argparse
import asyncio
import json
from pathlib import Path
import signal
import sys
import tempfile
import time
from typing import Any

ROOT = Path(__file__).resolve().parent.parent.parent
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

import httpx

from tools.bot.run import (
    connect_coop_peer,
    create_listed_coop_session,
    dev_login,
    ensure_character,
    list_active_sessions,
    log,
)


def build_metadata(
    *,
    session: dict[str, Any],
    host_email: str,
    host_account_id: str,
    host_character_id: str,
    host_player_id: str,
) -> dict[str, Any]:
    return {
        "ready": True,
        "session_id": str(session["session_id"]),
        "world_id": str(session.get("world_id", "")),
        "mode": str(session.get("mode", "")),
        "listed": bool(session.get("listed", False)),
        "host_email": host_email,
        "host_account_id": host_account_id,
        "host_character_id": host_character_id,
        "host_player_id": host_player_id,
    }


def write_metadata(path: Path, metadata: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", dir=path.parent, delete=False) as tmp:
        json.dump(metadata, tmp, sort_keys=True)
        tmp.write("\n")
        tmp_path = Path(tmp.name)
    tmp_path.replace(path)


async def wait_active_row(
    *,
    client: httpx.Client,
    token: str,
    session_id: str,
    timeout_s: float,
) -> dict[str, Any]:
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        rows = list_active_sessions(client, token)
        for row in rows:
            if str(row.get("session_id", "")) == session_id and int(row.get("connected_count", 0)) >= 1:
                return row
        await asyncio.sleep(0.1)
    raise TimeoutError(f"listed host session {session_id} did not become active-browser visible")


async def run(args: argparse.Namespace) -> int:
    stop = asyncio.Event()
    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        try:
            loop.add_signal_handler(sig, stop.set)
        except NotImplementedError:
            pass

    metadata_path = Path(args.metadata_file)
    with httpx.Client(base_url=args.base_url, timeout=10.0) as client:
        host_account_id, token = dev_login(client, args.email, args.dev_token)
        host_character_id = ensure_character(client, token, args.character_name)
        session = create_listed_coop_session(client, token, args.world_id, host_character_id, args.seed)
        peer = await connect_coop_peer(args.base_url, token, session, "client-join-preflight-host", args.world_id)
        pump_task = asyncio.create_task(_pump_until_stop(peer.ws, stop))
        wait_task: asyncio.Task[bool] | None = None
        try:
            await wait_active_row(
                client=client,
                token=token,
                session_id=str(session["session_id"]),
                timeout_s=args.ready_timeout_s,
            )
            write_metadata(
                metadata_path,
                build_metadata(
                    session=session,
                    host_email=args.email,
                    host_account_id=host_account_id,
                    host_character_id=host_character_id,
                    host_player_id=peer.state.local_player_id,
                ),
            )
            log("client join preflight ready", session["session_id"], "metadata", str(metadata_path))
            wait_task = asyncio.create_task(stop.wait())
            done, _pending = await asyncio.wait({wait_task, pump_task}, return_when=asyncio.FIRST_COMPLETED)
            if pump_task in done:
                pump_task.result()
        finally:
            stop.set()
            if wait_task is not None:
                wait_task.cancel()
            if not pump_task.done():
                pump_task.cancel()
            await asyncio.gather(*(task for task in (wait_task, pump_task) if task is not None), return_exceptions=True)
            try:
                await peer.ws.close()
            except Exception:
                pass
    return 0


async def _pump_until_stop(ws, stop: asyncio.Event) -> None:
    while not stop.is_set():
        try:
            await asyncio.wait_for(ws.recv(), timeout=0.5)
        except asyncio.TimeoutError:
            continue
        except asyncio.CancelledError:
            raise
        except Exception:
            if not stop.is_set():
                raise
            return


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="prepare and hold a listed co-op host for a Godot client-bot join scenario")
    parser.add_argument("--base-url", default="http://localhost:8888")
    parser.add_argument("--dev-token", default="local-dev-token")
    parser.add_argument("--world-id", default="dungeon_levels")
    parser.add_argument("--seed", default="")
    parser.add_argument("--email", required=True)
    parser.add_argument("--character-name", default="Join Host")
    parser.add_argument("--metadata-file", required=True)
    parser.add_argument("--ready-timeout-s", type=float, default=10.0)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    return asyncio.run(run(args))


if __name__ == "__main__":
    raise SystemExit(main())
