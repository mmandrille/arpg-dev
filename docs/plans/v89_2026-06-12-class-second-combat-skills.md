# v89 Plan - Class Second Combat Skills

Status: Complete
Goal: Add Barbarian Cleave and Sorcerer Ice Shard as server-authoritative, data-driven combat skills with bot and client visual proof.
Architecture: Extend the closed skill catalog with two new skill kinds/effects: cone area damage with knockback, and cold projectile damage with deterministic shard fan-out plus stackable slow. The Go sim owns all targeting, damage, pushback, slow state, projectile/shard resolution, RNG, cooldowns, and event emission. Godot renders only server-owned state/events: red cleave area, authoritative enemy movement, and light-blue slow tint.
Tech stack: shared JSON/schema, Go sim, Python protocol bot, Godot client, lifecycle docs.

## Baseline and shortcut decision

Baseline is v88 `skill-visual-rank-seeding` on `main`, with a clean worktree after maintenance.
This slice builds on v59/v61 skill catalogs, v70 class gates, v81/v88 skill visual tooling, and
existing projectile/combat text/client status-effect presentation.

Godot plugin adoption checklist: reject external plugins/assets. The requested visuals are small
code-native overlays/tints driven by server events and authoritative entity/effect state.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Add `cleave` and `ice_shard` rule rows. |
| Modify | `shared/rules/skills.v0.schema.json` | Validate cone, push, shard, slow, and skill hit policy fields. |
| Modify | `shared/assets/skill_presentations.v0.json` | Add skill presentation rows. |
| Modify | `shared/protocol/state_delta.v*.schema.json` | Allow new skill push/slow/shard/area presentation event fields if needed. |
| Modify | `shared/protocol/session_snapshot.v*.schema.json` | Allow slow/effect metadata in entity snapshots if schema requires it. |
| Modify | `tools/validate_shared.py` | Cross-check new skill rule capabilities and presentations. |
| Modify | `server/internal/game/rules.go` | Decode/validate new skill contracts. |
| Modify/Add | `server/internal/game/*skill*.go` | Implement focused skill helpers outside `sim.go` where practical. |
| Modify/Add | `server/internal/game/*test.go` | Add focused Cleave/Ice Shard tests outside the overlarge `game_test.go` where practical. |
| Modify | `tools/bot/run.py` | Add assertion helpers only if existing event/combat/entity assertions are insufficient. |
| Modify | `tools/bot/test_protocol.py` | Cover new bot assertion helpers and scenario discovery. |
| Add | `tools/bot/scenarios/45_class_second_combat_skills.json` | Protocol proof for both skills. |
| Modify | `tools/bot/test_skill_demo.py` | Verify skill demo catalog covers the new skills. |
| Modify | `tools/bot/test_skill_visual.py` | Verify skill visual matrix/wrapper includes new skills. |
| Modify/Add | `client/scripts/*skill*` | Render red Cleave cone and Ice Shard slow tint using focused helpers where practical. |
| Modify/Add | `client/tests/*skill*` | Focused GDScript presentation tests. |
| Add | `docs/as-built/v89_class-second-combat-skills.md` | As-built proof. |
| Modify | `PROGRESS.md` | Mark v89 complete and record deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [x] `server/internal/game/sim.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`

Decision:
- [x] Extract focused helper/module/test file as part of this slice where practical.
- [x] Defer extraction with rationale only for tiny wiring changes in existing coordinator files.

Rationale: v89 touched existing closed-contract dispatch, sim projectile/monster movement helpers,
and Godot event routing. A late broad extraction would have increased risk more than it reduced it.
The file-size baseline was updated with this documented exception, including the unrelated existing
`server/internal/http/auth_session_test.go` drift reported by the ratchet.

Verification:

```bash
make maintainability
```

## Task 1 - Shared skill contracts and content

Files:
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/assets/skill_presentations.v0.json`
- Modify: `shared/protocol/state_delta.v*.schema.json`
- Modify: `shared/protocol/session_snapshot.v*.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Add schema-backed fields for default skill hit policy: skills are `always_hit`
  unless data opts into normal combat resolution.
- [x] Step 1.2: Add a closed `cone_attack` or equivalent Cleave contract with range, angle degrees,
  damage source, push min/max distance, and target limit behavior.
