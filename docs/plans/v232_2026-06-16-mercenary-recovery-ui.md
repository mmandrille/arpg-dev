# v232 Plan - Mercenary Recovery UI

Status: Complete
Goal: Surface mercenary loss and recovery in the existing Godot mercenary panel.
Architecture: Reuse the authoritative `mercenary_lost` event and existing board rehire flow; keep
the change inside focused client mercenary panel/bridge files.
Tech stack: Godot panel/bridge, focused GDScript unit, client bot scenario, SDD docs.

## Baseline and shortcut decision

Do not add backend recovery semantics. v220 already proves loss and rehire. This slice only makes
that recovery path visible and bot-proven in the client. Asset/plugin decision: adopt existing
mercenary panel and bot framework; reject external assets/plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/mercenary_panel.gd` | Add loss/recovery state and status |
| Modify | `client/scripts/mercenary_panel_bridge.gd` | Route `mercenary_lost` events |
| Modify | `client/tests/test_mercenary_panel.gd` | Prove loss-state debug payload/status |
| Add | `tools/bot/scenarios/client/49_mercenary_recovery_ui.json` | Live client loss/recovery proof |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v232_mercenary-recovery-ui.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] None expected. Avoid touching `client/scripts/main.gd`, which is at its allowed ceiling.
- [x] Did every touched grandfathered file stay within the ratchet?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice: not needed; all touched source
  files are below 600 lines.
- [x] Defer extraction with rationale: existing focused panel/bridge split is sufficient.

Verification:
```bash
make maintainability
```

## Task 1 - Panel loss state

Files:
- Modify: `client/scripts/mercenary_panel.gd`
- Modify: `client/scripts/mercenary_panel_bridge.gd`

- [x] Step 1.1: Add `apply_lost_event` that clears the lost hire, displays a recovery status, and
  renders an empty roster.
- [x] Step 1.2: Route `mercenary_lost` through the bridge without changing `main.gd`.

## Task 2 - Unit and client scenario

Files:
- Modify: `client/tests/test_mercenary_panel.gd`
- Add: `tools/bot/scenarios/client/49_mercenary_recovery_ui.json`

- [x] Step 2.1: Extend the panel unit to prove loss status and `hired_count=0`.
- [x] Step 2.2: Add a client scenario that hires, waits for loss, asserts recovery UI, and rehires.

## Task 3 - Lifecycle proof

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v232_mercenary-recovery-ui.md`

- [x] Step 3.1: Record v232 as complete with focused proof and note final batch CI is pending.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- [x] `make client-unit`
- [x] `make bot-client scenario=49_mercenary_recovery_ui.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` is deferred to `$autoloop` after the selected queue commits.
