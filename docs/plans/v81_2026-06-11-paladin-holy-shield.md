# v81 Plan - Paladin Holy Shield

Status: Complete
Goal: Add a Paladin area defensive buff with server-owned effect state and visible holy shielding.
Architecture: Extend the closed skill catalog with an `area_stat_percent_buff` effect rather than
adding one-off Paladin code. The Go sim owns affected targets, stat effects, cooldowns, mana, and
expiry; Godot renders effect ids and status rows as presentation only. Effect state becomes
per-player so co-op allies can carry independent visible buffs.
Tech stack: shared JSON schemas, Go sim, Python protocol bot, Godot client, lifecycle docs.

## Baseline and shortcut decision

Builds on v61 Rage/Heal, v70 class gates, v73-v75 skill/status UI chrome, and v80 floating
presentation patterns. Godot plugin adoption checklist: reject external plugins/assets because the
slice needs a small code-native halo and status row, not a reusable VFX framework.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | Add closed area ally stat buff effect schema |
| Modify | `shared/rules/skills.v0.json` | Add `holy_shield` Paladin skill |
| Modify | `shared/assets/skill_presentations.v0.json` | Add icon/summary/effect visual metadata |
| Modify | `server/internal/game/rules.go` | Load/validate new skill kind/effect |
| Modify | `server/internal/game/sim.go` | Per-player effect state, stat application, expiry, entity views |
| Modify | `server/internal/game/handlers.go` | Cast area stat buff skill |
| Modify | `server/internal/game/game_test.go` | Holy Shield server behavior and regression tests |
| Modify | `tools/bot/run.py` | Assertions for effect events/state if needed |
| Add | `tools/bot/scenarios/43_paladin_holy_shield.json` | Protocol proof |
| Modify/Add | `client/scripts/*` | Holy Shield halo/status presentation |
| Modify | `client/tests/*` | Focused client unit proof |
| Modify | `PROGRESS.md` | Lifecycle and deferred scope |
| Add | `docs/as-built/v81_paladin-holy-shield.md` | As-built summary |

## Task 1 - Shared skill catalog

Files:
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/assets/skill_presentations.v0.json`

- [x] Step 1.1: Add a closed `area_stat_percent_buff` effect with target `allies`, include-caster,
  range, radius, stats, percent/rank scaling, duration, and optional `effect_id`.
- [x] Step 1.2: Add `holy_shield` as a Paladin skill using the new effect and conservative tuning.
- [x] Step 1.3: Add Holy Shield presentation metadata with a shield label and holy effect visual.

```bash
make validate-shared
```

## Task 2 - Server authority

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Extend Go skill structs/validation for `area_stat_buff` skill kind and effect.
- [x] Step 2.2: Generalize active skill effects to be per affected player while preserving Rage.
- [x] Step 2.3: Apply Holy Shield defensive stat effects in derived/effective stats and expire them
  deterministically.
- [x] Step 2.4: Emit effect-start/end events and entity updates carrying effect ids for affected
  heroes.
- [x] Step 2.5: Add Go tests for Paladin cast, non-Paladin rejection, ally targeting, defensive
  stat improvement, expiry, and Rage regression.

```bash
cd server && go test ./internal/game -run 'HolyShield|Rage|Skill'
```

## Task 3 - Protocol bot proof

Files:
- Add: `tools/bot/scenarios/43_paladin_holy_shield.json`
- Modify: `tools/bot/run.py` if existing assertions cannot prove effect state

- [x] Step 3.1: Create a Paladin skill progression lab scenario that learns Holy Shield.
- [x] Step 3.2: Cast Holy Shield and assert skill cast, cooldown, effect-start, and effect-end or
  active effect state.
- [x] Step 3.3: Keep assertions tuning-friendly; do not pin unrelated defensive formulas.

```bash
make bot scenario=43_paladin_holy_shield.json
```

## Task 4 - Godot presentation

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/status_effects_bar.gd`
- Add/Modify: focused effect helper script if useful
- Modify: `client/tests/test_status_effects_bar.gd`
- Modify: `client/tests/test_coop_client.gd` or another focused client test

- [x] Step 4.1: Render a code-native shining Holy Shield effect on any entity with the effect id.
- [x] Step 4.2: Remove the effect when entity state or effect-end events clear it.
- [x] Step 4.3: Feed Holy Shield effect-start events into the status-effect bar with remaining time.
- [x] Step 4.4: Add focused client tests for status row and world-effect attachment/removal.

```bash
make client-unit
```

## Task 5 - Lifecycle and CI

Files:
- Modify: `docs/specs/v81_spec-paladin-holy-shield.md`
- Modify: `docs/plans/v81_2026-06-11-paladin-holy-shield.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v81_paladin-holy-shield.md`

- [x] Step 5.1: Mark spec/plan complete, add lifecycle row/as-built, and record deferred scope.
- [x] Step 5.2: Run final verification.

```bash
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'HolyShield|Rage|Skill'`
- [x] `make bot scenario=43_paladin_holy_shield.json`
- [x] `make client-unit`
- [ ] `make ci`

## Deferred scope

- Production VFX/audio, shield absorb resources, thorns/reflect, taunt, invulnerability, and full
  buff/debuff taxonomy beyond the closed Holy Shield area stat buff.
