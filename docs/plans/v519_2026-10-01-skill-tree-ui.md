# v519 Plan — skill-tree-ui

- **Status:** Complete — focused slice gates and combined batch `make ci` passed.

- **Goal:** Improve the skill tree's state, prerequisite, point, hover, and selection readability and adopt `UiTheme`, while retaining all progression behavior.
- **Baseline:** v517 at `3a132626aa29581da6e3fdb36833c80e22b77b12` (detached worktree).
- **Prerequisite:** v518's exact integrated commit/change set must be transferred by the coordinator into this worktree before implementation touching skill, skill-icon, rarity, or UI-theme callsites. Preserve these docs and all local work during transfer. After transfer, run the dependency checks below, review the merged diff, and update this plan's recorded base/dependency evidence before `/execute`.
- **Transferred v518 prerequisite:** coordinator-applied v518 tracked/untracked change set against base `3a132626aa29581da6e3fdb36833c80e22b77b12`. It was intentionally uncommitted in the detached worktree; there was no separate prerequisite commit SHA. All v518 paths were verified in the coordinator's integrated tree before the combined batch commit. It did not touch `skills_panel.gd`, `skill_icon.gd`, `skill_tree_connectors.gd`, or `ui_theme.v0.json`.
- **Slice type:** Client presentation only. Final gate is focused slice verification; the coordinator runs combined `make ci` after all accepted slices integrate.

## Review gate result

The v519 spec has a bounded visual goal and preserves server-owned progression and spend eligibility. Each acceptance criterion maps to skill-panel assertions, theme token/validator checks, the existing client bot scenario, or real-renderer capture. No protocol/schema/golden, Go determinism, replay, rules, world preset, or gameplay tuning change is planned. The UI is an application presentation surface with no user-controlled HTML/data rendering changes, authentication, networking, or data-access work; the security router classifies this slice as pure UI/UX and out of scope for security implementation/review. ADR-0018 sourcing is recorded in the spec: adopt in-repo UI/theme/icon systems, borrow v515/v516 visual references, reject new dependencies/assets. No material spec gap or open product decision remains.

## Dependency and ownership map

| Surface | v519 ownership | v518 / coordinator boundary |
|---|---|---|
| `client/scripts/skills_panel.gd` | Skill-block state treatment, points line, hover/selection, tooltip/frame theme adoption and UI wiring, without changing spend or selection contract | Wait for transfer before editing; merge against v518's current callsites |
| `client/scripts/skill_tree_connectors.gd` | Met/unmet connector appearance and theme integration | Wait for transfer; preserve v518 icon/theme behavior where files overlap |
| `client/scripts/skill_icon.gd` | Touch only if needed to visually support state and only after v518 integration; no duplicate icon/presentation catalog | v518 may own icon changes; inspect and retain its intent |
| `shared/assets/ui_theme.v0.json` and schema | Add only skill-tree tokens/recipes required by v519; preserve v518's tokens and schema edits | Shared additive conflict; reconcile field-by-field after transfer |
| `tools/validate_ui_theme.py`, `tools/test_validate_ui_theme.py` | Add focused coverage for skill token/recipe requirements if not already covered by v518 | Shared validator/tests; integrate, do not replace |
| `client/tests/test_skills_panel.gd`, `client/tests/test_skill_tree_layout.gd`, `client/tests/test_ui_theme.gd` | Cover state/readability semantics, connector met/unmet style contract, existing debug/interactions, and theme application | Check for v518 additions before editing |
| `shared/i18n/en.json`, `shared/i18n/es.json`, `tools/validate_i18n.py` | Localized state badges and point-count label | No v518 overlap |
| `client/scripts/showme/visual_capture.gd` or focused capture driver; `skills/showme/scripts/render_focus.py`; screenshot catalog if needed | Provide repeatable real-renderer frame(s) for learned, spendable, locked, point count, prerequisites, hover, and selection | Do not alter unrelated visual fixtures from v518 |
| `docs/CODEMAP.md`, spec, plan, as-built evidence | Update ownership index and report evidence | Coordinator owns global `PROGRESS.md` and lifecycle closeout |

`skills_panel.gd` was checked against existing tests/debug accessors and is the behavioral coordinator for this feature. The transferred tree's maintainability baseline was 999 lines; v519 extracted style presentation into `SkillTreeStyles` and reduced the panel to 947 lines, lowering that baseline to 947. No ratchet was raised. Do not add new UI state that duplicates progression inputs.

## Ordered tasks

### P0 — Transfer v518 and establish integration baseline (must precede implementation)

- [x] Receive the coordinator's exact v518 patch transfer into this detached worktree. Do not implement overlapping callsites before transfer.
- [x] Verify `git status`, base history, and transferred paths. The transferred patch is staged against the assigned base; there are no v518 edits in `skills_panel.gd`, `skill_icon.gd`, `skill_tree_connectors.gd`, or `ui_theme.v0.json`. Reviewed the v518 `visual_capture.gd` change: it waits for the item thumbnail cache only for the inventory focus; preserve this condition when extending the skills capture. Reviewed the CODEMAP addition for v520; do not replace it when registering v519 helper files.
- [x] Run P0 dependency checks. After initializing Godot's global class cache with `godot --headless --editor --path client --import --quit`, `godot --headless --path client --script res://tests/test_ui_theme.gd` passed (131 assertions), `make validate-shared` passed (2,257 checks plus CODEMAP validation), `python tools/validate_ui_theme.py` passed, and `.venv/bin/pytest tools/test_validate_ui_theme.py` passed (9 tests). An initial test invocation before Godot initialization could not resolve global classes; the rerun passed. The local `.venv` was provisioned by the repository validation target.

### P1 — Lock behavior and theme contract

