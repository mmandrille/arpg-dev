# v239 Spec - Mercenary Stats Card

Status: Approved for autoloop
Date: 2026-06-17
Codename: mercenary-stats-card

## Purpose

Make hired mercenaries easier to inspect from the companion panel by adding a compact stats card for
the active roster. The card should summarize only client-visible companion state: name, HP, stance,
active state, and entity id.

## Non-goals

- No server/protocol changes, durable mercenary roster, mercenary gear, attack/armor exposure,
  recovery timers, per-mercenary commands, or new companion AI behavior.
- No external assets, portraits, or standalone mercenary detail window.

## Acceptance Criteria

- The mercenary panel renders a stats card when at least one companion is hired.
- The card hides when the roster is empty or after a mercenary loss clears the roster.
- Card lines include mercenary display name, HP current/max, stance, active state, and id.
- Mercenary panel debug state exposes card visibility/text/lines.
- Existing stance controls, hire status, recovery status, and companion roster behavior continue to
  work.
- A focused unit test proves card content and hide-on-loss behavior.
- A client bot scenario hires a mercenary and asserts card text via existing mercenary panel
  assertions.

## Scope and Likely Files

- Client: `client/scripts/mercenary_panel.gd`, `client/scripts/bot_mercenary_panel_assertions.gd`,
  `client/scripts/bot_step_catalog.gd`.
- Unit tests: `client/tests/test_mercenary_panel.gd`.
- Bot/scenario: `tools/bot/scenarios/client/56_mercenary_stats_card.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- `make bot-client scenario=56_mercenary_stats_card.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions. The card intentionally avoids stats that are not in client companion state.
- Risk: card text duplicates roster information. Keeping it compact and grouped makes it readable
  without adding new data contracts.
