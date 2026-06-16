# v205 Plan — Boss Enrage Phase

Status: Complete
Goal: Add a data-driven Cave Warden enrage transition that shortens future boss cooldowns below a health threshold.
Architecture: Boss enrage is owned by the Go sim and configured on boss templates. The server marks a boss enraged once when HP crosses the threshold, emits a `boss_enraged` event, exposes the state in boss entity views, and applies the configured cooldown multiplier only when scheduling future pattern cooldowns. Client presentation remains optional/debug-level and reuses existing boss health/phase UI.
Tech stack: shared JSON/schema, Go sim/rules/tests, protocol bot, optional Godot client bot, docs.

## Baseline and shortcut decision

Builds on v204 clean CI and existing boss floor/pattern infrastructure from v35/v58/v177/v178/v195. Asset/plugin decision: reject external assets/plugins because enrage uses current boss UI/debug surfaces and server-authored events; production VFX/audio remain deferred.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/boss_templates.v0.json` | Add Cave Warden enrage threshold and cooldown multiplier. |
| Modify | `shared/rules/boss_templates.v0.schema.json` | Validate enrage object shape and ranges. |
| Modify | `server/internal/game/rules.go` | Load enrage fields on boss templates. |
| Add | `server/internal/game/boss_template_rules.go` | Focused enrage validation helper for boss templates. |
| Modify | `server/internal/game/sim.go` | Add boss enrage runtime fields if entity struct/view wiring lives there. |
| Modify | `server/internal/game/types.go` | Expose additive boss enrage view/event fields. |
| Modify | `server/internal/game/boss_patterns.go` | Detect threshold crossing, emit event, and apply cooldown multiplier. |
| Add | `server/internal/game/boss_enrage_test.go` | Focused validation/state/cooldown tests. |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Additive event/entity fields if protocol validation requires them. |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Additive boss entity fields if snapshots expose them. |
| Add | `tools/bot/scenarios/87_boss_enrage_phase.json` | Protocol proof for `boss_enraged`. |
| Modify | `PROGRESS.md` | Mark v205 complete when shipped. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v205 lifecycle row. |
| Add | `docs/as-built/v205_boss-enrage-phase.md` | Capture shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/types.go`
- [x] `server/internal/game/game_test.go` not expected; add focused test file instead.
- [x] `tools/bot/run.py` only if needed.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: enrage tests go in `boss_enrage_test.go`; rule validation helper goes in `boss_template_rules.go`; runtime helpers stay in already focused `boss_patterns.go`. Avoided expanding `game_test.go` and `tools/bot/run.py`.

Verification:
```bash
make maintainability
```

## Task 1 — Shared enrage config

Files:
- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `shared/rules/boss_templates.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add an `enrage` object for Cave Warden with `health_ratio_threshold` and `cooldown_multiplier`.
- [x] Step 1.2: Extend schema validation for threshold `(0,1]` and positive cooldown multiplier.
- [x] Step 1.3: Extend Go template structs/validation with matching checks.
```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestBoss.*Enrage' -count=1
```

## Task 2 — Server enrage runtime

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/boss_patterns.go`
- Add: `server/internal/game/boss_enrage_test.go`

- [x] Step 2.1: Add boss runtime/view fields for `enraged` and enrage threshold data.
- [x] Step 2.2: Detect threshold crossing once before boss phase advancement.
- [x] Step 2.3: Emit `boss_enraged` with boss/template ids and threshold ratio; the same tick includes an entity update carrying HP/max HP and enraged state.
- [x] Step 2.4: Apply cooldown multiplier to future pattern cooldown scheduling, with a minimum 1 tick when base cooldown is positive.
- [x] Step 2.5: Add focused tests for validation, one-shot event/state, view exposure, and cooldown shortening.
```bash
cd server && go test ./internal/game/... -run 'TestBoss.*Enrage|TestBossPatternDeckCycles' -count=1
```

## Task 3 — Protocol and bot proof

Files:
- Modify: `shared/protocol/state_delta.v8.schema.json`
- Modify: `shared/protocol/session_snapshot.v8.schema.json`
- Add or Modify: `tools/bot/scenarios/NN_boss_enrage_phase.json` or `tools/bot/scenarios/24_boss_floor_gate.json`
- Modify: `tools/bot/run.py` only if needed

- [x] Step 3.1: Add schema coverage for additive boss entity/event fields when exposed.
- [x] Step 3.2: Add a protocol bot proof that damages Cave Warden below threshold and observes `boss_enraged`.
- [x] Step 3.3: Keep existing boss-floor gate scenario green; boss special drops remain covered by focused CI/bot coverage.
```bash
make bot scenario=boss_enrage_phase
make bot scenario=boss_floor_gate
```

## Task 4 — Client proof and lifecycle docs

Files:
- Modify: client boss UI/bot files only if needed for debug state
- Modify: `docs/specs/v205_spec-boss-enrage-phase.md`
- Modify: `docs/plans/v205_2026-06-15-boss-enrage-phase.md`
- Add: `docs/as-built/v205_boss-enrage-phase.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: No client UI change was required; protocol/state schema coverage and server view tests cover the additive entity fields.
- [x] Step 4.2: Mark spec/plan complete, add as-built proof, lifecycle row, and update progress.
```bash
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestBossEnrage|TestBossSummonedAdds'`
- [x] `make bot scenario=boss_enrage_phase`
- [x] `make bot scenario=boss_floor_gate`
- [x] `make ci`

## Deferred scope

- Damage multiplier enrage, new attack decks, weighted/random phase selection, bespoke client VFX/audio, boss portraits, co-op-specific enrage scaling, and production art remain future boss/combat work.
