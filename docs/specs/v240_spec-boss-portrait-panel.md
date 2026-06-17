# v240 Spec - Boss Portrait Panel

Status: Approved for autoloop
Date: 2026-06-17
Codename: boss-portrait-panel

## Purpose

Improve boss readability by adding a portrait tile to the existing screen-anchored boss health bar.
The portrait should give the Cave Warden a distinct HUD identity while continuing to use current
server-owned boss state.

## Non-goals

- No server/protocol changes, new boss templates, imported portrait art, model rendering, animation,
  boss combat changes, loot changes, or audio changes.
- No multi-boss selector, boss codex, or reward panel.

## Acceptance Criteria

- The boss health bar includes a visible code-drawn portrait tile while a boss is alive.
- The portrait derives its kind from `boss_template_id`, with a safe default for unknown bosses.
- The portrait hides/clears with the boss bar when no live boss is active.
- Boss health, phase countdown, fill ratios, and existing debug fields remain unchanged.
- Boss health bar debug state exposes portrait visibility and kind.
- A focused unit test proves portrait state while live and after hide/death.
- A client bot scenario reaches the Cave Warden and asserts the portrait kind through existing boss
  health bar assertions.

## Scope and Likely Files

- Client: `client/scripts/boss_health_bar.gd`, `client/scripts/bot_scenario_runner.gd`,
  `client/scripts/bot_step_catalog.gd`.
- Unit tests: `client/tests/test_boss_health_bar.gd`.
- Bot/scenario: `tools/bot/scenarios/client/57_boss_portrait_panel.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_boss_health_bar.gd`
- `make bot-client scenario=57_boss_portrait_panel.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. This slice rejects external art/assets and uses a deterministic code-drawn
  HUD portrait.
- Risk: code-drawn portrait art is placeholder quality. It is intentionally small and data-keyed so a
  future asset-backed portrait can replace it without server work.
