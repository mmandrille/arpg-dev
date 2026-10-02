# v519 Spec — skill-tree-ui

- **Status:** Complete — focused slice checks and combined batch CI passed.
- **Date:** 2026-10-01
- **Batch base:** `3a132626aa29581da6e3fdb36833c80e22b77b12` (detached v517 worktree)
- **Dependency:** v518 must be integrated and its exact integrated prerequisite transferred into this worktree before implementation in the overlapping skill, skill-icon, rarity, or UI-theme callsites. Spec and plan work can proceed against this base.
- **Sibling boundary:** v518 owns the theme and rarity/icon changes that overlap these callsites. v519 owns skill-tree state readability and the remaining skill-panel theme adoption after v518.

## Purpose

Make the skill tree easier to scan when choosing a skill: players should distinguish learned, spendable, and locked skills; understand prerequisite paths; see available points; and identify hover and selection. Bring the skill-tree panel into the existing `UiTheme` visual system. Treat the v515 inventory and v516 character screen as visual references; this is a focused readability pass, not a broader UI redesign.

The change is client presentation only. Server-owned progression, spend eligibility, skill rules, protocol payloads, and persisted progression remain authoritative and unchanged.

## Current-state inspection

At the assigned base:

- `client/scripts/skills_panel.gd` owns the panel, point count, skill blocks, rank labels, hover/selection, and tooltip. `_skill_visual_state()` already derives `highlight`, `normal`, and `disabled` from authoritative progression; `_skill_spend_enabled()` gates allocation. Its debug state exposes the values needed by tests.
- `client/scripts/skill_tree_connectors.gd` draws met and unmet prerequisite edges with hard-coded colors and width. `skill_tree_layout.gd` derives edges from skill requirements and current ranks.
- The panel, block, tooltip, point label, and connector colors/sizes are still locally styled. `UiTheme` and `shared/assets/ui_theme.v0.json` already exist. New skill-specific tokens and any validator/test changes must be coordinated with v518's integrated edits before they are added.
- `client/scripts/skill_icon.gd` already consumes the skill presentation catalog for icon shape/color/label. Preserve its gameplay identity and avoid duplicating skill-presentation data in the theme.
- `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json` is the requested client scenario (`make bot-visual scenario=client_skill_points_and_magic_bolt`). It currently exercises spending and class visibility; visual proof must include clearly distinct tree states.
- `client/scripts/showme/visual_capture.gd` already has a `skills` focus setup, and `skills/showme` can render it through the real renderer. Extend or add a focused state variant as needed so the capture includes learned, spendable, locked, connector, points, hover, and selection states.
- v515 inventory and v516 character-screen captures/as-built notes establish recent style references. v514's accepted UiTheme foundation supplies the token system. ADR-0018 requires in-repo assets and render evidence; this presentation slice does not need new art assets.

## Adopt / borrow / reject

| Candidate | Decision |
|---|---|
| Existing `UiTheme` and shared `ui_theme.v0.json` token system | **Adopt** after the v518 prerequisite is transferred. Add skill-prefixed tokens/recipes only where the current catalog has no appropriate semantic role; add matching validator/test coverage. |
| Existing `SkillIcon`, skill presentation catalog, and prerequisite layout model | **Adopt** unchanged as sources of icon identity and prerequisite truth. Improve their rendering only as required to express state; do not create duplicate rules or client-owned eligibility. |
| v515 inventory and v516 character-screen visual language | **Borrow** their restrained dark surfaces, clear hierarchy, consistent spacing, and shared theme recipes as visual references. |
| Third-party UI plugins, downloaded art, or a new icon pack | **Reject.** Existing Godot controls, icon renderer, and theme tokens are sufficient. No new dependency or art pipeline is justified. |

## Scope and design constraints