- [x] Add or extend tests in `client/tests/test_skills_panel.gd` for learned (`normal`), spendable (`highlight`), locked (`disabled`), 0/nonzero points, and hover/selection combinations. Preserve current signal behavior, class visibility, and debug-state fields. Check: `godot --headless --path client --script res://tests/test_skills_panel.gd` — PASS (150 assertions).
- [x] Add connector assertions in `client/tests/test_skill_tree_layout.gd` or a focused connector test for met/unmet edge data and the theme style contract. Keep prerequisite edge calculation in `SkillTreeLayout`. Check: `godot --headless --path client --script res://tests/test_skill_tree_layout.gd` — PASS (211 assertions).
- [x] Add namespaced skill-tree palette/frame/font/spacing tokens after comparing with v518's transferred catalog/schema. Keep presentation-only style values in `shared/assets/ui_theme.v0.json`; validator coverage checks the skill tokens/recipes without changing the shared schema.
- [x] Localize READY/LEARNED/LOCKED state labels and the skill-points label in the English and Spanish catalogs. `python tools/validate_i18n.py` — PASS.
- [x] Add validator assertions/tests that the skill-specific tokens or recipes this panel owns exist; preserve generic literal reference checks. `python tools/validate_ui_theme.py` and `.venv/bin/pytest tools/test_validate_ui_theme.py` — PASS.

### P2 — Implement readable panel states and connectors

- [x] Update block styling to give learned, spendable, and locked states distinct, legible cues, including a non-color cue. Keep spendability derived from the existing progression row and `interactive`; do not relax or duplicate `_skill_spend_enabled()`.
- [x] Restyle connector edges for met/unmet status using theme-backed values; preserve `from`, `to`, `met` debug data and the existing path geometry.
- [x] Apply theme recipes/font roles to panel, blocks, point line, tooltip, rank/key labels, and connector styles. Preserve all user-visible text/tooltip content and existing node/debug contracts unless additional presentation metadata is strictly needed and covered by tests.
- [x] Strengthen hover and selected feedback as separate states that can be seen on each progression state; keep tooltip positioning and interaction behavior intact.
- [x] Extract style construction/presentation to `SkillTreeStyles` and register it in `docs/CODEMAP.md`. `make maintainability` — PASS; baseline lowered from 999 to the resulting 947 lines.

### P3 — Bot and real-renderer proof

- [x] Extend the current renderer fixture to represent learned, spendable, locked, met/unmet prerequisites, zero/nonzero points, hover, and selection without changing production progression rules. Reuse the existing `skills` focus and preserve v518's inventory-only thumbnail-cache wait.
- [x] Run `make bot-visual scenario=client_skill_points_and_magic_bolt` — first attempt exited 1; rerun PASS (1 passed / 0 failed). If the bot does not expose the complete state matrix, use deterministic focused captures.
- [x] Render through windowed Godot / real renderer at 960×640 and inspect all three captures for contrast, text fit, clear connectors, hover/selection distinction, and clipping. Save evidence under `docs/as-built/assets/v519/`.
- [x] Do not claim performance improvement. UI-only styling adds no per-frame work by design; record that no frame-time/draw-call A/B was run.

### P4 — Regressions, docs, and handoff

- [x] Run focused final checks. Panel, layout, and theme Godot tests; Python validators; `make validate-shared`; `make maintainability`; the visible bot scenario; and `git diff --check` passed. The combined `make ci` remains coordinator-owned.

  ```bash
  godot --headless --path client --script res://tests/test_skills_panel.gd
  godot --headless --path client --script res://tests/test_skill_tree_layout.gd
  godot --headless --path client --script res://tests/test_ui_theme.gd
  .venv/bin/pytest tools/test_validate_ui_theme.py tools/test_validate_codemap.py tools/test_file_size_ratchet.py tools/test_extraction_coupling_ratchet.py
  python tools/validate_ui_theme.py && make validate-shared && make maintainability
  python tools/validate_i18n.py
  make bot-visual scenario=client_skill_points_and_magic_bolt
  git diff --check
  ```

- [x] Run and inspect repeatable real-renderer captures; document renderer, resolution, exact capture paths, and state matrix.
- [x] Update `docs/CODEMAP.md`, add `docs/as-built/v519_skill-tree-ui.md` with behavior, checks, visual evidence, and limitations; `PROGRESS.md` and the lifecycle index remain coordinator-owned.
- [x] Prepare handoff with prerequisite/base, dependency checks, complete changed/untracked path list, ignored evidence, outcomes, acceptance gaps, shared-file conflicts, and dirty state. No commit, push, `make ci`, `make ci-full`, or `/finish`.

## Maintainability and data ownership

- No gameplay tuning values are introduced or changed. New theme values describe presentation only and are owned by the existing shared UI theme catalog.
- Skill-tree graph topology (`SkillTreeLayout` origin, row/column gap, and node footprint) remains code-owned structural layout; theme tokens own the panel's visual spacing, typography, frames, and state styling.
- Do not raise file-size or extraction-coupling baselines. Keep `skills_panel.gd` at or below its assigned base line count, and keep any helper independently testable.
- No protocol schema, golden, server outcome, replay, or persisted progression change is planned.

## Handoff checklist

- [x] Record the v518 base and coordinator-applied change set; verify every v518 path is represented in the integrated tree. The prerequisite was intentionally uncommitted and had no separate commit SHA.
- [x] List complete changed/deleted/untracked paths and preserve renderer evidence under `docs/as-built/assets/v519/`.
- [x] Report commands and outcomes, named bot scenario, renderer captures/resolution, and visual-state coverage.
- [x] Report visual/runtime limits, shared-file reconciliation, clean/dirty handoff state, and final integrated revision.
