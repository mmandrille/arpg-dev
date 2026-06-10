# v59 Plan - Data-Driven Skill Catalog

Status: Complete
Goal: Move Magic Bolt into a schema-backed skill catalog with requirements, presentation metadata, and generic server/client handling.
Architecture: Shared skill rules remain authoritative mechanics data, with bounded evaluator/helper types instead of free-form formulas. Skill presentation is split into `shared/assets` so the client can render labels/tooltips/icons without affecting server authority. The Go sim enforces learning/cast requirements and owns mana, cooldown, projectiles, and damage; Godot only renders catalog/progression/cooldown state.
Tech stack: shared JSON schemas/rules, Go deterministic sim, Python protocol bot, Godot GDScript client tests and bot.

## Baseline and shortcut decision

Baseline is v58 `boss-pattern-variety` on `main`. This reuses the v44 skill-point/Magic Bolt loop,
the v55 `ItemRulesLoader` static singleton pattern, and the existing protocol/client bot scenarios.

Godot shortcut decision: **reject external skill-tree/gameplay plugins** for v59. Skill authority
must stay in Go/shared rules, and this slice only needs to make the existing in-repo one-skill UI
data-driven. No addon, asset pack, or editor-bound setup is adopted or borrowed.

Local data decision: the user approved deleting local characters for a clean baseline. Use
`make db-reset` before end-to-end bot proof if stale local data learned Magic Bolt without the new
`magic >= 15` requirement.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Expanded Magic Bolt mechanics catalog |
| Modify | `shared/rules/skills.v0.schema.json` | Validate catalog shape and evaluator/helper types |
| Add | `shared/assets/skill_presentations.v0.json` | Client presentation metadata |
| Add | `shared/assets/skill_presentations.v0.schema.json` | Presentation schema |
| Modify | `shared/golden/skill_points_and_magic_bolt.json` | Updated stat requirement proof |
| Modify | `shared/golden/skill_points_and_magic_bolt.v0.schema.json` | Golden schema if requirement fields are added |
| Modify | `tools/validate_shared.py` | Cross-check rules, presentations, requirements, goldens |
| Modify | `server/internal/game/rules.go` | Skill structs, validation, requirement/evaluator helpers |
| Modify | `server/internal/game/handlers.go` | Requirement enforcement on spend/cast |
| Modify | `server/internal/game/sim.go` | Skill helper shape compatibility if needed |
| Modify | `server/internal/game/game_test.go` | Requirement, validation, damage/cooldown tests |
| Modify | `server/internal/replay/replay_test.go` | Replay fixture expectations if affected |
| Add | `client/scripts/skill_rules_loader.gd` | Static shared skill/presentation loader |
| Modify | `client/scripts/skills_panel.gd` | Catalog-driven skill row, requirement status, tooltip |
| Modify | `client/scripts/skill_bar.gd` | Catalog-driven slot label/tooltip |
| Modify | `client/scripts/main.gd` | Wire catalog data into skill UI if needed |
| Modify | `client/scripts/bot_scenario_runner.gd` | Requirement/presentation assertions |
| Modify | `client/tests/test_skills_panel.gd` | Requirement/render state coverage |
| Modify | `client/tests/test_skill_bar.gd` | Presentation/cooldown state coverage |
| Modify | `client/tests/test_golden.gd` | Skill golden parity if needed |
| Modify | `tools/bot/scenarios/32_skill_points_and_magic_bolt.json` | Protocol proof update |
| Modify | `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json` | Client proof update |
| Add | `docs/researchs/data-driven-content-libraries.md` | Future full data-driven library manifest definition for v60 review |
| Add | `docs/as-built/v59_data-driven-skill-catalog.md` | Slice as-built close-out |
| Modify | `PROGRESS.md` | Lifecycle and backlog close-out |

## Task 1 - Shared skill catalog

Files:
- Add: `docs/researchs/data-driven-content-libraries.md`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/rules/skills.v0.schema.json`
- Add: `shared/assets/skill_presentations.v0.json`
- Add: `shared/assets/skill_presentations.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Expand Magic Bolt from the v44 flat shape to the v59 catalog shape: `class`,
  `tree.tier`, `tree.column`, `kind: projectile_attack`, `requirements.stats.magic = 15`,
  `cost.mana`, `damage.type = rank_linear_range`, `projectile`, and `cooldown`.
- [x] Step 1.2: Record the future full data-driven library manifest direction for v60 review.
- [x] Step 1.3: Add skill presentation metadata for Magic Bolt label/icon/summary/projectile visual
  in `shared/assets`.
- [x] Step 1.4: Update JSON schemas so unknown authoritative gameplay fields fail validation, and
  supported evaluator/helper types are explicit.