- Show learned, spendable, and locked states with distinguishable shape/border/contrast cues as well as color, so state is not conveyed by color alone.
- Make earned skill points visible and easy to find, including the zero-points state. Localize the new state and point labels through `TextCatalog`; do not change point counting or allocation behavior.
- Make met and unmet prerequisite connectors visually legible and distinguishable; keep connector endpoints and edge relationships derived from the existing `SkillTreeLayout` data.
- Strengthen hover and selected feedback without making either ambiguous with spendable state. Hover and selection must remain visible over each of the three skill states.
- Preserve skill names, rank values, tooltips, keyboard/mouse interactions, debug-state keys used by bot/tests, and the existing responsive/scroll behavior unless the implementation needs a narrowly justified presentation-only addition.
- Keep the existing draggable-window footprint and placement behavior; the panel need not fill the game viewport. Judge state readability within the skill window itself.
- Keep current node ownership and extract a focused style helper only if it reduces coupling or respects the file-size ratchet. Avoid growth in over-limit coordinators.
- Do not change skill definitions, progression data, protocol/schema/golden data, server behavior, balance, skill allocation rules, or new gameplay tuning.

## Acceptance criteria

1. A real-renderer skill-tree view makes learned, spendable, and locked states immediately distinguishable through multiple visual cues, with no state inferred from color alone.
2. Available skill points (including none available) are readable and visually subordinate to the tree title while still easy to locate.
3. Prerequisite connectors clearly distinguish met from unmet requirements and remain traceable between their correct skill blocks.
4. Hover and selection are visibly distinct from each other and from each block's progression state; the focused skill tooltip remains legible and correctly anchored.
5. All presentation colors, font roles/sizes, spacing, and frame styling owned by the skill tree use `UiTheme` tokens/recipes after the v518 transfer, with a validator/test that catches missing tokens or local hard-coded styling regressions.
6. Existing skill allocation and class filtering behavior remains unchanged, and tests cover the visual state matrix without duplicating tunable gameplay values.
7. `make bot-visual scenario=client_skill_points_and_magic_bolt` passes and produces usable real-renderer visual proof for the representative states. Add a showme/capture variant or other repeatable windowed-Godot capture if the scenario cannot expose all states in one frame.
8. Focused unit checks for skill-panel state/debug contracts, `UiTheme`, its validator, and maintainability/CODEMAP changes pass. Handoff records captures and limits; coordinator owns combined `make ci`.

## Likely surfaces

- Client: `client/scripts/skills_panel.gd`, `client/scripts/skill_tree_connectors.gd`, potentially `client/scripts/skill_icon.gd` and a small skill-tree styling helper; relevant `client/tests/` cases.
- Shared presentation: `shared/assets/ui_theme.v0.json` and `.schema.json` only if schema extension is needed; `tools/validate_ui_theme.py` and its tests.
- Text: `shared/i18n/en.json`, `shared/i18n/es.json`, and `tools/validate_i18n.py` for the new state and point labels.
- Visual proof: `client/scripts/showme/visual_capture.gd`, `skills/showme/scripts/render_focus.py`, and/or `tools/showme/screenshot_catalog.py`; reuse the client bot scenario above.
- Documentation: `docs/CODEMAP.md`, this spec, implementation plan, and as-built handoff evidence.
- No server, protocol, rules, or gameplay scenario changes expected.

## Focused verification

- `godot --headless --path client --script res://tests/test_skills_panel.gd`
- `godot --headless --path client --script res://tests/test_skill_tree_connectors.gd` (if connector behavior is extended; otherwise add assertions to the existing focused skill-tree test)
- `godot --headless --path client --script res://tests/test_ui_theme.gd`
- `pytest tools/test_validate_ui_theme.py tools/test_validate_codemap.py tools/test_file_size_ratchet.py tools/test_extraction_coupling_ratchet.py`
- `python tools/validate_ui_theme.py && make validate-shared && make maintainability`
- `make bot-visual scenario=client_skill_points_and_magic_bolt`
- Focused windowed renderer capture; inspect final images alongside v515/v516 reference captures.

## Dependencies and integration risks

- Wait for the coordinator to transfer the exact integrated v518 prerequisite into this worktree before editing overlapping skill, icon, rarity, or theme callsites. Preserve this spec/plan during transfer and run focused dependency checks before implementation.
- `shared/assets/ui_theme.v0.json`, its schema/validator/tests, and `skill_icon.gd` are likely shared conflicts; merge the intended v518 and v519 changes field-by-field, retaining v518's behavior and adding only skill-tree tokens needed by this scope.
- If `client/scripts/skills_panel.gd` is near its file-size limit, extract style/state presentation into an independent helper instead of increasing the coordinator's coupling or ratchet baseline.
- No unresolved product decision blocks planning.
