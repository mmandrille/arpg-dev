#!/usr/bin/env python3
"""Launch a focused Godot visual capture for fast feedback."""
from __future__ import annotations

import argparse
import subprocess
import sys
import time
from pathlib import Path


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def _default_output(root: Path, focus: str) -> Path:
    stamp = time.strftime("%Y%m%d-%H%M%S")
    return root / ".artifacts" / "showme" / f"{stamp}-{focus}.png"


def main() -> int:
    root = _repo_root()
    parser = argparse.ArgumentParser(description="Render a focused Godot client visual.")
    parser.add_argument("--focus", choices=["gear", "inventory", "skills", "shop", "character-menu", "join-menu", "hud", "stairs"], default="gear")
    parser.add_argument("--mode", choices=["screenshot", "live"], default="screenshot")
    parser.add_argument("--items", default="", help="Comma-separated item def ids for gear focus.")
    parser.add_argument("--output", default="", help="PNG output path for screenshot mode.")
    parser.add_argument("--width", type=int, default=640)
    parser.add_argument("--height", type=int, default=480)
    parser.add_argument("--duration", type=float, default=-1.0, help="Live mode timeout seconds; 0 keeps the window open until closed.")
    parser.add_argument("--godot", default="godot")
    args = parser.parse_args()

    output = Path(args.output) if args.output else _default_output(root, args.focus)
    output = output if output.is_absolute() else root / output
    output.parent.mkdir(parents=True, exist_ok=True)
    log_file = root / ".artifacts" / "showme" / "godot.log"
    log_file.parent.mkdir(parents=True, exist_ok=True)

    width = args.width
    height = args.height
    if args.focus == "inventory" and (args.width, args.height) == (640, 480):
        width, height = 960, 640
    if args.focus == "skills" and (args.width, args.height) == (640, 480):
        width, height = 960, 640
    if args.focus == "shop" and (args.width, args.height) == (640, 480):
        width, height = 1280, 760
    if args.focus in ["character-menu", "join-menu"] and (args.width, args.height) == (640, 480):
        width, height = 960, 640
    if args.focus == "stairs" and (args.width, args.height) == (640, 480):
        width, height = 960, 640

    duration = args.duration
    if args.mode == "live" and duration < 0.0:
        duration = 45.0

    gdscript = root / "skills" / "showme" / "scripts" / "visual_capture.gd"
    cmd = [
        args.godot,
        "--windowed",
        "--single-window",
        "--resolution",
        f"{width}x{height}",
    ]
    cmd += [
        "--disable-crash-handler",
        "--log-file",
        str(log_file),
        "--path",
        str(root / "client"),
        "--script",
        str(gdscript),
        "--",
        "--mode",
        args.mode,
        "--focus",
        args.focus,
        "--output",
        str(output),
        "--width",
        str(width),
        "--height",
        str(height),
        "--duration",
        str(duration),
    ]
    if args.items:
        cmd += ["--items", args.items]

    print("[showme] running:", " ".join(cmd))
    result = subprocess.run(cmd, cwd=root)
    if result.returncode != 0:
        return result.returncode
    if args.mode == "screenshot":
        if not output.exists():
            print(f"[showme] expected screenshot missing: {output}", file=sys.stderr)
            return 1
        print(f"[showme] screenshot: {output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