- [x] Step 1.5: Update `tools/validate_shared.py` cross-checks for skill rules, requirement stats,
  presentation keys, and existing golden expectations.

```bash
make validate-shared
```

## Task 2 - Server skill rules and requirements

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 2.1: Replace Magic Bolt-specific skill validation with catalog-wide validation over all
  supported skill definitions.
- [x] Step 2.2: Add typed structs and helpers for requirements, mana cost, rank-linear damage,
  projectile settings, and attack-interval cooldown.
- [x] Step 2.3: Enforce `magic >= 15` in both skill-point allocation and skill casting; rejection
  must not mutate points, mana, cooldown, projectiles, or damage.
- [x] Step 2.4: Preserve deterministic sorted skill progression views and existing skill event
  payloads.
- [x] Step 2.5: Add or update Go tests for requirement rejection, successful learn/cast after
  magic allocation, unsupported catalog type validation, and replay expectations.

```bash
cd server && go test ./internal/game/... -run Skill
cd server && go test ./internal/replay/... -run Skill
```

## Task 3 - Golden fixtures and protocol bot

Files:
- Modify: `shared/golden/skill_points_and_magic_bolt.json`
- Modify: `shared/golden/skill_points_and_magic_bolt.v0.schema.json`
- Modify: `tools/bot/scenarios/32_skill_points_and_magic_bolt.json`

- [x] Step 3.1: Update the skill golden to include the Magic Bolt `magic >= 15` requirement and
  new catalog fields where useful.
- [x] Step 3.2: Update bot assertions/helpers to prove requirement rejection and successful magic
  allocation before learning Magic Bolt.
- [x] Step 3.3: Update the protocol scenario so a clean character starts at magic 5, fails to learn
  Magic Bolt, levels enough to allocate ten stat points into magic, learns/casts, rejects recast,
  recovers, reconnects, replays, and proves fresh-session persistence.
- [x] Step 3.4: Reset local DB before full bot proof if stale characters would hide the new
  requirement behavior.

```bash
make validate-shared
make db-reset
make bot
```

## Task 4 - Godot skill catalog UI

Files:
- Add: `client/scripts/skill_rules_loader.gd`
- Modify: `client/scripts/skills_panel.gd`
- Modify: `client/scripts/skill_bar.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_skills_panel.gd`
- Modify: `client/tests/test_skill_bar.gd`
- Modify: `client/tests/test_golden.gd`
- Modify: `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json`

- [x] Step 4.1: Add a `SkillRulesLoader` static singleton for shared skill mechanics and
  presentation metadata.
- [x] Step 4.2: Refactor `SkillsPanel` to render Magic Bolt from catalog/progression state, expose
  requirement status in debug state or tooltip, and avoid `magic_bolt`-only branches where a single
  catalog row can drive the UI.
- [x] Step 4.3: Refactor `SkillBar` to resolve label/tooltip from presentation metadata while still
  using server `skill_cooldowns` for availability.
- [x] Step 4.4: Update Godot unit tests for disabled-before-requirement, enabled-after-magic, and
  presentation-driven label/tooltip.
- [x] Step 4.5: Update the client bot scenario to prove the skills panel disabled/enabled path and
  skill bar cooldown after learning.

```bash
make client-unit
make client-smoke
SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh
```

## Task 5 - Lifecycle docs and final verification

Files:
- Modify: `docs/plans/v59_2026-06-10-data-driven-skill-catalog.md`
- Modify: `docs/specs/v59_spec-data-driven-skill-catalog.md`
- Add: `docs/as-built/v59_data-driven-skill-catalog.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark plan checkboxes complete as implementation lands.
- [x] Step 5.2: Update the spec status if project convention for completed specs is used.
- [x] Step 5.3: Add v59 lifecycle row, update current status, and record newly closed/deferred
  skill-catalog scope in `PROGRESS.md`.
- [x] Step 5.4: Add v59 as-built notes.
- [x] Step 5.5: Run final CI.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -run Skill`
- [x] `cd server && go test ./internal/replay/... -run Skill`
- [x] `make client-unit`
- [x] client smoke via `make ci` phase 9
- [x] `make db-reset`
- [x] protocol bot + replay via `make ci` phase 7
- [x] `SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh`
- [x] `make ci`

## Deferred scope

- New active skills beyond Magic Bolt.
- Free-form formula/expression language.
- Class selection or class-locked character creation.
- Respec/refund, passive skills, alternate branches, and production skill VFX/audio.
- New skill capabilities such as AoE, DOT, homing, summons, traps, auras, chained projectiles, and
  status effects.
- Backward-compatible migration for stale local characters that learned Magic Bolt before the new
  requirement.