- [x] Step 1.3: Add a closed cold projectile shard contract with shard min/max count, shard damage
  divisor semantics, slow percent, slow duration ticks, slow cap percent, and projectile visual id.
- [x] Step 1.4: Add `cleave` and `ice_shard` content rows with conservative data-owned tuning.
- [x] Step 1.5: Add presentation metadata for both skills and validation cross-checks.

```bash
make validate-shared
```

## Task 2 - Server skill mechanics

Files:
- Modify: `server/internal/game/rules.go`
- Modify/Add: `server/internal/game/*skill*.go`
- Modify/Add: `server/internal/game/*test.go`

- [x] Step 2.1: Decode and validate Cleave and Ice Shard skill fields in Go.
- [x] Step 2.2: Make skill damage default to 100% hit chance unless a skill opts into normal
  hit/block resolution.
- [x] Step 2.3: Implement Cleave target selection using caster facing/cast direction, 3-unit range,
  50-degree cone, deterministic target ordering, melee/weapon damage, and server-owned pushback.
- [x] Step 2.4: Emit Cleave combat/presentation events and authoritative entity position changes
  for pushed enemies.
- [x] Step 2.5: Implement Ice Shard impact damage, slow effect application/refresh, stacked slow cap
  to 75%, light-blue effect id propagation, deterministic shard count/directions, and shard damage.
- [x] Step 2.6: Add focused Go tests for validation, 100% skill hit behavior, Cleave multi-hit/push,
  Ice Shard slow stacking/cap, deterministic shard secondary damage, and regressions for current skills.

```bash
cd server && go test ./internal/game -run 'Cleave|IceShard|Skill'
```

## Task 3 - Bot proof and skill visual tooling

Files:
- Add: `tools/bot/scenarios/45_class_second_combat_skills.json`
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Modify: `tools/bot/test_skill_demo.py`
- Modify: `tools/bot/test_skill_visual.py`

- [x] Step 3.1: Add a protocol lab scenario that seeds or earns both class skills, casts Cleave
  against clustered enemies, and asserts multi-hit plus push distance.
- [x] Step 3.2: Extend or reuse assertions to prove Ice Shard impact, shard secondary damage,
  slow effect state, and slow cap without pinning unrelated tuning.
- [x] Step 3.3: Update skill demo/visual tests so `cleave` and `ice_shard` appear dynamically.
- [x] Step 3.4: Verify skill visual dry-runs for both new skills.

```bash
.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_demo.py tools/bot/test_skill_visual.py -q
make bot scenario=45_class_second_combat_skills.json
make skill-visual-list
make skill-visual skill=cleave DRY_RUN=1
make skill-visual skill=ice_shard DRY_RUN=1
```

## Task 4 - Godot presentation

Files:
- Modify/Add: `client/scripts/*skill*`
- Modify: `client/scripts/main.gd`
- Modify/Add: `client/tests/*skill*`

- [x] Step 4.1: Render Cleave as a short red cone/area from the caster when the server emits the
  skill cast or presentation event.
- [x] Step 4.2: Render Ice Shard slow by tinting affected enemies light blue while their server
  effect id is active, and clear it on expiry/entity refresh.
- [x] Step 4.3: Add focused GDScript tests for Cleave presentation and Ice Shard slow tint state.
- [x] Step 4.4: Keep `main.gd` changes as thin wiring; place reusable presentation logic in focused
  helper scripts if more than trivial.

```bash
make client-unit
make client-smoke
```

## Task 5 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v89_spec-class-second-combat-skills.md`
- Modify: `docs/plans/v89_2026-06-12-class-second-combat-skills.md`
- Add: `docs/as-built/v89_class-second-combat-skills.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add v89 as-built notes.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice, lifecycle row, recently closed notes,
  and deferred scope.
- [x] Step 5.3: Run final verification and commit only after green CI.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'Cleave|IceShard|Skill'`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_demo.py tools/bot/test_skill_visual.py -q`
- [x] `make bot scenario=45_class_second_combat_skills.json`
- [x] `make skill-visual-list`
- [x] `make client-unit`
- [x] `make client-smoke` via `make ci` phase 9
- [x] `make ci`

## Deferred scope

- Production VFX/audio and custom assets.
- Ground-position skill targeting.
- PvP skill behavior.
- Full skill balance pass.
- General-purpose debuff taxonomy beyond the slow state needed for Ice Shard.
