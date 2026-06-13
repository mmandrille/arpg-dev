# v123 Plan - Ranger Piercing And Pinning Shots

Status: Complete
Goal: Add Ranger `Piercing Shot` and `Pinning Shot` as server-authoritative, data-driven physical bow skills with bot and client visual proof.
Architecture: Shared skill data owns projectile, pierce, root, mana, cooldown, and requirement tuning. The Go sim owns all hit selection, damage, root state, movement blocking, cooldowns, and events. The Godot client renders only server-owned projectile/effect events and root state.
Tech stack: shared JSON/schema, Go sim, Python protocol bot, Godot client presentation, SDD docs.

## Baseline And Shortcut Decision

Baseline is v122 `ranger-class-foundation` on `main`, committed as `03753c7e`.

Godot plugin adoption checklist: reject external plugins/assets. This slice adds authoritative
combat mechanics and small code-native VFX/root markers using existing presentation helpers.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Add closed pierce/root skill payloads. |
| Modify | `shared/rules/skills.v0.json` | Add `piercing_shot` and `pinning_shot` Ranger skill rows. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add icon/projectile/effect metadata. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Add localized skill names/summaries. |
| Modify | `tools/validate_shared.py` | Cross-check Ranger skill contracts/presentations. |
| Modify | `server/internal/game/rules.go` | Decode/validate new skill contracts. |
| Add | `server/internal/game/ranger_skills.go` | Focused Piercing Shot, Pinning Shot, root helpers. |
| Add | `server/internal/game/ranger_skills_test.go` | Focused unit coverage. |
| Modify | `server/internal/game/handlers.go`, `sim.go` | Thin dispatch/effect-state hooks. |
| Modify | `client/scripts/projectile_visuals.gd` | Arrow VFX variants for Ranger skills. |
| Modify | `client/scripts/player_status_effect_markers.gd` | Root/pin marker presentation. |
| Modify | `client/tests/test_projectile_visuals.gd`, `test_status_effect_presentation.gd` | Focused client proof. |
| Add | `tools/bot/scenarios/59_ranger_piercing_and_pinning_shots.json` | Protocol proof. |
| Modify | `tools/bot/test_protocol.py` | Scenario discovery and assertion coverage. |
| Add | `docs/as-built/v123_ranger-piercing-and-pinning-shots.md` | As-built proof. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/game_test.go`
- [x] `client/scripts/main.gd`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`

Decision:
- [x] Extract focused helper/module/test file as part of this slice where practical.
- [x] Defer extraction with rationale only for tiny wiring changes in existing coordinator files.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Skill Contracts And Content

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/assets/skill_presentations.v0.json`
- Modify: `shared/i18n/en.json`, `shared/i18n/es.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add schema-backed `pierce` payload with max hits and per-target damage falloff.
- [x] Step 1.2: Add schema-backed `root` payload with effect id and duration ticks.
- [x] Step 1.3: Add `piercing_shot` and `pinning_shot` Ranger skill rows with physical damage,
  projectile visuals, class requirements, mana costs, and cooldowns.
- [x] Step 1.4: Add presentation metadata and localized text for both skills.
- [x] Step 1.5: Add validator cross-checks so Ranger has exactly the expected first two skills.

```bash
make validate-shared
```

## Task 2 - Server Authority

Files:
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/ranger_skills.go`
- Add: `server/internal/game/ranger_skills_test.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/sim.go`

- [x] Step 2.1: Decode and validate pierce/root rule payloads.
- [x] Step 2.2: Dispatch Ranger skills from `handleCastSkill` without adding broad logic to
  `applyInput`.
- [x] Step 2.3: Implement Piercing Shot as deterministic line target selection, physical damage,
  ordered multi-hit events, mana spend, cooldown, and projectile presentation event.
- [x] Step 2.4: Implement Pinning Shot as first-hit physical damage plus root start/end events and
  root state stored in the deterministic sim snapshot.
- [x] Step 2.5: Prevent rooted monsters from moving while root is active, without blocking damage,
  death, or effect expiry.
- [x] Step 2.6: Add focused tests for rule validation, class gates, multi-hit piercing, root
  movement prevention, root expiry, and current projectile skill regressions.

```bash
cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'
```

## Task 3 - Bot Proof

Files:
- Add: `tools/bot/scenarios/59_ranger_piercing_and_pinning_shots.json`
- Modify: `tools/bot/test_protocol.py`
- Modify: `tools/bot/run.py` only if existing assertions are insufficient.

- [x] Step 3.1: Add a Ranger protocol scenario with seeded ranks for both skills.
- [x] Step 3.2: Prove Piercing Shot damages at least two lined-up monsters with one cast.
- [x] Step 3.3: Prove Pinning Shot starts root on a chase monster and that the monster remains
  pinned before expiry.
- [x] Step 3.4: Keep assertions semantic; do not pin unrelated damage tuning.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=59_ranger_piercing_and_pinning_shots
```

## Task 4 - Godot Presentation

Files:
- Modify: `client/scripts/projectile_visuals.gd`
- Modify: `client/scripts/player_status_effect_markers.gd`
- Modify: `client/scripts/main.gd` only for thin state/event wiring if required.
- Modify: `client/tests/test_projectile_visuals.gd`
- Modify: `client/tests/test_status_effect_presentation.gd`

- [x] Step 4.1: Render `piercing_shot_projectile` as a green arrow/bolt variant.
- [x] Step 4.2: Render `pinning_shot_projectile` as a green arrow with a brighter pin accent.
- [x] Step 4.3: Render rooted monsters with a visible ground/root marker while the root effect id is
  active.
- [x] Step 4.4: Add focused GDScript tests for the projectile variants and root marker lifecycle.

```bash
make client-unit
```

## Task 5 - Lifecycle Docs And CI

Files:
- Modify: `docs/specs/v123_spec-ranger-piercing-and-pinning-shots.md`
- Modify: `docs/plans/v123_2026-06-13-ranger-piercing-and-pinning-shots.md`
- Add: `docs/as-built/v123_ranger-piercing-and-pinning-shots.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice, next slice, lifecycle row, and
  recently closed notes.
- [x] Step 5.3: Run final verification and commit only after green CI.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=59_ranger_piercing_and_pinning_shots`
- [x] `make client-unit`
- [x] `make ci`

## Deferred Scope

- `Volley` and the Ranger visual showcase scenario are deferred to v124.
- Production animation/audio and custom authored bow attack clips.
- Full Ranger balance pass beyond conservative data-owned first-skill tuning.

