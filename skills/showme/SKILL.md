---
name: showme
description: Capture focused Godot client visuals for fast feedback while tuning presentation. Use when the user asks to show, render, preview, screenshot, or open the client for a specific visual under improvement, such as equipment models, paper-doll inventory UI, item icons, character facing, armor/helmet/boots/shield placement, or another isolated Godot client presentation detail.
---

# Show Me

Use this skill to produce quick visual proof for the exact client element being tuned, without running the full server game loop unless the request needs it.

## Workflow

1. Prefer a focused screenshot first. It is fast, deterministic, and easy to attach back to the user.
2. Use live mode only when the user asks to see the client window, needs rotation/interaction, or a screenshot is not enough.
3. Keep the capture scoped to the thing under review. Do not start `make play` unless the user specifically needs a full gameplay path.
4. After changing visuals, rerun the focused capture and inspect the image before asking for feedback.

## Script

Run from the repo root:

```bash
python3 skills/showme/scripts/render_focus.py --focus gear
```

The script prints the screenshot path under `.artifacts/showme/`.

Common examples:

```bash
# Full currently tuned equipment set: sword, shield, helm, armor, boots.
python3 skills/showme/scripts/render_focus.py --focus gear

# A specific item or small set.
python3 skills/showme/scripts/render_focus.py --focus gear --items helm,mail,boots

# Paper-doll inventory UI only.
python3 skills/showme/scripts/render_focus.py --focus inventory

# Open a live Godot preview window for 45 seconds.
python3 skills/showme/scripts/render_focus.py --focus gear --mode live --duration 45

# Keep a live Godot preview open until the user closes it.
python3 skills/showme/scripts/render_focus.py --focus gear --mode live --duration 0
```

Screenshot and live mode both use Godot's render-capable window path because the macOS headless/dummy renderer cannot produce viewport pixels. If sandbox GUI restrictions block either mode, rerun the command with escalation and a short approval prompt.

## Focus Values

- `gear`: isolated character model with selected equipment.
- `inventory`: isolated inventory/paper-doll panel with sample authoritative state.

Add more focus values only when a repeated feedback loop appears.

## Notes

- The renderer mirrors client presentation code and shared visual metadata; it must not mutate server rules or gameplay authority.
- For visual-equipment changes, usually run `godot --headless --path client --script res://tests/test_item_visuals.gd` after a capture.
- For UI layout changes, usually run `make client-unit` after the focused iteration stabilizes.
