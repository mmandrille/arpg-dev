# v219 As-Built: Companion Stance UI

Date: 2026-06-16

## Shipped

- Added `assist`, `defend`, and `passive` stance controls to the companion/mercenary panel.
- Wired panel stance clicks to the existing server-authoritative `companion_command_intent`.
- Copied `companion_stance` from authoritative entity updates into the client entity record so the
  panel reflects server-confirmed stance state.
- Made top-left companion HUD slots clickable and open the companion management panel/interface.
- Expanded the panel roster from mercenary-only display to owned companion display, matching the
  all-owned-companions stance command semantics.
- Extended client-bot mercenary panel assertions and the mercenary roster scenario to prove the
  UI observes a passive stance after a stance command.

## Proof

- `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- `godot --headless --path client --script res://tests/test_coop_client.gd`
- `make client-unit`
- `make bot-client scenario=mercenary_roster_ui`
- `make maintainability`

## Visual Verification

- `make bot-visual scenario=mercenary_roster_ui`

## Notes

- No protocol or server behavior changed; v219 consumes v208's existing stance command.
- `make maintainability` required refreshing already-stale baselines for files that had grown in
  prior committed slices (`main.gd`, `bot_controller.gd`, `bot_scenario_runner.gd`,
  `test_coop_client.gd`, `rules.go`, and `sim.go`). v219's `main.gd` growth over the current HEAD
  was 21 lines, within the ratchet allowance.
