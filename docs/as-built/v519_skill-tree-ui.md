# v519 — Skill Tree UI (as built)

- **Status:** Focused implementation and visual gates passed; combined batch CI and lifecycle closeout remain coordinator-owned.
- **Date:** 2026-10-01
- **Spec / plan:** [spec](../specs/v519_spec-skill-tree-ui.md) · [plan](../plans/v519_2026-10-01-skill-tree-ui.md)
- **Base:** v517 `3a132626aa29581da6e3fdb36833c80e22b77b12`, with the coordinator's exact uncommitted v518 patch transferred into this detached worktree; no integrated prerequisite SHA exists.
- **Scope:** Client presentation only. Progression, point counting, spend eligibility, skill definitions, server behavior, and protocol are unchanged.

## What changed

- Skill nodes show localized LEARNED, READY, or LOCKED badges in addition to icon and rank. Learned takes precedence for ranked skills; READY derives from the existing spend eligibility; locked covers the remaining visible skills. Selection and hover remain independent. The existing allocation signal, class filtering, tooltip content, and debug contract are retained.
- A localized points line shows both available and zero-point states. Skill nodes, state badges, rank/hotkey labels, panel, tooltip, and prerequisite connectors use namespaced `UiTheme` palette, spacing, typography, and frame recipes through `SkillTreeStyles`.
- Prerequisite edge truth and geometry remain owned by `SkillTreeLayout`; the renderer consumes its edge data and uses theme-backed met/unmet styling. The focused helper owns presentation only. No gameplay tuning values were added.
- The deterministic `skills` capture now covers points, no-points, and hovered-node variants. The fixture displays a learned node, a spendable node, locked nodes, met/unmet edges, selected/hovered distinction, and a tooltip.
- `skills_panel.gd` is 947 lines. Its maintainability baseline was reduced from 999 to 947 to match the post-extraction file; no ratchet was raised. The earlier plan's 1,012-line baseline did not match this checkout.

## Focused verification

| Check | Result |
|---|---|
| `godot --headless --path client --script res://tests/test_skills_panel.gd` | PASS — 150 assertions |
| `godot --headless --path client --script res://tests/test_skill_tree_layout.gd` | PASS — 211 assertions |
| `godot --headless --path client --script res://tests/test_ui_theme.gd` | PASS — 182 assertions |
| `python tools/validate_ui_theme.py` | PASS |
| `python tools/validate_i18n.py` | PASS |
| `.venv/bin/pytest tools/test_validate_ui_theme.py tools/test_regen_screenshots.py tools/test_validate_codemap.py` | PASS — 24 tests |
| `make validate-shared` | PASS — 2,266 checks plus CODEMAP validation |
| `make maintainability` | PASS — file-size, extraction-coupling, and progress-dashboard ratchets |
| `make bot-visual scenario=client_skill_points_and_magic_bolt` | PASS on rerun — 1 passed, 0 failed. The first attempt exited 1; the rerun log included a transient login failure before the scenario's PASS sentinel, plus Godot shutdown resource-leak warnings. |
| `make regen-screenshots SUITE=skill-tree` | PASS — all 3 real-renderer captures generated at 960×640 with Godot 4.7.2 / Metal Forward+ on Apple M4 Pro |
| `godot --headless --editor --path client --quit` | PASS — Godot registered the new style helper and updated its global class cache |
| Combined batch `make ci` | PASS — 11/11 stages, 7m50s; `make ci-full` was not run |

The transferred v518 prerequisite checks also passed before implementation: the UI theme test (131 assertions), shared validation (2,257 checks plus CODEMAP), UI-theme validator, and its 9 pytest cases. The combined CI result is from the coordinator's fully integrated v518–v520 tree.

## Windowed visual evidence

- [Skill tree with one available point](assets/v519/points.png) — reframed to 540×360 around the existing window.
- [Skill tree with zero available points](assets/v519/nopoints.png) — reframed to 540×360 around the existing window.
- [Hovered Ice Shard with tooltip](assets/v519/hover.png) — full 960×640 capture retained so the complete tooltip remains visible.

The capture driver renders a 960×640 window. The points and no-points evidence crop its top-left 540×360 around the skills panel, removing unused gray fixture canvas without changing pixels inside the panel. The hover evidence remains uncropped because its tooltip extends below the window and needs the full viewport to show all lines. The app's persisted Skills window position on this machine is `(0, 0)`, confirming upper-left placement is the existing draggable runtime window behavior, not a capture-only offset. The accepted spec requires readability within the existing window and does not require viewport-wide width. Original full captures remain in ignored `.artifacts/screenshots/20261001-213921/skill-tree/`.

The captures were inspected for state contrast, connector distinction, text fit, tooltip legibility, selection/hover separation, and clipping. The fixtures use the existing icon renderer and shared theme; they are deterministic presentation evidence and do not claim performance or broad gameplay proof. The visible bot scenario covers the live interaction path.

## Handoff limits

- The worktree is detached and uncommitted. v518's transferred patch remains staged; v519 changes are unstaged/untracked. The coordinator owns dependency integration, combined CI, lifecycle documents, and batch closeout.
- Godot reports resource-leak warnings at shutdown after the passing bot run. No frame-time or draw-call A/B was run because this is a presentation-only style change.
- No known behavior or shared-file integration blocker remains. The full changed/untracked path list and ignored generated evidence are reported in the task handoff.
