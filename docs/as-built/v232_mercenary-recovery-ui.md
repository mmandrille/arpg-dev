# v232 As-Built - Mercenary Recovery UI

Date: 2026-06-16

## What shipped

- Routed the existing server-authored `mercenary_lost` event through the focused
  `MercenaryPanelBridge`.
- Added mercenary panel loss state that clears the lost hire, renders an empty roster, and displays
  a clear "lost - hire a replacement" recovery status.
- Extended the focused mercenary panel unit to prove loss-state debug payload and status.
- Added `49_mercenary_recovery_ui.json`, which hires a mercenary, waits for the loss event, verifies
  the panel and companion HUD show no active hire, and rehires a replacement through the board.
- Avoided touching `main.gd`, preserving the existing file-size ratchet ceiling.

## Proof

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make client-unit
make bot-client scenario=49_mercenary_recovery_ui.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
passed after the selected v226-v232 feature queue completed.

Manual visual proof, if desired:

```bash
make bot-visual scenario=49_mercenary_recovery_ui.json
```

## Scope limits

- No durable mercenary roster, recovery timer, revive, insurance, refund, gear snapshot refresh,
  pricing/listing model, or backend contract change shipped.
- No new art or animation shipped.
- No per-mercenary command UI beyond the existing all-companions stance controls shipped.
