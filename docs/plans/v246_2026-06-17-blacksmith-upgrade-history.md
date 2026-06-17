# v246 Plan - Blacksmith Upgrade History

Status: Complete
Goal: Show recent blacksmith upgrade attempts in the panel.
Architecture: No server or protocol changes. The panel records history when
`update_after_upgrade` receives the authoritative result and delegates rendering/debug state to a
new focused `BlacksmithUpgradeHistory` helper.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v245 selected recipe IDs and existing upgrade result response handling.

Asset/plugin decision:
- Adopt existing text-first blacksmith UI.
- Borrow selected recipe label, item title, and response success/cost fields.
- Reject external assets/plugins and icons.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/blacksmith_upgrade_history.gd` | Render and expose recent attempt history |
| Modify | `client/scripts/blacksmith_panel.gd` | Record attempts and include history debug state |
| Modify | `client/tests/test_blacksmith_panel.gd` | Prove history behavior |
| Add | `tools/bot/scenarios/client/63_blacksmith_upgrade_history.json` | Client proof |
| Add | `docs/as-built/v246_blacksmith-upgrade-history.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/blacksmith_panel.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Add history in `blacksmith_upgrade_history.gd`, not directly in `main.gd`.
- [x] Keep `blacksmith_panel.gd` below 600 lines after the hook.

Verification:
```bash
make maintainability
```

## Task 1 - History helper and panel hook

Files:
- Add: `client/scripts/blacksmith_upgrade_history.gd`
- Modify: `client/scripts/blacksmith_panel.gd`

- [x] Build a compact helper that renders newest-first history rows and caps entries.
- [x] Record recipe label, item display name, success/failure, and cost after each upgrade result.
- [x] Expose history visibility, row count, and rows in blacksmith debug state.

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
```

## Task 2 - Tests and bot proof

Files:
- Modify: `client/tests/test_blacksmith_panel.gd`
- Add: `tools/bot/scenarios/client/63_blacksmith_upgrade_history.json`

- [x] Extend the blacksmith panel unit test for hidden empty history, recording, ordering, and cap.
- [x] Add a scenario that performs an upgrade and verifies the blacksmith remains in the upgraded
  state after history recording.

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=63_blacksmith_upgrade_history.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v246_blacksmith-upgrade-history.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=63_blacksmith_upgrade_history.json HEADLESS=1`
- [x] `make maintainability`
