# v232 Spec: Mercenary Recovery UI

Status: Complete
Date: 2026-06-16
Codename: mercenary-recovery-ui

## Purpose

Make mercenary loss visible and recoverable from the Godot mercenary panel. When the existing
server-owned `mercenary_lost` event fires, the panel should clear the hired roster, show a clear
lost/recovery status, and let the player rehire from the board.

## Baseline

Builds on v206 mercenary hiring, v207 mercenary roster UI, v219 stance controls, and v220
mercenary death loss. The server already removes the companion entity, emits `mercenary_lost`, and
allows rehire through the existing board path.

ADR alignment:
- ADR-0010: mercenary death/loss remains server authoritative and should have deterministic bot
  coverage.
- ADR-0014 D7/D9: companion losses should be readable and recoverable rather than silent.

Asset/plugin decision: adopt existing mercenary panel, bridge, panel unit, and client-bot scenario
framework; reject external assets/plugins.

## Non-goals

- No new durable mercenary roster, recovery timer, revive, insurance, refund, gear snapshot refresh,
  pricing/listing model, or backend contract change.
- No new art or animation.
- No per-mercenary command UI beyond the existing all-companions stance controls.

## Acceptance Criteria

- `mercenary_lost` is handled by the focused mercenary panel bridge.
- The panel clears the lost hired entity, displays a clear lost/recovery status, and renders an empty
  roster after the authoritative companion entity is removed.
- Reopening/clicking the mercenary board can hire a replacement and returns the panel to hired state.
- Focused mercenary panel unit coverage proves loss-state debug payload/status.
- A client-bot scenario hires a mercenary, observes `mercenary_lost`, verifies the panel and
  companion HUD show no active hire, then rehires a replacement.

## Scope and Likely Files

- `client/scripts/mercenary_panel.gd`
- `client/scripts/mercenary_panel_bridge.gd`
- `client/tests/test_mercenary_panel.gd`
- `tools/bot/scenarios/client/49_mercenary_recovery_ui.json`
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v232_mercenary-recovery-ui.md`

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- `make client-unit`
- `make bot-client scenario=49_mercenary_recovery_ui.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=49_mercenary_recovery_ui.json
```

## Open Questions and Risks

- No blocking questions. This slice deliberately uses the existing board rehire path as recovery.
- The live client scenario depends on the existing mercenary death-loss lab timing; use generous
  waits around the loss event.
