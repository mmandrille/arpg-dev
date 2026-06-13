# v136 Plan — Unique Chest Client Proof

Status: Complete
Goal: Add a focused Godot client bot scenario that proves the purple unique chest UI exposes named
uniques and readable effect summaries.
Architecture: Reuse the existing stash panel and client bot runner. Add generic stash row filters
for display name, container mode, and summary text instead of a unique-chest-specific assertion.
Tech stack: Godot GDScript client UI/debug state, client bot scenario JSON, lifecycle docs.

## Baseline And Shortcut Decision

Builds on v135 second named unique. Godot plugin adoption is rejected for this slice because it only
adds a client bot proof around existing UI/debug state; no new UI component or asset pipeline is
needed.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/stash_panel.gd` | Expose unique effect summary text in row debug state |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add generic stash row filters/assertions |
| Create | `tools/bot/scenarios/client/40_unique_chest_client_proof.json` | Client bot proof |
| Create | `docs/as-built/v136_unique-chest-client-proof.md` | As-built summary |
| Modify | `docs/specs/v136_spec-unique-chest-client-proof.md` | Status closeout |
| Modify | `docs/plans/v136_2026-06-13-unique-chest-client-proof.md` | Checkbox closeout |
| Modify | `PROGRESS.md` | Lifecycle and next-slice update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/stash_panel.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep stash panel change to debug-row composition only.
- [x] Keep runner changes generic row matching helpers, not chest-specific logic.

Verification:
```bash
make maintainability
```

## Task 1 — Bot-Readable Unique Tooltip State

Files:
- Modify: `client/scripts/stash_panel.gd`

- [x] Step 1.1: Include unique-effect text in stash row debug `summary_lines`.
- [x] Step 1.2: Preserve existing stash panel visible behavior and tooltip rendering.

## Task 2 — Generic Stash Assertions

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`

- [x] Step 2.1: Allow stash row matching by `display_name`.
- [x] Step 2.2: Allow stash row matching/assertion by `summary_contains`.
- [x] Step 2.3: Allow unique chest mode assertions through existing stash panel steps.

## Task 3 — Client Scenario

Files:
- Create: `tools/bot/scenarios/client/40_unique_chest_client_proof.json`

- [x] Step 3.1: Open the `town_unique_chest` interactable.
- [x] Step 3.2: Assert the panel is visible in `unique_chest` mode.
- [x] Step 3.3: Assert `Embercall Blade` and `Stormstring Bow` rows each appear with readable effect
  summaries.

```bash
make client-unit
SCENARIO=unique_chest_client_proof HEADLESS=1 make bot-client
```

## Task 4 — Lifecycle Docs And CI

Files:
- Create: `docs/as-built/v136_unique-chest-client-proof.md`
- Modify: `docs/specs/v136_spec-unique-chest-client-proof.md`
- Modify: `docs/plans/v136_2026-06-13-unique-chest-client-proof.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete.
- [x] Step 4.2: Record v136 completion and next backlog pointer in `PROGRESS.md`.
- [x] Step 4.3: Add the v136 as-built summary.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make client-unit`
- [x] `SCENARIO=unique_chest_client_proof HEADLESS=1 make bot-client`
- [x] `make maintainability`
- [x] `make ci`
